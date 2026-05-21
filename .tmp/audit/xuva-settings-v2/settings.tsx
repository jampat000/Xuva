import { createFileRoute } from "@tanstack/react-router";
import { useEffect, useRef, useState } from "react";
import {
  LayoutDashboard,
  User,
  Library,
  ScanSearch,
  Database,
  Play,
  Sliders,
  HardDrive,
  Wifi,
  ArrowRightLeft,
  Compass,
  Users,
  Link2,
  ShieldCheck,
  ClipboardList,
  Info,
  Check,
  KeyRound,
  Search,
} from "lucide-react";
import { Header } from "@/components/Header";

export const Route = createFileRoute("/settings")({
  head: () => ({
    meta: [
      { title: "Settings — Xuva" },
      {
        name: "description",
        content:
          "Configure your Xuva media server — metadata providers, libraries, playback, transcoding, and storage.",
      },
    ],
  }),
  component: SettingsPage,
});

type Group = "Media-Server" | "Devices" | "Advanced";
type Section = { id: string; label: string; icon: typeof User; group: Group; hint?: string };

const SECTIONS: Section[] = [
  { id: "dashboard", label: "Dashboard", icon: LayoutDashboard, group: "Media-Server", hint: "Overview & health" },
  { id: "general", label: "General", icon: User, group: "Media-Server", hint: "Server name, locale, theme" },
  { id: "libraries", label: "Libraries", icon: Library, group: "Media-Server", hint: "Folders, scanners, art" },
  { id: "scanning", label: "Scanning", icon: ScanSearch, group: "Media-Server", hint: "Cadence & rules" },
  { id: "metadata", label: "Metadata", icon: Database, group: "Media-Server", hint: "Providers, keys, matches" },
  { id: "playback", label: "Playback", icon: Play, group: "Media-Server", hint: "Streaming defaults" },
  { id: "transcoding", label: "Transcoding", icon: Sliders, group: "Media-Server", hint: "Hardware & quality" },
  { id: "storage", label: "Storage", icon: HardDrive, group: "Media-Server", hint: "Disks & cache" },
  { id: "network", label: "Network", icon: Wifi, group: "Media-Server", hint: "Ports, remote access" },
  { id: "migration", label: "Migration", icon: ArrowRightLeft, group: "Media-Server", hint: "Import from elsewhere" },
  { id: "discovery", label: "Discovery", icon: Compass, group: "Media-Server", hint: "Recommendations" },
  { id: "users", label: "Users", icon: Users, group: "Media-Server", hint: "Accounts & roles" },
  { id: "pairing", label: "Pairing", icon: Link2, group: "Devices", hint: "Link a new device" },
  { id: "approved", label: "Approved Devices", icon: ShieldCheck, group: "Devices", hint: "Trusted clients" },
  { id: "planned", label: "Planned", icon: ClipboardList, group: "Advanced", hint: "Coming soon" },
  { id: "about", label: "About", icon: Info, group: "Advanced", hint: "Build & licenses" },
];

type Mode = "balanced" | "automatic" | "advanced";

const MODES: { id: Mode; title: string; desc: string }[] = [
  {
    id: "balanced",
    title: "Balanced",
    desc: "Xuva uses its recommended metadata and artwork fallbacks for the best overall coverage.",
  },
  {
    id: "automatic",
    title: "Prefer local artwork",
    desc: "Keep metadata automatic but prefer artwork stored alongside your media before downloaded artwork.",
  },
  {
    id: "advanced",
    title: "Advanced provider settings",
    desc: "Choose exact providers, order, and keys when you need full control.",
  },
];

const PROVIDERS = [
  { id: "tmdb", name: "TMDB", blurb: "Movie, show, season, episode metadata and artwork" },
  { id: "tvdb", name: "TheTVDB", blurb: "TV and movie metadata, IDs, and ratings" },
  { id: "fanart", name: "Fanart.tv", blurb: "Logos, banners, thumbs, and extra artwork" },
  { id: "musicbrainz", name: "MusicBrainz", blurb: "Track, album, and artist metadata for music libraries" },
];

const GROUPS: Group[] = ["Media-Server", "Devices", "Advanced"];

function SettingsPage() {
  const [active, setActive] = useState("metadata");
  const [mode, setMode] = useState<Mode>("advanced");
  const [keys, setKeys] = useState<Record<string, string>>({});
  const [q, setQ] = useState("");
  const [headerScrolled, setHeaderScrolled] = useState(false);
  const mainRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const onScroll = () => setHeaderScrolled(window.scrollY > 80);
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  const current = SECTIONS.find((s) => s.id === active) ?? SECTIONS[4];
  const filtered = (g: Group) =>
    SECTIONS.filter((s) => s.group === g).filter(
      (s) => !q || s.label.toLowerCase().includes(q.toLowerCase())
    );

  const select = (id: string) => {
    setActive(id);
    mainRef.current?.scrollTo({ top: 0, behavior: "smooth" });
  };

  return (
    <div className="min-h-screen bg-background">
      <Header />

      <main className="px-6 pb-32 pt-24 md:px-12 md:pt-28 lg:px-20">
        {/* Compact editorial header — single line, no giant hero */}
        <header className="relative mb-10">
          <div
            aria-hidden
            className="pointer-events-none absolute -inset-x-6 -top-10 -z-10 h-[220px] opacity-60 md:-inset-x-12 lg:-inset-x-20"
            style={{
              background:
                "radial-gradient(50% 100% at 15% 0%, oklch(0.62 0.22 285 / 0.25), transparent 70%), radial-gradient(40% 100% at 90% 0%, oklch(0.72 0.16 255 / 0.18), transparent 70%)",
            }}
          />
          <div className="flex flex-wrap items-end justify-between gap-6">
            <div>
              <div className="mb-2 text-[10px] font-semibold uppercase tracking-[0.35em] text-primary-glow">
                Settings
              </div>
              <h1 className="font-serif-display text-[clamp(2rem,4vw,3.25rem)] leading-[1] tracking-tight">
                Media-<em>Server</em>
              </h1>
              <p className="mt-3 max-w-xl text-sm text-muted-foreground">
                Choose metadata sources, review matches, and tune how Xuva pulls
                artwork and information for your library.
              </p>
            </div>
            <div className="flex items-center gap-3">
              <span className="hairline inline-flex items-center gap-2 rounded-full bg-foreground/[0.04] px-3 py-1.5 text-[10px] font-semibold uppercase tracking-[0.25em] text-emerald-300">
                <span className="h-1.5 w-1.5 rounded-full bg-emerald-400 shadow-[0_0_10px_currentColor]" />
                Healthy
              </span>
              <div className="hidden text-right text-[11px] uppercase tracking-[0.22em] text-muted-foreground md:block">
                Updated <span className="text-foreground/90">10:39:30 am</span>
              </div>
            </div>
          </div>
        </header>

        {/* 2-column shell: sticky TOC + content panel */}
        <div className="grid gap-10 lg:grid-cols-[260px_minmax(0,1fr)] lg:gap-14">
          {/* TOC */}
          <aside className="lg:sticky lg:top-24 lg:self-start">
            <div className="relative mb-4">
              <Search className="pointer-events-none absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <input
                value={q}
                onChange={(e) => setQ(e.target.value)}
                placeholder="Search settings…"
                className="h-10 w-full rounded-full border border-border bg-surface/60 pl-10 pr-4 text-sm outline-none placeholder:text-muted-foreground/70 focus:border-primary/60 focus:bg-surface"
              />
            </div>

            <nav className="scrollbar-none -mx-1 max-h-[calc(100vh-12rem)] space-y-6 overflow-y-auto px-1">
              {GROUPS.map((g) => {
                const items = filtered(g);
                if (!items.length) return null;
                return (
                  <div key={g}>
                    <div className="px-2 pb-2 text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground/70">
                      {g}
                    </div>
                    <ul className="space-y-0.5">
                      {items.map((s) => {
                        const Icon = s.icon;
                        const isActive = s.id === active;
                        return (
                          <li key={s.id}>
                            <button
                              type="button"
                              onClick={() => select(s.id)}
                              className={`group/item relative flex w-full items-center gap-3 rounded-lg px-2.5 py-2 text-sm transition-all ${
                                isActive
                                  ? "bg-foreground/[0.06] text-foreground"
                                  : "text-muted-foreground hover:bg-foreground/[0.03] hover:text-foreground"
                              }`}
                            >
                              {isActive && (
                                <span className="absolute inset-y-1.5 left-0 w-0.5 rounded-full bg-primary-glow shadow-glow" />
                              )}
                              <Icon
                                className={`h-4 w-4 shrink-0 ${
                                  isActive ? "text-primary-glow" : "text-muted-foreground/80"
                                }`}
                              />
                              <span className="truncate">{s.label}</span>
                            </button>
                          </li>
                        );
                      })}
                    </ul>
                  </div>
                );
              })}
            </nav>
          </aside>

          {/* Content panel */}
          <div ref={mainRef} className="min-w-0">
            {/* Section header strip */}
            <div
              className={`sticky top-16 z-20 -mx-6 mb-8 flex items-center justify-between gap-4 border-b px-6 py-4 backdrop-blur-xl transition-colors md:top-18 md:-mx-12 md:px-12 lg:-mx-0 lg:px-0 ${
                headerScrolled
                  ? "border-border bg-background/80"
                  : "border-transparent bg-transparent"
              }`}
            >
              <div className="min-w-0">
                <div className="text-[10px] font-semibold uppercase tracking-[0.3em] text-muted-foreground/80">
                  {current.group}
                </div>
                <h2 className="font-serif-display mt-0.5 truncate text-2xl tracking-tight">
                  {current.label}
                </h2>
                {current.hint && (
                  <p className="mt-0.5 truncate text-xs text-muted-foreground">
                    {current.hint}
                  </p>
                )}
              </div>
              <div className="flex items-center gap-2">
                <button className="hairline rounded-full bg-foreground/[0.04] px-4 py-2 text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground transition-colors hover:bg-foreground/[0.08] hover:text-foreground">
                  Discard
                </button>
                <button className="rounded-full bg-gradient-primary px-5 py-2 text-xs font-semibold uppercase tracking-[0.18em] text-white shadow-glow ring-1 ring-white/20 transition hover:brightness-110">
                  Save changes
                </button>
              </div>
            </div>

            {/* Sections — render Metadata as the worked example; others are scaffold placeholders */}
            {active === "metadata" ? (
              <div className="space-y-12">
                <SettingBlock
                  title="Pick the level of control"
                  desc="Most libraries should stay on Automatic. Switch to Advanced only when you need exact providers, ordering, and keys."
                >
                  <div className="grid gap-3 md:grid-cols-3">
                    {MODES.map((m) => {
                      const isActive = mode === m.id;
                      return (
                        <button
                          key={m.id}
                          type="button"
                          onClick={() => setMode(m.id)}
                          className={`hairline group relative overflow-hidden rounded-2xl p-5 text-left transition-all duration-300 ${
                            isActive
                              ? "bg-surface-elevated/80 shadow-elev"
                              : "bg-surface/40 hover:bg-surface/70"
                          }`}
                        >
                          <div className="flex items-center justify-between gap-3">
                            <span
                              className={`flex h-4 w-4 items-center justify-center rounded-full border ${
                                isActive
                                  ? "border-primary-glow bg-primary-glow shadow-glow"
                                  : "border-border"
                              }`}
                            >
                              {isActive && <Check className="h-2.5 w-2.5 text-black" />}
                            </span>
                            {isActive && (
                              <span className="text-[10px] font-semibold uppercase tracking-[0.25em] text-primary-glow">
                                Selected
                              </span>
                            )}
                          </div>
                          <div className="mt-4 text-base font-semibold">{m.title}</div>
                          <p className="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">
                            {m.desc}
                          </p>
                        </button>
                      );
                    })}
                  </div>
                </SettingBlock>

                <SettingBlock
                  title="Provider keys"
                  desc="Missing keys stay visible and safe. Providers without a working key remain disabled until you add one."
                >
                  <div className="grid gap-3">
                    {PROVIDERS.map((p) => {
                      const value = keys[p.id] ?? "";
                      const hasKey = value.trim().length > 0;
                      return (
                        <article
                          key={p.id}
                          className="hairline rounded-2xl bg-surface/40 p-6 transition-colors hover:bg-surface/60"
                        >
                          <div className="flex flex-wrap items-start justify-between gap-5">
                            <div className="min-w-0 flex-1">
                              <div className="flex items-center gap-3">
                                <h3 className="font-serif-display text-xl tracking-tight">
                                  {p.name}
                                </h3>
                                {hasKey ? (
                                  <span className="hairline inline-flex items-center gap-1.5 rounded-full bg-emerald-400/10 px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.2em] text-emerald-300">
                                    <Check className="h-3 w-3" /> Active
                                  </span>
                                ) : (
                                  <span className="hairline inline-flex items-center gap-1.5 rounded-full bg-amber-400/10 px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.2em] text-amber-300">
                                    <KeyRound className="h-3 w-3" /> Key required
                                  </span>
                                )}
                              </div>
                              <p className="mt-1 text-sm text-muted-foreground">{p.blurb}</p>

                              <div className="mt-5 grid gap-3 sm:grid-cols-[1fr_auto]">
                                <div>
                                  <label className="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted-foreground">
                                    {p.name} API key
                                  </label>
                                  <input
                                    type="password"
                                    value={value}
                                    onChange={(e) =>
                                      setKeys((k) => ({ ...k, [p.id]: e.target.value }))
                                    }
                                    placeholder="Paste key"
                                    className="mt-2 h-11 w-full rounded-xl border border-border bg-background/40 px-4 text-sm outline-none placeholder:text-muted-foreground/60 focus:border-primary/60 focus:bg-background/70"
                                  />
                                </div>
                                <button
                                  type="button"
                                  onClick={() => setKeys((k) => ({ ...k, [p.id]: "" }))}
                                  disabled={!hasKey}
                                  className="hairline mt-2 self-end rounded-xl bg-foreground/[0.04] px-4 py-3 text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground transition-colors hover:bg-foreground/[0.08] hover:text-foreground disabled:cursor-not-allowed disabled:opacity-40 sm:mt-7"
                                >
                                  Clear key
                                </button>
                              </div>
                            </div>
                          </div>
                        </article>
                      );
                    })}
                  </div>
                </SettingBlock>
              </div>
            ) : (
              <Placeholder section={current} />
            )}
          </div>
        </div>
      </main>
    </div>
  );
}

function SettingBlock({
  title,
  desc,
  children,
}: {
  title: string;
  desc: string;
  children: React.ReactNode;
}) {
  return (
    <section className="grid gap-6 md:grid-cols-[280px_minmax(0,1fr)] md:gap-10">
      <div>
        <h3 className="font-serif-display text-lg tracking-tight">{title}</h3>
        <p className="mt-1.5 text-[13px] leading-relaxed text-muted-foreground">{desc}</p>
      </div>
      <div>{children}</div>
    </section>
  );
}

function Placeholder({ section }: { section: Section }) {
  const Icon = section.icon;
  return (
    <div className="hairline flex flex-col items-center justify-center rounded-3xl bg-surface/30 px-8 py-24 text-center">
      <div className="hairline flex h-14 w-14 items-center justify-center rounded-2xl bg-surface-elevated/70 text-primary-glow shadow-elev">
        <Icon className="h-6 w-6" />
      </div>
      <div className="font-serif-display mt-5 text-2xl tracking-tight">
        {section.label}
      </div>
      <p className="mt-2 max-w-sm text-sm text-muted-foreground">
        {section.hint ?? "This section is part of the Xuva settings surface."} —
        controls coming online soon.
      </p>
    </div>
  );
}
