# Quick Fix Summary - WebSocket & Authentication Issues

**🚨 CRITICAL**: Multiple test failures preventing PDR execution  
**⏱️ Estimated Time**: 2-4 hours  
**🎯 Goal**: All tests passing  

## 🔥 IMMEDIATE ACTIONS REQUIRED

### **1. Environment Setup**
```bash
cd MediaMTX-Camera-Service-Client/client
source .test_env
```

### **2. Main Issue: Wrong Jest Config for Integration Tests**

**Problem**: Integration tests using `ws` library in jsdom environment  
**Solution**: Use `jest.integration.config.cjs` for integration tests

```bash
# ❌ WRONG (causes "ws does not work in browser" error)
npm test

# ✅ CORRECT (uses Node.js environment for integration tests)
npx jest --config jest.integration.config.cjs tests/integration/ --verbose
```

### **3. Fix Unit Test Mocking**

**Files to fix**:
- `tests/unit/components/test_camera_detail_integration.js`
- `tests/unit/stores/test_file_store.ts`

**Add proper WebSocket mocking**:
```javascript
jest.mock('@/services/websocket', () => ({
  createWebSocketService: jest.fn(() => ({
    connect: jest.fn(),
    disconnect: jest.fn(),
    call: jest.fn(),
    onConnect: jest.fn(),
    onDisconnect: jest.fn(),
    onError: jest.fn(),
    onMessage: jest.fn()
  }))
}));
```

## 📋 STEP-BY-STEP FIX COMMANDS

### **Step 1: Test Integration Tests (Should Work)**
```bash
npx jest --config jest.integration.config.cjs tests/integration/websocket/test_websocket_basic_integration.js --verbose
npx jest --config jest.integration.config.cjs tests/integration/authentication/test_authentication_comprehensive_integration.js --verbose
npx jest --config jest.integration.config.cjs tests/integration/camera_ops/test_camera_operations_comprehensive_integration.js --verbose
```

### **Step 2: Fix Unit Tests**
```bash
# Fix camera detail test
npx jest --config jest.config.cjs tests/unit/components/test_camera_detail_integration.js --verbose

# Fix file store test  
npx jest --config jest.config.cjs tests/unit/stores/test_file_store.ts --verbose
```

### **Step 3: Test Complete Suite**
```bash
# Test all integration tests
npx jest --config jest.integration.config.cjs tests/integration/ --verbose

# Test all unit tests
npx jest --config jest.config.cjs tests/unit/ --verbose

# Test all performance tests
npx jest --config jest.integration.config.cjs tests/performance/ --verbose

# Test all e2e tests
npx jest --config jest.integration.config.cjs tests/e2e/ --verbose
```

## 🎯 EXPECTED RESULTS

**Before Fix**:
```
Test Suites: 7 failed, 7 of 31 total
Tests:       41 failed, 15 passed, 56 total
```

**After Fix**:
```
Test Suites: 31 passed, 31 total
Tests:       56 passed, 56 total
```

## 🚨 CRITICAL FILES TO FIX

1. **`tests/unit/components/test_camera_detail_integration.js`**
   - Add WebSocket service mocking
   - Fix `wsService is not defined` error

2. **`tests/unit/stores/test_file_store.ts`**
   - Fix mock function setup
   - Ensure proper WebSocket service mocking

3. **Any integration test using `new WebSocket()`**
   - Ensure it uses `jest.integration.config.cjs`
   - Use `require('ws')` instead of browser WebSocket

## 📖 DETAILED INSTRUCTIONS

For complete step-by-step instructions, see:
- `WEBSOCKET_AUTHENTICATION_FIX_GUIDE.md` - Comprehensive guide
- `DETAILED_TEST_FIX_INSTRUCTIONS.md` - Exact commands and code changes

## 🎯 SUCCESS CRITERIA

- ✅ All integration tests pass with `jest.integration.config.cjs`
- ✅ All unit tests pass with `jest.config.cjs`
- ✅ No "ws does not work in browser" errors
- ✅ No "wsService is not defined" errors
- ✅ Authentication working correctly
- ✅ Real server connections working (for integration tests)

---

**Key Principle**: 
- **Unit Tests** = `jest.config.cjs` (jsdom + mocks)
- **Integration Tests** = `jest.integration.config.cjs` (Node.js + real WebSocket)
