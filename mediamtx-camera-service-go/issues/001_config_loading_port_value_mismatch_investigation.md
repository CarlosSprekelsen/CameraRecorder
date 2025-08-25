# Issue 001: Configuration Loading Port Value Mismatch Investigation

**Issue ID:** 001  
**Type:** Investigation  
**Priority:** Medium  
**Status:** Open  
**Created:** 2025-01-15  
**Assigned:** Developer  
**Related Story:** S1.1 Configuration Management System  

## Summary

Investigation required for port value mismatch in `TestConfigManager_LoadConfig_ValidYAML` test. The test expects specific port values but receives different values, indicating potential configuration loading or test infrastructure issues.

## Problem Description

### **Test Failure Details**
- **Test:** `TestConfigManager_LoadConfig_ValidYAML`
- **Location:** `tests/unit/test_config_management_test.go:192-194`
- **Expected Values:** 8555, 8890, 8889
- **Actual Values:** 8554, 8889, 8888
- **Difference:** All values off by 1

### **Error Output**
```
Error: Not equal: 
        expected: 8555
        actual  : 8554
Error: Not equal: 
        expected: 8890
        actual  : 8889
Error: Not equal: 
        expected: 8889
        actual  : 8888
```

## Root Cause Analysis Required

### **Investigation Areas**

#### **1. Configuration Loading Logic**
- **Question:** Is the YAML file being loaded correctly?
- **Evidence:** Test YAML defines ports 8555, 8890, 8889
- **Investigation:** Verify configuration loading implementation

#### **2. Default Value Override**
- **Question:** Are default values overriding the YAML values?
- **Evidence:** Actual values (8554, 8889, 8888) match default configuration
- **Investigation:** Check if defaults are applied when YAML loading fails

#### **3. YAML Parsing Issues**
- **Question:** Is the YAML parsing working correctly?
- **Evidence:** Other configuration values load correctly
- **Investigation:** Verify YAML parsing for port fields specifically

#### **4. Test Infrastructure**
- **Question:** Is the test using the correct YAML file?
- **Evidence:** Test creates YAML file with correct values
- **Investigation:** Verify file creation and loading process

### **Code Analysis Required**

#### **Test YAML Content (Expected)**
```yaml
mediamtx:
  rtsp_port: 8555
  webrtc_port: 8890
  hls_port: 8889
```

#### **Default Configuration (Actual)**
```yaml
mediamtx:
  rtsp_port: 8554
  webrtc_port: 8889
  hls_port: 8888
```

## Investigation Steps

### **Step 1: Verify Configuration Loading**
1. Add debug logging to configuration loading process
2. Verify YAML file content is correct
3. Check if configuration manager loads the file

### **Step 2: Check Default Value Logic**
1. Review default value application logic
2. Verify when defaults are used vs YAML values
3. Check for any fallback mechanisms

### **Step 3: Validate YAML Parsing**
1. Test YAML parsing with minimal configuration
2. Verify port field mapping
3. Check for type conversion issues

### **Step 4: Test Infrastructure Validation**
1. Verify test file creation process
2. Check file permissions and access
3. Validate configuration manager initialization

## Expected Outcomes

### **Possible Root Causes**
1. **Configuration Loading Bug:** YAML file not loaded, defaults used instead
2. **Default Value Bug:** Defaults incorrectly override YAML values
3. **YAML Parsing Bug:** Port fields not parsed correctly
4. **Test Infrastructure Bug:** Wrong file or configuration used

### **Success Criteria**
- **Configuration loads YAML values correctly**
- **Test passes with expected port values**
- **No regression in other configuration loading**

## Impact Assessment

### **Current Impact**
- **Test Failure:** Prevents test suite from passing
- **Functionality:** Configuration system may not work as expected
- **Quality:** Indicates potential configuration loading issues

### **Risk Level**
- **Medium:** Affects configuration loading validation
- **Scope:** Limited to port value loading
- **Mitigation:** Investigation and fix required

## Next Steps

1. **Developer Investigation:** Perform root cause analysis
2. **Code Review:** Review configuration loading implementation
3. **Test Validation:** Verify fix resolves the issue
4. **Documentation:** Update test documentation if needed

## Related Files

- `tests/unit/test_config_management_test.go` (lines 192-194)
- `internal/config/config_manager.go` (configuration loading logic)
- `internal/config/config_types.go` (port field definitions)

## Notes

- **Test passes when run individually** but fails in full suite
- **Other configuration values load correctly**
- **Issue appears to be specific to port fields**
- **May indicate test isolation or configuration loading problem**

---

## Resolution

### **Root Cause Identified**
- **Issue:** Go test cache causing stale test results
- **Solution:** Clear cache with `go clean -cache`
- **Result:** Port values now load correctly (8555, 8890, 8889)

### **Lessons Learned**
- **Cache Issue:** Go test cache can cause unexpected test failures
- **Solution:** Added `go clean -cache` to testing guide
- **Prevention:** Clear cache when tests behave unexpectedly

**Issue Status:** Resolved  
**Resolution Date:** 2025-01-15  
**Priority:** Medium (resolved with cache clearing)
