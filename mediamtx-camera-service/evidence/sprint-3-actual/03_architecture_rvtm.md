# Requirements Verification Traceability Matrix (RVTM) - Sprint 3 CDR

Version: 1.0
Date: 2025-08-09
Prepared by: IV&V Team
Status: For system execution validation

## RVTM Overview
This RVTM maps every requirement from the Phase 2 scope (per `evidence/sprint-3-actual/02_requirements_inventory.md` and `docs/requirements/client-requirements.md` v1.1) to the architecture components defined in `docs/architecture/overview.md`. It provides both forward (Requirements → Architecture) and inverse (Architecture → Requirements) traceability, identifies critical gaps, and assesses adequacy of the current architecture to fulfill each requirement.

Architecture Components (as referenced below):
- WS Server: WebSocket JSON-RPC Server (`src/websocket_server/server.py`)
- SvcMgr: Service Manager orchestrator (`src/camera_service/service_manager.py`)
- MMX Ctrl: MediaMTX Controller (`src/mediamtx_wrapper/controller.py`)
- CamMon: Camera Discovery Monitor (`src/camera_discovery/hybrid_monitor.py`)
- Health: Health & Monitoring (`src/health_server.py` + controller health loop)
- Config: Configuration Management (`src/common/config.py`, `src/camera_service/config.py`)
- Security: Security Model (JWT/API Keys/Middleware) (`src/security/*`)
- Client: Client Applications (Web/Android) as per requirements doc

Traceability Status legend: Allocated, Partially Allocated, Unallocated
Adequacy Assessment legend: Adequate, Partial (needs work), Gap (missing support)

## Requirements → Architecture Mapping

| Requirement ID | Description | Allocated Architecture Component(s) | Traceability Status | Adequacy Assessment |
|---|---|---|---|---|
| F1.1.1 | Take photos using available cameras | WS Server, SvcMgr, MMX Ctrl | Allocated | Partial (business flow works; security enforcement pending) |
| F1.1.2 | Use `take_snapshot` JSON-RPC | WS Server | Allocated | Adequate (method exists; param validation added) |
| F1.1.3 | Display preview of captured photos | Client | Allocated | Partial (client responsibility) |
| F1.1.4 | Graceful photo error handling with feedback | WS Server, Client | Allocated | Partial (server errors standardized; client UX TBD) |
| F1.2.1 | Record videos using available cameras | WS Server, SvcMgr, MMX Ctrl | Allocated | Partial (flows implemented; security enforcement pending) |
| F1.2.2 | Unlimited duration recording mode | WS Server, MMX Ctrl | Allocated | Adequate (no duration auto-stop; stop via API) |
| F1.2.3 | Timed recording (sec/min/hr) with auto-stop | WS Server, MMX Ctrl | Allocated | Adequate (param validation + auto-stop + completion notification) |
| F1.2.4 | Manual stop recording | WS Server, MMX Ctrl | Allocated | Adequate |
| F1.2.5 | Recording session management via API | WS Server, MMX Ctrl | Allocated | Partial (conflict handling improving) |
| F1.3.1 | Auto file rollover on max size | MMX Ctrl | Allocated | Partial (depends on MediaMTX config) |
| F1.3.2 | Display recording status & elapsed time | WS Server, Client | Allocated | Partial (notifications emitted; UI pending) |
| F1.3.3 | Notify when recording completed | WS Server | Allocated | Adequate (completion notifications implemented) |
| F1.3.4 | Visual indicator for active recording | Client | Allocated | Partial (client responsibility) |
| F2.1.1 | Include location metadata | Client | Allocated | Partial (client device capability) |
| F2.1.2 | Include timestamp metadata | Client | Allocated | Adequate (client-controlled) |
| F2.1.3 | Request location permissions | Client | Allocated | Adequate (platform standard) |
| F2.2.1 | Naming format `[datetime]_[id].[ext]` | Client | Allocated | Adequate |
| F2.2.2 | DateTime format `YYYY-MM-DD_HH-MM-SS` | Client | Allocated | Adequate |
| F2.2.3 | Unique ID 6-char alphanumeric | Client | Allocated | Adequate |
| F2.2.4 | Example names (informative) | Client | Allocated | Adequate |
| F2.3.1 | Store media in configurable folder | Client | Allocated | Adequate |
| F2.3.2 | Folder selection interface | Client | Allocated | Adequate |
| F2.3.3 | Validate storage perms/space | Client | Allocated | Adequate |
| F2.3.4 | Default storage per platform | Client | Allocated | Adequate |
| F3.1.1 | Display list of available cameras | WS Server, CamMon, SvcMgr, Client | Allocated | Adequate (API + notifications in place) |
| F3.1.2 | Show camera status | WS Server, CamMon, SvcMgr, Client | Allocated | Adequate |
| F3.1.3 | Hot-plug notifications | WS Server, CamMon, SvcMgr, Client | Allocated | Adequate (multi-client broadcast hardened) |
| F3.1.4 | Camera switching interface | Client | Allocated | Adequate |
| F3.2.1 | Recording start/stop controls | Client, WS Server | Allocated | Adequate |
| F3.2.2 | Duration selector interface | Client | Allocated | Adequate |
| F3.2.3 | Show recording progress/elapsed | Client, WS Server | Allocated | Partial (UI pending) |
| F3.2.4 | Emergency stop | WS Server, MMX Ctrl | Allocated | Adequate |
| F3.2.5 | Operator role required for protected methods | Security, WS Server | Partially Allocated | Partial (server enforcement toggle pending full enablement) |
| F3.2.6 | Handle token expiration with re-auth | Security, WS Server, Client | Partially Allocated | Partial (token validation implemented; expiry handling tests pending) |
| F3.3.1 | Settings interface | Client | Allocated | Adequate |
| F3.3.2 | Validate and persist settings | Client | Allocated | Adequate |
| F3.3.3 | Reset settings to defaults | Client | Allocated | Adequate |
| W1.1 | Browser compatibility | Client | Allocated | Adequate |
| W1.2 | Responsive design | Client | Allocated | Adequate |
| W1.3 | PWA capabilities | Client | Allocated | Adequate |
| W1.4 | WebRTC preview when supported | Client | Allocated | Adequate |
| W2.1 | Browser download integration | Client | Allocated | Adequate |
| W2.2 | File naming preservation | Client | Allocated | Adequate |
| W2.3 | Large file download progress | Client | Allocated | Adequate |
| A1.1 | Android min API level 28 | Client | Allocated | Adequate |
| A1.2 | Target API level 34 | Client | Allocated | Adequate |
| A1.3 | Camera/audio permissions | Client | Allocated | Adequate |
| A1.4 | Storage permissions | Client | Allocated | Adequate |
| A1.5 | Location permissions | Client | Allocated | Adequate |
| A2.1 | MediaStore integration | Client | Allocated | Adequate |
| A2.2 | Background recording with foreground service | Client | Allocated | Partial (device instrumentation required) |
| A2.3 | Android notifications for recording | Client | Allocated | Adequate |
| A2.4 | Battery optimization exclusion guidance | Client | Allocated | Adequate |
| N1.1 | Startup <3s | WS Server, SvcMgr | Allocated | Partial (perf testing pending) |
| N1.2 | Camera list refresh <1s | WS Server, CamMon | Allocated | Partial (perf testing pending) |
| N1.3 | Photo capture resp <2s | WS Server, MMX Ctrl | Allocated | Partial (perf testing pending) |
| N1.4 | Recording start <2s | WS Server, MMX Ctrl | Allocated | Partial (perf testing pending) |
| N1.5 | UI feedback ≤200ms | Client | Allocated | Partial (UX perf testing pending) |
| N2.1 | Graceful disconnection handling | WS Server, SvcMgr | Allocated | Adequate |
| N2.2 | Auto reconnection with backoff | WS Server | Allocated | Adequate (retry/backoff patterns) |
| N2.3 | Preserve recording state across disconnects | WS Server, MMX Ctrl | Allocated | Partial (session state tests pending) |
| N2.4 | Validate all inputs and responses | WS Server, Config | Allocated | Adequate (param validation added; schema checks) |
| N3.1 | WSS in production | WS Server, Security, Config | Allocated | Partial (deployment config validation pending) |
| N3.2 | Validate JWT tokens & handle expiration | Security, WS Server | Allocated | Partial (expiry handling tests pending) |
| N3.3 | No plaintext sensitive credentials | Security, Config | Allocated | Adequate (hashed keys, config hygiene) |
| N3.4 | Inactive session timeout | WS Server, Security | Allocated | Partial (idle timeout policy to validate) |
| N4.1 | Clear error messages & recovery guidance | WS Server, Client | Allocated | Adequate (JSON-RPC taxonomy; client UX pending) |
| N4.2 | Consistent UI patterns | Client | Allocated | Adequate |
| N4.3 | Accessibility support | Client | Allocated | Partial (audit pending) |
| N4.4 | Offline mode limited functionality | Client | Allocated | Partial (design dependent) |
| I1.1 | WebSocket JSON-RPC 2.0 protocol | WS Server | Allocated | Adequate |
| I1.2 | Endpoint `ws://host:8002/ws` | WS Server, Config | Allocated | Adequate |
| I1.3 | Real-time notifications | WS Server, CamMon, SvcMgr | Allocated | Adequate |
| I1.4 | Supported method set | WS Server | Allocated | Adequate (ping, list, status, snapshot, start/stop) |
| I1.5 | JSON-RPC error taxonomy | WS Server | Allocated | Adequate (uses -32602/-32003 etc.) |
| I1.6 | Heartbeat ping every 30s | WS Server | Allocated | Partial (long-run validation pending) |
| I1.7 | Role-based access via JWT | Security, WS Server | Partially Allocated | Partial (enforcement enablement pending) |

## Architecture → Requirements Mapping (Inverse)

| Architecture Component | Requirements Fulfilled |
|---|---|
| WS Server | F1.1.2, F1.2.1, F1.2.2, F1.2.3, F1.2.4, F1.3.2 (notifications), F1.3.3, F3.1.1, F3.1.2, F3.1.3, F3.2.1, F3.2.3, F3.2.4, I1.1, I1.2, I1.4, I1.5, I1.6, N1.1–N1.4 (perf targets), N2.1–N2.4 (part), N3.1–N3.4 (with Security) |
| SvcMgr | F1.1.1, F1.2.1, F3.1.1–F3.1.3, N2.1, N2.2 |
| MMX Ctrl | F1.1.x, F1.2.x, F1.3.1, F3.2.4, N1.3–N1.4 (perf), N2.3 |
| CamMon | F3.1.1–F3.1.3, N1.2 |
| Health | N2.x suite (recovery, monitoring), I1.3 (indirect via stability) |
| Config | N2.4 (validation), I1.2 (endpoint), N3.1 (TLS/WSS config), N3.3 |
| Security | F3.2.5, F3.2.6, N3.1–N3.4, I1.7 |
| Client | F1.1.3, F1.3.2, F1.3.4, F2.x, F3.1.1–F3.1.4 (UI), F3.2.1–F3.2.3, F3.3.x, W1.x, W2.x, A1.x, A2.x, N1.5, N4.x |

## Gap Analysis

- Critical Gaps (Requirements without architecture allocation)
  - None identified at the architectural allocation level. All listed requirements have at least one responsible component.

- Partial/Needs Work (Risk to CDR if not completed)
  - F3.2.5 / I1.7 / N3.2 / N3.4: Security enforcement and session timeout require full enablement in WS Server and validation tests (token authentication method and protected-method checks must be enforced by default).
  - N1.x: Performance targets require empirical validation and potential tuning (WS Server, MMX Ctrl, CamMon).
  - N2.3: Recording state preservation across disconnects needs targeted validation.
  - I1.6: Heartbeat long-run stability tests pending.
  - A2.2: Android background recording requires device-level instrumentation.
  - N3.1: WSS/TLS production deployment verification pending in deployment docs and CI checks.

- Potential Over-Engineering (Components without clear requirement justification)
  - API Versioning framework (AD-5) present in WS Server; no explicit requirement now, but low risk and useful for evolution.
  - Extensive logging/metrics (AD-8, Health) exceed minimum stated requirements but support N2.x reliability and CDR evidence.

## Architecture Adequacy Conclusion

- Overall Adequacy: The architecture is capable of satisfying the Phase 2 functional and non-functional requirements. Core components are implemented and correctly allocated to requirements. 
- Primary Risks: Security enforcement completeness (F3.2.5/F3.2.6/I1.7/N3.x) and performance targets (N1.x) require final validation and CI gating. 
- Recommendation: Prioritize enabling strict authentication/authorization in WS Server and finalize performance/long-run tests to de-risk CDR.

---

This RVTM provides the required forward and inverse traceability to detect gaps and demonstrate architectural completeness for CDR.
