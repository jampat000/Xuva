# Xuva Desktop (Alpha Scaffold)

Windows-first launcher/tray shell for Desktop Alpha. The primary UI opens in the user's default browser; Electron is only used for tray/server supervision and small utility windows.

Current scaffold includes:

- Tray/taskbar launcher.
- Local server process supervision (start, stop, restart).
- First launch defaults to the bundled local server and opens the browser-based web setup flow.
- Browser-only folder browsing through the server filesystem API.
- Tray restart control for the supervised server process.
- Open Xuva action to `http://127.0.0.1:8097` in the default browser.
- Windows update handoff: the server stages a verified installer request and the launcher runs a detached updater to apply it after Xuva exits.

## Dev run

From `apps/desktop`:

```powershell
npm install
npm run dev
```

Notes:

- Default server command is `go run ./cmd/Xuva` with cwd `server/`.
- Packaged builds launch the bundled `resources/runtime/xuva-server.exe`.
- Packaged builds default `XUVA_HTTP_ADDR` to `0.0.0.0:8097`.
- Packaged builds prefer the machine runtime home `C:\ProgramData\Xuva` so one device has one durable server identity.
- If `C:\ProgramData\Xuva` cannot be created or written, packaged builds fall back to `%LOCALAPPDATA%\Xuva` and set `XUVA_RUNTIME_SCOPE=user-fallback`.
- Runtime folders are `data`, `logs`, `transcode`, `downloads`, `metadata`, `cache`, `temp`, and `trailers`.
- Staged update installers and handoff files live under `<runtime-home>\updates`.
- Override the runtime home with `XUVA_RUNTIME_HOME` when testing migrations or custom installs.
- Override using:
  - `XUVA_SERVER_CMD`
  - `XUVA_SERVER_ARGS`
  - `XUVA_SERVER_CWD`

## Windows package build

The repository package script builds the server runtime, bundles FFmpeg/FFprobe,
and then runs Electron Builder:

```powershell
./packaging/windows/build-package.ps1 -Version v0.0.x
```

Outputs:

- `dist/windows/xuva-v0.0.x-win-x64.exe` unsigned per-user installer.
- `dist/windows/xuva-v0.0.x-win-x64.zip` portable desktop package.

## Server discovery â€” when does this app need it?

This launcher **spawns the Xuva server itself** and opens
`http://127.0.0.1:8097` in the default browser. In the default single-machine
install there is no other server to discover; the one running on this box is
the one you want.

When the desktop app gains a multi-machine mode (connect to a Xuva on
the network instead of launching one locally), implement discovery using
the same protocol as the other clients:

```js
// In main process (Node side):
const { Bonjour } = require('bonjour-service');
const bonjour = new Bonjour();
const browser = bonjour.find({ type: 'xuva' }, (service) => {
  // service.name, service.host, service.port, service.txt.serverName, etc.
});
// IPC the list to the renderer for a "connect to remote server" UI.
```

Wire format documented in `apps/android-tv/README.md`. Same service type
(`_xuva._tcp`), same TXT records, same auto-pair flow as Apple.
