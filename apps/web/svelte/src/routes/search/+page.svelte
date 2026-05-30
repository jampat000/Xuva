<script lang="ts">
  import { page } from '$app/state';
  import { appState } from '$lib/stores/appState.svelte';
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';
  import { Search, Film, Tv, SlidersHorizontal, User, Layers } from 'lucide-svelte';
  import Header from '$lib/components/Header.svelte';
  import { runSearch, getSearchResults, isSearchLoading, getCatalogueSize } from '$lib/stores/searchStore.svelte';
  import type {
    SearchHit,
    SearchMovieHit,
    SearchSeriesHit,
    SearchPersonHit,
    SearchCollectionHit,
  } from '$lib/api/browse';

  // Read `q` from URL params
  let urlQuery = $derived(page.url.searchParams.get('q') ?? '');
  let localQuery = $state('');
  let filterType = $state<'all' | 'Movie' | 'Series' | 'Person' | 'Collection'>('all');
  let inputEl = $state<HTMLInputElement | null>(null);

  // Sync localQuery with URL on mount and when URL changes
  $effect(() => {
    localQuery = urlQuery;
  });

  // Trigger backend search whenever the query changes. /search wants more
  // results per type than the header dropdown, so request up to 40.
  $effect(() => {
    runSearch(localQuery, 40);
  });

  const resp = $derived(getSearchResults());
  const queryMatches = $derived(!!resp && resp.query === localQuery.trim());
  const movies = $derived<SearchMovieHit[]>(queryMatches ? resp!.movies : []);
  const series = $derived<SearchSeriesHit[]>(queryMatches ? resp!.series : []);
  const people = $derived<SearchPersonHit[]>(queryMatches ? resp!.people : []);
  const collections = $derived<SearchCollectionHit[]>(queryMatches ? resp!.collections : []);
  const totalCount = $derived(movies.length + series.length + people.length + collections.length);

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      const q = localQuery.trim();
      if (q) {
        goto(`/search?q=${encodeURIComponent(q)}`, { replaceState: true });
      }
    }
    if (e.key === 'Escape') {
      inputEl?.blur();
    }
  }

  function hitHref(hit: SearchHit): string {
    switch (hit.kind) {
      case 'movie':
        return `/movies/${hit.id}`;
      case 'series':
        return `/tv/${hit.id}`;
      case 'person':
        return `/people/${encodeURIComponent(hit.name)}`;
      case 'collection':
        return `/collections/${hit.id}`;
    }
  }

  onMount(() => {
    if (inputEl) inputEl.focus();
  });
</script>

<svelte:head>
  <title>{localQuery ? `"${localQuery}" — Search` : 'Search'} — {appState.serverName}</title>
  <meta name="description" content="Search your Xuva media library." />
</svelte:head>

<div class="min-h-screen bg-background">
  <Header />

  <div class="px-6 pb-32 pt-28 md:px-12 lg:px-20">
    <!-- Search input -->
    <div class="mx-auto max-w-2xl">
      <div class="relative">
        <Search class="pointer-events-none absolute left-5 top-1/2 h-5 w-5 -translate-y-1/2 text-muted-foreground" />
        <input
          bind:this={inputEl}
          bind:value={localQuery}
          type="search"
          autocomplete="off"
          placeholder="Search movies, shows, people, collections..."
          class="h-14 w-full rounded-2xl border border-border bg-surface/60 pl-14 pr-5 text-base text-foreground outline-none ring-0 transition-all placeholder:text-muted-foreground/60 focus:border-primary/60 focus:bg-surface focus:shadow-glow"
          onkeydown={handleKeydown}
        />
      </div>
    </div>

    {#if localQuery.trim()}
      <!-- Filter bar -->
      <div class="mt-8 flex flex-wrap items-center gap-3">
        <SlidersHorizontal class="h-4 w-4 text-muted-foreground" />
        {#each (['all', 'Movie', 'Series', 'Person', 'Collection'] as const) as f (f)}
          <button
            type="button"
            onclick={() => (filterType = f)}
            class={`rounded-full px-4 py-1.5 text-sm font-medium transition-colors ${
              filterType === f
                ? 'bg-foreground text-background'
                : 'hairline bg-foreground/[0.04] text-muted-foreground hover:bg-foreground/[0.08] hover:text-foreground'
            }`}
          >
            {#if f === 'all'}
              All ({totalCount})
            {:else if f === 'Movie'}
              Movies ({movies.length})
            {:else if f === 'Series'}
              TV ({series.length})
            {:else if f === 'Person'}
              People ({people.length})
            {:else}
              Collections ({collections.length})
            {/if}
          </button>
        {/each}
      </div>

      {#if isSearchLoading() && totalCount === 0}
        <div class="flex min-h-[30vh] items-center justify-center">
          <div class="h-8 w-8 animate-spin rounded-full border-2 border-border border-t-primary-glow"></div>
        </div>
      {:else if totalCount === 0}
        <div class="flex min-h-[30vh] flex-col items-center justify-center gap-3 text-center">
          <div class="flex h-16 w-16 items-center justify-center rounded-full bg-surface/60">
            <Search class="h-7 w-7 text-muted-foreground" />
          </div>
          <p class="text-base font-medium">No results for "{localQuery}"</p>
          <p class="text-sm text-muted-foreground">Try a different title, person, or collection</p>
        </div>
      {:else}
        <p class="mt-6 text-sm text-muted-foreground">
          {totalCount} result{totalCount !== 1 ? 's' : ''}
          for <span class="text-foreground">"{localQuery}"</span>
        </p>

        <!-- Movies -->
        {#if movies.length > 0 && (filterType === 'all' || filterType === 'Movie')}
          <section class="mt-8">
            {#if filterType === 'all'}
              <h2 class="mb-3 flex items-center gap-2 text-sm font-semibold uppercase tracking-wide text-muted-foreground">
                <Film class="h-4 w-4" /> Movies
              </h2>
            {/if}
            <div class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
              {#each movies as m (m.id)}
                {@const art = `/api/artwork/movie/${encodeURIComponent(m.id)}?type=poster&w=300`}
                <a href={hitHref(m)} class="group block">
                  <!-- Route through the artwork proxy by id rather than gating on
                       m.posterUrl: the search API omits posterUrl, so the old
                       `{#if m.posterUrl}` left every result as a bare icon. The
                       proxy serves the full-res poster by id; onerror reveals the
                       Film icon fallback if a title genuinely has no artwork. -->
                  <div class="relative aspect-[2/3] overflow-hidden rounded-xl bg-surface/60">
                    <div class="absolute inset-0 flex items-center justify-center"><Film class="h-8 w-8 text-muted-foreground" /></div>
                    <img src={art} alt={m.title} class="absolute inset-0 h-full w-full object-cover transition-transform group-hover:scale-105" loading="lazy" onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')} />
                  </div>
                  <p class="mt-2 truncate text-sm font-medium">{m.title}</p>
                  {#if m.year}
                    <p class="text-xs text-muted-foreground">{m.year}</p>
                  {/if}
                </a>
              {/each}
            </div>
          </section>
        {/if}

        <!-- Series -->
        {#if series.length > 0 && (filterType === 'all' || filterType === 'Series')}
          <section class="mt-8">
            {#if filterType === 'all'}
              <h2 class="mb-3 flex items-center gap-2 text-sm font-semibold uppercase tracking-wide text-muted-foreground">
                <Tv class="h-4 w-4" /> TV
              </h2>
            {/if}
            <div class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
              {#each series as s (s.id)}
                {@const art = `/api/artwork/series/${encodeURIComponent(s.id)}?type=poster&w=300`}
                <a href={hitHref(s)} class="group block">
                  <div class="relative aspect-[2/3] overflow-hidden rounded-xl bg-surface/60">
                    <div class="absolute inset-0 flex items-center justify-center"><Tv class="h-8 w-8 text-muted-foreground" /></div>
                    <img src={art} alt={s.title} class="absolute inset-0 h-full w-full object-cover transition-transform group-hover:scale-105" loading="lazy" onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')} />
                  </div>
                  <p class="mt-2 truncate text-sm font-medium">{s.title}</p>
                  {#if s.year}
                    <p class="text-xs text-muted-foreground">{s.year}</p>
                  {/if}
                </a>
              {/each}
            </div>
          </section>
        {/if}

        <!-- People -->
        {#if people.length > 0 && (filterType === 'all' || filterType === 'Person')}
          <section class="mt-8">
            {#if filterType === 'all'}
              <h2 class="mb-3 flex items-center gap-2 text-sm font-semibold uppercase tracking-wide text-muted-foreground">
                <User class="h-4 w-4" /> People
              </h2>
            {/if}
            <div class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6">
              {#each people as p (p.name)}
                <a href={hitHref(p)} class="group flex flex-col items-center text-center">
                  <div class="aspect-square w-full max-w-[140px] overflow-hidden rounded-full bg-surface/60">
                    {#if p.profileUrl}
                      <img src={p.profileUrl} alt={p.name} class="h-full w-full object-cover transition-transform group-hover:scale-105" loading="lazy" />
                    {:else}
                      <div class="flex h-full w-full items-center justify-center"><User class="h-10 w-10 text-muted-foreground" /></div>
                    {/if}
                  </div>
                  <p class="mt-2 truncate text-sm font-medium w-full">{p.name}</p>
                  <p class="text-xs text-muted-foreground">
                    {p.creditCount} credit{p.creditCount === 1 ? '' : 's'}
                  </p>
                </a>
              {/each}
            </div>
          </section>
        {/if}

        <!-- Collections -->
        {#if collections.length > 0 && (filterType === 'all' || filterType === 'Collection')}
          <section class="mt-8">
            {#if filterType === 'all'}
              <h2 class="mb-3 flex items-center gap-2 text-sm font-semibold uppercase tracking-wide text-muted-foreground">
                <Layers class="h-4 w-4" /> Collections
              </h2>
            {/if}
            <div class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
              {#each collections as c (c.id)}
                <a href={hitHref(c)} class="group block">
                  <div class="aspect-[2/3] overflow-hidden rounded-xl bg-surface/60">
                    {#if c.posterUrl}
                      <img src={c.posterUrl} alt={c.name} class="h-full w-full object-cover transition-transform group-hover:scale-105" loading="lazy" />
                    {:else}
                      <div class="flex h-full w-full items-center justify-center"><Layers class="h-8 w-8 text-muted-foreground" /></div>
                    {/if}
                  </div>
                  <p class="mt-2 truncate text-sm font-medium">{c.name}</p>
                  <p class="text-xs text-muted-foreground">
                    {c.movieCount} movie{c.movieCount === 1 ? '' : 's'}
                  </p>
                </a>
              {/each}
            </div>
          </section>
        {/if}
      {/if}
    {:else}
      <!-- Empty state -->
      <div class="mt-20 flex flex-col items-center gap-6 text-center">
        <div class="flex items-center gap-4">
          <div class="flex h-16 w-16 items-center justify-center rounded-2xl bg-surface/60">
            <Film class="h-7 w-7 text-muted-foreground" />
          </div>
          <div class="flex h-16 w-16 items-center justify-center rounded-2xl bg-surface/60">
            <Tv class="h-7 w-7 text-muted-foreground" />
          </div>
          <div class="flex h-16 w-16 items-center justify-center rounded-2xl bg-surface/60">
            <User class="h-7 w-7 text-muted-foreground" />
          </div>
          <div class="flex h-16 w-16 items-center justify-center rounded-2xl bg-surface/60">
            <Layers class="h-7 w-7 text-muted-foreground" />
          </div>
        </div>
        {#if getCatalogueSize() > 0}
          <p class="text-sm text-muted-foreground">Type to search movies, shows, people, and collections</p>
        {:else}
          <p class="text-sm text-muted-foreground">Type to search your library</p>
        {/if}
        <p class="text-xs text-muted-foreground/60">Press <kbd class="rounded border border-border bg-surface px-1.5 py-0.5">⌘K</kbd> from anywhere to search</p>
      </div>
    {/if}
  </div>
</div>
