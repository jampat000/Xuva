## Workflow rules for Claude sessions

These rules apply to every Claude Code session in this repo.

### Context window management — HARD RULE
- **Monitor context usage throughout every session.** When the conversation reaches approximately 75% of the context window, stop whatever you are doing and issue a clear warning before continuing:

  > ⚠️ **Context at ~75% — recommend starting a fresh session.**
  > Summarise any in-flight work (open PRs, pending edits, next steps) so nothing is lost, then pause and let the user decide whether to continue or open a new session.

- Do not wait until the context is exhausted. Degraded responses near the limit cost more to fix than a clean handoff.
- At the start of every new session, check the memory files and recent git log so you can resume seamlessly without re-reading the entire prior conversation.

### Branching & commits
- **All work goes through a PR into `main`** under normal circumstances. Direct pushes to `main` are permitted only when the repo owner explicitly instructs it in the chat (admin override). In that case, use `gh pr merge --merge` on the relevant PRs rather than `git push` — this keeps the commit history intact and triggers any post-merge hooks.
- Create a topic branch named for the work (e.g. `fix/<area>-<short>`, `chore/<area>-<short>`, `feat/<area>-<short>`).
- Open a PR with `gh pr create` (or the MCP equivalent). Wait for CI to be green before merging. Merge with `--merge` (preserve commit history), not squash.
- **Commit edits to a branch immediately after applying them.** Don't accumulate `Edit`/`Write` calls across multiple turns hoping to commit later — a silent file revert (by the user, a linter, or restored prior state) between your edit and your commit will lose the work without warning, and you'll only notice when someone says "I thought we shipped that." If a change is part of a larger PR and you can't commit the final state yet, push a WIP commit anyway so the work survives.

### Worktrees
- **Never commit from a Claude worktree under `.claude/worktrees/...`.** That isolates commits onto a worktree branch where they invisibly diverge from `main`.
- If the session CWD is `.claude/worktrees/...`, the agent must `cd` to `C:\Projects\Xuva` (or use `git -C C:\Projects\Xuva ...`) for every git operation.
- When in doubt, run `git rev-parse --show-toplevel` to confirm where commits would land.
- **Always use `git -C C:\Projects\Xuva ...` explicitly.** PowerShell's working directory can get stuck pointing at a removed worktree path (`Shell cwd was reset to ...`), which causes `git commit` to fail with `fatal: '/' is outside repository` while `git push` silently succeeds against the wrong ref. Belt-and-braces: pass `-C` on every git command, and write commit messages to a file (`git commit -F path/to/msg.txt`) rather than relying on PowerShell heredocs which can also misfire under the same cwd-reset.

### CI gate
- The `Security` workflow (`.github/workflows/security.yml`) runs `tools/agent-check.cjs` as a governance gate. It enforces three-way consistency between:
  1. `handleProtected` / `handleProtectedCSRF` registrations in `server/internal/api/router.go`
  2. Policy entries in `server/internal/api/authz.go`
  3. Route rows in `docs/route-policy.md`
- When adding a new protected route, update all three in the same PR or CI will block.

### Tests visibility
- The agent-check gate runs *before* `go test`. If agent-check is broken, test failures are hidden behind it. After fixing agent-check, expect previously-masked test failures to surface — fix them in the same PR or a follow-up.

### Installer changes — HARD RULE
- **PR-level CI does NOT run the Windows installer build.** The `Windows package script checks` job only validates PowerShell syntax. The actual NSIS compile + signing pipeline runs only in the release workflow on `v*.*.*` tag pushes. Three back-to-back v0.0.34 release runs failed because PR CI was green but the tag-push installer build wasn't.
- **Before pushing any change to `apps/desktop/installer.nsh`, `apps/desktop/package.json`'s `nsis` block, or `packaging/windows/build-package.ps1`, run `tools/check-installer-build.ps1` locally.** It exercises the same `npx electron-builder --win nsis --x64` path CI uses. Takes ~3 minutes on a warm cache. If it doesn't produce a `dist/*.exe`, don't push.
- This rule overrides the "wait for CI green" check on installer-touching PRs. Local installer build must pass first; CI is the second confirmation.
