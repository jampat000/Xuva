<script lang="ts">
  import { ChevronLeft, ChevronRight } from "lucide-svelte";
  import type { Media } from "$lib/mock-data";
  import PosterCard from "./PosterCard.svelte";

  let { title, eyebrow, items, variant = "poster" } = $props<{
    title: string;
    eyebrow?: string;
    items: Media[];
    variant?: "poster" | "wide";
  }>();

  let scrollerRef = $state<HTMLDivElement | null>(null);

  function scroll(dir: "l" | "r"): void {
    const el = scrollerRef;
    if (!el) return;
    const amount = el.clientWidth * 0.8;
    el.scrollBy({ left: dir === "l" ? -amount : amount, behavior: "smooth" });
  }
</script>

<section class="group/row relative">
  <div class="mb-5 flex items-end justify-between px-6 md:px-12 lg:px-20">
    <div>
      {#if eyebrow}
        <div class="mb-1.5 text-[10px] font-semibold uppercase tracking-[0.3em] text-primary-glow">
          {eyebrow}
        </div>
      {/if}
      <h2 class="font-serif-display text-3xl tracking-tight md:text-4xl">
        {title}
      </h2>
    </div>
    <div class="hidden gap-1.5 opacity-0 transition-opacity group-hover/row:opacity-100 md:flex">
      <button
        type="button"
        onclick={() => scroll("l")}
        class="hairline flex h-9 w-9 items-center justify-center rounded-full bg-background/60 backdrop-blur transition-colors hover:bg-surface-elevated"
        aria-label="Scroll left"
      >
        <ChevronLeft class="h-4 w-4" />
      </button>
      <button
        type="button"
        onclick={() => scroll("r")}
        class="hairline flex h-9 w-9 items-center justify-center rounded-full bg-background/60 backdrop-blur transition-colors hover:bg-surface-elevated"
        aria-label="Scroll right"
      >
        <ChevronRight class="h-4 w-4" />
      </button>
    </div>
  </div>
  <div
    bind:this={scrollerRef}
    class="scrollbar-none flex gap-4 overflow-x-auto scroll-smooth px-6 pb-6 pt-4 md:gap-5 md:px-12 lg:px-20"
  >
    {#each items as media (media.id)}
      <PosterCard {media} {variant} />
    {/each}
    <div class="w-2 shrink-0"></div>
  </div>
</section>
