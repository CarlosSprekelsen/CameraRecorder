# Bug Report #003: set_discovery_interval Internal Server Error

## **Severity:** HIGH
## **Priority:** P1
## **Component:** External Stream Discovery Configuration
## **Date:** 2025-09-27

## **Description**
The `set_discovery_interval` method returns internal server error instead of proper JSON-RPC error response.

## **Steps to Reproduce**
1. Authenticate with admin role token
2. Call `set_discovery_interval` method with valid parameters
3. Observe internal server error

## **Expected Behavior**
According to JSON-RPC API documentation, should return success confirmation or specific validation error.

## **Actual Behavior**
```
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32603,
    "message": "Internal server error"
  }
}
```

## **JSON-RPC Specification Compliance**
- **Violation:** Generic internal server error instead of specific error handling
- **Reference:** JSON-RPC 2.0 spec requires structured error responses
- **Impact:** Client cannot configure discovery intervals

## **Test Evidence**
```bash
# Test command used:
curl -X POST http://localhost:8002/ws -d '{
  "jsonrpc": "2.0",
  "method": "set_discovery_interval",
  "params": {"scan_interval": 60},
  "id": 1
}'
```

## **Environment**
- Server Version: 1.0.0
- Role: admin
- Authentication: JWT token
- Method: set_discovery_interval
- Parameters: scan_interval: 60

## **Fix Required**
1. Implement proper error handling in discovery interval configuration
2. Add parameter validation for scan_interval
3. Return specific error codes for invalid intervals

## **Validation Test**
Create dedicated test: `test_discovery_interval_configuration_error_handling`
