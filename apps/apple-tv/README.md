# Xuva Apple TV

Native Apple TV client.

Apple TV is the first native playback client target for Xuva alpha because it is available for home testing and fits the premium living-room product direction. The first app should prove the end-to-end product loop: install server, add libraries, pair Apple TV, browse media, and play through direct or adaptive streaming.

Source files live in `Sources/` and are ready to add to an Xcode tvOS SwiftUI app target. The visual system mirrors the server UI: cinema black, warm text, restrained cards, amber primary actions, teal focus, poster-led rows, and quiet playback diagnostics.

Responsibilities:

- Browse libraries.
- Play media through native Apple playback APIs where possible.
- Report client playback capabilities.
- Support local pairing.
- Support manual server URL entry.
- Expose playback quality and subtitle controls.

Alpha scope:

- Pair with a local Xuva server by code or manual URL.
- Register a tvOS playback capability profile.
- Browse Movies and TV Shows.
- Open movie, series, season, and episode detail screens.
- Start playback with AVPlayer using direct files or HLS adaptive streams.
- Select source version, audio track, and subtitle track before playback.
- Report playback heartbeat, resume position, and watched state.
- Show plain-language playback fallback when the server must convert video.

Current source status:

- Connects to `GET /api/client/bootstrap?clientProfile=apple-tv`.
- Creates `POST /api/pairing/requests`.
- Displays the six-digit pairing code.
- Polls `GET /api/pairing/requests/{id}` until approved.
- Enters the home shell after approval returns a `deviceId`.
- Loads `GET /api/client/home?clientProfile=apple-tv` and renders returned TV rows.
- Server contracts now exist for `GET /api/client/movies/{id}`, `GET /api/client/series/{id}`, `POST /api/client/playback/start`, `PATCH /api/client/playback/{id}`, and `POST /api/client/playback/{id}/stop`.

Deferred:

- Offline downloads.
- Purchases or subscription flows.
- Multi-server account sync.
- Vendor relay access.
