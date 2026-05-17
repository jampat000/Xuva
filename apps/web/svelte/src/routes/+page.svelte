<script lang="ts">
  import { onMount } from 'svelte';
  import heroFeatured from "$lib/assets/hero-featured.jpg";
  import CollectionsBento from "$lib/components/CollectionsBento.svelte";
  import ContentRow from "$lib/components/ContentRow.svelte";
  import Header from "$lib/components/Header.svelte";
  import Hero from "$lib/components/Hero.svelte";
  import Logo from "$lib/components/Logo.svelte";
  import MoodSelector from "$lib/components/MoodSelector.svelte";
  import Top10Row from "$lib/components/Top10Row.svelte";
  import { getClientHome } from '$lib/api/home';
  import { clientHomeItemToMedia } from '$lib/api/adapters';
  import type { Media, Collection } from '$lib/mock-data';

  const currentYear = new Date().getFullYear();

  let slides = $state<Media[]>([]);
  let continueWatching = $state<Media[]>([]);
  let recentMovies = $state<Media[]>([]);
  let recentSeries = $state<Media[]>([]);
  let topTen = $state<Media[]>([]);
  let collections = $state<Collection[]>([]);
  let loading = $state(true);

  onMount(async () => {
    try {
      const resp = await getClientHome();
      if (resp.hero) {
        slides = [clientHomeItemToMedia(resp.hero)];
      }
      for (const row of resp.rows ?? []) {
        const items = (row.items ?? []).map(clientHomeItemToMedia);
        const t = (row.title ?? '').toLowerCase();
        if (t.includes('continue') || t.includes('watching')) {
          continueWatching = items;
        } else if (t.includes('movie')) {
          recentMovies = items;
        } else if (t.includes('series') || t.includes('show') || t.includes('episode')) {
          recentSeries = items;
        } else if (t.includes('top')) {
          topTen = items;
        }
      }
    } catch {
      // Components render gracefully with empty arrays
    } finally {
      loading = false;
    }
  });
</script>

<svelte:head>
  <title>Xuva — Your cinema, everywhere</title>
  <meta
    name="description"
    content="Xuva is your personal streaming home for movies and series across every screen you own."
  />
  <meta property="og:title" content="Xuva — Your cinema, everywhere" />
  <meta
    property="og:description"
    content="A cinematic home for your personal library — continue watching, discover what is new, and jump between movies and series."
  />
  <meta property="og:type" content="website" />
  <meta property="og:image" content={heroFeatured} />
</svelte:head>

<div class="min-h-screen bg-background">
  <Header />
  <main class="pb-24">
    <Hero slides={slides.length ? slides : []} />

    <div class="relative z-10 -mt-12 space-y-16 md:-mt-16 md:space-y-20">
      {#if continueWatching.length > 0}
        <ContentRow
          eyebrow="Pick up where you left off"
          title="Continue watching"
          items={continueWatching}
          variant="wide"
        />
      {/if}

      <MoodSelector />

      {#if topTen.length > 0}
        <Top10Row items={topTen} />
      {/if}

      {#if collections.length > 0}
        <CollectionsBento items={collections} />
      {/if}

      {#if recentMovies.length > 0}
        <ContentRow
          eyebrow="Fresh in your library"
          title="New movies"
          items={recentMovies}
        />
      {/if}

      {#if recentSeries.length > 0}
        <ContentRow
          eyebrow="New episodes dropped"
          title="New series"
          items={recentSeries}
        />
      {/if}

      {#if !loading && continueWatching.length === 0 && recentMovies.length === 0 && recentSeries.length === 0}
        <div class="flex flex-col items-center justify-center px-6 py-24 text-center">
          <p class="font-serif-display text-3xl">Your library is ready.</p>
          <p class="mt-3 max-w-sm text-sm text-muted-foreground">
            Add a library in settings to start streaming your collection.
          </p>
          <a
            href="/settings"
            class="hairline mt-6 rounded-full bg-foreground/[0.06] px-6 py-3 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
          >
            Go to settings
          </a>
        </div>
      {/if}
    </div>

    <section class="relative mx-6 mt-28 overflow-hidden rounded-3xl border border-border bg-gradient-to-br from-surface/60 via-surface/30 to-background p-10 backdrop-blur md:mx-12 md:p-16 lg:mx-20">
      <div class="absolute -right-32 -top-32 h-[400px] w-[400px] rounded-full bg-primary/20 blur-[120px]"></div>
      <div class="absolute -bottom-32 -left-32 h-[400px] w-[400px] rounded-full bg-accent/20 blur-[120px]"></div>
      <div class="grain absolute inset-0"></div>
      <div class="relative grid items-center gap-12 md:grid-cols-[1.3fr_1fr]">
        <div>
          <div class="mb-4 text-[10px] font-semibold uppercase tracking-[0.35em] text-primary-glow">
            One library · Every screen
          </div>
          <h2 class="font-serif-display text-4xl leading-[1.05] tracking-tight md:text-6xl">
            Made for the couch, the commute, and everything between.
          </h2>
          <p class="mt-6 max-w-xl text-base leading-relaxed text-muted-foreground md:text-lg">
            Xuva adapts to every screen — a remote-friendly grid on tvOS and Android TV, a thumb-shaped feed on mobile, a sidebar-rich layout on tablet, and this cinematic surface on the web.
          </p>
        </div>
        <div class="grid grid-cols-2 gap-3 md:gap-4">
          {#each [
            { label: "Web", sub: "Cinematic" },
            { label: "Mobile", sub: "iOS · Android" },
            { label: "Tablet", sub: "iPadOS" },
            { label: "TV", sub: "tvOS · Android TV" }
          ] as device (device.label)}
            <div
              class="hairline rounded-2xl bg-background/40 p-5 backdrop-blur-md transition-all hover:-translate-y-1 hover:bg-background/60"
            >
              <div class="text-[10px] uppercase tracking-[0.25em] text-muted-foreground">
                {device.sub}
              </div>
              <div class="font-serif-display mt-2 text-2xl">
                {device.label}
              </div>
            </div>
          {/each}
        </div>
      </div>
    </section>

    <footer class="mt-24 border-t border-border px-6 pt-12 md:px-12 lg:px-20">
      <div class="flex flex-col items-start justify-between gap-6 pb-10 md:flex-row md:items-center">
        <div>
          <Logo />
          <p class="mt-4 max-w-sm text-sm leading-relaxed text-muted-foreground">
            Your personal cinema. Stream your collection on every screen you own.
          </p>
        </div>
        <div class="flex flex-wrap gap-x-10 gap-y-3 text-sm text-muted-foreground">
          <a class="transition-colors hover:text-foreground" href="/about">About</a>
          <a class="transition-colors hover:text-foreground" href="/apps">Apps</a>
          <a class="transition-colors hover:text-foreground" href="/support">Support</a>
          <a class="transition-colors hover:text-foreground" href="/privacy">Privacy</a>
          <a class="transition-colors hover:text-foreground" href="/terms">Terms</a>
        </div>
      </div>
      <div class="border-t border-border py-6 text-xs uppercase tracking-[0.2em] text-muted-foreground">
        © {currentYear} Xuva · Crafted for cinema lovers
      </div>
    </footer>
  </main>
</div>
