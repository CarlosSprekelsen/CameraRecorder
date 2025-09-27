# Bug Report #002: discover_external_streams Internal Server Error

## **Severity:** HIGH
## **Priority:** P1
## **Component:** External Stream Discovery
## **Date:** 2025-09-27

## **Description**
The `discover_external_streams` method returns internal server error instead of proper JSON-RPC error response.

## **Steps to Reproduce**
1. Authenticate with admin role token
2. Call `discover_external_streams` method
3. Observe internal server error

## **Expected Behavior**
According to JSON-RPC API documentation, should return structured error response or successful discovery result.

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
- **Reference:** JSON-RPC 2.0 spec requires structured error responses with specific error codes
- **Impact:** Client cannot determine if discovery failed due to configuration, network, or implementation issues

## **Test Evidence**
```bash
# Test command used:
curl -X POST http://localhost:8002/ws -d '{
  "jsonrpc": "2.0",
  "method": "discover_external_streams",
  "params": {},
  "id": 1
}'
```

## **Environment**
- Server Version: 1.0.0
- Role: admin
- Authentication: JWT token
- Method: discover_external_streams

## **Fix Required**
1. Implement proper error handling in external stream discovery
2. Return specific error codes for different failure scenarios
3. Add logging for debugging discovery failures

## **Validation Test**
Create dedicated test: `test_external_stream_discovery_error_handling`
