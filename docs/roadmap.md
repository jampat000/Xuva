# Roadmap

## Foundation

- Claim brand handles.
- Create private GitHub organization and repository.
- Write product principles.
- Write MVP scope.
- Decide initial architecture.
- Define playback decision model.
- Define lead TV design direction.
- Prototype core TV playback and setup surfaces.

## Playback Prototype

- Probe a local media folder with ffprobe.
- Store media probe results.
- Define client capability profile schema.
- Build playback decision engine.
- Serve direct-play files.
- Serve remuxed files.
- Start basic FFmpeg transcode jobs.
- Show playback decision explanations.
- Implement the version/audio/subtitle decision object used by the prototypes.

## Server MVP

- Real Lorivo player shell.
- Alpha desktop shell that starts the local server, opens the web UI, restarts the server, and provides native folder picking.
- Live dashboard updates through SSE.
- Playback session heartbeat and cleanup.
- Playback forecast with selected source, audio, and subtitle tracks.
- Resume progress and watched/unwatched state from item detail.
- Library scanning.
- Movie and TV grouping.
- Metadata matching.
- Multi-source ratings for IMDb, Rotten Tomatoes, TMDB, Metacritic, TVDB, and local/manual overrides.
- Local users.
- Device pairing.
- Server settings.
- Basic health dashboard.
- Windows, Linux, and Docker install paths.
- First-run setup flow.
- Remote access diagnostics without a vendor relay.

## First Clients

- Web admin and browser player.
- Apple TV client as the lead native alpha playback target.
- Android TV client after the Apple TV contract is proven.
- iOS and Android mobile apps after the TV playback loop is stable.
- Playback inspector in admin UI.
- Per-client capability reporting.
- Version selector.
- Audio and subtitle selector.
- Download selector.
- Series and episode detail.

## Premium Features

- Hardware transcoding.
- HDR tone mapping.
- Offline downloads.
- Subtitle conversion improvements.
- Intro and credits detection.
- Migration tools.

## Public Beta

- Installer polish.
- Crash reporting that does not expose media metadata by default.
- Documentation.
- Support workflow.
- Pricing page.
- Public website.
