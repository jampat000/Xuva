<script lang="ts">
  import { onMount } from 'svelte';
  import { appState } from '$lib/stores/appState.svelte';
  import Header from "$lib/components/Header.svelte";
  import LibraryGrid from "$lib/components/LibraryGrid.svelte";
  import ErrorState from '$lib/components/ErrorState.svelte';
  import { getMovies } from '$lib/api/browse';
  import { movieToMedia } from '$lib/api/adapters';
  import type { Media } from '$lib/mock-data';

  let items = $state<Media[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);

  async function load() {
    error = null;
    loading = true;
    try {
      const resp = await getMovies();
      items = (resp.movies ?? []).map(movieToMedia);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load movies';
    } finally {
      loading = false;
    }
  }

  onMount(load);
</script>

<svelte:head>
  <title>Movies — {appState.serverName}</title>
  <meta
    name="description"
    content="Browse your movie library on Xuva with editorial filtering, search, and density controls."
  />
  <meta property="og:title" content="Movies — Xuva" />
  <meta
    property="og:description"
    content="Rediscover your collection by genre, rating, year, and mood — all in one movie wall."
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
      eyebrow="Your library · Movies"
      title="Films worth a night."
      tagline="Every movie in your collection, organized the way you actually browse — by mood, by genre, by what you almost watched last weekend."
      {items}
      {loading}
      kind="Movies"
      baseHref="/movies"
    />
  {/if}
</div>
