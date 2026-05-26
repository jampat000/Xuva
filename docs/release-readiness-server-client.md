# Server/Client Release Readiness

This checklist gates releases for Xuva as a server/client system.

## 1) Runtime Reliability

- Server boots cleanly with persisted settings and runtime directories.
- Server restart recovers without manual database repair.
- Queue limits are enforced for scan/probe/transcode/download classes.
- Active playback stays responsive during background load.

## 2) Playback Correctness

- Deterministic route decisions for fixed source + client profile inputs.
- Direct/remux/transcode paths validated with representative fixtures.
- Subtitle and audio fallback paths validated for known edge cases.
- Resume and watched state durability verified across restart.

## 3) Security And Access

- Admin-only endpoints reject standard users.
- CSRF-protected mutation routes reject missing/invalid tokens.
- Stream-token protections validated for expiry, forgery, and cross-session misuse.
- Desktop bridge actions degrade safely when bridge is unavailable.

## 4) Client Contract Stability

- API contract tests pass for settings, playback, migration, and auth flows.
- No unintended breaking contract changes in route-policy-sensitive endpoints.
- SSE event payloads remain parseable by current clients.

## 5) Desktop Shell Operations

- Tray/taskbar shell can start, stop, and restart supervised server process.
- Native folder picker covers local, mapped, and UNC paths for signed-in user scope.
- Restart actions are explicit and visible to operators.
- Desktop logs include restart and bridge action audit events.
- Packaged Windows runtime includes Xuva, embedded web assets, desktop shell runtime, FFmpeg, FFprobe, CA certificates, and default runtime directory creation.

## 6) Verification Suite

Run before tagging a release:

```powershell
node tools/agent-check.cjs
node --test server/internal/webapp/frontend_tests/*.test.cjs
npm --prefix apps/web/svelte run check
go test ./...
git diff --check
```

## 7) Installation And Upgrade

- Clean install verified on target OS build.
- Upgrade from previous build verified with settings and data retained.
- Uninstall/reinstall verified without orphaning critical runtime state.
- Rollback procedure documented and rehearsed.
- Unsigned Windows artifact trust is covered by GitHub Release provenance plus published SHA256 checksums.
- No user-facing package assumes Go, Node.js, npm, SQLite tooling, FFmpeg, or FFprobe are already installed.
- Docker images include runtime prerequisites and expose a healthcheck suitable for orchestrators.

Rehearsal automation command:

```powershell
./tools/rehearse-install-upgrade-rollback.ps1
```
