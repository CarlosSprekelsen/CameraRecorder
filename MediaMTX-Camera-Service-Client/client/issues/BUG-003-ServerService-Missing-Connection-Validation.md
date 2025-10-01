# BUG-003: ServerService Missing Connection Validation Before API Calls

## Summary
ServerService methods do not validate WebSocket connection status before making API calls, causing inconsistent behavior compared to other services and failing connection validation tests.

## Type
Defect - Architectural Inconsistency

## Priority
High

## Severity
Major

## Affected Component
- **Service**: `ServerService`
- **File**: `src/services/server/ServerService.ts`
- **Methods**: All public API methods

## Environment
- **Version**: Current development branch
- **Architecture**: Client-side service layer
- **Test Framework**: Jest with React Testing Library

## Steps to Reproduce
1. Initialize `ServerService` with mock APIClient
2. Set `mockAPIClient.isConnected()` to return `false`
3. Call any ServerService method (e.g., `getServerInfo()`, `getStatus()`)
4. Observe that method succeeds instead of throwing expected connection error

## Expected Behavior
All ServerService methods should validate WebSocket connection before making API calls and throw `'WebSocket not connected'` error when connection is unavailable, consistent with other services like `AuthService`.

## Actual Behavior
ServerService methods proceed with API calls regardless of connection status, returning successful results from mocked API client instead of throwing connection errors.

## Root Cause Analysis

### Code Location
File: `src/services/server/ServerService.ts`

### Affected Methods
1. `getServerInfo()` (line 70-72)
2. `getStatus()` (line 74-76)
3. `getSystemStatus()` (line 78-80)
4. `getStorageInfo()` (line 82-84)
5. `getMetrics()` (line 86-88)
6. `subscribeEvents()` (line 90-92)
7. `unsubscribeEvents()` (line 94-96)
8. `getSubscriptionStats()` (line 98-100)
9. `ping()` (line 102-104)

### Code Evidence
```typescript
// Lines 70-72 - Missing connection validation
async getServerInfo(): Promise<ServerInfo> {
  return this.apiClient.call<ServerInfo>('get_server_info');
}

// Lines 74-76 - Missing connection validation
async getStatus(): Promise<SystemStatus> {
  return this.apiClient.call<SystemStatus>('get_status');
}
```

### Expected Pattern (from AuthService)
```typescript
// AuthService.ts lines 31-33 - Correct pattern
async authenticate(token: string): Promise<AuthenticateResult> {
  if (!this.apiClient.isConnected()) {
    throw new Error('WebSocket not connected');
  }
  // ... proceed with API call
}
```

### Why This Occurred
ServerService was implemented without following the established architectural pattern of connection validation that is consistently used in `AuthService` and other services in the codebase.

### Impact Assessment
- **Architectural Consistency**: Violates established service layer patterns
- **Error Handling**: Inconsistent error behavior across services
- **Testing**: 7 test failures due to missing connection validation
- **User Experience**: Services may appear to work when connection is actually unavailable

## Test Evidence

### Failing Tests
File: `tests/unit/services/server_service.test.ts`

Failing test cases (7 total):
```typescript
// Lines 50-54 - Expected connection validation
test('should throw error when WebSocket not connected', async () => {
  mockAPIClient.isConnected = jest.fn().mockReturnValue(false);
  
  await expect(serverService.getServerInfo()).rejects.toThrow('WebSocket not connected');
});

// Lines 92-96 - Expected connection validation
test('should throw error when WebSocket not connected', async () => {
  mockAPIClient.isConnected = jest.fn().mockReturnValue(false);
  
  await expect(serverService.getStatus()).rejects.toThrow('WebSocket not connected');
});
```

### Test Output
```
FAIL tests/unit/services/server_service.test.ts
  ● ServerService Unit Tests › REQ-SERVER-001: Server information retrieval › should throw error when WebSocket not connected

    expect(received).rejects.toThrow()

    Received promise resolved instead of rejected
    Resolved to value: {"architecture": "amd64", "build_date": "2025-01-15", ...}

  ● ServerService Unit Tests › REQ-SERVER-002: System status monitoring › should throw error when WebSocket not connected

    expect(received).rejects.toThrow()

    Received promise resolved instead of rejected
    Resolved to value: {"components": {"camera_monitor": "RUNNING", ...}}
```

## Comparison with Correct Implementation

### AuthService (Correct Pattern)
File: `src/services/auth/AuthService.ts`
```typescript
// Lines 31-33 - Correct connection validation
async authenticate(token: string): Promise<AuthenticateResult> {
  if (!this.apiClient.isConnected()) {
    throw new Error('WebSocket not connected');
  }
  
  const result = await this.apiClient.call<AuthenticateResult>('authenticate', { auth_token: token });
  return result;
}
```

### DeviceService (Correct Pattern)
File: `src/services/device/DeviceService.ts`
```typescript
// Similar pattern used throughout DeviceService methods
async getCameraList(): Promise<CameraListResult> {
  if (!this.apiClient.isConnected()) {
    throw new Error('WebSocket not connected');
  }
  
  return this.apiClient.call<CameraListResult>('get_camera_list');
}
```

## Related Documentation
- **Testing Guidelines**: `docs/development/client-testing-guidelines.md`
- **Architecture**: `docs/architecture/client-architechture.md`
- **Service Patterns**: ADR-007 (IAPIClient Abstraction)

## Acceptance Criteria
1. All ServerService methods validate `this.apiClient.isConnected()` before making API calls
2. All methods throw `'WebSocket not connected'` error when connection is unavailable
3. All 7 failing connection validation tests pass
4. Behavior matches AuthService and DeviceService patterns

## Proposed Fix

### Changes Required
Add connection validation to all ServerService methods:

**Method: getServerInfo()**
```typescript
// Lines 70-72: Change from
async getServerInfo(): Promise<ServerInfo> {
  return this.apiClient.call<ServerInfo>('get_server_info');
}
// To
async getServerInfo(): Promise<ServerInfo> {
  if (!this.apiClient.isConnected()) {
    throw new Error('WebSocket not connected');
  }
  
  return this.apiClient.call<ServerInfo>('get_server_info');
}
```

**Method: getStatus()**
```typescript
// Lines 74-76: Change from
async getStatus(): Promise<SystemStatus> {
  return this.apiClient.call<SystemStatus>('get_status');
}
// To
async getStatus(): Promise<SystemStatus> {
  if (!this.apiClient.isConnected()) {
    throw new Error('WebSocket not connected');
  }
  
  return this.apiClient.call<SystemStatus>('get_status');
}
```

**Apply same pattern to all remaining methods:**
- `getSystemStatus()`
- `getStorageInfo()`
- `getMetrics()`
- `subscribeEvents()`
- `unsubscribeEvents()`
- `getSubscriptionStats()`
- `ping()`

## Verification Steps
1. Apply proposed changes to `src/services/server/ServerService.ts`
2. Run unit tests: `npm run test:unit -- --testPathPattern=server_service.test.ts`
3. Verify all 7 previously failing connection validation tests now pass
4. Verify no regressions in other service tests
5. Verify consistent behavior with AuthService and DeviceService

## Additional Notes
- This fix ensures architectural consistency across all services
- Connection validation is a fundamental requirement for reliable WebSocket-based communication
- The pattern should be applied consistently to all service methods that make API calls

## Attachments
- Test failure output: See test run from 2025-10-01 02:21:57
- Service implementation: `src/services/server/ServerService.ts`
- Test expectations: `tests/unit/services/server_service.test.ts`
- Reference implementation: `src/services/auth/AuthService.ts`

