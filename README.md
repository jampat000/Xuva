# Xuva

Xuva is a local-first personal media server for private media libraries.

The goal is to combine polished TV-first playback with predictable local ownership: no required cloud account, no vendor relay dependency, no ads, and no streaming-service clutter.

Status: pre-alpha product and implementation foundation.

## Product Promise

Your media should play beautifully on the devices you own. When Xuva cannot direct-play a file, it should explain exactly why and choose the lightest possible fallback.

Playback preference order:

1. Direct play.
2. Remux only.
3. Audio-only transcode.
4. Subtitle conversion or burn-in.
5. Full video transcode.

## Principles

- Local-first by default.
- No required vendor cloud account.
- No vendor-hosted relay in v1.
- LAN playback works without internet.
- Remote access is user-owned and user-configured.
- Personal media is always the first-class experience.
- No ads in the paid product.
- Every playback decision must be explainable.
- The server should keep working if the company disappears.

## Repository Layout

```text
apps/
  android-tv/        Native Android TV client.
  apple-core/        Shared SwiftUI client code for Apple platforms.
  ios/               Native iPhone/iPad client.
  tvos/              Native Apple TV client.
  web/               Web admin and browser player.
docs/                Product, architecture, and business planning.
packages/
  client-profiles/   Device capability profiles and playback rules.
  media-probe/       Media probing schema and helpers.
server/              Xuva Server.
```

## Current Focus

The current web foundation is the SvelteKit Xuva app in `apps/web/svelte`.
It is the only production web UI and is served by the Go server from
`server/internal/webapp/static-next`.

Phase focus:

- Phase 1: desktop owner web only
- primary acceptance viewport: `1600x1000` (secondary: `1920x1080`)
- tablet/mobile web: smoke-only
- auth/users expansion: Phase 2
- native tvOS/Android clients: later phase

The first technical milestone is a playback decision prototype:

1. Scan a local media folder.
2. Probe files with FFmpeg/ffprobe.
3. Compare streams against a client capability profile.
4. Decide direct play, remux, partial transcode, or full transcode.
5. Stream the result.
6. Show the exact reason for the chosen path.

## Documentation

- [Agent Map](AGENTS.md)
- [Docs Index](docs/index.md)
- [Agent Harness](docs/agent-harness.md)
- [Quality Scorecard](docs/quality-score.md)
- [Branding](docs/branding.md)
- [Product Principles](docs/product-principles.md)
- [MVP Scope](docs/mvp-scope.md)
- [Architecture](docs/architecture.md)
- [Development mode](docs/development-mode.md)
- [Desktop owner mode](docs/desktop-owner-mode.md)
- [Security policy](SECURITY.md)
- [Repository hardening](docs/repository-hardening.md)
- [Roadmap phases](docs/roadmap-phases.md)
- [Playback Engine](docs/playback-engine.md)
- [Media Scanner](docs/media-scanner.md)
- [Remote Access](docs/remote-access.md)
- [Monetization](docs/monetization.md)
- [Roadmap](docs/roadmap.md)
- [Build Plan](docs/build-plan.md)
- [Competitive Notes](docs/competitive-notes.md)
- [Design](docs/design/README.md)
