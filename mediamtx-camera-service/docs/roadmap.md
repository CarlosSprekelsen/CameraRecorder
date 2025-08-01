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

### 🔧 Architecture Compliance (Priority: P0)
- [X] **[CLEANUP]** Remove outdated TODOs from all code files per audit
- [X] **[CLEANUP]** Remove development checklists from README.md
- [X] **[CLEANUP]** Delete or complete placeholder files in tests/ and examples/

### 🟠 Decisions Needing Immediate Resolution (Priority: P0)
- [X] **[DECISION]** Confirm WebSocket-only API vs hybrid REST+WebSocket
- [X] **[DECISION]** Select authentication strategy (None, JWT, API keys, or client certs)
- [X] **[DECISION]** Choose logging format (structured JSON vs traditional)
- [X] **[DECISION]** Define initial performance targets (latency, throughput)
- [X] **[DECISION]** Set resource limits (memory, CPU, storage)

### 💻 Core Implementation (Priority: P1)  
- [X] **[IMPL]** ServiceManager stub with architecture alignment
- [X] **[IMPL]** Logging configuration in logging_config.py
- [X] **[IMPL]** Camera discovery hybrid approach (udev + polling)

### 📚 Documentation (Priority: P2)
- [ ] **[DOCS]** Complete setup instructions in setup.md
- [ ] **[DOCS]** Finalize coding standards document

---

## 📅 How to use this Roadmap

- **Short-term work goes under Tasks:** keep this list actionable and up-to-date; use it for day-to-day and weekly planning.
- **Stories group related tasks into deliverable features or major work areas.**
- **Epics are for tracking broad, multi-milestone project goals.**
- **Review and update this file regularly. All contributors and AI must reference and update only this file for work status.**

