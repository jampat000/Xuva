<script lang="ts">
  import { ChevronLeft, ChevronRight } from "lucide-svelte";
  import type { Media } from "$lib/mock-data";

  let { items, eyebrow = 'From your library', title = 'Highest rated' } = $props<{ items: Media[], eyebrow?: string, title?: string }>();
  let scrollerRef = $state<HTMLDivElement | null>(null);

  function scroll(dir: "l" | "r"): void {
    const el = scrollerRef;
    if (!el) return;
    el.scrollBy({ left: dir === "l" ? -el.clientWidth * 0.8 : el.clientWidth * 0.8, behavior: "smooth" });
  }

  function detailHref(m: Media): string {
    return m.type === 'Series' ? `/tv/${m.id}` : `/movies/${m.id}`;
  }
</script>

<section class="group/row relative">
  <div class="mb-5 flex items-end justify-between px-6 md:px-12 lg:px-20">
    <div>
      <div class="mb-1.5 text-[10px] font-semibold uppercase tracking-[0.3em] text-primary-glow">
        {eyebrow}
      </div>
      <h2 class="font-serif-display text-3xl tracking-tight md:text-4xl">
        {title}
      </h2>
    </div>
    <div class="hidden gap-1.5 opacity-0 transition-opacity group-hover/row:opacity-100 md:flex">
      <button type="button" onclick={() => scroll("l")} aria-label="Scroll left" class="hairline flex h-9 w-9 items-center justify-center rounded-full bg-background/60 backdrop-blur hover:bg-surface-elevated">
        <ChevronLeft class="h-4 w-4" />
      </button>
      <button type="button" onclick={() => scroll("r")} aria-label="Scroll right" class="hairline flex h-9 w-9 items-center justify-center rounded-full bg-background/60 backdrop-blur hover:bg-surface-elevated">
        <ChevronRight class="h-4 w-4" />
      </button>
    </div>
  </div>

  <!-- Apple TV+ style: each card is just the poster; a small numeric badge
       overlays the top-left corner. Cards flow naturally so the first one
       aligns with the page gutter exactly like Continue Watching / New
       Movies — zero alignment math required. -->
  <div bind:this={scrollerRef} class="scrollbar-none flex gap-3 overflow-x-auto scroll-smooth px-6 pb-8 pt-4 md:gap-4 md:px-12 lg:px-20">
    {#each items as media, i (media.id)}
      <a
        href={detailHref(media)}
        class="group relative flex shrink-0 cursor-pointer flex-col"
      >
        <div
          class="shadow-poster relative aspect-[2/3] w-[140px] shrink-0 overflow-hidden rounded-lg transition-all duration-300 group-hover:-translate-y-2 group-hover:scale-[1.04] group-hover:ring-[3px] group-hover:ring-white/85 group-hover:shadow-[0_28px_60px_-12px_oklch(0_0_0/0.85)] md:w-[170px] lg:w-[180px]"
          style={`background: linear-gradient(135deg, ${media.palette[0]}, ${media.palette[1]})`}
        >
          {#if media.poster}
            <img
              src={media.poster}
              alt={media.title}
              loading="lazy"
              class="absolute inset-0 h-full w-full object-cover"
              onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
            />
          {:else}
            <div
              class="absolute inset-0 opacity-40 mix-blend-overlay"
              style="background-image: radial-gradient(circle at 30% 20%, rgba(255,255,255,0.4), transparent 50%), radial-gradient(circle at 80% 80%, rgba(0,0,0,0.6), transparent 60%);"
            ></div>
          {/if}

          <!-- Rank badge: small frosted-glass pill, top-left, Apple TV+ style.
               Uses min-w-7 so single-digit "1" stays circular-ish while "10"
               expands naturally. -->
          <div class="absolute left-2 top-2 flex h-7 min-w-[1.75rem] items-center justify-center rounded-md bg-black/55 px-1.5 backdrop-blur-md ring-1 ring-white/15">
            <span
              class="font-display text-xs font-bold tracking-tight text-white"
              style="text-shadow: 0 1px 2px rgba(0,0,0,0.5);"
            >
              {i + 1}
            </span>
          </div>

          <!-- Bottom gradient + meta -->
          <div class="absolute inset-x-0 bottom-0 h-1/2 bg-gradient-to-t from-black/85 to-transparent"></div>
          <div class="absolute inset-x-3 bottom-3">
            <h3
              class="font-display text-sm font-bold leading-tight text-white drop-shadow-md md:text-base"
              style="text-shadow: 0 2px 12px rgba(0,0,0,0.7);"
            >
              {media.title}
            </h3>
            <div class="mt-0.5 text-[10px] uppercase tracking-widest text-white/70">
              {#if media.year && media.year > 0}{media.year} • {/if}{media.type}
            </div>
          </div>
        </div>
      </a>
    {/each}
    <div class="w-2 shrink-0"></div>
  </div>
</section>
