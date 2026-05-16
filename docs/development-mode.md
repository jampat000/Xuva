# Xuva Development Mode

Xuva Phase 1 development targets a desktop owner workflow on local loopback.

## Canonical local mode

Use desktop owner development mode when working on server-owner settings and server/web UX.

- `XUVA_HTTP_ADDR=127.0.0.1:8097`
- auth remains enabled
- first run goes through `/signin` bootstrap, where the first user becomes admin

Run from repo root:

```powershell
./tools/run-desktop-owner.ps1
```

Live web dev mode (no republish/restart for Svelte edits):

```powershell
./tools/run-desktop-owner.ps1 -WebDev
```

In `-WebDev` mode, Xuva proxies web routes to Vite (`http://127.0.0.1:5173`) while API/playback routes still come from Go on `127.0.0.1:8097`.

## Why stale UI happens

Xuva web assets are embedded into Go from:

- `server/internal/webapp/static-next`

If you run without `-WebDev`, Xuva serves embedded static assets. Svelte source edits require republish/restart in that mode.

The canonical script rebuilds and republishes frontend assets before starting the server to avoid stale runtime mismatches.

## Production/default behavior

- owner sign-in/bootstrap behavior remains active in normal mode

## Scope guard for current phase

- desktop owner workflow is the acceptance target
- tablet/mobile web is smoke-only
- user-account/auth expansion is Phase 2 work
- tvOS/Android client work is Phase 4 work
