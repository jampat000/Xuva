import { useMemo, useState } from "react";
import { Search, SlidersHorizontal, Play, Plus } from "lucide-react";
import type { Media } from "@/lib/mock-data";

type Props = {
  eyebrow: string;
  title: string;
  tagline: string;
  items: Media[];
  /** field label shown on cards (e.g. "Movies" or "Series") */
  kind: "Movies" | "Series";
};

type Density = "S" | "M" | "L";
type Sort = "trending" | "rating" | "year-desc" | "az";

const DENSITY_GRID: Record<Density, string> = {
  S: "grid-cols-3 sm:grid-cols-5 md:grid-cols-7 lg:grid-cols-8 xl:grid-cols-10",
  M: "grid-cols-2 sm:grid-cols-3 md:grid-cols-5 lg:grid-cols-6 xl:grid-cols-7",
  L: "grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-4 xl:grid-cols-5",
};

export function LibraryGrid({ eyebrow, title, tagline, items, kind }: Props) {
  const [q, setQ] = useState("");
  const [genre, setGenre] = useState<string>("All");
  const [sort, setSort] = useState<Sort>("trending");
  const [density, setDensity] = useState<Density>("M");

  const genres = useMemo(() => {
    const s = new Set<string>();
    items.forEach((i) => i.genres.forEach((g) => s.add(g)));
    return ["All", ...Array.from(s).sort()];
  }, [items]);

  const filtered = useMemo(() => {
    let list = items.filter((m) => {
      const matchesGenre = genre === "All" || m.genres.includes(genre);
      const matchesQ = !q || m.title.toLowerCase().includes(q.toLowerCase());
      return matchesGenre && matchesQ;
    });
    const sorters: Record<Sort, (a: Media, b: Media) => number> = {
      trending: (a, b) => b.rating - a.rating,
      rating: (a, b) => b.rating - a.rating,
      "year-desc": (a, b) => b.year - a.year,
      az: (a, b) => a.title.localeCompare(b.title),
    };
    return [...list].sort(sorters[sort]);
  }, [items, genre, q, sort]);

  const featured = filtered[0] ?? items[0];

  return (
    <main className="pb-32">
      {/* Editorial header */}
      <section className="relative overflow-hidden px-6 pb-10 pt-32 md:px-12 md:pb-14 md:pt-40 lg:px-20">
        <div
          aria-hidden
          className="pointer-events-none absolute inset-x-0 top-0 -z-10 h-[520px] opacity-70"
          style={{
            background: featured
              ? `radial-gradient(60% 70% at 20% 0%, ${featured.palette[0]}55, transparent 60%), radial-gradient(50% 60% at 90% 10%, ${featured.palette[1]}40, transparent 70%)`
              : undefined,
          }}
        />
        <div className="grain pointer-events-none absolute inset-0 -z-10" />

        <div className="grid items-end gap-10 md:grid-cols-[1.6fr_1fr] lg:grid-cols-[2fr_1fr]">
          <div>
            <div className="mb-3 text-[10px] font-semibold uppercase tracking-[0.35em] text-primary-glow">
              {eyebrow}
            </div>
            <h1 className="font-serif-display text-[clamp(2.6rem,6vw,5rem)] leading-[0.95] tracking-tight">
              {title}
            </h1>
            <p className="mt-5 max-w-xl text-sm leading-relaxed text-muted-foreground md:text-base">
              {tagline}
            </p>
            <div className="mt-6 flex flex-wrap gap-x-6 gap-y-2 text-[11px] uppercase tracking-[0.22em] text-muted-foreground">
              <span><span className="text-foreground/90">{items.length}</span> in library</span>
              <span className="opacity-30">·</span>
              <span><span className="text-foreground/90">{genres.length - 1}</span> genres</span>
              <span className="opacity-30">·</span>
              <span><span className="text-foreground/90">4K · HDR · Atmos</span></span>
            </div>
          </div>

          {featured && (
            <article
              className="shadow-poster relative ml-auto aspect-[2/3] w-full max-w-[260px] overflow-hidden rounded-2xl md:max-w-[280px] lg:max-w-[320px]"
              style={{
                background: `linear-gradient(135deg, ${featured.palette[0]}, ${featured.palette[1]})`,
              }}
            >
              {featured.poster ? (
                <img
                  src={featured.poster}
                  alt={featured.title}
                  className="absolute inset-0 h-full w-full object-cover"
                />
              ) : (
                <div
                  className="absolute inset-0 opacity-40 mix-blend-overlay"
                  style={{
                    backgroundImage:
                      "radial-gradient(circle at 25% 15%, rgba(255,255,255,0.45), transparent 50%), radial-gradient(circle at 80% 85%, rgba(0,0,0,0.6), transparent 60%)",
                  }}
                />
              )}
              <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-black/10 to-transparent" />
              <div className="absolute inset-0 flex flex-col justify-between p-5">
                <div className="flex items-center gap-2 text-[9px] font-semibold uppercase tracking-[0.3em] text-white/85">
                  <span
                    className="h-1.5 w-1.5 rounded-full"
                    style={{ background: featured.accent, boxShadow: `0 0 12px ${featured.accent}` }}
                  />
                  Editor's pick
                </div>
                <div>
                  <div className="text-[10px] uppercase tracking-[0.28em] text-white/65">
                    {featured.year} · {featured.runtime ?? `${featured.seasons} Seasons`}
                  </div>
                  <h2
                    className="font-serif-display mt-1.5 text-2xl leading-[0.95] text-white md:text-3xl"
                    style={{ textShadow: "0 4px 24px rgba(0,0,0,0.5)" }}
                  >
                    {featured.title}
                  </h2>
                  <div className="mt-4 flex gap-2">
                    <button className="inline-flex items-center gap-1.5 rounded-full bg-white px-4 py-2 text-xs font-semibold text-black hover:bg-white/90">
                      <Play className="h-3.5 w-3.5 fill-black" /> Play
                    </button>
                    <button
                      aria-label="Add to list"
                      className="hairline flex h-9 w-9 items-center justify-center rounded-full bg-white/10 text-white backdrop-blur-md hover:bg-white/20"
                    >
                      <Plus className="h-4 w-4" />
                    </button>
                  </div>
                </div>
              </div>
            </article>
          )}
        </div>
      </section>

      {/* Sticky filter bar */}
      <div className="sticky top-16 z-30 -mb-px border-y border-border bg-background/75 backdrop-blur-xl md:top-18">
        <div className="flex flex-wrap items-center gap-3 px-6 py-4 md:px-12 lg:px-20">
          <div className="relative w-full max-w-xs">
            <Search className="pointer-events-none absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <input
              value={q}
              onChange={(e) => setQ(e.target.value)}
              placeholder={`Search ${kind.toLowerCase()}…`}
              className="h-10 w-full rounded-full border border-border bg-surface/60 pl-10 pr-4 text-sm outline-none placeholder:text-muted-foreground/70 focus:border-primary/60 focus:bg-surface"
            />
          </div>

          <div className="scrollbar-none -mx-2 flex flex-1 items-center gap-1.5 overflow-x-auto px-2">
            {genres.map((g) => {
              const active = g === genre;
              return (
                <button
                  key={g}
                  onClick={() => setGenre(g)}
                  className={`hairline shrink-0 rounded-full px-4 py-1.5 text-xs font-medium uppercase tracking-[0.15em] transition-colors ${
                    active
                      ? "bg-foreground text-background"
                      : "bg-foreground/[0.04] text-foreground/75 hover:bg-foreground/[0.08] hover:text-foreground"
                  }`}
                >
                  {g}
                </button>
              );
            })}
          </div>

          <div className="flex items-center gap-2">
            <div className="hairline flex items-center gap-1 rounded-full bg-foreground/[0.04] px-2 py-1 text-xs text-muted-foreground">
              <SlidersHorizontal className="ml-1 h-3.5 w-3.5" />
              <select
                value={sort}
                onChange={(e) => setSort(e.target.value as Sort)}
                className="cursor-pointer appearance-none bg-transparent px-2 py-1 text-foreground outline-none"
              >
                <option value="trending">Trending</option>
                <option value="rating">Highest rated</option>
                <option value="year-desc">Newest first</option>
                <option value="az">A → Z</option>
              </select>
            </div>
            <div className="hairline hidden items-center rounded-full bg-foreground/[0.04] p-1 md:flex">
              {(["S", "M", "L"] as Density[]).map((d) => (
                <button
                  key={d}
                  type="button"
                  aria-label={`${d === "S" ? "Small" : d === "M" ? "Medium" : "Large"} cards`}
                  onClick={() => setDensity(d)}
                  className={`flex h-7 min-w-7 items-center justify-center rounded-full px-2.5 text-[11px] font-semibold tracking-wider transition-colors ${
                    density === d ? "bg-foreground text-background" : "text-muted-foreground hover:text-foreground"
                  }`}
                >
                  {d}
                </button>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Grid */}
      <section className="px-6 pt-10 md:px-12 lg:px-20">
        {filtered.length === 0 ? (
          <div className="hairline flex flex-col items-center justify-center rounded-3xl bg-surface/30 py-24 text-center">
            <div className="font-serif-display text-3xl">Nothing matches that</div>
            <p className="mt-2 max-w-sm text-sm text-muted-foreground">
              Try a different genre or clear your search to see your full library.
            </p>
          </div>
        ) : (
          <div className={`grid gap-x-5 gap-y-10 ${DENSITY_GRID[density]}`}>
            {filtered.map((m) => (
              <article key={m.id} className="group cursor-pointer">
                <div
                  className="shadow-poster relative aspect-[2/3] overflow-hidden rounded-xl transition-all duration-500 group-hover:-translate-y-1.5 group-hover:shadow-glow"
                  style={{ background: `linear-gradient(135deg, ${m.palette[0]}, ${m.palette[1]})` }}
                >
                  {m.poster ? (
                    <img
                      src={m.poster}
                      alt={m.title}
                      loading="lazy"
                      className="absolute inset-0 h-full w-full object-cover"
                    />
                  ) : (
                    <div
                      className="absolute inset-0 opacity-40 mix-blend-overlay"
                      style={{
                        backgroundImage:
                          "radial-gradient(circle at 30% 20%, rgba(255,255,255,0.4), transparent 50%), radial-gradient(circle at 80% 80%, rgba(0,0,0,0.55), transparent 60%)",
                      }}
                    />
                  )}
                  <div className="absolute inset-x-0 bottom-0 h-1/2 bg-gradient-to-t from-black/80 to-transparent" />
                  <div className="absolute inset-x-3 bottom-3">
                    <h3
                      className="font-display text-[13px] font-semibold leading-tight text-white md:text-sm"
                      style={{ textShadow: "0 2px 12px rgba(0,0,0,0.7)" }}
                    >
                      {m.title}
                    </h3>
                  </div>
                  <div className="absolute inset-0 flex items-center justify-center bg-black/35 opacity-0 backdrop-blur-[1px] transition-opacity duration-300 group-hover:opacity-100">
                    <button
                      aria-label={`Play ${m.title}`}
                      className="flex h-12 w-12 items-center justify-center rounded-full bg-white text-black ring-1 ring-white/40"
                    >
                      <Play className="h-5 w-5 translate-x-0.5 fill-black" />
                    </button>
                  </div>
                </div>
                {density !== "S" && (
                  <>
                    <div className="mt-3 flex items-baseline justify-between gap-2 px-0.5">
                      <h4 className="truncate text-sm font-medium text-foreground">{m.title}</h4>
                      <span className="shrink-0 text-[11px] tabular-nums text-muted-foreground">
                        {m.year}
                      </span>
                    </div>
                    <p className="mt-0.5 truncate text-[11px] uppercase tracking-[0.18em] text-muted-foreground">
                      {m.genres.slice(0, 2).join(" · ")}
                      {typeof m.rating === "number" && m.rating > 0 && (
                        <span className="ml-2 normal-case tracking-normal text-foreground/70">
                          ★ {m.rating.toFixed(1)}
                        </span>
                      )}
                    </p>
                  </>
                )}
              </article>
            ))}
          </div>
        )}
      </section>
    </main>
  );
}
