# Xuva Native Client API

This document defines the current server contract for native clients (tvOS, Android TV, mobile).

Source-of-truth note:
- Xuva is the active implementation source.

Scope:
- local-server discovery
- bootstrap and pairing
- library/catalog browsing
- artwork and media playback routes
- session and playback progress updates

Out of scope:
- DLNA / SSDP / UPnP
- cloud relay
- fake device presence or fake online/offline state

## 1. Discovery

Current discovery protocol:
- mDNS / Bonjour
- service type: `_xuva._tcp.local.`
- service instance name: configured Xuva `Server Name` display value
- service port: active HTTP port

Discovery TXT records:
- `app=xuva`
- `api=/api/client/bootstrap`
- `serverName=<configured Xuva display name>`
- `hostName=<reachable connection host>`
- `web=<canonical or derived web origin>`

Naming rule:
- `serverName` is a friendly Xuva instance name for humans.
- `web`, `hostName`, the resolved mDNS target, manual URL, or `server.webUrl` are the connectivity authorities.
- Clients must not assume `serverName` is DNS-resolvable.
- Clients should prefer the `web` TXT record when present. The mDNS SRV target is a Xuva-branded `.local.` host label, not the operating-system hostname.

Manual fallback:
- clients can always use manual URL entry and call `GET /api/client/bootstrap`.

Not implemented:
- SSDP
- UPnP
- DLNA discovery

## 2. Bootstrap

Endpoint:
- `GET /api/client/bootstrap`

Purpose:
- safe server identity and capability handshake before sign-in or pairing.

Important fields:
- `server.name` / `server.displayName`
- `server.hostName`
- `server.baseUrl`
- `server.webUrl`
- `server.httpAddr`
- `auth.required`
- `auth.bootstrapAllowed`
- `features` flags
- `profiles` capability list
- `endpoints` route templates

Safe example response shape:

```json
{
  "server": {
    "name": "Xuva",
    "displayName": "Xuva",
    "hostName": "media-server.local",
    "baseUrl": "http://127.0.0.1:8097",
    "webUrl": "http://media-server.local:8097",
    "httpAddr": "127.0.0.1:8097"
  },
  "auth": {
    "required": true,
    "bootstrapAllowed": false
  },
  "features": {
    "directPlay": true,
    "hlsAdaptive": true,
    "resume": true,
    "watchedState": true,
    "vendorRelay": false
  },
  "endpoints": {
    "pairingCreate": "/api/pairing/requests",
    "pairingStatus": "/api/pairing/requests/{id}",
    "clientHome": "/api/client/home",
    "playbackDecision": "/api/playback/decision",
    "playbackRoute": "/api/playback/route"
  }
}
```

## 3. Pairing

Pairing create:
- `POST /api/pairing/requests`
- payload:
  - `deviceName`
  - `clientProfile`
  - `deviceId` (recommended stable client identifier)

Pairing status poll:
- `GET /api/pairing/requests/{id}`

Owner review and action:
- `GET /api/pairing/requests` (owner)
- `POST /api/pairing/requests/{id}/approve` (owner)
- `POST /api/pairing/requests/{id}/deny` (owner)

Current behavior:
- request receives a short-lived local pairing code.
- owner approves or denies.
- approved requests can become persisted approved-device entries.
- approved requests receive a native header auth credential in the approval response and in subsequent status polling by request id.

Approved credential shape:

```json
{
  "auth": {
    "method": "header_token",
    "sessionToken": "<opaque-token>",
    "expiresAt": "2026-05-22T01:00:00Z"
  }
}
```

Native clients send the token on protected routes as `X-Auth-Token: <opaque-token>` or `Authorization: Bearer <opaque-token>`.

## 4. Approved Devices

Owner APIs:
- `GET /api/devices`
- `POST /api/devices/{id}/revoke`

What is persisted:
- device identity metadata (name/profile/device id mapping)
- approval/revocation status
- approval timestamps and actor

What is not claimed:
- live online/offline state
- synthetic last-seen values

## 5. Auth and Session Model

Modes:
- normal auth mode:
  - owner/user session required for protected routes
  - browser mutations require CSRF token

Route classes:
- public bootstrap routes:
  - `GET /api/client/bootstrap`
  - `POST /api/pairing/requests`
  - `GET /api/pairing/requests/{id}`
- protected routes:
  - playback start/session mutation routes
  - owner settings mutation routes
  - owner pairing approval and approved-device management

## 6. Library and Catalog APIs

Home and browse:
- `GET /api/client/home` (protected)
- `GET /api/movies` (public read)
- `GET /api/series` (public read)
- `GET /api/movies/{id}` (public read)
- `GET /api/series/{id}` (public read)
- `GET /api/playback/recent` (public read)

Artwork:
- `GET /api/artwork/{kind}/{id}` (public read)
- if source artwork is unavailable, server returns safe fallback artwork.

Not-found behavior:
- missing movie/series detail returns `404`.

## 7. Playback APIs

Decision and routing:
- `GET /api/playback/decision`
- `GET /api/playback/route`

Session lifecycle:
- `POST /api/sessions` (protected)
- `PATCH /api/sessions/{id}` (protected)
- `DELETE /api/sessions/{id}` (protected)
- `PUT /api/playback/state/{id}` (protected)

Adaptive/direct stream support:
- `GET /api/media-sources/{id}/stream`
- `GET /api/media-sources/{id}/adaptive/master.m3u8`
- `GET /api/media-sources/{id}/adaptive/{variant}`
- `POST /api/media-sources/{id}/adaptive/session` (protected)
- `POST /api/media-sources/{id}/stream-token` (protected)

Current limitation:
- native clients must persist the pairing `auth.sessionToken` and refresh stored tokens when the server returns a replacement `X-Auth-Token` header.

## 8. Images and Asset Delivery

Artwork serving:
- uses metadata sources and local cache under server metadata storage.
- endpoint returns image bytes or safe fallback artwork.

Client caching expectations:
- clients should cache artwork URLs/content on their side.
- artwork may refresh as metadata updates run.

## 9. Error Model

Common status codes:
- `200` success
- `201` resource created
- `400` invalid payload/request
- `401` authentication required
- `403` authenticated but not allowed (or CSRF failure on browser mutation)
- `404` resource not found
- `409` request conflicts with current state
- `500` server error
- `503` capability unavailable

## 10. API Stability and Versioning

Current contract:
- bootstrap response is the runtime capability source of truth.
- route policy defines protected vs public expectations.

Recommendation:
- keep `GET /api/client/bootstrap` as the compatibility handshake.
- if explicit API versioning is added later, expose it there first to avoid parallel capability probes.

## Client Readiness Matrix

| Client need | Current support | Endpoint(s) | Ready for tvOS/Android | Blocker | Priority |
| --- | --- | --- | --- | --- | --- |
| Discover server on LAN | mDNS/Bonjour implemented | mDNS `_xuva._tcp.local.`, `GET /api/discovery/status` | yes | none for Bonjour-capable clients | P0 |
| Manual server URL entry | supported | `GET /api/client/bootstrap` | yes | none | P0 |
| Create pairing request | supported | `POST /api/pairing/requests` | yes | none | P0 |
| Poll pairing approval status | supported | `GET /api/pairing/requests/{id}` | yes | none | P0 |
| Owner approve/deny pairing | supported | `GET /api/pairing/requests`, `POST /approve`, `POST /deny` | yes | none | P0 |
| Persist approved device list | supported | `GET /api/devices`, `POST /api/devices/{id}/revoke` | yes | no public client read by design | P0 |
| Browse home feed | protected and functional | `GET /api/client/home` | yes | none after pairing credential is stored | P0 |
| Browse movies list | supported | `GET /api/movies` | yes | none | P0 |
| Browse TV list | supported | `GET /api/series` | yes | none | P0 |
| Load movie detail | supported | `GET /api/movies/{id}`, `GET /api/client/movies/{id}` | partial | choose one stable client path set | P1 |
| Load TV detail | supported | `GET /api/series/{id}`, `GET /api/client/series/{id}` | partial | choose one stable client path set | P1 |
| Load artwork/posters/backdrops | supported with fallback | `GET /api/artwork/{kind}/{id}` | yes | none | P0 |
| Start playback route | route APIs exist | `GET /api/playback/decision`, `GET /api/playback/route`, `POST /api/sessions` | yes | none after pairing credential is stored | P0 |
| Report playback progress/resume | supported | `PATCH /api/sessions/{id}`, `PUT /api/playback/state/{id}` | yes | none after pairing credential is stored | P0 |
| Watched/unwatched updates | supported via playback state/session | `PUT /api/playback/state/{id}` | yes | none after pairing credential is stored | P0 |
| Select version/source | basic support exists | detail/version payloads + playback route selection | partial | dedicated client-facing version selection contract needs tightening | P1 |
| Subtitle/audio track handling | APIs exist | `GET /api/media-sources/{id}/tracks`, subtitle routes | partial | needs client integration guidance and auth path clarity | P1 |
| Revoke paired device | owner support exists | `POST /api/devices/{id}/revoke` | yes | none | P1 |
| Discover server via SSDP/UPnP/DLNA | not implemented | n/a | no | protocol not implemented | P2 |

## Native Client Blockers Before Full tvOS/Android Build-out

1. Finalize one canonical detail endpoint family for native clients (`/api/client/...` vs mixed public catalog routes).
2. Add explicit native session refresh/revocation examples beyond the current auth-session token rotation header.
3. Decide whether approved-device revocation should also revoke any active native auth sessions for that device.
