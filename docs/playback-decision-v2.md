# Playback Decision Engine v2

Vyrden's playback decision contract is deterministic and explainable. The same normalized request and media facts produce the same route, reason code, and `decisionTraceId`.

## Request Inputs

The decision engine accepts:

- `mediaSourceId`: selected file version.
- `clientProfile`: target player profile such as `web`, `android-tv`, or `apple-tv`.
- `policy`: playback policy hint, for example original-only or fallback allowed.
- `routeType`: local, LAN, or remote route hint.
- `maxNetworkBitrate`: optional bitrate ceiling for constrained links.
- selected audio and subtitle track indexes/codecs.
- profile capabilities supplied by the device registry: containers, video codecs, audio codecs, and subtitle codecs.

Source facts come from the media check/probe result: container, video codec, resolution, bitrate, audio tracks, subtitles, and sidecar subtitles.

## Response Contract

Responses keep the existing `mode` and `reason` fields for UI compatibility and add v2 fields:

```json
{
  "mode": "Video Transcode",
  "reason": "This player profile needs video conversion for the selected file.",
  "reasonCode": "video_conversion_required",
  "reasonText": "This player profile needs video conversion for the selected file.",
  "decisionTraceId": "dec_0123456789abcdef",
  "containerAction": "transcode_or_remux",
  "videoAction": "transcode",
  "audioAction": "direct",
  "subtitleAction": "none",
  "estimatedCpuCost": "high",
  "estimatedGpuCost": "optional",
  "estimatedNetworkBitrate": 61000000,
  "selected": {
    "container": "matroska,webm",
    "videoCodec": "hevc",
    "resolution": "3840x2160"
  },
  "suggestedFixes": [
    "Try a player that supports the original file",
    "Allow temporary video conversion",
    "Unlock hardware acceleration to reduce CPU load"
  ]
}
```

## Decision Order

The engine evaluates in this stable order:

1. Missing media source or media check required.
2. Incomplete probe facts.
3. Subtitle burn-in requirement.
4. Network bitrate constraint.
5. Direct play.
6. Audio conversion with compatible video.
7. Container remux with compatible video.
8. Video conversion.

This order prevents ambiguous mixed outcomes from changing between runs.

## Reason Codes

- `source_required`
- `probe_required`
- `media_facts_incomplete`
- `subtitle_burn_required`
- `network_bitrate_limit`
- `direct_play_supported`
- `audio_conversion_required`
- `container_remux_required`
- `audio_and_container_conversion_required`
- `video_conversion_required`

## Compatibility Notes

Existing UI consumers can continue using `mode` and `reason`. New UI surfaces should prefer `reasonCode`, `reasonText`, `decisionTraceId`, `suggestedFixes`, and the action fields.
