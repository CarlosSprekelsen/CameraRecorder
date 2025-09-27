# MediaMTX On-Demand Configuration Issue - External Advisor Report

**Date:** 2025-09-27  
**Status:** Critical Issue Requiring External Expertise  
**Priority:** High - Blocking Recording Functionality  

## Executive Summary

The MediaMTX Camera Service is experiencing a critical issue where recording API calls succeed but no recording files are created. Investigation has identified that the MediaMTX on-demand path configuration is not being properly applied, preventing the FFmpeg publisher from starting and creating recording files.

## Problem Statement

### Observed Behavior
- `start_recording` API calls return success
- RTSP keepalive reader starts successfully
- No recording files are created in the configured directory
- MediaMTX path shows `runOnDemand: null` and `sourceOnDemand: false`

### Expected Behavior
- `start_recording` API calls return success
- MediaMTX automatically starts FFmpeg publisher via on-demand configuration
- Recording files are created in the configured directory
- MediaMTX path shows proper on-demand configuration

## Technical Investigation

### Architecture Overview
The system follows this flow:
1. **RecordingManager** → **ConfigIntegration.BuildPathConf** → **PathManager.CreatePath** → **MediaMTX API**
2. **RTSPKeepaliveReader** triggers on-demand stream activation
3. **MediaMTX** starts FFmpeg publisher and creates recording files

### Root Cause Analysis

#### Issue 1: PathConf Configuration Mismatch
**Location:** `internal/mediamtx/config_integration.go:BuildPathConf()`

**Current Implementation:**
```go
pathConf := &PathConf{
    Name:   pathName,
    Source: "", // Empty source for on-demand paths
    SourceOnDemand: true,
    RunOnDemand: ci.buildPathCommand(devicePath, pathName),
    RunOnDemandRestart: true,
}
```

**Problem:** MediaMTX API rejects this configuration combination.

#### Issue 2: MediaMTX API Response Analysis
**Evidence from MediaMTX API calls:**
```bash
# Path configuration after creation
curl -s "http://localhost:8888/v3/config/paths/get/camera0"
{
  "name": "camera0",
  "source": null,
  "runOnDemand": null,        # ❌ Should contain FFmpeg command
  "sourceOnDemand": false,    # ❌ Should be true
  "record": true,
  "ready": false
}
```

#### Issue 3: API Endpoint Behavior
**MediaMTX API Endpoints:**
- `POST /v3/config/paths/add/{name}` - Creates path
- `PATCH /v3/config/paths/edit/{name}` - Updates path
- `GET /v3/config/paths/get/{name}` - Retrieves path config

**Observation:** Path creation succeeds but on-demand configuration is not applied.

## Code Analysis

### Current Implementation Files

#### 1. PathConf Structure
**File:** `internal/mediamtx/types.go`
```go
type PathConf struct {
    Name                    string `json:"name"`
    Source                  string `json:"source,omitempty"`
    SourceOnDemand          bool   `json:"sourceOnDemand,omitempty"`
    SourceOnDemandStartTimeout string `json:"sourceOnDemandStartTimeout,omitempty"`
    SourceOnDemandCloseAfter   string `json:"sourceOnDemandCloseAfter,omitempty"`
    RunOnDemand             string `json:"runOnDemand,omitempty"`
    RunOnDemandRestart      bool   `json:"runOnDemandRestart,omitempty"`
    Record                  bool   `json:"record,omitempty"`
    // ... other fields
}
```

#### 2. Path Creation Logic
**File:** `internal/mediamtx/path_manager.go`
```go
func (pm *pathManager) CreatePath(ctx context.Context, name, source string, options *PathConf) error {
    // Builds request using marshalCreateUSBPathRequest
    data, err := marshalCreateUSBPathRequest(name, options.RunOnDemand)
    // Sends POST to /v3/config/paths/add/{name}
}
```

#### 3. USB Path Request Marshaling
**File:** `internal/mediamtx/client.go`
```go
func marshalCreateUSBPathRequest(name, ffmpegCommand string) ([]byte, error) {
    request := map[string]interface{}{
        "source":             "publisher", // Publisher source for on-demand paths
        "runOnDemand":        ffmpegCommand,
        "runOnDemandRestart": true,
    }
    return json.Marshal(request)
}
```

### Configuration Flow Analysis

#### Step 1: BuildPathConf
**Input:** Device path, path name, enable recording flag  
**Output:** PathConf with on-demand configuration  
**Status:** ✅ Correctly builds PathConf

#### Step 2: CreatePath
**Input:** PathConf from BuildPathConf  
**Output:** MediaMTX path creation  
**Status:** ❌ On-demand config not applied

#### Step 3: MediaMTX API
**Input:** JSON request from marshalCreateUSBPathRequest  
**Output:** Path creation response  
**Status:** ❌ Path created but on-demand config missing

## MediaMTX Documentation Analysis

### Required Research Areas

#### 1. On-Demand Configuration Syntax
**Question:** What is the correct JSON structure for on-demand paths?

**Current Attempts:**
```json
// Attempt 1: SourceOnDemand approach
{
  "source": "",
  "sourceOnDemand": true,
  "runOnDemand": "ffmpeg -f v4l2 -i /dev/video0 ...",
  "runOnDemandRestart": true
}

// Attempt 2: Publisher source approach  
{
  "source": "publisher",
  "runOnDemand": "ffmpeg -f v4l2 -i /dev/video0 ...",
  "runOnDemandRestart": true
}
```

#### 2. MediaMTX API Endpoint Behavior
**Question:** Are there different endpoints for different path types?

**Current Usage:**
- `POST /v3/config/paths/add/{name}` - General path creation
- May need specialized endpoint for on-demand paths

#### 3. Configuration Validation
**Question:** What validation rules does MediaMTX apply to path configurations?

**Observations:**
- Path creation succeeds (HTTP 200)
- Configuration not applied (fields remain null/false)
- No error messages returned

## Test Evidence

### Integration Test Results
**Test:** `TestWebSocket_FileLifecycle_Complete_Integration`
**Status:** ❌ FAILING
**Error:** `stat /tmp/recordings/camera0/camera0_2025-09-27_12-25-21.mp4: no such file or directory`

### Log Analysis
```
time="2025-09-27 12:25:22" level=info msg="Creating MediaMTX path"
time="2025-09-27 12:25:22" level=info msg="Sending CreatePath request to MediaMTX - FULL REQUEST"
time="2025-09-27 12:25:22" level=info msg="CreatePath API call completed - FULL RESPONSE"
time="2025-09-27 12:25:22" level=error msg="CreatePath HTTP request failed - investigating idempotency"
time="2025-09-27 12:25:22" level=info msg="MediaMTX path already exists, treating as success"
```

**Analysis:** Path creation reports success but on-demand configuration is not applied.

## External Documentation Requirements

### 1. MediaMTX Official Documentation
**Required:** MediaMTX on-demand path configuration documentation
**URL:** https://github.com/bluenviron/mediamtx
**Focus Areas:**
- On-demand path configuration syntax
- API endpoint specifications
- Configuration validation rules

### 2. MediaMTX API Reference
**Required:** Complete API reference for path management
**Focus Areas:**
- `POST /v3/config/paths/add/{name}` endpoint
- `PATCH /v3/config/paths/edit/{name}` endpoint
- Request/response schemas
- Error handling

### 3. MediaMTX Configuration Examples
**Required:** Working examples of on-demand path configurations
**Focus Areas:**
- V4L2 device on-demand paths
- FFmpeg command integration
- Recording-enabled paths

## Proposed Investigation Steps

### Phase 1: Documentation Research
1. **MediaMTX GitHub Repository**
   - Review official documentation
   - Analyze configuration examples
   - Study API endpoint specifications

2. **MediaMTX Community Resources**
   - Search for on-demand configuration examples
   - Review issue reports and solutions
   - Check community forums and discussions

### Phase 2: API Testing
1. **Direct MediaMTX API Testing**
   - Test different JSON configurations
   - Validate endpoint behavior
   - Document successful configurations

2. **Configuration Validation**
   - Test minimal working configurations
   - Identify required vs optional fields
   - Document validation rules

### Phase 3: Implementation Fix
1. **Code Updates**
   - Update PathConf structure if needed
   - Fix marshalCreateUSBPathRequest
   - Implement proper error handling

2. **Testing and Validation**
   - Verify recording file creation
   - Test integration scenarios
   - Validate performance characteristics

## Risk Assessment

### High Risk
- **Recording functionality completely non-functional**
- **Integration tests failing**
- **Production deployment blocked**

### Medium Risk
- **Architecture complexity increasing**
- **Maintenance burden**
- **Performance impact**

### Low Risk
- **Other functionality unaffected**
- **System stability maintained**
- **Rollback capability available**

## Success Criteria

### Primary Goals
1. **Recording files created successfully**
2. **Integration tests passing**
3. **On-demand configuration working**

### Secondary Goals
1. **Performance optimization**
2. **Error handling improvement**
3. **Documentation completeness**

## Conclusion

This issue requires external expertise in MediaMTX configuration and API usage. The current implementation follows logical patterns but may not align with MediaMTX's specific requirements for on-demand path configuration.

**Recommendation:** Engage MediaMTX community or experts to provide guidance on correct on-demand path configuration syntax and API usage patterns.

## Contact Information

**Project:** MediaMTX Camera Service Go Implementation  
**Repository:** https://github.com/camerarecorder/mediamtx-camera-service-go  
**Issue:** Recording file creation failure due to on-demand configuration  
**Priority:** Critical - Blocking core functionality  

---

**Report Generated:** 2025-09-27  
**Next Review:** After external consultation  
**Status:** Awaiting external expertise
