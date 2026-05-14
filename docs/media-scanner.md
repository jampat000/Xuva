# Media Scanner Prototype

The first practical Xuva engineering step is a local media scanner/prober that can run against a real library without moving or modifying files.

## Goal

Use real local media as a compatibility corpus:

- Containers.
- Codecs.
- HDR formats.
- Audio formats.
- Subtitle formats.
- Bitrates.
- File sizes.
- Client compatibility decisions.

## Rules

- Read-only access to media files.
- No upload.
- No copying.
- No metadata writes to the media folders.
- Output goes to local ignored `data/` files.
- Scans must be resumable.
- NAS scans should default to low concurrency.

## Current Tool

```text
tools/xuva_scan.py
```

Inventory mode works without FFmpeg:

```powershell
python tools/xuva_scan.py X:\ --no-probe --limit 100
```

Deep probe mode requires `ffprobe`:

```powershell
winget install --id Gyan.FFmpeg -e --source winget --accept-source-agreements --accept-package-agreements
```

```powershell
python tools/xuva_scan.py X:\ --ffprobe "C:\ffmpeg\bin\ffprobe.exe" --workers 2
```

If FFmpeg was installed through winget, the scanner also checks winget's local alias path automatically.

The aggregate `video_codecs` summary counts the primary playable video stream for each file. Secondary thumbnail/preview video streams are preserved in per-file records but do not inflate the headline video codec counts.

## Output

```text
data/probe-results.jsonl
data/probe-results-summary.json
```

The JSONL file contains one record per media file. The summary file contains aggregate counts for extensions, containers, codecs, subtitle formats, and prototype playback decisions.

Generate the local dashboard:

```powershell
python tools/xuva_dashboard.py data/probe-results.jsonl --out data/compatibility-dashboard.html
```

The dashboard is static HTML and stays local in `data/`.

## Why This Matters

The scanner turns a media library into a local compatibility test suite. Before Xuva has a full server, it can already identify:

- Direct-play candidates.
- Audio transcode candidates.
- Subtitle burn-in risks.
- Video transcode risks.
- Weird codecs and containers.
- Files that fail probing.
