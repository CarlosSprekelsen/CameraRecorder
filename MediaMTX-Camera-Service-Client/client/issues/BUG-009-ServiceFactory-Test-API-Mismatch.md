# BUG-009: ServiceFactory Test API Mismatch

## Summary
ServiceFactory tests expect deprecated methods `createWebSocketService()` and `getWebSocketService()` but the factory uses `createAPIClient()`. Tests need updating to match actual API.

## Type
Defect - Test Code / API Documentation Mismatch

## Priority
**MEDIUM**

## Severity
Minor (test-only issue, not production code bug)

## Affected Component
- **Test File**: `tests/unit/services/service_factory.test.ts`
- **Implementation**: `src/services/ServiceFactory.ts`
- **Failed Tests**: 15 tests

## Environment
- **Version**: Current development branch
- **Test Failures**: 15 tests in service_factory.test.ts

## Evidence

### Test Failures
```
● ServiceFactory Unit Tests › REQ-FACTORY-002: Service creation and caching
  TypeError: factory.createWebSocketService is not a function
  at Object.<anonymous> (tests/unit/services/service_factory.test.ts:XX)

● ServiceFactory Unit Tests › REQ-FACTORY-003: Service retrieval
  TypeError: factory.getWebSocketService is not a function
  at Object.<anonymous> (tests/unit/services/service_factory.test.ts:XX)
```

### Actual ServiceFactory API  
**File**: `src/services/ServiceFactory.ts` (lines 37-43)
```typescript
// ACTUAL API - Uses "APIClient" not "WebSocketService"
createAPIClient(wsService: any): IAPIClient {
  if (!this.apiClient) {
    this.apiClient = new APIClient(wsService, logger);
    logger.info('API Client created');
  }
  return this.apiClient;
}

getAPIClient(): IAPIClient | null {
  return this.apiClient;
}
```

### Test Expectations
**File**: `tests/unit/services/service_factory.test.ts`
```typescript
// Tests expect WRONG API - references old "WebSocketService" methods
factory.createWebSocketService()  // ❌ Does not exist
factory.getWebSocketService()     // ❌ Does not exist
```

## Root Cause Analysis
During architecture refactoring:
1. WebSocketService was abstracted behind IAPIClient interface
2. ServiceFactory was updated to create APIClient instead of WebSocketService
3. Tests were NOT updated to match the new API
4. All 15 ServiceFactory tests now fail because they reference non-existent methods

## Expected Behavior
Tests should use the actual ServiceFactory API:
```typescript
// CORRECT test code
const apiClient = factory.createAPIClient(mockWsService);
const retrievedClient = factory.getAPIClient();
```

## Actual Behavior
Tests use deprecated API that no longer exists:
```typescript
// WRONG test code - methods don't exist
const wsService = factory.createWebSocketService();
const retrievedService = factory.getWebSocketService();
```

## Impact
- 15 test failures in service_factory.test.ts
- All ServiceFactory functionality is untested
- Tests don't validate actual production code behavior
- False sense of test coverage (tests exist but don't run)

## Recommended Solution
Update service_factory.test.ts to use actual API:

```typescript
// BEFORE (WRONG)
test('Should create WebSocket service on first call', () => {
  const factory = ServiceFactory.getInstance();
  const ws1 = factory.createWebSocketService();
  expect(ws1).toBeDefined();
});

// AFTER (CORRECT)
test('Should create API client on first call', () => {
  const factory = ServiceFactory.getInstance();
  const mockWsService = createMockWebSocketService();
  const apiClient = factory.createAPIClient(mockWsService);
  expect(apiClient).toBeDefined();
  expect(apiClient).toBeInstanceOf(APIClient);
});
```

### Changes Required
Replace all occurrences in test file:
- `createWebSocketService()` → `createAPIClient(wsService)`
- `getWebSocketService()` → `getAPIClient()`
- Update test descriptions to reference "API Client" not "WebSocket Service"
- Update mock expectations to match IAPIClient interface

## Testing Requirements
- All 15 service_factory.test.ts tests should pass
- Tests should validate actual ServiceFactory methods:
  - `getInstance()` - singleton pattern
  - `createAPIClient()` - API client creation
  - `createAuthService()` - service creation with dependencies
  - `createServerService()` - service creation
  - `createDeviceService()` - service creation
  - `createRecordingService()` - service creation
  - `createFileService()` - service creation
  - `createStreamingService()` - service creation
  - `createExternalStreamService()` - service creation
  - `reset()` - factory reset

## Related Code
**Similar patterns exist in:**
- `src/services/auth/AuthService.ts` - Uses IAPIClient
- `src/services/server/ServerService.ts` - Uses IAPIClient
- `src/services/device/DeviceService.ts` - Uses IAPIClient

All services now depend on IAPIClient, not WebSocketService directly.

## Related Issues
- Part of broader architecture refactoring (WebSocket → APIClient abstraction)
- Similar test updates may be needed in other test files

## Classification
**Test Debt** - Tests lagging behind implementation changes. This is technical debt that should be addressed to maintain test coverage quality.

## Effort Estimate
- **Low complexity** - Simple find/replace in test file
- **15 tests affected** - All in single file
- **Estimated time**: 1-2 hours to update and verify

## Priority Justification
**Medium priority** because:
- ✅ Production code works correctly
- ❌ No test coverage for ServiceFactory (tests exist but fail)
- ❌ Could mask future ServiceFactory bugs
- ⚠️ Part of broader test suite health

