# Bug Report: HTTP Download Endpoints Status Code Bug

**Bug ID:** 049  
**Title:** HTTP Download Endpoints Status Code Bug  
**Severity:** Medium  
**Category:** API/HTTP  
**Status:** Identified  

## Summary

HTTP download endpoints for recordings and snapshots are returning status code 426 (Upgrade Required) instead of the expected status codes (200 for success, 404 for not found). This indicates a protocol or routing issue with the HTTP endpoints.

## Detailed Description

### Root Cause
The HTTP download endpoints are not properly configured or implemented. Status code 426 typically indicates that the server requires an upgrade to a different protocol (e.g., from HTTP to WebSocket), which suggests a routing or configuration issue.

### Impact
- File download functionality is broken
- Client applications cannot access media files
- Integration tests fail with unexpected status codes
- User experience degraded for file access

### Evidence
Test failure showing unexpected status code:
```
FAILED tests/integration/test_critical_interfaces.py::test_http_download_endpoints - AssertionError: Unexpected status: 426
```

The test expects status codes 200 (file exists) or 404 (file not found), but receives 426 (Upgrade Required).

## Recommended Actions

### Option 1: Fix HTTP Download Endpoints (Recommended)
1. **Investigate HTTP server configuration**
   - Check if HTTP server is properly configured
   - Verify routing for `/files/recordings/` and `/files/snapshots/` endpoints
   - Ensure HTTP server is running on correct port

2. **Implement proper file serving**
   - Add HTTP endpoints for file downloads
   - Handle authentication for file access
   - Return appropriate status codes (200, 404, 401, etc.)

3. **Add proper error handling**
   - Handle missing files (404)
   - Handle unauthorized access (401)
   - Handle server errors (500)

### Option 2: Use WebSocket for File Downloads
- Implement file download through WebSocket protocol
- Update client applications to use WebSocket for file access
- Remove HTTP download endpoints

### Option 3: Implement Stub HTTP Endpoints
- Add basic HTTP endpoints that return appropriate status codes
- Implement minimal file serving functionality
- Allow tests to pass with expected responses

## Implementation Priority

**High Priority:**
- Fix HTTP server configuration
- Implement basic file serving functionality
- Return correct status codes

**Medium Priority:**
- Add authentication to file downloads
- Implement proper error handling
- Add file type validation

## Test Validation

After implementation, validate with:
```bash
python3 -m pytest tests/integration/test_critical_interfaces.py::test_http_download_endpoints -v
```

Expected behavior:
- Status 200 for existing files
- Status 404 for non-existent files
- Status 401 for unauthorized access

## Technical Details

### Current Endpoints
- `http://localhost:8002/files/recordings/{filename}`
- `http://localhost:8002/files/snapshots/{filename}`

### Expected Behavior
- Return file content with status 200 for existing files
- Return status 404 for non-existent files
- Return status 401 for unauthorized requests

## Conclusion

This is a **medium-priority HTTP endpoint bug** that affects file download functionality. The HTTP download endpoints need to be properly configured and implemented to return correct status codes and serve files appropriately. This affects the user experience for accessing media files and the reliability of the file management system.
