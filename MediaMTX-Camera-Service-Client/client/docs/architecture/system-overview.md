# Unified System Architecture - MediaMTX Camera Service

## **Overview**

This document provides a comprehensive view of the MediaMTX Camera Service system architecture, showing the integration between the client and server components, data flow patterns, and communication protocols.

## **System Architecture Diagram**

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           MediaMTX Camera Service System                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────┐    WebSocket JSON-RPC    ┌─────────────────────────┐   │
│  │   Web Client    │◄────────────────────────►│   Camera Service        │   │
│  │   (React/TS)    │    ws://localhost:8002   │   (Python/FastAPI)      │   │
│  └─────────────────┘                          └─────────────────────────┘   │
│           │                                           │                     │
│           │ HTTP File Downloads                      │                     │
│           │ GET /files/recordings/                   │                     │
│           │ GET /files/snapshots/                    │                     │
│           │                                           │                     │
│  ┌─────────────────┐                          ┌─────────────────────────┐   │
│  │   User Browser  │                          │   Camera Discovery      │   │
│  │                 │                          │   Monitor (Python)      │   │
│  └─────────────────┘                          └─────────────────────────┘   │
│                                                           │                 │
│                                                   ┌─────────────────────────┐
│                                                   │   MediaMTX Server       │
│                                                   │   (Go)                  │
│                                                   │   Port: 8554 (RTSP)     │
│                                                   │   Port: 9997 (API)      │
│                                                   └─────────────────────────┘
│                                                           │                 │
│                                                   ┌─────────────────────────┐
│                                                   │   USB Camera Devices    │
│                                                   │   /dev/video0, /dev/video1 │
│                                                   └─────────────────────────┘
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## **Component Responsibilities**

### **Web Client (React/TypeScript)**
- **User Interface**: Camera management dashboard, recording controls, file browser
- **WebSocket Communication**: Real-time connection to server for live updates
- **File Downloads**: HTTP requests for recording and snapshot downloads
- **State Management**: Local state for UI responsiveness and offline capabilities
- **Error Handling**: User-friendly error messages and retry mechanisms

### **Camera Service (Python/FastAPI)**
- **WebSocket Server**: JSON-RPC 2.0 protocol implementation
- **Camera Discovery**: Integration with hybrid camera monitor
- **MediaMTX Control**: Stream management and recording operations
- **File Management**: Recording and snapshot file organization
- **HTTP Endpoints**: File download and health check endpoints
- **Real-time Notifications**: Event broadcasting to connected clients

### **Camera Discovery Monitor (Python)**
- **Device Monitoring**: USB camera hot-plug detection via udev
- **Capability Detection**: V4L2 capability parsing and format detection
- **Status Tracking**: Real-time camera connection status
- **Event Broadcasting**: Camera events to WebSocket clients

### **MediaMTX Server (Go)**
- **RTSP Streaming**: Real-time video streaming from USB cameras
- **Stream Management**: Camera stream lifecycle and routing
- **API Interface**: REST API for stream control and monitoring
- **Protocol Support**: RTSP, WebRTC, HLS streaming protocols

## **Communication Protocols**

### **WebSocket JSON-RPC 2.0**
**Endpoint**: `ws://localhost:8002/ws`

**Request Format**:
```json
{
  "jsonrpc": "2.0",
  "method": "get_camera_list",
  "params": {},
  "id": 1
}
```

**Response Format**:
```json
{
  "jsonrpc": "2.0",
  "result": { ... },
  "id": 1
}
```

**Notification Format**:
```json
{
  "jsonrpc": "2.0",
  "method": "camera_status_update",
  "params": { ... }
}
```

### **HTTP File Downloads**
**Recording Downloads**: `GET /files/recordings/{filename}`
**Snapshot Downloads**: `GET /files/snapshots/{filename}`

**Headers Required**:
- `Authorization: Bearer {jwt_token}` or `X-API-Key: {api_key}`

### **MediaMTX REST API**
**Endpoint**: `http://localhost:9997`
**Protocol**: HTTP REST API for stream management

## **Data Flow Patterns**

### **1. Camera Discovery Flow**
```
USB Camera Connected
        │
        ▼
┌─────────────────┐
│   udev Monitor  │
└─────────────────┘
        │
        ▼
┌─────────────────┐
│ Camera Discovery│
│   Monitor       │
└─────────────────┘
        │
        ▼
┌─────────────────┐
│ WebSocket       │
│ Notification    │
└─────────────────┘
        │
        ▼
┌─────────────────┐
│   Web Client    │
│   UI Update     │
└─────────────────┘
```

### **2. Recording Operation Flow**
```
User Initiates Recording
        │
        ▼
┌─────────────────┐
│   Web Client    │
│ JSON-RPC Call   │
└─────────────────┘
        │
        ▼
┌─────────────────┐
│ Camera Service  │
│ start_recording │
└─────────────────┘
        │
        ▼
┌─────────────────┐
│ MediaMTX API    │
│ Stream Creation │
└─────────────────┘
        │
        ▼
┌─────────────────┐
│ File System     │
│ Recording File  │
└─────────────────┘
        │
        ▼
┌─────────────────┐
│ WebSocket       │
│ Notification    │
└─────────────────┘
        │
        ▼
┌─────────────────┐
│   Web Client    │
│   UI Update     │
└─────────────────┘
```

### **3. Real-time Notification Flow**
```
Camera Event Occurs
        │
        ▼
┌─────────────────┐
│ Camera Discovery│
│   Monitor       │
└─────────────────┘
        │
        ▼
┌─────────────────┐
│ Event Processor │
└─────────────────┘
        │
        ▼
┌─────────────────┐
│ WebSocket       │
│ Broadcast       │
└─────────────────┘
        │
        ▼
┌─────────────────┐
│   Web Client    │
│ Event Handler   │
└─────────────────┘
        │
        ▼
┌─────────────────┐
│   UI Update     │
└─────────────────┘
```

## **Performance Requirements**

### **End-to-End Performance Targets**
- **Initial Page Load**: <3 seconds (client bundle + server connection)
- **WebSocket Connection**: <1 second establishment time
- **API Response Times**:
  - Status methods: <50ms (server guarantee)
  - Control methods: <100ms (server guarantee)
  - WebSocket notifications: <20ms (server guarantee)
- **File Download**: <5 seconds for 100MB recording
- **UI Responsiveness**: <100ms for user interactions

### **Client Performance Targets**
- **Bundle Size**: <2MB (gzipped)
- **Memory Usage**: <50MB sustained
- **CPU Usage**: <10% during normal operation
- **Network Efficiency**: Minimal polling, WebSocket-based updates

### **Server Performance Targets**
- **Concurrent Connections**: Support 10+ simultaneous clients
- **Camera Operations**: <100ms for all camera control operations
- **File Operations**: <200ms for file listing and metadata
- **Memory Usage**: <100MB per active camera stream

## **Error Handling and Recovery**

### **Connection Resilience**
```
WebSocket Disconnection
        │
        ▼
┌─────────────────┐
│ Auto-reconnect  │
│ (Exponential    │
│  Backoff)       │
└─────────────────┘
        │
        ▼
┌─────────────────┐
│ State Sync      │
│ (Re-fetch       │
│  camera list)   │
└─────────────────┘
        │
        ▼
┌─────────────────┐
│ UI Update       │
└─────────────────┘
```

### **Error Propagation**
```
API Error Occurs
        │
        ▼
┌─────────────────┐
│ Server Error    │
│ Response        │
└─────────────────┘
        │
        ▼
┌─────────────────┐
│ Client Error    │
│ Handler         │
└─────────────────┘
        │
        ▼
┌─────────────────┐
│ User Feedback   │
│ (Toast/Alert)   │
└─────────────────┘
```

## **Security Considerations**

### **Authentication**
- **JWT Tokens**: Bearer token authentication for file downloads
- **API Keys**: Alternative authentication method for automated clients
- **WebSocket**: Session-based authentication for real-time operations

### **Authorization**
- **File Access**: Token validation for recording and snapshot downloads
- **Camera Control**: Permission-based access to camera operations
- **API Methods**: Role-based access control for administrative functions

### **Data Protection**
- **File Storage**: Secure file system permissions for recordings
- **Network Security**: HTTPS/WSS for production deployments
- **Input Validation**: Comprehensive parameter validation on all endpoints

## **Deployment Architecture**

### **Development Environment**
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Client    │    │ Camera Service  │    │   MediaMTX      │
│   (localhost:3000)│    │ (localhost:8002) │    │ (localhost:8554) │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### **Production Environment**
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Load Balancer │    │ Camera Service  │    │   MediaMTX      │
│   (HTTPS/WSS)   │◄──►│   Cluster       │◄──►│   Cluster       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
        │                       │                       │
        ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Clients   │    │   File Storage  │    │   Camera Array  │
│   (Multiple)    │    │   (Network FS)  │    │   (USB/Network) │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## **Integration Patterns**

### **Real-time Communication**
- **WebSocket Connection**: Persistent connection for real-time updates
- **JSON-RPC Protocol**: Standardized request/response format
- **Event Broadcasting**: Server pushes notifications to all connected clients
- **Connection Management**: Auto-reconnect with exponential backoff

### **File Management**
- **Recording Storage**: Server manages recording files with metadata
- **Snapshot Storage**: Server manages snapshot files with timestamps
- **HTTP Downloads**: Secure file access with authentication
- **File Organization**: Structured directory layout for recordings/snapshots

### **Error Handling**
- **Graceful Degradation**: System continues operation with reduced functionality
- **Error Recovery**: Automatic retry mechanisms for transient failures
- **User Feedback**: Clear error messages and status indicators
- **Logging**: Comprehensive error logging for debugging

## **Monitoring and Observability**

### **Client Metrics**
- **Connection Status**: WebSocket connection health
- **API Performance**: Response times and error rates
- **User Interactions**: Feature usage and error patterns
- **Resource Usage**: Memory, CPU, and network consumption

### **Server Metrics**
- **Camera Status**: Connected cameras and stream health
- **API Performance**: Method response times and throughput
- **File Operations**: Storage usage and file access patterns
- **System Resources**: CPU, memory, and disk usage

### **Integration Metrics**
- **End-to-End Latency**: Complete operation timing
- **Error Correlation**: Client-server error relationship
- **User Experience**: Real user performance metrics
- **System Health**: Overall system availability and performance

---

**Document Version:** 1.0  
**Last Updated:** 2025-08-05  
**Status:** Approved for Implementation
