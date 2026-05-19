## Workflow rules for Claude sessions

These rules apply to every Claude Code session in this repo.

### Branching & commits
- **All work goes through a PR into `main`.** Never `git push` directly to `main`, even when admin override allows it. The repo has branch protection; respect it.
- Create a topic branch named for the work (e.g. `fix/<area>-<short>`, `chore/<area>-<short>`, `feat/<area>-<short>`).
- Open a PR with `gh pr create` (or the MCP equivalent). Wait for CI to be green before merging. Merge with `--merge` (preserve commit history), not squash.

### Worktrees
- **Never commit from a Claude worktree under `.claude/worktrees/...`.** That isolates commits onto a worktree branch where they invisibly diverge from `main`.
- If the session CWD is `.claude/worktrees/...`, the agent must `cd` to `C:\Projects\Xuva` (or use `git -C C:\Projects\Xuva ...`) for every git operation.
- When in doubt, run `git rev-parse --show-toplevel` to confirm where commits would land.

### CI gate
- The `Security` workflow (`.github/workflows/security.yml`) runs `tools/agent-check.cjs` as a governance gate. It enforces three-way consistency between:
  1. `handleProtected` / `handleProtectedCSRF` registrations in `server/internal/api/router.go`
  2. Policy entries in `server/internal/api/authz.go`
  3. Route rows in `docs/route-policy.md`
- When adding a new protected route, update all three in the same PR or CI will block.

### Tests visibility
- The agent-check gate runs *before* `go test`. If agent-check is broken, test failures are hidden behind it. After fixing agent-check, expect previously-masked test failures to surface — fix them in the same PR or a follow-up.
