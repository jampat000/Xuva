# Advanced Owner Tools Audit

## Purpose

This audit covers the remaining Vyrden-era advanced and future surfaces that are still relevant to Lorivo decommission planning after the core owner settings and local discovery work landed.

This is not an implementation plan. It is a product and engineering decision record for the remaining high-risk items.

## Classification summary

| Item | Classification | Target Lorivo location |
| --- | --- | --- |
| Advanced hardware owner tools | advanced-only | Future `Advanced` area, if Lorivo adds one |
| Source compatibility / Source Inspector redesign | implement later | Playback detail pages, not normal Settings |
| Remote diagnostics | reject permanently | None in normal Lorivo product UI |
| Persistent device registry | restored | `Settings > Access` |
| SSDP / UPnP / DLNA discovery | implement later | Future discovery layer, no new Settings surface yet |

## 1. Advanced hardware owner tools

- **Classification:** advanced-only
- **Old Vyrden evidence:**
  - `server/internal/webapp/static/app.js` had a `Hardware` tab in Settings.
  - Old UI exposed `GPU worker slots`, `Save hardware settings`, and `Test hardware acceleration`.
  - Old backend routes included `POST /api/settings/hardware/test`.
- **Current Lorivo backend support:**
  - `server/internal/api/router.go` still exposes `GET /api/settings/performance` and `POST /api/settings/hardware/test`.
  - `server/internal/config/config.go` still carries `GPUWorkers` and `HardwareUnlocked`.
  - Playback decisions and transcode jobs still understand hardware acceleration paths.
- **Current Lorivo frontend support:**
  - No dedicated hardware settings UI.
  - `Settings > Playback` shows only a small compatibility-support status derived from performance data.
- **User value:**
  - Moderate for power users who run heavier video conversion routes.
  - Low for normal owners who only need library, playback, storage, metadata, and access settings.
- **Risk:**
  - Medium to high.
  - Easy to overexpose FFmpeg/runtime details and regress the normal owner-facing Settings surface.
- **Recommendation:**
  - Keep backend support.
  - Do not restore this to normal Settings.
  - Only revisit inside a clearly-labeled future `Advanced` area after a product decision on whether Lorivo wants advanced owner tools at all.

## 2. Source compatibility / Source Inspector redesign

- **Classification:** implement later
- **Old Vyrden evidence:**
  - `server/internal/webapp/static/app.js` had a `Source Inspector` route and related playback drawer.
  - Old UI exposed `Inspect`, `Probe`, `Repackage while playing`, `Convert while playing`, `Create remote optimized version`, and `Raw Decision`.
  - Old route copy explicitly called it an internal safety net.
- **Current Lorivo backend support:**
  - Playback decision, probe, download, transcode, and device-profile routes still exist in `server/internal/api/router.go`.
  - Device capability data still exists in `server/internal/devices/service.go`.
  - Playback decision logic still exists in the server and media detail flows.
- **Current Lorivo frontend support:**
  - Lorivo already surfaces compatibility guidance in media detail flows and playback policy.
  - There is no dedicated Source Inspector page or raw decision UI in Lorivo Settings.
- **User value:**
  - High if narrowed to plain-language playback compatibility help.
  - Low if it returns as a diagnostic surface full of route-level internals.
- **Risk:**
  - High.
  - The old Vyrden tooling was useful, but it carried raw/technical framing that conflicts with Lorivo's current product direction.
- **Target Lorivo location:**
  - Playback detail pages, not Settings.
- **Recommendation:**
  - Redesign later as a Lorivo compatibility surface.
  - Do not restore `Source Inspector` or `Raw Decision`.
  - Treat this as playback UX work, not decommission-critical settings parity.

## 3. Remote diagnostics

- **Classification:** reject permanently
- **Old Vyrden evidence:**
  - `server/internal/webapp/static/app.js` had a `Remote Access` route.
  - Old UI exposed WAN detection and remote diagnostics.
  - Old backend routes included `GET /api/remote/access`, `POST /api/remote/diagnostics`, and `POST /api/remote/wan`.
- **Current Lorivo backend support:**
  - Those routes still exist in `server/internal/api/router.go`.
- **Current Lorivo frontend support:**
  - No remote diagnostics UI in Lorivo.
- **User value:**
  - Unclear.
  - Lorivo's current product direction has been local-first and avoids normal-settings clutter around networking internals.
- **Risk:**
  - High.
  - Easy to overclaim support, create confusing failure states, or imply a broader remote-access product that Lorivo has not committed to.
- **Target Lorivo location:**
  - None in normal product UI.
- **Recommendation:**
  - Do not restore remote diagnostics as a normal owner feature.
  - Keep this out of the decommission parity target.
  - If the backend routes remain for internal use or later cleanup, that does not justify a Lorivo UI surface.

## 4. Persistent device registry

- **Classification:** restored
- **Old Vyrden evidence:**
  - Old `Devices` tab showed pairing requests with `Approve` and `Deny`.
  - Approved requests received a generated `deviceId`, but the old flow still behaved like a runtime request queue rather than a durable device inventory.
- **Current Lorivo backend support:**
  - Pairing requests still remain runtime-only in `server/internal/pairing/service.go`.
  - Approved devices now persist in Lorivo's SQLite database and survive restart.
  - Lorivo now has a real revoke/remove flow for approved devices.
- **Current Lorivo frontend support:**
  - `Settings > Access` shows both pairing review and approved devices.
  - The UI stays honest about current limitations and does not claim live online/offline presence.
- **User value:**
  - High once pairing expands beyond a temporary review queue.
- **Risk:**
  - Medium.
  - The remaining risk is around future paired-client authentication and live presence, not basic registry persistence.
- **Target Lorivo location:**
  - `Settings > Access`
- **Recommendation:**
  - Keep the current minimal scope.
  - Do not expand it into fake connected-device or live presence UI without real session-backed tracking.

## 5. SSDP / UPnP / DLNA discovery

- **Classification:** implement later
- **Old Vyrden evidence:**
  - No prior Vyrden LAN discovery stack existed to restore here.
- **Current Lorivo backend support:**
  - `server/internal/discovery/service.go` now implements mDNS / Bonjour only.
  - `GET /api/discovery/status` exposes honest read-only status.
- **Current Lorivo frontend support:**
  - `Settings > About` shows local discovery status and uses the configured Server Name.
- **User value:**
  - Medium if future client targets need broader discovery beyond Bonjour/mDNS.
  - Low if current client strategy remains centered on Apple and other mDNS-capable clients.
- **Risk:**
  - Medium.
  - Broader discovery protocols create more surface area, more platform nuance, and more chances to overclaim DLNA-style behavior.
- **Target Lorivo location:**
  - No new Settings area yet.
  - If ever implemented, it should extend discovery status in `About`, not create a fake device section.
- **Recommendation:**
  - Implement later only if target clients require it.
  - Current mDNS / Bonjour support is enough to remove LAN discovery as a Vyrden decommission blocker.

## Overall recommendation

The remaining decommission blockers are now mostly classification and backlog management, not missing core owner settings.

- `Remote diagnostics` can be treated as rejected for normal Lorivo UI.
- `Advanced hardware owner tools` should stay backend-capable but out of normal Settings unless Lorivo deliberately adds an `Advanced` area.
- `Source compatibility` should come back only as a redesigned playback help surface.
- Optional broader discovery protocols and any future paired-client authentication work are the remaining device-platform items with clear product value.

That leaves Vyrden in the same overall state as the main checklist:

- safe to archive
- not yet the right candidate for permanent deletion
