# BUG-004: Test Suite Uses Inconsistent Mock Variable Names

## Summary
Test suite contains references to undefined mock variables (`mockWebSocketService`) instead of the correctly defined mock variables (`mockAPIClient`), causing ReferenceError failures in multiple test cases.

## Type
Defect - Test Implementation Error

## Priority
Medium

## Severity
Major

## Affected Component
- **Test File**: `tests/unit/services/server_service.test.ts`
- **Test Cases**: 5 connection validation tests
- **Mock Infrastructure**: Centralized mock utilities

## Environment
- **Version**: Current development branch
- **Test Framework**: Jest with React Testing Library
- **Mock System**: Centralized MockDataFactory

## Steps to Reproduce
1. Run unit tests: `npm run test:unit -- --testPathPattern=server_service.test.ts`
2. Observe ReferenceError failures in connection validation tests
3. Check test file for undefined variable references

## Expected Behavior
All test cases should reference correctly defined mock variables (`mockAPIClient`) and execute without ReferenceError exceptions.

## Actual Behavior
5 test cases fail with `ReferenceError: mockWebSocketService is not defined` because they reference an undefined mock variable instead of the correctly defined `mockAPIClient`.

## Root Cause Analysis

### Code Location
File: `tests/unit/services/server_service.test.ts`

### Affected Test Cases
1. REQ-SERVER-003: Storage information - connection validation test (line 131)
2. REQ-SERVER-004: System metrics collection - connection validation test (line 172)
3. REQ-SERVER-005: Event subscription management - connection validation test (line 292)
4. Ping functionality - connection validation test (line 321)
5. Error handling - WebSocket connection loss test (line 336)

### Code Evidence
```typescript
// Line 131 - Incorrect mock reference
test('should throw error when WebSocket not connected', async () => {
  mockWebSocketService.isConnected = false;  // ❌ Undefined variable
  
  await expect(serverService.getStorageInfo()).rejects.toThrow('WebSocket not connected');
});

// Line 172 - Incorrect mock reference
test('should throw error when WebSocket not connected', async () => {
  mockWebSocketService.isConnected = false;  // ❌ Undefined variable
  
  await expect(serverService.getMetrics()).rejects.toThrow('WebSocket not connected');
});

// Line 292 - Incorrect mock reference
test('should throw error when WebSocket not connected for subscriptions', async () => {
  mockWebSocketService.isConnected = false;  // ❌ Undefined variable
  
  await expect(serverService.subscribeEvents(['test'])).rejects.toThrow('WebSocket not connected');
});
```

### Correct Mock Definition
```typescript
// Lines 25-27 - Correctly defined mocks
const mockAPIClient = MockDataFactory.createMockAPIClient();
const mockLoggerService = MockDataFactory.createMockLoggerService();
```

### Expected Correct Usage
```typescript
// Should use the correctly defined mock
test('should throw error when WebSocket not connected', async () => {
  mockAPIClient.isConnected = jest.fn().mockReturnValue(false);  // ✅ Correct variable
  
  await expect(serverService.getStorageInfo()).rejects.toThrow('WebSocket not connected');
});
```

### Why This Occurred
The test file was likely created by copying from an older test that used `WebSocketService` directly, but the current architecture uses `IAPIClient` abstraction. The mock variable names were not updated to reflect the architectural change.

### Impact Assessment
- **Test Execution**: 5 test cases fail with ReferenceError
- **Test Coverage**: Connection validation scenarios are not properly tested
- **CI/CD**: Test suite cannot complete successfully
- **Code Quality**: Inconsistent mock usage across test files

## Test Evidence

### Failing Tests
File: `tests/unit/services/server_service.test.ts`

Test failure output:
```
FAIL tests/unit/services/server_service.test.ts
  ● ServerService Unit Tests › REQ-SERVER-003: Storage information › should throw error when WebSocket not connected

    ReferenceError: mockWebSocketService is not defined

      130 |
      131 |     test('should throw error when WebSocket not connected', async () => {
    > 132 |       mockWebSocketService.isConnected = false;
         |       ^
        133 |
      134 |       await expect(serverService.getStorageInfo()).rejects.toThrow('WebSocket not connected');

  ● ServerService Unit Tests › REQ-SERVER-004: System metrics collection › should throw error when WebSocket not connected

    ReferenceError: mockWebSocketService is not defined

      170 |
      171 |     test('should throw error when WebSocket not connected', async () => {
    > 172 |       mockWebSocketService.isConnected = false;
         |       ^
        173 |
      174 |       await expect(serverService.getMetrics()).rejects.toThrow('WebSocket not connected');

  ● ServerService Unit Tests › REQ-SERVER-005: Event subscription management › should throw error when WebSocket not connected for subscriptions

    ReferenceError: mockWebSocketService is not defined

      290 |
      291 |     test('should throw error when WebSocket not connected for subscriptions', async () => {
    > 292 |       mockWebSocketService.isConnected = false;
         |       ^
        293 |
      294 |       await expect(serverService.subscribeEvents(['test'])).rejects.toThrow('WebSocket not connected');
```

### Test Statistics
- **Total Tests**: 28
- **Passing Tests**: 21
- **Failing Tests**: 7 (5 due to ReferenceError, 2 due to missing connection validation)
- **ReferenceError Failures**: 5

## Comparison with Correct Implementation

### Other Test Files (Correct Pattern)
File: `tests/unit/services/auth_service.test.ts`
```typescript
// Correct mock usage
const mockAPIClient = MockDataFactory.createMockAPIClient();

test('should throw error when WebSocket not connected', async () => {
  mockAPIClient.isConnected = jest.fn().mockReturnValue(false);  // ✅ Correct
  
  await expect(authService.authenticate('token')).rejects.toThrow('WebSocket not connected');
});
```

### MockDataFactory Implementation
File: `tests/utils/mocks.ts`
```typescript
// Lines 813-822 - Correct mock creation
static createMockAPIClient(): jest.Mocked<IAPIClient> {
  return {
    call: jest.fn().mockResolvedValue({}),
    batchCall: jest.fn().mockResolvedValue([]),
    isConnected: jest.fn().mockReturnValue(true),
    getConnectionStatus: jest.fn().mockReturnValue({
      connected: true,
      ready: true
    })
  } as jest.Mocked<IAPIClient>;
}
```

## Related Documentation
- **Testing Guidelines**: `docs/development/client-testing-guidelines.md`
- **Mock Utilities**: `tests/utils/mocks.ts`
- **Architecture**: `docs/architecture/client-architechture.md`

## Acceptance Criteria
1. All test cases reference correctly defined mock variables
2. All 5 ReferenceError failures are resolved
3. Mock usage is consistent with centralized MockDataFactory pattern
4. Test execution completes without ReferenceError exceptions

## Proposed Fix

### Changes Required
Replace all `mockWebSocketService` references with `mockAPIClient` in the following test cases:

**Test Case 1: Storage information connection validation**
```typescript
// Line 131: Change from
mockWebSocketService.isConnected = false;
// To
mockAPIClient.isConnected = jest.fn().mockReturnValue(false);
```

**Test Case 2: System metrics connection validation**
```typescript
// Line 172: Change from
mockWebSocketService.isConnected = false;
// To
mockAPIClient.isConnected = jest.fn().mockReturnValue(false);
```

**Test Case 3: Event subscription connection validation**
```typescript
// Line 292: Change from
mockWebSocketService.isConnected = false;
// To
mockAPIClient.isConnected = jest.fn().mockReturnValue(false);
```

**Test Case 4: Ping functionality connection validation**
```typescript
// Line 321: Change from
mockWebSocketService.isConnected = false;
// To
mockAPIClient.isConnected = jest.fn().mockReturnValue(false);
```

**Test Case 5: Error handling connection loss**
```typescript
// Line 336: Change from
mockWebSocketService.isConnected = false;
// To
mockAPIClient.isConnected = jest.fn().mockReturnValue(false);
```

## Verification Steps
1. Apply proposed changes to `tests/unit/services/server_service.test.ts`
2. Run unit tests: `npm run test:unit -- --testPathPattern=server_service.test.ts`
3. Verify all 5 ReferenceError failures are resolved
4. Verify test execution completes successfully
5. Verify mock usage is consistent with other test files

## Additional Notes
- This fix aligns test implementation with the current architectural pattern using `IAPIClient`
- Mock usage should be consistent across all test files using the centralized MockDataFactory
- The fix maintains the intended test behavior while using the correct mock infrastructure

## Attachments
- Test failure output: See test run from 2025-10-01 02:21:57
- Test file: `tests/unit/services/server_service.test.ts`
- Mock utilities: `tests/utils/mocks.ts`
- Reference test: `tests/unit/services/auth_service.test.ts`

