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

- **E1: Robust Real-Time Camera Service Core**
    
    - **S1a: Architecture Scaffolding (COMPLETED)**
        - [x] [FIX] Implement all missing JSON-RPC method stubs in `server.py` as referenced in API docs.
            - Evidence: `src/websocket_server/server.py` (2025-08-02), `docs/api/json-rpc-methods.md`
            - Status: Architecture scaffolding complete (methods exist with proper signatures)
        - [x] [FIX] Correct all parameter typos in `api/json-rpc-methods.md`.
            - Evidence: `docs/api/json-rpc-methods.md` (2025-08-02)
        - [x] [IMPL] Add notification handler stubs in `server.py`.
            - Evidence: `src/websocket_server/server.py` (2025-08-02)
            - Status: Method signatures exist (business logic pending in S1b)
        - [x] [FIX] Standardize TODO comment formatting across codebase.
            - Evidence: All TODOs now follow `docs/development/principles.md` (2025-08-02)
        - [x] [IMPL] Add method-level API versioning stubs in `server.py`.
            - Evidence: `src/websocket_server/server.py` and `docs/architecture/overview.md` (2025-08-02)
        - [x] [IMPL] **HIGH PRIORITY**: Replace NotImplementedError with actual business logic in JSON-RPC methods
            - Evidence: `src/websocket_server/server.py` lines 380-550 (2025-08-02)
            - Status: Complete MediaMTX integration with error handling and logging
        - [x] [IMPL] **HIGH PRIORITY**: Replace `pass` statements with proper notification broadcasting logic  
            - Evidence: `src/websocket_server/server.py` lines 500-600 (2025-08-02)
            - Status: Complete JSON-RPC 2.0 notification system with client broadcasting
        - [x] [IMPL] **MEDIUM PRIORITY**: Actually integrate correlation ID in WebSocket logging (not TODO comments)
            - File: `src/websocket_server/server.py`
            - Fix: Remove TODO comments, implement actual correlation ID propagation in logs
            - Integration: Use existing `logging_config.py` CorrelationIdFilter
            - Evidence: File: `src/websocket_server/server.py` lines  10, 290-310, 200-250 (Date: 2025-08-02)
        - [x] [IMPL] **MEDIUM PRIORITY**: Actually refactor hard-coded values in hybrid_monitor.py
            - File: `src/camera_discovery/hybrid_monitor.py`
            - Fix: Remove hard-coded "camera0" return (line ~200) and placeholder CameraDevice values (line ~185)
            - Fix: Implement configurable stream naming and proper device detection
            - Evidence: File: src/camera_discovery/hybrid_monitor.py
                Sections:
                Lines ~270-290: get_stream_name_from_device_path - Implemented regex-based device number extraction replacing hard-coded "camera0"
                Lines ~185-220: _discover_cameras - Replaced hard-coded CameraDevice creation with proper device detection using device numbers and status determination
                Lines ~225-245: _determine_device_status - New method for proper device status determination
                Lines ~310-335: _probe_device_capabilities - Enhanced capability detection logic
                Date: 2025-08-02 Commit: Refactored all hard-coded values with working device detection logic per architecture requirements
        - [x] [IMPL] **MEDIUM PRIORITY**: Complete MediaMTX controller initialization in service_manager.py
            - Evidence: `src/camera_service/service_manager.py` lines 150-190 (2025-08-02)
            - Status: Complete MediaMTX controller startup with health verification, directory setup, and error handling
        - [x] [IMPL] **MEDIUM PRIORITY**: Complete camera monitor initialization in service_manager.py  
            - Evidence: `src/camera_service/service_manager.py` lines 170-200 (2025-08-02)
            - Status: Complete camera discovery monitor startup with HybridCameraMonitor, event handler registration, and error handling
        - [x] [IVV] **LOW PRIORITY**: Document and validate logging infrastructure implementation
            - Evidence: `src/camera_service/logging_config.py` (275 lines, 2025-08-02)
            - Status: Complete structured logging system with CorrelationIdFilter, JsonFormatter, ConsoleFormatter, and setup_logging function
            - Validation: All architectural requirements met - JSON production logging, human-readable development format, correlation ID propagation, configurable levels, file rotation support
            - Components: CorrelationIdFilter (lines 23-57), JsonFormatter (lines 60-105), ConsoleFormatter (lines 108-135), setup_logging (lines 138-222), helper functions (lines 225-275)
        - [x] [IVV] **LOW PRIORITY**: Document and validate configuration system implementation
            - Evidence: `src/camera_service/config.py` lines 1-470 (2025-08-02)
            - Status: Complete configuration system with all architectural requirements implemented

        - [x] [IMPL] **MEDIUM PRIORITY**: Implement environment variable overrides for configuration
            - Evidence: `src/camera_service/config.py` lines 150-230 (2025-08-02)
            - Status: Complete environment variable override system with proper type conversion and hierarchy

        - [x] [IMPL] **MEDIUM PRIORITY**: Implement configuration schema validation
            - Evidence: `src/camera_service/config.py` lines 290-380 (2025-08-02)
            - Status: Complete JSON Schema validation with comprehensive schema and fallback validation

        - [x] [IMPL] **LOW PRIORITY**: Implement runtime configuration updates
            - Evidence: `src/camera_service/config.py` lines 100-140 (2025-08-02)
            - Status: Complete runtime update system with validation, rollback, and change notifications

        - [x] [IMPL] **LOW PRIORITY**: Implement configuration hot reload capability
            - Evidence: `src/camera_service/config.py` lines 140-190 (2025-08-02)
            - Status: Complete hot reload system with file monitoring and automatic reload capability
        - [x] [IVV] **LOW PRIORITY**: Document and validate logging infrastructure implementation
            - Evidence: `src/camera_service/logging_config.py` (275 lines, 2025-08-02)
            - Status: Complete structured logging system with CorrelationIdFilter, JsonFormatter, ConsoleFormatter, and setup_logging function
            - Validation: All architectural requirements met - JSON production logging, human-readable development format, correlation ID propagation, configurable levels, file rotation support
            - Components: CorrelationIdFilter (lines 23-57), JsonFormatter (lines 60-105), ConsoleFormatter (lines 108-135), setup_logging (lines 138-222), helper functions (lines 225-275)

    - **S1b: Core Implementation (PENDING)**

### S1b: Core Implementation (PENDING)

        - [ ] [FIX] **CRITICAL**: TODO formatting standard in principles.md
            - File: docs/development/principles.md
            - Subtasks:
                - [ ] Add section specifying the standard TODO comment format (e.g., "# TODO: <priority>: <description> [IV&V|StoryRef]")
                - [ ] Remove completion claim in roadmap.md if standard is not specified
            - Evidence:

        - [ ] [FIX] **CRITICAL**: Document parameter typo corrections in API docs
            - File: docs/api/json-rpc-methods.md
            - Subtasks:
                - [ ] List all specific parameter typos fixed (before/after)
                - [ ] Remove completion claim in roadmap.md if none were found
            - Evidence:

        - [ ] [IMPL] **HIGH PRIORITY**: Complete MediaMTX Controller business logic (grouped)
            - File: src/mediamtx_wrapper/controller.py
            - Subtasks:
                - [ ] Implement health_check() with real status (no "unknown" placeholder)
                - [ ] Implement create_stream() and delete_stream() (replace placeholders with working REST calls)
                - [ ] Implement start_recording(), stop_recording(), take_snapshot()
                - [ ] Remove all remaining TODO comments
            - Evidence:

        - [ ] [IMPL] **HIGH PRIORITY**: Complete Service Manager core logic (grouped)
            - File: src/camera_service/service_manager.py
            - Subtasks:
                - [ ] Implement _start_websocket_server (replace pass/TODO)
                - [ ] Implement _start_health_monitor (replace pass/TODO)
                - [ ] Implement all _stop_* methods (replace placeholders)
                - [ ] Implement camera event handler integration (_handle_camera_connected, _handle_camera_disconnected, etc.; remove STOPPED/TODOs)
            - Evidence:

        - [ ] [IMPL] **MEDIUM PRIORITY**: Complete udev monitoring implementation (grouped)
            - File: src/camera_discovery/hybrid_monitor.py
            - Subtasks:
                - [ ] Implement _setup_udev_monitoring() (replace placeholder with actual udev logic)
                - [ ] Implement _udev_event_loop() (remove sleep-only loop, implement full event handling)
            - Evidence:

        - [ ] [FIX] **HIGH PRIORITY**: Clean up roadmap section S1b
            - File: docs/roadmap.md
            - Subtasks:
                - [ ] Remove duplicate/copy-paste entries in S1b
                - [ ] Clarify actual requirements
            - Evidence:

        - [ ] [IVV] **CRITICAL**: Add roadmap tracking for MediaMTX Controller implementation (phantom tracking)
            - File: docs/roadmap.md (Epic/Story tracking)
            - Subtasks:
                - [ ] Create or update roadmap.md to reflect all work in controller.py (REST API client, stream management, health monitoring)
                - [ ] Ensure all features in code have a corresponding roadmap item
            - Evidence:

        - [ ] [IVV] **MEDIUM PRIORITY**: Add roadmap tracking for missing core components
            - Files: main.py, requirements.txt, deployment/scripts/install.sh
            - Subtasks:
                - [ ] Review each implemented file/component for roadmap coverage
                - [ ] Ensure all implemented code has a corresponding roadmap.md task or item
            - Evidence:


    - **S2: Architecture Compliance IV&V (Control Point) - PENDING**
        - [ ] [IVV] Re-validate API docs and code alignment after S1b fixes
            - Task: Verify all JSON-RPC methods have working implementations (not NotImplementedError)
            - Task: Verify notification methods broadcast properly (not `pass` statements)
            - Evidence:
        - [ ] [IVV] Re-validate all issue resolution after actual implementation
            - Task: Verify hard-coded values are refactored (not just TODO comments)
            - Task: Verify correlation ID integration is working (not just TODO comments)
            - Evidence:
        - [ ] [IVV] Validate phantom implementation documentation is complete
            - Task: Ensure logging and configuration systems have proper roadmap tracking
            - Evidence:
        - [ ] [IVV] All stubs/modules correspond to architecture.
            - Evidence: `docs/architecture/overview.md`, codebase validation required
        - [ ] [IVV] No accidental scope/feature creep in implementation.
            - Evidence: Codebase review required after S1b completion
        - [ ] [IVV] Coding standards and docstrings are present.
            - Evidence: Codebase and `docs/development/principles.md` validation required

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

- **Hard-coded CameraDevice values in hybrid_monitor.py**
    - Status: TODO/STOP added, but actual refactoring still needed
    - Tracked in: S1b implementation tasks
    - Files: `src/camera_discovery/hybrid_monitor.py` lines ~185, ~200

- **Correlation ID integration in WebSocket logging**  
    - Status: TODO/STOP added, but actual integration still needed
    - Tracked in: S1b implementation tasks
    - Files: `src/websocket_server/server.py` lines ~290-310

- **Notification handler interface for camera_status_update**
    - Status: TODO/STOP added, but actual implementation still needed
    - Tracked in: S3 implementation tasks
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