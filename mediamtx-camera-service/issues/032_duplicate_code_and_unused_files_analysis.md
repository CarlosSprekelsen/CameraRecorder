# Issue 032: Duplicate Code and Unused Files Analysis

**Date:** 2025-01-27  
**Severity:** Medium  
**Category:** Code Quality  
**Status:** Identified  

## Summary

Analysis of the source code repository has identified several instances of duplicate code and unused files that were likely created by parallel developers working out of sync. This creates maintenance overhead and potential confusion.

## Duplicate Code Identified

### 1. Port Utility Functions (CRITICAL)

**Files:**
- `tests/fixtures/port_utils.py` (44 lines)
- `tests/utils/port_utils.py` (87 lines)

**Duplication:**
- Both files contain `check_websocket_server_port()` function with different signatures:
  - `fixtures/port_utils.py`: `check_websocket_server_port(port: int = 8002, host: str = "127.0.0.1")`
  - `utils/port_utils.py`: `check_websocket_server_port(port: int)`

**Usage:**
- Only `tests/utils/port_utils.py` is imported: `from tests.utils.port_utils import create_test_health_server, find_free_port`
- `tests/fixtures/port_utils.py` is **unused**

**Recommendation:** Delete `tests/fixtures/port_utils.py`

### 2. WebSocket Test Client Classes (REVISED - NOT DUPLICATES)

**Files and Their Specific Purposes:**
- `tests/fixtures/websocket_test_client.py` - `WebSocketTestClient` class
  - **Purpose:** General-purpose WebSocket client for real system testing
  - **Usage:** Connects to real service on port 8002
  - **Features:** Notification listening, message queuing, comprehensive JSON-RPC support

- `tests/fixtures/auth_utils.py` - `WebSocketAuthTestClient` class (line 314)
  - **Purpose:** Authentication-specific testing
  - **Usage:** Used by integration tests that start their own servers
  - **Features:** JWT authentication, protected method calls

- `tests/health/test_health_monitoring.py` - `WebSocketHealthClient` class (line 154)
  - **Purpose:** Health monitoring API testing
  - **Usage:** Connects to real service on port 8002
  - **Features:** Health-specific method calls

- `tests/performance/test_api_performance.py` - `WebSocketPerformanceClient` class (line 182)
  - **Purpose:** Performance testing with timing measurements
  - **Usage:** Connects to real service on port 8002
  - **Features:** Response time measurement, performance metrics

- `tests/integration/test_critical_error_handling.py` - `WebSocketTestClient` class (lines 52+)
  - **Purpose:** Error handling and network failure testing
  - **Usage:** Connects to real service on port 8002
  - **Features:** Error handling, graceful disconnection

**Analysis:**
- These are **purpose-specific implementations** rather than true duplicates
- Each client serves a distinct testing domain (auth, health, performance, error handling)
- The naming similarity (`WebSocketTestClient`) in two files is the only actual duplication

**Recommendation:** 
- **Keep all implementations** as they serve different purposes
- **Rename** the duplicate `WebSocketTestClient` in `test_critical_error_handling.py` to `WebSocketErrorTestClient` for clarity
- **Consider** creating a base `BaseWebSocketTestClient` class for common functionality

### 3. Token Generation Functions (MEDIUM)

**Files:**
- `tests/fixtures/auth_utils.py` - Multiple token generation functions (lines 33, 45, 57, 272, 277, 282)
- `tests/performance/test_api_performance.py` - `_generate_test_token()` (line 151)
- `tests/integration/test_service_manager_requirements.py` - `_generate_valid_jwt_token()` (line 44)
- `MediaMTX-Camera-Service-Client/client/generate-test-token.py` - `generate_test_token()` (line 8)

**Duplication:**
- Multiple implementations of test token generation
- Some functions have identical names but different implementations

**Usage:**
- `tests/fixtures/auth_utils.py` is widely imported and used
- Other implementations are used in their respective test files

**Recommendation:** Consolidate into `tests/fixtures/auth_utils.py` and remove duplicates

## Unused Files Identified

### 1. Configuration Files (LOW)

**Files:**
- `config/camera-service-fixed.yaml` - **No imports found**
- `config/production.yaml` - **No imports found**

**Usage:**
- Only referenced in documentation examples, not in actual code

**Recommendation:** Remove if not needed for deployment documentation

### 2. Scripts (LOW)

**Files:**
- `scripts/validate_multi_tier_snapshot.py` - **No imports found**
- `scripts/strip_bom.py` - **No imports found**

**Usage:**
- These scripts are not imported or referenced by any other code

**Recommendation:** Remove if not needed for manual operations

### 3. Empty Directories (LOW)

**Directories:**
- `dry_run/artifacts/` - **Empty**
- `evidence/sprint-3-actual/snapshots/` - **Empty**
- `evidence/sprint-3-actual/recordings/` - **Empty**

**Recommendation:** Remove empty directories or add `.gitkeep` files

### 4. Test Configuration (LOW)

**File:**
- `test_mediamtx.yml` - Only used in `tests/fixtures/real_system.py`

**Usage:**
- Minimal usage, could be consolidated with other test configurations

## Impact Assessment

### High Impact
- **Port utility duplication** creates confusion about which implementation to use ✅ **RESOLVED**
- **WebSocket client naming conflict** creates import confusion ✅ **RESOLVED**

### Medium Impact
- **Token generation duplication** increases maintenance overhead
- **Unused configuration files** create confusion about deployment options

### Low Impact
- **Unused scripts** and **empty directories** add repository clutter

## Recommendations

### Immediate Actions (High Priority)


### Long-term
1. **Establish** code review guidelines to prevent future duplications
2. **Implement** automated checks for duplicate function/class names
3. **Create** shared test utilities library to prevent reinvention

## Files to Consolidate

```
tests/fixtures/websocket_test_client.py
tests/integration/test_critical_error_handling.py (WebSocketTestClient class)
tests/health/test_health_monitoring.py (WebSocketHealthClient class)
tests/performance/test_api_performance.py (WebSocketPerformanceClient class)
```

## Root Cause

Parallel development without proper coordination led to:
1. **Reinvention** of common utilities (port checking, token generation) ✅ **PARTIALLY RESOLVED**
2. **Naming conflicts** for similar functionality (WebSocket clients)
3. **Lack of** shared test infrastructure for common patterns
4. **Missing** code review processes to catch duplications

