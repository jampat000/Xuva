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
