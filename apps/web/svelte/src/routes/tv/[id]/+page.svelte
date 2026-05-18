<script lang="ts">
  import { page } from '$app/state';
  import { onMount } from 'svelte';
  import { Star, ChevronLeft, Play, Plus, Tv } from 'lucide-svelte';
  import Header from '$lib/components/Header.svelte';
  import SubtitleSelector from '$lib/components/SubtitleSelector.svelte';
  import { getSeriesDetail } from '$lib/api/home';
  import { getMetadataRecords, refreshMetadataItem, applyMetadataMatch } from '$lib/api/browse';
  import type { SeriesDetailResponse } from '$lib/api/home';
  import type { MetadataRecord } from '$lib/api/browse';

  const id = $derived(page.params.id ?? '');

  let detail = $state<SeriesDetailResponse | null>(null);
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

  const seasons = $derived(detail?.seasons ?? []);
  const seasonCount = $derived(seasons.length);
  const episodeCount = $derived(seasons.reduce((sum, s) => sum + (s.episodes?.length ?? 0), 0));

  // Find first playable episode across all seasons
  const firstMediaSourceId = $derived(() => {
    for (const season of seasons) {
      for (const ep of season.episodes ?? []) {
        const msid = ep.versions?.[0]?.mediaSourceId;
        if (msid) return msid;
      }
    }
    return undefined;
  });

  // Build play URL with back-link and title for player chrome
  const basePlayUrl = $derived(
    firstMediaSourceId()
      ? `/play/${firstMediaSourceId()}?title=${encodeURIComponent(title)}&back=${encodeURIComponent(`/tv/${id}`)}`
      : ''
  );

  function episodePlayUrl(mediaSourceId: string, epTitle?: string): string {
    const t = epTitle ? `${title} — ${epTitle}` : title;
    return `/play/${mediaSourceId}?title=${encodeURIComponent(t)}&back=${encodeURIComponent(`/tv/${id}`)}`;
  }

  let playHref = $state('');

  async function load() {
    try {
      loading = true;
      error = null;
      const [detailResp, metaResp] = await Promise.all([
        getSeriesDetail(id),
        getMetadataRecords('series', id).catch(() => ({ best: null, records: [] }))
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
      await refreshMetadataItem({ kind: 'series', id });
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
      <a href="/tv" class="mb-8 inline-flex items-center gap-2 text-sm text-muted-foreground transition-colors hover:text-foreground">
        <ChevronLeft class="h-4 w-4" /> Back to TV
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
            {#if seasonCount > 0}
              <span class="opacity-30">·</span>
              <span class="flex items-center gap-1">
                <Tv class="h-3 w-3" />{seasonCount} Season{seasonCount !== 1 ? 's' : ''}
              </span>
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
            {#if firstMediaSourceId()}
              <a
                href={playHref || basePlayUrl || `/play/${firstMediaSourceId()}`}
                class="inline-flex items-center gap-2.5 rounded-full bg-foreground px-7 py-3.5 text-sm font-semibold text-background transition-all hover:bg-foreground/90"
              >
                <Play class="h-4 w-4 fill-background" /> Play from Start
              </a>
              {#if firstMediaSourceId()}
                <SubtitleSelector mediaSourceId={firstMediaSourceId()!} {basePlayUrl} bind:playHref />
              {/if}
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

          <!-- Episode counts strip -->
          {#if seasonCount > 0}
            <div class="mt-8 flex flex-wrap gap-x-6 gap-y-2 text-[11px] uppercase tracking-[0.2em] text-muted-foreground">
              <span><span class="text-foreground/80">{seasonCount}</span> {seasonCount === 1 ? 'season' : 'seasons'}</span>
              {#if episodeCount > 0}
                <span><span class="text-foreground/80">{episodeCount}</span> {episodeCount === 1 ? 'episode' : 'episodes'}</span>
              {/if}
            </div>
          {/if}

          <!-- Seasons list -->
          {#if seasons.length > 0}
            <div class="mt-10 border-t border-border pt-8">
              <h3 class="text-sm font-semibold uppercase tracking-[0.22em] text-muted-foreground">Seasons</h3>
              <div class="mt-4 space-y-2">
                {#each seasons as season, i (i)}
                  {@const epCount = season.episodes?.length ?? 0}
                  {@const firstEp = season.episodes?.[0]?.versions?.[0]?.mediaSourceId}
                  <div class="hairline flex items-center justify-between rounded-xl bg-surface/30 px-4 py-3">
                    <div>
                      <span class="font-medium">Season {i + 1}</span>
                      {#if epCount > 0}
                        <span class="ml-2 text-xs text-muted-foreground">{epCount} episode{epCount !== 1 ? 's' : ''}</span>
                      {/if}
                    </div>
                    {#if firstEp}
                      <a
                        href={episodePlayUrl(firstEp, `Season ${i + 1}, Episode 1`)}
                        class="inline-flex items-center gap-1.5 rounded-full bg-foreground/[0.06] px-3 py-1.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-foreground/[0.12] hover:text-foreground"
                      >
                        <Play class="h-3 w-3 fill-current" /> Play
                      </a>
                    {/if}
                  </div>
                {/each}
              </div>
            </div>
          {/if}

          <!-- Metadata correction panel -->
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
                            kind: 'series', id,
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
