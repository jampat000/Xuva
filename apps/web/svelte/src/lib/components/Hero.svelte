<script lang="ts">
  import { fade } from "svelte/transition";
  import { Check, Info, Play, Plus } from "lucide-svelte";
  import heroImg from "$lib/assets/hero-featured.jpg";
  import type { Media } from "$lib/mock-data";
  import { toggleWatchlist, isInWatchlist } from '$lib/stores/watchlistStore.svelte';

  // The `trailersEnabled` prop is kept for backwards compatibility with the
  // call site (the home +page.svelte passes it). The Hero no longer plays
  // trailers — those moved to the movie/TV detail pages where the user
  // explicitly opts in. The prop being here is just defensive: if some
  // caller still passes it, we don't error.
  let { slides: rawSlides } = $props<{ slides: Media[]; trailersEnabled?: boolean }>();

  function detailHref(m: Media): string {
    return m.type === 'Series' ? `/tv/${m.id}` : `/movies/${m.id}`;
  }

  // Prefer slides that have a real backdrop. Fall back to the full list only
  // if none have one — keeps the hero cinematic instead of flashing the
  // bundled placeholder JPG between titles.
  let slides = $derived.by(() => {
    const withBackdrop = rawSlides.filter((s: Media) => !!s.backdrop);
    return withBackdrop.length > 0 ? withBackdrop : rawSlides;
  });

  let idx = $state(0);
  let media = $derived(slides[idx] ?? slides[0]);
  let backdrop = $derived(media?.backdrop || heroImg);
  let words = $derived((media?.title ?? "").split(" "));

  // Slide rotation — fixed 12s cadence, no more trailer-aware timing.
  const SLIDE_MS = 12_000;
  $effect(() => {
    void media?.id; // re-run on slide change
    if (typeof window === "undefined" || slides.length <= 1) return;
    const timeoutId = window.setTimeout(() => {
      idx = (idx + 1) % slides.length;
    }, SLIDE_MS);
    return () => window.clearTimeout(timeoutId);
  });
</script>

<!-- Shrunk from 72-80vh (full screen of hero) to 48-58vh — still cinematic,
     but content rows are now visible without scrolling on a 1080p monitor. -->
<section class="relative -mt-16 h-[52vh] min-h-[420px] w-full overflow-hidden sm:h-[56vh] sm:min-h-[460px] lg:h-[60vh] lg:min-h-[500px]">
  <div class="absolute inset-0">
    {#key idx}
      <img
        src={backdrop}
        alt=""
        fetchpriority="high"
        class="absolute inset-0 animate-kenburns h-full w-full object-cover object-top"
        width="1920"
        height="1080"
        in:fade={{ duration: 800 }}
        out:fade={{ duration: 800 }}
        onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
      />
    {/key}

    <div class="absolute inset-0 bg-[radial-gradient(ellipse_at_center,transparent_30%,oklch(0.06_0.01_280/0.8)_100%)]"></div>
    <div class="absolute inset-0 bg-gradient-to-r from-background via-background/70 to-transparent"></div>
    <div class="absolute inset-x-0 bottom-0 h-2/3 bg-gradient-to-t from-background via-background/60 to-transparent"></div>
    <div class="absolute inset-x-0 top-0 h-24 bg-gradient-to-b from-background/80 to-transparent"></div>
  </div>

  <div class="grain pointer-events-none absolute inset-0"></div>

  <div class="relative flex h-full flex-col justify-between px-6 pb-10 pt-24 md:px-12 md:pb-12 lg:px-20 lg:pb-14">
    {#key media.id}
      <div class="animate-fade-up max-w-4xl flex-1 flex flex-col justify-end">
        <div class="mb-4 flex items-center gap-3 text-[10px] font-medium uppercase tracking-[0.4em] text-muted-foreground">
          <span class="h-px w-8 bg-foreground/40"></span>
          Xuva Presents
        </div>

        {#if media.logo}
          <!-- Clearlogo: SVG/PNG treatment matching Disney+/Netflix presentation -->
          <img
            src={media.logo}
            alt={media.title}
            class="max-h-20 w-auto max-w-[340px] object-contain drop-shadow-2xl md:max-h-24 lg:max-h-28"
            style="filter: drop-shadow(0 4px 24px oklch(0 0 0 / 0.6));"
          />
        {:else}
          <h1 class="font-serif-display whitespace-nowrap text-[clamp(2.25rem,5vw,4.5rem)] leading-[0.95] tracking-tight text-foreground">
            {#each words as word, i (i)}
              {#if i === words.length - 1}
                <em class="italic text-foreground/95">{word}</em>
              {:else}
                {word}
              {/if}
              {#if i < words.length - 1}
                {" "}
              {/if}
            {/each}
          </h1>
        {/if}

        <div class="mt-4 flex flex-wrap items-center gap-x-3 gap-y-1 text-[11px] font-medium uppercase tracking-[0.18em] text-muted-foreground">
          {#if media.year && media.year > 0}
            <span class="text-foreground/90">{media.year}</span>
            <span class="opacity-30">·</span>
          {/if}
          <span>{media.director ?? media.type}</span>
          {#if media.type === 'Series' && media.seasons}
            <span class="opacity-30">·</span>
            <span>{media.seasons} Season{media.seasons !== 1 ? 's' : ''}</span>
          {:else if media.runtime}
            <span class="opacity-30">·</span>
            <span>{media.runtime}</span>
          {/if}
          {#if media.rating > 0}
            <span class="opacity-30">·</span>
            <span class="flex items-center gap-1.5 normal-case tracking-normal text-foreground/90">
              <span class="text-amber-300">★</span>
              {media.rating.toFixed(1)}
            </span>
          {/if}
          {#if media.badge}
            <span class="opacity-30">·</span>
            <span class="rounded-sm border border-foreground/20 px-2 py-0.5 text-[10px] tracking-[0.2em] text-foreground/80">
              {media.badge}
            </span>
          {/if}
        </div>

        <p class="mt-3 max-w-3xl line-clamp-2 text-[14px] leading-relaxed text-foreground/75 md:text-[15px]">
          {media.synopsis}
        </p>

        <div class="mt-5 flex flex-wrap items-center gap-3">
          <a
            href={detailHref(media)}
            class="group inline-flex items-center gap-2.5 rounded-full bg-foreground px-6 py-3 text-sm font-semibold text-background transition-all hover:bg-foreground/90"
          >
            <Play class="h-4 w-4 fill-background" />
            Play
          </a>
          <a
            href={detailHref(media)}
            class="inline-flex items-center gap-2.5 rounded-full border border-foreground/15 bg-foreground/5 px-5 py-3 text-sm font-medium text-foreground backdrop-blur-md transition-colors hover:bg-foreground/10"
          >
            <Info class="h-4 w-4" />
            More info
          </a>
          <button
            type="button"
            onclick={() => toggleWatchlist({ id: media.id, kind: media.type === 'Series' ? 'series' : 'movie', title: media.title, year: media.year || undefined, posterUrl: media.poster, backdropUrl: media.backdrop, genres: media.genres })}
            aria-label={isInWatchlist(media.id, media.type === 'Series' ? 'series' : 'movie') ? 'Remove from watchlist' : 'Add to watchlist'}
            class="hairline flex h-11 w-11 items-center justify-center rounded-full bg-foreground/5 text-foreground backdrop-blur-md transition-colors hover:bg-foreground/10"
          >
            {#if isInWatchlist(media.id, media.type === 'Series' ? 'series' : 'movie')}
              <Check class="h-5 w-5 text-primary-glow" />
            {:else}
              <Plus class="h-5 w-5" />
            {/if}
          </button>
        </div>
      </div>
    {/key}

    <!-- Slide indicator dots — kept since multi-featured rotation is still on -->
    {#if slides.length > 1}
      <div class="pointer-events-none absolute inset-x-6 bottom-5 flex items-center md:inset-x-12 lg:inset-x-20">
        <div class="pointer-events-auto flex items-center gap-3">
          <span class="text-[9px] font-semibold uppercase tracking-[0.3em] text-foreground/40">
            Featured
          </span>
          {#each slides as slide, i (slide.id)}
            <button
              type="button"
              onclick={() => (idx = i)}
              aria-label={`Show ${slide.title}`}
              class="group/dot flex items-center"
            >
              <span
                class={`block h-[2px] transition-all duration-500 ${
                  i === idx ? "w-10 bg-foreground" : "w-5 bg-foreground/25 group-hover/dot:bg-foreground/50"
                }`}
              ></span>
            </button>
          {/each}
        </div>
      </div>
    {/if}
  </div>
</section>
