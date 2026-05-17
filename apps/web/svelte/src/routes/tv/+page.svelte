<script lang="ts">
  import Header from "$lib/components/Header.svelte";
  import LibraryGrid from "$lib/components/LibraryGrid.svelte";
  import { recentSeries, topTen } from "$lib/mock-data";

  const items = [
    ...recentSeries,
    ...topTen.filter((item) => item.type === "Series")
  ].filter((item, index, all) => all.findIndex((current) => current.id === item.id) === index);
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
  <LibraryGrid
    eyebrow="Your library · TV"
    title="Stories told in seasons."
    tagline="Pick up mid-episode, queue the next season, or fall down a rabbit hole — your full series shelf, beautifully laid out."
    items={items}
    kind="TV"
  />
</div>
