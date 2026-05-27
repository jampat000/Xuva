<script lang="ts">
  import { onMount } from 'svelte';
  import "../app.css";
  import { appState } from '$lib/stores/appState.svelte';
  import { profileStore } from '$lib/stores/profileStore.svelte';
  import { syncWatchlistFromServer } from '$lib/stores/watchlistStore.svelte';
  import { listProfiles } from '$lib/api/profiles';
  import WhoIsWatching from '$lib/components/WhoIsWatching.svelte';
  import type { ProfileCard } from '$lib/api/profiles';

  let { children } = $props();

  /** Whether the Who's Watching picker should show. */
  let showPicker = $derived(profileStore.showPicker);

  // sessionStorage key: persists across hard navigations within the same tab
  // but resets when the tab is closed — correct "who's watching" semantics.
  const PROFILE_SESSION_KEY = 'xuva-profile-card';

  onMount(async () => {
    if (typeof window === 'undefined') return;
    const path = window.location.pathname;
    if (path.startsWith('/setup') || path.startsWith('/signin')) return;

    try {
      const [setupResp, bootstrapResp] = await Promise.all([
        fetch('/api/setup/status'),
        fetch('/api/client/bootstrap'),
      ]);

      if (setupResp.ok) {
        const data = await setupResp.json() as { requiresSetup?: boolean };
        if (data.requiresSetup) { window.location.href = '/setup'; return; }
      }

      if (bootstrapResp.ok) {
        const data = await bootstrapResp.json() as { server?: { name?: string }; features?: { trailers?: boolean } };
        if (data.server?.name) appState.serverName = data.server.name;
        if (data.features?.trailers === false) appState.trailersEnabled = false;
        // Sync watchlist from server after bootstrap succeeds
        syncWatchlistFromServer();
      }
    } catch {
      // Server unreachable — stay on current page.
    }

    // If a profile was already selected this browser tab session, restore it
    // silently without showing the picker again.
    try {
      const stored = sessionStorage.getItem(PROFILE_SESSION_KEY);
      if (stored) {
        const card = JSON.parse(stored) as ProfileCard;
        profileStore.setActiveProfile(card);
        return;
      }
    } catch {
      try { sessionStorage.removeItem(PROFILE_SESSION_KEY); } catch { /* ignore */ }
    }

    // First visit in this tab — show picker if profiles exist and auth is on.
    try {
      const profiles = await listProfiles();
      if (profiles.length > 0) {
        profileStore.openPicker();
      }
    } catch {
      // Profiles endpoint unavailable (auth disabled or single-user) — skip.
    }
  });

  function handleProfileSelected(profile: ProfileCard) {
    profileStore.setActiveProfile(profile);
    try { sessionStorage.setItem(PROFILE_SESSION_KEY, JSON.stringify(profile)); } catch { /* ignore */ }
  }
</script>

<svelte:head>
  <title>{appState.serverName}</title>
</svelte:head>

{#if showPicker}
  <WhoIsWatching onselect={handleProfileSelected} />
{:else}
  {@render children()}
{/if}
