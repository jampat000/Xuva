# Deep QA Evidence

- **Build tested:** Xuva 0.1.0 (1), iOS simulator build from 2026-07-22 after local fixes
- **Date:** 2026-07-22
- **Devices:** iPhone 17 Pro simulator iOS 26.5; iPad Pro 13-inch simulator iOS 26.5
- **Verdict:** STOP
- **Gate proof:** FAIL
- **Perfection gate:** FAIL
- **Project identity matched:** YES
- **Apple-visible identity matched:** YES
- **Simulator target matched:** YES
- **Tested project root:** /Users/james/Projects/Xuva
- **Tested app name:** Xuva
- **Tested bundle id:** com.xuva.ios
- **Tested source commit:** 13f8161a941e1d626c53755547607af36d0c6bdc plus uncommitted local fixes
- **Open Blockers:** 4
- **Open Majors:** 3
- **Known unresolved issues:** Full player proof, physical-device proof, complete accessibility/performance/security/product-depth matrix
- **All noted issues fixed or escalated:** NO
- **Bot playthrough minimum met:** NO
- **Human-like playthrough minimum met:** NO
- **Onboarding/help passed:** PARTIAL
- **Apple device/runtime matrix passed:** NO
- **Pixel/geometry visual scan completed:** PARTIAL
- **Truncation/cropping/wrapping passed:** PARTIAL
- **Visual screenshot review completed:** PARTIAL
- **Fresh screenshot set captured for this QA run:** PARTIAL
- **Fresh screenshot set inspected for this QA run:** PARTIAL
- **Dynamic Type matrix passed:** NO
- **Dynamic Type low/default/large/accessibility covered:** NO
- **Control/component consistency passed:** PARTIAL
- **Visual system precision passed:** NO
- **Typography/component role coverage passed:** NO
- **Every-screen optical alignment passed:** NO
- **Studio visual finish passed:** NO
- **Asset/editorial/appearance passed:** PARTIAL
- **Visual regression/specialist review passed:** NO
- **Accessibility environment matrix passed:** NO
- **VoiceOver core path passed:** NO
- **Accessibility Nutrition Labels assessed:** NO
- **Accessibility common-task matrix passed:** NO
- **iPad adaptive layout/input passed:** PARTIAL
- **Touched screens fresh visual pass:** PARTIAL
- **State migration/update safety passed:** NO
- **Failure-mode UX passed:** PARTIAL
- **Interactive affordance audit passed:** NO
- **Post-action outcome clarity passed:** PARTIAL
- **Touch-target truth audit passed:** NO
- **Brand identity asset parity passed:** PARTIAL
- **Premium interaction detail passed:** NO
- **Screen destination/return audit passed:** NO
- **Scroll/restoration/micro-detail audit passed:** NO
- **Content completeness/duplication passed:** PARTIAL
- **Store truth/assets passed:** NO
- **Privacy-safe analytics/diagnostics passed:** PARTIAL
- **Interruption handling passed:** NO
- **End-to-end duration/value passed:** NO
- **Completion/replayability passed:** NO
- **Fun/design critique passed:** NO
- **Premium product rescue mode:** REQUIRED
- **Design intent/premium elevation passed:** NO
- **Full playflow coherence passed:** NO
- **Skill-over-luck/player-agency passed:** NOT APPLICABLE - media client
- **Free-to-paid desire passed:** NOT APPLICABLE - free app
- **Player experience pass:** NO
- **First-time pickup verified:** PARTIAL
- **Cognitive-load/clarity passed:** PARTIAL
- **First-session hook present:** PARTIAL
- **Emotional-journey playthrough completed:** NO
- **Evidence honesty classification applied:** YES
- **AAA visual bar passed:** NO
- **Motion/animation quality passed:** NO
- **Empty/loading/error state polish passed:** PARTIAL
- **Frontend experience engineering passed:** NO
- **Frontend architecture/state audit passed:** PARTIAL
- **Client state/loading/error polish passed:** PARTIAL
- **Frontend/backend integration truth passed:** PARTIAL
- **Component contract consistency passed:** PARTIAL
- **Frontend render/performance stability passed:** NO
- **Independent frontend challenge passed:** NO
- **Unresolved frontend findings:** launch packaging was fixed; player proof remains incomplete
- **Senior-dev bug/performance/security audit passed:** NO
- **Measured performance engineering passed:** NO
- **Security/threat/dependency engineering passed:** PARTIAL
- **Independent engineering challenge passed:** NO
- **Unresolved engineering findings:** process resource ceiling prevented completion of current-build player captures and validators
- **Adversarial no-BS critic passed:** NO

## Executive Summary

- **Overall read:** Xuva is a viable self-hosted media client concept, and the current lane fixed two real iOS blockers. It is not yet a shippable App Store build because full player, accessibility, performance, and product-depth gates are not complete.
- **What worked well:** The iOS app now launches on iOS 26.5 simulators; first-run connect UI is understandable; seeded paired-state home loads a scanned/probed movie from a disposable server on iPhone and iPad.
- **What failed or felt weak:** Before the scene-manifest fix, SpringBoard denied launch. Before the playback request fix, iOS direct-playable media was blocked by the default server policy. The run did not complete detail/player screenshots or the full device/accessibility/performance matrix.
- **Would I ship this build?:** No

## Gate

| Dimension | Result | Evidence |
| --- | --- | --- |
| Bot completion | STOP | `BotRuns.md` |
| Duration/value/retention | STOP | `DurationValue.md` |
| Onboarding/help | PARTIAL | `OnboardingHelp.md` |
| Visual/UI/UX/a11y | STOP | `VisualUX.md` |
| Visual-system precision | STOP | `VisualSystem.md` |
| Interactive playability | STOP | `InteractiveAudit.md` |
| Premium interaction detail | STOP | `PremiumDetail.md` |
| Frontend engineering | STOP | `FrontendEngineering.md` |
| Performance | STOP | `Performance.md` |
| Security/privacy | STOP | `SecurityPrivacy.md` |
| Release scorecard | STOP | `ReleaseScorecard.md` |
| Adversarial critic | STOP | `CriticReview.md` |

## Findings

| # | Severity | Area | Finding | Fix / Decision | Status |
| --- | --- | --- | --- | --- | --- |
| 1 | Blocker | iOS packaging | iOS 26.5 SpringBoard denied launch while built Info.plist contained a generated empty scene manifest. | Removed `INFOPLIST_KEY_UIApplicationSceneManifest_Generation` from both iOS build configurations. | FIXED |
| 2 | Blocker | Playback | iOS requested adaptive HLS by default, causing `blocked_by_policy` under server default `original_only` even for H.264/AAC direct-playable media. | iPhone/iPad now send `preferAdaptive: false`; tvOS keeps `true`. API proof returns direct route ready. | FIXED API, NEEDS SIM PLAYER PROOF |
| 3 | Blocker | Deep QA | Full player/detail/current-build matrix proof was interrupted by macOS process exhaustion during simulator screenshots. | Continue once simulator resources are stable; release Xuva simulators before rerun. | OPEN |
| 4 | Blocker | Real device | Physical iPhone/iPad devices were offline in Xcode readback. | Real-device proof unavailable in this lane. | OPEN |
| 5 | Major | Product depth | Home/library proof exists, but watchlist, settings/reset, detail/player, failure recovery, and persisted playback state were not fully proven on-device. | Continue capability audit. | OPEN |
| 6 | Major | Accessibility/performance | Dynamic Type, VoiceOver, accessibility environments, measured launch/memory/frame/network budgets are incomplete. | Run full matrix before TestFlight. | OPEN |
| 7 | Major | Release process | Validators are red and no ASC/TestFlight readback exists. | Do not upload or submit. | OPEN |
