Closes #128

## Summary
- New `DisableTrailers` bool config field (default `false` = trailers enabled). Stored in `settings.json` via `omitempty` — zero value means show trailers, so existing installs are unaffected.
- `GET /api/client/bootstrap` now includes `features.trailers: bool` — the SPA reads it before the first hero frame renders, avoiding any flicker.
- `PUT /api/settings` accepts `disableTrailers` field.
- `appState.trailersEnabled` (default `true`) — set from bootstrap on every page load.
- `Hero.svelte` now takes a `trailersEnabled` prop. When false, `hasLocalTrailer` and `hasYouTubeFallback` are forced false — no trailer timers fire, the static backdrop image is shown.
- **Settings → Libraries → Display options**: a live toggle card. Clicking it auto-saves via `PUT /api/settings` and propagates to `appState` immediately — no page reload required.

## Test plan
- [ ] Default state: hero plays trailers as before (no regression)
- [ ] Settings → Libraries: "Display options" card visible with "Show trailers in hero" toggle (defaults ON)
- [ ] Toggle OFF → hero immediately shows backdrop only, mute button disappears
- [ ] Toggle ON again → trailers resume on next slide change
- [ ] Reload page with toggle OFF → bootstrap returns `features.trailers: false`, hero starts in backdrop mode
- [ ] `GET /api/settings` response includes `config.disableTrailers: true` when toggled off

🤖 Generated with [Claude Code](https://claude.com/claude-code)
