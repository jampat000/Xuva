# Playback Engine

## Objective

Xuva should choose the cheapest playback path that works for the current client and network conditions.

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
- Adaptive stream.
- Audio transcode.
- Subtitle conversion.
- Subtitle burn-in.
- Video transcode.

The decision must also include a human-readable reason.

Examples:

- Direct play: client supports MKV, HEVC Main10, EAC3, and selected SRT subtitles.
- Remux: client supports H.264 and AAC but not MKV container.
- Adaptive stream: remote network limit is below the source bitrate and the client supports HLS.
- Audio transcode: client supports video stream but not DTS-HD MA audio.
- Subtitle burn-in: selected PGS subtitle cannot be rendered by this client.
- Video transcode: client cannot decode AV1 Main10 at this resolution.

## Decision Flow

```mermaid
flowchart TD
    Start([Client requests playback]) --> Probe[Probe selected media source<br/>container · video · audio · subtitle tracks]
    Probe --> Profile{Client capability<br/>profile match?}
    Profile -->|All streams supported<br/>+ container OK<br/>+ bitrate within budget| Direct[Direct Play<br/>serve original file]
    Profile -->|Streams OK,<br/>container not supported| Remux[Remux<br/>repack container, copy streams]
    Profile -->|Video OK,<br/>audio codec unsupported| Audio[Audio Transcode<br/>copy video, re-encode audio]
    Profile -->|Subtitle is image<br/>(PGS/DVB), client text-only| Burn[Subtitle Burn-In<br/>burn subs into video stream]
    Profile -->|Subtitle is text,<br/>codec unsupported| SubConv[Subtitle Conversion<br/>SRT/VTT mux]
    Profile -->|Network bitrate<br/>below source| Adaptive[Adaptive Stream<br/>HLS variants]
    Profile -->|Video codec or HDR<br/>unsupported| Transcode[Video Transcode<br/>re-encode video + audio]
    Direct --> Out([Decision returned<br/>with reason + actions])
    Remux --> Out
    Audio --> Out
    Burn --> Out
    SubConv --> Out
    Adaptive --> Out
    Transcode --> Out
```

The preference order is exactly that of [README](../README.md): direct play → remux → audio-only → subtitle convert/burn → full transcode. Every decision returns a `reasonCode` + `reasonText` so the UI can show *why* the chosen path was taken.

## Supported formats (current matrix)

What the server can ingest and play to current clients. Probe-data dependent; an item not in this table may still play via transcode.

| Containers | Video codecs | Audio codecs | Subtitles | Notes |
|---|---|---|---|---|
| MKV, MP4, MOV, WebM, AVI, TS, M4V | H.264 / AVC, H.265 / HEVC, AV1, VP9, MPEG-2, VP8 | AAC, AC-3, E-AC-3, DTS, DTS-HD, TrueHD, FLAC, MP3, Opus, Vorbis, PCM, ALAC | SRT, ASS / SSA, VTT, PGS, DVB-Sub, MOV-Text | HDR10, HDR10+, Dolby Vision (profile 5/8), HLG passed through on capable clients |

Client direct-play decisions are governed by `packages/client-profiles/` — each client (web, Apple TV, iOS, Android TV) declares which containers / codecs / HDR formats it can decode. Anything outside its profile triggers remux or transcode automatically.

## Decision Object

Every playback request should return a structured decision object:

```text
mode
reason
source_file
selected_version
selected_audio_track
selected_subtitle_track
video_action
audio_action
subtitle_action
container_action
estimated_cpu_cost
estimated_gpu_cost
estimated_network_bitrate
client_capability_match
suggested_fixes
```

The TV app, web admin, and playback inspector should all read from the same decision object so the product stays consistent.

## Design Rule

Never perform a heavier operation when a lighter one will work.

Decision order:

1. Try direct play for the chosen version and tracks.
2. Try container remux if streams are compatible but the container is not.
3. Try audio-only transcode if video and subtitles are compatible.
4. Try subtitle conversion or direct subtitle rendering before burn-in.
5. Use adaptive streaming for constrained remote routes before plain video transcode when the client supports it.
6. Burn subtitles only when the selected subtitle cannot render on the client.
7. Transcode video only when video compatibility, bitrate, HDR, or subtitle burn-in requires it.

## Version Selection

Versions should be ranked by the user's intent and route:

- Best local quality.
- Best remote quality.
- Most compatible.
- Smallest offline copy.
- User-pinned preference.

The selector should explain the tradeoff before playback. A lower-bitrate compatible version is often better than transcoding a larger file.

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

## Download Preparation

Offline downloads should use the same playback decision engine with a different target profile:

- Target device capability.
- Available storage.
- Offline subtitle renderer.
- Battery and network state.
- User quality preference.

The server can pre-transcode a download, but the app must show the resulting file size, included tracks, and playback compatibility before the user starts the job.

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
