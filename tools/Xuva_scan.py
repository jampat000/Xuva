#!/usr/bin/env python3
"""Local media inventory and ffprobe scanner for Xuva.

This tool only reads media files. It does not upload, copy, rename, delete, or
modify files. Output is written locally as JSONL plus a summary JSON file.
"""

from __future__ import annotations

import argparse
import concurrent.futures
import hashlib
import json
import os
import shutil
import subprocess
import sys
import threading
import time
from collections import Counter, defaultdict
from dataclasses import dataclass
from pathlib import Path
from typing import Any


MEDIA_EXTENSIONS = {
    ".3g2",
    ".3gp",
    ".avi",
    ".divx",
    ".flv",
    ".m2ts",
    ".m4v",
    ".mkv",
    ".mov",
    ".mp4",
    ".mpeg",
    ".mpg",
    ".mts",
    ".ogm",
    ".ogv",
    ".rmvb",
    ".ts",
    ".vob",
    ".webm",
    ".wmv",
}


SIDECAR_SUBTITLE_EXTENSIONS = {
    ".ass",
    ".idx",
    ".smi",
    ".srt",
    ".ssa",
    ".sub",
    ".sup",
    ".vtt",
}


SUBTITLE_EXTENSION_LABELS = {
    ".ass": "ASS",
    ".idx": "VobSub IDX",
    ".smi": "SAMI",
    ".srt": "SRT",
    ".ssa": "SSA",
    ".sub": "VobSub SUB",
    ".sup": "PGS/SUP",
    ".vtt": "WebVTT",
}


LANGUAGE_HINTS = {
    "ara": "Arabic",
    "chi": "Chinese",
    "chs": "Chinese Simplified",
    "cht": "Chinese Traditional",
    "dan": "Danish",
    "de": "German",
    "deu": "German",
    "dut": "Dutch",
    "en": "English",
    "eng": "English",
    "es": "Spanish",
    "spa": "Spanish",
    "fi": "Finnish",
    "fin": "Finnish",
    "fr": "French",
    "fre": "French",
    "fra": "French",
    "ger": "German",
    "ita": "Italian",
    "it": "Italian",
    "jpn": "Japanese",
    "kor": "Korean",
    "nl": "Dutch",
    "nor": "Norwegian",
    "por": "Portuguese",
    "pt": "Portuguese",
    "rus": "Russian",
    "sv": "Swedish",
    "swe": "Swedish",
}


DIR_ENTRY_CACHE: dict[Path, list[Path]] = {}
DIR_ENTRY_CACHE_LOCK = threading.Lock()


CLIENT_PROFILES = {
    "apple_tv_lan": {
        "containers": {"mov", "mp4", "m4v", "matroska,webm", "matroska"},
        "video": {"h264", "hevc", "mpeg4"},
        "audio": {"aac", "ac3", "eac3", "alac", "mp3", "truehd"},
        "subtitles": {"subrip", "webvtt", "mov_text", "ass", "ssa"},
    },
    "android_tv_modern": {
        "containers": {"mov", "mp4", "m4v", "matroska,webm", "matroska", "mpegts"},
        "video": {"h264", "hevc", "vp9", "av1", "mpeg4"},
        "audio": {"aac", "ac3", "eac3", "opus", "vorbis", "mp3", "flac"},
        "subtitles": {"subrip", "webvtt", "ass", "ssa"},
    },
    "browser": {
        "containers": {"mov", "mp4", "m4v", "webm", "matroska,webm"},
        "video": {"h264", "vp8", "vp9", "av1"},
        "audio": {"aac", "opus", "vorbis", "mp3"},
        "subtitles": {"webvtt", "mov_text"},
    },
}


IMAGE_SUBTITLE_CODECS = {
    "hdmv_pgs_subtitle",
    "dvd_subtitle",
    "dvb_subtitle",
    "xsub",
}


@dataclass(frozen=True)
class ScanConfig:
    root: Path
    output_base: Path
    ffprobe: str | None
    probe: bool
    workers: int
    limit: int | None
    timeout: int
    resume: bool
    hash_paths: bool


def stable_path(path: Path, hash_paths: bool) -> str:
    text = str(path)
    if not hash_paths:
        return text
    return hashlib.sha256(text.lower().encode("utf-8", errors="ignore")).hexdigest()


def library_section(root: Path, path: Path) -> str:
    try:
        relative = path.relative_to(root)
    except ValueError:
        return "Unknown"
    return relative.parts[0] if relative.parts else "Unknown"


def media_kind(section: str) -> str:
    normalized = section.casefold()
    if normalized == "movies":
        return "Movie"
    if normalized in {"tv", "tv shows", "shows", "series"}:
        return "TV"
    return "Unknown"


def record_key(path: Path, size: int, mtime_ns: int) -> str:
    raw = f"{str(path).lower()}|{size}|{mtime_ns}"
    return hashlib.sha256(raw.encode("utf-8", errors="ignore")).hexdigest()


def discover_media(root: Path, limit: int | None) -> tuple[list[Path], list[dict[str, str]]]:
    files: list[Path] = []
    errors: list[dict[str, str]] = []

    def on_error(error: OSError) -> None:
        errors.append({"path": getattr(error, "filename", ""), "error": str(error)})

    for dirpath, dirnames, filenames in os.walk(root, onerror=on_error):
        dirnames[:] = [name for name in dirnames if name not in {"$RECYCLE.BIN", "System Volume Information"}]
        for filename in filenames:
            path = Path(dirpath) / filename
            if path.suffix.lower() in MEDIA_EXTENSIONS:
                files.append(path)
                if limit is not None and len(files) >= limit:
                    return files, errors

    return files, errors


def directory_entries(path: Path) -> list[Path]:
    parent = path.parent
    with DIR_ENTRY_CACHE_LOCK:
        cached = DIR_ENTRY_CACHE.get(parent)
    if cached is not None:
        return cached

    try:
        entries = [entry for entry in parent.iterdir() if entry.is_file()]
    except OSError:
        entries = []

    with DIR_ENTRY_CACHE_LOCK:
        DIR_ENTRY_CACHE[parent] = entries
    return entries


def sidecar_matches(media_path: Path, candidate: Path) -> bool:
    if candidate.suffix.lower() not in SIDECAR_SUBTITLE_EXTENSIONS:
        return False
    media_stem = media_path.stem.casefold()
    candidate_stem = candidate.stem.casefold()
    return (
        candidate_stem == media_stem
        or candidate_stem.startswith(media_stem + ".")
        or candidate_stem.startswith(media_stem + "-")
        or candidate_stem.startswith(media_stem + "_")
    )


def infer_sidecar_language(media_path: Path, sidecar_path: Path) -> str | None:
    media_stem = media_path.stem.casefold()
    sidecar_stem = sidecar_path.stem.casefold()
    suffix = sidecar_stem.removeprefix(media_stem).strip(" ._-")
    tokens = [token for token in suffix.replace("-", ".").replace("_", ".").split(".") if token]
    for token in tokens:
        language = LANGUAGE_HINTS.get(token)
        if language:
            return language
    return None


def sidecar_tokens(media_path: Path, sidecar_path: Path) -> set[str]:
    media_stem = media_path.stem.casefold()
    sidecar_stem = sidecar_path.stem.casefold()
    suffix = sidecar_stem.removeprefix(media_stem).strip(" ._-")
    return {token for token in suffix.replace("-", ".").replace("_", ".").split(".") if token}


def detect_sidecars(path: Path, hash_paths: bool) -> list[dict[str, Any]]:
    sidecars: list[dict[str, Any]] = []
    for candidate in directory_entries(path):
        if not sidecar_matches(path, candidate):
            continue
        tokens = sidecar_tokens(path, candidate)
        sidecars.append(
            {
                "path": stable_path(candidate, hash_paths),
                "name": candidate.name if not hash_paths else None,
                "extension": candidate.suffix.lower(),
                "format": SUBTITLE_EXTENSION_LABELS.get(candidate.suffix.lower(), candidate.suffix.lower().lstrip(".").upper()),
                "language": infer_sidecar_language(path, candidate),
                "forced": "forced" in tokens,
                "hearing_impaired": bool(tokens & {"sdh", "hi", "hearing"}),
            }
        )
    return [{key: value for key, value in item.items() if value is not None} for item in sidecars]


def empty_summary() -> dict[str, Any]:
    return {
        "files_scanned": 0,
        "bytes_scanned": 0,
        "extensions": Counter(),
        "probe_status": Counter(),
        "containers": Counter(),
        "video_codecs": Counter(),
        "audio_codecs": Counter(),
        "subtitle_codecs": Counter(),
        "sidecar_subtitles": Counter(),
        "library_sections": Counter(),
        "media_kinds": Counter(),
        "files_with_embedded_subtitles": 0,
        "files_with_sidecar_subtitles": 0,
        "files_with_any_subtitles": 0,
        "decisions": defaultdict(Counter),
    }


def load_done_keys(jsonl_path: Path, summary: dict[str, Any] | None = None) -> set[str]:
    if not jsonl_path.exists():
        return set()

    done: set[str] = set()
    with jsonl_path.open("r", encoding="utf-8") as handle:
        for line in handle:
            line = line.strip()
            if not line:
                continue
            try:
                item = json.loads(line)
            except json.JSONDecodeError:
                continue
            key = item.get("key")
            if isinstance(key, str):
                done.add(key)
                if summary is not None:
                    update_summary(summary, item)
    return done


def run_ffprobe(path: Path, ffprobe: str, timeout: int) -> tuple[dict[str, Any] | None, str | None]:
    command = [
        ffprobe,
        "-v",
        "error",
        "-print_format",
        "json",
        "-show_format",
        "-show_streams",
        "-show_chapters",
        str(path),
    ]
    try:
        result = subprocess.run(
            command,
            check=False,
            capture_output=True,
            text=True,
            timeout=timeout,
            encoding="utf-8",
            errors="replace",
        )
    except subprocess.TimeoutExpired:
        return None, f"ffprobe timed out after {timeout}s"
    except OSError as error:
        return None, str(error)

    if result.returncode != 0:
        stderr = result.stderr.strip() or f"ffprobe exited {result.returncode}"
        return None, stderr[:1200]

    try:
        return json.loads(result.stdout), None
    except json.JSONDecodeError as error:
        return None, f"ffprobe returned invalid JSON: {error}"


def stream_groups(probe: dict[str, Any] | None) -> dict[str, list[dict[str, Any]]]:
    groups: dict[str, list[dict[str, Any]]] = {
        "video": [],
        "audio": [],
        "subtitle": [],
        "attachment": [],
        "other": [],
    }
    if not probe:
        return groups
    for stream in probe.get("streams", []):
        if not isinstance(stream, dict):
            continue
        disposition = stream.get("disposition")
        if isinstance(disposition, dict) and disposition.get("attached_pic") == 1:
            groups["attachment"].append(stream)
            continue
        codec_type = stream.get("codec_type")
        if codec_type in groups:
            groups[codec_type].append(stream)
        else:
            groups["other"].append(stream)
    return groups


def codec_name(stream: dict[str, Any] | None) -> str | None:
    if not stream:
        return None
    value = stream.get("codec_name")
    return value if isinstance(value, str) else None


def format_name(probe: dict[str, Any] | None) -> str | None:
    if not probe:
        return None
    fmt = probe.get("format")
    if not isinstance(fmt, dict):
        return None
    value = fmt.get("format_name")
    return value if isinstance(value, str) else None


def container_supported(container: str | None, supported: set[str]) -> bool:
    if not container:
        return True
    names = {name.strip().lower() for name in container.split(",") if name.strip()}
    names.add(container.lower())
    return bool(names & supported)


def bitrate(probe: dict[str, Any] | None) -> int | None:
    if not probe:
        return None
    fmt = probe.get("format")
    if not isinstance(fmt, dict):
        return None
    value = fmt.get("bit_rate")
    try:
        return int(value)
    except (TypeError, ValueError):
        return None


def duration_seconds(probe: dict[str, Any] | None) -> float | None:
    if not probe:
        return None
    fmt = probe.get("format")
    if not isinstance(fmt, dict):
        return None
    value = fmt.get("duration")
    try:
        return float(value)
    except (TypeError, ValueError):
        return None


def choose_default(streams: list[dict[str, Any]]) -> dict[str, Any] | None:
    for stream in streams:
        disposition = stream.get("disposition")
        if isinstance(disposition, dict) and disposition.get("default") == 1:
            return stream
    return streams[0] if streams else None


def decide_playback(probe: dict[str, Any] | None) -> dict[str, dict[str, Any]]:
    if not probe:
        return {}

    groups = stream_groups(probe)
    selected_video = choose_default(groups["video"])
    selected_audio = choose_default(groups["audio"])
    selected_subtitle = choose_default(groups["subtitle"])
    container = format_name(probe)
    video_codec = codec_name(selected_video)
    audio_codec = codec_name(selected_audio)
    subtitle_codec = codec_name(selected_subtitle)

    decisions: dict[str, dict[str, Any]] = {}
    for profile_name, profile in CLIENT_PROFILES.items():
        mode = "Direct Play"
        reason = "Container, video, audio, and selected subtitles match this client profile."
        actions = {
            "container": "direct",
            "video": "direct",
            "audio": "direct",
            "subtitle": "direct" if subtitle_codec else "none",
        }

        if video_codec and video_codec not in profile["video"]:
            mode = "Video Transcode"
            actions["video"] = "transcode"
            reason = f"Video codec {video_codec} is not in the {profile_name} direct-play profile."
        elif subtitle_codec and subtitle_codec in IMAGE_SUBTITLE_CODECS:
            mode = "Subtitle Burn"
            actions["subtitle"] = "burn"
            actions["video"] = "transcode"
            reason = f"Selected subtitle codec {subtitle_codec} is image-based and may require burn-in."
        elif audio_codec and audio_codec not in profile["audio"]:
            mode = "Audio Transcode"
            actions["audio"] = "transcode"
            reason = f"Audio codec {audio_codec} is not in the {profile_name} direct-play profile."
        elif not container_supported(container, profile["containers"]):
            mode = "Remux"
            actions["container"] = "remux"
            reason = f"Container {container} is not in the {profile_name} direct-play profile."

        decisions[profile_name] = {
            "mode": mode,
            "reason": reason,
            "actions": actions,
            "selected": {
                "container": container,
                "video_codec": video_codec,
                "audio_codec": audio_codec,
                "subtitle_codec": subtitle_codec,
            },
        }

    return decisions


def summarize_streams(probe: dict[str, Any] | None) -> dict[str, Any]:
    groups = stream_groups(probe)
    return {
        "container": format_name(probe),
        "duration_seconds": duration_seconds(probe),
        "bitrate": bitrate(probe),
        "video": [summarize_stream(stream) for stream in groups["video"]],
        "audio": [summarize_stream(stream) for stream in groups["audio"]],
        "subtitles": [summarize_stream(stream) for stream in groups["subtitle"]],
        "attachments": [summarize_stream(stream) for stream in groups["attachment"]],
        "chapters": len(probe.get("chapters", [])) if probe else 0,
    }


def summarize_stream(stream: dict[str, Any]) -> dict[str, Any]:
    tags = stream.get("tags") if isinstance(stream.get("tags"), dict) else {}
    disposition = stream.get("disposition") if isinstance(stream.get("disposition"), dict) else {}
    item = {
        "index": stream.get("index"),
        "codec": stream.get("codec_name"),
        "codec_long_name": stream.get("codec_long_name"),
        "profile": stream.get("profile"),
        "language": tags.get("language"),
        "title": tags.get("title"),
        "default": disposition.get("default") == 1,
        "forced": disposition.get("forced") == 1,
    }
    if stream.get("codec_type") == "video":
        item.update(
            {
                "width": stream.get("width"),
                "height": stream.get("height"),
                "pix_fmt": stream.get("pix_fmt"),
                "color_transfer": stream.get("color_transfer"),
                "color_space": stream.get("color_space"),
                "color_primaries": stream.get("color_primaries"),
                "avg_frame_rate": stream.get("avg_frame_rate"),
            }
        )
    if stream.get("codec_type") == "audio":
        item.update(
            {
                "channels": stream.get("channels"),
                "channel_layout": stream.get("channel_layout"),
                "sample_rate": stream.get("sample_rate"),
            }
        )
    return {key: value for key, value in item.items() if value is not None}


def scan_one(path: Path, config: ScanConfig) -> dict[str, Any]:
    started = time.time()
    try:
        stat = path.stat()
    except OSError as error:
        return {
            "path": stable_path(path, config.hash_paths),
            "error": str(error),
            "scanned_at": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
        }

    key = record_key(path, stat.st_size, stat.st_mtime_ns)
    probe: dict[str, Any] | None = None
    probe_error: str | None = None
    if config.probe and config.ffprobe:
        probe, probe_error = run_ffprobe(path, config.ffprobe, config.timeout)

    section = library_section(config.root, path)
    sidecars = detect_sidecars(path, config.hash_paths)
    record = {
        "key": key,
        "path": stable_path(path, config.hash_paths),
        "name": path.name if not config.hash_paths else None,
        "library_section": section,
        "media_kind": media_kind(section),
        "extension": path.suffix.lower(),
        "size_bytes": stat.st_size,
        "modified_time_ns": stat.st_mtime_ns,
        "probe_status": "ok" if probe else "not_run",
        "probe_error": probe_error,
        "sidecars": sidecars,
        "streams": summarize_streams(probe),
        "decisions": decide_playback(probe),
        "elapsed_ms": round((time.time() - started) * 1000),
        "scanned_at": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
    }
    if probe_error:
        record["probe_status"] = "error"
    return {key: value for key, value in record.items() if value is not None}


def write_json(path: Path, payload: dict[str, Any]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    tmp = path.with_suffix(path.suffix + ".tmp")
    with tmp.open("w", encoding="utf-8") as handle:
        json.dump(payload, handle, indent=2, sort_keys=True)
        handle.write("\n")
    tmp.replace(path)


def update_summary(summary: dict[str, Any], record: dict[str, Any]) -> None:
    summary["files_scanned"] += 1
    summary["bytes_scanned"] += int(record.get("size_bytes") or 0)
    summary["extensions"][record.get("extension") or "unknown"] += 1
    summary["probe_status"][record.get("probe_status") or "unknown"] += 1
    summary["library_sections"][record.get("library_section") or "Unknown"] += 1
    summary["media_kinds"][record.get("media_kind") or "Unknown"] += 1

    streams = record.get("streams")
    if not isinstance(streams, dict):
        return

    container = streams.get("container")
    if container:
        summary["containers"][container] += 1

    video_streams = streams.get("video", []) if isinstance(streams.get("video"), list) else []
    for stream in video_streams[:1]:
        codec = stream.get("codec")
        if codec:
            summary["video_codecs"][codec] += 1
    for stream in streams.get("audio", []) if isinstance(streams.get("audio"), list) else []:
        codec = stream.get("codec")
        if codec:
            summary["audio_codecs"][codec] += 1
    embedded_subtitles = streams.get("subtitles", []) if isinstance(streams.get("subtitles"), list) else []
    for stream in embedded_subtitles:
        codec = stream.get("codec")
        if codec:
            summary["subtitle_codecs"][codec] += 1

    sidecars = record.get("sidecars")
    sidecar_count = 0
    if isinstance(sidecars, list):
        for sidecar in sidecars:
            if isinstance(sidecar, dict):
                extension = sidecar.get("extension")
                if extension:
                    summary["sidecar_subtitles"][extension] += 1
                    sidecar_count += 1

    if embedded_subtitles:
        summary["files_with_embedded_subtitles"] += 1
    if sidecar_count:
        summary["files_with_sidecar_subtitles"] += 1
    if embedded_subtitles or sidecar_count:
        summary["files_with_any_subtitles"] += 1

    decisions = record.get("decisions")
    if isinstance(decisions, dict):
        for profile, decision in decisions.items():
            if isinstance(decision, dict):
                mode = decision.get("mode")
                if mode:
                    summary["decisions"][profile][mode] += 1


def finalize_summary(summary: dict[str, Any], config: ScanConfig, discovered: int, skipped: int, errors: list[dict[str, str]]) -> dict[str, Any]:
    output = {
        "root": str(config.root),
        "output_base": str(config.output_base),
        "ffprobe": config.ffprobe,
        "probe_enabled": config.probe and bool(config.ffprobe),
        "workers": config.workers,
        "limit": config.limit,
        "discovered_media_files": discovered,
        "skipped_existing_records": skipped,
        "walk_errors": errors,
        "files_scanned": summary["files_scanned"],
        "bytes_scanned": summary["bytes_scanned"],
    }
    for key in [
        "extensions",
        "probe_status",
        "containers",
        "video_codecs",
        "audio_codecs",
        "subtitle_codecs",
        "sidecar_subtitles",
        "library_sections",
        "media_kinds",
    ]:
        output[key] = dict(summary[key].most_common())
    output["files_with_embedded_subtitles"] = summary["files_with_embedded_subtitles"]
    output["files_with_sidecar_subtitles"] = summary["files_with_sidecar_subtitles"]
    output["files_with_any_subtitles"] = summary["files_with_any_subtitles"]
    output["decisions"] = {
        profile: dict(counter.most_common())
        for profile, counter in summary["decisions"].items()
    }
    return output


def parse_args(argv: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Scan a local or mapped NAS media folder for Xuva.")
    parser.add_argument("root", help="Media root, for example X:\\ or Z:\\Media")
    parser.add_argument("--out", default="data/probe-results", help="Output base path without extension.")
    parser.add_argument("--ffprobe", default=None, help="Path to ffprobe.exe. Defaults to ffprobe on PATH.")
    parser.add_argument("--no-probe", action="store_true", help="Inventory files only; do not run ffprobe.")
    parser.add_argument("--workers", type=int, default=2, help="Parallel file workers. Keep low for NAS scans.")
    parser.add_argument("--limit", type=int, default=None, help="Stop after discovering this many media files.")
    parser.add_argument("--timeout", type=int, default=90, help="ffprobe timeout per file in seconds.")
    parser.add_argument("--no-resume", action="store_true", help="Do not skip records already present in JSONL.")
    parser.add_argument("--hash-paths", action="store_true", help="Hash file paths in output for safer sharing.")
    return parser.parse_args(argv)


def find_ffprobe(explicit_path: str | None) -> str | None:
    if explicit_path:
        return explicit_path

    from_path = shutil.which("ffprobe")
    if from_path:
        return from_path

    local_app_data = os.environ.get("LOCALAPPDATA")
    if local_app_data:
        candidates = [
            Path(local_app_data) / "Microsoft" / "WinGet" / "Links" / "ffprobe.exe",
        ]
        packages = Path(local_app_data) / "Microsoft" / "WinGet" / "Packages"
        if packages.exists():
            candidates.extend(packages.glob("Gyan.FFmpeg_*/*/bin/ffprobe.exe"))
            candidates.extend(packages.glob("BtbN.FFmpeg.*/*/bin/ffprobe.exe"))
        for candidate in candidates:
            if candidate.exists():
                return str(candidate)

    return None


def main(argv: list[str]) -> int:
    args = parse_args(argv)
    root = Path(args.root).expanduser()
    if not root.exists():
        print(f"Media root does not exist: {root}", file=sys.stderr)
        return 2

    ffprobe = find_ffprobe(args.ffprobe)
    probe = not args.no_probe and bool(ffprobe)
    if not args.no_probe and not ffprobe:
        print("ffprobe was not found; running inventory-only mode. Use --ffprobe or install FFmpeg for deep probing.", file=sys.stderr)

    config = ScanConfig(
        root=root,
        output_base=Path(args.out),
        ffprobe=ffprobe,
        probe=probe,
        workers=max(1, args.workers),
        limit=args.limit,
        timeout=max(5, args.timeout),
        resume=not args.no_resume,
        hash_paths=args.hash_paths,
    )

    jsonl_path = config.output_base.with_suffix(".jsonl")
    summary_path = config.output_base.with_name(config.output_base.name + "-summary").with_suffix(".json")
    jsonl_path.parent.mkdir(parents=True, exist_ok=True)

    summary = empty_summary()

    files, walk_errors = discover_media(config.root, config.limit)
    done = load_done_keys(jsonl_path, summary) if config.resume else set()
    pending: list[Path] = []
    skipped = 0
    for path in files:
        try:
            stat = path.stat()
        except OSError:
            pending.append(path)
            continue
        if record_key(path, stat.st_size, stat.st_mtime_ns) in done:
            skipped += 1
        else:
            pending.append(path)

    print(f"Discovered {len(files)} media files under {config.root}")
    print(f"Pending {len(pending)} files, skipped {skipped} existing records")
    print(f"Writing {jsonl_path} and {summary_path}")

    with jsonl_path.open("a", encoding="utf-8") as handle:
        with concurrent.futures.ThreadPoolExecutor(max_workers=config.workers) as executor:
            futures = [executor.submit(scan_one, path, config) for path in pending]
            for index, future in enumerate(concurrent.futures.as_completed(futures), start=1):
                record = future.result()
                handle.write(json.dumps(record, ensure_ascii=False, sort_keys=True) + "\n")
                if index % 10 == 0:
                    handle.flush()
                update_summary(summary, record)
                if index == 1 or index % 25 == 0 or index == len(pending):
                    status = record.get("probe_status")
                    print(f"[{index}/{len(pending)}] {status} {record.get('extension')} {record.get('path')}")

    output = finalize_summary(summary, config, len(files), skipped, walk_errors)
    write_json(summary_path, output)
    print(json.dumps(output, indent=2, sort_keys=True))
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
