<script lang="ts">
  import { onMount } from 'svelte';
  import Header from "$lib/components/Header.svelte";
  import LibraryGrid from "$lib/components/LibraryGrid.svelte";
  import { getSeries } from '$lib/api/browse';
  import { seriesToMedia } from '$lib/api/adapters';
  import type { Media } from '$lib/mock-data';

  let items = $state<Media[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);

  onMount(async () => {
    try {
      const resp = await getSeries();
      items = (resp.series ?? []).map(seriesToMedia);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load TV shows';
    } finally {
      loading = false;
    }
  });
</script>

<svelte:head>
  <title>TV — Xuva</title>
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
    <div class="flex min-h-[60vh] flex-col items-center justify-center gap-3 px-6 text-center">
      <p class="text-base font-medium text-foreground/80">Can't reach your library</p>
      <p class="max-w-xs text-sm text-muted-foreground">Make sure your Xuva server is running, then try again.</p>
      <button
        onclick={() => { error = null; loading = true; getSeries().then(r => { items = (r.series ?? []).map(seriesToMedia); }).catch(e => { error = e instanceof Error ? e.message : 'Failed'; }).finally(() => { loading = false; }); }}
        class="mt-2 hairline rounded-full bg-foreground/[0.06] px-5 py-2.5 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
      >
        Try again
      </button>
    </div>
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
