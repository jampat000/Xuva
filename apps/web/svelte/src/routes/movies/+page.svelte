<script lang="ts">
  import Header from "$lib/components/Header.svelte";
  import LibraryGrid from "$lib/components/LibraryGrid.svelte";
  import { featured, spotlightSlides, recentMovies, topTen } from "$lib/mock-data";

  const items = [
    featured,
    ...spotlightSlides.filter((slide) => slide.type === "Movie"),
    ...recentMovies,
    ...topTen.filter((item) => item.type === "Movie")
  ].filter((item, index, all) => all.findIndex((current) => current.id === item.id) === index);
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
  <LibraryGrid
    eyebrow="Your library · Movies"
    title="Films worth a night."
    tagline="Every movie in your collection, organized the way you actually browse — by mood, by genre, by what you almost watched last weekend."
    items={items}
    kind="Movies"
  />
</div>
