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

- The production web UI is served from `server/internal/webapp/static`.
- Shared browser modules live in `server/internal/webapp/static/modules`.
- Critical frontend tests live in `server/internal/webapp/frontend_tests` and run with `node --test`.
- Architecture and contribution conventions are documented in `docs/frontend-architecture.md`.
