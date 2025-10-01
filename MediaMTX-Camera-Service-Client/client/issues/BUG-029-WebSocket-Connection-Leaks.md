# BUG-029: WebSocket Connection Leaks

## Summary
Jest is detecting open WebSocket connections that are not being properly closed after tests, causing potential memory leaks and preventing Jest from exiting cleanly.

## Detected Issues
```
Jest has detected the following 3 open handles potentially keeping Jest from exiting:
● TCPWRAP
  at Object.connect [as createConnection] (websocket.js:1054:14)
  at TestAPIClient.connectReal (api-client.ts:64:12)
```

## Affected Files
- `tests/integration/api/camera_operations.test.ts`
- `tests/integration/notification_security.test.ts` 
- `tests/integration/server_connectivity.test.ts`

## Root Cause Analysis
1. **Connection Cleanup**: WebSocket connections not properly closed after tests
2. **Test Lifecycle**: Connections created during test setup not cleaned up in teardown
3. **Resource Management**: TestAPIClient not properly disposing of connections

## Expected Behavior
- WebSocket connections should be closed after each test
- Jest should exit cleanly without open handles
- Test cleanup should properly dispose of all resources

## Impact
**MEDIUM** - Causes memory leaks and prevents clean test execution

## Priority
**MEDIUM** - Resource management issue, not blocking functionality

## Assignee
**Test Infrastructure Team**

## Files to Investigate
- `tests/utils/api-client.ts` (TestAPIClient implementation)
- Test setup and teardown procedures
- WebSocket connection lifecycle management
- Jest configuration for cleanup

## Resolution Steps
1. Add proper WebSocket connection cleanup in test teardown
2. Implement `disconnect()` method in TestAPIClient
3. Ensure all tests call cleanup methods after completion
4. Add connection lifecycle management to test utilities
5. Verify Jest exits cleanly after all tests complete
6. Consider adding connection pooling for test efficiency
