<script lang="ts">
  import { ChevronLeft, Monitor } from 'lucide-svelte';
  import Header from '$lib/components/Header.svelte';
</script>

<svelte:head>
  <title>Apps — Xuva</title>
  <meta name="description" content="Access Xuva through your web browser on any device." />
</svelte:head>

<div class="min-h-screen bg-background">
  <Header />

  <main class="px-6 pb-32 pt-24 md:px-12 md:pt-28 lg:px-20">
    <a href="/" class="mb-10 inline-flex items-center gap-2 text-sm text-muted-foreground transition-colors hover:text-foreground">
      <ChevronLeft class="h-4 w-4" /> Home
    </a>

    <div class="mx-auto max-w-3xl">
      <h1 class="font-serif-display text-[clamp(2rem,5vw,3.5rem)] leading-[1] tracking-tight">
        Xuva apps
      </h1>
      <p class="mt-4 max-w-xl text-base leading-relaxed text-muted-foreground">
        Xuva runs as a web app in your browser. Open it on any device on your local network.
      </p>

      <div class="mt-12 max-w-sm">
        <div class="hairline flex items-start gap-4 rounded-2xl bg-surface/40 p-6">
          <div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-surface-elevated/60 text-primary-glow">
            <Monitor class="h-5 w-5" />
          </div>
          <div>
            <div class="font-semibold">Web</div>
            <p class="mt-0.5 text-xs text-muted-foreground">Full browser experience — Chrome, Firefox, Safari</p>
            <span class="mt-3 inline-block rounded-full bg-emerald-400/10 px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.2em] text-emerald-300">
              Available now
            </span>
          </div>
        </div>
      </div>
    </div>
  </main>
</div>
