<script lang="ts">
  import { onMount } from 'svelte';
  import Header from "$lib/components/Header.svelte";
  import LibraryGrid from "$lib/components/LibraryGrid.svelte";
  import { getMovies } from '$lib/api/browse';
  import { movieToMedia } from '$lib/api/adapters';
  import type { Media } from '$lib/mock-data';

  let items = $state<Media[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);

  onMount(async () => {
    try {
      const resp = await getMovies();
      items = (resp.movies ?? []).map(movieToMedia);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load movies';
    } finally {
      loading = false;
    }
  });
</script>

<svelte:head>
  <title>Movies — Xuva</title>
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
    <div class="flex min-h-[60vh] flex-col items-center justify-center gap-3 px-6 text-center">
      <p class="text-base font-medium text-foreground/80">Can't reach your library</p>
      <p class="max-w-xs text-sm text-muted-foreground">Make sure your Xuva server is running, then try again.</p>
      <button
        onclick={() => { error = null; loading = true; getMovies().then(r => { items = (r.movies ?? []).map(movieToMedia); }).catch(e => { error = e instanceof Error ? e.message : 'Failed'; }).finally(() => { loading = false; }); }}
        class="mt-2 hairline rounded-full bg-foreground/[0.06] px-5 py-2.5 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
      >
        Try again
      </button>
    </div>
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
