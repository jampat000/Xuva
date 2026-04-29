# Vyrden Route Policy

This table documents the protected route policy introduced for local authorization.

Public routes remain readable without an authenticated session unless a later security task tightens them. Protected routes require a valid local session. Browser mutations also require a valid CSRF token.

| Pattern | Group | Action | Roles |
| --- | --- | --- | --- |
| `GET /api/auth/session` | auth | `session.read` | admin, standard |
| `POST /api/auth/logout` | auth | `session.logout` | admin, standard |
| `POST /api/libraries` | libraries | `library.save` | admin |
| `DELETE /api/libraries/{id}` | libraries | `library.delete` | admin |
| `POST /api/libraries/{id}/scan` | libraries | `library.scan` | admin |
| `POST /api/libraries/movies/scan` | libraries | `library.scan.movies` | admin |
| `POST /api/libraries/tv/scan` | libraries | `library.scan.tv` | admin |
| `POST /api/libraries/scan` | libraries | `library.scan.all` | admin |
| `PUT /api/metadata/match` | metadata | `metadata.match` | admin |
| `POST /api/metadata/refresh` | metadata | `metadata.refresh` | admin |
| `POST /api/metadata/refresh-batch` | metadata | `metadata.refresh.batch` | admin |
| `PUT /api/settings` | settings | `settings.update` | admin |
| `POST /api/settings/hardware/test` | settings | `settings.hardware.test` | admin |
| `POST /api/remote/wan` | remote | `remote.wan.lookup` | admin |
| `GET /api/media-sources/{id}/stream` | media | `media.stream` | admin, standard |
| `GET /api/media-sources/{id}/subtitles/{index}` | media | `media.subtitle.stream` | admin, standard |
| `POST /api/media-sources/{id}/probe` | media | `media.probe` | admin |
| `POST /api/probes` | media | `probe.start` | admin |
| `POST /api/work` | work | `work.start` | admin |
| `GET /api/work/{id}/file` | work | `work.file` | admin, standard |
| `POST /api/downloads` | downloads | `download.start` | admin |
| `GET /api/downloads/{id}/file` | downloads | `download.file` | admin, standard |
| `GET /api/sessions` | sessions | `sessions.list` | admin, standard |
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
