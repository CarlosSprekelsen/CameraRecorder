# Test Migration TODO - Stable Fixtures Implementation

**Date:** December 19, 2024  
**Priority:** URGENT - Complete within 24 hours  
**Status:** IN PROGRESS - Authentication fixed, fixtures ready

## 🎯 MIGRATION GOAL
Migrate ALL integration tests to use `stable-test-fixture.ts` for consistent, reliable testing against REAL MediaMTX server.

## ⚠️ CRITICAL COMPLIANCE RULES - MUST FOLLOW

### 1. **"Real Integration Always" Principle**
- ✅ **MUST**: Use REAL server (no mocks in integration tests)
- ✅ **MUST**: Connect to actual MediaMTX Camera Service
- ❌ **NEVER**: Use mock-server.ts in integration tests
- ❌ **NEVER**: Mock WebSocket or server responses

### 2. **Authentication Requirements**
- ✅ **MUST**: Use `stable-test-fixture.ts` which handles authentication automatically
- ✅ **MUST**: JWT tokens generated dynamically (no hardcoded credentials)
- ✅ **MUST**: Environment variables loaded via `set-test-env.sh`
- ❌ **NEVER**: Hardcode authentication tokens

### 3. **Endpoint Configuration**
- ✅ **MUST**: Use correct ports (8002 for WebSocket, 8003 for Health)
- ✅ **MUST**: Use `TEST_CONFIG` for endpoint URLs
- ❌ **NEVER**: Mix WebSocket methods with health endpoints

### 4. **Test Environment**
- ✅ **MUST**: Run from `client/` directory
- ✅ **MUST**: Source environment before tests: `source .test_env`
- ✅ **MUST**: Use `jest.integration.config.cjs` for integration tests

## 📋 MIGRATION TASKS

### ✅ COMPLETED
1. **Authentication Setup** - Fixed and automated
2. **Stable Fixtures** - Created and working
3. **WebSocket Integration** - Migrated successfully
4. **Polling Fallback** - Migrated successfully

### 🔄 IN PROGRESS - NEEDS COMPLETION

#### 1. **test_endpoint_validation.ts** - HIGH PRIORITY
**Status:** Partially migrated, needs testing
**Action:** Test and verify it works with fixed fixtures

#### 2. **test_server_integration_validation.ts** - HIGH PRIORITY
**Status:** ❌ FAILING - Server availability check broken
**Action:** 
```typescript
// REPLACE current server availability check with:
import { WebSocketTestFixture, HealthTestFixture } from '../fixtures/stable-test-fixture';

beforeAll(async () => {
  wsFixture = new WebSocketTestFixture();
  healthFixture = new HealthTestFixture();
  await wsFixture.initialize();
  await healthFixture.initialize();
});
```

#### 3. **test_network_integration_validation.ts** - HIGH PRIORITY
**Status:** ❌ FAILING - Server availability check broken
**Action:** Same as above - use stable fixtures

#### 4. **test_camera_operations_integration.ts** - MEDIUM PRIORITY
**Status:** ❌ FAILING - React DOM environment issues
**Action:** 
- Remove `renderHook` usage (React Testing Library)
- Use stable fixtures for WebSocket operations
- Test camera operations via WebSocket directly

#### 5. **test_mvp_functionality_validation.ts** - MEDIUM PRIORITY
**Status:** ❌ FAILING - Authentication issues
**Action:** Replace custom authentication with stable fixtures

#### 6. **test_authentication_comprehensive_integration.js** - LOW PRIORITY
**Status:** ❌ FAILING - Custom authentication logic
**Action:** Migrate to use stable fixtures authentication

#### 7. **test_security_features_integration.js** - LOW PRIORITY
**Status:** ❌ FAILING - Empty test suite
**Action:** Either implement tests or remove file

## 🛠️ MIGRATION TEMPLATE

### For Each Test File:
```typescript
/**
 * REQ-XXX: [Requirement being tested]
 * Coverage: INTEGRATION
 * Quality: HIGH
 */

import { WebSocketTestFixture, HealthTestFixture } from '../fixtures/stable-test-fixture';
import { TEST_CONFIG } from '../config/test-config';

describe('Test Description', () => {
  let wsFixture: WebSocketTestFixture;
  let healthFixture: HealthTestFixture;

  beforeAll(async () => {
    wsFixture = new WebSocketTestFixture();
    healthFixture = new HealthTestFixture();
    await wsFixture.initialize();
    await healthFixture.initialize();
  });

  afterAll(async () => {
    wsFixture.cleanup();
    healthFixture.cleanup();
  });

  describe('WebSocket Operations', () => {
    it('should perform operation', async () => {
      const result = await wsFixture.testConnection();
      expect(result).toBe(true);
    });
  });

  describe('Health Operations', () => {
    it('should check health', async () => {
      const result = await healthFixture.testSystemHealth();
      expect(result).toBe(true);
    });
  });
});
```

## 🚨 CRITICAL ISSUES TO FIX

### 1. **Server Availability Checks**
**Problem:** Tests failing with "MediaMTX Camera Service not available"
**Solution:** Use stable fixtures instead of custom availability checks

### 2. **React DOM Environment**
**Problem:** `document is not defined` errors
**Solution:** Remove React Testing Library usage in integration tests

### 3. **Authentication Failures**
**Problem:** "Authentication required: Not authenticated"
**Solution:** Use stable fixtures which handle authentication automatically

### 4. **Process.exit() Calls**
**Problem:** Tests calling process.exit() causing test runner termination
**Solution:** Remove process.exit() calls, use proper Jest error handling

## 📊 SUCCESS METRICS

### Target Results:
- **Integration Tests:** ≥70% pass rate
- **Authentication:** 100% working
- **Endpoint Usage:** 100% correct
- **Real Server Integration:** 100% working

### Current Status:
- **Unit Tests:** 75% pass rate ✅
- **Integration Tests:** 18% pass rate ❌ (TARGET: 70%)
- **Authentication:** Fixed ✅
- **Stable Fixtures:** Working ✅

## 🔧 TESTING COMMANDS

### Run All Integration Tests:
```bash
cd MediaMTX-Camera-Service-Client/client
source .test_env
npm run test:integration
```

### Run Specific Test:
```bash
npm run test:integration -- --testPathPattern="test_name"
```

### Run with Verbose Output:
```bash
npm run test:integration -- --verbose
```

## 📝 COMPLETION CHECKLIST

- [ ] test_endpoint_validation.ts - Test and verify
- [ ] test_server_integration_validation.ts - Migrate to stable fixtures
- [ ] test_network_integration_validation.ts - Migrate to stable fixtures
- [ ] test_camera_operations_integration.ts - Fix React DOM issues
- [ ] test_mvp_functionality_validation.ts - Fix authentication
- [ ] test_authentication_comprehensive_integration.js - Migrate to stable fixtures
- [ ] test_security_features_integration.js - Implement or remove
- [ ] Verify all tests pass against real server
- [ ] Update test quality table with new pass rates
- [ ] Document any remaining issues

## 🎯 NEXT STEPS AFTER MIGRATION

1. **Fix Unit Test Environment Issues** (React DOM problems)
2. **Fix E2E Tests** (Remove process.exit calls)
3. **Fix Performance Tests** (Configuration issues)
4. **Add Missing Edge Cases** (Rate limiting, concurrent operations)
5. **Improve Test Documentation**

---

**Team Instructions:** Follow this TODO exactly. Use stable fixtures for ALL integration tests. Test against REAL server only. Complete within 24 hours.
