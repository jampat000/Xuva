<script lang="ts">
  import { onMount } from 'svelte';
  import { Play, ChevronLeft } from 'lucide-svelte';
  import Header from '$lib/components/Header.svelte';
  import { getPlaybackRecent } from '$lib/api/home';
  import type { PlaybackRecentItem } from '$lib/api/home';

  let items = $state<PlaybackRecentItem[]>([]);
  let loading = $state(true);

  onMount(async () => {
    try {
      const resp = await getPlaybackRecent(undefined, 50);
      items = resp.recent ?? [];
    } catch {
      items = [];
    } finally {
      loading = false;
    }
  });

  function friendlyName(item: PlaybackRecentItem): string {
    if (item.name) return item.name;
    if (item.relPath) {
      return item.relPath
        .replace(/\.[a-z0-9]{2,4}$/i, '')
        .replace(/\s*\([^)]*(?:remux|bluray|1080p|2160p|720p|hdtv)[^)]*\)/gi, '')
        .split('/').pop() ?? item.relPath;
    }
    return 'Unknown';
  }

  function progressLabel(item: PlaybackRecentItem): string {
    if (item.watched) return 'Watched';
    if ((item.percent ?? 0) > 0) return `${Math.round(item.percent!)}%`;
    return 'Not started';
  }
</script>

<svelte:head>
  <title>Continue Watching — Xuva</title>
  <meta name="description" content="Pick up where you left off — all your in-progress films and shows in one place." />
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
      <div class="mb-2 text-[10px] font-semibold uppercase tracking-[0.35em] text-primary-glow">In progress</div>
      <h1 class="font-serif-display text-[clamp(2rem,4vw,3.25rem)] leading-[1] tracking-tight">Continue Watching</h1>
      <p class="mt-3 max-w-xl text-sm text-muted-foreground">
        Pick up where you left off across all your films and series.
      </p>
    </header>

    {#if loading}
      <div class="flex min-h-[40vh] items-center justify-center">
        <div class="h-8 w-8 animate-spin rounded-full border-2 border-border border-t-primary-glow"></div>
      </div>
    {:else if items.length === 0}
      <div class="flex flex-col items-center justify-center py-24 text-center">
        <div class="hairline flex h-14 w-14 items-center justify-center rounded-2xl bg-surface-elevated/70 text-primary-glow shadow-elev">
          <Play class="h-6 w-6" />
        </div>
        <p class="font-serif-display mt-5 text-2xl tracking-tight">Nothing in progress</p>
        <p class="mt-2 max-w-sm text-sm text-muted-foreground">
          Start watching something and it will appear here automatically.
        </p>
        <a
          href="/"
          class="hairline mt-6 inline-flex items-center gap-2 rounded-full bg-foreground/[0.06] px-6 py-3 text-sm font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.10] hover:text-foreground"
        >
          Browse library →
        </a>
      </div>
    {:else}
      <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {#each items as item (item.mediaSourceId)}
          <a
            href={`/play/${item.mediaSourceId}`}
            class="hairline group flex flex-col gap-3 rounded-2xl bg-surface/40 p-4 transition-colors hover:bg-surface/70"
          >
            <div class="flex items-center gap-3">
              <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-surface-elevated text-muted-foreground">
                <Play class="h-4 w-4" />
              </div>
              <div class="min-w-0 flex-1">
                <p class="truncate text-sm font-medium">{friendlyName(item)}</p>
                <p class="text-xs text-muted-foreground">{progressLabel(item)}</p>
              </div>
            </div>
            {#if (item.percent ?? 0) > 0 && !item.watched}
              <div class="h-1 w-full overflow-hidden rounded-full bg-foreground/10">
                <div
                  class="h-full rounded-full bg-primary-glow transition-all"
                  style="width: {item.percent}%"
                ></div>
              </div>
            {/if}
          </a>
        {/each}
      </div>
    {/if}
  </main>
</div>
