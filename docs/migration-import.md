# Migration Import

Lorivo migration import is designed to lower first-use friction for people moving from Plex, Emby, or Jellyfin without pretending to be a full ecosystem clone.

## Supported Today

- Watched state import.
- Resume position import.
- Movie external identifier import.
- Series external identifier import when the source row is an episode match.
- Dry-run preview with conflict classification.
- Selective execution by importable row.
- Rollback for completed runs.
- Post-import verification report.

## Supported Bundle Format

Schema: `lorivo.migration.v1`

Supported `source` values:

- `plex`
- `emby`
- `jellyfin`
- `generic`

Payload shape:

```json
{
  "schema": "lorivo.migration.v1",
  "source": "plex",
  "createdAt": "2026-04-30T09:00:00Z",
  "items": [
    {
      "id": "plex-heat",
      "kind": "movie",
      "title": "Heat",
      "year": 1995,
      "path": "Heat (1995)/Heat.1995.1080p.BluRay.mkv",
      "externalIds": {
        "tmdb": "949",
        "imdb": "tt0113277"
      },
      "watched": true,
      "progressSeconds": 10200,
      "durationSeconds": 10200,
      "lastPlayedAt": "2026-04-01T10:00:00Z"
    }
  ]
}
```

## Validation Rules

- The payload must be valid JSON.
- `schema` must be `lorivo.migration.v1`.
- Every item must target a `movie` or `episode`.
- Every item must include at least one safe locator:
  - exact path or relative path
  - external identifier
  - movie title plus year
  - series title plus season and episode numbers
- Playback imports require exactly one local media-source version.
- Metadata imports write identifiers only to matched local items.

## Matching Strategy

### Movies

1. Exact local path or relative path.
2. External identifiers.
3. Title plus year.

### Episodes

1. Exact local path or relative path.
2. Series external identifiers plus season and episode numbers.
3. Series title plus season and episode numbers.

If a playback row matches a logical movie or episode but Lorivo cannot safely pick one local file version, the row stays in conflict until the import bundle is narrowed.

## Conflict Classes

- unsupported item kind
- missing locator
- local item not found
- ambiguous match
- duplicate target
- version required for playback state
- no data for the selected import scopes

## Rollback Model

Completed runs store per-target backups before applying changes.

Rollback restores:

- previous playback state rows
- previous metadata external identifier rows

Failed imports do not partially apply because execution is transactional.

## Out Of Scope

- vendor cloud accounts
- plugins, playlists, collections, or users
- live sync with Plex, Emby, or Jellyfin
- one-click extraction from proprietary databases

## Admin Workflow

1. Open `Settings -> Advanced`.
2. Load or paste a migration bundle.
3. Run dry-run preview.
4. Review importable rows and conflicts.
5. Run the import.
6. Review verification summary.
7. Roll back the run if needed.
