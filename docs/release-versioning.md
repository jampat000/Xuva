# Release Versioning

Xuva uses semantic version tags in the form `vMAJOR.MINOR.PATCH`.

## Current Track

Until the installer, Docker image, upgrade flow, and local library workflows are proven on real machines, Xuva releases use:

```text
v0.0.x
```

This keeps the product clearly pre-1.0 while still allowing orderly public builds, upgrade testing, release notes, and support references.

Patch numbers should increase for every tagged package release:

```text
v0.0.1
v0.0.2
v0.0.3
```

Do not skip to `v0.1.x` or `v1.x.x` just to make the product look mature.

## Promotion To 1.x

Move to `v1.0.0` only when these are true:

- Windows package install, first run, restart, upgrade, rollback, and uninstall/reinstall are reliable.
- Docker install and upgrade docs are tested against the published image.
- SQLite schema migrations have fixture-based upgrade coverage.
- Bundled prerequisites are verified: FFmpeg, FFprobe, embedded web UI, and runtime directories.
- LAN access, canonical browser origin behavior, and local file/NAS browsing are predictable.
- The release workflow publishes all required artifacts and checksums from a tag.

After `v1.0.0`, use normal semantic versioning:

- Patch: safe fixes with no intentional behavior or API break.
- Minor: backward-compatible features.
- Major: intentional breaking changes or upgrade-contract changes.

## Tagging Rule

The release workflow is tag-driven. Normal early releases should be annotated tags:

```powershell
git tag -a v0.0.x -m "Xuva v0.0.x"
git push origin v0.0.x
```

Every release tag must point at a commit that has passed the release-readiness checks.
