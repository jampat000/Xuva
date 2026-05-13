# Lorivo Development Mode

Lorivo Phase 1 development targets a desktop owner workflow on local loopback.

## Canonical local mode

Use desktop owner development mode when working on server-owner settings and server/web UX.

- `LORIVO_DEV_AUTH_BYPASS=true`
- `LORIVO_HTTP_ADDR=127.0.0.1:8097`
- auth remains enabled; bypass only unlocks a local development owner session
- bypass is ignored on non-loopback binds

Run from repo root:

```powershell
./tools/run-desktop-owner.ps1
```

## Why stale UI happens

Lorivo web assets are embedded into Go from:

- `server/internal/webapp/static-next`

If you change Svelte files but do not republish static output and restart the Go server, the running app may still serve older assets.

The canonical script rebuilds and republishes frontend assets before starting the server to avoid stale runtime mismatches.

## Production/default behavior

- `LORIVO_DEV_AUTH_BYPASS` defaults to `false`
- production/non-loopback deployments should not use bypass
- owner sign-in/bootstrap behavior remains active in normal mode

## Scope guard for current phase

- desktop owner workflow is the acceptance target
- tablet/mobile web is smoke-only
- user-account/auth expansion is Phase 2 work
- tvOS/Android client work is Phase 4 work
