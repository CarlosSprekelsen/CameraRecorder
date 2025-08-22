# Bug Report: Port Binding Conflicts in Test Environment

**Bug ID:** 051  
**Title:** Port Binding Conflicts in Test Environment  
**Severity:** Medium  
**Category:** Testing/Environment  
**Status:** Identified  

## Summary

Multiple tests are failing due to port binding conflicts, specifically port 8003 (health server port) being already in use. This prevents tests from starting their own service instances and causes widespread test failures.

## Detailed Description

### Root Cause
The test environment has port conflicts where multiple tests try to use the same ports (primarily port 8003 for health server). This suggests:
1. A service is already running on port 8003
2. Tests are not properly cleaning up after themselves
3. Test isolation is not working correctly
4. Dynamic port allocation is not being used consistently

### Impact
- Multiple integration tests fail with port binding errors
- Test suite reliability is compromised
- Inability to run tests in parallel
- Test environment instability

### Evidence
Multiple test failures showing port binding conflicts:
```
FAILED tests/integration/test_ffmpeg_integration.py::test_ffmpeg_integration - OSError: [Errno 98] error while attempting to bind on address ('0.0.0.0', 8...
FAILED tests/integration/test_service_manager_e2e.py::test_e2e_connect_disconnect_creates_and_deletes_paths - OSError: [Errno 98] error while attempting to bind on address ('0.0.0.0', 8...
FAILED tests/integration/test_service_manager_e2e.py::test_e2e_resilience_on_mediamtx_failure - OSError: [Errno 98] error while attempting to bind on address ('0.0.0.0', 8...
```

And skipped tests due to port conflicts:
```
SKIPPED [1] tests/integration/test_config_component_integration.py:39: Service already running on port 8003, skipping test to avoid port conflict
SKIPPED [1] tests/integration/test_config_component_integration.py:128: Service already running on port 8003, skipping test to avoid port conflict
```

## Recommended Actions

### Option 1: Fix Test Isolation (Recommended)
1. **Ensure proper test cleanup**
   - Add proper teardown in all test fixtures
   - Ensure services are stopped after each test
   - Add cleanup verification

2. **Use dynamic port allocation consistently**
   - Ensure all tests use `find_free_port()` utility
   - Update any hardcoded port usage
   - Add port conflict detection and resolution

3. **Improve test environment management**
   - Add environment cleanup between test runs
   - Implement proper service lifecycle management
   - Add port availability checking

### Option 2: Implement Test Coordination
- Add test coordination to prevent port conflicts
- Implement test scheduling to avoid concurrent port usage
- Add port reservation system

### Option 3: Use Different Port Ranges
- Assign different port ranges to different test categories
- Implement port range management
- Add port conflict detection and resolution

## Implementation Priority

**High Priority:**
- Fix test cleanup and isolation
- Ensure consistent use of dynamic port allocation
- Add proper service lifecycle management

**Medium Priority:**
- Implement port conflict detection
- Add test coordination mechanisms
- Improve test environment stability

**Low Priority:**
- Add comprehensive port management
- Implement advanced test scheduling
- Add test environment monitoring

## Test Validation

After implementation, validate with:
```bash
python3 -m pytest tests/integration/ -v --tb=no
```

Expected behavior:
- No port binding conflicts
- All tests can run independently
- Proper test isolation maintained

## Technical Details

### Affected Ports
- Port 8003: Health server port (primary conflict)
- Other ports may be affected as well

### Current Issues
- Services not properly cleaned up after tests
- Hardcoded port usage in some tests
- Lack of port conflict detection
- Insufficient test isolation

### Required Fixes
1. Use `find_free_port()` consistently
2. Add proper service cleanup
3. Implement port conflict detection
4. Improve test isolation

## Conclusion

This is a **medium-priority testing environment bug** that affects test reliability and execution. The port binding conflicts need to be resolved to ensure proper test isolation and reliable test execution. This affects the ability to run tests consistently and detect actual software issues versus environment problems.
