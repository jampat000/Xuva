import { createFileRoute } from "@tanstack/react-router";
import { Header } from "@/components/Header";
import { LibraryGrid } from "@/components/LibraryGrid";
import { featured, spotlightSlides, recentMovies, topTen } from "@/lib/mock-data";

const all = [
  featured,
  ...spotlightSlides.filter((s) => s.type === "Movie"),
  ...recentMovies,
  ...topTen.filter((t) => t.type === "Movie"),
].filter((m, i, arr) => arr.findIndex((x) => x.id === m.id) === i);

export const Route = createFileRoute("/movies")({
  head: () => ({
    meta: [
      { title: "Movies — Xuva" },
      { name: "description", content: "Your movie library on Xuva — sort, filter, and rediscover what's worth a night in." },
      { property: "og:title", content: "Movies — Xuva" },
      { property: "og:description", content: "Your movie library on Xuva — sort, filter, and rediscover what's worth a night in." },
    ],
  }),
  component: MoviesPage,
});

function MoviesPage() {
  return (
    <div className="min-h-screen bg-background">
      <Header />
      <LibraryGrid
        eyebrow="Your library · Movies"
        title="Films worth a night."
        tagline="Every movie in your collection, organized the way you actually browse — by mood, by genre, by what you almost watched last weekend."
        items={all}
        kind="Movies"
      />
    </div>
  );
}
