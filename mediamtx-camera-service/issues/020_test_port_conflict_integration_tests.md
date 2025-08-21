# Issue 020: Test Port Conflict in Integration Tests

**Priority:** HIGH
**Category:** Test Infrastructure
**Status:** OPEN
**Created:** August 21, 2025
**Discovered By:** Test Suite Execution

## Description
Integration tests are failing because they try to start WebSocket servers on ports that are already in use by the running camera service. This prevents proper test execution and validation.

## Error Details
**Error:** `OSError: [Errno 98] error while attempting to bind on address ('0.0.0.0', 8002): address already in use`
**Location:** `tests/integration/test_critical_interfaces.py:145`
**Test:** `test_get_camera_list_success`
**Root Cause:** Tests trying to start WebSocket server on port 8002 while camera service is already running

## Ground Truth Analysis
### API Documentation Evidence
**`docs/api/json-rpc-methods.md`** defines the public API as JSON-RPC 2.0 over WebSocket:
- Tests should validate against the **running service**, not start their own servers
- Integration tests should use the deployed service for realistic validation

### Architecture Evidence
**`docs/architecture/overview.md`** shows the system architecture:
- Public interface: WebSocket JSON-RPC Server
- Tests should validate the **public API contract**, not internal implementation

### Requirements Evidence
**`docs/requirements/*.md`** contains no references to test servers:
- Requirements focus on **production service behavior**
- Tests should validate **real system behavior**

## Current Test Code (INCORRECT)
```python
# Tests are starting their own servers instead of using the running service
await self.server.start()  # WRONG: tries to bind to port 8002
```

## Correct Test Approach (Based on Ground Truth)
```python
# Tests should use the running service
websocket_url = "ws://localhost:8002/ws"  # Use running service
websocket_client = WebSocketRPCClient(websocket_url)
await websocket_client.authenticate()
result = await websocket_client.call("get_camera_list")
```

## Impact
- **Test Reliability:** Tests fail due to port conflicts
- **Integration Testing:** Missing validation of the complete client-server communication
- **API Compliance:** Tests don't validate the actual public API contract
- **Maintenance:** Tests break when service is running

## Affected Test Files
- `tests/integration/test_critical_interfaces.py` - Port 8002 conflict
- `tests/integration/test_protected_methods_authentication.py` - Port 8003 conflict
- Other integration tests may have similar issues

## Root Cause
The tests were designed to start their own WebSocket servers instead of testing against the running service. This violates the principle that integration tests should validate the public API contract against the actual deployed service.

## Proposed Solution
1. **Modify tests to use running service** instead of starting their own servers
2. **Use WebSocket client** to connect to existing service
3. **Test authentication flow** using documented auth methods
4. **Validate responses** against documented API specifications
5. **Remove server startup code** from integration tests

## Acceptance Criteria
- [ ] Tests use only the running camera service
- [ ] Tests authenticate using documented auth flow
- [ ] Tests validate responses against API documentation
- [ ] No port conflicts during test execution
- [ ] Tests cover the complete client-server communication flow

## Implementation Notes
- Integration tests should validate the **public API contract**
- Tests should use the **running service** for realistic validation
- Authentication should use the documented JWT token flow
- Response validation should match the documented API specifications

## Ground Truth Compliance
- ✅ **API Documentation**: Tests will use documented JSON-RPC methods
- ✅ **Architecture**: Tests will validate the public WebSocket interface
- ✅ **Requirements**: Tests will validate documented functionality

## Testing
- Verify tests connect to running service only
- Confirm authentication follows documented flow
- Validate responses match API documentation
- Ensure no port conflicts occur
