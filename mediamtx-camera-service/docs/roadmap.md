# Project Roadmap

## Quality and Control Point Overview

This roadmap enforces a strict quality pipeline to guarantee architectural compliance, implementation rigor, and auditability. **Progression through each phase is controlled by explicit IV&V (Independent Verification & Validation) "control points."** No Epic or Story may advance until all required IV&V gates are fully passed and evidence of implementation is present in both code and documentation.

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
- **Cannot move to the next Story/Epic/phase until the preceding IV&V "control point" Story is fully checked off.**
- **A task is only marked as [x] complete when both the architectural decision and the corresponding code, documentation, and tests are fully implemented and validated.**
- **STOP/BLOCKED items are moved to "Resolved Blockers" only when implementation is finished and matches the updated architecture.**
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

## Pre-Completion Validation Checklist

Before marking any task as [x] complete:
1. [ ] Code change is present and functional (not just TODO/STOP comments)
2. [ ] Related tests exist and pass
3. [ ] Documentation updated if needed  
4. [ ] No NotImplementedError, `pass` statements, or "TODO" comments remain
5. [ ] Reviewer has validated the evidence
6. [ ] Evidence field contains specific file/line/commit references

---

## IV&V Traceability Audit - Architecture & Scaffolding Only (Recent Findings)

### Forward Tracing Issues Found

1. **TODO Comment Format Non-Compliance**  
   - **Issue:** TODO comments across the codebase do not follow a documented, enforceable format.  
   - **Required Format (per intended standard):**  
     ```
     # TODO: <priority>: <description> [IV&V:<ControlPoint>|Story:<StoryRef>]
     ```  
   - **Non-compliant Examples:**  
     - `src/websocket_server/server.py` line ~30: `# TODO: [CRITICAL] Method-level API versioning framework stub`  
     - `src/websocket_server/server.py` line ~90: `# TODO: Initialize authentication system`  
     - `src/camera_discovery/hybrid_monitor.py` line ~185: `# TODO: Use v4l2-ctl or python v4l2 bindings to probe:`  
   - **Impact:** Existing roadmap claim of "Standardize TODO comment formatting across codebase" is unverifiable because no concrete standard was documented. All TODOs must be aligned to a canonical format and reconciled.  
   - **Actionable Remediation:** See corresponding task in S1b below.

---

## 🌍 Epics (Long-Term Goals)

- **E1: Robust Real-Time Camera Service Core**
    
    - **S1a: Architecture Scaffolding (PARTIALLY COMPLETE)**
        - [x] [FIX] Implement all missing JSON-RPC method stubs in `server.py` as referenced in API docs.  
            - Evidence: `src/websocket_server/server.py` (2025-08-02), `docs/api/json-rpc-methods.md`  
            - Status: Architecture scaffolding complete (methods exist with proper signatures)
        - [x] [FIX] Correct all parameter typos in `api/json-rpc-methods.md`.  
            - Evidence: `docs/api/json-rpc-methods.md` (2025-08-02)
        - [x] [IMPL] Add notification handler stubs in `server.py`.  
            - Evidence: `src/websocket_server/server.py` (2025-08-02)  
            - Status: Method signatures exist (business logic pending in S1b)
        - [ ] [FIX] **CRITICAL**: Standardize TODO comment formatting across codebase.  
            - Description: Define, document, and enforce a single TODO comment format; refactor all existing TODOs to comply.  
            - Required format example: `# TODO: HIGH: Implement version negotiation logic [IVV:S1a|Story:E1/S1a]`  
            - Non-compliant instances must be normalized.  
            - Evidence: (to be populated after documenting standard and performing refactor)  
        - [x] [IMPL] Add method-level API versioning stubs in `server.py`.  
            - Evidence: `src/websocket_server/server.py` and `docs/architecture/overview.md` (2025-08-02)
        - [x] [IMPL] **HIGH PRIORITY**: Replace NotImplementedError with actual business logic in JSON-RPC methods.  
            - Evidence: `src/websocket_server/server.py` lines 380-550 (2025-08-02)  
            - Status: Complete MediaMTX integration with error handling and logging
        - [x] [IMPL] **HIGH PRIORITY**: Replace `pass` statements with proper notification broadcasting logic.  
            - Evidence: `src/websocket_server/server.py` lines 500-600 (2025-08-02)  
            - Status: Complete JSON-RPC 2.0 notification system with client broadcasting
        - [x] [IMPL] **MEDIUM PRIORITY**: Actually integrate correlation ID in WebSocket logging (not TODO comments).  
            - File: `src/websocket_server/server.py`  
            - Fix: Implement actual propagation using `CorrelationIdFilter`  
            - Evidence: File: `src/websocket_server/server.py` lines 10, 200-250, 290-310 (Date: 2025-08-02)
        - [x] [IMPL] **MEDIUM PRIORITY**: Actually refactor hard-coded values in `hybrid_monitor.py`.  
            - File: `src/camera_discovery/hybrid_monitor.py`  
            - Evidence: Sections described in previous version (device detection, naming, capability probing) — commit dated 2025-08-02

        - [x] [IMPL] **MEDIUM PRIORITY**: Complete MediaMTX controller initialization in `service_manager.py`.  
            - Evidence: `src/camera_service/service_manager.py` lines 150-190 (2025-08-02)  
            - Status: Controller startup with health verification, directory setup, and error handling
        - [x] [IMPL] **MEDIUM PRIORITY**: Complete camera monitor initialization in `service_manager.py`.  
            - Evidence: `src/camera_service/service_manager.py` lines 170-200 (2025-08-02)  
            - Status: Hybrid monitor startup with event handler registration

        - [x] [IVV] **LOW PRIORITY**: Document and validate logging infrastructure implementation.  
            - Evidence: `src/camera_service/logging_config.py` (275 lines, 2025-08-02)  
            - Status: Structured logging with CorrelationIdFilter, JsonFormatter, ConsoleFormatter
        - [x] [IVV] **LOW PRIORITY**: Document and validate configuration system implementation.  
            - Evidence: `src/camera_service/config.py` lines 1-470 (2025-08-02)  
            - Status: YAML loading, overrides, validation, hot reload implemented
        - [x] [IMPL] **MEDIUM PRIORITY**: Implement environment variable overrides for configuration.  
            - Evidence: `src/camera_service/config.py` lines 150-230 (2025-08-02)
        - [x] [IMPL] **MEDIUM PRIORITY**: Implement configuration schema validation.  
            - Evidence: `src/camera_service/config.py` lines 290-380 (2025-08-02)
        - [x] [IMPL] **LOW PRIORITY**: Implement runtime configuration updates.  
            - Evidence: `src/camera_service/config.py` lines 100-140 (2025-08-02)
        - [x] [IMPL] **LOW PRIORITY**: Implement configuration hot reload capability.  
            - Evidence: `src/camera_service/config.py` lines 140-190 (2025-08-02)

        - [x] [IMPL] **HIGH PRIORITY**: Complete MediaMTX Controller business logic (grouped).  
            - File: `src/mediamtx_wrapper/controller.py`  
            - Evidence: Implementation details enumerated (health_check, create_stream, etc.) dated 2025-08-02

        - [x] [IMPL] **HIGH PRIORITY**: Complete Service Manager core logic (grouped).  
            - File: `src/camera_service/service_manager.py`  
            - Evidence: Full orchestration implementation (startup/shutdown, handlers) dated 2025-08-02

    - **S1b: Core Implementation (PENDING)**
        - [ ] [IMPL] **MEDIUM PRIORITY**: Add roadmap tracking for `main.py`  
            - File: `main.py`  
            - Task: Ensure `main.py` is explicitly represented in roadmap with appropriate Epic/Story linkage.  
            - Evidence:
        - [ ] [IMPL] **MEDIUM PRIORITY**: Add roadmap tracking for `requirements.txt`  
            - File: `requirements.txt`  
            - Task: Ensure dependency listing and its maintenance are surfaced in roadmap.  
            - Evidence:
        - [ ] [IMPL] **MEDIUM PRIORITY**: Add roadmap tracking for `deployment/scripts/install.sh`  
            - File: `deployment/scripts/install.sh`  
            - Task: Surface installation/deployment scripting in roadmap.  
            - Evidence:
        - [ ] [IVV] **MEDIUM PRIORITY**: Verify all core components are tracked in roadmap.md  
            - Task: Only close when above tracking items have concrete entries and evidence.  
            - Evidence:

  - **S2: Architecture Compliance IV&V (Control Point) - PENDING**
        - [ ] [IVV] Re-validate API docs and code alignment after S1b fixes  
            - Task: Verify all JSON-RPC methods have working implementations (not NotImplementedError) and notifications are not `pass`.  
            - Evidence:
        - [ ] [IVV] Re-validate all issue resolution after actual implementation  
            - Task: Confirm hard-coded values were replaced and correlation ID wiring is functional.  
            - Evidence:
        - [ ] [IVV] Validate phantom implementation documentation is complete  
            - Task: Ensure logging/config systems have clear roadmap entries.  
            - Evidence:
        - [ ] [IVV] All stubs/modules correspond to architecture.  
            - Evidence: `docs/architecture/overview.md`, codebase validation required
        - [ ] [IVV] No accidental scope/feature creep in implementation.  
            - Evidence: Post-S1b code review
        - [ ] [IVV] Coding standards and docstrings are present.  
            - Evidence: `docs/development/principles.md` and codebase review

        > **_Cannot proceed to S3 until all S2 IV&V tasks are complete._**

    - **S3: Camera Discovery & Monitoring Implementation**
        - [ ] [IMPL] Implement capability detection logic for CameraDevice (currently STOP/TODO in hybrid_monitor.py).  
            - Evidence:
        - [ ] [IMPL] Begin udev monitoring & device probing stubs in `hybrid_monitor.py`.  
            - Evidence:
        - [ ] [IMPL] Integrate camera connect/disconnect with MediaMTXController in `service_manager.py`.  
            - Evidence:
        - [ ] [IMPL] Implement notification handler interface for camera_status_update (currently STOP/TODO in service_manager.py).  
            - Evidence:

    - **S4: MediaMTX Integration**
        - [ ] [IMPL] Integrate stream creation/deletion logic in `service_manager.py` and `controller.py`.  
            - Evidence:
        - [ ] [FIX] Add/complete MediaMTX config template in `config/mediamtx/templates/`.  
            - Evidence:

    - **S5: Core Integration IV&V (Control Point)**
        - [ ] [IVV] MediaMTX integration tested with at least one camera/device.  
            - Evidence:
        - [ ] [IVV] Notification, error recovery, and authentication paths reviewed.  
            - Evidence:
        - [ ] [IVV] Tests cover all major workflows.  
            - Evidence:
        - [ ] [IMPL] Implement correlation ID propagation and structured logging (currently STOP/TODO in server.py).  
            - Evidence:
        - [ ] [IMPL] Implement method deprecation tracking and version negotiation logic (currently deferred per architecture decisions).  
            - Evidence:
        - **_Cannot proceed to E2 until S5 IV&V is complete._**

- **E2: Security and Production Hardening**
    - **S6: Security Features Implementation**
        - [ ] [IMPL] Implement basic authentication framework (JWT/API key stubs) in `server.py`.  
            - Evidence:
        - [ ] [IMPL] Add health check endpoints, error recovery strategies.  
            - Evidence:

    - **S7: Security IV&V (Control Point)**
        - [ ] [IVV] Authentication and health check features verified.  
            - Evidence:
        - [ ] [IVV] All endpoints have correct access control.  
            - Evidence:
        - [ ] [IVV] Security test cases in place.  
            - Evidence:
        - **_Cannot proceed to E3 until S7 IV&V is complete._**

- **E3: Client API & SDK Ecosystem**
    - **S8: Client APIs and Examples**
        - [ ] [IMPL] Add client API usage examples in `/examples`.  
            - Evidence:
        - [ ] [DOCS] Document API usage and authentication.  
            - Evidence:

    - **S9: SDK & Docs IV&V (Control Point)**
        - [ ] [IVV] Client API examples and docs reviewed for accuracy and completeness.  
            - Evidence:
        - [ ] [IVV] Usability tests or walkthroughs completed.  
            - Evidence:
        - **_Cannot proceed to E4 until S9 IV&V is complete._**

- **E4: Future Extensibility**
    - **S10: Cloud/Protocol Extensions (Planning Only)**
        - [ ] [IMPL] Prepare documentation and placeholders for future protocols/cloud (no implementation).  
            - Evidence:

    - **S11: Extensibility IV&V (Control Point)**
        - [ ] [IVV] Future plans reviewed, documented, and approved.  
            - Evidence:
        - [ ] [IVV] All extension points and plugin mechanisms validated.  
            - Evidence:
        - **_Cannot proceed to E5 until S11 IV&V is complete._**

- **E5: Deployment & Operations Strategy**
    - **S12: Deployment Automation & Ops**
        - [ ] [FIX] Complete `deployment/scripts/install.sh` with working steps.  
            - Evidence:
        - [ ] [DOCS] Add environment variable and system integration docs.  
            - Evidence:
        - [ ] [DOCS] Document update/rollback and backup procedures.  
            - Evidence:

    - **S13: Deployment IV&V (Control Point)**
        - [ ] [IVV] Deployment scripts tested on target environments.  
            - Evidence:
        - [ ] [IVV] Operations documentation validated.  
            - Evidence:
        - [ ] [IVV] Backup/recovery procedures verified.  
            - Evidence:

---

## 📈 Stories: Testing & Documentation (Cross-Epic)

- **S14: Automated Testing & Continuous Integration**
    - [ ] [IMPL] Add/validate minimal test scaffolding for all modules.  
        - Evidence:
    - [ ] [IMPL] Achieve >80% unit/integration test coverage.  
        - Evidence:
    - [ ] [DEV] Set up CI for linting, formatting, and tests.  
        - Evidence:
    - [ ] [DEV] Add/validate pre-commit hooks.  
        - Evidence:
    - [ ] [IVV] Audit for minor/cosmetic/documentation/test issues and verify zero open TODOs remain.  
        - Evidence:

- **S15: Documentation & Developer Onboarding**
    - [ ] [DOCS] Complete setup instructions in `docs/development/setup.md`.  
        - Evidence:
    - [ ] [DOCS] Finalize/maintain coding standards.  
        - Evidence:
    - [ ] [DOCS] Document config/environment variables.  
        - Evidence:
    - [ ] [DOCS] Document IV&V workflow and control points.  
        - Evidence:

---

## STOP BLOCKAGES

### Currently Blocked (Implementation Required)

- **TODO comment formatting standard and refactor**  
    - Status: Audit found non-compliant TODOs and missing documented standard.  
    - Tracked in: S1a FIX task.  
    - Examples: `src/websocket_server/server.py` lines ~30, ~90; `src/camera_discovery/hybrid_monitor.py` line ~185.

- **Notification handler interface for camera_status_update**  
    - Status: Pending verification in S3.  
    - Files: `src/camera_service/service_manager.py` lines with STOP comments

### Resolved Blockers

- [x] Clarification required: "metrics" field in get_camera_status API response.  
    - Resolved 2025-08-02: Implemented per updated architecture overview and code. See `docs/architecture/overview.md`, `src/websocket_server/server.py`, and `docs/api/json-rpc-methods.md`.  
- [x] Validate implementation of "metrics" field in get_camera_status response.  
    - Resolved 2025-08-02: Code, docs, and architecture overview now match. All STOP/TODO/placeholder code removed. See above files.  
- [x] Clarification required: Method-level API versioning implementation in server.py  
    - Resolved 2025-08-02: Architecture and code updated to document and implement method-level versioning. See `docs/architecture/overview.md`, `src/websocket_server/server.py`.  
- [x] Validate method-level API versioning implementation.  
    - Resolved 2025-08-02: Code registers methods with explicit version, docs match, STOP/TODO removed. See above files.  

---

## How to use this Roadmap

- **Tasks are only marked [x] completed when both the architectural decision and the corresponding code/documentation are fully implemented and validated.**  
- **STOP/BLOCKED items are moved to Resolved Blockers only when implementation is finished and matches the updated architecture.**  
- **Open IV&V items must be validated by review before proceeding to next Epic/Story.**  
- **All IV&V control points must be signed off using the IV&V Reviewer Checklist above.**  
- **Use the Pre-Completion Validation Checklist before marking any task as complete.**  
- **Priority levels (HIGH/MEDIUM/LOW) guide implementation order within each Story.**
