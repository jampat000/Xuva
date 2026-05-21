import { Radio } from "lucide-react";
import type { Media } from "@/lib/mock-data";

export function LiveStrip({ items }: { items: Media[] }) {
  return (
    <section className="px-6 md:px-12 lg:px-20">
      <div className="mb-5 flex items-end justify-between">
        <div>
          <div className="mb-1.5 flex items-center gap-2 text-[10px] font-semibold uppercase tracking-[0.3em] text-red-400">
            <span className="relative flex h-2 w-2">
              <span className="absolute inset-0 animate-ping rounded-full bg-red-500 opacity-75" />
              <span className="relative h-2 w-2 rounded-full bg-red-500" />
            </span>
            On Air Now
          </div>
          <h2 className="font-serif-display text-3xl tracking-tight md:text-4xl">
            Live <em className="italic opacity-70">channels</em>
          </h2>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-3 md:grid-cols-4 md:gap-4">
        {items.map((c) => (
          <article
            key={c.id}
            className="group relative aspect-[16/10] cursor-pointer overflow-hidden rounded-2xl"
            style={{ background: `linear-gradient(135deg, ${c.palette[0]}, ${c.palette[1]})` }}
          >
            <div
              className="absolute inset-0 opacity-40 mix-blend-overlay"
              style={{ backgroundImage: "radial-gradient(circle at 30% 30%, rgba(255,255,255,0.4), transparent 50%)" }}
            />
            <div className="absolute inset-0 flex flex-col justify-between p-5">
              <div className="flex items-center gap-1.5 rounded-full bg-red-500/90 px-2 py-0.5 text-[9px] font-bold uppercase tracking-[0.2em] text-white w-fit">
                <Radio className="h-2.5 w-2.5" />
                Live
              </div>
              <div>
                <h3 className="font-display text-lg font-bold leading-tight text-white md:text-xl" style={{ textShadow: "0 2px 12px rgba(0,0,0,0.6)" }}>
                  {c.title}
                </h3>
                <div className="mt-1 text-[10px] uppercase tracking-widest text-white/70">
                  Now: Featured film
                </div>
              </div>
            </div>
          </article>
        ))}
      </div>
    </section>
  );
}
