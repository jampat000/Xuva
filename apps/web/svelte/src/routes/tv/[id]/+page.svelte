<script lang="ts">
  import { page } from '$app/state';
  import { appState } from '$lib/stores/appState.svelte';
  import { onMount } from 'svelte';
  import {
    Star, ChevronLeft, Play, Plus, Check, Tv, User,
    Clapperboard, X, Shield, ChevronDown, Volume2, Captions, FileVideo, Film, Gauge, Layers,
  } from 'lucide-svelte';
  import { toggleWatchlist, isInWatchlist } from '$lib/stores/watchlistStore.svelte';
  import Header from '$lib/components/Header.svelte';
  import ErrorState from '$lib/components/ErrorState.svelte';
  import SubtitleSelector from '$lib/components/SubtitleSelector.svelte';
  import { getSeriesDetail, getSimilarSeries, type SimilarItem } from '$lib/api/home';
  import { getMetadataRecords, refreshMetadataItem, getMetadataCandidates } from '$lib/api/browse';
  import { getMediaSourceDetail, getMediaSourceTracks, getPlaybackDecision, type MediaSourceItem, type ProbeTrack, type PlaybackDecisionResponse } from '$lib/api/details';
  import { formatResolution, formatBitrate, formatChannels, formatCodec, formatLanguage, audioSummary, playabilityBadge, playabilityShortLabel } from '$lib/utils/mediaFormat';
  import { getPerformanceSettings } from '$lib/api/operator';
  import PlayabilityMatrix from '$lib/components/PlayabilityMatrix.svelte';
  import type { SeriesDetailResponse } from '$lib/api/home';
  import type { MetadataRecord, MetadataCredit, TMDBCandidate } from '$lib/api/browse';

  // ── Per-episode technical detail cache ────────────────────────────────────
  // Episodes load their codec/track info on-demand when the user clicks
  // "Details" — fetching this upfront for a 10-season show would mean
  // hundreds of extra requests. The Map persists across season collapses so
  // re-expanding doesn't refetch.
  const deviceProfiles = ['web', 'apple-tv', 'android-tv'] as const;
  const deviceLabels: Record<string, string> = { web: 'Web', 'apple-tv': 'Apple TV', 'android-tv': 'Android TV' };
  type DeviceDecisions = { web?: PlaybackDecisionResponse; 'apple-tv'?: PlaybackDecisionResponse; 'android-tv'?: PlaybackDecisionResponse };
  type EpisodeTech = {
    source: MediaSourceItem | null;
    audioTracks: ProbeTrack[];
    subtitleTracks: ProbeTrack[];
    decision: PlaybackDecisionResponse | null;
    decisions: DeviceDecisions;
    loading: boolean;
  };
  let episodeTech = $state(new Map<string, EpisodeTech>());
  let expandedEpisodes = $state<Set<string>>(new Set());

  async function loadEpisodeTech(mediaSourceId: string) {
    if (episodeTech.has(mediaSourceId) && !episodeTech.get(mediaSourceId)?.loading) return;
    const next = new Map(episodeTech);
    next.set(mediaSourceId, { source: null, audioTracks: [], subtitleTracks: [], decision: null, decisions: {}, loading: true });
    episodeTech = next;
    try {
      // Fetch source + tracks + all device profile decisions in parallel.
      // getPlaybackDecision is side-effect-free (does NOT start a transcode).
      const [src, tr, ...decResults] = await Promise.allSettled([
        getMediaSourceDetail(mediaSourceId),
        getMediaSourceTracks(mediaSourceId),
        ...deviceProfiles.map(p => getPlaybackDecision(mediaSourceId, { clientProfile: p })),
      ]);
      const decisions: DeviceDecisions = {};
      deviceProfiles.forEach((p, i) => {
        if (decResults[i].status === 'fulfilled') decisions[p] = decResults[i].value;
      });
      const updated = new Map(episodeTech);
      updated.set(mediaSourceId, {
        source: src.status === 'fulfilled' ? src.value : null,
        audioTracks: tr.status === 'fulfilled' ? (tr.value.audioTracks ?? []) : [],
        subtitleTracks: tr.status === 'fulfilled' ? (tr.value.subtitleTracks ?? []) : [],
        decision: decisions.web ?? null,
        decisions,
        loading: false,
      });
      episodeTech = updated;
    } catch {
      // Swallow — UI will show "couldn't load tracks" next to the button.
      const updated = new Map(episodeTech);
      updated.set(mediaSourceId, { source: null, audioTracks: [], subtitleTracks: [], decision: null, decisions: {}, loading: false });
      episodeTech = updated;
    }
  }

  function toggleEpisodeDetails(mediaSourceId: string) {
    const next = new Set(expandedEpisodes);
    if (next.has(mediaSourceId)) {
      next.delete(mediaSourceId);
    } else {
      next.add(mediaSourceId);
      loadEpisodeTech(mediaSourceId); // fire-and-forget
    }
    expandedEpisodes = next;
  }

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
  let showTrailer = $state(false);
  let showDataSources = $state(false);
  let selectedEncoderLabel = $state<string | null>(null);
  let similarItems = $state<SimilarItem[]>([]);
  let similarFallback = $state(false);
  let similarFallbackGenre = $state('');

  // ── Derived metadata fields ──────────────────────────────────────────────
  const title = $derived(metadata?.title ?? detail?.title ?? 'Unknown');
  const year = $derived(metadata?.year as number | undefined ?? (metadata?.firstAirDate ? parseInt(metadata.firstAirDate) : undefined));
  const overview = $derived(metadata?.overview ?? '');
  const tagline = $derived(metadata?.tagline ?? '');
  const posterUrl = $derived(metadata?.posterUrl ?? '');
  const backdropUrl = $derived(metadata?.backdropUrl ?? '');
  const logoUrl = $derived(metadata?.logoUrl ?? '');
  const genres = $derived(metadata?.genres ?? []);
  const rating = $derived(metadata?.voteAverage as number | undefined);
  const contentRating = $derived(metadata?.contentRating ?? '');
  const videoKey = $derived(metadata?.videoKey ?? '');
  const trailerPath = $derived(metadata?.trailerPath ?? '');
  // hasTrailer respects the user's "trailers on/off" setting from Settings →
  // General. When trailers are off globally, the button hides everywhere
  // (home hero already drops trailers entirely; this gates the detail page).
  const hasTrailer = $derived(appState.trailersEnabled !== false && !!(videoKey || trailerPath));
  const statusText = $derived(metadata?.statusText ?? '');

  const seasons = $derived(detail?.seasons ?? []);
  const seasonCount = $derived(seasons.length);
  const episodeCount = $derived(seasons.reduce((sum, s) => sum + (s.episodes?.length ?? 0), 0));

  // Credits
  const cast = $derived<MetadataCredit[]>(metadata?.cast ?? []);
  const creators = $derived<string[]>(metadata?.directors ?? []); // TMDB uses directors for series creators
  const writers = $derived<string[]>(metadata?.writers ?? []);

  // Network / studio chips
  const networkNames = $derived<string[]>([
    ...(metadata?.networks ?? []),
    ...(metadata?.studios ?? []),
  ].filter((v, i, a) => a.indexOf(v) === i).slice(0, 4));

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
    toggleWatchlist({ id, kind: 'series', title, year, posterUrl, backdropUrl, genres });
  }

  // Trailer
  function openTrailer() { showTrailer = true; }
  function closeTrailer() { showTrailer = false; }

  // Track which seasons are expanded. First season auto-expands.
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
    loading = true;
    error = null;
    // Fire metadata records request immediately so it runs in parallel,
    // but don't block the render on it — getSeriesDetail already embeds
    // the primary metadata record, so we can show the page right away.
    const metaPromise = getMetadataRecords('series', id).catch(() => ({ best: null, records: [] as typeof altRecords }));
    try {
      const detailResp = await getSeriesDetail(id);
      detail = detailResp;
      // Use the metadata embedded in the detail response immediately.
      metadata = (detailResp.metadata as typeof metadata) ?? null;
      loading = false; // Show the page — rich provider records load below
      const metaResp = await metaPromise;
      metadata = metaResp.best ?? (detailResp.metadata as typeof metadata) ?? null;
      altRecords = metaResp.records ?? [];
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load';
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

  onMount(() => {
    load();
    getPerformanceSettings().then(p => {
      selectedEncoderLabel = p.hardwareAcceleration?.selectedEncoder?.label ?? null;
    }).catch(() => {});
    // Fetch similar series lazily — don't block the main page load
    getSimilarSeries(id).then(r => {
      similarItems = r.items ?? [];
      similarFallback = r.genreFallback ?? false;
      similarFallbackGenre = r.fallbackGenre ?? '';
    }).catch(() => {});
  });
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
          <!-- Metadata strip -->
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
            {#if contentRating}
              <span class="opacity-30">·</span>
              <span class="inline-flex items-center gap-1 rounded border border-muted-foreground/40 px-1.5 py-0.5 text-[9px] font-semibold tracking-wider text-muted-foreground">
                <Shield class="h-2.5 w-2.5" />{contentRating}
              </span>
            {/if}
            {#if statusText}
              <span class="opacity-30">·</span>
              <span class={`rounded px-1.5 py-0.5 text-[9px] font-semibold tracking-wider ${statusText === 'Ended' ? 'bg-surface text-muted-foreground' : 'bg-green-500/20 text-green-400'}`}>
                {statusText}
              </span>
            {/if}
          </div>

          <!-- Logo or title -->
          {#if logoUrl}
            <div class="mt-4">
              <img
                src={logoUrl}
                alt={title}
                class="h-auto max-h-28 max-w-xs object-contain drop-shadow-[0_2px_16px_rgba(0,0,0,0.8)] md:max-h-36 md:max-w-sm"
                onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
              />
            </div>
          {:else}
            <h1 class="font-serif-display mt-3 text-[clamp(2rem,5vw,4rem)] leading-[0.95] tracking-tight">
              {title}
            </h1>
          {/if}

          <!-- Tagline -->
          {#if tagline}
            <p class="mt-2 text-sm italic text-muted-foreground/80">{tagline}</p>
          {/if}

          {#if overview}
            <p class="mt-5 max-w-2xl text-base leading-relaxed text-foreground/75">
              {overview}
            </p>
          {/if}

          <!-- Action row -->
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

            <!-- Trailer button -->
            {#if hasTrailer}
              <button
                type="button"
                onclick={openTrailer}
                class="hairline inline-flex items-center gap-2 rounded-full bg-foreground/[0.06] px-5 py-3.5 text-sm font-medium text-foreground backdrop-blur-sm transition-colors hover:bg-foreground/[0.12]"
              >
                <Clapperboard class="h-4 w-4" /> Trailer
              </button>
            {/if}

            <!-- Watchlist -->
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

          <!-- Episode counts + creators strip -->
          <div class="mt-6 flex flex-wrap gap-x-8 gap-y-3 text-xs text-muted-foreground">
            {#if seasonCount > 0}
              <div class="flex items-baseline gap-x-1.5">
                <span class="uppercase tracking-[0.2em] text-[10px]">Seasons</span>
                <span class="text-foreground/80">{seasonCount}</span>
              </div>
            {/if}
            {#if episodeCount > 0}
              <div class="flex items-baseline gap-x-1.5">
                <span class="uppercase tracking-[0.2em] text-[10px]">Episodes</span>
                <span class="text-foreground/80">{episodeCount}</span>
              </div>
            {/if}
            {#if creators.length > 0}
              <div class="flex flex-wrap items-baseline gap-x-1.5 gap-y-1">
                <span class="uppercase tracking-[0.2em] text-[10px]">{creators.length === 1 ? 'Creator' : 'Creators'}</span>
                {#each creators as c (c)}
                  <a href={`/people/${encodeURIComponent(c)}`} class="text-foreground/80 hover:text-foreground hover:underline">{c}</a>
                {/each}
              </div>
            {/if}
            {#if writers.length > 0}
              <div class="flex flex-wrap items-baseline gap-x-1.5 gap-y-1">
                <span class="uppercase tracking-[0.2em] text-[10px]">Writers</span>
                {#each writers.slice(0, 3) as w (w)}
                  <a href={`/people/${encodeURIComponent(w)}`} class="text-foreground/80 hover:text-foreground hover:underline">{w}</a>
                {/each}
              </div>
            {/if}
          </div>

          <!-- Network / studio chips -->
          {#if networkNames.length > 0}
            <div class="mt-4 flex flex-wrap gap-2">
              {#each networkNames as net (net)}
                <span class="hairline rounded-full bg-surface/40 px-3 py-1 text-[11px] text-muted-foreground">
                  {net}
                </span>
              {/each}
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
                                    class="absolute inset-0 flex items-center justify-center bg-black/30 opacity-0 transition-opacity duration-200 group-hover:opacity-100"
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
                                {#if epMsid}
                                  {@const isExpanded = expandedEpisodes.has(epMsid)}
                                  {@const tech = episodeTech.get(epMsid)}
                                  <button type="button" onclick={(e) => { e.preventDefault(); toggleEpisodeDetails(epMsid); }}
                                    class="mt-2 inline-flex items-center gap-1 rounded text-[11px] text-muted-foreground/75 transition-colors hover:text-foreground/90">
                                    <span>{isExpanded ? 'Hide' : 'File info & tracks'}</span>
                                    <ChevronDown class="h-3 w-3 transition-transform" style={isExpanded ? 'transform: rotate(180deg)' : ''} />
                                  </button>
                                  {#if isExpanded}
                                    <div class="mt-3 rounded-lg border border-border bg-background/50 p-3">
                                      {#if tech?.loading}
                                        <p class="text-[11px] italic text-muted-foreground/70">Loading tracks…</p>
                                      {:else if tech && (tech.source || tech.audioTracks.length > 0)}
                                        {#if true}
                                        {@const pills = [
                                          { icon: FileVideo, text: formatResolution(tech.source?.width, tech.source?.height) },
                                          { icon: Film, text: formatCodec(tech.source?.videoCodec) },
                                          { icon: Volume2, text: audioSummary(tech.audioTracks) },
                                          { icon: Layers, text: tech.source?.container ? tech.source.container.split(',')[0].toUpperCase() : '' },
                                          { icon: Gauge, text: formatBitrate(tech.source?.bitrate) },
                                          { icon: Captions, text: tech.subtitleTracks.length > 0 ? `${tech.subtitleTracks.length} subtitle${tech.subtitleTracks.length === 1 ? '' : 's'}` : '' },
                                        ].filter(p => p.text)}
                                        {#if pills.length > 0}
                                          <div class="flex flex-wrap gap-1.5 mb-3">
                                            {#each pills as p, pi (pi)}
                                              <span class="inline-flex items-center gap-1 rounded-full bg-surface/40 px-2.5 py-1 text-[11px] text-foreground/75">
                                                <p.icon class="h-3 w-3 text-muted-foreground/70" />
                                                {p.text}
                                              </span>
                                            {/each}
                                          </div>
                                        {/if}
                                        {/if}
                                        {#if tech.decisions && Object.keys(tech.decisions).length > 0}
                                          <div class="mb-3">
                                            <PlayabilityMatrix
                                              decisions={tech.decisions}
                                              {deviceProfiles}
                                              {deviceLabels}
                                              mediaSource={tech.source}
                                              audioTracks={tech.audioTracks}
                                              subtitleTracks={tech.subtitleTracks}
                                              encoderLabel={selectedEncoderLabel}
                                              compact={true}
                                            />
                                          </div>
                                        {/if}
                                        <div class="grid gap-3 sm:grid-cols-2">
                                          <!-- Audio -->
                                          <div>
                                            <div class="mb-1.5 flex items-center gap-1.5 text-[10px] uppercase tracking-[0.18em] text-muted-foreground">
                                              <Volume2 class="h-3 w-3" /> Audio · {tech.audioTracks.length}
                                            </div>
                                            {#if tech.audioTracks.length === 0}
                                              <p class="text-[11px] italic text-muted-foreground/60">No audio tracks.</p>
                                            {:else}
                                              <ul class="space-y-0.5">
                                                {#each tech.audioTracks as t, ti (t.index ?? ti)}
                                                  <li class="flex flex-wrap items-baseline gap-x-1.5 text-[11.5px] text-foreground/80">
                                                    <span class="font-medium">{formatCodec(t.codec) || '—'}</span>
                                                    {#if t.channels}<span class="text-foreground/60">{formatChannels(t.channels)}</span>{/if}
                                                    {#if t.language}<span class="italic text-muted-foreground">· {formatLanguage(t.language)}</span>{/if}
                                                    {#if t.default}<span class="ml-1 rounded bg-primary-glow/15 px-1 py-0 text-[9px] uppercase text-primary-glow">Default</span>{/if}
                                                    {#if t.forced}<span class="ml-1 rounded bg-amber-400/15 px-1 py-0 text-[9px] uppercase text-amber-300">Forced</span>{/if}
                                                  </li>
                                                {/each}
                                              </ul>
                                            {/if}
                                          </div>
                                          <!-- Subtitles -->
                                          <div>
                                            <div class="mb-1.5 flex items-center gap-1.5 text-[10px] uppercase tracking-[0.18em] text-muted-foreground">
                                              <Captions class="h-3 w-3" /> Subtitles · {tech.subtitleTracks.length}
                                            </div>
                                            {#if tech.subtitleTracks.length === 0}
                                              <p class="text-[11px] italic text-muted-foreground/60">None embedded.</p>
                                            {:else}
                                              <ul class="space-y-0.5">
                                                {#each tech.subtitleTracks as t, ti (t.index ?? ti)}
                                                  <li class="flex flex-wrap items-baseline gap-x-1.5 text-[11.5px] text-foreground/80">
                                                    <span class="font-medium">{formatLanguage(t.language) || formatCodec(t.codec) || '—'}</span>
                                                    {#if t.language && t.codec}<span class="text-foreground/55">· {formatCodec(t.codec)}</span>{/if}
                                                    {#if t.default}<span class="ml-1 rounded bg-primary-glow/15 px-1 py-0 text-[9px] uppercase text-primary-glow">Default</span>{/if}
                                                    {#if t.forced}<span class="ml-1 rounded bg-amber-400/15 px-1 py-0 text-[9px] uppercase text-amber-300">Forced</span>{/if}
                                                  </li>
                                                {/each}
                                              </ul>
                                            {/if}
                                          </div>
                                        </div>
                                      {:else}
                                        <p class="text-[11px] italic text-muted-foreground/60">Couldn't load tracks for this episode.</p>
                                      {/if}
                                    </div>
                                  {/if}
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
              <div class="scrollbar-none mt-5 -mx-1 flex gap-3 overflow-x-auto px-1 pb-3">
                {#each cast.slice(0, 16) as person, i (person.name ?? i)}
                  <a
                    href={`/people/${encodeURIComponent(person.name ?? '')}`}
                    class="group flex w-24 shrink-0 flex-col gap-2.5 text-center"
                  >
                    <div class="relative h-36 w-24 overflow-hidden rounded-xl bg-surface-elevated ring-2 ring-border/40 transition-all duration-300 group-hover:scale-[1.03] group-hover:ring-primary/40 group-hover:shadow-lg">
                      {#if person.profileUrl}
                        <img
                          src={person.profileUrl}
                          alt={person.name}
                          loading="lazy"
                          class="h-full w-full object-cover object-top transition-transform duration-300 group-hover:scale-105"
                          onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
                        />
                      {:else}
                        <div class="flex h-full w-full items-center justify-center text-muted-foreground/60">
                          <User class="h-10 w-10" />
                        </div>
                      {/if}
                    </div>
                    <div class="w-full min-w-0 px-0.5">
                      <p class="truncate text-[11px] font-semibold leading-tight text-foreground">{person.name ?? ''}</p>
                      {#if person.character}
                        <p class="mt-0.5 truncate text-[10px] leading-tight text-muted-foreground">{person.character}</p>
                      {/if}
                    </div>
                  </a>
                {/each}
              </div>
            </div>
          {/if}

          <!-- More Like This -->
          {#if similarItems.length > 0}
            <div class="mt-10 border-t border-border pt-8">
              <h3 class="text-sm font-semibold uppercase tracking-[0.22em] text-muted-foreground">
                {similarFallback && similarFallbackGenre ? `More in ${similarFallbackGenre}` : 'More Like This'}
              </h3>
              <div class="scrollbar-none mt-5 -mx-1 flex gap-3 overflow-x-auto pb-3 px-1">
                {#each similarItems as item (item.id)}
                  <a
                    href={`/tv/${item.id}`}
                    class="group relative flex shrink-0 flex-col gap-2"
                  >
                    <div class="relative aspect-[2/3] w-28 overflow-hidden rounded-xl bg-surface shadow-poster transition-all duration-300 group-hover:-translate-y-2 group-hover:scale-[1.04] group-hover:ring-[3px] group-hover:ring-white/85 group-hover:shadow-[0_28px_60px_-12px_oklch(0_0_0/0.85)] md:w-32">
                      {#if item.posterUrl}
                        <img
                          src={item.posterUrl}
                          alt={item.title}
                          loading="lazy"
                          class="h-full w-full object-cover"
                          onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
                        />
                      {:else}
                        <div class="flex h-full w-full items-center justify-center bg-surface-elevated">
                          <Tv class="h-6 w-6 text-muted-foreground/30" />
                        </div>
                      {/if}
                    </div>
                    <div class="min-w-0 px-0.5 w-28 md:w-32">
                      <p class="truncate text-xs font-medium leading-tight text-foreground">{item.title}</p>
                      {#if item.year}<p class="text-[11px] text-muted-foreground">{item.year}</p>{/if}
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
                {#if altRecords.length > 0}
                  <button
                    type="button"
                    onclick={() => { showDataSources = !showDataSources; }}
                    class="hairline rounded-full bg-foreground/[0.04] px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.08] hover:text-foreground"
                  >
                    Sources ({altRecords.length})
                  </button>
                {/if}
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

            <!-- ── Provider records (hidden by default) ──────────────────────── -->
            {#if altRecords.length > 0 && showDataSources}
              <div class="mt-4 space-y-1.5">
                <p class="text-[11px] font-medium uppercase tracking-[0.18em] text-muted-foreground/70">
                  Data sources ({altRecords.length})
                </p>
                {#each altRecords as rec (rec.provider ?? rec.itemId)}
                  <div class="hairline flex items-center gap-3 rounded-xl bg-surface/30 px-3 py-2.5">
                    {#if rec.posterUrl}
                      <img
                        src={rec.posterUrl}
                        alt={rec.title ?? ''}
                        class="h-[52px] w-[35px] shrink-0 rounded-md object-cover"
                        onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
                      />
                    {:else}
                      <div class="flex h-[52px] w-[35px] shrink-0 items-center justify-center rounded-md bg-surface-elevated/60 text-muted-foreground/50">
                        <Tv class="h-4 w-4" />
                      </div>
                    {/if}
                    <div class="min-w-0 flex-1">
                      <div class="flex flex-wrap items-center gap-1.5">
                        {#if rec.provider}
                          <span class="rounded px-1.5 py-0.5 text-[9px] font-bold uppercase tracking-wider bg-primary-glow/15 text-primary-glow">
                            {rec.provider}
                          </span>
                        {/if}
                        {#if rec.title}
                          <span class="truncate text-xs font-medium text-foreground">{rec.title}</span>
                          {#if rec.year}<span class="text-[11px] text-muted-foreground">{rec.year}</span>{/if}
                        {/if}
                      </div>
                      {#if rec.overview}
                        <p class="mt-0.5 line-clamp-1 text-[11px] text-muted-foreground">{rec.overview}</p>
                      {/if}
                    </div>
                  </div>
                {/each}
              </div>
            {/if}

            {#if showMetaPanel}
              <div class="mt-4 space-y-5">
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

<!-- ── Trailer modal ──────────────────────────────────────────────────────── -->
{#if showTrailer}
  <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/90 backdrop-blur-sm"
    role="dialog"
    aria-label="{title} Trailer"
    tabindex="-1"
    onclick={(e) => { if (e.target === e.currentTarget) closeTrailer(); }}
    onkeydown={(e) => e.key === 'Escape' && closeTrailer()}
  >
    <div class="relative w-full max-w-4xl px-4">
      <button
        type="button"
        onclick={closeTrailer}
        aria-label="Close trailer"
        class="absolute -top-12 right-4 flex h-9 w-9 items-center justify-center rounded-full bg-white/20 text-white transition-colors hover:bg-white/30"
      >
        <X class="h-5 w-5" />
      </button>
      <div class="aspect-video w-full overflow-hidden rounded-2xl bg-black shadow-2xl">
        {#if trailerPath}
          <!-- Local MP4 trailer — always works, no CSP concern -->
          <!-- svelte-ignore a11y_media_has_caption -->
          <video
            src={trailerPath}
            controls
            autoplay
            class="h-full w-full"
            title="{title} — Trailer"
          ></video>
        {:else if videoKey}
          <!-- YouTube embed — requires frame-src https://www.youtube.com in CSP -->
          <iframe
            src={`https://www.youtube.com/embed/${videoKey}?autoplay=1&rel=0`}
            class="h-full w-full"
            frameborder="0"
            allow="autoplay; fullscreen"
            title="{title} — Trailer"
          ></iframe>
        {/if}
      </div>
    </div>
  </div>
{/if}
