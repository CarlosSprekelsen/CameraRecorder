# MediaMTX Camera Service Client Architecture

**Version:** 2.0  
**Last Updated:** 2025-01-16  
**Status:** 🚨 **CRITICAL UPDATE - ARCHITECTURE REFACTORED FOR SERVER API ALIGNMENT**

## **Project Objectives**

### **Primary Goals**
- **Real-time Camera Management**: WebSocket-based camera control and monitoring
- **Health System Integration**: REST health endpoints for system monitoring
- **Role-Based Access Control**: JWT authentication with viewer/operator/admin roles
- **File Management**: Recording and snapshot file operations
- **Enhanced Recording Management**: Real-time recording status and conflict handling
- **Storage Monitoring**: Real-time storage usage and threshold management
- **Enhanced Error Handling**: Comprehensive error management and recovery

### **Technical Requirements**
- **WebSocket JSON-RPC 2.0**: Real-time bidirectional communication
- **REST Health Endpoints**: System monitoring and health checks
- **JWT Authentication**: Secure role-based access control
- **Material-UI**: Modern, responsive user interface
- **TypeScript**: Type-safe development with strict typing
- **Zustand**: Lightweight state management
- **React Query**: Server state management and caching

## **Architecture Overview**

### **Core Services**
- **WebSocket Service**: JSON-RPC 2.0 client with connection management
- **HTTP Health Client**: REST client for system monitoring
- **Authentication Service**: JWT token management and role-based access
- **File Download Service**: HTTP client for media file downloads
- **Recording State Manager**: Recording session management and error handling
- **Storage Monitor**: Storage usage monitoring via server API
- **Error Handler**: Enhanced error handling and user feedback

#### **UI Framework**
- **Material-UI (MUI)**: Comprehensive component library
- **Emotion**: CSS-in-JS for styling
- **React Router**: Client-side routing

#### **State Management**
- **Zustand**: Lightweight state management with persistence
- **React Query**: Server state management and caching
- **Connection State**: WebSocket connection status and health monitoring
- **Authentication State**: JWT token management and role-based access
- **Recording State**: Recording sessions and error tracking
- **Storage State**: Storage usage from server API
- **Error State**: Error tracking and user feedback

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

## **Enhanced Recording Management Architecture**

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
  
  // Error handling (no conflict objects - use error responses)
  handleRecordingError(error: JSONRPCError): void;
  validateRecordingRequest(device: string): ValidationResult;
  
  // Real-time updates
  onRecordingStatusChange(callback: (status: RecordingStatus) => void): void;
  onRecordingError(callback: (error: JSONRPCError) => void): void;
}
```

### **Storage Monitoring Architecture**
The client implements storage monitoring using only the server API:

```typescript
interface StorageMonitor {
  // Storage information (server API only)
  getStorageInfo(): Promise<StorageInfo>;
  
  // Validation
  validateStorageForRecording(): Promise<ValidationResult>;
  validateStorageForOperation(operation: string): Promise<ValidationResult>;
  
  // Event handling
  onStorageThresholdExceeded(callback: (threshold: ThresholdStatus) => void): void;
  onStorageCritical(callback: (status: StorageStatus) => void): void;
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

#### **6. Recording Manager**
```typescript
// RecordingManager.tsx - Enhanced recording management
- Recording State Display (active sessions)
- Recording Progress Tracking (elapsed time, file info)
- Error Display (JSON-RPC error responses)
- Session Management (start, stop operations)
- Real-time Status Updates (WebSocket notifications)
```

#### **7. Storage Monitor**
```typescript
// StorageMonitor.tsx - Storage monitoring and management
- Storage Usage Display (from server API)
- Threshold Status Indicators (warning, critical levels)
- Storage Warnings and Alerts (user-friendly messages)
- Storage Validation (pre-operation checks)
```

#### **8. Error Handler**
```typescript
// ErrorHandler.tsx - Enhanced error handling and recovery
- Error Display (user-friendly error messages)
- Error Logging (detailed error information)
- Error Prevention (conflict detection and resolution)
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
  // Health checks (server endpoints)
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
  
  // Recording state
  recording: {
    activeSessions: Map<string, RecordingSession>;
    errors: Map<string, JSONRPCError>;
    progress: Map<string, RecordingProgress>;
    isMonitoring: boolean;
  };
  
  // Storage state (server API only)
  storage: {
    info: StorageInfo | null;
    thresholdStatus: ThresholdStatus;
    warnings: string[];
    lastUpdate: Date | null;
  };
  
  // Error state
  errors: {
    currentErrors: ErrorInfo[];
    errorHistory: ErrorInfo[];
    isRecovering: boolean;
  };
}
```

## **API Integration Requirements**

### **JSON-RPC Methods (Server API Alignment)**
The client must support all server JSON-RPC methods:

```typescript
interface JSONRPCMethods {
  // Authentication
  authenticate(auth_token: string): Promise<AuthResult>;
  
  // Core methods
  ping(): Promise<string>;
  get_camera_list(): Promise<CameraListResponse>;
  get_camera_status(device: string): Promise<CameraStatusResponse>;
  
  // Control methods
  take_snapshot(device: string): Promise<SnapshotResponse>;
  start_recording(device: string): Promise<RecordingResponse>;
  stop_recording(device: string): Promise<RecordingResponse>;
  
  // File management
  list_recordings(device?: string, limit?: number, offset?: number): Promise<RecordingsResponse>;
  list_snapshots(device?: string, limit?: number, offset?: number): Promise<SnapshotsResponse>;
  get_recording_info(filename: string): Promise<RecordingInfoResponse>;
  get_snapshot_info(filename: string): Promise<SnapshotInfoResponse>;
  delete_recording(filename: string): Promise<DeleteResponse>;
  delete_snapshot(filename: string): Promise<DeleteResponse>;
  
  // Storage management
  get_storage_info(): Promise<StorageInfoResponse>;
  set_retention_policy(policy: RetentionPolicy): Promise<PolicyResponse>;
  cleanup_old_files(): Promise<CleanupResponse>;
  
  // System information
  get_metrics(): Promise<MetricsResponse>;
  get_status(): Promise<StatusResponse>;
  get_server_info(): Promise<ServerInfoResponse>;
  get_streams(): Promise<StreamsResponse>;
}
```

### **Health Endpoints (Server API Alignment)**
The client must support all server health endpoints:

```typescript
interface HealthEndpoints {
  // Health checks
  getSystemHealth(): Promise<SystemHealth>;
  getCameraHealth(): Promise<CameraHealth>;
  getMediaMTXHealth(): Promise<MediaMTXHealth>;
  getReadinessStatus(): Promise<ReadinessStatus>;
}
```

### **File Download Endpoints (Server API Alignment)**
The client must support file download endpoints:

```typescript
interface FileEndpoints {
  // File downloads
  downloadRecording(filename: string): Promise<Blob>;
  downloadSnapshot(filename: string): Promise<Blob>;
}
```

## **Error Handling Architecture**

### **JSON-RPC Error Codes**
The client must handle all server error codes:

```typescript
interface ErrorCodes {
  // Authentication errors
  -32001: "Authentication failed";
  -32002: "Insufficient permissions";
  
  // Recording errors
  -1006: "Recording conflict";
  -1007: "Recording failed";
  
  // Storage errors
  -1008: "Storage low";
  -1009: "Storage full";
  -1010: "Storage critical";
  
  // System errors
  -32600: "Invalid Request";
  -32601: "Method not found";
  -32602: "Invalid params";
  -32603: "Internal error";
}
```

### **Error Handling Strategy**
```typescript
interface ErrorHandler {
  // Error processing
  handleJSONRPCError(error: JSONRPCError): void;
  handleError(error: Error): void;
  
  // User feedback
  createUserFriendlyMessage(error: JSONRPCError): string;
  getErrorSeverity(error: JSONRPCError): ErrorSeverity;
  
  // Recovery
  isErrorRecoverable(error: JSONRPCError): boolean;
  attemptRecovery(error: JSONRPCError): Promise<RecoveryResult>;
}
```

## **Implementation Guidelines**

### **Architecture Compliance**
- **100% Server API Alignment**: All interfaces must match server ground truth
- **No Client-Only Features**: Remove features not provided by server
- **Error-First Design**: Handle all server error codes properly
- **Real-time Updates**: WebSocket notifications for all state changes

### **Performance Requirements**
- **WebSocket Latency**: <20ms for real-time updates
- **API Response Time**: <100ms for control operations
- **Health Check Frequency**: Configurable polling intervals
- **Error Recovery**: Automatic retry with exponential backoff

### **Security Requirements**
- **JWT Authentication**: All API calls require valid tokens
- **Role-Based Access**: Viewer/operator/admin permission levels
- **Secure WebSocket**: WSS for production environments
- **Input Validation**: All parameters validated before API calls

This architecture ensures complete alignment with the server API ground truth while maintaining a robust, scalable client application. 