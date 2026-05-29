<script lang="ts">
  import { onMount } from 'svelte';
  import { Bookmark, BookmarkX, ChevronDown, Film, Search, Tv, X } from 'lucide-svelte';
  import Header from '$lib/components/Header.svelte';
  import { appState } from '$lib/stores/appState.svelte';
  import { getAuthSession } from '$lib/api/auth';
  import { updateUserPreferences } from '$lib/api/operator';
  import { getWatchlist, removeFromWatchlist, type WatchlistItem } from '$lib/stores/watchlistStore.svelte';

  // ── Types ──────────────────────────────────────────────────────────────────
  type Density = 'S' | 'M' | 'L';
  type Sort = 'added-desc' | 'added-asc' | 'az' | 'za' | 'year-desc' | 'year-asc';
  type KindFilter = 'all' | 'movie' | 'series';

  const densityGrid: Record<Density, string> = {
    S: 'grid-cols-3 sm:grid-cols-5 md:grid-cols-7 lg:grid-cols-8 xl:grid-cols-10',
    M: 'grid-cols-2 sm:grid-cols-3 md:grid-cols-5 lg:grid-cols-6 xl:grid-cols-7',
    L: 'grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-4 xl:grid-cols-5',
  };

  const sortOptions: { value: Sort; label: string }[] = [
    { value: 'added-desc', label: 'Date Added — Newest' },
    { value: 'added-asc',  label: 'Date Added — Oldest' },
    { value: 'az',         label: 'Title A → Z' },
    { value: 'za',         label: 'Title Z → A' },
    { value: 'year-desc',  label: 'Year — Newest' },
    { value: 'year-asc',   label: 'Year — Oldest' },
  ];

  // ── Reactive items from store ──────────────────────────────────────────────
  const items = $derived(getWatchlist());

  // ── Control state ──────────────────────────────────────────────────────────
  let q          = $state('');
  let sort       = $state<Sort>('added-desc');
  const DENSITY_KEY = 'xuva:poster-size';
  function readCachedDensity(): Density {
    try {
      const v = typeof localStorage !== 'undefined' ? localStorage.getItem(DENSITY_KEY) : null;
      if (v === 'S' || v === 'M' || v === 'L') return v;
    } catch { /* SSR or privacy mode */ }
    return 'M';
  }
  let density    = $state<Density>(readCachedDensity());
  let kindFilter = $state<KindFilter>('all');
  let sortOpen   = $state(false);
  let mounted    = $state(false);

  // ── Derived filtered + sorted list ─────────────────────────────────────────
  const filtered = $derived.by(() => {
    let result = items.slice();

    // Kind filter
    if (kindFilter !== 'all') {
      result = result.filter(i => i.kind === kindFilter);
    }

    // Search
    if (q.trim()) {
      const needle = q.toLowerCase();
      result = result.filter(i => i.title.toLowerCase().includes(needle));
    }

    // Sort
    result.sort((a, b) => {
      switch (sort) {
        case 'added-desc': return new Date(b.addedAt).getTime() - new Date(a.addedAt).getTime();
        case 'added-asc':  return new Date(a.addedAt).getTime() - new Date(b.addedAt).getTime();
        case 'az':         return a.title.localeCompare(b.title);
        case 'za':         return b.title.localeCompare(a.title);
        case 'year-desc':  return (b.year ?? 0) - (a.year ?? 0);
        case 'year-asc':   return (a.year ?? 0) - (b.year ?? 0);
        default:           return 0;
      }
    });

    return result;
  });

  const sortLabel = $derived(sortOptions.find(o => o.value === sort)?.label ?? 'Sort');

  // ── Density ────────────────────────────────────────────────────────────────
  function setDensity(d: Density) {
    density = d;
    try { localStorage.setItem(DENSITY_KEY, d); } catch { /* privacy mode */ }
    updateUserPreferences({ posterSize: d }).catch(() => {});
  }

  onMount(() => {
    mounted = true;
    // Sync with server preference — localStorage already applied above, no flash.
    getAuthSession().then(s => {
      const size = s?.preferences?.posterSize;
      if (size === 'S' || size === 'M' || size === 'L') {
        density = size;
        try { localStorage.setItem(DENSITY_KEY, size); } catch { /* privacy mode */ }
      }
    }).catch(() => {});
  });
</script>

<svelte:head>
  <title>Watchlist — {appState.serverName}</title>
  <meta name="description" content="Your personal Xuva watchlist — films and shows you've saved to watch later." />
</svelte:head>

<div class="min-h-screen bg-background">
  <Header />

  <main class="px-6 pb-32 pt-24 md:px-12 md:pt-28 lg:px-20">

    <!-- Page header -->
    <header class="relative mb-0">
      <div
        aria-hidden="true"
        class="pointer-events-none absolute -inset-x-6 -top-10 -z-10 h-[180px] opacity-50 md:-inset-x-12 lg:-inset-x-20"
        style="background: radial-gradient(50% 100% at 15% 0%, oklch(0.62 0.22 285 / 0.20), transparent 70%);"
      ></div>
      <div class="mb-2 text-[10px] font-semibold uppercase tracking-[0.35em] text-primary-glow">Your list</div>
      <h1 class="font-serif-display text-[clamp(2rem,4vw,3.25rem)] leading-[1] tracking-tight">Watchlist</h1>
      {#if mounted && items.length > 0}
        <p class="mt-1.5 text-sm text-muted-foreground">
          {items.length} {items.length === 1 ? 'title' : 'titles'} saved
        </p>
      {:else}
        <p class="mt-1.5 max-w-xl text-sm text-muted-foreground">
          Films and shows you've saved to watch later, all in one place
        </p>
      {/if}
    </header>

    {#if !mounted}
      <!-- Avoid flash before localStorage loads -->
      <div class="flex min-h-[30vh] items-center justify-center">
        <div class="h-8 w-8 animate-spin rounded-full border-2 border-border border-t-primary-glow"></div>
      </div>

    {:else if items.length === 0}
      <div class="flex flex-col items-center justify-center py-24 text-center">
        <div class="hairline flex h-14 w-14 items-center justify-center rounded-2xl bg-surface-elevated/70 text-primary-glow shadow-elev">
          <Bookmark class="h-6 w-6" />
        </div>
        <p class="font-serif-display mt-5 text-2xl tracking-tight">Nothing saved yet</p>
        <p class="mt-2 max-w-sm text-sm text-muted-foreground">
          Tap the <span class="font-medium text-foreground/70">+</span> button on any movie or show to add it here.
        </p>
        <div class="mt-6 flex flex-wrap items-center justify-center gap-3">
          <a
            href="/movies"
            class="hairline inline-flex items-center gap-2 rounded-full bg-foreground/[0.06] px-6 py-3 text-sm font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.10] hover:text-foreground"
          >
            Browse movies →
          </a>
          <a
            href="/tv"
            class="hairline inline-flex items-center gap-2 rounded-full bg-foreground/[0.06] px-6 py-3 text-sm font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.10] hover:text-foreground"
          >
            Browse TV →
          </a>
        </div>
      </div>

    {:else}
      <!-- ── Control bar ──────────────────────────────────────────────────────── -->
      <div class="sticky top-14 z-30 -mx-6 mt-8 flex items-center gap-2 border-b border-border bg-background/80 px-6 py-3 backdrop-blur-md md:-mx-12 md:px-12 lg:-mx-20 lg:px-20">

        <!-- Search -->
        <div class="relative min-w-0 flex-1 max-w-xs">
          <Search class="pointer-events-none absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <input
            bind:value={q}
            placeholder="Search watchlist…"
            class="h-9 w-full rounded-full border border-border bg-surface/60 pl-10 pr-4 text-sm outline-none placeholder:text-muted-foreground/70 focus:border-primary/60 focus:bg-surface"
          />
          {#if q}
            <button type="button" onclick={() => (q = '')} class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground" aria-label="Clear search">
              <X class="h-3.5 w-3.5" />
            </button>
          {/if}
        </div>

        <!-- Kind filter pills -->
        <div class="hairline hidden items-center gap-1 rounded-full bg-foreground/[0.04] p-1 sm:flex">
          {#each ([['all', 'All'], ['movie', 'Movies'], ['series', 'TV']] as const) as [val, label]}
            <button
              type="button"
              onclick={() => (kindFilter = val)}
              class={`rounded-full px-3 py-1.5 text-xs font-medium transition-colors ${
                kindFilter === val
                  ? 'bg-foreground text-background'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              {label}
            </button>
          {/each}
        </div>

        <div class="ml-auto flex items-center gap-2">

          <!-- Sort picker -->
          <div class="relative">
            <button
              type="button"
              onclick={() => { sortOpen = !sortOpen; }}
              class={`hairline flex items-center gap-1.5 rounded-full bg-foreground/[0.04] px-3 py-1.5 text-xs transition-colors hover:bg-foreground/[0.08] ${sortOpen ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'}`}
            >
              {sortLabel}
              <ChevronDown class={`h-3.5 w-3.5 transition-transform ${sortOpen ? 'rotate-180' : ''}`} />
            </button>
            {#if sortOpen}
              <button type="button" class="fixed inset-0 z-40 cursor-default" aria-hidden="true" tabindex="-1" onclick={() => (sortOpen = false)}></button>
              <div class="absolute right-0 top-full z-50 mt-1.5 min-w-[188px] overflow-hidden rounded-xl border border-border bg-surface/95 py-1 shadow-xl backdrop-blur-xl">
                {#each sortOptions as opt (opt.value)}
                  <button
                    type="button"
                    onclick={() => { sort = opt.value; sortOpen = false; }}
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

          <!-- Density switcher -->
          <div class="hairline hidden items-center rounded-full bg-foreground/[0.04] p-1 md:flex">
            {#each (['S', 'M', 'L'] as const) as option (option)}
              <button
                type="button"
                aria-label={option === 'S' ? 'Small cards' : option === 'M' ? 'Medium cards' : 'Large cards'}
                onclick={() => setDensity(option)}
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

      <!-- ── Grid ──────────────────────────────────────────────────────────── -->
      <div class="mt-10">

        {#if filtered.length === 0}
          <div class="flex flex-col items-center justify-center py-24 text-center">
            <Search class="h-10 w-10 text-muted-foreground/20" />
            <p class="mt-4 text-muted-foreground">
              {#if q}
                No titles match "<span class="text-foreground">{q}</span>"
              {:else}
                No {kindFilter === 'movie' ? 'movies' : 'TV shows'} in your watchlist yet
              {/if}
            </p>
            <button
              type="button"
              onclick={() => { q = ''; kindFilter = 'all'; }}
              class="mt-3 text-sm text-primary-glow hover:underline"
            >
              Clear filters
            </button>
          </div>

        {:else}
          {#if q || kindFilter !== 'all' || filtered.length !== items.length}
            <p class="mb-6 text-xs text-muted-foreground">
              {filtered.length} of {items.length} titles
            </p>
          {/if}

          <div class={`grid gap-x-5 gap-y-10 ${densityGrid[density]}`}>
            {#each filtered as item (item.id + item.kind)}
              {@const href = item.kind === 'series' ? `/tv/${item.id}` : `/movies/${item.id}`}
              <div class="group relative">
                <a {href} class="block">
                  <!-- Poster -->
                  <div
                    class="shadow-poster relative aspect-[2/3] w-full overflow-hidden rounded-xl bg-surface transition-all duration-300 group-hover:-translate-y-2 group-hover:scale-[1.04] group-hover:ring-[3px] group-hover:ring-white/85 group-hover:shadow-[0_28px_60px_-12px_oklch(0_0_0/0.85)]"
                    style={item.posterUrl ? '' : 'background: linear-gradient(135deg, #1e3a5f, #0f172a)'}
                  >
                    {#if item.posterUrl}
                      <img
                        src={item.posterUrl}
                        alt={item.title}
                        loading="lazy"
                        class="absolute inset-0 h-full w-full object-cover"
                        onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
                      />
                    {/if}
                    <div class="absolute inset-x-0 bottom-0 h-1/2 bg-gradient-to-t from-black/80 to-transparent"></div>

                    <!-- Kind badge -->
                    {#if density !== 'S'}
                      <div class="absolute left-2 top-2 rounded-md bg-black/55 p-1 backdrop-blur-sm ring-1 ring-white/15">
                        {#if item.kind === 'series'}
                          <Tv class="h-3 w-3 text-white/80" />
                        {:else}
                          <Film class="h-3 w-3 text-white/80" />
                        {/if}
                      </div>
                    {/if}
                  </div>

                  <!-- Text -->
                  {#if density !== 'S'}
                    <div class="mt-2.5 px-0.5">
                      <h3 class="truncate text-sm font-medium text-foreground">{item.title}</h3>
                      {#if item.year}
                        <p class="mt-0.5 text-[11px] text-muted-foreground">{item.year}</p>
                      {/if}
                    </div>
                  {/if}
                </a>

                <!-- Remove button — top-right corner, appears on hover -->
                <button
                  type="button"
                  onclick={() => removeFromWatchlist(item.id, item.kind)}
                  aria-label="Remove from watchlist"
                  title="Remove from watchlist"
                  class="absolute right-1 top-1 flex h-7 w-7 items-center justify-center rounded-full bg-black/60 text-white/70 opacity-0 backdrop-blur-sm transition-all hover:text-white group-hover:opacity-100"
                >
                  <BookmarkX class="h-3.5 w-3.5" />
                </button>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    {/if}
  </main>
</div>
