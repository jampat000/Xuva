import { createFileRoute } from "@tanstack/react-router";
import { Header } from "@/components/Header";
import { LibraryGrid } from "@/components/LibraryGrid";
import { recentSeries, topTen } from "@/lib/mock-data";

const all = [
  ...recentSeries,
  ...topTen.filter((t) => t.type === "Series"),
].filter((m, i, arr) => arr.findIndex((x) => x.id === m.id) === i);

export const Route = createFileRoute("/series")({
  head: () => ({
    meta: [
      { title: "Series — Xuva" },
      { name: "description", content: "Your series library on Xuva — track what's next, queue up the season, and find the one you'll binge tonight." },
      { property: "og:title", content: "Series — Xuva" },
      { property: "og:description", content: "Your series library on Xuva — track what's next, queue up the season, and find the one you'll binge tonight." },
    ],
  }),
  component: SeriesPage,
});

function SeriesPage() {
  return (
    <div className="min-h-screen bg-background">
      <Header />
      <LibraryGrid
        eyebrow="Your library · Series"
        title="Stories told in seasons."
        tagline="Pick up mid-episode, queue the next season, or fall down a rabbit hole — your full series shelf, beautifully laid out."
        items={all}
        kind="Series"
      />
    </div>
  );
}
