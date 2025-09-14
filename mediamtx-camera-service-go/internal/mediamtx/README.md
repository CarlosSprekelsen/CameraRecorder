# MediaMTX Controller Module

## Architecture Overview
The MediaMTX Controller is the complete business logic layer that orchestrates all camera operations. It provides a unified interface for streaming, recording, and snapshot capabilities while maintaining proper abstraction between client-facing APIs and hardware implementation.

## Core Architecture Principles

### Single Source of Truth
- MediaMTX Controller is the only component that handles business logic
- All camera operations flow through the controller
- WebSocket server delegates all operations to MediaMTX Controller
- No direct hardware access from external components

### Abstraction Layer
- Client APIs use camera identifiers (camera0, camera1)
- Internal operations use device paths (/dev/video0, /dev/video1)
- Controller manages the mapping between identifiers and paths
- Hardware details are hidden from external consumers

### Component Integration
- All sub-components are orchestrated by the controller
- Centralized configuration and logging across all components
- Proper dependency injection and lifecycle management
- Clear separation of concerns between components

## Core Components

### 1. **MediaMTXController** (Main Interface)
**Role**: Single entry point for external integration (WebSocket, HTTP API)
**Location**: `controller.go`
**Responsibilities**:
- Orchestrates all MediaMTX operations
- Provides unified API for recording, streaming, and snapshots
- Manages component lifecycle and state
- Uses centralized config and logger
- Implements camera discovery integration
- Provides abstraction layer (camera0 ↔ /dev/video0)

**Key Methods**:
```go
// Camera discovery operations
GetCameraList(ctx) (*CameraListResponse, error)
GetCameraStatus(ctx, device) (*CameraStatusResponse, error)
ValidateCameraDevice(ctx, device) (bool, error)

// Recording operations
StartRecording(ctx, device, path) (*RecordingSession, error)
StopRecording(ctx, sessionID) error

// Streaming operations  
StartStreaming(ctx, device) (*Stream, error)
StopStreaming(ctx, device) error
GetStreamURL(ctx, device) (string, error)
GetStreamStatus(ctx, device) (*Stream, error)

// Snapshot operations
TakeSnapshot(ctx, device, path) (*Snapshot, error)
```

### 2. **StreamManager** (Stream Lifecycle)
**Role**: Manages FFmpeg processes and MediaMTX paths for different use cases
**Location**: `stream_manager.go`
**Responsibilities**:
- Creates and manages streams for recording, viewing, and snapshots
- Handles FFmpeg process lifecycle
- Generates stream names and URLs
- Manages use-case specific configurations

**Implemented Methods**:
```go
StartStream(ctx, devicePath) (*Stream, error)  // Single path for all operations
StopStream(ctx, device) error                  // Single stop method
StopStreaming(ctx, device) error               // Legacy compatibility
GenerateStreamName(devicePath, useCase) string
GenerateStreamURL(streamName) string
buildFFmpegCommand(devicePath, streamName) string
```

### 3. **PathManager** (MediaMTX Path Management)
**Role**: Creates and manages MediaMTX server paths
**Location**: `path_manager.go`
**Responsibilities**:
- Creates MediaMTX paths with proper configuration
- Manages path lifecycle (create/delete)
- Handles path validation and error handling

**Key Methods**:
```go
CreatePath(ctx, name, source, options) error
DeletePath(ctx, name) error
PathExists(ctx, name) bool
```

### 4. **RecordingManager** (Recording Operations)
**Role**: Manages recording sessions and file operations
**Location**: `recording_manager.go`
**Responsibilities**:
- Creates and manages recording sessions
- Handles file rotation and cleanup
- Integrates with StreamManager for recording streams

### 5. **SnapshotManager** (Multi-Tier Snapshot Operations)
**Role**: Manages intelligent snapshot capture with multi-tier fallback system
**Location**: `snapshot_manager.go`
**Responsibilities**:
- **Tier 1**: Direct FFmpeg capture from USB devices (`/dev/video*`) - fastest path
- **Tier 2**: RTSP immediate capture from existing MediaMTX streams
- **Tier 3**: RTSP stream activation (creates MediaMTX path, then captures)
- **Tier 4**: Error handling and fallback mechanisms
- Manages snapshot file storage and metadata
- Integrates with FFmpegManager for image processing
- Supports both current USB devices and future external RTSP sources (STANAG 4609 UAVs)
 
Notes:
- Stream naming uses `StreamManager.GenerateStreamName`; no path-based fallbacks.

### 6. **FFmpegManager** (FFmpeg Process Management)
**Role**: Manages FFmpeg processes for snapshots
**Location**: `ffmpeg_manager.go`
**Responsibilities**:
- Starts/stops FFmpeg processes
- Monitors process health
- Handles process cleanup

### 7. **HealthMonitor** (Health Monitoring)
**Role**: Monitors MediaMTX service health and implements circuit breaker pattern
**Location**: `health_monitor.go`
**Responsibilities**:
- Continuous health checking via HTTP endpoints
- Circuit breaker implementation for failure handling
- Health metrics collection and status reporting
- Automatic recovery from unhealthy states

**Integration**:
- Integrated with MediaMTXController lifecycle
- Exposed via GetHealth() API method
- Publishes health events via WebSocket
- Included in system metrics responses

Notes:
- Intervals/timeouts are validated by centralized config; runtime emergency fallbacks are removed.

### 8. **RTSPConnectionManager** (Connection Monitoring)
**Role**: Monitors RTSP connections and sessions
**Location**: `rtsp_connection_manager.go`
**Responsibilities**:
- Monitors active RTSP connections
- Provides connection health metrics
- Tracks session statistics

### 9. **ExternalStreamDiscovery** (External Stream Management)
**Role**: Discovers and manages external RTSP streams including UAVs
**Location**: `external_discovery.go`
**Responsibilities**:
- Network scanning for external RTSP streams
- UAV stream discovery (Skydio, generic models)
- RTSP stream validation and health monitoring
- On-demand and periodic discovery modes
- STANAG 4609 compliance for military UAVs
- Configurable network ranges and parameters

## Data Flow

### Recording Flow
```
Controller.StartRecording() 
→ RecordingManager.StartRecording()
→ StreamManager.StartRecordingStream()
→ PathManager.CreatePath()
→ FFmpeg process starts
→ MediaMTX receives stream
→ Recording begins
```

### Streaming Flow
```
Controller.StartStreaming()
→ StreamManager.StartStream()
→ PathManager.CreatePath()
→ FFmpeg process starts
→ MediaMTX receives stream
→ Stream available for viewing
```

### Snapshot Flow (Multi-Tier Architecture)
```
Controller.TakeSnapshot()
→ SnapshotManager.TakeSnapshot()
→ Multi-Tier Fallback System:
  ├─ Tier 1: Direct FFmpeg from /dev/video* (USB devices) - FASTEST
  ├─ Tier 2: RTSP immediate capture (from existing MediaMTX streams)
  ├─ Tier 3: RTSP stream activation (create MediaMTX path, then capture)
  └─ Tier 4: Error handling (all methods failed)
→ Image captured and file saved
```

### External Stream Discovery Flow
```
Controller.DiscoverExternalStreams()
→ ExternalStreamDiscovery.DiscoverExternalStreams()
→ Network Scanning Process:
  ├─ Skydio Discovery: Scan known IPs and network ranges
  ├─ Generic UAV Discovery: Scan configurable ranges and ports
  ├─ RTSP Validation: Test stream connectivity and capabilities
  └─ Stream Registration: Add discovered streams to system
→ External streams available for management
```

## Configuration and Logging Integration

### Centralized Configuration
All components use `*config.ConfigManager` for configuration:
- MediaMTX server connection settings
- Stream configuration parameters
- File storage paths and settings
- Security and authentication settings
- Performance tuning parameters

### Centralized Logging
All components use `*logging.Logger` for structured logging:
- Correlation IDs for request tracing
- Component-specific log contexts
- Error tracking and debugging
- Performance metrics logging
- Security event logging

## Stream Use Cases

### UseCaseRecording
- **Purpose**: Long-running streams for recording
- **Auto-close**: Never (0s)
- **Suffix**: "" (no suffix)
- **Restart**: true

### Single Path Architecture (OPTIMIZED)
- **Purpose**: Unified streaming for all operations (viewing, recording, snapshots)
- **Auto-close**: 0s (never auto-close - stable for recording)
- **Suffix**: "" (no suffix - simple path names like camera0, camera1)
- **Restart**: true
- **Use Case**: All operations use the same stable MediaMTX path
- **Integration**: Single path handles streaming, recording, and snapshot operations
- **Benefits**: Eliminates path duplication, reduces MediaMTX complexity, improves performance

## Integration Points

### External Integration
- **WebSocket**: Uses MediaMTXController interface
- **HTTP API**: Uses MediaMTXController interface
- **Other modules**: Access via MediaMTXController

### Internal Integration
- **StreamManager** uses **PathManager** for MediaMTX paths
- **RecordingManager** uses **StreamManager** for recording streams
- **SnapshotManager** uses **FFmpegManager** for direct image processing
- **SnapshotManager** uses **StreamManager.StartStream()** for external RTSP sources (Tier 3)
- **ExternalStreamDiscovery** uses **PathManager** for external stream integration
- **HealthMonitor** monitors **MediaMTXClient** for service health
- **Controller** orchestrates all components including health monitoring and external discovery


## Snapshot Architecture: Current vs Future Use Cases

### Current Use Case: USB Devices (`/dev/video*`)
```
Controller.TakeSnapshot("/dev/video0", path)
→ SnapshotManager.TakeSnapshot()
→ Tier 1: Direct FFmpeg from /dev/video0
→ Success (fastest path, ~100ms)
```

### Future Use Case: External RTSP Streams (STANAG 4609 UAVs)
```
Controller.TakeSnapshot("rtsp://uav-stream", path)
→ SnapshotManager.TakeSnapshot()
→ Tier 1: Direct FFmpeg fails (not USB device)
→ Tier 2: RTSP immediate capture fails (no existing stream)
→ Tier 3: StreamManager.StartStream() creates MediaMTX path
→ FFmpeg captures from RTSP stream
→ Success (fallback path, ~500ms)
```

### Why StreamManager.StartStream() is Required
- **External RTSP sources** cannot use direct FFmpeg from `/dev/video*`
- **MediaMTX paths must be created** to receive external RTSP streams
- **Single path architecture** handles all operations (viewing, recording, snapshots)
- **StreamManager handles MediaMTX path creation** for all stream types
- **SnapshotManager uses StreamManager** in Tier 3 for external sources
- **Architecture supports both current and future requirements**

## Event-Driven Architecture

### Overview

The MediaMTX Controller implements event-driven patterns to replace polling-based approaches, providing improved performance, responsiveness, and test efficiency.

### Event-Driven Readiness System

The controller provides an event-driven readiness system that notifies subscribers when the controller becomes ready for operations.

#### Key Features

- **Event Subscription**: Subscribe to readiness events instead of polling
- **Immediate Notification**: Receive events as soon as readiness state changes
- **Multiple Subscribers**: Support for multiple concurrent subscribers
- **Timeout Handling**: Built-in timeout support for all operations

#### Usage Example

```go
// Subscribe to readiness events
readinessChan := controller.SubscribeToReadiness()

// Wait for readiness with timeout
select {
case <-readinessChan:
    // Controller is ready
case <-time.After(10 * time.Second):
    // Handle timeout
}
```

#### Readiness Conditions

The controller is considered ready when:
- Controller is running
- Camera monitor has completed discovery
- Health monitor is healthy (if present)

### Event-Driven Test Orchestration

The system includes comprehensive event-driven testing capabilities that replace polling with efficient event subscription.

#### EventDrivenTestHelper

The `EventDrivenTestHelper` provides testing utilities for event-driven patterns:

```go
// Create event-driven test helper
eventHelper := helper.CreateEventDrivenTestHelper(t)
defer eventHelper.Cleanup()

// Start observing events (non-blocking)
eventHelper.ObserveReadiness()
eventHelper.ObserveHealthChanges()
eventHelper.ObserveCameraEvents()

// Try operations immediately with retries instead of waiting
var session *RecordingSession
for i := 0; i < 3; i++ {
    session, err = controller.StartRecording(...)
    if err == nil {
        break
    }
    time.Sleep(time.Second)
}

// Verify events occurred after operations complete
assert.True(t, eventHelper.DidEventOccur("readiness"))
```

#### Parallel Test Execution

Event-driven patterns enable parallel test execution by removing sequential execution bottlenecks:

```go
// Before: Sequential execution required
func TestWithPolling(t *testing.T) {
    EnsureSequentialExecution(t) // ❌ Unnecessary bottleneck
    // ... polling logic
}

// After: Parallel execution enabled
func TestWithEvents(t *testing.T) {
    // No sequential execution needed - only reads information
    eventHelper := helper.CreateEventDrivenTestHelper(t)
    defer eventHelper.Cleanup()
    // ... event-driven logic
}
```

### Performance Benefits

#### Reduced CPU Usage
- **Before**: Continuous polling consumes CPU cycles
- **After**: CPU only used when events occur

#### Improved Test Performance
- **Before**: Tests wait for fixed intervals
- **After**: Tests respond immediately to events

#### Better Responsiveness
- **Before**: Maximum delay = polling interval
- **After**: Immediate event notification

### Event Types

#### 1. Readiness Events
- **Purpose**: Notify when controller becomes ready
- **Trigger**: All readiness conditions met
- **Usage**: `controller.SubscribeToReadiness()`

#### 2. Health Events (Planned)
- **Purpose**: Notify when health status changes
- **Trigger**: Health status transitions
- **Usage**: `eventHelper.ObserveHealthChanges()`

#### 3. Camera Events (Planned)
- **Purpose**: Notify when camera discovery events occur
- **Trigger**: Camera connected/disconnected
- **Usage**: `eventHelper.ObserveCameraEvents()`

### Migration from Polling

#### Step 1: Identify Polling Patterns
```go
// Polling pattern
for {
    if controller.IsReady() {
        break
    }
    time.Sleep(100 * time.Millisecond)
}
```

#### Step 2: Replace with Event Subscription
```go
// Event-driven pattern
readinessChan := controller.SubscribeToReadiness()
select {
case <-readinessChan:
    // Controller is ready
case <-timeout:
    // Handle timeout
}
```

#### Step 3: Update Tests
```go
// Remove unnecessary sequential execution
// Use event-driven test helpers
// Enable parallel test execution
```

### Best Practices

1. **Always Use Timeouts**: Prevent infinite waiting
2. **Clean Up Resources**: Close channels and cancel contexts
3. **Handle Errors Gracefully**: Don't let event failures crash tests
4. **Use Buffered Channels**: Prevent blocking on slow consumers
5. **Test Event Aggregation**: Verify multiple event handling works

### Future Enhancements

- **Health Event System**: Event-driven health monitoring
- **Camera Event System**: Enhanced camera discovery events
- **Event Persistence**: Event storage for debugging
- **Event Filtering**: Selective event subscription
- **Event Metrics**: Performance monitoring

## Architecture Benefits

1. **Single Source of Truth**: MediaMTX Controller is the only business logic layer
2. **Proper Abstraction**: Clean separation between client APIs and hardware implementation
3. **Centralized Configuration**: All components use shared config and logger
4. **Component Integration**: Clear dependencies and orchestration
5. **Separation of Concerns**: Each component has a single, well-defined responsibility
6. **Extensible Design**: Easy to add new capabilities and use cases
7. **Future-Ready**: Supports current USB devices and future external RTSP sources
8. **External Stream Support**: Comprehensive UAV and network-based stream discovery
9. **Configurable Discovery**: Flexible network scanning with model-specific parameters
10. **STANAG 4609 Compliance**: Military-grade UAV stream support
11. **Event-Driven Architecture**: Efficient event subscription replacing polling patterns
12. **Parallel Test Execution**: Improved test performance through event-driven testing
13. **Enhanced Responsiveness**: Immediate event notification for better system behavior
