# Go Migration Project Plan

**Version:** 1.1  
**Date:** 2025-01-15  
**Status:** Approved Migration Strategy with Remediation  
**Related Epic/Story:** Go Implementation Migration  

## Executive Summary

This document outlines the comprehensive migration strategy from Python to Go implementation of the MediaMTX Camera Service. The plan follows a progressive vertical slice approach with foundation-first development, ensuring low risk and quick path to success.

### Migration Goals
- **Performance**: 5x improvement in response time and throughput
- **Concurrency**: 10x improvement (100 → 1000+ connections)
- **Resource Usage**: 50% reduction in memory footprint
- **Compatibility**: 100% API compatibility with Python implementation
- **Risk Management**: Incremental delivery with clear validation gates

### Success Criteria
- All JSON-RPC methods return identical responses to Python system
- Performance targets met: <50ms status, <100ms control operations
- Memory usage <60MB base, <200MB with 10 cameras
- 1000+ concurrent WebSocket connections supported

---

## Epic/Story/Task Breakdown

### **EPIC E1: Foundation Infrastructure** 
**Goal**: Establish core Go infrastructure and configuration management  
**Duration**: 2-3 sprints  
**Control Gate**: All foundation modules must pass unit tests and IV&V validation  
**Dependencies**: None  
**Status**: ✅ **COMPLETED** - All foundation modules implemented and validated 

#### **Story S1.1: Configuration Management System**
**Tasks**:
- **T1.1.1**: Implement Viper-based configuration loader (Developer) - *reference Python config patterns*
- **T1.1.2**: Create YAML configuration schema validation (Developer) 
- **T1.1.3**: Implement environment variable binding (Developer)
- **T1.1.4**: Add hot-reload capability (Developer)
- **T1.1.5**: Create configuration unit tests (Developer)
- **T1.1.6**: IV&V validate configuration system (IV&V)
- **T1.1.7**: PM approve foundation completion (PM)

**Rules (MANDATORY)**: /docs/testing/testing-guide.md,  docs/developemnt/go-coding-sandards
**Control Point**: Configuration system must load all settings from Python equivalent, no vilation of rules  
**Status**: All configuration sections implemented and functional  
**Remediation**: 1 sprint allowed, must demonstrate functional equivalence  
**Evidence**: Configuration loading tests, schema validation tests  

#### **Story S1.2: Logging Infrastructure**
**Tasks**:
- **T1.2.1**: ✅ Implement logrus structured logging (Developer) - *reference Python logging behavior* - **COMPLETED**
- **T1.2.2**: ✅ Add correlation ID support (Developer) - **COMPLETED**
- **T1.2.3**: ✅ Create log rotation configuration (Developer) - **COMPLETED**
- **T1.2.4**: ✅ Implement log level management (Developer) - **COMPLETED**
- **T1.2.5**: ✅ Create logging unit tests (Developer) - **COMPLETED**
- **T1.2.6**: ✅ **INTEGRATION TASK**: Integrate with Configuration Management System (Developer) - *use config from Epic E1* - **COMPLETED**
- **T1.2.7**: ✅ IV&V validate logging system (IV&V) - **COMPLETED**
- **T1.2.8**: ✅ PM approve logging completion (PM) - **COMPLETED**

**Rules (MANDATORY)**: /docs/testing/testing-guide.md,  docs/developemnt/go-coding-sandards
**Control Point**: Logging must produce identical format to Python system, no rules violation  
**Status**: ✅ FULLY COMPLETED - All tasks implemented with comprehensive integration
**Remediation**: 1 sprint allowed, must demonstrate format compatibility  
**Evidence**: Log format comparison tests, correlation ID tests, complete implementation with configuration integration  

#### **Story S1.3: Security Framework**
**Tasks**:
- **T1.3.1**: ✅ Implement JWT authentication with golang-jwt/jwt/v4 (Developer) - *reference Python auth patterns* - **COMPLETED**
- **T1.3.2**: ✅ Add role-based access control (Developer) - **COMPLETED**
- **T1.3.3**: ✅ Implement session management (Developer) - **COMPLETED**
- **T1.3.4**: ✅ Create security unit tests (Developer) - **COMPLETED**
- **T1.3.5**: ✅ IV&V validate security implementation (IV&V) - **COMPLETED**
- **T1.3.6**: ✅ PM approve security completion (PM) - **COMPLETED**

**Control Point**: Authentication must be functionally equivalent to Python system  
**Status**: ✅ FULLY COMPLETED - All security components implemented with comprehensive testing
**Remediation**: 1 sprint allowed, must demonstrate security parity  
**Evidence**: Authentication tests, role-based access tests, comprehensive security test suite  

---

### **EPIC E2: Camera Discovery System**
**Goal**: Implement USB camera detection and monitoring with 5x performance improvement  
**Duration**: 2-3 sprints  
**Control Gate**: Camera discovery must detect devices in <200ms  
**Dependencies**: Epic E1 (Foundation Infrastructure)  
**Status**: ✅ **COMPLETED** - All performance targets exceeded (73.7ms vs 200ms requirement)  

#### **Story S2.1: V4L2 Camera Interface**
**Tasks**:
- **T2.1.1**: ✅ Implement V4L2 device enumeration (Developer) - **COMPLETED**
- **T2.1.2**: ✅ Add camera capability probing (Developer) - **COMPLETED**
- **T2.1.3**: ✅ Implement device status monitoring (Developer) - **COMPLETED**
- **T2.1.4**: ✅ Create camera interface unit tests (Developer) - **COMPLETED**
- **T2.1.5**: ✅ IV&V validate camera detection (IV&V) - **COMPLETED**
- **T2.1.6**: ✅ PM approve camera interface (PM) - **COMPLETED**

**Status**: ✅ FULLY COMPLETED - All performance targets exceeded (73.7ms vs 200ms requirement)

#### **Story S2.2: Camera Monitor Service**
**Tasks**:
- **T2.2.1**: Implement goroutine-based camera monitoring (Developer)
- **T2.2.2**: Add hot-plug event handling (Developer)
- **T2.2.3**: Create event notification system (Developer)
- **T2.2.4**: Implement concurrent monitoring (Developer)
- **T2.2.5**: Create monitor unit tests (Developer)
- **T2.2.6**: IV&V validate monitoring system (IV&V)
- **T2.2.7**: PM approve monitoring completion (PM)

**Control Point**: Must handle connect/disconnect events with <20ms notification
**Evidence**: Event handling tests, notification latency tests

---

### **EPIC E3: WebSocket JSON-RPC Server**
**Goal**: Implement high-performance WebSocket server with 1000+ concurrent connections  
**Duration**: 3-4 sprints  
**Control Gate**: Server must handle 1000+ connections with <50ms response time  
**Dependencies**: Epic E1, Epic E2

#### **Story S3.1: WebSocket Infrastructure**
**Tasks**:
- **T3.1.1**: Implement gorilla/websocket server (Developer)
- **T3.1.2**: Add connection management (Developer)
- **T3.1.3**: Implement JSON-RPC 2.0 protocol (Developer)
- **T3.1.4**: Add authentication middleware (Developer)
- **T3.1.5**: Create WebSocket unit tests (Developer)
- **T3.1.6**: IV&V validate WebSocket implementation (IV&V)
- **T3.1.7**: PM approve WebSocket completion (PM)

**Control Point**: Must handle 1000+ concurrent connections
**Evidence**: Connection stress tests, performance benchmarks  

#### **Story S3.2: Core JSON-RPC Methods**
**Tasks**:
- **T3.2.1**: Implement `ping` method (Developer)
- **T3.2.2**: Implement `authenticate` method (Developer)
- **T3.2.3**: Implement `get_camera_list` method (Developer)
- **T3.2.4**: Implement `get_camera_status` method (Developer)
- **T3.2.5**: Create method unit tests (Developer)
- **T3.2.6**: IV&V validate core methods (IV&V)
- **T3.2.7**: PM approve core methods (PM)

**Control Point**: All methods must return identical responses to Python system
**Evidence**: API compatibility tests, response format validation  

---

### **EPIC E4: MediaMTX Integration**
**Goal**: Implement MediaMTX path management with FFmpeg integration  
**Duration**: 3-4 sprints  
**Control Gate**: Path creation must complete in <100ms  
**Dependencies**: Epic E1, Epic E2, Epic E3

#### **Story S4.1: MediaMTX Controller**
**Tasks**:
- **T4.1.1**: Implement MediaMTX REST API client (Developer)
- **T4.1.2**: Add dynamic path creation (Developer)
- **T4.1.3**: Implement FFmpeg command generation (Developer)
- **T4.1.4**: Add path lifecycle management (Developer)
- **T4.1.5**: Create controller unit tests (Developer)
- **T4.1.6**: IV&V validate MediaMTX integration (IV&V)
- **T4.1.7**: PM approve MediaMTX completion (PM)

**Control Point**: Must create paths in <100ms with FFmpeg integration
**Evidence**: Path creation tests, FFmpeg integration tests  

#### **Story S4.2: Stream Management**
**Tasks**:
- **T4.2.1**: Implement stream URL generation (Developer)
- **T4.2.2**: Add stream status monitoring (Developer)
- **T4.2.3**: Implement stream cleanup (Developer)
- **T4.2.4**: Create stream unit tests (Developer)
- **T4.2.5**: IV&V validate stream management (IV&V)
- **T4.2.6**: PM approve stream completion (PM)
- **T4.2.7**: Implement `get_streams` method (Developer)

**Control Point**: Must provide identical stream URLs to Python system
**Evidence**: Stream URL tests, get_streams method tests

---

### **EPIC E4.5: MediaMTX Integration Remediation**
**Goal**: Address critical MediaMTX integration gaps identified in Python vs Go analysis  
**Duration**: 1 sprint  
**Control Gate**: Must achieve functional equivalence with Python MediaMTX integration  
**Dependencies**: Epic E4 (MediaMTX Integration) - FAILED VALIDATION  

#### **Story S4.5.1: Stream Lifecycle Management**
**Tasks**:
- **T4.5.1.1**: Extend `StreamManager` interface in `internal/mediamtx/types.go` (Developer) - *Add CreateStreamWithUseCase method*
- **T4.5.1.2**: Add use case types to `internal/mediamtx/types.go` (Developer) - *Add StreamUseCase const (Recording, Viewing, Snapshot)*
- **T4.5.1.3**: Extend `streamManager` struct in `internal/mediamtx/stream_manager.go` (Developer) - *Add useCaseConfigs map[StreamUseCase]UseCaseConfig field*
- **T4.5.1.4**: Add `CreateStreamWithUseCase` method to `internal/mediamtx/stream_manager.go` (Developer) - *Implement use case differentiation logic*
- **T4.5.1.5**: Extend `controller` struct in `internal/mediamtx/controller.go` (Developer) - *Add CreateStreamForRecording, CreateStreamForViewing, CreateStreamForSnapshot methods*
- **T4.5.1.6**: Add unit tests in `tests/unit/test_mediamtx_stream_lifecycle_test.go` (Developer) - *New file for use case testing*
- **T4.5.1.7**: IV&V validate stream lifecycle implementation (IV&V)
- **T4.5.1.8**: PM approve stream lifecycle completion (PM)

**Control Point**: Must implement identical stream lifecycle management to Python system
**Evidence**: Stream lifecycle tests, use case differentiation tests, power efficiency validation  

#### **Story S4.5.2: Advanced Recording Features**
**Tasks**:
- **T4.5.2.1**: Add `RecordingConfig` struct to `internal/mediamtx/types.go` (Developer) - *Add SegmentFormat, StorageWarnPercent, StorageBlockPercent fields*
- **T4.5.2.2**: Extend `RecordingManager` in `internal/mediamtx/recording_manager.go` (Developer) - *Add storage monitoring fields and methods*
- **T4.5.2.3**: Add `CheckStorageSpace` method to `internal/mediamtx/recording_manager.go` (Developer) - *Implement storage threshold checking*
- **T4.5.2.4**: Add `StartSegmentedRecording` method to `internal/mediamtx/recording_manager.go` (Developer) - *Implement segment-based file rotation*
- **T4.5.2.5**: Extend `ffmpegManager` in `internal/mediamtx/ffmpeg_manager.go` (Developer) - *Add segment format support with strftime*
- **T4.5.2.6**: Add `HandleFileRotation` method to `internal/mediamtx/recording_manager.go` (Developer) - *Implement recording continuity during rotation*
- **T4.5.2.7**: Add unit tests in `tests/unit/test_mediamtx_recording_lifecycle_test.go` (Developer) - *New file for recording features*
- **T4.5.2.8**: IV&V validate recording management (IV&V)
- **T4.5.2.9**: PM approve recording management completion (PM)

**Control Point**: Must provide identical recording capabilities to Python system
**Evidence**: Recording management tests, segment rotation tests, storage monitoring validation  

#### **Story S4.5.3: Health Monitoring Enhancement**
**Tasks**:
- **T4.5.3.1**: Add `HealthState` struct to `internal/mediamtx/types.go` (Developer) - *Add ConsecutiveFailures, CircuitBreakerActive, LastSuccessTime fields*
- **T4.5.3.2**: Extend `healthMonitor` in `internal/mediamtx/health_monitor.go` (Developer) - *Add persistent state tracking fields*
- **T4.5.3.3**: Add `PersistHealthState` method to `internal/mediamtx/health_monitor.go` (Developer) - *Implement state persistence across restarts*
- **T4.5.3.4**: Add `ConfigurableBackoff` method to `internal/mediamtx/health_monitor.go` (Developer) - *Implement exponential backoff with jitter*
- **T4.5.3.5**: Extend `GetStatus` method in `internal/mediamtx/health_monitor.go` (Developer) - *Return comprehensive health metrics*
- **T4.5.3.6**: Add unit tests in `tests/unit/test_mediamtx_health_monitoring_test.go` (Developer) - *New file for health monitoring*
- **T4.5.3.7**: IV&V validate health monitoring (IV&V)
- **T4.5.3.8**: PM approve health monitoring completion (PM)

**Control Point**: Must provide identical health monitoring to Python system
**Evidence**: Health monitoring tests, circuit breaker tests, persistence validation

---

## **INTEGRATION REQUIREMENTS**

### **Existing Components to Extend (DO NOT CREATE NEW):**
- **Extend**: `StreamManager` interface in `internal/mediamtx/types.go`
- **Extend**: `streamManager` struct in `internal/mediamtx/stream_manager.go`  
- **Extend**: `controller` struct in `internal/mediamtx/controller.go`
- **Extend**: `RecordingManager` in `internal/mediamtx/recording_manager.go`
- **Extend**: `healthMonitor` in `internal/mediamtx/health_monitor.go`
- **Extend**: `ffmpegManager` in `internal/mediamtx/ffmpeg_manager.go`

### **Configuration Integration:**
- **Use**: Existing config system from `internal/config/`
- **Extend**: `MediaMTXConfig` struct in config package
- **Integration**: Use existing `ConfigManager` from Epic E1

### **Logging Integration:**
- **Use**: Existing logging from `internal/logging/`
- **Pattern**: Use existing `*logrus.Logger` instances
- **Integration**: Use existing correlation ID patterns

---

## **COMPILATION SAFETY CHECKLIST**

### **Before Implementation:**
- [ ] Identify which existing interfaces to extend
- [ ] Identify which existing structs to modify
- [ ] Plan backwards compatibility for existing methods
- [ ] Ensure new methods fit existing interface patterns

### **During Implementation:**
- [ ] Extend existing interfaces, don't create new ones
- [ ] Add fields to existing structs, don't create parallel structs
- [ ] Use existing constructor patterns
- [ ] Follow existing error handling patterns

### **After Implementation:**
- [ ] All existing tests still pass
- [ ] New functionality integrates cleanly
- [ ] No duplicate interface definitions
- [ ] No orphaned struct definitions
---

### **EPIC E5: Camera Control Operations**
**Goal**: Implement snapshot and recording functionality  
**Duration**: 2-3 sprints  
**Control Gate**: All operations must complete in <100ms  
**Dependencies**: Epic E4.5

#### **Story S5.1: Snapshot System**
**Tasks**:
- **T5.1.1**: Add `takeSnapshot` method to `internal/websocket/methods.go` (Developer) - *Add JSON-RPC method handler for take_snapshot*
- **T5.1.2**: Extend `MediaMTXController` interface in `internal/mediamtx/types.go` (Developer) - *Add TakeAdvancedSnapshot(ctx, devicePath, outputPath, options) method*
- **T5.1.3**: Add `TakeAdvancedSnapshot` method to `controller` struct in `internal/mediamtx/controller.go` (Developer) - *Implement snapshot orchestration using existing snapshotManager*
- **T5.1.4**: Extend `snapshotManager` in `internal/mediamtx/snapshot_manager.go` (Developer) - *Add metadata extraction and file management*
- **T5.1.5**: Add `SnapshotOptions` struct to `internal/mediamtx/types.go` (Developer) - *Add Quality, Format, Resolution, Timestamp fields*
- **T5.1.6**: Integrate with existing camera system in `internal/camera/hybrid_monitor.go` (Developer) - *Use GetCameraInfo for device validation*
- **T5.1.7**: Add unit tests in `tests/unit/test_websocket_snapshot_methods_test.go` (Developer) - *New file for JSON-RPC snapshot testing*
- **T5.1.8**: Add unit tests in `tests/unit/test_mediamtx_snapshot_operations_test.go` (Developer) - *New file for snapshot operations testing*
- **T5.1.9**: IV&V validate snapshot system (IV&V)
- **T5.1.10**: PM approve snapshot completion (PM)

**Control Point**: Must produce identical snapshot files to Python system
**Evidence**: Snapshot file tests, metadata validation tests  

#### **Story S5.2: Recording System**
**Tasks**:
- **T5.2.1**: Add `startRecording` method to `internal/websocket/methods.go` (Developer) - *Add JSON-RPC method handler for start_recording*
- **T5.2.2**: Add `stopRecording` method to `internal/websocket/methods.go` (Developer) - *Add JSON-RPC method handler for stop_recording*
- **T5.2.3**: Extend `MediaMTXController` interface in `internal/mediamtx/types.go` (Developer) - *Add StartRecording, StopRecording, GetRecordingStatus methods*
- **T5.2.4**: Add recording methods to `controller` struct in `internal/mediamtx/controller.go` (Developer) - *Implement recording orchestration using existing recordingManager*
- **T5.2.5**: Extend `recordingManager` in `internal/mediamtx/recording_manager.go` (Developer) - *Add session tracking and file management using existing sessions map*
- **T5.2.6**: Add `RecordingOptions` struct to `internal/mediamtx/types.go` (Developer) - *Add Duration, Quality, FileRotation, SegmentSize fields*
- **T5.2.7**: Integrate with existing camera system in `internal/camera/hybrid_monitor.go` (Developer) - *Use GetCameraInfo for device validation*
- **T5.2.8**: Use existing security system in `internal/security/role_manager.go` (Developer) - *Validate recording permissions using existing roles*
- **T5.2.9**: Add unit tests in `tests/unit/test_websocket_recording_methods_test.go` (Developer) - *New file for JSON-RPC recording testing*
- **T5.2.10**: Add unit tests in `tests/unit/test_mediamtx_recording_operations_test.go` (Developer) - *New file for recording operations testing*
- **T5.2.11**: IV&V validate recording system (IV&V)
- **T5.2.12**: PM approve recording completion (PM)

**Control Point**: Must produce identical recording files to Python system
**Evidence**: Recording file tests, metadata validation tests, session management tests

---

## **INTEGRATION REQUIREMENTS**

### **Existing Components to Extend (DO NOT CREATE NEW):**
- **WebSocket**: Extend `internal/websocket/methods.go` for JSON-RPC method handlers
- **MediaMTX**: Extend `internal/mediamtx/controller.go` for camera control orchestration
- **MediaMTX**: Extend `internal/mediamtx/recording_manager.go` for recording operations
- **MediaMTX**: Extend `internal/mediamtx/snapshot_manager.go` for snapshot operations
- **MediaMTX**: Extend `internal/mediamtx/ffmpeg_manager.go` for FFmpeg operations
- **Types**: Extend `internal/mediamtx/types.go` for new data structures

### **Existing Systems Integration:**
- **Configuration**: Use existing `internal/config/config_manager.go` for snapshot/recording settings
- **Logging**: Use existing `internal/logging/logger.go` with correlation IDs  
- **Camera**: Use existing `internal/camera/hybrid_monitor.go` for device validation
- **Security**: Use existing `internal/security/role_manager.go` for operation authorization
- **WebSocket**: Use existing `internal/websocket/server.go` infrastructure

### **File Integration Patterns:**
- **Use**: Existing `controller.sessions` map in `internal/mediamtx/controller.go`
- **Use**: Existing `recordingManager` and `snapshotManager` fields in controller
- **Use**: Existing FFmpeg command patterns from `ffmpegManager`
- **Use**: Existing error types from `internal/mediamtx/errors.go`

---

## **COMPILATION SAFETY CHECKLIST**

### **Before Implementation:**
- [ ] Review existing JSON-RPC method patterns in `internal/websocket/methods.go`
- [ ] Check existing MediaMTX controller structure in `internal/mediamtx/controller.go`
- [ ] Verify existing manager interfaces in `internal/mediamtx/types.go`
- [ ] Confirm existing recording/snapshot manager implementation

### **During Implementation:**
- [ ] Add methods to existing WebSocket methods file, don't create new file
- [ ] Extend existing MediaMTX interfaces, don't create new interfaces
- [ ] Use existing manager instances in controller, don't create new managers
- [ ] Follow existing constructor and initialization patterns

### **After Implementation:**
- [ ] All existing JSON-RPC methods still work
- [ ] All existing MediaMTX operations still function  
- [ ] New methods follow existing naming conventions
- [ ] No interface conflicts or compilation errors  

---

### **EPIC E6: File Management System**
**Goal**: Implement file listing, metadata, and deletion operations  
**Duration**: 2 sprints  
**Control Gate**: All file operations must be functionally equivalent  
**Dependencies**: Epic E5

#### **Story S6.1: File Listing Operations**
**Tasks**:
- **T6.1.1**: Add `listRecordings` method to `internal/websocket/methods.go` (Developer) - *Add JSON-RPC method handler for list_recordings*
- **T6.1.2**: Add `listSnapshots` method to `internal/websocket/methods.go` (Developer) - *Add JSON-RPC method handler for list_snapshots*
- **T6.1.3**: Extend `MediaMTXController` interface in `internal/mediamtx/types.go` (Developer) - *Add ListRecordings, ListSnapshots methods*
- **T6.1.4**: Add file listing methods to `controller` struct in `internal/mediamtx/controller.go` (Developer) - *Add ListRecordings, ListSnapshots using existing recordingManager and snapshotManager*
- **T6.1.5**: Extend `recordingManager` in `internal/mediamtx/recording_manager.go` (Developer) - *Add GetRecordingsList with file scanning and metadata extraction*
- **T6.1.6**: Extend `snapshotManager` in `internal/mediamtx/snapshot_manager.go` (Developer) - *Add GetSnapshotsList with file scanning and metadata extraction*
- **T6.1.7**: Add `FileMetadata` struct to `internal/mediamtx/types.go` (Developer) - *Add FileName, FileSize, CreatedAt, Duration fields*
- **T6.1.8**: Add unit tests in `tests/unit/test_websocket_file_listing_methods_test.go` (Developer) - *New file for JSON-RPC file listing testing*
- **T6.1.9**: Add unit tests in `tests/unit/test_mediamtx_file_operations_test.go` (Developer) - *New file for file operations testing*
- **T6.1.10**: IV&V validate file listing (IV&V)
- **T6.1.11**: PM approve file listing (PM)

**Control Point**: Must return identical file lists to Python system
**Evidence**: File listing tests, metadata extraction tests  

#### **Story S6.2: File Lifecycle Management**
**Tasks**:
- **T6.2.1**: Add `getRecordingInfo` method to `internal/websocket/methods.go` (Developer) - *Add JSON-RPC method handler for get_recording_info*
- **T6.2.2**: Add `getSnapshotInfo` method to `internal/websocket/methods.go` (Developer) - *Add JSON-RPC method handler for get_snapshot_info*
- **T6.2.3**: Add `deleteRecording` method to `internal/websocket/methods.go` (Developer) - *Add JSON-RPC method handler for delete_recording*
- **T6.2.4**: Add `deleteSnapshot` method to `internal/websocket/methods.go` (Developer) - *Add JSON-RPC method handler for delete_snapshot*
- **T6.2.5**: Extend `MediaMTXController` interface in `internal/mediamtx/types.go` (Developer) - *Add GetRecordingInfo, GetSnapshotInfo, DeleteRecording, DeleteSnapshot methods*
- **T6.2.6**: Add file lifecycle methods to `controller` struct in `internal/mediamtx/controller.go` (Developer) - *Implement info and deletion methods using existing managers*
- **T6.2.7**: Add file operations to `recordingManager` in `internal/mediamtx/recording_manager.go` (Developer) - *Add GetRecordingInfo, DeleteRecording with session cleanup*
- **T6.2.8**: Add file operations to `snapshotManager` in `internal/mediamtx/snapshot_manager.go` (Developer) - *Add GetSnapshotInfo, DeleteSnapshot with metadata*
- **T6.2.9**: Use existing security system in `internal/security/role_manager.go` (Developer) - *Validate file operation permissions using existing roles*
- **T6.2.10**: Add unit tests in `tests/unit/test_websocket_file_lifecycle_methods_test.go` (Developer) - *New file for JSON-RPC file lifecycle testing*
- **T6.2.11**: Add unit tests in `tests/unit/test_mediamtx_file_lifecycle_test.go` (Developer) - *New file for file lifecycle operations testing*
- **T6.2.12**: IV&V validate file management (IV&V)
- **T6.2.13**: PM approve file management (PM)

**Control Point**: Must handle file operations identically to Python system
**Evidence**: File operation tests, deletion validation tests

---

## **INTEGRATION REQUIREMENTS**

### **Existing Components to Extend (DO NOT CREATE NEW):**
- **WebSocket**: Extend `internal/websocket/methods.go` for JSON-RPC method handlers
- **MediaMTX**: Extend `internal/mediamtx/controller.go` for file operation orchestration
- **MediaMTX**: Extend `internal/mediamtx/recording_manager.go` for recording file operations
- **MediaMTX**: Extend `internal/mediamtx/snapshot_manager.go` for snapshot file operations
- **Types**: Extend `internal/mediamtx/types.go` for file metadata structures

### **Existing Systems Integration:**
- **Configuration**: Use existing `internal/config/config_manager.go` for file storage settings
- **Logging**: Use existing `internal/logging/logger.go` with correlation IDs
- **Security**: Use existing `internal/security/role_manager.go` for file operation authorization
- **WebSocket**: Use existing `internal/websocket/server.go` infrastructure  

---

### **EPIC E7: System Management & Monitoring**
**Goal**: Implement system metrics, health monitoring, and observability  
**Duration**: 3 sprints  
**Control Gate**: All monitoring must provide identical data to Python system  
**Dependencies**: Epic E3, Epic E4.5

#### **Story S7.1: System Metrics**
**Tasks**:
- **T7.1.1**: Add `getMetrics` method to `internal/websocket/methods.go` (Developer) - *Add JSON-RPC method handler for get_metrics*
- **T7.1.2**: Extend `MediaMTXController` interface in `internal/mediamtx/types.go` (Developer) - *Add GetSystemMetrics() method*
- **T7.1.3**: Add `GetSystemMetrics` method to `controller` struct in `internal/mediamtx/controller.go` (Developer) - *Implement metrics collection using existing managers*
- **T7.1.4**: Add `SystemMetrics` struct to `internal/mediamtx/types.go` (Developer) - *Add RequestCount, ResponseTime, ErrorCount, ActiveConnections fields*
- **T7.1.5**: Extend `healthMonitor` in `internal/mediamtx/health_monitor.go` (Developer) - *Add performance metrics collection and resource tracking*
- **T7.1.6**: Use existing WebSocket server in `internal/websocket/server.go` (Developer) - *Integrate metrics collection in existing connection handler*
- **T7.1.7**: Use existing configuration system in `internal/config/config_manager.go` (Developer) - *Add metrics configuration settings*
- **T7.1.8**: Add unit tests in `tests/unit/test_websocket_metrics_methods_test.go` (Developer) - *New file for JSON-RPC metrics testing*
- **T7.1.9**: Add unit tests in `tests/unit/test_mediamtx_system_metrics_test.go` (Developer) - *New file for system metrics testing*
- **T7.1.10**: IV&V validate metrics system (IV&V)
- **T7.1.11**: PM approve metrics completion (PM)

**Control Point**: Must provide identical metrics to Python system
**Evidence**: Metrics comparison tests, performance tracking tests  

#### **Story S7.2: Health Monitoring**
**Tasks**:
- **T7.2.1**: Add `getStatus` method to `internal/websocket/methods.go` (Developer) - *Add JSON-RPC method handler for get_status*
- **T7.2.2**: Extend existing `GetStatus` method in `controller` struct in `internal/mediamtx/controller.go` (Developer) - *Enhance to return comprehensive health data*
- **T7.2.3**: Extend `HealthStatus` struct in `internal/mediamtx/types.go` (Developer) - *Add ComponentStatus, ErrorCount, LastCheck, CircuitBreakerState fields*
- **T7.2.4**: Add component health methods to `healthMonitor` in `internal/mediamtx/health_monitor.go` (Developer) - *Add CheckAllComponents, GetDetailedStatus methods*
- **T7.2.5**: Integrate with existing camera system in `internal/camera/hybrid_monitor.go` (Developer) - *Use existing GetCameraStatus for health reporting*
- **T7.2.6**: Use existing logging system in `internal/logging/logger.go` (Developer) - *Add health monitoring logs with correlation IDs*
- **T7.2.7**: Add unit tests in `tests/unit/test_websocket_status_methods_test.go` (Developer) - *New file for JSON-RPC status testing*
- **T7.2.8**: Add unit tests in `tests/unit/test_mediamtx_health_monitoring_test.go` (Developer) - *New file for health monitoring testing*
- **T7.2.9**: IV&V validate health system (IV&V)
- **T7.2.10**: PM approve health completion (PM)

**Control Point**: Must provide identical health data to Python system
**Evidence**: Health check tests, component status tests  

### **Existing Systems Integration:**
- **Configuration**: Use existing `internal/config/config_manager.go` for file storage settings
- **Logging**: Use existing `internal/logging/logger.go` with correlation IDs
- **Security**: Use existing `internal/security/role_manager.go` for file operation authorization
- **WebSocket**: Use existing `internal/websocket/server.go` infrastructure  

---

### **EPIC E8: Integration & Validation**
**Goal**: End-to-end integration testing and performance validation  
**Duration**: 2-3 sprints  
**Control Gate**: Complete functional equivalence with 5x performance improvement  
**Dependencies**: All previous epics  

#### **Story S8.1: Integration Testing**
**Tasks**:
- **T8.1.1**: Create `tests/integration/test_end_to_end_camera_operations.go` (Developer) - *New file for complete camera workflow testing*
- **T8.1.2**: Create `tests/integration/test_websocket_api_integration.go` (Developer) - *New file for WebSocket JSON-RPC API testing*
- **T8.1.3**: Create `tests/integration/test_mediamtx_integration.go` (Developer) - *New file for MediaMTX path/stream integration testing*
- **T8.1.4**: Add performance benchmarks in `tests/benchmarks/benchmark_api_performance_test.go` (Developer) - *New file for API performance benchmarking*
- **T8.1.5**: Add stress tests in `tests/stress/test_concurrent_connections.go` (Developer) - *New file for 1000+ connection testing*
- **T8.1.6**: Create compatibility validator in `tests/integration/test_python_api_compatibility.go` (Developer) - *New file for Python vs Go API response validation*
- **T8.1.7**: Add unit tests in `tests/unit/test_integration_validation_test.go` (Developer) - *New file for integration validation testing*
- **T8.1.8**: IV&V validate integration (IV&V)
- **T8.1.9**: PM approve integration completion (PM)

**Control Point**: Must pass all integration tests with performance targets
**Evidence**: Integration test results, performance benchmarks  

#### **Story S8.2: Documentation & Deployment**
**Tasks**:
- **T8.2.1**: Update existing `docs/api/json-rpc-methods.md` (Developer) - *Add new method documentation for camera operations and file management*
- **T8.2.2**: Update existing `docs/deployment/deployment-guide.md` (Developer) - *Add Go service deployment procedures*
- **T8.2.3**: Update existing `docs/operations/operational-procedures.md` (Developer) - *Add Go service operational procedures*
- **T8.2.4**: Create `docs/migration/python-to-go-migration-guide.md` (Developer) - *New file for migration procedures*
- **T8.2.5**: Update existing `README.md` (Developer) - *Update with Go implementation instructions*
- **T8.2.6**: Update existing `docker/Dockerfile` (Developer) - *Add Go service container configuration*
- **T8.2.7**: Update existing `scripts/build.sh` (Developer) - *Add Go build procedures*
- **T8.2.8**: IV&V validate documentation (IV&V)
- **T8.2.9**: PM approve final delivery (PM)

**Control Point**: Must have complete documentation and deployment procedures
**Evidence**: Documentation completeness, deployment validation

---

## **INTEGRATION REQUIREMENTS**

### **Existing Components to Extend (DO NOT CREATE NEW):**
- **WebSocket**: Extend `internal/websocket/methods.go` for all JSON-RPC method handlers
- **MediaMTX**: Extend `internal/mediamtx/controller.go` for all operation orchestration
- **MediaMTX**: Extend existing managers in `internal/mediamtx/` for specific operations
- **Types**: Extend `internal/mediamtx/types.go` for all new data structures
- **Tests**: Use existing test infrastructure and patterns
- **Documentation**: Update existing documentation files, don't create parallel docs

### **Existing Systems Integration:**
- **Configuration**: Use existing `internal/config/config_manager.go` for all settings
- **Logging**: Use existing `internal/logging/logger.go` with correlation IDs  
- **Camera**: Use existing `internal/camera/hybrid_monitor.go` for device operations
- **Security**: Use existing `internal/security/role_manager.go` for all authorization
- **WebSocket**: Use existing `internal/websocket/server.go` infrastructure  

---

## Control Point Rules

### **Go/No-Go Gates**
- **Foundation Epic (E1)**: Must complete before proceeding to functional epics
- **Core Epics (E2-E3)**: Must complete before proceeding to integration epics
- **Remediation Epic (E4.5)**: Must complete before proceeding to remaining epics
- **Integration Epic (E8)**: Final validation gate

### **Remediation Policy**
- **1 Sprint Remediation**: Allowed for most control points
- **2 Sprint Remediation**: Allowed for integration testing only
- **No Carry-Over**: Failed control points must be remediated before proceeding
- **PM Approval Required**: All remediation must be approved by PM

### **Role Responsibilities**
- **Developer**: Implementation, unit tests, evidence creation
- **IV&V**: Integration validation, quality gates, functional verification
- **PM**: Final approval, scope control, remediation decisions

---

## Risk Management

### **Technical Risks**
- **MediaMTX Integration Complexity**: Addressed by Epic E4.5 remediation sprint
- **Performance Targets**: Mitigated by incremental validation and benchmarking
- **API Compatibility**: Mitigated by comprehensive testing against Python system

### **Schedule Risks**
- **Foundation Dependencies**: Mitigated by clear dependency mapping
- **Integration Complexity**: Mitigated by progressive vertical slice approach
- **Remediation Buffer**: E4.5 provides 1 sprint buffer for critical gap resolution

### **Quality Risks**
- **Functional Equivalence**: Mitigated by comprehensive IV&V validation
- **Performance Regression**: Mitigated by continuous benchmarking

---

## Success Metrics

### **Performance Targets**
- **Response Time**: 5x improvement (500ms → 100ms)
- **Concurrency**: 10x improvement (100 → 1000+ connections)
- **Throughput**: 5x improvement (200 → 1000+ requests/second)
- **Memory Usage**: 50% reduction (80% → 60%)
- **CPU Usage**: 30% reduction (70% → 50%)

### **Quality Targets**
- **API Compatibility**: 100% functional equivalence
- **Test Coverage**: >90% unit test coverage
- **Documentation**: Complete API and deployment documentation
- **Performance**: All targets met in integration testing

### **Delivery Targets**
- **Timeline**: 13-17 sprints total (including remediation)
- **Risk Management**: No more than 2 remediation sprints per epic
- **Quality Gates**: All control points passed with IV&V validation

---

**Document Status**: Approved migration plan with remediation sprint  
**Last Updated**: 2025-01-15  
**Progress**: 
- Epic E1: ✅ COMPLETED (Foundation Infrastructure)
- Epic E2: ✅ COMPLETED (Camera Discovery System)  
- Ready for Epic E3: WebSocket JSON-RPC Server