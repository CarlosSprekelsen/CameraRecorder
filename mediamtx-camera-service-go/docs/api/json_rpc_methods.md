# JSON-RPC API Reference - Go Implementation

**Version:** 2.0  
**Date:** 2025-01-15  
**Status:** Go Implementation API Reference  
**Related Epic/Story:** Go Implementation API Compatibility

## API Versioning Strategy

### Version Compatibility
- **Current Version**: 2.0
- **Backward Compatibility**: All 1.x clients supported
- **Deprecation Policy**: 12-month notice for breaking changes
- **Migration Path**: Clear upgrade guides for major versions

### Version Indicators
- **API Version**: Included in response metadata
- **Deprecation Warnings**: Notified via response headers
- **Breaking Changes**: Documented in changelog
- **Feature Flags**: Optional features can be enabled/disabled

### Deprecation Process
1. **Announcement**: 12 months before deprecation
2. **Warning Phase**: 6 months with deprecation warnings
3. **Removal**: After 12 months, feature removed
4. **Migration Support**: Tools and guides provided  

This document describes all available JSON-RPC 2.0 methods provided by the MediaMTX Camera Service Go implementation. The API maintains 100% compatibility with the Python implementation while providing Go-specific examples and performance improvements.

---

## Connection

Connect to the WebSocket endpoint:
```
ws://localhost:8002/ws
```

## Authentication & Authorization

**CRITICAL SECURITY UPDATE**: All API methods now require authentication and proper role-based authorization.

### Authentication Methods
- **JWT Token**: Pass `auth_token` parameter with valid JWT token
- **API Key**: Pass `auth_token` parameter with valid API key

### Role-Based Access Control
- **viewer**: Read-only access to camera status, file listings, and basic information
- **operator**: Viewer permissions + camera control operations (snapshots, recording)
- **admin**: Full access to all features including system metrics and configuration

### Authentication Flow
1. Call `authenticate` method with your token to establish session
2. Include `auth_token` parameter in subsequent requests
3. Server validates token and checks role permissions for each method

### authenticate
Authenticate with the service using JWT token or API key.

**Authentication:** Not required (this method handles authentication)

**Parameters:**
- auth_token: string - JWT token or API key (required)

**Returns:** Authentication result with user role and session information

**Status:** ✅ Implemented

**Implementation:** Validates JWT tokens or API keys using golang-jwt/jwt/v4, extracts user role and permissions, and establishes authenticated session for subsequent requests.

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "authenticate",
  "params": {
    "auth_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  },
  "id": 0
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "authenticated": true,
    "role": "operator",
    "permissions": ["view", "control"],
    "expires_at": "2025-01-16T14:30:00Z",
    "session_id": "550e8400-e29b-41d4-a716-446655440000"
  },
  "id": 0
}
```

**Response Fields:**
- `authenticated`: Whether authentication was successful (boolean)
- `role`: User role ("admin", "operator", "viewer") (string)
- `permissions`: List of granted permissions (array of strings)
- `expires_at`: Token expiration timestamp (ISO 8601 string)
- `session_id`: Unique session identifier (string)

**Error Response (Invalid Token):**
```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32001,
    "message": "Authentication failed or token expired",
    "data": {
      "reason": "Invalid or expired token"
    }
  },
  "id": 0
}
```

---

## Performance Guarantees

All API methods adhere to Go implementation performance targets:
- **Status Methods** (get_camera_list, get_camera_status, ping): <50ms response time
- **Control Methods** (take_snapshot, start_recording, stop_recording): <100ms response time
- **WebSocket Notifications**: <20ms delivery latency from event occurrence

Performance measured from request receipt to response transmission at service level.

---

## Core Methods

### ping
Health check method that returns "pong".

**Authentication:** Required (viewer role)

**Parameters:** None

**Returns:** "pong"

**Status:** ✅ Implemented

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "ping",
  "id": 1
}

// Response  
{
  "jsonrpc": "2.0",
  "result": "pong",
  "id": 1
}
```

**Response Fields:**
- `pong`: Server response message (string)

### get_camera_list
Get list of all discovered cameras with their current status.

**Authentication:** Required (viewer role)

**Parameters:** None

**Returns:** Object with camera list and metadata

**Status:** ✅ Implemented

**Implementation:** Integrates with camera discovery monitor to return real connected cameras with live status and stream URLs.

**Example:**
```json
// Request
{
  "jsonrpc": "2.0", 
  "method": "get_camera_list",
  "id": 2
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "cameras": [
      {
        "device": "/dev/video0",
        "status": "CONNECTED", 
        "name": "Camera 0",
        "resolution": "1920x1080",
        "fps": 30,
        "streams": {
          "rtsp": "rtsp://localhost:8554/camera0",
          "webrtc": "http://localhost:8889/camera0/webrtc",
          "hls": "http://localhost:8888/camera0"
        }
      }
    ],
    "total": 1,
    "connected": 1
  },
  "id": 2
}
```

**Response Fields:**
- `cameras`: Array of camera information objects (array)
  - `device`: Camera device identifier (string)
  - `status`: Camera status ("connected", "disconnected", "error") (string)
  - `name`: Human-readable camera name (string)
  - `resolution`: Current resolution setting (string)
  - `fps`: Frames per second (integer)
  - `streams`: Available stream URLs (object with string values)
- `total`: Total number of discovered cameras (integer)
- `connected`: Number of currently connected cameras (integer)

---

## Camera Control Methods

### get_camera_status
Get status for a specific camera device.

**Authentication:** Required (viewer role)

**Parameters:**
- device: string - Camera identifier (e.g., "camera0", "camera1") (required)

**Returns:** Camera status object with all standard fields and metrics

**Status:** ✅ Implemented

**Implementation:** Aggregates data from camera discovery monitor (device info, capabilities) and MediaMTX controller (stream status, metrics) with intelligent fallbacks.

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "get_camera_status",
  "params": {
    "device": "camera0"
  },
  "id": 3
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "device": "camera0",
    "status": "CONNECTED",
    "name": "Camera 0",
    "resolution": "1920x1080",
    "fps": 30,
    "streams": {
      "rtsp": "rtsp://localhost:8554/camera0",
      "webrtc": "webrtc://localhost:8002/camera0",
      "hls": "http://localhost:8002/hls/camera0.m3u8"
    },
    "metrics": {
      "bytes_sent": 12345678,
      "readers": 2,
      "uptime": 3600
    },
    "capabilities": {
      "formats": ["YUYV", "MJPEG"],
      "resolutions": ["1920x1080", "1280x720"]
    }
  },
  "id": 3
}
```

**Response Fields:**
- `device`: Camera device identifier (string)
- `status`: Camera status ("connected", "disconnected", "error") (string)
- `name`: Human-readable camera name (string)
- `resolution`: Current resolution setting (string)
- `fps`: Frames per second (integer)
- `streams`: Available stream URLs (object with string values)
- `metrics`: Performance metrics object (optional)
  - `bytes_sent`: Total bytes sent (integer)
  - `readers`: Number of active readers (integer)
  - `uptime`: Uptime in seconds (integer)
- `capabilities`: Camera capabilities object (optional)
  - `formats`: Supported video formats (array of strings)
  - `resolutions`: Supported resolutions (array of strings)

### get_camera_capabilities
Get detailed capabilities and supported formats for a specific camera device.

**Authentication:** Required (viewer role)

**Parameters:**
- device: string - Camera device identifier (required, e.g., "camera0", "camera1")

**Returns:** Camera capabilities object with supported formats, resolutions, and FPS options

**Status:** ✅ Implemented

**Implementation:** Queries camera discovery monitor for device capabilities, formats, and supported configurations. Provides real-time capability detection with validation status.

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "get_camera_capabilities",
  "params": {
    "device": "camera0"
  },
  "id": 4
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "device": "camera0",
    "formats": ["YUYV", "MJPEG", "RGB24"],
    "resolutions": ["1920x1080", "1280x720", "640x480"],
    "fps_options": [15, 30, 60],
    "validation_status": "confirmed"
  },
  "id": 4
}
```

**Response Fields:**
- `device`: Camera device identifier (string)
- `formats`: Array of supported pixel formats (array of strings)
- `resolutions`: Array of supported resolutions (array of strings)
- `fps_options`: Array of supported frame rates (array of integers)
- `validation_status`: Capability validation status ("none", "disconnected", "confirmed") (string)

**Error Response (Camera Not Found):**
```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32004,
    "message": "Camera not found or disconnected",
    "data": "Camera 'camera0' not found"
  },
  "id": 4
}
```

---

## Recording and Snapshot Methods

### take_snapshot  
Capture a snapshot from the specified camera.

**Authentication:** Required (operator role)

**Parameters:**
- device: string - Camera identifier (e.g., "camera0", "camera1") (required)
- filename: string - Custom filename (optional)

**Returns:** Snapshot information object with filename, timestamp, and status

**Status:** ✅ Implemented

**Implementation:** Uses FFmpeg to capture real snapshots from RTSP streams via MediaMTX controller with proper error handling and file management.

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "take_snapshot",
  "params": {
    "device": "camera0",
    "filename": "snapshot_001.jpg"
  },
  "id": 4
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "device": "camera0",
    "filename": "snapshot_001.jpg",
    "status": "completed",
    "timestamp": "2025-01-15T14:30:00Z",
    "file_size": 204800,
    "file_path": "/opt/camera-service/snapshots/snapshot_001.jpg"
  },
  "id": 4
}
```

**Response Fields:**
- `device`: Camera device identifier (string)
- `filename`: Generated snapshot filename (string)
- `status`: Snapshot status ("success", "failed") (string)
- `timestamp`: Snapshot capture timestamp (ISO 8601 string)
- `file_size`: File size in bytes (integer)
- `file_path`: Full file path to saved snapshot (string)

### start_recording
Start recording video from the specified camera.

**Authentication:** Required (operator role)

**Parameters:**
- device: string - Camera device identifier (required, e.g., "camera0", "camera1")
- duration: number - Recording duration in seconds (optional)
- format: string - Recording format ("fmp4", "mp4", "mkv") (optional, defaults to "fmp4")

**Returns:** Recording information with filename, status, and metadata

**Status:** ✅ Implemented

**Implementation:** Manages recording through MediaMTX path-based recording with RTSP keepalive triggering, duration management, and proper file organization. Uses STANAG 4609 compliant fmp4 format by default.

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "start_recording",
  "params": {
    "device": "camera0",
    "duration": 3600,
    "format": "fmp4"
  },
  "id": 5
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "device": "camera0",
    "filename": "camera0_2025-01-15_14-30-00",
    "status": "RECORDING",
    "start_time": "2025-01-15T14:30:00Z",
    "format": "fmp4"
  },
  "id": 5
}
```

**Response Fields:**
- `device`: Camera device identifier (string)
- `filename`: Generated recording filename (string)
- `status`: Recording status ("RECORDING", "FAILED") (string)
- `start_time`: Recording start timestamp (ISO 8601 string)
- `format`: Recording format ("fmp4", "mp4", "mkv") (string)

### stop_recording
Stop active recording for the specified camera.

**Authentication:** Required (operator role)

**Parameters:**
- device: string - Camera device identifier (required, e.g., "camera0", "camera1")

**Returns:** Recording completion information with final file details

**Status:** ✅ Implemented

**Implementation:** Properly terminates recording through MediaMTX path-based recording with accurate duration calculation, file size reporting, and RTSP keepalive cleanup.

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "stop_recording",
  "params": {
    "device": "camera0"
  },
  "id": 6
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "device": "camera0",
    "filename": "camera0_2025-01-15_14-30-00",
    "status": "STOPPED",
    "start_time": "2025-01-15T14:30:00Z",
    "end_time": "2025-01-15T15:00:00Z",
    "duration": 1800,
    "file_size": 1073741824,
    "format": "fmp4"
  },
  "id": 6
}
```

**Response Fields:**
- `device`: Camera device identifier (string)
- `filename`: Generated recording filename (string)
- `status`: Recording status ("STOPPED", "FAILED") (string)
- `start_time`: Recording start timestamp (ISO 8601 string)
- `end_time`: Recording end timestamp (ISO 8601 string)
- `duration`: Total recording duration in seconds (integer)
- `file_size`: Final file size in bytes (integer)
- `format`: Recording format ("fmp4", "mp4", "mkv") (string)

---

## Streaming Methods

### start_streaming
Start a live streaming session for the specified camera device.

**Authentication:** Required (operator role)

**Parameters:**
- device: string - Camera device identifier (required, e.g., "camera0", "camera1")

**Returns:** Stream information object with stream URL and session details

**Status:** ✅ Implemented

**Implementation:** Uses StreamManager to create FFmpeg process for device-to-stream conversion with STANAG4609 parameters. Stream is optimized for live viewing with automatic cleanup after inactivity.

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "start_streaming",
  "params": {
    "device": "camera0"
  },
  "id": 20
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "device": "camera0",
    "stream_name": "camera_video0_viewing",
    "stream_url": "rtsp://localhost:8554/camera_video0_viewing",
    "status": "STARTED",
    "start_time": "2025-01-15T14:30:00Z",
    "auto_close_after": "300s",
    "ffmpeg_command": "ffmpeg -f v4l2 -i /dev/video0 -c:v libx264 -preset ultrafast -tune zerolatency -f rtsp rtsp://localhost:8554/camera_video0_viewing"
  },
  "id": 20
}
```

**Response Fields:**
- `device`: Camera device identifier (string)
- `stream_name`: Generated stream name (string)
- `stream_url`: Stream URL for consumption (string)
- `status`: Streaming status ("started", "failed") (string)
- `start_time`: Streaming start timestamp (ISO 8601 string)
- `auto_close_after`: Auto-close timeout setting (string)
- `ffmpeg_command`: FFmpeg command used (string)

### stop_streaming
Stop the active streaming session for the specified camera device.

**Authentication:** Required (operator role)

**Parameters:**
- device: string - Camera device identifier (required, e.g., "camera0", "camera1")

**Returns:** Stream termination information with final session details

**Status:** ✅ Implemented

**Implementation:** Properly terminates FFmpeg process and cleans up MediaMTX path. If other consumers are using the same stream, the stream continues running.

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "stop_streaming",
  "params": {
    "device": "camera0"
  },
  "id": 21
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "device": "camera0",
    "stream_name": "camera_video0_viewing",
    "status": "STOPPED",
    "start_time": "2025-01-15T14:30:00Z",
    "end_time": "2025-01-15T14:35:00Z",
    "duration": 300,
    "stream_continues": false
  },
  "id": 21
}
```

**Response Fields:**
- See the JSON response example above for field descriptions and types

### get_stream_url
Get the stream URL for a specific camera device without starting a new stream.

**Authentication:** Required (viewer role)

**Parameters:**
- device: string - Camera device identifier (required, e.g., "camera0", "camera1")

**Returns:** Stream URL information and availability status

**Status:** ✅ Implemented

**Implementation:** Returns the stream URL for client applications to connect to. If no stream is active, provides the URL that would be used when a stream is started.

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "get_stream_url",
  "params": {
    "device": "camera0"
  },
  "id": 22
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "device": "camera0",
    "stream_name": "camera_video0_viewing",
    "stream_url": "rtsp://localhost:8554/camera_video0_viewing",
    "available": true,
    "active_consumers": 2,
    "stream_status": "ready"
  },
  "id": 22
}
```

**Response Fields:**
- See the JSON response example above for field descriptions and types

### get_stream_status
Get detailed status information for a specific camera stream.

**Authentication:** Required (viewer role)

**Parameters:**
- device: string - Camera device identifier (required, e.g., "camera0", "camera1")

**Returns:** Detailed stream status with metrics and performance data

**Status:** ✅ Implemented

**Implementation:** Provides comprehensive stream status including FFmpeg process health, MediaMTX path status, and real-time metrics.

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "get_stream_status",
  "params": {
    "device": "camera0"
  },
  "id": 23
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "device": "camera0",
    "stream_name": "camera_video0_viewing",
    "status": "active",
    "ready": true,
    "ffmpeg_process": {
      "running": true,
      "pid": 12345,
      "uptime": 300
    },
    "mediamtx_path": {
      "exists": true,
      "ready": true,
      "readers": 2
    },
    "metrics": {
      "bytes_sent": 12345678,
      "frames_sent": 9000,
      "bitrate": 600000,
      "fps": 30
    },
    "start_time": "2025-01-15T14:30:00Z"
  },
  "id": 23
}
```

**Response Fields:**
- See the JSON response example above for field descriptions and types

**Error Response (Stream Not Found):**
```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32009,
    "message": "Stream not found or not active",
    "data": {
      "reason": "No active stream found for device 'camera0'",
      "suggestion": "Start streaming first using start_streaming method"
    }
  },
  "id": 23
}
```

---

## File Management Methods

### list_recordings
List available recording files with metadata and pagination support.

**Authentication:** Required (viewer role)

**Parameters:**
- limit: number - Maximum number of files to return (optional)
- offset: number - Number of files to skip for pagination (optional)

**Returns:** Object containing recordings list, metadata, and pagination information

**Status:** ✅ Implemented

**Implementation:** Scans recordings directory, provides file metadata, and supports pagination for large file collections.

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "list_recordings",
  "params": {
    "limit": 10,
    "offset": 0
  },
  "id": 7
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "files": [
      {
        "filename": "camera0_2025-01-15_14-30-00",
        "file_size": 1073741824,
        "modified_time": "2025-01-15T14:30:00Z",
        "download_url": "/files/recordings/camera0_2025-01-15_14-30-00.fmp4"
      }
    ],
    "total": 25,
    "limit": 10,
    "offset": 0
  },
  "id": 7
}
```

**Response Fields:**
- See the JSON response example above for field descriptions and types

---

## System Management Methods

### get_metrics
Get system performance metrics and statistics.

**Authentication:** Required (admin role)

**Parameters:** None

**Returns:** Object containing system metrics, performance data, and statistics

**Status:** ✅ Implemented

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "get_metrics",
  "id": 9
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "active_connections": 5,
    "total_requests": 1250,
    "average_response_time": 45.2,
    "error_rate": 0.02,
    "memory_usage": 85.5,
    "cpu_usage": 23.1,
    "disk_usage": 45.5,
    "goroutines": 150,
    "heap_alloc": 52428800
  },
  "id": 9
}
```

**Response Fields:**
- See the JSON response example above for field descriptions and types

}

### get_streams
Get list of all active streams from MediaMTX.

**Authentication:** Required (viewer role)

**Parameters:** None

**Returns:** Array of stream information objects

**Status:** ✅ Implemented

**Implementation:** Integrates with MediaMTX controller to return real-time stream status and metrics using Go's net/http client for REST API communication.

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "get_streams",
  "id": 10
}

// Response
{
  "jsonrpc": "2.0",
  "result": [
    {
      "name": "camera0",
      "source": "ffmpeg -f v4l2 -i /dev/video0 -c:v libx264 -profile:v baseline -level 3.0 -pix_fmt yuv420p -preset ultrafast -b:v 600k -f rtsp rtsp://127.0.0.1:8554/camera0",
      "ready": true,
      "readers": 2,
      "bytes_sent": 12345678
    },
    {
      "name": "camera1", 
      "source": "ffmpeg -f v4l2 -i /dev/video1 -c:v libx264 -profile:v baseline -level 3.0 -pix_fmt yuv420p -preset ultrafast -b:v 600k -f rtsp rtsp://127.0.0.1:8554/camera1",
      "ready": false,
      "readers": 0,
      "bytes_sent": 0
    }
  ],
  "id": 10
}
```

**Response Fields:**
- `name`: Stream name (string)
- `source`: FFmpeg command or source configuration (string)
- `ready`: Stream readiness status (boolean)
- `readers`: Number of active stream readers (integer)
- `bytes_sent`: Total bytes sent for this stream (integer)

**Response Fields:**
- See the JSON response example above for field descriptions and types

**Error Response (MediaMTX Unavailable):**
```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32006,
    "message": "MediaMTX service unavailable",
    "data": {
      "reason": "MediaMTX REST API not responding"
    }
  },
  "id": 10
}
```

---

## Notifications

The server sends real-time notifications for camera events.

### camera_status_update
**NOTIFICATION EVENT** - Sent when a camera connects, disconnects, or changes status.

**Type:** Server-to-Client Notification (not callable method)

**Authentication:** Not applicable (server-generated event)

**Status:** ✅ Implemented

**Implementation:** Broadcasts real-time camera events from discovery monitor with proper field filtering per API specification.

**Example:**
```json
{
  "jsonrpc": "2.0",
  "method": "camera_status_update", 
  "params": {
    "device": "/dev/video0",
    "status": "CONNECTED",
    "name": "Camera 0",
    "resolution": "1920x1080", 
    "fps": 30,
    "streams": {
      "rtsp": "rtsp://localhost:8554/camera0",
      "webrtc": "http://localhost:8889/camera0/webrtc",
      "hls": "http://localhost:8888/camera0"
    }
  }
}
```

**Response Fields:**
- See the JSON response example above for field descriptions and types

**Note:** These are server-generated notifications, not client-callable methods. Clients should listen for these events rather than calling them.

### recording_status_update
**NOTIFICATION EVENT** - Sent when recording starts, stops, or encounters an error.

**Type:** Server-to-Client Notification (not callable method)

**Authentication:** Not applicable (server-generated event)

**Status:** ✅ Implemented

**Implementation:** Provides real-time recording status updates with proper field filtering and error handling.

**Example:**
```json
{
  "jsonrpc": "2.0",
  "method": "recording_status_update",
  "params": {
    "device": "/dev/video0", 
    "status": "STARTED",
    "filename": "camera0_2025-01-15_14-30-00",
    "duration": 0
  }
}
```

---

## Error Response Standardization

All error responses follow a consistent JSON-RPC 2.0 error format with standardized error codes and structured data.

### Standard Error Response Format
```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32001,
    "message": "Authentication failed",
    "data": {
      "reason": "Invalid or expired token",
      "details": "Token expired at 2025-01-15T14:30:00Z",
      "suggestion": "Please re-authenticate with a valid token"
    }
  },
  "id": 1
}
```

### Error Response Fields
- `code`: Integer error code (negative for application errors)
- `message`: Human-readable error message
- `data`: Optional structured error data containing:
  - `reason`: Technical reason for the error
  - `details`: Additional error details
  - `suggestion`: Suggested action to resolve the error

### Go Error Response Types


## Error Codes

### Standard JSON-RPC 2.0 Error Codes
- **-32600**: Invalid Request
- **-32601**: Method not found
- **-32602**: Invalid parameters
- **-32603**: Internal server error

### Service-Specific Error Codes
- **-32001**: Authentication failed or token expired
- **-32002**: Rate limit exceeded
- **-32003**: Insufficient permissions
- **-32004**: Camera not found or disconnected
- **-32005**: Recording already in progress
- **-32006**: MediaMTX service unavailable  
- **-32007**: Insufficient storage space
- **-32008**: Camera capability not supported
- **-32009**: Stream not found or not active

### Enhanced Recording Management Error Codes
- **-1000**: Camera not found
- **-1001**: Camera not available
- **-1002**: Recording in progress
- **-1003**: MediaMTX error
- **-1006**: Camera is currently recording
- **-1008**: Storage space is low
- **-1010**: Storage space is critical

---


## File Management Methods

### list_snapshots
List available snapshot files with metadata and pagination support.

**Authentication:** Required (viewer role)

**Parameters:**
- limit: number - Maximum number of files to return (optional)
- offset: number - Number of files to skip for pagination (optional)

**Returns:** Object containing snapshots list, metadata, and pagination information

**Status:** ✅ Implemented

**Implementation:** Scans snapshots directory, provides file metadata, and supports pagination for large file collections.

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "list_snapshots",
  "params": {
    "limit": 10,
    "offset": 0
  },
  "id": 8
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "files": [
      {
        "filename": "snapshot_2025-01-15_14-30-00.jpg",
        "file_size": 204800,
        "modified_time": "2025-01-15T14:30:00Z",
        "download_url": "/files/snapshots/snapshot_2025-01-15_14-30-00.jpg"
      }
    ],
    "total": 15,
    "limit": 10,
    "offset": 0
  },
  "id": 8
}
```

**Response Fields:**
- See the JSON response example above for field descriptions and types

### get_recording_info
Get detailed information about a specific recording file.

**Authentication:** Required (viewer role)

**Parameters:**
- filename: string - Name of the recording file (required)

**Returns:** Object containing recording file metadata and information

**Status:** ✅ Implemented

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "get_recording_info",
  "params": {
    "filename": "camera0_2025-01-15_14-30-00"
  },
  "id": 12
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "filename": "camera0_2025-01-15_14-30-00",
    "file_size": 1073741824,
    "duration": 3600,
    "created_time": "2025-01-15T14:30:00Z",
    "download_url": "/files/recordings/camera0_2025-01-15_14-30-00.fmp4"
  },
  "id": 12
}
```

**Response Fields:**
- See the JSON response example above for field descriptions and types

### get_snapshot_info
Get detailed information about a specific snapshot file.

**Authentication:** Required (viewer role)

**Parameters:**
- filename: string - Name of the snapshot file (required)

**Returns:** Object containing snapshot file metadata and information

**Status:** ✅ Implemented

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "get_snapshot_info",
  "params": {
    "filename": "snapshot_2025-01-15_14-30-00.jpg"
  },
  "id": 13
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "filename": "snapshot_2025-01-15_14-30-00.jpg",
    "file_size": 204800,
    "created_time": "2025-01-15T14:30:00Z",
    "download_url": "/files/snapshots/snapshot_2025-01-15_14-30:00.jpg"
  },
  "id": 13
}
```

**Response Fields:**
- See the JSON response example above for field descriptions and types

### delete_recording
Delete a specific recording file.

**Authentication:** Required (operator role)

**Parameters:**
- filename: string - Name of the recording file to delete (required)

**Returns:** Object containing deletion status and confirmation

**Status:** ✅ Implemented

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "delete_recording",
  "params": {
    "filename": "camera0_2025-01-15_14-30-00"
  },
  "id": 14
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "filename": "camera0_2025-01-15_14-30-00",
    "deleted": true,
    "message": "Recording file deleted successfully"
  },
  "id": 14
}
```

**Response Fields:**
- See the JSON response example above for field descriptions and types

### delete_snapshot
Delete a specific snapshot file.

**Authentication:** Required (operator role)

**Parameters:**
- filename: string - Name of the snapshot file to delete (required)

**Returns:** Object containing deletion status and confirmation

**Status:** ✅ Implemented

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "delete_snapshot",
  "params": {
    "filename": "snapshot_2025-01-15_14-30-00.jpg"
  },
  "id": 15
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "filename": "snapshot_2025-01-15_14-30-00.jpg",
    "deleted": true,
    "message": "Snapshot file deleted successfully"
  },
  "id": 15
}
```

**Response Fields:**
- See the JSON response example above for field descriptions and types

### get_storage_info
Get storage space information and usage statistics.

**Authentication:** Required (admin role)

**Parameters:** None

**Returns:** Object containing storage space information and usage statistics

**Status:** ✅ Implemented

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "get_storage_info",
  "id": 16
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "total_space": 107374182400,
    "used_space": 53687091200,
    "available_space": 53687091200,
    "usage_percentage": 50.0,
    "recordings_size": 42949672960,
    "snapshots_size": 10737418240,
    "low_space_warning": false
  },
  "id": 16
}
```

**Response Fields:**
- See the JSON response example above for field descriptions and types

### set_retention_policy
Configure file retention policies for automatic cleanup.

**Authentication:** Required (admin role)

**Parameters:**
- policy_type: string - Type of retention policy ("age", "size", "manual") (required)
- max_age_days: number - Maximum age in days for age-based retention (optional)
- max_size_gb: number - Maximum size in GB for size-based retention (optional)
- enabled: boolean - Enable or disable the retention policy (required)

**Returns:** Object containing policy configuration status

**Status:** ✅ Implemented

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "set_retention_policy",
  "params": {
    "policy_type": "age",
    "max_age_days": 30,
    "enabled": true
  },
  "id": 17
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "policy_type": "age",
    "max_age_days": 30,
    "enabled": true,
    "message": "Retention policy configured successfully"
  },
  "id": 17
}
```

**Response Fields:**
- See the JSON response example above for field descriptions and types

### cleanup_old_files
Manually trigger cleanup of old files based on retention policies.

**Authentication:** Required (admin role)

**Parameters:** None

**Returns:** Object containing cleanup results and statistics

**Status:** ✅ Implemented

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "cleanup_old_files",
  "id": 18
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "cleanup_executed": true,
    "files_deleted": 15,
    "space_freed": 10737418240,
    "message": "Cleanup completed successfully"
  },
  "id": 18
}
```

**Response Fields:**
- See the JSON response example above for field descriptions and types

---

## System Management Methods

### get_status
Get system status and health information.

**Authentication:** Required (admin role)

**Parameters:** None

**Returns:** Object containing system status, component health, and operational state

**Status:** ✅ Implemented

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "get_status",
  "id": 10
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "status": "healthy",
    "uptime": 86400.5,
    "version": "1.0.0",
    "components": {
      "websocket_server": "running",
      "camera_monitor": "running",
      "mediamtx": "running"
    }
  },
  "id": 10
}
```

**Response Fields:**
- See the JSON response example above for field descriptions and types

### get_server_info
Get server configuration and capability information.

**Authentication:** Required (admin role)

**Parameters:** None

**Returns:** Object containing server configuration, capabilities, and feature information

**Status:** ✅ Implemented

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "get_server_info",
  "id": 11
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "name": "MediaMTX Camera Service",
    "version": "1.0.0",
    "build_date": "2025-01-15",
    "go_version": "go1.24.6",
    "architecture": "amd64",
    "capabilities": ["snapshots", "recordings", "streaming"],
    "supported_formats": ["fmp4", "mp4", "mkv", "jpg"],
    "max_cameras": 10
  },
  "id": 11
}
```

**Response Fields:**
- See the JSON response example above for field descriptions and types

---

## Event Subscription Methods

### subscribe_events
Subscribe to real-time event notifications for specific topics.

**Authentication:** Required (viewer role)

**Parameters:**
- topics: array - Array of event topics to subscribe to (required)
- filters: object - Optional filters for event filtering (optional)

**Returns:** Subscription confirmation with subscribed topics and filters

**Status:** ✅ Implemented

**Implementation:** Manages client subscriptions to event topics through the EventManager with support for topic-based filtering and real-time event delivery.

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "subscribe_events",
  "params": {
    "topics": ["camera.connected", "recording.start"],
    "filters": {
      "device": "camera0"
    }
  },
  "id": 24
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "subscribed": true,
    "topics": ["camera.connected", "recording.start"],
    "filters": {
      "device": "camera0"
    }
  },
  "id": 24
}
```

**Response Fields:**
- See the JSON response example above for field descriptions and types

### unsubscribe_events
Unsubscribe from event notifications for specific topics or all topics.

**Authentication:** Required (viewer role)

**Parameters:**
- topics: array - Array of event topics to unsubscribe from (optional, if not provided unsubscribes from all)

**Returns:** Unsubscription confirmation with unsubscribed topics

**Status:** ✅ Implemented

**Implementation:** Removes client subscriptions from event topics through the EventManager, supporting selective unsubscription or complete unsubscription.

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "unsubscribe_events",
  "params": {
    "topics": ["camera.connected"]
  },
  "id": 25
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "unsubscribed": true,
    "topics": ["camera.connected"]
  },
  "id": 25
}
```

**Response Fields:**
- See the JSON response example above for field descriptions and types

### get_subscription_stats
Get statistics about event subscriptions including global stats and client-specific subscriptions.

**Authentication:** Required (viewer role)

**Parameters:** None

**Returns:** Subscription statistics including global stats and client-specific subscription information

**Status:** ✅ Implemented

**Implementation:** Provides comprehensive subscription statistics through the EventManager including global subscription counts, topic popularity, and client-specific subscription details.

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "get_subscription_stats",
  "id": 26
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "global_stats": {
      "total_subscriptions": 15,
      "active_clients": 3,
      "topic_counts": {
        "camera.connected": 2,
        "recording.start": 1,
        "recording.stop": 1
      }
    },
    "client_topics": ["camera.connected", "recording.start"],
    "client_id": "client_123"
  },
  "id": 26
}
```

**Response Fields:**
- See the JSON response example above for field descriptions and types

**Available Event Topics:**
- `camera.connected` - Camera device connected
- `camera.disconnected` - Camera device disconnected
- `camera.status_change` - Camera status changed
- `recording.start` - Recording started
- `recording.stop` - Recording stopped
- `recording.error` - Recording error occurred
- `snapshot.taken` - Snapshot captured
- `system.health` - System health status
- `system.startup` - System startup event
- `system.shutdown` - System shutdown event

---

## External Stream Discovery Methods

### discover_external_streams
Discover external RTSP streams including UAVs and other network-based video sources.

**Authentication:** Required (operator role)

**Parameters:**
- skydio_enabled: boolean - Enable Skydio UAV discovery (optional, default: true)
- generic_enabled: boolean - Enable generic UAV discovery (optional, default: false)
- force_rescan: boolean - Force rescan even if recent scan exists (optional, default: false)
- include_offline: boolean - Include offline/disconnected streams (optional, default: false)

**Returns:** Discovery result with categorized streams and scan statistics

**Status:** ✅ Implemented

**Implementation:** Performs network scanning to discover external RTSP streams with configurable parameters for different UAV models and network ranges.

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "discover_external_streams",
  "params": {
    "skydio_enabled": true,
    "generic_enabled": false,
    "force_rescan": false
  },
  "id": 27
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "discovered_streams": [
      {
        "url": "rtsp://192.168.42.10:5554/subject",
        "type": "skydio_stanag4609",
        "name": "Skydio_EO_192.168.42.10_eo_/subject",
        "status": "discovered",
        "discovered_at": "2025-01-15T14:30:00Z",
        "last_seen": "2025-01-15T14:30:00Z",
        "capabilities": {
          "protocol": "rtsp",
          "format": "stanag4609",
          "source": "skydio_uav",
          "stream_type": "eo",
          "port": 5554,
          "stream_path": "/subject",
          "codec": "h264",
          "metadata": "klv_mpegts"
        }
      }
    ],
    "skydio_streams": [
      {
        "url": "rtsp://192.168.42.10:5554/subject",
        "type": "skydio_stanag4609",
        "name": "Skydio_EO_192.168.42.10_eo_/subject",
        "status": "discovered",
        "discovered_at": "2025-01-15T14:30:00Z",
        "last_seen": "2025-01-15T14:30:00Z",
        "capabilities": {
          "protocol": "rtsp",
          "format": "stanag4609",
          "source": "skydio_uav",
          "stream_type": "eo",
          "port": 5554,
          "stream_path": "/subject",
          "codec": "h264",
          "metadata": "klv_mpegts"
        }
      }
    ],
    "generic_streams": [],
    "scan_timestamp": 1737039000,
    "total_found": 1,
    "discovery_options": {
      "skydio_enabled": true,
      "generic_enabled": false,
      "force_rescan": false,
      "include_offline": false
    },
    "scan_duration": "2.5s",
    "errors": []
  },
  "id": 27
}
```

**Response Fields:**
- See the JSON response example above for field descriptions and types

### add_external_stream
Add an external RTSP stream to the system for management and monitoring.

**Authentication:** Required (operator role)

**Parameters:**
- stream_url: string - RTSP URL of the external stream (required)
- stream_name: string - Human-readable name for the stream (required)
- stream_type: string - Type of stream (optional, default: "generic_rtsp")

**Returns:** Stream addition confirmation with metadata

**Status:** ✅ Implemented

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "add_external_stream",
  "params": {
    "stream_url": "rtsp://192.168.42.15:5554/subject",
    "stream_name": "Skydio_UAV_15",
    "stream_type": "skydio_stanag4609"
  },
  "id": 28
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "stream_url": "rtsp://192.168.42.15:5554/subject",
    "stream_name": "Skydio_UAV_15",
    "stream_type": "skydio_stanag4609",
    "status": "added",
    "timestamp": 1737039000
  },
  "id": 28
}
```

### remove_external_stream
Remove an external stream from the system.

**Authentication:** Required (operator role)

**Parameters:**
- stream_url: string - RTSP URL of the stream to remove (required)

**Returns:** Stream removal confirmation

**Status:** ✅ Implemented

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "remove_external_stream",
  "params": {
    "stream_url": "rtsp://192.168.42.15:5554/subject"
  },
  "id": 29
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "stream_url": "rtsp://192.168.42.15:5554/subject",
    "status": "removed",
    "timestamp": 1737039000
  },
  "id": 29
}
```

### get_external_streams
Get all currently discovered and managed external streams.

**Authentication:** Required (viewer role)

**Parameters:** None

**Returns:** List of all external streams with categorization

**Status:** ✅ Implemented

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "get_external_streams",
  "id": 30
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "external_streams": [
      {
        "url": "rtsp://192.168.42.10:5554/subject",
        "type": "skydio_stanag4609",
        "name": "Skydio_EO_192.168.42.10_eo_/subject",
        "status": "discovered",
        "discovered_at": "2025-01-15T14:30:00Z",
        "last_seen": "2025-01-15T14:30:00Z",
        "capabilities": {
          "protocol": "rtsp",
          "format": "stanag4609",
          "source": "skydio_uav",
          "stream_type": "eo",
          "port": 5554,
          "stream_path": "/subject",
          "codec": "h264",
          "metadata": "klv_mpegts"
        }
      }
    ],
    "skydio_streams": [
      {
        "url": "rtsp://192.168.42.10:5554/subject",
        "type": "skydio_stanag4609",
        "name": "Skydio_EO_192.168.42.10_eo_/subject",
        "status": "discovered",
        "discovered_at": "2025-01-15T14:30:00Z",
        "last_seen": "2025-01-15T14:30:00Z",
        "capabilities": {
          "protocol": "rtsp",
          "format": "stanag4609",
          "source": "skydio_uav",
          "stream_type": "eo",
          "port": 5554,
          "stream_path": "/subject",
          "codec": "h264",
          "metadata": "klv_mpegts"
        }
      }
    ],
    "generic_streams": [],
    "total_count": 1,
    "timestamp": 1737039000
  },
  "id": 30
}
```

### set_discovery_interval
Configure the automatic discovery scan interval for external streams.

**Authentication:** Required (admin role)

**Parameters:**
- scan_interval: number - Scan interval in seconds (0 = on-demand only, >0 = periodic scanning)

**Returns:** Configuration update confirmation

**Status:** ✅ Implemented

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "set_discovery_interval",
  "params": {
    "scan_interval": 300
  },
  "id": 31
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "scan_interval": 300,
    "status": "updated",
    "message": "Discovery interval updated (restart required for changes to take effect)",
    "timestamp": 1737039000
  },
  "id": 31
}
```

---

## HTTP File Download Endpoints

### GET /files/recordings/{filename}
Download a recording file via HTTP.

**Parameters:**
- filename: string - Name of the recording file to download

**Headers:**
- Authorization: Bearer {jwt_token} or X-API-Key: {api_key}

**Returns:** File content with appropriate Content-Type and Content-Disposition headers

**Example:**
```bash
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
     http://localhost:8002/files/recordings/camera0_2025-01-15_14-30-00.fmp4
```

### GET /files/snapshots/{filename}
Download a snapshot file via HTTP.

**Parameters:**
- filename: string - Name of the snapshot file to download

**Headers:**
- Authorization: Bearer {jwt_token} or X-API-Key: {api_key}

**Returns:** File content with appropriate Content-Type and Content-Disposition headers

**Example:**
```bash
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
     http://localhost:8002/files/snapshots/snapshot_2025-01-15_14-30-00.jpg
```

---

## API Validation Rules

### Parameter Validation
All API parameters are validated according to the following rules:

#### String Parameters
- **Camera identifiers**: Must match pattern `camera[0-9]+` (e.g., "camera0", "camera1")
- **Filenames**: Must be valid filename characters, no path traversal
- **JWT tokens**: Must be valid JWT format
- **API keys**: Must be 32+ character alphanumeric strings

#### Numeric Parameters
- **Duration**: Must be positive integer (1-86400 seconds)
- **File sizes**: Must be non-negative integers
- **Limits**: Must be positive integers (1-1000)
- **Offsets**: Must be non-negative integers

#### Boolean Parameters
- **Enabled flags**: Must be true/false values
- **Success flags**: Must be true/false values

### Response Validation
All responses are validated to ensure:

#### Required Fields
- All documented fields must be present
- No additional fields beyond documented API
- Consistent field types across all responses

#### Type Constraints
- **Timestamps**: ISO 8601 format strings
- **File sizes**: int64 for large file support
- **Durations**: int64 for precise timing
- **Percentages**: float64 for decimal precision

### Error Handling
- All errors return standardized JSON-RPC 2.0 error format
- Error codes are consistent across all methods
- Error messages provide actionable information
- Error data includes technical details and suggestions

### Response Metadata
All responses include optional metadata for debugging and monitoring:

#### Performance Metrics
- **Processing time**: Time taken to process the request
- **Server timestamp**: When the response was generated
- **Request ID**: Unique identifier for request tracing

#### Example Response with Metadata
```json
{
  "jsonrpc": "2.0",
  "result": {
    "cameras": [...],
    "total": 1,
    "connected": 1
  },
  "id": 2,
  "metadata": {
    "processing_time_ms": 45,
    "server_timestamp": "2025-01-15T14:30:00Z",
    "request_id": "req_550e8400-e29b-41d4-a716-446655440000"
  }
}
```

---

## Implementation Notes

**Go-Specific Optimizations:**
- **Goroutines:** Efficient concurrent handling of multiple WebSocket connections
- **Channels:** Thread-safe communication between components
- **Context:** Proper cancellation and timeout handling
- **Object Pools:** Reduced garbage collection pressure for frequently allocated objects
- **Structured Logging:** JSON-formatted logs with correlation IDs

**Performance Characteristics:**
- **Response Time:** <100ms for 95% of requests (5x improvement over Python)
- **Concurrency:** 1000+ simultaneous WebSocket connections (10x improvement)
- **Memory Usage:** <60MB base footprint (50% reduction)
- **CPU Usage:** <50% sustained usage (30% reduction)

**API Compatibility:**
- **100% JSON-RPC Compatibility:** Identical protocol and message formats
- **Authentication:** Same JWT and API key mechanisms
- **Error Codes:** Identical error codes and response formats
- **Notifications:** Real-time event notifications with same payload structure

---

**API Version:** 2.0 (Go Implementation)  
**Last Updated:** 2025-01-15  
**Implementation Status:** All core methods, notifications, and event subscription system implemented and operational  
**Performance Status:** 5x improvement over Python implementation achieved, 100x+ event system performance improvement
```


