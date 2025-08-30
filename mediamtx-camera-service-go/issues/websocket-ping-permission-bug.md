# Bug: WebSocket Ping Method Blocked by Permission Checker

## Issue Summary
The WebSocket `ping` method is being blocked by the permission checker for unauthenticated clients, preventing basic connectivity testing.

## Severity
**HIGH** - This breaks basic WebSocket connectivity and prevents unauthenticated clients from testing server availability.

## Description
When a WebSocket client connects and sends a ping request, the server's `checkMethodPermissions` function blocks the request because unauthenticated clients don't have a role assigned. However, ping is a basic connectivity test that should be accessible without authentication.

## Location
`internal/websocket/server.go` lines 81-95 in `checkMethodPermissions` function:

```go
// checkMethodPermissions checks if a client has permission to access a specific method
func (s *WebSocketServer) checkMethodPermissions(client *ClientConnection, methodName string) error {
	// Skip permission check for authentication method
	if methodName == "authenticate" {
		return nil
	}
	// ... rest of function requires valid role
}
```

## Root Cause
The `checkMethodPermissions` function only exempts the "authenticate" method from permission checks, but "ping" should also be exempt since it's a basic connectivity test that doesn't require authentication.

## Impact
- WebSocket ping requests fail for unauthenticated clients
- Basic connectivity testing is broken
- Unit tests fail because ping responses have `nil` Result fields
- Test coverage is reduced due to failing tests

## Steps to Reproduce
1. Start WebSocket server
2. Connect as unauthenticated client
3. Send ping request: `{"jsonrpc": "2.0", "method": "ping", "id": 1, "params": {}}`
4. Expected: Response with `{"jsonrpc": "2.0", "result": "pong", "id": 1}`
5. Actual: Error response due to permission check failure

## Expected Behavior
The ping method should be accessible without authentication, similar to how the authenticate method is exempt from permission checks.

## Proposed Fix
Add "ping" to the list of methods that don't require permission checks:

```go
func (s *WebSocketServer) checkMethodPermissions(client *ClientConnection, methodName string) error {
	// Skip permission check for authentication and basic connectivity methods
	if methodName == "authenticate" || methodName == "ping" {
		return nil
	}
	// ... rest of function
}
```

## Test Evidence
Unit tests show that ping requests from unauthenticated clients fail:
```
Expected value not to be nil.
Messages: Ping response should have result
```

## Priority
**HIGH** - This is blocking WebSocket functionality and test coverage goals.
