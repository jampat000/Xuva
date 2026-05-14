# Xuva Tools

## Agent Harness Check

`agent-check.cjs` verifies that the repository remains legible to future agent runs:

```powershell
node tools/agent-check.cjs
```

It checks the agent map, docs index, execution-plan folders, and protected route policy alignment.

## Media Scanner

`xuva_scan.py` inventories a local or mapped NAS media folder and optionally runs `ffprobe` for stream metadata.

The tool is read-only. It does not upload, copy, rename, delete, or modify media files.

Example inventory-only scan:

```powershell
python tools/xuva_scan.py X:\ --no-probe --limit 100
```

Example deep probe scan after FFmpeg is installed:

```powershell
winget install --id Gyan.FFmpeg -e --source winget --accept-source-agreements --accept-package-agreements
```

```powershell
python tools/xuva_scan.py X:\ --ffprobe "C:\ffmpeg\bin\ffprobe.exe" --workers 2
```

On Windows, the scanner also checks winget's FFmpeg alias location automatically:

```text
%LOCALAPPDATA%\Microsoft\WinGet\Links\ffprobe.exe
```

Outputs are local and ignored by git:

```text
data/probe-results.jsonl
data/probe-results-summary.json
```

Generate a local compatibility dashboard from a scan:

```powershell
python tools/xuva_dashboard.py data/probe-results.jsonl --out data/compatibility-dashboard.html
```

Use low worker counts for NAS paths. `--workers 2` is the default because network storage can become slower when too many files are probed at once.

The summary counts the primary playable video stream per file. Secondary thumbnail/preview video streams remain in JSONL records but do not affect headline video codec totals.

Useful options:

- `--limit 100` scans a small sample.
- `--no-probe` skips FFmpeg and only inventories file paths, sizes, extensions, and modified times.
- `--hash-paths` hashes paths in output before sharing results.
- `--no-resume` ignores prior JSONL records and scans again.

Scanner records include the top-level library section, media kind, embedded subtitles, sidecar subtitles, primary video stream, audio streams, and prototype playback decisions.
