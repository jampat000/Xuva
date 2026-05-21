import { Play, Star } from "lucide-react";
import type { Media } from "@/lib/mock-data";

type Props = {
  media: Media;
  /** "poster" = 2:3 portrait, "wide" = 16:9 landscape (for continue watching) */
  variant?: "poster" | "wide";
};

export function PosterCard({ media, variant = "poster" }: Props) {
  const isWide = variant === "wide";
  const gradient = `linear-gradient(135deg, ${media.palette[0]}, ${media.palette[1]})`;
  const art = isWide ? media.backdrop ?? media.poster : media.poster ?? media.backdrop;

  return (
    <article
      className={`group relative shrink-0 cursor-pointer ${
        isWide ? "w-[260px] md:w-[320px] lg:w-[360px]" : "w-[140px] md:w-[170px] lg:w-[180px]"
      }`}
    >
      <div
        className={`relative overflow-hidden rounded-2xl shadow-poster transition-all duration-500 group-hover:-translate-y-1 group-hover:shadow-glow ${
          isWide ? "aspect-video" : "aspect-[2/3]"
        }`}
        style={{ background: gradient }}
      >
        {art && (
          <img
            src={art}
            alt={media.title}
            loading="lazy"
            className="absolute inset-0 h-full w-full object-cover"
          />
        )}
        {/* Texture overlay (only when no art) */}
        {!art && (
          <div
            className="absolute inset-0 opacity-30 mix-blend-overlay"
            style={{
              backgroundImage:
                "radial-gradient(circle at 30% 20%, rgba(255,255,255,0.4), transparent 50%), radial-gradient(circle at 80% 80%, rgba(0,0,0,0.5), transparent 60%)",
            }}
          />
        )}
        {/* Diagonal sheen */}
        <div
          className="absolute -inset-1 opacity-0 transition-opacity duration-500 group-hover:opacity-100"
          style={{
            background:
              "linear-gradient(115deg, transparent 30%, rgba(255,255,255,0.18) 50%, transparent 70%)",
          }}
        />

        {/* Bottom shade so title is legible over real artwork */}
        {art && (
          <div className="absolute inset-x-0 bottom-0 h-2/3 bg-gradient-to-t from-black/85 via-black/30 to-transparent" />
        )}

        {/* Title block baked into "poster" */}
        <div className="absolute inset-0 flex flex-col justify-between p-4">
          <div className="flex items-center gap-1.5 text-[10px] font-medium uppercase tracking-widest text-white/80">
            <span
              className="h-1.5 w-1.5 rounded-full"
              style={{ background: media.accent, boxShadow: `0 0 10px ${media.accent}` }}
            />
            {media.genres[0]}
          </div>
          <div>
            <h3
              className="font-display text-lg font-bold leading-tight text-white drop-shadow-lg md:text-xl"
              style={{ textShadow: "0 2px 12px rgba(0,0,0,0.6)" }}
            >
              {media.title}
            </h3>
            <div className="mt-1 flex items-center gap-2 text-[11px] text-white/75">
              <span>{media.year}</span>
              <span className="opacity-50">•</span>
              <Star className="h-3 w-3 fill-current text-amber-300" />
              <span>{media.rating.toFixed(1)}</span>
            </div>
          </div>
        </div>

        {/* Hover play button */}
        <div className="absolute inset-0 flex items-center justify-center bg-black/30 opacity-0 backdrop-blur-[2px] transition-opacity duration-300 group-hover:opacity-100">
          <button
            type="button"
            className="flex h-14 w-14 items-center justify-center rounded-full bg-gradient-primary shadow-glow ring-1 ring-white/30 transition-transform hover:scale-110"
            aria-label={`Play ${media.title}`}
          >
            <Play className="h-6 w-6 translate-x-0.5 fill-white text-white" />
          </button>
        </div>

        {/* Progress bar for continue watching */}
        {typeof media.progress === "number" && (
          <div className="absolute inset-x-3 bottom-3">
            <div className="h-1 overflow-hidden rounded-full bg-white/20 backdrop-blur">
              <div
                className="h-full rounded-full bg-gradient-primary"
                style={{ width: `${Math.max(media.progress * 100, 3)}%` }}
              />
            </div>
          </div>
        )}
      </div>

      {/* Metadata under wide cards */}
      {isWide && (
        <div className="mt-3 px-1">
          <h4 className="truncate text-sm font-semibold text-foreground">{media.title}</h4>
          <p className="mt-0.5 text-xs text-muted-foreground">
            {media.year} · {media.type}
            {typeof media.progress === "number" && (
              <> · Resume from {Math.round(media.progress * 100)}%</>
            )}
          </p>
        </div>
      )}
      {!isWide && (
        <div className="mt-2 px-1">
          <p className="truncate text-xs text-muted-foreground">
            {media.type === "Series"
              ? `${media.seasons} season${(media.seasons ?? 0) > 1 ? "s" : ""} · ${media.episodes} ep`
              : media.runtime}
          </p>
        </div>
      )}
    </article>
  );
}
