# xuva web frontend

This app is the SvelteKit + TypeScript product frontend.

- Source path: `apps/web/svelte`
- Root mount: `/`
- Go-served build output: `server/internal/webapp/static-next`

## Component Foundation

Reusable primitives now live in:

- `src/lib/components/shell/*`
- `src/lib/components/ui/*`
- `src/lib/components/media/*`
- `src/lib/components/operator/*`

These primitives now power the root Svelte routes.

## Install

```powershell
npm --prefix apps/web/svelte install
```

## Run Dev Server

```powershell
npm --prefix apps/web/svelte run dev
```

Open [http://localhost:5174/](http://localhost:5174/) when using the repo dev scripts. Set `XUVA_WEB_DEV_PORT` to override the Vite port and `XUVA_API_ORIGIN` to override the Go API proxy target.

## Type Check

```powershell
npm --prefix apps/web/svelte run check
```

## Build / Publish To Go Static

```powershell
npm --prefix apps/web/svelte run publish:go-static
```

This command:

1. Cleans `server/internal/webapp/static-next` (keeps `.gitignore` and `README.md`).
2. Builds the SvelteKit app with static adapter output into `server/internal/webapp/static-next`.

You can also run build without the clean step:

```powershell
npm --prefix apps/web/svelte run build
```

## Route Smoke

```powershell
npm --prefix apps/web/svelte run smoke:root -- http://127.0.0.1:18100
```

## Test Script

```powershell
npm --prefix apps/web/svelte run test
```
