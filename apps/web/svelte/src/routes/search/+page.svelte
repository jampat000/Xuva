<script lang="ts">
  import { page } from '$app/state';
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';
  import { Search, Film, Tv, SlidersHorizontal } from 'lucide-svelte';
  import Header from '$lib/components/Header.svelte';
  import PosterCard from '$lib/components/PosterCard.svelte';
  import { primeSearchCatalogue, searchCatalogue, isSearchLoading, getCatalogueSize } from '$lib/stores/searchStore.svelte';
  import type { Media } from '$lib/mock-data';

  // Read `q` from URL params
  let urlQuery = $derived(page.url.searchParams.get('q') ?? '');
  let localQuery = $state('');
  let filterType = $state<'all' | 'Movie' | 'Series'>('all');
  let inputEl = $state<HTMLInputElement | null>(null);

  // Sync localQuery with URL on mount and when URL changes
  $effect(() => {
    localQuery = urlQuery;
  });

  const rawResults = $derived(searchCatalogue(localQuery, 200));
  const results = $derived(
    filterType === 'all'
      ? rawResults
      : rawResults.filter((m) => m.type === filterType)
  );
  const movies = $derived(rawResults.filter((m) => m.type === 'Movie').length);
  const shows = $derived(rawResults.filter((m) => m.type === 'Series').length);

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

  onMount(() => {
    primeSearchCatalogue();
    // Auto-focus the search input on page load
    if (inputEl) inputEl.focus();
  });
</script>

<svelte:head>
  <title>{localQuery ? `"${localQuery}" — Search` : 'Search'} — Xuva</title>
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
          placeholder="Search movies, shows, people..."
          class="h-14 w-full rounded-2xl border border-border bg-surface/60 pl-14 pr-5 text-base text-foreground outline-none ring-0 transition-all placeholder:text-muted-foreground/60 focus:border-primary/60 focus:bg-surface focus:shadow-glow"
          onkeydown={handleKeydown}
        />
      </div>
    </div>

    {#if localQuery.trim()}
      <!-- Filter bar -->
      <div class="mt-8 flex items-center gap-3">
        <SlidersHorizontal class="h-4 w-4 text-muted-foreground" />
        {#each (['all', 'Movie', 'Series'] as const) as f (f)}
          <button
            type="button"
            onclick={() => (filterType = f)}
            class={`rounded-full px-4 py-1.5 text-sm font-medium transition-colors ${
              filterType === f
                ? 'bg-foreground text-background'
                : 'hairline bg-foreground/[0.04] text-muted-foreground hover:bg-foreground/[0.08] hover:text-foreground'
            }`}
          >
            {f === 'all' ? `All (${rawResults.length})` : f === 'Movie' ? `Movies (${movies})` : `TV (${shows})`}
          </button>
        {/each}
      </div>

      <!-- Results -->
      {#if isSearchLoading()}
        <div class="flex min-h-[30vh] items-center justify-center">
          <div class="h-8 w-8 animate-spin rounded-full border-2 border-border border-t-primary-glow"></div>
        </div>
      {:else if results.length === 0}
        <div class="flex min-h-[30vh] flex-col items-center justify-center gap-3 text-center">
          <div class="flex h-16 w-16 items-center justify-center rounded-full bg-surface/60">
            <Search class="h-7 w-7 text-muted-foreground" />
          </div>
          <p class="text-base font-medium">No results for "{localQuery}"</p>
          <p class="text-sm text-muted-foreground">Try a different title or keyword</p>
        </div>
      {:else}
        <p class="mt-6 text-sm text-muted-foreground">
          {results.length} result{results.length !== 1 ? 's' : ''}
          {filterType !== 'all' ? ` in ${filterType === 'Movie' ? 'Movies' : 'TV'}` : ''}
          for <span class="text-foreground">"{localQuery}"</span>
        </p>
        <div class="mt-6 grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
          {#each results as m (m.id)}
            <PosterCard media={m} />
          {/each}
        </div>
      {/if}
    {:else}
      <!-- Empty state — catalogue stats -->
      <div class="mt-20 flex flex-col items-center gap-6 text-center">
        <div class="flex items-center gap-4">
          <div class="flex h-16 w-16 items-center justify-center rounded-2xl bg-surface/60">
            <Film class="h-7 w-7 text-muted-foreground" />
          </div>
          <div class="flex h-16 w-16 items-center justify-center rounded-2xl bg-surface/60">
            <Tv class="h-7 w-7 text-muted-foreground" />
          </div>
        </div>
        {#if getCatalogueSize() > 0}
          <p class="text-sm text-muted-foreground">
            Search across <span class="text-foreground font-medium">{getCatalogueSize()}</span> titles in your library
          </p>
        {:else if isSearchLoading()}
          <p class="text-sm text-muted-foreground">Loading your library…</p>
        {:else}
          <p class="text-sm text-muted-foreground">Type to search your library</p>
        {/if}
        <p class="text-xs text-muted-foreground/60">Press <kbd class="rounded border border-border bg-surface px-1.5 py-0.5">⌘K</kbd> from anywhere to search</p>
      </div>
    {/if}
  </div>
</div>
