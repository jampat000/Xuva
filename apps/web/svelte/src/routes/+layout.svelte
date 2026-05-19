<script lang="ts">
  import { onMount } from 'svelte';
  import "../app.css";
  import { appState } from '$lib/stores/appState.svelte';
  import { profileStore } from '$lib/stores/profileStore.svelte';
  import { readProfileToken } from '$lib/api/profile-token-store';
  import { listProfiles } from '$lib/api/profiles';
  import WhoIsWatching from '$lib/components/WhoIsWatching.svelte';
  import type { ProfileCard } from '$lib/api/profiles';

  let { children } = $props();

  /** Whether the Who's Watching picker should show. */
  let showPicker = $derived(profileStore.showPicker);

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
        const data = await bootstrapResp.json() as { server?: { name?: string } };
        if (data.server?.name) appState.serverName = data.server.name;
      }
    } catch {
      // Server unreachable — stay on current page.
    }

    // Restore the active profile from the stored token if we have one.
    // If auth is disabled or the session has a single user, profiles may not
    // exist — we silently skip and show content without a profile.
    const token = readProfileToken();
    if (token) {
      try {
        const profiles = await listProfiles();
        // The server validates the token; if valid it will include activeProfile
        // in /api/auth/session. For now re-fetch profiles and trust the first
        // match (the token is validated server-side on every request anyway).
        // We can't decode the token client-side — show the picker instead.
        profileStore.openPicker();
      } catch {
        // Profiles endpoint unavailable (auth disabled) — skip picker.
      }
    } else {
      // No token — show the picker if profiles exist and auth is enabled.
      try {
        const profiles = await listProfiles();
        if (profiles.length > 0) {
          profileStore.openPicker();
        }
      } catch {
        // Profiles unavailable (auth disabled or single-user) — skip.
      }
    }
  });

  function handleProfileSelected(profile: ProfileCard) {
    profileStore.setActiveProfile(profile);
  }
</script>

<svelte:head>
  <title>{appState.serverName}</title>
  <link rel="icon" type="image/svg+xml" href="/favicon.svg" />
  <link rel="icon" type="image/png" sizes="32x32" href="/favicon-32.png" />
  <link rel="icon" type="image/png" sizes="16x16" href="/favicon-16.png" />
  <link rel="apple-touch-icon" href="/apple-touch-icon.png" />
  <meta name="theme-color" content="#8c5cff" />
</svelte:head>

{#if showPicker}
  <WhoIsWatching onselect={handleProfileSelected} />
{:else}
  {@render children()}
{/if}
