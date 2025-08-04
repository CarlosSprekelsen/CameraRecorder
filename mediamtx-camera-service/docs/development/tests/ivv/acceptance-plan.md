# S5 End-to-End Acceptance Test Plan

**Version:** 1.0  
**Authors:** Solo Engineer  
**Date:** 2025-08-04  
**Status:** Approved  
**Related Epic/Story:** E1 / S5  

## 1. Overview and Scope

This document defines acceptance criteria and test scenarios for validating the complete end-to-end functionality of the MediaMTX Camera Service, covering the full flow from camera discovery through MediaMTX integration to WebSocket notifications.

**Test Scope:**
- Camera discovery → MediaMTX stream/record/snapshot → WebSocket notification → shutdown/error recovery
- Core API methods: `get_camera_list`, `get_camera_status`, `take_snapshot`, `start_recording`, `stop_recording`
- Real-time notifications: `camera_status_update`, `recording_status_update`
- Health monitoring and error recovery scenarios

**Exclusions:**
- Authentication and security features (deferred to S6)
- Performance benchmarking under load (deferred to S7)
- Multi-camera concurrent scenarios beyond basic validation

## 2. Test Environment Requirements

### System Dependencies
- Ubuntu 22.04+ test environment
- Python 3.10+ with all project dependencies installed
- MediaMTX server (latest stable version)
- USB camera or virtual V4L2 device for testing
- WebSocket client for API testing

### Test Infrastructure
- Camera Service running on localhost:8002
- MediaMTX running on localhost:8554 (RTSP), localhost:8889 (WebRTC)
- Test data directory for recordings and snapshots
- Mock camera device for error injection scenarios

## 3. Core Happy Path Scenarios

### Scenario HP-1: Complete Camera Lifecycle
**Objective:** Validate full camera connect-to-disconnect flow
**Prerequisites:** Clean system state, no cameras detected
**Steps:**
1. Connect USB camera to system
2. Verify camera discovery event within 5 seconds
3. Validate MediaMTX stream configuration creation
4. Check camera_status_update notification with correct schema
5. Verify get_camera_list returns connected camera with stream URLs
6. Disconnect camera
7. Verify camera removal notification
8. Confirm MediaMTX stream cleanup

**Success Criteria:**
- Camera detected within 5 seconds of connection
- Notification contains required fields: device, status, name, resolution, fps, streams
- Stream URLs are accessible and valid
- Clean disconnection with proper resource cleanup

### Scenario HP-2: Streaming and Recording Operations
**Objective:** Validate core media operations
**Prerequisites:** Camera connected and discovered
**Steps:**
1. Get camera status to verify streaming readiness
2. Access RTSP stream URL and verify video data
3. Start recording via start_recording method
4. Verify recording_status_update notification (status=STARTED)
5. Wait 10 seconds for recording capture
6. Stop recording via stop_recording method
7. Verify recording_status_update notification (status=STOPPED)
8. Validate recording file exists and has expected duration
9. Take snapshot via take_snapshot method
10. Verify snapshot file creation and image validity

**Success Criteria:**
- RTSP stream provides valid video data
- Recording start/stop notifications received within 2 seconds
- Recording file duration matches expected timeframe (±2 seconds)
- Snapshot file is valid image format

### Scenario HP-3: WebSocket Client Integration
**Objective:** Validate real-time notification delivery
**Prerequisites:** Camera Service running
**Steps:**
1. Establish WebSocket connection to ws://localhost:8002/ws
2. Send ping method and verify pong response
3. Connect camera while WebSocket client is listening
4. Verify camera_status_update notification received
5. Start recording and verify recording_status_update received
6. Test multiple concurrent WebSocket connections
7. Verify all clients receive notifications

**Success Criteria:**
- WebSocket connection established successfully
- All notifications received by connected clients
- JSON-RPC 2.0 format compliance in all messages
- No message loss during normal operations

## 4. Error Handling and Recovery Scenarios

### Scenario ER-1: MediaMTX Service Disruption
**Objective:** Validate recovery from MediaMTX downtime
**Prerequisites:** Camera connected, MediaMTX running
**Steps:**
1. Verify camera streaming normally
2. Stop MediaMTX service
3. Attempt camera operations (get_status, start_recording)
4. Verify appropriate error responses
5. Restart MediaMTX service
6. Verify automatic recovery and stream restoration
7. Test camera operations resume normally

**Success Criteria:**
- Error responses include meaningful error codes (-1003: MediaMTX error)
- Service recovers automatically within 30 seconds of MediaMTX restart
- No camera state corruption during outage

### Scenario ER-2: Camera Disconnection During Recording
**Objective:** Validate graceful handling of unexpected camera removal
**Prerequisites:** Camera connected and recording
**Steps:**
1. Start recording on connected camera
2. Physically disconnect camera during recording
3. Verify recording_status_update notification (status=ERROR)
4. Verify partial recording file preservation
5. Reconnect camera
6. Verify new recording session can be started

**Success Criteria:**
- Error notification received within 10 seconds
- Partial recording file saved and accessible
- Camera reconnection handled cleanly

### Scenario ER-3: Invalid API Requests
**Objective:** Validate robust error handling for malformed requests
**Prerequisites:** Service running
**Steps:**
1. Send invalid JSON to WebSocket endpoint
2. Send JSON-RPC request with missing required fields
3. Attempt operations on non-existent camera device
4. Send requests with invalid parameter types
5. Verify all error responses include proper JSON-RPC error format

**Success Criteria:**
- All error responses follow JSON-RPC 2.0 error specification
- Error codes match API documentation (-32700, -32600, -32601, -32602, -1000, -1001)
- Service remains stable after error conditions

## 5. Performance and Resource Scenarios

### Scenario PR-1: Resource Usage Validation
**Objective:** Verify service operates within acceptable resource limits
**Prerequisites:** Single camera connected
**Steps:**
1. Monitor CPU and memory usage during normal operations
2. Start recording and monitor resource impact
3. Take multiple snapshots and verify resource cleanup
4. Run continuous streaming for 10 minutes
5. Verify no memory leaks or resource exhaustion

**Success Criteria:**
- CPU usage <20% during normal operations
- Memory usage <100MB total
- No memory leaks detected over 10-minute run
- All temporary files cleaned up properly

### Scenario PR-2: Multiple Client Handling
**Objective:** Validate WebSocket server handles multiple clients
**Prerequisites:** Service running, camera connected
**Steps:**
1. Connect 10 WebSocket clients simultaneously
2. Trigger camera event and verify all clients receive notifications
3. Send API requests from multiple clients concurrently
4. Disconnect clients and verify proper cleanup

**Success Criteria:**
- All clients receive notifications simultaneously
- No cross-client request interference
- Connection cleanup prevents resource leaks

## 6. Integration Smoke Test Implementation

The core integration smoke test implements a subset of these scenarios focusing on the most critical end-to-end paths:

**Covered Scenarios:**
- HP-1: Camera discovery and notification
- HP-2: Basic recording and snapshot operations  
- HP-3: WebSocket notification delivery
- ER-1: Basic error handling

**Test Implementation Location:** `tests/ivv/test_integration_smoke.py`

## 7. Test Execution and Reporting

### Execution Environment
- Run tests on clean Ubuntu 22.04+ system
- Use virtual camera device for reproducible testing
- Execute tests in isolated environment (Docker recommended)

### Success Criteria Matrix
| Scenario | Must Pass | Duration | Dependencies |
|----------|-----------|----------|--------------|
| HP-1 | Yes | <30s | USB camera or virtual device |
| HP-2 | Yes | <60s | Camera + MediaMTX |
| HP-3 | Yes | <30s | WebSocket client |
| ER-1 | Recommended | <90s | Service control permissions |
| ER-2 | Recommended | <45s | Physical camera control |
| ER-3 | Yes | <15s | None |
| PR-1 | Recommended | <15min | Resource monitoring tools |
| PR-2 | Optional | <60s | None |

### Test Report Format
Each test execution must produce:
- Pass/Fail status for each scenario
- Execution duration and resource usage
- Error details for any failures
- Log excerpts for key events
- Screenshot/recording samples where applicable

## 8. Deployment Validation

### Pre-Production Checklist
- [ ] All HP scenarios pass
- [ ] Critical ER scenarios pass  
- [ ] Resource usage within limits
- [ ] Log output follows structured format
- [ ] Configuration validates correctly
- [ ] Service starts/stops cleanly

### Known Limitations
- Authentication testing deferred to S6
- Load testing deferred to S7
- Multi-camera concurrency requires additional test infrastructure

---

**Next Steps:**
1. Implement core integration smoke test
2. Execute initial test run and document results
3. Address any discovered gaps or failures
4. Prepare test automation for CI/CD integration