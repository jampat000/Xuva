import { createFileRoute } from "@tanstack/react-router";
import {
  Play,
  RotateCcw,
  Film,
  Check,
  Heart,
  MoreHorizontal,
  Search,
  Settings,
  Menu,
  ChevronLeft,
  Volume2,
  Captions,
  HardDrive,
  Star,
  ExternalLink,
} from "lucide-react";
import backdropImg from "@/assets/backdrop-speed.jpg";
import posterImg from "@/assets/poster-speed.jpg";

export const Route = createFileRoute("/")({
  component: MoviePage,
});

const cast = [
  { name: "Keanu Reeves", role: "Jack Traven", hue: 18 },
  { name: "Sandra Bullock", role: "Annie Porter", hue: 340 },
  { name: "Dennis Hopper", role: "Howard Payne", hue: 28 },
  { name: "Jeff Daniels", role: "Harry Temple", hue: 200 },
  { name: "Joe Morton", role: "Capt. McMahon", hue: 260 },
  { name: "Alan Ruck", role: "Stephens", hue: 50 },
  { name: "Glenn Plummer", role: "Jaguar Owner", hue: 130 },
  { name: "Beth Grant", role: "Helen", hue: 310 },
];

const related = [
  { title: "Die Hard", year: 1988, hue: 0 },
  { title: "Point Break", year: 1991, hue: 195 },
  { title: "The Matrix", year: 1999, hue: 130 },
  { title: "Aliens", year: 1986, hue: 30 },
  { title: "Heat", year: 1995, hue: 220 },
];

function MoviePage() {
  return (
    <div className="min-h-screen bg-background text-foreground">
      <TopBar />

      <main>
        {/* HERO */}
        <section className="relative">
          <div className="relative h-[72vh] w-full overflow-hidden">
            <img
              src={backdropImg}
              alt="Speed backdrop"
              width={1920}
              height={1080}
              className="absolute inset-0 w-full h-full object-cover animate-fade-in"
            />
            <div className="absolute inset-0 bg-gradient-to-r from-background via-background/85 to-background/40" />
            <div className="absolute inset-0 bg-gradient-to-t from-background via-background/40 to-transparent" />

            <div className="relative h-full max-w-[1400px] mx-auto px-8 pt-24 pb-12 flex items-end">
              <div className="grid grid-cols-12 gap-10 items-end w-full animate-fade-up">
                {/* Poster */}
                <div className="col-span-12 lg:col-span-3">
                  <div className="relative">
                    <div className="absolute -inset-2 bg-gradient-brand opacity-30 blur-2xl rounded-2xl" />
                    <img
                      src={posterImg}
                      alt="Speed poster"
                      width={800}
                      height={1216}
                      loading="lazy"
                      className="surface relative w-full max-w-[280px] aspect-[2/3] object-cover"
                    />
                  </div>
                </div>

                {/* Title & meta */}
                <div className="col-span-12 lg:col-span-9 space-y-6">
                  <button className="inline-flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground transition-colors">
                    <ChevronLeft className="size-3.5" /> Back to Movies
                  </button>

                  <div>
                    <h1 className="text-5xl md:text-6xl xl:text-7xl font-semibold tracking-tight mb-3">
                      Speed
                    </h1>
                    <div className="flex flex-wrap items-center gap-x-4 gap-y-2 text-sm text-muted-foreground">
                      <span className="inline-flex items-center gap-1.5 text-foreground">
                        <Star className="size-3.5 fill-cyan text-cyan" />
                        <span className="font-medium">7.3</span>
                      </span>
                      <Dot />
                      <span>1994</span>
                      <Dot />
                      <span>1h 55m</span>
                      <Dot />
                      <span className="px-1.5 py-0.5 rounded border border-border text-[10px] font-medium">
                        R
                      </span>
                      <Dot />
                      <span>Action · Thriller</span>
                    </div>
                  </div>

                  {/* Quality chips */}
                  <div className="flex flex-wrap gap-2">
                    <Chip primary>4K Dolby Vision</Chip>
                    <Chip>HEVC</Chip>
                    <Chip>Atmos 7.1</Chip>
                    <Chip>Direct Play</Chip>
                  </div>

                  <p className="max-w-2xl text-sm md:text-base text-muted-foreground leading-relaxed">
                    A young police officer must prevent a bomb exploding aboard a city
                    bus by keeping its speed above 50 mph. Tensions run high when a
                    crazed bomber rigs a Los Angeles bus with a device that will kill
                    everyone on board if the vehicle's speed dips below fifty.
                  </p>

                  {/* Actions */}
                  <div className="flex flex-wrap items-center gap-3 pt-2">
                    <button className="btn btn-primary px-6">
                      <Play className="size-4 fill-current" />
                      Resume
                    </button>
                    <button className="btn btn-secondary">
                      <RotateCcw className="size-4" />
                      Play From Start
                    </button>
                    <button className="btn btn-ghost">
                      <Film className="size-4" />
                      Trailer
                    </button>

                    <div className="flex items-center gap-1.5 ml-1">
                      <IconBtn label="Mark watched"><Check className="size-4" /></IconBtn>
                      <IconBtn label="Favorite"><Heart className="size-4" /></IconBtn>
                      <IconBtn label="More"><MoreHorizontal className="size-4" /></IconBtn>
                    </div>
                  </div>

                  {/* Progress */}
                  <div className="max-w-md pt-2">
                    <div className="flex justify-between text-[11px] text-muted-foreground mb-1.5">
                      <span><span className="text-cyan font-medium">36% watched</span> · 42:15</span>
                      <span>1h 13m left</span>
                    </div>
                    <div className="h-1 w-full bg-secondary rounded-full overflow-hidden">
                      <div className="h-full bg-gradient-progress rounded-full" style={{ width: "36%" }} />
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* DETAIL GRID */}
        <section className="max-w-[1400px] mx-auto px-8 py-16">
          <div className="grid grid-cols-12 gap-8">
            {/* Main column */}
            <div className="col-span-12 lg:col-span-8 space-y-14">
              {/* About */}
              <Section title="About" subtitle="Story, credits, and source media profile.">
                <div className="grid grid-cols-2 md:grid-cols-4 gap-x-6 gap-y-5 mb-8">
                  <Meta label="Director" value="Jan de Bont" />
                  <Meta label="Studio" value="20th Century Fox" />
                  <Meta label="Country" value="United States" />
                  <Meta label="Released" value="Jun 10, 1994" />
                </div>
                <p className="text-sm text-muted-foreground leading-relaxed max-w-3xl">
                  Jack Traven is a LAPD SWAT operative who is sent to diffuse a bomb
                  that a revenge-seeking extortionist has planted on a bus. Until Jack
                  can find a way to dismantle the device, he and the passengers driving
                  the bus must keep it racing through the streets of Los Angeles at
                  more than 50 mph — or it explodes.
                </p>
                <div className="flex flex-wrap gap-4 mt-6">
                  <ExtLink href="#">IMDb</ExtLink>
                  <ExtLink href="#">TheMovieDB</ExtLink>
                  <ExtLink href="#">Letterboxd</ExtLink>
                  <ExtLink href="#">Trakt</ExtLink>
                </div>
              </Section>

              {/* Cast */}
              <Section title="Cast & Crew" subtitle="Top billed cast and key roles.">
                <div className="flex gap-5 overflow-x-auto pb-4 no-scrollbar -mx-2 px-2">
                  {cast.map((c) => (
                    <div key={c.name} className="flex-none w-32 group cursor-pointer">
                      <div
                        className="surface surface-hover aspect-square rounded-full mb-3 overflow-hidden"
                        style={{
                          background: `linear-gradient(135deg, hsl(${c.hue} 40% 30%) 0%, hsl(${c.hue} 30% 14%) 100%)`,
                        }}
                      >
                        <div className="w-full h-full grid place-items-center text-2xl font-semibold text-foreground/40">
                          {c.name.split(" ").map((s) => s[0]).join("")}
                        </div>
                      </div>
                      <p className="text-sm font-medium text-foreground leading-tight text-center">
                        {c.name}
                      </p>
                      <p className="text-[11px] text-muted-foreground text-center mt-0.5">
                        {c.role}
                      </p>
                    </div>
                  ))}
                </div>
              </Section>

              {/* More like this */}
              <Section title="More Like This" subtitle="Picks based on this title.">
                <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 gap-5">
                  {related.map((r) => (
                    <div key={r.title} className="group cursor-pointer">
                      <div
                        className="surface surface-hover aspect-[2/3] mb-3 overflow-hidden relative"
                        style={{
                          background: `linear-gradient(160deg, hsl(${r.hue} 45% 32%) 0%, hsl(${r.hue} 35% 12%) 100%)`,
                        }}
                      >
                        <div className="absolute inset-0 grid place-items-center p-4 text-center">
                          <span className="text-lg font-semibold text-foreground/85 leading-tight">
                            {r.title}
                          </span>
                        </div>
                      </div>
                      <p className="text-sm font-medium">{r.title}</p>
                      <p className="text-[11px] text-muted-foreground">{r.year}</p>
                    </div>
                  ))}
                </div>
              </Section>
            </div>

            {/* Sidebar */}
            <aside className="col-span-12 lg:col-span-4 space-y-5 lg:sticky lg:top-24 self-start">
              <SpecCard
                icon={<Film className="size-3.5" />}
                title="Video"
                rows={[
                  ["Title", "4K Dolby Vision HEVC"],
                  ["Codec", "HEVC · Main 10"],
                  ["Dolby Profile", "Profile 8.1 (HDR10 compat)"],
                  ["Resolution", "3840 × 2160"],
                  ["Aspect", "16:9"],
                  ["Framerate", "23.976 fps"],
                  ["Bitrate", "63 Mbps"],
                  ["Color", "BT.2020 · SMPTE 2084"],
                  ["Bit Depth", "10 bit"],
                ]}
              />
              <SpecCard
                icon={<Volume2 className="size-3.5" />}
                title="Audio"
                rows={[
                  ["Title", "English DTS-HD MA 5.1"],
                  ["Codec", "DTS-HD MA"],
                  ["Channels", "5.1 · 6 ch"],
                  ["Sample Rate", "48,000 Hz"],
                  ["Bit Depth", "24 bit"],
                  ["Language", "English (Default)"],
                ]}
              />
              <SpecCard
                icon={<Captions className="size-3.5" />}
                title="Subtitles"
                rows={[
                  ["Title", "English (SRT)"],
                  ["Codec", "SRT"],
                  ["Language", "English"],
                  ["External", "Yes"],
                  ["Default", "No"],
                ]}
              />
              <SpecCard
                icon={<HardDrive className="size-3.5" />}
                title="File"
                rows={[
                  ["Container", "MKV"],
                  ["Size", "51.4 GB"],
                  ["Added", "Jan 26, 2026"],
                  ["Source", "Remux-2160p"],
                ]}
              >
                <div className="pt-3 mt-3 border-t border-border">
                  <p className="text-[10px] uppercase tracking-wider text-muted-foreground mb-1.5">
                    Path
                  </p>
                  <p className="text-[11px] font-mono text-foreground/70 break-all leading-relaxed">
                    Z:\Movies\Speed (1994)\Speed (1994) (Remux-2160p).mkv
                  </p>
                </div>
              </SpecCard>
            </aside>
          </div>
        </section>
      </main>
    </div>
  );
}

/* ── Components ─────────────────────────────────────────────── */

function TopBar() {
  return (
    <header className="fixed top-0 inset-x-0 z-50 h-16 px-6 flex items-center justify-between bg-background/60 backdrop-blur-xl border-b border-border/60">
      <div className="flex items-center gap-4">
        <button aria-label="Menu" className="btn-icon size-9">
          <Menu className="size-5" />
        </button>
        <div className="flex items-center gap-2">
          <div className="size-7 rounded-md bg-gradient-brand grid place-items-center shadow-brand">
            <Play className="size-3.5 fill-white text-white" />
          </div>
          <span className="text-lg font-semibold tracking-tight">
            <span className="text-gradient-brand">X</span>uva
          </span>
        </div>
      </div>

      <div className="flex-1 max-w-xl mx-8 hidden md:block">
        <div className="surface-glass h-9 px-4 rounded-full flex items-center gap-3 text-sm text-muted-foreground">
          <Search className="size-4" />
          <span>Search</span>
        </div>
      </div>

      <div className="flex items-center gap-2">
        <button aria-label="Settings" className="btn-icon size-9">
          <Settings className="size-4" />
        </button>
        <div className="size-8 rounded-full bg-gradient-brand grid place-items-center text-xs font-semibold text-white">
          A
        </div>
      </div>
    </header>
  );
}

function Dot() {
  return <span className="size-1 rounded-full bg-muted-foreground/40" />;
}

function Chip({ children, primary }: { children: React.ReactNode; primary?: boolean }) {
  return (
    <span
      className={
        "px-2.5 py-1 rounded-md text-[10px] font-semibold uppercase tracking-wider " +
        (primary
          ? "bg-primary/15 text-primary ring-1 ring-primary/30"
          : "bg-secondary/60 text-muted-foreground ring-1 ring-border")
      }
    >
      {children}
    </span>
  );
}

function IconBtn({
  children,
  label,
}: {
  children: React.ReactNode;
  label: string;
}) {
  return (
    <button aria-label={label} className="btn-icon">
      {children}
    </button>
  );
}

function Section({
  title,
  subtitle,
  children,
}: {
  title: string;
  subtitle?: string;
  children: React.ReactNode;
}) {
  return (
    <section>
      <div className="mb-6 pb-4 border-b border-border">
        <h2 className="text-xl font-semibold tracking-tight">{title}</h2>
        {subtitle ? (
          <p className="text-xs text-muted-foreground mt-1">{subtitle}</p>
        ) : null}
      </div>
      {children}
    </section>
  );
}

function Meta({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="text-[10px] uppercase tracking-wider text-muted-foreground mb-1">
        {label}
      </p>
      <p className="text-sm text-foreground">{value}</p>
    </div>
  );
}

function ExtLink({ href, children }: { href: string; children: React.ReactNode }) {
  return (
    <a
      href={href}
      className="inline-flex items-center gap-1.5 text-xs text-muted-foreground hover:text-primary transition-colors"
    >
      {children}
      <ExternalLink className="size-3" />
    </a>
  );
}

function SpecCard({
  icon,
  title,
  rows,
  children,
}: {
  icon: React.ReactNode;
  title: string;
  rows: [string, string][];
  children?: React.ReactNode;
}) {
  return (
    <div className="surface surface-hover p-5">
      <div className="flex items-center gap-2 mb-4">
        <span className="size-6 rounded-md bg-primary/15 text-primary grid place-items-center">
          {icon}
        </span>
        <h3 className="text-sm font-semibold">{title}</h3>
      </div>
      <div className="space-y-2.5">
        {rows.map(([k, v]) => (
          <div key={k} className="flex items-baseline justify-between gap-3 text-xs">
            <span className="text-muted-foreground">{k}</span>
            <span className="text-foreground text-right font-medium truncate">{v}</span>
          </div>
        ))}
      </div>
      {children}
    </div>
  );
}
