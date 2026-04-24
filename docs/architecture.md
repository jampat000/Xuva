# Architecture

## High-Level Shape

Vyrden is a local server with native and web clients.

```text
Client app
  -> Server API
  -> Library database
  -> Media probe data
  -> Playback decision engine
  -> Streaming/transcode workers
  -> Local media files
```

## Server

Preferred starting stack:

- Language: Go.
- Database: SQLite.
- Media probing: ffprobe.
- Media processing: FFmpeg.
- API: HTTP/JSON first, WebSocket or SSE for session events.
- Packaging: Windows installer, Linux packages, Docker.

Go is the pragmatic default because it is fast enough, easy to distribute, straightforward to operate, and simpler to hire for than more specialized stacks.

Core server services:

- Library scanner.
- Metadata matcher.
- Media probe service.
- Playback decision service.
- Streaming session service.
- Transcode worker service.
- Download preparation service.
- Device capability service.
- Remote access diagnostics service.
- License service with local cache.

## Web

Preferred starting stack:

- TypeScript.
- React.
- Vite or equivalent lightweight build tooling.
- Tailwind or CSS modules, depending on final design direction.

The web app should cover setup, admin, library management, diagnostics, and browser playback.

## Native TV Apps

Android TV:

- Kotlin.
- Native player APIs first.
- Capability reporting back to the server.

Apple TV:

- Swift.
- AVFoundation first.
- Capability reporting back to the server.

## Data Model Areas

- Libraries.
- Media items.
- Media sources.
- Probe results.
- People and metadata.
- Users.
- Devices.
- Playback sessions.
- Transcode jobs.
- Downloads.
- Server settings.
- Licenses.

## Playback Path

Playback should move through a deterministic contract:

```text
Client request
  -> selected media source
  -> selected version
  -> selected audio track
  -> selected subtitle track
  -> client capability profile
  -> network estimate
  -> playback decision
  -> stream plan
```

The UI should never guess playback state. It should show the server's decision object, including the chosen mode and reason.

## Resource Scheduling

The server should distinguish between normal background work and playback-critical work.

Resource controls:

- CPU limits.
- GPU transcode queue.
- RAM cache limits.
- Disk cache limits.
- Per-user stream limits.
- Per-device quality limits.
- Burst mode.
- Quiet mode.

Playback-critical jobs should preempt background scans, chapter extraction, intro detection, and preview generation. The scheduler should protect active playback from avoidable resource spikes.
