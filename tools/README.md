# Vyrden Tools

## Media Scanner

`vyrden_scan.py` inventories a local or mapped NAS media folder and optionally runs `ffprobe` for stream metadata.

The tool is read-only. It does not upload, copy, rename, delete, or modify media files.

Example inventory-only scan:

```powershell
python tools/vyrden_scan.py X:\ --no-probe --limit 100
```

Example deep probe scan after FFmpeg is installed:

```powershell
python tools/vyrden_scan.py X:\ --ffprobe "C:\ffmpeg\bin\ffprobe.exe" --workers 2
```

Outputs are local and ignored by git:

```text
data/probe-results.jsonl
data/probe-results-summary.json
```

Use low worker counts for NAS paths. `--workers 2` is the default because network storage can become slower when too many files are probed at once.

Useful options:

- `--limit 100` scans a small sample.
- `--no-probe` skips FFmpeg and only inventories file paths, sizes, extensions, and modified times.
- `--hash-paths` hashes paths in output before sharing results.
- `--no-resume` ignores prior JSONL records and scans again.
