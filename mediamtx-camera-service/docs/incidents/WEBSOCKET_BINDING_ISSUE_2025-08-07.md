# Incident Report: WebSocket Server Binding Issue

**Date:** August 7, 2025  
**Severity:** High  
**Status:** Resolved  
**Incident ID:** INC-2025-08-07-001  

---

## Executive Summary

The MediaMTX Camera Service WebSocket server was not binding to port 8002, preventing client connections. The issue was caused by multiple configuration and import path problems that were not caught during development and testing phases.

### **Impact:**
- **Service Availability:** WebSocket server unavailable for client connections
- **Client Development:** Blocked client development and testing
- **Sprint 3:** Delayed Sprint 3 authorization due to blocking issue
- **User Experience:** Service appeared to be running but was not accessible

---

## Incident Timeline

### **Discovery (2025-08-07 00:47)**
- Service reported as running but WebSocket endpoint not accessible
- Port 8002 not listening despite service process active
- Client connection attempts failing

### **Initial Investigation (2025-08-07 00:48-00:52)**
- Service status showed `activating (auto-restart)` with exit code 1
- Logs revealed `ModuleNotFoundError: No module named 'src.camera'`
- Identified Python module import path issues

### **Root Cause Analysis (2025-08-07 00:53-00:56)**
- Discovered multiple import path problems in service files
- Found configuration schema mismatches
- Identified missing directory permissions

### **Resolution Implementation (2025-08-07 00:57-01:01)**
- Fixed Python import paths in all service files
- Corrected configuration file schema
- Updated systemd service configuration
- Created required directories with proper permissions

---

## Root Cause Analysis

### **Primary Issues:**

#### **1. Python Module Import Path Problems**
**Root Cause:** Service installed with absolute imports instead of relative imports
**Files Affected:**
- `src/camera_service/service_manager.py`
- `src/camera_service/logging_config.py`
- `src/mediamtx_wrapper/controller.py`
- `src/websocket_server/server.py`

**Technical Details:**
```python
# INCORRECT (absolute imports)
from mediamtx_wrapper.controller import MediaMTXController
from camera_service.logging_config import set_correlation_id

# CORRECT (relative imports)
from ..mediamtx_wrapper.controller import MediaMTXController
from ..camera_service.logging_config import set_correlation_id
```

#### **2. Configuration Schema Mismatches**
**Root Cause:** Configuration file parameters didn't match expected schema
**Issues Fixed:**
- `api_url` → `api_port` (MediaMTXConfig expects `api_port`)
- `file` → `file_path` (LoggingConfig expects `file_path`)
- `max_size` → `max_file_size` (LoggingConfig expects `max_file_size`)
- Removed `api_timeout` (not a valid MediaMTXConfig parameter)
- Removed `enabled` from recording section (not a valid RecordingConfig parameter)

#### **3. Missing Directory Permissions**
**Root Cause:** Installation script didn't create required directories with proper permissions
**Directories Missing:**
- `/var/recordings`
- `/var/snapshots`

**Permission Issues:**
- Directories didn't exist
- Service user `camera-service` couldn't access directories
- Service failed during startup validation

#### **4. Service Configuration Issues**
**Root Cause:** Systemd service configuration had incorrect module path
**Issue Fixed:**
```bash
# INCORRECT
ExecStart=/opt/camera-service/venv/bin/python -m src.camera_service

# CORRECT
ExecStart=/opt/camera-service/venv/bin/python -m src.camera_service.main
```

---

## Resolution Actions

### **1. Fixed Python Import Paths**
**Files Updated:**
- `src/camera_service/service_manager.py` - Fixed `mediamtx_wrapper` and `camera_discovery` imports
- `src/camera_service/logging_config.py` - Fixed `camera_service.config` import
- `src/mediamtx_wrapper/controller.py` - Fixed `camera_service.logging_config` import
- `src/websocket_server/server.py` - Fixed `camera_service.logging_config` import

**Changes Made:**
```python
# Before
from mediamtx_wrapper.controller import MediaMTXController
from camera_service.logging_config import set_correlation_id

# After
from ..mediamtx_wrapper.controller import MediaMTXController
from ..camera_service.logging_config import set_correlation_id
```

### **2. Corrected Configuration Schema**
**File:** `config/camera-service.yaml`
**Changes Made:**
```yaml
# Before
mediamtx:
  api_url: "http://localhost:9997"
  api_timeout: 30

logging:
  file: "/var/log/camera-service/camera-service.log"
  max_size: "10MB"

# After
mediamtx:
  host: "localhost"
  api_port: 9997

logging:
  file_enabled: true
  file_path: "/var/log/camera-service/camera-service.log"
  max_file_size: "10MB"
```

### **3. Updated Service Configuration**
**File:** `/etc/systemd/system/camera-service.service`
**Change Made:**
```bash
# Before
ExecStart=/opt/camera-service/venv/bin/python -m src.camera_service

# After
ExecStart=/opt/camera-service/venv/bin/python -m src.camera_service.main
```

### **4. Created Required Directories**
**Commands Executed:**
```bash
sudo mkdir -p /var/recordings /var/snapshots
sudo chown camera-service:camera-service /var/recordings /var/snapshots
sudo chmod 755 /var/recordings /var/snapshots
```

---

## Why This Was Not Caught in Tests

### **1. Environment Differences**
- **Development:** Tests run from project root with `PYTHONPATH` set
- **Production:** Service runs from `/opt/camera-service` with different Python path
- **Gap:** No production-like environment testing

### **2. Configuration Testing Gaps**
- **Unit Tests:** Test individual components with mocked configurations
- **Integration Tests:** Test component interactions but not production config files
- **Gap:** No validation of actual configuration file schema

### **3. Installation Testing Gaps**
- **Current Tests:** Test service functionality but not installation process
- **Missing:** Installation validation, directory creation, permission setup
- **Gap:** No end-to-end installation testing

### **4. Service Lifecycle Testing Gaps**
- **Current Tests:** Test service components but not systemd service startup
- **Missing:** Service startup validation, WebSocket binding verification
- **Gap:** No production service lifecycle testing

---

## Impact Assessment

### **Immediate Impact:**
- **Service Availability:** WebSocket server unavailable for ~1 hour
- **Client Development:** Blocked client development and testing
- **Sprint 3:** Delayed Sprint 3 authorization

### **Long-term Impact:**
- **Process Improvements:** Identified gaps in testing and deployment processes
- **Documentation:** Need for better installation and troubleshooting guides
- **Testing:** Need for production-like testing environment

---

## Lessons Learned

### **1. Testing Environment Gaps**
- Need production-like testing environment
- Must test actual installation process
- Must validate service startup in production conditions

### **2. Configuration Management**
- Need better configuration schema validation
- Must test actual configuration files, not just mocked configs
- Need configuration file format documentation

### **3. Deployment Process**
- Need pre-deployment validation scripts
- Must test complete service lifecycle
- Need better error handling and logging

### **4. Documentation Gaps**
- Need comprehensive installation troubleshooting guide
- Must document configuration schema requirements
- Need production deployment validation procedures

---

## Action Items

### **Immediate (This Sprint):**
- [x] Fix Python import paths in all service files
- [x] Correct configuration file schema
- [x] Update systemd service configuration
- [x] Create required directories with proper permissions
- [ ] Update installation script to include directory creation
- [ ] Add installation validation tests
- [ ] Re-run integration tests with production-like environment

### **Short Term (Next Sprint):**
- [ ] Update all documentation with correct procedures
- [ ] Implement pre-deployment validation script
- [ ] Add production-like testing environment
- [ ] Update CI/CD pipeline with deployment validation

### **Long Term (Ongoing):**
- [ ] Implement comprehensive deployment testing
- [ ] Add automated installation validation
- [ ] Create production deployment checklist
- [ ] Establish deployment quality gates

---

## Success Criteria

### **Technical Resolution:**
- [x] Service starts successfully without errors
- [x] WebSocket server binds to port 8002
- [x] Configuration loads without schema errors
- [x] All required directories exist with proper permissions

### **Process Improvements:**
- [ ] Installation script includes directory creation
- [ ] Integration tests include production-like environment
- [ ] Configuration validation tests implemented
- [ ] Pre-deployment validation script created

### **Documentation Updates:**
- [ ] Installation manual updated with directory creation steps
- [ ] Configuration guide updated with correct schema
- [ ] Troubleshooting guide expanded with common issues
- [ ] Deployment validation procedures documented

---

## Conclusion

This incident was successfully resolved with minimal service downtime. The root causes were identified and fixed, and the incident revealed important gaps in our testing and deployment processes that need to be addressed to prevent similar issues in the future.

**Key Takeaways:**
1. Need for production-like testing environment
2. Importance of configuration schema validation
3. Need for comprehensive installation testing
4. Value of pre-deployment validation scripts

The incident has been resolved and the service is now operational. All action items have been identified and prioritized for implementation.

---

**Incident Resolution:** ✅ COMPLETE  
**Service Status:** ✅ OPERATIONAL  
**Next Steps:** Implement action items to prevent recurrence 