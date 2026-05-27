# Local Discovery

Xuva can advertise itself on the local network with mDNS / Bonjour so nearby client apps can find a server instance and connect to its network address.

## What It Does

- Advertises a custom service on the local network: `_xuva._tcp.local.`
- Uses the configured **Xuva Server Name** as the advertised instance name
- Advertises the current HTTP port
- Prefers real LAN adapters (Ethernet/Wi-Fi) and avoids virtual adapters (WSL/Docker/Hyper-V/Tailscale) unless no physical adapter is available.
- Publishes only safe TXT records:
  - `app=xuva`
  - `api=/api/client/bootstrap`
  - `hostName=<reachable connection host>`
  - `serverName=<configured name>`
  - `web=<canonical web origin, or a derived LAN IP URL>`

Xuva does not advertise secrets, tokens, credentials, filesystem paths, or runtime folder details.

## Server Name Versus Network Name

Xuva has two separate names:

- **Xuva Server Name** is the friendly instance/display name. It appears in browser titles, setup screens, client discovery lists, and `/api/client/bootstrap`.
- **Connection host / canonical web address** is the routable address clients use to connect. It comes from an explicitly configured canonical web origin or a derived LAN IP URL. It is not inferred from the Windows/Linux machine name for native clients.

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

## Canonical Address And Discovery URL

If a canonical web address is configured, discovery advertises that address in the `web=` TXT record.

If it is blank, discovery now prefers a reachable advertised LAN IP such as `http://10.1.1.103:8097` instead of a bare Windows hostname. This avoids Apple/TV clients failing when they cannot resolve `DESKTOP-...` style hostnames.

When Xuva listens on `0.0.0.0`, discovery advertises IPv4 LAN addresses only. This avoids IPv6-first client failures on networks where IPv6 routing is incomplete.

The mDNS SRV target also uses a Xuva-branded `.local.` host label derived from the configured server name, for example `family-room-xuva.local.`, rather than the operating-system hostname. This prevents native clients from displaying development or bare-metal machine names as the server URL.

`GET /api/discovery/status` exposes the active discovery interfaces and advertised IPs so installer and LAN diagnostics can confirm which adapters are being used.

## Current Protocol

- Implemented now: mDNS / Bonjour
- Not implemented in this pass:
  - SSDP
  - UPnP
  - DLNA service discovery

## Windows Firewall

Windows installs must allow both the HTTP server and the discovery responder on trusted LAN networks:

- TCP `8097` inbound to `xuva-server.exe` for web/API access from phones, TVs, and browsers on the LAN
- UDP `5353` inbound to `xuva-server.exe` for Bonjour/mDNS discovery

The Windows installer provisions these rules for Private and Domain profiles only. Xuva should not open Public network profiles automatically. If a machine is accidentally marked as Public in Windows network settings, discovery may fail until the network is changed to Private or equivalent firewall rules are added by the operator.

## Current Product Boundary

Local discovery, device pairing, and the approved-device registry are separate.

- Discovery helps clients find the Xuva server on the home network.
- Pairing requires owner approval.
- Approved devices are stored persistently after approval and survive restart.
- Xuva does not claim live online/offline presence or connected-device polling.
