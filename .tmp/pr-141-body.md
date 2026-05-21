## Summary
- The hero "Add to watchlist" `+` button had no `onclick` handler and no `watchlistStore` import — clicking it was a no-op
- Import `toggleWatchlist` / `isInWatchlist` from `watchlistStore.svelte` and `Check` from lucide-svelte
- `onclick` calls `toggleWatchlist` with the correct `id`, `kind` (mapped from `featured.type`), `title`, `year`, `posterUrl`, `backdropUrl`, and `genres`
- Button visually toggles: highlighted ring + check icon when watchlisted; plain + plus icon when not

## Test plan
- [ ] Open `/movies` — hero card shows plus button
- [ ] Click plus — button turns to check with highlight ring; Watchlist count in nav increments
- [ ] Refresh page — button still shows check (persisted to localStorage)
- [ ] Click check — button reverts to plus; count decrements
- [ ] Repeat on `/tv` page for a series

Closes #141

🤖 Generated with [Claude Code](https://claude.com/claude-code)
