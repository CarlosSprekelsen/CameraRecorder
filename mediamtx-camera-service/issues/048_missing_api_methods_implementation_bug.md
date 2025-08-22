# Bug Report: Missing API Methods Implementation

**Bug ID:** 048  
**Title:** Missing API Methods Implementation  
**Severity:** Medium  
**Category:** API/Implementation  
**Status:** Identified  

## Summary

Several API methods that are required by the updated requirements baseline and documented in the API specification are not implemented in the WebSocket server. These methods return "Method not found" errors when called. The requirements have been updated to include comprehensive file lifecycle management capabilities.

## Detailed Description

### Root Cause
The following API methods are missing from the WebSocket server implementation:
- `get_recording_info`
- `get_snapshot_info` 
- `delete_recording`
- `delete_snapshot`
- `get_storage_info`
- `set_retention_policy`
- `cleanup_old_files`

These methods are required by the updated requirements baseline (REQ-CLIENT-034 through REQ-CLIENT-041) and documented in the API specification. They are essential for comprehensive file lifecycle management and storage space control.

### Impact
- Integration tests fail with "Method not found" errors
- API functionality is incomplete for file lifecycle management
- Client applications cannot access file deletion and management features
- Storage system will grow indefinitely without cleanup capabilities
- No storage space monitoring or retention policy enforcement
- Critical operational risk of storage exhaustion

### Evidence
Test failures showing "Method not found" errors:
```
FAILED tests/integration/test_critical_interfaces.py::test_get_recording_info_success - AssertionError: Response should contain 'result' field
FAILED tests/integration/test_critical_interfaces.py::test_get_snapshot_info_success - AssertionError: Response should contain 'result' field  
FAILED tests/integration/test_critical_interfaces.py::test_delete_recording_success - AssertionError: Response should contain 'result' field
```

## Recommended Actions

### Option 1: Implement Missing Methods (Recommended)
1. **Implement `get_recording_info` method**
   - Add method to WebSocket server
   - Return recording file metadata (filename, size, duration, timestamp, etc.)
   - Follow API specification format
   - Require viewer role authentication

2. **Implement `get_snapshot_info` method**
   - Add method to WebSocket server
   - Return snapshot file metadata (filename, size, resolution, timestamp, etc.)
   - Follow API specification format
   - Require viewer role authentication

3. **Implement `delete_recording` method**
   - Add method to WebSocket server
   - Handle file deletion with proper error handling
   - Return deletion status
   - Require operator role authentication

4. **Implement `delete_snapshot` method**
   - Add method to WebSocket server
   - Handle snapshot file deletion with proper error handling
   - Return deletion status
   - Require operator role authentication

5. **Implement `get_storage_info` method**
   - Add method to WebSocket server
   - Return storage space information and usage statistics
   - Require admin role authentication

6. **Implement `set_retention_policy` method**
   - Add method to WebSocket server
   - Configure file retention policies for automatic cleanup
   - Require admin role authentication

7. **Implement `cleanup_old_files` method**
   - Add method to WebSocket server
   - Manually trigger cleanup based on retention policies
   - Require admin role authentication

### Option 2: Remove Tests for Unimplemented Methods
- Remove or skip tests for unimplemented methods
- Update API documentation to reflect current implementation
- Mark methods as "not yet implemented"

### Option 3: Implement Stub Methods
- Add stub implementations that return appropriate error responses
- Document methods as "not yet implemented"
- Allow tests to pass with expected error responses

## Implementation Priority

**Critical Priority:**
- `delete_recording` - Required for file cleanup operations (prevents storage exhaustion)
- `delete_snapshot` - Required for file cleanup operations (prevents storage exhaustion)

**High Priority:**
- `get_recording_info` - Required for recording management
- `get_snapshot_info` - Required for snapshot management
- `get_storage_info` - Required for storage monitoring

**Medium Priority:**
- `set_retention_policy` - Required for automated cleanup configuration
- `cleanup_old_files` - Required for automated cleanup execution

## Test Validation

After implementation, validate with:
```bash
python3 -m pytest tests/integration/test_critical_interfaces.py::test_get_recording_info_success -v
python3 -m pytest tests/integration/test_critical_interfaces.py::test_get_snapshot_info_success -v
python3 -m pytest tests/integration/test_critical_interfaces.py::test_delete_recording_success -v
```

## Conclusion

This is a **critical-priority implementation gap** where required API functionality for file lifecycle management is missing. The missing methods are essential for preventing storage exhaustion and providing comprehensive file management capabilities. The requirements baseline has been updated to include these capabilities, and the API specification has been documented. Implementation should be prioritized to address the operational risk of unlimited storage growth.
