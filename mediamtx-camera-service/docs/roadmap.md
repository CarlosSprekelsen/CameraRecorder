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

**How to use this Roadmap:**
- **Cannot move to the next Story/Epic until the preceding IV&V Story is fully checked off.**
- **Control points are explicit workflow blockers, ensuring quality before progress.**
