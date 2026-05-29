<script lang="ts">
  import { onMount } from 'svelte';
  import "../app.css";
  import { appState } from '$lib/stores/appState.svelte';
  import { profileStore } from '$lib/stores/profileStore.svelte';
  import { syncWatchlistFromServer } from '$lib/stores/watchlistStore.svelte';
  import { listProfiles } from '$lib/api/profiles';
  import NavProgress from '$lib/components/NavProgress.svelte';
  import { connectEventStream } from '$lib/api/events';
  import type { ProfileCard } from '$lib/api/profiles';
  import type { Component } from 'svelte';

  let { children } = $props();

  /** Whether the Who's Watching picker should show. */
  let showPicker = $derived(profileStore.showPicker);

  // Lazy-loaded reference to WhoIsWatching. Previously this was a static import
  // at the top of the file, which dragged the picker's transitive dependencies
  // (lucide icons, profile API client, PIN pad) into the module graph of every
  // page including /signin. If any of those imports had a side-effect that
  // threw during module evaluation, the layout's downstream onMount callbacks
  // — including AuthEntry.svelte's initialize() — would silently never run,
  // leaving /signin stuck on "Checking authentication status…" forever.
  // Loading WhoIsWatching on demand isolates the picker's failure modes from
  // the sign-in flow that's required to even reach the picker. This is the
  // primary fix for #418.
  let WhoIsWatching = $state<Component<{ onselect: (p: ProfileCard, remember: boolean) => void }> | null>(null);

  // Persistence keys for the active profile. sessionStorage = tab-scoped
  // (default — user has to repick on each new tab). localStorage = device-
  // scoped (set when the user ticks "Remember me on this device" in the
  // picker — survives tab close, browser restart, until next logout).
  const PROFILE_SESSION_KEY = 'xuva-profile-card';
  const PROFILE_DEVICE_KEY = 'xuva-profile-card-device';

  onMount(() => {
    if (typeof window === 'undefined') return;
    const path = window.location.pathname;
    if (path.startsWith('/setup') || path.startsWith('/signin')) return;

    // Subscribe to backend events for live cache invalidation. The
    // connection is per-tab and idempotent; cleanup tears it down when the
    // layout unmounts (i.e. on full page navigation away).
    const disconnect = connectEventStream();

    // Restore active profile. Try localStorage first (device-remembered)
    // then sessionStorage (tab-remembered). Synchronous so the picker
    // never flashes on subsequent navigations.
    try {
      const deviceStored = localStorage.getItem(PROFILE_DEVICE_KEY);
      const tabStored = !deviceStored ? sessionStorage.getItem(PROFILE_SESSION_KEY) : null;
      const stored = deviceStored ?? tabStored;
      if (stored) profileStore.setActiveProfile(JSON.parse(stored) as ProfileCard);
    } catch {
      try { localStorage.removeItem(PROFILE_DEVICE_KEY); } catch { /* ignore */ }
      try { sessionStorage.removeItem(PROFILE_SESSION_KEY); } catch { /* ignore */ }
    }

    // Fire bootstrap, setup-check, and profile discovery in parallel without
    // awaiting — first paint is no longer gated on the network. Each handler
    // updates app state when its response lands.
    void fetch('/api/setup/status').then(async (resp) => {
      if (!resp.ok) return;
      const data = await resp.json() as { requiresSetup?: boolean };
      if (data.requiresSetup) window.location.href = '/setup';
    }).catch(() => { /* server unreachable — stay put */ });

    void fetch('/api/client/bootstrap').then(async (resp) => {
      if (!resp.ok) return;
      const data = await resp.json() as { server?: { name?: string }; features?: { trailers?: boolean } };
      if (data.server?.name) appState.serverName = data.server.name;
      if (data.features?.trailers === false) appState.trailersEnabled = false;
      syncWatchlistFromServer();
    }).catch(() => { /* server unreachable — stay put */ });

    // Skip the picker if a profile was already chosen this tab session.
    if (profileStore.activeProfile) return;

    void listProfiles().then((profiles) => {
      if (profiles.length > 0 && !profileStore.activeProfile) profileStore.openPicker();
    }).catch(() => { /* auth disabled or endpoint unavailable */ });

    return () => { disconnect(); };
  });

  // When the picker becomes visible, lazy-load its module. The first display
  // pays a one-time import cost (~50-100ms); subsequent shows reuse the
  // cached module. Failure to load doesn't break the rest of the app — the
  // picker just stays hidden and the user can still navigate manually.
  $effect(() => {
    if (showPicker && !WhoIsWatching) {
      void import('$lib/components/WhoIsWatching.svelte').then(m => {
        WhoIsWatching = m.default as unknown as Component<{ onselect: (p: ProfileCard, remember: boolean) => void }>;
      }).catch(err => {
        if (typeof console !== 'undefined') console.error('[layout] WhoIsWatching lazy-load failed', err);
      });
    }
  });

  function handleProfileSelected(profile: ProfileCard, remember: boolean) {
    profileStore.setActiveProfile(profile);
    // Always write tab-scoped storage so navigation within the tab keeps the
    // pick (matches existing behaviour).
    try { sessionStorage.setItem(PROFILE_SESSION_KEY, JSON.stringify(profile)); } catch { /* ignore */ }
    // Only write device-scoped storage when the user opted in via the
    // "Remember me on this device" checkbox. If they didn't, clear any
    // stale device-stored value from a previous session.
    try {
      if (remember) {
        localStorage.setItem(PROFILE_DEVICE_KEY, JSON.stringify(profile));
      } else {
        localStorage.removeItem(PROFILE_DEVICE_KEY);
      }
    } catch { /* localStorage blocked — fall back to tab-only persistence */ }
  }
</script>

<svelte:head>
  <title>{appState.serverName}</title>
</svelte:head>

<NavProgress />

{#if showPicker && WhoIsWatching}
  <WhoIsWatching onselect={handleProfileSelected} />
{:else}
  {@render children()}
{/if}
