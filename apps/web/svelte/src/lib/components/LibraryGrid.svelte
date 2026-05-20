<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/state';
  import { goto } from '$app/navigation';
  import { ChevronDown, Play, Plus, RotateCcw, Search, SlidersHorizontal, X } from "lucide-svelte";
  import type { Media } from "$lib/mock-data";

  let { eyebrow, title, tagline, items, kind, loading = false, baseHref = "" } = $props<{
    eyebrow: string;
    title: string;
    tagline: string;
    items: Media[];
    kind: "Movies" | "TV";
    loading?: boolean;
    baseHref?: string;
  }>();

  // ── Types ──────────────────────────────────────────────────────────────────
  type Density = "S" | "M" | "L";
  type Sort =
    | "az" | "za"
    | "year-desc" | "year-asc"
    | "rating-desc"
    | "runtime-asc" | "runtime-desc"
    | "parental-asc"
    | "random";

  // ── Density grid ──────────────────────────────────────────────────────────
  const densityGrid: Record<Density, string> = {
    S: "grid-cols-3 sm:grid-cols-5 md:grid-cols-7 lg:grid-cols-8 xl:grid-cols-10",
    M: "grid-cols-2 sm:grid-cols-3 md:grid-cols-5 lg:grid-cols-6 xl:grid-cols-7",
    L: "grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-4 xl:grid-cols-5"
  };

  // ── Sort options ───────────────────────────────────────────────────────────
  const sortOptions: { value: Sort; label: string }[] = [
    { value: "az",            label: "Title A → Z" },
    { value: "za",            label: "Title Z → A" },
    { value: "year-desc",     label: "Year — Newest" },
    { value: "year-asc",      label: "Year — Oldest" },
    { value: "rating-desc",   label: "Rating — Highest" },
    { value: "runtime-asc",   label: "Runtime — Shortest" },
    { value: "runtime-desc",  label: "Runtime — Longest" },
    { value: "parental-asc",  label: "Parental Rating" },
    { value: "random",        label: "Random" },
  ];

  // Parental-rating sort order (G → most restricted)
  const RATING_ORDER = ['G','PG','PG-13','R','NC-17','TV-Y','TV-Y7','TV-Y7-FV','TV-G','TV-PG','TV-14','TV-MA'];

  // ── Filter & UI state ─────────────────────────────────────────────────────
  let q            = $state("");
  let sort         = $state<Sort>("az");
  let density      = $state<Density>("M");
  let sortOpen     = $state(false);
  let filterOpen   = $state(false);
  let randomSeed   = $state(Date.now());

  // Multi-select filter sets
  let selectedGenres   = $state(new Set<string>());
  let selectedDecades  = $state(new Set<string>());
  let selectedRatings  = $state(new Set<string>());
  let onlyNeedsReview  = $state(false);
  let onlyMultiVersion = $state(false);

  // ── Pseudo-random helper ───────────────────────────────────────────────────
  function pseudoRandom(id: string, seed: number): number {
    let h = seed & 0x7fffffff;
    for (let i = 0; i < id.length; i++) h = (Math.imul(h, 31) + id.charCodeAt(i)) & 0x7fffffff;
    return h / 0x7fffffff;
  }

  // ── Decade helper ─────────────────────────────────────────────────────────
  function getDecade(year: number): string {
    if (!year || year <= 0) return '';
    if (year < 1970) return 'Older';
    return `${Math.floor(year / 10) * 10}s`;
  }

  // ── Derived filter options ─────────────────────────────────────────────────
  type GenreChip = { name: string; count: number };

  const genreChips = $derived.by<GenreChip[]>(() => {
    const counts = new Map<string, number>();
    for (const item of items) {
      for (const g of item.genres ?? []) {
        counts.set(g, (counts.get(g) ?? 0) + 1);
      }
    }
    const arr: GenreChip[] = [];
    for (const [name, count] of counts) arr.push({ name, count });
    arr.sort((a, b) => (b.count - a.count) || a.name.localeCompare(b.name));
    return arr;
  });

  const availableDecades = $derived.by<string[]>(() => {
    const set = new Set<string>();
    for (const item of items) {
      const d = getDecade(item.year);
      if (d) set.add(d);
    }
    return [...set].sort((a, b) => {
      if (a === 'Older') return 1;
      if (b === 'Older') return -1;
      return parseInt(b) - parseInt(a);
    });
  });

  const availableRatings = $derived.by<string[]>(() => {
    const set = new Set<string>();
    for (const item of items) {
      const r = item.contentRating?.trim().toUpperCase() || 'NR';
      set.add(r);
    }
    return [...set].sort((a, b) => {
      const ai = RATING_ORDER.indexOf(a);
      const bi = RATING_ORDER.indexOf(b);
      if (ai >= 0 && bi >= 0) return ai - bi;
      if (ai >= 0) return -1;
      if (bi >= 0) return 1;
      return a.localeCompare(b);
    });
  });

  const hasNeedsReview     = $derived(items.some((i: Media) => i.needsReview));
  const hasMultipleVersion = $derived(items.some((i: Media) => (i.versionCount ?? 1) > 1));

  // Snap invalid genre selections to empty if genres disappear (library reload).
  $effect(() => {
    if (selectedGenres.size === 0) return;
    const valid = new Set(genreChips.map(g => g.name));
    let changed = false;
    const next = new Set<string>();
    for (const g of selectedGenres) {
      if (valid.has(g)) next.add(g);
      else changed = true;
    }
    if (changed) selectedGenres = next;
  });

  // ── Active filter count (for badge) ───────────────────────────────────────
  const activeFilterCount = $derived(
    selectedGenres.size + selectedDecades.size + selectedRatings.size +
    (onlyNeedsReview ? 1 : 0) + (onlyMultiVersion ? 1 : 0)
  );

  // ── Sort label ─────────────────────────────────────────────────────────────
  const sortLabel = $derived(sortOptions.find(o => o.value === sort)?.label ?? "Sort");

  // ── Filtered + sorted list ─────────────────────────────────────────────────
  const filtered = $derived.by(() => {
    let list = items.filter((item: Media) => {
      if (q && !item.title.toLowerCase().includes(q.toLowerCase())) return false;
      if (selectedGenres.size > 0 && !(item.genres ?? []).some(g => selectedGenres.has(g))) return false;
      if (selectedDecades.size > 0 && !selectedDecades.has(getDecade(item.year))) return false;
      if (selectedRatings.size > 0) {
        const r = item.contentRating?.trim().toUpperCase() || 'NR';
        if (!selectedRatings.has(r)) return false;
      }
      if (onlyNeedsReview && !item.needsReview) return false;
      if (onlyMultiVersion && (item.versionCount ?? 1) <= 1) return false;
      return true;
    });

    switch (sort) {
      case 'za':           list = [...list].sort((a, b) => b.title.localeCompare(a.title)); break;
      case 'year-desc':    list = [...list].sort((a, b) => (b.year || 0) - (a.year || 0)); break;
      case 'year-asc':     list = [...list].sort((a, b) => (a.year || 0) - (b.year || 0)); break;
      case 'rating-desc':  list = [...list].sort((a, b) => (b.rating || 0) - (a.rating || 0)); break;
      case 'runtime-asc':  list = [...list].sort((a, b) => (a.runtimeMins || 9999) - (b.runtimeMins || 9999)); break;
      case 'runtime-desc': list = [...list].sort((a, b) => (b.runtimeMins || 0) - (a.runtimeMins || 0)); break;
      case 'parental-asc': list = [...list].sort((a, b) => {
        const ra = RATING_ORDER.indexOf(a.contentRating?.trim().toUpperCase() || 'NR');
        const rb = RATING_ORDER.indexOf(b.contentRating?.trim().toUpperCase() || 'NR');
        return (ra < 0 ? 999 : ra) - (rb < 0 ? 999 : rb);
      }); break;
      case 'random': {
        const seed = randomSeed;
        list = [...list].sort((a, b) => pseudoRandom(a.id, seed) - pseudoRandom(b.id, seed));
        break;
      }
      case 'az':
      default:
        list = [...list].sort((a, b) => a.title.localeCompare(b.title));
    }
    return list;
  });

  const featured = $derived(filtered[0] ?? items[0]);

  // ── Toggle helpers ─────────────────────────────────────────────────────────
  function toggleGenre(name: string) {
    const next = new Set(selectedGenres);
    if (next.has(name)) next.delete(name); else next.add(name);
    selectedGenres = next;
  }
  function toggleDecade(d: string) {
    const next = new Set(selectedDecades);
    if (next.has(d)) next.delete(d); else next.add(d);
    selectedDecades = next;
  }
  function toggleRating(r: string) {
    const next = new Set(selectedRatings);
    if (next.has(r)) next.delete(r); else next.add(r);
    selectedRatings = next;
  }
  function clearFilters() {
    selectedGenres   = new Set();
    selectedDecades  = new Set();
    selectedRatings  = new Set();
    onlyNeedsReview  = false;
    onlyMultiVersion = false;
  }

  // ── URL state ──────────────────────────────────────────────────────────────
  // Read initial state from URL params on mount, then sync changes back.
  onMount(() => {
    const p = page.url.searchParams;
    const sortParam = p.get('sort') as Sort | null;
    if (sortParam && sortOptions.some(o => o.value === sortParam)) sort = sortParam;
    const qParam = p.get('q');
    if (qParam) q = qParam;
    const g = p.get('genres');
    if (g) selectedGenres = new Set(g.split(',').filter(Boolean));
    const dec = p.get('decades');
    if (dec) selectedDecades = new Set(dec.split(',').filter(Boolean));
    const rat = p.get('ratings');
    if (rat) selectedRatings = new Set(rat.split(',').filter(Boolean));
    if (p.get('review') === '1') onlyNeedsReview = true;
    if (p.get('multi') === '1') onlyMultiVersion = true;
    const seed = parseInt(p.get('seed') ?? '', 10);
    if (!isNaN(seed) && seed > 0) randomSeed = seed;
  });

  // Sync filter state → URL (replaceState, no scroll, no focus loss)
  $effect(() => {
    const p = new URLSearchParams();
    if (q) p.set('q', q);
    if (sort !== 'az') p.set('sort', sort);
    if (sort === 'random') p.set('seed', String(randomSeed));
    if (selectedGenres.size)  p.set('genres',  [...selectedGenres].sort().join(','));
    if (selectedDecades.size) p.set('decades', [...selectedDecades].sort().join(','));
    if (selectedRatings.size) p.set('ratings', [...selectedRatings].sort().join(','));
    if (onlyNeedsReview)  p.set('review', '1');
    if (onlyMultiVersion) p.set('multi', '1');
    const qs = p.toString();
    const cur = page.url.searchParams.toString();
    if (qs !== cur) {
      goto(`?${qs}`, { replaceState: true, noScroll: true, keepFocus: true });
    }
  });
</script>

<main class="pb-32">
  <!-- ── Hero header ──────────────────────────────────────────────────────── -->
  <section class="relative isolate overflow-hidden px-6 pb-10 pt-32 md:px-12 md:pb-14 md:pt-40 lg:px-20">
    {#if featured?.backdrop}
      {#key featured.id}
        <img
          src={featured.backdrop}
          alt=""
          aria-hidden="true"
          class="pointer-events-none absolute inset-x-0 top-0 -z-20 h-full w-full object-cover object-top"
          style="opacity: 0.35;"
          onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
        />
      {/key}
    {/if}
    <div class="pointer-events-none absolute inset-x-0 bottom-0 -z-10 h-2/3 bg-gradient-to-t from-background to-transparent"></div>
    <div
      aria-hidden="true"
      class="pointer-events-none absolute inset-x-0 top-0 -z-10 h-[520px] opacity-70"
      style={`background: ${featured
        ? `radial-gradient(60% 70% at 20% 0%, ${featured.palette[0]}55, transparent 60%), radial-gradient(50% 60% at 90% 10%, ${featured.palette[1]}40, transparent 70%)`
        : ""}`}
    ></div>
    <div class="grain pointer-events-none absolute inset-0 -z-10"></div>

    <div class="grid items-end gap-10 md:grid-cols-[1.6fr_1fr] lg:grid-cols-[2fr_1fr]">
      <div>
        <div class="mb-3 text-[10px] font-semibold uppercase tracking-[0.35em] text-primary-glow">{eyebrow}</div>
        <h1 class="font-serif-display text-[clamp(2.6rem,6vw,5rem)] leading-[0.95] tracking-tight">{title}</h1>
        <p class="mt-5 max-w-xl text-sm leading-relaxed text-muted-foreground md:text-base">{tagline}</p>
        <div class="mt-6 flex flex-wrap gap-x-6 gap-y-2 text-[11px] uppercase tracking-[0.22em] text-muted-foreground">
          <span><span class="text-foreground/90">{items.length}</span> in library</span>
          <span class="opacity-30">·</span>
          <span><span class="text-foreground/90">{genreChips.length}</span> genres</span>
        </div>
      </div>

      {#if featured}
        <article
          class="shadow-poster relative ml-auto aspect-[2/3] w-full max-w-[260px] overflow-hidden rounded-2xl md:max-w-[280px] lg:max-w-[320px]"
          style={`background: linear-gradient(135deg, ${featured.palette[0]}, ${featured.palette[1]})`}
        >
          {#if featured.poster}
            <img src={featured.poster} alt={featured.title} class="absolute inset-0 h-full w-full object-cover" onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')} />
          {:else}
            <div class="absolute inset-0 opacity-40 mix-blend-overlay" style="background-image: radial-gradient(circle at 25% 15%, rgba(255,255,255,0.45), transparent 50%), radial-gradient(circle at 80% 85%, rgba(0,0,0,0.6), transparent 60%);"></div>
          {/if}
          <div class="absolute inset-0 bg-gradient-to-t from-black/80 via-black/10 to-transparent"></div>
          <div class="absolute inset-0 flex flex-col justify-between p-5">
            <div class="flex items-center gap-2 text-[9px] font-semibold uppercase tracking-[0.3em] text-white/85">
              <span class="h-1.5 w-1.5 rounded-full" style={`background: ${featured.accent}; box-shadow: 0 0 12px ${featured.accent};`}></span>
              Editor's pick
            </div>
            <div>
              <div class="text-[10px] uppercase tracking-[0.28em] text-white/65">
                {featured.year || ""}
                {#if featured.type === "Series" && featured.seasons}
                  {featured.year ? " · " : ""}{featured.seasons} Season{featured.seasons !== 1 ? "s" : ""}
                {:else if featured.runtime}
                  {featured.year ? " · " : ""}{featured.runtime}
                {/if}
              </div>
              <h2 class="font-serif-display mt-1.5 text-2xl leading-[0.95] text-white md:text-3xl" style="text-shadow: 0 4px 24px rgba(0,0,0,0.5);">
                {featured.title}
              </h2>
              <div class="mt-4 flex gap-2">
                <a
                  href={baseHref ? `${baseHref}/${featured.id}` : (featured.type === 'Series' ? `/tv/${featured.id}` : `/movies/${featured.id}`)}
                  class="inline-flex items-center gap-1.5 rounded-full bg-white px-4 py-2 text-xs font-semibold text-black hover:bg-white/90"
                >
                  <Play class="h-3.5 w-3.5 fill-black" /> Play
                </a>
                <button aria-label="Add to list" class="hairline flex h-9 w-9 items-center justify-center rounded-full bg-white/10 text-white backdrop-blur-md hover:bg-white/20">
                  <Plus class="h-4 w-4" />
                </button>
              </div>
            </div>
          </div>
        </article>
      {/if}
    </div>
  </section>

  <!-- ── Toolbar (sticky) ─────────────────────────────────────────────────── -->
  <div class="sticky top-16 z-30 -mb-px border-y border-border bg-background/75 backdrop-blur-xl md:top-18">
    <div class="flex flex-wrap items-center gap-2 px-6 py-3 md:px-12 lg:px-20">

      <!-- Search -->
      <div class="relative min-w-0 flex-1 max-w-xs">
        <Search class="pointer-events-none absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <input
          bind:value={q}
          placeholder={`Search ${kind === "TV" ? "TV shows" : "movies"}...`}
          class="h-9 w-full rounded-full border border-border bg-surface/60 pl-10 pr-4 text-sm outline-none placeholder:text-muted-foreground/70 focus:border-primary/60 focus:bg-surface"
        />
        {#if q}
          <button type="button" onclick={() => (q = '')} class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground" aria-label="Clear search">
            <X class="h-3.5 w-3.5" />
          </button>
        {/if}
      </div>

      <!-- Filters toggle button -->
      <div class="relative">
        <button
          type="button"
          onclick={() => { filterOpen = !filterOpen; sortOpen = false; }}
          class={`hairline flex items-center gap-1.5 rounded-full px-3 py-1.5 text-xs font-medium transition-colors ${
            filterOpen || activeFilterCount > 0
              ? 'bg-primary-glow/15 text-foreground ring-1 ring-primary-glow/30'
              : 'bg-foreground/[0.04] text-muted-foreground hover:bg-foreground/[0.08] hover:text-foreground'
          }`}
          aria-expanded={filterOpen}
        >
          <SlidersHorizontal class="h-3.5 w-3.5" />
          Filters
          {#if activeFilterCount > 0}
            <span class="flex h-4 min-w-4 items-center justify-center rounded-full bg-primary-glow px-1 text-[10px] font-bold text-white">
              {activeFilterCount}
            </span>
          {/if}
          <ChevronDown class={`h-3.5 w-3.5 transition-transform ${filterOpen ? 'rotate-180' : ''}`} />
        </button>
      </div>

      <div class="ml-auto flex items-center gap-2">
        <!-- Sort picker -->
        <div class="flex items-center gap-1">
          <!-- Shuffle button shown alongside when random is active -->
          {#if sort === 'random'}
            <button
              type="button"
              title="Shuffle"
              onclick={() => { randomSeed = Date.now(); }}
              class="flex h-7 w-7 items-center justify-center rounded-full bg-foreground/[0.04] text-muted-foreground transition-colors hover:bg-foreground/[0.08] hover:text-foreground"
              aria-label="Shuffle"
            >
              <RotateCcw class="h-3 w-3" />
            </button>
          {/if}
          <div class="relative">
            <button
              type="button"
              onclick={() => { sortOpen = !sortOpen; filterOpen = false; }}
              class={`hairline flex items-center gap-1.5 rounded-full bg-foreground/[0.04] px-3 py-1.5 text-xs transition-colors hover:bg-foreground/[0.08] ${sortOpen ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'}`}
            >
              {sortLabel}
              <ChevronDown class={`h-3.5 w-3.5 transition-transform ${sortOpen ? 'rotate-180' : ''}`} />
            </button>
            {#if sortOpen}
              <button type="button" class="fixed inset-0 z-40 cursor-default" aria-hidden="true" tabindex="-1" onclick={() => (sortOpen = false)}></button>
              <div class="absolute right-0 top-full z-50 mt-1.5 min-w-[172px] overflow-hidden rounded-xl border border-border bg-surface/95 py-1 shadow-xl backdrop-blur-xl">
                {#each sortOptions as opt (opt.value)}
                  <button
                    type="button"
                    onclick={() => { sort = opt.value; sortOpen = false; if (opt.value === 'random') randomSeed = Date.now(); }}
                    class={`flex w-full items-center px-3.5 py-2 text-left text-xs transition-colors ${
                      sort === opt.value
                        ? 'bg-foreground/[0.07] font-medium text-foreground'
                        : 'text-muted-foreground hover:bg-foreground/[0.05] hover:text-foreground'
                    }`}
                  >
                    {opt.label}
                  </button>
                {/each}
              </div>
            {/if}
          </div>
        </div>

        <!-- Density switcher -->
        <div class="hairline hidden items-center rounded-full bg-foreground/[0.04] p-1 md:flex">
          {#each (["S", "M", "L"] as const) as option (option)}
            <button
              type="button"
              aria-label={option === "S" ? "Small cards" : option === "M" ? "Medium cards" : "Large cards"}
              onclick={() => (density = option)}
              class={`flex h-7 min-w-7 items-center justify-center rounded-full px-2.5 text-[11px] font-semibold tracking-wider transition-colors ${
                density === option ? 'bg-foreground text-background' : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              {option}
            </button>
          {/each}
        </div>
      </div>
    </div>

    <!-- ── Filter panel ──────────────────────────────────────────────────── -->
    {#if filterOpen}
      <div class="border-t border-border bg-surface/30 px-6 py-5 md:px-12 lg:px-20">
        <div class="space-y-5">

          <!-- Genre -->
          {#if genreChips.length > 0}
            <div>
              <div class="mb-2 text-[10px] font-semibold uppercase tracking-[0.25em] text-muted-foreground">Genre</div>
              <div class="flex flex-wrap gap-1.5">
                {#each genreChips as chip (chip.name)}
                  {@const active = selectedGenres.has(chip.name)}
                  <button
                    type="button"
                    onclick={() => toggleGenre(chip.name)}
                    class={`inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-medium transition-colors ${
                      active
                        ? 'bg-foreground text-background'
                        : 'bg-foreground/[0.05] text-foreground/70 hover:bg-foreground/[0.10] hover:text-foreground'
                    }`}
                    aria-pressed={active}
                  >
                    {chip.name}
                    <span class={`tabular-nums text-[10px] ${active ? 'text-background/60' : 'text-foreground/40'}`}>{chip.count}</span>
                  </button>
                {/each}
              </div>
            </div>
          {/if}

          <!-- Year / decade -->
          {#if availableDecades.length > 0}
            <div>
              <div class="mb-2 text-[10px] font-semibold uppercase tracking-[0.25em] text-muted-foreground">Year</div>
              <div class="flex flex-wrap gap-1.5">
                {#each availableDecades as decade (decade)}
                  {@const active = selectedDecades.has(decade)}
                  <button
                    type="button"
                    onclick={() => toggleDecade(decade)}
                    class={`rounded-full px-3 py-1 text-xs font-medium transition-colors ${
                      active
                        ? 'bg-foreground text-background'
                        : 'bg-foreground/[0.05] text-foreground/70 hover:bg-foreground/[0.10] hover:text-foreground'
                    }`}
                    aria-pressed={active}
                  >
                    {decade}
                  </button>
                {/each}
              </div>
            </div>
          {/if}

          <!-- Parental rating -->
          {#if availableRatings.length > 0}
            <div>
              <div class="mb-2 text-[10px] font-semibold uppercase tracking-[0.25em] text-muted-foreground">Rating</div>
              <div class="flex flex-wrap gap-1.5">
                {#each availableRatings as r (r)}
                  {@const active = selectedRatings.has(r)}
                  <button
                    type="button"
                    onclick={() => toggleRating(r)}
                    class={`rounded-full border px-3 py-1 text-xs font-medium transition-colors ${
                      active
                        ? 'border-foreground/60 bg-foreground text-background'
                        : 'border-border bg-transparent text-foreground/70 hover:border-foreground/30 hover:text-foreground'
                    }`}
                    aria-pressed={active}
                  >
                    {r}
                  </button>
                {/each}
              </div>
            </div>
          {/if}

          <!-- Flags -->
          {#if hasNeedsReview || hasMultipleVersion}
            <div>
              <div class="mb-2 text-[10px] font-semibold uppercase tracking-[0.25em] text-muted-foreground">Show only</div>
              <div class="flex flex-wrap gap-1.5">
                {#if hasNeedsReview}
                  <button
                    type="button"
                    onclick={() => (onlyNeedsReview = !onlyNeedsReview)}
                    class={`rounded-full px-3 py-1 text-xs font-medium transition-colors ${
                      onlyNeedsReview
                        ? 'bg-amber-500/20 text-amber-300 ring-1 ring-amber-500/30'
                        : 'bg-foreground/[0.05] text-foreground/70 hover:bg-foreground/[0.10] hover:text-foreground'
                    }`}
                    aria-pressed={onlyNeedsReview}
                  >
                    Needs Review
                  </button>
                {/if}
                {#if hasMultipleVersion}
                  <button
                    type="button"
                    onclick={() => (onlyMultiVersion = !onlyMultiVersion)}
                    class={`rounded-full px-3 py-1 text-xs font-medium transition-colors ${
                      onlyMultiVersion
                        ? 'bg-primary-glow/15 text-foreground ring-1 ring-primary-glow/30'
                        : 'bg-foreground/[0.05] text-foreground/70 hover:bg-foreground/[0.10] hover:text-foreground'
                    }`}
                    aria-pressed={onlyMultiVersion}
                  >
                    Multiple Versions
                  </button>
                {/if}
              </div>
            </div>
          {/if}

          <!-- Clear / status row -->
          <div class="flex items-center justify-between pt-1">
            <p class="text-xs text-muted-foreground">
              {filtered.length === items.length
                ? `${items.length} ${kind === 'TV' ? 'shows' : 'movies'}`
                : `${filtered.length} of ${items.length} ${kind === 'TV' ? 'shows' : 'movies'}`}
            </p>
            {#if activeFilterCount > 0}
              <button
                type="button"
                onclick={clearFilters}
                class="flex items-center gap-1.5 rounded-full bg-foreground/[0.06] px-3 py-1 text-xs text-muted-foreground transition-colors hover:bg-foreground/[0.10] hover:text-foreground"
              >
                <X class="h-3 w-3" /> Clear {activeFilterCount} filter{activeFilterCount !== 1 ? 's' : ''}
              </button>
            {/if}
          </div>
        </div>
      </div>
    {/if}
  </div>

  <!-- ── Grid ─────────────────────────────────────────────────────────────── -->
  <section class="px-6 pt-10 md:px-12 lg:px-20">
    {#if loading}
      <div class={`grid gap-x-5 gap-y-10 ${densityGrid[density]}`}>
        {#each { length: 18 } as _, i (i)}
          <div class="animate-pulse">
            <div class="aspect-[2/3] rounded-xl bg-foreground/[0.07]"></div>
            {#if density !== "S"}
              <div class="mt-3 h-4 w-3/4 rounded bg-foreground/[0.07]"></div>
              <div class="mt-1.5 h-3 w-1/2 rounded bg-foreground/[0.05]"></div>
            {/if}
          </div>
        {/each}
      </div>
    {:else if filtered.length === 0}
      <div class="hairline flex flex-col items-center justify-center rounded-3xl bg-surface/30 py-24 text-center">
        <div class="font-serif-display text-3xl">Nothing matches that</div>
        <p class="mt-2 max-w-sm text-sm text-muted-foreground">
          Try different filters or clear them to see your full library.
        </p>
        {#if activeFilterCount > 0 || q}
          <button
            type="button"
            onclick={() => { clearFilters(); q = ''; }}
            class="mt-6 inline-flex items-center gap-2 rounded-full bg-foreground/[0.06] px-5 py-2.5 text-sm text-muted-foreground transition-colors hover:bg-foreground/[0.10] hover:text-foreground"
          >
            <X class="h-4 w-4" /> Clear all filters
          </button>
        {/if}
      </div>
    {:else}
      <div class={`grid gap-x-5 gap-y-10 ${densityGrid[density]}`}>
        {#each filtered as media (media.id)}
          <a
            href={baseHref ? `${baseHref}/${media.id}` : (media.type === "Series" ? `/tv/${media.id}` : `/movies/${media.id}`)}
            class="group block"
          >
            <div
              class="shadow-poster relative aspect-[2/3] overflow-hidden rounded-xl transition-all duration-500 group-hover:-translate-y-1.5 group-hover:shadow-glow"
              style={`background: linear-gradient(135deg, ${media.palette[0]}, ${media.palette[1]})`}
            >
              {#if media.poster}
                <img
                  src={media.poster}
                  alt={media.title}
                  loading="lazy"
                  class="absolute inset-0 h-full w-full object-cover"
                  onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
                />
              {:else}
                <div class="absolute inset-0 opacity-40 mix-blend-overlay" style="background-image: radial-gradient(circle at 30% 20%, rgba(255,255,255,0.4), transparent 50%), radial-gradient(circle at 80% 80%, rgba(0,0,0,0.55), transparent 60%);"></div>
              {/if}
              <div class="absolute inset-x-0 bottom-0 h-1/2 bg-gradient-to-t from-black/80 to-transparent"></div>
              <div class="absolute inset-x-3 bottom-3">
                <h3 class="font-display text-[13px] font-semibold leading-tight text-white md:text-sm" style="text-shadow: 0 2px 12px rgba(0,0,0,0.7);">
                  {media.title}
                </h3>
              </div>
              <!-- Needs review badge -->
              {#if media.needsReview}
                <div class="absolute right-2 top-2 rounded-md bg-amber-500/80 px-1.5 py-0.5 text-[9px] font-bold uppercase tracking-wide text-black backdrop-blur-sm">
                  Review
                </div>
              {/if}
              <div class="absolute inset-0 flex items-center justify-center bg-black/35 opacity-0 backdrop-blur-[1px] transition-opacity duration-300 group-hover:opacity-100">
                <button aria-label={`Play ${media.title}`} class="flex h-12 w-12 items-center justify-center rounded-full bg-white text-black ring-1 ring-white/40">
                  <Play class="h-5 w-5 translate-x-0.5 fill-black" />
                </button>
              </div>
            </div>
            {#if density !== "S"}
              <div class="mt-3 flex items-baseline justify-between gap-2 px-0.5">
                <h4 class="truncate text-sm font-medium text-foreground">{media.title}</h4>
                <span class="shrink-0 text-[11px] tabular-nums text-muted-foreground">{media.year || ''}</span>
              </div>
              <p class="mt-0.5 truncate text-[11px] uppercase tracking-[0.18em] text-muted-foreground">
                {media.genres.slice(0, 2).join(" · ")}
                {#if typeof media.rating === "number" && media.rating > 0}
                  <span class="ml-2 normal-case tracking-normal text-foreground/70">★ {media.rating.toFixed(1)}</span>
                {/if}
              </p>
            {/if}
          </a>
        {/each}
      </div>
    {/if}
  </section>
</main>
