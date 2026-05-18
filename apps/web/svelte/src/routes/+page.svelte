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
  <title>Xuva — Your home cinema, on every screen</title>
  <meta
    name="description"
    content="Xuva is your personal media server for movies and series — stream your collection on every screen in your home."
  />
  <meta property="og:title" content="Xuva — Your home cinema, on every screen" />
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
    {#if slides.length > 0}
      <Hero slides={slides} />
    {/if}

    <div class={`relative z-10 space-y-16 md:space-y-20 ${slides.length > 0 ? '-mt-12 md:-mt-16' : 'pt-24 md:pt-28'}`}>
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
        <div class="relative flex flex-col items-center justify-center px-6 py-32 text-center">
          <div
            aria-hidden="true"
            class="pointer-events-none absolute inset-0 -z-10"
            style="background: radial-gradient(ellipse at 50% 40%, oklch(0.62 0.22 285 / 0.12), transparent 60%), radial-gradient(ellipse at 30% 80%, oklch(0.72 0.16 255 / 0.08), transparent 55%);"
          ></div>
          <div class="hairline mb-6 flex h-16 w-16 items-center justify-center rounded-2xl bg-surface-elevated/70 text-primary-glow shadow-elev">
            <svg class="h-7 w-7" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M3.375 19.5h17.25m-17.25 0a1.125 1.125 0 0 1-1.125-1.125M3.375 19.5h7.5c.621 0 1.125-.504 1.125-1.125m-9.75 0V5.625m0 12.75v-1.5c0-.621.504-1.125 1.125-1.125m18.375 2.625V5.625m0 12.75c0 .621-.504 1.125-1.125 1.125m1.125-1.125v-1.5c0-.621-.504-1.125-1.125-1.125m0 3.75h-7.5A1.125 1.125 0 0 1 12 18.375m9.75-12.75c0-.621-.504-1.125-1.125-1.125H3.375c-.621 0-1.125.504-1.125 1.125m19.5 0v1.5c0 .621-.504 1.125-1.125 1.125M2.25 5.625v1.5c0 .621.504 1.125 1.125 1.125m0 0h17.25m-17.25 0c-.621 0-1.125.504-1.125 1.125v7.5" />
            </svg>
          </div>
          <p class="font-serif-display text-3xl tracking-tight">Your cinema awaits.</p>
          <p class="mt-3 max-w-sm text-sm leading-relaxed text-muted-foreground">
            Point Xuva at your media folders to start streaming your collection on every screen.
          </p>
          <a
            href="/settings"
            class="hairline mt-6 inline-flex items-center gap-2 rounded-full bg-foreground/[0.06] px-6 py-3 text-sm font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.10] hover:text-foreground"
          >
            Open Settings →
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
