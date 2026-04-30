# Plan: Apple TV Alpha

## Goal

Build the first native playback client path: pair Apple TV to the local server, browse TV-shaped rows, open detail screens, and play direct/HLS routes with session heartbeat.

## Context

- Apple TV is the lead native client because it can be tested at home.
- Server remains local-first with no vendor relay.
- SwiftUI source lives in `apps/apple-tv/Sources`.
- Contract details live in `docs/apple-tv-alpha.md`.

## In Scope

- Local server bootstrap.
- Local pairing code flow.
- TV home rows.
- Movie/series detail contracts.
- AVPlayer direct stream and HLS playback.
- Session heartbeat, resume, watched state.

## Out Of Scope

- App Store purchase flows.
- Vendor cloud account.
- Offline downloads.
- Android TV parity.

## Steps

- [x] Add client bootstrap contract.
- [x] Add local pairing request/approval flow.
- [x] Add TV home contract.
- [x] Add SwiftUI starter shell and pairing flow.
- [ ] Import Swift source into Xcode tvOS target.
- [ ] Fix compile/signing issues on Mac.
- [ ] Add movie/series detail client contract optimized for TV.
- [ ] Add playback-start contract for direct/HLS route selection.
- [ ] Add AVPlayer playback shell.
- [ ] Add session heartbeat and resume from tvOS.

## Validation

- `go test ./internal/api`
- `go test ./internal/pairing ./internal/api`
- `go test ./...`
- Xcode build on Apple TV hardware once Mac is available.

## Risks And Rollback

- Pairing currently returns a device ID but no durable credential. Rollback is to keep pairing disabled from bootstrap and use browser auth until credential design lands.
- Swift source has not compiled on Xcode yet. Keep changes isolated in `apps/apple-tv/Sources` until first Mac validation.

## Decision Log

- 2026-04-30: Apple TV chosen as lead native alpha client.
- 2026-04-30: Local pairing code chosen over cloud account or vendor relay.
