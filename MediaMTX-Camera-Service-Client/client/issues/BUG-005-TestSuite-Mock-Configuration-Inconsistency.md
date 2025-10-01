# BUG-005: Test Suite Mock Configuration Inconsistency

## Summary
Test cases use inconsistent mock configuration patterns for `isConnected()` method, with some tests using direct property assignment (`mockAPIClient.isConnected = false`) and others using Jest function mocking (`mockAPIClient.isConnected = jest.fn().mockReturnValue(false)`), causing test execution inconsistencies.

## Type
Defect - Test Implementation Inconsistency

## Priority
Medium

## Severity
Minor

## Affected Component
- **Test File**: `tests/unit/services/server_service.test.ts`
- **Mock Configuration**: Jest mock setup patterns
- **Test Cases**: Connection validation tests

## Environment
- **Version**: Current development branch
- **Test Framework**: Jest with React Testing Library
- **Mock System**: Jest mocking utilities

## Steps to Reproduce
1. Examine test file `tests/unit/services/server_service.test.ts`
2. Compare mock configuration patterns across different test cases
3. Observe inconsistent usage of mock setup methods

## Expected Behavior
All test cases should use consistent mock configuration patterns, preferably Jest function mocking for better test isolation and control.

## Actual Behavior
Test cases use mixed mock configuration approaches:
- Some tests use: `mockAPIClient.isConnected = jest.fn().mockReturnValue(false)`
- Other tests use: `mockAPIClient.isConnected = false` (direct property assignment)

## Root Cause Analysis

### Code Location
File: `tests/unit/services/server_service.test.ts`

### Inconsistent Mock Patterns

**Pattern 1: Jest Function Mocking (Correct)**
```typescript
// Lines 51, 93 - Correct pattern
test('should throw error when WebSocket not connected', async () => {
  mockAPIClient.isConnected = jest.fn().mockReturnValue(false);  // ✅ Jest function mock
  
  await expect(serverService.getServerInfo()).rejects.toThrow('WebSocket not connected');
});
```

**Pattern 2: Direct Property Assignment (Inconsistent)**
```typescript
// Lines 131, 172, 292, 321, 336 - Inconsistent pattern
test('should throw error when WebSocket not connected', async () => {
  mockAPIClient.isConnected = false;  // ❌ Direct property assignment
  
  await expect(serverService.getStorageInfo()).rejects.toThrow('WebSocket not connected');
});
```

### Mock Definition Context
```typescript
// Lines 25-27 - Mock creation uses Jest function
const mockAPIClient = MockDataFactory.createMockAPIClient();

// MockDataFactory creates isConnected as Jest function:
static createMockAPIClient(): jest.Mocked<IAPIClient> {
  return {
    isConnected: jest.fn().mockReturnValue(true),  // Jest function
    // ... other properties
  } as jest.Mocked<IAPIClient>;
}
```

### Why This Occurred
The test file was likely developed incrementally with different developers or at different times, leading to inconsistent mock configuration patterns. Some tests were written using Jest function mocking while others used direct property assignment.

### Impact Assessment
- **Test Consistency**: Mixed patterns make test maintenance difficult
- **Test Isolation**: Direct property assignment may not properly reset between tests
- **Mock Control**: Jest function mocking provides better control over mock behavior
- **Code Quality**: Inconsistent patterns reduce code maintainability

## Test Evidence

### Mock Configuration Analysis
File: `tests/unit/services/server_service.test.ts`

**Correct Pattern Usage (2 instances):**
```typescript
// Line 51 - REQ-SERVER-001 connection validation
mockAPIClient.isConnected = jest.fn().mockReturnValue(false);

// Line 93 - REQ-SERVER-002 connection validation  
mockAPIClient.isConnected = jest.fn().mockReturnValue(false);
```

**Inconsistent Pattern Usage (5 instances):**
```typescript
// Line 131 - REQ-SERVER-003 connection validation
mockAPIClient.isConnected = false;

// Line 172 - REQ-SERVER-004 connection validation
mockAPIClient.isConnected = false;

// Line 292 - REQ-SERVER-005 connection validation
mockAPIClient.isConnected = false;

// Line 321 - Ping functionality connection validation
mockAPIClient.isConnected = false;

// Line 336 - Error handling connection validation
mockAPIClient.isConnected = false;
```

### Test Execution Context
Both patterns may work in the current test environment, but Jest function mocking is the recommended approach for:
- Better test isolation
- Proper mock reset between tests
- More explicit mock behavior control
- Consistency with Jest best practices

## Comparison with Best Practices

### Jest Documentation (Recommended Pattern)
```typescript
// Jest recommended pattern for function mocking
const mockFunction = jest.fn();
mockFunction.mockReturnValue(false);

// Or inline
const mockFunction = jest.fn().mockReturnValue(false);
```

### Other Test Files (Consistent Pattern)
File: `tests/unit/services/auth_service.test.ts`
```typescript
// Consistent use of Jest function mocking
test('should throw error when WebSocket not connected', async () => {
  mockAPIClient.isConnected = jest.fn().mockReturnValue(false);  // ✅ Consistent pattern
  
  await expect(authService.authenticate('token')).rejects.toThrow('WebSocket not connected');
});
```

### MockDataFactory Pattern
File: `tests/utils/mocks.ts`
```typescript
// MockDataFactory creates Jest functions
static createMockAPIClient(): jest.Mocked<IAPIClient> {
  return {
    isConnected: jest.fn().mockReturnValue(true),  // Jest function
    call: jest.fn().mockResolvedValue({}),         // Jest function
    // ... all methods are Jest functions
  } as jest.Mocked<IAPIClient>;
}
```

## Related Documentation
- **Jest Documentation**: Function mocking best practices
- **Testing Guidelines**: `docs/development/client-testing-guidelines.md`
- **Mock Utilities**: `tests/utils/mocks.ts`

## Acceptance Criteria
1. All test cases use consistent Jest function mocking pattern
2. No direct property assignments on mock functions
3. All mock configurations follow Jest best practices
4. Test patterns are consistent across the test file

## Proposed Fix

### Changes Required
Replace all direct property assignments with Jest function mocking:

**Test Case 1: Storage information connection validation**
```typescript
// Line 131: Change from
mockAPIClient.isConnected = false;
// To
mockAPIClient.isConnected = jest.fn().mockReturnValue(false);
```

**Test Case 2: System metrics connection validation**
```typescript
// Line 172: Change from
mockAPIClient.isConnected = false;
// To
mockAPIClient.isConnected = jest.fn().mockReturnValue(false);
```

**Test Case 3: Event subscription connection validation**
```typescript
// Line 292: Change from
mockAPIClient.isConnected = false;
// To
mockAPIClient.isConnected = jest.fn().mockReturnValue(false);
```

**Test Case 4: Ping functionality connection validation**
```typescript
// Line 321: Change from
mockAPIClient.isConnected = false;
// To
mockAPIClient.isConnected = jest.fn().mockReturnValue(false);
```

**Test Case 5: Error handling connection validation**
```typescript
// Line 336: Change from
mockAPIClient.isConnected = false;
// To
mockAPIClient.isConnected = jest.fn().mockReturnValue(false);
```

### Additional Improvements
Consider adding `beforeEach` setup for consistent mock reset:
```typescript
beforeEach(() => {
  jest.clearAllMocks();
  // Reset mock to default connected state
  mockAPIClient.isConnected = jest.fn().mockReturnValue(true);
});
```

## Verification Steps
1. Apply proposed changes to `tests/unit/services/server_service.test.ts`
2. Run unit tests: `npm run test:unit -- --testPathPattern=server_service.test.ts`
3. Verify all tests pass with consistent mock patterns
4. Verify mock behavior is properly isolated between tests
5. Compare with other test files for pattern consistency

## Additional Notes
- Jest function mocking provides better test isolation and control
- Consistent patterns improve test maintainability and readability
- This fix aligns with Jest best practices and other test files in the codebase
- Mock reset between tests ensures proper test isolation

## Attachments
- Test file: `tests/unit/services/server_service.test.ts`
- Mock utilities: `tests/utils/mocks.ts`
- Reference implementation: `tests/unit/services/auth_service.test.ts`
- Jest documentation: Function mocking patterns

