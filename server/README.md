# Vyrden Server

The server owns libraries, users, playback decisions, streaming, and media processing.

Preferred starting stack:

- Go.
- SQLite.
- FFmpeg and ffprobe.
- HTTP/JSON API.
- SSE event stream for live admin updates.
- WebSocket later for targeted two-way playback control.

Structure:

- `cmd/vyrden`: process entrypoint.
- `internal/api`: HTTP/JSON and SSE.
- `internal/app`: service wiring and lifecycle.
- `internal/database`: SQLite connection, pragmas, and migrations.
- `internal/catalog`: SQLite-backed library and media catalog persistence.
- `internal/libraries`: configured movie and TV roots.
- `internal/scanner`: shared filesystem walking and media extension detection.
- `internal/media`: shared media source model.
- `internal/movies`: movie-specific classification and domain logic.
- `internal/tv`: series/season/episode classification and domain logic.
- `internal/playback`: playback decision engine.
- `internal/resources`: workload classes and resource limits.
- `internal/jobs`: bounded work queues.

Initial scan endpoints:

- `GET /`
- `GET /api/libraries`
- `GET /api/scans`
- `GET /api/scans/{id}`
- `POST /api/libraries/movies/scan`
- `POST /api/libraries/tv/scan`
- `POST /api/libraries/scan`

`GET /` serves the local dev console for quick testing.

Scan requests return a background scan job with HTTP 202. Progress is emitted through `GET /api/events` as `scan.queued`, `scan.started`, `scan.progress`, `scan.completed`, and `scan.failed`.

Scan requests can pass `path`, or use `VYRDEN_MOVIES_PATH` and `VYRDEN_TV_PATH` when those environment variables are configured.

Library paths are not NAS-specific. The scanner accepts any OS-visible folder, including local disks, removable USB drives, network shares, and mounted volumes. Library records include a `storageType` field so future scan/probe behavior can tune itself per storage class.

Catalog endpoints:

- `GET /api/catalog/summary`
- `GET /api/movies`
- `GET /api/movies/{id}`
- `GET /api/series`
- `GET /api/series/{id}`
- `GET /api/review`
- `GET /api/metadata/suggestions`
- `GET /api/versions`
- `GET /api/catalog/health`
- `GET /api/settings/performance`
- `GET /api/media-sources`
- `GET /api/media-sources/{id}`
- `GET /api/media-sources/{id}/stream`
- `POST /api/media-sources/{id}/probe`
- `GET /api/playback/decision?mediaSourceId={id}&clientProfile=web`

First milestone:

1. Probe a media file.
2. Store stream metadata.
3. Evaluate a client profile.
4. Return a playback decision with reasons.
5. Stream direct-play media.

## Authentication and Session Security

Vyrden now supports local credential authentication with server-side sessions.

Environment variables:

- `VYRDEN_AUTH_DISABLED=false`
- `VYRDEN_ADMIN_USERNAME=admin`
- `VYRDEN_ADMIN_PASSWORD=...`

Bootstrap behavior:

- If auth is enabled and no credentialed user exists yet, Vyrden creates the initial local admin user on startup.
- If `VYRDEN_ADMIN_PASSWORD` is omitted for first boot, Vyrden generates a random bootstrap password and logs it once to the server log.
- Auth bootstrap settings are environment-only and are not written back into `settings.json`.

Session behavior:

- Passwords are hashed with `argon2id`.
- Browser auth uses an `HttpOnly` session cookie plus a companion CSRF cookie/token pair.
- Session expiry is extended on valid use.
- Session secrets rotate periodically during active use.
- Logout revokes the current session immediately.
- Invalid login bursts trigger a temporary lockout window.

Protected routes in this phase:

- browser write operations
- playback session management
- direct media stream endpoints
- subtitle stream endpoints
- file download endpoints
- `/play/{id}`

Operators should keep auth enabled outside local development. Role-based authorization and finer route policy are tracked separately in `P0.2`.
