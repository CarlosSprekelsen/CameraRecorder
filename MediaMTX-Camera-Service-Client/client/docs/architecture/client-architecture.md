# MediaMTX Camera Service - Client Architecture

## **Project Overview**

### **Purpose**
The MediaMTX Camera Service Client is a React/TypeScript Progressive Web App (PWA) that provides a modern web interface for managing USB cameras, monitoring their status, and controlling recording/snapshot operations through the MediaMTX Camera Service.

### **Service Architecture Integration**
The client integrates with the MediaMTX Camera Service which provides **two distinct endpoints**:

1. **WebSocket JSON-RPC Endpoint** (`ws://localhost:8002/ws`) - Primary API for camera operations
2. **HTTP Health Endpoints** (`http://localhost:8003`) - System monitoring and health checks

### **Project Objectives**
1. **Real-time Camera Management**: Provide instant visibility into camera status and capabilities
2. **Recording Control**: Enable snapshot capture and recording start/stop operations
3. **Mobile-First Design**: Responsive PWA that works seamlessly on smartphones and desktops
4. **Intuitive UX**: Clean, modern interface that requires minimal training
5. **Reliable Communication**: WebSocket-based real-time updates with polling fallback
6. **System Health Monitoring**: Integration with health endpoints for operational visibility

### **Target Users**
- **System Administrators**: Monitor camera health and manage recordings
- **Security Personnel**: Quick access to camera status and snapshot capture
- **Mobile Users**: Responsive interface for on-the-go camera management
- **Developers**: API integration reference and testing interface
- **DevOps Teams**: System health monitoring and operational status

## **Architecture Overview**

### **High-Level Architecture**

```
┌────────────────────────────────────────────────────────────┐
│                    React PWA Client                        │
├─────────────────────────────────────────────────────────────┤
│                WebSocket JSON-RPC Client                   │
│     • Real-time notifications (camera_status_update)       │
│     • RPC method calls (get_camera_list, take_snapshot)   │
│     • Automatic reconnection and error handling            │
│     • Polling fallback for missed notifications           │
│     • JWT authentication and role-based access control     │
├─────────────────────────────────────────────────────────────┤
│                HTTP Health Client                          │
│     • System health monitoring (/health/system)           │
│     • Camera system health (/health/cameras)              │
│     • MediaMTX integration health (/health/mediamtx)      │
│     • Kubernetes readiness probes (/health/ready)         │
├─────────────────────────────────────────────────────────────┤
│                React Component Architecture                 │
│     • Dashboard (camera grid, status overview)            │
│     • Camera Detail (capabilities, controls, history)     │
│     • Settings (server configuration, PWA settings)       │
│     • Notifications (real-time status updates)            │
│     • Health Monitor (system status, component health)    │
├─────────────────────────────────────────────────────────────┤
│                State Management (Zustand)                  │
│     • Camera state (connected devices, status, metadata)  │
│     • UI state (selected camera, view mode, settings)     │
│     • Connection state (WebSocket status, error handling)  │
│     • Health state (system status, component health)      │
└─────────────────────┬───────────────────────────────────────┘
                      │ Dual-Endpoint Service Integration
┌─────────────────────▼───────────────────────────────────────┐
│              MediaMTX Camera Service                        │
│                (Backend Server)                            │
├─────────────────────────────────────────────────────────────┤
│  WebSocket JSON-RPC Server (Port 8002)                     │
│  • Camera operations and control                           │
│  • Real-time notifications                                 │
│  • File management                                         │
│  • Authentication and authorization                        │
├─────────────────────────────────────────────────────────────┤
│  HTTP Health Server (Port 8003)                            │
│  • System health monitoring                                │
│  • Component status checks                                 │
│  • Kubernetes readiness/liveness probes                    │
│  • Operational metrics                                     │
└─────────────────────────────────────────────────────────────┘
```

### **Technology Stack**

#### **Frontend Framework**
- **React 18+**: Modern React with hooks and concurrent features
- **TypeScript**: Type safety and better developer experience
- **Vite**: Fast build tool and development server
- **PWA Support**: Service workers and offline capabilities

#### **Service Integration**
- **WebSocket Service**: JSON-RPC 2.0 client with connection management
- **HTTP Health Client**: REST client for system monitoring
- **Authentication Service**: JWT token management and role-based access
- **File Download Service**: HTTP client for media file downloads

#### **UI Framework**
- **Material-UI (MUI)**: Comprehensive component library
- **Emotion**: CSS-in-JS for styling
- **React Router**: Client-side routing

#### **State Management**
- **Zustand**: Lightweight state management with persistence
- **React Query**: Server state management and caching
- **Connection State**: WebSocket connection status and health monitoring
- **Authentication State**: JWT token management and role-based access

#### **Communication**
- **WebSocket**: Real-time bidirectional communication with JSON-RPC 2.0
- **HTTP Health**: REST endpoints for system monitoring and health checks
- **JSON-RPC 2.0**: Standard protocol with 2-parameter method calls (method, params)
- **HTTP Client**: Health endpoint monitoring and file downloads

### **WebSocket Service Interface Requirements**

The client must implement a WebSocket service with these exact method signatures to align with the server ground truth:

```typescript
interface WebSocketService {
  // Connection Management (REQUIRED)
  connect(): Promise<void>;           // Establish WebSocket connection
  disconnect(): Promise<void>;        // Close WebSocket connection
  isConnected(): boolean;             // Check connection status
  
  // JSON-RPC Communication (REQUIRED)
  call(method: string, params?: Record<string, unknown>): Promise<unknown>;
  
  // Event Handlers (REQUIRED)
  onConnect(handler: () => void): void;
  onDisconnect(handler: () => void): void;
  onError(handler: (error: Error) => void): void;
  onMessage(handler: (message: unknown) => void): void;
}
```

**Critical Implementation Notes:**
- All JSON-RPC calls must use exactly 2 parameters: `method` and optional `params`
- Do not use 3-parameter calls (method, params, true) - this is incorrect
- Connection management methods are required for proper lifecycle handling
- **Polling Fallback**: Backup for missed WebSocket messages

#### **Testing**
- **Jest**: Unit testing framework
- **React Testing Library**: Component testing
- **MSW (Mock Service Worker)**: API mocking
- **Cypress**: End-to-end testing

## **Component Architecture**

### **Core Components**

#### **1. App Shell**
```typescript
// App.tsx - Main application shell
- Header (navigation, connection status, health indicators)
- Sidebar (camera list, settings, health monitor)
- Main Content Area (dashboard, detail views)
- Notification System (real-time updates, health alerts)
```

#### **2. Dashboard**
```typescript
// Dashboard.tsx - Camera overview
- Camera Grid (status cards for all cameras)
- Quick Actions (snapshot, record buttons)
- Status Summary (connected/disconnected counts)
- Real-time Updates (WebSocket notifications)
- System Health Overview (health endpoint integration)
```

#### **3. Health Monitor**
```typescript
// HealthMonitor.tsx - System health monitoring
- System Status Overview (/health/system)
- Camera System Health (/health/cameras)
- MediaMTX Integration Health (/health/mediamtx)
- Kubernetes Readiness Status (/health/ready)
- Health History and Trends
```

#### **4. Camera Detail**
```typescript
// CameraDetail.tsx - Individual camera management
- Camera Status Display (real-time from WebSocket)
- Camera Controls (snapshot, recording)
- Stream Information (RTSP, WebRTC, HLS URLs)
- Recording History and Management
- Camera Capabilities and Metrics
```

#### **5. File Manager**
```typescript
// FileManager.tsx - Media file management
- Recording Files List (with pagination)
- Snapshot Files List (with pagination)
- File Metadata Display
- Download Functionality
- File Deletion (with role-based permissions)
```

## **Service Integration Architecture**

### **WebSocket JSON-RPC Integration**

#### **Connection Management**
```typescript
interface WebSocketManager {
  // Connection lifecycle
  connect(): Promise<void>;
  disconnect(): void;
  reconnect(): Promise<void>;
  
  // Authentication
  authenticate(token: string): Promise<AuthResult>;
  
  // RPC calls
  call(method: string, params?: any): Promise<any>;
  
  // Event handling
  onNotification(callback: (notification: any) => void): void;
  onConnectionChange(callback: (status: ConnectionStatus) => void): void;
}
```

#### **Authentication Flow**
```typescript
interface AuthManager {
  // Token management
  setToken(token: string): void;
  getToken(): string | null;
  clearToken(): void;
  
  // Role-based access
  hasPermission(operation: string): boolean;
  getUserRole(): UserRole;
  
  // Session management
  refreshToken(): Promise<void>;
  validateSession(): Promise<boolean>;
}
```

### **HTTP Health Integration**

#### **Health Monitoring**
```typescript
interface HealthMonitor {
  // Health checks
  getSystemHealth(): Promise<SystemHealth>;
  getCameraHealth(): Promise<CameraHealth>;
  getMediaMTXHealth(): Promise<MediaMTXHealth>;
  getReadinessStatus(): Promise<ReadinessStatus>;
  
  // Monitoring
  startHealthPolling(interval: number): void;
  stopHealthPolling(): void;
  
  // Event handling
  onHealthChange(callback: (health: SystemHealth) => void): void;
}
```

## **State Management Architecture**

### **Store Structure**
```typescript
interface AppState {
  // Connection state
  connection: {
    websocket: ConnectionStatus;
    health: ConnectionStatus;
    lastError?: string;
  };
  
  // Authentication state
  auth: {
    token: string | null;
    role: UserRole | null;
    isAuthenticated: boolean;
    permissions: string[];
  };
  
  // Camera state
  cameras: {
    list: Camera[];
    selectedCamera: string | null;
    status: Record<string, CameraStatus>;
  };
  
  // Health state
  health: {
    system: SystemHealth | null;
    cameras: CameraHealth | null;
    mediamtx: MediaMTXHealth | null;
    readiness: ReadinessStatus | null;
    lastUpdate: Date | null;
  };
  
  // File state
  files: {
    recordings: FileInfo[];
    snapshots: FileInfo[];
    loading: boolean;
    pagination: PaginationState;
  };
  
  // UI state
  ui: {
    currentView: ViewType;
    notifications: Notification[];
    settings: UserSettings;
  };
}
```

## **Performance Architecture**

### **Performance Targets**
The client must meet these performance targets aligned with server capabilities:

- **Application Startup**: <3 seconds (includes service connection <1s)
- **Camera List Refresh**: <1 second (service API <50ms + UI rendering)
- **Photo Capture Response**: <2 seconds (service processing <100ms + file transfer)
- **Video Recording Start**: <2 seconds (service API <100ms + MediaMTX setup)
- **UI Interactions**: <200ms immediate feedback (excludes service calls)
- **Health Endpoint Responses**: <100ms response time

### **Optimization Strategies**

#### **Caching Strategy**
```typescript
interface CacheManager {
  // Camera data caching
  cacheCameraList(cameras: Camera[]): void;
  getCachedCameraList(): Camera[] | null;
  
  // Health data caching
  cacheHealthData(health: SystemHealth): void;
  getCachedHealthData(): SystemHealth | null;
  
  // File list caching
  cacheFileList(files: FileInfo[], type: 'recordings' | 'snapshots'): void;
  getCachedFileList(type: 'recordings' | 'snapshots'): FileInfo[] | null;
}
```

#### **Connection Optimization**
- **WebSocket**: Persistent connection with automatic reconnection
- **Health Polling**: Configurable intervals (default: 30 seconds)
- **Request Debouncing**: Prevent rapid successive API calls
- **Optimistic Updates**: Immediate UI feedback for better UX

## **Security Architecture**

### **Authentication & Authorization**
- **JWT Token Management**: Secure token storage and refresh
- **Role-Based Access Control**: UI adaptation based on user permissions
- **Session Management**: Automatic session validation and renewal
- **Secure Communication**: HTTPS/WSS in production environments

### **Data Protection**
- **Input Validation**: Client-side validation before server submission
- **Output Sanitization**: Safe rendering of server responses
- **Error Handling**: Secure error messages without information disclosure
- **Token Security**: Secure storage and transmission of authentication tokens

## **Error Handling Architecture**

### **Error Categories**
1. **Connection Errors**: WebSocket/HTTP connection failures
2. **Authentication Errors**: Token expiration, invalid credentials
3. **Authorization Errors**: Insufficient permissions for operations
4. **Service Errors**: Server-side errors and exceptions
5. **UI Errors**: Component rendering and interaction errors

### **Error Recovery Strategies**
- **Automatic Retry**: Exponential backoff for transient failures
- **Graceful Degradation**: Fallback to cached data when services unavailable
- **User Feedback**: Clear error messages with recovery guidance
- **Logging**: Comprehensive error logging for debugging

## **Testing Architecture**

### **Test Strategy**
- **Unit Tests**: Individual component and utility function testing
- **Integration Tests**: Service API communication testing
- **E2E Tests**: Complete user workflow validation
- **Performance Tests**: Response time and resource usage validation
- **Security Tests**: Authentication and authorization validation

### **Test Coverage Targets**
- **Unit Test Coverage**: >90% code coverage
- **Integration Test Coverage**: All API endpoints and workflows
- **E2E Test Coverage**: Critical user journeys
- **Performance Test Coverage**: All performance targets validated

## **Deployment Architecture**

### **Build Configuration**
- **Development**: Hot reload with development server
- **Production**: Optimized build with service worker
- **PWA**: Progressive Web App capabilities for mobile installation

### **Environment Configuration**
- **Development**: Local service endpoints
- **Staging**: Staging service endpoints
- **Production**: Production service endpoints with HTTPS/WSS

## **Monitoring & Observability**

### **Client-Side Monitoring**
- **Performance Metrics**: Response times, error rates
- **User Experience**: Page load times, interaction responsiveness
- **Error Tracking**: Client-side error collection and reporting
- **Usage Analytics**: Feature usage and user behavior tracking

### **Health Integration**
- **System Health**: Real-time system status monitoring
- **Component Health**: Individual service component status
- **Operational Metrics**: Performance and availability metrics
- **Alerting**: Proactive notification of system issues

---

**⚠️ IMPORTANT**: This architecture is designed to integrate with the MediaMTX Camera Service. Always refer to the server API documentation for the most current service specifications and capabilities. 