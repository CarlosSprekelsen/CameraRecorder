# Project Roadmap

## 🌍 Epics (Long-Term Goals)
- **E1: Robust Real-Time Camera Service Core**
    - Deliver a stable, production-ready camera monitoring and control backend fully aligned with the approved architecture.
- **E2: Security and Production Hardening**
    - Complete security features, error recovery, health checks, and deployment automation for a professional release.
- **E3: Client API & SDK Ecosystem**
    - Provide reliable, well-documented client APIs/examples for easy third-party integration.
- **E4: Future Extensibility**
    - Lay groundwork for cloud integration, advanced protocols, and pluggable camera sources.

- **E5: Deployment & Operations Strategy**
    - Story: Review and confirm system integration (systemd vs others)
    - Story: Decide on user permission model (dedicated vs existing)
    - Story: Validate and document production file system layout
    - Story: Define update and rollback strategy
    - Story: Establish backup & recovery procedures

---

## 📈 Stories (Major Features/Phases)
- **S1: Complete Architecture Compliance**
    - Ensure every module, config, and doc matches the approved architecture and principles.
- **S2: Camera Discovery & Monitoring Implementation**
    - Implement hybrid (udev + polling) camera discovery and monitoring logic.
- **S3: MediaMTX Integration**
    - Implement and verify REST API integration with MediaMTX for stream, record, and snapshot control.
- **S4: WebSocket JSON-RPC API**
    - Develop the JSON-RPC server, core methods, and client connection handling.
- **S5: Automated Testing & Continuous Integration**
    - Achieve >80% unit/integration test coverage with CI setup and test maintenance.
- **S6: Documentation & Developer Onboarding**
    - Complete usage, API, and contribution docs for new contributors.

---

## 📝 Tasks (Immediate/Short-Term)

### 💻 Core Implementation (Priority: P1)

- [ ] **[IMPL]** MediaMTXController stub: Create `src/mediamtx_wrapper/controller.py` to define an async client/controller class for managing MediaMTX streams, recording, and health (stub only, with docstrings and TODOs).
- [ ] **[IMPL]** MediaMTX config template: Add a starter YAML config file for MediaMTX in `config/mediamtx/templates/`, parameterized for your service’s needs.
- [ ] **[IMPL]** Integrate camera discovery with MediaMTXController: Update `ServiceManager` and stubs so that camera connect/disconnect events will (eventually) trigger stream config updates via MediaMTXController (stub the integration point only).
- [ ] **[IMPL]** WebSocket JSON-RPC server scaffold: Create a stub server in `src/websocket_server/server.py` with class, methods, and event handler structure per architecture (no business logic yet).
- [ ] **[IMPL]** Define basic JSON-RPC method specs: Add minimal stub methods for `ping`, `get_camera_list`, etc. in the server scaffold.
- [ ] **[IMPL]** API documentation stubs: For any new public API methods, add their names and parameters to `docs/api/json-rpc-methods.md` (as "not yet implemented" if so).

---

### 🟢 (Optional, parallel) – Developer Experience/CI

- [ ] **[DEV]** Add/validate test stubs for all new modules (unit test skeletons, “test import” at minimum).
- [ ] **[DEV]** Update or add pre-commit hooks, linter configs, or CI workflow files for new paths/modules if not already present.


### 📚 Documentation (Priority: P2)
- [ ] **[DOCS]** Complete setup instructions in setup.md

---

## 📅 How to use this Roadmap

- **Short-term work goes under Tasks:** keep this list actionable and up-to-date; use it for day-to-day and weekly planning.
- **Stories group related tasks into deliverable features or major work areas.**
- **Epics are for tracking broad, multi-milestone project goals.**
- **Review and update this file regularly. All contributors and AI must reference and update only this file for work status.**

