# Bug Report #001: get_system_status Permission Denied for Admin Role

## **Severity:** HIGH
## **Priority:** P1
## **Component:** Authentication/Authorization
## **Date:** 2025-09-27

## **Description**
The `get_system_status` method returns "Permission denied" error for admin role users, which violates the JSON-RPC API specification.

## **Steps to Reproduce**
1. Authenticate with admin role token
2. Call `get_system_status` method
3. Observe permission denied error

## **Expected Behavior**
According to JSON-RPC API documentation, admin role should have access to system status information.

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
- **Violation:** Admin role should have system status access
- **Reference:** API documentation specifies admin permissions include system monitoring
- **Impact:** Client cannot validate system health for admin users

## **Test Evidence**
```bash
# Test command used:
curl -X POST http://localhost:8002/ws -d '{
  "jsonrpc": "2.0",
  "method": "get_system_status",
  "params": {},
  "id": 1
}'
```

## **Environment**
- Server Version: 1.0.0
- Role: admin
- Authentication: JWT token
- Method: get_system_status

## **Fix Required**
Update permission matrix to allow admin role access to `get_system_status` method.

## **Validation Test**
Create dedicated test: `test_admin_system_status_access`
