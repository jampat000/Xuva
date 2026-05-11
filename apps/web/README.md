# Lorivo Web

The web app covers setup, admin, library management, diagnostics, and browser playback.

Primary screens:

- First-run setup.
- Library management.
- Metadata matching.
- Users.
- Devices.
- Playback sessions.
- Server health.
- Remote access diagnostics.

Frontend runtime notes:

- Source for the production web UI lives in `apps/web/svelte`.
- The Go server serves the published Svelte build from `server/internal/webapp/static-next`.
- Critical frontend tests live in `server/internal/webapp/frontend_tests` and run with `node --test`.
- Architecture and contribution conventions are documented in `docs/frontend-architecture.md`.
