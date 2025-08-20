# Client API Reference

## Performance Targets
- Status methods: <50ms
- Control methods: <100ms  
- WebSocket notifications: <20ms

## WebSocket Endpoint
```
ws://localhost:8002/ws
```

## Core Methods

### ping()
Returns: "pong"

### get_camera_list()
Returns: `{ cameras: Camera[], total: number, connected: number }`

### get_camera_status(device: string)
Returns: `Camera`

### take_snapshot(device: string, filename?: string)
Returns: `{ device, filename, status, timestamp, file_size, file_path }`

### start_recording(device: string, duration_seconds?: number, duration_minutes?: number, duration_hours?: number, format?: string)
Returns: `{ device, session_id, filename, status, start_time, duration, format }`

### stop_recording(device: string)
Returns: `{ device, session_id, filename, status, start_time, end_time, duration, file_size }`

### list_recordings(limit?: number, offset?: number)
Returns: `{ files: FileInfo[], total, limit, offset }`

### list_snapshots(limit?: number, offset?: number)
Returns: `{ files: FileInfo[], total, limit, offset }`

### authenticate(token: string)
Returns: `{ authenticated: boolean, role?: string }`

## HTTP Endpoints

### GET /files/recordings/{filename}
Headers: `Authorization: Bearer {token}`

### GET /files/snapshots/{filename}
Headers: `Authorization: Bearer {token}`

## Notifications

### camera_status_update
```json
{ "device", "status", "name", "resolution", "fps", "streams": { "rtsp", "webrtc", "hls" } }
```

### recording_status_update
```json
{ "device", "status", "filename", "duration" }
```

## Error Codes
- -32700: Parse error (Invalid JSON)
- -32600: Invalid Request  
- -32601: Method not found
- -32602: Invalid params
- -32603: Internal error
- -32001: Camera not found or disconnected
- -32002: Recording already in progress
- -32003: MediaMTX service unavailable  
- -32004: Authentication required or token expired
- -32005: Insufficient storage space
- -32006: Camera capability not supported

## Type Definitions
```typescript
interface Camera {
  device: string;
  status: 'CONNECTED' | 'DISCONNECTED' | 'ERROR';
  name: string;
  resolution: string;
  fps: number;
  streams: { rtsp: string; webrtc: string; hls: string; };
}

interface FileInfo {
  filename: string;
  file_size: number;
  modified_time: string;
  download_url: string;
}
```

---
**Rule**: Must match server implementation exactly. Server API is authoritative. 