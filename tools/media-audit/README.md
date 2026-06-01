# media-audit

Overnight ffprobe-driven health check for a Xuva media library.

Walks one or more directories, runs `ffprobe` on every media file
(`.mkv`, `.mp4`, `.m4v`, `.avi`, `.mov`, `.webm`, `.ts`, `.mpg`,
`.m2ts`, …), and writes two reports:

- `audit-report.json` — full structured `FileAudit` per file (codecs,
  resolution, FPS, HDR format, bitrate, audio tracks, subtitle tracks,
  detected issues).
- `audit-summary.txt` — human-readable counts (codecs, containers,
  HDR formats, issues) and the flagged-file list.

It is **resumable**: if `audit-report.json` already exists, files whose
size + mtime are unchanged are reused without reprobing. Useful for
splitting a large run across nights or for catching new files only.

It is **graceful on Ctrl-C**: SIGINT / SIGTERM stop the walk and write
the partial report so an overnight run isn't lost.

It does not need Xuva to be installed or scanned — it's a file-walk
tool, not a DB query (the companion `server/cmd/xuva-audit` is the
DB-driven sibling for after-scan audits).

## Quick start

```bash
# Run against one or more library roots; reports land in cwd.
tools/media-audit/run.sh /mnt/media/Movies /mnt/media/TV

# Run overnight (resumable on next start):
nohup tools/media-audit/run.sh /mnt/media/Movies > audit.log 2>&1 &

# Cancel cleanly — partial report is written before exit:
kill -SIGTERM <pid>
```

Environment variables:

- `XUVA_AUDIT_OUT_DIR` — directory to write reports (default: cwd).
- `XUVA_FFPROBE` — path to ffprobe binary (default: `ffprobe` from PATH).

## What gets flagged

Per file, the audit raises a flag for any of:

- ffprobe failed (file may be corrupt / unreadable)
- no video stream (likely audio-only) or no audio streams
- variable frame rate (`r_frame_rate` ≠ `avg_frame_rate` ±2 %)
- HDR (HDR10, Dolby Vision, HLG) — verify tone-mapping on SDR clients
- codec that web clients can't direct-play (HEVC / VP9 / AV1)
- bitrate above 50 Mbps (may stress slow networks)

The summary text file groups these by count so you can scan for
hotspots. The JSON report has the full per-file data for scripting
follow-ups.

## Exit codes

- `0` — all files probed cleanly, no flags
- `1` — one or more files flagged or failed ffprobe
- `2` — fatal startup error (missing arguments, ffprobe not on PATH,
  output directory not writable)

## Compared to `server/cmd/xuva-audit`

| Tool | Source of file list | Output | When to use |
|---|---|---|---|
| `tools/media-audit/` | walks a directory | `audit-report.json` + `audit-summary.txt` | one-off health check on a foreign drive or before a fresh scan |
| `server/cmd/xuva-audit` | reads Xuva DB | Markdown report | after a Xuva scan, to flag what *Xuva* thinks is wrong |
