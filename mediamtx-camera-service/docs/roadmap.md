# Project Roadmap

## 🌍 Epics (Long-Term Goals)

- **E1: Robust Real-Time Camera Service Core**
    - **S1: Complete Architecture Compliance**
        - [ ] [FIX] Implement all missing JSON-RPC method stubs in `server.py` as referenced in API docs.
        - [ ] [FIX] Correct all parameter typos in `api/json-rpc-methods.md`.
        - [ ] [IMPL] Add notification handler stubs in `server.py`.
        - [ ] [FIX] Refactor all hard-coded values (e.g., `hybrid_monitor.py`).
        - [ ] [FIX] Standardize TODO comment formatting across codebase.
        - [ ] [IMPL] Integrate correlation ID in WebSocket request logging.
        - [ ] [IMPL] Add method-level API versioning stubs in `server.py`.

    - **S2: Architecture Compliance IV&V (Control Point)**
        - [ ] [IVV] API docs and code stubs match (method names, params, status).
        - [ ] [IVV] All stubs/modules correspond to architecture.
        - [ ] [IVV] No accidental scope/feature creep in stubs.
        - [ ] [IVV] Coding standards and docstrings are present.
        - [ ] [IVV] All CRITICAL/MEDIUM issues from IV&V resolved.
        - **_Cannot proceed to S3 until all S2 IV&V tasks are complete._**

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

- [ ] [BLOCKED] Clarification required: “metrics” field in get_camera_status API response.
      - Context: While implementing get_camera_status (src/websocket_server/server.py), found that API doc (docs/api/json-rpc-methods.md) defines a "metrics" field (bytes_sent, readers, uptime), but this is NOT present or permitted per docs/architecture/overview.md.
      - Action: STOPPED implementation and flagged with a rich TODO in code as required by docs/development/principles.md.
      - Rationale: Per principles.md, do not implement undocumented features; require project architect decision whether "metrics" should be included in camera status response.
      - IV&V Reference: Architecture Compliance IV&V - S2
      - Needed: Update to overview.md or project owner guidance to resolve.

- [ ] [IVV] Validate implementation of "metrics" field in get_camera_status response:
    - Confirm code in `src/websocket_server/server.py` returns metrics field with real values.
    - Confirm API docs and architecture overview match implementation.
    - Confirm all STOP/TODO/placeholder code is removed.
    - Confirm tests (if present) cover metrics field.
    - Reference: Architecture Decision 2025-08-02, IV&V Story S1.

- [ ] [BLOCKED] Clarification required: Hard-coded CameraDevice values in hybrid_monitor.py
    - Context: src/camera_discovery/hybrid_monitor.py, lines 322–328
    - Action taken: Implementation stopped, rich TODO added to document intentional stub and block further development until v4l2 capability detection is integrated.
    - Rationale: Architecture overview requires real capability detection; hard-coded values violate production requirements.
    - IV&V reference: Story S1, Epic E1
    - Needed: Confirmation of capability detection implementation plan and removal of hard-coded placeholders.

- [ ] [IVV] Validate removal of hard-coded CameraDevice values and implementation of capability detection:
    - Confirm code in `src/camera_discovery/hybrid_monitor.py` uses real capability detection.
    - Confirm architecture overview and code match.
    - Confirm all STOP/TODO/placeholder code is removed.
    - Confirm tests (if present) cover capability detection.
    - Reference: Architecture Decision 2025-08-02, IV&V Story S1.

- [ ] [BLOCKED] Clarification required: Method-level API versioning implementation in server.py
    - Context: src/websocket_server/server.py, class WebSocketJsonRpcServer, method registration/versioning stubs
    - Action taken: Implementation stopped, rich TODO added for version negotiation and migration logic.
    - Rationale: Architecture overview requires method-level versioning, but details for negotiation and migration are not documented.
    - IV&V reference: Architecture Decisions v6, API Versioning Strategy, Story S1
    - Needed: Explicit documentation of versioning requirements, negotiation, and migration strategy.

- [ ] [IVV] Validate method-level API versioning implementation:
    - Confirm code in `src/websocket_server/server.py` registers methods with explicit version.
    - Confirm architecture overview and code match.
    - Confirm all STOP/TODO/placeholder code is removed.
    - Confirm tests (if present) cover versioning logic.
    - Reference: Architecture Decision 2025-08-02, IV&V Story S1.

- [ ] [BLOCKED] Clarification required: Correlation ID integration in WebSocket logging
    - Context: src/websocket_server/server.py, all message handling and notification logging methods
    - Action taken: Implementation stopped, rich TODO added for correlation ID extraction and propagation in logs.
    - Rationale: Architecture overview requires structured logging with correlation IDs, but correlation strategy and propagation details are not documented.
    - IV&V reference: Architecture Decisions v6, Logging Format, Story S1
    - Needed: Explicit documentation of correlation ID format, propagation, and integration with logging system.

- [ ] [IVV] Validate correlation ID integration in WebSocket logging:
    - Confirm code in `src/websocket_server/server.py` includes correlation ID in all relevant logs.
    - Confirm architecture overview and code match.
    - Confirm all STOP/TODO/placeholder code is removed.
    - Confirm tests (if present) cover logging logic.
    - Reference: Architecture Decision 2025-08-02, IV&V Story S1.

- [ ] [BLOCKED] Clarification required: Notification handler interface for camera_status_update in service_manager.py
    - Context: src/camera_service/service_manager.py, _handle_camera_connected/_disconnected/_status_changed methods
    - Action taken: Implementation stopped, rich TODO added to document required notification integration.
    - Rationale: Architecture overview requires camera_status_update notifications, but notification handler interface is not finalized.
    - IV&V reference: Story S1, Epic E1
    - Needed: Finalized notification handler interface and integration guidance.

- [ ] [IVV] Validate notification handler interface and integration for camera_status_update:
    - Confirm code in `src/camera_service/service_manager.py` implements notification handler as per architecture.
    - Confirm architecture overview and code match.
    - Confirm all STOP/TODO/placeholder code is removed.
    - Confirm tests (if present) cover notification logic.
    - Reference: Architecture Decision 2025-08-02, IV&V Story S1.

**How to use this Roadmap:**
- **Cannot move to the next Story/Epic until the preceding IV&V Story is fully checked off.**
- **Control points are explicit workflow blockers, ensuring quality before progress.**
