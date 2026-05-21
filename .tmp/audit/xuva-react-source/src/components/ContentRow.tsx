import { useRef, type ReactNode } from "react";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { PosterCard } from "./PosterCard";
import type { Media } from "@/lib/mock-data";

type Props = {
  title: ReactNode;
  eyebrow?: string;
  items: Media[];
  variant?: "poster" | "wide";
};

export function ContentRow({ title, eyebrow, items, variant = "poster" }: Props) {
  const scrollerRef = useRef<HTMLDivElement>(null);

  const scroll = (dir: "l" | "r") => {
    const el = scrollerRef.current;
    if (!el) return;
    const amount = el.clientWidth * 0.8;
    el.scrollBy({ left: dir === "l" ? -amount : amount, behavior: "smooth" });
  };

  return (
    <section className="group/row relative">
      <div className="mb-5 flex items-end justify-between px-6 md:px-12 lg:px-20">
        <div>
          {eyebrow && (
            <div className="mb-1.5 text-[10px] font-semibold uppercase tracking-[0.3em] text-primary-glow">
              {eyebrow}
            </div>
          )}
          <h2 className="font-serif-display text-3xl tracking-tight md:text-4xl">
            {title}
          </h2>
        </div>
        <div className="hidden gap-1.5 opacity-0 transition-opacity group-hover/row:opacity-100 md:flex">
          <button
            type="button"
            onClick={() => scroll("l")}
            className="hairline flex h-9 w-9 items-center justify-center rounded-full bg-background/60 backdrop-blur transition-colors hover:bg-surface-elevated"
            aria-label="Scroll left"
          >
            <ChevronLeft className="h-4 w-4" />
          </button>
          <button
            type="button"
            onClick={() => scroll("r")}
            className="hairline flex h-9 w-9 items-center justify-center rounded-full bg-background/60 backdrop-blur transition-colors hover:bg-surface-elevated"
            aria-label="Scroll right"
          >
            <ChevronRight className="h-4 w-4" />
          </button>
        </div>
      </div>
      <div
        ref={scrollerRef}
        className="scrollbar-none flex gap-4 overflow-x-auto scroll-smooth px-6 pb-6 md:gap-5 md:px-12 lg:px-20"
      >
        {items.map((m) => (
          <PosterCard key={m.id} media={m} variant={variant} />
        ))}
        <div className="w-2 shrink-0" />
      </div>
    </section>
  );
}
