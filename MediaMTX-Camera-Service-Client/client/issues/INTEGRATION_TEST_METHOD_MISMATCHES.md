# Integration Test Method Mismatches - Critical Issue

**Date:** August 20, 2025  
**Status:** CRITICAL - Blocking integration test validation  
**Priority:** HIGH - Required for PDR execution  
**Scope:** All integration tests failing due to server API misalignment  

---

## Executive Summary

Integration tests are failing due to **server method mismatches** and **API alignment issues**. The tests are calling methods that either don't exist on the MediaMTX server or have different signatures than expected.

### **Critical Issues Identified:**
- **Request Timeouts**: Methods timing out due to non-existent server methods
- **Method Not Found**: Tests calling methods that don't exist on server
- **Error Code Mismatches**: Tests expecting specific error codes not returned by server
- **Connection Issues**: WebSocket connection problems in test environment

---

## Detailed Issue Analysis

### **1. Request Timeout Issues**

#### **Problem:**
Multiple methods are timing out with 5-second timeouts:
- `ping` method
- `get_camera_list` method  
- `list_recordings` method
- `list_snapshots` method

#### **Root Cause:**
These methods either don't exist on the MediaMTX server or have different names/signatures.

#### **Affected Tests:**
```
tests/integration/test_websocket_integration.ts
tests/integration/test_mvp_functionality_validation.ts
tests/integration/test_server_integration_validation.ts
```

#### **Error Pattern:**
```
WebSocketError: Request timeout for method: ping
WebSocketError: Request timeout for method: get_camera_list
WebSocketError: Request timeout for method: list_recordings
```

### **2. Method Not Found Errors**

#### **Problem:**
Tests expecting specific error codes for invalid methods, but receiving `undefined` instead.

#### **Expected vs Actual:**
```javascript
// Expected:
expect(error.code).toBe(ERROR_CODES.METHOD_NOT_FOUND); // -32601

// Actual:
expect(error.code).toBe(undefined); // No error code returned
```

#### **Affected Error Codes:**
- `ERROR_CODES.METHOD_NOT_FOUND` (-32601)
- `ERROR_CODES.INVALID_PARAMS` (-32602)  
- `ERROR_CODES.CAMERA_NOT_FOUND_OR_DISCONNECTED` (-32001)

### **3. Connection Issues**

#### **Problem:**
WebSocket connection tests failing with connection state mismatches.

#### **Error Pattern:**
```javascript
// Expected:
expect(wsService.isConnected).toBe(true);

// Actual:
expect(wsService.isConnected).toBe(false);
```

#### **Affected Tests:**
- Connection resilience tests
- Reconnection tests
- Network interruption tests

### **4. Server API Method Validation**

#### **Methods That May Not Exist:**
1. `ping` - May not be implemented on server
2. `get_camera_list` - May have different name
3. `list_recordings` - May have different name
4. `list_snapshots` - May have different name
5. `get_camera_status` - May have different signature
6. `take_snapshot` - May have different parameters
7. `start_recording` - May have different parameters
8. `stop_recording` - May have different parameters

---

## Impact Assessment

### **Test Suite Impact:**
- **Integration Tests**: 56 failed, 7 passed (11% pass rate)
- **MVP Functionality**: All critical workflows failing
- **Server Integration**: All validation tests failing
- **Performance Tests**: All performance validation failing

### **Business Impact:**
- **PDR Execution**: Blocked - cannot validate real server integration
- **Quality Assurance**: No confidence in server communication
- **Release Readiness**: Cannot verify production functionality

---

## Required Actions

### **Phase 1: Server API Discovery (IMMEDIATE)**

#### **1.1: Document Actual Server Methods**
```bash
# Connect to MediaMTX server and discover available methods
# Use WebSocket client to list all available JSON-RPC methods
```

#### **1.2: Create Server API Documentation**
- Document all available JSON-RPC methods
- Document method signatures and parameters
- Document error codes and responses
- Create API compatibility matrix

#### **1.3: Validate Server Endpoints**
```bash
# Test each method individually to verify:
# - Method exists
# - Parameters are correct
# - Error handling works
# - Response format matches expectations
```

### **Phase 2: Test Alignment (HIGH PRIORITY)**

#### **2.1: Update Test Methods**
- Align test method calls with actual server API
- Update method names and signatures
- Fix parameter structures
- Update error code expectations

#### **2.2: Fix Connection Tests**
- Investigate WebSocket connection issues
- Fix connection state management
- Update reconnection logic
- Fix network interruption tests

#### **2.3: Update Error Handling**
- Align error code expectations with server responses
- Update error message validation
- Fix timeout handling
- Update error recovery tests

### **Phase 3: Validation (MEDIUM PRIORITY)**

#### **3.1: Comprehensive Testing**
- Test all methods against real server
- Validate error scenarios
- Test performance under load
- Validate real-time notifications

#### **3.2: Documentation Update**
- Update test documentation
- Create server API reference
- Document integration test requirements
- Update troubleshooting guides

---

## Immediate Next Steps

### **1. Server Method Discovery**
```bash
# Connect to MediaMTX server and discover available methods
# This is CRITICAL for fixing the integration tests
```

### **2. Create Method Mapping**
```javascript
// Map test methods to actual server methods
const METHOD_MAPPING = {
  'ping': 'actual_ping_method',
  'get_camera_list': 'actual_camera_list_method',
  'list_recordings': 'actual_recordings_method',
  // ... etc
};
```

### **3. Update Test Configuration**
```javascript
// Update test configuration to use correct methods
// and handle server-specific behavior
```

---

## Success Criteria

### **Phase 1 Success:**
- [ ] All server methods documented
- [ ] API compatibility matrix created
- [ ] Method mapping completed

### **Phase 2 Success:**
- [ ] Integration test pass rate >80%
- [ ] All critical workflows working
- [ ] Error handling validated

### **Phase 3 Success:**
- [ ] 100% integration test pass rate
- [ ] Performance targets met
- [ ] Real-time features validated

---

## Risk Assessment

### **High Risk:**
- **Server API Changes**: If server API changes, tests will break again
- **Method Deprecation**: Server methods may be deprecated
- **Version Mismatches**: Client and server version incompatibilities

### **Mitigation:**
- **API Versioning**: Implement proper API versioning
- **Backward Compatibility**: Maintain backward compatibility
- **Automated Validation**: Implement automated API validation

---

## Conclusion

The integration test failures are primarily due to **server API misalignment**. The tests are calling methods that either don't exist or have different signatures than expected. 

**Immediate Action Required:** Discover and document the actual MediaMTX server API methods to align the integration tests with the real server implementation.

**Priority:** HIGH - This is blocking PDR execution and quality validation.

---

**Next Action:** Begin server method discovery to create accurate API documentation and fix integration test alignment.
