# MediaMTX Module Architecture

## Overview
The MediaMTX module provides unified streaming, recording, and snapshot capabilities for camera devices. It integrates with MediaMTX server to manage RTSP streams, FFmpeg processes, and file operations.

## Core Components

### 1. **MediaMTXController** (Main Interface)
**Role**: Single entry point for external integration (WebSocket, HTTP API)
**Location**: `controller.go`
**Responsibilities**:
- Orchestrates all MediaMTX operations
- Provides unified API for recording, streaming, and snapshots
- Manages component lifecycle and state
- Uses centralized config and logger

**Key Methods**:
```go
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
StartRecordingStream(ctx, devicePath) (*Stream, error)
StartViewingStream(ctx, devicePath) (*Stream, error) 
StartSnapshotStream(ctx, devicePath) (*Stream, error)
StopViewingStream(ctx, device) error
StopStreaming(ctx, device) error
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

### 5. **SnapshotManager** (Snapshot Operations)
**Role**: Manages snapshot capture and file operations
**Location**: `snapshot_manager.go`
**Responsibilities**:
- Captures snapshots from camera devices
- Manages snapshot file storage
- Integrates with FFmpegManager for image processing

### 6. **FFmpegManager** (FFmpeg Process Management)
**Role**: Manages FFmpeg processes for snapshots
**Location**: `ffmpeg_manager.go`
**Responsibilities**:
- Starts/stops FFmpeg processes
- Monitors process health
- Handles process cleanup

### 7. **RTSPConnectionManager** (Connection Monitoring)
**Role**: Monitors RTSP connections and sessions
**Location**: `rtsp_connection_manager.go`
**Responsibilities**:
- Monitors active RTSP connections
- Provides connection health metrics
- Tracks session statistics

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
→ StreamManager.StartViewingStream()
→ PathManager.CreatePath()
→ FFmpeg process starts
→ MediaMTX receives stream
→ Stream available for viewing
```

### Snapshot Flow
```
Controller.TakeSnapshot()
→ SnapshotManager.TakeSnapshot()
→ FFmpegManager.StartProcess()
→ Image captured
→ File saved
```

## Configuration Integration

### Centralized Config Usage
All components use `*MediaMTXConfig` from centralized config:
```go
type MediaMTXConfig struct {
    Host        string
    Port        int
    RTSPPort    int
    // ... other fields
}
```

### Centralized Logger Usage
All components use `*logging.Logger` from centralized logging:
```go
logger.WithFields(map[string]interface{}{
    "device": device,
    "action": "start_streaming",
}).Info("Starting streaming session")
```

## Stream Use Cases

### UseCaseRecording
- **Purpose**: Long-running streams for recording
- **Auto-close**: Never (0s)
- **Suffix**: "" (no suffix)
- **Restart**: true

### UseCaseViewing  
- **Purpose**: Live viewing streams
- **Auto-close**: 300s (5 minutes after last viewer)
- **Suffix**: "_viewing"
- **Restart**: true

### UseCaseSnapshot
- **Purpose**: Quick snapshot capture
- **Auto-close**: 60s (1 minute after capture)
- **Suffix**: "_snapshot" 
- **Restart**: false

## Integration Points

### External Integration
- **WebSocket**: Uses MediaMTXController interface
- **HTTP API**: Uses MediaMTXController interface
- **Other modules**: Access via MediaMTXController

### Internal Integration
- **StreamManager** uses **PathManager** for MediaMTX paths
- **RecordingManager** uses **StreamManager** for recording streams
- **SnapshotManager** uses **FFmpegManager** for image processing
- **Controller** orchestrates all components


## Architecture Benefits

1. **Separation of Concerns**: Each component has a clear, single responsibility
2. **Centralized Configuration**: All components use shared config and logger
3. **Unified Interface**: MediaMTXController provides single entry point
4. **Reusable Components**: StreamManager handles all stream types
5. **Proper Integration**: Components work together without duplication
6. **Extensible**: Easy to add new use cases or stream types
