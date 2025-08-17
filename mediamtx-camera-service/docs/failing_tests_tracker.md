# Failing Tests Tracker - Systematic Fix Approach

**Status**: 🟡 STARTING FROM BASE ZERO  
**Last Updated**: 2025-08-17  
**Total Failing Tests**: 54  
**Approach**: One test at a time, controlled decision making with GLOBAL requirements view

## Decision Framework

For each failing test, we will:
1. **ANALYZE**: What does this test actually validate?
2. **CHECK REQUIREMENTS**: Does it map to a real requirement in the system?
3. **DECIDE**: 
   - 🟢 **FIX**: Test validates real behavior, simple fix needed
   - 🟡 **REDESIGN**: Test designed to pass, needs redesign for real behavior
   - 🔴 **DELETE**: Test validates non-existent functionality (requires authorization)
   - ⏭️ **SKIP**: Complex but useful, defer for later
4. **EXECUTE**: Fix/redesign/delete one test at a time
5. **VERIFY**: Confirm fix works, update table
6. **CHECK GLOBAL IMPACT**: Ensure no requirements coverage is lost

## Global Requirements Coverage (ALL Requirements)

### Camera Requirements (REQ-CAM-*)
- **REQ-CAM-001**: Camera discovery automatic
- **REQ-CAM-002**: Frame rate extraction  
- **REQ-CAM-003**: Resolution detection
- **REQ-CAM-004**: Camera status monitoring
- **REQ-CAM-005**: Advanced camera capabilities

### Configuration Requirements (REQ-CONFIG-*)
- **REQ-CONFIG-001**: Configuration validation
- **REQ-CONFIG-002**: Hot reload configuration
- **REQ-CONFIG-003**: Configuration error handling

### Error Handling Requirements (REQ-ERROR-*)
- **REQ-ERROR-001**: WebSocket MediaMTX failures
- **REQ-ERROR-002**: WebSocket client disconnection
- **REQ-ERROR-003**: MediaMTX service unavailability
- **REQ-ERROR-004**: System stability during config failures
- **REQ-ERROR-005**: System stability during logging failures
- **REQ-ERROR-006**: System stability during WebSocket failures
- **REQ-ERROR-007**: System stability during MediaMTX failures
- **REQ-ERROR-008**: System stability during service failures
- **REQ-ERROR-009**: Error propagation handling
- **REQ-ERROR-010**: Error recovery mechanisms

### Health Monitoring Requirements (REQ-HEALTH-*)
- **REQ-HEALTH-001**: Health monitoring
- **REQ-HEALTH-002**: Structured logging
- **REQ-HEALTH-003**: Correlation IDs

### Integration Requirements (REQ-INT-*)
- **REQ-INT-001**: System integration
- **REQ-INT-002**: MediaMTX service integration

### Media Requirements (REQ-MEDIA-*)
- **REQ-MEDIA-001**: Media processing
- **REQ-MEDIA-002**: Stream management
- **REQ-MEDIA-003**: Health monitoring
- **REQ-MEDIA-004**: Service failure handling

### Service Requirements (REQ-SVC-*)
- **REQ-SVC-001**: Service lifecycle management
- **REQ-SVC-002**: Startup/shutdown handling
- **REQ-SVC-003**: Component orchestration

### WebSocket Requirements (REQ-WS-*)
- **REQ-WS-001**: WebSocket server aggregation
- **REQ-WS-002**: WebSocket capability metadata
- **REQ-WS-003**: WebSocket status aggregation
- **REQ-WS-004**: WebSocket notifications
- **REQ-WS-005**: WebSocket message handling
- **REQ-WS-006**: WebSocket error handling
- **REQ-WS-007**: WebSocket connection management

## Current Failing Tests (To Be Analyzed One by One)

| # | Test File | Test Method | Status | Decision | Requirements Impact | Action Required | Notes |
|---|-----------|-------------|--------|----------|-------------------|-----------------|-------|
| 1 | `test_server_notifications.py` | `test_notification_correlation_id_handling` | 🟢 **FIXED** | ✅ **COMPLETED** | REQ-WS-004, REQ-WS-005, REQ-WS-006, REQ-WS-007 | **REDESIGNED** - Now uses real WebSocket communication | Tests real correlation ID propagation through WebSocket |
| 2 | `test_server_notifications.py` | `test_recording_status_notification_field_filtering_with_real_client` | 🟢 **FIXED** | ✅ **COMPLETED** | REQ-WS-005 | **FIXED** - Removed MediaMTX dependencies | Tests real WebSocket field filtering without MediaMTX |
| 3 | `test_server_notifications.py` | `test_broadcast_notification_to_real_clients` | 🟢 **FIXED** | ✅ **COMPLETED** | REQ-WS-004, REQ-WS-007 | **FIXED** - Removed MediaMTX dependencies | Tests real WebSocket broadcasting without MediaMTX |
| 4 | `test_server_notifications.py` | `test_send_notification_to_specific_real_client` | 🟢 **FIXED** | ✅ **COMPLETED** | REQ-WS-004, REQ-WS-007 | **FIXED** - Removed MediaMTX dependencies | Tests real WebSocket targeted notifications without MediaMTX |
| 5 | `test_hybrid_monitor_enhanced.py` | `test_camera_status_monitoring_adaptive_polling_interval` | 🔴 **DELETED** | ✅ **COMPLETED** | REQ-CAM-004 | **DELETED** - Tests non-existent adaptive polling functionality | Tests features that don't exist in implementation |
| 6 | `test_controller_stream_operations_real.py` | `test_unlimited_duration_recording_f1_2_2` | 🟢 **FIXED** | ✅ **COMPLETED** | F1.2.2, REQ-MEDIA-002 | **FIXED** - Updated expected error message | Tests real MediaMTX error handling |
| 7 | `test_controller_stream_operations_real.py` | `test_timed_recording_f1_2_3` | 🟢 **FIXED** | ✅ **COMPLETED** | F1.2.3, REQ-MEDIA-002 | **FIXED** - Updated expected error message | Tests real MediaMTX error handling |
| 8 | `test_controller_stream_operations_real.py` | `test_manual_recording_stop_f1_2_4` | 🟢 **FIXED** | ✅ **COMPLETED** | F1.2.4, REQ-MEDIA-002 | **FIXED** - Updated expected error message | Tests real MediaMTX error handling |
| 9 | `test_controller_stream_operations_real.py` | `test_recording_session_management_f1_2_5` | 🟢 **FIXED** | ✅ **COMPLETED** | F1.2.5, REQ-MEDIA-002 | **FIXED** - Updated expected error message | Tests real MediaMTX error handling |
| 10 | `test_controller_stream_operations_real.py` | `test_invalid_stream_recording_error` | 🟢 **FIXED** | ✅ **COMPLETED** | REQ-MEDIA-002, REQ-MEDIA-009, REQ-MTX-009 | **FIXED** - Updated expected error message | Tests real MediaMTX error handling |
| 11 | `test_controller_stream_operations_real.py` | `test_invalid_format_error` | 🟢 **FIXED** | ✅ **COMPLETED** | REQ-MEDIA-009, REQ-MTX-009 | **FIXED** - Updated expected error message | Tests real MediaMTX error handling |
| 12 | `test_controller_stream_operations_real.py` | `test_full_recording_lifecycle_integration` | 🟢 **FIXED** | ✅ **COMPLETED** | REQ-MEDIA-002, REQ-MEDIA-005, REQ-MTX-001 | **FIXED** - Updated expected error message | Tests real MediaMTX error handling |
| 13 | `test_real_integration_fixed.py` | `test_real_camera_status_integration` | 🟢 **FIXED** | ✅ **COMPLETED** | REQ-WS-001 | **FIXED** - Fixed async fixture usage | Tests real WebSocket server integration |
| 14 | `test_real_integration_fixed.py` | `test_real_camera_list_integration` | 🟢 **FIXED** | ✅ **COMPLETED** | REQ-WS-002 | **FIXED** - Fixed async fixture usage | Tests real WebSocket server integration |
| 15 | `test_real_integration_fixed.py` | `test_mediamtx_integration_with_mock` | 🔴 **DELETED** | ✅ **COMPLETED** | REQ-WS-001, REQ-WS-003 | **DELETED** - Uses mocks, requirements already covered | Tests mock behavior, not real system behavior |
| 16 | `test_real_integration_fixed.py` | `test_error_handling_with_invalid_device` | 🟢 **FIXED** | ✅ **COMPLETED** | REQ-ERROR-004 | **FIXED** - Fixed async fixture usage | Tests real error handling |
| 17 | `test_real_integration_fixed.py` | `test_missing_device_parameter_handling` | 🟢 **FIXED** | ✅ **COMPLETED** | REQ-ERROR-004 | **FIXED** - Fixed async fixture usage | Tests real error handling |
| 18 | `test_real_integration_fixed.py` | `test_stream_name_generation` | 🟢 **FIXED** | ✅ **COMPLETED** | REQ-WS-001 | **FIXED** - Fixed async fixture usage | Tests real stream name generation |
| 19 | **PENDING** | **PENDING** | 🔄 **WAITING** | 🔄 **PENDING** | **NEEDS ANALYSIS** | **NEEDS ANALYSIS** | **NEEDS ANALYSIS** |

## Progress Summary

- **Total Failing Tests**: 54
- **Analyzed**: 18
- **Fixed**: 15
- **Redesigned**: 1
- **Deleted**: 2
- **Skipped**: 0
- **Remaining**: 36

## Requirements Coverage Status

### ✅ Currently Covered (Working Tests)
- **REQ-CAM-001**: Camera discovery automatic - COVERED by `test_capability_detection.py`
- **REQ-CAM-003**: Resolution detection - COVERED by `test_capability_detection.py`
- **REQ-CONFIG-001**: Configuration validation - COVERED by `test_configuration_validation.py`
- **REQ-SVC-001**: Service lifecycle management - COVERED by `test_service_manager_lifecycle.py`

### ⚠️ Potentially At Risk (Need Analysis)
- **REQ-CAM-004**: Camera status monitoring - NEEDS ANALYSIS
- **REQ-ERROR-***: All error handling requirements - NEEDS ANALYSIS
- **REQ-HEALTH-***: All health monitoring requirements - NEEDS ANALYSIS
- **REQ-WS-***: All WebSocket requirements - NEEDS ANALYSIS
- **REQ-MEDIA-***: All media requirements - NEEDS ANALYSIS

## Next Steps

1. **GET FIRST FAILING TEST**: Run test suite and identify first failing test
2. **ANALYZE**: What does this test validate?
3. **CHECK REQUIREMENTS**: Which requirements does it cover?
4. **DECIDE**: FIX/REDESIGN/DELETE/SKIP
5. **EXECUTE**: Take action on one test
6. **VERIFY**: Confirm action worked
7. **UPDATE TABLE**: Remove from failing list
8. **CHECK GLOBAL IMPACT**: Ensure no requirements lost
9. **CONTINUE**: Move to next test

## Authorization Log

| Date | Test | Action | Authorized By | Requirements Impact | Reason |
|------|------|--------|---------------|-------------------|--------|
| 2025-08-17 | `test_notification_correlation_id_handling` | REDESIGN | User | REQ-WS-004, REQ-WS-005, REQ-WS-006, REQ-WS-007 | Converted from mock-based to real WebSocket communication |
| 2025-08-17 | `test_recording_status_notification_field_filtering_with_real_client` | FIX | User | REQ-WS-005 | Removed MediaMTX dependencies for WebSocket field filtering test |
| 2025-08-17 | `test_camera_status_monitoring_adaptive_polling_interval` | DELETE | User | REQ-CAM-004 | Tests non-existent adaptive polling functionality |
| 2025-08-17 | `test_unlimited_duration_recording_f1_2_2` | FIX | User | F1.2.2, REQ-MEDIA-002 | Updated expected error message to match actual system behavior |
| 2025-08-17 | `test_timed_recording_f1_2_3` | FIX | User | F1.2.3, REQ-MEDIA-002 | Updated expected error message to match actual system behavior |
| 2025-08-17 | `test_manual_recording_stop_f1_2_4` | FIX | User | F1.2.4, REQ-MEDIA-002 | Updated expected error message to match actual system behavior |
| 2025-08-17 | `test_recording_session_management_f1_2_5` | FIX | User | F1.2.5, REQ-MEDIA-002 | Updated expected error message to match actual system behavior |
| 2025-08-17 | `test_invalid_stream_recording_error` | FIX | User | REQ-MEDIA-002, REQ-MEDIA-009, REQ-MTX-009 | Updated expected error message to match actual system behavior |
| 2025-08-17 | `test_invalid_format_error` | FIX | User | REQ-MEDIA-009, REQ-MTX-009 | Updated expected error message to match actual system behavior |
| 2025-08-17 | `test_full_recording_lifecycle_integration` | FIX | User | REQ-MEDIA-002, REQ-MEDIA-005, REQ-MTX-001 | Updated expected error message to match actual system behavior |
| 2025-08-17 | `test_real_camera_status_integration` | FIX | User | REQ-WS-001 | Fixed async fixture usage for real WebSocket integration |
| 2025-08-17 | `test_real_camera_list_integration` | FIX | User | REQ-WS-002 | Fixed async fixture usage for real WebSocket integration |
| 2025-08-17 | `test_mediamtx_integration_with_mock` | DELETE | User | REQ-WS-001, REQ-WS-003 | Uses mocks, requirements already covered by real integration tests |
| 2025-08-17 | `test_error_handling_with_invalid_device` | FIX | User | REQ-ERROR-004 | Fixed async fixture usage for real error handling |
| 2025-08-17 | `test_missing_device_parameter_handling` | FIX | User | REQ-ERROR-004 | Fixed async fixture usage for real error handling |
| 2025-08-17 | `test_stream_name_generation` | FIX | User | REQ-WS-001 | Fixed async fixture usage for real stream name generation |

---

**NEXT ACTION**: Get the first failing test from the 54 failing tests and analyze it individually.

**CRITICAL**: Before any deletion, we must verify that the requirements it covers are either:
1. Already covered by other working tests, OR
2. Not actually implemented in the system (non-existent functionality)
