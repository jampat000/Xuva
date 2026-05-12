# Vyrden Decommission Checklist

## Status

Recommended status: **Safe to archive but not delete**

Reason:
- Lorivo now covers the core owner settings that were still product-relevant from old Vyrden.
- Local discovery is now implemented in Lorivo with mDNS / Bonjour using the configured Server Name.
- The remaining Vyrden surfaces are either advanced-only owner tools, future platform work, or product directions Lorivo has already rejected.
- It is still premature to delete Vyrden outright until the remaining advanced/future items are tracked explicitly and one final pass confirms nothing required still lives only in Vyrden.

## Already restored in Lorivo

- [x] Server name / identity
  - `1a2e468` / `d053e1c` Add configurable Lorivo server name
- [x] Library management
  - `b4ee043` Restore Lorivo library playback and access settings
- [x] Playback policy editing
  - `b4ee043` Restore Lorivo library playback and access settings
- [x] Access basics
  - `b4ee043` Restore Lorivo library playback and access settings
- [x] Scanning automation
  - `8c2f2c7` Restore Lorivo scanning automation settings
- [x] Settings visibility and local dev parity
  - `8038e96` Fix Lorivo settings visibility in dev
- [x] Development auth bypass for owner-only settings work
  - `443d351` Add development auth bypass for Lorivo settings
- [x] Restored settings polish
  - `1429303` Polish Lorivo restored settings
- [x] Settings navigation / settings mode reconciliation
  - `1cac500` Reconcile Lorivo settings navigation cleanup
- [x] Storage settings
  - `cf569f3` Restore Lorivo storage settings
- [x] Metadata source settings
  - `180d720` Restore Lorivo metadata source settings
- [x] Metadata review
  - `765af27` Restore Lorivo metadata review settings
- [x] Version group summary
  - `765af27` Restore Lorivo metadata review settings
- [x] Pairing review
  - `17f97bd` Restore Lorivo pairing review settings
- [x] Local network discovery using Lorivo Server Name
  - `5ab69a0` Add Lorivo local network discovery
  - Implemented as mDNS / Bonjour over `_lorivo._tcp.local.`

## Backlogged / deliberately deferred

### Advanced-only

- [ ] Hardware acceleration settings and hardware test
- [ ] GPU worker count
- [ ] Source compatibility / Source Inspector tools
- [ ] Remote diagnostics
- [ ] Advanced runtime controls, if Lorivo still needs them after product review

### Backlog / future

- [ ] Persistent device registry
- [ ] SSDP / UPnP / DLNA discovery, if Lorivo later needs more than mDNS / Bonjour
- [ ] Source compatibility tools redesigned for Lorivo
- [ ] Advanced hardware owner tools
- [ ] Optional diagnostics page

## Rejected permanently

- [x] Admin / Operator wording
- [x] Raw config grid
- [x] Raw performance JSON
- [x] Raw Decision
- [x] Provider API key fields in normal Settings
- [x] Migration import / rollback in normal Settings
- [x] Fake discovery controls
- [x] Preview / demo content
- [x] Vyrden branding
- [x] FFmpeg / FFprobe / worker / allowed-origin controls in normal Settings

## Remaining Vyrden items by decision

### Keep only if Lorivo later wants an Advanced area

- Hardware acceleration status, tuning, and test flow
- Source compatibility / Source Inspector tools
- Remote diagnostics
- Advanced runtime controls

### Keep as tracked future platform work

- Persistent device registry
- SSDP / UPnP / DLNA discovery expansion, if product needs it

### Do not restore

- Any Settings surface that reads like admin tooling or operator telemetry
- Any raw backend or configuration dump
- Any API-key-based normal-user metadata setup
- Any fake device or discovery UX

## Final deletion criteria

- [x] all core owner settings restored
- [x] advanced/operator items either restored, rejected, or tracked as issues
- [x] LAN discovery implemented and device registry tracked as future work
- [ ] no required code/assets/docs remain only in Vyrden
- [x] final Lorivo validation passed
- [ ] optional final archive/tag created
- [ ] Vyrden repo archived/deleted

## Recommended next step

Archive Vyrden first. Do not delete it yet.

That is the safer sequence because Lorivo has reached practical owner-settings parity, but the remaining advanced and future items should be closed out intentionally rather than by losing the old reference repo too early.

See also: [Advanced owner tools audit](advanced-owner-tools-audit.md)

## Optional issue drafts

### 1. Add persistent device registry

**Title**  
Add persistent device registry for approved clients

**Problem**  
Current pairing review is real, but pairing requests are runtime-only and there is no durable registry of approved devices.

**Scope**
- persist approved device records
- define owner-visible device list behavior
- define revoke / forget device flow only if backend is real
- keep pairing separate from discovery

**Acceptance**
- approved devices survive restart
- Lorivo can show a truthful owner-facing device list
- no fake connected devices UI before persistence exists

### 2. Extend local discovery beyond mDNS / Bonjour

**Title**  
Decide whether Lorivo needs SSDP / UPnP / DLNA discovery

**Problem**  
Lorivo now advertises itself over mDNS / Bonjour, but does not implement SSDP, UPnP, or DLNA-style discovery.

**Scope**
- decide whether non-Bonjour discovery protocols are needed at all
- define protocol scope if broader client support is required
- keep settings and docs honest about what is and is not discoverable
- do not add fake discovery UI before backend support exists

**Acceptance**
- clear keep/defer/reject decision for SSDP / UPnP / DLNA
- if kept, tracked as future platform work instead of implied support

### 3. Design Advanced hardware owner tools

**Title**  
Decide whether Lorivo needs advanced hardware owner tools

**Problem**  
Old Vyrden exposed hardware acceleration status, GPU worker controls, and hardware tests. Lorivo has intentionally kept those out of normal Settings.

**Scope**
- decide whether Lorivo should surface hardware tools at all
- classify hardware status, testing, and worker controls
- keep normal Settings owner-friendly
- only consider a future Advanced area if the product still needs one

**Acceptance**
- documented product decision on hardware owner tooling
- clear keep/defer/reject decision for GPU worker controls and hardware test flow

### 4. Redesign Source Compatibility tools

**Title**  
Redesign Source Compatibility tools for Lorivo

**Problem**  
Old Vyrden Source Inspector was useful but too diagnostic and internal for Lorivo's current product direction.

**Scope**
- decide which compatibility facts are owner-useful
- remove backend/internal framing
- avoid reviving raw decision tooling

**Acceptance**
- clear Lorivo-oriented compatibility UX proposal
- no dependency on old Vyrden operator language

### 5. Decide Remote Diagnostics future

**Title**  
Decide whether Remote Diagnostics belongs in Lorivo

**Problem**  
Old Vyrden had remote diagnostics tooling, but it is not part of current Lorivo owner settings.

**Scope**
- decide whether remote diagnostics is a future product need
- determine whether it belongs in docs, advanced tools, or nowhere
- avoid normal-settings clutter

**Acceptance**
- explicit keep/defer/reject decision
- if kept, tracked as an advanced feature rather than normal Settings

## Decommission summary

Lorivo has restored the core owner-settings surface that justified keeping old Vyrden around as a live reference:

- identity
- libraries
- scanning
- playback
- storage
- metadata sources
- metadata review
- access basics
- pairing review
- local network discovery via mDNS / Bonjour

What remains from Vyrden is not a blocking parity gap for normal owner settings. It is either:

- advanced tooling,
- future platform capability,
- or product direction Lorivo has already rejected.

That means Vyrden is **safe to archive now**, but **not yet the right candidate for permanent deletion** until the remaining advanced/future items are explicitly tracked or closed out.
