## Summary
- The three sync-behaviour checkboxes (Auto-scrobble, Sync ratings, Import watchlists) were hardcoded `checked` with no state binding — toggling them had zero effect and there was no way to save
- Added reactive `wlSyncEdit` / `wlSyncSaved` state with localStorage persistence
- Bound each checkbox to `wlSyncEdit[opt.id]` via `checked` + `onchange`
- Save/Discard button row appears only when changes are pending (`wlSyncDirty`)
- Save commits the edit state and writes to `xuva-wl-sync-opts` in localStorage; Discard reverts

Note: the underlying service integrations (Trakt, Letterboxd, Simkl) still use client-side mocks. These settings will be wired to the server once backend endpoints exist.

## Test plan
- [ ] Open Settings → Watchlist Services → scroll to "Sync behaviour"
- [ ] All three checkboxes start checked (default state)
- [ ] Uncheck one → Save/Discard buttons appear
- [ ] Click Discard → checkbox reverts, buttons disappear
- [ ] Uncheck one → Click Save → buttons disappear, preference persists
- [ ] Refresh page → unchecked checkbox remains unchecked

Closes #145

🤖 Generated with [Claude Code](https://claude.com/claude-code)
