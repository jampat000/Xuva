# Lorivo Agent Map

This file is the short entry point for agent work. Do not turn it into a full manual. The repository-local docs below are the system of record.

## Product Direction

- [Product principles](docs/product-principles.md)
- [MVP scope](docs/mvp-scope.md)
- [Roadmap](docs/roadmap.md)
- [Apple TV alpha](docs/apple-tv-alpha.md)
- [Alpha desktop packaging](docs/alpha-desktop.md)

## Architecture

- [Architecture](docs/architecture.md)
- [Route policy](docs/route-policy.md)
- [Frontend architecture](docs/frontend-architecture.md)
- [Playback engine](docs/playback-engine.md)
- [Playback decision v2](docs/playback-decision-v2.md)
- [Adaptive streaming](docs/adaptive-streaming.md)

## Design

- [Design index](docs/design/README.md)
- [UI direction](docs/design/ui-direction.md)
- [TV experience](docs/design/tv-experience.md)
- [Design tokens](docs/design/design-tokens.md)
- [Apple TV design](apps/apple-tv/Design.md)

## Execution

- [Docs index](docs/index.md)
- [Active plans](docs/plans/active/)
- [Completed plans](docs/plans/completed/)
- [Technical debt tracker](docs/tech-debt-tracker.md)
- [Quality scorecard](docs/quality-score.md)
- [Agent harness](docs/agent-harness.md)
- [Handover journal](docs/handover-journal.md)

## Required Local Checks

Before merging normal server/web changes:

```powershell
node tools/agent-check.cjs
node --test server/internal/webapp/frontend_tests/*.test.cjs
go test ./... 
git diff --check
```

For security-sensitive or release work also run:

```powershell
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

## Operating Rules

- Keep knowledge in repo docs, not chat memory.
- Prefer small execution plans for multi-step work.
- Add tests or mechanical checks when a rule should persist.
- Keep APIs and UI states legible to future agents.
- Preserve local-first constraints: no required cloud account, no vendor relay, user-owned remote access.
