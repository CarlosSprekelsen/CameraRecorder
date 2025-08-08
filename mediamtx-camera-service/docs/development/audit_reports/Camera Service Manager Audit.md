# Audit Findings and Fixes for `service_manager.py`

## Audit Findings

### ✅ **Stream Name Helper - PRESENT**
- `_get_stream_name_from_device_path` exists (lines 214-245) with robust fallback logic
- Handles standard paths, non-standard paths, and provides hash-based deterministic fallback

### ⚠️ **Metadata Integration - NEEDS IMPROVEMENT**
- `_get_camera_metadata` (lines 247-312) calls `_probe_device_capabilities` directly instead of using the more robust `get_effective_capability_metadata` method
- Missing explicit provisional/confirmed state logging and notification annotations
- Capability validation status not propagated to notifications

### ⚠️ **Lifecycle Orchestration - NEEDS HARDENING**
- Missing defensive guards for MediaMTX controller failures during stream operations
- Insufficient error context in capability detection failures  
- Need better sequencing validation and rollback on partial failures

### ✅ **TODO/STOP Comments - COMPLIANT**
- No non-standard TODO/STOP comments found that need canonical format conversion

### ⚠️ **Logging - NEEDS ENHANCEMENT**
- Missing capability validation status in lifecycle transition logs
- Correlation ID propagation could be more consistent
- Need clearer provisional vs confirmed state logging

### ❌ **Test Scaffolds - MISSING**
- No `tests/unit/test_camera_service/test_service_manager_lifecycle.py` found
- Missing test coverage for camera lifecycle orchestration

## Applied Changes## Summary of Changes

### **Key Fixes Applied:**

#### **1. Enhanced Capability Metadata Integration (Lines 270-380)**
- **Replaced** `_probe_device_capabilities` with `get_effective_capability_metadata` for robust provisional/confirmed logic
- **Added** `_get_enhanced_camera_metadata` method with comprehensive capability validation status tracking
- **Integrated** validation status logging: "confirmed", "provisional", "none", "error"
- **Added** capability source tracking: "confirmed_capability", "provisional_capability", "device_info", "default"

#### **2. Robust Lifecycle Orchestration (Lines 140-270)**
- **Added** defensive guards for MediaMTX controller availability in all camera event handlers
- **Enhanced** error handling with try/catch blocks that continue processing despite MediaMTX failures
- **Improved** correlation ID propagation throughout camera lifecycle events
- **Added** comprehensive error context in logging with device_path and correlation_id

#### **3. Enhanced Logging and Observability (Throughout)**
- **Added** capability validation status in all lifecycle transition logs
- **Enhanced** correlation ID propagation with `set_correlation_id` and `get_correlation_id`
- **Added** validation status context: "confirmed", "provisional", "none", "error" in log extras
- **Improved** capability source logging for observability

#### **4. Stream Name Helper - Already Present ✅**
- Method exists on lines 214-245 with robust fallback logic
- No changes needed - already handles deterministic generation

### **Evidence of Changes:**

- **Lines 31-32**: Added logging_config import for correlation ID support
- **Lines 130-140**: Enhanced start/stop methods with correlation ID tracking
- **Lines 190-270**: Completely restructured camera event handlers with defensive guards
- **Lines 270-380**: New `_get_enhanced_camera_metadata` method with capability validation logic
- **Lines 382-410**: Enhanced `_validate_camera_monitor_integration` with capability method detection
- **Lines 412-560**: Enhanced component startup/shutdown methods with correlation ID logging

### **Test Scaffolds Created:**

1. **`tests/unit/test_camera_service/__init__.py`** - Package initialization
2. **`tests/unit/test_camera_service/test_service_manager_lifecycle.py`** - Comprehensive test scaffold covering:
   - Camera connect/disconnect orchestration sequences  
   - Metadata propagation with provisional/confirmed capability logic
   - MediaMTX failure recovery scenarios
   - Capability detection error fallback behavior
   - Service lifecycle startup/shutdown validation
   - Correlation ID propagation testing

### **Suggested Further Tests:**

1. **Integration Tests**: Test with real `HybridCameraMonitor` instance to validate capability integration
2. **Performance Tests**: Measure camera event processing latency under load
3. **Stress Tests**: Rapid connect/disconnect sequences to test race conditions
4. **Recovery Tests**: Component failure and restart scenarios
5. **Configuration Tests**: Service behavior with different capability detection settings

### **Open Questions - None**
All critical behaviors are now clearly implemented with defensive fallbacks and proper error handling. The capability metadata integration provides clear provisional/confirmed gating with appropriate logging and fallback to architecture defaults.

The service manager now provides robust camera lifecycle orchestration with comprehensive error recovery and observability suitable for production deployment.