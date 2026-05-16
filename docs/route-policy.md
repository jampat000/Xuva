# Xuva Route Policy

This table documents the protected route policy introduced for local authorization.

Public routes remain readable without an authenticated session unless a later security task tightens them. Protected routes require a valid local session. Browser mutations also require a valid CSRF token.

Public client bootstrap routes:

| Pattern | Purpose |
| --- | --- |
| `GET /api/client/bootstrap` | Read-only native-client bootstrap for server identity, auth requirement, feature flags, profiles, and endpoint templates. |
| `POST /api/pairing/requests` | Creates a short-lived local pairing code for a TV/native app. |
| `GET /api/pairing/requests/{id}` | Polls pairing status for the requesting TV/native app. |

| Pattern | Group | Action | Roles |
| --- | --- | --- | --- |
| `GET /api/auth/session` | auth | `session.read` | admin, standard |
| `POST /api/auth/logout` | auth | `session.logout` | admin, standard |
| `GET /api/users` | auth | `users.list` | admin |
| `POST /api/users` | auth | `user.create` | admin |
| `PATCH /api/users/{id}` | auth | `user.update` | admin |
| `DELETE /api/users/{id}` | auth | `user.delete` | admin |
| `POST /api/users/{id}/password` | auth | `user.password.update` | admin |
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
| `GET /api/client/series/{id}` | client | `client.series.detail` | admin, standard |
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
| `GET /api/media-sources/{id}/adaptive/master.m3u8` | media | `media.adaptive.master` | admin, standard |
| `GET /api/media-sources/{id}/adaptive/{variant}` | media | `media.adaptive.variant` | admin, standard |
| `POST /api/media-sources/{id}/adaptive/session` | media | `media.adaptive.session` | admin, standard |
| `POST /api/adaptive/telemetry` | media | `media.adaptive.telemetry` | admin, standard |
| `POST /api/media-sources/{id}/stream-token` | media | `media.stream.token` | admin, standard |
| `GET /api/media-sources/{id}/subtitles/{index}` | media | `media.subtitle.stream` | admin, standard |
| `POST /api/media-sources/{id}/subtitles/{index}/convert` | media | `media.subtitle.convert` | admin, standard |
| `POST /api/media-sources/{id}/probe` | media | `media.probe` | admin |
| `POST /api/probes` | media | `probe.start` | admin |
| `POST /api/work` | work | `work.start` | admin |
| `DELETE /api/work/{id}` | work | `work.cancel` | admin |
| `GET /api/work/{id}/file` | work | `work.file` | admin, standard |
| `POST /api/downloads` | downloads | `download.start` | admin |
| `GET /api/downloads/{id}/file` | downloads | `download.file` | admin, standard |
| `GET /api/pairing/requests` | pairing | `pairing.list` | admin |
| `POST /api/pairing/requests/{id}/approve` | pairing | `pairing.approve` | admin |
| `POST /api/pairing/requests/{id}/deny` | pairing | `pairing.deny` | admin |
| `GET /api/devices` | devices | `devices.list` | admin |
| `POST /api/devices/{id}/revoke` | devices | `devices.revoke` | admin |
| `GET /api/sessions` | sessions | `sessions.list` | admin, standard |
| `GET /api/sessions/{id}/inspector` | sessions | `sessions.inspector` | admin, standard |
| `POST /api/sessions` | sessions | `session.start` | admin, standard |
| `PATCH /api/sessions/{id}` | sessions | `session.update` | admin, standard |
| `DELETE /api/sessions/{id}` | sessions | `session.stop` | admin, standard |
| `PUT /api/playback/state/{id}` | playback | `playback.state.update` | admin, standard |
| `GET /play/{id}` | playback | `player.open` | admin, standard |

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
