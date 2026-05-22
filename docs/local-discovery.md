# Local Discovery

Xuva can advertise itself on the local network with mDNS / Bonjour so nearby client apps can find a server instance and connect to its network address.

## What It Does

- Advertises a custom service on the local network: `_xuva._tcp.local.`
- Uses the configured **Xuva Server Name** as the advertised instance name
- Advertises the current HTTP port
- Publishes only safe TXT records:
  - `app=xuva`
  - `api=/api/client/bootstrap`
  - `hostName=<operating-system hostname>`
  - `serverName=<configured name>`
  - `web=<canonical or derived web origin>`

Xuva does not advertise secrets, tokens, credentials, filesystem paths, or runtime folder details.

## Server Name Versus Network Name

Xuva has two separate names:

- **Xuva Server Name** is the friendly instance/display name. It appears in browser titles, setup screens, client discovery lists, and `/api/client/bootstrap`.
- **Network host name / canonical web address** is the routable address clients use to connect. It comes from the operating system, local DNS, mDNS, a reverse proxy, Docker/container networking, or the configured canonical web origin.

The saved **Xuva Server Name** is used for:

- the browser title
- `/api/client/bootstrap`
- local discovery instance naming

It is not treated as a DNS name and does not need to resolve on the network.

If you change the Xuva Server Name, restart Xuva so discovery can advertise the updated instance name.

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
