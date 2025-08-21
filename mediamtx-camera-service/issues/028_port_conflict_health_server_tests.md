# Bug Report: Port conflict in health server tests

**Issue ID:** 028  
**Severity:** HIGH  
**Status:** OPEN  
**Date:** 2025-01-06  
**Reporter:** IV&V Team  

## Summary
Multiple test failures occur due to port conflicts when the health server tries to bind to port 8003, which is already in use. This affects multiple service manager lifecycle tests.

## Description
When running the test suite, multiple tests fail with the following error:

```
OSError: [Errno 98] error while attempting to bind on address ('0.0.0.0', 8003): [errno 98] address already in use
```

This affects the following tests:
- `test_real_connect_flow`
- `test_real_disconnect_flow`
- `test_real_mediamtx_failure_keeps_service_running`
- `test_real_capability_metadata`
- `test_real_camera_monitor_integration`
- `test_real_camera_event_processing_latency`
- `test_real_rapid_connect_disconnect_stress`
- `test_real_unknown_camera_event_types`
- `test_real_service_recovery_after_errors`
- `test_real_concurrent_event_processing`

## Root Cause
The health server is configured to bind to port 8003, but this port is already in use by another process or a previous test that didn't properly clean up.

## Impact
- Multiple service manager lifecycle tests fail
- Prevents validation of critical service functionality
- Blocks comprehensive test coverage validation
- Violates testing guide requirements for real system testing

## Steps to Reproduce
1. Run the test suite: `python3 -m pytest tests/unit/test_camera_service/test_service_manager_lifecycle.py -v`
2. Observe port binding errors in multiple tests

## Expected Behavior
Tests should use available ports or properly clean up after themselves to avoid conflicts.

## Actual Behavior
Tests fail due to port conflicts when trying to start the health server.

## Affected Files
- `tests/unit/test_camera_service/test_service_manager_lifecycle.py`
- `src/health_server.py`
- `src/camera_service/service_manager.py`

## Requirements Impact
- Service lifecycle management requirements
- Health monitoring functionality
- Service recovery and error handling
- Real system integration testing

## Fix Required
1. Implement dynamic port allocation for health server tests
2. Ensure proper cleanup of health server resources after each test
3. Add port availability checking before starting health server
4. Consider using unique ports for each test or test isolation

## Testing Guide Compliance
- **Violation**: Tests do not properly isolate system resources
- **Impact**: Prevents execution of service lifecycle validation tests
- **Priority**: HIGH - Service lifecycle testing is critical for system validation
