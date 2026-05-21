## Summary
- Cast portraits were 72×72 px circles — faces were tiny and the circle crop often cut off hair/chin
- Upgraded to 96×144 px (2:3 ratio) portrait cards with rounded corners, matching the industry standard used by TMDB, Plex, and IMDb
- `object-top` keeps faces visible instead of centre-cropping
- Subtle hover lift (`scale-[1.03]` + `shadow-lg`) makes the strip feel interactive
- Fallback User icon scaled up from `h-7 w-7` to `h-10 w-10` to fill the larger frame
- Name text bumped from `font-medium` to `font-semibold` for better legibility at small size

Applies to both `routes/movies/[id]/+page.svelte` and `routes/tv/[id]/+page.svelte`.

## Test plan
- [ ] Open any movie detail page — cast strip shows portrait cards, not circles
- [ ] Open any TV series detail page — same portrait card style
- [ ] Verify faces are visible and not cropped (object-top)
- [ ] Hover a card — slight scale lift and shadow appear
- [ ] Verify fallback placeholder renders correctly for cast members with no profile photo
- [ ] Verify horizontal scroll still works when there are many cast members

🤖 Generated with [Claude Code](https://claude.com/claude-code)
