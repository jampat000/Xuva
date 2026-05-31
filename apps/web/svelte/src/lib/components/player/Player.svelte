<script lang="ts">
  import { onMount, onDestroy, tick } from 'svelte';
  import Hls from 'hls.js';
  import {
    Play, Pause, Volume2, VolumeX, Maximize, Minimize,
    Subtitles, Mic2, ChevronLeft, AlertTriangle
  } from 'lucide-svelte';
  // RouteBadge + SkipBack/Forward + Settings icons removed in the v2 player
  // simplification — see the trail of comments below for context.
  import TrackMenu from './TrackMenu.svelte';
  import { type QualityOption } from './QualityMenu.svelte';

  // Local copy of the quality ladder (mirrors QualityMenu.QUALITY_OPTIONS)
  const QUALITY_OPTIONS: QualityOption[] = [
    { id: 'auto', label: 'Auto', sublabel: 'Adapts to your connection' },
    { id: 'original', label: 'Original', sublabel: 'Direct stream if possible' },
    { id: '2160p', label: '4K Ultra HD', sublabel: '2160p · ~40 Mbps' },
    { id: '1440p', label: '2K QHD', sublabel: '1440p · ~16 Mbps' },
    { id: '1080p', label: 'Full HD', sublabel: '1080p · ~8 Mbps' },
    { id: '720p', label: 'HD', sublabel: '720p · ~4 Mbps' },
    { id: '480p', label: 'Standard', sublabel: '480p · ~1.5 Mbps' },
  ];
  // Inspector + KeyboardShortcuts overlay removed in the player simplification
  // pass (PR for issue: "player is too complicated"). Codec/bitrate diagnostics
  // moved to the movie/TV detail page.
  import type {
    PlaybackRouteResponse, PlaybackDecisionResponse,
    MediaSourceItem, ProbeTrack, PlaybackStateResponse
  } from '$lib/api/details';
  import {
    getPlaybackRoute, getMediaSourceTracks, getStreamToken,
    heartbeatClientPlayback, stopClientPlayback, setPlaybackState
  } from '$lib/api/details';
  import { getChapters, type ChaptersResponse, type UserPreferences } from '$lib/api/operator';
  import { getAuthSession } from '$lib/api/auth';
  import { invalidateHomeCache } from '$lib/api/home';
  import { fmt, parseTimestampVTT, thumbForTime as thumbForTimeHelper, type ThumbnailCue, type ChapterCue } from './helpers.js';

  // ─── Props ───────────────────────────────────────────────────────────────
  interface Props {
    mediaSourceId: string;
    title?: string;
    initialRoute: PlaybackRouteResponse;
    initialState?: PlaybackStateResponse;
    mediaSource?: MediaSourceItem;
    clientSessionId?: string;
    deviceId?: string;
    defaultSubtitlesEnabled?: boolean;
    backHref?: string;
  }

  let {
    mediaSourceId,
    title = '',
    initialRoute,
    initialState,
    mediaSource,
    clientSessionId,
    deviceId = 'web',
    defaultSubtitlesEnabled = false,
    backHref = '/'
  }: Props = $props();

  // ─── Video element ────────────────────────────────────────────────────────
  let videoEl = $state<HTMLVideoElement | undefined>(undefined);
  let containerEl = $state<HTMLDivElement | undefined>(undefined);
  let hls: Hls | null = null;

  // Resume seek for the native-HLS / direct-stream paths. Setting currentTime
  // before the media is seekable silently no-ops, so we stash the target and
  // apply it once `loadedmetadata` fires. (The hls.js path uses Hls.startPosition
  // instead — see loadSource.)
  let pendingSeek: number | null = null;

  // ─── Mid-stream stream-token refresh ───────────────────────────────────────
  // Stream URLs carry a short-lived signed token (?sessionId&deviceId&token).
  // On a long title the token can expire mid-playback, after which segment / byte-
  // range requests start failing with 401/403. We detect that, fetch a fresh
  // token, splice it into the URL and reload at the current position so playback
  // resumes seamlessly. Guarded against tight loops when the failure isn't
  // actually auth-related.
  let tokenRefreshing = false;
  let tokenRefreshAttempts = 0;
  const MAX_TOKEN_REFRESHES = 3;

  // ─── Playback state ───────────────────────────────────────────────────────
  let currentTime = $state(0);
  let duration = $state(0);
  let paused = $state(true);
  let buffered = $state(0);
  let volume = $state(1);
  let muted = $state(false);
  let fullscreen = $state(false);
  let loading = $state(true);
  let error = $state<string | null>(null);
  let seeking = $state(false);
  let hasStarted = $state(false);

  // ─── Route / decision ─────────────────────────────────────────────────────
  // svelte-ignore state_referenced_locally — intentional one-time snapshot of the prop
  let route = $state<PlaybackRouteResponse>({ ...initialRoute });
  let decision = $derived(route.decision);
  let isAdaptive = $derived(route.protocol === 'hls');

  // ─── Tracks ───────────────────────────────────────────────────────────────
  let audioTracks = $state<ProbeTrack[]>([]);
  let subtitleTracks = $state<ProbeTrack[]>([]);
  let activeAudioIndex = $state<number | null>(null);
  let activeSubtitleIndex = $state<number | null>(null);

  // ─── UI state ─────────────────────────────────────────────────────────────
  let controlsVisible = $state(true);
  let hideTimer: ReturnType<typeof setTimeout> | null = null;
  let showAudioMenu = $state(false);
  let showSubMenu = $state(false);
  let activeQualityId = $state('auto');
  let resumeToast = $state<string | null>(null);
  let seekToast = $state<string | null>(null);
  let seekToastTimer: ReturnType<typeof setTimeout> | null = null;
  let doubleTapLeft = $state(false);
  let doubleTapRight = $state(false);
  let doubleTapTimer: ReturnType<typeof setTimeout> | null = null;

  // ─── Chapter markers (intro / credits) ───────────────────────────────────
  let chaptersData = $state<ChaptersResponse | null>(null);
  let userPrefs = $state<UserPreferences>({});
  let skipIntroDismissed = $state(false);

  const showSkipIntro = $derived(
    !!chaptersData?.intro &&
    !skipIntroDismissed &&
    currentTime >= (chaptersData.intro.start) &&
    currentTime <= (chaptersData.intro.end)
  );

  const showCreditsMarker = $derived(
    !!chaptersData?.credits &&
    currentTime >= (chaptersData.credits.start)
  );

  function skipIntro() {
    if (!videoEl || !chaptersData?.intro) return;
    videoEl.currentTime = chaptersData.intro.end;
    skipIntroDismissed = true;
  }

  // ─── Session lifecycle ────────────────────────────────────────────────────
  let heartbeatInterval: ReturnType<typeof setInterval> | null = null;

  // ─── Derived helpers ──────────────────────────────────────────────────────
  const progressPercent = $derived(duration > 0 ? (currentTime / duration) * 100 : 0);
  const bufferedPercent = $derived(duration > 0 ? (buffered / duration) * 100 : 0);

  // fmt, parseTimestampVTT, thumbForTime imported from helpers.ts

  function fmtVolume(v: number): string {
    return `${Math.round(v * 100)}%`;
  }

  // ─── Controls visibility ──────────────────────────────────────────────────
  function showControls() {
    controlsVisible = true;
    if (hideTimer) clearTimeout(hideTimer);
    if (!paused) {
      hideTimer = setTimeout(() => {
        if (!showAudioMenu && !showSubMenu) {
          controlsVisible = false;
        }
      }, 2500);
    }
  }

  function keepControlsVisible() {
    if (hideTimer) clearTimeout(hideTimer);
    controlsVisible = true;
  }

  // ─── HLS + video setup ────────────────────────────────────────────────────
  async function loadSource(r: PlaybackRouteResponse, resumePosition?: number) {
    if (!videoEl) return;

    // Tear down existing HLS
    if (hls) {
      hls.destroy();
      hls = null;
    }

    error = null;
    loading = true;

    const url = r.manifestUrl || r.url || '';
    const protocol = r.protocol ?? 'http';

    const resumeAt = resumePosition && resumePosition > 5 ? resumePosition : null;

    if (protocol === 'hls' && url.includes('.m3u8')) {
      if (Hls.isSupported()) {
        hls = new Hls({
          // Aggressive start — load media segments immediately
          startLevel: -1, // auto
          // Seek to the resume point natively. hls.js loads the segment covering
          // startPosition first, so the seek can't be lost to a not-yet-seekable
          // range the way setting videoEl.currentTime at MANIFEST_PARSED could.
          startPosition: resumeAt ?? -1,
          abrEwmaDefaultEstimate: 5_000_000, // assume 5 Mbps until we know better
          maxBufferLength: 60,
          maxMaxBufferLength: 120,
          enableWorker: true,
          lowLatencyMode: false,
          backBufferLength: 30,
        });
        hls.loadSource(url);
        hls.attachMedia(videoEl);
        hls.on(Hls.Events.MANIFEST_PARSED, () => {
          loading = false;
          videoEl!.play().catch(() => {});
        });
        hls.on(Hls.Events.ERROR, (_, data) => {
          if (!data.fatal) return;
          if (data.type === Hls.ErrorTypes.NETWORK_ERROR) {
            // A 401/403 means the stream token expired — refresh it and reload
            // rather than retrying the same dead URL forever.
            const code = data.response?.code;
            if (code === 401 || code === 403) {
              refreshTokenAndReload();
            } else {
              hls?.startLoad();
            }
          } else if (data.type === Hls.ErrorTypes.MEDIA_ERROR) {
            hls?.recoverMediaError();
          } else {
            error = 'Stream error. Check server logs.';
            loading = false;
          }
        });
      } else if (videoEl.canPlayType('application/vnd.apple.mpegurl')) {
        // Safari native HLS — seek after metadata (see pendingSeek).
        pendingSeek = resumeAt;
        videoEl.src = url;
        videoEl.load();
        videoEl.play().catch(() => {});
        loading = false;
      }
    } else {
      // Direct play or remux — plain HTTP stream. Seek after metadata.
      pendingSeek = resumeAt;
      videoEl.src = url;
      videoEl.load();
      videoEl.play().catch(() => {});
      loading = false;
    }
  }

  // Splice a fresh `?sessionId&deviceId&token` query onto a stream URL, dropping
  // any stale auth params first so we don't end up with two token= values.
  function applyStreamTokenQuery(rawUrl: string, query: string): string {
    if (!rawUrl) return rawUrl;
    const cleanQuery = query.replace(/^\?/, '');
    // Strip existing sessionId / deviceId / token / expires params.
    const stripped = rawUrl
      .replace(/([?&])(sessionId|deviceId|token|expires|expiresAt)=[^&]*/gi, '$1')
      .replace(/[?&]+$/g, '')
      .replace(/&{2,}/g, '&')
      .replace(/\?&/, '?');
    const sep = stripped.includes('?') ? '&' : '?';
    return `${stripped}${sep}${cleanQuery}`;
  }

  // Fetch a new stream token and reload the source at the current position.
  async function refreshTokenAndReload() {
    if (tokenRefreshing) return;
    if (!clientSessionId) { error = 'Stream authorization expired. Reload the page to continue.'; loading = false; return; }
    if (tokenRefreshAttempts >= MAX_TOKEN_REFRESHES) {
      error = 'Stream authorization expired. Reload the page to continue.';
      loading = false;
      return;
    }
    tokenRefreshing = true;
    tokenRefreshAttempts += 1;
    const resumeAt = videoEl ? videoEl.currentTime : 0;
    try {
      const tok = await getStreamToken(mediaSourceId, clientSessionId, deviceId);
      if (!tok.query) { error = 'Stream authorization expired. Reload the page to continue.'; loading = false; return; }
      const next: PlaybackRouteResponse = { ...route };
      if (next.manifestUrl) next.manifestUrl = applyStreamTokenQuery(next.manifestUrl, tok.query);
      if (next.url) next.url = applyStreamTokenQuery(next.url, tok.query);
      route = next;
      await loadSource(next, resumeAt);
    } catch {
      error = 'Stream authorization expired. Reload the page to continue.';
      loading = false;
    } finally {
      tokenRefreshing = false;
    }
  }

  // ─── Track switching (without full restart) ───────────────────────────────
  async function switchAudioTrack(index: number | null) {
    if (!videoEl) return;
    const captured = videoEl.currentTime;
    showAudioMenu = false;

    const newRoute = await getPlaybackRoute(mediaSourceId, {
      clientProfile: 'web',
      audioTrackIndex: index ?? 0,
      subtitleTrackIndex: activeSubtitleIndex ?? undefined,
      subtitleTrackActive: activeSubtitleIndex !== null,
      supportsAdaptive: true,
    }).catch(() => null);

    if (!newRoute) return;
    route = newRoute;
    activeAudioIndex = index;
    await loadSource(newRoute, captured);
  }

  async function switchSubtitleTrack(index: number | null) {
    if (!videoEl) return;
    const captured = videoEl.currentTime;
    showSubMenu = false;

    const newRoute = await getPlaybackRoute(mediaSourceId, {
      clientProfile: 'web',
      audioTrackIndex: activeAudioIndex ?? 0,
      subtitleTrackIndex: index ?? undefined,
      subtitleTrackActive: index !== null,
      supportsAdaptive: true,
    }).catch(() => null);

    if (!newRoute) return;
    route = newRoute;
    activeSubtitleIndex = index;
    await loadSource(newRoute, captured);
  }

  // switchQuality is kept (unused by the UI now that the picker is removed)
  // because it's invoked indirectly by the adaptive HLS layer to reflect the
  // currently-selected variant. Removing it would break the auto badge.
  async function switchQuality(qualityId: string) {
    if (!videoEl || !isAdaptive) return;
    const captured = videoEl.currentTime;
    activeQualityId = qualityId;

    // Map quality ID to max bitrate
    const bitrateMap: Record<string, number> = {
      'auto': 0,
      'original': 0,
      '2160p': 40_000_000,
      '1440p': 16_000_000,
      '1080p': 8_000_000,
      '720p': 4_000_000,
      '480p': 1_500_000,
    };

    if (hls && qualityId === 'auto') {
      hls.currentLevel = -1;
      return;
    }

    const maxBitrate = bitrateMap[qualityId] ?? 0;
    const newRoute = await getPlaybackRoute(mediaSourceId, {
      clientProfile: 'web',
      audioTrackIndex: activeAudioIndex ?? 0,
      subtitleTrackIndex: activeSubtitleIndex ?? undefined,
      subtitleTrackActive: activeSubtitleIndex !== null,
      supportsAdaptive: qualityId === 'auto' || qualityId === 'original',
      maxNetworkBitrate: maxBitrate > 0 ? maxBitrate : undefined,
      routeType: qualityId === 'original' ? 'direct' : undefined,
    }).catch(() => null);

    if (!newRoute) return;
    route = newRoute;
    await loadSource(newRoute, captured);
  }

  // ─── Video event handlers ─────────────────────────────────────────────────
  function onTimeUpdate() {
    if (!videoEl) return;
    currentTime = videoEl.currentTime;

    // Update buffered amount
    if (videoEl.buffered.length > 0) {
      buffered = videoEl.buffered.end(videoEl.buffered.length - 1);
    }

    // Reset skip-dismissed flag when scrubbing back before intro
    if (chaptersData?.intro && currentTime < chaptersData.intro.start) {
      skipIntroDismissed = false;
    }

    // Auto-skip intro if preference is set
    if (userPrefs.autoSkipIntros && chaptersData?.intro && !skipIntroDismissed) {
      if (currentTime >= chaptersData.intro.start && currentTime < chaptersData.intro.end) {
        videoEl.currentTime = chaptersData.intro.end;
        skipIntroDismissed = true;
      }
    }
  }

  function onLoadedMetadata() {
    if (!videoEl) return;
    duration = videoEl.duration;
    loading = false;
    // Apply a deferred resume seek now that the media is seekable (native HLS /
    // direct streams — the hls.js path resumes via Hls.startPosition instead).
    if (pendingSeek != null) {
      const target = pendingSeek;
      pendingSeek = null;
      try { videoEl.currentTime = target; } catch { /* not seekable yet — leave at 0 */ }
    }
  }

  function onPlay() {
    paused = false;
    hasStarted = true;
    showControls();
  }

  function onPause() {
    paused = true;
    keepControlsVisible();
  }

  // ─── Remux seek restart ───────────────────────────────────────────────────
  // For live-pipe fMP4 streams (route === "remux") the server pipes data from
  // the beginning of the file. The browser cannot seek beyond what has been
  // buffered. When a seek lands beyond the buffered range the video stalls;
  // we detect this in onWaiting and restart the stream at the requested
  // position by setting a new src with ?startTime=<seconds>.
  let remuxRestartTimer: ReturnType<typeof setTimeout> | null = null;

  function maybeRestartRemuxAtSeekPosition() {
    if (!videoEl || route.route !== 'remux' || !route.url) return;
    const target = videoEl.currentTime;
    // Check if target is beyond the buffered end
    let bufferedEnd = 0;
    for (let i = 0; i < videoEl.buffered.length; i++) {
      if (videoEl.buffered.start(i) <= target + 0.5) {
        bufferedEnd = Math.max(bufferedEnd, videoEl.buffered.end(i));
      }
    }
    if (target <= bufferedEnd + 1) return; // within buffer — no restart needed

    // Restart the stream at the seek target by replacing the startTime param
    const baseUrl = route.url.replace(/([?&])startTime=[^&]*/g, '').replace(/&$|[?]$/, '');
    const sep = baseUrl.includes('?') ? '&' : '?';
    const newUrl = `${baseUrl}${sep}startTime=${target.toFixed(3)}`;
    videoEl.src = newUrl;
    videoEl.currentTime = 0; // server seeks to startTime — client starts from 0
    videoEl.load();
    videoEl.play().catch(() => {});
  }

  function onWaiting() {
    seeking = true;
    // Debounce — only attempt a remux restart if stalled for more than 400 ms
    if (route.route === 'remux') {
      if (remuxRestartTimer) clearTimeout(remuxRestartTimer);
      remuxRestartTimer = setTimeout(() => {
        remuxRestartTimer = null;
        maybeRestartRemuxAtSeekPosition();
      }, 400);
    }
  }

  function onCanPlay() {
    seeking = false;
    loading = false;
    // Playback is healthy again — reset the token-refresh budget so a later,
    // independent expiry later in the title gets its own refresh allowance.
    tokenRefreshAttempts = 0;
  }

  function onEnded() {
    paused = true;
    keepControlsVisible();
    // Write final state
    if (clientSessionId) {
      stopClientPlayback(clientSessionId, { positionSeconds: Math.floor(duration), completed: true }).catch(() => {});
    }
    setPlaybackState(mediaSourceId, {
      progressSeconds: Math.floor(duration),
      durationSeconds: Math.floor(duration),
      watched: true,
    }).catch(() => {});
    // Bust home cache so continue-watching refreshes on next visit
    invalidateHomeCache();
  }

  function onVolumeChange() {
    if (!videoEl) return;
    volume = videoEl.volume;
    muted = videoEl.muted;
  }

  function onError() {
    // For the direct / native-HLS paths a <video> error is the only signal we
    // get when the stream token expires (the element can't surface the HTTP
    // status). If we have a session and a token in the URL, try a token refresh
    // before giving up — but only when hls.js isn't driving the element (its own
    // ERROR handler covers that path with the real HTTP code).
    if (!hls && clientSessionId && tokenRefreshAttempts < MAX_TOKEN_REFRESHES &&
        ((route.url && route.url.includes('token=')) || (route.manifestUrl && route.manifestUrl.includes('token=')))) {
      refreshTokenAndReload();
      return;
    }
    error = 'Could not load media. The file may be unavailable or the format is unsupported.';
    loading = false;
  }

  // ─── Controls ─────────────────────────────────────────────────────────────
  function togglePlay() {
    if (!videoEl) return;
    if (videoEl.paused) {
      videoEl.play().catch(() => {});
    } else {
      videoEl.pause();
    }
  }

  function skip(seconds: number) {
    if (!videoEl) return;
    videoEl.currentTime = Math.max(0, Math.min(duration, videoEl.currentTime + seconds));
    showSeekToast(seconds);
    showControls();
  }

  function showSeekToast(seconds: number) {
    seekToast = seconds > 0 ? `+${seconds}s` : `${seconds}s`;
    if (seekToastTimer) clearTimeout(seekToastTimer);
    seekToastTimer = setTimeout(() => { seekToast = null; }, 800);
  }

  function setVolume(v: number) {
    if (!videoEl) return;
    videoEl.volume = Math.max(0, Math.min(1, v));
    videoEl.muted = false;
  }

  function toggleMute() {
    if (!videoEl) return;
    videoEl.muted = !videoEl.muted;
  }

  function toggleFullscreen() {
    if (!containerEl) return;
    if (!document.fullscreenElement) {
      containerEl.requestFullscreen().catch(() => {});
    } else {
      document.exitFullscreen().catch(() => {});
    }
  }

  function onFullscreenChange() {
    fullscreen = !!document.fullscreenElement;
  }

  // ─── Thumbnails + chapters ────────────────────────────────────────────────
  // ThumbnailCue, ChapterCue interfaces imported from helpers.ts

  let thumbCues = $state<ThumbnailCue[]>([]);
  let chapterCues = $state<ChapterCue[]>([]);
  let spriteUrl = $state('');

  async function loadThumbnailVTT() {
    try {
      const res = await fetch(`/api/media-sources/${mediaSourceId}/thumbnails/thumbnails.vtt`);
      if (!res.ok) return;
      const text = await res.text();
      const cues: ThumbnailCue[] = [];
      const lines = text.split('\n');
      for (let i = 0; i < lines.length; i++) {
        const line = lines[i].trim();
        if (!line.includes('-->')) continue;
        const [startStr, endStr] = line.split('-->').map(s => s.trim());
        const next = (lines[i + 1] ?? '').trim();
        const match = next.match(/#xywh=(\d+),(\d+),(\d+),(\d+)/);
        if (match) {
          cues.push({
            start: parseTimestampVTT(startStr),
            end: parseTimestampVTT(endStr),
            x: parseInt(match[1]), y: parseInt(match[2]),
            w: parseInt(match[3]), h: parseInt(match[4]),
          });
        }
      }
      if (cues.length > 0) {
        thumbCues = cues;
        spriteUrl = `/api/media-sources/${mediaSourceId}/thumbnails/sprite.jpg`;
      }
    } catch { /* thumbnail VTT unavailable — scrubber works without it */ }
  }

  async function loadChaptersVTT() {
    try {
      const res = await fetch(`/api/media-sources/${mediaSourceId}/thumbnails/chapters.vtt`);
      if (!res.ok) return;
      const text = await res.text();
      const cues: ChapterCue[] = [];
      const lines = text.split('\n');
      for (let i = 0; i < lines.length; i++) {
        const line = lines[i].trim();
        if (!line.includes('-->')) continue;
        const [startStr, endStr] = line.split('-->').map(s => s.trim());
        const title = (lines[i + 1] ?? '').trim();
        if (title && !title.includes('-->')) {
          cues.push({
            start: parseTimestampVTT(startStr),
            end: parseTimestampVTT(endStr),
            title,
          });
        }
      }
      chapterCues = cues;
    } catch { /* chapters unavailable */ }
  }

  function thumbForTime(t: number): ThumbnailCue | null {
    return thumbForTimeHelper(t, thumbCues);
  }

  // ─── Scrubber hover preview ───────────────────────────────────────────────
  let hoverPercent = $state<number | null>(null);
  let hoverTime = $state(0);

  // ─── Seek bar ─────────────────────────────────────────────────────────────
  let seekBarEl = $state<HTMLDivElement | undefined>(undefined);
  let isScrubbing = $state(false);

  function seekTo(clientX: number) {
    if (!seekBarEl || !videoEl || !duration) return;
    const rect = seekBarEl.getBoundingClientRect();
    const ratio = Math.max(0, Math.min(1, (clientX - rect.left) / rect.width));
    videoEl.currentTime = ratio * duration;
  }

  function onSeekBarPointerDown(e: PointerEvent) {
    isScrubbing = true;
    seekBarEl?.setPointerCapture(e.pointerId);
    seekTo(e.clientX);
  }

  function onSeekBarPointerMove(e: PointerEvent) {
    // Always track hover position (for thumbnail preview)
    if (seekBarEl && duration) {
      const rect = seekBarEl.getBoundingClientRect();
      const ratio = Math.max(0, Math.min(1, (e.clientX - rect.left) / rect.width));
      hoverPercent = ratio * 100;
      hoverTime = ratio * duration;
    }
    if (!isScrubbing) return;
    seekTo(e.clientX);
    showControls();
  }

  function onSeekBarPointerLeave() {
    if (!isScrubbing) hoverPercent = null;
  }

  function onSeekBarPointerUp(e: PointerEvent) {
    isScrubbing = false;
    hoverPercent = null;
    seekBarEl?.releasePointerCapture(e.pointerId);
  }

  // ─── Keyboard shortcuts ───────────────────────────────────────────────────
  function onKeyDown(e: KeyboardEvent) {
    // Don't intercept when typing in an input
    if ((e.target as HTMLElement)?.tagName === 'INPUT') return;

    switch (e.key) {
      case ' ':
      case 'k':
      case 'K':
        e.preventDefault();
        togglePlay();
        showControls();
        break;
      case 'j':
      case 'J':
        e.preventDefault();
        skip(-10);
        break;
      case 'l':
      case 'L':
        e.preventDefault();
        skip(10);
        break;
      case 'ArrowLeft':
        e.preventDefault();
        skip(e.shiftKey ? -30 : -10);
        break;
      case 'ArrowRight':
        e.preventDefault();
        skip(e.shiftKey ? 30 : 10);
        break;
      case 'ArrowUp':
        e.preventDefault();
        setVolume(volume + 0.1);
        showControls();
        break;
      case 'ArrowDown':
        e.preventDefault();
        setVolume(volume - 0.1);
        showControls();
        break;
      case 'm':
      case 'M':
        e.preventDefault();
        toggleMute();
        showControls();
        break;
      case 'f':
      case 'F':
        e.preventDefault();
        toggleFullscreen();
        break;
      case 'c':
      case 'C':
        // Keep the subtitles shortcut — it's the most-used menu and 'C' is
        // the long-standing convention (YouTube, VLC, native browser players).
        e.preventDefault();
        showSubMenu = !showSubMenu;
        showAudioMenu = false;
        break;
      case 'Escape':
        showAudioMenu = false;
        showSubMenu = false;
        break;
      case '0': case '1': case '2': case '3': case '4':
      case '5': case '6': case '7': case '8': case '9':
        e.preventDefault();
        if (videoEl && duration) {
          videoEl.currentTime = (parseInt(e.key) / 10) * duration;
          showControls();
        }
        break;
    }
  }

  // ─── Mobile double-tap ────────────────────────────────────────────────────
  let lastTapTime = 0;
  let lastTapSide: 'left' | 'right' | null = null;

  function onVideoTap(e: MouseEvent | TouchEvent) {
    if (!containerEl) return;
    const rect = containerEl.getBoundingClientRect();
    const x = 'touches' in e ? e.changedTouches[0].clientX : (e as MouseEvent).clientX;
    const side = x < rect.left + rect.width / 2 ? 'left' : 'right';
    const now = Date.now();

    if (now - lastTapTime < 300 && lastTapSide === side) {
      // Double tap
      skip(side === 'left' ? -10 : 10);
      if (side === 'left') {
        doubleTapLeft = true;
        setTimeout(() => { doubleTapLeft = false; }, 600);
      } else {
        doubleTapRight = true;
        setTimeout(() => { doubleTapRight = false; }, 600);
      }
      lastTapTime = 0;
      lastTapSide = null;
    } else {
      lastTapTime = now;
      lastTapSide = side;
      // Single tap — toggle controls
      showControls();
      if ('ontouchstart' in window) {
        // On touch devices, single tap toggles controls rather than play
      } else {
        togglePlay();
      }
    }
  }

  // ─── Heartbeat ────────────────────────────────────────────────────────────
  function startHeartbeat() {
    if (!clientSessionId) return;
    heartbeatInterval = setInterval(() => {
      if (!videoEl) return;
      heartbeatClientPlayback(clientSessionId!, {
        positionSeconds: Math.floor(videoEl.currentTime),
        isPaused: videoEl.paused,
      }).catch(() => {});
    }, 10_000);
  }

  // ─── Persist progress on unload / backgrounding (sendBeacon) ───────────────
  // `beforeunload` alone is unreliable on mobile — iOS/Android Safari and Chrome
  // frequently skip it, and the OS can kill a backgrounded tab without ever
  // firing it. We additionally listen for `pagehide` (the spec'd replacement,
  // fires on real navigations) and `visibilitychange` → hidden (fires when the
  // tab is backgrounded, which is the last reliable hook before an OS kill).
  //
  // `endSession` distinguishes a genuine teardown (stop the playback session)
  // from a mere backgrounding (just checkpoint progress — the user may return,
  // and the heartbeat session must survive a tab switch).
  function writeProgressBeacon(endSession: boolean) {
    if (!clientSessionId || !videoEl || typeof navigator === 'undefined' || !navigator.sendBeacon) return;
    const pos = Math.floor(videoEl.currentTime);
    const dur = Math.floor(duration);
    // Always checkpoint the resume position.
    navigator.sendBeacon(
      `/api/playback/state/${mediaSourceId}`,
      JSON.stringify({ progressSeconds: pos, durationSeconds: dur })
    );
    // Only end the playback session on a real unload, not a tab switch.
    if (endSession) {
      navigator.sendBeacon(
        `/api/client/playback/${clientSessionId}/stop`,
        JSON.stringify({ positionSeconds: pos, completed: dur > 0 && pos >= dur - 5 })
      );
    }
  }

  function onBeforeUnload() { writeProgressBeacon(true); }
  function onPageHide() { writeProgressBeacon(true); }
  function onVisibilityChange() {
    if (typeof document !== 'undefined' && document.visibilityState === 'hidden') {
      writeProgressBeacon(false);
    }
  }

  // ─── Lifecycle ────────────────────────────────────────────────────────────
  onMount(async () => {
    await tick();
    if (!videoEl) return;

    // Restore resume position
    let resumePos: number | undefined;
    const savedProgress = initialState?.progressSeconds ?? 0;
    const savedDuration = initialState?.durationSeconds ?? 0;
    const savedPercent = savedDuration > 0 ? savedProgress / savedDuration : 0;

    if (savedProgress > 10 && savedPercent < 0.95) {
      resumePos = savedProgress;
      resumeToast = `Resuming from ${fmt(savedProgress)}`;
      setTimeout(() => { resumeToast = null; }, 3000);
    }

    // Fire track fetch and video load in parallel — track metadata is not needed
    // before the video can buffer, and saving ~200ms here is meaningful.
    loadThumbnailVTT();
    loadChaptersVTT();
    getChapters(mediaSourceId).then(ch => { chaptersData = ch; }).catch(() => {});
    getAuthSession().then(s => { if (s?.preferences) userPrefs = s.preferences as UserPreferences; }).catch(() => {});

    const [tracksResult] = await Promise.allSettled([
      getMediaSourceTracks(mediaSourceId),
      loadSource(initialRoute, resumePos),
    ]);

    // Apply track results (the video is already loading in parallel)
    if (tracksResult.status === 'fulfilled' && tracksResult.value != null) {
      const tracks = tracksResult.value;
      audioTracks = tracks.audioTracks ?? [];
      subtitleTracks = tracks.subtitleTracks ?? [];
      if (audioTracks.length > 0) {
        activeAudioIndex = audioTracks.find(t => t.default)?.index ?? audioTracks[0].index ?? 0;
      }
    }

    // Auto-enable subtitles if the user opted in (Settings → Playback). Done
    // after initial load so we don't block first frame. Picks the first
    // non-forced track; user can change in the subtitle menu.
    if (defaultSubtitlesEnabled && activeSubtitleIndex === null && subtitleTracks.length > 0) {
      const preferred = subtitleTracks.find(t => !t.forced) ?? subtitleTracks[0];
      const idx = preferred?.index;
      if (typeof idx === 'number') {
        switchSubtitleTrack(idx).catch(() => {});
      }
    }

    // Start heartbeat
    startHeartbeat();

    // Fullscreen change listener
    document.addEventListener('fullscreenchange', onFullscreenChange);
    window.addEventListener('beforeunload', onBeforeUnload);
    window.addEventListener('pagehide', onPageHide);
    document.addEventListener('visibilitychange', onVisibilityChange);

    // Initial controls show
    showControls();
  });

  onDestroy(() => {
    if (hls) { hls.destroy(); hls = null; }
    if (heartbeatInterval) clearInterval(heartbeatInterval);
    if (hideTimer) clearTimeout(hideTimer);
    if (seekToastTimer) clearTimeout(seekToastTimer);
    if (remuxRestartTimer) clearTimeout(remuxRestartTimer);
    document.removeEventListener('fullscreenchange', onFullscreenChange);
    window.removeEventListener('beforeunload', onBeforeUnload);
    window.removeEventListener('pagehide', onPageHide);
    document.removeEventListener('visibilitychange', onVisibilityChange);

    // Final position write on component destroy (SPA navigation)
    if (videoEl && clientSessionId) {
      const pos = Math.floor(videoEl.currentTime);
      stopClientPlayback(clientSessionId, { positionSeconds: pos }).catch(() => {});
      setPlaybackState(mediaSourceId, {
        progressSeconds: pos,
        durationSeconds: Math.floor(duration),
      }).catch(() => {});
      // Bust home cache so continue-watching refreshes on next visit
      invalidateHomeCache();
    }
  });

  // ─── Build track lists for menus ─────────────────────────────────────────
  function audioTrackOptions() {
    return audioTracks.map(t => ({
      index: t.index ?? 0,
      label: t.title || langLabel(t.language) || `Track ${(t.index ?? 0) + 1}`,
      sublabel: [t.language?.toUpperCase(), t.channels ? `${t.channels}ch` : ''].filter(Boolean).join(' · '),
    }));
  }

  function subtitleTrackOptions() {
    const none = [{ index: -1, label: 'None', isNone: true }];
    const tracks = subtitleTracks.map(t => ({
      index: t.index ?? 0,
      label: t.title || langLabel(t.language) || `Track ${(t.index ?? 0) + 1}`,
      sublabel: [
        t.language?.toUpperCase(),
        t.forced ? 'Forced' : '',
      ].filter(Boolean).join(' · '),
    }));
    return [...none, ...tracks];
  }

  function langLabel(lang: string | undefined): string {
    if (!lang) return '';
    try {
      return new Intl.DisplayNames(['en'], { type: 'language' }).of(lang) ?? lang;
    } catch {
      return lang;
    }
  }

  // Quality options — filter to only show resolutions the source can support
  const qualityOptions = $derived((): QualityOption[] => {
    const sourceH = mediaSource?.height ?? 0;
    return QUALITY_OPTIONS.filter(opt => {
      if (opt.id === 'auto' || opt.id === 'original') return true;
      const h = parseInt(opt.id);
      return !sourceH || h <= sourceH;
    });
  });
</script>

<svelte:window onkeydown={onKeyDown} />

<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<div
  bind:this={containerEl}
  class={`relative h-screen w-screen overflow-hidden bg-black ${controlsVisible ? 'cursor-default' : 'cursor-none'}`}
  onmousemove={showControls}
  onclick={onVideoTap}
  onkeydown={onKeyDown}
  role="application"
  aria-label="Video player"
  tabindex="-1"
>
  <!-- ─── VIDEO ─────────────────────────────────────────────────────────── -->
  <video
    bind:this={videoEl}
    class="h-full w-full object-contain"
    preload="auto"
    playsinline
    ontimeupdate={onTimeUpdate}
    onloadedmetadata={onLoadedMetadata}
    onplay={onPlay}
    onpause={onPause}
    onwaiting={onWaiting}
    oncanplay={onCanPlay}
    onended={onEnded}
    onvolumechange={onVolumeChange}
    onerror={onError}
  ></video>

  <!-- ─── LOADING SPINNER ───────────────────────────────────────────────── -->
  {#if loading || seeking}
    <div class="pointer-events-none absolute inset-0 flex items-center justify-center">
      <div class="h-12 w-12 animate-spin rounded-full border-2 border-white/20 border-t-white/80"></div>
    </div>
  {/if}

  <!-- ─── ERROR ─────────────────────────────────────────────────────────── -->
  {#if error}
    <div class="absolute inset-0 flex flex-col items-center justify-center gap-4 bg-black/80 p-8 text-center">
      <div class="flex h-14 w-14 items-center justify-center rounded-2xl bg-white/10 text-white/60">
        <AlertTriangle class="h-6 w-6" />
      </div>
      <p class="max-w-md text-sm text-white/70 leading-relaxed">{error}</p>
      <div class="flex flex-wrap items-center justify-center gap-2">
        <a href={backHref} class="rounded-full bg-white/10 px-6 py-2.5 text-sm text-white transition-colors hover:bg-white/20">
          ← Back
        </a>
      </div>
    </div>
  {/if}

  <!-- ─── DOUBLE-TAP RIPPLE ─────────────────────────────────────────────── -->
  <div
    class={`pointer-events-none absolute inset-y-0 left-0 w-1/2 flex items-center justify-center transition-opacity duration-200 ${doubleTapLeft ? 'opacity-100' : 'opacity-0'}`}
  >
    <div class="flex h-20 w-20 items-center justify-center rounded-full bg-white/20">
      <span class="text-white text-2xl font-bold">-10</span>
    </div>
  </div>
  <div
    class={`pointer-events-none absolute inset-y-0 right-0 w-1/2 flex items-center justify-center transition-opacity duration-200 ${doubleTapRight ? 'opacity-100' : 'opacity-0'}`}
  >
    <div class="flex h-20 w-20 items-center justify-center rounded-full bg-white/20">
      <span class="text-white text-2xl font-bold">+10</span>
    </div>
  </div>

  <!-- ─── SEEK TOAST ────────────────────────────────────────────────────── -->
  {#if seekToast}
    <div
      class="pointer-events-none absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 rounded-xl bg-black/70 px-5 py-3 text-2xl font-bold text-white backdrop-blur-sm"
    >
      {seekToast}
    </div>
  {/if}

  <!-- ─── RESUME TOAST ───────────────────────────────────────────────────── -->
  {#if resumeToast}
    <div
      class="pointer-events-none absolute bottom-28 left-1/2 -translate-x-1/2 rounded-full bg-black/70 px-5 py-2.5 text-sm text-white/90 backdrop-blur-sm"
    >
      {resumeToast}
    </div>
  {/if}

  <!-- ─── SKIP INTRO BUTTON ─────────────────────────────────────────────── -->
  {#if showSkipIntro}
    <div class="absolute bottom-28 right-6">
      <button
        type="button"
        onclick={skipIntro}
        class="rounded-md border border-white/50 bg-black/60 px-5 py-2.5 text-sm font-semibold text-white backdrop-blur-sm transition-colors hover:bg-white hover:text-black"
      >
        Skip Intro
      </button>
    </div>
  {/if}

  <!-- ─── CREDITS MARKER ────────────────────────────────────────────────── -->
  {#if showCreditsMarker}
    <div class="absolute bottom-28 right-6">
      <div class="rounded-md border border-white/30 bg-black/60 px-5 py-2.5 text-sm text-white/80 backdrop-blur-sm">
        Credits
      </div>
    </div>
  {/if}

  <!-- ─── CONTROLS OVERLAY ─────────────────────────────────────────────── -->
  <div
    class={`absolute inset-0 flex flex-col justify-between transition-opacity duration-300 ${controlsVisible ? 'opacity-100' : 'opacity-0 pointer-events-none'}`}
    role="toolbar"
    aria-label="Player controls"
    tabindex="0"
    onclick={(e) => e.stopPropagation()}
    onkeydown={(e) => e.stopPropagation()}
  >
    <!-- Top bar: back + title + route badge -->
    <div
      class="flex items-center gap-3 px-4 pt-4 pb-8 md:px-6 md:pt-5"
      style="background: linear-gradient(to bottom, rgba(0,0,0,0.75) 0%, transparent 100%);"
    >
      <a
        href={backHref}
        class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-white/70 transition-colors hover:bg-white/10 hover:text-white"
        aria-label="Back"
        onclick={() => {
          // Stop session before navigating
          if (clientSessionId && videoEl) {
            stopClientPlayback(clientSessionId, { positionSeconds: Math.floor(videoEl.currentTime) }).catch(() => {});
          }
        }}
      >
        <ChevronLeft class="h-[18px] w-[18px]" />
      </a>

      {#if title}
        <span class="flex-1 truncate text-[13px] font-medium text-white/85 md:text-sm">{title}</span>
      {/if}

      <!-- RouteBadge + DV/HDR pills removed from the top chrome — they were
           tech-overlay noise on a player. The same info lives on the detail
           page now (File Info pills). -->
    </div>

    <!-- Spacer -->
    <div class="flex-1"></div>

    <!-- Bottom controls -->
    <div
      class="px-4 pb-4 md:px-6 md:pb-5 space-y-3"
      style="background: linear-gradient(to top, rgba(0,0,0,0.85) 0%, transparent 100%);"
    >
      <!-- Seek bar -->
      <div class="flex items-center gap-3">
        <span class="min-w-[3rem] text-right text-[11px] tabular-nums text-white/55">{fmt(currentTime)}</span>

        <div class="relative flex-1">
          <!-- Thumbnail preview tooltip (shown on hover) -->
          {#if hoverPercent !== null && thumbCues.length > 0}
            {@const cue = thumbForTime(hoverTime)}
            {#if cue}
              <div
                class="pointer-events-none absolute bottom-[calc(100%+12px)] z-50 overflow-hidden rounded-lg shadow-2xl ring-1 ring-white/10"
                style="left: clamp(80px, {hoverPercent}%, calc(100% - 80px)); transform: translateX(-50%); width: {cue.w}px; height: {cue.h}px;"
              >
                <img
                  src={spriteUrl}
                  alt=""
                  style="position:absolute; top: -{cue.y}px; left: -{cue.x}px; width: auto; height: auto; image-rendering: pixelated;"
                />
                <div class="absolute bottom-0 inset-x-0 bg-black/50 py-0.5 text-center text-[10px] font-mono text-white/90">
                  {fmt(hoverTime)}
                </div>
              </div>
            {/if}
          {:else if hoverPercent !== null}
            <!-- Time-only tooltip when no thumbnails available -->
            <div
              class="pointer-events-none absolute bottom-[calc(100%+10px)] z-50 rounded-md bg-black/70 px-2 py-1 text-[11px] font-mono text-white backdrop-blur-sm"
              style="left: clamp(20px, {hoverPercent}%, calc(100% - 20px)); transform: translateX(-50%);"
            >
              {fmt(hoverTime)}
            </div>
          {/if}

          <div
            bind:this={seekBarEl}
            class="group relative h-1 w-full cursor-pointer rounded-full bg-white/15 hover:h-1.5 transition-all duration-150"
            onpointerdown={onSeekBarPointerDown}
            onpointermove={onSeekBarPointerMove}
            onpointerup={onSeekBarPointerUp}
            onpointerleave={onSeekBarPointerLeave}
            role="slider"
            aria-label="Seek"
            aria-valuemin={0}
            aria-valuemax={duration}
            aria-valuenow={currentTime}
            aria-valuetext={fmt(currentTime)}
            tabindex="0"
            onkeydown={(e) => {
              if (e.key === 'ArrowLeft') { e.stopPropagation(); e.preventDefault(); skip(-5); }
              if (e.key === 'ArrowRight') { e.stopPropagation(); e.preventDefault(); skip(5); }
            }}
          >
            <!-- Buffer -->
            <div
              class="absolute inset-y-0 left-0 rounded-full bg-white/25"
              style="width: {bufferedPercent}%"
            ></div>
            <!-- Progress -->
            <div
              class="absolute inset-y-0 left-0 rounded-full bg-white"
              style="width: {progressPercent}%"
            ></div>
            <!-- Chapter markers -->
            {#each chapterCues as ch (ch.start)}
              {#if ch.start > 0 && duration > 0}
                <div
                  class="absolute top-1/2 -translate-y-1/2 w-[3px] rounded-sm bg-white/50 h-[150%] opacity-60"
                  style="left: {(ch.start / duration) * 100}%;"
                  title={ch.title}
                ></div>
              {/if}
            {/each}
            <!-- Hover indicator -->
            {#if hoverPercent !== null}
              <div
                class="pointer-events-none absolute inset-y-0 left-0 rounded-full bg-white/40"
                style="width: {hoverPercent}%"
              ></div>
            {/if}
            <!-- Thumb -->
            <div
              class="absolute top-1/2 -translate-y-1/2 -translate-x-1/2 h-4 w-4 rounded-full bg-white shadow-md opacity-0 group-hover:opacity-100 transition-opacity"
              style="left: {progressPercent}%"
            ></div>
          </div>
        </div>

        <span class="min-w-[3rem] font-mono text-xs text-white/70">{fmt(duration)}</span>
      </div>

      <!-- Button row -->
      <div class="flex items-center gap-0.5">
        <!-- Play / pause -->
        <button
          onclick={togglePlay}
          class="flex h-8 w-8 items-center justify-center rounded-full text-white transition-colors hover:bg-white/10"
          aria-label={paused ? 'Play' : 'Pause'}
        >
          {#if paused}
            <Play class="h-[17px] w-[17px] translate-x-px fill-white" />
          {:else}
            <Pause class="h-[17px] w-[17px] fill-white" />
          {/if}
        </button>

        <!-- Skip back / forward buttons removed in the v2 simplification pass.
             Keyboard ←/→ (and J/L) still skip ±10s; mobile double-tap-left /
             double-tap-right ripples still work. YouTube dropped these on
             desktop years ago; we follow the same logic — they were just
             clutter on the bar. -->

        <!-- Mute toggle (volume slider removed — keyboard ↑↓ adjusts) -->
        <button
          onclick={toggleMute}
          class="flex h-8 w-8 items-center justify-center rounded-full text-white/60 transition-colors hover:bg-white/10 hover:text-white"
          aria-label={muted ? 'Unmute' : 'Mute'}
        >
          {#if muted || volume === 0}
            <VolumeX class="h-[15px] w-[15px]" />
          {:else}
            <Volume2 class="h-[15px] w-[15px]" />
          {/if}
        </button>

        <!-- Spacer -->
        <div class="flex-1"></div>

        <!-- Subtitles -->
        {#if subtitleTracks.length > 0}
          <div class="relative">
            <button
              onclick={(e) => { e.stopPropagation(); showSubMenu = !showSubMenu; showAudioMenu = false; }}
              class={`flex h-8 w-8 items-center justify-center rounded-full transition-colors hover:bg-white/10 ${activeSubtitleIndex !== null ? 'text-white' : 'text-white/60'}`}
              aria-label="Subtitles"
              aria-expanded={showSubMenu}
            >
              <Subtitles class="h-[15px] w-[15px]" />
            </button>
            <TrackMenu
              open={showSubMenu}
              title="Subtitles"
              tracks={subtitleTrackOptions()}
              activeIndex={activeSubtitleIndex}
              onSelect={switchSubtitleTrack}
              onClose={() => { showSubMenu = false; }}
            />
          </div>
        {/if}

        <!-- Audio tracks (only when source has 2+ audio tracks) -->
        {#if audioTracks.length > 1}
          <div class="relative">
            <button
              onclick={(e) => { e.stopPropagation(); showAudioMenu = !showAudioMenu; showSubMenu = false; }}
              class={`flex h-8 w-8 items-center justify-center rounded-full transition-colors hover:bg-white/10 ${showAudioMenu ? 'text-white' : 'text-white/60'}`}
              aria-label="Audio track"
              aria-expanded={showAudioMenu}
            >
              <Mic2 class="h-[15px] w-[15px]" />
            </button>
            <TrackMenu
              open={showAudioMenu}
              title="Audio"
              tracks={audioTrackOptions()}
              activeIndex={activeAudioIndex}
              onSelect={switchAudioTrack}
              onClose={() => { showAudioMenu = false; }}
            />
          </div>
        {/if}

        <!-- Fullscreen -->
        <button
          onclick={toggleFullscreen}
          class="flex h-8 w-8 items-center justify-center rounded-full text-white/60 transition-colors hover:bg-white/10 hover:text-white"
          aria-label={fullscreen ? 'Exit fullscreen' : 'Enter fullscreen'}
        >
          {#if fullscreen}
            <Minimize class="h-[15px] w-[15px]" />
          {:else}
            <Maximize class="h-[15px] w-[15px]" />
          {/if}
        </button>
      </div>
    </div>
  </div>

  <!-- Inspector panel removed — was a developer/debug tool that showed
       playback decision codes, codecs, and bitrates. That info now lives on
       the movie/TV detail page where it belongs (File Info pills + Audio
       and Subtitles cards).

       Keyboard-shortcuts overlay also removed. Remaining shortcuts
       (Space/K, arrows, M, F, C, Esc) are conventional. -->
</div>
