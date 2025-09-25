# MediaMTX Camera Service Client - Implementation Plan

**Document Version:** 1.0  
**Date:** January 2025  
**Classification:** Development Reference  
**Status:** Active Implementation

---

## 1. Implementation Strategy Overview

### 1.1 Architecture Alignment

This implementation plan is **perfectly aligned** with the client architecture document and follows a **layered, sprint-based approach** that maps directly to the architectural layers:

| Sprint | Architecture Layer | Focus Area |
|--------|-------------------|------------|
| 1 | Infrastructure + Service | WebSocket, Authentication, Basic RPC |
| 2 | Application + Presentation | Device Discovery, Stream Links |
| 3 | Application + Presentation | Command Operations (Snapshot/Recording) |
| 4 | Application + Presentation | File Management (List/Download/Delete) |
| 5 | Cross-cutting | Security, Performance, Polish |

### 1.2 RPC Method Alignment (Authoritative)

Based on section 5.3.1 of the architecture document:

#### **Discovery Methods:**
- `get_camera_list` → cameras with stream fields
- `get_streams` → MediaMTX active streams  
- `get_stream_url` → URL for specific device

#### **Command Methods:**
- `take_snapshot(device[, filename])`
- `start_recording(device[, duration][, format])`
- `stop_recording(device)`

#### **File Methods:**
- `list_recordings(limit, offset)`
- `list_snapshots(limit, offset)`
- `get_recording_info(filename)`
- `get_snapshot_info(filename)`
- `delete_recording(filename)`
- `delete_snapshot(filename)`

#### **Status/Admin Methods:**
- `get_status`, `get_storage_info`, `get_server_info`, `get_metrics`
- `subscribe_events`, `unsubscribe_events`, `get_subscription_stats`

---

## 2. Sprint Implementation Details

### 2.1 Sprint 1 — Auth + WS/RPC Foundation

**Architecture Layer:** Infrastructure + Service  
**Duration:** 1 week  
**Focus:** WebSocket connection, authentication, basic RPC client
MANDATORY: 100% adherence to architechture documents

#### **Methods to Implement:**
- `ping` - Connectivity check
- `authenticate(auth_token)` - Session establishment
- `get_server_info()` - Server metadata
- `get_status()` - System health

#### **UI Components:**
- `LoginPage` - Token entry and authentication
- `AboutPage` - Server information display
- `TopBar` - WebSocket status indicator
- `AppLayout` - Main application shell

#### **State Management:**
```typescript
// authStore
interface AuthState {
  token: string | null;
  role: 'admin' | 'operator' | 'viewer' | null;
  session_id: string | null;
  isAuthenticated: boolean;
  expires_at: string | null;
}

// connectionStore  
interface ConnectionState {
  status: 'connected' | 'connecting' | 'disconnected' | 'error';
  lastError: string | null;
  reconnectAttempts: number;
}

// serverStore
interface ServerState {
  info: ServerInfo | null;
  status: SystemStatus | null;
  loading: boolean;
  error: string | null;
}
```

#### **Enhanced DoD (Definition of Done):**
- [ ] WebSocket connection with automatic reconnection
- [ ] Token-based authentication with session persistence
- [ ] Server info display on About page
- [ ] System status indicator in top bar
- [ ] Redirect to `/login` if unauthenticated
- [ ] Error boundary for connection failures
- [ ] Loading states for all async operations
- [ ] Session storage for token persistence
- [ ] Exponential backoff for reconnection attempts

#### **Technical Implementation:**
```typescript
// WebSocket Service with reconnection
class WebSocketService {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  
  async connect(url: string): Promise<void> {
    // Implementation with exponential backoff
  }
  
  async sendRPC(method: string, params?: any): Promise<any> {
    // JSON-RPC 2.0 implementation
  }
}

// Authentication Service
class AuthenticationService {
  async authenticate(token: string): Promise<AuthResult> {
    // Call authenticate RPC method
  }
  
  async refreshToken(): Promise<void> {
    // Token refresh logic
  }
}
```

### 2.2 Sprint 2 — Cameras & Links

**Architecture Layer:** Application + Presentation  
**Duration:** 1 week  
**Focus:** Device discovery, stream link management
**MANDATORY:** 100% adherence to architechture documents

#### **Methods to Implement:**
- `get_camera_list()` - Device enumeration
- `get_stream_url(device)` - Stream URL retrieval
- `get_streams()` - MediaMTX stream status
- `subscribe_events(['camera_status_update'])` - Real-time updates

#### **UI Components:**
- `CameraPage` - Main device table
- `CameraTable` - Device list with status
- `DeviceActions` - Per-device action menu
- `CopyLinkButton` - Stream URL copying

#### **Enhanced DoD:**
- [ ] Device table with real-time status updates
- [ ] Copy link functionality (HLS/WebRTC URLs)
- [ ] Stream availability indicators
- [ ] Real-time status updates via WebSocket
- [ ] Device capability display
- [ ] Loading states for device operations
- [ ] Error handling for device failures
- [ ] Responsive table design

#### **State Management:**
```typescript
// deviceStore
interface DeviceState {
  cameras: Camera[];
  streams: StreamInfo[];
  loading: boolean;
  error: string | null;
  lastUpdated: string | null;
}

interface Camera {
  device: string;
  status: 'CONNECTED' | 'DISCONNECTED' | 'ERROR';
  name: string;
  resolution: string;
  fps: number;
  streams: {
    rtsp: string;
    hls: string;
  };
}
```

### 2.3 Sprint 3 — Commands (Snapshot & Recording)

**Architecture Layer:** Application + Presentation  
**Duration:** 1 week  
**Focus:** Device control operations
**MANDATORY:** 100% adherence to architechture documents


#### **Methods to Implement:**
- `take_snapshot(device[, filename])` - Image capture
- `start_recording(device[, duration][, format])` - Video recording
- `stop_recording(device)` - Recording termination
- `subscribe_events(['recording_status_update'])` - Recording status

#### **UI Components:**
- `DeviceActions` - Enhanced with command buttons
- `SnapshotButton` - Image capture
- `RecordingControls` - Start/stop/timed recording
- `TimedRecordDialog` - Duration picker
- `RecordingStatus` - Current recording indicators

#### **Enhanced DoD:**
- [ ] Snapshot capture with success feedback
- [ ] Recording start/stop with status updates
- [ ] Timed recording with duration picker
- [ ] Real-time recording status display
- [ ] Command acknowledgment toasts
- [ ] Error handling for failed commands
- [ ] Recording state persistence
- [ ] Concurrent recording limits
- [ ] 100% adherence to architechture documents
- [ ] All previous sprints DoD validated.

#### **State Management:**
```typescript
// recordingStore
interface RecordingState {
  activeRecordings: Map<string, RecordingInfo>;
  recordingHistory: RecordingInfo[];
  loading: boolean;
  error: string | null;
}

interface RecordingInfo {
  device: string;
  filename: string;
  status: 'RECORDING' | 'STOPPED' | 'ERROR';
  startTime: string;
  duration?: number;
  format: string;
}
```

### 2.4 Sprint 4 — Files (List/Download/Delete)

**Architecture Layer:** Application + Presentation  
**Duration:** 1 week  
**Focus:** Server-side file management
**MANDATORY:** 100% adherence to architechture documents

#### **Methods to Implement:**
- `list_recordings(limit, offset)` - Recording enumeration
- `list_snapshots(limit, offset)` - Snapshot enumeration
- `get_recording_info(filename)` - Recording metadata
- `get_snapshot_info(filename)` - Snapshot metadata
- `delete_recording(filename)` - Recording deletion
- `delete_snapshot(filename)` - Snapshot deletion

#### **UI Components:**
- `FilesPage` - Main file management interface
- `FileTabs` - Recordings/Snapshots tabs
- `FileTable` - Paginated file list
- `FileActions` - Download/delete actions
- `ConfirmDialog` - Delete confirmation
- `Pagination` - Page navigation

#### **Enhanced DoD:**
- [ ] Paginated file tables (recordings/snapshots)
- [ ] Download functionality via server URLs
- [ ] Delete confirmation dialogs
- [ ] File size formatting (human-readable)
- [ ] Sort by date/size/name
- [ ] Bulk operations support
- [ ] Download progress indicators
- [ ] Real-time file list updates

#### **State Management:**
```typescript
// fileStore
interface FileState {
  recordings: RecordingFile[];
  snapshots: SnapshotFile[];
  loading: boolean;
  error: string | null;
  pagination: {
    limit: number;
    offset: number;
    total: number;
  };
  selectedFiles: string[];
}

interface RecordingFile {
  filename: string;
  file_size: number;
  modified_time: string;
  download_url: string;
}
```

### 2.5 Sprint 5 — Hardening & Polish

**Architecture Layer:** Cross-cutting concerns  
**Duration:** 1 week  
**Focus:** Security, performance, accessibility
**MANDATORY:** 100% adherence to architechture documents

#### **Security Enhancements:**
- Role-based UI hiding (viewer vs operator vs admin)
- Permission-based action availability
- Secure token storage
- Input validation and sanitization

#### **Performance Optimizations:**
- Loading skeletons for better UX
- Debounced search/filter inputs
- Lazy loading for non-critical components
- Performance monitoring (Core Web Vitals)

#### **Accessibility Improvements:**
- Keyboard navigation support
- Screen reader compatibility
- ARIA labels and semantic HTML
- Focus management

#### **Enhanced DoD:**
- [ ] Role-based access control
- [ ] Keyboard shortcuts for common actions
- [ ] Offline detection and messaging
- [ ] Performance monitoring implementation
- [ ] Lighthouse audit pass
- [ ] Accessibility compliance (WCAG 2.1 AA)
- [ ] Error boundary coverage
- [ ] Loading state consistency
- [ ] 100% alingment with architechture
- [ ] All previous sprints DoD validated.


---

## 3. Technical Implementation Guidelines

### 3.1 Type Safety Strategy

```typescript
// Generate TypeScript types from OpenRPC spec
// This ensures compile-time safety for all RPC calls

// Example generated types:
interface AuthenticateParams {
  auth_token: string;
}

interface AuthenticateResult {
  authenticated: boolean;
  role: 'admin' | 'operator' | 'viewer';
  permissions: string[];
  expires_at: string;
  session_id: string;
}
```

### 3.2 State Management Pattern

```typescript
// Use Zustand stores aligned with architecture layers:

// Infrastructure Layer
const useConnectionStore = create<ConnectionState>((set) => ({
  status: 'disconnected',
  lastError: null,
  reconnectAttempts: 0,
  // ... actions
}));

// Service Layer  
const useAuthStore = create<AuthState>((set) => ({
  token: null,
  role: null,
  session_id: null,
  isAuthenticated: false,
  // ... actions
}));

// Application Layer
const useDeviceStore = create<DeviceState>((set) => ({
  cameras: [],
  streams: [],
  loading: false,
  error: null,
  // ... actions
}));
```

### 3.3 Error Handling Strategy

```typescript
// Implement error boundaries and recovery patterns
// from section 8.1 of architecture

class ErrorBoundary extends React.Component {
  // Error boundary implementation
}

// Error recovery patterns:
const useErrorRecovery = () => {
  const retryConnection = useCallback(() => {
    // Reconnection logic
  }, []);
  
  const retryCommand = useCallback((command: () => Promise<void>) => {
    // Command retry logic
  }, []);
  
  return { retryConnection, retryCommand };
};
```

### 3.4 Performance Monitoring

```typescript
// Track metrics from section 10.1 of architecture:
// - Command Ack ≤ 200ms (p95)
// - Event-to-UI ≤ 100ms (p95)  
// - Initial Load Time < 3 seconds

const usePerformanceMonitoring = () => {
  const trackCommandLatency = useCallback((method: string, startTime: number) => {
    const latency = Date.now() - startTime;
    // Track latency metrics
  }, []);
  
  const trackEventToUI = useCallback((eventType: string, startTime: number) => {
    const latency = Date.now() - startTime;
    // Track event processing time
  }, []);
  
  return { trackCommandLatency, trackEventToUI };
};
```

---

## 4. Architecture Compliance Checklist

### 4.1 Section 5.3.1 RPC Method Alignment
- ✅ **Discovery**: `get_camera_list`, `get_stream_url`, `get_streams`
- ✅ **Commands**: `take_snapshot`, `start_recording`, `stop_recording`  
- ✅ **Files**: `list_recordings`, `list_snapshots`, `get_recording_info`, `get_snapshot_info`, `delete_recording`, `delete_snapshot`
- ✅ **Status**: `get_status`, `get_server_info`, `subscribe_events`

### 4.2 Section 8.3 Security Architecture
- ✅ **Authentication**: Token-based with session management
- ✅ **Authorization**: Role-based access control
- ✅ **Transport**: WebSocket encryption
- ✅ **Input Validation**: Client-side validation
- ✅ **Session Management**: Automatic timeout and renewal

### 4.3 Section 13.4 Scope Compliance
- ✅ **No embedded playback**: HLS/WebRTC links only
- ✅ **Server-authoritative timers**: Duration passed to server
- ✅ **File operations server-side**: Download URLs, server deletions
- ✅ **Control plane only**: WebSocket/JSON-RPC
- ✅ **Delete confirmations**: User confirmation required

---

## 5. Quality Assurance Strategy

### 5.1 Testing Approach
- **Unit Tests**: Store logic, service methods, utility functions
- **Integration Tests**: RPC communication, state management
- **E2E Tests**: User workflows, error scenarios
- **Performance Tests**: Load time, responsiveness metrics

### 5.2 Code Quality
- **ESLint/Prettier**: Automated code formatting
- **TypeScript**: Strict type checking
- **Architecture Compliance**: Regular architecture reviews
- **Security Audits**: Token handling, input validation

### 5.3 Documentation
- **API Documentation**: Generated from OpenRPC spec
- **Component Documentation**: Storybook for UI components
- **Architecture Documentation**: Regular updates to architecture doc
- **User Documentation**: Help text, tooltips, error messages

---

## 6. Risk Mitigation

### 6.1 Technical Risks
- **WebSocket Instability**: Robust reconnection logic
- **State Synchronization**: Clear recovery procedures
- **Performance Degradation**: Monitoring and optimization
- **Security Vulnerabilities**: Regular audits and updates

### 6.2 Implementation Risks
- **Scope Creep**: Strict adherence to architecture constraints
- **Technical Debt**: Regular refactoring cycles
- **Dependency Issues**: Regular dependency updates
- **Browser Compatibility**: Cross-browser testing

---

**Document Status:** Active Implementation  
**Classification:** Development Reference  
**Review Cycle:** Weekly during implementation  
**Approval:** Development Team Lead
