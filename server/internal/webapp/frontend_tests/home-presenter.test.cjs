const test = require("node:test");
const assert = require("node:assert/strict");

const home = require("./modules/home-presenter.js");

test("home presenter cleans noisy movie filenames", () => {
  const item = home.presentHomeItem({
    id: "movie-1",
    kind: "movie",
    title: "The.Matrix.1999.1080p.BluRay.x264-Group.mkv",
    subtitle: "1999",
  });

  assert.equal(item.displayTitle, "The Matrix");
  assert.equal(item.displayYear, "1999");
  assert.equal(item.displaySubtitle, "1999");
  assert.equal(item.displayMeta, "1999 \u00B7 Movie");
});

test("home presenter extracts clean episode labels", () => {
  const item = home.presentHomeItem({
    id: "episode-1",
    kind: "episode",
    title: "The.Bear.S01E03.2160p.WEBRip.x265.mkv",
    subtitle: "Resume from 42%",
    percent: 0.42,
  });

  assert.equal(item.displayTitle, "The Bear");
  assert.equal(item.displaySubtitle, "S1 \u00B7 E3");
  assert.equal(item.progressPercent, 42);
  assert.equal(item.displayMeta, "S1 \u00B7 E3 \u00B7 Episode");
});

test("home presenter uses poster as backdrop fallback", () => {
  const item = home.presentHomeItem({
    id: "movie-2",
    kind: "movie",
    title: "Arrival",
    posterUrl: "https://example.com/poster.jpg",
  });

  assert.equal(item.posterUrl, "https://example.com/poster.jpg");
  assert.equal(item.backdropUrl, "https://example.com/poster.jpg");
  assert.equal(item.thumbUrl, "https://example.com/poster.jpg");
  assert.equal(item.usePosterBackdrop, true);
  assert.equal(item.hasArtwork, true);
});

test("home presenter keeps synthetic artwork marker for home-only placeholders", () => {
  const item = home.presentHomeItem({
    id: "movie-2b",
    kind: "movie",
    title: "Arrival",
  }, {
    posterUrl: "/api/artwork/movie/movie-2b?style=neutral&type=poster",
    syntheticArtworkOnly: true,
  });

  assert.equal(item.syntheticArtworkOnly, true);
});

test("home presenter can generate premium placeholder artwork", () => {
  const item = home.presentHomeItem({
    id: "movie-2c",
    kind: "movie",
    title: "12 Angry Men",
    subtitle: "1957",
  });

  const poster = home.buildPlaceholderArtwork(item, "poster");
  const hero = home.buildPlaceholderArtwork(item, "hero");

  assert.match(poster, /^data:image\/svg\+xml/);
  assert.match(hero, /^data:image\/svg\+xml/);
});

test("home presenter flags placeholder treatment when artwork is missing", () => {
  const item = home.presentHomeItem({
    id: "movie-3",
    kind: "movie",
    title: "Unknown.File.2024.WEBRip.mkv",
  });

  assert.equal(item.needsPlaceholder, true);
  assert.equal(item.hasArtwork, false);
  assert.equal(item.displayTitle, "Unknown File");
  assert.equal(item.displayYear, "2024");
});

test("home presenter cleans filename noise and encoding artifacts", () => {
  const item = home.presentHomeItem({
    id: "movie-4",
    kind: "movie",
    title: "'night, Mother (1986) (WEBRip-1080p).mp4",
    subtitle: "1986",
  });

  assert.equal(item.displayTitle, "'night, Mother");
  assert.equal(item.displaySubtitle, "1986");
});
