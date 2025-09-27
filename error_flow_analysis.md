# Error Flow Analysis: discover_external_streams

## Error Flow Diagram

```
WebSocket Client
    ↓ JSON-RPC Request
WebSocketServer.MethodDiscoverExternalStreams()
    ↓ authenticatedMethodWrapper()
    ↓ methodWrapper()
    ↓ translateErrorToJsonRpc()
    ↓ mediaMTXController.DiscoverExternalStreams()
Controller.DiscoverExternalStreams()
    ↓ hasExternalDiscovery() check
    ↓ ExternalDiscoveryDisabledError (if disabled)
    ↓ externalDiscovery.DiscoverExternalStreamsAPI()
ExternalStreamDiscovery.DiscoverExternalStreamsAPI()
    ↓ DiscoverExternalStreams()
    ↓ config == nil check
    ↓ fmt.Errorf("external stream discovery is not configured")
```

## Error Information Flow Analysis

### 1. WebSocket Layer (`internal/websocket/methods.go`)
- **Method**: `MethodDiscoverExternalStreams`
- **Error Handling**: Returns raw error from controller
- **Translation**: `translateErrorToJsonRpc()` converts to JSON-RPC format

### 2. Controller Layer (`internal/mediamtx/controller.go`)
- **Method**: `DiscoverExternalStreams`
- **Error Check**: `hasExternalDiscovery()` 
- **Error Type**: `ExternalDiscoveryDisabledError`
- **Message**: "External stream discovery is disabled in configuration"

### 3. External Discovery Layer (`internal/mediamtx/external_discovery.go`)
- **Method**: `DiscoverExternalStreamsAPI` → `DiscoverExternalStreams`
- **Error Check**: `config == nil`
- **Error Type**: `fmt.Errorf`
- **Message**: "external stream discovery is not configured"

## Critical Issues Found

### Issue 1: Message Mismatch
- **Controller**: "External stream discovery is disabled in configuration"
- **External Discovery**: "external stream discovery is not configured"
- **String Matching**: `translateErrorToJsonRpc` looks for controller message
- **Result**: External discovery error falls through to generic `INTERNAL_ERROR`

### Issue 2: Error Type Inconsistency
- **Controller**: Uses `ExternalDiscoveryDisabledError` (structured)
- **External Discovery**: Uses `fmt.Errorf` (generic)
- **Translation**: Only handles controller error type

### Issue 3: Error Flow Paths
```
Path A (Controller): ExternalDiscoveryDisabledError → translateErrorToJsonRpc → UNSUPPORTED
Path B (External Discovery): fmt.Errorf → translateErrorToJsonRpc → INTERNAL_ERROR
```

## Root Cause Analysis

The string matching in `translateErrorToJsonRpc` is too specific:
```go
if strings.Contains(errMsg, "external stream discovery is disabled in configuration") {
    return NewJsonRpcError(UNSUPPORTED, "feature_disabled", ...)
}
```

This only matches the controller error, not the external discovery error.

## Recommended Fixes

### Fix 1: Broaden String Matching
```go
if strings.Contains(errMsg, "external stream discovery") && 
   (strings.Contains(errMsg, "disabled") || strings.Contains(errMsg, "not configured")) {
    return NewJsonRpcError(UNSUPPORTED, "feature_disabled", ...)
}
```

### Fix 2: Standardize Error Messages
- Use consistent error message across all layers
- Or use error type checking instead of string matching

### Fix 3: Error Type Hierarchy
- Create common error interface for feature-disabled errors
- Use type assertions instead of string matching

## Current Test Status
- **Expected**: `UNSUPPORTED` (-32030) with structured error data
- **Actual**: `INTERNAL_ERROR` (-32603) with generic error
- **Cause**: String matching fails for external discovery error path
