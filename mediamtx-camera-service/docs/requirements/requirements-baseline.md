# Requirements Baseline Document

**Version:** 1.0  
**Date:** 2025-01-15  
**Role:** Project Manager  
**Status:** 🚀 MASTER REQUIREMENTS REGISTER  
**Related Documents:** `docs/requirements/client-requirements.md`, `docs/requirements/performance-requirements.md`

---

## Executive Summary

This document serves as the master requirements register for the MediaMTX Camera Service project, providing a single source of truth for all requirements across the system. It consolidates requirements from all sources and establishes traceability between business needs, client requirements, and technical specifications.

---

## 1. Requirements Structure

### 1.1 Requirements Categories

#### Functional Requirements (REQ-FUNC-*)
- **Client Application Requirements:** User interface and application functionality
- **Service Integration Requirements:** MediaMTX service integration and API functionality
- **File Management Requirements:** Media file handling and storage
- **Camera Control Requirements:** Camera discovery, control, and status management

#### Non-Functional Requirements (REQ-NFUNC-*)
- **Performance Requirements (REQ-PERF-*):** Response times, throughput, scalability
- **Security Requirements (REQ-SEC-*):** Authentication, authorization, data protection
- **Reliability Requirements (REQ-REL-*):** Availability, fault tolerance, recovery
- **Usability Requirements (REQ-USE-*):** User experience, accessibility, documentation

#### Technical Requirements (REQ-TECH-*)
- **Architecture Requirements:** System architecture and design constraints
- **Integration Requirements:** External system integration and APIs
- **Deployment Requirements:** Deployment, operations, and maintenance
- **Compliance Requirements:** Standards, regulations, and compliance

---

## 2. Requirements Master List

### 2.1 Functional Requirements

| REQ-ID | Category | Description | Priority | Source | Status |
|--------|----------|-------------|----------|--------|--------|
| REQ-FUNC-001 | Client | Photo capture functionality | Critical | Client Requirements | ✅ Implemented |
| REQ-FUNC-002 | Client | Video recording functionality | Critical | Client Requirements | ✅ Implemented |
| REQ-FUNC-003 | Client | Camera discovery and selection | High | Client Requirements | ✅ Implemented |
| REQ-FUNC-004 | Service | WebSocket JSON-RPC API | Critical | Architecture | ✅ Implemented |
| REQ-FUNC-005 | Service | MediaMTX integration | Critical | Architecture | ✅ Implemented |
| REQ-FUNC-006 | File | Metadata management | High | Client Requirements | ✅ Implemented |
| REQ-FUNC-007 | File | Storage configuration | High | Client Requirements | ✅ Implemented |
| REQ-FUNC-008 | File | JSON-RPC list_recordings method | Critical | Epic E6 | ✅ Implemented |
| REQ-FUNC-009 | File | JSON-RPC list_snapshots method | Critical | Epic E6 | ✅ Implemented |
| REQ-FUNC-010 | File | HTTP recording file download endpoint | Critical | Epic E6 | ✅ Implemented |
| REQ-FUNC-011 | File | HTTP snapshot file download endpoint | Critical | Epic E6 | ✅ Implemented |
| REQ-FUNC-012 | File | Nginx routing for file endpoints | Critical | Epic E6 | ✅ Implemented |

### 2.2 Non-Functional Requirements

| REQ-ID | Category | Description | Priority | Source | Status |
|--------|----------|-------------|----------|--------|--------|
| REQ-PERF-001 | Performance | API response time < 500ms (Python) | High | Performance Requirements | ⚠️ Validating |
| REQ-PERF-002 | Performance | Camera discovery < 10 seconds | High | Performance Requirements | ⚠️ Validating |
| REQ-PERF-003 | Performance | 50-100 concurrent connections | High | Performance Requirements | ⚠️ Validating |
| REQ-PERF-004 | Performance | Resource usage limits | High | Performance Requirements | ⚠️ Validating |
| REQ-PERF-005 | Performance | Throughput 100-200 req/s | Medium | Performance Requirements | ⚠️ Validating |
| REQ-PERF-006 | Performance | Scalability requirements | Medium | Performance Requirements | ⚠️ Validating |
| REQ-SEC-001 | Security | JWT authentication | Critical | Security Requirements | ⚠️ Validating |
| REQ-SEC-002 | Security | API key validation | Critical | Security Requirements | ⚠️ Validating |
| REQ-SEC-003 | Security | Input validation | High | Security Requirements | ⚠️ Validating |
| REQ-SEC-004 | Security | Data encryption | High | Security Requirements | ⚠️ Validating |

### 2.3 Technical Requirements

| REQ-ID | Category | Description | Priority | Source | Status |
|--------|----------|-------------|----------|--------|--------|
| REQ-TECH-001 | Architecture | Python implementation | High | Technical Requirements | ✅ Implemented |
| REQ-TECH-002 | Architecture | WebSocket communication | Critical | Architecture | ✅ Implemented |
| REQ-TECH-003 | Integration | MediaMTX API integration | Critical | Architecture | ✅ Implemented |
| REQ-TECH-004 | Deployment | Docker containerization | High | Deployment Requirements | ⚠️ Validating |
| REQ-TECH-005 | Operations | Monitoring and alerting | High | Operations Requirements | ⚠️ Validating |

---

## 3. Requirements Traceability Matrix

### 3.1 Business Need to Requirements Mapping

| Business Need | Client Requirements | Performance Requirements | Technical Requirements |
|---------------|-------------------|-------------------------|----------------------|
| Real-time camera control | REQ-FUNC-001, REQ-FUNC-002 | REQ-PERF-001, REQ-PERF-002 | REQ-TECH-002, REQ-TECH-003 |
| Multi-user support | REQ-FUNC-003 | REQ-PERF-003, REQ-PERF-006 | REQ-TECH-001, REQ-TECH-004 |
| Secure operations | REQ-FUNC-004 | REQ-SEC-001, REQ-SEC-002 | REQ-TECH-005 |
| Reliable file management | REQ-FUNC-006, REQ-FUNC-007, REQ-FUNC-008, REQ-FUNC-009, REQ-FUNC-010, REQ-FUNC-011, REQ-FUNC-012 | REQ-PERF-004, REQ-PERF-005 | REQ-TECH-001 |

### 3.2 Requirements to Test Mapping

| Requirement | Test Category | Test Files | Validation Status |
|-------------|---------------|------------|-------------------|
| REQ-PERF-001 | Performance | `tests/performance/test_response_times.py` | ⚠️ In Progress |
| REQ-PERF-002 | Performance | `tests/performance/test_camera_discovery.py` | ⚠️ In Progress |
| REQ-PERF-003 | Performance | `tests/performance/test_concurrent_connections.py` | ⚠️ In Progress |
| REQ-SEC-001 | Security | `tests/security/test_authentication.py` | ⚠️ In Progress |

---

## 4. Epic E6 File Management Requirements Details

### 4.1 REQ-FUNC-008: JSON-RPC list_recordings Method
**Description:** Server shall provide JSON-RPC method `list_recordings` to enumerate available recording files in `/opt/camera-service/recordings` directory.

**Parameters:** 
- `limit` (optional): Maximum number of files to return (default: 100)
- `offset` (optional): Number of files to skip for pagination (default: 0)

**Response:**
```json
{
  "files": [
    {
      "filename": "camera0_2025-01-15_14-30-00.mp4",
      "size": 1048576,
      "timestamp": "2025-01-15T14:30:00Z",
      "duration": 30.5,
      "download_url": "/files/recordings/camera0_2025-01-15_14-30-00.mp4"
    }
  ],
  "total_count": 150,
  "has_more": true
}
```

**Error Handling:**
- Directory not accessible: Return error with code -32001
- Permission denied: Return error with code -32002
- Invalid parameters: Return error with code -32602

### 4.2 REQ-FUNC-009: JSON-RPC list_snapshots Method
**Description:** Server shall provide JSON-RPC method `list_snapshots` to enumerate available snapshot files in `/opt/camera-service/snapshots` directory.

**Parameters:**
- `limit` (optional): Maximum number of files to return (default: 100)
- `offset` (optional): Number of files to skip for pagination (default: 0)

**Response:**
```json
{
  "files": [
    {
      "filename": "camera0_2025-01-15_14-30-00.jpg",
      "size": 524288,
      "timestamp": "2025-01-15T14:30:00Z",
      "download_url": "/files/snapshots/camera0_2025-01-15_14-30-00.jpg"
    }
  ],
  "total_count": 75,
  "has_more": false
}
```

**Error Handling:**
- Directory not accessible: Return error with code -32001
- Permission denied: Return error with code -32002
- Invalid parameters: Return error with code -32602

### 4.3 REQ-FUNC-010: HTTP Recording File Download Endpoint
**Description:** Server shall provide HTTP endpoint `/files/recordings/{filename}` for downloading recording files.

**Features:**
- MIME type detection for video files (mp4, avi, mov, etc.)
- Content-Disposition header for proper file downloads
- 404 response for non-existent files
- File access logging for security audit trail
- Support for range requests for large files

**Response Headers:**
```
Content-Type: video/mp4
Content-Disposition: attachment; filename="camera0_2025-01-15_14-30-00.mp4"
Content-Length: 1048576
Accept-Ranges: bytes
```

### 4.4 REQ-FUNC-011: HTTP Snapshot File Download Endpoint
**Description:** Server shall provide HTTP endpoint `/files/snapshots/{filename}` for downloading snapshot files.

**Features:**
- MIME type detection for image files (jpg, png, bmp, etc.)
- Content-Disposition header for proper file downloads
- 404 response for non-existent files
- File access logging for security audit trail

**Response Headers:**
```
Content-Type: image/jpeg
Content-Disposition: attachment; filename="camera0_2025-01-15_14-30-00.jpg"
Content-Length: 524288
```

### 4.5 REQ-FUNC-012: Nginx Routing for File Endpoints
**Description:** Server nginx configuration shall include location blocks for file download endpoints while preserving existing functionality.

**Configuration Requirements:**
- `/files/recordings/` location block pointing to server file handler
- `/files/snapshots/` location block pointing to server file handler
- Preserve existing WebSocket routing on port 8002
- Preserve existing health endpoint routing on port 8003
- Maintain SSL/HTTPS functionality for all endpoints

**Security Considerations:**
- Prevent directory traversal attacks
- Implement proper file access controls
- Log all file download requests

---

## 5. Requirements Validation Status

### 5.1 Epic E6 Requirements Status
- **REQ-FUNC-008:** ✅ Implemented - JSON-RPC list_recordings method operational
- **REQ-FUNC-009:** ✅ Implemented - JSON-RPC list_snapshots method operational  
- **REQ-FUNC-010:** ✅ Implemented - HTTP recording file download endpoint operational
- **REQ-FUNC-011:** ✅ Implemented - HTTP snapshot file download endpoint operational
- **REQ-FUNC-012:** ✅ Implemented - Nginx routing for file endpoints operational

### 5.2 Validation Criteria
- All file management API methods shall be tested with real file system operations
- File download endpoints shall be validated with various file formats and sizes
- Nginx configuration shall be tested for proper routing and SSL functionality
- Performance impact on existing server operations shall be measured
- Security validation shall include directory traversal prevention testing

---

**Document Status:** Updated for Epic E6 file management requirements (COMPLETED)
**Last Updated:** 2025-01-15
**Next Review:** Epic E6 completed successfully
