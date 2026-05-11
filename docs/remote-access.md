# Remote Access

## Principle

Lorivo should help users configure remote access, but it should not require Lorivo-owned infrastructure.

The setup experience should be honest: Lorivo can test, explain, and generate guidance, but the user controls the route.

## Supported Paths

### LAN

The default experience. Discovery and pairing should work without an internet connection.

### Direct Remote

The user exposes their server through router port forwarding or a reverse proxy.

Lorivo should provide:

- Reachability checks.
- HTTPS checks.
- Port checks.
- Upload speed checks.
- Clear warnings about exposing services.

The UI should prefer a tested HTTPS URL and show whether the route is LAN-only, public internet, reverse proxy, or private network.

### User-Managed Reverse Proxy

Lorivo should document and generate config examples for:

- Caddy.
- Nginx.
- Traefik.

### User-Managed Private Network

Lorivo should work cleanly over:

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

## Product Rules

- LAN pairing must work without internet.
- Manual server URL entry must always exist.
- Remote access diagnostics must say what failed in plain language.
- Lorivo must not imply that unsupported NAT or ISP setups are the user's fault.
- Paid features can improve diagnostics and download preparation, but core local access must not depend on vendor infrastructure.

## Explicit Non-Goals

- No vendor relay in v1.
- No mandatory cloud auth.
- No vendor-hosted media proxy.
- No promise that every NAT or ISP setup can be reached remotely.
