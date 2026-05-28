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

  // Non-blocking load — see routes/movies/+page.svelte for rationale.
  let items = $state<Media[]>([]);
  let userError = $state<string | null | undefined>(undefined);
  let loading = $state(true);

  const error = $derived(userError ?? null);

  onMount(() => {
    let cancelled = false;

    void data.itemsPromise.then((full) => {
      if (cancelled) return;
      items = full;
      loading = false;
    });
    void data.loadErrorPromise.then((err) => {
      if (cancelled) return;
      if (err) userError = err;
    });

    // SWR push subscription — fires when SSE invalidation or background
    // refresh produces fresh data, so a later scan / metadata update flows
    // in without a reload.
    const unsubscribe = subscribeSeries(0, (resp) => {
      if (cancelled) return;
      items = (resp.series ?? []).map(seriesToMedia);
      loading = false;
    });
    return () => { cancelled = true; unsubscribe(); };
  });

  async function reload() {
    userError = null;
    loading = true;
    try {
      const resp = await getSeries();
      items = (resp.series ?? []).map(seriesToMedia);
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
