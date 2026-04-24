# Remote Access

## Principle

Vyrden should help users configure remote access, but it should not require Vyrden-owned infrastructure.

## Supported Paths

### LAN

The default experience. Discovery and pairing should work without an internet connection.

### Direct Remote

The user exposes their server through router port forwarding or a reverse proxy.

Vyrden should provide:

- Reachability checks.
- HTTPS checks.
- Port checks.
- Upload speed checks.
- Clear warnings about exposing services.

### User-Managed Reverse Proxy

Vyrden should document and generate config examples for:

- Caddy.
- Nginx.
- Traefik.

### User-Managed Private Network

Vyrden should work cleanly over:

- WireGuard.
- Tailscale.
- ZeroTier.
- Other private network tools.

The product should not require direct integration to be useful. A custom server URL and certificate handling are enough for v1.

### Offline Downloads

Offline downloads are the cleanest remote playback path because no company relay is needed.

Download features should include:

- Resume support.
- Quality presets.
- Audio and subtitle selection.
- Server-side pre-transcode.
- Storage limits.
- Download next unwatched episodes.
- Integrity verification.
- Offline playback without contacting the server.

## Explicit Non-Goals

- No vendor relay in v1.
- No mandatory cloud auth.
- No vendor-hosted media proxy.
- No promise that every NAT or ISP setup can be reached remotely.

