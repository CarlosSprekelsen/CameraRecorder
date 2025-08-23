# Client API Reference

## **Ground Truth Reference**

**⚠️ IMPORTANT**: This client API reference is based on the authoritative server API documentation. For the complete, up-to-date API specification, always refer to the server documentation:

- **Primary API Reference**: [`mediamtx-camera-service/docs/api/json-rpc-methods.md`](../../../mediamtx-camera-service/docs/api/json-rpc-methods.md)
- **Health Endpoints**: [`mediamtx-camera-service/docs/api/health-endpoints.md`](../../../mediamtx-camera-service/docs/api/health-endpoints.md)

## **Service Architecture Overview**

The MediaMTX Camera Service provides **two distinct endpoints**:

### **1. WebSocket JSON-RPC Endpoint**
- **URL**: `ws://localhost:8002/ws`
- **Purpose**: Primary API for camera operations, real-time notifications
- **Protocol**: JSON-RPC 2.0 over WebSocket
- **Authentication**: JWT token-based with role-based access control

### **2. HTTP Health Endpoints**
- **URL**: `http://localhost:8003`
- **Purpose**: System health monitoring, Kubernetes probes
- **Protocol**: REST HTTP
- **Authentication**: None (monitoring endpoints)

## **Performance Targets**

The client application must meet these performance targets aligned with server capabilities:

- **Status methods**: <50ms response time
- **Control methods**: <100ms response time  
- **WebSocket notifications**: <20ms delivery latency
- **Health endpoint responses**: <100ms response time

## **Authentication & Authorization**

### **Authentication Flow**
1. **Establish WebSocket Connection**: Connect to `ws://localhost:8002/ws`
2. **Authenticate Session**: Call `authenticate` method with JWT token
3. **Role-Based Access**: Server validates permissions for each method
4. **Session Management**: Token remains valid for subsequent requests

### **Role-Based Access Control**
- **viewer**: Read-only access to camera status, file listings
- **operator**: Viewer permissions + camera control operations (snapshots, recording)
- **admin**: Full access including system metrics and configuration

### **Authentication Example**
```typescript
// 1. Connect to WebSocket
const ws = new WebSocket('ws://localhost:8002/ws');

// 2. Authenticate with JWT token
const authRequest = {
  jsonrpc: "2.0",
  method: "authenticate",
  params: {
    token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    auth_type: "jwt"
  },
  id: 1
};

// 3. Server validates and establishes session
// 4. Subsequent requests use the authenticated session
```

## **Core WebSocket JSON-RPC Methods**

### **Connection & Authentication**
- `ping()` - Health check, returns "pong"
- `authenticate(params: { token: string, auth_type?: string })` - Establish authenticated session

### **Camera Operations**
- `get_camera_list()` - List all discovered cameras
- `get_camera_status(params: { device: string })` - Get specific camera status
- `take_snapshot(params: { device: string, filename?: string })` - Capture still image
- `start_recording(params: { device: string, duration?: number, format?: string })` - Begin video recording
- `stop_recording(params: { device: string })` - End video recording

### **File Management**
- `list_recordings(params?: { limit?: number, offset?: number })` - List recording files
- `list_snapshots(params?: { limit?: number, offset?: number })` - List snapshot files
- `get_recording_info(params: { filename: string })` - Get recording metadata
- `get_snapshot_info(params: { filename: string })` - Get snapshot metadata
- `delete_recording(params: { filename: string })` - Delete recording file
- `delete_snapshot(params: { filename: string })` - Delete snapshot file

### **System Management (Admin Only)**
- `get_metrics()` - System performance metrics
- `get_status()` - System health status
- `get_server_info()` - Server configuration
- `get_storage_info()` - Storage space information
- `set_retention_policy(params: { policy_type: string, max_age_days?: number, max_size_gb?: number, enabled: boolean })` - Configure retention
- `cleanup_old_files()` - Manual cleanup trigger

## **HTTP File Download Endpoints**

### **Recording Downloads**
```
GET /files/recordings/{filename}
Headers: Authorization: Bearer {jwt_token}
```

### **Snapshot Downloads**
```
GET /files/snapshots/{filename}
Headers: Authorization: Bearer {jwt_token}
```

## **Health Endpoints (Monitoring)**

### **System Health**
```
GET http://localhost:8003/health/system
```

### **Camera System Health**
```
GET http://localhost:8003/health/cameras
```

### **MediaMTX Integration Health**
```
GET http://localhost:8003/health/mediamtx
```

### **Kubernetes Readiness Probe**
```
GET http://localhost:8003/health/ready
```

## **Real-Time Notifications**

### **Camera Status Updates**
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

### **Recording Status Updates**
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

## **Error Codes**

### **Standard JSON-RPC 2.0 Errors**
- `-32700`: Parse error (Invalid JSON)
- `-32600`: Invalid Request
- `-32601`: Method not found
- `-32602`: Invalid params
- `-32603`: Internal error

### **Service-Specific Errors**
- `-32001`: Authentication failed or token expired
- `-32002`: Rate limit exceeded
- `-32003`: Insufficient permissions (role-based access control)
- `-32004`: Camera not found or disconnected
- `-32005`: Recording already in progress
- `-32006`: MediaMTX service unavailable
- `-32007`: Insufficient storage space
- `-32008`: Camera capability not supported

## **Type Definitions**

```typescript
interface Camera {
  device: string;
  status: 'CONNECTED' | 'DISCONNECTED' | 'ERROR';
  name: string;
  resolution: string;
  fps: number;
  streams: {
    rtsp: string;
    webrtc: string;
    hls: string;
  };
  metrics?: {
    bytes_sent: number;
    readers: number;
    uptime: number;
  };
  capabilities?: {
    formats: string[];
    resolutions: string[];
  };
}

interface FileInfo {
  filename: string;
  file_size: number;
  modified_time: string;
  download_url: string;
}

interface RecordingInfo extends FileInfo {
  duration?: number;
  created_time: string;
}

interface SnapshotInfo extends FileInfo {
  resolution?: string;
  created_time: string;
}
```

## **WebSocket Service Interface (Ground Truth)**

### **Required WebSocket Service Methods**
The client must implement a WebSocket service with these exact method signatures:

```typescript
interface WebSocketService {
  // Connection Management
  connect(): Promise<void>;           // Establish WebSocket connection
  disconnect(): Promise<void>;        // Close WebSocket connection
  isConnected(): boolean;             // Check connection status
  
  // JSON-RPC Communication
  call(method: string, params?: Record<string, unknown>): Promise<unknown>;
  
  // Event Handlers
  onConnect(handler: () => void): void;
  onDisconnect(handler: () => void): void;
  onError(handler: (error: Error) => void): void;
  onMessage(handler: (message: unknown) => void): void;
}
```

### **JSON-RPC Call Pattern**
All method calls must follow this exact pattern:
```typescript
// ✅ CORRECT - Only 2 parameters
const result = await wsService.call('get_camera_list');
const result = await wsService.call('get_camera_status', { device: '/dev/video0' });

// ❌ WRONG - Do not use 3 parameters
const result = await wsService.call('get_camera_list', {}, true);
```

## **Client Implementation Guidelines**

### **Connection Management**
- Implement automatic WebSocket reconnection with exponential backoff
- Use polling fallback for missed notifications
- Handle connection state changes gracefully

### **Authentication Handling**
- Store JWT tokens securely (not in localStorage)
- Implement token refresh before expiration
- Handle authentication failures with user feedback

### **Error Handling**
- Map service error codes to user-friendly messages
- Implement retry logic for transient failures
- Provide clear recovery guidance for users

### **Performance Optimization**
- Cache camera list and status data
- Implement request debouncing for rapid operations
- Use optimistic updates for better UX

### **Security Considerations**
- Validate all server responses
- Sanitize user inputs before sending to server
- Implement proper session timeout handling
- Use HTTPS/WSS in production environments

---

**⚠️ IMPORTANT**: Always refer to the server API documentation for the most current and complete specification. This client reference is a summary and may not include all details or recent changes. 