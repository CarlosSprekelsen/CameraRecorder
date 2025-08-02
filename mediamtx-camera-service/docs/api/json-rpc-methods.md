# JSON-RPC API Reference

This document describes all available JSON-RPC 2.0 methods provided by the Camera Service.

## Connection

Connect to the WebSocket endpoint:
```
ws://localhost:8002/ws
```

## Core Methods

### ping
Health check method that returns "pong".

**Parameters:** None

**Returns:** "pong"

**Status:** Implemented

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

**Parameters:** None

**Returns:** Object with camera list and metadata

**Status:** Implemented

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
        "name": "USB Camera",
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

**Parameters:**
- device: string - Camera device path (required)

**Returns:** Camera status object with all standard fields and metrics

**Example:**
```json
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

**Parameters:**
- device: string - Camera device path (required)
- filename: string - Custom filename (optional)

**Returns:** Snapshot information object with filename, timestamp, and status

**Status:** Not yet implemented

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
    "file_size": 204800
  },
  "id": 4
}
```

### start_recording
Start recording video from the specified camera.

**Parameters:**
- device: string - Camera device path (required)
- duration: number - Recording duration in seconds (optional)
- format: string - Recording format ("mp4", "mkv") (optional)

**Returns:** Recording session information with filename, status, and metadata

**Status:** Not yet implemented

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
    "filename": "camera0_2025-01-15_14-30-00.mp4",
    "status": "started",
    "start_time": "2025-01-15T14:30:00Z",
    "duration": 3600,
    "format": "mp4"
  },
  "id": 5
}
```

### stop_recording
Stop active recording for the specified camera.

**Parameters:**
- device: string - Camera device path (required)

**Returns:** Recording completion information with final file details

**Status:** Not yet implemented

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
    "filename": "camera0_2025-01-15_14-30-00.mp4",
    "status": "completed",
    "duration": 1800,
    "file_size": 1073741824,
    "end_time": "2025-01-15T15:00:00Z"
  },
  "id": 6
}
```

## Notifications

The server sends real-time notifications for camera events.

### camera_status_update
Sent when a camera connects, disconnects, or changes status.

**Status:** Not yet implemented

**Example:**
```json
{
  "jsonrpc": "2.0",
  "method": "camera_status_update", 
  "params": {
    "device": "/dev/video0",
    "status": "CONNECTED",
    "name": "USB Camera",
    "resolution": "1920x1080", 
    "fps": 30,
    "streams": {
      "rtsp": "rtsp://localhost:8554/camera0",
      "webrtc": "http://localhost:8889/camera0/webrtc"
    }
  }
}
```

### recording_status_update
Sent when recording starts, stops, or encounters an error.

**Status:** Not yet implemented

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