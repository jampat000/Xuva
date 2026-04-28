# Vyrden Handover Journal

Operational handover for building Vyrden into a production-grade media server competitor.

Use this as the single source of truth for execution, delivery tracking, and acceptance.

---

## How To Use This Journal

- **Execution model**: continuous flow, no sprint framing.
- **Task state**: tick checkboxes when complete.
- **Definition of done**: every task requires artifacts and acceptance evidence.
- **Order of execution**: follow the Priority Sections in sequence unless dependencies allow parallel work.
- **Change control**: if scope changes, update this journal first.

### Status Legend

- `[ ]` Not started
- `[-]` In progress
- `[x]` Completed
- `[!]` Blocked

### Evidence Requirements (applies to all tasks)

Each completed task must include:

1. Code/PR references.
2. Test evidence (automated or manual steps).
3. Observability evidence (logs/metrics/events if relevant).
4. Risk notes + rollback notes.

Add these under each task’s **Completion Notes** subsection when done.

---

## Program Objectives (Non-Negotiable Outcomes)

1. Secure-by-default server and API model.
2. Transparent playback decisions users can understand without forum help.
3. Reliable playback under load with background-job isolation.
4. Large-library performance without degraded UX.
5. Remote access that is user-owned, diagnosable, and honest.

---

## Release Gates (Must Pass Before Broad Public Beta)

### Gate A: Security
- [ ] Authenticated API access for protected routes.
- [ ] Role-based authorization for admin actions.
- [ ] Signed short-lived streaming URLs.
- [ ] Security logging + auditability for sensitive events.

### Gate B: Playback Transparency
- [ ] Every route decision includes explicit reason and suggested fix.
- [ ] Subtitle impact shown before playback starts.
- [ ] Mid-session inspector exposes route, load, and change history.

### Gate C: Reliability
- [ ] Sessions and jobs survive restart/recovery.
- [ ] No zombie transcode/download/session states.
- [ ] Operational health checks and failure diagnostics in place.

### Gate D: Scale
- [ ] Stable behavior on 5,000+ file library tests.
- [ ] Scans/probes do not starve active playback.
- [ ] Catalog and dashboard remain responsive during background work.

### Gate E: Remote Experience
- [ ] Remote diagnostics identify likely failure class.
- [ ] Guided next actions are actionable and specific.
- [ ] Adaptive streaming path available for unstable links.

---

## Priority 0: Security & Trust Foundation

### P0.1 Local Authentication and Session Security
**Objective**: prevent anonymous control/stream abuse and establish identity model.

**Tasks**
- [ ] Implement user credential model with strong password hashing (`argon2id`).
- [ ] Implement login/logout/session lifecycle (expiry and revocation).
- [ ] Add CSRF protection to browser-initiated write operations.
- [ ] Add brute-force mitigation (rate limits and temporary lockouts).
- [ ] Add secure cookie/session settings and token rotation strategy.

**Deliverables**
- Auth service module and protected middleware integration.
- Updated API route policy map (public vs protected).
- Session invalidation behavior documented and tested.

**Acceptance Criteria**
- Unauthorized requests to protected endpoints return `401`.
- Role violations return `403`.
- Expired sessions cannot perform API actions.
- Logout invalidates current session immediately.

**Dependencies**
- None.

**Completion Notes**
- PR(s):
- Tests:
- Metrics/log evidence:
- Risks/rollback:

---

### P0.2 Authorization and Route Hardening
**Objective**: enforce least-privilege across administrative and media operations.

**Tasks**
- [ ] Add role checks (`admin`, `standard`) for settings/library/system actions.
- [ ] Enforce authorization for `/api/media-sources/*`, `/api/work/*`, `/api/downloads/*`, `/api/settings*`.
- [ ] Create centralized policy registry for route-level access.
- [ ] Ensure internal service operations record acting user context.

**Deliverables**
- Authorization middleware.
- Route policy definition file/table.
- Audit fields in high-impact mutations.

**Acceptance Criteria**
- Standard users cannot access admin-only routes.
- Audit logs include actor identity for sensitive operations.

**Dependencies**
- P0.1

**Completion Notes**
- PR(s):
- Tests:
- Metrics/log evidence:
- Risks/rollback:

---

### P0.3 Signed Streaming URLs and Playback Session Binding
**Objective**: prevent hotlinking and token replay on media streams.

**Tasks**
- [ ] Replace direct file stream access with signed URL tokens.
- [ ] Add token TTL and signature verification.
- [ ] Bind token to session/user/device context.
- [ ] Optionally bind token to client IP/fingerprint with tolerance rules.
- [ ] Enforce per-user and per-device stream concurrency limits.

**Deliverables**
- Stream token issuance and verification logic.
- Updated player flow to request authorized stream URL.
- Rejection/error taxonomy for invalid token states.

**Acceptance Criteria**
- Expired tokens are rejected.
- Reused/replayed tokens are rejected when policy requires single-use/session binding.
- Stream access without valid session context is denied.

**Dependencies**
- P0.1, P0.2

**Completion Notes**
- PR(s):
- Tests:
- Metrics/log evidence:
- Risks/rollback:

---

### P0.4 Security Baseline and Auditability
**Objective**: establish a repeatable security baseline and event traceability.

**Tasks**
- [ ] Add secure headers and strict CORS policy.
- [ ] Validate and sanitize all user-controlled inputs.
- [ ] Add path traversal and filesystem safety tests.
- [ ] Add dependency vulnerability checks to CI.
- [ ] Add security/audit event stream (`auth`, `settings`, `library`, `stream`).

**Deliverables**
- Security middleware package.
- CI security checks configuration.
- Audit event schema documentation.

**Acceptance Criteria**
- Security checks run in CI for every merge request.
- Sensitive actions can be reconstructed from audit logs.

**Dependencies**
- P0.1, P0.2

**Completion Notes**
- PR(s):
- Tests:
- Metrics/log evidence:
- Risks/rollback:

---

## Priority 1: Playback Correctness & User Trust

### P1.1 Playback Decision Engine v2
**Objective**: replace coarse heuristics with deterministic, explainable compatibility decisions.

**Tasks**
- [ ] Extend input model: profile/level/bit-depth/HDR/frame-rate/audio/subtitle capabilities.
- [ ] Add network-aware policy inputs (estimated throughput, route type).
- [ ] Formalize decision order and tie-break rules.
- [ ] Emit canonical decision object with machine and human fields.
- [ ] Add decision trace IDs for debug correlation.

**Deliverables**
- Expanded decision schema and engine implementation.
- Backward-compatible API response strategy or versioned endpoint.
- Decision contract documentation.

**Acceptance Criteria**
- Identical inputs produce identical decisions.
- Decision outputs include exact reason and suggested alternatives.
- Regression tests cover direct/remux/audio/subtitle/video branches.

**Dependencies**
- P0 gates strongly recommended before broad rollout.

**Completion Notes**
- PR(s):
- Tests:
- Metrics/log evidence:
- Risks/rollback:

---

### P1.2 Subtitle Pipeline (Core Differentiator)
**Objective**: eliminate subtitle-driven playback confusion and performance surprises.

**Tasks**
- [ ] Classify subtitle types (text/image) with capability matching.
- [ ] Add conversion workflow where feasible.
- [ ] Burn-in only as last resort with explicit warning.
- [ ] Add pre-play forecast updates for subtitle selections.
- [ ] Expose subtitle impact in inspector and decision payload.

**Deliverables**
- Subtitle compatibility matrix by client profile.
- Subtitle conversion/burn-in decision branch implementation.
- User-facing warning and fix suggestions.

**Acceptance Criteria**
- Subtitle toggle changes forecast before playback starts.
- Image subtitles triggering burn-in are clearly indicated.
- Users can identify why subtitles changed server load.

**Dependencies**
- P1.1

**Completion Notes**
- PR(s):
- Tests:
- Metrics/log evidence:
- Risks/rollback:

---

### P1.3 Transcode Reliability and Failure Classification
**Objective**: make transcode behavior deterministic, debuggable, and recoverable.

**Tasks**
- [ ] Parse ffmpeg failures into normalized error categories.
- [ ] Add retry/backoff policy for retryable failures.
- [ ] Add hard timeout and cancellation cleanup.
- [ ] Track job state transitions with persisted reasons.
- [ ] Surface actionable remediation tips in API/UI.

**Deliverables**
- Error taxonomy spec and parser.
- Job lifecycle enhancements with robust status transitions.
- Operator-visible failure diagnostics.

**Acceptance Criteria**
- Transcode failure responses contain categorized reason and next action.
- Cancelled jobs release resources and clean temp artifacts.
- No orphaned running states after process termination/restart.

**Dependencies**
- P1.1; durable job storage from P2.1 strongly recommended.

**Completion Notes**
- PR(s):
- Tests:
- Metrics/log evidence:
- Risks/rollback:

---

### P1.4 Live Playback Inspector
**Objective**: make active playback operationally transparent to users and admins.

**Tasks**
- [ ] Include current route, mode, selected tracks, bitrate, buffer health.
- [ ] Include server impact and transcode speed/load.
- [ ] Add route-change events (if adaptation occurs).
- [ ] Ensure inspector updates via SSE without blocking server work.

**Deliverables**
- Inspector API enhancements.
- Dashboard/player inspector UI.
- Event schema for session updates.

**Acceptance Criteria**
- Active session state is visible and updates in near real-time.
- Users can see exact reason for current route and server impact.

**Dependencies**
- P1.1, P1.2, P1.3

**Completion Notes**
- PR(s):
- Tests:
- Metrics/log evidence:
- Risks/rollback:

---

## Priority 2: Durability, Operations, and Scale

### P2.1 Persistent Sessions and Job Recovery
**Objective**: remove memory-only runtime risk and survive restarts cleanly.

**Tasks**
- [ ] Persist sessions, transcode jobs, probe jobs, download jobs.
- [ ] Implement startup reconciliation/recovery process.
- [ ] Add heartbeats with TTL cleanup for stale entities.
- [ ] Add crash-safe state transitions.

**Deliverables**
- New/updated DB tables and migration set.
- Recovery worker logic.
- Operational docs for restart behavior.

**Acceptance Criteria**
- Restart does not lose active/recent operational state.
- Stale entities transition to terminal states with reason.

**Dependencies**
- P0.1 for user identity references.

**Completion Notes**
- PR(s):
- Tests:
- Metrics/log evidence:
- Risks/rollback:

---

### P2.2 Observability and Operational Diagnostics
**Objective**: make production behavior diagnosable without guesswork.

**Tasks**
- [ ] Add structured logging with request/session correlation IDs.
- [ ] Add metrics for API latency/errors, queue depth, worker utilization, playback failures.
- [ ] Add health/readiness endpoints with subsystem granularity.
- [ ] Add alert thresholds for queue saturation and transcode error spikes.

**Deliverables**
- Logging standard and field schema.
- Metrics instrumentation and dashboard baseline.
- Health endpoints and runbook.

**Acceptance Criteria**
- Any playback failure can be traced via correlated logs/metrics.
- Operators can see impending resource saturation before user impact.

**Dependencies**
- None; can run in parallel with P2.1.

**Completion Notes**
- PR(s):
- Tests:
- Metrics/log evidence:
- Risks/rollback:

---

### P2.3 Large Library Performance
**Objective**: handle 5k+ media files without degrading interactivity.

**Tasks**
- [ ] Implement incremental scan mode and changed-file prioritization.
- [ ] Add filesystem watcher mode with throttled batching.
- [ ] Tune DB indexes based on observed query hotspots.
- [ ] Add storage-type-aware scan/probe worker defaults.
- [ ] Add benchmark suite for scan/probe/catalog responsiveness.

**Deliverables**
- Incremental scanning implementation.
- Performance tuning configuration.
- Benchmark report template and baseline results.

**Acceptance Criteria**
- Re-scan duration materially reduced for low-change libraries.
- Playback remains stable during scan/probe on large datasets.
- Dashboard remains responsive during background jobs.

**Dependencies**
- P2.1, P2.2 recommended.

**Completion Notes**
- PR(s):
- Tests:
- Metrics/log evidence:
- Risks/rollback:

---

### P2.4 Frontend Maintainability and Safety
**Objective**: reduce regression risk by decomposing the large web app script.

**Tasks**
- [ ] Refactor web app into domain modules (dashboard/playback/metadata/settings/activity).
- [ ] Introduce typed API client contracts (if moving to TS, define migration path).
- [ ] Add component/unit tests for playback forecast + inspector rendering.
- [ ] Add error boundary and retry UX patterns.

**Deliverables**
- Modular frontend structure.
- Test coverage for critical UI logic.
- Frontend architecture notes.

**Acceptance Criteria**
- UI regressions in forecast/decision visibility are caught by automated tests.
- New contributors can make localized changes without touching monolithic script.

**Dependencies**
- None; can begin early but most valuable after P1 contracts stabilize.

**Completion Notes**
- PR(s):
- Tests:
- Metrics/log evidence:
- Risks/rollback:

---

## Priority 3: Remote Playback and Adoption Levers

### P3.1 Remote Diagnostics Assistant
**Objective**: convert remote-access uncertainty into clear troubleshooting outcomes.

**Tasks**
- [ ] Add connectivity checks (LAN/WAN reachability, TLS validity, route classification).
- [ ] Identify probable failure class (NAT, DNS, cert, firewall, throughput).
- [ ] Provide targeted next-action guidance (proxy, VPN, port-forward paths).
- [ ] Preserve local-first principle (no mandatory vendor relay).

**Deliverables**
- Diagnostics API and UI.
- User guidance templates.
- Known-limitations messaging.

**Acceptance Criteria**
- Failed remote setup produces specific failure class + concrete next actions.

**Dependencies**
- P2.2 recommended for diagnostics quality.

**Completion Notes**
- PR(s):
- Tests:
- Metrics/log evidence:
- Risks/rollback:

---

### P3.2 Adaptive Streaming (HLS/DASH)
**Objective**: improve playback continuity on unstable/limited networks.

**Tasks**
- [ ] Implement adaptive packaging path with bitrate ladders.
- [ ] Add route selection logic for direct/remux/adaptive.
- [ ] Add client capability checks for adaptive support.
- [ ] Add monitoring for adaptation behavior and stall rates.

**Deliverables**
- Adaptive stream generation + serving path.
- Playback route policy update.
- Telemetry for adaptation outcomes.

**Acceptance Criteria**
- Remote sessions degrade gracefully instead of hard buffering/failure.

**Dependencies**
- P1.1, P1.3, P2.2

**Completion Notes**
- PR(s):
- Tests:
- Metrics/log evidence:
- Risks/rollback:

---

### P3.3 Migration and Adoption Tooling
**Objective**: reduce switching cost from Plex/Emby/Jellyfin.

**Tasks**
- [ ] Define import format support for watched/resume/metadata IDs.
- [ ] Build dry-run import with conflict report.
- [ ] Build selective import and rollback/undo safety.
- [ ] Add post-import verification report.

**Deliverables**
- Import service + validation tooling.
- Mapping diagnostics and conflict resolution UX.
- Migration documentation.

**Acceptance Criteria**
- User can preview and execute migration with clear outcome report and minimal manual cleanup.

**Dependencies**
- P2.1, metadata integrity from P1/P2.

**Completion Notes**
- PR(s):
- Tests:
- Metrics/log evidence:
- Risks/rollback:

---

## Cross-Cutting Backlog (Continuous)

### CX.1 Quality Engineering
- [ ] Golden media corpus covering container/codec/subtitle/HDR edge cases.
- [ ] End-to-end tests for scan -> probe -> decision -> route -> session.
- [ ] Fault-injection tests for ffmpeg/provider/storage/network failure modes.
- [ ] Compatibility matrix tests by client profile.

### CX.2 Documentation
- [ ] Operator runbook (install, configure, monitor, recover).
- [ ] API contract reference for playback/decision/session endpoints.
- [ ] Troubleshooting guide for top failure classes.
- [ ] Security model documentation and threat assumptions.

### CX.3 Product Safeguards
- [ ] Paid feature boundaries remain non-blocking for core local playback.
- [ ] In-product messaging stays explicit about local-first and user-owned remote model.

---

## Dependency Graph (Execution Order)

1. P0.1 -> P0.2 -> P0.3 -> P0.4
2. P1.1 -> P1.2 -> P1.3 -> P1.4
3. P2.1 and P2.2 (parallel) -> P2.3 -> P2.4
4. P3.1 -> P3.2
5. P3.3 after P2.1 and metadata stability work
6. Cross-cutting work continues throughout

---

## Master Checklist (Single View)

### Security
- [ ] P0.1 Local Authentication and Session Security
- [ ] P0.2 Authorization and Route Hardening
- [ ] P0.3 Signed Streaming URLs and Playback Session Binding
- [ ] P0.4 Security Baseline and Auditability

### Playback
- [ ] P1.1 Playback Decision Engine v2
- [ ] P1.2 Subtitle Pipeline
- [ ] P1.3 Transcode Reliability and Failure Classification
- [ ] P1.4 Live Playback Inspector

### Reliability & Scale
- [ ] P2.1 Persistent Sessions and Job Recovery
- [ ] P2.2 Observability and Operational Diagnostics
- [ ] P2.3 Large Library Performance
- [ ] P2.4 Frontend Maintainability and Safety

### Remote & Growth
- [ ] P3.1 Remote Diagnostics Assistant
- [ ] P3.2 Adaptive Streaming (HLS/DASH)
- [ ] P3.3 Migration and Adoption Tooling

### Continuous
- [ ] CX.1 Quality Engineering
- [ ] CX.2 Documentation
- [ ] CX.3 Product Safeguards

---

## Risks Register Template

Use this block as tasks progress:

- **Risk**:
- **Area**:
- **Impact**:
- **Likelihood**:
- **Mitigation**:
- **Owner**:
- **Status**:

---

## Handover Notes For Incoming Engineers

- Focus first on **P0 + P1**; they directly define trust and user-perceived quality.
- Do not ship major remote features before security and reliability gates are green.
- Keep playback-decision outputs backward compatible where possible; UI and telemetry depend on stable contract fields.
- Treat subtitle handling as a core product capability, not a follow-up enhancement.
- Preserve local-first principles in every architecture and product decision.

