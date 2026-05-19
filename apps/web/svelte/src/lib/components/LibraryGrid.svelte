<script lang="ts">
  import { Play, Plus, Search, SlidersHorizontal } from "lucide-svelte";
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

  type Density = "S" | "M" | "L";
  type Sort = "trending" | "rating" | "year-desc" | "az";

  const densityGrid: Record<Density, string> = {
    S: "grid-cols-3 sm:grid-cols-5 md:grid-cols-7 lg:grid-cols-8 xl:grid-cols-10",
    M: "grid-cols-2 sm:grid-cols-3 md:grid-cols-5 lg:grid-cols-6 xl:grid-cols-7",
    L: "grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-4 xl:grid-cols-5"
  };

  let q = $state("");
  let genre = $state("All");
  let sort = $state<Sort>("trending");
  let density = $state<Density>("M");
  let sortOpen = $state(false);

  const sortOptions: { value: Sort; label: string }[] = [
    { value: "trending",  label: "Trending" },
    { value: "rating",    label: "Highest rated" },
    { value: "year-desc", label: "Newest first" },
    { value: "az",        label: "A → Z" },
  ];
  const sortLabel = $derived(sortOptions.find((o) => o.value === sort)?.label ?? "Sort");

  // Aggregate genres with counts so chips show "Action 47" and the strip is
  // sorted by populousness — what the user actually has the most of is the
  // first thing they see. "All" leads, then descending count, ties broken
  // alphabetically.
  type GenreChip = { name: string; count: number };
  let genreChips = $derived.by<GenreChip[]>(() => {
    const counts = new Map<string, number>();
    for (const item of items) {
      for (const g of item.genres ?? []) {
        counts.set(g, (counts.get(g) ?? 0) + 1);
      }
    }
    const arr: GenreChip[] = [];
    for (const [name, count] of counts) arr.push({ name, count });
    arr.sort((a, b) => (b.count - a.count) || a.name.localeCompare(b.name));
    return [{ name: "All", count: items.length }, ...arr];
  });

  // If the active filter selection disappears (library updates, switched
  // libraries, etc.) snap back to "All" so the grid never silently shows an
  // empty result for a vanished selection.
  $effect(() => {
    if (genre === "All") return;
    if (!genreChips.some((g) => g.name === genre)) {
      genre = "All";
    }
  });

  let filtered = $derived.by(() => {
    let list = items.filter((item: Media) => {
      const matchesGenre = genre === "All" || (item.genres ?? []).includes(genre);
      const matchesQ = !q || item.title.toLowerCase().includes(q.toLowerCase());
      return matchesGenre && matchesQ;
    });

    const sorters: Record<Sort, (a: Media, b: Media) => number> = {
      trending: (a, b) => b.rating - a.rating,
      rating: (a, b) => b.rating - a.rating,
      "year-desc": (a, b) => b.year - a.year,
      az: (a, b) => a.title.localeCompare(b.title)
    };

    return [...list].sort(sorters[sort]);
  });

  let featured = $derived(filtered[0] ?? items[0]);
</script>

<main class="pb-32">
  <section class="relative isolate overflow-hidden px-6 pb-10 pt-32 md:px-12 md:pb-14 md:pt-40 lg:px-20">
    <!-- Backdrop photo — atmospheric base layer, fades toward bottom -->
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
    <!-- Bottom fade so the backdrop doesn't bleed into the filter bar -->
    <div class="pointer-events-none absolute inset-x-0 bottom-0 -z-10 h-2/3 bg-gradient-to-t from-background to-transparent"></div>
    <!-- Radial colour accent from the featured item's palette -->
    <div
      aria-hidden="true"
      class="pointer-events-none absolute inset-x-0 top-0 -z-10 h-[520px] opacity-70"
      style={`background: ${
        featured
          ? `radial-gradient(60% 70% at 20% 0%, ${featured.palette[0]}55, transparent 60%), radial-gradient(50% 60% at 90% 10%, ${featured.palette[1]}40, transparent 70%)`
          : ""
      }`}
    ></div>
    <div class="grain pointer-events-none absolute inset-0 -z-10"></div>

    <div class="grid items-end gap-10 md:grid-cols-[1.6fr_1fr] lg:grid-cols-[2fr_1fr]">
      <div>
        <div class="mb-3 text-[10px] font-semibold uppercase tracking-[0.35em] text-primary-glow">
          {eyebrow}
        </div>
        <h1 class="font-serif-display text-[clamp(2.6rem,6vw,5rem)] leading-[0.95] tracking-tight">
          {title}
        </h1>
        <p class="mt-5 max-w-xl text-sm leading-relaxed text-muted-foreground md:text-base">
          {tagline}
        </p>
        <div class="mt-6 flex flex-wrap gap-x-6 gap-y-2 text-[11px] uppercase tracking-[0.22em] text-muted-foreground">
          <span><span class="text-foreground/90">{items.length}</span> in library</span>
          <span class="opacity-30">·</span>
          <span><span class="text-foreground/90">{genreChips.length - 1}</span> genres</span>
          <span class="opacity-30">·</span>
          <span><span class="text-foreground/90">4K · HDR · Atmos</span></span>
        </div>
      </div>

      {#if featured}
        <article
          class="shadow-poster relative ml-auto aspect-[2/3] w-full max-w-[260px] overflow-hidden rounded-2xl md:max-w-[280px] lg:max-w-[320px]"
          style={`background: linear-gradient(135deg, ${featured.palette[0]}, ${featured.palette[1]})`}
        >
          {#if featured.poster}
            <img
              src={featured.poster}
              alt={featured.title}
              class="absolute inset-0 h-full w-full object-cover"
              onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
            />
          {:else}
            <div
              class="absolute inset-0 opacity-40 mix-blend-overlay"
              style="background-image: radial-gradient(circle at 25% 15%, rgba(255,255,255,0.45), transparent 50%), radial-gradient(circle at 80% 85%, rgba(0,0,0,0.6), transparent 60%);"
            ></div>
          {/if}
          <div class="absolute inset-0 bg-gradient-to-t from-black/80 via-black/10 to-transparent"></div>
          <div class="absolute inset-0 flex flex-col justify-between p-5">
            <div class="flex items-center gap-2 text-[9px] font-semibold uppercase tracking-[0.3em] text-white/85">
              <span
                class="h-1.5 w-1.5 rounded-full"
                style={`background: ${featured.accent}; box-shadow: 0 0 12px ${featured.accent};`}
              ></span>
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
              <h2
                class="font-serif-display mt-1.5 text-2xl leading-[0.95] text-white md:text-3xl"
                style="text-shadow: 0 4px 24px rgba(0,0,0,0.5);"
              >
                {featured.title}
              </h2>
              <div class="mt-4 flex gap-2">
                <button class="inline-flex items-center gap-1.5 rounded-full bg-white px-4 py-2 text-xs font-semibold text-black hover:bg-white/90">
                  <Play class="h-3.5 w-3.5 fill-black" /> Play
                </button>
                <button
                  aria-label="Add to list"
                  class="hairline flex h-9 w-9 items-center justify-center rounded-full bg-white/10 text-white backdrop-blur-md hover:bg-white/20"
                >
                  <Plus class="h-4 w-4" />
                </button>
              </div>
            </div>
          </div>
        </article>
      {/if}
    </div>
  </section>

  <div class="sticky top-16 z-30 -mb-px border-y border-border bg-background/75 backdrop-blur-xl md:top-18">
    <div class="flex flex-wrap items-center gap-3 px-6 py-4 md:px-12 lg:px-20">
      <div class="relative w-full max-w-xs">
        <Search class="pointer-events-none absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <input
          bind:value={q}
          placeholder={`Search ${kind === "TV" ? "TV shows" : "movies"}...`}
          class="h-10 w-full rounded-full border border-border bg-surface/60 pl-10 pr-4 text-sm outline-none placeholder:text-muted-foreground/70 focus:border-primary/60 focus:bg-surface"
        />
      </div>

      <div class="scrollbar-none -mx-2 flex flex-1 items-center gap-1.5 overflow-x-auto px-2">
        {#each genreChips as chip (chip.name)}
          {@const isActive = chip.name === genre}
          <button
            onclick={() => (genre = isActive && chip.name !== "All" ? "All" : chip.name)}
            class={`hairline group/chip shrink-0 inline-flex items-center gap-1.5 rounded-full px-4 py-1.5 text-xs font-medium uppercase tracking-[0.15em] transition-colors ${
              isActive
                ? "bg-foreground text-background"
                : "bg-foreground/[0.04] text-foreground/75 hover:bg-foreground/[0.08] hover:text-foreground"
            }`}
            aria-pressed={isActive}
          >
            <span>{chip.name}</span>
            <span class={`text-[10px] tabular-nums ${isActive ? "text-background/60" : "text-foreground/40"}`}>
              {chip.count}
            </span>
          </button>
        {/each}
      </div>

      <div class="flex items-center gap-2">
        <!-- Custom sort picker — replaces native <select> to avoid
             the OS white-background dropdown that can't be themed. -->
        <div class="relative">
          <button
            type="button"
            onclick={() => (sortOpen = !sortOpen)}
            class={`hairline flex items-center gap-1.5 rounded-full bg-foreground/[0.04] px-3 py-1.5 text-xs transition-colors hover:bg-foreground/[0.08] ${sortOpen ? "text-foreground" : "text-muted-foreground hover:text-foreground"}`}
          >
            <SlidersHorizontal class="h-3.5 w-3.5" />
            {sortLabel}
          </button>
          {#if sortOpen}
            <!-- invisible scrim: click outside closes the menu -->
            <button
              type="button"
              class="fixed inset-0 z-40 cursor-default"
              aria-hidden="true"
              tabindex="-1"
              onclick={() => (sortOpen = false)}
            ></button>
            <div class="absolute right-0 top-full z-50 mt-1.5 min-w-[148px] overflow-hidden rounded-xl border border-border bg-surface/95 py-1 shadow-xl backdrop-blur-xl">
              {#each sortOptions as opt (opt.value)}
                <button
                  type="button"
                  onclick={() => { sort = opt.value; sortOpen = false; }}
                  class={`flex w-full items-center px-3.5 py-2 text-left text-xs transition-colors ${
                    sort === opt.value
                      ? "text-foreground font-medium bg-foreground/[0.07]"
                      : "text-muted-foreground hover:text-foreground hover:bg-foreground/[0.05]"
                  }`}
                >
                  {opt.label}
                </button>
              {/each}
            </div>
          {/if}
        </div>
        <div class="hairline hidden items-center rounded-full bg-foreground/[0.04] p-1 md:flex">
          {#each ["S", "M", "L"] as option (option)}
            <button
              type="button"
              aria-label={`${option === "S" ? "Small" : option === "M" ? "Medium" : "Large"} cards`}
              onclick={() => (density = option as Density)}
              class={`flex h-7 min-w-7 items-center justify-center rounded-full px-2.5 text-[11px] font-semibold tracking-wider transition-colors ${
                density === option ? "bg-foreground text-background" : "text-muted-foreground hover:text-foreground"
              }`}
            >
              {option}
            </button>
          {/each}
        </div>
      </div>
    </div>
  </div>

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
          Try a different genre or clear your search to see your full library.
        </p>
      </div>
    {:else}
      <div class={`grid gap-x-5 gap-y-10 ${densityGrid[density]}`}>
        {#each filtered as media (media.id)}
          <a href={baseHref ? `${baseHref}/${media.id}` : (media.type === "Series" ? `/tv/${media.id}` : `/movies/${media.id}`)} class="group block">
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
                <div
                  class="absolute inset-0 opacity-40 mix-blend-overlay"
                  style="background-image: radial-gradient(circle at 30% 20%, rgba(255,255,255,0.4), transparent 50%), radial-gradient(circle at 80% 80%, rgba(0,0,0,0.55), transparent 60%);"
                ></div>
              {/if}
              <div class="absolute inset-x-0 bottom-0 h-1/2 bg-gradient-to-t from-black/80 to-transparent"></div>
              <div class="absolute inset-x-3 bottom-3">
                <h3
                  class="font-display text-[13px] font-semibold leading-tight text-white md:text-sm"
                  style="text-shadow: 0 2px 12px rgba(0,0,0,0.7);"
                >
                  {media.title}
                </h3>
              </div>
              <div class="absolute inset-0 flex items-center justify-center bg-black/35 opacity-0 backdrop-blur-[1px] transition-opacity duration-300 group-hover:opacity-100">
                <button
                  aria-label={`Play ${media.title}`}
                  class="flex h-12 w-12 items-center justify-center rounded-full bg-white text-black ring-1 ring-white/40"
                >
                  <Play class="h-5 w-5 translate-x-0.5 fill-black" />
                </button>
              </div>
            </div>
            {#if density !== "S"}
              <div class="mt-3 flex items-baseline justify-between gap-2 px-0.5">
                <h4 class="truncate text-sm font-medium text-foreground">{media.title}</h4>
                <span class="shrink-0 text-[11px] tabular-nums text-muted-foreground">
                  {media.year}
                </span>
              </div>
              <p class="mt-0.5 truncate text-[11px] uppercase tracking-[0.18em] text-muted-foreground">
                {media.genres.slice(0, 2).join(" · ")}
                {#if typeof media.rating === "number" && media.rating > 0}
                  <span class="ml-2 normal-case tracking-normal text-foreground/70">
                    ★ {media.rating.toFixed(1)}
                  </span>
                {/if}
              </p>
            {/if}
          </a>
        {/each}
      </div>
    {/if}
  </section>
</main>

