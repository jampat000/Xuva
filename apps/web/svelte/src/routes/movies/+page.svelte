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

  // Local overrides for prop-derived values. `null` means "fall through to the
  // load() prop"; once set, the override wins. This avoids the Svelte 5
  // state_referenced_locally warning that fires when $state(prop.x) shadows a
  // prop value (the prop would not propagate further changes into the state).
  let swrItems = $state<Media[] | null>(null);
  let userError = $state<string | null | undefined>(undefined);
  let loading = $state(false);

  const items = $derived(swrItems ?? data.items);
  const error = $derived(userError === undefined ? data.loadError : userError);

  // Background-merge the rest of the library (the first-page paint already
  // mounted the page with FIRST_PAGE items) and stay subscribed so a later
  // background SWR refresh — e.g. when a scan adds new movies — flows in
  // without a manual reload.
  onMount(() => {
    if (error) return;
    let cancelled = false;
    if (data.hasMore) {
      void getMovies(undefined, 0).then((resp) => {
        if (cancelled) return;
        const full = (resp.movies ?? []).map(movieToMedia);
        // Only replace if the full list is materially larger.
        if (full.length > items.length) swrItems = full;
      }).catch(() => { /* keep the first page on failure */ });
    }
    const unsubscribe = subscribeMovies(0, (resp) => {
      if (cancelled) return;
      swrItems = (resp.movies ?? []).map(movieToMedia);
    });
    return () => { cancelled = true; unsubscribe(); };
  });

  // Only used by the "Try again" button — re-runs the full fetch
  async function reload() {
    userError = null;
    loading = true;
    try {
      const resp = await getMovies();
      swrItems = (resp.movies ?? []).map(movieToMedia);
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
      title="Movies."
      tagline="Every movie in your collection, organized the way you actually browse — by mood, by genre, by what you almost watched last weekend."
      {items}
      {loading}
      kind="Movies"
      baseHref="/movies"
      showHero={false}
    />
  {/if}
</div>
