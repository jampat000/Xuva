<script lang="ts">
  import { onMount } from 'svelte';
  import { Layers } from 'lucide-svelte';
  import Header from '$lib/components/Header.svelte';
  import ErrorState from '$lib/components/ErrorState.svelte';
  import { appState } from '$lib/stores/appState.svelte';
  import { getCollections } from '$lib/api/home';
  import type { CollectionListItem } from '$lib/api/home';

  let items = $state<CollectionListItem[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);

  async function load() {
    error = null;
    loading = true;
    try {
      const resp = await getCollections();
      items = resp.collections ?? [];
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load collections';
    } finally {
      loading = false;
    }
  }

  onMount(load);
</script>

<svelte:head>
  <title>Collections — {appState.serverName}</title>
</svelte:head>

<div class="min-h-screen bg-background">
  <Header />

  {#if error}
    <ErrorState
      title="Can't load collections"
      message="Make sure your Xuva server is running, then try again."
      actions={[{ label: 'Try again', onClick: load }]}
      diagnosticInfo={error}
    />
  {:else}
    <main class="px-6 pb-32 pt-28 md:px-12 lg:px-20">
      <div class="mb-2 text-[10px] font-semibold uppercase tracking-[0.35em] text-primary-glow">
        Your library
      </div>
      <h1 class="font-serif-display text-[clamp(2rem,5vw,3.5rem)] leading-[0.95] tracking-tight">
        Collections.
      </h1>
      <p class="mt-4 max-w-xl text-sm text-muted-foreground">
        Film franchises and multi-part series grouped together from your library.
      </p>

      {#if loading}
        <div class="mt-16 grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
          {#each { length: 12 } as _}
            <div class="flex flex-col gap-2">
              <div class="aspect-[2/3] animate-pulse rounded-xl bg-surface"></div>
              <div class="h-3 w-3/4 animate-pulse rounded bg-surface"></div>
            </div>
          {/each}
        </div>
      {:else if items.length === 0}
        <div class="mt-24 flex flex-col items-center justify-center gap-4 text-center">
          <Layers class="h-14 w-14 text-muted-foreground/20" />
          <p class="text-muted-foreground">No collections found in your library.</p>
          <p class="text-sm text-muted-foreground/60">Collections are created automatically when your movies belong to a TMDB franchise.</p>
        </div>
      {:else}
        <div class="mt-12 grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
          {#each items as collection (collection.id)}
            <a
              href="/collections/{collection.id}"
              class="group flex flex-col gap-2"
            >
              <div class="relative overflow-hidden rounded-xl bg-surface shadow-poster transition-transform duration-200 group-hover:scale-[1.03]">
                {#if collection.posterUrl}
                  <img
                    src={collection.posterUrl}
                    alt={collection.name}
                    class="aspect-[2/3] w-full object-cover"
                    onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
                    loading="lazy"
                  />
                {:else}
                  <div class="flex aspect-[2/3] w-full items-center justify-center bg-surface-elevated">
                    <Layers class="h-8 w-8 text-muted-foreground/30" />
                  </div>
                {/if}
                <div class="absolute inset-x-0 bottom-0 h-1/3 bg-gradient-to-t from-black/60 to-transparent opacity-0 transition-opacity group-hover:opacity-100"></div>
              </div>
              <div class="min-w-0 px-0.5">
                <p class="truncate text-sm font-medium leading-tight">{collection.name}</p>
                <p class="text-xs text-muted-foreground">{collection.movieCount} {collection.movieCount === 1 ? 'film' : 'films'}</p>
              </div>
            </a>
          {/each}
        </div>
      {/if}
    </main>
  {/if}
</div>
