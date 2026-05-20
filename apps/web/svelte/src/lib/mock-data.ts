export type Media = {
  id: string;
  title: string;
  year: number;
  type: "Movie" | "Series" | "Live";
  director?: string;
  genres: string[];
  rating: number;
  runtime?: string;
  runtimeMins?: number;    // parsed minutes for sorting
  seasons?: number;
  episodes?: number;
  progress?: number;
  synopsis: string;
  poster?: string;          // 2:3 portrait, primary card art (TMDB + Fanart)
  backdrop?: string;        // 16:9 widescreen, hero/detail bg (TMDB + Fanart)
  logo?: string;            // transparent PNG, clearlogo treatment (TMDB + Fanart)
  thumbnail?: string;       // landscape still, episode/scene art (TMDB StillPath + Fanart)
  banner?: string;          // ultra-wide marquee art, Fanart only
  videoKey?: string;        // YouTube key (fallback if no local trailer cached)
  trailerUrl?: string;      // local self-hosted MP4 path, preferred over YouTube
  parentId?: string;        // For Continue Watching: id of the parent movie/series
  parentKind?: string;      // For Continue Watching: "movie" or "series"
  contentRating?: string;   // Parental rating: G, PG, PG-13, R, NC-17, TV-MA, etc.
  needsReview?: boolean;    // Flagged as needing metadata review
  versionCount?: number;    // Number of media files for this title
  palette: [string, string];
  accent: string;
  badge?: string;
};

export type Collection = {
  id: string;
  title: string;
  count: number;
  palette: [string, string];
  accent: string;
  span?: "wide" | "tall" | "default";
};

export const featured: Media = {
  id: "the-burbs",
  title: "The 'Burbs",
  year: 1989,
  type: "Movie",
  director: "Joe Dante",
  genres: ["Comedy", "Mystery", "Thriller"],
  rating: 7.1,
  runtime: "1h 41m",
  badge: "Director's Cut - 4K HDR",
  synopsis:
    "An overstressed suburbanite and his neighbors become convinced that the new family in town are part of a satanic cult.",
  palette: ["#3a1f6b", "#a23b8f"],
  accent: "#c084fc"
};

export const spotlightSlides: Media[] = [
  featured,
  {
    id: "blade-runner",
    title: "Blade Runner 2049",
    year: 2017,
    type: "Movie",
    director: "Denis Villeneuve",
    genres: ["Sci-Fi", "Drama"],
    rating: 8,
    runtime: "2h 44m",
    badge: "Dolby Vision - Atmos",
    synopsis:
      "A young blade runner's discovery of a long-buried secret leads him on a quest to find Rick Deckard.",
    palette: ["#7a3b1f", "#d97706"],
    accent: "#fb923c"
  },
  {
    id: "past-lives",
    title: "Past Lives",
    year: 2023,
    type: "Movie",
    director: "Celine Song",
    genres: ["Romance", "Drama"],
    rating: 7.9,
    runtime: "1h 45m",
    badge: "Awards Season",
    synopsis:
      "Two childhood friends, reunited in New York after twenty years, confront notions of love and destiny.",
    palette: ["#581c87", "#be185d"],
    accent: "#f9a8d4"
  }
];

export const continueWatching: Media[] = [
  {
    id: "and-the-band",
    title: "And the Band Played On",
    year: 1993,
    type: "Movie",
    genres: ["Drama", "History"],
    rating: 7.7,
    runtime: "2h 21m",
    progress: 0.09,
    synopsis: "",
    palette: ["#1e3a5f", "#6b2d5c"],
    accent: "#60a5fa"
  },
  {
    id: "quiet-place-2",
    title: "A Quiet Place Part II",
    year: 2021,
    type: "Movie",
    genres: ["Horror"],
    rating: 7.2,
    runtime: "1h 37m",
    progress: 0.01,
    synopsis: "",
    palette: ["#1a1f2e", "#5b1f1f"],
    accent: "#f87171"
  },
  {
    id: "civil-action",
    title: "A Civil Action",
    year: 1998,
    type: "Movie",
    genres: ["Drama"],
    rating: 6.6,
    runtime: "1h 55m",
    progress: 0.34,
    synopsis: "",
    palette: ["#2d1f4a", "#8b5cf6"],
    accent: "#a78bfa"
  },
  {
    id: "blade-runner-cw",
    title: "Blade Runner 2049",
    year: 2017,
    type: "Movie",
    genres: ["Sci-Fi"],
    rating: 8,
    runtime: "2h 44m",
    progress: 0.62,
    synopsis: "",
    palette: ["#7a3b1f", "#d97706"],
    accent: "#fb923c"
  },
  {
    id: "terminal-list",
    title: "The Terminal List",
    year: 2022,
    type: "Series",
    genres: ["Thriller"],
    rating: 8,
    seasons: 1,
    episodes: 8,
    progress: 0.45,
    synopsis: "",
    palette: ["#1e3a8a", "#0f172a"],
    accent: "#60a5fa"
  }
];

export const topTen: Media[] = [
  { id: "t1", title: "Dune: Part Two", year: 2024, type: "Movie", genres: ["Sci-Fi"], rating: 8.5, runtime: "2h 46m", synopsis: "", palette: ["#78350f", "#ea580c"], accent: "#fdba74" },
  { id: "t2", title: "Oppenheimer", year: 2023, type: "Movie", genres: ["Biography"], rating: 8.3, runtime: "3h 0m", synopsis: "", palette: ["#1c1917", "#7c2d12"], accent: "#fb923c" },
  { id: "t3", title: "The Bear", year: 2022, type: "Series", genres: ["Drama"], rating: 8.6, seasons: 3, episodes: 28, synopsis: "", palette: ["#450a0a", "#dc2626"], accent: "#fca5a5" },
  { id: "t4", title: "Severance", year: 2022, type: "Series", genres: ["Sci-Fi"], rating: 8.7, seasons: 2, episodes: 19, synopsis: "", palette: ["#0c4a6e", "#155e75"], accent: "#67e8f9" },
  { id: "t5", title: "Anatomy of a Fall", year: 2023, type: "Movie", genres: ["Drama"], rating: 7.7, runtime: "2h 32m", synopsis: "", palette: ["#1f2937", "#475569"], accent: "#cbd5e1" },
  { id: "t6", title: "Shogun", year: 2024, type: "Series", genres: ["Drama"], rating: 8.7, seasons: 1, episodes: 10, synopsis: "", palette: ["#3f0f0f", "#9b1c1c"], accent: "#fbbf24" },
  { id: "t7", title: "The Zone of Interest", year: 2023, type: "Movie", genres: ["Drama"], rating: 7.4, runtime: "1h 45m", synopsis: "", palette: ["#1a1a1a", "#404040"], accent: "#a3a3a3" },
  { id: "t8", title: "Poor Things", year: 2023, type: "Movie", genres: ["Drama"], rating: 7.8, runtime: "2h 21m", synopsis: "", palette: ["#581c87", "#7e22ce"], accent: "#d8b4fe" },
  { id: "t9", title: "True Detective: Night Country", year: 2024, type: "Series", genres: ["Thriller"], rating: 7.4, seasons: 1, episodes: 6, synopsis: "", palette: ["#0f172a", "#1e3a8a"], accent: "#60a5fa" },
  { id: "t10", title: "Slow Horses", year: 2022, type: "Series", genres: ["Thriller"], rating: 8.3, seasons: 4, episodes: 24, synopsis: "", palette: ["#1f1f1f", "#525252"], accent: "#a3a3a3" }
];

export const recentMovies: Media[] = [
  { id: "m1", title: "Cloverfield Lane", year: 2016, type: "Movie", genres: ["Thriller"], rating: 7.2, runtime: "1h 43m", synopsis: "", palette: ["#0f172a", "#1e3a8a"], accent: "#60a5fa" },
  { id: "m2", title: "10,000 BC", year: 2008, type: "Movie", genres: ["Adventure"], rating: 5.1, runtime: "1h 49m", synopsis: "", palette: ["#3f2417", "#a16207"], accent: "#fbbf24" },
  { id: "m3", title: "12 Angry Men", year: 1957, type: "Movie", genres: ["Drama"], rating: 9, runtime: "1h 36m", synopsis: "", palette: ["#1f1f1f", "#dc2626"], accent: "#fbbf24" },
  { id: "m4", title: "12 Feet Deep", year: 2017, type: "Movie", genres: ["Thriller"], rating: 5.6, runtime: "1h 25m", synopsis: "", palette: ["#0c4a6e", "#0369a1"], accent: "#38bdf8" },
  { id: "m5", title: "12 Strong", year: 2018, type: "Movie", genres: ["War"], rating: 6.6, runtime: "2h 10m", synopsis: "", palette: ["#451a03", "#b45309"], accent: "#f59e0b" },
  { id: "m6", title: "Past Lives", year: 2023, type: "Movie", genres: ["Romance"], rating: 7.9, runtime: "1h 45m", synopsis: "", palette: ["#581c87", "#be185d"], accent: "#f9a8d4" },
  { id: "m7", title: "Killers of the Flower Moon", year: 2023, type: "Movie", genres: ["Crime"], rating: 7.6, runtime: "3h 26m", synopsis: "", palette: ["#3f1f0a", "#92400e"], accent: "#fb923c" },
  { id: "m8", title: "The Holdovers", year: 2023, type: "Movie", genres: ["Comedy"], rating: 7.9, runtime: "2h 13m", synopsis: "", palette: ["#1c1917", "#78350f"], accent: "#fbbf24" }
];

export const recentSeries: Media[] = [
  { id: "s1", title: "911: Lone Star", year: 2020, type: "Series", genres: ["Drama"], rating: 7.4, seasons: 5, episodes: 72, synopsis: "", palette: ["#7f1d1d", "#f97316"], accent: "#fb923c" },
  { id: "s2", title: "90 Day Fiance", year: 2014, type: "Series", genres: ["Reality"], rating: 6.2, seasons: 11, episodes: 220, synopsis: "", palette: ["#831843", "#be185d"], accent: "#f472b6" },
  { id: "s3", title: "The Day of the Jackal", year: 2024, type: "Series", genres: ["Thriller"], rating: 8.1, seasons: 1, episodes: 10, synopsis: "", palette: ["#450a0a", "#991b1b"], accent: "#fca5a5" },
  { id: "s4", title: "Shogun", year: 2024, type: "Series", genres: ["Drama"], rating: 8.7, seasons: 1, episodes: 10, synopsis: "", palette: ["#3f0f0f", "#9b1c1c"], accent: "#fbbf24" },
  { id: "s5", title: "Severance", year: 2022, type: "Series", genres: ["Sci-Fi"], rating: 8.7, seasons: 2, episodes: 19, synopsis: "", palette: ["#0c4a6e", "#155e75"], accent: "#67e8f9" },
  { id: "s6", title: "The Bear", year: 2022, type: "Series", genres: ["Drama"], rating: 8.6, seasons: 3, episodes: 28, synopsis: "", palette: ["#450a0a", "#dc2626"], accent: "#fca5a5" },
  { id: "s7", title: "Slow Horses", year: 2022, type: "Series", genres: ["Thriller"], rating: 8.3, seasons: 4, episodes: 24, synopsis: "", palette: ["#1f1f1f", "#525252"], accent: "#a3a3a3" },
  { id: "s8", title: "The Terminal List", year: 2022, type: "Series", genres: ["Action"], rating: 8, seasons: 1, episodes: 8, synopsis: "", palette: ["#1e3a8a", "#0f172a"], accent: "#60a5fa" }
];

export const collections: Collection[] = [
  { id: "c1", title: "Neo-Noir Essentials", count: 24, palette: ["#0f172a", "#7c2d12"], accent: "#fb923c", span: "wide" },
  { id: "c2", title: "Studio Ghibli", count: 21, palette: ["#0c4a6e", "#0e7490"], accent: "#67e8f9" },
  { id: "c3", title: "A24 Anthology", count: 38, palette: ["#1c1917", "#7f1d1d"], accent: "#fca5a5", span: "tall" },
  { id: "c4", title: "Wes Anderson", count: 12, palette: ["#7c2d12", "#facc15"], accent: "#fef08a" },
  { id: "c5", title: "Cinema of the 90s", count: 47, palette: ["#581c87", "#a21caf"], accent: "#f0abfc" },
  { id: "c6", title: "Late-Night Horror", count: 33, palette: ["#0a0a0a", "#450a0a"], accent: "#f87171", span: "wide" }
];

export const liveNow: Media[] = [
  { id: "l1", title: "MUBI Curated", year: 2026, type: "Live", genres: ["Channel"], rating: 0, synopsis: "", palette: ["#0a0a0a", "#262626"], accent: "#fbbf24", badge: "ON AIR" },
  { id: "l2", title: "Criterion 24", year: 2026, type: "Live", genres: ["Channel"], rating: 0, synopsis: "", palette: ["#1e3a8a", "#0f172a"], accent: "#60a5fa", badge: "ON AIR" },
  { id: "l3", title: "Studio Classics", year: 2026, type: "Live", genres: ["Channel"], rating: 0, synopsis: "", palette: ["#3f1f0a", "#92400e"], accent: "#fb923c", badge: "ON AIR" },
  { id: "l4", title: "Indie Lab", year: 2026, type: "Live", genres: ["Channel"], rating: 0, synopsis: "", palette: ["#581c87", "#be185d"], accent: "#f9a8d4", badge: "ON AIR" }
];
