# Media Scanner Prototype

The first practical Vyrden engineering step is a local media scanner/prober that can run against a real library without moving or modifying files.

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
tools/vyrden_scan.py
```

Inventory mode works without FFmpeg:

```powershell
python tools/vyrden_scan.py X:\ --no-probe --limit 100
```

Deep probe mode requires `ffprobe`:

```powershell
python tools/vyrden_scan.py X:\ --ffprobe "C:\ffmpeg\bin\ffprobe.exe" --workers 2
```

## Output

```text
data/probe-results.jsonl
data/probe-results-summary.json
```

The JSONL file contains one record per media file. The summary file contains aggregate counts for extensions, containers, codecs, subtitle formats, and prototype playback decisions.

## Why This Matters

The scanner turns a media library into a local compatibility test suite. Before Vyrden has a full server, it can already identify:

- Direct-play candidates.
- Audio transcode candidates.
- Subtitle burn-in risks.
- Video transcode risks.
- Weird codecs and containers.
- Files that fail probing.
