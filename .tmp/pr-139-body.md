## Summary
- `altRecords` were fetched from `getMetadataRecords()` and stored in `$state` on both movie and TV detail pages, but there was no `{#each}` render loop — so the Metadata section showed nothing even when TMDB/Fanart data was available
- Added a "Data sources" list that renders each `MetadataRecord` with a provider badge, mini poster, title/year, and one-line overview
- The list appears whenever `altRecords.length > 0`, above the existing "Fix match" panel

## Test plan
- [ ] Open any movie detail page that has metadata → scroll to the Metadata section → "Data sources (N)" list should appear with provider badges (TMDB, Fanart, etc.)
- [ ] Open any TV series detail page — same result
- [ ] Open a movie with no metadata yet → Data sources section is hidden (no empty state clutter)
- [ ] Verify poster thumbnails load; broken image fallback shows Film/Tv icon

Closes #139

🤖 Generated with [Claude Code](https://claude.com/claude-code)
