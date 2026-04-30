# Agent Harness

Vyrden should be optimized for agent-assisted development. The goal is not process for its own sake; the goal is to make product intent, architecture, UI behavior, tests, logs, and quality constraints directly legible from the repository.

This alignment is based on the harness-engineering model described by OpenAI: humans steer, agents execute, repository knowledge is the source of truth, and mechanical checks enforce architecture and taste.

## Principles

1. **Repository knowledge wins**
   - Product decisions, architecture, acceptance criteria, and rollback notes belong in docs or tests.
   - Chat-only decisions must be promoted into repository artifacts when they affect future work.

2. **Small map, deep references**
   - `AGENTS.md` is a table of contents.
   - Detailed guidance belongs in focused docs and execution plans.

3. **Agent legibility is product leverage**
   - The app must be runnable locally.
   - UI states, logs, metrics, and API contracts should be inspectable by tools.
   - Validation steps should be scripted where possible.

4. **Mechanical taste**
   - If a rule matters repeatedly, encode it in `tools/agent-check.cjs`, tests, or CI.
   - Prefer structured API responses and explicit route policy entries over implicit behavior.

5. **Continuous garbage collection**
   - Small cleanup PRs are preferred over large rescue refactors.
   - Quality gaps go into [quality-score.md](quality-score.md) or [tech-debt-tracker.md](tech-debt-tracker.md).

## Work Loop

1. Inspect current git status and related docs.
2. Create or update an execution plan for substantial work.
3. Implement the smallest product-valid slice.
4. Run local tests and harness checks.
5. Update docs if behavior, contracts, or product direction changed.
6. Capture residual risks and follow-up debt.

## App Legibility Targets

- Dev server starts consistently on `127.0.0.1:8097`.
- Static web UI is testable with Node frontend tests.
- API routes have route policy documentation when protected.
- Native client contracts are documented before app implementation depends on them.
- Logs and operational events include correlation IDs for workflows that cross services.

## Current Harness Checks

Run:

```powershell
node tools/agent-check.cjs
```

The check currently verifies:

- `AGENTS.md` exists and links to core docs.
- docs index and execution-plan directories exist.
- route policy and API route registrations stay aligned for protected routes.
- docs that should exist for active product tracks are present.

Expand this script whenever repeated review feedback becomes mechanical.
