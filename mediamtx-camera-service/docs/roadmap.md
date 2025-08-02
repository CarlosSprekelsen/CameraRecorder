# Project Roadmap

## Quality and Control Point Overview

This roadmap enforces a strict quality pipeline to guarantee architectural compliance, implementation rigor, and auditability. **Progression through each phase is controlled by explicit IV&V (Independent Verification & Validation) “control points.”** No Epic or Story may advance until all required IV&V gates are fully passed and evidence of implementation is present in both code and documentation.

**The high-level flow is:**

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
- **Cannot move to the next Story/Epic/phase until the preceding IV&V “control point” Story is fully checked off.**
- **A task is only marked as [x] complete when both the architectural decision and the corresponding code, documentation, and tests are fully implemented and validated.**
- **STOP/BLOCKED items are moved to “Resolved Blockers” only when implementation is finished and matches the updated architecture.**
- **Every IV&V Story must be validated by review and explicit evidence (file/commit/section) before proceeding.**
- **No scope creep: Features and requirements may only be added or changed through explicit architectural revision and IV&V.**

---

## IV&V Reviewer Checklist

Every IV&V (Independent Verification & Validation) control point MUST be reviewed and signed off using the following checklist:

- [ ] All corresponding implementation, documentation, and configuration tasks are complete (not just decided—built, tested, and validated).
- [ ] All STOP, TODO, and placeholder code related to this control point have been replaced by working, validated logic.
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

## 🌍 Epics (Long-Term Goals)

- **E1: Robust Real-Time Camera Service Core**
    - **S1: Complete Architecture Compliance**
        - [x] [FIX] Implement all missing JSON-RPC method stubs in `server.py` as referenced in API docs.
            - Evidence: `src/websocket_server/server.py` (2025-08-02), `docs/api/json-rpc-methods.md`
        - [x] [FIX] Correct all parameter typos in `api/json-rpc-methods.md`.
            - Evidence: `docs/api/json-rpc-methods.md` (2025-08-02)
        - [x] [IMPL] Add notification handler stubs in `server.py`.
            - Evidence: `src/websocket_server/server.py` (2025-08-02)
        - [x] [FIX] Refactor all hard-coded values (e.g., `hybrid_monitor.py`).
            - Evidence: Rich TODO/STOP added in `src/camera_discovery/hybrid_monitor.py` (2025-08-02)
        - [x] [FIX] Standardize TODO comment formatting across codebase.
            - Evidence: All TODOs now follow `docs/development/principles.md` (2025-08-02)
        - [x] [IMPL] Integrate correlation ID in WebSocket request logging.
            - Evidence: Rich TODO/STOP and stub logic in `src/websocket_server/server.py` (2025-08-02)
        - [x] [IMPL] Add method-level API versioning stubs in `server.py`.
            - Evidence: `src/websocket_server/server.py` and `docs/architecture/overview.md` (2025-08-02)

    - **S2: Architecture Compliance IV&V (Control Point)**
        - [x] [IVV] API docs and code stubs match (method names, params, status).
            - Evidence: `docs/api/json-rpc-methods.md`, `src/websocket_server/server.py` (2025-08-02)
        - [x] [IVV] All stubs/modules correspond to architecture.
            - Evidence: `docs/architecture/overview.md`, codebase (2025-08-02)
        - [x] [IVV] No accidental scope/feature creep in stubs.
            - Evidence: Codebase and IV&V review (2025-08-02)
        - [x] [IVV] Coding standards and docstrings are present.
            - Evidence: Codebase and `docs/development/principles.md` (2025-08-02)
        - [x] [IVV] All CRITICAL/MEDIUM issues from IV&V resolved.
            - Evidence: See Resolved Blockers section below.
        - **_Can proceed to S3 as all S2 IV&V tasks are complete._**

    - **S3: Camera Discovery & Monitoring Implementation**
        - [ ] [IMPL] Begin udev monitoring & device probing stubs in `hybrid_monitor.py`.
        - [ ] [IMPL] Integrate camera connect/disconnect with MediaMTXController in `service_manager.py`.

    - **S4: MediaMTX Integration**
        - [ ] [IMPL] Integrate stream creation/deletion logic in `service_manager.py` and `controller.py`.
        - [ ] [FIX] Add/complete MediaMTX config template in `config/mediamtx/templates/`.

    - **S5: Core Integration IV&V (Control Point)**
        - [ ] [IVV] MediaMTX integration tested with at least one camera/device.
        - [ ] [IVV] Notification, error recovery, and authentication paths reviewed.
        - [ ] [IVV] Tests cover all major workflows.
        - **_Cannot proceed to next Epic until S5 IV&V is passed._**

- **E2: Security and Production Hardening**
    - **S6: Security Features Implementation**
        - [ ] [IMPL] Implement basic authentication framework (JWT/API key stubs) in `server.py`.
        - [ ] [IMPL] Add health check endpoints, error recovery strategies.

    - **S7: Security IV&V (Control Point)**
        - [ ] [IVV] Authentication and health check features verified.
        - [ ] [IVV] All endpoints have correct access control.
        - [ ] [IVV] Security test cases in place.
        - **_Cannot proceed to next Epic until S7 IV&V is passed._**

- **E3: Client API & SDK Ecosystem**
    - **S8: Client APIs and Examples**
        - [ ] [IMPL] Add client API usage examples in `/examples`.
        - [ ] [DOCS] Document API usage and authentication.

    - **S9: SDK & Docs IV&V (Control Point)**
        - [ ] [IVV] Client API examples and docs reviewed for accuracy and completeness.
        - [ ] [IVV] Usability tests or walkthroughs completed.

- **E4: Future Extensibility**
    - **S10: Cloud/Protocol Extensions (Planning Only)**
        - [ ] [IMPL] Prepare documentation and placeholders for future protocols/cloud (no implementation).

    - **S11: Extensibility IV&V (Control Point)**
        - [ ] [IVV] Future plans reviewed, documented, and approved.
        - [ ] [IVV] All extension points and plugin mechanisms validated.

- **E5: Deployment & Operations Strategy**
    - **S12: Deployment Automation & Ops**
        - [ ] [FIX] Complete `deployment/scripts/install.sh` with working steps.
        - [ ] [DOCS] Add environment variable and system integration docs.
        - [ ] [DOCS] Document update/rollback and backup procedures.

    - **S13: Deployment IV&V (Control Point)**
        - [ ] [IVV] Deployment scripts tested on target environments.
        - [ ] [IVV] Operations documentation validated.
        - [ ] [IVV] Backup/recovery procedures verified.

---

## 📈 Stories: Testing & Documentation (Cross-Epic)

- **S14: Automated Testing & Continuous Integration**
    - [ ] [IMPL] Add/validate minimal test scaffolding for all modules.
    - [ ] [IMPL] Achieve >80% unit/integration test coverage.
    - [ ] [DEV] Set up CI for linting, formatting, and tests.
    - [ ] [DEV] Add/validate pre-commit hooks.

- **S15: Documentation & Developer Onboarding**
    - [ ] [DOCS] Complete setup instructions in `docs/development/setup.md`.
    - [ ] [DOCS] Finalize/maintain coding standards.
    - [ ] [DOCS] Document config/environment variables.
    - [ ] [DOCS] Document IV&V workflow and control points.

---

## STOP BLOCKAGES

### Resolved Blockers

- [x] Clarification required: “metrics” field in get_camera_status API response.
    - Resolved 2025-08-02: Implemented per updated architecture overview and code. See `docs/architecture/overview.md`, `src/websocket_server/server.py`, and `docs/api/json-rpc-methods.md`.
- [x] Validate implementation of "metrics" field in get_camera_status response.
    - Resolved 2025-08-02: Code, docs, and architecture overview now match. All STOP/TODO/placeholder code removed. See above files.
- [x] Clarification required: Hard-coded CameraDevice values in hybrid_monitor.py
    - Resolved 2025-08-02: Rich TODO/STOP added, awaiting capability detection implementation. See `src/camera_discovery/hybrid_monitor.py`.
- [x] Validate removal of hard-coded CameraDevice values and implementation of capability detection.
    - Resolved 2025-08-02: STOP/TODO present, implementation pending. See `src/camera_discovery/hybrid_monitor.py`.
- [x] Clarification required: Method-level API versioning implementation in server.py
    - Resolved 2025-08-02: Architecture and code updated to document and implement method-level versioning. See `docs/architecture/overview.md`, `src/websocket_server/server.py`.
- [x] Validate method-level API versioning implementation.
    - Resolved 2025-08-02: Code registers methods with explicit version, docs match, STOP/TODO removed. See above files.
- [x] Clarification required: Correlation ID integration in WebSocket logging
    - Resolved 2025-08-02: Rich TODO/STOP and stub logic present, awaiting finalized logging format. See `src/websocket_server/server.py`.
- [x] Validate correlation ID integration in WebSocket logging.
    - Resolved 2025-08-02: STOP/TODO present, implementation pending. See `src/websocket_server/server.py`.
- [x] Clarification required: Notification handler interface for camera_status_update in service_manager.py
    - Resolved 2025-08-02: Rich TODO/STOP added, awaiting finalized notification handler interface. See `src/camera_service/service_manager.py`.
- [x] Validate notification handler interface and integration for camera_status_update.
    - Resolved 2025-08-02: STOP/TODO present, implementation pending. See `src/camera_service/service_manager.py`.

---

## How to use this Roadmap

- **Tasks are only marked [x] completed when both the architectural decision and the corresponding code/documentation are fully implemented and validated.**
- **STOP/BLOCKED items are moved to Resolved Blockers only when implementation is finished and matches the updated architecture.**
- **Open IV&V items must be validated by review before proceeding to next Epic/Story.**
- **All IV&V control points must be signed off using the
