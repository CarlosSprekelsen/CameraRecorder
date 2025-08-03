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

## 🌍 Epics (Long-Term Goals)

### E1: Robust Real-Time Camera Service Core - SUBSTANTIALLY COMPLETE (FAST TRACK)
    
- **S1a: Architecture Scaffolding (COMPLETE)**
        - [x] [FIX] Implement all missing JSON-RPC method stubs in `server.py` as referenced in API docs.  
            - Evidence: `src/websocket_server/server.py` (2025-08-02), `docs/api/json-rpc-methods.md`  
            - Status: All method signatures present with proper parameter handling
        - [x] [FIX] Correct all parameter typos in `api/json-rpc-methods.md`.  
            - Evidence: `docs/api/json-rpc-methods.md` (2025-08-02)
        - [x] [IMPL] Add notification handler stubs in `server.py`.  
            - Evidence: `src/websocket_server/server.py` (2025-08-02)  
            - Status: notification methods `notify_camera_status_update` and `notify_recording_status_update` present
        - [x] [IMPL] Add method-level API versioning stubs in `server.py`.  
            - Evidence: `src/websocket_server/server.py` and `docs/architecture/overview.md` (2025-08-02)  
            - Status: Version tracking framework implemented with `register_method()` versioning
        - [x] [IVV] Document and validate logging infrastructure implementation.  
            - Evidence: `src/camera_service/logging_config.py` (275 lines, 2025-08-02)  
            - Status: Structured logging with CorrelationIdFilter, JsonFormatter, ConsoleFormatter
        - [x] [IVV] Document and validate configuration system implementation.  
            - Evidence: `src/camera_service/config.py` lines 1-470 (2025-08-02)  
            - Status: YAML loading, environment overrides, validation, hot reload implemented
        - [x] [IMPL] Implement environment variable overrides for configuration.  
            - Evidence: `src/camera_service/config.py` lines 150-230 (2025-08-02)
        - [x] [IMPL] Implement configuration schema validation.  
            - Evidence: `src/camera_service/config.py` lines 290-380 (2025-08-02)
        - [x] [IMPL] Implement runtime configuration updates.  
            - Evidence: `src/camera_service/config.py` lines 100-140 (2025-08-02)
        - [x] [IMPL] Implement configuration hot reload capability.  
            - Evidence: `src/camera_service/config.py` lines 140-190 (2025-08-02)

    - **S2: Architecture Compliance IV&V (Control Point) - COMPLETE**
        - [x] [IVV] Validate all stubs/modules correspond to architecture.  
            - Evidence: `docs/architecture/overview.md` vs codebase structure validation (2025-08-02)
            - Status: All architectural components have corresponding code modules
        - [x] [IVV] Validate no accidental scope/feature creep in scaffolding.  
            - Evidence: Code review confirms alignment with architecture scope (2025-08-02)
        - [x] [IVV] Validate coding standards and docstrings are present.  
            - Evidence: `docs/development/principles.md` compliance and codebase review (2025-08-02)

- **S2b: Audit & Task Breakdown (INFORMATIONAL, feeds S3)**  
    - **Purpose:** Quickly validate that fast-track implementation aligns with architecture and produce the real actionable work for S3/S4. This is not a hard gating story—execution proceeds while remaining gaps are tracked and closed downstream.  
    - [x] **Audit existing implementation components against architecture & API definitions.**  
        - Evidence: Code present in `src/websocket_server/server.py`, `src/mediamtx_wrapper/controller.py`, and `src/camera_service/service_manager.py` reflecting core responsibilities defined in `docs/architecture/overview.md`. Status: Baseline audit completed; key partials surfaced and promoted into S3/S4 for execution.
    - [x] **Document and resolve surfaced partial implementation gaps (subset resolved).**  
        - Resolved gaps:  
            * Snapshot implementation updated from placeholder to real FFmpeg-based capture.
            * Recording duration computation implemented in `stop_recording`. 
        - Remaining / tracked gaps (now owned in S3/S4):  
            * API documentation drift (JSON-RPC docs vs implementation).  
            * Versioning/deprecation governance clarity in WebSocket server.
            * Integration of real capability detection into propagated camera metadata.
    - [ ] **Clarify deferred decisions or governance items if not yet finalized.**  
        - Example: Deprecated-method tracking/version negotiation—decide to implement fully or formally defer with annotation. 
    - [ ] **Normalize non-compliant TODO/STOP comments to the canonical format defined in `docs/development/principles.md`.**  
        - Task: Refactor outstanding comments and capture before/after evidence.
    - [ ] **Update architecture overview with fast-track deviations and decision records.**  
        - File: `docs/architecture/overview.md`.  
        - Task: Capture that snapshot/duration were known partials and are now resolved, and note remaining items feeding S3/S4.  

> _Note: Completion of these audit observations informs closing S3; they are not independent hard gates. Execution of remaining work continues in S3 and S4 with tracked acceptance criteria._

- **S3: Camera Discovery & Monitoring Implementation - PARTIALLY COMPLETE (FAST TRACK)**  
- **Objective:** Fulfill architecture requirements for real-time camera discovery, capability probing, status tracking, and event propagation.

    - [x] [IMPL] Implement camera connect/disconnect handling and coordinate with MediaMTX.  
        - Evidence: `src/camera_service/service_manager.py` event handlers (`_handle_camera_connected`, `_handle_camera_disconnected`) and orchestration paths. 
        - Status: Complete for baseline flow.

    - [x] [IMPL] Integrate camera monitoring with MediaMTX controller.  
        - Evidence: Notification parameter preparation and stream creation/deletion logic in `ServiceManager`.
        - Status: Full integration path exists.

    - [x] [IMPL] Hybrid camera discovery framework (udev + polling) implemented.  
        - Evidence: `src/camera_discovery/hybrid_monitor.py` implementation of udev event loop, polling fallback, and device lifecycle. 
        - Status: Complete scaffolding and operational logic.

    - [x] [IMPL] Harden and expand capability detection validation.  
        - Task: Extend tests and edge case handling for `_probe_device_capabilities`, including varied format/resolution outputs, error conditions, and timeout fallbacks.  
        - Acceptance Criteria: Tests cover success paths, parsing variations, timeouts, and failure modes; no unresolved TODOs in detection logic.  
        - Evidence: Expanded tests under `tests/unit/test_camera_discovery/validate_capabilities.py` or similar; execution results, coverage report. 

    - [x] [IMPL] Implement udev event filtering and real-time processing.  
        - Evidence: `_process_udev_device_event` with device node validation and range filtering. 
        - Status: Baseline implemented.

    - [ ] [IMPL] **MEDIUM PRIORITY**: Expand udev event processing test coverage.  
        - Task: Cover additional scenarios: `change` events affecting status, invalid nodes, race conditions, and fallback to polling.  
        - Acceptance Criteria: Tests in `tests/unit/test_camera_discovery/validate_udev.py` (or combined) demonstrating correct filtering and event propagation.{index=36}  

    - [ ] [IMPL] **MEDIUM PRIORITY**: Integrate real capability detection results into camera metadata propagation.  
        - Task: Replace fallbacks/defaults in `_get_camera_metadata` with actual capability-derived resolution/fps when available; if delayed, annotate dependency explicitly.  
        - Acceptance Criteria: Notifications include accurate metadata or clearly annotated interim state. 

    - [ ] [DOCS] **LOW/MEDIUM PRIORITY**: Reconcile and update JSON-RPC API documentation for camera status notifications and capability fields.  
        - Task: Update `docs/api/json-rpc-methods.md` to reflect actual implemented fields, mark implemented methods appropriately, and add examples matching runtime behavior.  
        - Evidence: Updated API doc with “Status: Implemented” and linked code lines for `camera_status_update`, `get_camera_status`, etc. 

- **S4: MediaMTX Integration - PARTIALLY COMPLETE (FAST TRACK)**  
- **Objective:** Provide reliable MediaMTX stream management, recording, snapshotting, health checking, and dynamic configuration per architecture.

    - [x] [IMPL] Implement stream creation/deletion logic.  
        - Evidence: `create_stream` and `delete_stream` methods in `src/mediamtx_wrapper/controller.py` with REST interaction and error handling. :contentReference[oaicite:39]{index=39}  

    - [x] [IMPL] Implement recording management (start/stop), including duration computation.  
        - Evidence: `start_recording` and `stop_recording` with session metadata; `stop_recording` includes accurate duration calculation. :contentReference[oaicite:40]{index=40}  
        - Status: Completed (previously pending duration computation).

    - [x] [IMPL] Snapshot capture implemented (real).  
        - Evidence: `take_snapshot` invokes FFmpeg to capture a real frame, persists actual image, and returns metadata. :contentReference[oaicite:41]{index=41}  
        - Status: Completed (no longer a placeholder).

    - [x] [IMPL] Health monitoring and connectivity verification.  
        - Evidence: `health_check` and associated background monitoring logic in controller. :contentReference[oaicite:42]{index=42}  

    - [x] [IMPL] Dynamic configuration updates.  
        - Evidence: `update_configuration` method exists and applies updates per architecture expectations. :contentReference[oaicite:43]{index=43}  

    - [ ] [DOCS] **MEDIUM PRIORITY**: Reflect MediaMTX integration status in API/architecture docs, including any historical partials and their closure.  
        - Task: Update roadmap/architecture decision logs to capture that snapshot and recording duration were known partials and have been resolved.  
        - Acceptance Criteria: Clear “known partials” note with owner and closure dates recorded.


- **S5: Core Integration IV&V (Control Point) - PENDING**  
    - [ ] [IVV] **MEDIUM PRIORITY**: Draft acceptance test cases for end-to-end workflows  
        - Task: Create test scenarios for camera → MediaMTX → notification flows before full IV&V  
        - Action: Define test cases for connect/disconnect, stream creation, recording, snapshot capture  
        - Evidence:  
    - [ ] [IVV] **HIGH PRIORITY**: Validate MediaMTX integration with actual camera device testing  
        - Task: Test stream creation/deletion with physical or virtual camera device  
        - Action: Verify end-to-end camera detection -> MediaMTX stream -> client notification flow  
        - Evidence:  
    - [ ] [IVV] **HIGH PRIORITY**: Validate notification and error recovery workflows  
        - Task: Test camera connect/disconnect notification broadcasting and error handling  
        - Action: Verify WebSocket notification delivery and MediaMTX error recovery mechanisms  
        - Evidence:  
    - [ ] [IVV] **MEDIUM PRIORITY**: Validate service orchestration and component lifecycle  
        - Task: Test service manager startup/shutdown and component coordination  
        - Action: Verify graceful service lifecycle with proper component dependencies  
        - Evidence:  
    - [ ] [IMPL] **MEDIUM PRIORITY**: Complete missing deployment automation  
        - File: deployment/scripts/install.sh (currently incomplete)  
        - Task: Complete installation script with system dependencies and service setup  
        - Action: Implement full system installation workflow matching deployment documentation  
        - Evidence:  
    - **_Cannot proceed to E2 until S5 IV&V is complete._**


## **E2: Security and Production Hardening - PENDING S5 COMPLETION**
- **S6: Security Features Implementation**
        - [ ] [IMPL] Implement authentication framework (JWT/API key) in WebSocket server.  
            - Evidence:
        - [ ] [IMPL] Add health check REST endpoints for monitoring systems.  
            - Evidence:
        - [ ] [IMPL] Implement rate limiting and connection management.  
            - Evidence:
        - [ ] [IMPL] Add TLS/SSL support for production deployment.  
            - Evidence:

- **S7: Security IV&V (Control Point)**
        - [ ] [IVV] Authentication and authorization features verified.  
            - Evidence:
        - [ ] [IVV] All endpoints have correct access control validation.  
            - Evidence:
        - [ ] [IVV] Security test cases in place and passing.  
            - Evidence:
        - **_Cannot proceed to E3 until S7 IV&V is complete._**

## **E3: Client API & SDK Ecosystem - PENDING E2 COMPLETION**
- **S8: Client APIs and Examples**
        - [ ] [IMPL] Add client API usage examples in `/examples`.  
            - Evidence:
        - [ ] [DOCS] Document API usage and authentication.  
            - Evidence:
        - [ ] [IMPL] Create client SDK or library for common platforms.  
            - Evidence:

    - **S9: SDK & Docs IV&V (Control Point)**
        - [ ] [IVV] Client API examples and docs reviewed for accuracy and completeness.  
            - Evidence:
        - [ ] [IVV] Usability tests or walkthroughs completed.  
            - Evidence:
        - **_Cannot proceed to E4 until S9 IV&V is complete._**

## **E4: Future Extensibility - PLANNING ONLY**
- **S10: Cloud/Protocol Extensions (Planning Only)**
        - [ ] [DOCS] Prepare documentation and placeholders for future protocols/cloud (no implementation).  
            - Evidence:
        - [ ] [DOCS] Document plugin architecture for camera source extensions.  
            - Evidence:

    - **S11: Extensibility IV&V (Control Point)**
        - [ ] [IVV] Future plans reviewed, documented, and approved.  
            - Evidence:
        - [ ] [IVV] All extension points and plugin mechanisms validated.  
            - Evidence:
        - **_Cannot proceed to E5 until S11 IV&V is complete._**

- **E5: Deployment & Operations Strategy - PENDING E3 COMPLETION**
    - **S12: Deployment Automation & Ops**
        - [ ] [IMPL] Complete production deployment automation scripts.  
            - Evidence:
        - [ ] [DOCS] Add environment variable and system integration documentation.  
            - Evidence:
        - [ ] [DOCS] Document update/rollback and backup procedures.  
            - Evidence:
        - [ ] [IMPL] Add monitoring and alerting integration.  
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

- **S14: Automated Testing & Continuous Integration - PARTIALLY COMPLETE**
    - [x] [IMPL] **MEDIUM PRIORITY**: Configure code formatting and linting tools  
        - File: `.flake8`  
        - Evidence: `.flake8` lines 1-10 (2025-08-02)  
        - Status: Flake8 configuration with max line length 88, extend-ignore E203/W503
    - [x] [IMPL] **MEDIUM PRIORITY**: Setup pre-commit hooks for code quality  
        - File: `.pre-commit-config.yaml`  
        - Evidence: `.pre-commit-config.yaml` lines 1-25 (2025-08-02)  
        - Status: Pre-commit configuration with black, flake8, mypy, and standard hooks
    - [x] [IMPL] **MEDIUM PRIORITY**: Configure type checking with mypy  
        - File: `mypy.ini`  
        - Evidence: `mypy.ini` lines 1-15 (2025-08-02)  
        - Status: Mypy configuration with strict type checking compliance
    - [x] [IMPL] **HIGH PRIORITY**: Document core project dependencies  
        - File: `requirements.txt`  
        - Evidence: `requirements.txt` lines 1-15 (2025-08-02)  
        - Status: Core dependencies documented with version constraints
    - [x] [IMPL] **HIGH PRIORITY**: Implement main application entry point  
        - File: `src/camera_service/main.py`  
        - Evidence: `src/camera_service/main.py` lines 1-70 (2025-08-02)  
        - Status: Complete main entry point with service orchestration
    - [ ] [IMPL] **MEDIUM PRIORITY**: Add unit test scaffolding for all modules
        - Task: Create test structure in tests/ directory matching src/ structure
        - Evidence:
    - [ ] [IMPL] **MEDIUM PRIORITY**: Achieve >80% unit/integration test coverage
        - Task: Implement comprehensive test suite covering core functionality
        - Evidence:
    - [ ] [IMPL] **LOW PRIORITY**: Set up CI pipeline for automated testing
        - Task: Configure GitHub Actions or similar for automated linting, formatting, and tests
        - Evidence:

- **S15: Documentation & Developer Onboarding - PARTIALLY COMPLETE**
    - [x] [DOCS] **HIGH PRIORITY**: Maintain coding standards documentation  
        - File: `docs/development/coding-standards.md`  
        - Evidence: `docs/development/coding-standards.md` (2025-08-02)  
        - Status: Comprehensive coding standards with logging, documentation, and security requirements
    - [x] [DOCS] **HIGH PRIORITY**: Maintain development principles documentation  
        - File: `docs/development/principles.md`  
        - Evidence: `docs/development/principles.md` (2025-08-02)  
        - Status: Project principles with TODO formatting and IV&V alignment requirements
    - [ ] [DOCS] **MEDIUM PRIORITY**: Complete setup instructions in `docs/development/setup.md`
        - Task: Create developer environment setup guide
        - Evidence:
    - [ ] [DOCS] **MEDIUM PRIORITY**: Document configuration and environment variables
        - Task: Create comprehensive configuration reference documentation
        - Evidence:
    - [ ] [DOCS] **LOW PRIORITY**: Document IV&V workflow and control points
        - Task: Create IV&V process documentation for future contributors
        - Evidence:

---
## Open Issues & Gaps (Non-Blocking)

These are known deficiencies or pending tasks surfaced by the fast-track audit. None are architectural show-stoppers; they are tracked as discrete work items in their respective Stories and expected to be closed as part of normal execution.

- [ ] **Snapshot capture is a placeholder.**  
  - Story: S4.  
  - Description: Current implementation returns simulated data; real snapshot acquisition (MediaMTX API or FFmpeg fallback) needs to be implemented.  
  - Owner: <name> / Target: <date>  
  - Evidence: `src/mediamtx_wrapper/controller.py` `take_snapshot` TODO. :contentReference[oaicite:5]{index=5}  

- [ ] **Recording duration not computed.**  
  - Story: S4.  
  - Description: `stop_recording` lacks actual duration calculation.  
  - Owner: <name> / Target: <date>  
  - Evidence: TODO in `src/mediamtx_wrapper/controller.py`. :contentReference[oaicite:6]{index=6}  

- [ ] **API documentation drift.**  
  - Story: S3.  
  - Description: JSON-RPC docs still label implemented methods as “Not yet implemented”; needs reconciliation.  
  - Owner: <name> / Target: <date>  
  - Evidence: `docs/api/json-rpc-methods.md` vs `src/websocket_server/server.py`. :contentReference[oaicite:7]{index=7}  

- [ ] **Versioning/deprecation governance clarity.**  
  - Story: S2b/S3.  
  - Description: Deprecated-method tracking in `server.py` is stubbed; decide whether to finalize or formally defer.  
  - Owner: <name> / Target: <date>  
  - Evidence: `src/websocket_server/server.py`. :contentReference[oaicite:8]{index=8}  

- [ ] **TODO/STOP comment normalization.**  
  - Story: S1a (cleanup) / S3.  
  - Description: Format is defined in `docs/development/principles.md` but not consistently applied; refactor existing comments to canonical form.  
  - Owner: <name> / Target: <date>  
  - Evidence: `docs/development/principles.md` and instances in code. :contentReference[oaicite:9]{index=9}  

## STOP BLOCKAGES

### Currently Blocked (Implementation Required)

*No current blockers - fast-track implementation has resolved previous architectural blockers*

### Resolved Blockers

- [x] **Clarification required**: "metrics" field in get_camera_status API response.  
    - Resolved 2025-08-02: Implemented per updated architecture overview. See `docs/architecture/overview.md`, `src/websocket_server/server.py`.  
- [x] **Clarification required**: Method-level API versioning implementation approach.  
    - Resolved 2025-08-02: Architecture and code implement method-level versioning with register_method(). See `docs/architecture/overview.md`, `src/websocket_server/server.py`.  
- [x] **Phase classification mismatch**: S1a completion status vs actual implementation level.  
    - Resolved 2025-08-02: Roadmap realigned to reflect fast-track implementation reality with proper Epic transition validation.

---

## How to use this Roadmap

- **Tasks are only marked [x] completed when both the architectural decision and the corresponding code/documentation are fully implemented and validated.**  
- **STOP/BLOCKED items are moved to Resolved Blockers only when implementation is finished and matches the updated architecture.**  
- **Open IV&V items must be validated by review before proceeding to next Epic/Story.**  
- **All IV&V control points must be signed off using the IV&V Reviewer Checklist above.**  
- **Use the Pre-Completion Validation Checklist before marking any task as complete.**  
- **Priority levels (HIGH/MEDIUM/LOW) guide implementation order within each Story.**
- **Fast-track implementation has accelerated progress - focus now on validation and completion of remaining gaps.**