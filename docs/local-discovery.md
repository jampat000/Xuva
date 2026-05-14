# Local Discovery

Xuva can advertise itself on the local network with mDNS / Bonjour so nearby client apps can find the server by name.

## What It Does

- Advertises a custom service on the local network: `_xuva._tcp.local.`
- Uses the configured **Server Name** as the advertised instance name
- Advertises the current HTTP port
- Publishes only safe TXT records:
  - `app=xuva`
  - `api=/api/client/bootstrap`
  - `serverName=<configured name>`

Xuva does not advertise secrets, tokens, credentials, filesystem paths, or runtime folder details.

## How Server Name Is Used

The saved **Server Name** is used for:

- the browser title
- `/api/client/bootstrap`
- local discovery instance naming

If you change the Server Name, restart Xuva so discovery can advertise the updated name.

## Enable Or Disable Discovery

Environment variable:

```powershell
XUVA_DISCOVERY_ENABLED=true
```

Default behavior:

- discovery is enabled by default
- Xuva only advertises when the server is listening on a LAN-safe address
- if Xuva is bound only to loopback such as `127.0.0.1` or `localhost`, discovery does not start

Optional service type override:

```powershell
XUVA_DISCOVERY_SERVICE_TYPE=_xuva._tcp
```

Normal installs should keep the default service type.

## Current Protocol

- Implemented now: mDNS / Bonjour
- Not implemented in this pass:
  - SSDP
  - UPnP
  - DLNA service discovery

## Current Product Boundary

Local discovery, device pairing, and the approved-device registry are separate.

- Discovery helps clients find the Xuva server on the home network.
- Pairing requires owner approval.
- Approved devices are stored persistently after approval and survive restart.
- Xuva does not claim live online/offline presence or connected-device polling.
