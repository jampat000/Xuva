# Vyrden

Vyrden is a local-first personal media server for private media libraries.

The goal is to combine polished TV-first playback with predictable local ownership: no required cloud account, no vendor relay dependency, no ads, and no streaming-service clutter.

Status: pre-alpha product planning.

## Product Promise

Your media should play beautifully on the devices you own. When Vyrden cannot direct-play a file, it should explain exactly why and choose the lightest possible fallback.

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
  apple-tv/          Native Apple TV client.
  web/               Web admin and browser player.
docs/                Product, architecture, and business planning.
packages/
  client-profiles/   Device capability profiles and playback rules.
  media-probe/       Media probing schema and helpers.
server/              Vyrden Server.
```

## Current Focus

The first technical milestone is a playback decision prototype:

1. Scan a local media folder.
2. Probe files with FFmpeg/ffprobe.
3. Compare streams against a client capability profile.
4. Decide direct play, remux, partial transcode, or full transcode.
5. Stream the result.
6. Show the exact reason for the chosen path.

## Documentation

- [Branding](docs/branding.md)
- [Product Principles](docs/product-principles.md)
- [MVP Scope](docs/mvp-scope.md)
- [Architecture](docs/architecture.md)
- [Playback Engine](docs/playback-engine.md)
- [Remote Access](docs/remote-access.md)
- [Monetization](docs/monetization.md)
- [Roadmap](docs/roadmap.md)
- [Competitive Notes](docs/competitive-notes.md)

