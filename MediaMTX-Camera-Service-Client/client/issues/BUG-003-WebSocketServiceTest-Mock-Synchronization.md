# BUG-003: Mock State Synchronization Issue in WebSocket Service Test

## Summary
Test `websocket_service_simple.test.ts` fails to properly synchronize mock state between `isConnected()` and `getConnectionStatus()` methods, resulting in test failure.

## Type
Defect - Test Implementation

## Priority
Low

## Severity
Minor

## Affected Component
- **Test File**: `tests/unit/services/websocket_service_simple.test.ts`
- **Test**: "should handle connection state changes"
- **Lines**: 38-45

## Environment
- **Version**: Current development branch
- **Test Framework**: Jest
- **Architecture**: Unit tests for IAPIClient interface

## Steps to Reproduce
1. Run unit tests: `npm run test:unit -- --testPathPattern=websocket_service_simple.test.ts`
2. Observe test failure in "should handle connection state changes"
3. Review test output showing mock state mismatch

## Expected Behavior
When mocking `isConnected()` to return `false`, the `getConnectionStatus()` mock should also be updated to reflect the disconnected state:

```typescript
(apiClient.isConnected as jest.Mock).mockReturnValue(false);
(apiClient.getConnectionStatus as jest.Mock).mockReturnValue({
  connected: false,
  ready: false
});

expect(apiClient.isConnected()).toBe(false);
expect(apiClient.getConnectionStatus()).toEqual({
  connected: false,
  ready: false
});
```

## Actual Behavior
Test only mocks `isConnected()` but forgets to update `getConnectionStatus()` mock, causing assertion failure:

```typescript
(apiClient.isConnected as jest.Mock).mockReturnValue(false);
expect(apiClient.isConnected()).toBe(false);
expect(apiClient.getConnectionStatus()).toEqual({
  connected: false,  // FAILS - still returns true
  ready: false       // FAILS - still returns true
});
```

## Root Cause Analysis

### Code Location
File: `tests/unit/services/websocket_service_simple.test.ts`
Lines: 38-45

### Problematic Test Code
```typescript
test('should handle connection state changes', () => {
  // Test initial state
  expect(apiClient.isConnected()).toBe(true);

  // Mock state changes
  (apiClient.isConnected as jest.Mock).mockReturnValue(false);
  expect(apiClient.isConnected()).toBe(false);
  
  // BUG: getConnectionStatus() mock is not updated
  expect(apiClient.getConnectionStatus()).toEqual({
    connected: false,  // FAILS
    ready: false       // FAILS
  });
});
```

### Mock Initialization
File: `tests/unit/services/websocket_service_simple.test.ts`
Lines: 6-10

```typescript
const mockAPIClient = MockDataFactory.createMockAPIClient();

// Initial mock setup
// createMockAPIClient() returns:
{
  call: jest.fn().mockResolvedValue({}),
  batchCall: jest.fn().mockResolvedValue([]),
  isConnected: jest.fn().mockReturnValue(true),
  getConnectionStatus: jest.fn().mockReturnValue({
    connected: true,
    ready: true
  })
}
```

### Why This Occurred
Test author updated `isConnected()` mock but forgot that `getConnectionStatus()` is a separate mock that needs independent updating. The two mocks don't automatically synchronize.

### Impact Assessment
- **Test Validity**: Test doesn't properly validate mock state management
- **Code Coverage**: False negative - test appears to fail but code may be correct
- **Maintenance**: Other developers may be confused by the failure
- **CI/CD**: Test suite appears broken when it's actually a test bug

## Test Evidence

### Test Output
```
FAIL tests/unit/services/websocket_service_simple.test.ts
  ● IAPIClient Interface Tests (WebSocket Abstraction) › REQ-WS-001: IAPIClient connection management › should handle connection state changes
    
    expect(received).toEqual(expected) // deep equality
    
    - Expected  - 2
    + Received  + 2
    
      Object {
    -   "connected": false,
    -   "ready": false,
    +   "connected": true,
    +   "ready": true,
      }
    
      39 |       (apiClient.isConnected as jest.Mock).mockReturnValue(false);
      40 |       expect(apiClient.isConnected()).toBe(false);
    > 41 |       expect(apiClient.getConnectionStatus()).toEqual({
         |                                               ^
      42 |         connected: false,
      43 |         ready: false
      44 |       });
```

### Mock Factory Implementation
File: `tests/utils/mocks.ts`
Lines: Related to `createMockAPIClient()`

```typescript
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

Note: The factory correctly creates independent mocks, but the test doesn't handle them independently.

## Related Documentation
- **Testing Guidelines**: `docs/development/client-testing-guidelines.md`
- **IAPIClient Interface**: `src/services/abstraction/IAPIClient.ts`

## Acceptance Criteria
1. Test properly synchronizes both `isConnected()` and `getConnectionStatus()` mocks
2. Test passes without errors
3. Test correctly validates state change behavior
4. Mock synchronization pattern is clear and maintainable

## Proposed Fix

### Option 1: Synchronize Both Mocks Manually (Recommended)

**File**: `tests/unit/services/websocket_service_simple.test.ts`
**Lines**: 38-45

```typescript
// BEFORE
test('should handle connection state changes', () => {
  expect(apiClient.isConnected()).toBe(true);
  
  (apiClient.isConnected as jest.Mock).mockReturnValue(false);
  expect(apiClient.isConnected()).toBe(false);
  expect(apiClient.getConnectionStatus()).toEqual({
    connected: false,
    ready: false
  });
});

// AFTER
test('should handle connection state changes', () => {
  expect(apiClient.isConnected()).toBe(true);
  
  // Update BOTH mocks when changing connection state
  (apiClient.isConnected as jest.Mock).mockReturnValue(false);
  (apiClient.getConnectionStatus as jest.Mock).mockReturnValue({
    connected: false,
    ready: false
  });
  
  expect(apiClient.isConnected()).toBe(false);
  expect(apiClient.getConnectionStatus()).toEqual({
    connected: false,
    ready: false
  });
});
```

### Option 2: Create Helper Function for State Changes

```typescript
// Add helper function at top of test file
function setMockConnectionState(
  client: jest.Mocked<IAPIClient>,
  connected: boolean,
  ready: boolean = connected
) {
  (client.isConnected as jest.Mock).mockReturnValue(connected);
  (client.getConnectionStatus as jest.Mock).mockReturnValue({
    connected,
    ready
  });
}

// Use in test
test('should handle connection state changes', () => {
  expect(apiClient.isConnected()).toBe(true);
  
  // Use helper to synchronize mocks
  setMockConnectionState(apiClient, false);
  
  expect(apiClient.isConnected()).toBe(false);
  expect(apiClient.getConnectionStatus()).toEqual({
    connected: false,
    ready: false
  });
});
```

### Option 3: Use beforeEach for Each State (Most Explicit)

```typescript
describe('connection state transitions', () => {
  describe('when connected', () => {
    beforeEach(() => {
      (apiClient.isConnected as jest.Mock).mockReturnValue(true);
      (apiClient.getConnectionStatus as jest.Mock).mockReturnValue({
        connected: true,
        ready: true
      });
    });

    test('should report connected state', () => {
      expect(apiClient.isConnected()).toBe(true);
      expect(apiClient.getConnectionStatus().connected).toBe(true);
    });
  });

  describe('when disconnected', () => {
    beforeEach(() => {
      (apiClient.isConnected as jest.Mock).mockReturnValue(false);
      (apiClient.getConnectionStatus as jest.Mock).mockReturnValue({
        connected: false,
        ready: false
      });
    });

    test('should report disconnected state', () => {
      expect(apiClient.isConnected()).toBe(false);
      expect(apiClient.getConnectionStatus().connected).toBe(false);
    });
  });
});
```

## Recommended Solution
**Option 1** is recommended for this specific test because:
- Minimal changes required
- Clear and explicit
- Easy to understand for maintainers
- Doesn't require refactoring test structure

## Verification Steps
1. Apply proposed fix to `tests/unit/services/websocket_service_simple.test.ts`
2. Run test: `npm run test:unit -- --testPathPattern=websocket_service_simple.test.ts`
3. Verify test passes
4. Verify test still validates the intended behavior
5. Check that similar patterns in other tests are consistent

## Additional Considerations

### Similar Issues in Other Tests
Check if the same pattern exists in:
- `tests/unit/services/websocket_service.test.ts`
- Other tests that mock `IAPIClient`

### Best Practices for Mock State Management
Consider adding to testing guidelines:
1. Always synchronize related mocks when changing state
2. Document dependencies between mocked methods
3. Consider using helper functions for complex mock state
4. Use descriptive variable names for mock states

## Not a Code Bug
**Important**: This is a TEST bug, not a CODE bug. The IAPIClient interface and its implementation are correct. The test simply doesn't properly manage mock state.

## Attachments
- Test failure output: See test run from 2025-10-01
- Test file: `tests/unit/services/websocket_service_simple.test.ts`
- Mock factory: `tests/utils/mocks.ts`
- Interface definition: `src/services/abstraction/IAPIClient.ts`

