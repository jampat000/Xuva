# Competitive Notes

## Positioning

Xuva is a local-first movie and TV server/client platform. It should not try to beat Plex, Emby, and Jellyfin by copying their surface area. It should beat them by making the painful parts of personal media obvious, explainable, and controllable.

The product promise:

- Your server stays yours.
- Playback decisions are visible before you press play.
- Remote access is user-owned and guided, not vendor-hosted.
- Library health is treated as a first-class dashboard, not hidden admin work.
- TV and movie experiences are polished enough for a living room, but detailed enough for power users.

## Competitor Layout Patterns

### Plex

Plex is strongest at polished browsing, client reach, and non-technical setup. Its layout usually centers around home rows, pinned libraries, continue watching, recommendations, player surfaces, account features, and server settings behind a separate admin area.

Xuva should learn from:

- Strong first-run onboarding.
- Continue watching as a first-class surface.
- Familiar poster walls, detail pages, and player controls.
- Device coverage and simple client discovery.

Xuva should avoid:

- Personal-library UI competing with streaming-service clutter.
- Local playback depending on cloud account state.
- Hiding why media is transcoding.
- Making remote access feel like magic until it fails.

### Emby

Emby is closer to a traditional self-hosted media server with commercial polish and paid feature gates. It has server dashboard, libraries, users, devices, metadata tools, live TV, playback/transcoding settings, and client apps.

Xuva should learn from:

- Clear server administration.
- Practical metadata management.
- User/device administration.
- Feature gating that can support a business.

Xuva should avoid:

- Dense admin surfaces that feel older than the client experience.
- Settings spread across too many places.
- Paid boundaries that surprise users during core playback.

### Jellyfin

Jellyfin is strongest on open-source trust, local control, and avoiding vendor lock-in. It is weaker on polish, client consistency, and support expectations.

Xuva should learn from:

- Local-first trust.
- No required cloud dependency.
- Transparent server behavior.
- Community-friendly file and codec support.

Xuva should avoid:

- Making users become their own support department.
- Rough client UX, especially on TV devices.
- Weak download/offline flows.
- Metadata and subtitle workflows that require manual forum archaeology.

## Repeated User Complaint Themes

These are the complaint themes that should become product requirements:

- Remote access is confusing, fragile, or too dependent on a vendor account.
- Users do not know why something is transcoding.
- Subtitles unexpectedly force transcoding or burn-in.
- Metadata matches are wrong, hard to correct, or inconsistent between movies and TV.
- Users want clear ratings from multiple trusted sources instead of one opaque score.
- Large libraries scan slowly or make the server feel busy.
- Playback stutters because background jobs compete with active sessions.
- Downloads are unreliable, unclear, or locked behind confusing feature boundaries.
- Different clients support different formats without explaining the difference.
- Admin settings are powerful but hard to navigate.
- UI can feel cluttered, dated, or not designed for a living room.
- Paid features can feel like they interrupt a core personal-media workflow.

## Xuva Responses

### Playback Transparency

Every playable item must show:

- Direct play, remux, audio transcode, subtitle burn, or video transcode.
- The exact reason.
- The selected client profile.
- The source container, video codec, bitrate, audio track, and subtitle track.
- The expected server impact.
- Suggested fixes when playback is expensive.

Acceptance test: a user should never need a forum post to understand why the server is working hard.

### Subtitle Control

Subtitles need their own workflow, not a hidden player dropdown.

Required behavior:

- Show text vs image subtitles.
- Show direct subtitle support vs burn-in requirement.
- Warn when PGS/VobSub will cause burn-in on a client.
- Prefer external SRT/WEBVTT when available.
- Support subtitle extraction/conversion as a background job.

Acceptance test: changing subtitles should update the playback forecast before playback starts.

### Metadata Trust

Metadata should be visible, explainable, and correctable.

Required behavior:

- Filename/local fallback records.
- Provider records with confidence scores.
- Manual override that clearly wins.
- Review queue for uncertain matches.
- Movie and TV-specific matching rules.
- Metadata/artwork cache path configurable outside the install drive.
- External IDs for IMDb, TMDB, TVDB, and provider-specific records.
- Ratings as separate source-attributed values, not one blended score.

Acceptance test: a bad match should be fixable from the item page without leaving the app.

### Ratings Clarity

Xuva should show ratings the way users actually talk about movies and TV: multiple trusted signals side by side.

Required behavior:

- IMDb rating.
- Rotten Tomatoes critics score.
- Rotten Tomatoes audience score.
- TMDB community rating.
- Metacritic score when available.
- TVDB rating for series/episodes when available.
- Source attribution and last-updated timestamp.
- Per-user preference for which ratings appear first.

Acceptance test: a user should be able to compare IMDb, Rotten Tomatoes, and TMDB without guessing which provider Xuva used.

### Library Health Command Centre

The dashboard should be useful before, during, and after playback.

Required signals:

- What is playing now.
- Active jobs and whether they compete with playback.
- Source count, movies, series, episodes.
- Probe coverage.
- Direct-play percentage.
- High bitrate / 4K / subtitle / unsupported counts.
- CPU, memory, heap, disk capacity.
- Runtime folders for metadata, cache, transcode temp, downloads, and scratch.

Acceptance test: opening the dashboard should tell a user what is healthy, what is busy, and what needs attention.

### Remote Access Without Xuva Servers

Xuva should not host relays or maintain user media infrastructure.

Required behavior:

- LAN URL.
- WAN IP detection.
- Port and reachability diagnostics.
- Router/NAT warning states.
- Reverse proxy guidance.
- Tailscale/WireGuard/manual tunnel guidance.
- Clear language that users own remote access.

Acceptance test: Xuva helps configure and diagnose remote access without becoming a relay provider.

### Large Library Performance

The app must assume thousands of files.

Required behavior:

- Separate Movies and TV scan flows.
- Shared walker but separate classifiers.
- Background jobs with bounded queues.
- Scan, probe, transcode, download, and metadata jobs separated.
- Playback-critical work prioritized over background work.
- Incremental rescans and idempotent upserts.
- Visible progress and job history.

Acceptance test: a 5,000+ file library should not make the UI feel frozen or mysterious.

### Paid Feature Boundary

Free must feel complete for local playback.

Free baseline:

- Local server.
- Movie and TV libraries.
- Metadata matching.
- Direct play.
- Basic remux/transcode.
- Local users.
- Web admin.
- Remote access helpers.

Premium candidates:

- Hardware transcoding.
- HDR tone mapping.
- Offline downloads.
- Advanced subtitle conversion.
- Intro/credits detection.
- Advanced parental controls.
- Migration tools.
- Enhanced server health/history.

Acceptance test: users should understand the value of premium without feeling core local playback is held hostage.

## Implementation Backlog

### Next Functional Slice

1. Make the web player a real app surface, not a plain HTML video page.
2. Wire playback sessions into live dashboard updates.
3. Add item-level play buttons that open the player for the selected movie/episode/version.
4. Show playback forecast before playback starts.
5. Persist resume progress and watched/unwatched state.
6. Add audio/subtitle selection into the play decision.
7. Make the dashboard update live through SSE.

### Server Work

1. Finish SSE consumer in the web UI.
2. Add session heartbeat cleanup for abandoned sessions.
3. Add playback decision endpoint inputs for selected audio/subtitle track.
4. Add remux path that can produce streamable MP4/HLS when direct stream is not enough.
5. Make download jobs resumable and visible.
6. Add metadata provider integration behind the existing metadata records.
7. Add configurable metadata/artwork cache use.

### UI Work

1. Replace the raw `/play/{id}` page with the Xuva player shell.
2. Add play/resume/mark watched buttons on movie and episode detail screens.
3. Add source, audio, and subtitle selectors into item detail.
4. Add a real review queue screen for metadata.
5. Add a live activity screen using server events.
6. Add setup wizard for library folders, runtime folders, and remote access diagnostics.

## Sources To Recheck Before Public Claims

These notes are product planning, not marketing copy. Before public comparison pages, recheck official docs and current product behavior for Plex, Emby, and Jellyfin.
