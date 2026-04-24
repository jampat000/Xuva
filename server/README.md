# Vyrden Server

The server owns libraries, users, playback decisions, streaming, and media processing.

Preferred starting stack:

- Go.
- SQLite.
- FFmpeg and ffprobe.
- HTTP/JSON API.

First milestone:

1. Probe a media file.
2. Store stream metadata.
3. Evaluate a client profile.
4. Return a playback decision with reasons.
5. Stream direct-play media.

