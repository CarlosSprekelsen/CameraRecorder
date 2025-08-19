# WebSocket & Authentication Test Fix Guide

**Status**: CRITICAL - Multiple test failures due to WebSocket environment mismatches and authentication issues  
**Priority**: HIGH - Required for PDR execution  
**Created**: $(date)  
**Scope**: All integration, performance, and e2e tests  

## 🚨 CRITICAL ISSUES IDENTIFIED

### **Issue #1: WebSocket Environment Mismatch**
**Error**: `ws does not work in the browser. Browser clients must use the native WebSocket object`

**Root Cause**: Integration tests are using Node.js `ws` library in Jest's `jsdom` environment (browser simulation)

**Affected Tests**:
- `tests/integration/websocket/test_websocket_basic_integration.js`
- `tests/integration/authentication/test_authentication_comprehensive_integration.js`
- `tests/integration/camera_ops/test_camera_operations_comprehensive_integration.js`
- `tests/integration/test_realtime_features_integration.js`
- `tests/performance/test_realtime_updates_performance.js`

### **Issue #2: Authentication Token Generation**
**Error**: Various authentication failures due to incorrect JWT token generation

**Root Cause**: Tests using hardcoded secrets, custom JWT generation, or wrong environment variables

**Affected Tests**:
- `tests/integration/test_authentication_setup_integration.js`
- `tests/performance/test_notification_timing_performance.js`
- `tests/e2e/test_take_snapshot_e2e.js`

### **Issue #3: Unit Test Mocking Issues**
**Error**: `ReferenceError: wsService is not defined` and `TypeError: Cannot read properties of undefined`

**Root Cause**: Unit tests not properly mocking WebSocket services

**Affected Tests**:
- `tests/unit/components/test_camera_detail_integration.js`
- `tests/unit/stores/test_file_store.ts`

## 🛠️ SOLUTION ARCHITECTURE

### **Dual Jest Configuration Strategy**
```
jest.config.cjs          → Unit tests (jsdom + mocks)
jest.integration.config.cjs → Integration tests (Node.js + real WebSocket)
```

### **Test Environment Separation**
- **Unit Tests**: Fast, isolated, mocked WebSocket (jsdom environment)
- **Integration Tests**: Real server connections (Node.js environment)
- **Performance Tests**: Real server connections (Node.js environment)
- **E2E Tests**: Real server connections (Node.js environment)

## 📋 STEP-BY-STEP FIX INSTRUCTIONS

### **Step 1: Environment Setup**
```bash
cd MediaMTX-Camera-Service-Client/client
source .test_env
```

**Verify Environment Variables**:
```bash
echo $CAMERA_SERVICE_JWT_SECRET  # Should show 64-character hex string
echo $TEST_SERVER_URL           # Should show ws://localhost:8002/ws
```

### **Step 2: Fix Integration Tests (Use Integration Config)**

**For each integration test file, ensure it uses the integration config**:

```bash
# Test with integration config
npx jest --config jest.integration.config.cjs [test-file] --verbose
```

**Required Changes for Integration Tests**:

1. **Add Node.js ws import**:
```javascript
const WebSocket = require('ws');
```

2. **Use Node.js event listeners**:
```javascript
// ✅ CORRECT (Node.js ws library)
ws.on('open', () => {});
ws.on('message', (data) => {});
ws.on('error', (error) => {});
ws.on('close', () => {});

// ❌ WRONG (Browser WebSocket API)
ws.onopen = () => {};
ws.onmessage = (event) => {};
```

3. **Parse message data correctly**:
```javascript
// ✅ CORRECT (Node.js ws library)
const response = JSON.parse(data.toString());

// ❌ WRONG (Browser WebSocket API)
const response = JSON.parse(event.data);
```

4. **Use proper JWT generation**:
```javascript
const jwt = require('jsonwebtoken');

function generateValidToken() {
  if (!process.env.CAMERA_SERVICE_JWT_SECRET) {
    throw new Error('CAMERA_SERVICE_JWT_SECRET environment variable not set');
  }
  
  const payload = {
    user_id: 'test-user',
    role: 'operator',
    iat: Math.floor(Date.now() / 1000),
    exp: Math.floor(Date.now() / 1000) + (24 * 60 * 60)
  };
  
  return jwt.sign(payload, process.env.CAMERA_SERVICE_JWT_SECRET, { algorithm: 'HS256' });
}
```

### **Step 3: Fix Unit Tests (Use Default Config)**

**For each unit test file, ensure it uses the default config**:

```bash
# Test with default config
npx jest --config jest.config.cjs [test-file] --verbose
```

**Required Changes for Unit Tests**:

1. **Proper WebSocket mocking**:
```javascript
// Mock WebSocket service
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

2. **Mock WebSocket global**:
```javascript
// In tests/setup.ts
class MockWebSocket {
  static CONNECTING = 0;
  static OPEN = 1;
  static CLOSING = 2;
  static CLOSED = 3;
  
  readyState = MockWebSocket.CONNECTING;
  onopen = null;
  onclose = null;
  onmessage = null;
  onerror = null;
  
  send = jest.fn();
  close = jest.fn();
}

global.WebSocket = MockWebSocket;
```

### **Step 4: Test Execution Commands**

**Unit Tests (Fast, Mocked)**:
```bash
npm run test:unit
# or
npx jest --config jest.config.cjs --testPathPattern='tests/unit'
```

**Integration Tests (Real Server)**:
```bash
npm run test:integration
# or
npx jest --config jest.integration.config.cjs --testPathPattern='tests/integration'
```

**Performance Tests (Real Server)**:
```bash
npm run test:performance
# or
npx jest --config jest.integration.config.cjs --testPathPattern='tests/performance'
```

**E2E Tests (Real Server)**:
```bash
npm run test:e2e
# or
npx jest --config jest.integration.config.cjs --testPathPattern='tests/e2e'
```

**All Tests**:
```bash
npm run test:all
```

## 🔧 SPECIFIC FIXES BY TEST FILE

### **Integration Tests to Fix**

1. **`tests/integration/websocket/test_websocket_basic_integration.js`**
   - ✅ Already fixed - uses `require('ws')` and Node.js event listeners

2. **`tests/integration/authentication/test_authentication_comprehensive_integration.js`**
   - ✅ Already fixed - uses `require('ws')` and proper JWT generation

3. **`tests/integration/camera_ops/test_camera_operations_comprehensive_integration.js`**
   - ✅ Already fixed - uses `require('ws')` and proper JWT generation

4. **`tests/integration/test_realtime_features_integration.js`**
   - ❌ Needs fix: Add Jest test functions
   - Add `describe()` and `test()` blocks

5. **`tests/integration/test_authentication_setup_integration.js`**
   - ✅ Already fixed - uses environment variables and proper JWT generation

### **Performance Tests to Fix**

1. **`tests/performance/test_notification_timing_performance.js`**
   - ✅ Already fixed - uses `require('ws')` and `jsonwebtoken`

2. **`tests/performance/test_realtime_updates_performance.js`**
   - ❌ Needs fix: Add Jest test functions
   - Add `describe()` and `test()` blocks

### **E2E Tests to Fix**

1. **`tests/e2e/test_take_snapshot_e2e.js`**
   - ✅ Already fixed - uses `require('ws')` and `jsonwebtoken`

### **Unit Tests to Fix**

1. **`tests/unit/components/test_camera_detail_integration.js`**
   - ❌ Needs fix: Proper WebSocket service mocking
   - Fix `wsService is not defined` error

2. **`tests/unit/stores/test_file_store.ts`**
   - ❌ Needs fix: Proper WebSocket service mocking
   - Fix mock function calls

## 🎯 VALIDATION CHECKLIST

### **Before Starting**
- [ ] MediaMTX server running on localhost:8002
- [ ] Environment variables set: `source .test_env`
- [ ] JWT secret available: `echo $CAMERA_SERVICE_JWT_SECRET`

### **After Each Fix**
- [ ] Test passes with correct config
- [ ] No WebSocket environment errors
- [ ] Authentication working correctly
- [ ] Real server connection established (for integration tests)

### **Final Validation**
- [ ] All unit tests pass: `npm run test:unit`
- [ ] All integration tests pass: `npm run test:integration`
- [ ] All performance tests pass: `npm run test:performance`
- [ ] All e2e tests pass: `npm run test:e2e`
- [ ] Complete test suite: `npm run test:all`

## 🚀 QUICK START COMMANDS

```bash
# 1. Setup environment
cd MediaMTX-Camera-Service-Client/client
source .test_env

# 2. Test integration tests (should work)
npx jest --config jest.integration.config.cjs tests/integration/websocket/test_websocket_basic_integration.js --verbose

# 3. Test unit tests (may need fixes)
npx jest --config jest.config.cjs tests/unit/components/test_camera_detail_integration.js --verbose

# 4. Fix issues one by one following the patterns above
```

## 📞 SUPPORT

**If you encounter issues**:
1. Check environment variables are set correctly
2. Verify MediaMTX server is running
3. Ensure you're using the correct Jest config for each test type
4. Follow the WebSocket API patterns (Node.js vs Browser)

**Key Principle**: 
- **Unit Tests** = Fast, Mocked, jsdom environment
- **Integration Tests** = Real Server, Node.js environment

---

**Remember**: The goal is to have ALL tests passing with proper separation between unit and integration testing environments.
