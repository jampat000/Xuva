<script lang="ts">
  import { ArrowUpRight } from "lucide-svelte";
  import type { Collection } from "$lib/mock-data";

  let { items } = $props<{ items: Collection[] }>();
</script>

<section class="px-6 md:px-12 lg:px-20">
  <div class="mb-6">
    <div class="mb-1.5 text-[10px] font-semibold uppercase tracking-[0.3em] text-primary-glow">
      Curated by Xuva
    </div>
    <h2 class="font-serif-display text-3xl tracking-tight md:text-4xl">
      Collections worth a <em class="italic opacity-70">long evening</em>
    </h2>
  </div>

  <div class="grid auto-rows-[170px] grid-cols-2 gap-3 md:grid-cols-4 md:gap-4">
    {#each items as collection (collection.id)}
      <article
        class={`group relative cursor-pointer overflow-hidden rounded-2xl ${
          collection.span === "wide"
            ? "md:col-span-2 md:row-span-1"
            : collection.span === "tall"
              ? "md:row-span-2"
              : ""
        }`}
        style={`background: linear-gradient(135deg, ${collection.palette[0]}, ${collection.palette[1]})`}
      >
        <div
          class="absolute inset-0 opacity-50 mix-blend-overlay transition-opacity duration-700 group-hover:opacity-70"
          style="background-image: radial-gradient(circle at 20% 20%, rgba(255,255,255,0.4), transparent 55%), radial-gradient(circle at 80% 80%, rgba(0,0,0,0.6), transparent 60%);"
        ></div>
        <div
          class="absolute -inset-1 opacity-0 transition-opacity duration-700 group-hover:opacity-100"
          style="background: linear-gradient(115deg, transparent 30%, rgba(255,255,255,0.12) 50%, transparent 70%);"
        ></div>
        <div class="absolute inset-0 flex flex-col justify-between p-5 md:p-6">
          <div class="flex items-start justify-between">
            <div class="text-[10px] font-semibold uppercase tracking-[0.3em] text-white/70">
              {collection.count} titles
            </div>
            <div class="hairline flex h-8 w-8 items-center justify-center rounded-full bg-background/30 text-white/80 backdrop-blur transition-transform group-hover:rotate-45">
              <ArrowUpRight class="h-4 w-4" />
            </div>
          </div>
          <h3
            class="font-serif-display text-2xl leading-tight text-white md:text-3xl"
            style="text-shadow: 0 2px 18px rgba(0,0,0,0.5);"
          >
            {collection.title}
          </h3>
        </div>
      </article>
    {/each}
  </div>
</section>
