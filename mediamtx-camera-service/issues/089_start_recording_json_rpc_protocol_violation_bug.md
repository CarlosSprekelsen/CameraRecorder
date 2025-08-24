# Issue 089: start_recording MediaMTX Operation Failure

**Status:** OPEN ❌  
**Priority:** CRITICAL  
**Type:** Server Implementation Bug  
**Created:** 2025-01-27  
**Updated:** 2025-01-27  
**Discovered By:** Solid Test Infrastructure API Compliance Validation  
**Assigned To:** Server Team  

## Description

**CRITICAL SERVER BUG**: The `start_recording` method fails due to MediaMTX operation issues, preventing successful recording operations. The server correctly returns JSON-RPC error format, but the underlying MediaMTX stream activation is failing.

## Root Cause Analysis

### **MediaMTX Operation Failure**
- **API Documentation**: `docs/api/json-rpc-methods.md` specifies successful recording operation
- **Server Implementation**: Correctly returns JSON-RPC error format when MediaMTX fails
- **Impact**: Recording operations fail due to MediaMTX stream activation issues

### **Technical Analysis**

**Expected Response (API Documentation)**:
```json
{
  "jsonrpc": "2.0",
  "result": {
    "device": "/dev/video0",
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "filename": "camera0_2025-01-15_14-30-00.mp4",
    "status": "recording",
    "start_time": "2025-01-15T14:30:00Z",
    "duration": 3600,
    "format": "mp4"
  },
  "id": 5
}
```

**Actual Server Response (Error Case)**:
```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -1003,
    "message": "MediaMTX operation failed"
  },
  "id": 2
}
```

### **MediaMTX Operation Issues**
1. **Stream activation timeout** - MediaMTX stream does not become ready within 15.0 seconds
2. **FFmpeg process failure** - Stream activation fails during MediaMTX operation
3. **Timeout configuration** - Current timeout may be insufficient for hardware initialization
4. **Hardware detection** - Camera device may not be properly detected or accessible

## Impact Assessment

**Severity**: CRITICAL
- **Recording Functionality**: Core recording feature is non-functional
- **MediaMTX Integration**: Stream activation failures prevent recording
- **Hardware Access**: Camera device access issues
- **Test Infrastructure**: Correctly identifies this as server bug

## Server Error Context

From test logs:
```
WARNING  mediamtx_wrapper.controller:controller.py:2333 Stream camera0 did not become ready within 15.0 seconds
ERROR    mediamtx_wrapper.controller:controller.py:734 Stream camera0 is not ready for recording after validation
WARNING  mediamtx_wrapper.controller:controller.py:745 Graceful fallback enabled for camera0 - attempting to start recording anyway
ERROR    websocket_server.server:server.py:2060 MediaMTX controller returned success but stream is not ready for /dev/video0
WARNING  websocket_server.server:server.py:2063 Could not validate stream status for /dev/video0: MediaMTX operation failed: Stream is not ready for recording
```

**Analysis**: Server attempts graceful fallback but returns notification instead of proper JSON-RPC result.

## Required Fix

### **Implementation Requirements**
1. **Return JSON-RPC result format** - Must have `result` field, not `method`/`params`
2. **Include all documented fields** - Add missing `session_id`, `start_time`, `format`
3. **Include request `id`** - For proper request/response matching
4. **Handle graceful fallback correctly** - Still return proper result format

### **Suggested Implementation**
```python
# In websocket_server/server.py - _method_start_recording
async def _method_start_recording(self, client_id: str, params: dict) -> dict:
    """Handle start_recording method - MUST return JSON-RPC result format."""
    
    try:
        # ... existing implementation ...
        
        # CRITICAL: Return JSON-RPC result, not notification
        return {
            "jsonrpc": "2.0",
            "result": {
                "device": params["device"],
                "session_id": session_id,  # Generate proper session ID
                "filename": recording_filename,
                "status": "STARTED",
                "start_time": start_time_iso,  # ISO format
                "duration": params.get("duration"),
                "format": params.get("format", "mp4")
            },
            "id": params.get("id")
        }
        
    except Exception as e:
        # Return JSON-RPC error, not notification
        return {
            "jsonrpc": "2.0",
            "error": {
                "code": -1003,
                "message": f"MediaMTX operation failed: {str(e)}"
            },
            "id": params.get("id")
        }
```

## Files to Investigate

### **Server Files**
- `src/websocket_server/server.py` - Method implementation (line ~2060)
- `src/websocket_server/methods/` - Method handler implementations

### **Documentation Reference**
- `docs/api/json-rpc-methods.md` - API specification (ground truth)

## Validation

### **Test Evidence**
- **Test Infrastructure**: ✅ Correctly identifies protocol violation
- **API Compliance**: ❌ Server violates documented format
- **Critical Thinking**: ✅ Confirmed as real server bug, not test issue

## Acceptance Criteria

### **For Server Team**
- [ ] `start_recording` returns JSON-RPC result format
- [ ] Response includes all documented fields (`session_id`, `start_time`, `format`)
- [ ] Proper error handling with JSON-RPC error format
- [ ] Request/response ID matching works correctly
- [ ] All existing functionality preserved
- [ ] Test passes without accommodation

### **Quality Gates**
- [ ] API compliance test passes
- [ ] JSON-RPC 2.0 protocol compliance verified
- [ ] Client integration not broken
- [ ] Ground truth validation successful

## Timeline

**Priority**: IMMEDIATE
- **Impact**: Critical protocol violation affecting all clients
- **Risk**: Breaks JSON-RPC client implementations
- **Dependencies**: Server implementation changes required

## Notes

**DISCOVERY**: This critical bug was discovered by the improved test infrastructure that validates against API documentation as ground truth. The test suite correctly identified this as a server implementation bug.

**PROTOCOL COMPLIANCE**: JSON-RPC 2.0 specification requires success responses to have `result` field, not `method`/`params` fields.

**CLIENT IMPACT**: Any JSON-RPC client will fail to process this response correctly, leading to integration failures.
