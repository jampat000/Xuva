# Plan: Desktop Alpha

## Goal

Ship Lorivo as a user-launched desktop/tray app that supervises the local Go server, opens the web UI, provides native folder picking, and supports restart controls.

## Context

- Lorivo should not default to a system service.
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

## Steps

- [x] Add browser fallback folder browse API.
- [x] Add web bridge contract: `window.lorivoDesktop.pickFolder`.
- [x] Move Libraries into Settings and add Folders/Devices tabs.
- [ ] Choose desktop shell implementation.
- [ ] Implement native folder picker.
- [ ] Implement restart control.
- [ ] Package Windows alpha installer.

## Validation

- Runtime folder save does not visually revert before restart.
- Native picker can choose local, mapped, and UNC paths available to the signed-in user.
- Restart applies runtime folder changes.

## Risks And Rollback

- Desktop shell choice may add weight. Keep the web/server fallback working so packaging can change without breaking core server.

## Decision Log

- 2026-04-30: Desktop install should be taskbar/tray user-mode app, not service-first.
