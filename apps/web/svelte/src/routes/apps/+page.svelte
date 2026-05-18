<script lang="ts">
  import { ChevronLeft, Tv, Smartphone, Tablet, Monitor } from 'lucide-svelte';
  import Header from '$lib/components/Header.svelte';
</script>

<svelte:head>
  <title>Apps — Xuva</title>
  <meta name="description" content="Download Xuva on every screen — iOS, Android, tvOS, Android TV, and web." />
</svelte:head>

<div class="min-h-screen bg-background">
  <Header />

  <main class="px-6 pb-32 pt-24 md:px-12 md:pt-28 lg:px-20">
    <a href="/" class="mb-10 inline-flex items-center gap-2 text-sm text-muted-foreground transition-colors hover:text-foreground">
      <ChevronLeft class="h-4 w-4" /> Home
    </a>

    <div class="mx-auto max-w-3xl">
      <div class="mb-2 text-[10px] font-semibold uppercase tracking-[0.35em] text-primary-glow">Every screen</div>
      <h1 class="font-serif-display text-[clamp(2rem,5vw,3.5rem)] leading-[1] tracking-tight">
        Xuva apps
      </h1>
      <p class="mt-4 max-w-xl text-base leading-relaxed text-muted-foreground">
        Your library travels with you. Xuva is designed for every form factor — from couch to commute.
      </p>

      <div class="mt-12 grid gap-4 sm:grid-cols-2">
        {#each [
          { icon: Monitor, label: 'Web', sub: 'Cinematic browser experience', status: 'Available now', available: true },
          { icon: Smartphone, label: 'iOS & Android', sub: 'Thumb-friendly mobile app', status: 'Coming soon', available: false },
          { icon: Tablet, label: 'iPad', sub: 'Sidebar-rich tablet layout', status: 'Coming soon', available: false },
          { icon: Tv, label: 'tvOS & Android TV', sub: 'Remote-friendly 10-foot UI', status: 'Coming soon', available: false },
        ] as app (app.label)}
          <div class={`hairline flex items-start gap-4 rounded-2xl bg-surface/40 p-6 ${!app.available ? 'opacity-60' : ''}`}>
            <div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-surface-elevated/60 text-primary-glow">
              <app.icon class="h-5 w-5" />
            </div>
            <div>
              <div class="font-semibold">{app.label}</div>
              <p class="mt-0.5 text-xs text-muted-foreground">{app.sub}</p>
              <span class={`mt-3 inline-block rounded-full px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.2em] ${app.available ? 'bg-emerald-400/10 text-emerald-300' : 'bg-foreground/[0.06] text-muted-foreground'}`}>
                {app.status}
              </span>
            </div>
          </div>
        {/each}
      </div>
    </div>
  </main>
</div>
