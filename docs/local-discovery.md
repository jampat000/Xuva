# Local Discovery

Lorivo can advertise itself on the local network with mDNS / Bonjour so nearby client apps can find the server by name.

## What It Does

- Advertises a custom service on the local network: `_lorivo._tcp.local.`
- Uses the configured **Server Name** as the advertised instance name
- Advertises the current HTTP port
- Publishes only safe TXT records:
  - `app=lorivo`
  - `api=/api/client/bootstrap`
  - `serverName=<configured name>`

Lorivo does not advertise secrets, tokens, credentials, filesystem paths, or runtime folder details.

## How Server Name Is Used

The saved **Server Name** is used for:

- the browser title
- `/api/client/bootstrap`
- local discovery instance naming

If you change the Server Name, restart Lorivo so discovery can advertise the updated name.

## Enable Or Disable Discovery

Environment variable:

```powershell
LORIVO_DISCOVERY_ENABLED=true
```

Default behavior:

- discovery is enabled by default
- Lorivo only advertises when the server is listening on a LAN-safe address
- if Lorivo is bound only to loopback such as `127.0.0.1` or `localhost`, discovery does not start

Optional service type override:

```powershell
LORIVO_DISCOVERY_SERVICE_TYPE=_lorivo._tcp
```

Normal installs should keep the default service type.

## Current Protocol

- Implemented now: mDNS / Bonjour
- Not implemented in this pass:
  - SSDP
  - UPnP
  - DLNA service discovery
  - persistent device registry

## Current Product Boundary

Local discovery and device pairing are separate.

- Discovery helps clients find the Lorivo server on the home network.
- Pairing still requires owner approval.
- Lorivo does not yet expose a persistent connected-device registry.
