<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import Player from '$lib/components/player/Player.svelte';
  import {
    getPlaybackRoute,
    getPlaybackState,
    getMediaSourceDetail,
    startClientPlayback,
    type PlaybackRouteResponse,
    type PlaybackStateResponse,
    type MediaSourceItem,
  } from '$lib/api/details';

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
  let loadError = $state<string | null>(null);
  let loading = $state(true);

  onMount(async () => {
    if (!mediaSourceId) {
      loadError = 'No media source specified.';
      loading = false;
      return;
    }

    try {
      // Parallel fetch: route + saved state + file detail
      const [routeResp, stateResp, sourceResp] = await Promise.allSettled([
        getPlaybackRoute(mediaSourceId, {
          clientProfile: 'web',
          supportsAdaptive: true,
        }),
        getPlaybackState(mediaSourceId),
        getMediaSourceDetail(mediaSourceId),
      ]);

      if (routeResp.status === 'fulfilled') {
        route = routeResp.value;
      } else {
        loadError = 'Could not determine a playback route for this file. The server may be unavailable or the file is not playable.';
        loading = false;
        return;
      }

      if (stateResp.status === 'fulfilled') {
        savedState = stateResp.value;
      }

      if (sourceResp.status === 'fulfilled') {
        mediaSource = sourceResp.value;
      }

      // Start a client playback session
      try {
        const session = await startClientPlayback({
          mediaSourceId,
          positionSeconds: savedState?.progressSeconds ?? 0,
          clientProfile: 'web',
        });
        clientSessionId = session.id;
      } catch {
        // Non-fatal — heartbeat and stop will just be no-ops
      }

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
  <!-- Full-screen black loader while fetching route -->
  <div class="flex h-screen w-screen items-center justify-center bg-black">
    <div class="h-10 w-10 animate-spin rounded-full border-2 border-white/20 border-t-white/70"></div>
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
      <p class="mt-2 max-w-sm text-sm leading-relaxed text-white/50">
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
    {backHref}
  />
{/if}
