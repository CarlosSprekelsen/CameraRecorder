# Recording Management Quick Reference

**Version:** 1.0  
**Date:** 2025-01-15  
**Audience:** Implementation Team, Testing Team, Client Team  
**Status:** 🚨 CRITICAL - GROUND TRUTH CHANGED  

---

## 🚨 CRITICAL CHANGES SUMMARY

### **What Changed:**
- **17 new requirements** added for recording management
- **Enhanced error codes** (-1006, -1008, -1010)
- **New API response fields** for recording status
- **Configurable parameters** via environment variables
- **No backwards compatibility** required (server not production-ready)

### **Why Changed:**
- **System protection** against resource exhaustion
- **Storage safety** with configurable thresholds
- **User experience** improvements with clear error messages
- **Data integrity** with file rotation and state management

---

## 📋 NEW ERROR CODES

| Code | Message | When | Data Fields |
|------|---------|------|-------------|
| **-1006** | "Camera is currently recording" | Device already recording | `camera_id`, `session_id` |
| **-1008** | "Storage space is low" | Below 10% available | `available_space`, `total_space` |
| **-1010** | "Storage space is critical" | Below 5% available | `available_space`, `total_space` |

---

## 🔧 CONFIGURATION VARIABLES

```bash
# Recording Management Configuration
RECORDING_ROTATION_MINUTES=30    # File rotation interval (default: 30)
STORAGE_WARN_PERCENT=80          # Warning threshold (default: 80%)
STORAGE_BLOCK_PERCENT=90         # Block new recordings (default: 90%)
```

---

## 📡 API RESPONSE CHANGES

### **Enhanced Camera Status:**
```json
{
  "jsonrpc": "2.0",
  "result": {
    "camera_id": "camera0",
    "device": "/dev/video0",
    "status": "connected",
    "recording": true,                    // NEW
    "recording_session": "uuid",          // NEW
    "current_file": "camera0_timestamp.mp4", // NEW
    "elapsed_time": 1800                  // NEW
  },
  "id": 5
}
```

### **New Error Response:**
```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -1006,
    "message": "Camera is currently recording",
    "data": {
      "camera_id": "camera0",
      "session_id": "550e8400-e29b-41d4-a716-446655440000"
    }
  },
  "id": 5
}
```

---

## 🧪 TESTING TEAM CHECKLIST

### **New Test Categories Required:**
- [ ] **Recording State Management Tests**
  - [ ] Test recording conflicts (error code -1006)
  - [ ] Test per-device recording limits
  - [ ] Test recording status in API responses

- [ ] **Storage Protection Tests**
  - [ ] Test storage validation (error codes -1008, -1010)
  - [ ] Test configurable thresholds
  - [ ] Test storage monitoring

- [ ] **File Rotation Tests**
  - [ ] Test configurable rotation intervals
  - [ ] Test timestamped file naming
  - [ ] Test recording continuity across rotations

### **Test Infrastructure Updates:**
- [ ] Add recording management test fixtures
- [ ] Add storage simulation utilities
- [ ] Update test expectations for new error codes
- [ ] Add recording state validation helpers

---

## 💻 CLIENT TEAM CHECKLIST

### **Error Handling Updates:**
- [ ] Handle error code -1006 (recording conflict)
- [ ] Handle error code -1008 (storage low)
- [ ] Handle error code -1010 (storage critical)
- [ ] Display user-friendly error messages

### **Recording State Management:**
- [ ] Track recording state per camera
- [ ] Display recording status in UI
- [ ] Show current file and elapsed time
- [ ] Handle recording session information

### **Storage Monitoring:**
- [ ] Integrate `get_storage_info` method
- [ ] Display storage warnings to users
- [ ] Block recording when storage critical
- [ ] Show storage usage indicators

### **User Experience:**
- [ ] Update camera list with recording status
- [ ] Add recording progress indicators
- [ ] Show file rotation information
- [ ] Provide clear user guidance

---

## 🏗️ IMPLEMENTATION TEAM CHECKLIST

### **Server Implementation:**
- [ ] Add recording state tracking per device
- [ ] Implement storage space validation
- [ ] Add configurable environment variables
- [ ] Implement file rotation management
- [ ] Add enhanced error responses
- [ ] Update camera status responses

### **MediaMTX Integration:**
- [ ] Add file rotation to FFmpeg commands
- [ ] Implement timestamped file naming
- [ ] Maintain recording continuity
- [ ] Add storage monitoring

### **API Updates:**
- [ ] Implement new error codes (-1006, -1008, -1010)
- [ ] Add recording status to camera responses
- [ ] Implement `get_storage_info` method
- [ ] Add user-friendly error messages

---

## 📊 REQUIREMENTS SUMMARY

### **Critical Requirements (8):**
- REQ-REC-001.1: Per-device recording limits
- REQ-REC-001.2: Error handling for conflicts
- REQ-REC-002.3: Recording continuity
- REQ-REC-003.1: Storage space validation
- REQ-REC-003.2: Configurable storage thresholds
- REQ-REC-003.3: Storage error handling
- REQ-REC-004.1: Storage monitoring
- REQ-REC-004.2: Storage information API

### **High Priority Requirements (9):**
- REQ-REC-001.3: Recording status integration
- REQ-REC-002.1: Configurable file rotation
- REQ-REC-002.2: Timestamped file naming
- REQ-REC-003.4: No auto-deletion policy
- REQ-REC-004.3: Health integration
- REQ-REC-005.1: User-friendly error messages
- REQ-REC-005.2: Recording progress information
- REQ-REC-005.3: Real-time notifications
- REQ-REC-006.1: Configuration management

---

## 🚀 IMMEDIATE ACTIONS

### **All Teams:**
1. **Review** `recording-management-requirements.md`
2. **Update** expectations for new error codes
3. **Implement** against new ground truth
4. **Test** new functionality thoroughly
5. **Validate** against requirements

### **Timeline:**
- **Week 1-2**: Core functionality implementation
- **Week 3-4**: Enhanced features and testing
- **Week 5-6**: Integration and validation

---

## 📞 SUPPORT

### **Documentation:**
- **Requirements**: `docs/requirements/recording-management-requirements.md`
- **API Reference**: `docs/api/json-rpc-methods.md`
- **Architecture**: `docs/architecture/overview.md`
- **Implementation Overview**: `docs/requirements/recording-management-implementation-overview.md`

### **Ground Truth:**
- All requirements documents are now the authoritative source
- No backwards compatibility required
- Implement against new specifications immediately

---

**This quick reference ensures all teams have the essential information needed to implement the new recording management requirements.** 🎯
