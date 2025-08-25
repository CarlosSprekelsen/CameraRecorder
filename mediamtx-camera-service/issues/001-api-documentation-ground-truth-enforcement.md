# Issue #001: API Documentation Gap - get_streams Method

**Status**: Critical  
**Priority**: High  
**Type**: Documentation Gap  
**Created**: 2025-01-15  
**Assigned**: Documentation Team  

## Problem Statement

The `get_streams` method is **implemented and functional** in the Python server but has **incomplete documentation** in the API reference. This violates the ground truth principle where API documentation should drive implementation, not follow it.

## Evidence of Implementation

### 1. Server Implementation Exists
**File**: `src/websocket_server/server.py` (lines 2446-2480)
```python
async def _method_get_streams(self, params: Optional[Dict[str, Any]] = None) -> List[Dict[str, Any]]:
    """
    Get list of all active streams from MediaMTX.
    Returns: List of stream information dictionaries
    """
```

### 2. Method Registration Confirmed
**File**: `src/websocket_server/server.py` (line 1270)
```python
self.register_method("get_streams", self._method_get_streams, version="1.0"
```

### 3. Authentication Permissions Set
**File**: `src/websocket_server/server.py` (line 734)
```python
"get_streams": "viewer",
```

### 4. Tests Exist and Pass
**Files**: Multiple test files reference and test `get_streams` method
- `tests/integration/test_critical_interfaces.py` (lines 237-293)
- `tests/integration/test_sdk_response_format.py` (lines 163-215)
- `tests/integration/test_security_authentication.py` (lines 259-291)

## Documentation Gap Identified

### Current API Documentation Status
**File**: `docs/api/json-rpc-methods.md`
- ✅ Method is listed in API reference
- ❓ **Documentation may be incomplete or missing examples**
- ❓ **Response format may not be fully documented**
- ❓ **Error conditions may not be documented**

### Required Documentation Updates
1. **Complete method documentation** with request/response examples
2. **Response format specification** with all field descriptions
3. **Error condition documentation** with proper error codes
4. **Usage examples** for client integration
5. **Authentication requirements** clarification

## Impact Assessment

### High Impact Areas
- **Client Development**: Incomplete documentation hinders client integration
- **Go Migration**: Documentation gap affects API compatibility validation
- **Quality Assurance**: Missing documentation makes testing incomplete
- **Ground Truth Violation**: Implementation exists without proper documentation

### Risk Mitigation
- **Immediate**: Complete `get_streams` method documentation
- **Short-term**: Review all API methods for similar gaps
- **Long-term**: Establish documentation-first development process

## Acceptance Criteria

1. **Complete `get_streams` documentation** - Full request/response examples
2. **Response format specification** - All fields documented with types
3. **Error condition documentation** - Proper error codes and messages
4. **Usage examples provided** - Client integration examples
5. **Documentation review completed** - All API methods verified

## Files Requiring Updates

### Primary Documentation
- `docs/api/json-rpc-methods.md` - Complete `get_streams` documentation

### Supporting Documentation
- `docs/api/examples/` - Add usage examples
- `docs/api/error-codes.md` - Document error conditions

## Related Issues
- Go Migration: API compatibility validation
- Ground Truth Enforcement: Documentation-first development

## Notes
This issue demonstrates a violation of the ground truth principle where implementation exists without proper documentation. The `get_streams` method is functional but needs complete documentation to serve as proper ground truth for the Go migration project.
