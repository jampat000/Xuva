# Alpha Desktop Packaging

Xuva's first alpha should ship as a user-launched desktop app with a tray/taskbar presence, not as a background system service. The server still runs locally, but the desktop shell owns install-time ergonomics: open the web UI, restart the server, expose native folder pickers, and show basic health.

## Folder Selection

The web UI supports a desktop bridge at `window.xuvaDesktop.pickFolder(request)`.

Request:

```json
{
  "title": "Library folder",
  "currentPath": "D:\\Media\\Movies",
  "purpose": "library"
}
```

Response:

```json
{ "path": "D:\\Media\\Movies" }
```

The bridge should open the operating system's native folder picker. On Windows this gives users local drives, mapped drives, Quick Access, and reachable NAS/UNC paths under the signed-in user's permissions. If the bridge is absent, returns no path, or the user is running headless/dev, the web UI falls back to `GET /api/settings/folders/browse`.

The desktop shell also exposes `window.xuvaDesktop.restartServer()` for explicit "Restart Xuva" actions.

## Alpha Scope

- Windows tray/taskbar app starts and supervises the Go server.
- First launch defaults to local-owner mode and opens the web first-start/setup flow; remote server selection remains available from the tray.
- Open Xuva launches `http://127.0.0.1:8097`.
- Restart Xuva restarts the supervised server process.
- Native folder picker fills library and runtime folder fields.
- Web fallback remains available for dev/headless/local browser usage.
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
- Revisit trigger: if package size, startup time, or maintenance overhead becomes a release blocker, re-evaluate shell after alpha feedback.

## Validation

- Add a movie library from a local drive.
- Add a TV library from a NAS or mapped drive.
- Move transcode/cache folders and verify settings remain visible before restart.
- Restart from the desktop shell and verify active runtime folders use the saved locations.
- Confirm the same screens still work in dev without the desktop bridge.

## Restart Runbook

- `Restart Xuva` is an explicit operator action from the desktop shell bridge.
- Expected behavior: desktop shell restarts the supervised Go child process and the web UI reconnects to `127.0.0.1:8097`.
- If restart fails, keep current process state and surface a user-visible error instead of silent exit.
- If bridge is unavailable, web UI keeps server/web fallback behavior and does not pretend restart was executed.
