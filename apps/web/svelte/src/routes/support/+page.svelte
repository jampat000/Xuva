<script lang="ts">
  import { ChevronLeft, MessageSquare, BookOpen, GitBranch } from 'lucide-svelte';
  import Header from '$lib/components/Header.svelte';
</script>

<svelte:head>
  <title>Support — Xuva</title>
  <meta name="description" content="Get help with Xuva — documentation, community, and bug reports." />
</svelte:head>

<div class="min-h-screen bg-background">
  <Header />

  <main class="px-6 pb-32 pt-24 md:px-12 md:pt-28 lg:px-20">
    <a href="/" class="mb-10 inline-flex items-center gap-2 text-sm text-muted-foreground transition-colors hover:text-foreground">
      <ChevronLeft class="h-4 w-4" /> Home
    </a>

    <div class="mx-auto max-w-2xl">
      <h1 class="font-serif-display text-[clamp(2rem,5vw,3.5rem)] leading-[1] tracking-tight">
        Support
      </h1>
      <p class="mt-4 text-base leading-relaxed text-muted-foreground">
        Need help with Xuva? Here are the best places to start.
      </p>

      <div class="mt-10 space-y-4">
        {#each [
          { icon: BookOpen, label: 'Documentation', desc: 'Setup guides, configuration reference, and API docs.', href: '#', cta: 'Read the docs' },
          { icon: GitBranch, label: 'GitHub Issues', desc: 'Found a bug or have a feature request? Open an issue on GitHub.', href: '#', cta: 'Open an issue' },
          { icon: MessageSquare, label: 'Community', desc: 'Discuss Xuva with other users and get community support.', href: '#', cta: 'Join the discussion' },
        ] as item (item.label)}
          <div class="hairline flex items-start gap-5 rounded-2xl bg-surface/40 p-6 transition-colors hover:bg-surface/60">
            <div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-surface-elevated/60 text-primary-glow">
              <item.icon class="h-5 w-5" />
            </div>
            <div class="flex-1">
              <div class="font-semibold">{item.label}</div>
              <p class="mt-1 text-sm text-muted-foreground">{item.desc}</p>
              <a href={item.href} class="mt-3 inline-block text-xs font-medium text-primary-glow underline-offset-2 hover:underline">
                {item.cta} →
              </a>
            </div>
          </div>
        {/each}
      </div>
    </div>
  </main>
</div>
