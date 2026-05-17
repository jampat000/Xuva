<script lang="ts">
  const moods = [
    { id: "all", label: "All" },
    { id: "comfort", label: "Comfort watch", hint: "Easy, warm, familiar" },
    { id: "edge", label: "On the edge", hint: "Thrillers & tension" },
    { id: "mind", label: "Mind-bending", hint: "Sci-fi & puzzles" },
    { id: "tears", label: "A good cry", hint: "Drama & romance" },
    { id: "laugh", label: "Need a laugh", hint: "Comedies" },
    { id: "epic", label: "Epic & grand", hint: "Long, beautiful, immersive" },
    { id: "dark", label: "Late & dark", hint: "Horror & noir" }
  ];

  let active = $state("all");
  let current = $derived(moods.find((mood) => mood.id === active));
</script>

<section class="px-6 md:px-12 lg:px-20">
  <div class="mb-6 flex items-end justify-between gap-6">
    <div>
      <div class="mb-1.5 text-[10px] font-semibold uppercase tracking-[0.3em] text-primary-glow">
        Tonight, I'm in the mood for...
      </div>
      <h2 class="font-serif-display text-3xl tracking-tight md:text-4xl">
        What's the vibe?
      </h2>
    </div>
    {#if current?.hint && active !== "all"}
      <div class="hidden text-right text-sm text-muted-foreground md:block">
        {current.hint}
      </div>
    {/if}
  </div>

  <div class="scrollbar-none -mx-6 flex gap-2 overflow-x-auto px-6 md:mx-0 md:flex-wrap md:px-0">
    {#each moods as mood (mood.id)}
      <button
        type="button"
        onclick={() => (active = mood.id)}
        class={`hairline shrink-0 rounded-full px-5 py-2.5 text-sm font-medium transition-all ${
          mood.id === active
            ? "bg-foreground text-background"
            : "bg-foreground/[0.04] text-foreground/80 hover:bg-foreground/[0.08] hover:text-foreground"
        }`}
      >
        {mood.label}
      </button>
    {/each}
  </div>
</section>
