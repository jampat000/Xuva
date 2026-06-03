# Release Versioning

Xuva uses [semantic versioning](https://semver.org) — `vMAJOR.MINOR.PATCH`.

## Cadence

- **Patch** (`v1.0.x`) — safe fixes, no intentional behaviour or API change. Released as needed (typically weekly during active development, monthly during stabilisation).
- **Minor** (`v1.x.0`) — backward-compatible features. Released when a meaningful set of new functionality is ready.
- **Major** (`vX.0.0`) — intentional breaking changes or upgrade-contract changes. Released rarely and with prior notice.

Pre-release tags follow the `-(alpha|beta|rc).N` suffix convention:

```text
v1.1.0-alpha.1
v1.1.0-beta.2
v1.1.0-rc.1
```

A pre-release tag publishes its artifacts but never updates the `:latest` Docker tag — only stable releases do.

## Tagging Rule

The release workflow is tag-driven. Releases are annotated tags:

```bash
git tag -a v1.2.3 -m "Xuva v1.2.3"
git push origin v1.2.3
```

Every release tag must point at a `main` commit that has passed CI.

`tools/validate-release-tag.cjs` enforces the strict semver shape. The guard rejects malformed tags (e.g. `v1.2`, `1.2.3`, `v1.2.3-foo`) at the start of the release workflow so a typo can't publish a broken artifact.

## Promotion History

- **v0.0.x** — pre-public iteration. Used while the Windows MSI install/upgrade/uninstall, Docker image, native pairing, and library scanning were being validated on real hardware.
- **v1.0.0** — public release. Server, Web, Apple TV / iOS, Windows MSI, and Docker images are box-validated; auto-update is end-to-end proven.
