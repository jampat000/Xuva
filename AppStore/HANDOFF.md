# Xuva App Store Handoff

Current phase: DEEP_QA_IN_PROGRESS
Next skill: ejm-03-deep-qa
Next command: continue Deep QA on `/Users/james/Projects/Xuva` only after freeing simulator resources, then rerun validators before TestFlight.

## Identity

- App name: Xuva
- Bundle ID: com.xuva.ios
- Version/build: 0.1.0 (1)
- Device family: universal iPhone/iPad
- Project root: /Users/james/Projects/Xuva
- Starting commit: 13f8161a941e1d626c53755547607af36d0c6bdc
- Branch: fix/apple-tv-focus-and-default-server

## Current Verdict

STOP. Xuva is conceptually shippable as a self-hosted media-server client, but this build is not TestFlight/App Store ready.

## Fresh Evidence From 2026-07-22

- iOS Simulator build passed on iPhoneSimulator 26.5 SDK after source fixes.
- Local Xuva server scanned and probed a synthetic H.264/AAC MP4: `Orbit Proof (2026).mp4`.
- iPhone 17 Pro simulator on iOS 26.5 launched after removing generated empty scene manifest.
- iPad Pro 13-inch simulator on iOS 26.5 launched after the same fix.
- First-run connect UI captured on iPhone and iPad.
- Seeded paired-state home/library UI loaded the scanned/probed movie on iPhone and iPad.
- API playback decision was blocked before the Swift fix when iOS forced adaptive HLS under the server default `original_only` policy.
- API playback decision returned direct-play ready after the Swift fix changed iPhone/iPad playback requests to not prefer adaptive HLS.

## Fixed In This Lane

- Removed generated empty iOS scene manifest setting that made SpringBoard deny launching `com.xuva.ios` on iOS 26.5 simulators.
- Changed iPhone/iPad playback start requests to honor the default original-only server policy and direct play compatible files; tvOS keeps adaptive-preferred behavior.
- Updated Svelte dependencies with `npm audit fix`; subsequent web checks/smokes passed before the simulator resource ceiling.
- Recreated EJM workflow and DeepQA evidence structure.

## Remaining True Blockers

- Full current-build player proof is incomplete: auto-open detail API requests succeeded, but simulator screenshots failed when macOS hit `Resource temporarily unavailable`.
- Latest-runtime physical iPhone/iPad proof is unavailable: attached physical devices were offline in `xcrun xctrace list devices`.
- Deep QA validator remains red because the complete matrix is not proven: dynamic type, accessibility environments/VoiceOver, full interaction coverage, performance budgets, security/privacy evidence depth, release scorecard, and at least 12 fresh app screenshots are not complete.
- Workflow validator remains red until the source/evidence is committed, pushed, and the Deep QA gate passes.
- No TestFlight upload or App Store Connect submission has been attempted.

Read first: AppStore/HANDOFF.md
