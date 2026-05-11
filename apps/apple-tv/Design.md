# Lorivo Apple TV Design

The Apple TV app should feel like the living-room version of the Lorivo server UI, not a separate product. It keeps the same cinema-black palette, warm typography, amber action color, teal focus color, and clear playback language, but adapts the layout for remote-first browsing and AVPlayer playback.

## Visual Rules

- Artwork leads the screen: backdrops and posters carry mood.
- UI chrome stays quiet: no dense admin panels, no generic file-browser language.
- Focus is large and obvious from a couch: teal ring, subtle lift, and scale.
- Cards stay restrained at 8 px corner radius.
- Playback state is present but quiet: Direct Play, Adaptive Stream, Remux, Audio Conversion, Subtitle Burn, or Video Conversion.
- Server ownership remains visible through pairing, route, and playback diagnostics, not through heavy infrastructure copy.

## Home

- Full-bleed backdrop from the focused item.
- Left-side title, metadata, and primary action.
- Horizontal rows for Continue Watching, Movies, TV Shows, and Recently Added.
- Focus changes update the hero.

## Detail

- Backdrop plus poster.
- Play/Resume is primary amber.
- Version, Audio, Subtitles, and Quality are peer controls, not hidden settings.
- Playback forecast appears before playback starts.

## Player

- AVPlayer owns the video layer.
- Overlay appears on remote interaction.
- Controls: timeline, play/pause, skip back, skip forward, audio, subtitles, quality, inspector.
- Route badge stays small: Direct Play or Adaptive Stream should be visible without distracting from the movie.

## Pairing

- First launch asks for server URL or discovered server.
- Pairing code flow comes before library browse.
- Manual URL should feel first-class, not like an error path.
