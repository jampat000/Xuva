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

