# Vyrden TV Experience

## Objective

Vyrden TV should feel like opening a private cinema library. It should make movies and TV shows visually prominent, support remote-first navigation, and expose playback quality without turning the home screen into a diagnostics tool.

## Core Screens

### Home

Purpose: resume watching, browse personal library rows, and surface one featured item from the user's own collection.

Required sections:

- Global navigation.
- Featured media hero.
- Continue Watching.
- Movies.
- TV Shows.
- Recently Added.
- Downloads, if available on device.

### Media Detail

Purpose: explain what the item is and make playback choices obvious.

Required sections:

- Backdrop.
- Poster.
- Title.
- Year, runtime, rating, resolution, HDR, audio format.
- Play.
- Resume.
- Download.
- Trailer, if available locally or through metadata provider.
- Version selector.
- Audio selector.
- Subtitle selector.
- Collection context.
- Playback compatibility preview.

Design intent:

- Keep the screen cinematic and poster-led.
- Show version selection before playback.
- Surface the selected audio and subtitle tracks as first-class choices.
- Forecast the playback path before the user presses play.
- Explain whether the server will direct-play, remux, transcode audio, burn subtitles, or transcode video.
- Make offline download quality choices visible without hiding them in settings.

### Playback Overlay

Purpose: support watching without visual clutter.

Required sections:

- Timeline.
- Current time and remaining time.
- Play/pause.
- Skip back/forward.
- Audio.
- Subtitles.
- Quality.
- More.
- Quiet direct-play/transcode state.

Design intent:

- Keep play/pause, skip, version, audio, subtitles, quality, inspector, and watched state visible in one predictable control row.
- Let audio, subtitle, version, and quality focus open a lightweight selector without leaving playback.
- Keep direct-play/remux/transcode state visible as a small status, not as a distracting dashboard.
- Preserve a full inspector path for users who want to know exactly why Vyrden chose the current playback route.
- Make subtitle burn-in risk visible before the user selects a subtitle track.

### Audio And Subtitle Selector

Purpose: make track choice obvious, fast, and technically honest.

Required sections:

- Audio tracks with language, codec, channel layout, default/commentary flags, and route impact.
- Subtitle tracks with language, forced/default flags, format, render path, and burn-in risk.
- Combined playback forecast for the selected audio and subtitle pair.
- Subtitle delay.
- Audio delay.
- Remember choice scope.
- Apply and cancel.

Design intent:

- Treat audio and subtitles as first-class playback choices, not hidden advanced settings.
- Preview direct render, direct play, audio transcode, styled subtitle render, and subtitle burn-in before apply.
- Keep the forecast combined across both columns so users can understand the real outcome of the selected pair.
- Make forced/default/language flags scannable at TV distance.
- Keep sync and delay controls reachable from the selector without opening general settings.

### Playback Inspector

Purpose: make Vyrden's technical advantage visible.

Required sections:

- Session path: Direct Play, Remux, Audio Transcode, Subtitle Burn, or Video Transcode.
- Exact reason.
- Source stream details.
- Output stream details.
- Server CPU/GPU load.
- Buffer health.
- Suggested fix.

### Downloads

Purpose: make offline viewing reliable and understandable.

Required sections:

- Download queue.
- Completed downloads.
- Storage limit.
- Quality preset.
- Audio/subtitle selection.
- Error recovery and resume state.

## Remote Navigation Rules

- Every actionable item must have a strong focus state.
- Focus movement should be predictable in rows and columns.
- The hero should update as poster focus changes.
- Back should return to the previous context, not always home.
- Long-press should expose item actions only when the platform supports it.
- Text must be readable at 10 feet on a 55-inch TV.

## Movie And TV Specific Rules

- Posters use vertical aspect ratios.
- TV episodes use landscape stills when available.
- Series pages distinguish series, season, and episode clearly.
- "Continue Watching" must show episode title, series name, season/episode, and progress.
- Movie versions should not be hidden in admin-only screens.
- Audio and subtitles are first-class playback choices.

## Platform Notes

Apple tvOS and Android TV both rely heavily on focus-driven interaction. Vyrden should design focus states as part of the component, not as an afterthought.

Relevant external guidance:

- Apple tvOS Human Interface Guidelines.
- Android TV design and focus/navigation guidance.
