import { createFileRoute } from "@tanstack/react-router";
import { Header } from "@/components/Header";
import { Hero } from "@/components/Hero";
import { ContentRow } from "@/components/ContentRow";
import { Top10Row } from "@/components/Top10Row";
import { CollectionsBento } from "@/components/CollectionsBento";
import { MoodSelector } from "@/components/MoodSelector";
import { Logo } from "@/components/Logo";
import {
  spotlightSlides,
  continueWatching,
  topTen,
  collections,
  recentMovies,
  recentSeries,
} from "@/lib/mock-data";

export const Route = createFileRoute("/")({
  head: () => ({
    meta: [
      { title: "Xuva — Your cinema, everywhere" },
      {
        name: "description",
        content:
          "Xuva is your personal streaming home — movies and series on web, mobile, tablet, tvOS, and Android. One library. Every screen.",
      },
    ],
  }),
  component: HomePage,
});

function HomePage() {
  return (
    <div className="min-h-screen bg-background">
      <Header />
      <main className="pb-24">
        <Hero slides={spotlightSlides} />

        <div className="relative z-10 -mt-12 space-y-16 md:-mt-16 md:space-y-20">
          <ContentRow
            eyebrow="Pick up where you left off"
            title="Continue watching"
            items={continueWatching}
            variant="wide"
          />

          <MoodSelector />

          <Top10Row items={topTen} />

          <CollectionsBento items={collections} />

          <ContentRow
            eyebrow="Fresh in your library"
            title="New movies"
            items={recentMovies}
          />

          <ContentRow
            eyebrow="New episodes dropped"
            title="New series"
            items={recentSeries}
          />
        </div>

        {/* Cross-device band */}
        <section className="relative mx-6 mt-28 overflow-hidden rounded-3xl border border-border bg-gradient-to-br from-surface/60 via-surface/30 to-background p-10 backdrop-blur md:mx-12 md:p-16 lg:mx-20">
          <div className="absolute -right-32 -top-32 h-[400px] w-[400px] rounded-full bg-primary/20 blur-[120px]" />
          <div className="absolute -bottom-32 -left-32 h-[400px] w-[400px] rounded-full bg-accent/20 blur-[120px]" />
          <div className="grain absolute inset-0" />
          <div className="relative grid items-center gap-12 md:grid-cols-[1.3fr_1fr]">
            <div>
              <div className="mb-4 text-[10px] font-semibold uppercase tracking-[0.35em] text-primary-glow">
                One library · Every screen
              </div>
              <h2 className="font-serif-display text-4xl leading-[1.05] tracking-tight md:text-6xl">
                Made for the couch, the commute, and everything between.
              </h2>
              <p className="mt-6 max-w-xl text-base leading-relaxed text-muted-foreground md:text-lg">
                Xuva adapts to every screen — a remote-friendly grid on tvOS and Android TV, a thumb-shaped feed on mobile, a sidebar-rich layout on tablet, and this cinematic surface on the web.
              </p>
            </div>
            <div className="grid grid-cols-2 gap-3 md:gap-4">
              {[
                { label: "Web", sub: "Cinematic" },
                { label: "Mobile", sub: "iOS · Android" },
                { label: "Tablet", sub: "iPadOS" },
                { label: "TV", sub: "tvOS · Android TV" },
              ].map((d) => (
                <div
                  key={d.label}
                  className="hairline rounded-2xl bg-background/40 p-5 backdrop-blur-md transition-all hover:-translate-y-1 hover:bg-background/60"
                >
                  <div className="text-[10px] uppercase tracking-[0.25em] text-muted-foreground">
                    {d.sub}
                  </div>
                  <div className="font-serif-display mt-2 text-2xl">
                    {d.label}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* Footer */}
        <footer className="mt-24 border-t border-border px-6 pt-12 md:px-12 lg:px-20">
          <div className="flex flex-col items-start justify-between gap-6 pb-10 md:flex-row md:items-center">
            <div>
              <Logo />
              <p className="mt-4 max-w-sm text-sm leading-relaxed text-muted-foreground">
                Your personal cinema. Stream your collection on every screen you own.
              </p>
            </div>
            <div className="flex flex-wrap gap-x-10 gap-y-3 text-sm text-muted-foreground">
              <a className="transition-colors hover:text-foreground" href="#">About</a>
              <a className="transition-colors hover:text-foreground" href="#">Apps</a>
              <a className="transition-colors hover:text-foreground" href="#">Support</a>
              <a className="transition-colors hover:text-foreground" href="#">Privacy</a>
              <a className="transition-colors hover:text-foreground" href="#">Terms</a>
            </div>
          </div>
          <div className="border-t border-border py-6 text-xs uppercase tracking-[0.2em] text-muted-foreground">
            © {new Date().getFullYear()} Xuva · Crafted for cinema lovers
          </div>
        </footer>
      </main>
    </div>
  );
}
