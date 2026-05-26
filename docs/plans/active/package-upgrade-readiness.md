# Package And Upgrade Readiness Plan

## Goal

Make Xuva releasable as a server/client product with robust upgrades and predictable packaging:

- Windows desktop/server package.
- Docker image.
- Later Linux bare-metal packages.

Early package releases use `v0.0.x` tags. `v1.0.0` is reserved until install, upgrade, rollback, and local file/NAS workflows are proven.

The release package contains the Xuva server, embedded web UI, runtime dependencies, and desktop shell where applicable. It does not contain Apple, Android, iOS, tvOS, or other native app builds.

## Packaging Direction

### Windows

- Ship an unsigned per-user installer/package until paid code signing is viable.
- Publish SHA256 checksums with every GitHub Release artifact.
- Build from a tagged source tree with version metadata embedded into the server binary.
- Include all normal-user prerequisites:
  - `Xuva.exe` desktop shell.
  - `xuva-server.exe`.
  - Embedded Svelte web UI.
  - Desktop shell runtime.
  - `ffmpeg.exe` and `ffprobe.exe`.
  - CA certificate bundle if required by the runtime.
  - Default runtime directory creation.
- Do not require users to install Go, Node.js, npm, FFmpeg, FFprobe, SQLite tools, or Docker for the Windows desktop path.
- Run in the signed-in user session by default, not as a Windows service, so local disks, mapped drives, removable drives, SMB shares, and UNC paths match what the operator can browse.
- Keep mutable state outside the install directory:
  - database
  - logs
  - settings
  - backups
  - cache
  - transcode work
  - generated metadata
- Preferred Windows runtime home is `C:\ProgramData\Xuva` so server identity, approvals, libraries, backups, logs, and upgrade state are per device/server rather than per Windows user. `%LOCALAPPDATA%\Xuva` is fallback-only when the shared runtime home is unavailable.
- Windows desktop process remains a signed-in user session app by default so mapped drives, removable disks, and UNC/NAS paths match the operator's File Explorer access.
- In-app upgrades should update application binaries through the desktop shell/update framework while preserving the runtime home and running pre-upgrade database backup/integrity checks before replacing the server binary.

### Docker

- Ship an all-in-one server image with embedded web UI.
- Include FFmpeg, FFprobe, CA certificates, timezone data, and a container healthcheck.
- Keep `/data` as the persistent runtime volume.
- Treat media libraries as explicit bind mounts.
- Do not bundle native client source trees or build artifacts.

### Linux Bare Metal

- Defer until Windows and Docker are stable.
- Same runtime contract: package prerequisites, run under an operator-controlled user, and require explicit media/cache/transcode paths.

## Upgrade Schema Direction

Replace the current ad-hoc SQL slice with a versioned migration ledger before public packaging:

- `schema_migrations` table with migration id, name, checksum, applied timestamp, and duration.
- Ordered immutable migrations.
- Transaction per migration where SQLite supports it.
- Hard failure on checksum mismatch or partial state.
- `PRAGMA integrity_check` before and after upgrade.
- Automatic pre-upgrade database backup.
- Fixture-based upgrade tests from older database snapshots.
- Public version/status endpoint exposing app version, commit, schema version, data directory, and upgrade state.
- Structured JSONL file logging under the runtime home for upgrade diagnostics and support bundles.

## MediaMop Findings To Reuse

MediaMop has useful release patterns that fit Xuva:

- Tag-driven release workflow.
- Windows package plus Docker image from the same release tag.
- Unsigned Windows installer accepted as a documented constraint.
- SHA256 checksums for release artifacts.
- Bundled FFmpeg/FFprobe instead of assuming host installation.
- Per-user desktop runtime instead of Windows service mode.
- Docker healthcheck and container smoke testing.
- Upgrade completion should verify the running app version, not just installer exit status.

MediaMop implementation details that should not be copied directly:

- Python virtualenv/PyInstaller flow.
- Alembic-specific migration stack.
- Legacy updater service path.

## Required Work

1. Build versioned SQLite schema migration package.
2. Add app/system version endpoint.
3. Add pre-upgrade backup and restore/rollback rehearsal around real fixture DBs.
4. Add Windows package script that builds web assets, compiles Go, vendors FFmpeg/FFprobe with checksum verification, and assembles the first desktop installer/package.
5. Add release workflow artifacts for Windows installer/package and checksums.
6. Add Docker healthcheck and container smoke test to CI/release.
7. Document Windows filesystem access model for local disks, mapped drives, UNC paths, NAS, and removable media.
8. Add package verification checklist covering clean install, upgrade, rollback, uninstall/reinstall, and first-run setup.

## Non-Goals

- Paid code signing.
- Windows service mode as the default desktop install.
- Bundling Apple, Android, iOS, tvOS, or native client builds.
- Requiring cloud accounts or vendor relay infrastructure for local use.
