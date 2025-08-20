# Detailed Test Fix Instructions

**For Developer**: Step-by-step guide to fix all failing tests  
**Priority**: CRITICAL - Required for PDR execution  
**Estimated Time**: 2-4 hours  

## 🎯 OVERVIEW

This document provides exact commands and code changes needed to fix each failing test. Follow the instructions in order.

## 📋 PREREQUISITES

```bash
# 1. Navigate to client directory
cd MediaMTX-Camera-Service-Client/client

# 2. Set up environment
source .test_env

# 3. Verify environment
echo $CAMERA_SERVICE_JWT_SECRET  # Should show 64-char hex string
echo $TEST_SERVER_URL           # Should show ws://localhost:8002/ws

# 4. Ensure MediaMTX server is running
curl -s http://localhost:8002/api/health || echo "Server not running"
```

## 🔧 FIX 1: Integration Tests (WebSocket Environment Mismatch)

### **Problem**: Integration tests using `ws` library in jsdom environment

### **Solution**: Use integration config for all integration tests

#### **Step 1.1: Test Current Status**
```bash
# Test with integration config (should work)
npx jest --config jest.integration.config.cjs tests/integration/websocket/test_websocket_basic_integration.js --verbose

# Test with default config (should fail)
npx jest --config jest.config.cjs tests/integration/websocket/test_websocket_basic_integration.js --verbose
```

#### **Step 1.2: Fix Authentication Integration Test**
```bash
# Test current status
npx jest --config jest.integration.config.cjs tests/integration/authentication/test_authentication_comprehensive_integration.js --verbose
```

**Expected**: Should pass (already fixed)

#### **Step 1.3: Fix Camera Operations Integration Test**
```bash
# Test current status
npx jest --config jest.integration.config.cjs tests/integration/camera_ops/test_camera_operations_comprehensive_integration.js --verbose
```

**Expected**: Should pass (already fixed)

#### **Step 1.4: Fix Realtime Features Integration Test**
```bash
# Test current status
npx jest --config jest.integration.config.cjs tests/integration/test_realtime_features_integration.js --verbose
```

**If it fails with "no tests found"**, add Jest test functions:

```javascript
// Add at the end of the file
describe('Realtime Features Integration Tests', () => {
  test('should test realtime features', async () => {
    // Your existing test logic here
    await expect(yourTestFunction()).resolves.not.toThrow();
  }, 30000);
});
```

## 🔧 FIX 2: Performance Tests

#### **Step 2.1: Test Notification Timing Performance**
```bash
npx jest --config jest.integration.config.cjs tests/performance/test_notification_timing_performance.js --verbose
```

**Expected**: Should pass (already fixed)

#### **Step 2.2: Fix Realtime Updates Performance Test**
```bash
# Test current status
npx jest --config jest.integration.config.cjs tests/performance/test_realtime_updates_performance.js --verbose
```

**If it fails with "no tests found"**, add Jest test functions:

```javascript
// Add at the end of the file
describe('Realtime Updates Performance Tests', () => {
  test('should test realtime updates performance', async () => {
    // Your existing test logic here
    await expect(yourTestFunction()).resolves.not.toThrow();
  }, 60000);
});
```

## 🔧 FIX 3: Unit Tests (Mocking Issues)

### **Problem**: Unit tests not properly mocking WebSocket services

#### **Step 3.1: Fix Camera Detail Integration Unit Test**

**File**: `tests/unit/components/test_camera_detail_integration.js`

**Issue**: `ReferenceError: wsService is not defined`

**Fix**: Add proper WebSocket service mocking at the top of the file:

```javascript
// Add these mocks at the top of the file
jest.mock('@/services/websocket', () => ({
  createWebSocketService: jest.fn(() => ({
    connect: jest.fn(),
    disconnect: jest.fn(),
    call: jest.fn(),
    onConnect: jest.fn(),
    onDisconnect: jest.fn(),
    onError: jest.fn(),
    onMessage: jest.fn(),
    addEventListener: jest.fn(),
    send: jest.fn()
  }))
}));

// Mock WebSocket global
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
  addEventListener = jest.fn();
  removeEventListener = jest.fn();
}

global.WebSocket = MockWebSocket;
```

**Test the fix**:
```bash
npx jest --config jest.config.cjs tests/unit/components/test_camera_detail_integration.js --verbose
```

#### **Step 3.2: Fix File Store Unit Test**

**File**: `tests/unit/stores/test_file_store.ts`

**Issue**: Mock functions not being called properly

**Fix**: Update the mock setup:

```typescript
// Update the mock at the top of the file
const mockWebSocketService = {
  connect: jest.fn(),
  disconnect: jest.fn(),
  call: jest.fn(),
  onConnect: jest.fn(),
  onDisconnect: jest.fn(),
  onError: jest.fn(),
  onMessage: jest.fn()
};

const mockCreateWebSocketService = jest.fn(() => mockWebSocketService);

jest.mock('@/services/websocket', () => ({
  createWebSocketService: mockCreateWebSocketService
}));

// Export for use in tests
export { mockWebSocketService, mockCreateWebSocketService };
```

**Test the fix**:
```bash
npx jest --config jest.config.cjs tests/unit/stores/test_file_store.ts --verbose
```

## 🔧 FIX 4: E2E Tests

#### **Step 4.1: Test Take Snapshot E2E**
```bash
npx jest --config jest.integration.config.cjs tests/e2e/test_take_snapshot_e2e.js --verbose
```

**Expected**: Should pass (already fixed)

## 🧪 VALIDATION STEPS

### **Step 1: Test Integration Tests**
```bash
# Test all integration tests
npx jest --config jest.integration.config.cjs tests/integration/ --verbose
```

**Expected**: All integration tests should pass

### **Step 2: Test Unit Tests**
```bash
# Test all unit tests
npx jest --config jest.config.cjs tests/unit/ --verbose
```

**Expected**: All unit tests should pass

### **Step 3: Test Performance Tests**
```bash
# Test all performance tests
npx jest --config jest.integration.config.cjs tests/performance/ --verbose
```

**Expected**: All performance tests should pass

### **Step 4: Test E2E Tests**
```bash
# Test all e2e tests
npx jest --config jest.integration.config.cjs tests/e2e/ --verbose
```

**Expected**: All e2E tests should pass

### **Step 5: Test Complete Suite**
```bash
# Test everything
npm run test:all
```

**Expected**: All tests should pass

## 🚨 TROUBLESHOOTING

### **If Integration Tests Still Fail**

**Check 1**: Environment variables
```bash
echo $CAMERA_SERVICE_JWT_SECRET
echo $TEST_SERVER_URL
```

**Check 2**: Server is running
```bash
curl -s http://localhost:8002/api/health
```

**Check 3**: WebSocket connection
```bash
# Test WebSocket manually
node -e "
const WebSocket = require('ws');
const ws = new WebSocket('ws://localhost:8002/ws');
ws.on('open', () => { console.log('✅ Connected'); ws.close(); });
ws.on('error', (e) => console.log('❌ Error:', e.message));
"
```

### **If Unit Tests Still Fail**

**Check 1**: Mock setup
```bash
# Verify mocks are working
npx jest --config jest.config.cjs tests/unit/ --verbose --no-coverage
```

**Check 2**: Import paths
```bash
# Check if @/services/websocket exists
ls -la src/services/websocket*
```

### **If Authentication Fails**

**Check 1**: JWT secret
```bash
echo $CAMERA_SERVICE_JWT_SECRET | wc -c  # Should be 65 (64 chars + newline)
```

**Check 2**: Test JWT generation
```bash
node -e "
const jwt = require('jsonwebtoken');
const secret = process.env.CAMERA_SERVICE_JWT_SECRET;
const token = jwt.sign({user_id:'test',role:'operator'}, secret, {algorithm:'HS256'});
console.log('Token generated:', token ? '✅' : '❌');
"
```

## 📊 SUCCESS METRICS

**Target**: All tests passing
- ✅ Integration tests: 100% pass rate
- ✅ Unit tests: 100% pass rate  
- ✅ Performance tests: 100% pass rate
- ✅ E2E tests: 100% pass rate

**Expected Output**:
```
Test Suites: 31 passed, 31 total
Tests:       56 passed, 56 total
Snapshots:   0 total
Time:        60s
```

## 🎯 NEXT STEPS

After all tests pass:

1. **Commit the fixes**
2. **Update documentation**
3. **Run PDR validation**
4. **Monitor test stability**

---

**Remember**: The key is using the correct Jest config for each test type:
- **Unit tests**: `jest.config.cjs` (jsdom + mocks)
- **Integration/Performance/E2E**: `jest.integration.config.cjs` (Node.js + real WebSocket)
