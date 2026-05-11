#!/usr/bin/env python3
"""Generate a local Lorivo compatibility dashboard from scanner JSONL output."""

from __future__ import annotations

import argparse
import html
import json
import statistics
from collections import Counter, defaultdict
from pathlib import Path
from typing import Any


CLIENT_LABELS = {
    "apple_tv_lan": "Apple TV LAN",
    "android_tv_modern": "Android TV Modern",
    "browser": "Browser",
}


def load_records(path: Path) -> list[dict[str, Any]]:
    records: list[dict[str, Any]] = []
    with path.open("r", encoding="utf-8") as handle:
        for line in handle:
            line = line.strip()
            if not line:
                continue
            records.append(json.loads(line))
    return records


def fmt_int(value: int | float | None) -> str:
    if value is None:
        return "0"
    return f"{int(value):,}"


def fmt_bytes(value: int | float | None) -> str:
    if not value:
        return "0 B"
    units = ["B", "KB", "MB", "GB", "TB", "PB"]
    amount = float(value)
    unit = units[0]
    for unit in units:
        if amount < 1024 or unit == units[-1]:
            break
        amount /= 1024
    if amount >= 100:
        return f"{amount:,.0f} {unit}"
    return f"{amount:,.1f} {unit}"


def fmt_bitrate(value: int | float | None) -> str:
    if not value:
        return "0 Mbps"
    amount = float(value)
    if amount >= 1_000_000_000:
        return f"{amount / 1_000_000_000:,.2f} Gbps"
    if amount >= 1_000_000:
        return f"{amount / 1_000_000:,.1f} Mbps"
    if amount >= 1_000:
        return f"{amount / 1_000:,.0f} Kbps"
    return f"{amount:,.0f} bps"


def fmt_percent(count: int, total: int) -> str:
    if total <= 0:
        return "0%"
    return f"{(count / total) * 100:.1f}%"


def streams(record: dict[str, Any], key: str) -> list[dict[str, Any]]:
    stream_data = record.get("streams")
    if not isinstance(stream_data, dict):
        return []
    value = stream_data.get(key)
    return value if isinstance(value, list) else []


def primary_video(record: dict[str, Any]) -> dict[str, Any]:
    videos = streams(record, "video")
    return videos[0] if videos else {}


def all_audio(record: dict[str, Any]) -> list[dict[str, Any]]:
    return streams(record, "audio")


def embedded_subtitles(record: dict[str, Any]) -> list[dict[str, Any]]:
    return streams(record, "subtitles")


def sidecars(record: dict[str, Any]) -> list[dict[str, Any]]:
    value = record.get("sidecars")
    return value if isinstance(value, list) else []


def decision_mode(record: dict[str, Any], profile: str) -> str:
    decisions = record.get("decisions")
    if not isinstance(decisions, dict):
        return "Unknown"
    decision = decisions.get(profile)
    if not isinstance(decision, dict):
        return "Unknown"
    value = decision.get("mode")
    return value if isinstance(value, str) else "Unknown"


def bitrate(record: dict[str, Any]) -> int:
    stream_data = record.get("streams")
    if isinstance(stream_data, dict):
        value = stream_data.get("bitrate")
        if isinstance(value, int):
            return value
        if isinstance(value, str) and value.isdigit():
            return int(value)
    return 0


def duration(record: dict[str, Any]) -> float:
    stream_data = record.get("streams")
    if isinstance(stream_data, dict):
        value = stream_data.get("duration_seconds")
        if isinstance(value, (int, float)):
            return float(value)
    return 0.0


def record_label(record: dict[str, Any]) -> str:
    name = record.get("name")
    if isinstance(name, str) and name:
        return name
    path = record.get("path")
    if isinstance(path, str) and path:
        return path
    return record.get("key", "Unknown")


def counter_table(counter: Counter[str], total: int, title: str) -> str:
    rows = []
    for label, count in counter.most_common(12):
        rows.append(
            f"<tr><td>{html.escape(str(label))}</td><td>{fmt_int(count)}</td><td>{fmt_percent(count, total)}</td></tr>"
        )
    return f"""
      <section class="panel">
        <h2>{html.escape(title)}</h2>
        <table>
          <thead><tr><th>Type</th><th>Count</th><th>Share</th></tr></thead>
          <tbody>{''.join(rows)}</tbody>
        </table>
      </section>
    """


def decision_table(decisions: dict[str, Counter[str]], total: int) -> str:
    cards = []
    for profile, counter in decisions.items():
        rows = []
        for mode, count in counter.most_common():
            rows.append(
                f"<tr><td>{html.escape(mode)}</td><td>{fmt_int(count)}</td><td>{fmt_percent(count, total)}</td></tr>"
            )
        cards.append(
            f"""
            <section class="panel">
              <h2>{html.escape(CLIENT_LABELS.get(profile, profile))}</h2>
              <table>
                <thead><tr><th>Mode</th><th>Files</th><th>Share</th></tr></thead>
                <tbody>{''.join(rows)}</tbody>
              </table>
            </section>
            """
        )
    return "".join(cards)


def file_table(records: list[dict[str, Any]], title: str, subtitle: str) -> str:
    rows = []
    for record in records[:20]:
        video = primary_video(record)
        audio_codecs = Counter(stream.get("codec") or "unknown" for stream in all_audio(record))
        sub_count = len(embedded_subtitles(record)) + len(sidecars(record))
        rows.append(
            "<tr>"
            f"<td>{html.escape(record_label(record))}</td>"
            f"<td>{html.escape(str(record.get('media_kind') or 'Unknown'))}</td>"
            f"<td>{html.escape(str(video.get('codec') or 'unknown'))}</td>"
            f"<td>{fmt_bytes(record.get('size_bytes'))}</td>"
            f"<td>{fmt_bitrate(bitrate(record))}</td>"
            f"<td>{html.escape(', '.join(f'{codec} x{count}' for codec, count in audio_codecs.items()) or 'none')}</td>"
            f"<td>{fmt_int(sub_count)}</td>"
            "</tr>"
        )
    return f"""
      <section class="panel wide">
        <div class="panel-head">
          <h2>{html.escape(title)}</h2>
          <span>{html.escape(subtitle)}</span>
        </div>
        <table>
          <thead><tr><th>File</th><th>Kind</th><th>Video</th><th>Size</th><th>Bitrate</th><th>Audio</th><th>Subs</th></tr></thead>
          <tbody>{''.join(rows)}</tbody>
        </table>
      </section>
    """


def build_stats(records: list[dict[str, Any]]) -> dict[str, Any]:
    total = len(records)
    total_bytes = sum(int(record.get("size_bytes") or 0) for record in records)
    bitrates = [bitrate(record) for record in records if bitrate(record) > 0]
    sizes = [int(record.get("size_bytes") or 0) for record in records if int(record.get("size_bytes") or 0) > 0]

    media_kinds = Counter(str(record.get("media_kind") or "Unknown") for record in records)
    sections = Counter(str(record.get("library_section") or "Unknown") for record in records)
    containers = Counter(str(record.get("streams", {}).get("container") or "unknown") for record in records if isinstance(record.get("streams"), dict))
    extensions = Counter(str(record.get("extension") or "unknown") for record in records)
    video_codecs = Counter(str(primary_video(record).get("codec") or "unknown") for record in records)
    audio_codecs: Counter[str] = Counter()
    embedded_subtitle_codecs: Counter[str] = Counter()
    sidecar_subtitle_formats: Counter[str] = Counter()
    decision_counts: dict[str, Counter[str]] = defaultdict(Counter)

    files_with_embedded_subs = 0
    files_with_sidecars = 0
    files_with_any_subs = 0
    audio_transcode_causes: Counter[str] = Counter()
    browser_video_transcode_causes: Counter[str] = Counter()

    for record in records:
        for audio in all_audio(record):
            audio_codecs[str(audio.get("codec") or "unknown")] += 1
        embedded = embedded_subtitles(record)
        for subtitle in embedded:
            embedded_subtitle_codecs[str(subtitle.get("codec") or "unknown")] += 1
        sc = sidecars(record)
        for sidecar in sc:
            sidecar_subtitle_formats[str(sidecar.get("extension") or "unknown")] += 1
        if embedded:
            files_with_embedded_subs += 1
        if sc:
            files_with_sidecars += 1
        if embedded or sc:
            files_with_any_subs += 1

        for profile in CLIENT_LABELS:
            mode = decision_mode(record, profile)
            decision_counts[profile][mode] += 1

        for profile in ["apple_tv_lan", "android_tv_modern", "browser"]:
            if decision_mode(record, profile) == "Audio Transcode":
                decisions = record.get("decisions", {})
                selected = decisions.get(profile, {}).get("selected", {}) if isinstance(decisions, dict) else {}
                audio_transcode_causes[str(selected.get("audio_codec") or "unknown")] += 1

        if decision_mode(record, "browser") == "Video Transcode":
            browser_video_transcode_causes[str(primary_video(record).get("codec") or "unknown")] += 1

    largest = sorted(records, key=lambda record: int(record.get("size_bytes") or 0), reverse=True)
    highest_bitrate = sorted(records, key=bitrate, reverse=True)
    browser_transcodes = [record for record in records if decision_mode(record, "browser") == "Video Transcode"]
    subtitle_risk = [
        record
        for record in records
        if not embedded_subtitles(record) and not sidecars(record)
    ]

    return {
        "total": total,
        "total_bytes": total_bytes,
        "average_size": statistics.mean(sizes) if sizes else 0,
        "median_size": statistics.median(sizes) if sizes else 0,
        "average_bitrate": statistics.mean(bitrates) if bitrates else 0,
        "media_kinds": media_kinds,
        "sections": sections,
        "containers": containers,
        "extensions": extensions,
        "video_codecs": video_codecs,
        "audio_codecs": audio_codecs,
        "embedded_subtitle_codecs": embedded_subtitle_codecs,
        "sidecar_subtitle_formats": sidecar_subtitle_formats,
        "decision_counts": decision_counts,
        "files_with_embedded_subs": files_with_embedded_subs,
        "files_with_sidecars": files_with_sidecars,
        "files_with_any_subs": files_with_any_subs,
        "audio_transcode_causes": audio_transcode_causes,
        "browser_video_transcode_causes": browser_video_transcode_causes,
        "largest": largest,
        "highest_bitrate": highest_bitrate,
        "browser_transcodes": browser_transcodes,
        "subtitle_risk": subtitle_risk,
    }


def render_dashboard(records: list[dict[str, Any]], source: Path) -> str:
    stats = build_stats(records)
    total = stats["total"]
    cards = [
        ("Files", fmt_int(total), "Media files scanned"),
        ("Library Size", fmt_bytes(stats["total_bytes"]), "Bytes across scanned files"),
        ("Movies", fmt_int(stats["media_kinds"].get("Movie", 0)), "Files under Movies"),
        ("TV", fmt_int(stats["media_kinds"].get("TV", 0)), "Files under TV"),
        ("Any Subtitles", fmt_int(stats["files_with_any_subs"]), fmt_percent(stats["files_with_any_subs"], total)),
        ("Avg Bitrate", fmt_bitrate(stats["average_bitrate"]), "Primary stream estimate"),
    ]

    card_html = "".join(
        f"<section class=\"metric\"><span>{html.escape(label)}</span><strong>{html.escape(value)}</strong><em>{html.escape(note)}</em></section>"
        for label, value, note in cards
    )

    generated = html.escape(source.name)
    body = f"""
      <header>
        <div class="brand"><div class="mark">V</div><div><h1>Lorivo Compatibility Dashboard</h1><p>Local library intelligence generated from <code>{generated}</code>.</p></div></div>
      </header>
      <section class="metrics">{card_html}</section>
      <section class="grid three">{decision_table(stats["decision_counts"], total)}</section>
      <section class="grid">
        {counter_table(stats["media_kinds"], total, "Movies vs TV")}
        {counter_table(stats["video_codecs"], total, "Primary Video Codecs")}
        {counter_table(stats["audio_codecs"], sum(stats["audio_codecs"].values()), "Audio Streams")}
        {counter_table(stats["embedded_subtitle_codecs"], max(sum(stats["embedded_subtitle_codecs"].values()), 1), "Embedded Subtitles")}
        {counter_table(stats["sidecar_subtitle_formats"], max(sum(stats["sidecar_subtitle_formats"].values()), 1), "Sidecar Subtitles")}
        {counter_table(stats["audio_transcode_causes"], max(sum(stats["audio_transcode_causes"].values()), 1), "Audio Transcode Causes")}
        {counter_table(stats["browser_video_transcode_causes"], max(sum(stats["browser_video_transcode_causes"].values()), 1), "Browser Video Transcode Causes")}
        {counter_table(stats["containers"], total, "Containers")}
      </section>
      {file_table(stats["largest"], "Largest Files", "Good candidates for version/download optimization")}
      {file_table(stats["highest_bitrate"], "Highest Bitrate", "Likely to stress remote playback")}
      {file_table(stats["browser_transcodes"], "Browser Video Transcode", "Usually HEVC or AV1 browser support gaps")}
      {file_table(stats["subtitle_risk"], "No Subtitle Tracks Detected", "Files without embedded or sidecar subtitles")}
    """

    return f"""<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Lorivo Compatibility Dashboard</title>
    <style>
      :root {{
        color-scheme: dark;
        --bg: #080705;
        --panel: #14120f;
        --text: #fff7ea;
        --muted: #b9b0a2;
        --quiet: #746c60;
        --line: #312b23;
        --focus: #4fe6c8;
        --action: #f6b756;
      }}
      * {{ box-sizing: border-box; }}
      body {{
        margin: 0;
        min-height: 100vh;
        background:
          linear-gradient(132deg, rgba(246, 183, 86, 0.24) 0 13%, transparent 13% 45%, rgba(79, 230, 200, 0.2) 45% 62%, transparent 62%),
          linear-gradient(135deg, #080705 0%, #11100d 52%, #102520 100%);
        color: var(--text);
        font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      }}
      main {{
        width: min(1540px, calc(100vw - 48px));
        margin: 0 auto;
        padding: 44px 0;
      }}
      header {{
        display: flex;
        align-items: center;
        justify-content: space-between;
        margin-bottom: 26px;
      }}
      .brand {{
        display: flex;
        align-items: center;
        gap: 16px;
      }}
      .mark {{
        display: grid;
        place-items: center;
        width: 54px;
        height: 54px;
        border: 1px solid rgba(79, 230, 200, 0.36);
        border-radius: 14px;
        background: rgba(79, 230, 200, 0.08);
        color: var(--focus);
        font-size: 28px;
        font-weight: 780;
      }}
      h1 {{
        margin: 0;
        font-size: 42px;
        line-height: 1;
        letter-spacing: 0;
      }}
      h2 {{
        margin: 0 0 14px;
        font-size: 19px;
      }}
      p {{
        margin: 8px 0 0;
        color: var(--muted);
        font-size: 15px;
      }}
      code {{
        color: var(--focus);
      }}
      .metrics {{
        display: grid;
        grid-template-columns: repeat(6, minmax(0, 1fr));
        gap: 14px;
        margin-bottom: 18px;
      }}
      .metric,
      .panel {{
        border: 1px solid rgba(255, 247, 234, 0.12);
        border-radius: 10px;
        background: rgba(20, 18, 15, 0.8);
        box-shadow: 0 26px 70px rgba(0, 0, 0, 0.28);
        backdrop-filter: blur(20px);
      }}
      .metric {{
        display: grid;
        gap: 7px;
        min-height: 126px;
        padding: 16px;
      }}
      .metric span,
      .panel-head span {{
        color: var(--quiet);
        font-size: 12px;
        font-weight: 820;
        text-transform: uppercase;
      }}
      .metric strong {{
        font-size: 28px;
        line-height: 1;
      }}
      .metric em {{
        color: var(--muted);
        font-size: 13px;
        font-style: normal;
        font-weight: 680;
      }}
      .grid {{
        display: grid;
        grid-template-columns: repeat(2, minmax(0, 1fr));
        gap: 18px;
        margin-bottom: 18px;
      }}
      .grid.three {{
        grid-template-columns: repeat(3, minmax(0, 1fr));
      }}
      .panel {{
        overflow: hidden;
        padding: 18px;
      }}
      .panel.wide {{
        margin-bottom: 18px;
      }}
      .panel-head {{
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 18px;
        margin-bottom: 14px;
      }}
      table {{
        width: 100%;
        border-collapse: collapse;
      }}
      th,
      td {{
        padding: 10px 8px;
        border-top: 1px solid rgba(255, 247, 234, 0.08);
        text-align: left;
        vertical-align: top;
        font-size: 13px;
      }}
      th {{
        color: var(--quiet);
        font-size: 11px;
        font-weight: 820;
        text-transform: uppercase;
      }}
      td {{
        color: var(--text);
        font-weight: 690;
      }}
      td:first-child {{
        max-width: 520px;
        overflow-wrap: anywhere;
      }}
      @media (max-width: 1180px) {{
        .metrics,
        .grid,
        .grid.three {{
          grid-template-columns: 1fr;
        }}
      }}
    </style>
  </head>
  <body>
    <main>{body}</main>
  </body>
</html>
"""


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Generate a local Lorivo compatibility dashboard.")
    parser.add_argument("jsonl", help="Scanner JSONL path.")
    parser.add_argument("--out", default="data/compatibility-dashboard.html", help="Output HTML path.")
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    source = Path(args.jsonl)
    records = load_records(source)
    output = Path(args.out)
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_text(render_dashboard(records, source), encoding="utf-8")
    print(f"Wrote {output} from {len(records):,} records")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
