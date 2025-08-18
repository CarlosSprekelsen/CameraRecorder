# MediaMTX Camera Service - Client API Reference

## **Overview**

This document provides the API reference for the MediaMTX Camera Service Client, which communicates with the MediaMTX Camera Service via WebSocket JSON-RPC 2.0 protocol. For complete server API documentation, see the main server documentation.

## **API References**

### **Server Documentation**
- **Complete API Reference**: `docs/api/json-rpc-methods.md`
- **WebSocket Protocol**: `docs/api/websocket-protocol.md`
- **Error Codes**: `docs/api/error-codes.md`
- **Authentication**: `docs/api/authentication.md` (future)

## **Client-Specific API Usage**

### **WebSocket Connection**

#### **Connection Setup**
```typescript
// Connect to MediaMTX Camera Service WebSocket
const ws = new WebSocket('ws://localhost:8002/ws');

// Handle connection events
ws.onopen = () => {
  console.log('Connected to camera service');
};

ws.onclose = () => {
  console.log('Disconnected from camera service');
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};
```

#### **JSON-RPC Client Implementation**
```typescript
// JSON-RPC client for camera service
class CameraServiceClient {
  private ws: WebSocket;
  private requestId = 0;

  constructor(url: string) {
    this.ws = new WebSocket(url);
    this.setupEventHandlers();
  }

  private setupEventHandlers() {
    this.ws.onmessage = (event) => {
      const response = JSON.parse(event.data);
      this.handleResponse(response);
    };
  }

  async call(method: string, params: any = {}): Promise<any> {
    const request = {
      jsonrpc: '2.0',
      method,
      params,
      id: ++this.requestId
    };

    return new Promise((resolve, reject) => {
      this.ws.send(JSON.stringify(request));
      // Store promise resolvers for response handling
    });
  }
}
```

### **Core Camera Operations**

#### **Get Camera List**
```typescript
// Get all connected cameras
const cameraList = await client.call('get_camera_list', {});

// Response format
{
  "jsonrpc": "2.0",
  "result": {
    "cameras": [
      {
        "device": "/dev/video0",
        "name": "USB Camera",
        "status": "CONNECTED",
        "capabilities": {
          "resolution": "1920x1080",
          "fps": 30,
          "formats": ["YUYV", "MJPEG"]
        },
        "streams": {
          "rtsp": "rtsp://localhost:8554/camera0",
          "webrtc": "http://localhost:8889/camera0",
          "hls": "http://localhost:8888/camera0"
        }
      }
    ],
    "total": 1,
    "connected": 1
  },
  "id": 1
}
```

#### **Get Camera Status**
```typescript
// Get status of specific camera
const cameraStatus = await client.call('get_camera_status', {
  device: '/dev/video0'
});

// Response format
{
  "jsonrpc": "2.0",
  "result": {
    "device": "/dev/video0",
    "status": "CONNECTED",
    "name": "USB Camera",
    "capabilities": {
      "resolution": "1920x1080",
      "fps": 30,
      "validation_status": "confirmed",
      "formats": ["YUYV", "MJPEG"]
    },
    "streams": {
      "rtsp": "rtsp://localhost:8554/camera0",
      "webrtc": "http://localhost:8889/camera0",
      "hls": "http://localhost:8888/camera0"
    },
    "metrics": {
      "bytes_sent": 12345,
      "readers": 1,
      "uptime": 30
    }
  },
  "id": 2
}
```

### **Recording Operations**

#### **Start Recording**
```typescript
// Start recording for a camera
const recordingResult = await client.call('start_recording', {
  device: '/dev/video0',
  duration: 60,        // Optional: recording duration in seconds
  format: 'mp4'        // Optional: mp4, avi, mkv
});

// Response format
{
  "jsonrpc": "2.0",
  "result": {
    "success": true,
    "session_id": "uuid-12345",
    "file_path": "/opt/camera-service/recordings/camera0_20250127_143022.mp4",
    "duration": 60,
    "format": "mp4"
  },
  "id": 3
}
```

#### **Stop Recording**
```typescript
// Stop recording for a camera
const stopResult = await client.call('stop_recording', {
  device: '/dev/video0'
});

// Response format
{
  "jsonrpc": "2.0",
  "result": {
    "success": true,
    "session_id": "uuid-12345",
    "file_path": "/opt/camera-service/recordings/camera0_20250127_143022.mp4",
    "duration": 45,
    "format": "mp4"
  },
  "id": 4
}
```

### **Snapshot Operations**

#### **Take Snapshot**
```typescript
// Take a snapshot from a camera
const snapshotResult = await client.call('take_snapshot', {
  device: '/dev/video0',
  format: 'jpg',       // Optional: jpg, png
  quality: 85,         // Optional: 1-100
  filename: 'custom_name' // Optional: custom filename
});

// Response format
{
  "jsonrpc": "2.0",
  "result": {
    "success": true,
    "file_path": "/opt/camera-service/snapshots/camera0_20250127_143022.jpg",
    "format": "jpg",
    "quality": 85,
    "size": 245760
  },
  "id": 5
}
```

### **Utility Operations**

#### **Ping Server**
```typescript
// Test server connectivity
const pingResult = await client.call('ping', {});

// Response format
{
  "jsonrpc": "2.0",
  "result": "pong",
  "id": 6
}
```

#### **Get Server Info**
```typescript
// Get server information
const serverInfo = await client.call('get_server_info', {});

// Response format
{
  "jsonrpc": "2.0",
  "result": {
    "version": "1.0.0",
    "uptime": 3600,
    "cameras_connected": 2,
    "total_recordings": 15,
    "total_snapshots": 45
  },
  "id": 7
}
```

## **Real-time Notifications**

### **Camera Status Updates**
```typescript
// Listen for camera status updates
ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  
  if (message.method === 'camera_status_update') {
    const { device, status, capabilities, streams } = message.params;
    
    // Update UI with new camera status
    updateCameraStatus(device, status, capabilities, streams);
  }
};

// Example notification
{
  "jsonrpc": "2.0",
  "method": "camera_status_update",
  "params": {
    "device": "/dev/video0",
    "status": "CONNECTED",
    "capabilities": {
      "resolution": "1920x1080",
      "fps": 30,
      "validation_status": "confirmed"
    },
    "streams": {
      "rtsp": "rtsp://localhost:8554/camera0",
      "webrtc": "http://localhost:8889/camera0",
      "hls": "http://localhost:8888/camera0"
    }
  }
}
```

### **Recording Status Updates**
```typescript
// Listen for recording status updates
ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  
  if (message.method === 'recording_status_update') {
    const { device, session_id, status, progress } = message.params;
    
    // Update UI with recording status
    updateRecordingStatus(device, session_id, status, progress);
  }
};

// Example notification
{
  "jsonrpc": "2.0",
  "method": "recording_status_update",
  "params": {
    "device": "/dev/video0",
    "session_id": "uuid-12345",
    "status": "RECORDING",
    "progress": 45,  // seconds recorded
    "duration": 60   // total duration
  }
}
```

## **Error Handling**

### **Error Response Format**
```typescript
// Error response example
{
  "jsonrpc": "2.0",
  "error": {
    "code": -1000,
    "message": "Camera device not found"
  },
  "id": 8
}
```

### **Common Error Codes**
```typescript
// Error code constants
const ERROR_CODES = {
  CAMERA_NOT_FOUND: -1000,
  MEDIAMTX_ERROR: -1003,
  INVALID_PARAMS: -32602,
  INTERNAL_ERROR: -32603,
  PARSE_ERROR: -32700
};

// Error handling example
try {
  const result = await client.call('get_camera_status', {
    device: '/dev/video999'
  });
} catch (error) {
  if (error.code === ERROR_CODES.CAMERA_NOT_FOUND) {
    showError('Camera not found. Please check device path.');
  } else {
    showError(`Server error: ${error.message}`);
  }
}
```

## **TypeScript Type Definitions**

### **Core Types**
```typescript
// Camera-related types
interface CameraDevice {
  device: string;
  name: string;
  status: 'CONNECTED' | 'DISCONNECTED' | 'ERROR';
  capabilities?: CameraCapabilities;
  streams?: CameraStreams;
}

interface CameraCapabilities {
  resolution: string;
  fps: number;
  validation_status: 'provisional' | 'confirmed';
  formats: string[];
  all_resolutions?: string[];
}

interface CameraStreams {
  rtsp?: string;
  webrtc?: string;
  hls?: string;
}

// Recording types
interface RecordingRequest {
  device: string;
  duration?: number;
  format?: 'mp4' | 'avi' | 'mkv';
}

interface RecordingResponse {
  success: boolean;
  session_id: string;
  file_path: string;
  duration?: number;
  format: string;
}

// Snapshot types
interface SnapshotRequest {
  device: string;
  format?: 'jpg' | 'png';
  quality?: number;
  filename?: string;
}

interface SnapshotResponse {
  success: boolean;
  file_path: string;
  format: string;
  quality: number;
  size: number;
}

// JSON-RPC types
interface JsonRpcRequest {
  jsonrpc: '2.0';
  method: string;
  params?: any;
  id: number;
}

interface JsonRpcResponse {
  jsonrpc: '2.0';
  result?: any;
  error?: JsonRpcError;
  id: number;
}

interface JsonRpcError {
  code: number;
  message: string;
  data?: any;
}

// Notification types
interface CameraStatusNotification {
  jsonrpc: '2.0';
  method: 'camera_status_update';
  params: {
    device: string;
    status: string;
    capabilities?: CameraCapabilities;
    streams?: CameraStreams;
  };
}

interface RecordingStatusNotification {
  jsonrpc: '2.0';
  method: 'recording_status_update';
  params: {
    device: string;
    session_id: string;
    status: 'RECORDING' | 'STOPPED' | 'ERROR';
    progress?: number;
    duration?: number;
  };
}
```

## **Client Implementation Examples**

### **React Hook for Camera Operations**
```typescript
// useCamera.ts - Custom hook for camera operations
import { useState, useEffect } from 'react';

export function useCamera(device: string) {
  const [camera, setCamera] = useState<CameraDevice | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const getCameraStatus = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const result = await client.call('get_camera_status', { device });
      setCamera(result);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const takeSnapshot = async (format = 'jpg', quality = 85) => {
    try {
      const result = await client.call('take_snapshot', {
        device,
        format,
        quality
      });
      return result;
    } catch (err) {
      setError(err.message);
      throw err;
    }
  };

  const startRecording = async (duration = 60, format = 'mp4') => {
    try {
      const result = await client.call('start_recording', {
        device,
        duration,
        format
      });
      return result;
    } catch (err) {
      setError(err.message);
      throw err;
    }
  };

  const stopRecording = async () => {
    try {
      const result = await client.call('stop_recording', { device });
      return result;
    } catch (err) {
      setError(err.message);
      throw err;
    }
  };

  useEffect(() => {
    getCameraStatus();
  }, [device]);

  return {
    camera,
    loading,
    error,
    getCameraStatus,
    takeSnapshot,
    startRecording,
    stopRecording
  };
}
```

### **WebSocket Connection Hook**
```typescript
// useWebSocket.ts - Custom hook for WebSocket connection
import { useState, useEffect, useCallback } from 'react';

export function useWebSocket(url: string) {
  const [ws, setWs] = useState<WebSocket | null>(null);
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const connect = useCallback(() => {
    const websocket = new WebSocket(url);
    
    websocket.onopen = () => {
      setConnected(true);
      setError(null);
    };
    
    websocket.onclose = () => {
      setConnected(false);
    };
    
    websocket.onerror = (event) => {
      setError('WebSocket connection error');
    };
    
    setWs(websocket);
  }, [url]);

  const disconnect = useCallback(() => {
    if (ws) {
      ws.close();
      setWs(null);
    }
  }, [ws]);

  const send = useCallback((data: any) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify(data));
    }
  }, [ws]);

  useEffect(() => {
    connect();
    return () => disconnect();
  }, [connect, disconnect]);

  return {
    ws,
    connected,
    error,
    send,
    connect,
    disconnect
  };
}
```

## **Configuration**

### **Client Configuration**
```typescript
// Client configuration interface
interface ClientConfig {
  serverUrl: string;
  reconnectInterval: number;
  maxReconnectAttempts: number;
  pollingInterval: number;
  requestTimeout: number;
}

// Default configuration
const defaultConfig: ClientConfig = {
  serverUrl: 'ws://localhost:8002/ws',
  reconnectInterval: 5000,
  maxReconnectAttempts: 10,
  pollingInterval: 5000,
  requestTimeout: 10000
};
```

### **Environment Variables**
```typescript
// Environment-based configuration
const config: ClientConfig = {
  serverUrl: process.env.REACT_APP_SERVER_URL || 'ws://localhost:8002/ws',
  reconnectInterval: parseInt(process.env.REACT_APP_RECONNECT_INTERVAL || '5000'),
  maxReconnectAttempts: parseInt(process.env.REACT_APP_MAX_RECONNECT_ATTEMPTS || '10'),
  pollingInterval: parseInt(process.env.REACT_APP_POLLING_INTERVAL || '5000'),
  requestTimeout: parseInt(process.env.REACT_APP_REQUEST_TIMEOUT || '10000')
};
```

## **Testing Examples**

### **Mock Service Worker Setup**
```typescript
// mocks/handlers.ts - MSW handlers for testing
import { rest } from 'msw';

export const handlers = [
  // Mock WebSocket connection
  rest.get('/ws', (req, res, ctx) => {
    return res(
      ctx.status(101),
      ctx.set('Upgrade', 'websocket'),
      ctx.set('Connection', 'Upgrade')
    );
  }),

  // Mock camera list response
  rest.post('/api/rpc', (req, res, ctx) => {
    const { method } = req.body;
    
    if (method === 'get_camera_list') {
      return res(
        ctx.json({
          jsonrpc: '2.0',
          result: {
            cameras: [
              {
                device: '/dev/video0',
                name: 'Test Camera',
                status: 'CONNECTED',
                capabilities: {
                  resolution: '1920x1080',
                  fps: 30,
                  formats: ['YUYV', 'MJPEG']
                }
              }
            ],
            total: 1,
            connected: 1
          },
          id: 1
        })
      );
    }
  })
];
```

### **Component Testing**
```typescript
// CameraCard.test.tsx - Component test example
import { render, screen, fireEvent } from '@testing-library/react';
import { CameraCard } from './CameraCard';

describe('CameraCard', () => {
  const mockCamera = {
    device: '/dev/video0',
    name: 'Test Camera',
    status: 'CONNECTED',
    capabilities: {
      resolution: '1920x1080',
      fps: 30
    }
  };

  it('renders camera information', () => {
    render(<CameraCard camera={mockCamera} />);
    
    expect(screen.getByText('Test Camera')).toBeInTheDocument();
    expect(screen.getByText('CONNECTED')).toBeInTheDocument();
    expect(screen.getByText('1920x1080')).toBeInTheDocument();
  });

  it('calls takeSnapshot when snapshot button is clicked', () => {
    const mockTakeSnapshot = jest.fn();
    render(<CameraCard camera={mockCamera} onTakeSnapshot={mockTakeSnapshot} />);
    
    fireEvent.click(screen.getByText('Take Snapshot'));
    expect(mockTakeSnapshot).toHaveBeenCalledWith('/dev/video0');
  });
});
```

---

**API Reference**: Complete  
**Status**: Ready for Implementation  
**References**: Links to server documentation  
**Examples**: Client-specific usage patterns provided 