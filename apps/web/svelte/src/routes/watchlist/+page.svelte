<script lang="ts">
  import { onMount } from 'svelte';
  import { Bookmark, BookmarkX, ChevronLeft, Film, Tv } from 'lucide-svelte';
  import Header from '$lib/components/Header.svelte';
  import { appState } from '$lib/stores/appState.svelte';
  import { getWatchlist, removeFromWatchlist, type WatchlistItem } from '$lib/stores/watchlistStore.svelte';

  // Items are reactive — the store uses $state internally
  const items = $derived(getWatchlist());

  // Trigger init on mount so localStorage is read client-side
  let mounted = $state(false);
  onMount(() => { mounted = true; });
</script>

<svelte:head>
  <title>Watchlist — {appState.serverName}</title>
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
      {#if mounted && items.length > 0}
        <p class="mt-3 text-sm text-muted-foreground">
          {items.length} {items.length === 1 ? 'title' : 'titles'} saved
        </p>
      {:else}
        <p class="mt-3 max-w-xl text-sm text-muted-foreground">
          Films and shows you've saved to watch later, all in one place.
        </p>
      {/if}
    </header>

    {#if !mounted}
      <!-- Avoid flash before localStorage loads -->
      <div class="flex min-h-[30vh] items-center justify-center">
        <div class="h-8 w-8 animate-spin rounded-full border-2 border-border border-t-primary-glow"></div>
      </div>

    {:else if items.length === 0}
      <div class="flex flex-col items-center justify-center py-24 text-center">
        <div class="hairline flex h-14 w-14 items-center justify-center rounded-2xl bg-surface-elevated/70 text-primary-glow shadow-elev">
          <Bookmark class="h-6 w-6" />
        </div>
        <p class="font-serif-display mt-5 text-2xl tracking-tight">Nothing saved yet</p>
        <p class="mt-2 max-w-sm text-sm text-muted-foreground">
          Tap the <span class="font-medium text-foreground/70">+</span> button on any movie or show to add it here.
        </p>
        <div class="mt-6 flex flex-wrap items-center justify-center gap-3">
          <a
            href="/movies"
            class="hairline inline-flex items-center gap-2 rounded-full bg-foreground/[0.06] px-6 py-3 text-sm font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.10] hover:text-foreground"
          >
            Browse movies →
          </a>
          <a
            href="/tv"
            class="hairline inline-flex items-center gap-2 rounded-full bg-foreground/[0.06] px-6 py-3 text-sm font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.10] hover:text-foreground"
          >
            Browse TV →
          </a>
        </div>
      </div>

    {:else}
      <div class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
        {#each items as item (item.id + item.kind)}
          {@const href = item.kind === 'series' ? `/tv/${item.id}` : `/movies/${item.id}`}
          <div class="group relative">
            <a {href} class="block">
              <!-- Poster -->
              <div
                class="shadow-poster relative aspect-[2/3] w-full overflow-hidden rounded-xl bg-surface transition-all duration-300 group-hover:-translate-y-1 group-hover:shadow-glow"
                style={item.posterUrl ? '' : 'background: linear-gradient(135deg, #1e3a5f, #0f172a)'}
              >
                {#if item.posterUrl}
                  <img
                    src={item.posterUrl}
                    alt={item.title}
                    loading="lazy"
                    class="absolute inset-0 h-full w-full object-cover transition-transform duration-300 group-hover:scale-105"
                    onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
                  />
                {/if}
                <div class="absolute inset-x-0 bottom-0 h-1/2 bg-gradient-to-t from-black/80 to-transparent"></div>

                <!-- Kind badge -->
                <div class="absolute left-2 top-2 rounded-md bg-black/55 p-1 backdrop-blur-sm ring-1 ring-white/15">
                  {#if item.kind === 'series'}
                    <Tv class="h-3 w-3 text-white/80" />
                  {:else}
                    <Film class="h-3 w-3 text-white/80" />
                  {/if}
                </div>
              </div>

              <!-- Text -->
              <div class="mt-2.5 px-0.5">
                <h3 class="truncate text-sm font-medium text-foreground">{item.title}</h3>
                {#if item.year}
                  <p class="mt-0.5 text-[11px] text-muted-foreground">{item.year}</p>
                {/if}
              </div>
            </a>

            <!-- Remove button — top-right corner, appears on hover -->
            <button
              type="button"
              onclick={() => removeFromWatchlist(item.id, item.kind)}
              aria-label="Remove from watchlist"
              title="Remove from watchlist"
              class="absolute right-1 top-1 flex h-7 w-7 items-center justify-center rounded-full bg-black/60 text-white/70 opacity-0 backdrop-blur-sm transition-all hover:text-white group-hover:opacity-100"
            >
              <BookmarkX class="h-3.5 w-3.5" />
            </button>
          </div>
        {/each}
      </div>
    {/if}
  </main>
</div>
