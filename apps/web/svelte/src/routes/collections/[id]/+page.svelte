<script lang="ts">
  import { page } from '$app/state';
  import { onMount } from 'svelte';
  import { ChevronLeft, Play, Star, Clock, Film } from 'lucide-svelte';
  import Header from '$lib/components/Header.svelte';
  import ErrorState from '$lib/components/ErrorState.svelte';
  import { appState } from '$lib/stores/appState.svelte';
  import { getCollectionDetail } from '$lib/api/home';
  import type { CollectionDetailResponse, CollectionMovie } from '$lib/api/home';

  const id = $derived(page.params.id ?? '');

  let data = $state<CollectionDetailResponse | null>(null);
  let loading = $state(true);
  let error = $state<string | null>(null);

  const collection = $derived(data?.collection);
  const movies = $derived(data?.movies ?? []);
  const sortedMovies = $derived([...movies].sort((a, b) => (a.year ?? 0) - (b.year ?? 0)));

  // Deterministic palette fallback (matches adapters.ts)
  const PALETTES: Array<{ bg: string; accent: string }> = [
    { bg: 'linear-gradient(135deg,#0f172a,#1e3a8a)', accent: '#60a5fa' },
    { bg: 'linear-gradient(135deg,#1c1917,#7c2d12)', accent: '#fb923c' },
    { bg: 'linear-gradient(135deg,#0f172a,#4c1d95)', accent: '#a78bfa' },
    { bg: 'linear-gradient(135deg,#0a0a1a,#1e3a5f)', accent: '#93c5fd' },
    { bg: 'linear-gradient(135deg,#1a0a2e,#6b21a8)', accent: '#d8b4fe' },
  ];
  function fallbackGradient(id: string): string {
    let h = 0;
    for (let i = 0; i < id.length; i++) h = (h + id.charCodeAt(i)) % PALETTES.length;
    return PALETTES[h].bg;
  }

  function formatRuntime(minutes?: number): string {
    if (!minutes) return '';
    const h = Math.floor(minutes / 60);
    const m = minutes % 60;
    return h > 0 ? `${h}h ${m}m` : `${m}m`;
  }

  async function load() {
    try {
      loading = true;
      error = null;
      data = await getCollectionDetail(id);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load collection';
    } finally {
      loading = false;
    }
  }

  onMount(load);
</script>

<svelte:head>
  <title>{collection?.name ?? 'Collection'} — {appState.serverName}</title>
</svelte:head>

<div class="min-h-screen bg-background">
  <Header />

  {#if loading}
    <div class="flex min-h-[60vh] items-center justify-center">
      <div class="h-10 w-10 animate-spin rounded-full border-2 border-border border-t-primary-glow"></div>
    </div>

  {:else if error}
    <ErrorState
      title="Collection not found"
      message={error}
      actions={[{ label: 'Browse movies', href: '/movies' }]}
      diagnosticInfo={`Collection ID: ${id}\nError: ${error}`}
    />

  {:else}
    <!-- Cinematic header with collection backdrop -->
    <div class="relative -mt-16 h-[50vh] min-h-[360px] w-full overflow-hidden">
      {#if collection?.backdropUrl}
        <img
          src={collection.backdropUrl}
          alt=""
          class="h-full w-full object-cover"
          onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
        />
      {:else if movies[0]?.backdropUrl}
        <img
          src={movies[0].backdropUrl}
          alt=""
          class="h-full w-full object-cover"
          onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
        />
      {:else}
        <div class="h-full w-full" style="background: linear-gradient(135deg, #0f172a, #4c1d95)"></div>
      {/if}
      <div class="absolute inset-0 bg-gradient-to-r from-background via-background/60 to-transparent"></div>
      <div class="absolute inset-x-0 bottom-0 h-3/4 bg-gradient-to-t from-background to-transparent"></div>
      <div class="absolute inset-x-0 top-0 h-24 bg-gradient-to-b from-background/80 to-transparent"></div>
    </div>

    <div class="relative -mt-40 px-6 pb-32 md:px-12 lg:px-20">
      <a href="/movies" class="mb-8 inline-flex items-center gap-2 text-sm text-muted-foreground transition-colors hover:text-foreground">
        <ChevronLeft class="h-4 w-4" /> Back to Movies
      </a>

      <!-- Collection header -->
      <div class="flex items-end gap-8">
        {#if collection?.posterUrl}
          <img
            src={collection.posterUrl}
            alt={collection.name}
            class="shadow-poster hidden h-48 w-32 shrink-0 rounded-xl object-cover md:block"
            onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
          />
        {/if}
        <div class="min-w-0">
          <div class="text-[10px] font-semibold uppercase tracking-[0.35em] text-primary-glow">Collection</div>
          <h1 class="font-serif-display mt-2 text-[clamp(2rem,5vw,4rem)] leading-[0.95] tracking-tight">
            {collection?.name ?? 'Collection'}
          </h1>
          {#if movies.length > 0}
            <p class="mt-4 text-sm text-muted-foreground">
              {movies.length} {movies.length === 1 ? 'movie' : 'movies'} in your library
            </p>
          {/if}
        </div>
      </div>

      <!-- Movie grid -->
      {#if sortedMovies.length === 0}
        <div class="mt-16 flex flex-col items-center justify-center gap-3 text-center">
          <Film class="h-12 w-12 text-muted-foreground/30" />
          <p class="text-muted-foreground">No movies from this collection are in your library.</p>
        </div>
      {:else}
        <div class="mt-12 grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {#each sortedMovies as movie (movie.id)}
            <a
              href={movie.id ? `/movies/${movie.id}` : '#'}
              class="group hairline block overflow-hidden rounded-2xl bg-surface/40 transition-all duration-300 hover:-translate-y-1 hover:bg-surface/70"
            >
              <!-- Backdrop / poster area -->
              <div
                class="relative aspect-video w-full overflow-hidden"
                style={movie.backdropUrl ? '' : fallbackGradient(movie.id ?? '')}
              >
                {#if movie.backdropUrl}
                  <img
                    src={movie.backdropUrl}
                    alt={movie.title}
                    loading="lazy"
                    class="h-full w-full object-cover transition-transform duration-500 group-hover:scale-105"
                    onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
                  />
                {/if}
                <div class="absolute inset-0 bg-gradient-to-t from-background/80 via-transparent to-transparent"></div>

                <!-- Year badge -->
                {#if movie.year}
                  <div class="absolute right-3 top-3 rounded-md bg-black/50 px-2 py-0.5 text-[10px] font-semibold text-white backdrop-blur-sm">
                    {movie.year}
                  </div>
                {/if}

                <!-- Play overlay -->
                <div class="absolute inset-0 flex items-center justify-center bg-black/30 opacity-0 backdrop-blur-[1px] transition-opacity duration-200 group-hover:opacity-100">
                  <span class="flex h-12 w-12 items-center justify-center rounded-full bg-white/90">
                    <Play class="h-5 w-5 translate-x-0.5 fill-black text-black" />
                  </span>
                </div>

                <!-- Poster thumbnail floating left -->
                {#if movie.posterUrl}
                  <img
                    src={movie.posterUrl}
                    alt=""
                    loading="lazy"
                    class="absolute bottom-3 left-3 h-16 w-11 rounded-md object-cover shadow-lg ring-1 ring-white/10"
                    onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
                  />
                {/if}
              </div>

              <!-- Details below -->
              <div class="p-4">
                <h3 class="truncate font-semibold text-foreground">{movie.title ?? 'Unknown'}</h3>
                <div class="mt-1.5 flex flex-wrap items-center gap-x-3 gap-y-1 text-[11px] uppercase tracking-[0.15em] text-muted-foreground">
                  {#if movie.genres && movie.genres.length > 0}
                    <span>{movie.genres.slice(0, 2).join(' · ')}</span>
                  {/if}
                  {#if movie.runtimeMinutes}
                    <span class="flex items-center gap-1">
                      <Clock class="h-3 w-3" />{formatRuntime(movie.runtimeMinutes)}
                    </span>
                  {/if}
                  {#if movie.voteAverage && movie.voteAverage > 0}
                    <span class="flex items-center gap-1 normal-case tracking-normal text-amber-300">
                      <Star class="h-3 w-3 fill-current" />{movie.voteAverage.toFixed(1)}
                    </span>
                  {/if}
                </div>
                {#if movie.overview}
                  <p class="mt-2 line-clamp-2 text-xs leading-relaxed text-muted-foreground">{movie.overview}</p>
                {/if}
              </div>
            </a>
          {/each}
        </div>
      {/if}
    </div>
  {/if}
</div>
