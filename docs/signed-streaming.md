# Signed Streaming URLs

Vyrden direct media streams require a short-lived signed token when local authentication is enabled.

Token claims:

- `mediaSourceId`
- `sessionId`
- `userId`
- `deviceId`
- `exp`
- `nonce`
- `kid`

The token is HMAC-SHA256 signed. The active signing key has a key ID (`kid`), and previous keys can remain loaded for validation during rotation.

## Playback Flow

1. The player creates an authenticated playback session.
2. The player requests `POST /api/media-sources/{id}/stream-token` with the active `sessionId` and `deviceId`.
3. The server verifies that the session belongs to the authenticated user, device, and media source.
4. The server returns a signed `streamUrl`.
5. The direct stream endpoint validates token signature, expiry, session binding, user binding, device binding, and stream limits before serving the file.

## Rejection Classes

- Missing token: `401`
- Expired token: `403`
- Forged or malformed token: `403`
- Token/session/media/user/device mismatch: `403`
- Stream limit exceeded: `429`

Denied streams emit `audit.stream.denied` with user, session, media source, route path, and reason.

## Stream Limits

The runtime tracks active streams by user and by device. Defaults:

- per user: `4`
- per user/device pair: `2`

These limits are in-memory for the current server process. Persistent multi-process coordination remains future work.

## Key Rotation

The streaming service supports key IDs and a rotation entry point. Current production bootstrap uses a generated local runtime key. A later installer/config task should persist the signing key material in the protected runtime data directory so sessions can survive process restarts.

Rollback path:

- Temporarily set `VYRDEN_AUTH_DISABLED=true` in local development.
- Or restore direct stream handlers to bypass token validation while keeping auth/session middleware intact.
