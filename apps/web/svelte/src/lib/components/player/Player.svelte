<script lang="ts">
  import { onMount, onDestroy, tick } from 'svelte';
  import Hls from 'hls.js';
  import {
    Play, Pause, Volume2, VolumeX, Maximize, Minimize,
    SkipBack, SkipForward, Subtitles, Mic2, Settings, Info, ChevronLeft, Keyboard
  } from 'lucide-svelte';
  import RouteBadge from './RouteBadge.svelte';
  import TrackMenu from './TrackMenu.svelte';
  import QualityMenu, { type QualityOption } from './QualityMenu.svelte';

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
  import Inspector from './Inspector.svelte';
  import type {
    PlaybackRouteResponse, PlaybackDecisionResponse,
    MediaSourceItem, ProbeTrack, PlaybackStateResponse
  } from '$lib/api/details';
  import {
    getPlaybackRoute, getMediaSourceTracks,
    heartbeatClientPlayback, stopClientPlayback, setPlaybackState
  } from '$lib/api/details';

  // ─── Props ───────────────────────────────────────────────────────────────
  interface Props {
    mediaSourceId: string;
    title?: string;
    initialRoute: PlaybackRouteResponse;
    initialState?: PlaybackStateResponse;
    mediaSource?: MediaSourceItem;
    clientSessionId?: string;
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
    defaultSubtitlesEnabled = false,
    backHref = '/'
  }: Props = $props();

  // ─── Video element ────────────────────────────────────────────────────────
  let videoEl = $state<HTMLVideoElement | undefined>(undefined);
  let containerEl = $state<HTMLDivElement | undefined>(undefined);
  let hls: Hls | null = null;

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
  let showInspector = $state(false);
  let showAudioMenu = $state(false);
  let showSubMenu = $state(false);
  let showQualityMenu = $state(false);
  let showShortcuts = $state(false);
  let activeQualityId = $state('auto');
  let resumeToast = $state<string | null>(null);
  let seekToast = $state<string | null>(null);
  let seekToastTimer: ReturnType<typeof setTimeout> | null = null;
  let doubleTapLeft = $state(false);
  let doubleTapRight = $state(false);
  let doubleTapTimer: ReturnType<typeof setTimeout> | null = null;

  // ─── Session lifecycle ────────────────────────────────────────────────────
  let heartbeatInterval: ReturnType<typeof setInterval> | null = null;

  // ─── Derived helpers ──────────────────────────────────────────────────────
  const progressPercent = $derived(duration > 0 ? (currentTime / duration) * 100 : 0);
  const bufferedPercent = $derived(duration > 0 ? (buffered / duration) * 100 : 0);

  function fmt(s: number): string {
    const h = Math.floor(s / 3600);
    const m = Math.floor((s % 3600) / 60);
    const sec = Math.floor(s % 60);
    if (h > 0) return `${h}:${String(m).padStart(2, '0')}:${String(sec).padStart(2, '0')}`;
    return `${m}:${String(sec).padStart(2, '0')}`;
  }

  function fmtVolume(v: number): string {
    return `${Math.round(v * 100)}%`;
  }

  // ─── Controls visibility ──────────────────────────────────────────────────
  function showControls() {
    controlsVisible = true;
    if (hideTimer) clearTimeout(hideTimer);
    if (!paused) {
      hideTimer = setTimeout(() => {
        if (!showAudioMenu && !showSubMenu && !showQualityMenu) {
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

    if (protocol === 'hls' && url.includes('.m3u8')) {
      if (Hls.isSupported()) {
        hls = new Hls({
          // Aggressive start — load media segments immediately
          startLevel: -1, // auto
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
          if (resumePosition && resumePosition > 5) {
            videoEl!.currentTime = resumePosition;
          }
          videoEl!.play().catch(() => {});
        });
        hls.on(Hls.Events.ERROR, (_, data) => {
          if (data.fatal) {
            if (data.type === Hls.ErrorTypes.NETWORK_ERROR) {
              hls?.startLoad();
            } else if (data.type === Hls.ErrorTypes.MEDIA_ERROR) {
              hls?.recoverMediaError();
            } else {
              error = 'Stream error. Check server logs.';
              loading = false;
            }
          }
        });
      } else if (videoEl.canPlayType('application/vnd.apple.mpegurl')) {
        // Safari native HLS
        videoEl.src = url;
        if (resumePosition && resumePosition > 5) {
          videoEl.currentTime = resumePosition;
        }
        videoEl.load();
        videoEl.play().catch(() => {});
        loading = false;
      }
    } else {
      // Direct play or remux — plain HTTP stream
      videoEl.src = url;
      if (resumePosition && resumePosition > 5) {
        // Wait for loadedmetadata before seeking on direct streams
        const seekOnce = () => {
          videoEl!.currentTime = resumePosition;
          videoEl!.removeEventListener('loadedmetadata', seekOnce);
        };
        videoEl.addEventListener('loadedmetadata', seekOnce);
      }
      videoEl.load();
      videoEl.play().catch(() => {});
      loading = false;
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

  async function switchQuality(qualityId: string) {
    if (!videoEl || !isAdaptive) return;
    const captured = videoEl.currentTime;
    showQualityMenu = false;
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
  }

  function onLoadedMetadata() {
    if (!videoEl) return;
    duration = videoEl.duration;
    loading = false;
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

  function onWaiting() {
    seeking = true;
  }

  function onCanPlay() {
    seeking = false;
    loading = false;
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
  }

  function onVolumeChange() {
    if (!videoEl) return;
    volume = videoEl.volume;
    muted = videoEl.muted;
  }

  function onError() {
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
  interface ThumbnailCue { start: number; end: number; x: number; y: number; w: number; h: number; }
  interface ChapterCue   { start: number; end: number; title: string; }

  let thumbCues = $state<ThumbnailCue[]>([]);
  let chapterCues = $state<ChapterCue[]>([]);
  let spriteUrl = $state('');

  function parseTimestampVTT(ts: string): number {
    const parts = ts.trim().split(':');
    if (parts.length === 3) {
      return parseFloat(parts[0]) * 3600 + parseFloat(parts[1]) * 60 + parseFloat(parts[2]);
    }
    if (parts.length === 2) {
      return parseFloat(parts[0]) * 60 + parseFloat(parts[1]);
    }
    return parseFloat(parts[0]);
  }

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
    if (!thumbCues.length) return null;
    return thumbCues.find(c => t >= c.start && t < c.end) ?? thumbCues[thumbCues.length - 1];
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
      case 'i':
      case 'I':
        e.preventDefault();
        showInspector = !showInspector;
        if (showInspector) keepControlsVisible();
        break;
      case 'c':
      case 'C':
        e.preventDefault();
        showSubMenu = !showSubMenu;
        showAudioMenu = false;
        showQualityMenu = false;
        break;
      case 'a':
      case 'A':
        e.preventDefault();
        showAudioMenu = !showAudioMenu;
        showSubMenu = false;
        showQualityMenu = false;
        break;
      case '?':
        e.preventDefault();
        showShortcuts = !showShortcuts;
        if (showShortcuts) keepControlsVisible();
        break;
      case 'Escape':
        if (showShortcuts) { showShortcuts = false; break; }
        showAudioMenu = false;
        showSubMenu = false;
        showQualityMenu = false;
        showInspector = false;
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

  // ─── Stop on unload (sendBeacon) ──────────────────────────────────────────
  function onBeforeUnload() {
    if (!clientSessionId || !videoEl) return;
    const pos = Math.floor(videoEl.currentTime);
    const dur = Math.floor(duration);
    // sendBeacon for guaranteed delivery on tab close
    const payload = JSON.stringify({ positionSeconds: pos, completed: pos >= dur - 5 });
    navigator.sendBeacon(`/api/client/playback/${clientSessionId}/stop`, payload);
    // Also write final playback state
    const statePayload = JSON.stringify({ progressSeconds: pos, durationSeconds: dur });
    navigator.sendBeacon(`/api/playback/state/${mediaSourceId}`, statePayload);
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

    // Fetch tracks
    try {
      const tracks = await getMediaSourceTracks(mediaSourceId);
      audioTracks = tracks.audioTracks ?? [];
      subtitleTracks = tracks.subtitleTracks ?? [];
      if (audioTracks.length > 0) {
        activeAudioIndex = audioTracks.find(t => t.default)?.index ?? audioTracks[0].index ?? 0;
      }
    } catch {
      // Non-fatal — tracks just won't show in menus
    }

    // Load thumbnail VTT and chapters (best-effort, non-blocking)
    loadThumbnailVTT();
    loadChaptersVTT();

    // Load the video
    await loadSource(initialRoute, resumePos);

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

    // Initial controls show
    showControls();
  });

  onDestroy(() => {
    if (hls) { hls.destroy(); hls = null; }
    if (heartbeatInterval) clearInterval(heartbeatInterval);
    if (hideTimer) clearTimeout(hideTimer);
    if (seekToastTimer) clearTimeout(seekToastTimer);
    document.removeEventListener('fullscreenchange', onFullscreenChange);
    window.removeEventListener('beforeunload', onBeforeUnload);

    // Final position write on component destroy (SPA navigation)
    if (videoEl && clientSessionId) {
      const pos = Math.floor(videoEl.currentTime);
      stopClientPlayback(clientSessionId, { positionSeconds: pos }).catch(() => {});
      setPlaybackState(mediaSourceId, {
        progressSeconds: pos,
        durationSeconds: Math.floor(duration),
      }).catch(() => {});
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
      <div class="text-4xl">⚠️</div>
      <p class="max-w-md text-sm text-white/70 leading-relaxed">{error}</p>
      <a href={backHref} class="mt-2 rounded-full bg-white/10 px-6 py-2.5 text-sm text-white transition-colors hover:bg-white/20">
        ← Back
      </a>
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

  <!-- ─── CONTROLS OVERLAY ─────────────────────────────────────────────── -->
  <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
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
        class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-black/40 text-white/80 backdrop-blur-sm transition-colors hover:bg-black/60 hover:text-white"
        aria-label="Back"
        onclick={() => {
          // Stop session before navigating
          if (clientSessionId && videoEl) {
            stopClientPlayback(clientSessionId, { positionSeconds: Math.floor(videoEl.currentTime) }).catch(() => {});
          }
        }}
      >
        <ChevronLeft class="h-5 w-5" />
      </a>

      {#if title}
        <span class="flex-1 truncate text-sm font-semibold text-white/90 md:text-base">{title}</span>
      {/if}

      <RouteBadge {decision} />
      {#if decision?.reasonCode === 'dolby_vision_pass_through'}
        <span class="hidden rounded-md bg-blue-900/60 px-2 py-0.5 text-[9px] font-bold uppercase tracking-[0.12em] text-blue-300 ring-1 ring-blue-500/30 backdrop-blur-sm md:inline">DV</span>
      {:else if decision?.reasonCode === 'hdr_pass_through'}
        <span class="hidden rounded-md bg-amber-900/60 px-2 py-0.5 text-[9px] font-bold uppercase tracking-[0.12em] text-amber-300 ring-1 ring-amber-500/30 backdrop-blur-sm md:inline">HDR</span>
      {/if}
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
        <span class="min-w-[3rem] text-right font-mono text-xs text-white/70">{fmt(currentTime)}</span>

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
            class="group relative h-1.5 w-full cursor-pointer rounded-full bg-white/20 hover:h-2.5 transition-all duration-150"
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
              if (e.key === 'ArrowLeft') skip(-5);
              if (e.key === 'ArrowRight') skip(5);
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
      <div class="flex items-center gap-1 md:gap-2">
        <!-- Skip back -->
        <button
          onclick={() => skip(-10)}
          class="flex h-10 w-10 items-center justify-center rounded-full text-white/80 transition-colors hover:bg-white/10 hover:text-white"
          aria-label="Skip back 10 seconds"
        >
          <SkipBack class="h-5 w-5" />
        </button>

        <!-- Play / pause -->
        <button
          onclick={togglePlay}
          class="flex h-12 w-12 items-center justify-center rounded-full bg-white text-black shadow-lg transition-transform hover:scale-105 active:scale-95"
          aria-label={paused ? 'Play' : 'Pause'}
        >
          {#if paused}
            <Play class="h-5 w-5 translate-x-0.5" />
          {:else}
            <Pause class="h-5 w-5" />
          {/if}
        </button>

        <!-- Skip forward -->
        <button
          onclick={() => skip(10)}
          class="flex h-10 w-10 items-center justify-center rounded-full text-white/80 transition-colors hover:bg-white/10 hover:text-white"
          aria-label="Skip forward 10 seconds"
        >
          <SkipForward class="h-5 w-5" />
        </button>

        <!-- Volume -->
        <div class="group/vol relative ml-1 flex items-center gap-2">
          <button
            onclick={toggleMute}
            class="flex h-10 w-10 items-center justify-center rounded-full text-white/80 transition-colors hover:bg-white/10 hover:text-white"
            aria-label={muted ? 'Unmute' : 'Mute'}
          >
            {#if muted || volume === 0}
              <VolumeX class="h-5 w-5" />
            {:else}
              <Volume2 class="h-5 w-5" />
            {/if}
          </button>
          <!-- Volume slider (appears on hover) -->
          <div class="hidden w-20 md:block">
            <input
              type="range"
              min="0"
              max="1"
              step="0.05"
              value={muted ? 0 : volume}
              oninput={(e) => setVolume(parseFloat((e.target as HTMLInputElement).value))}
              class="h-1 w-full cursor-pointer appearance-none rounded-full bg-white/30 accent-white"
              aria-label="Volume"
            />
          </div>
        </div>

        <!-- Spacer -->
        <div class="flex-1"></div>

        <!-- Subtitles -->
        {#if subtitleTracks.length > 0}
          <div class="relative">
            <button
              onclick={(e) => { e.stopPropagation(); showSubMenu = !showSubMenu; showAudioMenu = false; showQualityMenu = false; }}
              class={`flex h-10 w-10 items-center justify-center rounded-full transition-colors hover:bg-white/10 ${activeSubtitleIndex !== null ? 'text-white' : 'text-white/60'}`}
              aria-label="Subtitles"
              aria-expanded={showSubMenu}
            >
              <Subtitles class="h-5 w-5" />
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

        <!-- Audio tracks -->
        {#if audioTracks.length > 1}
          <div class="relative">
            <button
              onclick={(e) => { e.stopPropagation(); showAudioMenu = !showAudioMenu; showSubMenu = false; showQualityMenu = false; }}
              class={`flex h-10 w-10 items-center justify-center rounded-full transition-colors hover:bg-white/10 ${showAudioMenu ? 'text-white' : 'text-white/60'}`}
              aria-label="Audio track"
              aria-expanded={showAudioMenu}
            >
              <Mic2 class="h-5 w-5" />
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

        <!-- Quality -->
        {#if isAdaptive || (qualityOptions().length > 2)}
          <div class="relative">
            <button
              onclick={(e) => { e.stopPropagation(); showQualityMenu = !showQualityMenu; showAudioMenu = false; showSubMenu = false; }}
              class={`flex h-10 w-10 items-center justify-center rounded-full transition-colors hover:bg-white/10 ${showQualityMenu ? 'text-white' : 'text-white/60'}`}
              aria-label="Quality"
              aria-expanded={showQualityMenu}
            >
              <Settings class="h-5 w-5" />
            </button>
            <QualityMenu
              open={showQualityMenu}
              options={qualityOptions()}
              activeId={activeQualityId}
              onSelect={switchQuality}
              onClose={() => { showQualityMenu = false; }}
            />
          </div>
        {/if}

        <!-- Keyboard shortcuts -->
        <button
          onclick={(e) => { e.stopPropagation(); showShortcuts = !showShortcuts; if (showShortcuts) keepControlsVisible(); }}
          class={`hidden md:flex h-10 w-10 items-center justify-center rounded-full transition-colors hover:bg-white/10 ${showShortcuts ? 'text-white' : 'text-white/60'}`}
          aria-label="Keyboard shortcuts"
          aria-pressed={showShortcuts}
        >
          <Keyboard class="h-5 w-5" />
        </button>

        <!-- Inspector toggle -->
        <button
          onclick={(e) => { e.stopPropagation(); showInspector = !showInspector; }}
          class={`flex h-10 w-10 items-center justify-center rounded-full transition-colors hover:bg-white/10 ${showInspector ? 'text-white' : 'text-white/60'}`}
          aria-label="Toggle inspector"
          aria-pressed={showInspector}
        >
          <Info class="h-5 w-5" />
        </button>

        <!-- Fullscreen -->
        <button
          onclick={toggleFullscreen}
          class="flex h-10 w-10 items-center justify-center rounded-full text-white/80 transition-colors hover:bg-white/10 hover:text-white"
          aria-label={fullscreen ? 'Exit fullscreen' : 'Enter fullscreen'}
        >
          {#if fullscreen}
            <Minimize class="h-5 w-5" />
          {:else}
            <Maximize class="h-5 w-5" />
          {/if}
        </button>
      </div>
    </div>
  </div>

  <!-- ─── INSPECTOR PANEL ───────────────────────────────────────────────── -->
  <Inspector
    open={showInspector}
    {decision}
    {mediaSource}
    positionSeconds={currentTime}
    durationSeconds={duration}
    onClose={() => { showInspector = false; }}
  />

  <!-- ─── KEYBOARD SHORTCUTS PANEL ────────────────────────────────────────── -->
  {#if showShortcuts}
    <!-- Backdrop -->
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div
      class="absolute inset-0 z-40 flex items-center justify-center bg-black/60 backdrop-blur-sm"
      onclick={(e) => { if (e.target === e.currentTarget) showShortcuts = false; }}
    >
      <div
        class="w-full max-w-sm overflow-hidden rounded-2xl border border-white/10 bg-[#0e0e0e]/95 shadow-2xl"
        onclick={(e) => e.stopPropagation()}
        role="dialog"
        aria-label="Keyboard shortcuts"
        tabindex="-1"
      >
        <!-- Header -->
        <div class="flex items-center justify-between border-b border-white/10 px-5 py-4">
          <span class="text-sm font-semibold text-white">Keyboard Shortcuts</span>
          <button
            onclick={() => { showShortcuts = false; }}
            class="flex h-7 w-7 items-center justify-center rounded-full text-white/50 transition-colors hover:bg-white/10 hover:text-white"
            aria-label="Close"
          >✕</button>
        </div>

        <!-- Shortcut rows -->
        <div class="divide-y divide-white/[0.06] px-5 py-2">
          {#each [
            { keys: ['Space', 'K'], label: 'Play / Pause' },
            { keys: ['J'], label: 'Rewind 10 s' },
            { keys: ['L'], label: 'Forward 10 s' },
            { keys: ['←', '→'], label: 'Skip ±10 s' },
            { keys: ['⇧←', '⇧→'], label: 'Skip ±30 s' },
            { keys: ['↑', '↓'], label: 'Volume ±10%' },
            { keys: ['M'], label: 'Mute / Unmute' },
            { keys: ['F'], label: 'Fullscreen' },
            { keys: ['C'], label: 'Subtitles' },
            { keys: ['A'], label: 'Audio track' },
            { keys: ['0–9'], label: 'Jump to 0%–90%' },
            { keys: ['?'], label: 'This panel' },
          ] as row (row.label)}
            <div class="flex items-center justify-between py-2.5">
              <span class="text-sm text-white/70">{row.label}</span>
              <div class="flex items-center gap-1.5">
                {#each row.keys as k (k)}
                  <kbd class="rounded-md border border-white/20 bg-white/10 px-2 py-0.5 font-mono text-[11px] text-white/90">{k}</kbd>
                {/each}
              </div>
            </div>
          {/each}
        </div>

        <div class="border-t border-white/10 px-5 py-3 text-center text-[11px] text-white/30">
          Press <kbd class="rounded border border-white/15 bg-white/10 px-1.5 font-mono text-[10px] text-white/40">Esc</kbd> or click outside to close
        </div>
      </div>
    </div>
  {/if}
</div>
