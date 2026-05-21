import { useRef } from "react";
import { ChevronLeft, ChevronRight } from "lucide-react";
import type { Media } from "@/lib/mock-data";

export function Top10Row({ items }: { items: Media[] }) {
  const scrollerRef = useRef<HTMLDivElement>(null);
  const scroll = (dir: "l" | "r") => {
    const el = scrollerRef.current;
    if (!el) return;
    el.scrollBy({ left: dir === "l" ? -el.clientWidth * 0.8 : el.clientWidth * 0.8, behavior: "smooth" });
  };

  return (
    <section className="group/row relative">
      <div className="mb-5 flex items-end justify-between px-6 md:px-12 lg:px-20">
        <div>
          <div className="mb-1.5 text-[10px] font-semibold uppercase tracking-[0.3em] text-primary-glow">
            Trending this week
          </div>
          <h2 className="font-serif-display text-3xl tracking-tight md:text-4xl">
            Top 10 <em className="italic opacity-70">in your country</em>
          </h2>
        </div>
        <div className="hidden gap-1.5 opacity-0 transition-opacity group-hover/row:opacity-100 md:flex">
          <button type="button" onClick={() => scroll("l")} aria-label="Scroll left" className="hairline flex h-9 w-9 items-center justify-center rounded-full bg-background/60 backdrop-blur hover:bg-surface-elevated">
            <ChevronLeft className="h-4 w-4" />
          </button>
          <button type="button" onClick={() => scroll("r")} aria-label="Scroll right" className="hairline flex h-9 w-9 items-center justify-center rounded-full bg-background/60 backdrop-blur hover:bg-surface-elevated">
            <ChevronRight className="h-4 w-4" />
          </button>
        </div>
      </div>

      <div ref={scrollerRef} className="scrollbar-none flex gap-3 overflow-x-auto scroll-smooth px-6 pb-8 md:gap-4 md:px-12 lg:px-20">
        {items.map((m, i) => (
          <article key={m.id} className="group relative flex shrink-0 cursor-pointer items-end gap-1 md:gap-2">
            {/* Giant numeral */}
            <span
              className="font-serif-display select-none text-[8rem] leading-[0.78] tracking-tighter text-transparent md:text-[12rem]"
              style={{
                WebkitTextStroke: "1.5px oklch(0.45 0.04 280)",
                textShadow: "0 12px 40px oklch(0 0 0 / 0.5)",
                marginRight: "-0.35em",
                paddingLeft: i === 0 ? "0.1em" : 0,
              }}
            >
              {i + 1}
            </span>
            {/* Poster */}
            <div
              className="shadow-poster relative aspect-[2/3] w-[120px] shrink-0 overflow-hidden rounded-lg transition-all duration-500 group-hover:-translate-y-1.5 md:w-[160px]"
              style={{ background: `linear-gradient(135deg, ${m.palette[0]}, ${m.palette[1]})` }}
            >
              {m.poster && (
                <img
                  src={m.poster}
                  alt={m.title}
                  loading="lazy"
                  className="absolute inset-0 h-full w-full object-cover"
                />
              )}
              {!m.poster && (
                <div
                  className="absolute inset-0 opacity-40 mix-blend-overlay"
                  style={{ backgroundImage: "radial-gradient(circle at 30% 20%, rgba(255,255,255,0.4), transparent 50%), radial-gradient(circle at 80% 80%, rgba(0,0,0,0.6), transparent 60%)" }}
                />
              )}
              <div className="absolute inset-x-0 bottom-0 h-1/2 bg-gradient-to-t from-black/85 to-transparent" />
              <div className="absolute inset-x-3 bottom-3">
                <h3 className="font-display text-sm font-bold leading-tight text-white drop-shadow-md md:text-base" style={{ textShadow: "0 2px 12px rgba(0,0,0,0.7)" }}>
                  {m.title}
                </h3>
                <div className="mt-0.5 text-[10px] uppercase tracking-widest text-white/70">
                  {m.year} · {m.type}
                </div>
              </div>
            </div>
          </article>
        ))}
        <div className="w-2 shrink-0" />
      </div>
    </section>
  );
}
