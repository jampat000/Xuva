<script lang="ts">
  import { Radio } from "lucide-svelte";
  import type { Media } from "$lib/mock-data";

  let { items } = $props<{ items: Media[] }>();
</script>

<section class="px-6 md:px-12 lg:px-20">
  <div class="mb-5 flex items-end justify-between">
    <div>
      <div class="mb-1.5 flex items-center gap-2 text-[10px] font-semibold uppercase tracking-[0.3em] text-red-400">
        <span class="relative flex h-2 w-2">
          <span class="absolute inset-0 animate-ping rounded-full bg-red-500 opacity-75"></span>
          <span class="relative h-2 w-2 rounded-full bg-red-500"></span>
        </span>
        On Air Now
      </div>
      <h2 class="font-serif-display text-3xl tracking-tight md:text-4xl">
        Live <em class="italic opacity-70">channels</em>
      </h2>
    </div>
  </div>

  <div class="grid grid-cols-2 gap-3 md:grid-cols-4 md:gap-4">
    {#each items as item (item.id)}
      <article
        class="group relative aspect-[16/10] cursor-pointer overflow-hidden rounded-2xl"
        style={`background: linear-gradient(135deg, ${item.palette[0]}, ${item.palette[1]})`}
      >
        <div
          class="absolute inset-0 opacity-40 mix-blend-overlay"
          style="background-image: radial-gradient(circle at 30% 30%, rgba(255,255,255,0.4), transparent 50%);"
        ></div>
        <div class="absolute inset-0 flex flex-col justify-between p-5">
          <div class="flex w-fit items-center gap-1.5 rounded-full bg-red-500/90 px-2 py-0.5 text-[9px] font-bold uppercase tracking-[0.2em] text-white">
            <Radio class="h-2.5 w-2.5" />
            Live
          </div>
          <div>
            <h3 class="font-display text-lg font-bold leading-tight text-white md:text-xl" style="text-shadow: 0 2px 12px rgba(0,0,0,0.6);">
              {item.title}
            </h3>
            <div class="mt-1 text-[10px] uppercase tracking-widest text-white/70">
              Now: Featured film
            </div>
          </div>
        </div>
      </article>
    {/each}
  </div>
</section>
