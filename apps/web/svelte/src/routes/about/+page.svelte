<script lang="ts">
  import { ChevronLeft } from 'lucide-svelte';
  import Header from '$lib/components/Header.svelte';
  import Logo from '$lib/components/Logo.svelte';
</script>

<svelte:head>
  <title>About — Xuva</title>
  <meta name="description" content="About Xuva — your home cinema, on every screen." />
</svelte:head>

<div class="min-h-screen bg-background">
  <Header />

  <main class="px-6 pb-32 pt-24 md:px-12 md:pt-28 lg:px-20">
    <a href="/" class="mb-10 inline-flex items-center gap-2 text-sm text-muted-foreground transition-colors hover:text-foreground">
      <ChevronLeft class="h-4 w-4" /> Home
    </a>

    <div class="mx-auto max-w-2xl">
      <div class="mb-8">
        <Logo />
      </div>

      <h1 class="font-serif-display text-[clamp(2rem,5vw,3.5rem)] leading-[1] tracking-tight">
        Your home cinema, <em>on every screen.</em>
      </h1>
      <p class="mt-6 text-base leading-relaxed text-muted-foreground">
        Xuva is a personal media server designed for cinema lovers. Stream your collection on every screen in your home — from a 77-inch OLED to a tablet on the couch.
      </p>
      <p class="mt-4 text-base leading-relaxed text-muted-foreground">
        Built on open-source foundations including SvelteKit, Go, and FFmpeg. Xuva puts you in control of your media — no subscriptions, no cloud dependency, no compromises.
      </p>

      <div class="mt-12 border-t border-border pt-8">
        <h2 class="font-serif-display text-xl tracking-tight">Built with</h2>
        <ul class="mt-4 space-y-3 text-sm text-muted-foreground">
          {#each ['SvelteKit + Svelte 5', 'Go', 'FFmpeg', 'Tailwind CSS', 'Lucide icons'] as tech (tech)}
            <li class="flex items-center gap-3">
              <span class="h-1 w-1 rounded-full bg-primary-glow"></span>
              {tech}
            </li>
          {/each}
        </ul>
      </div>
    </div>
  </main>
</div>
