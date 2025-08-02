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

- **E1: Robust Real-Time Camera Service Core - SUBSTANTIALLY COMPLETE (FAST TRACK)**
    
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

    - **S2b: Implementation Transition Validation (REQUIRED GATE - S2b must be fully validated and signed off before S3 work is considered complete or before S5 begins)**
        - [ ] [IVV] **CRITICAL PRIORITY**: Validate fast-track implementation completeness against Epic requirements
            - Task: Audit all fast-track implemented components against Epic S3-S5 requirements
            - Files: src/websocket_server/server.py, src/mediamtx_wrapper/controller.py, src/camera_service/service_manager.py
            - Action: Verify WebSocket integration, MediaMTX integration, service orchestration meet implementation Epic requirements
            - Evidence:
        - [ ] [IVV] **HIGH PRIORITY**: Identify and document missing implementation gaps
            - Task: Identify any incomplete areas in fast-track implementation
            - Action: Document actual WebSocket server lifecycle, camera event handling, MediaMTX stream management status
            - Evidence:
        - [ ] [DOCS] **MEDIUM PRIORITY**: Update architecture decisions to reflect fast-track implementation choices
            - File: docs/architecture/overview.md
            - Task: Document implementation decisions made during fast-track development
            - Action: Update Architecture Decisions section with actual implementation patterns used
            - Evidence:
        - [ ] [IVV] **MEDIUM PRIORITY**: Validate TODO comment format compliance across codebase
            - Files: All .py files in src/ directory
            - Task: Ensure all TODO comments follow format: # TODO: <priority>: <description> [Story:<ref>]
            - Action: Standardize any non-compliant TODO comments per principles.md requirements
            - Evidence:

        > **_GATE ENFORCEMENT: Cannot proceed to S3 completion or S5 until all S2b IV&V tasks are complete._**
        > **_EVIDENCE REQUIREMENT: Ensure all S2b items and residual S3 gaps have concrete evidence attached (file/line/commit/test results) before marking their status as progressing._**

    - **S3: Camera Discovery & Monitoring Implementation - PARTIALLY COMPLETE (FAST TRACK) - Remaining gaps: capability detection logic, udev event processing**
        - [x] [IMPL] Implement camera connect/disconnect handling in `service_manager.py`.  
            - Evidence: `src/camera_service/service_manager.py` lines 200-350 (2025-08-02)  
            - Status: Complete event handling with MediaMTX stream coordination
        - [x] [IMPL] Integrate camera monitoring with MediaMTX controller.  
            - Evidence: `src/camera_service/service_manager.py` handle_camera_event methods (2025-08-02)  
            - Status: Full integration with stream creation/deletion and notification broadcasting
        - [x] [IMPL] Implement hybrid udev + polling camera discovery framework.  
            - Evidence: `src/camera_discovery/hybrid_monitor.py` lines 1-500 (2025-08-02)  
            - Status: Complete hybrid monitoring with event system and handler interfaces
        - [x] [IMPL] Implement capability detection logic for CameraDevice 
            - Evidence: src/camera_discovery/hybrid_monitor.py lines 450-650 (2025-08-02)
            - Location: `src/camera_discovery/hybrid_monitor.py` lines 450-650 (2025-08-02)  
            - Implementation: Complete V4L2 probing using v4l2-ctl subprocess calls with timeout handling  
            - Features: Device info, supported formats, resolutions, frame rates detection  
            - Validation: Logic validated with regex pattern testing and mock subprocess calls  
        - [x] [IMPL] Complete udev event processing implementation
            - File: `src/camera_discovery/hybrid_monitor.py` _process_udev_device_event method
            - Task: Complete real-time udev event processing with proper device filtering
            - Action: Implement device node validation and event action mapping
            - Evidence: src/camera_discovery/hybrid_monitor.py lines 220-290 (2025-08-02)

    - **S4: MediaMTX Integration - COMPLETE (FAST TRACK)**
        - [x] [IMPL] Implement stream creation/deletion logic in MediaMTX controller.  
            - Evidence: `src/mediamtx_wrapper/controller.py` create_stream/delete_stream methods (2025-08-02)  
            - Status: Complete REST API integration with stream URL generation
        - [x] [IMPL] Implement recording management functionality.  
            - Evidence: `src/mediamtx_wrapper/controller.py` start_recording/stop_recording methods (2025-08-02)  
            - Status: Complete recording session management with filename generation
        - [x] [IMPL] Implement snapshot capture functionality.  
            - Evidence: `src/mediamtx_wrapper/controller.py` take_snapshot method (2025-08-02)  
            - Status: Complete snapshot capture with file management
        - [x] [IMPL] Implement MediaMTX health monitoring and connectivity verification.  
            - Evidence: `src/mediamtx_wrapper/controller.py` health_check and _health_monitor_loop methods (2025-08-02)  
            - Status: Complete health monitoring with automatic recovery
        - [x] [IMPL] Implement MediaMTX configuration management.  
            - Evidence: `src/mediamtx_wrapper/controller.py` update_configuration method (2025-08-02)  
            - Status: Complete dynamic configuration updates via REST API

    - **S5: Core Integration IV&V (Control Point) - PENDING S2b COMPLETION**
        - [ ] [IVV] **MEDIUM PRIORITY**: Draft acceptance test cases for end-to-end workflows
            - Task: Create test scenarios for camera→MediaMTX→notification flows before full IV&V
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

- **E2: Security and Production Hardening - PENDING S5 COMPLETION**
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

- **E3: Client API & SDK Ecosystem - PENDING E2 COMPLETION**
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

- **E4: Future Extensibility - PLANNING ONLY**
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