# Package Verification

This checklist verifies that release artifacts work as packages, not just as source builds.

## Version Track

Early package releases use `v0.0.x` tags. Do not tag `v1.0.0` until install, upgrade, rollback, and filesystem workflows are proven on real machines.

## Windows Package

Artifact expectations:

- Unsigned package zip exists on the GitHub Release.
- Matching `.sha256` file exists.
- Package contains `app/xuva.exe`.
- Package contains `app/Start-Xuva.ps1`.
- Package contains `app/bin/ffmpeg.exe`.
- Package contains `app/bin/ffprobe.exe`.
- Package does not contain Apple, Android, iOS, tvOS, or native client build outputs.

Automated package verification:

```powershell
./packaging/windows/verify-package.ps1 -PackagePath "dist/windows/xuva-v0.0.x-win-x64.zip"
```

Local package builds restore `server/internal/webapp/static-next` after compiling the Go binary so release rehearsal does not leave generated web assets in the working tree. Use `-LeavePublishedStatic` only when intentionally refreshing committed embedded assets.

Clean install smoke:

1. Verify package checksum.
2. Extract to a clean directory.
3. Run `powershell -ExecutionPolicy Bypass -File .\app\Start-Xuva.ps1`.
4. Open `http://localhost:8097/api/system/version`.
5. Confirm `version` equals the release tag and `schemaVersion` is populated.
6. Open `http://localhost:8097/`.
7. Complete first-run setup.
8. Add a local media folder.
9. Add a mapped drive or UNC NAS media folder.
10. Restart Xuva and verify settings/database persisted.

Upgrade smoke:

1. Install/extract previous `v0.0.x` package.
2. Complete setup and add at least one library.
3. Stop Xuva.
4. Start the newer package against the same `XUVA_DATA_DIR`.
5. Confirm `/api/system/version` reports the newer tag.
6. Confirm `data/backups/schema-upgrade-*.db` exists when a schema migration was pending.
7. Confirm users, libraries, devices, settings, and playback state remain intact.

Rollback smoke:

1. Preserve the pre-upgrade package.
2. Preserve the pre-upgrade data directory or schema backup.
3. Restore the data directory or database backup.
4. Start the previous package.
5. Confirm setup does not reappear for an established install.

Local rollback rehearsal:

```powershell
./tools/rehearse-install-upgrade-rollback.ps1 -DataDir "data"
```

If `xuva.db-wal` exists, stop Xuva before running a full database rehearsal. Use `-SkipDatabase` only when validating settings rollback while the server is still running.

## Docker Package

Artifact expectations:

- GHCR image has the release tag.
- Image includes FFmpeg and FFprobe.
- Image exposes `/api/health`.
- Image exposes `/api/system/version`.
- Image declares a healthcheck.
- Image does not include native app source trees or build outputs.

Clean install smoke:

```bash
docker run --rm \
  -p 8097:8097 \
  -v xuva_data:/data \
  -v /srv/media/movies:/movies:ro \
  -v /srv/media/tv:/tv:ro \
  ghcr.io/jampat000/xuva:v0.0.x
```

Checks:

- `GET /api/health` returns 200.
- `GET /api/system/version` returns the release tag.
- First-run setup works.
- Libraries can be added from mounted container paths.
- Settings/database persist after container replacement.

Upgrade smoke:

1. Start previous image tag with a persistent `/data` volume.
2. Complete setup and add libraries.
3. Replace container with newer image tag using the same `/data`.
4. Confirm version endpoint reports the newer tag.
5. Confirm schema backup is present when a migration was pending.
6. Confirm libraries and settings remain intact.

## Required Local Checks Before Tagging

```powershell
./tools/check.ps1 -Release -SkipFrontendInstall
```

The check runner intentionally executes Go commands from `server/`, because that is the Go module root.

## Release Dry Run

Before creating a real tag, run the Release workflow manually with:

```powershell
gh workflow run Release --ref main -f version=v0.0.1 -f dry_run=true
gh run watch --exit-status
```

The dry run builds and verifies the unsigned Windows package and builds/smokes the Docker image without publishing a GitHub Release or pushing GHCR tags.
