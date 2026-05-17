<script lang="ts">
  import { ChevronLeft, ChevronRight } from "lucide-svelte";
  import type { Media } from "$lib/mock-data";

  let { items } = $props<{ items: Media[] }>();
  let scrollerRef = $state<HTMLDivElement | null>(null);

  function scroll(dir: "l" | "r"): void {
    const el = scrollerRef;
    if (!el) return;
    el.scrollBy({ left: dir === "l" ? -el.clientWidth * 0.8 : el.clientWidth * 0.8, behavior: "smooth" });
  }
</script>

<section class="group/row relative">
  <div class="mb-5 flex items-end justify-between px-6 md:px-12 lg:px-20">
    <div>
      <div class="mb-1.5 text-[10px] font-semibold uppercase tracking-[0.3em] text-primary-glow">
        Trending this week
      </div>
      <h2 class="font-serif-display text-3xl tracking-tight md:text-4xl">
        Top 10 <em class="italic opacity-70">in your country</em>
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

  <div bind:this={scrollerRef} class="scrollbar-none flex gap-3 overflow-x-auto scroll-smooth px-6 pb-8 md:gap-4 md:px-12 lg:px-20">
    {#each items as media, i (media.id)}
      <article class="group relative flex shrink-0 cursor-pointer items-end gap-1 md:gap-2">
        <span
          class="font-serif-display select-none text-[8rem] leading-[0.78] tracking-tighter text-transparent md:text-[12rem]"
          style={`-webkit-text-stroke: 1.5px oklch(0.45 0.04 280); text-shadow: 0 12px 40px oklch(0 0 0 / 0.5); margin-right: -0.35em; padding-left: ${i === 0 ? "0.1em" : "0"};`}
        >
          {i + 1}
        </span>
        <div
          class="shadow-poster relative aspect-[2/3] w-[120px] shrink-0 overflow-hidden rounded-lg transition-all duration-500 group-hover:-translate-y-1.5 md:w-[160px]"
          style={`background: linear-gradient(135deg, ${media.palette[0]}, ${media.palette[1]})`}
        >
          {#if media.poster}
            <img
              src={media.poster}
              alt={media.title}
              loading="lazy"
              class="absolute inset-0 h-full w-full object-cover"
            />
          {:else}
            <div
              class="absolute inset-0 opacity-40 mix-blend-overlay"
              style="background-image: radial-gradient(circle at 30% 20%, rgba(255,255,255,0.4), transparent 50%), radial-gradient(circle at 80% 80%, rgba(0,0,0,0.6), transparent 60%);"
            ></div>
          {/if}
          <div class="absolute inset-x-0 bottom-0 h-1/2 bg-gradient-to-t from-black/85 to-transparent"></div>
          <div class="absolute inset-x-3 bottom-3">
            <h3 class="font-display text-sm font-bold leading-tight text-white drop-shadow-md md:text-base" style="text-shadow: 0 2px 12px rgba(0,0,0,0.7);">
              {media.title}
            </h3>
            <div class="mt-0.5 text-[10px] uppercase tracking-widest text-white/70">
              {media.year} • {media.type}
            </div>
          </div>
        </div>
      </article>
    {/each}
    <div class="w-2 shrink-0"></div>
  </div>
</section>
