# Bug Report #005: camera_status_update Permission Denied - RESOLVED

## **Severity:** INFORMATIONAL (Not a Bug)
## **Priority:** N/A
## **Component:** Authentication/Authorization
## **Date:** 2025-09-27
## **Status:** RESOLVED - Correct Security Behavior

## **Description**
The `camera_status_update` method returns "Permission denied" error for admin role users. **This is CORRECT SECURITY BEHAVIOR, not a bug.**

## **Root Cause Analysis**
- **Issue**: `camera_status_update` is a **server-generated notification**, not a client-callable method
- **API Documentation**: Clearly states it's a "NOTIFICATION EVENT" and "Server-to-Client Notification (not callable method)"
- **Security Design**: Method is intentionally blocked to prevent clients from sending fake notifications
- **Permission Matrix**: Method is correctly NOT included in any role permissions

## **Actual Behavior (CORRECT)**
```
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32002,
    "message": "Permission denied"
  }
}
```

## **JSON-RPC Specification Compliance**
- **✅ COMPLIANT**: Method correctly returns permission denied for security
- **Reference**: API documentation clearly states this is a notification-only method
- **Security**: Prevents clients from sending fake camera status updates

## **Why This Was Missed in Testing**
- **Test Coverage Gap**: No integration test existed to validate notification method security
- **Missing Test**: `TestMissingAPI_CameraStatusUpdate_Integration` was not implemented
- **Client Misunderstanding**: Client team attempted to call server-generated notification method

## **Solution Implemented**
1. **Added Test Coverage**: Created `TestMissingAPI_CameraStatusUpdate_Integration` test
2. **Validated Security**: Test confirms method is properly blocked with `-32002` (Permission Denied)
3. **Added Test Client Method**: `CameraStatusUpdate()` method for testing security validation
4. **Documentation**: Clarified that this is correct security behavior, not a bug

## **Test Evidence**
```bash
# Test validates correct security behavior:
go test -v ./internal/websocket -run TestMissingAPI_CameraStatusUpdate_Integration
# Result: PASS - Security properly enforced
```

## **Client Team Action Required**
- **DO NOT** attempt to call `camera_status_update` directly
- **USE** event subscription system (`subscribe_events`) to receive camera status updates
- **LISTEN** for `camera.connected` and `camera.disconnected` events instead
- **UNDERSTAND** that this is a server-generated notification, not a client-callable method

## **Correct Usage**
```javascript
// WRONG - Don't call this directly:
// camera_status_update({"device": "camera0", "status": "connected"})

// CORRECT - Subscribe to events:
subscribe_events({"topics": ["camera.connected", "camera.disconnected"]})
// Then listen for incoming notifications from the server
```

## **Resolution**
- **✅ Bug Report**: Updated to reflect correct understanding
- **✅ Test Coverage**: Added comprehensive security validation test
- **✅ Documentation**: Clarified API usage for client team
- **✅ Security**: Validated notification methods are properly protected
