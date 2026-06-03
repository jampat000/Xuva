# Contributing to Xuva

Thanks for considering a contribution. This document covers the practical things — how to get the project building, what CI expects, and how to land a PR.

## Repository layout

```
apps/
  android-tv/   Native Android TV client (Kotlin + Compose-for-TV)
  apple-core/   Shared SwiftUI code for the Apple clients
  desktop/      Electron tray shell (bundled into the Windows MSI)
  ios/          iPhone / iPad client
  tvos/         Apple TV client
  web/svelte/   Web UI + browser player (SvelteKit)
docker/         Container entrypoint script
docs/           Architecture, ops, decisions, playback engine
packaging/      Windows MSI build (WiX v5)
server/         Go server, embeds the web UI via go:embed
tools/          Repo-wide automation (CI governance, audit, validators)
```

## Building

You'll need the toolchain for whichever surface you're changing — everything is path-filtered in CI.

### Server (Go + embedded web UI)

```bash
cd apps/web/svelte && npm ci && npm run publish:go-static
cd ../../../server && go build -o xuva ./cmd/Xuva
./xuva
```

The web app uses `npm run publish:go-static` (not `npm run build`) because it preserves the `.gitignore` and `README.md` in the embed target and writes `build-info.json`. Plain `npm run build` rimrafs the target dir and produces a malformed snapshot.

### Apple (iOS / tvOS)

```bash
xcodebuild build \
  -project apps/ios/XuvaIOS.xcodeproj \
  -scheme "Xuva iOS" \
  -destination 'generic/platform=iOS Simulator' \
  CODE_SIGNING_ALLOWED=NO
```

Swap `apps/tvos/XuvaTV.xcodeproj` / `Xuva TV` / `tvOS Simulator` for the Apple TV build. CI uses Xcode 16.4 — a recent local Xcode is more lenient about SDK availability; **trust CI for tvOS-specific availability errors** (e.g. `controlSize` is iOS-only and must be `#if !os(tvOS)`-gated).

### Android TV

No local Android toolchain is required — CI builds it on every PR (`.github/workflows/android.yml`). Locally:

```bash
cd apps/android-tv
gradle assembleDebug
```

Requires JDK 17 + Gradle 8.11 + Android SDK with `compileSdk 35`.

### Windows MSI

The MSI is built in CI on every PR (`.github/workflows/security.yml`, "Windows MSI build" job). Locally it needs Windows + `dotnet tool install --global wix` + the build script at `packaging/windows/build-msi.ps1`.

### Docker

```bash
docker compose up --build
```

The compose file is hardened for self-hosting (PUID/PGID, cap drops, `no-new-privileges`). For development you can drop the cap restrictions in a `docker-compose.override.yml`.

## Tests

```bash
cd server && go test ./...
cd apps/web/svelte && npm run check && npm run test:frontend
```

There is a known pre-existing `TestCatalogSummaryUpdatesAfterScans` macOS-only failure that's green in Linux CI. If only that test fails locally on macOS, ignore it; if anything else fails, fix it.

## CI gate (agent-check)

`tools/agent-check.cjs` is a governance gate that runs before `go test`. It enforces three-way consistency between:

1. `handleProtected` / `handleProtectedCSRF` registrations in `server/internal/api/router.go`
2. Policy entries in `server/internal/api/authz.go`
3. Route rows in `docs/route-policy.md`

When you add a new protected route, update all three in the same PR or CI will block. Run `node tools/agent-check.cjs` locally before pushing.

## Pull request workflow

1. **Branch off `main`**, name it for the work (`fix/<area>-<short>`, `feat/<area>-<short>`, `chore/<area>-<short>`).
2. **Open a PR with `gh pr create`.** Wait for the full CI suite (Go tests, frontend, Docker smoke, Windows MSI build, Android build when touched, Apple build when touched).
3. **Merge with `--squash --delete-branch`.** `main` requires linear history (no merge commits) and one approval.

## Code style

- **Go**: `gofmt` / `go vet`. Errors propagate as `error` values, not panic. Prefer composition over interface zoos.
- **Svelte / TypeScript**: prettier defaults, `svelte-check` zero warnings.
- **Kotlin / Android**: official Kotlin style (`kotlin.code.style=official`), no `!!` on nullable values without a clear contract.
- **Swift**: SwiftUI idiomatic, `@MainActor` annotations where required.
- **Commits**: imperative present tense ("fix scanner deadlock"), `area: subject` prefix when it clarifies (`fix(catalog):`, `feat(android-tv):`).

## Security disclosures

See [SECURITY.md](SECURITY.md) — please use GitHub Security Advisories for private disclosure rather than opening public issues.

## License

By contributing you agree your code is licensed under the [MIT License](LICENSE).
