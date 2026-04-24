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

- `GET /api/libraries`
- `POST /api/libraries/movies/scan`
- `POST /api/libraries/tv/scan`
- `POST /api/libraries/scan`

Scan requests can pass `path`, or use `VYRDEN_MOVIES_PATH` and `VYRDEN_TV_PATH` when those environment variables are configured.

Catalog endpoints:

- `GET /api/catalog/summary`

First milestone:

1. Probe a media file.
2. Store stream metadata.
3. Evaluate a client profile.
4. Return a playback decision with reasons.
5. Stream direct-play media.
