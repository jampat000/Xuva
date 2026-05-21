## Problem

Hero cards on the home page and detail pages can autoplay a trailer (YouTube embed). There is currently no way to disable this — it plays unconditionally when trailer data is available.

Users in metered-bandwidth environments, privacy-conscious setups (no YouTube CDN), or households that prefer a static cinematic backdrop should be able to turn trailers off entirely.

## Proposed behaviour

Add a **Show Trailers** toggle to **Settings → Libraries** (or Settings → General).

| State | Hero behaviour |
|---|---|
| On (default) | Trailer autoplays when available, as today |
| Off | Hero always shows the backdrop image; trailer button is hidden |

The setting should be:
- Server-level (one setting for all users), controlled by an admin
- Persisted in `settings.json` / the existing settings store
- Exposed via `GET /api/settings` and respected in `PUT /api/settings`
- Also surfaced through `GET /api/client/bootstrap` so the frontend knows before the first hero renders (avoids a flicker)

## Acceptance criteria

- [ ] `enableTrailers` boolean field added to settings struct (default `true`)
- [ ] `GET /api/settings` returns `enableTrailers`
- [ ] `PUT /api/settings` accepts and persists `enableTrailers`
- [ ] `GET /api/client/bootstrap` includes `enableTrailers` in its feature flags object
- [ ] When `enableTrailers = false`: hero shows backdrop image, trailer autoplay is suppressed, no trailer button is rendered
- [ ] Settings → Libraries (or General) shows a labelled toggle for this option
- [ ] Governance gate stays green (if a new route is added, all three files updated)

## Implementation notes

- The frontend hero already has a `trailerKey` prop path — just guard the autoplay block with `if (enableTrailers && trailerKey)`.
- Bootstrap response already has a shape for feature flags — add `trailers: boolean` there.
- No database migration needed; the setting lives in the JSON config file.
