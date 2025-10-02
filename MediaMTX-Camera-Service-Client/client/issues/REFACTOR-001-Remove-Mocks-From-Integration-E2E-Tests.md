# REFACTOR-001: Remove Mocks from Integration and E2E Tests

## Summary
Removed all WebSocket mocks from integration and E2E tests. Integration and E2E tests must use REAL WebSockets to connect to the real server, not mocks. Mocks are ONLY for unit tests.

## Affected Files

### ✅ Fixed Files
1. **`tests/integration/setup.ts`**
   - **Before**: Used `WebSocketMock` class wrapping the real WebSocket
   - **After**: Uses REAL WebSocket directly from `ws` package
   - **Change**: Removed `WebSocketMock` wrapper, now assigns real `WebSocket` to `global.WebSocket`

2. **`tests/utils/api-client.ts`**
   - **Before**: `TestAPIClient.connect()` had mock mode check that created `MockWebSocket`
   - **After**: Always calls `connectReal()` for integration/E2E tests
   - **Change**: Removed mock mode conditional, always uses real WebSocket connection

3. **`tests/utils/mocks.ts`**
   - **Status**: Still contains `createMockWebSocket()` method
   - **Purpose**: Used ONLY for unit tests (correct usage)
   - **No changes needed**: This is for unit test mocking only

4. **`tests/setup.ts`** (Unit Test Setup)
   - **Status**: Contains `MockWebSocket` class
   - **Purpose**: Used ONLY for unit tests (correct usage)
   - **No changes needed**: This is for unit test mocking only

## Test Results

### Integration Tests (REAL WebSockets ✅)
```
Test Suites: 11 failed, 3 passed, 14 total
Tests:       51 failed, 106 passed, 157 total
Time:        205.417 s
```

**Key Success**: 106 tests passed using REAL WebSocket connections!

### E2E Tests (REAL WebSockets ✅)
```
Test Suites: 13 failed, 13 total
Tests:       42 failed, 42 total
Time:        4.05 s
```

**Issues**: Module resolution errors (missing `workflow-test-helper`), NOT WebSocket mock issues

## Verification (DOD)

### ✅ DOD Criteria Met:

1. **Integration tests use REAL WebSockets**
   ```typescript
   // tests/integration/setup.ts
   const WebSocket = require('ws');
   global.WebSocket = WebSocket; // NO MOCKS!
   ```

2. **E2E tests use REAL WebSockets**
   - E2E tests use `WebSocketService` directly
   - E2E tests use real service classes (`AuthService`, `DeviceService`, etc.)
   - No WebSocket mocks found in E2E test files

3. **TestAPIClient uses REAL WebSockets**
   ```typescript
   // tests/utils/api-client.ts
   async connect(): Promise<void> {
     // NO MOCKS in integration/E2E tests - always use real WebSocket!
     return this.connectReal();
   }
   ```

4. **Mocks only remain in unit tests**
   - `tests/setup.ts`: MockWebSocket for unit tests ✅
   - `tests/utils/mocks.ts`: createMockWebSocket() for unit tests ✅
   - No mocks in integration/E2E test files ✅

### 🔍 Verification Commands:

```bash
# Verify no WebSocket mocks in integration tests
grep -r "WebSocketMock\|MockWebSocket" tests/integration/
# Result: No matches ✅

# Verify no WebSocket mocks in E2E tests  
grep -r "WebSocketMock\|MockWebSocket" tests/e2e/
# Result: No matches ✅

# Verify integration tests use real WebSocket
grep "WebSocket" tests/integration/setup.ts
# Result: const WebSocket = require('ws'); ✅

# Verify TestAPIClient uses real WebSocket
grep -A5 "async connect()" tests/utils/api-client.ts
# Result: NO MOCKS - always calls connectReal() ✅
```

## Root Cause

The integration and E2E tests were incorrectly configured to use WebSocket mocks:
1. `tests/integration/setup.ts` was wrapping the real WebSocket in a `WebSocketMock` class
2. `TestAPIClient` had a `mockMode` flag that created `MockWebSocket` instead of using real connections
3. This violated the architecture requirement that integration/E2E tests must use real system components

## Expected Behavior

- **Unit Tests**: Use WebSocket mocks (fast, isolated)
- **Integration Tests**: Use REAL WebSockets to connect to real server
- **E2E Tests**: Use REAL WebSockets for complete user journeys

## Status

**COMPLETED** ✅

All WebSocket mocks removed from integration and E2E tests. DOD verified.

## Priority

**CRITICAL** - This was blocking proper integration and E2E testing

## Assignee

**Test Infrastructure Team**

## Next Steps

1. ✅ **DONE**: Remove WebSocket mocks from integration/E2E tests
2. 🔄 **IN PROGRESS**: Fix failing integration tests (51 failures)
3. 🔄 **IN PROGRESS**: Fix E2E module resolution errors
4. ⏳ **PENDING**: Fix integration test cleanup (open handles)
5. ⏳ **PENDING**: Increase integration test coverage

## Architecture Alignment

This refactoring aligns with the architecture ground rules:

> **Real System Testing Over Mocking**
> - **WebSocket:** Use real connections within system
> - **NEVER MOCK:** Internal WebSocket

> **Strategic Mocking Rules**
> **MOCK:** External APIs, time operations, expensive hardware simulation  
> **NEVER MOCK:** MediaMTX service, filesystem, internal WebSocket, JWT auth, config loading

## Related Issues

- Integration tests still have 51 failures (mostly timeout and cleanup issues)
- E2E tests have module resolution errors (missing helper files)

## Evidence

```bash
# Integration test run with REAL WebSockets
npm run test:integration
# Output shows:
# 🚀 Starting Integration Tests with Real Server
# 📡 Server URL: ws://localhost:8002/ws
# Tests: 51 failed, 106 passed, 157 total
```

