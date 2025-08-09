# Requirements Inventory - Sprint 3 CDR

Version: 1.0
Date: 2025-08-09
Prepared by: IV&V Team
Status: For architecture component inventory

## Executive Summary
This document catalogs all known requirements from client and system sources for the applications interfacing with the MediaMTX Camera Service. It includes functional (F), non-functional (N), platform-specific (Web/Android), integration, and implied system requirements. Each requirement is categorized, prioritized, and assessed for testability. Customer-critical domains are photo capture (F1.1), recording (F1.2, F1.3), camera selection and notifications (F3.1), and security enforcement (F3.2, N3.x).

Highlights:
- Customer-critical: F1.1, F1.2, F3.1, F3.2
- System-critical: Protocol, error taxonomy, heartbeat, discovery integration
- Security-critical: F3.2.5, F3.2.6, N3.1-N3.4
- Performance-critical: N1.1-N1.5

## Requirements Catalog
Table columns: ID, Description, Category, Priority, Testability Status

### Functional - F1: Camera Interface Requirements

| ID | Description | Category | Priority | Testability Status |
|---|---|---|---|---|
| F1.1.1 | Allow users to take photos using available cameras | Customer-critical | High | API test via `take_snapshot`; optional UI E2E |
| F1.1.2 | Use service `take_snapshot` JSON-RPC method | System-critical | High | API contract test |
| F1.1.3 | Display preview of captured photos | Customer-critical | Medium | UI test; manual visual verification |
| F1.1.4 | Graceful photo error handling with user feedback | Customer-critical | High | API error injection + UI message check |
| F1.2.1 | Allow users to record videos using available cameras | Customer-critical | High | API test via `start_recording`/`stop_recording` |
| F1.2.2 | Unlimited duration recording mode | Customer-critical | High | API test: start without duration; ensure not auto-stopped |
| F1.2.3 | Timed recording with seconds/minutes/hours | Customer-critical | High | API tests per unit; verify auto-stop and completion notice |
| F1.2.4 | Manual stop of recording | Customer-critical | High | API test via `stop_recording` |
| F1.2.5 | Recording session management via service API | System-critical | High | API tests: state transitions, conflicts |
| F1.3.1 | Auto file rollover on max size (service) | System-critical | Medium | Integration test with MediaMTX emulator/config observation |
| F1.3.2 | Display recording status and elapsed time | Customer-critical | Medium | UI test; assert notifications arrival |
| F1.3.3 | Notify when recording is completed | Customer-critical | High | API notification test |
| F1.3.4 | Visual indicator for active recording state | Customer-critical | Medium | UI test |

### Functional - F2: File Management Requirements

| ID | Description | Category | Priority | Testability Status |
|---|---|---|---|---|
| F2.1.1 | Include location metadata when available | System-critical | Medium | Instrumented test; manual metadata inspection |
| F2.1.2 | Include timestamp metadata | System-critical | Medium | File metadata inspection |
| F2.1.3 | Request device location permissions appropriately | System-critical | Medium | Platform permission flow test |
| F2.2.1 | Default naming format `[datetime]_[id].[ext]` | System-critical | Medium | File name pattern test |
| F2.2.2 | DateTime format `YYYY-MM-DD_HH-MM-SS` | System-critical | Medium | Name parsing test |
| F2.2.3 | Unique ID 6-char alphanumeric | System-critical | Medium | Regex validation |
| F2.2.4 | Example names (informative) | Informational | Low | Not applicable |
| F2.3.1 | Store media in user-configurable default folder | Customer-critical | Medium | UI config + file existence test |
| F2.3.2 | Provide folder selection interface | Customer-critical | Medium | UI test |
| F2.3.3 | Validate storage permissions and available space | System-critical | Medium | Platform permission/space checks |
| F2.3.4 | Default storage per platform | System-critical | Medium | Config check per platform |

### Functional - F3: UI and Security Requirements

| ID | Description | Category | Priority | Testability Status |
|---|---|---|---|---|
| F3.1.1 | Display list of available cameras from service | Customer-critical | High | API `get_camera_list` + UI render |
| F3.1.2 | Show camera status (connected/disconnected) | Customer-critical | High | API `get_camera_status` + UI |
| F3.1.3 | Handle hot-plug events via real-time notifications | Customer-critical | High | Notification delivery tests (multi-client) |
| F3.1.4 | Provide camera switching interface | Customer-critical | Medium | UI test |
| F3.2.1 | Recording start/stop controls | Customer-critical | High | UI flows + API invocation |
| F3.2.2 | Duration selector interface | Customer-critical | Medium | UI test |
| F3.2.3 | Show recording progress and elapsed time | Customer-critical | Medium | UI + notification assertions |
| F3.2.4 | Emergency stop functionality | System-critical | High | API `stop_recording` under failure scenarios |
| F3.2.5 | Protected methods require operator role (JWT) | Security-critical | High | AuthN/AuthZ tests; negative/positive cases |
| F3.2.6 | Handle token expiration with re-authentication | Security-critical | High | Token TTL tests + retry flow |
| F3.3.1 | Settings interface (server, storage, quality, notifications) | Customer-critical | Medium | UI + persistence tests |
| F3.3.2 | Validate and persist user settings | System-critical | Medium | Persistence/integration test |
| F3.3.3 | Reset settings to defaults | Customer-critical | Medium | UI + config reset test |

### Platform-Specific - Web (PWA)

| ID | Description | Category | Priority | Testability Status |
|---|---|---|---|---|
| W1.1 | Browser compatibility Chrome/Firefox/Safari | System-critical | Medium | Compatibility test matrix |
| W1.2 | Responsive design | Customer-critical | Medium | UI responsive tests |
| W1.3 | PWA capabilities | System-critical | Medium | Lighthouse/PWA audit |
| W1.4 | WebRTC preview when supported | System-critical | Medium | Feature detection + preview test |
| W2.1 | Browser download integration | Customer-critical | Medium | E2E download tests |
| W2.2 | File naming preservation | System-critical | Medium | Download filename assertion |
| W2.3 | Large file download progress | Customer-critical | Medium | UI progress tests |

### Platform-Specific - Android

| ID | Description | Category | Priority | Testability Status |
|---|---|---|---|---|
| A1.1 | Min API level 28 | System-critical | Medium | Build/config check |
| A1.2 | Target API level 34 | System-critical | Medium | Build/config check |
| A1.3 | Camera/audio permissions | System-critical | High | Permission flow tests |
| A1.4 | Storage permissions | System-critical | High | Permission flow tests |
| A1.5 | Location permissions | System-critical | Medium | Permission flow tests |
| A2.1 | MediaStore integration | System-critical | Medium | Instrumented test |
| A2.2 | Background recording with foreground service | System-critical | High | Instrumented/OS behavior test |
| A2.3 | Android notification integration for recording | Customer-critical | Medium | Instrumented UI/notification test |
| A2.4 | Battery optimization exclusion guidance | System-critical | Low | UX guidance presence check |

### Non-Functional Requirements

| ID | Description | Category | Priority | Testability Status |
|---|---|---|---|---|
| N1.1 | Startup under 3s | Performance-critical | Medium | Performance test |
| N1.2 | Camera list refresh < 1s | Performance-critical | Medium | Performance test |
| N1.3 | Photo capture response < 2s | Performance-critical | Medium | Performance test |
| N1.4 | Recording start within 2s | Performance-critical | Medium | Performance test |
| N1.5 | UI feedback within 200ms | Performance-critical | Medium | UX/perf test |
| N2.1 | Graceful service disconnection handling | System-critical | High | Integration tests |
| N2.2 | Auto reconnection with backoff | System-critical | High | Long-run integration test |
| N2.3 | Preserve recording state across disconnections | System-critical | High | Fault-injection integration test |
| N2.4 | Validate all user inputs and responses | System-critical | High | Static checks + runtime validation tests |
| N3.1 | Use WSS in production | Security-critical | High | Deployment/config validation |
| N3.2 | Validate JWT tokens and handle expiration | Security-critical | High | Security unit/integration tests |
| N3.3 | No plaintext sensitive credential storage | Security-critical | High | Code/config review |
| N3.4 | Inactive session timeout | Security-critical | Medium | Integration test with timer |
| N4.1 | Clear error messages and recovery guidance | Customer-critical | Medium | UX copy review |
| N4.2 | Consistent UI patterns | Customer-critical | Low | UX review |
| N4.3 | Accessibility support | Customer-critical | Medium | Accessibility audit |
| N4.4 | Offline mode with limited functionality | System-critical | Low | Offline mode test |

### Integration & Implied System Requirements

| ID | Description | Category | Priority | Testability Status |
|---|---|---|---|---|
| I1.1 | WebSocket JSON-RPC 2.0 protocol | System-critical | High | Protocol compliance tests |
| I1.2 | Connection endpoint `ws://host:8002/ws` | System-critical | High | Config validation |
| I1.3 | Real-time notifications for camera/recording | Customer-critical | High | Notification delivery tests |
| I1.4 | Supported methods set (list/status/snapshot/recording) | System-critical | High | API discovery/contract tests |
| I1.5 | Error handling uses JSON-RPC error taxonomy | System-critical | High | Negative tests per error code |
| I1.6 | Heartbeat ping every 30s | System-critical | Medium | Long-run connection test |
| I1.7 | Role-based access control via JWT | Security-critical | High | AuthZ tests (positive/negative) |

## Priority Analysis

- High Priority
  - Customer-critical: F1.1.x, F1.2.x, F3.1.x, F3.2.5, F3.2.6
  - Security-critical: N3.1-N3.3, I1.7
  - System-critical: Protocol, error taxonomy, availability and notifications (I1.1, I1.5, I1.3)
- Medium Priority
  - Performance targets N1.x; UI progress/elapsed time; platform permissions
- Low Priority
  - UX polish (consistency), offline limited mode, guidance items

## Testability Assessment

- API-centric tests cover: F1.1, F1.2, F3.1, F3.2, I1.x, N2.1-N2.4, N3.2
- UI/instrumented tests cover: F1.1.3, F1.3.2, F3.2.2-F3.2.3, settings interfaces, platform permissions
- Deployment/config validation: N3.1 (WSS), endpoint configuration, target API levels (Android)
- Performance tests: N1.x thresholds using smoke/perf harness
- Documentation/inspection: credential storage (N3.3), accessibility (N4.3)

Coverage Gaps and Notes
- Some file/metadata items require device/emulator or file system inspection (F2.1.x)
- Android background recording (A2.2) needs device-level instrumentation
- Accessibility and responsiveness require specialized tooling/audits

---

This inventory is provided to IV&V for the next step: architecture component inventory and mapping of requirements to components and tests.
