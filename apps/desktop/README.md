# Xuva Desktop (Alpha Scaffold)

Windows-first Electron shell for Desktop Alpha.

Current scaffold includes:

- Tray/taskbar shell window.
- Local server process supervision (start, stop, restart).
- `window.xuvaDesktop.pickFolder(request)` bridge.
- `window.xuvaDesktop.restartServer()` bridge.
- Open Xuva action to `http://127.0.0.1:8097`.

## Dev run

From `apps/desktop`:

```powershell
npm install
npm run dev
```

Notes:

- Default server command is `go run ./cmd/Xuva` with cwd `server/`.
- Override using:
  - `XUVA_SERVER_CMD`
  - `XUVA_SERVER_ARGS`
  - `XUVA_SERVER_CWD`

## Server discovery — when does this app need it?

This Electron shell **spawns the Xuva server itself** and points to
`http://127.0.0.1:8097`. In the default single-machine install there is
no other server to discover — the one running on this box is the one
you want.

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
