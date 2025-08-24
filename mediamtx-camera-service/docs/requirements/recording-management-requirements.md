# Recording Management Requirements Document

**Version:** 1.0  
**Date:** 2025-01-15  
**Role:** IV&V  
**Status:** 🚀 RECORDING MANAGEMENT REQUIREMENTS ESTABLISHED  
**Related Documents:** `docs/requirements/requirements-baseline.md`, `docs/api/json-rpc-methods.md`, `docs/architecture/overview.md`

---

## Executive Summary

This document defines comprehensive requirements for recording management in the MediaMTX Camera Service, addressing resource protection, file rotation, storage monitoring, and user experience. These requirements ensure reliable recording operations while preventing system resource exhaustion and providing clear user feedback.

---

## 1. Recording State Management Requirements

### REQ-REC-001: Recording State Management
**Requirement:** The system SHALL implement comprehensive recording state management to prevent conflicts and ensure reliable operation

#### REQ-REC-001.1: Per-Device Recording Limits
**Requirement:** The system SHALL prevent multiple simultaneous recordings on the same camera device
**Specifications:**
- **State Tracking:** Track active recording sessions per camera with session_id and device identifier
- **Conflict Prevention:** Reject new recording requests for cameras already recording
- **Session Management:** Maintain recording session state throughout recording lifecycle
- **Cleanup:** Properly clean up recording sessions on completion or failure

**Acceptance Criteria:**
- Only one recording session allowed per camera device
- Recording state properly tracked and persisted
- Session cleanup occurs on recording completion or failure
- Recording conflicts properly detected and prevented

#### REQ-REC-001.2: Error Handling for Recording Conflicts
**Requirement:** The system SHALL return appropriate error responses when recording conflicts occur
**Specifications:**
- **Error Code:** Return error code -1006 for recording conflicts
- **User Message:** Provide user-friendly message "Camera is currently recording"
- **Session Data:** Include camera_id and session_id in error response
- **API Alignment:** Use consistent camera identifiers in all responses

**Acceptance Criteria:**
- Error code -1006 returned for recording conflicts
- User-friendly error messages without technical device details
- Consistent camera identifiers used in all API responses
- Session information included in error responses

#### REQ-REC-001.3: Recording Status Integration
**Requirement:** The system SHALL integrate recording status into camera information responses
**Specifications:**
- **Camera List:** Include recording status in `get_camera_list` responses
- **Camera Status:** Include recording status in `get_camera_status` responses
- **Status Fields:** Include recording: true/false, session_id, current_file, elapsed_time
- **Real-time Updates:** Provide real-time recording status updates via notifications

**Acceptance Criteria:**
- Recording status included in camera list and status responses
- Real-time status updates provided via WebSocket notifications
- Recording metadata properly tracked and reported
- Status information accurate and up-to-date

---

## 2. File Rotation Management Requirements

### REQ-REC-002: File Rotation Management
**Requirement:** The system SHALL implement configurable file rotation to manage recording file sizes and improve reliability

#### REQ-REC-002.1: Configurable File Rotation
**Requirement:** The system SHALL create new recording files at configurable intervals during continuous recording
**Specifications:**
- **Default Interval:** 30 minutes per recording file
- **Configurability:** Rotation interval configurable via `RECORDING_ROTATION_MINUTES` environment variable
- **Continuity:** Maintain recording continuity across file rotations without interruption
- **Session Preservation:** Preserve recording session across file rotations

**Acceptance Criteria:**
- New files created at configurable intervals (default: 30 minutes)
- Recording continuity maintained across file rotations
- Configuration changes applied without service restart
- Session information preserved across rotations

#### REQ-REC-002.2: Timestamped File Naming
**Requirement:** The system SHALL generate timestamped filenames for recording files
**Specifications:**
- **Naming Format:** `camera{camera_id}_{YYYY-MM-DD_HH-MM-SS}.mp4`
- **Timestamp Accuracy:** Use precise timestamps for file creation
- **Sequential Ordering:** Ensure files are properly ordered by creation time
- **Metadata Preservation:** Include timestamp information in file metadata

**Acceptance Criteria:**
- Files named with accurate timestamps
- Naming format consistent and sortable
- Timestamp information preserved in metadata
- Files properly ordered by creation time

#### REQ-REC-002.3: Recording Continuity
**Requirement:** The system SHALL maintain recording continuity across file rotations
**Specifications:**
- **No Interruption:** Recording continues without interruption during file rotation
- **Session Continuity:** Recording session maintained across multiple files
- **Metadata Consistency:** Consistent metadata across all files in session
- **Progress Tracking:** Accurate progress tracking across file rotations

**Acceptance Criteria:**
- Recording continues without interruption during rotation
- Session information consistent across all files
- Progress tracking accurate across file boundaries
- Metadata properly maintained across rotations

---

## 3. Storage Protection Requirements

### REQ-REC-003: Storage Protection
**Requirement:** The system SHALL implement comprehensive storage protection to prevent system resource exhaustion

#### REQ-REC-003.1: Storage Space Validation
**Requirement:** The system SHALL validate available storage space before starting recordings
**Specifications:**
- **Pre-recording Check:** Verify sufficient storage space before starting any recording
- **Threshold Validation:** Check against configurable storage thresholds
- **Error Prevention:** Prevent recording start when storage is insufficient
- **User Notification:** Provide clear error messages for storage issues

**Acceptance Criteria:**
- Storage space validated before recording start
- Configurable thresholds properly enforced
- Clear error messages for storage issues
- Recording prevented when storage insufficient

#### REQ-REC-003.2: Configurable Storage Thresholds
**Requirement:** The system SHALL implement configurable storage thresholds for warnings and blocking
**Specifications:**
- **Warning Threshold:** 80% storage usage (configurable via `STORAGE_WARN_PERCENT`)
- **Blocking Threshold:** 90% storage usage (configurable via `STORAGE_BLOCK_PERCENT`)
- **Critical Threshold:** 95% storage usage (hard-coded for system protection)
- **Threshold Application:** Apply thresholds consistently across all operations

**Acceptance Criteria:**
- Configurable thresholds properly applied
- System protection maintained at critical levels
- Thresholds consistent across all operations
- Configuration changes applied without restart

#### REQ-REC-003.3: Storage Error Handling
**Requirement:** The system SHALL provide appropriate error responses for storage-related issues
**Specifications:**
- **Error Code -1008:** "Storage space is low" when below 10% available
- **Error Code -1010:** "Storage space is critical" when below 5% available
- **User-Friendly Messages:** Clear, non-technical error messages
- **Actionable Information:** Provide guidance for resolving storage issues

**Acceptance Criteria:**
- Appropriate error codes returned for storage issues
- User-friendly error messages provided
- Clear guidance for resolving storage issues
- Error responses consistent with API standards

#### REQ-REC-003.4: No Auto-Deletion Policy
**Requirement:** The system SHALL NOT automatically delete any recording files
**Specifications:**
- **User Control:** Users maintain full control over their recording data
- **No Automatic Cleanup:** No automatic deletion of recordings based on age or size
- **Manual Management:** Users responsible for managing their recording storage
- **Data Preservation:** All recording data preserved until user action

**Acceptance Criteria:**
- No automatic deletion of recording files
- User control maintained over all recording data
- Data preservation until explicit user action
- Clear user responsibility for storage management

---

## 4. Resource Monitoring Requirements

### REQ-REC-004: Resource Monitoring
**Requirement:** The system SHALL implement comprehensive resource monitoring during recording operations

#### REQ-REC-004.1: Storage Monitoring
**Requirement:** The system SHALL monitor disk space usage during recording operations
**Specifications:**
- **Real-time Monitoring:** Monitor storage usage during active recordings
- **Threshold Tracking:** Track usage against configured thresholds
- **Alert Generation:** Generate alerts when thresholds exceeded
- **Logging:** Log storage events with structured logging

**Acceptance Criteria:**
- Real-time storage monitoring during recordings
- Threshold tracking and alert generation
- Structured logging of storage events
- Monitoring data available via API

#### REQ-REC-004.2: Storage Information API
**Requirement:** The system SHALL provide storage usage information via API
**Specifications:**
- **get_storage_info Method:** Provide total, used, and available storage space
- **Percentage Calculation:** Calculate and report storage usage percentages
- **Threshold Status:** Include threshold status in storage information
- **Real-time Data:** Provide current storage information on request

**Acceptance Criteria:**
- Storage information available via API
- Accurate storage calculations provided
- Threshold status included in responses
- Real-time data available on request

#### REQ-REC-004.3: Health Integration
**Requirement:** The system SHALL integrate storage status into health monitoring
**Specifications:**
- **Health Check Inclusion:** Include storage status in health check responses
- **Threshold Monitoring:** Monitor storage thresholds in health checks
- **Status Reporting:** Report storage status in system health information
- **Alert Integration:** Integrate storage alerts with health monitoring

**Acceptance Criteria:**
- Storage status included in health checks
- Threshold monitoring in health system
- Status reporting in health information
- Alert integration with health monitoring

---

## 5. User Experience Requirements

### REQ-REC-005: User Experience and API Alignment
**Requirement:** The system SHALL provide excellent user experience with API-aligned interactions

#### REQ-REC-005.1: User-Friendly Error Messages
**Requirement:** The system SHALL provide user-friendly error messages without technical details
**Specifications:**
- **Non-Technical Language:** Use user-friendly language in error messages
- **No Device Details:** Avoid technical device path information in user messages
- **Actionable Guidance:** Provide clear guidance for resolving issues
- **Consistent Messaging:** Use consistent error message format

**Acceptance Criteria:**
- User-friendly error messages provided
- No technical device details in user messages
- Clear guidance for issue resolution
- Consistent error message format

#### REQ-REC-005.2: Recording Progress Information
**Requirement:** The system SHALL provide comprehensive recording progress information
**Specifications:**
- **Current File:** Report current recording file information
- **Elapsed Time:** Track and report recording elapsed time
- **File Size:** Monitor and report current file size
- **Session Information:** Provide complete session information

**Acceptance Criteria:**
- Current file information provided
- Accurate elapsed time tracking
- File size monitoring and reporting
- Complete session information available

#### REQ-REC-005.3: Real-time Notifications
**Requirement:** The system SHALL provide real-time recording status notifications
**Specifications:**
- **WebSocket Notifications:** Send recording status updates via WebSocket
- **Status Changes:** Notify on all recording status changes
- **Progress Updates:** Provide periodic progress updates
- **Error Notifications:** Notify on recording errors and issues

**Acceptance Criteria:**
- Real-time notifications via WebSocket
- Status change notifications provided
- Progress updates available
- Error notifications sent

---

## 6. Configuration Requirements

### REQ-REC-006: Configuration Management
**Requirement:** The system SHALL support configurable recording management parameters

#### REQ-REC-006.1: Environment Variable Configuration
**Specifications:**
- **RECORDING_ROTATION_MINUTES:** File rotation interval (default: 30)
- **STORAGE_WARN_PERCENT:** Storage warning threshold (default: 80)
- **STORAGE_BLOCK_PERCENT:** Storage blocking threshold (default: 90)
- **Configuration Validation:** Validate configuration values on startup

**Acceptance Criteria:**
- Environment variables properly read and applied
- Default values used when not specified
- Configuration validation on startup
- Configuration changes applied without restart

---

## 7. Error Code Specifications

### Enhanced Error Codes for Recording Operations

| Error Code | Message | Description | Data Fields |
|------------|---------|-------------|-------------|
| -1006 | "Camera is currently recording" | Device already has active recording | camera_id, session_id |
| -1008 | "Storage space is low" | Available storage below 10% | available_space, total_space |
| -1010 | "Storage space is critical" | Available storage below 5% | available_space, total_space |

### Error Response Format
```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -1006,
    "message": "Camera is currently recording",
    "data": {
      "camera_id": "camera0",
      "session_id": "550e8400-e29b-41d4-a716-446655440000"
    }
  },
  "id": 5
}
```

---

## 8. API Response Enhancements

### Updated Camera Status Response
```json
{
  "jsonrpc": "2.0",
  "result": {
    "camera_id": "camera0",
    "device": "/dev/video0",
    "status": "connected",
    "recording": true,
    "recording_session": "550e8400-e29b-41d4-a716-446655440000",
    "current_file": "camera0_2025-01-15_14-30-00.mp4",
    "elapsed_time": 1800
  },
  "id": 5
}
```

---

## 9. Requirements Traceability

### Mapping to Existing Requirements
- **REQ-CLIENT-010**: Automatic file creation when maximum file size reached
- **REQ-CLIENT-035**: Configurable retention policies
- **REQ-CLIENT-036**: Storage space monitoring and alerts
- **REQ-SEC-026**: File size limits enforcement
- **REQ-API-028**: Storage information method
- **REQ-API-033**: Real-time storage monitoring

### Test Coverage Requirements
- Unit tests for recording state management
- Integration tests for file rotation
- Storage monitoring tests
- Error handling tests
- Configuration validation tests

---

**Document Status:** Complete recording management requirements established
**Last Updated:** 2025-01-15
**Next Review:** After implementation validation
