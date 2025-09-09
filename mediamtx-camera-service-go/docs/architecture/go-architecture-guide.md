# Go Architecture Guide - MediaMTX Camera Service

**Version:** 1.0  
**Date:** 2025-01-15  
**Status:** Go Implementation Architecture  
**Related Epic/Story:** Go Implementation Architecture  

## Table of Contents

1. [Security Architecture](#security-architecture)
2. [System Overview](#system-overview)
3. [Component Architecture](#component-architecture)
4. [Core Architecture Patterns](#core-architecture-patterns)
5. [Testing Architecture](#testing-architecture)
6. [Implementation Guidelines](#implementation-guidelines)
7. [JSON-RPC API Contract](#json-rpc-api-contract)
8. [Performance Targets](#performance-targets)
9. [Technology Stack](#technology-stack)

---

## Security Architecture

**CRITICAL**: Security is the foundation of the entire system. All components must implement proper authentication, authorization, and security controls.

### Security-First Design Principles

1. **Zero Trust Architecture**: Every request must be authenticated and authorized
2. **Defense in Depth**: Multiple security layers at every component
3. **Least Privilege**: Users get minimum required permissions
4. **Audit Everything**: All security events must be logged
5. **Secure by Default**: All components start in secure state

### Security Middleware Design

The security layer is designed to integrate seamlessly with existing systems rather than creating parallel infrastructure:

```
┌────────────────────────────────────────────────────────────┐
│                    Security Layer                          │
├─────────────────────────────────────────────────────────────┤
│            Authentication Middleware                       │
│     • JWT token validation                                │
│     • Session management                                  │
│     • Uses existing SecurityConfig                        │
├─────────────────────────────────────────────────────────────┤
│            RBAC Middleware                                │
│     • Role-based access control                           │
│     • Permission matrix enforcement                       │
│     • Integrates with existing PermissionChecker          │
├─────────────────────────────────────────────────────────────┤
│            Rate Limiting                                  │
│     • Per-method rate limits                              │
│     • DDoS protection                                     │
│     • Uses existing SecurityConfig values                 │
├─────────────────────────────────────────────────────────────┤
│            Input Validation                               │
│     • Parameter sanitization                              │
│     • Type safety enforcement                             │
│     • Centralized validation logic                        │
├─────────────────────────────────────────────────────────────┤
│            Audit Logging                                  │
│     • Security event tracking                             │
│     • Uses existing LoggingConfig                         │
│     • File rotation and retention                         │
└─────────────────────┬──────────────────────────────────────┘
                      │ Configuration Integration
┌─────────────────────▼──────────────────────────────────────┐
│            Existing Configuration System                   │
│     • SecurityConfig for rate limits and JWT settings     │
│     • LoggingConfig for audit log configuration           │
│     • No hard-coded values or parallel infrastructure    │
└─────────────────────────────────────────────────────────────┘
```

### Security Integration Principles

1. **Leverage Existing Systems**: Use `SecurityConfig`, `LoggingConfig`, and existing logger
2. **No Hard-coded Values**: All security parameters come from configuration
3. **Transparent Integration**: Security middleware works seamlessly with existing code
4. **Configuration Adapter Pattern**: Bridge between security middleware and existing config
5. **Audit Trail**: Comprehensive logging of all security events

### Security Middleware Components

- **AuthMiddleware**: Centralized authentication enforcement
- **RBACMiddleware**: Role-based access control with existing permission matrix
- **EnhancedRateLimiter**: Rate limiting using existing configuration values
- **InputValidator**: Centralized input validation and sanitization
- **SecurityAuditLogger**: Comprehensive security event logging
- **ConfigAdapter**: Bridge between security middleware and existing configuration

### Role-Based Access Control (RBAC)

#### Role Definitions

- **viewer**: Read-only access to camera status, file listings, and basic information
- **operator**: Viewer permissions + camera control operations (snapshots, recording)
- **admin**: Full access to all features including system metrics and configuration

#### Permission Matrix

| Method | viewer | operator | admin |
|--------|--------|----------|-------|
| ping | ✅ | ✅ | ✅ |
| authenticate | ✅ | ✅ | ✅ |
| get_camera_list | ✅ | ✅ | ✅ |
| get_camera_status | ✅ | ✅ | ✅ |
| get_camera_capabilities | ✅ | ✅ | ✅ |
| take_snapshot | ❌ | ✅ | ✅ |
| start_recording | ❌ | ✅ | ✅ |
| stop_recording | ❌ | ✅ | ✅ |
| list_recordings | ✅ | ✅ | ✅ |
| list_snapshots | ✅ | ✅ | ✅ |
| delete_recording | ❌ | ✅ | ✅ |
| delete_snapshot | ❌ | ✅ | ✅ |
| get_metrics | ❌ | ❌ | ✅ |
| get_storage_info | ❌ | ❌ | ✅ |
| set_retention_policy | ❌ | ❌ | ✅ |
| cleanup_old_files | ❌ | ❌ | ✅ |

### Security Enforcement Points

#### WebSocket Server Security
- **Authentication**: Every WebSocket connection must authenticate
- **Authorization**: Every method call must be authorized
- **Rate Limiting**: Per-client rate limits to prevent abuse
- **Input Validation**: All parameters must be validated and sanitized

#### MediaMTX Controller Security
- **Device Access Control**: Only authorized devices can be accessed
- **File System Security**: Secure file operations with proper permissions
- **Process Security**: FFmpeg processes run with minimal privileges
- **Network Security**: Secure communication with MediaMTX server

### Security Compliance Checklist

Before implementing any changes, verify security compliance:

#### ✅ Authentication Compliance
- [ ] All WebSocket connections require authentication
- [ ] JWT tokens are properly validated
- [ ] Session management is secure
- [ ] Authentication failures are logged

#### ✅ Authorization Compliance
- [ ] RBAC is enforced for all methods
- [ ] Permission matrix is properly implemented
- [ ] Role escalation is prevented
- [ ] Authorization failures are logged

#### ✅ Input Validation Compliance
- [ ] All parameters are validated
- [ ] SQL injection prevention
- [ ] Path traversal prevention
- [ ] XSS prevention

#### ✅ Audit Logging Compliance
- [ ] All security events are logged
- [ ] Logs include user context
- [ ] Logs are tamper-proof
- [ ] Log retention policies are enforced

---

## System Overview

The MediaMTX Camera Service Go implementation is a high-performance wrapper around MediaMTX, providing:

1. **Real-time USB camera discovery and monitoring** (5x faster detection)
2. **WebSocket JSON-RPC 2.0 API** (1000+ concurrent connections)
3. **Dynamic MediaMTX configuration management** (100ms response time)
4. **Streaming, recording, and snapshot coordination** (5x throughput improvement)
5. **Resilient error recovery and health monitoring** (50% resource reduction)
6. **Secure access control and authentication** (Go crypto libraries)

### System Goals
- **Performance**: High-performance camera service with real-time capabilities
- **Resource Usage**: Efficient memory and CPU utilization
- **Compatibility**: Standards-compliant API with broad client support
- **Risk Management**: Working software first, integration incrementally

### Success Criteria
- Camera detection <200ms latency
- WebSocket server handles 1000+ concurrent connections
- Memory usage <60MB base, <200MB with 10 cameras
- 1000+ concurrent WebSocket connections supported
- **Working Service**: Fully functional camera service
- **Basic Integration**: Added incrementally when platform systems exist


## Component Architecture

```
┌────────────────────────────────────────────────────────────┐
│                    Client Applications                      │
│            (Web browsers, mobile apps, etc.)               │
│  • Use camera identifiers (camera0, camera1)               │
│  • Hardware-independent interface                          │
└─────────────────────┬──────────────────────────────────────┘
                      │ WebSocket JSON-RPC 2.0
┌─────────────────────▼──────────────────────────────────────┐
│            WebSocket JSON-RPC Server (gorilla/websocket)  │
│                 (THIN PROTOCOL LAYER)                     │
│     • Client connection management (1000+ concurrent)     │
│     • JSON-RPC 2.0 protocol handling                      │
│     • Real-time notifications (<20ms latency)             │
│     • Authentication and authorization (golang-jwt/jwt/v4) │
│     • Security middleware with RBAC enforcement            │
│     • Rate limiting and DDoS protection                   │
│     • NO business logic - delegates to MediaMTX           │
└─────────────────────┬──────────────────────────────────────┘
                      │ Delegates to
┌─────────────────────▼──────────────────────────────────────┐
│                MediaMTX Controller                         │
│              (COMPLETE BUSINESS LOGIC LAYER)              │
│     • Camera discovery integration                         │
│     • API abstraction layer (camera0 ↔ /dev/video0)       │
│     • All camera operations (recording, snapshots, etc.)  │
│     • Stream management and lifecycle                     │
│     • File operations and storage management              │
│     • Event management and notifications                  │
│     • Single source of truth for all camera operations   │
├─────────────────────────────────────────────────────────────┤
│             Camera Discovery Monitor (goroutines)         │
│     • USB camera detection (<200ms)                       │
│     • Camera status tracking                              │
│     • Hot-plug event handling                             │
│     • Concurrent monitoring with channels                 │
│     • Internal device path management (/dev/video*)       │
├─────────────────────────────────────────────────────────────┤
│            MediaMTX Path Manager (net/http)               │
│     • Dynamic path creation via REST API                  │
│     • FFmpeg command generation                           │
│     • Path lifecycle management                           │
│     • Error handling and recovery                         │
│     • Internal device path operations                     │
├─────────────────────────────────────────────────────────────┤
│               Health & Monitoring (logrus)                │
│     • Service health checks                               │
│     • Resource usage monitoring                           │
│     • Error tracking and recovery                         │
│     • Configuration management (viper)                    │
└─────────────────────┬───────────────────────────────────────┘
                      │ HTTP REST API
┌─────────────────────▼───────────────────────────────────────┐
│                   MediaMTX Server (Go)                      │
├─────────────────────────────────────────────────────────────┤
│                Media Processing                           │
│     • RTSP/WebRTC/HLS streaming                           │
│     • FFmpeg process management                           │
│     • Multi-protocol support                              │
│     • Recording and snapshot generation                   │
│     • Internal device path operations                     │
└─────────────────────┬───────────────────────────────────────┘
                      │ FFmpeg Processes
┌─────────────────────▼───────────────────────────────────────┐
│                 USB Cameras                                 │
│         /dev/video0, /dev/video1, etc.                     │
│  • Hardware layer (internal only)                          │
│  • Not exposed to clients                                  │
└─────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

#### WebSocket JSON-RPC Server (gorilla/websocket) - THIN PROTOCOL LAYER
- **SECURITY FIRST**: Authentication, authorization, and input validation for every request
- **ONLY** client connection management and authentication (1000+ concurrent)
- **ONLY** JSON-RPC 2.0 protocol implementation
- **ONLY** real-time event notifications (<20ms latency)
- **ONLY** session management and authorization
- **ONLY** rate limiting and DDoS protection
- **NO business logic** - all operations delegate to MediaMTX Controller
- **NO camera operations** - MediaMTX Controller handles all camera logic
- **NO file operations** - MediaMTX Controller handles all file logic
- **NO stream operations** - MediaMTX Controller handles all stream logic

#### MediaMTX Controller - COMPLETE BUSINESS LOGIC LAYER
- **Single source of truth** for all camera operations
- Camera discovery integration and management
- API abstraction layer (camera0 ↔ /dev/video0 mapping)
- All camera operations (recording, snapshots, streaming)
- Stream management and lifecycle coordination
- File operations and storage management
- Event management and notifications
- Orchestrates all sub-components (Camera Monitor, Path Manager, etc.)

#### Camera Discovery Monitor (goroutines) - HARDWARE ABSTRACTION LAYER
- USB camera detection via V4L2 (<200ms)
- Hot-plug event handling with channels
- Concurrent monitoring of multiple cameras
- Camera capability probing and status tracking
- **Internal to MediaMTX Controller** - not directly accessible

#### MediaMTX Path Manager (net/http) - STREAM INFRASTRUCTURE LAYER
- Dynamic path creation via MediaMTX REST API
- FFmpeg command generation and management
- Path lifecycle management (create, update, delete)
- Error handling and automatic recovery
- **Internal to MediaMTX Controller** - not directly accessible

#### Health & Monitoring (logrus) - OBSERVABILITY LAYER
- Structured logging with correlation IDs
- Service health monitoring and reporting
- Resource usage tracking and alerts
- Configuration management with hot-reload

---

## Core Architecture Patterns

### 1. Single Source of Truth Architecture

The system implements a **single source of truth** pattern where MediaMTX Controller is the complete business logic layer, with WebSocket server being a thin protocol layer that delegates all operations.

#### Architectural Layers

```
┌─────────────────────────────────────────────────────────────┐
│                    CLIENT LAYER                             │
│  • Works with camera identifiers (camera0, camera1)         │
│  • No knowledge of internal device paths                    │
│  • Clean, abstract API interface                           │
│  • Hardware-independent client code                        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                WEBSOCKET PROTOCOL LAYER                     │
│  • JSON-RPC 2.0 protocol handling                          │
│  • Authentication and authorization                         │
│  • Request/response formatting                             │
│  • NO business logic - delegates to MediaMTX               │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                MEDIAMTX CONTROLLER                          │
│              (SINGLE SOURCE OF TRUTH)                      │
│  • API abstraction layer (camera0 ↔ /dev/video0)          │
│  • All camera operations                                   │
│  • All business logic                                      │
│  • Orchestrates all sub-components                         │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                  HARDWARE LAYER                             │
│  • Camera discovery monitor                                │
│  • MediaMTX path manager                                   │
│  • Device path operations (/dev/video0, /dev/video1)      │
│  • Hardware-specific operations                            │
└─────────────────────────────────────────────────────────────┘
```

#### Implementation Details

**WebSocket Server - Thin Protocol Layer**
```go
// WebSocket Server: ONLY protocol handling
func (s *WebSocketServer) MethodTakeSnapshot(params map[string]interface{}, client *ClientConnection) (*JsonRpcResponse, error) {
    cameraID := params["device"].(string) // "camera0"
    
    // Simple delegation to MediaMTX Controller
    result, err := s.mediaMTXController.TakeSnapshot(context.Background(), cameraID, options)
    if err != nil {
        return s.createErrorResponse(request.ID, err)
    }
    
    return &JsonRpcResponse{Result: result, ID: request.ID}, nil
}
```

**MediaMTX Controller - Complete Business Logic**
```go
// MediaMTX Controller: ALL business logic
func (c *controller) TakeSnapshot(ctx context.Context, cameraID, options string) (*Snapshot, error) {
    // Validate camera identifier format
    if !c.validateCameraIdentifier(cameraID) {
        return nil, fmt.Errorf("invalid camera identifier: %s", cameraID)
    }
    
    // Map to internal device path
    devicePath := c.getDevicePathFromCameraIdentifier(cameraID) // "/dev/video0"
    
    // Validate camera exists
    if !c.cameraMonitor.DeviceExists(devicePath) {
        return nil, fmt.Errorf("camera device %s not found", cameraID)
    }
    
    // Perform snapshot operation
    return c.snapshotManager.TakeSnapshot(ctx, devicePath, options)
}
```

### 2. Strict Separation of Concerns

The architecture enforces strict separation of concerns with clear boundaries and responsibilities.

#### WebSocket Server Dependencies

```go
// WebSocket Server: ONLY depends on MediaMTX Controller
type WebSocketServer struct {
    mediaMTXController MediaMTXController  // ONLY dependency
    jwtHandler         *security.JWTHandler
    eventManager       *EventManager
    // NO camera monitor, NO business logic components
}

// Constructor enforces single dependency
func NewWebSocketServer(
    configManager *config.ConfigManager,
    logger *logging.Logger,
    mediaMTXController MediaMTXController,  // ONLY business logic dependency
    jwtHandler *security.JWTHandler,
) (*WebSocketServer, error)
```

#### MediaMTX Controller Dependencies

```go
// MediaMTX Controller: Orchestrates all business logic components
type MediaMTXController interface {
    // Camera discovery operations
    GetCameraList(ctx context.Context) (*CameraListResponse, error)
    GetCameraStatus(ctx context.Context, device string) (*CameraStatusResponse, error)
    ValidateCameraDevice(ctx context.Context, device string) (bool, error)
    
    // All camera operations
    TakeSnapshot(ctx context.Context, device, options string) (*Snapshot, error)
    StartRecording(ctx context.Context, device, path string) (*RecordingSession, error)
    StopRecording(ctx context.Context, sessionID string) error
    
    // All other operations...
}

// Internal implementation has access to all sub-components
type controller struct {
    cameraMonitor      camera.CameraMonitor      // Internal only
    pathManager        PathManager               // Internal only
    streamManager      StreamManager             // Internal only
    recordingManager   *RecordingManager         // Internal only
    snapshotManager    *SnapshotManager          // Internal only
    // All business logic components
}
```

#### Dependency Injection in Main

```go
func main() {
    // Create camera monitor (hardware layer)
    cameraMonitor := camera.NewHybridCameraMonitor(...)
    
    // Create MediaMTX controller with camera monitor integration
    mediaMTXController, err := mediamtx.ControllerWithConfigManager(
        configManager, 
        cameraMonitor,  // Camera monitor integrated into MediaMTX
        logger,
    )
    
    // Create WebSocket server with ONLY MediaMTX dependency
    wsServer, err := websocket.NewWebSocketServer(
        configManager,
        logger,
        mediaMTXController,  // ONLY business logic dependency
        jwtHandler,
    )
    
    // Clean architecture: WebSocket → MediaMTX → Camera Monitor
    // No circular dependencies, clear separation of concerns
}
```

### 3. Architecture Enforcement Rules

The architecture document enforces strict rules to prevent the current broken implementation:

#### **RULE 1: WebSocket Server MUST Be Thin**
- **FORBIDDEN**: Direct camera monitor access (`s.cameraMonitor.GetConnectedCameras()`)
- **FORBIDDEN**: Business logic in WebSocket methods
- **FORBIDDEN**: Camera validation in WebSocket layer
- **FORBIDDEN**: File operations in WebSocket layer
- **REQUIRED**: All operations delegate to MediaMTX Controller

#### **RULE 2: MediaMTX Controller MUST Be Complete**
- **REQUIRED**: Camera monitor integration
- **REQUIRED**: All camera discovery methods
- **REQUIRED**: All business logic operations
- **REQUIRED**: Single source of truth for camera operations
- **REQUIRED**: Proper abstraction layer (camera0 ↔ /dev/video0)

#### **RULE 3: No Direct Hardware Access**
- **FORBIDDEN**: WebSocket server accessing camera monitor directly
- **FORBIDDEN**: WebSocket server accessing file system directly
- **FORBIDDEN**: WebSocket server accessing MediaMTX server directly
- **REQUIRED**: All hardware access through MediaMTX Controller

#### **RULE 4: Clear Dependency Chain**
```
WebSocket Server → MediaMTX Controller → Camera Monitor
                → MediaMTX Controller → Path Manager
                → MediaMTX Controller → File Manager
```

#### **RULE 5: No Duplicated Logic**
- **FORBIDDEN**: Abstraction layer in both WebSocket and MediaMTX
- **FORBIDDEN**: Camera validation in multiple places
- **FORBIDDEN**: Stream URL generation in multiple places
- **REQUIRED**: Single implementation in MediaMTX Controller

### 4. Event-Driven Architecture

The event system is managed by MediaMTX Controller, not WebSocket server.

#### Event System Components

**EventManager (Central Hub)**
```go
type EventManager struct {
    subscriptions      map[string]*EventSubscription
    topicSubscriptions map[EventTopic]map[string]*EventSubscription
    eventHandlers      map[EventTopic][]func(*EventMessage) error
    mu                 sync.RWMutex
    logger             *logrus.Logger
}

// High-performance event delivery
func (em *EventManager) PublishEvent(topic EventTopic, data map[string]interface{}) error {
    // Only send to interested clients
    subscribers := em.GetSubscribersForTopic(topic)
    for _, clientID := range subscribers {
        // Deliver event to specific client
        em.deliverEventToClient(clientID, topic, data)
    }
    return nil
}
```

**Event Topics and Filtering**
```go
const (
    // Camera events
    TopicCameraConnected    EventTopic = "camera.connected"
    TopicCameraDisconnected EventTopic = "camera.disconnected"
    TopicCameraStatusChange EventTopic = "camera.status_change"
    
    // Recording events
    TopicRecordingStart EventTopic = "recording.start"
    TopicRecordingStop  EventTopic = "recording.stop"
    
    // System events
    TopicSystemHealth  EventTopic = "system.health"
    TopicSystemStartup EventTopic = "system.startup"
)

// Client subscription with filters
subscription := &EventSubscription{
    ClientID: "client1",
    Topics:   []EventTopic{TopicCameraConnected, TopicRecordingStart},
    Filters: map[string]interface{}{
        "device": "/dev/video0", // Only interested in specific device
    },
}
```

#### Performance Characteristics

**Before (Broadcast System)**
- **Network Traffic**: Events sent to ALL clients regardless of interest
- **Processing**: Every client processes every event
- **Scalability**: Linear degradation with client count
- **Performance**: O(n) where n = total clients

**After (Topic-Based System)**
- **Network Traffic**: Events sent only to interested clients
- **Processing**: Clients only process relevant events
- **Scalability**: Logarithmic scaling with client count
- **Performance**: O(log n) where n = interested clients
- **Improvement**: 100x+ faster event delivery

#### Event Integration Layer

**Component Adapters**
```go
// Camera Event Notifier
type CameraEventNotifier struct {
    eventManager *EventManager
    logger       *logrus.Logger
}

func (n *CameraEventNotifier) NotifyCameraConnected(device *camera.CameraDevice) {
    eventData := map[string]interface{}{
        "device":    device.Path,
        "name":      device.Name,
        "status":    string(device.Status),
        "timestamp": time.Now().Format(time.RFC3339),
    }
    
    // Publish to event system
    n.eventManager.PublishEvent(TopicCameraConnected, eventData)
}

// MediaMTX Event Notifier
type MediaMTXEventNotifier struct {
    eventManager *EventManager
    logger       *logrus.Logger
}

func (n *MediaMTXEventNotifier) NotifyRecordingStarted(device, sessionID, filename string) {
    eventData := map[string]interface{}{
        "device":     device,
        "session_id": sessionID,
        "filename":   filename,
        "timestamp":  time.Now().Format(time.RFC3339),
    }
    
    n.eventManager.PublishEvent(TopicMediaMTXRecordingStarted, eventData)
}
```

### 4. Stream Lifecycle Management

Stream lifecycle management ensures reliable recording operations while maintaining power efficiency through on-demand activation.

#### Stream Lifecycle Types

**Recording Streams**
- **Purpose**: Long-duration video recording with file rotation
- **Lifecycle**: Manual start/stop, no auto-close
- **MediaMTX Settings**:
  ```yaml
  runOnDemandCloseAfter: 0s  # Never auto-close
  runOnDemandRestart: yes
  runOnDemandStartTimeout: 10s
  ```

**Viewing Streams**
- **Purpose**: Live stream viewing for monitoring
- **Lifecycle**: Auto-close after inactivity
- **MediaMTX Settings**:
  ```yaml
  runOnDemandCloseAfter: 300s  # 5 minutes after last viewer
  runOnDemandRestart: yes
  runOnDemandStartTimeout: 10s
  ```

**Snapshot Streams**
- **Purpose**: Quick photo capture
- **Lifecycle**: Immediate activation/deactivation
- **MediaMTX Settings**:
  ```yaml
  runOnDemandCloseAfter: 60s  # 1 minute after capture
  runOnDemandRestart: no
  runOnDemandStartTimeout: 5s
  ```

#### Go Implementation Pattern

```go
type StreamLifecycleManager struct {
    config     *config.ConfigManager
    logger     *logging.Logger
    mediamtx   *MediaMTXPathManager
}

func (slm *StreamLifecycleManager) CreateRecordingStream(devicePath string) error {
    // Create MediaMTX path with recording lifecycle settings
    pathConfig := &MediaMTXPathConfig{
        Name:                    fmt.Sprintf("camera_%s", devicePath),
        RunOnDemand:            slm.buildFFmpegCommand(devicePath),
        RunOnDemandCloseAfter:  "0s",  // Never auto-close
        RunOnDemandRestart:     true,
        RunOnDemandStartTimeout: "10s",
    }
    
    return slm.mediamtx.CreatePath(pathConfig)
}
```

### 2. On-Demand Stream Activation

On-demand stream activation optimizes power efficiency by only starting FFmpeg processes when needed.

#### How It Works

1. **Camera Detection Phase**
   - MediaMTX path created with `runOnDemand` configuration
   - No FFmpeg process started immediately
   - Path configured but inactive (`ready: false`, `source: null`)

2. **On-Demand Activation Phase**
   - First access triggers FFmpeg process start via `runOnDemand`
   - FFmpeg captures from camera and publishes to MediaMTX
   - Stream becomes active (`ready: true`, `source: {...}`)

#### Go Implementation Pattern

```go
type OnDemandStreamManager struct {
    mediamtx *MediaMTXPathManager
    logger   *logging.Logger
}

func (odsm *OnDemandStreamManager) ActivateStream(streamName string) error {
    // Check if stream is already active
    status, err := odsm.mediamtx.GetPathStatus(streamName)
    if err != nil {
        return fmt.Errorf("failed to get stream status: %w", err)
    }
    
    if status.Ready {
        odsm.logger.Info("Stream already active", "stream", streamName)
        return nil
    }
    
    // Trigger on-demand activation
    odsm.logger.Info("Activating stream on-demand", "stream", streamName)
    return odsm.mediamtx.TriggerOnDemand(streamName)
}
```

### 3. Multi-Tier Snapshot Capture

Multi-tier snapshot capture provides optimal user experience while maintaining power efficiency.

#### Tier 1: Immediate RTSP Capture (Fastest Path)
- **Response Time**: < 0.5 seconds
- **Use Case**: Stream already running
- **Process**: Quick RTSP readiness check and immediate capture

#### Tier 2: Quick Stream Activation (Balanced Path)
- **Response Time**: 1-3 seconds
- **Use Case**: First snapshot after idle period
- **Process**: Trigger on-demand activation and capture

#### Tier 3: Direct Camera Capture (Fallback Path)
- **Response Time**: 2-5 seconds
- **Use Case**: MediaMTX issues, emergency capture
- **Process**: Bypass MediaMTX entirely

#### Go Implementation Pattern

```go
type MultiTierSnapshotManager struct {
    mediamtx *MediaMTXPathManager
    logger   *logging.Logger
    config   *config.ConfigManager
}

func (mtsm *MultiTierSnapshotManager) TakeSnapshot(devicePath string) (*SnapshotResult, error) {
    // Tier 1: Try immediate RTSP capture
    if result, err := mtsm.tier1ImmediateCapture(devicePath); err == nil {
        mtsm.logger.Info("Tier 1 capture successful", "device", devicePath)
        return result, nil
    }
    
    // Tier 2: Try quick stream activation
    if result, err := mtsm.tier2StreamActivation(devicePath); err == nil {
        mtsm.logger.Info("Tier 2 capture successful", "device", devicePath)
        return result, nil
    }
    
    // Tier 3: Fallback to direct camera capture
    mtsm.logger.Info("Using Tier 3 direct capture", "device", devicePath)
    return mtsm.tier3DirectCapture(devicePath)
}
```

### 4. Codec Compatibility (H.264 STANAG 4406)

Ensure H.264 streams are compatible with STANAG 4406 requirements.

#### STANAG 4406 Requirements
- **Profile**: Constrained Baseline Profile (CBP) or Baseline Profile
- **Level**: 3.0 or lower for compatibility
- **Pixel Format**: 4:2:0 (yuv420p)
- **Bitrate**: Variable, typically 64kbps to 2Mbps
- **Resolution**: Up to 720p (1280x720) for Level 3.0

#### Go Implementation Pattern

```go
type CodecManager struct {
    config *config.ConfigManager
    logger *logging.Logger
}

func (cm *CodecManager) BuildSTANAG4406Command(devicePath, outputURL string) string {
    return fmt.Sprintf(
        "ffmpeg -f v4l2 -i %s -c:v libx264 -profile:v baseline -level 3.0 "+
        "-pix_fmt yuv420p -preset ultrafast -b:v 600k -f rtsp %s",
        devicePath, outputURL,
    )
}

func (cm *CodecManager) ValidateCodecCompatibility(streamConfig *StreamConfig) error {
    if streamConfig.Codec != "h264" {
        return fmt.Errorf("unsupported codec: %s, only H.264 supported", streamConfig.Codec)
    }
    
    if streamConfig.Profile != "baseline" && streamConfig.Profile != "constrained_baseline" {
        return fmt.Errorf("unsupported profile: %s, STANAG 4406 requires baseline profile", streamConfig.Profile)
    }
    
    return nil
}
```

---

## Testing Architecture

### Single Systemd-Managed MediaMTX Instance

**Decision**: All tests MUST use the single systemd-managed MediaMTX service instance.

#### Service Configuration
```bash
# MediaMTX service configuration
sudo systemctl start mediamtx
sudo systemctl enable mediamtx
sudo systemctl status mediamtx
```

#### Go Test Integration
```go
type RealMediaMTXServer struct {
    client *http.Client
    logger *logging.Logger
}

func (rms *RealMediaMTXServer) Start() error {
    // Check if MediaMTX service is running via systemd
    cmd := exec.Command("systemctl", "is-active", "mediamtx")
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("MediaMTX systemd service is not running: %w", err)
    }
    
    // Wait for MediaMTX API to be ready
    return rms.waitForMediaMTXReady()
}

func (rms *RealMediaMTXServer) waitForMediaMTXReady() error {
    healthURL := "http://127.0.0.1:9997/v3/config/global/get"
    
    for i := 0; i < 30; i++ {
        resp, err := rms.client.Get(healthURL)
        if err == nil && resp.StatusCode == 200 {
            return nil
        }
        time.Sleep(1 * time.Second)
    }
    
    return fmt.Errorf("MediaMTX API not ready after 30 seconds")
}
```

#### Port Configuration
- **API Port**: 9997 (fixed systemd service port)
- **RTSP Port**: 8554 (fixed systemd service port)
- **WebRTC Port**: 8889 (fixed systemd service port)
- **HLS Port**: 8888 (fixed systemd service port)

---

## Implementation Guidelines

### 1. Go Best Practices

#### Error Handling
```go
// Use wrapped errors with context
if err := someOperation(); err != nil {
    return fmt.Errorf("failed to perform operation: %w", err)
}

// Use custom error types for specific scenarios
type StreamNotReadyError struct {
    StreamName string
    Reason     string
}

func (e StreamNotReadyError) Error() string {
    return fmt.Sprintf("stream %s not ready: %s", e.StreamName, e.Reason)
}
```

#### Context Usage
```go
// Use context for cancellation and timeouts
func (cm *CameraMonitor) MonitorCameras(ctx context.Context) error {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            if err := cm.scanCameras(ctx); err != nil {
                cm.logger.Error("Camera scan failed", "error", err)
            }
        }
    }
}
```

#### Goroutine Management
```go
// Use errgroup for coordinated goroutine management
func (cm *CameraMonitor) StartMonitoring(ctx context.Context) error {
    g, ctx := errgroup.WithContext(ctx)
    
    // Camera discovery goroutine
    g.Go(func() error {
        return cm.discoverCameras(ctx)
    })
    
    // Event handling goroutine
    g.Go(func() error {
        return cm.handleEvents(ctx)
    })
    
    return g.Wait()
}
```

### 2. Configuration Management

#### Viper Configuration
```go
type Config struct {
    Camera struct {
        AutoStartStreams bool   `mapstructure:"auto_start_streams"`
        ScanInterval     string `mapstructure:"scan_interval"`
        DetectionTimeout string `mapstructure:"detection_timeout"`
    } `mapstructure:"camera"`
    
    WebSocket struct {
        Port            int    `mapstructure:"port"`
        MaxConnections  int    `mapstructure:"max_connections"`
        ReadTimeout     string `mapstructure:"read_timeout"`
        WriteTimeout    string `mapstructure:"write_timeout"`
    } `mapstructure:"websocket"`
    
    MediaMTX struct {
        APIURL     string `mapstructure:"api_url"`
        RTSPPort   int    `mapstructure:"rtsp_port"`
        WebRTCPort int    `mapstructure:"webrtc_port"`
    } `mapstructure:"mediamtx"`
}

func LoadConfig(configPath string) (*Config, error) {
    viper.SetConfigFile(configPath)
    viper.AutomaticEnv()
    
    if err := viper.ReadInConfig(); err != nil {
        return nil, fmt.Errorf("failed to read config: %w", err)
    }
    
    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }
    
    return &config, nil
}
```

### 3. Logging with Logrus

#### Structured Logging
```go
type Logger struct {
    *logrus.Logger
}

func NewLogger(config *LoggingConfig) *Logger {
    logger := logrus.New()
    
    // Set log level
    level, err := logrus.ParseLevel(config.Level)
    if err != nil {
        level = logrus.InfoLevel
    }
    logger.SetLevel(level)
    
    // Set formatter
    logger.SetFormatter(&logrus.JSONFormatter{
        TimestampFormat: time.RFC3339,
    })
    
    return &Logger{Logger: logger}
}

func (l *Logger) InfoWithContext(ctx context.Context, msg string, fields logrus.Fields) {
    if correlationID := ctx.Value("correlation_id"); correlationID != nil {
        fields["correlation_id"] = correlationID
    }
    l.WithFields(fields).Info(msg)
}
```

### 4. WebSocket JSON-RPC Implementation

#### Connection Management
```go
type WebSocketServer struct {
    upgrader websocket.Upgrader
    clients  map[*websocket.Conn]bool
    mutex    sync.RWMutex
    logger   *logging.Logger
}

func (wss *WebSocketServer) handleConnection(w http.ResponseWriter, r *http.Request) {
    conn, err := wss.upgrader.Upgrade(w, r, nil)
    if err != nil {
        wss.logger.Error("WebSocket upgrade failed", "error", err)
        return
    }
    
    wss.mutex.Lock()
    wss.clients[conn] = true
    wss.mutex.Unlock()
    
    defer func() {
        wss.mutex.Lock()
        delete(wss.clients, conn)
        wss.mutex.Unlock()
        conn.Close()
    }()
    
    wss.handleMessages(conn)
}
```

#### JSON-RPC Method Handling
```go
type JSONRPCHandler struct {
    cameraMonitor *CameraMonitor
    logger        *logging.Logger
}

func (jrh *JSONRPCHandler) HandleRequest(conn *websocket.Conn, request *JSONRPCRequest) {
    var response *JSONRPCResponse
    
    switch request.Method {
    case "ping":
        response = jrh.handlePing(request)
    case "get_camera_list":
        response = jrh.handleGetCameraList(request)
    case "get_camera_status":
        response = jrh.handleGetCameraStatus(request)
    default:
        response = &JSONRPCResponse{
            ID:    request.ID,
            Error: &JSONRPCError{Code: -32601, Message: "Method not found"},
        }
    }
    
    jrh.sendResponse(conn, response)
}
```

---

## Performance Targets

### Response Time Targets
- **Camera Detection**: <200ms latency
- **WebSocket Response**: <50ms for JSON-RPC methods
- **Stream Activation**: <3s for on-demand activation
- **Snapshot Capture**: <0.5s (Tier 1), <3s (Tier 2), <5s (Tier 3)

### Concurrency Targets
- **WebSocket Connections**: 1000+ concurrent connections
- **Camera Monitoring**: 10+ cameras with concurrent monitoring
- **FFmpeg Processes**: 10+ concurrent FFmpeg processes

### Resource Usage Targets
- **Memory Usage**: <60MB base, <200MB with 10 cameras
- **CPU Usage**: <20% idle, <80% under load
- **Network**: <100Mbps per camera stream

---

## JSON-RPC API Contract

The MediaMTX Camera Service implements a comprehensive JSON-RPC 2.0 API over WebSocket connections. **All API methods are implemented by MediaMTX Controller, with WebSocket server providing only protocol handling.**

### Connection
- **Protocol**: WebSocket
- **Endpoint**: `ws://localhost:8002/ws`
- **Authentication**: JWT token or API key required for all methods

### Architecture Compliance
- **WebSocket Server**: Thin protocol layer, delegates all operations to MediaMTX Controller
- **MediaMTX Controller**: Implements all business logic and camera operations
- **No Direct Access**: WebSocket server cannot access camera monitor, file system, or MediaMTX server directly

### Authentication & Authorization

#### Authentication Methods
- **JWT Token**: Pass `auth_token` parameter with valid JWT token
- **API Key**: Pass `auth_token` parameter with valid API key

#### Role-Based Access Control
- **viewer**: Read-only access to camera status, file listings, and basic information
- **operator**: Viewer permissions + camera control operations (snapshots, recording)
- **admin**: Full access to all features including system metrics and configuration

### Core Methods

#### ping
**Purpose**: Health check and connection validation
**Authentication**: Required (viewer role)
**Parameters**: None
**Returns**: Pong response with timestamp

#### authenticate
**Purpose**: Establish authenticated session
**Authentication**: Not required (handles authentication)
**Parameters**: `auth_token` (string) - JWT token or API key
**Returns**: Authentication result with user role and session information

#### get_camera_list
**Purpose**: Get list of all discovered cameras with current status
**Authentication**: Required (viewer role)
**Parameters**: None
**Returns**: Object with camera list and metadata

#### get_camera_status
**Purpose**: Get detailed status of specific camera
**Authentication**: Required (viewer role)
**Parameters**: `device` (string) - Camera device path
**Returns**: Detailed camera status and capabilities

#### get_camera_capabilities
**Purpose**: Get camera capabilities and supported formats
**Authentication**: Required (viewer role)
**Parameters**: `device` (string) - Camera device path
**Returns**: Camera capabilities and supported formats

### Camera Control Methods

#### take_snapshot
**Purpose**: Capture photo from camera
**Authentication**: Required (operator role)
**Parameters**: 
- `device` (string) - Camera device path
- `format` (string) - Image format (jpg, png)
- `quality` (int) - Image quality (1-100)
**Returns**: Snapshot result with file path and metadata

#### start_recording
**Purpose**: Start video recording from camera
**Authentication**: Required (operator role)
**Parameters**:
- `device` (string) - Camera device path
- `duration` (int) - Recording duration in seconds (optional)
- `format` (string) - Video format (mp4, avi)
**Returns**: Recording result with file path and metadata

#### stop_recording
**Purpose**: Stop active recording
**Authentication**: Required (operator role)
**Parameters**: `device` (string) - Camera device path
**Returns**: Recording stop result

### File Management Methods

#### list_recordings
**Purpose**: List all recorded video files
**Authentication**: Required (viewer role)
**Parameters**: None
**Returns**: List of recording files with metadata

#### list_snapshots
**Purpose**: List all captured snapshot files
**Authentication**: Required (viewer role)
**Parameters**: None
**Returns**: List of snapshot files with metadata

#### get_recording_info
**Purpose**: Get detailed information about recording file
**Authentication**: Required (viewer role)
**Parameters**: `filename` (string) - Recording filename
**Returns**: Detailed recording information

#### get_snapshot_info
**Purpose**: Get detailed information about snapshot file
**Authentication**: Required (viewer role)
**Parameters**: `filename` (string) - Snapshot filename
**Returns**: Detailed snapshot information

#### delete_recording
**Purpose**: Delete recording file
**Authentication**: Required (operator role)
**Parameters**: `filename` (string) - Recording filename
**Returns**: Deletion result

#### delete_snapshot
**Purpose**: Delete snapshot file
**Authentication**: Required (operator role)
**Parameters**: `filename` (string) - Snapshot filename
**Returns**: Deletion result

### System Management Methods

#### get_metrics
**Purpose**: Get system performance metrics
**Authentication**: Required (admin role)
**Parameters**: None
**Returns**: System metrics and performance data

#### get_streams
**Purpose**: Get active stream information
**Authentication**: Required (viewer role)
**Parameters**: None
**Returns**: Active stream information

#### get_status
**Purpose**: Get overall system status
**Authentication**: Required (viewer role)
**Parameters**: None
**Returns**: System status and health information

#### get_server_info
**Purpose**: Get server information and capabilities
**Authentication**: Required (viewer role)
**Parameters**: None
**Returns**: Server information and supported features

### Storage Management Methods

#### get_storage_info
**Purpose**: Get storage usage and capacity information
**Authentication**: Required (admin role)
**Parameters**: None
**Returns**: Storage information and usage statistics

#### set_retention_policy
**Purpose**: Configure file retention policies
**Authentication**: Required (admin role)
**Parameters**:
- `max_age_days` (int) - Maximum age in days
- `max_size_gb` (int) - Maximum size in GB
**Returns**: Policy update result

#### cleanup_old_files
**Purpose**: Manually trigger cleanup of old files
**Authentication**: Required (admin role)
**Parameters**: None
**Returns**: Cleanup result with statistics

### Real-Time Notifications

#### camera_status_update
**Purpose**: Real-time camera status updates
**Authentication**: Required (viewer role)
**Parameters**: None
**Returns**: Camera status change notifications

#### recording_status_update
**Purpose**: Real-time recording status updates
**Authentication**: Required (operator role)
**Parameters**: None
**Returns**: Recording status change notifications

### Error Handling

#### Standard Error Response Format
```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32000,
    "message": "Camera not found",
    "data": {
      "device": "/dev/video0",
      "available_devices": ["/dev/video1", "/dev/video2"]
    }
  },
  "id": 1
}
```

#### Error Codes
- **-32600**: Invalid Request
- **-32601**: Method not found
- **-32602**: Invalid params
- **-32603**: Internal error
- **-32000**: Camera not found
- **-32001**: Authentication required
- **-32002**: Insufficient permissions
- **-32003**: Device busy
- **-32004**: Recording in progress
- **-32005**: File not found

### Go Implementation Pattern

```go
// WebSocket Server: Thin protocol layer
type WebSocketServer struct {
    mediaMTXController MediaMTXController  // ONLY business logic dependency
    jwtHandler         *security.JWTHandler
    eventManager       *EventManager
    logger             *logging.Logger
}

func (s *WebSocketServer) HandleRequest(conn *websocket.Conn, request *JSONRPCRequest) {
    // Validate authentication
    if err := s.validateAuth(request); err != nil {
        s.sendError(conn, request.ID, -32001, "Authentication required", err)
        return
    }
    
    // Route to appropriate method handler - ALL delegate to MediaMTX
    var response *JSONRPCResponse
    switch request.Method {
    case "ping":
        response = s.handlePing(request)
    case "authenticate":
        response = s.handleAuthenticate(request)
    case "get_camera_list":
        // Simple delegation to MediaMTX Controller
        result, err := s.mediaMTXController.GetCameraList(context.Background())
        response = s.createResponse(request.ID, result, err)
    case "take_snapshot":
        // Simple delegation to MediaMTX Controller
        cameraID := request.Params["device"].(string)
        result, err := s.mediaMTXController.TakeSnapshot(context.Background(), cameraID, options)
        response = s.createResponse(request.ID, result, err)
    case "start_recording":
        // Simple delegation to MediaMTX Controller
        cameraID := request.Params["device"].(string)
        path := request.Params["path"].(string)
        result, err := s.mediaMTXController.StartRecording(context.Background(), cameraID, path)
        response = s.createResponse(request.ID, result, err)
    // ... all other methods delegate to MediaMTX Controller
    default:
        response = &JSONRPCResponse{
            ID:    request.ID,
            Error: &JSONRPCError{Code: -32601, Message: "Method not found"},
        }
    }
    
    s.sendResponse(conn, response)
}
```

## Technology Stack

### Core Technologies
- **Language**: Go 1.21+
- **WebSocket**: gorilla/websocket
- **HTTP Client**: net/http (standard library)
- **Configuration**: spf13/viper
- **Logging**: sirupsen/logrus
- **Authentication**: golang-jwt/jwt/v4
- **Testing**: testing (standard library) + testify

### External Dependencies
- **MediaMTX**: v1.13.1+ (systemd-managed service)
- **FFmpeg**: v6.0+ (for video processing)
- **V4L2**: Linux Video4Linux2 (for camera access)

### Development Tools
- **Linting**: golangci-lint
- **Formatting**: gofmt
- **Testing**: go test with coverage
- **Documentation**: godoc

---

## Architecture Decisions

### Technology Choices
- **Goroutines**: Efficient concurrency for I/O-bound workloads
- **Channels**: Thread-safe communication without explicit locking
- **Context**: Standard Go pattern for timeout and cancellation
- **Structured logging**: JSON format with correlation IDs for observability

---

## Architecture Compliance Checklist

Before implementing any changes, verify compliance with these architecture rules:

### ✅ Security Compliance (CRITICAL)
- [ ] All WebSocket connections require authentication
- [ ] RBAC is enforced for all methods
- [ ] Input validation is implemented for all parameters
- [ ] Rate limiting is active for all clients
- [ ] Security events are logged and audited
- [ ] No hard-coded security values
- [ ] Security configuration is externalized

### ✅ WebSocket Server Compliance
- [ ] WebSocket server has ONLY MediaMTX Controller as business logic dependency
- [ ] NO direct camera monitor access (`s.cameraMonitor.GetConnectedCameras()`)
- [ ] NO business logic in WebSocket methods
- [ ] NO camera validation in WebSocket layer
- [ ] NO file operations in WebSocket layer
- [ ] ALL operations delegate to MediaMTX Controller
- [ ] WebSocket server is under 500 lines of code
- [ ] Security middleware is properly integrated

### ✅ MediaMTX Controller Compliance
- [ ] MediaMTX Controller has camera monitor integration
- [ ] MediaMTX Controller implements all camera discovery methods
- [ ] MediaMTX Controller implements all business logic operations
- [ ] MediaMTX Controller is single source of truth for camera operations
- [ ] MediaMTX Controller has proper abstraction layer (camera0 ↔ /dev/video0)
- [ ] MediaMTX Controller validates all camera operations
- [ ] MediaMTX Controller enforces device access controls

### ✅ No Direct Hardware Access
- [ ] WebSocket server does NOT access camera monitor directly
- [ ] WebSocket server does NOT access file system directly
- [ ] WebSocket server does NOT access MediaMTX server directly
- [ ] ALL hardware access goes through MediaMTX Controller

### ✅ No Duplicated Logic
- [ ] Abstraction layer exists ONLY in MediaMTX Controller
- [ ] Camera validation exists ONLY in MediaMTX Controller
- [ ] Stream URL generation exists ONLY in MediaMTX Controller
- [ ] NO duplicated business logic between layers
- [ ] Security logic is centralized and not duplicated

---

**Document Status**: Updated architecture guide with strict enforcement rules to prevent broken implementations  
**Last Updated**: 2025-01-15  
**Next Review**: Before any WebSocket server or MediaMTX controller changes
