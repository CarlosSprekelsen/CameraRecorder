# Issue 084: Missing API Methods Implementation Bug

**Status:** OPEN  
**Priority:** Critical  
**Type:** API Implementation Bug  
**Created:** 2025-01-23  
**Discovered By:** Test Infrastructure API Compliance Validation  
**Assigned To:** Server Team  

## Description

Multiple API methods documented in `docs/api/json-rpc-methods.md` are not implemented in the server, causing widespread test failures and breaking client functionality. The server returns `-32601` (Method not found) errors for documented methods.

## Root Cause Analysis

### API Documentation vs Implementation Mismatch:
- **Documented Methods**: Listed in `docs/api/json-rpc-methods.md` as available
- **Server Reality**: Methods return `-32601` (Method not found) error
- **Impact**: All clients attempting to use these methods will fail

### Missing Methods Identified:

#### **File Management Methods:**
- `get_recording_info` - Returns `-32601` instead of recording metadata
- `get_snapshot_info` - Returns `-32601` instead of snapshot metadata
- `delete_recording` - Returns `-32601` instead of deletion confirmation
- `delete_snapshot` - Returns `-32601` instead of deletion confirmation

#### **Recording Control Methods:**
- `start_recording` - Returns `-32601` instead of recording session
- `list_recordings` - Returns `-32601` instead of recording list
- `list_snapshots` - Returns `-32601` instead of snapshot list

#### **Storage Management Methods:**
- `get_storage_info` - Returns `-32601` instead of storage statistics
- `configure_storage_thresholds` - Returns `-32601` instead of configuration

#### **HTTP Download Methods:**
- `get_download_url` - Returns `-32601` instead of download URLs

## Technical Analysis

### Expected Behavior (from documentation):
```json
{
  "jsonrpc": "2.0",
  "method": "get_recording_info",
  "params": {"recording_id": "rec_123"},
  "id": 1
}
```

Expected Response:
```json
{
  "jsonrpc": "2.0",
  "result": {
    "recording_id": "rec_123",
    "metadata": {...},
    "file_size": 1024000,
    "duration": 60
  },
  "id": 1
}
```

### Actual Behavior (from implementation):
```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32601,
    "message": "Method not found"
  },
  "id": 1
}
```

## Test Evidence

### Affected Test Files:
- `tests/integration/test_critical_interfaces.py` - 4 failures
- `tests/integration/test_file_management_integration.py` - 4 failures
- `tests/integration/test_file_metadata_tracking.py` - 7 failures
- `tests/integration/test_http_file_download.py` - 6 failures
- `tests/integration/test_storage_space_monitoring.py` - 7 failures
- `tests/security/test_file_management_security.py` - 4 failures

### Specific Test Failures:
```
FAILED test_start_recording_success - AssertionError: Response should contain 'result' field
FAILED test_list_recordings_negative - AssertionError: Should return error for unauthenticated request
FAILED test_get_recording_info_success - AssertionError: Response should contain 'result' field
FAILED test_get_snapshot_info_success - AssertionError: Response should contain 'result' field
FAILED test_delete_recording_success - AssertionError: Response should contain 'result' field
FAILED test_delete_snapshot_success - AssertionError: Response should contain 'result' field
```

## Impact Assessment

**Severity**: CRITICAL
- **Client Integration**: All file management functionality broken
- **API Compliance**: Server violates documented API contract
- **Test Coverage**: 32+ test failures due to missing methods
- **User Experience**: Core functionality unavailable

## Required Fix

### Implementation Priority:
1. **HIGH PRIORITY** - File management methods (`get_recording_info`, `get_snapshot_info`, `delete_recording`, `delete_snapshot`)
2. **HIGH PRIORITY** - Recording control methods (`start_recording`, `list_recordings`, `list_snapshots`)
3. **MEDIUM PRIORITY** - Storage management methods (`get_storage_info`, `configure_storage_thresholds`)
4. **MEDIUM PRIORITY** - HTTP download methods (`get_download_url`)

### Implementation Requirements:
1. **Register methods** in WebSocket server method registry
2. **Implement method handlers** with proper authentication
3. **Return documented response formats** exactly as specified
4. **Handle error cases** according to API documentation
5. **Validate parameters** as documented

### Suggested Implementation Pattern:
```python
# In websocket_server/server.py
async def _method_get_recording_info(self, client_id: str, params: dict) -> dict:
    """Handle get_recording_info method."""
    # Validate authentication
    if not self._is_authenticated(client_id):
        return self._error_response(-32001, "Authentication required")
    
    # Validate parameters
    recording_id = params.get("recording_id")
    if not recording_id:
        return self._error_response(-32602, "Missing required parameter: recording_id")
    
    # Implement recording info retrieval
    try:
        recording_info = await self._get_recording_metadata(recording_id)
        return {
            "jsonrpc": "2.0",
            "result": recording_info,
            "id": params.get("id")
        }
    except FileNotFoundError:
        return self._error_response(-32603, "Recording not found")
```

## Files to Investigate

### Server Files:
- `src/websocket_server/server.py` - Method registration and routing
- `src/file_management/` - File management implementation
- `src/recording/` - Recording control implementation
- `src/storage/` - Storage management implementation

### Documentation Files:
- `docs/api/json-rpc-methods.md` - API specification (ground truth)

## Acceptance Criteria

### For Server Team:
- [ ] All documented methods implemented and registered
- [ ] Methods return documented response formats
- [ ] Proper authentication and authorization implemented
- [ ] Error handling matches API documentation
- [ ] All affected tests pass
- [ ] No breaking changes to existing functionality

### For Test Infrastructure:
- [ ] All file management tests pass
- [ ] All recording control tests pass
- [ ] All storage management tests pass
- [ ] All HTTP download tests pass
- [ ] API compliance validation confirms implementation

## Timeline

**Priority**: IMMEDIATE
- **Impact**: Core functionality completely broken
- **Risk**: Client applications cannot function
- **Dependencies**: None - server-only implementation required

## Related Issues

- **Issue 083**: Authentication method `expires_at` field type mismatch
- **Test Infrastructure**: Now properly validates API compliance

## Notes

This issue was discovered by the improved test infrastructure that validates against API documentation as ground truth. The test suite is working correctly - it's the server implementation that is missing critical functionality.

**CRITICAL**: This is not a test infrastructure issue. The tests are correctly failing because the server implementation is missing documented API methods. The server team must implement all documented methods to match the API specification.

**ESTIMATED IMPACT**: 32+ test failures, affecting core file management, recording control, and storage functionality.
