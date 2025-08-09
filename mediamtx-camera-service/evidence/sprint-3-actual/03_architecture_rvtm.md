# Architecture RVTM (Requirements Verification Traceability Matrix)
**Version:** 1.0  
**Authors:** IV&V Team  
**Date:** 2025-08-09  
**Status:** in review  
**Related Epic/Story:** Sprint 3 CDR

## Purpose/Overview
This document provides a comprehensive Requirements Verification Traceability Matrix (RVTM) for Phase 2 requirements, mapping every requirement from the inventory to the approved architecture components defined in `docs/architecture/overview.md`. It verifies that the architecture is capable of satisfying the requirements, identifies gaps where no component is allocated, highlights components with no requirement justification, and assesses adequacy.

Architecture components referenced (per approved overview):
- WebSocket JSON-RPC Server
- Camera Discovery Monitor
- MediaMTX Controller
- Health & Monitoring
- Security Model
- Configuration Management
- MediaMTX Server (external dependency)
- Client Applications (external to service; noted for completeness where requirements are explicitly client/UI)

Notes on scope:
- Requirements that are purely client/UI or mobile/web platform concerns are marked External (Client Applications). These are not gaps in the service architecture; they are delivered by client-side systems enabled by the service interfaces.
- “Critical Gap” is reserved for requirements that must be fulfilled by the service architecture but currently lack an allocated component or a feasible mechanism.

---

## Requirements → Architecture Mapping (Complete RVTM)
Columns: Requirement ID, Description, Allocated Architecture Component(s), Traceability Status, Adequacy Assessment

### Functional - F1: Camera Interface Requirements

| Requirement ID | Description | Allocated Architecture Component(s) | Traceability Status | Adequacy Assessment |
|---|---|---|---|---|
| F1.1.1 | Allow users to take photos using available cameras | WebSocket JSON-RPC Server; MediaMTX Controller; MediaMTX Server | Allocated | Adequate. Snapshot capability supported via service API and MediaMTX “recording and snapshot generation”. |
| F1.1.2 | Use service `take_snapshot` JSON-RPC method | WebSocket JSON-RPC Server | Allocated | Adequate. Requires API method defined in JSON-RPC catalog and request/response schema. |
| F1.1.3 | Display preview of captured photos | Client Applications (External) | External | Adequate via enablement. Service provides snapshot payload/URL; rendering is client responsibility. |
| F1.1.4 | Graceful photo error handling with user feedback | WebSocket JSON-RPC Server; Client Applications (External) | Partially allocated | Service provides error taxonomy (I1.5). Client displays messages. Adequate with taxonomy coverage. |
| F1.2.1 | Allow users to record videos using available cameras | WebSocket JSON-RPC Server; MediaMTX Controller; MediaMTX Server | Allocated | Adequate. Recording start/stop via API; MediaMTX handles media pipeline. |
| F1.2.2 | Unlimited duration recording mode | MediaMTX Controller; MediaMTX Server | Allocated | Adequate. Omit duration parameter; ensure controller does not set auto-stop. |
| F1.2.3 | Timed recording with seconds/minutes/hours | MediaMTX Controller; WebSocket JSON-RPC Server | Allocated | Adequate. Duration parameter handling and timer-driven stop in controller. |
| F1.2.4 | Manual stop of recording | WebSocket JSON-RPC Server; MediaMTX Controller | Allocated | Adequate. `stop_recording` supported; immediate pipeline halt. |
| F1.2.5 | Recording session management via service API | WebSocket JSON-RPC Server; MediaMTX Controller; Health & Monitoring | Allocated | Adequate. Session state transitions and conflict checks covered. |
| F1.3.1 | Auto file rollover on max size (service) | MediaMTX Controller; MediaMTX Server | Partially allocated | Risk. Requires MediaMTX recording segmentation/rollover by size; confirm feature and expose config. Potential gap if unsupported. |
| F1.3.2 | Display recording status and elapsed time | Health & Monitoring; WebSocket JSON-RPC Server; Client Applications (External) | Partially allocated | Service emits status/elapsed notifications; client renders UI. Adequate with event payloads. |
| F1.3.3 | Notify when recording is completed | WebSocket JSON-RPC Server; Health & Monitoring | Allocated | Adequate. Real-time notifications specified. |
| F1.3.4 | Visual indicator for active recording state | Client Applications (External) | External | Adequate via enablement. Service provides status; client renders indicator. |

### Functional - F2: File Management Requirements

| Requirement ID | Description | Allocated Architecture Component(s) | Traceability Status | Adequacy Assessment |
|---|---|---|---|---|
| F2.1.1 | Include location metadata when available | Client Applications (External) | External | Client-only. Service host typically lacks location; not a service concern. |
| F2.1.2 | Include timestamp metadata | MediaMTX Controller; MediaMTX Server | Partially allocated | Adequate for filename/timecoding; embedded metadata requires MediaMTX/FFmpeg tagging—confirm and configure. |
| F2.1.3 | Request device location permissions appropriately | Client Applications (External) | External | Client-only. |
| F2.2.1 | Default naming format `[datetime]_[id].[ext]` | MediaMTX Controller; Configuration Management | Allocated | Adequate. Naming template configurable and enforced server-side. |
| F2.2.2 | DateTime format `YYYY-MM-DD_HH-MM-SS` | MediaMTX Controller; Configuration Management | Allocated | Adequate. Deterministic formatting implemented in controller. |
| F2.2.3 | Unique ID 6-char alphanumeric | MediaMTX Controller | Allocated | Adequate. Add collision-safe ID generator and tests. |
| F2.2.4 | Example names (informative) | — | Not applicable | Informational only. |
| F2.3.1 | Store media in user-configurable default folder | Configuration Management; MediaMTX Controller | Allocated | Adequate. Configurable path; validated and applied. |
| F2.3.2 | Provide folder selection interface | Client Applications (External) | External | Client-only. |
| F2.3.3 | Validate storage permissions and available space | Health & Monitoring; Configuration Management | Allocated | Adequate. Disk space checks and warnings; permission validation on service host. |
| F2.3.4 | Default storage per platform | Client Applications (External); Configuration Management | Partially allocated | Client defaults external; service defaults via config for Linux host. |

### Functional - F3: UI and Security Requirements

| Requirement ID | Description | Allocated Architecture Component(s) | Traceability Status | Adequacy Assessment |
|---|---|---|---|---|
| F3.1.1 | Display list of available cameras from service | Camera Discovery Monitor; WebSocket JSON-RPC Server | Allocated | Adequate. Discovery provides list; API exposes to clients. |
| F3.1.2 | Show camera status (connected/disconnected) | Camera Discovery Monitor; Health & Monitoring; WebSocket JSON-RPC Server | Allocated | Adequate. Real-time status maintained and exposed. |
| F3.1.3 | Handle hot-plug events via real-time notifications | Camera Discovery Monitor; WebSocket JSON-RPC Server | Allocated | Adequate. Event stream supported. |
| F3.1.4 | Provide camera switching interface | Client Applications (External); WebSocket JSON-RPC Server | Partially allocated | Service supports selection/switch APIs; UI control external. |
| F3.2.1 | Recording start/stop controls | WebSocket JSON-RPC Server; MediaMTX Controller | Allocated | Adequate. |
| F3.2.2 | Duration selector interface | Client Applications (External); WebSocket JSON-RPC Server | Partially allocated | Service supports duration parameter; UI external. |
| F3.2.3 | Show recording progress and elapsed time | Health & Monitoring; WebSocket JSON-RPC Server; Client Applications (External) | Partially allocated | Service emits progress; UI external. |
| F3.2.4 | Emergency stop functionality | WebSocket JSON-RPC Server; MediaMTX Controller | Allocated | Adequate. Priority operation defined. |
| F3.2.5 | Protected methods require operator role (JWT) | Security Model; WebSocket JSON-RPC Server | Allocated | Adequate. Role-based access (viewer/operator/admin). |
| F3.2.6 | Handle token expiration with re-authentication | Security Model; WebSocket JSON-RPC Server; Client Applications (External) | Partially allocated | Service enforces expiry; client performs re-auth. |
| F3.3.1 | Settings interface (server, storage, quality, notifications) | Client Applications (External); Configuration Management | Partially allocated | Service provides configurable options/API; UI external. |
| F3.3.2 | Validate and persist user settings | Configuration Management; WebSocket JSON-RPC Server | Allocated | Adequate. Validation and persistence on service. |
| F3.3.3 | Reset settings to defaults | Configuration Management; WebSocket JSON-RPC Server | Allocated | Adequate. Reset operation supported. |

### Platform-Specific - Web (PWA)

| Requirement ID | Description | Allocated Architecture Component(s) | Traceability Status | Adequacy Assessment |
|---|---|---|---|---|
| W1.1 | Browser compatibility Chrome/Firefox/Safari | Client Applications (External) | External | Client-only. Service is protocol-compliant and browser-agnostic. |
| W1.2 | Responsive design | Client Applications (External) | External | Client-only. |
| W1.3 | PWA capabilities | Client Applications (External) | External | Client-only. |
| W1.4 | WebRTC preview when supported | MediaMTX Server; WebSocket JSON-RPC Server | Partially allocated | Service/MediaMTX support WebRTC; UI rendering external. |
| W2.1 | Browser download integration | Client Applications (External) | External | Client-only. |
| W2.2 | File naming preservation | Client Applications (External); MediaMTX Controller | Partially allocated | Service ensures filenames; client preserves on download. |
| W2.3 | Large file download progress | Client Applications (External) | External | Client-only. |

### Platform-Specific - Android

| Requirement ID | Description | Allocated Architecture Component(s) | Traceability Status | Adequacy Assessment |
|---|---|---|---|---|
| A1.1 | Min API level 28 | Client Applications (External) | External | Client-only. |
| A1.2 | Target API level 34 | Client Applications (External) | External | Client-only. |
| A1.3 | Camera/audio permissions | Client Applications (External) | External | Client-only. |
| A1.4 | Storage permissions | Client Applications (External) | External | Client-only. |
| A1.5 | Location permissions | Client Applications (External) | External | Client-only. |
| A2.1 | MediaStore integration | Client Applications (External) | External | Client-only. |
| A2.2 | Background recording with foreground service | Client Applications (External) | External | Client-only. |
| A2.3 | Android notification integration for recording | Client Applications (External) | External | Client-only. |
| A2.4 | Battery optimization exclusion guidance | Client Applications (External) | External | Client-only. |

### Non-Functional Requirements

| Requirement ID | Description | Allocated Architecture Component(s) | Traceability Status | Adequacy Assessment |
|---|---|---|---|---|
| N1.1 | Startup under 3s | WebSocket JSON-RPC Server; Health & Monitoring | Allocated | Adequate. Architecture targets quick startup; verify with perf tests. |
| N1.2 | Camera list refresh < 1s | Camera Discovery Monitor; WebSocket JSON-RPC Server | Allocated | Adequate. Hybrid udev+polling with 0.1s default supports target. |
| N1.3 | Photo capture response < 2s | MediaMTX Controller; MediaMTX Server | Allocated | Adequate. Depends on pipeline; monitor and tune. |
| N1.4 | Recording start within 2s | MediaMTX Controller; MediaMTX Server | Allocated | Adequate with health verification checks. |
| N1.5 | UI feedback within 200ms | WebSocket JSON-RPC Server; Client Applications (External) | Partially allocated | Service emits immediate ACK/notifications; UI rendering external. |
| N2.1 | Graceful service disconnection handling | Health & Monitoring; WebSocket JSON-RPC Server | Allocated | Adequate. Recovery flows defined. |
| N2.2 | Auto reconnection with backoff | Health & Monitoring | Allocated | Adequate. Exponential backoff and circuit breaker specified. |
| N2.3 | Preserve recording state across disconnections | MediaMTX Controller; Health & Monitoring | Allocated | Adequate with state tracking and restoration. |
| N2.4 | Validate all user inputs and responses | WebSocket JSON-RPC Server; Configuration Management | Allocated | Adequate. Schema validation and runtime checks. |
| N3.1 | Use WSS in production | Security Model; Configuration Management | Allocated | Adequate. TLS/mTLS supported. |
| N3.2 | Validate JWT tokens and handle expiration | Security Model; WebSocket JSON-RPC Server | Allocated | Adequate. Token verification and expiry enforced. |
| N3.3 | No plaintext sensitive credential storage | Security Model | Allocated | Adequate. API keys hashed (bcrypt) per decision. |
| N3.4 | Inactive session timeout | WebSocket JSON-RPC Server; Security Model | Allocated | Adequate. Session tracking and timeout policy. |
| N4.1 | Clear error messages and recovery guidance | WebSocket JSON-RPC Server; Client Applications (External) | Partially allocated | Service provides structured errors; UX copy external. |
| N4.2 | Consistent UI patterns | Client Applications (External) | External | Client-only. |
| N4.3 | Accessibility support | Client Applications (External) | External | Client-only. |
| N4.4 | Offline mode with limited functionality | Client Applications (External); WebSocket JSON-RPC Server | Partially allocated | Client offline behavior external; service can queue/retry and provide cached state when connected. |

### Integration & Implied System Requirements

| Requirement ID | Description | Allocated Architecture Component(s) | Traceability Status | Adequacy Assessment |
|---|---|---|---|---|
| I1.1 | WebSocket JSON-RPC 2.0 protocol | WebSocket JSON-RPC Server | Allocated | Adequate. Protocol compliance defined. |
| I1.2 | Connection endpoint `ws://host:8002/ws` | WebSocket JSON-RPC Server; Configuration Management | Allocated | Adequate. Configurable endpoint. |
| I1.3 | Real-time notifications for camera/recording | WebSocket JSON-RPC Server; Health & Monitoring | Allocated | Adequate. Notification pipeline specified. |
| I1.4 | Supported methods set (list/status/snapshot/recording) | WebSocket JSON-RPC Server | Allocated | Adequate. Methods enumerated in API docs. |
| I1.5 | Error handling uses JSON-RPC error taxonomy | WebSocket JSON-RPC Server | Allocated | Adequate. Negative test coverage planned. |
| I1.6 | Heartbeat ping every 30s | WebSocket JSON-RPC Server | Allocated | Adequate. Keepalive policy defined. |
| I1.7 | Role-based access control via JWT | Security Model; WebSocket JSON-RPC Server | Allocated | Adequate. Roles enforced per method. |

---

## Architecture → Requirements Mapping (Inverse Traceability)

For each architecture component, the requirements it fulfills:

- WebSocket JSON-RPC Server: F1.1.1, F1.1.2, F1.1.4, F1.2.1, F1.2.3, F1.2.4, F1.2.5, F1.3.2, F1.3.3, F3.1.1, F3.1.2, F3.1.3, F3.1.4 (partial), F3.2.1, F3.2.2 (partial), F3.2.3 (partial), F3.2.4, F3.2.5, F3.2.6 (partial), F3.3.2, F3.3.3, N1.1, N1.2, N1.5 (partial), N2.1, N2.4, N3.1, N3.2, N3.4, N4.1 (partial), N4.4 (partial), I1.1, I1.2, I1.3, I1.4, I1.5, I1.6, I1.7
- Camera Discovery Monitor: F3.1.1, F3.1.2, F3.1.3, N1.2
- MediaMTX Controller: F1.1.1, F1.2.1, F1.2.2, F1.2.3, F1.2.4, F1.2.5, F1.3.1 (partial), F2.2.1, F2.2.2, F2.2.3, F2.3.1, N1.3, N1.4, N2.3
- Health & Monitoring: F1.2.5, F1.3.2, F1.3.3, F3.1.2, N1.1, N2.1, N2.2, N2.3
- Security Model: F3.2.5, F3.2.6 (partial), N3.1, N3.2, N3.3, N3.4, I1.7
- Configuration Management: F2.2.1, F2.2.2, F2.3.1, F3.3.1 (partial), F3.3.2, F3.3.3, N2.4, N3.1, I1.2
- MediaMTX Server (external): F1.1.1, F1.2.1, F1.2.2, F1.3.1 (partial), N1.3, N1.4, W1.4 (partial)
- Client Applications (external): F1.1.3, F1.1.4 (partial), F1.3.2 (UI), F1.3.4, F3.1.1 (render), F3.1.2 (render), F3.1.4, F3.2.2, F3.2.3, F3.2.6 (re-auth), F3.3.1, W1.1, W1.2, W1.3, W1.4 (render), W2.1, W2.2, W2.3, A1.1–A1.5, A2.1–A2.4, N1.5 (render), N4.1–N4.3, N4.4 (offline UX)

---

## Gap Analysis

- Requirements without service architecture allocation (Critical Gaps):
  - F1.3.1 Auto file rollover on max size: Partially allocated. Critical if MediaMTX cannot enforce size-based rollover. Action: confirm MediaMTX capability; if absent, design controller-managed segmentation or post-process.
  - F2.1.2 Timestamp embedded in file metadata (beyond filename): Partially allocated. Action: verify MediaMTX/FFmpeg tagging support and configure; otherwise implement post-process tagging.

- Requirements explicitly external to service (not gaps):
  - All UI/UX-only and platform-specific items: F1.1.3, F1.3.4, F3.1.4, F3.2.2, F3.2.3 (UI), F3.3.1, W1.1–W2.3, A1.1–A2.4, N4.1–N4.3, N1.5 (UI rendering), N4.4 (offline UX).

- Architecture components with minimal/no direct requirement justification (potential over-engineering):
  - None identified. Every core component maps to multiple requirements. The only caution is ensuring size-based rollover (F1.3.1) is truly supported; otherwise, avoid over-design in Controller for a feature MediaMTX may already provide.

---

## Architecture Adequacy Conclusion

- The approved architecture is adequate to satisfy all service-side Phase 2 requirements cataloged, with two items requiring confirmation/configuration work: size-based recording rollover (F1.3.1) and embedded timestamp metadata (F2.1.2).
- Client/UI and platform-specific requirements are appropriately externalized; the service provides enabling APIs, notifications, and protocols to support their fulfillment.
- Security, resilience, and performance targets are explicitly supported through Security Model, Health & Monitoring, Controller, and WebSocket Server components as defined in `docs/architecture/overview.md`.

### Next Steps/Actions
- Confirm MediaMTX capability and configuration for F1.3.1 (size-based rollover) and F2.1.2 (embedded metadata). Update `docs/decisions.md` and API/config docs accordingly.
- Ensure API documentation (`docs/api/json_rpc_methods.md`) enumerates snapshot and recording methods with schemas and error taxonomy (I1.4, I1.5).
- Add performance and negative tests aligned to N1.x and I1.5 to validate adequacy claims.

### Evidence/References
- Architecture Overview: `docs/architecture/overview.md`
- Documentation Guidelines: `docs/development/documentation-guidelines.md`
- Ground Rules: `docs/development/project-ground-rules.md`
- Roles & Responsibilities: `docs/development/roles-responsibilities.md`
- Requirements Inventory: `evidence/sprint-3-actual/02_requirements_inventory.md`


