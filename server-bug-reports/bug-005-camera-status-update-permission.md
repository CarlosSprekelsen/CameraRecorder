# Bug Report #005: camera_status_update Permission Denied for Admin Role

## **Severity:** MEDIUM
## **Priority:** P2
## **Component:** Authentication/Authorization
## **Date:** 2025-09-27

## **Description**
The `camera_status_update` method returns "Permission denied" error for admin role users, which violates the JSON-RPC API specification.

## **Steps to Reproduce**
1. Authenticate with admin role token
2. Call `camera_status_update` method with valid parameters
3. Observe permission denied error

## **Expected Behavior**
According to JSON-RPC API documentation, admin role should have access to camera status update notifications.

## **Actual Behavior**
```
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32041,
    "message": "Permission denied"
  }
}
```

## **JSON-RPC Specification Compliance**
- **Violation:** Admin role should have camera status update access
- **Reference:** API documentation specifies admin permissions include camera management
- **Impact:** Client cannot receive or send camera status updates

## **Test Evidence**
```bash
# Test command used:
curl -X POST http://localhost:8002/ws -d '{
  "jsonrpc": "2.0",
  "method": "camera_status_update",
  "params": {"device": "camera0", "status": "connected"},
  "id": 1
}'
```

## **Environment**
- Server Version: 1.0.0
- Role: admin
- Authentication: JWT token
- Method: camera_status_update
- Parameters: device: "camera0", status: "connected"

## **Fix Required**
Update permission matrix to allow admin role access to `camera_status_update` method.

## **Validation Test**
Create dedicated test: `test_admin_camera_status_update_access`
