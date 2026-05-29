<script lang="ts">
  import { onMount } from 'svelte';
  import { appState } from '$lib/stores/appState.svelte';
  import Header from "$lib/components/Header.svelte";
  import LibraryGrid from "$lib/components/LibraryGrid.svelte";
  import ErrorState from '$lib/components/ErrorState.svelte';
  import { getMovies, subscribeMovies } from '$lib/api/browse';
  import { movieToMedia } from '$lib/api/adapters';
  import type { Media } from '$lib/mock-data';

  let { data } = $props();

  // load() in +page.ts returns the full library as an unresolved Promise so
  // SvelteKit mounts the page immediately with a skeleton — see +page.ts
  // for why the old two-stage fetch is gone.
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

    // SWR pushes (background refresh / SSE invalidation → refetch) flow in
    // here so a later scan or metadata update lands without a reload.
    const unsubscribe = subscribeMovies(0, (resp) => {
      if (cancelled) return;
      items = (resp.movies ?? []).map(movieToMedia);
      loading = false;
    });
    return () => { cancelled = true; unsubscribe(); };
  });

  // "Try again" — re-run the fetch and clear errors.
  async function reload() {
    userError = null;
    loading = true;
    try {
      const resp = await getMovies();
      items = (resp.movies ?? []).map(movieToMedia);
    } catch (e) {
      userError = e instanceof Error ? e.message : 'Failed to load movies';
    } finally {
      loading = false;
    }
  }
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
      actions={[{ label: 'Try again', onClick: reload }]}
      diagnosticInfo={error}
    />
  {:else}
    <LibraryGrid
      eyebrow="Your library · Movies"
      title="Movies"
      tagline="Every movie in your collection, organized the way you actually browse — by mood, by genre, by what you almost watched last weekend."
      {items}
      {loading}
      kind="Movies"
      baseHref="/movies"
      showHero={false}
    />
  {/if}
</div>
