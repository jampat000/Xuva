## Root cause

Web browsers cannot natively play AC3, DTS, TrueHD, or similar surround audio codecs. When a file has one of these as its default audio track, the Go server's route decision engine requires audio conversion — but if the server's playback policy is set to `original_only` (or similar), the route request returns `blocked_by_policy` and the play page was showing an error screen.

Before this fix the error was a silent blank video (no early-exit existed). After PR #152 added the early-exit, the error became visible but still unplayable.

## Change

### `PlaybackQueryOptions` / `playbackQueryString` (`details.ts`)
- Add `forcePlayable?: boolean` field
- Serialise it as `forcePlayable=true` in the query string when set

### Play page (`play/[mediaSourceId]/+page.svelte`)
Replace the hard early-exit for `blocked_by_policy` with a two-attempt strategy:

1. **First attempt** — normal route request (respects policy)
2. **If blocked** → automatically retry with `forcePlayable=true` (lets the server pick the best route it can actually serve, bypassing the policy gate)
3. **Error screen only shown** if the retry is also blocked or returns no usable URL

All subsequent references to `initialRoute` in the stream-token phase are updated to use `finalAttemptRoute` (the result of whichever attempt succeeded).

## Test plan
- [ ] Play a file whose default audio track is AC3/DTS → plays (no error screen)
- [ ] Play a file with native AAC audio → plays as before (single route request, no retry)
- [ ] Temporarily change server policy to one that blocks all transcoding → error screen still shows (retry also blocked)
- [ ] Confirm stream-token auth still works (signed URL used, not plain URL)

🤖 Generated with [Claude Code](https://claude.com/claude-code)
