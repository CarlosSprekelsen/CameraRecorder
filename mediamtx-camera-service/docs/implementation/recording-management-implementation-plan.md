# Recording Management Implementation Plan

**Version:** 1.0  
**Date:** 2025-01-15  
**Team:** Implementation Team  
**Status:** 🚀 READY FOR IMPLEMENTATION  
**Related Documents:** `docs/requirements/recording-management-requirements.md`, `docs/api/json-rpc-methods.md`

---

## 🎯 Implementation Team Scope

### **What We Need to Implement:**
- **Server-side recording state management**
- **Storage protection and monitoring**
- **File rotation management**
- **Enhanced error handling**
- **API response updates**

### **What We DON'T Need to Worry About:**
- ❌ Client-side implementation (separate team)
- ❌ Test implementation (separate team)
- ❌ Backwards compatibility (server not production-ready)

---

## 📋 Implementation Requirements Summary

### **17 New Requirements to Implement:**
- **8 Critical**: System protection and core functionality
- **9 High Priority**: User experience and configuration

### **Key Changes:**
- **Enhanced error codes**: -1006, -1008, -1010
- **New API response fields**: Recording status in camera responses
- **Configurable parameters**: Environment variables for thresholds
- **File rotation**: Configurable intervals with timestamped naming

---

## 🏗️ Implementation Plan

### **Phase 1: Core Infrastructure (Week 1)**

#### **1.1 Recording State Management**
**Files to Modify:**
- `src/websocket_server/server.py`
- `src/camera_service/service_manager.py`

**Implementation Tasks:**

1. **Add recording state tracking**
   ```python
   # Add to WebSocketJsonRpcServer class
   self._active_recordings = {}  # device_path -> session_info
   ```

2. **Implement conflict detection in `_method_start_recording`**
   ```python
   # Check if device already recording
   if device_path in self._active_recordings:
       return self._create_error_response(
           request_id, -1006, "Camera is currently recording",
           {"camera_id": f"camera{device_id}", "session_id": existing_session_id}
       )
   ```

3. **Add recording status to camera responses**
   ```python
   # Update _method_get_camera_status
   result = {
       "camera_id": f"camera{device_id}",
       "device": device_path,
       "status": "connected",
       "recording": device_path in self._active_recordings,
       "recording_session": self._active_recordings.get(device_path, {}).get("session_id"),
       "current_file": self._active_recordings.get(device_path, {}).get("current_file"),
       "elapsed_time": self._active_recordings.get(device_path, {}).get("elapsed_time", 0)
   }
   ```

#### **1.2 Configuration Management**
**Files to Modify:**
- `src/camera_service/config.py`
- `src/websocket_server/server.py`

**Implementation Tasks:**

1. **Add environment variable configuration**
   ```python
   # Add to config.py
   RECORDING_ROTATION_MINUTES = int(os.getenv('RECORDING_ROTATION_MINUTES', '30'))
   STORAGE_WARN_PERCENT = int(os.getenv('STORAGE_WARN_PERCENT', '80'))
   STORAGE_BLOCK_PERCENT = int(os.getenv('STORAGE_BLOCK_PERCENT', '90'))
   ```

2. **Add configuration validation**
   ```python
   def validate_recording_config():
       if RECORDING_ROTATION_MINUTES < 1 or RECORDING_ROTATION_MINUTES > 1440:
           raise ValueError("RECORDING_ROTATION_MINUTES must be between 1 and 1440")
       if STORAGE_WARN_PERCENT >= STORAGE_BLOCK_PERCENT:
           raise ValueError("STORAGE_WARN_PERCENT must be less than STORAGE_BLOCK_PERCENT")
   ```

### **Phase 2: Storage Protection (Week 2)**

#### **2.1 Storage Space Validation**
**Files to Modify:**
- `src/websocket_server/server.py`
- `src/camera_service/service_manager.py`

**Implementation Tasks:**

1. **Add storage space checking**
   ```python
   def check_storage_space(self):
       """Check available storage space and return status"""
       try:
           statvfs = os.statvfs(self._storage_path)
           total_space = statvfs.f_frsize * statvfs.f_blocks
           available_space = statvfs.f_frsize * statvfs.f_bavail
           used_percent = ((total_space - available_space) / total_space) * 100
           
           return {
               "total_space": total_space,
               "available_space": available_space,
               "used_percent": used_percent,
               "warn_threshold": STORAGE_WARN_PERCENT,
               "block_threshold": STORAGE_BLOCK_PERCENT
           }
       except Exception as e:
           self._logger.error(f"Storage check failed: {e}")
           return None
   ```

2. **Integrate storage validation in recording start**
   ```python
   # Add to _method_start_recording before starting recording
   storage_info = self.check_storage_space()
   if storage_info:
       if storage_info["used_percent"] >= STORAGE_BLOCK_PERCENT:
           return self._create_error_response(
               request_id, -1010, "Storage space is critical",
               {"available_space": storage_info["available_space"], 
                "total_space": storage_info["total_space"]}
           )
       elif storage_info["used_percent"] >= STORAGE_WARN_PERCENT:
           self._logger.warning(f"Storage usage high: {storage_info['used_percent']}%")
   ```

#### **2.2 Storage Information API**
**Files to Modify:**
- `src/websocket_server/server.py`

**Implementation Tasks:**

1. **Implement `get_storage_info` method**
   ```python
   async def _method_get_storage_info(self, request_id: int, params: dict) -> dict:
       """Get storage information and status"""
       storage_info = self.check_storage_space()
       if not storage_info:
           return self._create_error_response(request_id, -1005, "Storage information unavailable")
       
       return {
           "jsonrpc": "2.0",
           "result": storage_info,
           "id": request_id
       }
   ```

### **Phase 3: File Rotation Management (Week 3)**

#### **3.1 MediaMTX Controller Updates**
**Files to Modify:**
- `src/mediamtx_wrapper/controller.py`
- `src/mediamtx_wrapper/path_manager.py`

**Implementation Tasks:**

1. **Update FFmpeg commands for file rotation**
   ```python
   # Modify in path_manager.py
   ffmpeg_command = (
       f"ffmpeg -f v4l2 -i {device_path} -c:v {codec} -profile:v {video_profile} "
       f"-level {video_level} -pix_fmt {pixel_format} -preset {preset} -b:v {bitrate} "
       f"-f segment -segment_time {RECORDING_ROTATION_MINUTES * 60} "
       f"-reset_timestamps 1 -c copy "
       f"-strftime 1 /path/to/recordings/camera{camera_id}_%Y-%m-%d_%H-%M-%S.mp4"
   )
   ```

2. **Add recording session tracking**
   ```python
   # Add to controller.py
   async def start_recording_with_rotation(self, stream_name: str, duration: int, format_type: str):
       session_id = str(uuid.uuid4())
       start_time = datetime.utcnow()
       
       # Store session information
       self._recording_sessions[stream_name] = {
           "session_id": session_id,
           "start_time": start_time,
           "current_file": f"camera{stream_name}_{start_time.strftime('%Y-%m-%d_%H-%M-%S')}.{format_type}",
           "elapsed_time": 0
       }
       
       return await self.start_recording(stream_name, duration, format_type)
   ```

#### **3.2 Recording Continuity**
**Implementation Tasks:**

1. **Maintain session across file rotations**
   ```python
   # Add session tracking to recording management
   def update_recording_session(self, stream_name: str, current_file: str, elapsed_time: int):
       if stream_name in self._recording_sessions:
           self._recording_sessions[stream_name].update({
               "current_file": current_file,
               "elapsed_time": elapsed_time
           })
   ```

### **Phase 4: Enhanced Error Handling (Week 4)**

#### **4.1 Error Code Implementation**
**Files to Modify:**
- `src/websocket_server/server.py`

**Implementation Tasks:**

1. **Add new error codes to error handling**
   ```python
   # Add to error handling constants
   ERROR_CAMERA_ALREADY_RECORDING = -1006
   ERROR_STORAGE_LOW = -1008
   ERROR_STORAGE_CRITICAL = -1010
   ```

2. **Update error response creation**
   ```python
   def _create_error_response(self, request_id: int, error_code: int, message: str, data: dict = None):
       error_response = {
           "jsonrpc": "2.0",
           "error": {
               "code": error_code,
               "message": message
           },
           "id": request_id
       }
       if data:
           error_response["error"]["data"] = data
       return error_response
   ```

#### **4.2 User-Friendly Error Messages**
**Implementation Tasks:**

1. **Implement user-friendly error messages**
   ```python
   # Error message mapping
   ERROR_MESSAGES = {
       -1006: "Camera is currently recording",
       -1008: "Storage space is low",
       -1010: "Storage space is critical"
   }
   ```

2. **Update error responses to use user-friendly messages**
   ```python
   # Use user-friendly messages instead of technical details
   message = ERROR_MESSAGES.get(error_code, "An error occurred")
   ```

---

## 🔧 Implementation Checklist

### **Week 1: Core Infrastructure**
- [ ] Add recording state tracking to WebSocket server
- [ ] Implement conflict detection in start_recording
- [ ] Add recording status to camera responses
- [ ] Add environment variable configuration
- [ ] Implement configuration validation

### **Week 2: Storage Protection**
- [ ] Implement storage space checking
- [ ] Add storage validation to recording start
- [ ] Implement get_storage_info method
- [ ] Add storage monitoring to health checks
- [ ] Test storage threshold enforcement

### **Week 3: File Rotation**
- [ ] Update FFmpeg commands for file rotation
- [ ] Implement recording session tracking
- [ ] Add timestamped file naming
- [ ] Maintain recording continuity
- [ ] Test file rotation functionality

### **Week 4: Error Handling**
- [ ] Add new error codes to constants
- [ ] Implement user-friendly error messages
- [ ] Update error response creation
- [ ] Test all error scenarios
- [ ] Validate error message consistency

---

## 🧪 Testing Strategy

### **Unit Testing:**
- Test recording state management functions
- Test storage space validation
- Test error code handling
- Test configuration validation

### **Integration Testing:**
- Test complete recording workflows
- Test storage protection scenarios
- Test file rotation functionality
- Test error handling end-to-end

### **Performance Testing:**
- Verify storage checking doesn't impact performance
- Test recording state tracking under load
- Validate error response times

---

## 📊 Success Criteria

### **Functional Requirements:**
- ✅ All 17 recording management requirements implemented
- ✅ Enhanced error codes (-1006, -1008, -1010) working
- ✅ Storage protection with configurable thresholds active
- ✅ File rotation with timestamped naming functional
- ✅ Recording state tracking per device working

### **Technical Requirements:**
- ✅ Configuration via environment variables
- ✅ User-friendly error messages
- ✅ Recording status in API responses
- ✅ Storage information API functional
- ✅ No performance degradation

### **Quality Requirements:**
- ✅ Comprehensive error handling
- ✅ Proper logging and monitoring
- ✅ Configuration validation
- ✅ Resource cleanup on errors

---

## 🚨 Risk Mitigation

### **Technical Risks:**
- **Risk**: MediaMTX integration complexity
  - **Mitigation**: Start with simple state tracking, add rotation incrementally
- **Risk**: Storage monitoring performance impact
  - **Mitigation**: Cache storage information, check only when needed

### **Implementation Risks:**
- **Risk**: Configuration complexity
  - **Mitigation**: Use sensible defaults, validate configuration on startup
- **Risk**: Error handling consistency
  - **Mitigation**: Centralize error response creation, use constants

---

## 📞 Support & Resources

### **Documentation:**
- **Requirements**: `docs/requirements/recording-management-requirements.md`
- **API Reference**: `docs/api/json-rpc-methods.md`
- **Architecture**: `docs/architecture/overview.md`

### **Implementation Guidelines:**
- Follow existing code patterns and style
- Add comprehensive logging for debugging
- Implement proper error handling and cleanup
- Test each component thoroughly before integration

---

**This plan focuses specifically on the implementation team's scope and provides clear, actionable tasks for implementing the recording management requirements.** 🎯

**Document Status:** Ready for implementation team use
**Next Review:** After Phase 1 completion
