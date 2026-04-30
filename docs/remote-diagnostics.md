# Remote Diagnostics

Vyrden helps users diagnose remote access without operating a vendor relay. The product checks the route the user owns, classifies the likely failure, and returns concrete next actions.

## Result Schema

`POST /api/remote/diagnostics`

Input:

```json
{
  "publicUrl": "https://media.example.com:8097",
  "requiredMbps": 25,
  "measuredMbps": 12
}
```

Output:

```json
{
  "status": "needs_action",
  "failureClass": "nat_firewall",
  "route": "public",
  "summary": "Vyrden could not open a TCP connection to the remote address.",
  "target": { "scheme": "https", "host": "media.example.com", "port": 8097 },
  "checks": [],
  "nextActions": [],
  "privacy": []
}
```

## Failure Taxonomy

| Class | Meaning | User Actions |
| --- | --- | --- |
| `ready` | DNS, port, TLS, and optional throughput checks passed. | Pair the remote client, keep HTTPS enabled, lower remote quality if buffering appears. |
| `not_configured` | No public URL, VPN name, reverse proxy, or forwarded route was provided. | Add the route the user expects remote clients to use. |
| `private_route` | The route is LAN/VPN/mesh only. | Use it for private networks, or configure public DNS/reverse proxy/port forward for internet access. |
| `dns` | Hostname does not resolve. | Fix A/AAAA/CNAME records or dynamic DNS. |
| `nat_firewall` | DNS resolves but TCP cannot connect. | Check port forwarding, router firewall, OS firewall, and ISP CGNAT. |
| `certificate` | TCP connects but HTTPS certificate validation fails. | Install or renew a valid cert for the exact hostname. |
| `throughput` | Route is reachable but measured bandwidth is below the selected target. | Lower remote quality, use adaptive streaming, or create an optimized version. |
| `invalid_input` | URL includes unsafe/sensitive or unsupported parts. | Enter only a base `http` or `https` URL with optional port. |

## Privacy Rules

- Diagnostics return only scheme, host, and port.
- URL paths, query strings, fragments, usernames, passwords, and tokens are rejected.
- No remote diagnostic request uses Vyrden-hosted relay infrastructure.
- WAN IP lookup remains a separate explicit action.
- Logs should use the failure class and check code, not full user-supplied URLs.

## Operator Examples

DNS failure:

1. User enters `https://media.example.com`.
2. Vyrden returns `dns`.
3. User creates or repairs the DNS record, waits for propagation, then runs diagnostics again.

NAT/firewall failure:

1. DNS resolves, but TCP connection fails.
2. Vyrden returns `nat_firewall`.
3. User forwards the public port to the reverse proxy or Vyrden host, opens the OS firewall, or switches to VPN/mesh if CGNAT blocks inbound access.

Certificate failure:

1. TCP connects, TLS fails.
2. Vyrden returns `certificate`.
3. User fixes certificate renewal, hostname mismatch, or proxy TLS configuration.

Throughput failure:

1. Route is reachable but measured bandwidth is lower than required.
2. Vyrden returns `throughput`.
3. User lowers remote quality, uses adaptive streaming, or creates an optimized version.
