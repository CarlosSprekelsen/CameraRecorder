# Project Roadmap

## Quality and Control Point Overview

This roadmap enforces a strict quality pipeline to guarantee architectural compliance, implementation rigor, and auditability. **Progression through each phase is controlled by explicit IV&V (Independent Verification & Validation) "control points."** No Epic or Story may advance until all required IV&V gates are fully passed and evidence of implementation is present in both code and documentation.

**The high-level flow:**

1. **Architecture & Scaffolding**
    - Complete and approve all architecture documents, API contracts, and configuration structures.
    - Implement all method and handler stubs to mirror the architecture (no business logic yet).
    - [IV&V Control Point Gate]: All stubs, APIs, and docs must match the approved architecture before logic is written.

2. **Implementation & Integration**
    - Fill in all business logic, handlers, and error recovery as described in the architecture and stories.
    - Replace all placeholders and STOP/TODOs with working logic.
    - [IV&V Control Point Gate]: Implementation and integration must be reviewed and validated against requirements and architecture.

3. **Testing & Verification**
    - Develop and run all unit, integration, and workflow tests.
    - Ensure code coverage and behavioral correctness.
    - [IV&V Control Point Gate]: All tests must pass, and IV&V review must confirm feature completeness, coding standards, and no accidental scope creep.

4. **Release & Operations**
    - Prepare deployment scripts, operational docs, and rollback strategies.
    - Validate deployment on target environments.
    - [IV&V Control Point Gate]: All deployment and operations steps must be tested, documented, and signed off.

**Key Control Point Rules:**
- **Cannot move to the next Story/Epic/phase until the preceding IV&V "control point" Story is fully checked off.**
- **A task is only marked as [x] complete when both the architectural decision and the corresponding code, documentation, and tests are fully implemented and validated.**
- **STOP/BLOCKED items are moved to "Resolved Blockers" only when implementation is finished and matches the updated architecture.**
- **Every IV&V Story must be validated by review and explicit evidence (file/commit/section) before proceeding.**
- **No scope creep: Features and requirements may only be added or changed through explicit architectural revision and IV&V.**

---

## IV&V Reviewer Checklist

Every IV&V (Independent Verification & Validation) control point MUST be reviewed and signed off using the following checklist:

- [ ] All corresponding implementation, documentation, and configuration tasks are complete (not just decided—built, tested, and validated).
- [ ] All STOP, TODO, and placeholder code related to this control point have been replaced by working, validated logic or explicitly deferred with rationale.
- [ ] No accidental scope/feature creep: The codebase and docs match only what is authorized in the architecture and roadmap.
- [ ] Evidence of completion is present: All roadmap tasks reference the file(s), doc(s), or commit(s) demonstrating completion.
- [ ] All cross-referenced IV&V or Epic/Story dependencies are also satisfied (no unaddressed downstream blockers).
- [ ] Tests for this control point are present, passing, and cover all critical flows.
- [ ] Documentation and changelogs are updated to reflect the change.
- [ ] STOP/BLOCKED items are only marked resolved when actual implementation is present.
- [ ] Any ambiguities or deviations from architecture are annotated for further decision/action (no silent acceptance).
- [ ] Reviewer/lead signs off and records date and name.

> _No Epic, Story, or phase may advance until all checklist items above are checked, with evidence and reviewer signature/date._

---

## Pre-Completion Validation Checklist

Before marking any task as [x] complete:
1. [ ] Code change is present and functional (not just TODO/STOP comments)  
2. [ ] Related tests exist and pass  
3. [ ] Documentation updated if needed  
4. [ ] No NotImplementedError, `pass` statements, or "TODO" comments remain unaddressed (or are canonical deferred decisions)  
5. [ ] Reviewer has validated the evidence  
6. [ ] Evidence field contains specific file/line/commit references  

---

## 🌍 Epics (Long-Term Goals)

### E1: Robust Real-Time Camera Service Core - SUBSTANTIALLY COMPLETE (FAST TRACK)

- **S1a: Architecture Scaffolding (COMPLETE)**  
    - Status: ✅ Complete  
    - Summary: API contracts, configuration structures, method/handler stubs, and documentation frameworks are in place and aligned with the approved architecture.  

- **S2: Architecture Compliance IV&V (Control Point) - COMPLETE**  
    - Status: ✅ Complete  
    - Summary: Stubs and scaffolding validated against architecture; no accidental scope creep; coding standards and docstring requirements confirmed.  
    - Evidence Sources: `docs/architecture/overview.md`, `docs/development/principles.md`, audit artifacts.

- **S2b: Fast-Track Audit Baseline (Informational)**  
    - Status: ✅ Completed / Baseline Captured  
    - Purpose: Capture the actual implementation state from fast-track work to feed into S3/S4 closure. Not a blocking gate if clarifications remain; findings were folded into subsequent stories.  
    - Audit Artifacts: `WebSocket Server Code Audit.md`, `MediaMTX Controller Code Audit.md`, `Camera Service Manager Audit.md`, `Camera Discovery Module Security Audit.md`  
    - Summary Findings: Core modules largely implemented; remaining deficiencies identified around metadata confirmation, observability, health recovery logic, and test scaffolds.  

- **S3: Camera Discovery & Monitoring Implementation - PARTIALLY COMPLETE**  
    - Status: 🟡 In progress  
    - Bug / Partial Summary:  
        - ✅ Service manager lifecycle and observability hardening completed: Enhanced camera event orchestration with deterministic sequencing, robust error recovery, and comprehensive observability. Implemented provisional/confirmed metadata state tracking in notifications with correlation ID traceability throughout lifecycle events. Added defensive guards for MediaMTX controller failures with fallback behaviors.
        - Udev event processing test coverage gaps (change events, race conditions, invalid nodes, fallback to polling).  
    - Key Deliverables Remaining:  
        - Expand discovery test coverage for udev event processing edge cases.
    - Evidence: `src/camera_service/service_manager.py` (lines 650-750), `tests/unit/test_camera_service/test_service_manager_lifecycle.py` (complete test suite), notification schema enhanced with metadata validation flags (2025-08-04). 

- **S4: MediaMTX Integration - PARTIALLY COMPLETE**  
    - Status: 🟡 In progress  
    - Bug / Partial Summary:  
        - ✅ Health monitor edge-case testing completed (2025-08-04): Circuit breaker flapping resistance, recovery confirmation logic, and backoff/jitter behavior comprehensively validated with 33 new test methods.
        - Snapshot capture and recording duration implementation hardened; closure documentation pending.  
        - Logging and error context improvements mostly applied; verify consistency.  
    - Key Deliverables Remaining:  
        - Add explicit architectural decision log entries marking snapshot and duration partials as closed.  
        - Finalize any remaining test gaps for robustness and observability.

- **S5: Core Integration IV&V (Control Point) - PENDING**  
    - Status: 🔴 Pending  
    - Summary: End-to-end acceptance test scenarios, lifecycle validation, error recovery workflows, and deployment bootstrap are not yet validated.  
    - Key Work:  
        - Define and execute acceptance test matrix (camera → MediaMTX → notification flows).  
        - Validate service startup/shutdown orchestration.  
        - Test error injection and recovery for critical paths.  
        - Harden deployment script for repeatable environments.  

### E2: Security and Production Hardening - PENDING E1 COMPLETION

- **S6: Security Features Implementation**  
    - Status: ⬜ Pending  
    - Tasks: Authentication (JWT/API key), health check endpoints, rate limiting/connection control, TLS/SSL support.  

- **S7: Security IV&V (Control Point)**  
    - Status: ⬜ Pending  
    - Gate: Authentication/authorization, access control, security test cases must be reviewed and passing before proceeding to E3.  

### E3: Client API & SDK Ecosystem - PENDING E2 COMPLETION

- **S8: Client APIs and Examples**  
    - Status: ⬜ Pending  
    - Tasks: Add client usage examples, document authentication/usage, create SDKs.  

- **S9: SDK & Docs IV&V (Control Point)**  
    - Status: ⬜ Pending  
    - Gate: Review docs/examples for accuracy, usability testing.  

### E4: Future Extensibility - PLANNING ONLY

- **S10: Cloud/Protocol Extensions (Planning Only)**  
    - Status: ⬜ Planning  
    - Tasks: Placeholder docs for future protocols/cloud integration and plugin architectures.  

- **S11: Extensibility IV&V (Control Point)**  
    - Status: ⬜ Pending  
    - Gate: Review and approve future extension points before E5.

### E5: Deployment & Operations Strategy - PENDING E3 COMPLETION

- **S12: Deployment Automation & Ops**  
    - Status: ⬜ Pending  
    - Tasks: Complete deployment scripts, document environment integration, rollback/backup procedures, monitoring/alerting.  

- **S13: Deployment IV&V (Control Point)**  
    - Status: ⬜ Pending  
    - Gate: Validate deployment on target environments, verify ops docs, and backup/recovery.

---

## 🌱 Cross-Epic Stories

### S14: Automated Testing & Continuous Integration - PARTIALLY COMPLETE
- Status: 🟡 In progress  
- Summary: Test scaffolds exist for core modules (websocket server, MediaMTX controller, service manager, hybrid monitor). Supporting components (`main.py`, `config.py`, `logging_config.py`) require proper stubs and integration. CI pipeline is not fully wired to enforce quality gates.  
- Key Deliverables:  
    - Fill out missing test scaffolds for camera_service support modules.  
    - Add README for test conventions.  
    - Bootstrap CI (lint, type-check, run unit tests).  
    - Surface integration/happy-path test harness for S5.

### S15: Documentation & Developer Onboarding - PARTIALLY COMPLETE
- Status: 🟡 In progress  
- Summary: Core principles, coding standards, and architectural overview exist. Need to (a) capture resolved partials and decisions, (b) sync API docs, (c) provide a concise test/acceptance guide for upcoming IV&V work.  
- Key Deliverables:  
    - Update API docs to reflect actual implemented fields and behaviors.  
    - Document capability confirmation and health recovery policies.  
    - Provide a lightweight acceptance test plan for S5.  

---

## Sprint Planning (Post-Audit)

### Objectives for Next Sprint

#### Objective A: Harden & Close S3 Remaining Gaps
- Expand udev event processing tests (`change`, races, invalid nodes, polling fallback).  
- Finalize and validate service manager metadata flow (provisional/confirmed) and lifecycle observability.  
- Reconcile capability merging output with consumed camera metadata end-to-end.

#### Objective B: Close S4 Partial Polishes & Document Closure  
- Add or expand tests for health monitor recovery/flapping behavior with consecutive-success logic.  
- Record and publish decision log entries marking snapshot/recording duration partials as closed.  

#### Objective C: Prepare S5 End-to-End IV&V  
- Draft and codify acceptance test matrix (full flow scenarios).  
- Implement a “happy path” integration harness.  
- Validate service lifecycle (startup/shutdown) with failure injection.  
- Improve deployment bootstrap script for repeatable environments.

#### Objective D: Solidify Test Coverage & CI (S14)  
- Create stubs for `main.py`, `config.py`, and `logging_config.py`.  
- Publish a test README summarizing scope and naming conventions.  
- Wire a minimal CI to run key quality checks automatically.

---

## Backlog (Prioritized)

1. [DONE] Expand udev testing and metadata reconciliation (S3)  
2. [DONE] Harden and validate service manager lifecycle and observability (S3)  
3. [DONE] Add MediaMTX edge-case health monitor tests (S4)  
4. Document closure of resolved partials (S4)  
5. Draft S5 acceptance test plan and implement core integration smoke test  
6. Create missing camera_service support module test stubs (S14)  
7. Add tests README and conventions doc (S14)  
8. Improve deployment/install script (S5)  
9. Enable CI to enforce tests, linting, and type checking (S14)  
10. Begin security feature implementation groundwork (E2)

---

## Status Summary

- **Architecture & Scaffolding (S1a/S2):** ✅ Complete  
- **Fast-track Audit (S2b):** ✅ Baseline captured and folded into stories  
- **Camera Discovery & Monitoring (S3):** 🟡 Partial — hardening/testing remaining  
- **MediaMTX Integration (S4):** 🟡 Partial — polishing and formal closure pending  
- **Core Integration IV&V (S5):** 🔴 Pending  
- **Testing & CI (S14):** 🟡 In progress  
- **Documentation & Onboarding (S15):** 🟡 In progress  
- **Security (E2):** ⬜ Pending  
- **Client APIs/SDK (E3):** ⬜ Pending  
- **Extensibility (E4):** ⬜ Planning  
- **Deployment & Ops (E5):** ⬜ Pending  

---

## How to Use This Roadmap

- Use the **story-level summaries** above as the authoritative view of current state; the backlog items are the concrete next steps.  
- When a story is completed, update its status, add precise evidence (file/line/commit, test names, doc updates), and sign off per the IV&V Reviewer Checklist.  
- Archive detailed per-file micro-tasks (from prior iterations) in a separate “audit journal” document if needed, keeping this file lean.  
- Reference audit artifacts (the four audit MDs) when closing related story gaps to preserve traceability.

---

## Resolved Blockers / Notes

- [x] **Architecture vs implementation misalignment**: Resolved through S2b audits and updated story definitions.  
- [x] **Method-level versioning ambiguity**: Deferred with canonical STOP-style note in `server.py`; decision logged for revisit post-1.0.  
- [x] **Capability merging policy instability**: Updated to frequency-weighted merge with confirmation window; documented and feeding into S3 metadata flow.  
- [x] **Health monitor recovery flapping risk**: Addressed by introducing consecutive-success confirmation logic (configurable).  

---

*Audit artifacts to refer to for detailed origin of issues:*  
- `WebSocket Server Code Audit.md`  
- `MediaMTX Controller Code Audit.md`  
- `Camera Service Manager Audit.md`  
- `Camera Discovery Module Security Audit.md`  
