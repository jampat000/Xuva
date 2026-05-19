<script lang="ts">
  import { page } from '$app/state';
  import { appState } from '$lib/stores/appState.svelte';
  import { onMount } from 'svelte';
  import { Play, Plus, Check, Star, Clock, ChevronLeft, User, Film } from 'lucide-svelte';
  import { toggleWatchlist, isInWatchlist } from '$lib/stores/watchlistStore.svelte';
  import Header from '$lib/components/Header.svelte';
  import ErrorState from '$lib/components/ErrorState.svelte';
  import SubtitleSelector from '$lib/components/SubtitleSelector.svelte';
  import { getMovieDetail } from '$lib/api/home';
  import { getMetadataRecords, refreshMetadataItem, getMetadataCandidates } from '$lib/api/browse';
  import type { MovieDetailResponse } from '$lib/api/home';
  import type { MetadataRecord, TMDBCandidate } from '$lib/api/browse';

  const id = $derived(page.params.id ?? '');

  let detail = $state<MovieDetailResponse | null>(null);
  let metadata = $state<MetadataRecord | null>(null);
  let altRecords = $state<MetadataRecord[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);
  let fixingMeta = $state(false);
  let showMetaPanel = $state(false);
  let refreshing = $state(false);
  let tmdbCandidates = $state<TMDBCandidate[]>([]);
  let tmdbCandidatesLoading = $state(false);
  let tmdbCandidatesError = $state<string | null>(null);
  let manualTmdbId = $state('');
  let manualTmdbError = $state<string | null>(null);

  const title = $derived(
    (detail?.metadata as Record<string, unknown> | undefined)?.title as string
      ?? detail?.title
      ?? 'Unknown'
  );
  const year = $derived(
    (detail?.metadata as Record<string, unknown> | undefined)?.year as number | undefined
  );
  const overview = $derived(detail?.metadata?.overview ?? '');
  const posterUrl = $derived(
    (detail?.metadata as Record<string, unknown> | undefined)?.posterUrl as string | undefined
  );
  const backdropUrl = $derived(
    (detail?.metadata as Record<string, unknown> | undefined)?.backdropUrl as string | undefined
  );
  const genres = $derived(
    (detail?.metadata as Record<string, unknown> | undefined)?.genres as string[] | undefined ?? []
  );
  const rating = $derived(
    (detail?.metadata as Record<string, unknown> | undefined)?.voteAverage as number | undefined
  );
  const runtime = $derived(
    (detail?.metadata as Record<string, unknown> | undefined)?.runtime as string | undefined
  );
  const versionCount = $derived(detail?.versions?.length ?? 0);
  const mediaSourceId = $derived(detail?.versions?.[0]?.mediaSourceId);

  // Cast / crew / collection — from the merged best metadata record.
  // MetadataRecord has [key: string]: unknown so we cast explicitly.
  type Credit = { name?: string; character?: string; role?: string; profileUrl?: string; sortOrder?: number };
  type CollInfo = { id?: string; name?: string; posterUrl?: string; backdropUrl?: string; logoUrl?: string };
  const metaAny = $derived(metadata as Record<string, unknown> | null);
  const cast = $derived<Credit[]>((metaAny?.cast as Credit[] | undefined) ?? []);
  const directors = $derived<string[]>((metaAny?.directors as string[] | undefined) ?? []);
  const collection = $derived<CollInfo | undefined>(metaAny?.collection as CollInfo | undefined);

  // Build play URL including back-link and title for the player chrome
  const basePlayUrl = $derived(
    mediaSourceId
      ? `/play/${mediaSourceId}?title=${encodeURIComponent(title)}&back=${encodeURIComponent(`/movies/${id}`)}`
      : ''
  );
  let playHref = $state('');

  // Watchlist
  const inWatchlist = $derived(isInWatchlist(id, 'movie'));
  function handleWatchlist() {
    toggleWatchlist({
      id,
      kind: 'movie',
      title,
      year,
      posterUrl,
      backdropUrl,
      genres,
    });
  }

  async function load() {
    try {
      loading = true;
      error = null;
      const [detailResp, metaResp] = await Promise.all([
        getMovieDetail(id),
        getMetadataRecords('movie', id).catch(() => ({ best: null, records: [] }))
      ]);
      detail = detailResp;
      metadata = metaResp.best ?? null;
      altRecords = metaResp.records ?? [];
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load';
    } finally {
      loading = false;
    }
  }

  async function refreshMeta() {
    refreshing = true;
    try {
      await refreshMetadataItem({ kind: 'movie', id });
      await load();
    } finally {
      refreshing = false;
    }
  }

  async function openMetaPanel() {
    showMetaPanel = !showMetaPanel;
    if (showMetaPanel && tmdbCandidates.length === 0) {
      await fetchTMDBCandidates();
    }
  }

  async function fetchTMDBCandidates() {
    tmdbCandidatesLoading = true;
    tmdbCandidatesError = null;
    try {
      const res = await getMetadataCandidates('movie', title, year);
      tmdbCandidates = res.candidates ?? [];
    } catch (e) {
      tmdbCandidatesError = e instanceof Error ? e.message : 'Search failed';
    } finally {
      tmdbCandidatesLoading = false;
    }
  }

  async function pickTMDBCandidate(candidate: TMDBCandidate) {
    fixingMeta = true;
    try {
      await refreshMetadataItem({ kind: 'movie', id, tmdbOverrideId: candidate.id });
      showMetaPanel = false;
      tmdbCandidates = [];
      await load();
    } finally {
      fixingMeta = false;
    }
  }

  async function applyManualTmdbId() {
    const numId = parseInt(manualTmdbId.trim(), 10);
    if (!numId || numId <= 0) {
      manualTmdbError = 'Enter a valid TMDB ID (numbers only)';
      return;
    }
    fixingMeta = true;
    manualTmdbError = null;
    try {
      await refreshMetadataItem({ kind: 'movie', id, tmdbOverrideId: numId });
      showMetaPanel = false;
      manualTmdbId = '';
      tmdbCandidates = [];
      await load();
    } finally {
      fixingMeta = false;
    }
  }

  onMount(load);
</script>

<svelte:head>
  <title>{title} — {appState.serverName}</title>
  <meta name="description" content={overview || `Watch ${title} on Xuva.`} />
</svelte:head>

<div class="min-h-screen bg-background">
  <Header />

  {#if loading}
    <div class="flex min-h-[60vh] items-center justify-center">
      <div class="h-10 w-10 animate-spin rounded-full border-2 border-border border-t-primary-glow"></div>
    </div>

  {:else if error}
    <ErrorState
      title="Can't load this title"
      message="Make sure your Xuva server is running, then try again."
      actions={[
        { label: 'Try again', onClick: load },
        { label: 'Browse movies', href: '/movies' },
      ]}
      diagnosticInfo={`Movie ID: ${id}\nError: ${error}`}
    />

  {:else}
    <!-- Backdrop -->
    <div class="relative -mt-16 h-[60vh] min-h-[480px] w-full overflow-hidden">
      {#if backdropUrl}
        <img
          src={backdropUrl}
          alt=""
          class="h-full w-full object-cover object-top"
          onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
        />
      {/if}
      <div class="absolute inset-0 bg-gradient-to-r from-background via-background/70 to-transparent"></div>
      <div class="absolute inset-x-0 bottom-0 h-2/3 bg-gradient-to-t from-background to-transparent"></div>
      <div class="absolute inset-x-0 top-0 h-24 bg-gradient-to-b from-background/80 to-transparent"></div>
    </div>

    <div class="relative -mt-48 px-6 pb-32 md:px-12 lg:px-20">
      <a href="/movies" class="mb-8 inline-flex items-center gap-2 text-sm text-muted-foreground transition-colors hover:text-foreground">
        <ChevronLeft class="h-4 w-4" /> Back to Movies
      </a>

      <div class="grid gap-10 md:grid-cols-[240px_minmax(0,1fr)] lg:grid-cols-[280px_minmax(0,1fr)] lg:gap-16">
        <!-- Poster -->
        <div class="shrink-0">
          <div
            class="shadow-poster aspect-[2/3] w-full overflow-hidden rounded-2xl"
            style="background: linear-gradient(135deg, #1e3a5f, #0f172a);"
          >
            {#if posterUrl}
              <img
                src={posterUrl}
                alt={title}
                class="h-full w-full object-cover"
                onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
              />
            {/if}
          </div>
        </div>

        <!-- Details -->
        <div>
          <div class="flex flex-wrap items-center gap-x-3 gap-y-1 text-[11px] uppercase tracking-[0.22em] text-muted-foreground">
            {#if year}<span>{year}</span>{/if}
            {#if year && genres.length}<span class="opacity-30">·</span>{/if}
            {#each genres as g, i (g)}
              <span>{g}</span>{#if i < genres.length - 1}<span class="opacity-30">·</span>{/if}
            {/each}
            {#if runtime}
              <span class="opacity-30">·</span>
              <span class="flex items-center gap-1"><Clock class="h-3 w-3" />{runtime}</span>
            {/if}
            {#if rating}
              <span class="opacity-30">·</span>
              <span class="flex items-center gap-1 text-amber-300"><Star class="h-3 w-3 fill-current" />{rating.toFixed(1)}</span>
            {/if}
          </div>

          <h1 class="font-serif-display mt-3 text-[clamp(2rem,5vw,4rem)] leading-[0.95] tracking-tight">
            {title}
          </h1>

          {#if overview}
            <p class="mt-5 max-w-2xl text-base leading-relaxed text-foreground/75">
              {overview}
            </p>
          {/if}

          <div class="mt-8 flex flex-wrap items-center gap-3">
            {#if mediaSourceId}
              <a
                href={playHref || basePlayUrl || `/play/${mediaSourceId}`}
                class="inline-flex items-center gap-2.5 rounded-full bg-foreground px-7 py-3.5 text-sm font-semibold text-background transition-all hover:bg-foreground/90"
              >
                <Play class="h-4 w-4 fill-background" /> Play
              </a>
              <SubtitleSelector {mediaSourceId} {basePlayUrl} bind:playHref />
            {:else}
              <button
                disabled
                class="inline-flex cursor-not-allowed items-center gap-2.5 rounded-full bg-foreground/30 px-7 py-3.5 text-sm font-semibold text-background/60"
              >
                <Play class="h-4 w-4 fill-background/60" /> No source
              </button>
            {/if}
            <button
              type="button"
              onclick={handleWatchlist}
              aria-label={inWatchlist ? 'Remove from watchlist' : 'Add to watchlist'}
              title={inWatchlist ? 'Remove from watchlist' : 'Add to watchlist'}
              class={`hairline flex h-12 w-12 items-center justify-center rounded-full backdrop-blur-md transition-all ${
                inWatchlist
                  ? 'bg-primary/20 text-primary-glow hover:bg-primary/30'
                  : 'bg-foreground/5 text-foreground hover:bg-foreground/10'
              }`}
            >
              {#if inWatchlist}
                <Check class="h-5 w-5" />
              {:else}
                <Plus class="h-5 w-5" />
              {/if}
            </button>
          </div>

          <!-- Technical info strip -->
          {#if versionCount > 0}
            <div class="mt-8 flex flex-wrap gap-x-6 gap-y-2 text-[11px] uppercase tracking-[0.2em] text-muted-foreground">
              <span><span class="text-foreground/80">{versionCount}</span> {versionCount === 1 ? 'version' : 'versions'}</span>
              {#if directors.length > 0}
                <span>Dir. <span class="text-foreground/80">{directors[0]}</span></span>
              {/if}
            </div>
          {/if}

          <!-- Cast strip -->
          {#if cast.length > 0}
            <div class="mt-10 border-t border-border pt-8">
              <h3 class="text-sm font-semibold uppercase tracking-[0.22em] text-muted-foreground">Cast</h3>
              <div class="scrollbar-none mt-5 -mx-1 flex gap-4 overflow-x-auto px-1 pb-3">
                {#each cast.slice(0, 16) as person, i (person.name ?? i)}
                  <a
                    href={`/people/${encodeURIComponent(person.name ?? '')}`}
                    class="group flex w-[72px] shrink-0 flex-col items-center gap-2 text-center"
                  >
                    <div class="relative h-[72px] w-[72px] overflow-hidden rounded-full bg-surface-elevated ring-2 ring-border/40 transition-all duration-300 group-hover:ring-primary/40">
                      {#if person.profileUrl}
                        <img
                          src={person.profileUrl}
                          alt={person.name}
                          loading="lazy"
                          class="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105"
                          onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
                        />
                      {:else}
                        <div class="flex h-full w-full items-center justify-center text-muted-foreground">
                          <User class="h-7 w-7" />
                        </div>
                      {/if}
                    </div>
                    <div class="w-full min-w-0">
                      <p class="truncate text-[11px] font-medium leading-tight text-foreground">{person.name ?? ''}</p>
                      {#if person.character}
                        <p class="mt-0.5 truncate text-[10px] leading-tight text-muted-foreground">{person.character}</p>
                      {/if}
                    </div>
                  </a>
                {/each}
              </div>
            </div>
          {/if}

          <!-- Collection banner -->
          {#if collection?.id && collection.name}
            <div class="mt-10 border-t border-border pt-8">
              <h3 class="text-sm font-semibold uppercase tracking-[0.22em] text-muted-foreground">Part of a Collection</h3>
              <a
                href={`/collections/${encodeURIComponent(collection.id)}`}
                class="hairline mt-4 flex items-center gap-0 overflow-hidden rounded-2xl bg-surface/30 transition-all duration-300 hover:bg-surface/60 hover:-translate-y-0.5"
              >
                {#if collection.posterUrl}
                  <img
                    src={collection.posterUrl}
                    alt={collection.name}
                    loading="lazy"
                    class="h-24 w-16 shrink-0 object-cover"
                    onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
                  />
                {:else}
                  <div class="flex h-24 w-16 shrink-0 items-center justify-center bg-surface-elevated text-muted-foreground">
                    <Film class="h-6 w-6" />
                  </div>
                {/if}
                <div class="flex min-w-0 flex-1 flex-col gap-1 p-5">
                  <span class="text-[10px] uppercase tracking-[0.3em] text-muted-foreground">Collection</span>
                  <span class="truncate font-semibold text-foreground">{collection.name}</span>
                  <span class="text-xs text-muted-foreground">View all movies →</span>
                </div>
              </a>
            </div>
          {/if}

          <!-- Metadata correction panel (#9) -->
          <div class="mt-10 border-t border-border pt-8">
            <div class="flex items-center justify-between">
              <h3 class="text-sm font-semibold uppercase tracking-[0.22em] text-muted-foreground">Metadata</h3>
              <div class="flex items-center gap-2">
                <button
                  type="button"
                  onclick={refreshMeta}
                  disabled={refreshing}
                  class="hairline rounded-full bg-foreground/[0.04] px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.08] hover:text-foreground disabled:opacity-40"
                >
                  {refreshing ? 'Refreshing…' : 'Refresh'}
                </button>
                <button
                  type="button"
                  onclick={openMetaPanel}
                  class="hairline rounded-full bg-foreground/[0.04] px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.08] hover:text-foreground"
                >
                  Fix match
                </button>
              </div>
            </div>

            {#if showMetaPanel}
              <div class="mt-4 space-y-5">
                <!-- TMDB Candidates -->
                <div>
                  <div class="mb-2 flex items-center justify-between">
                    <p class="text-xs font-medium text-muted-foreground">TMDB matches for "{title}"</p>
                    <button
                      type="button"
                      onclick={fetchTMDBCandidates}
                      disabled={tmdbCandidatesLoading}
                      class="text-[11px] text-primary-glow hover:underline disabled:opacity-40"
                    >
                      {tmdbCandidatesLoading ? 'Searching…' : 'Search again'}
                    </button>
                  </div>
                  {#if tmdbCandidatesLoading}
                    <div class="flex items-center gap-2 py-4 text-xs text-muted-foreground">
                      <div class="h-4 w-4 animate-spin rounded-full border-2 border-border border-t-primary-glow"></div>
                      Searching TMDB…
                    </div>
                  {:else if tmdbCandidatesError}
                    <p class="text-xs text-red-400">{tmdbCandidatesError}</p>
                  {:else if tmdbCandidates.length === 0}
                    <p class="text-xs text-muted-foreground">No TMDB results found. Try the manual ID below.</p>
                  {:else}
                    <div class="space-y-2">
                      {#each tmdbCandidates as c (c.id)}
                        <button
                          type="button"
                          disabled={fixingMeta}
                          onclick={() => pickTMDBCandidate(c)}
                          class="hairline flex w-full items-start gap-4 rounded-xl bg-surface/40 p-4 text-left transition-colors hover:bg-surface/70 disabled:opacity-50"
                        >
                          {#if c.posterUrl}
                            <img src={c.posterUrl} alt={c.title} class="h-16 w-11 shrink-0 rounded-lg object-cover" onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')} />
                          {:else}
                            <div class="flex h-16 w-11 shrink-0 items-center justify-center rounded-lg bg-surface-elevated/60 text-muted-foreground">
                              <Film class="h-5 w-5" />
                            </div>
                          {/if}
                          <div class="min-w-0 flex-1">
                            <div class="flex items-baseline gap-2">
                              <span class="font-semibold">{c.title}</span>
                              {#if c.year}<span class="text-xs text-muted-foreground">{c.year}</span>{/if}
                            </div>
                            {#if c.voteAverage && c.voteAverage > 0}
                              <div class="mt-0.5 flex items-center gap-1 text-[11px] text-amber-300">
                                <Star class="h-3 w-3 fill-current" />{c.voteAverage.toFixed(1)}
                                {#if c.voteCount}<span class="text-muted-foreground">({c.voteCount.toLocaleString()} votes)</span>{/if}
                              </div>
                            {/if}
                            {#if c.overview}
                              <p class="mt-1 line-clamp-2 text-xs text-muted-foreground">{c.overview}</p>
                            {/if}
                            <div class="mt-1 text-[10px] text-muted-foreground/50">TMDB #{c.id}</div>
                          </div>
                        </button>
                      {/each}
                    </div>
                  {/if}
                </div>

                <!-- Manual TMDB ID -->
                <div class="border-t border-border/40 pt-4">
                  <p class="mb-2 text-xs font-medium text-muted-foreground">Or enter a TMDB ID manually</p>
                  <div class="flex gap-2">
                    <input
                      type="number"
                      placeholder="e.g. 12345"
                      bind:value={manualTmdbId}
                      class="hairline min-w-0 flex-1 rounded-xl bg-surface/40 px-3 py-2 text-sm placeholder-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary-glow/50"
                    />
                    <button
                      type="button"
                      disabled={fixingMeta || !manualTmdbId}
                      onclick={applyManualTmdbId}
                      class="hairline shrink-0 rounded-xl bg-foreground/[0.08] px-4 py-2 text-xs font-medium text-foreground transition-colors hover:bg-foreground/[0.14] disabled:opacity-40"
                    >
                      {fixingMeta ? 'Applying…' : 'Apply'}
                    </button>
                  </div>
                  {#if manualTmdbError}
                    <p class="mt-1 text-xs text-red-400">{manualTmdbError}</p>
                  {/if}
                </div>
              </div>
            {/if}
          </div>
        </div>
      </div>
    </div>
  {/if}
</div>
