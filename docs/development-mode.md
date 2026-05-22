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

In `-WebDev` mode, Xuva proxies web routes to Vite (`http://127.0.0.1:5174` by default, or `XUVA_WEB_DEV_PORT`) while API/playback routes still come from Go on `127.0.0.1:8097`. Vite proxies `/api` to `XUVA_API_ORIGIN` when set, otherwise `http://127.0.0.1:8097`.

## Canonical Web Address

Browsers isolate cookies and local storage by origin, so `localhost`, `127.0.0.1`, the LAN IP, and the machine hostname cannot safely share one browser session. Xuva therefore uses one canonical web origin for browser UI routes:

- Set `XUVA_CANONICAL_WEB_ORIGIN` or Settings -> General -> Canonical web address for a fixed address such as `http://media-server.local:8097`.
- If it is blank and the server listens on the LAN, raw IP and loopback browser visits redirect to the operating-system hostname on the same port.
- API, media, health, pairing, artwork, and stream routes do not redirect. Native clients can keep using the LAN URL they discovered.

This is separate from the Xuva Server Name. The Xuva Server Name is the friendly display/discovery instance name; it is not a DNS name. The canonical web origin is the real browser authority, for example an operating-system hostname, LAN DNS name, mDNS name, reverse proxy hostname, or container host address.

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
