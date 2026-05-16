# Plan: Desktop Server Shell

## Goal

Ship Xuva as a user-launched desktop/tray shell that operates the local server runtime, opens the web UI, provides native folder picking, and supports restart controls.

## Context

- Xuva should not default to a system service.
- Runtime and library folder selection should use the signed-in user's local, mapped, and NAS/UNC access.
- Web fallback remains necessary for dev/headless/remote-admin use.

## In Scope

- Desktop shell selection.
- Server process supervision.
- Open web UI.
- Native folder picker bridge.
- Restart now control.
- Installer-ready settings defaults.

## Out Of Scope

- Vendor relay.
- Cloud account requirement.
- TV/mobile client packaging.

## Fit-for-Purpose Criteria

- Desktop shell actions must never block active playback-critical server work.
- Restart must be explicit, observable, and safe with process supervision recovery.
- Native folder selection must preserve signed-in user scope (local, mapped, UNC/NAS) without elevation.
- Web fallback behavior must remain valid when no desktop bridge is available.

## Steps

- [x] Add browser fallback folder browse API.
- [x] Add web bridge contract: `window.xuvaDesktop.pickFolder`.
- [x] Move Libraries into Settings and add Folders/Devices tabs.
- [x] Choose desktop shell implementation.
- [x] Implement native folder picker.
- [x] Implement restart control.
- [ ] Package Windows alpha installer.

## Validation

- Runtime folder save does not visually revert before restart.
- Native picker can choose local, mapped, and UNC paths available to the signed-in user.
- Restart applies runtime folder changes.

## Risks And Rollback

- Desktop shell choice may add weight. Keep the web/server fallback working so packaging can change without breaking core server.

## Decision Log

- 2026-04-30: Desktop install should be taskbar/tray user-mode app, not service-first.
- 2026-05-16: Desktop shell selected: Electron (Windows-first) for tray/taskbar UX, native folder picker bridge, and Go process supervision in one runtime.
- 2026-05-16: Added `apps/desktop` Electron scaffold with Go process supervision and bridge handlers for `pickFolder` and `restartServer`; remaining work is packaging polish and production installer workflow.
