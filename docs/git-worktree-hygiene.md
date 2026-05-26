# Git Worktree Hygiene

Xuva is a mixed server/web/native monorepo. Git operations must stay predictable on Windows, Linux, and NAS-backed workspaces.

## Rules

- Do not install local checkout hooks that mutate tracked files after `git checkout`, `git switch`, `git merge`, or `git pull`.
- Do not use local hooks to force-reset Apple, tvOS, Android, generated web assets, or server files.
- If a path needs generated output refreshed, use an explicit script and document it.
- Keep Git maintenance disabled locally on NAS-backed Windows workspaces when pack/index permission errors appear:

```powershell
git config --local maintenance.auto false
git config --local gc.auto 0
```

## Known Windows/NAS Failure Mode

A local `.git/hooks/post-checkout` hook previously ran:

```powershell
git checkout HEAD -- apps/apple-core/Sources/XuvaClientCore/ apps/tvos/Sources/XuvaTVApp/
```

On the NAS-backed Windows checkout this spawned many stuck Git workers and blocked normal branch switching. The correct fix is to disable that local hook, not to change Apple client source.

## Recommended Recovery

If Git commands hang after switching branches:

```powershell
Get-CimInstance Win32_Process |
  Where-Object { $_.Name -match '^git(\.exe)?$' -and $_.CommandLine -like '*checkout HEAD --*apps/apple-core*apps/tvos*' } |
  ForEach-Object { Stop-Process -Id $_.ProcessId -Force }

Remove-Item .git/index.lock -Force -ErrorAction SilentlyContinue
git status --short --branch
```

Only remove `.git/index.lock` after confirming no legitimate Git command is still running.
