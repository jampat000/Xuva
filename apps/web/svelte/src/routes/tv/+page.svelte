<script lang="ts">
  import { onMount } from 'svelte';
  import { appState } from '$lib/stores/appState.svelte';
  import Header from "$lib/components/Header.svelte";
  import LibraryGrid from "$lib/components/LibraryGrid.svelte";
  import ErrorState from '$lib/components/ErrorState.svelte';
  import { getSeries } from '$lib/api/browse';
  import { seriesToMedia } from '$lib/api/adapters';
  import type { Media } from '$lib/mock-data';

  let items = $state<Media[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);

  async function load() {
    error = null;
    loading = true;
    try {
      const resp = await getSeries();
      items = (resp.series ?? []).map(seriesToMedia);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load TV shows';
    } finally {
      loading = false;
    }
  }

  onMount(load);
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
      actions={[{ label: 'Try again', onClick: load }]}
      diagnosticInfo={error}
    />
  {:else}
    <LibraryGrid
      eyebrow="Your library · TV"
      title="Stories told in seasons."
      tagline="Pick up mid-episode, queue the next season, or fall down a rabbit hole — your full series shelf, beautifully laid out."
      {items}
      {loading}
      kind="TV"
      baseHref="/tv"
    />
  {/if}
</div>
