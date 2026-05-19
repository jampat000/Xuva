<script lang="ts">
  import { onMount } from 'svelte';
  import "../app.css";
  import { appState } from '$lib/stores/appState.svelte';

  let { children } = $props();

  onMount(async () => {
    if (typeof window === 'undefined') return;
    // Don't redirect if already on the setup or sign-in pages.
    const path = window.location.pathname;
    if (path.startsWith('/setup') || path.startsWith('/signin')) return;

    // Both endpoints are public (no auth required).
    // /api/setup/status carries the setup flag; /api/client/bootstrap carries the server name.
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
  });
</script>

<svelte:head>
  <title>{appState.serverName}</title>
  <link rel="icon" type="image/svg+xml" href="/favicon.svg" />
  <link rel="icon" type="image/png" sizes="32x32" href="/favicon-32.png" />
  <link rel="icon" type="image/png" sizes="16x16" href="/favicon-16.png" />
  <link rel="apple-touch-icon" href="/apple-touch-icon.png" />
  <meta name="theme-color" content="#8c5cff" />
</svelte:head>

{@render children()}
