# Technical Debt Tracker

Debt belongs here when it is known, real, and not being fixed in the current slice. Keep entries short and actionable.

| ID | Area | Debt | Impact | Next Action |
| --- | --- | --- | --- | --- |
| TD-001 | Apple TV auth | Pairing approval returns a device ID but not a durable device credential. | TV app can prove pairing flow, but cannot yet authenticate as a persistent device when auth is enabled. | Design local device credential issue/rotation/revocation. |
| TD-002 | Desktop alpha | Native folder picker bridge is defined but no tray/taskbar shell exists. | Installed UX cannot yet use OS-native folder picker or Restart now. | Choose desktop shell and implement `window.xuvaDesktop.pickFolder`. |
| TD-003 | Observability | Logs and metrics are available locally but not queryable through an agent-friendly stack. | Agents rely on shell logs and API metrics instead of correlated queries. | Add local dev observability adapter or structured log query script. |
| TD-004 | tvOS build | Swift source exists but has not compiled in Xcode/tvOS SDK yet. | Compile errors may appear when Mac hardware arrives. | Import into Xcode tvOS target and record fixes in `docs/plans/active/apple-tv-alpha.md`. |

## Rules

- Do not use this file for vague wishlist items.
- Every row needs a concrete next action.
- Remove or move rows to completed plans when fixed.
