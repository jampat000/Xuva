<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import Player from '$lib/components/player/Player.svelte';
  import {
    getPlaybackRoute,
    getPlaybackState,
    getMediaSourceDetail,
    startClientPlayback,
    getStreamToken,
    type PlaybackRouteResponse,
    type PlaybackStateResponse,
    type MediaSourceItem,
  } from '$lib/api/details';
  import { buildCapabilityReport } from '$lib/api/capabilities';

  // ─── Route param ──────────────────────────────────────────────────────────
  const mediaSourceId = $derived($page.params.mediaSourceId ?? '');

  // ─── URL search params for back navigation + title ────────────────────────
  const backHref = $derived($page.url.searchParams.get('back') ?? '/');
  const titleParam = $derived($page.url.searchParams.get('title') ?? '');

  // ─── Async state ──────────────────────────────────────────────────────────
  let route = $state<PlaybackRouteResponse | null>(null);
  let savedState = $state<PlaybackStateResponse | null>(null);
  let mediaSource = $state<MediaSourceItem | null>(null);
  let clientSessionId = $state<string | undefined>(undefined);
  let defaultSubtitlesEnabled = $state(false);
  let loadError = $state<string | null>(null);
  let loading = $state(true);

  // Phase label shown inside the loading screen so users understand delays.
  // "probing" is only shown when startClientPlayback triggers foreground ffprobe.
  type LoadPhase = 'resolving' | 'probing' | 'authorizing';
  let loadPhase = $state<LoadPhase>('resolving');
  const PHASE_LABELS: Record<LoadPhase, string> = {
    resolving:   'Resolving stream…',
    probing:     'Analysing file — this may take a moment…',
    authorizing: 'Authorising stream…',
  };

  onMount(async () => {
    if (!mediaSourceId) {
      loadError = 'No media source specified.';
      loading = false;
      return;
    }

    try {
      // Detect client capabilities once — synchronous canPlayType calls
      const caps = buildCapabilityReport();

      // ── Phase 1: Parallel — route + saved state + file detail ────────────
      loadPhase = 'resolving';
      const [routeResp, stateResp, sourceResp] = await Promise.allSettled([
        getPlaybackRoute(mediaSourceId, {
          clientProfile: 'web',
          supportsAdaptive: true,
          supportsHdr: caps.supportsHdr,
          maxBitDepth: caps.maxVideoBitDepth,
          videoCodecs: caps.videoCodecs,
          audioCodecs: caps.audioCodecs,
        }),
        getPlaybackState(mediaSourceId),
        getMediaSourceDetail(mediaSourceId),
      ]);

      if (routeResp.status !== 'fulfilled') {
        loadError = 'Could not determine a playback route for this file. The server may be unavailable or the file is not playable.';
        loading = false;
        return;
      }

      const initialRoute = routeResp.value;

      if (stateResp.status === 'fulfilled') savedState = stateResp.value;
      if (sourceResp.status === 'fulfilled') mediaSource = sourceResp.value;

      // ── Policy-block retry with forcePlayable ─────────────────────────────
      // Web browsers can't play AC3/DTS/TrueHD natively — audio conversion is
      // unavoidable regardless of the server's playback policy. When the first
      // route request is blocked, retry with forcePlayable=true to let the
      // server pick the best route it can actually serve. Only show the error
      // screen if the retry is also blocked (or returns nothing useful).
      let finalAttemptRoute = initialRoute;
      if (initialRoute.route === 'blocked' || initialRoute.status === 'blocked_by_policy') {
        try {
          const retried = await getPlaybackRoute(mediaSourceId, {
            clientProfile: 'web',
            supportsAdaptive: true,
            supportsHdr: caps.supportsHdr,
            maxBitDepth: caps.maxVideoBitDepth,
            videoCodecs: caps.videoCodecs,
            audioCodecs: caps.audioCodecs,
            forcePlayable: true,
          });
          finalAttemptRoute = retried;
        } catch {
          // Fall through to error display below with the original route
        }
      }

      // If still blocked (or no URL) after the forcePlayable retry, surface error
      if (finalAttemptRoute.route === 'blocked' || finalAttemptRoute.status === 'blocked_by_policy') {
        const reason = finalAttemptRoute.decision?.reasonText || finalAttemptRoute.decision?.reason || 'Playback is blocked by your server policy.';
        const hints = (finalAttemptRoute.fallbackOptions ?? []).map(f => f.label).filter(Boolean);
        loadError = hints.length
          ? `${reason}\n\nSuggested fixes: ${hints.join(', ')}`
          : reason;
        loading = false;
        return;
      }

      // ── Early-exit for transcode-needed but no ready URL ──────────────────
      // When a transcode job is queuing or in-progress, surface a clear message
      // rather than silently setting <video src=""> and freezing.
      if (!finalAttemptRoute.url && !finalAttemptRoute.manifestUrl) {
        const status = finalAttemptRoute.status ?? 'unknown';
        const mode   = finalAttemptRoute.route  ?? 'transcode';
        if (status === 'queuing' || status === 'transcoding') {
          loadError = `This file needs to be transcoded before it can play (${mode}). Wait a moment and try again, or check Activity in Settings.`;
        } else {
          loadError = `No playback URL was returned (route: ${mode}, status: ${status}). Check the server logs.`;
        }
        loading = false;
        return;
      }

      // ── Phase 2: Create playback session ─────────────────────────────────
      // startClientPlayback may trigger a synchronous foreground ffprobe if
      // this file has never been probed before — that can take up to 45s.
      // We surface a "Analysing file…" message so users know why it's slow.
      loadPhase = 'probing';
      let sessionId: string | undefined;
      let sessionDeviceId: string | undefined;

      try {
        const sessionTimer = setTimeout(() => { loadPhase = 'probing'; }, 300);
        const session = await startClientPlayback({
          mediaSourceId,
          positionSeconds: savedState?.progressSeconds ?? 0,
          clientProfile: 'web',
          deviceId: 'web',
          clientCapabilities: caps,
        });
        clearTimeout(sessionTimer);
        // Server returns "sessionId" and "deviceId" — capture both.
        // The session was created with deviceId:'web' above; we mirror that
        // value here so getStreamToken receives a matching deviceId, avoiding
        // the 403 that occurs when they differ (server validates equality).
        sessionId = session.sessionId;
        sessionDeviceId = session.deviceId ?? 'web';
        clientSessionId = session.sessionId;
        defaultSubtitlesEnabled = Boolean(session.defaultSubtitlesEnabled);
        // The start response embeds the resolved route — use it to avoid a
        // redundant getPlaybackRoute round-trip on the happy path.
        if (session.route?.url || session.route?.manifestUrl) {
          finalAttemptRoute = { ...finalAttemptRoute, ...session.route };
        }
      } catch {
        // Non-fatal if auth is disabled; proceed with the plain URL.
        // If auth IS enabled, the stream will 403 and the player will show an error.
      }

      // ── Phase 3: Get signed stream URL ────────────────────────────────────
      // The native <video> element does NOT send X-Auth-Token headers.
      // authorizeStreamRequest on the server requires ?sessionId=&deviceId=&token=
      // to be present in the URL. We fetch those params here and patch the route
      // before handing it to the Player component.
      loadPhase = 'authorizing';
      let finalRoute = finalAttemptRoute;

      if (sessionId && finalAttemptRoute.url && !finalAttemptRoute.url.includes('token=')) {
        try {
          const tokenResp = await getStreamToken(mediaSourceId, sessionId, sessionDeviceId ?? 'web');
          if (tokenResp.streamUrl) {
            finalRoute = { ...finalAttemptRoute, url: tokenResp.streamUrl };
          }
        } catch {
          // Auth is disabled — plain URL will work as-is. Continue.
        }
      }

      // For adaptive streams, append the query string to the manifest URL too.
      if (sessionId && finalAttemptRoute.manifestUrl && !finalAttemptRoute.manifestUrl.includes('token=')) {
        try {
          const tokenResp = await getStreamToken(mediaSourceId, sessionId, sessionDeviceId ?? 'web');
          if (tokenResp.query) {
            const sep = finalAttemptRoute.manifestUrl.includes('?') ? '&' : '?';
            finalRoute = { ...finalRoute, manifestUrl: finalAttemptRoute.manifestUrl + sep + tokenResp.query.replace(/^\?/, '') };
          }
        } catch {
          // Auth disabled — continue with plain manifest URL.
        }
      }

      route = finalRoute;

    } catch (e) {
      loadError = `Unexpected error: ${(e as Error)?.message ?? e}`;
    } finally {
      loading = false;
    }
  });

  // Derive a display title from URL param or media source name
  const displayTitle = $derived(titleParam || mediaSource?.name || '');
</script>

<svelte:head>
  <title>{displayTitle ? `${displayTitle} — Xuva` : 'Playing — Xuva'}</title>
</svelte:head>

{#if loading}
  <!-- Full-screen black loader with phase label so users see what's happening -->
  <div class="flex h-screen w-screen flex-col items-center justify-center gap-4 bg-black">
    <div class="h-10 w-10 animate-spin rounded-full border-2 border-white/20 border-t-white/70"></div>
    <p class="text-xs tracking-widest text-white/40 uppercase">{PHASE_LABELS[loadPhase]}</p>
  </div>

{:else if loadError || !route}
  <!-- Error state -->
  <div class="flex h-screen w-screen flex-col items-center justify-center gap-6 bg-black p-8 text-center">
    <div class="flex h-16 w-16 items-center justify-center rounded-2xl bg-rose-500/20 text-rose-400">
      <svg class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
        <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
      </svg>
    </div>
    <div>
      <p class="font-serif-display text-2xl text-white">Playback unavailable</p>
      <p class="mt-2 max-w-sm text-sm leading-relaxed text-white/50" style="white-space: pre-line;">
        {loadError ?? 'No playback route could be determined.'}
      </p>
    </div>
    <a
      href={backHref}
      class="rounded-full bg-white/10 px-6 py-2.5 text-sm text-white transition-colors hover:bg-white/20"
    >
      ← Back
    </a>
  </div>

{:else}
  <Player
    {mediaSourceId}
    title={displayTitle}
    initialRoute={route}
    initialState={savedState ?? undefined}
    mediaSource={mediaSource ?? undefined}
    {clientSessionId}
    {defaultSubtitlesEnabled}
    {backHref}
  />
{/if}
