<script lang="ts">
  import { ChevronLeft } from 'lucide-svelte';
  import Header from '$lib/components/Header.svelte';

  const currentYear = new Date().getFullYear();
</script>

<svelte:head>
  <title>Privacy Policy — Xuva</title>
  <meta name="description" content="Xuva privacy policy — your data stays on your hardware." />
</svelte:head>

<div class="min-h-screen bg-background">
  <Header />

  <main class="px-6 pb-32 pt-24 md:px-12 md:pt-28 lg:px-20">
    <a href="/" class="mb-10 inline-flex items-center gap-2 text-sm text-muted-foreground transition-colors hover:text-foreground">
      <ChevronLeft class="h-4 w-4" /> Home
    </a>

    <div class="mx-auto max-w-2xl">
      <h1 class="font-serif-display text-[clamp(2rem,5vw,3.5rem)] leading-[1] tracking-tight">
        Privacy Policy
      </h1>
      <p class="mt-2 text-xs uppercase tracking-[0.2em] text-muted-foreground">Last updated {currentYear}</p>

      <div class="mt-10 space-y-8 text-sm leading-relaxed text-muted-foreground">
        <section>
          <h2 class="font-serif-display mb-3 text-xl tracking-tight text-foreground">Your data stays on your hardware</h2>
          <p>Xuva is a self-hosted media server. All your media, watch history, and account data lives on the machine you install Xuva on. We don't collect, transmit, or store any personal data on Xuva servers.</p>
        </section>

        <section>
          <h2 class="font-serif-display mb-3 text-xl tracking-tight text-foreground">Metadata providers</h2>
          <p>When Xuva fetches metadata (posters, descriptions, ratings), it sends movie and TV show titles to third-party providers such as TMDB and TheTVDB. No personal information is sent. Review each provider's privacy policy for details.</p>
        </section>

        <section>
          <h2 class="font-serif-display mb-3 text-xl tracking-tight text-foreground">Watchlist services</h2>
          <p>If you connect optional services like Trakt or Letterboxd, Xuva will share watch events with those services per your configuration. You can disconnect at any time from Settings → Watchlist Services.</p>
        </section>

        <section>
          <h2 class="font-serif-display mb-3 text-xl tracking-tight text-foreground">No analytics, no telemetry</h2>
          <p>Xuva does not include any analytics or usage telemetry. Your viewing habits are yours alone.</p>
        </section>
      </div>
    </div>
  </main>
</div>
