# Apple TV Alpha

Apple TV is Xuva's first native playback client target. It should prove that Xuva is more than a server admin console: a user can install the local server, add media, pair the TV app, browse their library, and play a movie or episode from the couch.

## Product Goals

- Native living-room playback feels fast, quiet, and premium.
- Pairing works on a local network without a cloud account.
- Manual server URL entry works for advanced users and remote self-hosted routes.
- Playback uses direct play or HLS adaptive streams before falling back to heavier server conversion.
- Every fallback is explained in plain language.

## First App Scope

- Server discovery or manual URL entry.
- Pairing code flow against the local Xuva server.
- tvOS capability registration.
- Movies and TV Shows home rows.
- Movie detail, series detail, season, and episode screens.
- Version, audio, and subtitle selection.
- AVPlayer playback for direct streams and HLS adaptive streams.
- Playback heartbeat, resume position, and watched/unwatched updates.
- Clear connection and playback error states.

## Server Contracts Needed

- Stable client bootstrap endpoint with server identity, feature flags, and URLs: `GET /api/client/bootstrap?clientProfile=apple-tv`.
- Pairing create/claim/approve flow for TV devices.
- Client capability profile registration for tvOS.
- Catalog endpoints sized for TV home rows and detail screens.
- Image endpoints that provide poster, backdrop, and thumbnail sizes appropriate for TV.
- Playback start endpoint that returns the selected route, signed stream URLs, audio/subtitle options, and heartbeat interval.
- HLS master playlist path for adaptive sessions.
- Session heartbeat and stop endpoints.
- Resume and watched state endpoints.

## Client Bootstrap

`GET /api/client/bootstrap?clientProfile=apple-tv` is public and read-only so a TV app can validate the local server before pairing or login. It returns:

- server identity, base URL, LAN URLs, and start time.
- whether authentication is required.
- the requested client profile and all known playback profiles.
- feature flags for direct play, HLS adaptive streaming, resume, watched state, and relay-free operation.
- endpoint templates for auth, catalog, playback decisions, sessions, direct streams, HLS manifests, tracks, subtitles, and remote access diagnostics.

It must not return library contents, user data, tokens, or media paths.

## Local Pairing

The first pairing loop is local and cloud-free:

1. Apple TV calls `POST /api/pairing/requests` with `deviceName` and `clientProfile`.
2. Apple TV displays the returned six-digit `code`.
3. Web admin opens `Settings -> Devices` and approves or denies the request.
4. Apple TV polls `GET /api/pairing/requests/{id}` until the status changes.
5. Approved responses include a `deviceId`; the long-lived authenticated device credential remains a later hardening task.

The SwiftUI starter in `apps/apple-tv/Sources` already implements this first loop against the local API: manual URL, bootstrap, pairing request creation, code display, polling, and transition into the home shell after approval.

Admin-only routes:

- `GET /api/pairing/requests`
- `POST /api/pairing/requests/{id}/approve`
- `POST /api/pairing/requests/{id}/deny`

## TV Home

`GET /api/client/home?clientProfile=apple-tv` returns a native-client home payload:

- `hero`: first useful item for the full-screen hero.
- `rows`: Continue Watching, Movies, TV Shows, and Recently Added.
- normalized row items with `id`, `kind`, `title`, `subtitle`, artwork URLs, optional `mediaSourceId`, and route label.
- endpoint templates for movie detail, series detail, and playback route.

This route is protected when auth is enabled because it exposes library contents. The SwiftUI starter loads it after local pairing succeeds.

## TV Detail And Playback Start

The next native playback contract is now present on the server:

- `GET /api/client/movies/{id}`
- `GET /api/client/series/{id}`
- `POST /api/client/playback/start`
- `PATCH /api/client/playback/{id}`
- `POST /api/client/playback/{id}/stop`

Detail payloads return TV-shaped item metadata, version payloads, audio tracks, embedded subtitle tracks, discovered sidecar subtitles, and a playback decision preview for each version.

Playback start returns:

- `sessionId`
- `heartbeatUrl`
- `stopUrl`
- `playbackStateUrl`
- `heartbeatIntervalMs`
- selected playback `decision`
- a route payload for direct, adaptive, remux, or transcode playback

Current hardening limit: when auth is enabled, native playback start fails explicitly until the paired-device credential can exchange for protected stream URLs. That is intentional and tracked as remaining device-auth work rather than silently issuing unusable URLs.

## Deferred

- Offline downloads.
- Mobile remote control.
- App Store purchase/subscription flows.
- Vendor-hosted relay.
- Multi-server cloud sync.

## Acceptance Test

1. Start the Xuva desktop alpha server.
2. Add a Movies library and scan it.
3. Install or sideload the Apple TV alpha app.
4. Pair the Apple TV app to the server on the LAN.
5. Browse Movies from the TV app.
6. Play one direct-play source.
7. Play one adaptive HLS source.
8. Stop playback and verify resume position appears in the web admin and TV app.
