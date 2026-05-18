<script lang="ts">
  import { page } from '$app/state';
  import { onMount } from 'svelte';
  import { Play, Plus, Star, Clock, ChevronLeft } from 'lucide-svelte';
  import Header from '$lib/components/Header.svelte';
  import SubtitleSelector from '$lib/components/SubtitleSelector.svelte';
  import { getMovieDetail } from '$lib/api/home';
  import { getMetadataRecords, refreshMetadataItem, applyMetadataMatch } from '$lib/api/browse';
  import type { MovieDetailResponse } from '$lib/api/home';
  import type { MetadataRecord } from '$lib/api/browse';

  const id = $derived(page.params.id ?? '');

  let detail = $state<MovieDetailResponse | null>(null);
  let metadata = $state<MetadataRecord | null>(null);
  let altRecords = $state<MetadataRecord[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);
  let fixingMeta = $state(false);
  let showMetaPanel = $state(false);
  let refreshing = $state(false);

  const title = $derived(
    (detail?.metadata as Record<string, unknown> | undefined)?.title as string
      ?? detail?.title
      ?? 'Unknown'
  );
  const year = $derived(
    (detail?.metadata as Record<string, unknown> | undefined)?.year as number | undefined
  );
  const overview = $derived(detail?.metadata?.overview ?? '');
  const posterUrl = $derived(
    (detail?.metadata as Record<string, unknown> | undefined)?.posterUrl as string | undefined
  );
  const backdropUrl = $derived(
    (detail?.metadata as Record<string, unknown> | undefined)?.backdropUrl as string | undefined
  );
  const genres = $derived(
    (detail?.metadata as Record<string, unknown> | undefined)?.genres as string[] | undefined ?? []
  );
  const rating = $derived(
    (detail?.metadata as Record<string, unknown> | undefined)?.voteAverage as number | undefined
  );
  const runtime = $derived(
    (detail?.metadata as Record<string, unknown> | undefined)?.runtime as string | undefined
  );
  const versionCount = $derived(detail?.versions?.length ?? 0);
  const mediaSourceId = $derived(detail?.versions?.[0]?.mediaSourceId);

  // Build play URL including back-link and title for the player chrome
  const basePlayUrl = $derived(
    mediaSourceId
      ? `/play/${mediaSourceId}?title=${encodeURIComponent(title)}&back=${encodeURIComponent(`/movies/${id}`)}`
      : ''
  );
  let playHref = $state('');

  async function load() {
    try {
      loading = true;
      error = null;
      const [detailResp, metaResp] = await Promise.all([
        getMovieDetail(id),
        getMetadataRecords('movie', id).catch(() => ({ best: null, records: [] }))
      ]);
      detail = detailResp;
      metadata = metaResp.best ?? null;
      altRecords = metaResp.records ?? [];
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load';
    } finally {
      loading = false;
    }
  }

  async function refreshMeta() {
    refreshing = true;
    try {
      await refreshMetadataItem({ kind: 'movie', id });
      await load();
    } finally {
      refreshing = false;
    }
  }

  onMount(load);
</script>

<svelte:head>
  <title>{title} — Xuva</title>
  <meta name="description" content={overview || `Watch ${title} on Xuva.`} />
</svelte:head>

<div class="min-h-screen bg-background">
  <Header />

  {#if loading}
    <div class="flex min-h-[60vh] items-center justify-center">
      <div class="h-10 w-10 animate-spin rounded-full border-2 border-border border-t-primary-glow"></div>
    </div>

  {:else if error}
    <div class="flex min-h-[60vh] flex-col items-center justify-center gap-3 px-6 text-center">
      <p class="text-base font-medium text-foreground/80">Can't load this title</p>
      <p class="max-w-xs text-sm text-muted-foreground">Make sure your Xuva server is running, then try again.</p>
      <button onclick={load} class="mt-2 hairline rounded-full bg-foreground/[0.06] px-5 py-2.5 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground">
        Try again
      </button>
    </div>

  {:else}
    <!-- Backdrop -->
    <div class="relative -mt-16 h-[60vh] min-h-[480px] w-full overflow-hidden">
      {#if backdropUrl}
        <img
          src={backdropUrl}
          alt=""
          class="h-full w-full object-cover"
          onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
        />
      {/if}
      <div class="absolute inset-0 bg-gradient-to-r from-background via-background/70 to-transparent"></div>
      <div class="absolute inset-x-0 bottom-0 h-2/3 bg-gradient-to-t from-background to-transparent"></div>
      <div class="absolute inset-x-0 top-0 h-24 bg-gradient-to-b from-background/80 to-transparent"></div>
    </div>

    <div class="relative -mt-48 px-6 pb-32 md:px-12 lg:px-20">
      <a href="/movies" class="mb-8 inline-flex items-center gap-2 text-sm text-muted-foreground transition-colors hover:text-foreground">
        <ChevronLeft class="h-4 w-4" /> Back to Movies
      </a>

      <div class="grid gap-10 md:grid-cols-[240px_minmax(0,1fr)] lg:grid-cols-[280px_minmax(0,1fr)] lg:gap-16">
        <!-- Poster -->
        <div class="shrink-0">
          <div
            class="shadow-poster aspect-[2/3] w-full overflow-hidden rounded-2xl"
            style="background: linear-gradient(135deg, #1e3a5f, #0f172a);"
          >
            {#if posterUrl}
              <img
                src={posterUrl}
                alt={title}
                class="h-full w-full object-cover"
                onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
              />
            {/if}
          </div>
        </div>

        <!-- Details -->
        <div>
          <div class="flex flex-wrap items-center gap-x-3 gap-y-1 text-[11px] uppercase tracking-[0.22em] text-muted-foreground">
            {#if year}<span>{year}</span>{/if}
            {#if year && genres.length}<span class="opacity-30">·</span>{/if}
            {#each genres as g, i (g)}
              <span>{g}</span>{#if i < genres.length - 1}<span class="opacity-30">·</span>{/if}
            {/each}
            {#if runtime}
              <span class="opacity-30">·</span>
              <span class="flex items-center gap-1"><Clock class="h-3 w-3" />{runtime}</span>
            {/if}
            {#if rating}
              <span class="opacity-30">·</span>
              <span class="flex items-center gap-1 text-amber-300"><Star class="h-3 w-3 fill-current" />{rating.toFixed(1)}</span>
            {/if}
          </div>

          <h1 class="font-serif-display mt-3 text-[clamp(2rem,5vw,4rem)] leading-[0.95] tracking-tight">
            {title}
          </h1>

          {#if overview}
            <p class="mt-5 max-w-2xl text-base leading-relaxed text-foreground/75">
              {overview}
            </p>
          {/if}

          <div class="mt-8 flex flex-wrap items-center gap-3">
            {#if mediaSourceId}
              <a
                href={playHref || basePlayUrl || `/play/${mediaSourceId}`}
                class="inline-flex items-center gap-2.5 rounded-full bg-foreground px-7 py-3.5 text-sm font-semibold text-background transition-all hover:bg-foreground/90"
              >
                <Play class="h-4 w-4 fill-background" /> Play
              </a>
              <SubtitleSelector {mediaSourceId} {basePlayUrl} bind:playHref />
            {:else}
              <button
                disabled
                class="inline-flex cursor-not-allowed items-center gap-2.5 rounded-full bg-foreground/30 px-7 py-3.5 text-sm font-semibold text-background/60"
              >
                <Play class="h-4 w-4 fill-background/60" /> No source
              </button>
            {/if}
            <button
              type="button"
              aria-label="Add to watchlist"
              class="hairline flex h-12 w-12 items-center justify-center rounded-full bg-foreground/5 text-foreground backdrop-blur-md transition-colors hover:bg-foreground/10"
            >
              <Plus class="h-5 w-5" />
            </button>
          </div>

          <!-- Technical info strip -->
          {#if versionCount > 0}
            <div class="mt-8 flex flex-wrap gap-x-6 gap-y-2 text-[11px] uppercase tracking-[0.2em] text-muted-foreground">
              <span><span class="text-foreground/80">{versionCount}</span> {versionCount === 1 ? 'version' : 'versions'}</span>
            </div>
          {/if}

          <!-- Metadata correction panel (#9) -->
          <div class="mt-10 border-t border-border pt-8">
            <div class="flex items-center justify-between">
              <h3 class="text-sm font-semibold uppercase tracking-[0.22em] text-muted-foreground">Metadata</h3>
              <div class="flex items-center gap-2">
                <button
                  type="button"
                  onclick={refreshMeta}
                  disabled={refreshing}
                  class="hairline rounded-full bg-foreground/[0.04] px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.08] hover:text-foreground disabled:opacity-40"
                >
                  {refreshing ? 'Refreshing…' : 'Refresh'}
                </button>
                <button
                  type="button"
                  onclick={() => (showMetaPanel = !showMetaPanel)}
                  class="hairline rounded-full bg-foreground/[0.04] px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.08] hover:text-foreground"
                >
                  Fix match
                </button>
              </div>
            </div>

            {#if showMetaPanel}
              <div class="mt-4 space-y-3">
                {#if altRecords.length === 0}
                  <p class="text-sm text-muted-foreground">No alternative metadata records found. Try refreshing.</p>
                {:else}
                  <p class="text-xs text-muted-foreground">Select the correct match for this title:</p>
                  {#each altRecords as rec ((rec.provider ?? '') + (rec.title ?? ''))}
                    <button
                      type="button"
                      onclick={async () => {
                        fixingMeta = true;
                        try {
                          await applyMetadataMatch({
                            kind: 'movie', id,
                            title: rec.title ?? title,
                            year: rec.year,
                            overview: rec.overview,
                            provider: rec.provider ?? '',
                            posterUrl: rec.posterUrl,
                            backdropUrl: rec.backdropUrl,
                            review: false
                          });
                          showMetaPanel = false;
                          await load();
                        } finally {
                          fixingMeta = false;
                        }
                      }}
                      class="hairline flex w-full items-start gap-4 rounded-xl bg-surface/40 p-4 text-left transition-colors hover:bg-surface/70"
                    >
                      {#if rec.posterUrl}
                        <img src={rec.posterUrl} alt={rec.title} class="h-16 w-11 shrink-0 rounded-lg object-cover" onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')} />
                      {:else}
                        <div class="h-16 w-11 shrink-0 rounded-lg bg-surface-elevated/60"></div>
                      {/if}
                      <div class="min-w-0">
                        <div class="font-semibold">{rec.title ?? 'Unknown'}</div>
                        <div class="mt-0.5 text-xs text-muted-foreground">{rec.year ?? ''} · {rec.provider ?? ''}</div>
                        {#if rec.overview}
                          <p class="mt-1 line-clamp-2 text-xs text-muted-foreground">{rec.overview}</p>
                        {/if}
                      </div>
                    </button>
                  {/each}
                {/if}
              </div>
            {/if}
          </div>
        </div>
      </div>
    </div>
  {/if}
</div>
