<script lang="ts">
  import { Bookmark, ChevronLeft } from 'lucide-svelte';
  import Header from '$lib/components/Header.svelte';
</script>

<svelte:head>
  <title>Watchlist — Xuva</title>
  <meta name="description" content="Your personal Xuva watchlist — films and shows you've saved to watch later." />
</svelte:head>

<div class="min-h-screen bg-background">
  <Header />

  <main class="px-6 pb-32 pt-24 md:px-12 md:pt-28 lg:px-20">
    <a href="/" class="mb-8 inline-flex items-center gap-2 text-sm text-muted-foreground transition-colors hover:text-foreground">
      <ChevronLeft class="h-4 w-4" /> Home
    </a>

    <header class="relative mb-12">
      <div
        aria-hidden="true"
        class="pointer-events-none absolute -inset-x-6 -top-10 -z-10 h-[180px] opacity-50 md:-inset-x-12 lg:-inset-x-20"
        style="background: radial-gradient(50% 100% at 15% 0%, oklch(0.62 0.22 285 / 0.20), transparent 70%);"
      ></div>
      <div class="mb-2 text-[10px] font-semibold uppercase tracking-[0.35em] text-primary-glow">Your list</div>
      <h1 class="font-serif-display text-[clamp(2rem,4vw,3.25rem)] leading-[1] tracking-tight">Watchlist</h1>
      <p class="mt-3 max-w-xl text-sm text-muted-foreground">
        Films and shows you've saved to watch later, all in one place.
      </p>
    </header>

    <div class="flex flex-col items-center justify-center py-24 text-center">
      <div class="hairline flex h-14 w-14 items-center justify-center rounded-2xl bg-surface-elevated/70 text-primary-glow shadow-elev">
        <Bookmark class="h-6 w-6" />
      </div>
      <p class="font-serif-display mt-5 text-2xl tracking-tight">Nothing saved yet</p>
      <p class="mt-2 max-w-sm text-sm text-muted-foreground">
        Tap the <span class="font-medium text-foreground/70">+</span> button on any movie or show to add it here.
      </p>
      <a
        href="/movies"
        class="hairline mt-6 inline-flex items-center gap-2 rounded-full bg-foreground/[0.06] px-6 py-3 text-sm font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.10] hover:text-foreground"
      >
        Browse movies →
      </a>
    </div>
  </main>
</div>
