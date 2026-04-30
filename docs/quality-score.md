# Quality Scorecard

This scorecard gives agents a compact way to understand current quality by domain. Scores are intentionally rough and should be updated when work materially changes risk.

| Domain | Score | Evidence | Gaps |
| --- | ---: | --- | --- |
| Local-first product direction | 8 | Product principles, remote diagnostics, desktop alpha docs, Apple TV alpha docs | First-run setup still needs packaging validation. |
| Playback decision engine | 7 | Playback v2 docs, decision tests, adaptive streaming tests | Native client playback-start contract needs hardening. |
| Web admin UI | 7 | Static modules, frontend tests, settings tabs, folder browser | Needs more browser-driven journey tests. |
| Runtime folders and storage | 7 | Browse API, saved config overlay tests, runtime docs | Native desktop picker bridge not yet implemented in a shell. |
| Apple TV alpha | 4 | SwiftUI shell, bootstrap, pairing, TV home contract | Not compiled on Xcode yet; no AVPlayer route playback. |
| Auth and authorization | 7 | Local auth, route policies, CSRF, protected tests | Device credentials after pairing are not persistent yet. |
| Observability | 6 | Request IDs, metrics, events, operations runbook | No local queryable logs/traces stack yet. |
| Agent harness | 4 | AGENTS map, docs index, plan structure, harness check | CI does not run harness check yet. |

## Update Rules

- Raise a score only when tests, docs, or working product evidence improve.
- Lower a score when a gap is found, even if the code still passes tests.
- Add a row for every major product/runtime domain that agents will touch repeatedly.
- Link durable evidence in docs or tests instead of relying on chat history.
