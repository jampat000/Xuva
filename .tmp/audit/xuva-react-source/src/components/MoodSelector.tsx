import { useState } from "react";

const MOODS = [
  { id: "all", label: "All" },
  { id: "comfort", label: "Comfort watch", hint: "Easy, warm, familiar" },
  { id: "edge", label: "On the edge", hint: "Thrillers & tension" },
  { id: "mind", label: "Mind-bending", hint: "Sci-fi & puzzles" },
  { id: "tears", label: "A good cry", hint: "Drama & romance" },
  { id: "laugh", label: "Need a laugh", hint: "Comedies" },
  { id: "epic", label: "Epic & grand", hint: "Long, beautiful, immersive" },
  { id: "dark", label: "Late & dark", hint: "Horror & noir" },
];

export function MoodSelector() {
  const [active, setActive] = useState("all");
  const current = MOODS.find((m) => m.id === active);

  return (
    <section className="px-6 md:px-12 lg:px-20">
      <div className="mb-6 flex items-end justify-between gap-6">
        <div>
          <div className="mb-1.5 text-[10px] font-semibold uppercase tracking-[0.3em] text-primary-glow">
            Tonight, I'm in the mood for…
          </div>
          <h2 className="font-serif-display text-3xl tracking-tight md:text-4xl">
            What's the vibe?
          </h2>
        </div>
        {current?.hint && active !== "all" && (
          <div className="hidden text-right text-sm text-muted-foreground md:block">
            {current.hint}
          </div>
        )}
      </div>

      <div className="scrollbar-none -mx-6 flex gap-2 overflow-x-auto px-6 md:mx-0 md:flex-wrap md:px-0">
        {MOODS.map((m) => {
          const isActive = m.id === active;
          return (
            <button
              key={m.id}
              type="button"
              onClick={() => setActive(m.id)}
              className={`hairline shrink-0 rounded-full px-5 py-2.5 text-sm font-medium transition-all ${
                isActive
                  ? "bg-foreground text-background"
                  : "bg-foreground/[0.04] text-foreground/80 hover:bg-foreground/[0.08] hover:text-foreground"
              }`}
            >
              {m.label}
            </button>
          );
        })}
      </div>
    </section>
  );
}
