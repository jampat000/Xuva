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
- `internal/media`: shared media source model.
- `internal/movies`: movie-specific domain logic.
- `internal/tv`: series/season/episode domain logic.
- `internal/playback`: playback decision engine.
- `internal/resources`: workload classes and resource limits.
- `internal/jobs`: bounded work queues.

First milestone:

1. Probe a media file.
2. Store stream metadata.
3. Evaluate a client profile.
4. Return a playback decision with reasons.
5. Stream direct-play media.
