# Playback Engine

## Objective

Vyrden should choose the cheapest playback path that works for the current client and network conditions.

## Inputs

- Media container.
- Video codec, profile, level, bit depth, resolution, frame rate, HDR format.
- Audio codec, channel count, sample rate, language, default flag.
- Subtitle codec, language, forced flag, default flag.
- Client capability profile.
- Network bandwidth estimate.
- User quality preference.
- Server hardware capabilities.

## Output

Playback mode:

- Direct play.
- Remux.
- Audio transcode.
- Subtitle conversion.
- Subtitle burn-in.
- Video transcode.

The decision must also include a human-readable reason.

Examples:

- Direct play: client supports MKV, HEVC Main10, EAC3, and selected SRT subtitles.
- Remux: client supports H.264 and AAC but not MKV container.
- Audio transcode: client supports video stream but not DTS-HD MA audio.
- Subtitle burn-in: selected PGS subtitle cannot be rendered by this client.
- Video transcode: client cannot decode AV1 Main10 at this resolution.

## Design Rule

Never perform a heavier operation when a lighter one will work.

## Subtitle Priorities

Subtitle support should be treated as a headline feature, not a leftover part of playback.

Initial targets:

- SRT.
- WebVTT.
- ASS/SSA through libass-compatible rendering.
- PGS.
- VobSub.
- Forced subtitles.
- Subtitle delay controls.
- Per-user subtitle language preferences.

Future targets:

- OCR image subtitles to text.
- Encoding repair.
- Automatic subtitle sync suggestions.

## Inspector

Every active session should have an inspector with:

- User.
- Client.
- File.
- Selected tracks.
- Playback mode.
- Exact reason.
- Bitrate.
- Server CPU/GPU load.
- Transcode speed.
- Buffer health.

