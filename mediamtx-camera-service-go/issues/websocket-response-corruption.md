# Critical Bug: WebSocket Server Response Corruption

## Issue Summary
The WebSocket server is corrupting JSON-RPC responses by overwriting the response object returned by method handlers.

## Severity
**CRITICAL** - This breaks the core WebSocket JSON-RPC functionality.

## Description
When a JSON-RPC method handler (e.g., `MethodPing`) returns a response with a `Result` field, the `handleRequest` method in `internal/websocket/server.go` overwrites the response object, potentially corrupting the `Result` field.

## Location
`internal/websocket/server.go` lines 610-611:
```go
// Set JSON-RPC version and ID
response.JSONRPC = "2.0"
response.ID = request.ID
```

## Root Cause
The server is modifying the response object returned by method handlers instead of creating a new response or properly preserving the existing fields.

## Impact
- JSON-RPC responses have `nil` Result fields even when methods return valid results
- Breaks all WebSocket API functionality
- Affects ping, authentication, and all other JSON-RPC methods

## Steps to Reproduce
1. Start WebSocket server
2. Send ping request via WebSocket
3. Expected: Response with `Result: "pong"`
4. Actual: Response with `Result: null`

## Expected Behavior
The server should preserve the `Result` field from method handlers while setting the JSON-RPC version and ID.

## Proposed Fix
The `handleRequest` method should either:
1. Create a new response object and copy fields from the handler response, or
2. Only set fields that are not already set by the handler

## Test Evidence
Unit tests show that ping responses have `nil` Result fields:
```
Expected value not to be nil.
Messages: Ping response should have result
```

## Priority
**HIGH** - This is blocking WebSocket functionality and needs immediate attention from the implementation team.
