## Root cause

The SvelteKit player never called the stream-token endpoint. The Go server's `authorizeStreamRequest()` requires `?sessionId=&deviceId=&token=` query params on every `/api/media-sources/{id}/stream` request. A native `<video>` element doesn't send `X-Auth-Token` headers, so **every stream request returned 403 when auth was enabled** — playback silently failed.

Four additional issues compounded the problem.

## Changes

### 1. Stream auth (root fix — `play/+page.svelte`, `details.ts`)
- Add `getStreamToken(mediaSourceId, sessionId, deviceId)` → `POST /api/media-sources/{id}/stream-token` (endpoint already exists in server)
- After `startClientPlayback()` returns a `sessionId`, call `getStreamToken` and replace `route.url` / `route.manifestUrl` with the signed URLs before passing them to `<Player>`
- Falls back to plain URL gracefully when auth is disabled (`getStreamToken` throws → catch and continue)
- Adaptive (HLS) streams: append `tokenResp.query` to the manifest URL

### 2. Probe transparency (`play/+page.svelte`)
- `startClientPlayback()` runs a synchronous foreground ffprobe (up to 45 s) on first play of any unprobed file
- New three-phase loader: **"Resolving stream…" → "Analysing file…" → "Authorising stream…"** — user sees clear progress instead of a stuck black spinner
- Phase label only shows "Analysing file…" when the session request takes > 300 ms

### 3. Blocked / transcoding routes now surface errors (`play/+page.svelte`)
- `route.status === 'blocked_by_policy'` → error message with `decision.reasonText` + suggested fallback options
- `route.status === 'queuing' | 'transcoding'` with no URL → "needs to be transcoded" message with next steps
- No more silent empty-src on the `<video>` element

### 4. Parallel track loading (`Player.svelte`)
- `getMediaSourceTracks()` was `await`-ed before `loadSource()`, adding ~200 ms of sequential latency
- Now fired in parallel via `Promise.allSettled()`; video starts buffering immediately while track metadata loads alongside

### 5. Detail page progressive loading (`movies/[id]`, `tv/[id]`)
- `getMetadataRecords()` (multi-provider DB query) was blocking the whole page via `Promise.all()`
- Both detail endpoints already embed primary metadata in their response
- Now: await `getMovieDetail/getSeriesDetail` first → `loading = false` immediately using embedded metadata → update `metadata` + `altRecords` when provider records arrive in background

## Test plan
- [ ] Play a movie/episode with auth enabled → video starts playing (not 403)
- [ ] Play a file that has never been probed → "Analysing file…" label visible during probe, then playback starts
- [ ] Restart server, play a previously-probed file → loads in < 500 ms total
- [ ] Play a file that needs transcoding (or temporarily change policy to block) → clear error message shown, not silent failure
- [ ] Switch audio tracks mid-playback → works (still uses signed URL)
- [ ] Open a movie detail page → renders immediately even if metadata provider records are slow
- [ ] Open a TV detail page → same
- [ ] Confirm audio/subtitle menus still populate correctly (parallel track load)

🤖 Generated with [Claude Code](https://claude.com/claude-code)
