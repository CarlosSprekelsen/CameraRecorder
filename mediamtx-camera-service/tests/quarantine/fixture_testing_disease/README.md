# Quarantined Tests - Fixture Testing Disease

**Date:** August 20, 2025  
**Reason:** Tests moved to quarantine due to "Fixture Testing Disease" - excessive mocking instead of real system integration  
**Status:** Permanently quarantined - replaced with real integration tests  

## Overview

These tests were quarantined because they violate the testing guide principles by:
- Mocking real system components instead of testing actual integration
- Using excessive @patch decorators for file operations
- Testing mock responses instead of real system behavior
- Creating false confidence while hiding real integration issues

## Quarantined Test Files

### 1. `test_websocket_server/` (Entire Directory)
**Reason:** All WebSocket tests mocked real connections instead of testing actual WebSocket communication
**Issues:**
- Mocked WebSocket connections with Mock() objects
- No real WebSocket communication testing
- No real client disconnection scenarios
- No real-time notification validation

**Replacement:** `tests/integration/test_websocket_real_integration.py`

### 2. `test_health_server_file_downloads.py`
**Reason:** Excessive mocking of file system operations with 15+ @patch decorators
**Issues:**
- Mocked all file system operations (os.path.join, os.path.exists, etc.)
- No real file system testing
- No real error scenarios for file operations
- Tests mock responses instead of actual file behavior

**Replacement:** Real file system testing in integration tests

### 3. `test_critical_requirements_minimal.py`
**Reason:** Extensive mocking of all components instead of real system integration
**Issues:**
- Mocked MediaMTX controller with Mock() and AsyncMock()
- Mocked camera monitor with Mock() objects
- No real system integration testing
- Tests validate mock responses, not real behavior

**Replacement:** Real integration tests in `tests/integration/`

## Testing Guide Violations

### ❌ **NEVER MOCK Rules Violated:**
1. **MediaMTX Service** - Should use systemd-managed service
2. **WebSocket Connections** - Should use real connections
3. **File Operations** - Should use tempfile
4. **JWT Authentication** - Should use real tokens

### ❌ **Fixture Testing Disease Symptoms:**
1. **Mock Responses** - Tests validate mock responses instead of real behavior
2. **Internal Wrapper Testing** - Tests mock internal methods instead of integration
3. **False Confidence** - Tests pass while real integration fails
4. **Production Risk** - Integration failures discovered in production

## Replacement Strategy

### ✅ **New Real Integration Tests:**
1. **WebSocket Real Integration** - `test_websocket_real_integration.py`
   - Real WebSocket connections
   - Real authentication testing
   - Real client disconnection scenarios
   - Real concurrent connection testing

2. **MediaMTX Real Integration** - `test_mediamtx_real_integration.py`
   - Real systemd-managed MediaMTX service
   - Real stream creation and management
   - Real recording operations
   - Real service failure scenarios

3. **System Real Integration** - `test_system_real_integration.py`
   - Real end-to-end workflows
   - Real cross-component data flow
   - Real concurrent operations
   - Real error scenarios

4. **Error Handling Real** - `test_error_handling_real.py`
   - Real network failures
   - Real resource exhaustion
   - Real service failures
   - Real recovery mechanisms

## Migration Status

- ✅ **Quarantine Complete** - All problematic tests moved to quarantine
- ✅ **Foundation Infrastructure** - Real system test infrastructure created
- ✅ **Authentication Infrastructure** - Real JWT authentication testing created
- ✅ **Requirements Coverage Framework** - Coverage mapping and validation created
- ✅ **Edge Case Framework** - Comprehensive edge case testing framework created
- ✅ **First Real Integration Test** - WebSocket real integration test created
- 🔄 **In Progress** - MediaMTX real integration tests
- ⏳ **Pending** - System integration tests
- ⏳ **Pending** - Error handling tests
- ⏳ **Pending** - Edge case tests

## Quality Improvement

### **Before Quarantine:**
- 25% real integration coverage
- 75% mock-based testing
- 0% adequate WebSocket testing
- 15% adequate MediaMTX testing
- 20% adequate error handling

### **After Redesign (Target):**
- 90%+ real integration coverage
- 0% mock-based testing of internal services
- 100% real WebSocket testing
- 100% real MediaMTX testing
- 100% real error handling

## Conclusion

These tests were quarantined to prevent "Fixture Testing Disease" from spreading and to ensure the test suite provides real confidence in system integration. The new real integration tests will catch actual integration failures before they reach production.

**DO NOT RESTORE** these tests - they violate testing guide principles and create false confidence.
