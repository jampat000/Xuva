<script lang="ts">
  import { onMount } from 'svelte';
  import { ChevronDown, Layers, RotateCcw, Search, X } from 'lucide-svelte';
  import Header from '$lib/components/Header.svelte';
  import ErrorState from '$lib/components/ErrorState.svelte';
  import { appState } from '$lib/stores/appState.svelte';
  import { getCollections } from '$lib/api/home';
  import { getAuthSession } from '$lib/api/auth';
  import { updateUserPreferences } from '$lib/api/operator';
  import type { CollectionListItem } from '$lib/api/home';

  // ── Types ──────────────────────────────────────────────────────────────────
  type Density = 'S' | 'M' | 'L';
  type Sort = 'az' | 'za' | 'count-desc' | 'count-asc' | 'random';

  const densityGrid: Record<Density, string> = {
    S: 'grid-cols-3 sm:grid-cols-5 md:grid-cols-7 lg:grid-cols-8 xl:grid-cols-10',
    M: 'grid-cols-2 sm:grid-cols-3 md:grid-cols-5 lg:grid-cols-6 xl:grid-cols-7',
    L: 'grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-4 xl:grid-cols-5',
  };

  const sortOptions: { value: Sort; label: string }[] = [
    { value: 'az',         label: 'Title A → Z' },
    { value: 'za',         label: 'Title Z → A' },
    { value: 'count-desc', label: 'Most Films' },
    { value: 'count-asc',  label: 'Fewest Films' },
    { value: 'random',     label: 'Random' },
  ];

  // ── Data state ─────────────────────────────────────────────────────────────
  let items   = $state<CollectionListItem[]>([]);
  let loading = $state(true);
  let error   = $state<string | null>(null);

  // ── Control state ──────────────────────────────────────────────────────────
  let q          = $state('');
  let sort       = $state<Sort>('az');
  const DENSITY_KEY = 'xuva:poster-size';
  function readCachedDensity(): Density {
    try {
      const v = typeof localStorage !== 'undefined' ? localStorage.getItem(DENSITY_KEY) : null;
      if (v === 'S' || v === 'M' || v === 'L') return v;
    } catch { /* SSR or privacy mode */ }
    return 'M';
  }
  let density    = $state<Density>(readCachedDensity());
  let sortOpen   = $state(false);
  let randomSeed = $state(Date.now());

  // ── Helpers ────────────────────────────────────────────────────────────────
  function pseudoRandom(id: string, seed: number): number {
    let h = seed & 0x7fffffff;
    for (let i = 0; i < id.length; i++) h = (Math.imul(h, 31) + id.charCodeAt(i)) & 0x7fffffff;
    return h / 0x7fffffff;
  }

  // ── Derived filtered + sorted list ─────────────────────────────────────────
  const filtered = $derived.by(() => {
    let result = items.slice();
    if (q.trim()) {
      const needle = q.toLowerCase();
      result = result.filter(c => c.name.toLowerCase().includes(needle));
    }
    result.sort((a, b) => {
      switch (sort) {
        case 'az':         return a.name.localeCompare(b.name);
        case 'za':         return b.name.localeCompare(a.name);
        case 'count-desc': return b.movieCount - a.movieCount;
        case 'count-asc':  return a.movieCount - b.movieCount;
        case 'random':     return pseudoRandom(a.id, randomSeed) - pseudoRandom(b.id, randomSeed);
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

  // ── Data load ──────────────────────────────────────────────────────────────
  async function load() {
    error = null;
    loading = true;
    try {
      const resp = await getCollections();
      items = resp.collections ?? [];
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load collections';
    } finally {
      loading = false;
    }
  }

  onMount(async () => {
    load();
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
  <title>Collections — {appState.serverName}</title>
</svelte:head>

<div class="min-h-screen bg-background">
  <Header />

  {#if error}
    <ErrorState
      title="Can't load collections"
      message="Make sure your Xuva server is running, then try again."
      actions={[{ label: 'Try again', onClick: load }]}
      diagnosticInfo={error}
    />
  {:else}
    <!-- Page header -->
    <div class="px-6 pb-0 pt-28 md:px-12 lg:px-20">
      <div class="mb-2 text-[10px] font-semibold uppercase tracking-[0.35em] text-primary-glow">
        Your library
      </div>
      <h1 class="font-serif-display text-[clamp(2rem,5vw,3.5rem)] leading-[0.95] tracking-tight">
        Collections.
      </h1>
      <p class="mt-4 max-w-xl text-sm text-muted-foreground">
        Film franchises and multi-part series grouped together from your library.
      </p>
    </div>

    <!-- ── Control bar ──────────────────────────────────────────────────────── -->
    <div class="sticky top-14 z-30 mt-8 flex items-center gap-2 border-b border-border bg-background/80 px-6 py-3 backdrop-blur-md md:px-12 lg:px-20">

      <!-- Search -->
      <div class="relative min-w-0 flex-1 max-w-xs">
        <Search class="pointer-events-none absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <input
          bind:value={q}
          placeholder="Search collections…"
          class="h-9 w-full rounded-full border border-border bg-surface/60 pl-10 pr-4 text-sm outline-none placeholder:text-muted-foreground/70 focus:border-primary/60 focus:bg-surface"
        />
        {#if q}
          <button type="button" onclick={() => (q = '')} class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground" aria-label="Clear search">
            <X class="h-3.5 w-3.5" />
          </button>
        {/if}
      </div>

      <div class="ml-auto flex items-center gap-2">

        <!-- Sort picker -->
        <div class="flex items-center gap-1">
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
              onclick={() => { sortOpen = !sortOpen; }}
              class={`hairline flex items-center gap-1.5 rounded-full bg-foreground/[0.04] px-3 py-1.5 text-xs transition-colors hover:bg-foreground/[0.08] ${sortOpen ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'}`}
            >
              {sortLabel}
              <ChevronDown class={`h-3.5 w-3.5 transition-transform ${sortOpen ? 'rotate-180' : ''}`} />
            </button>
            {#if sortOpen}
              <button type="button" class="fixed inset-0 z-40 cursor-default" aria-hidden="true" tabindex="-1" onclick={() => (sortOpen = false)}></button>
              <div class="absolute right-0 top-full z-50 mt-1.5 min-w-[160px] overflow-hidden rounded-xl border border-border bg-surface/95 py-1 shadow-xl backdrop-blur-xl">
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

    <!-- ── Grid ────────────────────────────────────────────────────────────── -->
    <main class="px-6 pb-32 pt-10 md:px-12 lg:px-20">

      {#if loading}
        <div class={`grid gap-x-5 gap-y-10 ${densityGrid[density]}`}>
          {#each { length: 12 } as _}
            <div class="flex flex-col gap-2">
              <div class="aspect-[2/3] animate-pulse rounded-xl bg-surface"></div>
              {#if density !== 'S'}
                <div class="h-3 w-3/4 animate-pulse rounded bg-surface"></div>
              {/if}
            </div>
          {/each}
        </div>

      {:else if items.length === 0}
        <div class="mt-24 flex flex-col items-center justify-center gap-4 text-center">
          <Layers class="h-14 w-14 text-muted-foreground/20" />
          <p class="text-muted-foreground">No collections found in your library.</p>
          <p class="text-sm text-muted-foreground/60">Collections are created automatically when your movies belong to a TMDB franchise.</p>
        </div>

      {:else if filtered.length === 0}
        <div class="mt-24 flex flex-col items-center justify-center gap-4 text-center">
          <Search class="h-10 w-10 text-muted-foreground/20" />
          <p class="text-muted-foreground">No collections match "<span class="text-foreground">{q}</span>"</p>
          <button type="button" onclick={() => (q = '')} class="text-sm text-primary-glow hover:underline">Clear search</button>
        </div>

      {:else}
        {#if q || filtered.length !== items.length}
          <p class="mb-6 text-xs text-muted-foreground">
            {filtered.length} of {items.length} collections
          </p>
        {/if}

        <div class={`grid gap-x-5 gap-y-10 ${densityGrid[density]}`}>
          {#each filtered as collection (collection.id)}
            <a href="/collections/{collection.id}" class="group flex flex-col gap-2">
              <div class="relative overflow-hidden rounded-xl bg-surface shadow-poster transition-transform duration-200 group-hover:scale-[1.03]">
                {#if collection.posterUrl}
                  <img
                    src={collection.posterUrl}
                    alt={collection.name}
                    class="aspect-[2/3] w-full object-cover"
                    onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
                    loading="lazy"
                  />
                {:else}
                  <div class="flex aspect-[2/3] w-full items-center justify-center bg-surface-elevated">
                    <Layers class="h-8 w-8 text-muted-foreground/30" />
                  </div>
                {/if}
                <div class="absolute inset-x-0 bottom-0 h-1/3 bg-gradient-to-t from-black/60 to-transparent opacity-0 transition-opacity group-hover:opacity-100"></div>
              </div>
              {#if density !== 'S'}
                <div class="min-w-0 px-0.5">
                  <p class="truncate text-sm font-medium leading-tight">{collection.name}</p>
                  <p class="text-xs text-muted-foreground">{collection.movieCount} {collection.movieCount === 1 ? 'film' : 'films'}</p>
                </div>
              {/if}
            </a>
          {/each}
        </div>
      {/if}
    </main>
  {/if}
</div>
