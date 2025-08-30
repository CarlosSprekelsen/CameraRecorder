# MediaMTX Client Migration Guide

**Version:** 1.0  
**Date:** 2025-08-30  
**Status:** Migration Guide for MediaMTX Test Infrastructure  

## 🎯 **Migration Goal**

Eliminate code duplication and create a centralized, maintainable MediaMTX test infrastructure by migrating from individual client creation to shared utilities.

## 📊 **Current State Analysis**

### **Problems Identified:**
- **50+ instances** of `mediamtx.NewClient("http://localhost:9997", testConfig, logger)` across test files
- **Code duplication** in every MediaMTX test file
- **Inconsistent configurations** across different tests
- **Hardcoded URLs** repeated everywhere
- **Poor maintainability** - changes require updating multiple files

### **Files Requiring Migration:**
1. `test_mediamtx_client_test.go` (15 instances)
2. `test_mediamtx_path_manager_test.go` (10 instances)
3. `test_mediamtx_health_monitor_test.go` (9 instances)
4. `test_mediamtx_stream_lifecycle_test.go` (4 instances)
5. `test_mediamtx_path_integration_test.go` (1 instance)
6. `test_mediamtx_stream_manager_test.go` (1 instance)

## 🔄 **Migration Pattern**

### **OLD PATTERN (To Be Replaced):**
```go
func TestMediaMTXFeature(t *testing.T) {
    // Individual setup in each test
    testConfig := &mediamtx.MediaMTXConfig{
        BaseURL:       "http://localhost:9997",
        Timeout:       30 * time.Second,
        RetryAttempts: 3,
        RetryDelay:    1 * time.Second,
    }
    
    client := mediamtx.NewClient("http://localhost:9997", testConfig, logger)
    
    // Test logic...
}
```

### **NEW PATTERN (Using Utilities):**
```go
func TestMediaMTXFeature(t *testing.T) {
    // COMMON PATTERN: Use shared test environment
    env := utils.SetupTestEnvironment(t)
    defer utils.TeardownTestEnvironment(t, env)

    // NEW PATTERN: Use centralized MediaMTX client setup
    client := utils.SetupMediaMTXTestClient(t, env)
    defer utils.TeardownMediaMTXTestClient(t, client)

    // Test logic using client.Client...
}
```

## 🛠️ **Available Utilities**

### **1. Basic Client Setup**
```go
// Standard client with default configuration
client := utils.SetupMediaMTXTestClient(t, env)

// Client with custom configuration
customConfig := utils.CreateMediaMTXTestConfigWithTimeout(env.TempDir, 5*time.Second)
client := utils.SetupMediaMTXTestClientWithConfig(t, env, customConfig)
```

### **2. Manager Setup**
```go
// Health monitor
healthMonitor := utils.SetupMediaMTXHealthMonitor(t, client)

// Stream manager
streamManager := utils.SetupMediaMTXStreamManager(t, client)

// Path manager
pathManager := utils.SetupMediaMTXPathManager(t, client)

// Recording manager
recordingManager := utils.SetupMediaMTXRecordingManager(t, client)

// Snapshot manager
snapshotManager := utils.SetupMediaMTXSnapshotManager(t, client)
```

### **3. Test Data Creation**
```go
// Test path configuration
testPath := utils.CreateMediaMTXTestPath("test-path")

// Test stream configuration
testStream := utils.CreateMediaMTXTestStream("test-stream")
```

### **4. Connection Testing**
```go
// Test if MediaMTX service is accessible
isAccessible := utils.TestMediaMTXConnection(t, client)
if !isAccessible {
    t.Skip("MediaMTX service not accessible, skipping test")
}
```

## 📋 **Migration Checklist**

### **Phase 1: Infrastructure Setup** ✅
- [x] Create `mediamtx_test_utils.go` with centralized utilities
- [x] Create migration example test file
- [x] Verify utilities work with real MediaMTX service
- [x] Document migration patterns

### **Phase 2: Progressive Migration**
- [ ] Migrate `test_mediamtx_client_test.go` (15 instances)
- [ ] Migrate `test_mediamtx_path_manager_test.go` (10 instances)
- [ ] Migrate `test_mediamtx_health_monitor_test.go` (9 instances)
- [ ] Migrate `test_mediamtx_stream_lifecycle_test.go` (4 instances)
- [ ] Migrate `test_mediamtx_path_integration_test.go` (1 instance)
- [ ] Migrate `test_mediamtx_stream_manager_test.go` (1 instance)

### **Phase 3: Validation**
- [ ] Run all MediaMTX tests to ensure they still pass
- [ ] Verify coverage improvements
- [ ] Remove old pattern code
- [ ] Update documentation

## 🔧 **Migration Steps for Each File**

### **Step 1: Update Imports**
```go
// Add utils import if not present
import (
    "github.com/camerarecorder/mediamtx-camera-service-go/tests/utils"
)
```

### **Step 2: Replace Client Creation**
```go
// OLD:
testConfig := &mediamtx.MediaMTXConfig{...}
client := mediamtx.NewClient("http://localhost:9997", testConfig, logger)

// NEW:
client := utils.SetupMediaMTXTestClient(t, env)
defer utils.TeardownMediaMTXTestClient(t, client)
```

### **Step 3: Update Client Usage**
```go
// OLD:
data, err := client.Get(ctx, "/v3/config/global/get")

// NEW:
data, err := client.Client.Get(ctx, "/v3/config/global/get")
```

### **Step 4: Add Environment Setup**
```go
// Add at the beginning of each test function
env := utils.SetupTestEnvironment(t)
defer utils.TeardownTestEnvironment(t, env)
```

## 🚨 **Critical Migration Rules**

### **1. Separation of Concerns**
- **DO NOT** modify implementation code to make tests pass
- **DO** fix test infrastructure issues
- **STOP** if you find implementation errors

### **2. Progressive Migration**
- Migrate one file at a time
- Test each migration before proceeding
- Keep old and new patterns working during transition

### **3. Error Handling**
- If MediaMTX service is not accessible, skip tests gracefully
- Log connection issues for debugging
- Don't fail tests due to infrastructure issues

### **4. Configuration Management**
- Use standardized configurations from utilities
- Only create custom configs when absolutely necessary
- Document any deviations from standard config

## 📈 **Expected Benefits**

### **Immediate Benefits:**
- **Eliminate 50+ code duplications**
- **Centralized configuration management**
- **Consistent test behavior**
- **Easier maintenance**

### **Long-term Benefits:**
- **Improved test reliability**
- **Better coverage measurement**
- **Faster test development**
- **Reduced technical debt**

## 🔍 **Validation Commands**

### **Test New Utilities:**
```bash
go test -tags=unit -coverpkg=./internal/mediamtx ./tests/unit/test_mediamtx_client_migration_example_test.go -v
```

### **Test Specific File Migration:**
```bash
go test -tags=unit -coverpkg=./internal/mediamtx ./tests/unit/test_mediamtx_client_test.go -v
```

### **Check Coverage Improvement:**
```bash
go test -tags=unit -coverpkg=./internal/mediamtx ./tests/unit/ -coverprofile=mediamtx_coverage.out
go tool cover -func=mediamtx_coverage.out
```

## 📞 **Support**

If you encounter issues during migration:
1. Check the example test file for patterns
2. Verify MediaMTX service is running (`systemctl status mediamtx`)
3. Test connection manually (`curl http://localhost:9997/v3/config/global/get`)
4. Review this migration guide for patterns

**Remember:** The goal is to create solid test infrastructure, not to make tests pass. If you find implementation errors, STOP and report them.
