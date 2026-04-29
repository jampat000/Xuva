# Live Playback Inspector

The playback inspector exposes the current session route in a user-readable and machine-readable shape.

## Inspector API

`GET /api/sessions/{id}/inspector`

Example:

```json
{
  "sessionId": "session_20260430T101500_1",
  "mediaSourceId": "source-1",
  "deviceId": "web",
  "clientProfile": "web",
  "title": "Movie",
  "route": "transcode",
  "mode": "Video Transcode",
  "reasonCode": "video_conversion_required",
  "reasonText": "Video conversion required.",
  "selectedTracks": {
    "audio": "Default",
    "subtitles": "English PGS"
  },
  "bitrate": 61000000,
  "serverImpact": "High server load",
  "progressSeconds": 120,
  "durationSeconds": 7200,
  "status": "playing",
  "routeHistory": [
    {
      "fromRoute": "direct",
      "toRoute": "transcode",
      "fromReason": "direct_play_supported",
      "toReason": "video_conversion_required"
    }
  ]
}
```

## SSE Events

The existing `/api/events` stream publishes:

- `session.started`
- `session.updated`
- `session.inspector.updated`
- `session.route.changed`
- `session.stopped`

The event bus uses bounded per-subscriber buffers and non-blocking publish. Slow consumers can miss events, but they do not block playback-critical session updates.

## Player Integration

The web player updates the inspector when:

- a route is selected,
- a route changes from direct to prepared/transcode,
- progress/status changes,
- subtitle selection changes.

This keeps the dashboard and future player inspector in sync without full page refreshes.
