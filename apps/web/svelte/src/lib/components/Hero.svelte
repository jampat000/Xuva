<script lang="ts">
  import { goto } from "$app/navigation";
  import { Info, Play, Plus, Volume2, VolumeX } from "lucide-svelte";
  import heroImg from "$lib/assets/hero-featured.jpg";
  import type { Media } from "$lib/mock-data";

  let { slides } = $props<{ slides: Media[] }>();

  function detailHref(m: Media): string {
    return m.type === 'Series' ? `/tv/${m.id}` : `/movies/${m.id}`;
  }

  let idx = $state(0);
  let muted = $state(true);
  let media = $derived(slides[idx] ?? slides[0]);
  let backdrop = $derived(media?.backdrop ?? heroImg);
  let words = $derived((media?.title ?? "").split(" "));

  $effect(() => {
    if (typeof window === "undefined") return;
    const t = setInterval(() => {
      idx = (idx + 1) % slides.length;
    }, 9000);
    return () => clearInterval(t);
  });
</script>

<section class="relative -mt-16 h-[92vh] min-h-[680px] w-full overflow-hidden">
  <div class="absolute inset-0">
    {#key idx}
      <img
        src={backdrop}
        alt=""
        fetchpriority="high"
        class="animate-kenburns h-full w-full object-cover"
        width="1920"
        height="1080"
        onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
      />
    {/key}
    <div class="absolute inset-0 bg-[radial-gradient(ellipse_at_center,transparent_30%,oklch(0.06_0.01_280/0.8)_100%)]"></div>
    <div class="absolute inset-0 bg-gradient-to-r from-background via-background/70 to-transparent"></div>
    <div class="absolute inset-x-0 bottom-0 h-2/3 bg-gradient-to-t from-background via-background/60 to-transparent"></div>
    <div class="absolute inset-x-0 top-0 h-32 bg-gradient-to-b from-background/80 to-transparent"></div>
  </div>

  <div class="grain pointer-events-none absolute inset-0"></div>

  <div class="relative flex h-full flex-col justify-between px-6 pb-20 pt-28 md:px-12 md:pb-24 lg:px-20 lg:pb-28">
    {#key media.id}
      <div class="animate-fade-up max-w-4xl flex-1 flex flex-col justify-end">
        <div class="mb-6 flex items-center gap-3 text-[10px] font-medium uppercase tracking-[0.4em] text-muted-foreground">
          <span class="h-px w-8 bg-foreground/40"></span>
          Xuva Presents
        </div>

        <h1 class="font-serif-display whitespace-nowrap text-[clamp(2.75rem,7vw,6.5rem)] leading-[0.92] tracking-tight text-foreground">
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

        <div class="mt-6 flex flex-wrap items-center gap-x-4 gap-y-2 text-[12px] font-medium uppercase tracking-[0.18em] text-muted-foreground">
          <span class="text-foreground/90">{media.year}</span>
          <span class="opacity-30">·</span>
          <span>{media.director ?? media.type}</span>
          <span class="opacity-30">·</span>
          {#if media.type === 'Series' && media.seasons}
            <span>{media.seasons} Season{media.seasons !== 1 ? 's' : ''}</span>
          {:else if media.runtime}
            <span>{media.runtime}</span>
          {/if}
          <span class="opacity-30">·</span>
          <span class="flex items-center gap-1.5 normal-case tracking-normal text-foreground/90">
            <span class="text-amber-300">★</span>
            {media.rating.toFixed(1)}
          </span>
          {#if media.badge}
            <span class="opacity-30">·</span>
            <span class="rounded-sm border border-foreground/20 px-2 py-0.5 text-[10px] tracking-[0.2em] text-foreground/80">
              {media.badge}
            </span>
          {/if}
        </div>

        <p class="mt-5 max-w-3xl text-[15px] leading-relaxed text-foreground/75 md:text-base">
          {media.synopsis}
        </p>

        <div class="mt-5 flex flex-wrap items-center gap-x-3 text-[12px] uppercase tracking-[0.2em] text-muted-foreground">
          {#each media.genres as genre, i (genre)}
            <span class="flex items-center gap-3">
              {genre}
              {#if i < media.genres.length - 1}
                <span class="opacity-30">·</span>
              {/if}
            </span>
          {/each}
        </div>

        <div class="mt-9 flex flex-wrap items-center gap-3">
          <a
            href={detailHref(media)}
            class="group inline-flex items-center gap-2.5 rounded-full bg-foreground px-7 py-3.5 text-sm font-semibold text-background transition-all hover:bg-foreground/90"
          >
            <Play class="h-4 w-4 fill-background" />
            Play
          </a>
          <a
            href={detailHref(media)}
            class="inline-flex items-center gap-2.5 rounded-full border border-foreground/15 bg-foreground/5 px-6 py-3.5 text-sm font-medium text-foreground backdrop-blur-md transition-colors hover:bg-foreground/10"
          >
            <Info class="h-4 w-4" />
            More info
          </a>
          <button
            type="button"
            aria-label="Add to list"
            class="hairline flex h-12 w-12 items-center justify-center rounded-full bg-foreground/5 text-foreground backdrop-blur-md transition-colors hover:bg-foreground/10"
          >
            <Plus class="h-5 w-5" />
          </button>
        </div>
      </div>
    {/key}

    <div class="pointer-events-none absolute inset-x-6 bottom-8 flex items-center justify-between md:inset-x-12 lg:inset-x-20">
      <div class="pointer-events-auto flex items-center gap-2">
        {#each slides as slide, i (slide.id)}
          <button
            type="button"
            onclick={() => (idx = i)}
            aria-label={`Show ${slide.title}`}
            class="group/dot flex items-center"
          >
            <span
              class={`block h-[2px] transition-all duration-500 ${
                i === idx ? "w-12 bg-foreground" : "w-6 bg-foreground/25 group-hover/dot:bg-foreground/50"
              }`}
            ></span>
          </button>
        {/each}
      </div>
      <button
        type="button"
        onclick={() => (muted = !muted)}
        aria-label={muted ? "Unmute" : "Mute"}
        class="pointer-events-auto hairline flex h-10 w-10 items-center justify-center rounded-full bg-background/40 text-foreground/80 backdrop-blur-md transition-colors hover:text-foreground"
      >
        {#if muted}
          <VolumeX class="h-4 w-4" />
        {:else}
          <Volume2 class="h-4 w-4" />
        {/if}
      </button>
    </div>
  </div>
</section>
