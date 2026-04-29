# Subtitle Pipeline

Vyrden treats subtitles as part of the playback route, not as a hidden afterthought.

## Subtitle Classes

- `text`: SRT, SubRip, WebVTT, ASS/SSA, mov_text.
- `image`: PGS, DVD subtitle, DVB subtitle.
- `unknown`: formats that require inspection before Vyrden can safely choose a route.

## Compatibility Rules

Text subtitles should avoid video conversion whenever possible:

- compatible text subtitle: direct subtitle path.
- unsupported text subtitle: convert to temporary WebVTT sidecar where supported.
- unknown text/container subtype: report inspection required.

Image subtitles are burn-in only as a last resort:

- if the client profile supports the image subtitle, Vyrden can keep video unchanged.
- if the client profile does not support it, Vyrden returns `subtitle_burn_required` with high server impact.

## Forecast Contract

Playback decisions include subtitle-specific fields:

```json
{
  "subtitleAction": "convert",
  "subtitleClass": "text",
  "subtitleImpact": {
    "class": "text",
    "action": "convert",
    "codec": "ass",
    "serverLoad": "low",
    "output": "webvtt sidecar",
    "userMessage": "Vyrden can convert this text subtitle to WebVTT without video conversion."
  }
}
```

Changing `subtitleTrackActive`, `subtitleTrackIndex`, or `subtitleCodec` on `/api/playback/decision` changes the forecast before playback starts.

## Conversion Entry Point

`POST /api/media-sources/{id}/subtitles/{index}/convert?clientProfile=web`

This returns a conversion plan for sidecar subtitles. It does not modify the original media file.

Example:

```json
{
  "conversion": {
    "status": "available",
    "sourceFormat": "ass",
    "outputFormat": "webvtt",
    "outputBehavior": "Generate a temporary WebVTT sidecar for playback; do not modify the original media file.",
    "serverImpact": "low",
    "reasonCode": "subtitle_text_conversion_available"
  }
}
```

## Burn-In Performance Note

Subtitle burn-in requires video conversion because the subtitle image must become part of each video frame. Expect high CPU use without hardware acceleration. With hardware acceleration unlocked and available, GPU conversion can reduce CPU load, but burn-in is still heavier than direct subtitle delivery.
