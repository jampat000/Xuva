# Adaptive Streaming

Lorivo uses adaptive streaming for remote or constrained routes where the original file bitrate is higher than the selected network limit. The first implementation targets HLS because it is broadly supported by browsers and native players. DASH remains a later protocol option behind the same planning contract.

## Route Selection

Adaptive streaming is considered after media facts are known and before plain video transcode for constrained remote playback.

Required inputs:

- HLS-capable client profile.
- Probed source bitrate and resolution.
- Remote, WAN, internet, or constrained route, or an explicit network bitrate limit.
- Source bitrate above the selected network limit.
- No subtitle burn-in requirement.

If any input is missing, Lorivo falls back to the existing decision engine path: direct play, remux, audio conversion, video conversion, or a policy block.

## Profile Ladder

The initial ladder is intentionally conservative:

| Variant | Resolution | Target Bitrate | Codec |
| --- | --- | ---: | --- |
| `2160p` | 3840x2160 | 35 Mbps | H.264/AAC |
| `1440p` | 2560x1440 | 20 Mbps | H.264/AAC |
| `1080p` | 1920x1080 | 8 Mbps | H.264/AAC |
| `720p` | 1280x720 | 4 Mbps | H.264/AAC |
| `480p` | 854x480 | 1.5 Mbps | H.264/AAC |

The plan filters out variants above the source resolution and above the selected network target. If no variant fits the network target but the source is otherwise eligible, Lorivo keeps the lowest source-compatible rung so the route can still degrade cleanly rather than failing the session.

## API Surface

- `POST /api/media-sources/{id}/adaptive/session` returns the adaptive plan and manifest URL.
- `GET /api/media-sources/{id}/adaptive/master.m3u8` returns the HLS master playlist for the selected route.
- `GET /api/media-sources/{id}/adaptive/variant-{id}.m3u8` returns the variant playlist envelope.
- `POST /api/adaptive/telemetry` records startup, variant, stall, recovery, and ended events.

Protected adaptive routes use the media route policy group and are available to admin and standard users. Browser mutations require CSRF.

## Telemetry Snapshot

Telemetry events are normalized to the `adaptive.*` namespace and published onto the local event bus, which also feeds `/api/metrics`.

Example stall event:

```json
{
  "event": "adaptive.stall",
  "sessionId": "sess_1",
  "mediaSourceId": "source_1",
  "clientProfile": "web",
  "variantId": "720p",
  "bufferSeconds": 0.4,
  "stallMs": 1200,
  "observedBitrate": 3800000,
  "correlationId": "req_abc"
}
```

Operational use:

- Stall rate: count `adaptive.stall` over active adaptive sessions.
- Recovery rate: compare `adaptive.recovered` with `adaptive.stall`.
- Variant pressure: watch frequent `adaptive.variant_changed` events down to `480p`.
- Correlation: use `correlationId` to connect API requests, playback route selection, and client adaptation reports.

## Resource Impact

Adaptive streaming is heavier than direct play and remux because each active rung may require FFmpeg video encoding. The decision engine reports medium CPU cost and optional GPU cost so UI surfaces can explain the route plainly. Operators should prefer hardware encoding for remote adaptive sessions and keep concurrent transcode worker counts conservative on low-power hosts.

The initial API plan and playlists are local-only control plane work; FFmpeg remains the intended media operation backend for actual segment generation. Lorivo does not use vendor relay, CDN, or hosted packaging infrastructure.

## Rollback Plan

Rollback is low risk:

1. Disable adaptive selection by removing HLS support from the affected device profile or by using the Original Only or Light playback policy.
2. Existing remote playback falls back to video transcode or policy-blocked fallback options.
3. Remove or ignore adaptive telemetry consumers; events are additive and isolated under `adaptive.*`.
4. Revert the adaptive API routes and playback `Adaptive Stream` mode if a release rollback is required.

