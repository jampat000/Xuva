<script lang="ts">
  import { page } from '$app/state';
  import { appState } from '$lib/stores/appState.svelte';
  import { onMount } from 'svelte';
  import { Star, ChevronLeft, Play, Plus, Check, Tv, User } from 'lucide-svelte';
  import { toggleWatchlist, isInWatchlist } from '$lib/stores/watchlistStore.svelte';
  import Header from '$lib/components/Header.svelte';
  import ErrorState from '$lib/components/ErrorState.svelte';
  import SubtitleSelector from '$lib/components/SubtitleSelector.svelte';
  import { getSeriesDetail } from '$lib/api/home';
  import { getMetadataRecords, refreshMetadataItem, getMetadataCandidates } from '$lib/api/browse';
  import type { SeriesDetailResponse } from '$lib/api/home';
  import type { MetadataRecord, TMDBCandidate } from '$lib/api/browse';

  const id = $derived(page.params.id ?? '');

  let detail = $state<SeriesDetailResponse | null>(null);
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

  const seasons = $derived(detail?.seasons ?? []);
  const seasonCount = $derived(seasons.length);
  const episodeCount = $derived(seasons.reduce((sum, s) => sum + (s.episodes?.length ?? 0), 0));

  // Cast from merged best metadata record.
  type Credit = { name?: string; character?: string; role?: string; profileUrl?: string; sortOrder?: number };
  const metaAny = $derived(metadata as Record<string, unknown> | null);
  const cast = $derived<Credit[]>((metaAny?.cast as Credit[] | undefined) ?? []);

  // Find first playable episode across all seasons
  const firstMediaSourceId = $derived(() => {
    for (const season of seasons) {
      for (const ep of season.episodes ?? []) {
        const msid = ep.versions?.[0]?.mediaSourceId;
        if (msid) return msid;
      }
    }
    return undefined;
  });

  // Build play URL with back-link and title for player chrome
  const basePlayUrl = $derived(
    firstMediaSourceId()
      ? `/play/${firstMediaSourceId()}?title=${encodeURIComponent(title)}&back=${encodeURIComponent(`/tv/${id}`)}`
      : ''
  );

  function episodePlayUrl(mediaSourceId: string, epTitle?: string): string {
    const t = epTitle ? `${title} — ${epTitle}` : title;
    return `/play/${mediaSourceId}?title=${encodeURIComponent(t)}&back=${encodeURIComponent(`/tv/${id}`)}`;
  }

  let playHref = $state('');

  // Watchlist
  const inWatchlist = $derived(isInWatchlist(id, 'series'));
  function handleWatchlist() {
    toggleWatchlist({
      id,
      kind: 'series',
      title,
      year,
      posterUrl,
      backdropUrl,
      genres,
    });
  }

  // Track which seasons are expanded. First season auto-expands once data
  // loads so users see episodes immediately without an extra click.
  let expandedSeasons = $state<Set<string>>(new Set());
  $effect(() => {
    if (seasons.length > 0 && expandedSeasons.size === 0) {
      const first = seasons[0];
      if (first?.id) expandedSeasons = new Set([first.id]);
    }
  });
  function toggleSeason(seasonId: string) {
    const next = new Set(expandedSeasons);
    if (next.has(seasonId)) next.delete(seasonId);
    else next.add(seasonId);
    expandedSeasons = next;
  }

  async function load() {
    try {
      loading = true;
      error = null;
      const [detailResp, metaResp] = await Promise.all([
        getSeriesDetail(id),
        getMetadataRecords('series', id).catch(() => ({ best: null, records: [] }))
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
      await refreshMetadataItem({ kind: 'series', id });
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
      const res = await getMetadataCandidates('series', title, year);
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
      await refreshMetadataItem({ kind: 'series', id, tmdbOverrideId: candidate.id });
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
      await refreshMetadataItem({ kind: 'series', id, tmdbOverrideId: numId });
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
        { label: 'Browse TV', href: '/tv' },
      ]}
      diagnosticInfo={`Series ID: ${id}\nError: ${error}`}
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
      <a href="/tv" class="mb-8 inline-flex items-center gap-2 text-sm text-muted-foreground transition-colors hover:text-foreground">
        <ChevronLeft class="h-4 w-4" /> Back to TV
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
            {#if seasonCount > 0}
              <span class="opacity-30">·</span>
              <span class="flex items-center gap-1">
                <Tv class="h-3 w-3" />{seasonCount} Season{seasonCount !== 1 ? 's' : ''}
              </span>
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
            {#if firstMediaSourceId()}
              <a
                href={playHref || basePlayUrl || `/play/${firstMediaSourceId()}`}
                class="inline-flex items-center gap-2.5 rounded-full bg-foreground px-7 py-3.5 text-sm font-semibold text-background transition-all hover:bg-foreground/90"
              >
                <Play class="h-4 w-4 fill-background" /> Play from Start
              </a>
              {#if firstMediaSourceId()}
                <SubtitleSelector mediaSourceId={firstMediaSourceId()!} {basePlayUrl} bind:playHref />
              {/if}
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

          <!-- Episode counts strip -->
          {#if seasonCount > 0}
            <div class="mt-8 flex flex-wrap gap-x-6 gap-y-2 text-[11px] uppercase tracking-[0.2em] text-muted-foreground">
              <span><span class="text-foreground/80">{seasonCount}</span> {seasonCount === 1 ? 'season' : 'seasons'}</span>
              {#if episodeCount > 0}
                <span><span class="text-foreground/80">{episodeCount}</span> {episodeCount === 1 ? 'episode' : 'episodes'}</span>
              {/if}
            </div>
          {/if}

          <!-- Seasons + Episodes -->
          {#if seasons.length > 0}
            <div class="mt-10 border-t border-border pt-8">
              <h3 class="text-sm font-semibold uppercase tracking-[0.22em] text-muted-foreground">Seasons</h3>
              <div class="mt-4 space-y-3">
                {#each seasons as season, i (season.id ?? i)}
                  {@const epCount = season.episodes?.length ?? 0}
                  {@const firstEp = season.episodes?.[0]?.versions?.[0]?.mediaSourceId}
                  {@const seasonRecord = season as Record<string, unknown>}
                  {@const seasonPoster = seasonRecord.posterUrl as string | undefined}
                  {@const seasonName = (seasonRecord.name as string | undefined) || `Season ${i + 1}`}
                  {@const seasonOverview = seasonRecord.overview as string | undefined}
                  {@const isOpen = expandedSeasons.has(season.id ?? `s${i}`)}
                  <div class="hairline overflow-hidden rounded-xl bg-surface/30">
                    <!-- Season header (clickable to toggle) -->
                    <button
                      type="button"
                      onclick={() => toggleSeason(season.id ?? `s${i}`)}
                      class="flex w-full items-center gap-4 px-4 py-3 text-left transition-colors hover:bg-surface/60"
                    >
                      {#if seasonPoster}
                        <img
                          src={seasonPoster}
                          alt={seasonName}
                          loading="lazy"
                          class="h-16 w-12 shrink-0 rounded-md object-cover shadow-md"
                          onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
                        />
                      {:else}
                        <div class="h-16 w-12 shrink-0 rounded-md bg-gradient-to-br from-surface to-surface-elevated"></div>
                      {/if}
                      <div class="min-w-0 flex-1">
                        <div class="flex items-center gap-3">
                          <span class="font-medium text-foreground">{seasonName}</span>
                          {#if epCount > 0}
                            <span class="text-xs text-muted-foreground">{epCount} ep{epCount !== 1 ? 's' : ''}</span>
                          {/if}
                        </div>
                        {#if seasonOverview}
                          <p class="mt-0.5 line-clamp-1 text-xs text-muted-foreground">{seasonOverview}</p>
                        {/if}
                      </div>
                      {#if firstEp}
                        <a
                          href={episodePlayUrl(firstEp, `${seasonName}, Episode 1`)}
                          onclick={(e) => e.stopPropagation()}
                          class="inline-flex items-center gap-1.5 rounded-full bg-foreground/[0.06] px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.12] hover:text-foreground"
                        >
                          <Play class="h-3 w-3 fill-current" /> Play
                        </a>
                      {/if}
                      <span class="text-muted-foreground transition-transform duration-200" style={`transform: rotate(${isOpen ? 90 : 0}deg)`}>›</span>
                    </button>

                    <!-- Episode list (collapsed by default for all but the first season) -->
                    {#if isOpen && epCount > 0}
                      <div class="border-t border-border bg-background/40 p-3">
                        <ul class="space-y-2">
                          {#each season.episodes ?? [] as ep, epIdx (ep.id ?? epIdx)}
                            {@const epRecord = ep as Record<string, unknown>}
                            {@const epThumb = (epRecord.thumbnailUrl as string | undefined) || seasonPoster}
                            {@const epOverview = epRecord.overview as string | undefined}
                            {@const epRuntime = epRecord.runtimeMinutes as number | undefined}
                            {@const epAirDate = epRecord.airDate as string | undefined}
                            {@const epMsid = ep.versions?.[0]?.mediaSourceId}
                            {@const epLabel = `E${String(ep.episodeNumber ?? epIdx + 1).padStart(2, '0')}`}
                            {@const epTitle = ep.title || epLabel}
                            <li class="group flex gap-3 rounded-lg p-2 transition-colors hover:bg-surface/40">
                              <!-- 16:9 thumbnail with play overlay -->
                              <div class="relative aspect-video w-32 shrink-0 overflow-hidden rounded-md bg-gradient-to-br from-surface to-surface-elevated md:w-40">
                                {#if epThumb}
                                  <img
                                    src={epThumb}
                                    alt={epTitle}
                                    loading="lazy"
                                    class="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105"
                                    onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
                                  />
                                {/if}
                                {#if epMsid}
                                  <a
                                    href={episodePlayUrl(epMsid, `${seasonName} — ${epLabel}`)}
                                    aria-label={`Play ${epTitle}`}
                                    class="absolute inset-0 flex items-center justify-center bg-black/30 opacity-0 backdrop-blur-[1px] transition-opacity duration-200 group-hover:opacity-100"
                                  >
                                    <span class="flex h-10 w-10 items-center justify-center rounded-full bg-foreground/95">
                                      <Play class="h-4 w-4 translate-x-0.5 fill-background text-background" />
                                    </span>
                                  </a>
                                {/if}
                                <div class="absolute left-1.5 top-1.5 rounded-sm bg-black/60 px-1.5 py-0.5 text-[10px] font-semibold tracking-wider text-white">
                                  {epLabel}
                                </div>
                              </div>

                              <!-- Episode meta -->
                              <div class="min-w-0 flex-1">
                                <div class="flex items-start justify-between gap-3">
                                  <h4 class="truncate text-sm font-semibold text-foreground">{epTitle}</h4>
                                  {#if epRuntime}
                                    <span class="shrink-0 text-[11px] uppercase tracking-wider text-muted-foreground">{epRuntime}m</span>
                                  {/if}
                                </div>
                                {#if epAirDate}
                                  <div class="mt-0.5 text-[10px] uppercase tracking-[0.15em] text-muted-foreground">{epAirDate}</div>
                                {/if}
                                {#if epOverview}
                                  <p class="mt-1 line-clamp-2 text-xs text-muted-foreground">{epOverview}</p>
                                {/if}
                              </div>
                            </li>
                          {/each}
                        </ul>
                      </div>
                    {/if}
                  </div>
                {/each}
              </div>
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

          <!-- Metadata correction panel -->
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
                              <Tv class="h-5 w-5" />
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
