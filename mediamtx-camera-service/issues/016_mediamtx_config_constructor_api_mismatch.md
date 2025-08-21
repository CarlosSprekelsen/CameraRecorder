# Issue 016: MediaMTXConfig Constructor API Mismatch

## Summary
Tests are passing a non-existent `codec` parameter to the MediaMTXConfig constructor.

## Status
**OPEN** 🚨

## Priority
**MEDIUM**

## Category
**Test Infrastructure**

## Description
A test is failing with `TypeError: MediaMTXConfig.__init__() got an unexpected keyword argument 'codec'` because it's passing a parameter that doesn't exist in the constructor.

## Root Cause
**API Inconsistency**: The requirements specify STANAG 4406 H.264 codec support, but the MediaMTXConfig API doesn't include a `codec` parameter. The ConfigManager is trying to pass a `codec` parameter from configuration data to MediaMTXConfig constructor, but MediaMTXConfig doesn't support it.

## Impact
- Test failure preventing test suite execution
- **API inconsistency between requirements and implementation**
- STANAG 4406 H.264 codec requirements not properly supported in API

## Evidence
```
TypeError: MediaMTXConfig.__init__() got an unexpected keyword argument 'codec'
```

## Ground Truth Analysis
**Requirements vs API Mismatch:**

1. **Requirements (Ground Truth)**:
   - STANAG 4406 H.264 compliance is a stakeholder requirement
   - System must support H.264 codec configuration
   - STANAG 4406 compliance tests exist and expect codec configuration

2. **Current API (Implementation)**:
   - MediaMTXConfig constructor has NO `codec` parameter
   - Available parameters: host, api_port, rtsp_port, webrtc_port, hls_port, config_path, recordings_path, snapshots_path, health_* parameters
   - **Missing**: codec configuration parameter

3. **Configuration Flow**:
   - ConfigManager loads configuration data including `codec` setting
   - Tries to pass `**mediamtx_data` to MediaMTXConfig constructor
   - MediaMTXConfig rejects `codec` parameter because it doesn't exist

## Investigation Results
**Root Cause Found**: This is an API design inconsistency where:
- Requirements specify STANAG 4406 H.264 codec support
- MediaMTXConfig API doesn't include codec configuration
- ConfigManager expects codec parameter to be supported
- Tests expect codec configuration to work

## Requirements Traceability
- **STANAG 4406 Compliance**: Requires H.264 codec configuration
- **API Design**: MediaMTXConfig should support codec parameters
- **Configuration Management**: ConfigManager expects codec support

## Requirements Traceability
- **Test Infrastructure**: Tests must use correct API signatures
- **API Consistency**: Test expectations must match actual implementation

## Next Steps
1. Investigate the specific test code causing this error
2. Check MediaMTXConfig constructor documentation
3. Determine correct API usage
4. Update test to use proper API signature
