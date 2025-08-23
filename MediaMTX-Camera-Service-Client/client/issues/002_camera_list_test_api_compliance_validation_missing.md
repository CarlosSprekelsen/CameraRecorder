# Issue 002: Camera List Test API Compliance Validation Missing

**Status:** OPEN  
**Priority:** High  
**Type:** Test Compliance Issue  
**Created:** 2025-01-23  
**Discovered By:** API Compliance Validation  
**Category:** Test Suite Non-Compliance  

## Description

The camera list test in `tests/integration/test_camera_list_integration.js` lacks proper API compliance validation and does not validate against the frozen API documentation ground truth.

## Ground Truth Reference

**Source:** `mediamtx-camera-service/docs/api/json-rpc-methods.md` (FROZEN)  
**Method:** `get_camera_list`  
**Documented Response Format:** `{ cameras: array, total: number, connected: number }`

## Current Test Implementation

**File:** `tests/integration/test_camera_list_integration.js`  
**Lines:** 22-29  
**Current Issues:**
1. No authentication (required per API documentation)
2. No validation of response format against API documentation
3. No ground truth references in test header
4. No API compliance validation rules

## API Documentation Ground Truth

**Required Authentication:** Yes (viewer role)  
**Documented Response Format:**
```json
{
  "jsonrpc": "2.0",
  "result": {
    "cameras": [
      {
        "device": "/dev/video0",
        "status": "CONNECTED",
        "name": "Camera 0",
        "resolution": "1920x1080",
        "fps": 30,
        "streams": {
          "rtsp": "rtsp://localhost:8554/camera0",
          "webrtc": "http://localhost:8889/camera0/webrtc",
          "hls": "http://localhost:8888/camera0"
        }
      }
    ],
    "total": 1,
    "connected": 1
  },
  "id": 2
}
```

## Impact Assessment

**Severity:** HIGH
- **Test Failure:** Test will fail due to missing authentication
- **API Compliance:** Test does not validate against ground truth
- **Missing Validation:** No verification of response format
- **Ground Truth Violation:** Test does not follow API compliance rules

## Required Changes

### 1. Add Authentication
**Current (Missing):**
```javascript
// No authentication - will fail
const request = {
  jsonrpc: '2.0',
  method: 'get_camera_list',
  id: 1
};
```

**Required (With Authentication):**
```javascript
// First authenticate
await sendRequest(ws, 'authenticate', {
  auth_token: authToken
});

// Then call get_camera_list
const request = {
  jsonrpc: '2.0',
  method: 'get_camera_list',
  id: 1
};
```

### 2. Add API Compliance Validation
- Validate response format matches API documentation
- Check for all required fields: `cameras`, `total`, `connected`
- Validate camera object structure if cameras exist
- Check for required camera fields: `device`, `status`, `name`, `resolution`, `fps`, `streams`

### 3. Update Test Documentation
- Add ground truth references to test header
- Document API compliance validation approach
- Reference frozen API documentation
- Add proper requirements coverage

### 4. Add Error Handling Validation
- Test unauthorized access (should fail with proper error code)
- Validate error response format per API documentation

## Files Affected

### Primary Files:
- `tests/integration/test_camera_list_integration.js` (entire file)

### Related Files:
- `tests/fixtures/stable-test-fixture.ts` (lines 347-348)
- Other camera-related integration tests

## Acceptance Criteria

- [ ] Test includes proper authentication flow
- [ ] Test validates against documented response format
- [ ] Test includes proper ground truth references
- [ ] Test follows API compliance validation rules
- [ ] Test validates error responses for unauthorized access
- [ ] No adaptation to existing implementation flaws

## Testing Rules Compliance

**✅ Ground Truth Validation:** Test must validate against frozen API documentation  
**❌ Current Status:** Test does not validate against API documentation  
**✅ No Code Peeking:** Test should not reference implementation code  
**❌ Current Status:** Test lacks proper validation structure  

## Resolution Priority

**HIGH** - This blocks proper API compliance validation and will cause test failures. The test must be updated to include authentication and proper validation against the frozen API documentation.

## Related Issues

- Issue 001: Authentication Test Parameter Format Non-Compliance
- Other camera operation test compliance issues
