# Security Baseline

Xuva keeps the local server secure by default while still supporting a LAN-first media workflow.

## Web and API Headers

Every router response passes through the security middleware and receives:

- `Content-Security-Policy`: restricts scripts, media, images, and connections to the local app origin.
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `Referrer-Policy: no-referrer`
- `Cross-Origin-Resource-Policy: same-origin`
- `Permissions-Policy`: disables camera, microphone, and geolocation by default.

## CORS Policy

Browser requests with no `Origin` header are treated as direct local/server requests.

Browser requests with an `Origin` header are allowed only when the origin is:

- explicitly configured in `allowedOrigins` or `XUVA_ALLOWED_ORIGINS`, or
- local to the machine/LAN runtime, such as `localhost` or `127.0.0.1`.

Unknown origins are rejected before reaching API handlers.

## Path Safety

Handlers that serve cached files validate each user-controlled path segment before joining paths. Segments containing directory separators, empty values, `.` or `..` are rejected. Joined paths must remain under the intended runtime folder before Xuva reads or writes cache files.

## Audit Events

Sensitive activity is emitted to the local event bus using structured audit events.

Example login event:

```json
{
  "type": "audit.auth",
  "data": {
    "userId": "admin",
    "username": "admin",
    "role": "admin",
    "method": "POST",
    "path": "/api/auth/login",
    "action": "login",
    "result": "allowed",
    "createdAt": "2026-04-29T10:00:00Z"
  }
}
```

Example settings event:

```json
{
  "type": "audit.settings",
  "data": {
    "userId": "admin",
    "action": "settings.update",
    "result": "allowed",
    "restartRequired": true
  }
}
```

Example library event:

```json
{
  "type": "audit.library",
  "data": {
    "userId": "admin",
    "action": "library.save",
    "result": "allowed",
    "libraryKind": "movies",
    "storageType": "local"
  }
}
```

Denied streams continue to emit `audit.stream.denied` with the media source, session, route, and denial reason.

## Dependency Scanning

GitHub Actions runs `go test ./...` and `govulncheck ./...` for every pull request and push to `main`. Critical reachable Go vulnerabilities fail the security workflow.

False positives should be handled by linking the advisory, recording reachability, and either upgrading the dependency or documenting why the vulnerable symbol is not reachable before merging.
