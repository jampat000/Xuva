<script lang="ts">
  import { page } from '$app/state';
  import { appState } from '$lib/stores/appState.svelte';
  import { onMount } from 'svelte';
  import {
    Play, Plus, Check, Star, Clock, ChevronLeft, User, Film,
    Clapperboard, X, Shield, Volume2, Captions, FileVideo, Gauge, Layers, Languages,
  } from 'lucide-svelte';
  import { toggleWatchlist, isInWatchlist } from '$lib/stores/watchlistStore.svelte';
  import Header from '$lib/components/Header.svelte';
  import ErrorState from '$lib/components/ErrorState.svelte';
  import SubtitleSelector from '$lib/components/SubtitleSelector.svelte';
  import { getMovieDetail } from '$lib/api/home';
  import { getMetadataRecords, refreshMetadataItem, getMetadataCandidates } from '$lib/api/browse';
  import { getMediaSourceDetail, getMediaSourceTracks, type MediaSourceItem, type ProbeTrack } from '$lib/api/details';
  import type { MovieDetailResponse } from '$lib/api/home';
  import type { MetadataRecord, MetadataCredit, TMDBCandidate } from '$lib/api/browse';

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
  let showTrailer = $state(false);

  // ── Derived metadata fields ──────────────────────────────────────────────
  const title = $derived(metadata?.title ?? detail?.title ?? 'Unknown');
  const year = $derived(metadata?.year as number | undefined);
  const overview = $derived(metadata?.overview ?? '');
  const tagline = $derived(metadata?.tagline ?? '');
  const posterUrl = $derived(metadata?.posterUrl ?? '');
  const backdropUrl = $derived(metadata?.backdropUrl ?? '');
  const logoUrl = $derived(metadata?.logoUrl ?? '');
  const genres = $derived(metadata?.genres ?? []);
  const rating = $derived(metadata?.voteAverage as number | undefined);
  const runtime = $derived(metadata?.runtime as string | undefined ?? (
    metadata?.runtimeMinutes ? `${Math.floor(metadata.runtimeMinutes / 60)}h ${metadata.runtimeMinutes % 60}m` : undefined
  ));
  const contentRating = $derived(metadata?.contentRating ?? '');
  const videoKey = $derived(metadata?.videoKey ?? '');
  const trailerPath = $derived(metadata?.trailerPath ?? '');
  const hasTrailer = $derived(!!(videoKey || trailerPath));
  const versionCount = $derived(detail?.versions?.length ?? 0);
  const versions = $derived(detail?.versions ?? []);
  // Selected version drives the "File Info" and Play button. Defaults to the
  // first listed version (usually highest quality). Users with multiple versions
  // (e.g. 1080p and 4K of the same movie) can click another card to switch.
  let selectedVersionIdx = $state(0);
  const selectedVersion = $derived(versions[selectedVersionIdx]);
  const mediaSourceId = $derived(selectedVersion?.mediaSourceId);
  const qualityLabel = $derived(selectedVersion?.qualityLabel ?? '');

  // ── Per-version technical detail (codec, bitrate, tracks) ─────────────────
  // Fetched in parallel with the metadata records once a mediaSourceId is known.
  let mediaSource = $state<MediaSourceItem | null>(null);
  let audioTracks = $state<ProbeTrack[]>([]);
  let subtitleTracks = $state<ProbeTrack[]>([]);
  let tracksLoading = $state(false);

  // Re-fetch tracks + source detail whenever the selected version changes.
  // Wrapped in an effect so switching versions in the UI reactively updates the
  // File Info + Tracks panels without leaving stale data on screen.
  $effect(() => {
    const id = mediaSourceId;
    if (!id) {
      mediaSource = null; audioTracks = []; subtitleTracks = [];
      return;
    }
    tracksLoading = true;
    Promise.allSettled([getMediaSourceDetail(id), getMediaSourceTracks(id)])
      .then(([src, tr]) => {
        if (src.status === 'fulfilled') mediaSource = src.value; else mediaSource = null;
        if (tr.status === 'fulfilled') {
          audioTracks = tr.value.audioTracks ?? [];
          subtitleTracks = tr.value.subtitleTracks ?? [];
        } else {
          audioTracks = []; subtitleTracks = [];
        }
      })
      .finally(() => { tracksLoading = false; });
  });

  // Credits
  const cast = $derived<MetadataCredit[]>(metadata?.cast ?? []);
  const directors = $derived<string[]>(metadata?.directors ?? []);
  const writers = $derived<string[]>(metadata?.writers ?? []);

  // Studios / production companies
  const studioNames = $derived<string[]>([
    ...(metadata?.studios ?? []),
    ...(metadata?.productionCompanies ?? []),
  ].filter((v, i, a) => a.indexOf(v) === i).slice(0, 4));

  // Collection from metadata
  type CollInfo = { id?: string; name?: string; posterUrl?: string; backdropUrl?: string; logoUrl?: string };
  const collection = $derived<CollInfo | undefined>(metadata?.collection as CollInfo | undefined);

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
    toggleWatchlist({ id, kind: 'movie', title, year, posterUrl, backdropUrl, genres });
  }

  // Trailer
  function openTrailer() { showTrailer = true; }
  function closeTrailer() { showTrailer = false; }

  async function load() {
    loading = true;
    error = null;
    // Fire metadata records request immediately so it runs in parallel,
    // but don't block the render on it — getMovieDetail already embeds
    // the primary metadata record, so we can show the page right away.
    const metaPromise = getMetadataRecords('movie', id).catch(() => ({ best: null, records: [] as typeof altRecords }));
    try {
      const detailResp = await getMovieDetail(id);
      detail = detailResp;
      // Use the metadata embedded in the detail response immediately.
      // metaRecords will override this once it resolves (usually within a few
      // hundred ms), giving us the full multi-provider record set.
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

  // ─── Formatting helpers (match the iOS DetailScreen style) ────────────────
  function formatResolution(w?: number, h?: number): string {
    if (!h) return '';
    // Round to common bucket names. Treat anything within ±50px as the bucket.
    if (h >= 2100) return '4K';
    if (h >= 1000) return '1080p';
    if (h >= 700)  return '720p';
    if (h >= 400)  return '480p';
    return `${w}×${h}`;
  }
  function formatBitrate(bps?: number): string {
    if (!bps || bps <= 0) return '';
    if (bps >= 1_000_000) return `${(bps / 1_000_000).toFixed(1)} Mbps`;
    if (bps >= 1_000) return `${Math.round(bps / 1_000)} kbps`;
    return `${bps} bps`;
  }
  function formatChannels(n?: number): string {
    if (!n || n <= 0) return '';
    if (n === 1) return 'Mono';
    if (n === 2) return 'Stereo';
    if (n === 6) return '5.1';
    if (n === 8) return '7.1';
    return `${n}ch`;
  }
  function formatCodec(codec?: string): string {
    if (!codec) return '';
    // FFmpeg names → display names. Keep recognisable; uppercase short codes.
    const map: Record<string, string> = {
      h264: 'H.264', hevc: 'HEVC', av1: 'AV1', vp9: 'VP9', mpeg4: 'MPEG-4',
      aac: 'AAC', ac3: 'AC3', eac3: 'E-AC3', dts: 'DTS', truehd: 'TrueHD',
      flac: 'FLAC', mp3: 'MP3', opus: 'Opus', vorbis: 'Vorbis', alac: 'ALAC',
      pgs: 'PGS', srt: 'SRT', subrip: 'SRT', webvtt: 'WebVTT', vtt: 'WebVTT',
      ass: 'ASS', ssa: 'SSA', mov_text: 'MOV Text',
      hdmv_pgs_subtitle: 'PGS', dvd_subtitle: 'VobSub', dvb_subtitle: 'DVB',
    };
    const lower = codec.toLowerCase();
    return map[lower] ?? codec.toUpperCase();
  }
  function formatLanguage(code?: string): string {
    if (!code) return '';
    const c = code.toLowerCase();
    const map: Record<string, string> = {
      en: 'English', eng: 'English',
      es: 'Spanish', spa: 'Spanish',
      fr: 'French',  fre: 'French', fra: 'French',
      de: 'German',  ger: 'German', deu: 'German',
      it: 'Italian', ita: 'Italian',
      ja: 'Japanese', jpn: 'Japanese',
      ko: 'Korean',  kor: 'Korean',
      zh: 'Chinese', chi: 'Chinese', zho: 'Chinese',
      pt: 'Portuguese', por: 'Portuguese',
      ru: 'Russian', rus: 'Russian',
      hi: 'Hindi',   hin: 'Hindi',
      ar: 'Arabic',  ara: 'Arabic',
      nl: 'Dutch',   dut: 'Dutch', nld: 'Dutch',
      sv: 'Swedish', swe: 'Swedish',
      no: 'Norwegian', nor: 'Norwegian',
      da: 'Danish',  dan: 'Danish',
      fi: 'Finnish', fin: 'Finnish',
      pl: 'Polish',  pol: 'Polish',
      tr: 'Turkish', tur: 'Turkish',
      he: 'Hebrew',  heb: 'Hebrew',
      und: 'Unknown',
    };
    return map[c] ?? code.toUpperCase();
  }
  function formatFileSize(bytes?: number): string {
    if (!bytes) return '';
    if (bytes >= 1_073_741_824) return `${(bytes / 1_073_741_824).toFixed(1)} GB`;
    if (bytes >= 1_048_576) return `${Math.round(bytes / 1_048_576)} MB`;
    return `${bytes} B`;
  }

  // Audio summary like the iOS displayAudioSummary: "AC3 5.1" or "AAC Stereo".
  // Uses the FIRST audio track, since this is just the headline summary; the
  // full per-track breakdown shows up in the Audio section below.
  function audioSummary(tracks: ProbeTrack[]): string {
    const first = tracks?.[0];
    if (!first) return '';
    const codec = formatCodec(first.codec);
    const channels = formatChannels(first.channels);
    return [codec, channels].filter(Boolean).join(' ');
  }
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
          <!-- Metadata strip: year · genres · runtime · rating · content-rating -->
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
            {#if contentRating}
              <span class="opacity-30">·</span>
              <span class="inline-flex items-center gap-1 rounded border border-muted-foreground/40 px-1.5 py-0.5 text-[9px] font-semibold tracking-wider text-muted-foreground">
                <Shield class="h-2.5 w-2.5" />{contentRating}
              </span>
            {/if}
            {#if qualityLabel}
              <span class="opacity-30">·</span>
              <span class="rounded bg-primary/20 px-1.5 py-0.5 text-[9px] font-bold tracking-wider text-primary-glow">{qualityLabel}</span>
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

          <!-- Credits + studio strip -->
          <div class="mt-6 flex flex-wrap gap-x-8 gap-y-3 text-xs text-muted-foreground">
            {#if directors.length > 0}
              <div class="flex flex-wrap items-baseline gap-x-1.5 gap-y-1">
                <span class="uppercase tracking-[0.2em] text-[10px]">{directors.length === 1 ? 'Director' : 'Directors'}</span>
                {#each directors as d (d)}
                  <a href={`/people/${encodeURIComponent(d)}`} class="text-foreground/80 hover:text-foreground hover:underline">{d}</a>
                {/each}
              </div>
            {/if}
            {#if writers.length > 0}
              <div class="flex flex-wrap items-baseline gap-x-1.5 gap-y-1">
                <span class="uppercase tracking-[0.2em] text-[10px]">{writers.length === 1 ? 'Writer' : 'Writers'}</span>
                {#each writers.slice(0, 3) as w (w)}
                  <a href={`/people/${encodeURIComponent(w)}`} class="text-foreground/80 hover:text-foreground hover:underline">{w}</a>
                {/each}
              </div>
            {/if}
            {#if versionCount > 0}
              <div class="flex items-baseline gap-x-1.5">
                <span class="uppercase tracking-[0.2em] text-[10px]">Versions</span>
                <span class="text-foreground/80">{versionCount}</span>
              </div>
            {/if}
          </div>

          <!-- Studio chips -->
          {#if studioNames.length > 0}
            <div class="mt-4 flex flex-wrap gap-2">
              {#each studioNames as studio (studio)}
                <span class="hairline rounded-full bg-surface/40 px-3 py-1 text-[11px] text-muted-foreground">
                  {studio}
                </span>
              {/each}
            </div>
          {/if}

          <!-- ── Versions (only when multiple files exist for this movie) ─── -->
          {#if versions.length > 1}
            <div class="mt-10 border-t border-border pt-8">
              <h3 class="font-serif-display text-lg tracking-tight text-foreground/90">Versions</h3>
              <p class="mt-1 text-[12.5px] text-muted-foreground">
                Pick which file plays. The file info and tracks below update to match your choice.
              </p>
              <div class="scrollbar-none mt-4 -mx-1 flex gap-3 overflow-x-auto px-1 pb-2">
                {#each versions as v, idx (v.mediaSourceId ?? idx)}
                  {@const isSelected = idx === selectedVersionIdx}
                  <button type="button" onclick={() => { selectedVersionIdx = idx; }}
                    class={`shrink-0 rounded-xl border p-4 text-left transition-all w-[220px] ${isSelected ? 'border-primary/60 bg-primary-glow/[0.08] ring-1 ring-primary/40' : 'border-border bg-surface/30 hover:bg-surface/50 hover:border-border/80'}`}>
                    <div class="flex items-baseline justify-between gap-2">
                      <span class="font-serif-display text-[15px] tracking-tight text-foreground/90 truncate">
                        {v.qualityLabel || v.edition || `Version ${idx + 1}`}
                      </span>
                      {#if isSelected}<Check class="h-4 w-4 shrink-0 text-primary-glow" />{/if}
                    </div>
                    {#if v.edition && v.qualityLabel}
                      <div class="mt-1 text-[11px] text-muted-foreground/80 truncate">{v.edition}</div>
                    {/if}
                    {#if v.sizeBytes}
                      <div class="mt-2 text-[12px] text-foreground/55 tabular-nums">{formatFileSize(v.sizeBytes)}</div>
                    {/if}
                    {#if v.relPath}
                      <div class="mt-1 truncate font-mono text-[10px] text-muted-foreground/55" title={v.relPath}>{v.relPath}</div>
                    {/if}
                  </button>
                {/each}
              </div>
            </div>
          {/if}

          <!-- ── File Info pills (codec, resolution, bitrate, audio, container) -->
          {#if mediaSource || audioTracks.length > 0}
            {#if true}
            {@const pills = [
              { icon: FileVideo,  text: formatResolution(mediaSource?.width, mediaSource?.height) },
              { icon: Film,       text: formatCodec(mediaSource?.videoCodec) },
              { icon: Volume2,    text: audioSummary(audioTracks) },
              { icon: Layers,     text: mediaSource?.container ? mediaSource.container.split(',')[0].toUpperCase() : '' },
              { icon: Gauge,      text: formatBitrate(mediaSource?.bitrate) },
              { icon: Captions,   text: subtitleTracks.length > 0 ? `${subtitleTracks.length} subtitle${subtitleTracks.length === 1 ? '' : 's'}` : '' },
            ].filter(p => p.text)}
            {#if pills.length > 0}
              <div class="mt-10 border-t border-border pt-8">
                <h3 class="font-serif-display text-lg tracking-tight text-foreground/90">File Info</h3>
                <div class="mt-4 flex flex-wrap gap-2">
                  {#each pills as p, i (i)}
                    <span class="hairline inline-flex items-center gap-1.5 rounded-full bg-surface/40 px-3 py-1.5 text-[12px] text-foreground/75">
                      <p.icon class="h-3.5 w-3.5 text-muted-foreground/70" />
                      {p.text}
                    </span>
                  {/each}
                </div>
              </div>
            {/if}
            {/if}
          {/if}

          <!-- ── Audio & Subtitles (per-track list) ────────────────────────── -->
          {#if audioTracks.length > 0 || subtitleTracks.length > 0}
            <div class="mt-10 border-t border-border pt-8">
              <h3 class="font-serif-display text-lg tracking-tight text-foreground/90">Audio &amp; Subtitles</h3>
              <p class="mt-1 text-[12.5px] text-muted-foreground">
                Pick a track when you press Play. Default-flagged tracks are used automatically.
              </p>
              <div class="mt-4 grid gap-4 md:grid-cols-2">
                <!-- Audio tracks -->
                <div class="rounded-xl border border-border bg-surface/30 p-4">
                  <div class="mb-3 flex items-center gap-2 text-[11px] uppercase tracking-[0.18em] text-muted-foreground">
                    <Volume2 class="h-3.5 w-3.5" /> Audio · {audioTracks.length}
                  </div>
                  {#if audioTracks.length === 0}
                    <p class="text-[12px] italic text-muted-foreground/60">No audio tracks detected.</p>
                  {:else}
                    <ul class="space-y-1.5">
                      {#each audioTracks as t, i (t.index ?? i)}
                        <li class="flex items-baseline gap-2 rounded-lg px-2 py-1.5 transition-colors hover:bg-foreground/[0.025]">
                          <span class="w-5 text-center text-[10px] tabular-nums text-muted-foreground/55">{(t.index ?? i + 1)}</span>
                          <div class="flex-1 min-w-0">
                            <div class="flex flex-wrap items-baseline gap-x-2 gap-y-0.5">
                              <span class="text-[13px] font-medium text-foreground/90">{formatCodec(t.codec) || '—'}</span>
                              {#if t.channels}<span class="text-[12px] text-foreground/65">{formatChannels(t.channels)}</span>{/if}
                              {#if t.language}<span class="text-[12px] italic text-muted-foreground">· {formatLanguage(t.language)}</span>{/if}
                              {#if t.title}<span class="truncate text-[11.5px] text-muted-foreground/70">· {t.title}</span>{/if}
                            </div>
                            <div class="mt-0.5 flex gap-1.5">
                              {#if t.default}<span class="rounded bg-primary-glow/15 px-1.5 py-0.5 text-[9px] font-medium uppercase tracking-wide text-primary-glow">Default</span>{/if}
                              {#if t.forced}<span class="rounded bg-amber-400/15 px-1.5 py-0.5 text-[9px] font-medium uppercase tracking-wide text-amber-300">Forced</span>{/if}
                            </div>
                          </div>
                        </li>
                      {/each}
                    </ul>
                  {/if}
                </div>

                <!-- Subtitle tracks -->
                <div class="rounded-xl border border-border bg-surface/30 p-4">
                  <div class="mb-3 flex items-center gap-2 text-[11px] uppercase tracking-[0.18em] text-muted-foreground">
                    <Captions class="h-3.5 w-3.5" /> Subtitles · {subtitleTracks.length}
                  </div>
                  {#if subtitleTracks.length === 0}
                    <p class="text-[12px] italic text-muted-foreground/60">No subtitle tracks. Drop an SRT next to the file to add one.</p>
                  {:else}
                    <ul class="space-y-1.5">
                      {#each subtitleTracks as t, i (t.index ?? i)}
                        <li class="flex items-baseline gap-2 rounded-lg px-2 py-1.5 transition-colors hover:bg-foreground/[0.025]">
                          <span class="w-5 text-center text-[10px] tabular-nums text-muted-foreground/55">{(t.index ?? i + 1)}</span>
                          <div class="flex-1 min-w-0">
                            <div class="flex flex-wrap items-baseline gap-x-2 gap-y-0.5">
                              <span class="text-[13px] font-medium text-foreground/90">{formatLanguage(t.language) || formatCodec(t.codec) || '—'}</span>
                              {#if t.language && t.codec}<span class="text-[11.5px] text-muted-foreground/65">· {formatCodec(t.codec)}</span>{/if}
                              {#if t.title}<span class="truncate text-[11.5px] text-muted-foreground/70">· {t.title}</span>{/if}
                            </div>
                            <div class="mt-0.5 flex gap-1.5">
                              {#if t.default}<span class="rounded bg-primary-glow/15 px-1.5 py-0.5 text-[9px] font-medium uppercase tracking-wide text-primary-glow">Default</span>{/if}
                              {#if t.forced}<span class="rounded bg-amber-400/15 px-1.5 py-0.5 text-[9px] font-medium uppercase tracking-wide text-amber-300">Forced</span>{/if}
                            </div>
                          </div>
                        </li>
                      {/each}
                    </ul>
                  {/if}
                </div>
              </div>
              {#if tracksLoading}
                <p class="mt-3 text-[11px] italic text-muted-foreground/60">Loading tracks…</p>
              {/if}
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

            <!-- ── Provider records ──────────────────────────────────────────── -->
            {#if altRecords.length > 0}
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
                        <Film class="h-4 w-4" />
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
