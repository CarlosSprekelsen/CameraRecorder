# JSON-RPC API Reference

This document describes all available JSON-RPC 2.0 methods provided by the Camera Service.

## Connection

Connect to the WebSocket endpoint:
`
ws://localhost:8002/ws
`

## Core Methods

### ping
Health check method that returns "pong".

**Parameters:** None

**Returns:** "pong"

**Example:**
`json
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
`

### get_camera_list
Get list of all discovered cameras with their current status.

**Parameters:** None

**Returns:** Object with camera list and metadata

**Example:**
`json
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
`

## Camera Control Methods

### get_camera_status
Get detailed status for a specific camera.

**Parameters:**
- device (string): Camera device path (e.g., "/dev/video0")

**Returns:** Detailed camera status object

### take_snapshot  
Capture a snapshot from the specified camera.

**Parameters:**
- device (string): Camera device path
- ilename (string, optional): Custom filename

**Returns:** Snapshot information object

### start_recording
Start recording video from the specified camera.

**Parameters:**
- device (string): Camera device path  
- duration (number, optional): Recording duration in seconds
- ormat (string, optional): Recording format ("mp4", "mkv")

**Returns:** Recording session information

### stop_recording
Stop active recording for the specified camera.

**Parameters:**
- device (string): Camera device path

**Returns:** Recording completion information

## Notifications

The server sends real-time notifications for camera events.

### camera_status_update
Sent when a camera connects, disconnects, or changes status.

**Example:**
`json
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
`

### recording_status_update
Sent when recording starts, stops, or encounters an error.

**Example:**
`json
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
`

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
