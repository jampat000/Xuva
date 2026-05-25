# Xuva Route Policy

This table documents the protected route policy introduced for local authorization.

Public routes remain readable without an authenticated session unless a later security task tightens them. Protected routes require a valid local session. Browser mutations also require a valid CSRF token.

Public client bootstrap routes:

| Pattern | Purpose |
| --- | --- |
| `GET /api/client/bootstrap` | Read-only native-client bootstrap for server identity, auth requirement, feature flags, profiles, and endpoint templates. |
| `POST /api/pairing/requests` | Creates a short-lived local pairing code for a TV/native app. |
| `GET /api/pairing/requests/{id}` | Polls pairing status for the requesting TV/native app. |
| `DELETE /api/pairing/requests/{id}` | Withdraws a still-pending pairing request; authorized by matching deviceId from the body or query. |

| Pattern | Group | Action | Roles |
| --- | --- | --- | --- |
| `GET /api/auth/session` | auth | `session.read` | admin, standard |
| `POST /api/auth/logout` | auth | `session.logout` | admin, standard |
| `GET /api/profiles` | auth | `profiles.list` | admin, standard |
| `POST /api/auth/switch-profile` | auth | `profile.switch` | admin, standard |
| `GET /api/users` | auth | `users.list` | admin |
| `POST /api/users` | auth | `user.create` | admin |
| `PATCH /api/users/{id}` | auth | `user.update` | admin |
| `DELETE /api/users/{id}` | auth | `user.delete` | admin |
| `POST /api/users/{id}/password` | auth | `user.password.update` | admin |
| `POST /api/users/{id}/pin` | auth | `user.pin.update` | admin |
| `POST /api/libraries` | libraries | `library.save` | admin |
| `DELETE /api/libraries/{id}` | libraries | `library.delete` | admin |
| `POST /api/libraries/{id}/scan` | libraries | `library.scan` | admin |
| `POST /api/libraries/movies/scan` | libraries | `library.scan.movies` | admin |
| `POST /api/libraries/tv/scan` | libraries | `library.scan.tv` | admin |
| `POST /api/libraries/scan` | libraries | `library.scan.all` | admin |
| `PUT /api/metadata/match` | metadata | `metadata.match` | admin |
| `POST /api/metadata/refresh` | metadata | `metadata.refresh` | admin |
| `POST /api/metadata/refresh-batch` | metadata | `metadata.refresh.batch` | admin |
| `GET /api/migrations/formats` | migration | `migration.formats` | admin |
| `GET /api/migrations/runs` | migration | `migration.runs.list` | admin |
| `GET /api/migrations/runs/{id}` | migration | `migration.run.read` | admin |
| `POST /api/migrations/dry-run` | migration | `migration.dry_run` | admin |
| `POST /api/migrations/import` | migration | `migration.import` | admin |
| `POST /api/migrations/runs/{id}/rollback` | migration | `migration.rollback` | admin |
| `GET /api/client/home` | client | `client.home` | admin, standard |
| `GET /api/client/movies/{id}` | client | `client.movie.detail` | admin, standard |
| `GET /api/client/movies/{id}/similar` | client | `client.movie.similar` | admin, standard |
| `GET /api/client/series/{id}` | client | `client.series.detail` | admin, standard |
| `GET /api/client/series/{id}/similar` | client | `client.series.similar` | admin, standard |
| `POST /api/client/playback/start` | client | `client.playback.start` | admin, standard |
| `PATCH /api/client/playback/{id}` | client | `client.playback.heartbeat` | admin, standard |
| `POST /api/client/playback/{id}/stop` | client | `client.playback.stop` | admin, standard |
| `PUT /api/settings` | settings | `settings.update` | admin |
| `PUT /api/settings/metadata-sources` | settings | `settings.metadata_sources.update` | admin |
| `POST /api/settings/hardware/test` | settings | `settings.hardware.test` | admin |
| `POST /api/remote/diagnostics` | remote | `remote.diagnostics.run` | admin |
| `POST /api/remote/wan` | remote | `remote.wan.lookup` | admin |
| `GET /api/settings/folders/browse` | settings | `settings.folders.browse` | admin |
| `GET /api/media-sources/{id}/stream` | media | `media.stream` | admin, standard |
| `GET /api/media-sources/{id}/download` | media | `media.download` | admin, standard |
| `GET /api/media-sources/{id}/adaptive/master.m3u8` | media | `media.adaptive.master` | admin, standard |
| `GET /api/media-sources/{id}/adaptive/{variant}` | media | `media.adaptive.variant` | admin, standard |
| `POST /api/media-sources/{id}/adaptive/session` | media | `media.adaptive.session` | admin, standard |
| `POST /api/adaptive/telemetry` | media | `media.adaptive.telemetry` | admin, standard |
| `POST /api/media-sources/{id}/stream-token` | media | `media.stream.token` | admin, standard |
| `GET /api/media-sources/{id}/subtitles/{index}` | media | `media.subtitle.stream` | admin, standard |
| `POST /api/media-sources/{id}/subtitles/{index}/convert` | media | `media.subtitle.convert` | admin, standard |
| `POST /api/media-sources/{id}/probe` | media | `media.probe` | admin |
| `POST /api/probes` | media | `probe.start` | admin |
| `GET /api/jobs` | jobs | `jobs.status` | admin |
| `POST /api/work` | work | `work.start` | admin |
| `DELETE /api/work/{id}` | work | `work.cancel` | admin |
| `GET /api/work/{id}/file` | work | `work.file` | admin, standard |
| `POST /api/downloads` | downloads | `download.start` | admin |
| `GET /api/downloads/{id}/file` | downloads | `download.file` | admin, standard |
| `GET /api/pairing/requests` | pairing | `pairing.list` | admin |
| `POST /api/pairing/requests/{id}/approve` | pairing | `pairing.approve` | admin |
| `POST /api/pairing/requests/{id}/deny` | pairing | `pairing.deny` | admin |
| `POST /api/pairing/qr` | pairing | `pairing.qr.generate` | admin |
| `GET /api/devices` | devices | `devices.list` | admin |
| `POST /api/devices/{id}/revoke` | devices | `devices.revoke` | admin |
| `GET /api/sessions` | sessions | `sessions.list` | admin, standard |
| `GET /api/sessions/{id}/inspector` | sessions | `sessions.inspector` | admin, standard |
| `POST /api/sessions` | sessions | `session.start` | admin, standard |
| `PATCH /api/sessions/{id}` | sessions | `session.update` | admin, standard |
| `DELETE /api/sessions/{id}` | sessions | `session.stop` | admin, standard |
| `PUT /api/playback/state/{id}` | playback | `playback.state.update` | admin, standard |
| `GET /api/metrics` | system | `system.metrics` | admin |
| `GET /api/events` | system | `system.events` | admin |
| `GET /api/architecture` | system | `system.architecture` | admin |
| `GET /api/system/status` | system | `system.status` | admin |
| `GET /api/remote/access` | remote | `remote.access.read` | admin |
| `GET /api/libraries` | libraries | `library.list` | admin |
| `GET /api/catalog/summary` | catalog | `catalog.summary` | admin |
| `GET /api/catalog/health` | catalog | `catalog.health` | admin |
| `GET /api/catalog/codecs` | catalog | `catalog.codecs` | admin |
| `GET /api/review` | metadata | `metadata.review` | admin |
| `GET /api/metadata/providers` | metadata | `metadata.providers` | admin |
| `GET /api/metadata/suggestions` | metadata | `metadata.suggestions` | admin |
| `GET /api/metadata/{kind}/{id}` | metadata | `metadata.read` | admin, standard |
| `GET /api/metadata/candidates` | metadata | `metadata.candidates` | admin |
| `GET /api/metadata/backfill` | metadata | `metadata.backfill.read` | admin |
| `POST /api/metadata/backfill` | metadata | `metadata.backfill.start` | admin |
| `DELETE /api/metadata/backfill` | metadata | `metadata.backfill.stop` | admin |
| `GET /api/artwork/{kind}/{id}` | metadata | `artwork.read` | admin, standard |
| `GET /api/trailers/{tmdbId}` | metadata | `trailer.read` | admin, standard |
| `GET /api/settings` | settings | `settings.read` | admin |
| `GET /api/settings/performance` | settings | `settings.performance.read` | admin |
| `GET /api/probes` | media | `probe.list` | admin |
| `GET /api/probes/{id}` | media | `probe.read` | admin |
| `GET /api/work` | work | `work.list` | admin |
| `GET /api/downloads` | downloads | `downloads.list` | admin |
| `GET /api/downloads/{id}` | downloads | `downloads.read` | admin |
| `GET /api/scans` | libraries | `scans.list` | admin |
| `GET /api/scans/{id}` | libraries | `scans.read` | admin |
| `GET /api/movies` | catalog | `movies.list` | admin, standard |
| `GET /api/movies/{id}` | catalog | `movies.read` | admin, standard |
| `GET /api/series` | catalog | `series.list` | admin, standard |
| `GET /api/series/{id}` | catalog | `series.read` | admin, standard |
| `GET /api/versions` | catalog | `versions.read` | admin, standard |
| `GET /api/media-sources` | media | `media.list` | admin, standard |
| `GET /api/media-sources/{id}` | media | `media.read` | admin, standard |
| `DELETE /api/media-sources/{id}` | media | `media.delete` | admin |
| `GET /api/media-sources/{id}/tracks` | media | `media.tracks` | admin, standard |
| `GET /api/media-sources/{id}/subtitles` | media | `media.subtitles` | admin, standard |
| `GET /api/media-sources/{id}/thumbnails/status` | media | `media.thumbnails.status` | admin, standard |
| `POST /api/media-sources/{id}/thumbnails/generate` | media | `media.thumbnails.generate` | admin |
| `GET /api/media-sources/{id}/thumbnails/sprite.jpg` | media | `media.thumbnails.sprite` | admin, standard |
| `GET /api/media-sources/{id}/thumbnails/thumbnails.vtt` | media | `media.thumbnails.vtt` | admin, standard |
| `GET /api/media-sources/{id}/thumbnails/chapters.vtt` | media | `media.thumbnails.chapters` | admin, standard |
| `GET /api/devices/profiles` | devices | `devices.profiles` | admin, standard |
| `GET /api/playback/recent` | playback | `playback.recent` | admin, standard |
| `GET /api/playback/state/{id}` | playback | `playback.state.read` | admin, standard |
| `GET /api/playback/decision` | playback | `playback.decision` | admin, standard |
| `GET /api/playback/route` | playback | `playback.route` | admin, standard |
| `GET /api/client/collections` | client | `client.collections.list` | admin, standard |
| `GET /api/client/collections/{id}` | client | `client.collection.detail` | admin, standard |
| `GET /api/client/people/{name}` | client | `client.person.detail` | admin, standard |
| `GET /api/client/search` | client | `client.search` | admin, standard |
| `GET /api/client/watchlist` | client | `client.watchlist.list` | admin, standard |
| `POST /api/client/watchlist` | client | `client.watchlist.add` | admin, standard |
| `DELETE /api/client/watchlist/{id}` | client | `client.watchlist.remove` | admin, standard |
| `POST /api/setup/complete` | setup | `setup.complete` | admin |
| `GET /api/backup/export` | backup | `backup.export` | admin |
| `POST /api/backup/import` | backup | `backup.import` | admin |
| `GET /api/notifications` | notifications | `notifications.list` | admin, standard |
| `POST /api/notifications/{id}/dismiss` | notifications | `notifications.dismiss` | admin, standard |
| `POST /api/notifications/dismiss-all` | notifications | `notifications.dismiss_all` | admin, standard |
| `GET /api/media-sources/{id}/chapters` | chapters | `chapters.read` | admin, standard |
| `POST /api/media-sources/{id}/chapters/analyze` | chapters | `chapters.analyze` | admin |
| `PATCH /api/users/me/preferences` | auth | `user.preferences.update` | admin, standard |

Audit events are emitted as `audit.route` with:

- `userId`
- `username`
- `role`
- `method`
- `path`
- `pattern`
- `group`
- `action`
- `result`
- `reason`
- `createdAt`
