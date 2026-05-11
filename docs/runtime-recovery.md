# Runtime Recovery

Lorivo persists operational runtime state in SQLite so a restart does not hide active playback, conversions, media checks, or prepared download jobs.

## Stored Entities

Runtime snapshots are written to `runtime_entities` with an entity type and JSON payload:

- `session`: active playback sessions and their latest heartbeat/update.
- `transcode`: remux and transcode work.
- `probe`: media check batches and single-file checks.
- `download`: optimized/offline copy jobs.

The table is append-safe through upserts: each lifecycle transition replaces the previous snapshot for the same entity id.

## Startup Reconciliation

On startup each persistent service reloads its own snapshots.

- Recent sessions are restored to the active session list.
- Sessions with expired heartbeats are marked `stale` and are not shown as active.
- Queued or running transcode, probe, and download jobs are moved to `failed` because the worker process that owned them no longer exists.
- Completed and failed records remain visible for diagnostics until cleanup.

This prevents zombie `running` jobs after an unclean shutdown while still preserving enough recent state for the dashboard and support evidence.

## Heartbeat And Cleanup Policy

Playback session updates act as heartbeats. The default stale threshold is 15 minutes. Runtime maintenance runs at startup and every 5 minutes:

- active sessions older than the heartbeat threshold become `stale`;
- terminal session/job records older than 24 hours are removed.

## Rollback

The migration only adds `runtime_entities` and indexes. Rolling back the runtime persistence feature can leave the table in place because older builds ignore it. If a hard rollback is required, stop Lorivo and drop `runtime_entities`; no media library, metadata, or user records depend on it.
