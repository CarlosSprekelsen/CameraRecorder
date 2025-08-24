# MediaMTX Camera Service - Client Architecture

## **Project Overview**

### **Purpose**
The MediaMTX Camera Service Client is a React/TypeScript Progressive Web App (PWA) that provides a modern web interface for managing USB cameras, monitoring their status, and controlling recording/snapshot operations through the MediaMTX Camera Service.

### **Service Architecture Integration**
The client integrates with the MediaMTX Camera Service which provides **two distinct endpoints**:

1. **WebSocket JSON-RPC Endpoint** (`ws://localhost:8002/ws`) - Primary API for camera operations
2. **HTTP Health Endpoints** (`http://localhost:8003`) - System monitoring and health checks

**🚨 CRITICAL UPDATE:** The service now includes enhanced recording management capabilities with storage protection, conflict prevention, and real-time monitoring features.

### **Project Objectives**
1. **Real-time Camera Management**: Provide instant visibility into camera status and capabilities
2. **Enhanced Recording Control**: Enable snapshot capture and recording start/stop operations with conflict prevention
3. **Storage Protection**: Monitor storage usage and prevent system resource exhaustion
4. **Mobile-First Design**: Responsive PWA that works seamlessly on smartphones and desktops
5. **Intuitive UX**: Clean, modern interface that requires minimal training
6. **Reliable Communication**: WebSocket-based real-time updates with polling fallback
7. **System Health Monitoring**: Integration with health endpoints for operational visibility
8. **Configuration Management**: Dynamic configuration with environment variable support

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
│     • Recording Manager (recording state, conflicts)      │
│     • Storage Monitor (storage usage, thresholds)         │
├─────────────────────────────────────────────────────────────┤
│                State Management (Zustand)                  │
│     • Camera state (connected devices, status, metadata)  │
│     • UI state (selected camera, view mode, settings)     │
│     • Connection state (WebSocket status, error handling)  │
│     • Health state (system status, component health)      │
│     • Recording state (sessions, conflicts, progress)     │
│     • Storage state (usage, thresholds, warnings)         │
│     • Configuration state (settings, environment vars)    │
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
- **Recording State Manager**: Recording session management and conflict prevention
- **Storage Monitor**: Storage usage monitoring and threshold management
- **Configuration Manager**: Environment variable and dynamic configuration management

#### **UI Framework**
- **Material-UI (MUI)**: Comprehensive component library
- **Emotion**: CSS-in-JS for styling
- **React Router**: Client-side routing

#### **State Management**
- **Zustand**: Lightweight state management with persistence
- **React Query**: Server state management and caching
- **Connection State**: WebSocket connection status and health monitoring
- **Authentication State**: JWT token management and role-based access
- **Recording State**: Recording sessions, conflicts, and progress tracking
- **Storage State**: Storage usage, thresholds, and warning management
- **Configuration State**: Environment variables and dynamic settings

#### **Communication**
- **WebSocket**: Real-time bidirectional communication with JSON-RPC 2.0
- **HTTP Health**: REST endpoints for system monitoring and health checks
- **JSON-RPC 2.0**: Standard protocol with 2-parameter method calls (method, params)
- **HTTP Client**: Health endpoint monitoring and file downloads
- **Enhanced Error Codes**: Support for -1006 (recording conflict), -1008 (storage low), -1010 (storage critical)
- **Real-time Notifications**: Enhanced recording status and storage monitoring notifications

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

## **Enhanced Recording Management Architecture (NEW)**

### **Recording State Management**
The client implements comprehensive recording state management to support the new ground truth requirements:

```typescript
interface RecordingStateManager {
  // Session tracking
  activeRecordings: Map<string, RecordingSession>;
  
  // State management
  startRecording(device: string): Promise<void>;
  stopRecording(device: string): Promise<void>;
  isRecording(device: string): boolean;
  getRecordingSession(device: string): RecordingSession | null;
  
  // Conflict prevention
  canStartRecording(device: string): boolean;
  validateRecordingRequest(device: string): ValidationResult;
  
  // Real-time updates
  onRecordingStatusChange(callback: (status: RecordingStatus) => void): void;
  onRecordingConflict(callback: (conflict: RecordingConflict) => void): void;
}
```

### **Storage Monitoring Architecture**
The client implements real-time storage monitoring with configurable thresholds:

```typescript
interface StorageMonitor {
  // Storage information
  getStorageInfo(): Promise<StorageInfo>;
  getStorageUsage(): Promise<StorageUsage>;
  
  // Threshold management
  checkStorageThresholds(): Promise<ThresholdStatus>;
  isStorageAvailable(): Promise<boolean>;
  
  // Validation
  validateStorageForRecording(): Promise<ValidationResult>;
  validateStorageForOperation(operation: string): Promise<ValidationResult>;
  
  // Monitoring
  startStorageMonitoring(interval: number): void;
  stopStorageMonitoring(): void;
  
  // Event handling
  onStorageThresholdExceeded(callback: (threshold: ThresholdStatus) => void): void;
  onStorageCritical(callback: (status: StorageStatus) => void): void;
}
```

### **Configuration Management Architecture**
The client supports dynamic configuration management with environment variable support:

```typescript
interface ConfigurationManager {
  // Environment variables
  getRecordingRotationMinutes(): number;
  getStorageWarnPercent(): number;
  getStorageBlockPercent(): number;
  
  // Configuration validation
  validateConfiguration(): ValidationResult;
  getConfigurationErrors(): string[];
  
  // Dynamic updates
  updateConfiguration(config: Partial<AppConfig>): void;
  reloadConfiguration(): Promise<void>;
  
  // Default values
  getDefaultConfiguration(): AppConfig;
}
```

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

#### **6. Recording Manager (NEW)**
```typescript
// RecordingManager.tsx - Enhanced recording management
- Recording State Display (active sessions, conflicts)
- Recording Progress Tracking (elapsed time, file info)
- Conflict Prevention Interface (disable controls for active recordings)
- Session Management (start, stop, pause operations)
- Real-time Status Updates (WebSocket notifications)
```

#### **7. Storage Monitor (NEW)**
```typescript
// StorageMonitor.tsx - Storage monitoring and management
- Storage Usage Display (total, used, available space)
- Threshold Status Indicators (warning, critical levels)
- Storage Warnings and Alerts (user-friendly messages)
- Storage Validation (pre-operation checks)
- Configuration Interface (threshold settings)
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

### **ErrorRecoveryService Architecture**
The ErrorRecoveryService follows a **pure utility service pattern**:

- **Design Pattern**: Dependency injection with function parameters
- **Separation of Concerns**: Service layer isolated from state management
- **No Circular Dependencies**: Service doesn't import stores or components
- **Usage Pattern**: Stores inject their operations into the service

**Example Usage:**
```typescript
// Store calls ErrorRecoveryService with operation function
const result = await errorRecoveryService.executeWithRetry(
  () => wsService.call('get_camera_list', {}),
  'get_camera_list'
);
```

**Key Benefits:**
- ✅ Testable: Easy to mock operations
- ✅ Reusable: Works with any async operation
- ✅ Type Safe: Full TypeScript support
- ✅ SOLID Compliant: Single responsibility and dependency inversion

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

## **Architecture Update Notes**

### **Ground Truth Alignment**
This architecture has been updated to align with the new recording management ground truth requirements. The following enhancements have been integrated:

1. **Enhanced Recording Management**: Comprehensive recording state management with conflict prevention
2. **Storage Protection**: Real-time storage monitoring with configurable thresholds
3. **Configuration Management**: Dynamic configuration with environment variable support
4. **Enhanced Error Handling**: Support for new error codes and user-friendly messages
5. **Real-time Notifications**: Enhanced WebSocket notifications for recording and storage status

### **Implementation Priority**
The new architecture components should be implemented in the following order:
1. **Phase 1**: Recording State Manager and Storage Monitor services
2. **Phase 2**: Enhanced Error Handling and Configuration Management
3. **Phase 3**: UI Components and integration

---

**⚠️ IMPORTANT**: This architecture is designed to integrate with the MediaMTX Camera Service. Always refer to the server API documentation for the most current service specifications and capabilities. 