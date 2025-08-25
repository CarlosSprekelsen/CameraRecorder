# Requirements Baseline Document - Go Implementation

**Version:** 4.0  
**Date:** 2025-01-15  
**Role:** IV&V  
**Status:** Go Implementation Requirements Register  
**Related Documents:** `docs/requirements/client-requirements.md`, `docs/requirements/performance-requirements.md`, `docs/requirements/security-requirements.md`, `docs/architecture/overview.md`, `docs/api/`

---

## Executive Summary

This document serves as the master requirements register for the MediaMTX Camera Service Go implementation, providing a single source of truth for all requirements across the system. It consolidates requirements from authoritative sources with proper "SHALL" statements and clear acceptance criteria, updated for Go technology stack.

---

## 1. Technology Stack Requirements

### 1.1 Go Implementation Requirements

| REQ-ID | Description | Priority | Source Reference |
|--------|-------------|----------|------------------|
| REQ-TECH-GO-001 | The system SHALL be implemented in Go 1.24.6+ | Critical | Go Implementation |
| REQ-TECH-GO-002 | The system SHALL use gorilla/websocket for WebSocket implementation | Critical | Go Implementation |
| REQ-TECH-GO-003 | The system SHALL use golang-jwt/jwt/v4 for JWT authentication | Critical | Go Implementation |
| REQ-TECH-GO-004 | The system SHALL use golang.org/x/crypto/bcrypt for password hashing | Critical | Go Implementation |
| REQ-TECH-GO-005 | The system SHALL use viper for configuration management | High | Go Implementation |
| REQ-TECH-GO-006 | The system SHALL use logrus for structured logging | High | Go Implementation |
| REQ-TECH-GO-007 | The system SHALL use testify for testing utilities | High | Go Implementation |
| REQ-TECH-GO-008 | The system SHALL be statically linked for deployment | Critical | Go Implementation |
| REQ-TECH-GO-009 | The system SHALL use goroutines for concurrent operations | Critical | Go Implementation |
| REQ-TECH-GO-010 | The system SHALL use channels for inter-goroutine communication | Critical | Go Implementation |

### 1.2 Performance Targets (Go Implementation)

| REQ-ID | Description | Priority | Source Reference |
|--------|-------------|----------|------------------|
| REQ-PERF-GO-001 | API Response Time: <100ms for 95% of requests (5x improvement) | Critical | Performance Requirements |
| REQ-PERF-GO-002 | Concurrent Connections: 1000+ simultaneous WebSocket connections (10x improvement) | Critical | Performance Requirements |
| REQ-PERF-GO-003 | Throughput: 1000+ requests/second (5x improvement) | Critical | Performance Requirements |
| REQ-PERF-GO-004 | Memory Usage: <60MB base footprint, <200MB with 10 cameras (50% reduction) | High | Performance Requirements |
| REQ-PERF-GO-005 | CPU Usage: <50% sustained usage under normal load (30% reduction) | High | Performance Requirements |
| REQ-PERF-GO-006 | Goroutine Limit: <1000 concurrent goroutines maximum | High | Performance Requirements |

---

## 2. Client Application Requirements

### 2.1 Photo Capture Requirements

| REQ-ID | Description | Priority | Source Reference |
|--------|-------------|----------|------------------|
| REQ-CLIENT-001 | The application SHALL allow users to take photos using available cameras via service's `take_snapshot` JSON-RPC method with preview display | Critical | F1.1.1 |
| REQ-CLIENT-004 | The application SHALL handle photo capture errors gracefully with user feedback | High | F1.1.4 |

### 2.2 Video Recording Requirements

| REQ-ID | Description | Priority | Source Reference |
|--------|-------------|----------|------------------|
| REQ-CLIENT-005 | The application SHALL allow users to record videos using available cameras | Critical | F1.2.1 |
| REQ-CLIENT-006 | The application SHALL support unlimited duration recording mode | Critical | F1.2.2 |
| REQ-CLIENT-007 | The application SHALL support timed recording with user-specified duration in seconds, minutes, or hours | Critical | F1.2.3 |
| REQ-CLIENT-008 | The application SHALL allow users to manually stop video recording | High | F1.2.4 |
| REQ-CLIENT-009 | The application SHALL handle recording session management via service API | High | F1.2.5 |

### 2.3 Recording Management Requirements

| REQ-ID | Description | Priority | Source Reference |
|--------|-------------|----------|------------------|
| REQ-CLIENT-010 | The application SHALL automatically create new video files when maximum file size is reached (handled by service) | High | F1.3.1 |
| REQ-CLIENT-011 | The application SHALL display recording status and elapsed time in real-time | High | F1.3.2 |
| REQ-CLIENT-012 | The application SHALL notify users when video recording is completed | High | F1.3.3 |
| REQ-CLIENT-013 | The application SHALL provide visual indicators for active recording state | High | F1.3.4 |

### 2.4 Metadata Management Requirements

| REQ-ID | Description | Priority | Source Reference |
|--------|-------------|----------|------------------|
| REQ-CLIENT-014 | The application SHALL ensure photos and videos include location metadata (when available) | High | F2.1.1 |
| REQ-CLIENT-015 | The application SHALL ensure photos and videos include timestamp metadata | High | F2.1.2 |
| REQ-CLIENT-016 | The application SHALL request device location permissions appropriately | High | F2.1.3 |

### 2.5 File Naming Requirements

| REQ-ID | Description | Priority | Source Reference |
|--------|-------------|----------|------------------|
| REQ-CLIENT-017 | The application SHALL use default naming format: `[datetime]_[unique_id].[extension]` | High | F2.2.1 |
| REQ-CLIENT-018 | DateTime format SHALL be: `YYYY-MM-DD_HH-MM-SS` | High | F2.2.2 |
| REQ-CLIENT-019 | Unique ID SHALL be a 6-character alphanumeric string | High | F2.2.3 |

### 2.6 Storage Configuration Requirements

| REQ-ID | Description | Priority | Source Reference |
|--------|-------------|----------|------------------|
| REQ-CLIENT-020 | The application SHALL store media files in a user-configurable default folder | High | F2.3.1 |
| REQ-CLIENT-021 | The application SHALL provide folder selection interface | High | F2.3.2 |
| REQ-CLIENT-022 | The application SHALL validate storage permissions and available space | High | F2.3.3 |
| REQ-CLIENT-023 | Default storage location SHALL be platform-appropriate | High | F2.3.4 |

### 2.7 File Lifecycle Management Requirements

| REQ-ID | Description | Priority | Source Reference |
|--------|-------------|----------|------------------|
| REQ-CLIENT-034 | The application SHALL provide file deletion capabilities for recordings and snapshots via service API | High | F2.4.1 |
| REQ-CLIENT-035 | The application SHALL implement configurable retention policies for media files (age-based, size-based, or manual) | High | F2.4.2 |
| REQ-CLIENT-036 | The application SHALL provide storage space monitoring and alerts when space is low | High | F2.4.3 |
| REQ-CLIENT-037 | The application SHALL support automatic cleanup of old files based on retention policies | High | F2.4.4 |
| REQ-CLIENT-038 | The application SHALL provide manual file management interface for bulk operations | High | F2.4.5 |
| REQ-CLIENT-039 | The application SHALL support file archiving to external storage before deletion | Medium | F2.4.6 |
| REQ-CLIENT-040 | The application SHALL provide file metadata viewing capabilities (size, duration, creation date, etc.) | High | F2.4.7 |
| REQ-CLIENT-041 | The application SHALL implement role-based access control for file deletion (admin/operator roles only) | Critical | F2.4.8 |

### 2.8 Camera Selection Requirements

| REQ-ID | Description | Priority | Source Reference |
|--------|-------------|----------|------------------|
| REQ-CLIENT-024 | The application SHALL display list of available cameras from service API | Critical | F3.1.1 |
| REQ-CLIENT-025 | The application SHALL show camera status (connected/disconnected) | High | F3.1.2 |
| REQ-CLIENT-026 | The application SHALL handle camera hot-plug events via real-time notifications | High | F3.1.3 |
| REQ-CLIENT-027 | The application SHALL provide camera switching interface | High | F3.1.4 |

### 2.9 Recording Controls Requirements

| REQ-ID | Description | Priority | Source Reference |
|--------|-------------|----------|------------------|
| REQ-CLIENT-028 | The application SHALL provide intuitive recording start/stop controls | Critical | F3.2.1 |
| REQ-CLIENT-029 | The application SHALL display recording duration selector interface | High | F3.2.2 |
| REQ-CLIENT-030 | The application SHALL show recording progress and elapsed time | High | F3.2.3 |
| REQ-CLIENT-031 | The application SHALL provide emergency stop functionality | High | F3.2.4 |
| REQ-CLIENT-032 | The application SHALL implement role-based access control with viewer, operator, and admin permissions for all protected operations | Critical | F3.2.5 |
| REQ-CLIENT-033 | The application SHALL handle token expiration by re-authenticating before retrying protected operations | High | F3.2.6 |

---

## 3. API Requirements

### 3.1 JSON-RPC API Requirements

| REQ-ID | Description | Priority | Source Reference |
|--------|-------------|----------|------------------|
| REQ-API-001 | The system SHALL provide WebSocket JSON-RPC 2.0 API endpoint at `ws://localhost:8002/ws` | Critical | API Reference |
| REQ-API-002 | The system SHALL implement `ping` method for health checks | Critical | API Reference |
| REQ-API-003 | The system SHALL implement `get_camera_list` method for camera enumeration | Critical | API Reference |
| REQ-API-004 | The system SHALL implement `get_camera_status` method for camera status | Critical | API Reference |
| REQ-API-005 | The system SHALL implement `take_snapshot` method for photo capture | Critical | API Reference |
| REQ-API-006 | The system SHALL implement `start_recording` method for video recording | Critical | API Reference |
| REQ-API-007 | The system SHALL implement `stop_recording` method for video recording | Critical | API Reference |
| REQ-API-008 | The system SHALL implement `authenticate` method for authentication | Critical | API Reference |
| REQ-API-009 | The system SHALL implement role-based access control with viewer, operator, and admin permissions for all protected methods | Critical | API Reference |
| REQ-API-011 | API methods SHALL respond within specified time limits: Status methods <50ms, Control methods <100ms | High | API Reference |
| REQ-API-013 | WebSocket Notifications SHALL be delivered within <20ms | High | API Reference |
| REQ-API-014 | The system SHALL implement `get_streams` method for stream enumeration | Critical | API Reference |
| REQ-API-015 | The system SHALL implement `list_recordings` method for recording file management | Critical | API Reference |
| REQ-API-016 | The system SHALL implement `list_snapshots` method for snapshot file management | Critical | API Reference |
| REQ-API-017 | The system SHALL implement `get_metrics` method for system performance metrics | Critical | API Reference |
| REQ-API-018 | The system SHALL implement `get_status` method for system health information | Critical | API Reference |
| REQ-API-019 | The system SHALL implement `get_server_info` method for server configuration | Critical | API Reference |
| REQ-API-020 | The system SHALL provide real-time camera status update notifications | Critical | API Reference |
| REQ-API-021 | The system SHALL provide real-time recording status update notifications | Critical | API Reference |
| REQ-API-022 | The system SHALL implement HTTP file download endpoints for recordings | Critical | API Reference |
| REQ-API-023 | The system SHALL implement HTTP file download endpoints for snapshots | Critical | API Reference |

### 3.2 Health Endpoints Requirements

| REQ-ID | Description | Priority | Source Reference |
|--------|-------------|----------|------------------|
| REQ-API-014 | The system SHALL provide REST health endpoints at `http://localhost:8003/health/` for system, cameras, and MediaMTX monitoring | Critical | Health Endpoints |
| REQ-API-017 | Health endpoints SHALL return JSON responses with status and timestamp | High | Health Endpoints |
| REQ-API-018 | Health endpoints SHALL return 200 OK for healthy status | High | Health Endpoints |
| REQ-API-019 | Health endpoints SHALL return 500 Internal Server Error for unhealthy status | High | Health Endpoints |

---

## 4. Security Requirements

### 4.1 Authentication Requirements

| REQ-ID | Description | Priority | Source Reference |
|--------|-------------|----------|------------------|
| REQ-SEC-001 | The system SHALL implement JWT token-based authentication for all API access | Critical | REQ-SEC-001 |
| REQ-SEC-002 | Token Format: JSON Web Token (JWT) with standard claims | Critical | REQ-SEC-001 |
| REQ-SEC-003 | Token Expiration: Configurable expiration time (default: 24 hours) | Critical | REQ-SEC-001 |
| REQ-SEC-004 | Token Refresh: Support for token refresh mechanism | Critical | REQ-SEC-001 |
| REQ-SEC-005 | Token Validation: Proper signature validation and claim verification | Critical | REQ-SEC-001 |
| REQ-SEC-006 | The system SHALL validate API keys for service-to-service communication | Critical | REQ-SEC-002 |
| REQ-SEC-007 | API Key Format: Secure random string (32+ characters) | Critical | REQ-SEC-002 |
| REQ-SEC-008 | Key Storage: Secure storage of API keys | Critical | REQ-SEC-002 |
| REQ-SEC-009 | Key Rotation: Support for API key rotation | Critical | REQ-SEC-002 |

### 4.2 Authorization Requirements

| REQ-ID | Description | Priority | Source Reference |
|--------|-------------|----------|------------------|
| REQ-SEC-010 | The system SHALL implement role-based access control for different user types | Critical | REQ-SEC-003 |
| REQ-SEC-011 | User Roles: Admin, User, Read-Only roles | Critical | REQ-SEC-003 |
| REQ-SEC-012 | Permission Matrix: Clear permission definitions for each role | Critical | REQ-SEC-003 |
| REQ-SEC-013 | Access Control: Enforcement of role-based permissions | Critical | REQ-SEC-003 |
| REQ-SEC-014 | The system SHALL control access to camera resources and media files | Critical | REQ-SEC-004 |
| REQ-SEC-015 | Camera Access: Users can only access authorized cameras | Critical | REQ-SEC-004 |
| REQ-SEC-016 | File Access: Users can only access authorized media files | Critical | REQ-SEC-004 |
| REQ-SEC-017 | Resource Isolation: Proper isolation between user resources | Critical | REQ-SEC-004 |
| REQ-SEC-018 | Access Logging: Logging of all resource access attempts | Critical | REQ-SEC-004 |

### 4.3 Input Validation Requirements

| REQ-ID | Description | Priority | Source Reference |
|--------|-------------|----------|------------------|
| REQ-SEC-019 | The system SHALL sanitize and validate all input data | High | REQ-SEC-005 |
| REQ-SEC-020 | Input Validation: Comprehensive validation of all input parameters | High | REQ-SEC-005 |
| REQ-SEC-021 | Sanitization: Proper sanitization of user input | High | REQ-SEC-005 |
| REQ-SEC-022 | Injection Prevention: Prevention of SQL injection, XSS, and command injection | High | REQ-SEC-005 |
| REQ-SEC-023 | Parameter Validation: Validation of parameter types and ranges | High | REQ-SEC-005 |
| REQ-SEC-024 | The system SHALL implement secure file upload handling | High | REQ-SEC-006 |
| REQ-SEC-025 | File Type Validation: Validation of uploaded file types | High | REQ-SEC-006 |
| REQ-SEC-026 | File Size Limits: Enforcement of file size limits | High | REQ-SEC-006 |
| REQ-SEC-027 | Virus Scanning: Scanning of uploaded files for malware | High | REQ-SEC-006 |
| REQ-SEC-028 | Secure Storage: Secure storage of uploaded files | High | REQ-SEC-006 |

### 4.4 Data Protection Requirements

| REQ-ID | Description | Priority | Source Reference |
|--------|-------------|----------|------------------|
| REQ-SEC-029 | The system SHALL encrypt sensitive data in transit and at rest | High | REQ-SEC-007 |
| REQ-SEC-030 | Transport Encryption: TLS 1.2+ for all communications | High | REQ-SEC-007 |
| REQ-SEC-031 | Storage Encryption: Encryption of sensitive data at rest | High | REQ-SEC-007 |
| REQ-SEC-032 | The system SHALL implement comprehensive audit logging for all security-relevant events | Critical | Security Requirements |
| REQ-SEC-033 | The system SHALL implement rate limiting to prevent abuse and DoS attacks | Critical | Security Requirements |
| REQ-SEC-034 | The system SHALL implement configurable session timeout for all authenticated sessions | Critical | Security Requirements |
| REQ-SEC-035 | The system SHALL implement data encryption at rest for all sensitive data storage | Critical | Security Requirements |

---

## 5. Testing Requirements

### 5.1 Testing Architecture Requirements

| REQ-ID | Description | Priority | Source Reference |
|--------|-------------|----------|------------------|
| REQ-TEST-001 | All tests SHALL use the single systemd-managed MediaMTX service instance | Critical | AD-001 |
| REQ-TEST-002 | Tests SHALL NOT create multiple MediaMTX instances or start their own MediaMTX processes | Critical | AD-001 |
| REQ-TEST-003 | Tests SHALL validate against actual production MediaMTX service | Critical | AD-001 |
| REQ-TEST-004 | Tests SHALL use fixed systemd service ports (API: 9997, RTSP: 8554, WebRTC: 8889, HLS: 8888) | High | AD-001 |
| REQ-TEST-005 | Tests SHALL coordinate on shared service with proper test isolation | High | AD-001 |
| REQ-TEST-006 | Tests SHALL verify MediaMTX service is running via systemd before execution | High | AD-001 |

### 5.2 Test Coverage Requirements

| REQ-ID | Description | Priority | Source Reference |
|--------|-------------|----------|------------------|
| REQ-TEST-007 | The system SHALL have comprehensive test coverage for all API methods | Critical | Testing Guide |
| REQ-TEST-008 | The system SHALL have real system integration tests using actual MediaMTX service | Critical | Testing Guide |
| REQ-TEST-009 | The system SHALL have authentication and authorization test coverage | Critical | Testing Guide |
| REQ-TEST-010 | The system SHALL have error handling and edge case test coverage | High | Testing Guide |
| REQ-TEST-011 | The system SHALL have performance test coverage for response time requirements | High | Testing Guide |
| REQ-TEST-012 | The system SHALL have security test coverage for all security requirements | High | Testing Guide |

---

## 6. Deployment Requirements

### 6.1 Go Deployment Requirements

| REQ-ID | Description | Priority | Source Reference |
|--------|-------------|----------|------------------|
| REQ-DEPLOY-GO-001 | The system SHALL be deployable as a single statically linked binary | Critical | Go Implementation |
| REQ-DEPLOY-GO-002 | The system SHALL support SystemD service management | Critical | Go Implementation |
| REQ-DEPLOY-GO-003 | The system SHALL support OCI-compatible container deployment | High | Go Implementation |
| REQ-DEPLOY-GO-004 | The system SHALL support CNCF-compliant orchestration deployment | Medium | Go Implementation |
| REQ-DEPLOY-GO-005 | The system SHALL provide automated installation scripts | High | Go Implementation |
| REQ-DEPLOY-GO-006 | The system SHALL provide automated uninstallation scripts | High | Go Implementation |
| REQ-DEPLOY-GO-007 | The system SHALL support configuration hot-reload | High | Go Implementation |
| REQ-DEPLOY-GO-008 | The system SHALL provide health check endpoints for monitoring | Critical | Go Implementation |

---

## 7. Requirements Summary

### 7.1 Requirements by Category

| Category | Count | Critical | High | Medium | Total |
|----------|-------|----------|------|--------|-------|
| Technology Stack (Go) | 10 | 6 | 4 | 0 | 10 |
| Performance (Go) | 6 | 3 | 3 | 0 | 6 |
| Client Application | 41 | 8 | 33 | 0 | 41 |
| API | 23 | 18 | 5 | 0 | 23 |
| Security | 35 | 20 | 15 | 0 | 35 |
| Testing | 12 | 6 | 6 | 0 | 12 |
| Deployment (Go) | 8 | 3 | 4 | 1 | 8 |
| **TOTAL** | **135** | **64** | **70** | **1** | **135** |

### 7.2 Requirements by Priority

| Priority | Count | Percentage |
|----------|-------|------------|
| Critical | 64 | 47.4% |
| High | 70 | 51.9% |
| Medium | 1 | 0.7% |
| **TOTAL** | **135** | **100%** |

---

## 8. Go Implementation Benefits

### 8.1 Performance Improvements
- **Response Time:** 5x improvement (500ms → 100ms)
- **Concurrency:** 10x improvement (100 → 1000+ connections)
- **Throughput:** 5x improvement (200 → 1000+ requests/second)
- **Memory Usage:** 50% reduction (80% → 60%)
- **CPU Usage:** 30% reduction (70% → 50%)

### 8.2 Deployment Benefits
- **Single Binary:** No runtime dependencies
- **Static Linking:** Simplified deployment
- **Container Support:** Efficient OCI-compatible container images
- **Cross-Platform:** Linux, macOS, Windows support
- **Resource Efficiency:** Lower memory and CPU requirements

### 8.3 Development Benefits
- **Type Safety:** Compile-time error checking
- **Concurrency:** Built-in goroutines and channels
- **Tooling:** Excellent development tools
- **Testing:** Built-in testing and benchmarking
- **Documentation:** Automatic documentation generation

---

**Document Status:** Complete requirements baseline with Go technology updates  
**Last Updated:** 2025-01-15  
**Next Review:** After Go implementation validation
