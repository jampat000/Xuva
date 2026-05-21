import { ArrowUpRight } from "lucide-react";
import type { Collection } from "@/lib/mock-data";

export function CollectionsBento({ items }: { items: Collection[] }) {
  return (
    <section className="px-6 md:px-12 lg:px-20">
      <div className="mb-6">
        <div className="mb-1.5 text-[10px] font-semibold uppercase tracking-[0.3em] text-primary-glow">
          Curated by Xuva
        </div>
        <h2 className="font-serif-display text-3xl tracking-tight md:text-4xl">
          Collections worth a <em className="italic opacity-70">long evening</em>
        </h2>
      </div>

      <div className="grid auto-rows-[170px] grid-cols-2 gap-3 md:grid-cols-4 md:gap-4">
        {items.map((c) => {
          const span =
            c.span === "wide"
              ? "md:col-span-2 md:row-span-1"
              : c.span === "tall"
                ? "md:row-span-2"
                : "";
          return (
            <article
              key={c.id}
              className={`group relative cursor-pointer overflow-hidden rounded-2xl ${span}`}
              style={{ background: `linear-gradient(135deg, ${c.palette[0]}, ${c.palette[1]})` }}
            >
              <div
                className="absolute inset-0 opacity-50 mix-blend-overlay transition-opacity duration-700 group-hover:opacity-70"
                style={{ backgroundImage: "radial-gradient(circle at 20% 20%, rgba(255,255,255,0.4), transparent 55%), radial-gradient(circle at 80% 80%, rgba(0,0,0,0.6), transparent 60%)" }}
              />
              {/* Sheen */}
              <div
                className="absolute -inset-1 opacity-0 transition-opacity duration-700 group-hover:opacity-100"
                style={{ background: "linear-gradient(115deg, transparent 30%, rgba(255,255,255,0.12) 50%, transparent 70%)" }}
              />
              <div className="absolute inset-0 flex flex-col justify-between p-5 md:p-6">
                <div className="flex items-start justify-between">
                  <div className="text-[10px] font-semibold uppercase tracking-[0.3em] text-white/70">
                    {c.count} titles
                  </div>
                  <div className="hairline flex h-8 w-8 items-center justify-center rounded-full bg-background/30 text-white/80 backdrop-blur transition-transform group-hover:rotate-45">
                    <ArrowUpRight className="h-4 w-4" />
                  </div>
                </div>
                <h3
                  className="font-serif-display text-2xl leading-tight text-white md:text-3xl"
                  style={{ textShadow: "0 2px 18px rgba(0,0,0,0.5)" }}
                >
                  {c.title}
                </h3>
              </div>
            </article>
          );
        })}
      </div>
    </section>
  );
}
