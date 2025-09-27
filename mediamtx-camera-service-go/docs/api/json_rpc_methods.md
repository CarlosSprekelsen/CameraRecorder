# JSON-RPC API Reference

**Version:** 1.0.0  
**Date:** 2025-01-15  
**Status:** Camera Service API Reference  

## API Versioning Strategy

### Version Compatibility

- **Current Version**: 1.0.0
- **Deprecation Policy**: 12-month notice for breaking changes
- **Migration Path**: Clear upgrade guides for major versions

### Version Indicators

- **API Version**: Included in response metadata
- **Deprecation Warnings**: Notified via metadata or notifications
- **Breaking Changes**: Documented in changelog
- **Feature Flags**: Optional features can be enabled/disabled

### Deprecation Process

1. **Announcement**: 12 months before deprecation
2. **Warning Phase**: 6 months with deprecation warnings
3. **Removal**: After 12 months, feature removed
4. **Migration Support**: Tools and guides provided  

This document describes all available JSON-RPC 2.0 methods provided by the MediaMTX Camera Service Go implementation. The API provides high-performance WebSocket communication with comprehensive camera management capabilities.

## JSON-RPC 2.0 Compliance (Authoritative)

* **Protocol:** JSON-RPC 2.0 over WebSocket.
* **Envelope:** Every message MUST include `"jsonrpc":"2.0"`, and EITHER `"result"` OR `"error"`, plus `"id"` for calls that expect a response.
* **Notifications:** Supported (requests **without** `"id"`). No response will be sent.
* **Batch:** **Not supported.** If a batch is received, return `-32600 Invalid Request`.
* **IDs:** MAY be string or number; MUST be echoed unchanged in the response.
* **Errors:** Use JSON-RPC standard codes plus the **vendor range `-32000..-32099`** defined in this spec (see Error Catalog).

---

## Connection

Connect to the WebSocket endpoint:

```text
ws://localhost:8002/ws
```

## Authentication & Authorization

**CRITICAL SECURITY NOTE**: Most API methods require authentication and proper role-based authorization. The `ping` method is the only exception for connectivity testing.

### Authentication Methods

- **JWT Token**: Pass `auth_token` parameter with valid JWT token
- **API Key**: Pass `auth_token` parameter with valid API key

### Role-Based Access Control

- **viewer**: Read-only access to camera status, file listings, and basic information
- **operator**: Viewer permissions + camera control operations (snapshots, recording)
- **admin**: Full access to all features including system metrics and configuration

### Authentication Flow

**Session Authentication (WebSocket):**
Call `authenticate` once after opening the WebSocket. If it succeeds, the **connection becomes authenticated** for its lifetime. **Do not** include tokens in subsequent method calls on the same socket. If the socket reconnects, authenticate again.

**Identifier Policy (External Contract):**

* Clients MUST see **only** logical camera IDs: `"camera0"`, `"camera1"`, …
* Linux device paths like `"/dev/video0"` are **internal** and MUST NOT appear in public responses.
* When needed for admin/debug, expose internal paths under `debug.internal_device_path` **only** when `role=admin` and `debug=true`.

**Streams Object Contract:**

* `rtsp`: e.g., `"rtsp://<host>:8554/camera0"`
* `hls`: e.g., `"https://<host>/hls/camera0.m3u8"`
* `webrtc`: **(reserved)** will be added once the signaling API is published.

**Pagination (Standard):**

* `limit` (default **50**, max **1000**)
* `offset` (default **0**)
* **Sort order:** unless stated otherwise, lists are sorted by `created_at` **desc**.
* Empty sets MUST return `"result": { "items": [], "total": 0 }`.

## Permissions Matrix

| Method               | Viewer | Operator | Admin |
| -------------------- | :----: | :------: | :---: |
| `ping`               |    ✅   |     ✅    |   ✅   |
| `authenticate`       |    ✅   |     ✅    |   ✅   |
| `get_camera_list`    |    ✅   |     ✅    |   ✅   |
| `get_camera_status`  |    ✅   |     ✅    |   ✅   |
| `get_camera_capabilities` | ✅ | ✅ | ✅ |
| `take_snapshot`      |    ❌   |     ✅    |   ✅   |
| `start_recording`    |    ❌   |     ✅    |   ✅   |
| `stop_recording`     |    ❌   |     ✅    |   ✅   |
| `list_recordings`    |    ✅   |     ✅    |   ✅   |
| `list_snapshots`     |    ✅   |     ✅    |   ✅   |
| `get_recording_info` |    ✅   |     ✅    |   ✅   |
| `get_snapshot_info` |    ✅   |     ✅    |   ✅   |
| `delete_recording`   |    ❌   |     ❌    |   ✅   |
| `delete_snapshot`     |    ❌   |     ❌    |   ✅   |
| `start_streaming`    |    ❌   |     ✅    |   ✅   |
| `stop_streaming`     |    ❌   |     ✅    |   ✅   |
| `get_stream_url`     |    ✅   |     ✅    |   ✅   |
| `get_stream_status`  |    ✅   |     ✅    |   ✅   |
| `get_streams`        |    ✅   |     ✅    |   ✅   |
| `get_metrics`        |    ❌   |     ❌    |   ✅   |
| `get_status`         |    ❌   |     ❌    |   ✅   |
| `get_system_status`  |    ✅   |     ✅    |   ✅   |
| `get_server_info`    |    ✅   |     ✅    |   ✅   |
| `get_storage_info`   |    ❌   |     ❌    |   ✅   |
| `set_retention_policy` |  ❌   |     ❌    |   ✅   |
| `cleanup_old_files`  |    ❌   |     ❌    |   ✅   |
| `subscribe_events`   |    ✅   |     ✅    |   ✅   |
| `unsubscribe_events` |    ✅   |     ✅    |   ✅   |
| `get_subscription_stats` | ✅ | ✅ | ✅ |
| `discover_external_streams` | ❌ | ✅ | ✅ |
| `add_external_stream` | ❌ | ✅ | ✅ |
| `remove_external_stream` | ❌ | ✅ | ✅ |
| `get_external_streams` | ✅ | ✅ | ✅ |
| `set_discovery_interval` | ❌ | ❌ | ✅ |

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

**Authentication:** **Not required**
**Purpose:** Connectivity + envelope sanity check **before** `authenticate`.

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
        "device": "camera0",
        "status": "CONNECTED", 
        "name": "Camera 0",
        "resolution": "1920x1080",
        "fps": 30,
        "streams": {
          "rtsp": "rtsp://localhost:8554/camera0",
          "hls": "https://localhost/hls/camera0.m3u8"
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
  - `status`: Camera status ("CONNECTED", "DISCONNECTED", "ERROR") (string)
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
      "hls": "https://localhost/hls/camera0.m3u8"
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
- `status`: Camera status ("CONNECTED", "DISCONNECTED", "ERROR") (string)
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
    "validation_status": "CONFIRMED"
  },
  "id": 4
}
```

**Response Fields:**

- `device`: Camera device identifier (string)
- `formats`: Array of supported pixel formats (array of strings)
- `resolutions`: Array of supported resolutions (array of strings)
- `fps_options`: Array of supported frame rates (array of integers)
- `validation_status`: Capability validation status ("NONE", "DISCONNECTED", "CONFIRMED") (string)

**Error Response (Camera Not Found):**

```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32010,
    "message": "Camera not found",
    "data": { "resource": "camera", "id": "camera0" }
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
    "status": "SUCCESS",
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
- `status`: Snapshot status ("SUCCESS", "FAILED") (string)
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
- `status`: Recording status ("RECORDING", "STARTING", "STOPPING", "PAUSED", "ERROR", "FAILED") (string)
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
- `status`: Recording status ("STOPPED", "STARTING", "STOPPING", "PAUSED", "ERROR", "FAILED") (string)
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
- `status`: Streaming status ("STARTED", "FAILED") (string)
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

- `device`: Camera device identifier (string)
- `stream_name`: Generated stream name (string)
- `status`: Streaming status ("STOPPED", "FAILED") (string)
- `start_time`: Streaming start timestamp (ISO 8601 string)
- `end_time`: Streaming end timestamp (ISO 8601 string)
- `duration`: Total streaming duration in seconds (integer)
- `stream_continues`: Whether stream continues for other consumers after this stop (boolean, true if other clients still connected, false if this was the last consumer)
- `message`: Success message (string)

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
    "stream_status": "READY"
  },
  "id": 22
}
```

**Response Fields:**

- `device`: Camera device identifier (string)
- `stream_name`: Generated stream name (string)
- `stream_url`: Stream URL for consumption (string)
- `available`: Whether stream is available (boolean)
- `active_consumers`: Number of active stream consumers (integer)
- `stream_status`: Stream readiness status ("READY", "NOT_READY", "ERROR") (string)

**Error Response (Device Not Found):**

```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32004,
    "message": "Camera not found or disconnected",
    "data": {
      "reason": "Device 'camera0' not found",
      "suggestion": "Use get_camera_list to see available cameras"
    }
  },
  "id": 22
}
```

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
    "status": "ACTIVE",
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

- `device`: Camera device identifier (string)
- `stream_name`: Generated stream name (string)
- `status`: Stream status ("ACTIVE", "INACTIVE", "ERROR", "STARTING", "STOPPING") (string)
- `ready`: Whether stream is ready for consumption (boolean)
- `ffmpeg_process`: FFmpeg process information (object)
  - `running`: Whether FFmpeg process is running (boolean)
  - `pid`: Process ID (integer)
  - `uptime`: Process uptime in seconds (integer)
- `mediamtx_path`: MediaMTX path information (object)
  - `exists`: Whether path exists (boolean)
  - `ready`: Whether path is ready (boolean)
  - `readers`: Number of active readers (integer)
- `metrics`: Stream performance metrics (object)
  - `bytes_sent`: Total bytes sent (integer)
  - `frames_sent`: Total frames sent (integer)
  - `bitrate`: Current bitrate (integer)
  - `fps`: Frames per second (integer)
- `start_time`: Stream start timestamp (ISO 8601 string)

**Error Response (Stream Not Found):**

```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32010,
    "message": "Stream not found",
    "data": { "resource": "stream", "id": "camera0" }
  },
  "id": 23
}
```

---

## Recording File Management

### list_recordings

List available recording files with metadata and pagination support.

**Authentication:** Required (viewer role)

**Parameters:**

- limit: number - Maximum number of files to return (optional, default: 50, max: 1000)
- offset: number - Number of files to skip for pagination (optional, default: 0)

**Returns:** Object containing recordings list, metadata, and pagination information

**Status:** ✅ Implemented

**Implementation:** Scans recordings directory, provides file metadata, and supports pagination for large file collections.

Note (Empty Set Semantics):

- When no recording files exist, this method MUST return a successful JSON-RPC response with an empty result object, not an error.
- Required structure for empty sets:
  {
    "jsonrpc": "2.0",
    "result": { "files": [], "total": 0, "limit": <int>, "offset": <int> },
    "id": <id>
  }
- Errors (e.g., directory inaccessible) MUST use documented JSON-RPC error codes with structured data.

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

- `files`: Array of recording file information objects (array)
  - `filename`: Recording filename without extension (string)
  - `file_size`: File size in bytes (integer)
  - `modified_time`: File modification timestamp (ISO 8601 string)
  - `download_url`: HTTP download URL for the file (string)
- `total`: Total number of recording files (integer)
- `limit`: Maximum number of files requested (integer)
- `offset`: Number of files skipped for pagination (integer)

**Error Response (Directory Not Found):**

```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32603,
    "message": "Internal server error",
    "data": {
      "reason": "Recordings directory not found or inaccessible",
      "suggestion": "Check storage configuration and permissions"
    }
  },
  "id": 7
}
```

---

## System Metrics and Monitoring

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
    "timestamp": "2025-01-15T14:30:00Z",
    "system_metrics": {
      "cpu_usage": 23.1,
      "memory_usage": 85.5,
      "disk_usage": 45.5,
      "goroutines": 150
    },
    "camera_metrics": {
      "connected_cameras": 2,
      "cameras": {
        "camera0": {
          "path": "camera0",
          "name": "USB 2.0 Camera: USB 2.0 Camera",
          "status": "CONNECTED",
          "device_num": 0,
          "last_seen": "2025-01-15T14:30:00Z",
          "capabilities": {
            "driver_name": "uvcvideo",
            "card_name": "USB 2.0 Camera: USB 2.0 Camera",
            "bus_info": "usb-0000:00:1a.0-1.2",
            "version": "6.14.8",
            "capabilities": ["0x84a00001", "Video Capture", "Metadata Capture", "Streaming", "Extended Pix Format"],
            "device_caps": ["0x04200001", "Video Capture", "Streaming", "Extended Pix Format"]
          },
          "formats": [
            {
              "pixel_format": "YUYV",
              "width": 640,
              "height": 480,
              "frame_rates": ["30.000", "20.000", "15.000", "10.000", "5.000"]
            }
          ]
        }
      }
    },
    "recording_metrics": {},
    "stream_metrics": {
      "active_streams": 0,
      "total_streams": 4,
      "total_viewers": 0
    }
  },
  "id": 9
}
```

**Response Fields:**

- `timestamp`: Metrics collection timestamp (ISO 8601 string)
- `system_metrics`: System-level performance metrics (object)
  - `cpu_usage`: CPU usage percentage (0.0-100.0) (float64)
  - `memory_usage`: Memory usage percentage (0.0-100.0) (float64)
  - `disk_usage`: Disk usage percentage (0.0-100.0) (float64)
  - `goroutines`: Number of active goroutines (integer)
- `camera_metrics`: Camera-specific metrics and information (object)
  - `connected_cameras`: Number of currently connected cameras (integer)
  - `cameras`: Detailed camera information indexed by device path (object)
    - Each camera contains: path, name, status, device_num, last_seen, capabilities, formats
- `recording_metrics`: Recording performance metrics (object, currently empty)
- `stream_metrics`: Streaming metrics and statistics (object)
  - `active_streams`: Number of currently active streams (integer)
  - `total_streams`: Total number of configured streams (integer)
  - `total_viewers`: Total number of active stream viewers (integer)

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

**Error Response (MediaMTX Unavailable):**

```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32050,
    "message": "Dependency failed",
    "data": { "dependency": "MediaMTX", "detail": "REST API not responding" }
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
    "device": "camera0",
    "status": "CONNECTED",
    "name": "Camera 0",
    "resolution": "1920x1080", 
    "fps": 30,
    "streams": {
      "rtsp": "rtsp://localhost:8554/camera0",
      "hls": "https://localhost/hls/camera0.m3u8"
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
    "device": "camera0", 
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

## Error Catalog

| Code       | Name              | When to use                            | `error.data` fields          |
| ---------- | ----------------- | -------------------------------------- | ---------------------------- |
| **-32600** | Invalid Request   | Bad JSON-RPC envelope                  | `hint`                       |
| **-32601** | Method Not Found  | Unknown `"method"`                     | `hint`                       |
| **-32602** | Invalid Params    | Fails validation rules                 | `param`, `rule`, `hint`      |
| **-32603** | Internal Error    | Unhandled server error                 | `request_id`                 |
| **-32001** | Auth Failed       | Invalid/expired token                  | `reason`                     |
| **-32002** | Permission Denied | Role lacks permission                  | `required_role`, `have_role` |
| **-32010** | Not Found         | Recording/file/camera not found        | `resource`, `id`             |
| **-32020** | Invalid State     | Operation not allowed in current state | `state`, `allowed_states`    |
| **-32030** | Unsupported       | Feature/capability not available       | `feature`                    |
| **-32040** | Rate Limited      | Too many requests                      | `retry_after_ms`             |
| **-32050** | Dependency Failed | MediaMTX/FFmpeg error                  | `dependency`, `detail`       |

---

## Snapshot File Management

### list_snapshots

List available snapshot files with metadata and pagination support.

**Authentication:** Required (viewer role)

**Parameters:**

- limit: number - Maximum number of files to return (optional, default: 50, max: 1000)
- offset: number - Number of files to skip for pagination (optional, default: 0)

**Returns:** Object containing snapshots list, metadata, and pagination information

**Status:** ✅ Implemented

**Implementation:** Scans snapshots directory, provides file metadata, and supports pagination for large file collections.

Note (Empty Set Semantics):

- When no snapshot files exist, this method MUST return a successful JSON-RPC response with an empty result object, not an error.
- Required structure for empty sets:
  {
    "jsonrpc": "2.0",
    "result": { "files": [], "total": 0, "limit": <int>, "offset": <int> },
    "id": <id>
  }
- Errors (e.g., directory inaccessible) MUST use documented JSON-RPC error codes with structured data.

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

- `files`: Array of snapshot file information objects (array)
  - `filename`: Snapshot filename (string)
  - `file_size`: File size in bytes (integer)
  - `modified_time`: File modification timestamp (ISO 8601 string)
  - `download_url`: HTTP download URL for the file (string)
- `total`: Total number of snapshot files (integer)
- `limit`: Maximum number of files requested (integer)
- `offset`: Number of files skipped for pagination (integer)

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

- `filename`: Recording filename without extension (string)
- `file_size`: File size in bytes (integer)
- `duration`: Recording duration in seconds (integer)
- `created_time`: File creation timestamp (ISO 8601 string)
- `download_url`: HTTP download URL for the file (string)

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

- `filename`: Snapshot filename (string)
- `file_size`: File size in bytes (integer)
- `created_time`: File creation timestamp (ISO 8601 string)
- `download_url`: HTTP download URL for the file (string)

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

- `filename`: Recording filename that was deleted (string)
- `deleted`: Whether deletion was successful (boolean)
- `message`: Deletion status message (string)

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

- `filename`: Snapshot filename that was deleted (string)
- `deleted`: Whether deletion was successful (boolean)
- `message`: Deletion status message (string)

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

- `total_space`: Total storage space in bytes (integer)
- `used_space`: Used storage space in bytes (integer)
- `available_space`: Available storage space in bytes (integer)
- `usage_percentage`: Storage usage percentage (0.0-100.0) (float64)
- `recordings_size`: Total size of recording files in bytes (integer)
- `snapshots_size`: Total size of snapshot files in bytes (integer)
- `low_space_warning`: Whether low space warning is active (boolean)

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

- `policy_type`: Type of retention policy ("age", "size", "manual") (string)
- `max_age_days`: Maximum age in days for age-based retention (integer)
- `max_size_gb`: Maximum size in GB for size-based retention (integer)
- `enabled`: Whether the retention policy is enabled (boolean)
- `message`: Policy configuration status message (string)

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

- `files`: Array of recording file information objects (array)
  - `filename`: Recording filename without extension (string)
  - `file_size`: File size in bytes (integer)
  - `modified_time`: File modification timestamp (ISO 8601 string)
  - `download_url`: HTTP download URL for the file (string)
- `total`: Total number of recording files (integer)
- `limit`: Maximum number of files requested (integer)
- `offset`: Number of files skipped for pagination (integer)

---

## System Status and Health

### get_status

Get system status and health information with comprehensive health monitoring and threshold-based status determination.

**Authentication:** Required (admin role)

**Parameters:** None

**Returns:** Object containing system status, component health, and operational state

**Status:** ✅ Implemented

**System Status Values:**

The system status is determined by comprehensive monitoring of multiple metrics and thresholds:

- **`"HEALTHY"`** - All systems operational within normal parameters
  - MediaMTX connectivity: API responding within 10s timeout
  - Memory usage: < 90% (configurable threshold)
  - Error rate: < 5% (configurable threshold)
  - Response time: < 1000ms average (configurable threshold)
  - Active connections: < 900 (configurable threshold)
  - Goroutines: < 1000 (configurable threshold)
  - Storage space: > 30% available (warn at 70%, block at 85%)
  - Health check failures: < 5 consecutive failures (configurable threshold)

- **`"DEGRADED"`** - System experiencing performance issues but core functionality available
  - MediaMTX connectivity: API responding but with delays (>5s response time)
  - Memory usage: 90-95% (approaching critical threshold)
  - Error rate: 5-10% (elevated error rate)
  - Response time: 1000-2000ms average (slow but acceptable)
  - Active connections: 900-950 (approaching limit)
  - Goroutines: 1000-1200 (elevated but manageable)
  - Storage space: 15-30% available (warning zone)
  - Health check failures: 3-4 consecutive failures (approaching threshold)

- **`"UNHEALTHY"`** - System experiencing critical failures impacting core functionality
  - MediaMTX connectivity: API not responding or >10s timeout
  - Memory usage: >95% (critical memory pressure)
  - Error rate: >10% (high failure rate)
  - Response time: >2000ms average (unacceptable delays)
  - Active connections: >950 (at or near limit)
  - Goroutines: >1200 (potential goroutine leak)
  - Storage space: <15% available (critical storage)
  - Health check failures: ≥5 consecutive failures (threshold exceeded)

**Component Status Values:**

Each component reports its operational state:

- **`"RUNNING"`** - Component operational and healthy
- **`"STOPPED"`** - Component intentionally stopped or disabled
- **`"ERROR"`** - Component experiencing errors or failures
- **`"STARTING"`** - Component in startup process
- **`"STOPPING"`** - Component in shutdown process

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
    "status": "HEALTHY",
    "uptime": 86400.5,
    "version": "1.0.0",
    "components": {
      "websocket_server": "RUNNING",
      "camera_monitor": "RUNNING",
      "mediamtx": "RUNNING"
    }
  },
  "id": 10
}
```

**Response Fields:**

- `status`: System health status ("HEALTHY", "DEGRADED", "UNHEALTHY") (string)
- `uptime`: System uptime in seconds with sub-second precision (float64)
- `version`: Service version string (string)
- `components`: Object containing component operational states (object)
  - `websocket_server`: WebSocket server status ("RUNNING", "STOPPED", "ERROR", "STARTING", "STOPPING") (string)
  - `camera_monitor`: Camera discovery monitor status ("RUNNING", "STOPPED", "ERROR", "STARTING", "STOPPING") (string)
  - `mediamtx`: MediaMTX service connectivity status ("RUNNING", "STOPPED", "ERROR", "STARTING", "STOPPING") (string)

### get_system_status

Get system readiness and initialization status for client applications.

**Authentication:** Required (viewer role)

**Parameters:** None

**Returns:** Object containing system readiness status, available cameras, and initialization progress

**Status:** ✅ Implemented

**Implementation:** Provides real-time system readiness information including camera discovery status, available cameras, and initialization progress. This method is designed for client applications to determine when the system is ready to accept requests and what resources are currently available.

**System Readiness Status Values:**

The system readiness status indicates the current initialization state:

- **`"starting"`** - System is initializing, components are starting up
  - Camera discovery is in progress
  - MediaMTX service is starting
  - No cameras are available yet
  - Client should wait before making camera-specific requests

- **`"partial"`** - System is partially ready, some components available
  - Some cameras have been discovered and are available
  - Camera discovery may still be in progress
  - Client can make requests for available cameras
  - System is functional but may still be discovering additional resources

- **`"ready"`** - System is fully operational
  - All components are initialized and running
  - Camera discovery is complete
  - All discovered cameras are available for use
  - Client can make full use of all system capabilities

**Example:**

```json
// Request
{
  "jsonrpc": "2.0",
  "method": "get_system_status",
  "id": 12
}

// Response (System Starting)
{
  "jsonrpc": "2.0",
  "result": {
    "status": "starting",
    "message": "System is initializing, please wait",
    "available_cameras": [],
    "discovery_active": true
  },
  "id": 12
}

// Response (System Partially Ready)
{
  "jsonrpc": "2.0",
  "result": {
    "status": "partial",
    "message": "Some cameras available, discovery in progress",
    "available_cameras": ["camera0"],
    "discovery_active": true
  },
  "id": 12
}

// Response (System Ready)
{
  "jsonrpc": "2.0",
  "result": {
    "status": "ready",
    "message": "System is fully operational",
    "available_cameras": ["camera0", "camera1"],
    "discovery_active": false
  },
  "id": 12
}
```

**Response Fields:**

- `status`: System readiness status ("starting", "partial", "ready") (string)
- `message`: Human-readable status message describing current state (string)
- `available_cameras`: Array of currently available camera device identifiers (array of strings)
- `discovery_active`: Whether camera discovery is currently in progress (boolean)

**Use Cases:**

- **Client Initialization**: Check system readiness before making camera requests
- **Progressive Loading**: Determine which cameras are available for immediate use
- **Status Monitoring**: Monitor system initialization progress in real-time
- **Error Prevention**: Avoid making requests for cameras that haven't been discovered yet

**Note:** This method provides different information than `get_status`:
- `get_system_status`: Real-time readiness and initialization state (viewer accessible)
- `get_status`: Comprehensive system health monitoring with metrics (admin only)

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

- `name`: Service name (string)
- `version`: Service version (string)
- `build_date`: Build date (string)
- `go_version`: Go version used (string)
- `architecture`: System architecture (string)
- `capabilities`: Array of supported capabilities (array of strings)
- `supported_formats`: Array of supported file formats (array of strings)
- `max_cameras`: Maximum number of supported cameras (integer)

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

- `subscribed`: Whether subscription was successful (boolean)
- `topics`: Array of successfully subscribed topics (array of strings)
- `filters`: Applied filters for event filtering (object)

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

- `unsubscribed`: Whether unsubscription was successful (boolean)
- `topics`: Array of unsubscribed topics (array of strings)

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

- `global_stats`: Global subscription statistics (object)
  - `total_subscriptions`: Total number of subscriptions (integer)
  - `active_clients`: Number of active clients (integer)
  - `topic_counts`: Count of subscriptions per topic (object)
- `client_topics`: Array of topics subscribed by current client (array of strings)
- `client_id`: Unique client identifier (string)

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

**Note:** When external discovery is disabled in configuration, this method returns a structured error response indicating that the feature is not available.

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
        "status": "DISCOVERED",
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
        "status": "DISCOVERED",
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
    "scan_timestamp": "2025-01-15T14:30:00Z",
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

**Error Response (External Discovery Disabled):**

```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32030,
    "message": "Unsupported",
    "data": {
      "reason": "feature_disabled",
      "details": "External stream discovery is disabled in configuration",
      "suggestion": "Enable external discovery in configuration"
    }
  },
  "id": 27
}
```

**Response Fields:**

- `discovered_streams`: Array of all discovered streams (array of objects)
  - `url`: Stream RTSP URL (string)
  - `type`: Stream type identifier (string)
  - `name`: Human-readable stream name (string)
  - `status`: Discovery status ("DISCOVERED", "ERROR") (string)
  - `discovered_at`: Discovery timestamp (ISO 8601 string)
  - `last_seen`: Last seen timestamp (ISO 8601 string)
  - `capabilities`: Stream capabilities object (object)
- `skydio_streams`: Array of Skydio-specific streams (array of objects)
- `generic_streams`: Array of generic RTSP streams (array of objects)
- `scan_timestamp`: Scan completion timestamp (ISO 8601 string)
- `total_found`: Total number of streams found (integer)
- `discovery_options`: Options used for discovery (object)
- `scan_duration`: Time taken for scan (string)
- `errors`: Array of scan errors (array of strings)

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
    "status": "ADDED",
    "timestamp": "2025-01-15T14:30:00Z"
  },
  "id": 28
}
```

**Response Fields:**

- `stream_url`: RTSP URL that was added (string)
- `stream_name`: Human-readable name assigned (string)
- `stream_type`: Type of stream added (string)
- `status`: Addition status ("ADDED", "ERROR") (string)
- `timestamp`: Addition timestamp (ISO 8601 string)

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
    "status": "REMOVED",
    "timestamp": "2025-01-15T14:30:00Z"
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
        "status": "DISCOVERED",
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
        "status": "DISCOVERED",
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
    "timestamp": "2025-01-15T14:30:00Z"
  },
  "id": 30
}
```

**Response Fields:**

- `external_streams`: Array of all external streams (array of objects)
  - `url`: Stream RTSP URL (string)
  - `type`: Stream type identifier (string)
  - `name`: Human-readable stream name (string)
  - `status`: Stream status ("DISCOVERED", "ERROR") (string)
  - `discovered_at`: Discovery timestamp (ISO 8601 string)
  - `last_seen`: Last seen timestamp (ISO 8601 string)
  - `capabilities`: Stream capabilities object (object)
- `skydio_streams`: Array of Skydio-specific streams (array of objects)
- `generic_streams`: Array of generic RTSP streams (array of objects)
- `total_count`: Total number of streams (integer)
- `timestamp`: Response timestamp (ISO 8601 string)

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
    "status": "UPDATED",
    "message": "Discovery interval updated (restart required for changes to take effect)",
    "timestamp": "2025-01-15T14:30:00Z"
  },
  "id": 31
}
```

**Response Fields:**

- `scan_interval`: Configured scan interval in seconds (integer)
- `status`: Configuration status ("UPDATED", "ERROR") (string)
- `message`: Status message (string)
- `timestamp`: Response timestamp (ISO 8601 string)

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

On any validation failure, the server MUST return `-32602 Invalid Params` with `error.data.rule` and `error.data.param`.

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

### Response Metadata (Optional)

Responses MAY include a top-level `"metadata"` object with:

* `processing_time_ms` (number)
* `server_timestamp` (ISO 8601 string)
* `request_id` (string, for tracing)

**Example:**

```json
{
  "jsonrpc": "2.0",
  "result": { "...": "..." },
  "id": 42,
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

- **Response Time:** <100ms for 95% of requests (high-performance Go implementation)
- **Concurrency:** 1000+ simultaneous WebSocket connections (10x improvement)
- **Memory Usage:** <60MB base footprint (50% reduction)
- **CPU Usage:** <50% sustained usage (30% reduction)

**API Compatibility:**

- **100% JSON-RPC Compatibility:** Identical protocol and message formats
- **Authentication:** Same JWT and API key mechanisms
- **Error Codes:** Identical error codes and response formats
- **Notifications:** Real-time event notifications with same payload structure

---

**API Version:** 1.0.0 (Go Implementation)  
**Last Updated:** 2025-01-15  
**Implementation Status:** All core methods, notifications, and event subscription system implemented and operational  
**Performance Status:** High-performance Go implementation achieved, 100x+ event system performance improvement
