# Alpha Desktop Packaging

Xuva's first alpha should ship as a user-launched server/web launcher with a tray/taskbar presence, not as a background system service and not as an embedded-app browser. The server still runs locally, but the launcher owns install-time ergonomics: start the server, open the web UI in the user's default browser, restart the server, and show basic health.

## Folder Selection

The primary web UI runs in the user's browser and uses `GET /api/settings/folders/browse` to browse folders from the server process. On Windows this process runs in the signed-in user's session, so it can see the same local drives, mapped drives, removable disks, and reachable NAS/UNC paths that the operator can access.

Native picker bridges are optional utility-surface work only. They must not be required for normal browser-based setup, because the browser session is the product UI.

## Alpha Scope

- Windows tray/taskbar launcher starts and supervises the Go server.
- First launch defaults to local-owner mode and opens the web first-start/setup flow; remote server selection remains available from the tray.
- Open Xuva launches `http://127.0.0.1:8097` in the default browser.
- Restart Xuva restarts the supervised server process.
- Server-side folder browsing fills library and runtime folder fields.
- Browser-first behavior remains valid for dev, headless, local browser, and LAN browser usage.
- Package bundles every runtime dependency needed by normal users: `Xuva.exe` desktop shell, `xuva-server.exe`, embedded web UI, FFmpeg, FFprobe, CA certificates, and default runtime directory creation.
- No vendor relay infrastructure is introduced.
- No Apple, Android, iOS, tvOS, or native mobile/TV client builds are included in the Windows desktop package.

## Windows Package Policy

- Release packages are unsigned unless/until signing is affordable.
- The user-facing Windows entry point is `Xuva.exe`. PowerShell scripts are not the normal launch path.
- Publish SHA256 checksums and GitHub Release provenance for every installer/package artifact.
- Do not require users to install Go, Node.js, npm, FFmpeg, FFprobe, or a database engine.
- Prefer per-user install and user-session execution so mapped drives, removable disks, SMB shares, and UNC paths work under the same permissions the operator sees in File Explorer.
- Keep writable runtime state outside the application install directory under the machine runtime home `C:\ProgramData\Xuva\` by default: `data`, `logs`, `transcode`, `downloads`, `metadata`, `cache`, `temp`, and `trailers`.
- If the machine runtime home is not writable, fall back to `%LOCALAPPDATA%\Xuva\` and expose that as single-user fallback mode, not the preferred server identity model.
- If Windows SmartScreen warns because the installer is unsigned, document that clearly rather than pretending the warning will not happen.

## Shell Choice (Alpha)

- Shell: Electron (Windows-first alpha).
- Why now: fastest path to tray/taskbar UX, native folder picker bridge, and local Go process supervision with minimal extra platform plumbing.
- Revisit trigger: if package size, startup time, or maintenance overhead becomes a release blocker, replace Electron with a smaller native tray launcher after alpha feedback.

## Validation

- Add a movie library from a local drive.
- Add a TV library from a NAS or mapped drive.
- Move transcode/cache folders and verify settings remain visible before restart.
- Restart from the tray launcher and verify active runtime folders use the saved locations.
- Confirm the same screens work from the default browser and from another LAN browser.

## Restart Runbook

- `Restart Xuva` is an explicit operator action from the tray launcher.
- Expected behavior: desktop shell restarts the supervised Go child process and the web UI reconnects to `127.0.0.1:8097`.
- If restart fails, keep current process state and surface a user-visible error instead of silent exit.
- If the launcher is unavailable, the web UI keeps server/web fallback behavior and does not pretend restart was executed.

## Upgrade Surface

- The web settings UI exposes an Updates section that checks the latest GitHub Release and offers Windows/Docker downloads.
- Automatic self-apply is deliberately deferred until a separate updater supervisor exists. A running server process should not replace its own executable or installer payload mid-request.
- The durable updater design is: web UI requests update, server verifies metadata and checksum, launcher/updater stops Xuva, applies package, restarts Xuva, then the web UI verifies `/api/system/version`.
