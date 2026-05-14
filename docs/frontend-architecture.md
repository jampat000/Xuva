# Frontend Architecture

Xuva's web UI remains a lightweight server-rendered static app: no framework runtime, no build step, and no client bundle pipeline. The app is split into browser-safe modules that also run under Node tests.

## Module Boundaries

- `api-client.js`: all HTTP calls, JSON parsing, request timeouts, retries, request IDs, normalized user-facing errors, and simple API shape validators.
- `error-boundary.js`: consistent full-page and inline retry UX for navigation failures and live refresh failures.
- `playback-presenter.js`: playback readiness, server impact language, source inspector facts, and critical playback labels.
- `domain-boundaries.js`: ownership map for dashboard, playback, metadata, settings, and activity modules.
- `app.js`: app state, screen composition, DOM wiring, and domain render orchestration.

New frontend work should move pure formatting, labeling, and API contract logic into modules first, then keep DOM mutation in `app.js` or a future domain renderer. This keeps the UI testable without forcing a framework rewrite.

## API Contract Pattern

Use the shared client:

```js
const payload = await api("/api/sessions");
```

For critical lists, validate the expected shape before rendering:

```js
const sessions = XuvaApi.validateArrayPayload(payload, "sessions");
```

Errors should surface through `XuvaErrors.renderErrorBoundary` for full views or `XuvaErrors.renderInlineError` for live patches. Avoid raw `error.message` in user-facing UI.

## Critical Tests

Frontend module tests live in `server/internal/webapp/frontend_tests` and use Node's built-in test runner:

```powershell
node --test server/internal/webapp/frontend_tests/*.test.cjs
```

Required coverage for future changes:

- Playback readiness labels and server impact language.
- Source inspector facts and escaping.
- API error normalization, retries, and contract validators.
- Error boundary retry rendering.

## Demo Checklist

- App boots on `/` and the dashboard renders.
- Density and theme controls still update immediately.
- Movies open to source detail and source inspector actions still render.
- Playback readiness wording never says a file cannot play when conversion is only an impact concern.
- Live dashboard/activity updates patch in place rather than full-refreshing on every session tick.
- API failures show a retry path instead of raw or silent errors.

## Rollback

The migration is staged. If a module causes trouble, revert the script include and the small delegate calls in `app.js`; the previous inline functions remain as fallbacks for playback labels while the module boundary settles.
