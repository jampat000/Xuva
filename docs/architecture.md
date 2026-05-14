# Architecture

## High-Level Shape

Xuva is a local server with native and web clients.

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
- API: HTTP/JSON for commands and reads, SSE for live server-to-client events, WebSocket only for future two-way realtime control.
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

The first implementation should be one installable server process with separated internal services. This keeps installation light while preventing unrelated workloads from sharing uncontrolled queues.

Desktop installs should run as a user-launched tray/taskbar app by default, not as an always-on system service. Runtime folder browsing and selection should therefore use the signed-in user's permissions: local drives, mapped drives, and reachable NAS/UNC paths are available when that user can access them. Xuva should not require elevated service permissions just to choose media, metadata, cache, download, or transcode folders. Packaged desktop builds should expose a small web bridge for native folder picking and server restart controls; the browser-only folder API remains the fallback for dev, headless, and remote-admin use.

Storage must be library-path agnostic. Xuva should support local internal disks, removable USB disks, NAS/SMB/NFS/network shares, and mounted volumes. NAS safety is one tuning profile, not the only target.

Initial package layout:

```text
server/
  cmd/xuva/        Process entrypoint.
  internal/api/      HTTP routes, JSON, SSE.
  internal/app/      Dependency wiring and lifecycle.
  internal/config/   Runtime configuration.
  internal/database/ SQLite connection and migrations.
  internal/catalog/  SQLite-backed media catalog persistence.
  internal/libraries Library folders and scan scheduling.
  internal/scanner/  Shared filesystem walking and media extension filtering.
  internal/probe/    ffprobe worker pool.
  internal/media/    Shared media sources, streams, versions.
  internal/movies/   Movie naming, grouping, editions, cuts, collections.
  internal/tv/       Series, seasons, episodes, specials, next-up.
  internal/playback/ Playback decision engine.
  internal/streaming Direct file/range streaming.
  internal/transcode FFmpeg jobs and worker pools.
  internal/subtitles Embedded and sidecar subtitle logic.
  internal/devices/  Client profiles and capability reports.
  internal/sessions/ Active playback sessions.
  internal/downloads Offline preparation and queue.
  internal/events/   Internal event bus and SSE fanout.
  internal/resources CPU/GPU/disk/network limits.
  internal/jobs/     Bounded background queues.
```

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

## Movie And TV Separation

Movies and TV should be separate domain packages, but they should share the technical media and playback engine.

```text
Movie -> movie version -> media source -> playback decision
Episode -> episode version -> media source -> playback decision
```

Rules:

- Browsing is movie/TV-domain based.
- Playback is media-source based.
- Movies own collections, editions, cuts, and extras.
- TV owns series, seasons, episodes, specials, and next-up.
- Scanner, probe, subtitles, versions, streaming, transcode, downloads, devices, and sessions are shared services.
- The filesystem walker is shared; movie and TV classification logic is separate.
- `X:\Movies` and `X:\TV` should be scanned as separate libraries, even though their files feed the same media source and playback engine.
- A movie scan can be rescheduled without touching TV, and a TV scan can be rescheduled without touching movies.
- Scan results persist into SQLite as libraries, media sources, movie versions, series, seasons, episodes, and scan runs.
- Libraries carry a storage type (`local`, `removable`, `network`, `mounted`, or `unknown`) so future defaults can differ for local SSDs, USB drives, NAS shares, and mounted volumes.
- Metadata matching starts local-first from filenames and folders, with review queues exposed before any provider integration.
- Version intelligence is represented as grouped movie/episode versions instead of duplicate standalone items.

## Resource Scheduling

The server should distinguish between normal background work and playback-critical work.

Workload classes:

- Playback-critical: direct streaming, active remux/transcode, subtitle delivery, session heartbeats.
- Interactive: API requests, dashboard reads, users, devices, active sessions.
- Background: library scans, ffprobe jobs, metadata matching, image generation, chapter extraction, intro detection, cleanup.

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

Implementation rules:

- Scanning has its own bounded queue.
- Scan requests enqueue background jobs and return a job ID immediately.
- Scan progress is published over SSE so the UI never blocks on a full NAS scan.
- ffprobe has a separate low/medium concurrency pool. Network shares and removable drives should default to more conservative probing than local internal drives.
- ffprobe results persist separately from media source identity so source paths and technical stream facts can update independently.
- Transcoding has a strict worker pool.
- GPU jobs have a separate limiter.
- SQLite writes should use short transactions and batching where appropriate.
- SQLite should run locally in WAL mode with a short busy timeout and foreign keys enabled.
- SSE fanout must be non-blocking so a slow browser cannot block playback or scanning.
- Direct streaming must not wait behind scan/probe jobs.
- Direct media streaming is served from trusted catalog paths and supports browser range requests.
- Performance settings expose queue limits, active/queued work, and storage-aware recommendations before packaging.
- Movies and TV can scan independently, but both feed the shared media source store.
