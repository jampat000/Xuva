<script lang="ts">
  import { onMount } from 'svelte';
  import { ChevronLeft } from 'lucide-svelte';
  import Header from '$lib/components/Header.svelte';
  import Logo from '$lib/components/Logo.svelte';

  // Pulled from the server's public /api/system/version endpoint (no auth, no
  // PII — just release/upgrade identity). Rendered in the "Version" block at
  // the bottom of the page so users can quote it in bug reports without
  // hunting through Settings or the installer receipt.
  let version = $state<string | null>(null);
  let commit = $state<string | null>(null);
  let buildDate = $state<string | null>(null);

  onMount(async () => {
    try {
      const res = await fetch('/api/system/version');
      if (!res.ok) return;
      const data = await res.json();
      if (typeof data?.version === 'string') version = data.version;
      if (typeof data?.commit === 'string') commit = data.commit;
      if (typeof data?.buildDate === 'string') buildDate = data.buildDate;
    } catch {
      // best-effort: leave version null so the block renders "unknown"
    }
  });

  // Show the short commit (first 7 chars) — the full SHA is noisy and the
  // short form is what GitHub links and `git log` display by default.
  const shortCommit = $derived(commit ? commit.slice(0, 7) : null);
</script>

<svelte:head>
  <title>About — Xuva</title>
  <meta name="description" content="About Xuva — a self-hosted personal media server." />
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
        Your personal media library.
      </h1>
      <p class="mt-6 text-base leading-relaxed text-muted-foreground">
        Xuva is a self-hosted media server for your movie and TV collection. It runs on your own hardware — no subscriptions, no cloud dependency.
      </p>
      <p class="mt-4 text-base leading-relaxed text-muted-foreground">
        Built on open-source foundations including SvelteKit, Go, and FFmpeg. Xuva puts you in control of your media — no subscriptions, no cloud dependency, no compromises.
      </p>

      <div class="mt-12 border-t border-border pt-8">
        <h2 class="font-serif-display text-xl tracking-tight">Version</h2>
        <dl class="mt-4 grid grid-cols-[max-content_1fr] gap-x-6 gap-y-2 text-sm">
          <dt class="text-muted-foreground">Release</dt>
          <dd class="font-mono">{version ?? 'unknown'}</dd>
          {#if shortCommit}
            <dt class="text-muted-foreground">Commit</dt>
            <dd class="font-mono">{shortCommit}</dd>
          {/if}
          {#if buildDate}
            <dt class="text-muted-foreground">Built</dt>
            <dd class="font-mono">{buildDate}</dd>
          {/if}
        </dl>
      </div>

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
