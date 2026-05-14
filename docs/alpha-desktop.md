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

## Alpha Scope

- Windows tray/taskbar app starts and supervises the Go server.
- Open Xuva launches `http://127.0.0.1:8097`.
- Restart Xuva restarts the supervised server process.
- Native folder picker fills library and runtime folder fields.
- Web fallback remains available for dev/headless/local browser usage.
- No vendor relay infrastructure is introduced.

## Validation

- Add a movie library from a local drive.
- Add a TV library from a NAS or mapped drive.
- Move transcode/cache folders and verify settings remain visible before restart.
- Restart from the desktop shell and verify active runtime folders use the saved locations.
- Confirm the same screens still work in dev without the desktop bridge.
