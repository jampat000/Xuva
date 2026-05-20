<script lang="ts">
  import { onMount } from 'svelte';
  import { appState } from '$lib/stores/appState.svelte';
  import heroFeatured from "$lib/assets/hero-featured.jpg";
  import CollectionsBento from "$lib/components/CollectionsBento.svelte";
  import ContentRow from "$lib/components/ContentRow.svelte";
  import Header from "$lib/components/Header.svelte";
  import Hero from "$lib/components/Hero.svelte";
  import Logo from "$lib/components/Logo.svelte";
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
  let topRowTitle = $state('Highest rated');
  let topRowEyebrow = $state('From your library');
  let collections = $state<Collection[]>([]);
  let loading = $state(true);

  onMount(async () => {
    try {
      const resp = await getClientHome();
      if (resp.heroes?.length) {
        slides = resp.heroes.map(clientHomeItemToMedia);
      }
      for (const row of resp.rows ?? []) {
        const items = (row.items ?? []).map(clientHomeItemToMedia);
        const id = (row.id ?? '').toLowerCase();
        const t = (row.title ?? '').toLowerCase();
        if (id === 'continue' || t.includes('continue') || t.includes('watching')) {
          continueWatching = items;
        } else if (id === 'movies' || t.includes('movie')) {
          recentMovies = items;
        } else if (id === 'tv' || t.includes('series') || t.includes('show') || t.includes('episode')) {
          recentSeries = items;
        } else if (id === 'top' || t.includes('top') || t.includes('trend') || t.includes('rated')) {
          topTen = items;
          const r = row as Record<string, unknown>;
          if (r.title) topRowTitle = r.title as string;
          if (r.eyebrow) topRowEyebrow = r.eyebrow as string;
        }
        // 'recently-added' is intentionally not rendered on web — movies and
        // series are already shown in their own dedicated rows above.
      }
    } catch {
      // Components render gracefully with empty arrays
    } finally {
      loading = false;
    }
  });
</script>

<svelte:head>
  <title>{appState.serverName} — Your personal media library</title>
  <meta
    name="description"
    content="Xuva is your personal media server for movies and series."
  />
  <meta property="og:title" content="Xuva — Your personal media library" />
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
      <Hero slides={slides} trailersEnabled={appState.trailersEnabled} />
    {/if}

    <div class={`relative z-10 space-y-16 md:space-y-20 ${slides.length > 0 ? 'pt-10 md:pt-14' : 'pt-24 md:pt-28'}`}>
      {#if continueWatching.length > 0}
        <ContentRow
          eyebrow="Pick up where you left off"
          title="Continue watching"
          items={continueWatching}
          variant="wide"
        />
      {/if}

      {#if topTen.length > 0}
        <Top10Row items={topTen} title={topRowTitle} eyebrow={topRowEyebrow} />
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

    <footer class="mt-24 border-t border-border px-6 py-6 md:px-12 lg:px-20">
      <div class="flex flex-col items-start justify-between gap-4 md:flex-row md:items-center">
        <div class="flex items-center gap-5">
          <Logo />
          <span class="text-[11px] uppercase tracking-[0.2em] text-muted-foreground">
            © {currentYear} Xuva
          </span>
        </div>
        <div class="flex flex-wrap gap-x-8 gap-y-2 text-sm text-muted-foreground">
          <a class="transition-colors hover:text-foreground" href="/about">About</a>
          <a class="transition-colors hover:text-foreground" href="/support">Support</a>
          <a class="transition-colors hover:text-foreground" href="/privacy">Privacy</a>
          <a class="transition-colors hover:text-foreground" href="/terms">Terms</a>
        </div>
      </div>
    </footer>
  </main>
</div>
