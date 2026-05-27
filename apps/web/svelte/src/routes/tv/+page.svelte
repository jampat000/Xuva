<script lang="ts">
  import { onMount } from 'svelte';
  import { appState } from '$lib/stores/appState.svelte';
  import Header from "$lib/components/Header.svelte";
  import LibraryGrid from "$lib/components/LibraryGrid.svelte";
  import ErrorState from '$lib/components/ErrorState.svelte';
  import { getSeries, subscribeSeries } from '$lib/api/browse';
  import { seriesToMedia } from '$lib/api/adapters';
  import type { Media } from '$lib/mock-data';

  let { data } = $props();

  // Same prop-override pattern as /movies — avoids state_referenced_locally
  // warnings while still letting SWR + the reload button mutate.
  let swrItems = $state<Media[] | null>(null);
  let userError = $state<string | null | undefined>(undefined);
  let loading = $state(false);

  const items = $derived(swrItems ?? data.items);
  const error = $derived(userError === undefined ? data.loadError : userError);

  // Background-merge full list and stay subscribed to SWR refreshes — same
  // pattern as /movies. See movies/+page.svelte for rationale.
  onMount(() => {
    if (error) return;
    let cancelled = false;
    if (data.hasMore) {
      void getSeries(undefined, 0).then((resp) => {
        if (cancelled) return;
        const full = (resp.series ?? []).map(seriesToMedia);
        if (full.length > items.length) swrItems = full;
      }).catch(() => { /* keep first page on failure */ });
    }
    const unsubscribe = subscribeSeries(0, (resp) => {
      if (cancelled) return;
      swrItems = (resp.series ?? []).map(seriesToMedia);
    });
    return () => { cancelled = true; unsubscribe(); };
  });

  // Only used by the "Try again" button — re-runs the full fetch
  async function reload() {
    userError = null;
    loading = true;
    try {
      const resp = await getSeries();
      swrItems = (resp.series ?? []).map(seriesToMedia);
    } catch (e) {
      userError = e instanceof Error ? e.message : 'Failed to load TV shows';
    } finally {
      loading = false;
    }
  }
</script>

<svelte:head>
  <title>TV — {appState.serverName}</title>
  <meta
    name="description"
    content="Browse your TV library on Xuva, track what is next, and move through your series shelf season by season."
  />
  <meta property="og:title" content="TV — Xuva" />
  <meta
    property="og:description"
    content="Your TV library, organized for bingeing, catching up, and finding the next episode fast."
  />
</svelte:head>

<div class="min-h-screen bg-background">
  <Header />
  {#if error}
    <ErrorState
      title="Can't reach your library"
      message="Make sure your Xuva server is running, then try again."
      actions={[{ label: 'Try again', onClick: reload }]}
      diagnosticInfo={error}
    />
  {:else}
    <LibraryGrid
      eyebrow="Your library · TV"
      title="TV."
      tagline="Pick up mid-episode, queue the next season, or fall down a rabbit hole — your full series shelf, beautifully laid out."
      {items}
      {loading}
      kind="TV"
      baseHref="/tv"
      showHero={false}
    />
  {/if}
</div>
