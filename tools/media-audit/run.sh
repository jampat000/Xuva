#!/usr/bin/env bash
# tools/media-audit/run.sh — convenience wrapper around server/cmd/media-audit.
#
# Usage:
#     tools/media-audit/run.sh /path/to/library [/path/to/another/library ...]
#
# Output (audit-report.json + audit-summary.txt) lands in $XUVA_AUDIT_OUT_DIR
# if set, otherwise the current working directory.
#
# ffprobe is taken from $XUVA_FFPROBE if set, otherwise PATH.
#
# Run as a one-off:
#     tools/media-audit/run.sh /mnt/media/Movies /mnt/media/TV
#
# Run overnight (resumable on next start):
#     nohup tools/media-audit/run.sh /mnt/media/Movies > audit.log 2>&1 &
#
# Cancel cleanly:
#     kill -SIGTERM <pid>   # the tool writes whatever it has so far

set -euo pipefail

if [ "$#" -lt 1 ]; then
    echo "usage: $0 <library-path> [<library-path>...]" >&2
    exit 2
fi

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
OUT_DIR="${XUVA_AUDIT_OUT_DIR:-$(pwd)}"
FFPROBE="${XUVA_FFPROBE:-ffprobe}"

args=(--out-dir "$OUT_DIR" --ffprobe "$FFPROBE")
for lib in "$@"; do
    args+=(--library "$lib")
done

cd "$REPO_ROOT/server"
exec go run ./cmd/media-audit "${args[@]}"
