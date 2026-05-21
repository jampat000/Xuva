## Problem

The Movies and TV pages show a horizontal strip of genre chips for filtering. On medium-sized libraries this already overflows the viewport (chips are cut off on the right edge). There is no sort control at all. As the library grows this becomes a significant UX gap vs Plex / Emby / Jellyfin.

## Proposed solution

Replace the overflow chip strip with a collapsible **Filter & Sort** panel that sits above the grid. Selected filters collapse to a compact "active filters" badge row so they don't eat vertical space when not in use.

---

## Sort options (priority order)

| Option | Notes |
|---|---|
| **Title A–Z** (default) | Current implicit behaviour |
| **Title Z–A** | Reverse alpha |
| **Year — Newest first** | Release year desc |
| **Year — Oldest first** | Release year asc |
| **Date Added — Newest** | When item was scanned in |
| **Date Added — Oldest** | |
| **IMDb Rating — High to Low** | Needs `imdb_rating` in metadata |
| **Runtime — Shortest** | Needs `runtime_minutes` in metadata |
| **Runtime — Longest** | |
| **Random** | Shuffle on each page load / button press |
| **Parental Rating** | G → NC-17 / TV-Y → TV-MA order |

Deferred to a later phase (require probe data): Bitrate, Resolution, Video codec, Framerate.

---

## Filter dimensions

| Dimension | Type | Source |
|---|---|---|
| **Genre** | Multi-select chips (replace current strip) | Metadata `genres` array |
| **Year** | Range slider or decade buckets (e.g. 2020s, 2010s…) | `release_year` |
| **Parental Rating** | Multi-select (G, PG, PG-13, R, NC-17 / TV-Y … TV-MA) | `content_rating` |
| **Watched / Unwatched** | Toggle (All / Unwatched / Watched) | Play-state |
| **Studio** | Multi-select (top 20 by count) | Metadata `studio` field |
| **Resolution** | Multi-select (4K, 1080p, 720p, SD) | Probe data |
| **Video Codec** | Multi-select (H.265/HEVC, H.264, AV1, VP9…) | Probe data |
| **Audio Codec** | Multi-select (DTS, TrueHD, AAC, EAC3…) | Probe data |
| **Has Subtitles** | Toggle | Probe subtitle tracks |
| **Missing Metadata** | Toggle — surfaces items with no TMDB match | `metadata_records` |

Novel filters that competitors lack:
- **Needs Review** — matched but low-confidence score
- **Unprobed** — file was scanned but never ffprobed
- **Has Multiple Versions** — more than one media source for the same title

---

## Acceptance criteria

- [ ] Genre overflow is eliminated — chips either wrap gracefully or move into the panel
- [ ] Sort dropdown / segmented control is visible on Movies and TV pages
- [ ] Filter panel is accessible (keyboard-navigable, ARIA labels)
- [ ] Active filter count badge shows on the collapsed panel button
- [ ] Filters that require probe data are hidden when no probe data exists in the library
- [ ] URL reflects active sort + filters (query params) so links are shareable and browser Back works
- [ ] Performance: filtering stays client-side for libraries < 2k items; server-side pagination for larger libraries

## Implementation notes

- Sort: for now, fetch the full list and sort client-side (already the pattern for < 2k items). Add `?sort=` param to the catalog API for server-side once pagination (#91) lands.
- Filters: same pattern — client-side first, server-side query params in a follow-up.
- Resolution / codec filters require probe data to be populated — gate those chips on `hasProbeData` being true for at least some items.
- URL state: use SvelteKit `$page.url.searchParams` + `goto()` with `replaceState: true`.
