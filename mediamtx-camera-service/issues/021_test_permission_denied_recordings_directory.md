# Issue 021: Test Permission Denied for Recordings Directory

**Priority:** HIGH
**Category:** Test Infrastructure
**Status:** OPEN
**Created:** August 21, 2025
**Discovered By:** Test Suite Execution

## Description
Unit tests are failing because they try to write to `/opt/camera-service/recordings/` directory which requires elevated permissions. This prevents proper test execution and validation.

## Error Details
**Error:** `PermissionError: [Errno 13] Permission denied: '/opt/camera-service/recordings/.write_test_*'`
**Location:** `src/mediamtx_wrapper/controller.py:268`
**Test:** Multiple tests in `test_disconnect_handling.py`
**Root Cause:** Tests trying to write to system directory without proper permissions

## Ground Truth Analysis
### API Documentation Evidence
**`docs/api/json-rpc-methods.md`** defines the public API as JSON-RPC 2.0 over WebSocket:
- Tests should validate **API behavior**, not file system permissions
- Unit tests should use **mock/stub components** for external dependencies

### Architecture Evidence
**`docs/architecture/overview.md`** shows the system architecture:
- Tests should validate **component interfaces**, not system-level operations
- Unit tests should isolate components from external dependencies

### Requirements Evidence
**`docs/requirements/*.md`** contains no references to file system testing:
- Requirements focus on **functional behavior**
- Tests should validate **business logic**, not system permissions

## Current Test Code (INCORRECT)
```python
# Tests are trying to write to system directories
await real_service_manager.start()  # WRONG: tries to write to /opt/camera-service/recordings/
```

## Correct Test Approach (Based on Ground Truth)
```python
# Tests should use mock components or temporary directories
with tempfile.TemporaryDirectory() as temp_dir:
    # Configure service to use temporary directory
    config.recordings.path = temp_dir
    await service_manager.start()
```

## Impact
- **Test Reliability:** Tests fail due to permission issues
- **Unit Testing:** Missing validation of component behavior
- **API Compliance:** Tests don't validate the actual business logic
- **Maintenance:** Tests break when run without elevated permissions

## Affected Test Files
- `tests/unit/test_camera_discovery/test_disconnect_handling.py` - Multiple permission failures
- Other unit tests may have similar issues with file system access

## Root Cause
The tests were designed to use real file system operations instead of using mock components or temporary directories. This violates the principle that unit tests should isolate components from external dependencies.

## Proposed Solution
1. **Use temporary directories** for file system operations
2. **Mock external dependencies** where appropriate
3. **Configure test environment** to use writable locations
4. **Isolate components** from system-level operations
5. **Focus on business logic** validation, not system permissions

## Acceptance Criteria
- [ ] Tests use temporary directories for file operations
- [ ] Tests mock external dependencies appropriately
- [ ] Tests validate business logic, not system permissions
- [ ] No permission errors during test execution
- [ ] Tests can run without elevated privileges

## Implementation Notes
- Unit tests should validate **component behavior**
- Use **temporary directories** for file system operations
- **Mock external dependencies** for isolation
- Focus on **business logic validation**

## Ground Truth Compliance
- ✅ **API Documentation**: Tests will validate documented behavior
- ✅ **Architecture**: Tests will validate component interfaces
- ✅ **Requirements**: Tests will validate functional requirements

## Testing
- Verify tests use temporary directories
- Confirm no permission errors occur
- Validate business logic is tested
- Ensure tests can run without elevated privileges
