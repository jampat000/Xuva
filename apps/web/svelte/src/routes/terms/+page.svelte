<script lang="ts">
  import { ChevronLeft } from 'lucide-svelte';
  import Header from '$lib/components/Header.svelte';

  const currentYear = new Date().getFullYear();
</script>

<svelte:head>
  <title>Terms of Service — Xuva</title>
  <meta name="description" content="Xuva terms of service — personal use, self-hosted media server." />
</svelte:head>

<div class="min-h-screen bg-background">
  <Header />

  <main class="px-6 pb-32 pt-24 md:px-12 md:pt-28 lg:px-20">
    <a href="/" class="mb-10 inline-flex items-center gap-2 text-sm text-muted-foreground transition-colors hover:text-foreground">
      <ChevronLeft class="h-4 w-4" /> Home
    </a>

    <div class="mx-auto max-w-2xl">
      <h1 class="font-serif-display text-[clamp(2rem,5vw,3.5rem)] leading-[1] tracking-tight">
        Terms of Service
      </h1>
      <p class="mt-2 text-xs uppercase tracking-[0.2em] text-muted-foreground">Last updated {currentYear}</p>

      <div class="mt-10 space-y-8 text-sm leading-relaxed text-muted-foreground">
        <section>
          <h2 class="font-serif-display mb-3 text-xl tracking-tight text-foreground">Personal use</h2>
          <p>Xuva is intended for personal, private use to stream media you own or have the right to access. By using Xuva you agree to comply with the copyright laws in your jurisdiction.</p>
        </section>

        <section>
          <h2 class="font-serif-display mb-3 text-xl tracking-tight text-foreground">No warranty</h2>
          <p>Xuva is provided as-is without warranty of any kind. We make no guarantees about uptime, data integrity, or fitness for a particular purpose. Always maintain your own backups.</p>
        </section>

        <section>
          <h2 class="font-serif-display mb-3 text-xl tracking-tight text-foreground">Open source</h2>
          <p>Xuva incorporates open-source software. Each component is subject to its own license. You may use, modify, and distribute Xuva subject to the terms of the applicable open-source licenses.</p>
        </section>

        <section>
          <h2 class="font-serif-display mb-3 text-xl tracking-tight text-foreground">Changes to these terms</h2>
          <p>We may update these terms from time to time. Continued use of Xuva after updates constitutes acceptance of the revised terms.</p>
        </section>
      </div>
    </div>
  </main>
</div>
