import { useEffect, useState } from "react";
import { Play, Plus, Info, Volume2, VolumeX } from "lucide-react";
import heroImg from "@/assets/hero-featured.jpg";
import type { Media } from "@/lib/mock-data";

type Props = { slides: Media[] };

export function Hero({ slides }: Props) {
  const [idx, setIdx] = useState(0);
  const [muted, setMuted] = useState(true);
  const media = slides[idx];

  useEffect(() => {
    const t = setInterval(() => setIdx((i) => (i + 1) % slides.length), 9000);
    return () => clearInterval(t);
  }, [slides.length]);

  const backdrop = media.backdrop ?? heroImg;

  return (
    <section className="relative -mt-16 h-[92vh] min-h-[680px] w-full overflow-hidden">
      {/* Backdrop with Ken Burns */}
      <div className="absolute inset-0">
        <img
          key={`${idx}-${backdrop}`}
          src={backdrop}
          alt=""
          fetchPriority="high"
          className="animate-kenburns h-full w-full object-cover"
          width={1920}
          height={1080}
        />
        {/* Cinematic vignettes */}
        <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_center,transparent_30%,oklch(0.06_0.01_280/0.8)_100%)]" />
        <div className="absolute inset-0 bg-gradient-to-r from-background via-background/70 to-transparent" />
        <div className="absolute inset-x-0 bottom-0 h-2/3 bg-gradient-to-t from-background via-background/60 to-transparent" />
        <div className="absolute inset-x-0 top-0 h-32 bg-gradient-to-b from-background/80 to-transparent" />
      </div>

      {/* Film grain */}
      <div className="grain pointer-events-none absolute inset-0" />

      {/* Content */}
      <div className="relative flex h-full flex-col justify-between px-6 pb-20 pt-28 md:px-12 md:pb-24 lg:px-20 lg:pb-28">
        <div key={media.id} className="animate-fade-up max-w-4xl flex-1 flex flex-col justify-end">
          {/* Brand presents */}
          <div className="mb-6 flex items-center gap-3 text-[10px] font-medium uppercase tracking-[0.4em] text-muted-foreground">
            <span className="h-px w-8 bg-foreground/40" />
            Xuva Presents
          </div>

          {/* Title — serif display, sized to fit on one line at desktop */}
          <h1 className="font-serif-display whitespace-nowrap text-[clamp(2.75rem,7vw,6.5rem)] leading-[0.92] tracking-tight text-foreground">
            {media.title.split(" ").map((w, i, arr) => (
              <span key={i}>
                {i === arr.length - 1 ? <em className="italic text-foreground/95">{w}</em> : w}
                {i < arr.length - 1 && " "}
              </span>
            ))}
          </h1>

          {/* Metadata line */}
          <div className="mt-6 flex flex-wrap items-center gap-x-4 gap-y-2 text-[12px] font-medium uppercase tracking-[0.18em] text-muted-foreground">
            <span className="text-foreground/90">{media.year}</span>
            <span className="opacity-30">·</span>
            <span>{media.director ?? media.type}</span>
            <span className="opacity-30">·</span>
            <span>{media.runtime ?? `${media.seasons} Seasons`}</span>
            <span className="opacity-30">·</span>
            <span className="flex items-center gap-1.5 normal-case tracking-normal text-foreground/90">
              <span className="text-amber-300">★</span>
              {media.rating.toFixed(1)}
            </span>
            {media.badge && (
              <>
                <span className="opacity-30">·</span>
                <span className="rounded-sm border border-foreground/20 px-2 py-0.5 text-[10px] tracking-[0.2em] text-foreground/80">
                  {media.badge}
                </span>
              </>
            )}
          </div>

          {/* Synopsis — wider so it breaks less */}
          <p className="mt-5 max-w-3xl text-[15px] leading-relaxed text-foreground/75 md:text-base">
            {media.synopsis}
          </p>

          {/* Genre chips — restrained */}
          <div className="mt-5 flex flex-wrap items-center gap-x-3 text-[12px] uppercase tracking-[0.2em] text-muted-foreground">
            {media.genres.map((g, i) => (
              <span key={g} className="flex items-center gap-3">
                {g}
                {i < media.genres.length - 1 && <span className="opacity-30">·</span>}
              </span>
            ))}
          </div>

          {/* CTAs — Apple TV+ style: filled white + ghost */}
          <div className="mt-9 flex flex-wrap items-center gap-3">
            <button
              type="button"
              className="group inline-flex items-center gap-2.5 rounded-full bg-foreground px-7 py-3.5 text-sm font-semibold text-background transition-all hover:bg-foreground/90"
            >
              <Play className="h-4 w-4 fill-background" />
              Play
            </button>
            <button
              type="button"
              className="inline-flex items-center gap-2.5 rounded-full border border-foreground/15 bg-foreground/5 px-6 py-3.5 text-sm font-medium text-foreground backdrop-blur-md transition-colors hover:bg-foreground/10"
            >
              <Info className="h-4 w-4" />
              More info
            </button>
            <button
              type="button"
              aria-label="Add to list"
              className="hairline flex h-12 w-12 items-center justify-center rounded-full bg-foreground/5 text-foreground backdrop-blur-md transition-colors hover:bg-foreground/10"
            >
              <Plus className="h-5 w-5" />
            </button>
          </div>
        </div>

        {/* Bottom bar — slide dots + mute */}
        <div className="pointer-events-none absolute inset-x-6 bottom-8 flex items-center justify-between md:inset-x-12 lg:inset-x-20">
          <div className="pointer-events-auto flex items-center gap-2">
            {slides.map((s, i) => (
              <button
                key={s.id}
                type="button"
                onClick={() => setIdx(i)}
                aria-label={`Show ${s.title}`}
                className="group/dot flex items-center"
              >
                <span
                  className={`block h-[2px] transition-all duration-500 ${
                    i === idx ? "w-12 bg-foreground" : "w-6 bg-foreground/25 group-hover/dot:bg-foreground/50"
                  }`}
                />
              </button>
            ))}
          </div>
          <button
            type="button"
            onClick={() => setMuted((m) => !m)}
            aria-label={muted ? "Unmute" : "Mute"}
            className="pointer-events-auto hairline flex h-10 w-10 items-center justify-center rounded-full bg-background/40 text-foreground/80 backdrop-blur-md transition-colors hover:text-foreground"
          >
            {muted ? <VolumeX className="h-4 w-4" /> : <Volume2 className="h-4 w-4" />}
          </button>
        </div>
      </div>
    </section>
  );
}
