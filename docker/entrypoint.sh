#!/bin/sh
# Xuva container entrypoint.
#
# When PUID/PGID are set, switch the in-container "xuva" user/group to those
# IDs and chown the data volume so the host-mounted media (typically owned by
# the host's media user) is readable without permission errors. This is the
# linuxserver.io convention also used by Plex and Jellyfin.
#
# When PUID/PGID are unset, the build-time defaults (1000/1000) apply and we
# avoid the usermod/chown roundtrip — `docker run` with no extras keeps working.
set -e

# Bail-out helper: print the error to stderr (becomes a container log line)
# then exit non-zero so the orchestrator restarts / surfaces a failure.
die() {
    printf 'xuva entrypoint: %s\n' "$1" >&2
    exit 1
}

CURRENT_UID="$(id -u xuva 2>/dev/null || echo 1000)"
CURRENT_GID="$(id -g xuva 2>/dev/null || echo 1000)"

if [ -n "${PUID}" ] && [ "${PUID}" != "${CURRENT_UID}" ]; then
    case "${PUID}" in
        ''|*[!0-9]*) die "PUID must be a positive integer, got '${PUID}'" ;;
    esac
    usermod -o -u "${PUID}" xuva
fi

if [ -n "${PGID}" ] && [ "${PGID}" != "${CURRENT_GID}" ]; then
    case "${PGID}" in
        ''|*[!0-9]*) die "PGID must be a positive integer, got '${PGID}'" ;;
    esac
    groupmod -o -g "${PGID}" xuva
fi

# Re-chown only when we actually changed the ID; pointless on a fresh stock
# container and slow on a large /data with millions of generated files.
if [ -n "${PUID}${PGID}" ] && [ -d "${XUVA_DATA_DIR:-/data}" ]; then
    chown xuva:xuva "${XUVA_DATA_DIR:-/data}"
fi

# Drop privileges and exec — su-exec keeps PID 1 (tini → su-exec → xuva)
# so signal handling stays clean.
exec su-exec xuva /usr/local/bin/xuva "$@"
