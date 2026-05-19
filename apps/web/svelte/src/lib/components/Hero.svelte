<script lang="ts">
  import { goto } from "$app/navigation";
  import { fade } from "svelte/transition";
  import { Info, Play, Plus, Volume2, VolumeX } from "lucide-svelte";
  import heroImg from "$lib/assets/hero-featured.jpg";
  import type { Media } from "$lib/mock-data";

  let { slides: rawSlides } = $props<{ slides: Media[] }>();

  function detailHref(m: Media): string {
    return m.type === 'Series' ? `/tv/${m.id}` : `/movies/${m.id}`;
  }

  // Prefer slides that have a real backdrop. Fall back to all slides only if
  // nothing in the set has one — that keeps the hero cinematic instead of
  // flashing the bundled placeholder JPG between titles.
  let slides = $derived.by(() => {
    const withBackdrop = rawSlides.filter((s: Media) => !!s.backdrop);
    return withBackdrop.length > 0 ? withBackdrop : rawSlides;
  });

  let idx = $state(0);
  let muted = $state(true);
  let media = $derived(slides[idx] ?? slides[0]);
  let backdrop = $derived(media?.backdrop || heroImg);
  let words = $derived((media?.title ?? "").split(" "));

  // ── Trailer playback ────────────────────────────────────────────────────
  // Two-tier strategy:
  //   1. If the slide has a `trailerUrl` (local self-hosted MP4), render a
  //      native <video> — no ads, no branding, no iframe, perfect loop.
  //   2. Else if it has a `videoKey` (YouTube fallback), render a muted
  //      youtube-nocookie iframe with the same fade-in behavior.
  //   3. Else, the backdrop image alone owns the hero.
  //
  // Either way we crossfade in over the backdrop after 1.5 s so the user
  // perceives the still poster frame first, then the trailer "wakes up".
  let trailerVisible = $state(false);
  let videoEl = $state<HTMLVideoElement | null>(null);
  let iframeEl = $state<HTMLIFrameElement | null>(null);

  const hasLocalTrailer = $derived(!!media?.trailerUrl);

  // YouTube embed can be blocked for two reasons:
  //   • The owner disabled embedding (error 101/150)
  //   • The video is age-restricted (also 101/150 in embedded context)
  // When either fires we set this flag, hide the iframe, and let the
  // backdrop image take over — user never sees the YouTube error wall.
  let youtubeEmbedBlocked = $state(false);

  // Reset the blocked flag whenever the slide changes so the next slide
  // gets a fresh attempt even if the previous one was blocked.
  $effect(() => {
    void media?.id;
    youtubeEmbedBlocked = false;
  });

  // Listen for YouTube IFrame API postMessage errors. enablejsapi=1 is
  // already in the embed URL so these events are always emitted.
  $effect(() => {
    if (typeof window === "undefined") return;
    function onMessage(e: MessageEvent) {
      try {
        const data = typeof e.data === "string" ? JSON.parse(e.data) : e.data;
        // 101 / 150 = embedding not allowed (includes age-restricted content).
        // 2 = invalid videoId — shouldn't occur but handle defensively.
        if (data?.event === "onError" && (data?.info === 101 || data?.info === 150 || data?.info === 2)) {
          youtubeEmbedBlocked = true;
          trailerVisible = false;
        }
      } catch { /* non-JSON messages from other origins — safe to ignore */ }
    }
    window.addEventListener("message", onMessage);
    return () => window.removeEventListener("message", onMessage);
  });

  const hasYouTubeFallback = $derived(!media?.trailerUrl && !!media?.videoKey && !youtubeEmbedBlocked);

  // Reset trailer state when slide changes; fade it in after a short hold.
  $effect(() => {
    void media?.id;
    trailerVisible = false;
    if (typeof window === "undefined") return;
    if (!media?.trailerUrl && !media?.videoKey) return;
    const wakeup = window.setTimeout(() => {
      trailerVisible = true;
      // Sync mute state to whichever player is mounted.
      if (videoEl) videoEl.muted = muted;
    }, 1500);
    return () => window.clearTimeout(wakeup);
  });

  // Slide rotation — respects trailer length instead of a fixed interval.
  //
  // Per-slide schedule:
  //   • Local MP4 trailer → wait for video metadata, then advance after
  //     (duration + 2 s buffer), clamped to [20 s, 90 s]. We let the trailer
  //     play through once. With `loop` on the <video>, short trailers loop
  //     during the buffer rather than freezing on the last frame.
  //   • YouTube fallback → 45 s (we can't introspect iframe playback without
  //     the YT IFrame API; 45 s is roughly one trailer length).
  //   • No trailer → 12 s (snappy carousel of static backdrops).
  //
  // Re-runs whenever the slide changes; cleans up its timer on next change.
  const SLIDE_MIN_MS = 20_000;
  const SLIDE_LOCAL_CAP_MS = 90_000;
  const SLIDE_YOUTUBE_MS = 45_000;
  const SLIDE_STATIC_MS = 12_000;

  $effect(() => {
    void media?.id; // re-run on slide change
    if (typeof window === "undefined" || slides.length <= 1) return;

    let timeoutId: number | null = null;
    const advance = () => {
      idx = (idx + 1) % slides.length;
    };
    const scheduleMs = (ms: number) => {
      if (timeoutId !== null) window.clearTimeout(timeoutId);
      timeoutId = window.setTimeout(advance, ms);
    };

    if (hasLocalTrailer) {
      // Need to wait for the video element to mount AND its metadata to
      // load before we know the real duration. Until then, fall back to
      // the cap so we never get stuck.
      scheduleMs(SLIDE_LOCAL_CAP_MS);
      const tryScheduleByDuration = () => {
        const v = videoEl;
        if (!v || !isFinite(v.duration) || v.duration <= 0) return false;
        const fullPlay = v.duration * 1000 + 2_000;
        scheduleMs(Math.min(SLIDE_LOCAL_CAP_MS, Math.max(SLIDE_MIN_MS, fullPlay)));
        return true;
      };
      if (!tryScheduleByDuration()) {
        // Poll briefly for the video element + metadata to arrive. The
        // {#key media.id} block remounts the video each slide change, so
        // the binding might not be ready at the moment this effect runs.
        const poll = window.setInterval(() => {
          if (tryScheduleByDuration()) window.clearInterval(poll);
        }, 250);
        // Stop polling after 3 s no matter what — we keep the cap timer.
        const stopPoll = window.setTimeout(() => window.clearInterval(poll), 3_000);
        return () => {
          window.clearInterval(poll);
          window.clearTimeout(stopPoll);
          if (timeoutId !== null) window.clearTimeout(timeoutId);
        };
      }
    } else if (hasYouTubeFallback) {
      scheduleMs(SLIDE_YOUTUBE_MS);
    } else {
      scheduleMs(SLIDE_STATIC_MS);
    }

    return () => {
      if (timeoutId !== null) window.clearTimeout(timeoutId);
    };
  });

  function toggleMute() {
    muted = !muted;
    if (videoEl) videoEl.muted = muted;
    // YouTube postMessage: { event, func, args }. We only send for iframe.
    const w = iframeEl?.contentWindow;
    if (w) {
      w.postMessage(
        JSON.stringify({ event: "command", func: muted ? "mute" : "unMute", args: [] }),
        "*"
      );
    }
  }

  // Build YouTube embed URL for the fallback path.
  function youtubeEmbedUrl(key: string): string {
    const origin = typeof window !== "undefined" ? encodeURIComponent(window.location.origin) : "";
    const params = [
      "autoplay=1",
      "mute=1",
      "controls=0",
      `loop=1`,
      `playlist=${encodeURIComponent(key)}`,
      "modestbranding=1",
      "playsinline=1",
      "rel=0",
      "iv_load_policy=3",
      "disablekb=1",
      "fs=0",
      "enablejsapi=1",
      `origin=${origin}`
    ].join("&");
    return `https://www.youtube-nocookie.com/embed/${encodeURIComponent(key)}?${params}`;
  }
</script>

<section class="relative -mt-16 h-[72vh] min-h-[500px] w-full overflow-hidden sm:h-[76vh] sm:min-h-[540px] lg:h-[80vh] lg:min-h-[580px]">
  <div class="absolute inset-0">
    <!-- Wrapper fades OUT when trailer fades IN (synchronized crossfade).
         The img inside uses Svelte fade for slide-to-slide transitions — two
         separate concerns, no CSS/JS opacity conflict. -->
    <div
      class={`absolute inset-0 transition-opacity duration-[2000ms] ${trailerVisible && (hasLocalTrailer || hasYouTubeFallback) ? 'opacity-0' : 'opacity-100'}`}
    >
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
    </div>

    {#if hasLocalTrailer}
      <!-- LOCAL TRAILER (preferred): native <video>, self-hosted MP4. No
           ads, no branding, perfect loop, supports Range requests for
           instant seek. The poster prop keeps the backdrop visible while
           the video buffers — zero flicker. -->
      {#key media.id}
        <div
          class={`absolute inset-0 transition-opacity duration-[2000ms] ${trailerVisible ? "opacity-100" : "opacity-0"}`}
          style="pointer-events: none;"
          aria-hidden="true"
        >
          <video
            bind:this={videoEl}
            src={media.trailerUrl}
            poster={media.backdrop}
            autoplay
            muted
            loop
            playsinline
            preload="auto"
            class="absolute inset-0 h-full w-full object-cover object-top"
          ></video>
        </div>
      {/key}
    {:else if hasYouTubeFallback}
      <!-- FALLBACK: YouTube iframe. Used until the local downloader catches
           up, or for items yt-dlp can't fetch. -->
      {#key media.id}
        <div
          class={`absolute inset-0 transition-opacity duration-[2000ms] ${trailerVisible ? "opacity-100" : "opacity-0"}`}
          style="pointer-events: none;"
          aria-hidden="true"
        >
          <iframe
            bind:this={iframeEl}
            src={youtubeEmbedUrl(media.videoKey!)}
            title="Hero trailer"
            allow="autoplay; encrypted-media"
            loading="lazy"
            class="absolute left-1/2 top-0 -translate-x-1/2"
            style="width: 177.78vh; min-width: 100%; height: 56.25vw; min-height: 100%;"
            frameborder="0"
          ></iframe>
        </div>
      {/key}
    {/if}

    <div class="absolute inset-0 bg-[radial-gradient(ellipse_at_center,transparent_30%,oklch(0.06_0.01_280/0.8)_100%)]"></div>
    <div class="absolute inset-0 bg-gradient-to-r from-background via-background/70 to-transparent"></div>
    <div class="absolute inset-x-0 bottom-0 h-2/3 bg-gradient-to-t from-background via-background/60 to-transparent"></div>
    <div class="absolute inset-x-0 top-0 h-32 bg-gradient-to-b from-background/80 to-transparent"></div>
  </div>

  <div class="grain pointer-events-none absolute inset-0"></div>

  <div class="relative flex h-full flex-col justify-between px-6 pb-14 pt-28 md:px-12 md:pb-18 lg:px-20 lg:pb-22">
    {#key media.id}
      <div class="animate-fade-up max-w-4xl flex-1 flex flex-col justify-end">
        <div class="mb-6 flex items-center gap-3 text-[10px] font-medium uppercase tracking-[0.4em] text-muted-foreground">
          <span class="h-px w-8 bg-foreground/40"></span>
          Xuva Presents
        </div>

        {#if media.logo}
          <!-- Clearlogo: SVG/PNG treatment matching Disney+/Netflix presentation -->
          <img
            src={media.logo}
            alt={media.title}
            class="max-h-24 w-auto max-w-[380px] object-contain drop-shadow-2xl md:max-h-28 lg:max-h-36"
            style="filter: drop-shadow(0 4px 24px oklch(0 0 0 / 0.6));"
          />
        {:else}
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
        {/if}

        <div class="mt-6 flex flex-wrap items-center gap-x-4 gap-y-2 text-[12px] font-medium uppercase tracking-[0.18em] text-muted-foreground">
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
                i === idx ? "w-12 bg-foreground" : "w-6 bg-foreground/25 group-hover/dot:bg-foreground/50"
              }`}
            ></span>
          </button>
        {/each}
      </div>
      {#if hasLocalTrailer || hasYouTubeFallback}
        <button
          type="button"
          onclick={toggleMute}
          aria-label={muted ? "Unmute trailer" : "Mute trailer"}
          class="pointer-events-auto hairline flex h-10 w-10 items-center justify-center rounded-full bg-background/40 text-foreground/80 backdrop-blur-md transition-colors hover:text-foreground"
        >
          {#if muted}
            <VolumeX class="h-4 w-4" />
          {:else}
            <Volume2 class="h-4 w-4" />
          {/if}
        </button>
      {/if}
    </div>
  </div>
</section>
