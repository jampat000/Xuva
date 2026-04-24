# Build Plan

## Current Definition Of Complete

The current planning/design phase is complete when Vyrden has:

- A clear local-first product stance.
- A credible monetization boundary.
- A playback decision architecture.
- A TV-first visual direction.
- Prototype coverage for the core playback decisions.
- A concrete engineering path into the first server prototype.

The repo now aims to satisfy that definition before moving into production code.

## Immediate Engineering Milestone

Build a playback decision prototype before building the full UI.

Required behavior:

- Probe a local folder with `ffprobe`.
- Store probe results in SQLite.
- Define a client capability profile.
- Select a media version, audio track, and subtitle track.
- Return a playback decision object.
- Explain direct play, remux, audio transcode, subtitle burn, and video transcode.

## Server Skeleton

Start with:

- Go HTTP API.
- SQLite database.
- File scanner.
- `ffprobe` integration.
- Playback decision package.
- Basic direct file streaming.
- FFmpeg worker wrapper.
- Session event stream.

Internal separation:

- REST for commands and reads.
- SSE for live dashboard updates.
- WebSocket later only for two-way playback/remote-control surfaces.
- Separate packages for Movies, TV, shared media sources, playback, subtitles, scanner, probe, transcode, sessions, downloads, events, and resources.
- Separate bounded queues for scan, probe, and transcode work.
- Shared filesystem walker with separate movie and TV classifiers.
- Separate scan endpoints for Movies, TV, and combined library scans.
- SQLite migrations for libraries, media sources, movies, TV series/seasons/episodes, versions, and scan runs.
- Scan endpoint persistence with idempotent upserts.
- Background scan jobs with SSE progress and scan job lookup endpoints.
- Catalog browse APIs for movies, series, review items, and media sources.
- ffprobe parser and media probe persistence.
- Playback decision v1 using stored probe facts.
- Direct media source streaming with HTTP range support.

## Client Skeleton

Start with web admin before native TV clients:

- First-run setup.
- Library scan progress.
- Media detail page.
- Playback decision inspector.
- Active sessions.
- Transcode queue.

Native TV apps should follow once the server decision contract is stable.

## Product Gates

Do not move to public beta until:

- LAN playback works without an internet connection.
- Every transcode has a visible reason.
- Subtitles are handled deliberately.
- Downloads resume safely.
- Install and update paths are boring.
- Paid feature boundaries are clear.
