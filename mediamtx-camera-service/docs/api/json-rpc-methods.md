# JSON-RPC API Reference

This document describes all available JSON-RPC 2.0 methods provided by the Camera Service.

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

**Implementation:** Validates JWT tokens or API keys, extracts user role and permissions, and establishes authenticated session for subsequent requests.

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

**Error Response (Invalid Token):**
```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32001,
    "message": "Authentication failed",
    "data": {
      "reason": "Invalid or expired token"
    }
  },
  "id": 0
}
```

## Performance Guarantees

All API methods adhere to architecture performance targets:
- **Status Methods** (get_camera_list, get_camera_status, ping): <50ms response time
- **Control Methods** (take_snapshot, start_recording, stop_recording): <100ms response time
- **WebSocket Notifications**: <20ms delivery latency from event occurrence

Performance measured from request receipt to response transmission at service level.

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

## Camera Control Methods

### get_camera_status
Get status for a specific camera device.

**Authentication:** Required (viewer role)

**Parameters:**
- device: string - Camera device path (required)

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
    "device": "/dev/video0"
  },
  "id": 3
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "device": "/dev/video0",
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

### take_snapshot  
Capture a snapshot from the specified camera.

**Authentication:** Required (operator role)

**Parameters:**
- device: string - Camera device path (required)
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
    "device": "/dev/video0",
    "filename": "snapshot_001.jpg"
  },
  "id": 4
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "device": "/dev/video0",
    "filename": "snapshot_001.jpg",
    "status": "completed",
    "timestamp": "2025-01-15T14:30:00Z",
    "file_size": 204800,
    "file_path": "/opt/camera-service/snapshots/snapshot_001.jpg"
  },
  "id": 4
}
```

### start_recording
Start recording video from the specified camera.

**Authentication:** Required (operator role)

**Parameters:**
- device: string - Camera device path (required)
- duration: number - Recording duration in seconds (optional)
- format: string - Recording format ("mp4", "mkv") (optional)

**Returns:** Recording session information with filename, status, and metadata

**Status:** ✅ Implemented

**Implementation:** Manages recording sessions through MediaMTX controller with session tracking, duration management, and proper file organization.

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "start_recording",
  "params": {
    "device": "/dev/video0",
    "duration": 3600,
    "format": "mp4"
  },
  "id": 5
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "device": "/dev/video0",
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "filename": "camera0_2025-01-15_14-30-00.mp4",
    "status": "STARTED",
    "start_time": "2025-01-15T14:30:00Z",
    "duration": 3600,
    "format": "mp4"
  },
  "id": 5
}
```

### stop_recording
Stop active recording for the specified camera.

**Authentication:** Required (operator role)

**Parameters:**
- device: string - Camera device path (required)

**Returns:** Recording completion information with final file details

**Status:** ✅ Implemented

**Implementation:** Properly terminates recording sessions with accurate duration calculation, file size reporting, and session cleanup.

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "stop_recording",
  "params": {
    "device": "/dev/video0"
  },
  "id": 6
}

// Response
{
  "jsonrpc": "2.0",
  "result": {
    "device": "/dev/video0",
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "filename": "camera0_2025-01-15_14-30-00.mp4",
    "status": "STOPPED",
    "start_time": "2025-01-15T14:30:00Z",
    "end_time": "2025-01-15T15:00:00Z",
    "duration": 1800,
    "file_size": 1073741824
  },
  "id": 6
}
```

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
        "filename": "camera0_2025-01-15_14-30-00.mp4",
        "file_size": 1073741824,
        "modified_time": "2025-01-15T14:30:00Z",
        "download_url": "/files/recordings/camera0_2025-01-15_14-30-00.mp4"
      }
    ],
    "total": 25,
    "limit": 10,
    "offset": 0
  },
  "id": 7
}
```

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
    "cpu_usage": 23.1
  },
  "id": 9
}
```

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
    "uptime": 86400,
    "version": "1.0.0",
    "components": {
      "websocket_server": "running",
      "camera_monitor": "running",
      "mediamtx_controller": "running"
    }
  },
  "id": 10
}
```

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
    "capabilities": ["snapshots", "recordings", "streaming"],
    "supported_formats": ["mp4", "mkv", "jpg"],
    "max_cameras": 10
  },
  "id": 11
}
```

### get_streams
Get list of all active streams from MediaMTX.

**Authentication:** Required (viewer role)

**Parameters:** None

**Returns:** Array of stream information objects

**Status:** ✅ Implemented

**Implementation:** Integrates with MediaMTX controller to return real-time stream status and metrics.

**Example:**
```json
// Request
{
  "jsonrpc": "2.0",
  "method": "get_streams",
  "id": 12
}

// Response
{
  "jsonrpc": "2.0",
  "result": [
    {
      "name": "camera0",
      "source": "ffmpeg -f v4l2 -i /dev/video0 -c:v libx264...",
      "ready": true,
      "readers": 2,
      "bytes_sent": 12345678
    },
    {
      "name": "camera1", 
      "source": "ffmpeg -f v4l2 -i /dev/video1 -c:v libx264...",
      "ready": false,
      "readers": 0,
      "bytes_sent": 0
    }
  ],
  "id": 12
}
```

**Response Fields:**
- `name`: Stream name (string)
- `source`: FFmpeg command or source configuration (string)
- `ready`: Stream readiness status (boolean)
- `readers`: Number of active stream readers (integer)
- `bytes_sent`: Total bytes sent for this stream (integer)

## HTTP File Download Endpoints

### GET /files/recordings/{filename}
Download a recording file via HTTP.

**Parameters:**
- filename: string - Name of the recording file to download

**Headers:**
- Authorization: Bearer {jwt_token} or X-API-Key: {api_key}

**Returns:** File content with appropriate Content-Type and Content-Disposition headers

**Status:** ✅ Implemented

**Implementation:** Serves recording files with proper MIME type detection, security validation, and access logging.

**Example:**
```bash
curl -H "Authorization: Bearer your_jwt_token" \
     http://localhost:8002/files/recordings/camera0_2025-01-15_14-30-00.mp4 \
     -o recording.mp4
```

### GET /files/snapshots/{filename}
Download a snapshot file via HTTP.

**Parameters:**
- filename: string - Name of the snapshot file to download

**Headers:**
- Authorization: Bearer {jwt_token} or X-API-Key: {api_key}

**Returns:** File content with appropriate Content-Type and Content-Disposition headers

**Status:** ✅ Implemented

**Implementation:** Serves snapshot files with proper MIME type detection, security validation, and access logging.

**Example:**
```bash
curl -H "Authorization: Bearer your_jwt_token" \
     http://localhost:8002/files/snapshots/snapshot_2025-01-15_14-30-00.jpg \
     -o snapshot.jpg
```

## Notifications

The server sends real-time notifications for camera events.

### camera_status_update
Sent when a camera connects, disconnects, or changes status.

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

### recording_status_update
Sent when recording starts, stops, or encounters an error.

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
    "filename": "camera0_2025-01-15_14-30-00.mp4",
    "duration": 0
  }
}
```

## Error Codes

Standard JSON-RPC 2.0 error codes:

- -32700: Parse error (Invalid JSON)
- -32600: Invalid Request  
- -32601: Method not found
- -32602: Invalid params
- -32603: Internal error

Custom error codes:
- -1000: Camera not found
- -1001: Camera not available
- -1002: Recording in progress
- -1003: MediaMTX error

## Implementation Notes

**Camera Data Integration:**
- All camera methods integrate with the hybrid camera discovery monitor
- Capability detection provides real format and resolution data when available
- Graceful fallbacks to default values when capability data is unavailable

**MediaMTX Integration:**
- Stream management through MediaMTX REST API
- Real snapshot capture using FFmpeg from RTSP streams  
- Recording session management with accurate duration tracking
- Health monitoring and error recovery

**Real-time Notifications:**
- Event-driven notifications from camera discovery system
- Proper field filtering per API specification
- Correlation ID support for request tracing

**Error Handling:**
- Comprehensive error responses with meaningful messages
- Graceful degradation when dependencies unavailable
- Proper cleanup and resource management

## Error Codes

Standard JSON-RPC 2.0 error codes plus service-specific codes:
- **-32001**: Authentication failed or token expired
- **-32002**: Rate limit exceeded
- **-32003**: Insufficient permissions (role-based access control)
- **-32004**: Camera not found or disconnected
- **-32005**: Recording already in progress
- **-32006**: MediaMTX service unavailable  
- **-32007**: Insufficient storage space
- **-32008**: Camera capability not supported

---

**API Version:** 1.0  
**Last Updated:** 2025-08-03  
**Implementation Status:** All core methods and notifications implemented and operational