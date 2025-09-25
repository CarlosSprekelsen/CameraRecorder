# MediaMTX Camera Service Client - UI Surface Specification

**Document Version:** 1.0  
**Date:** January 2025  
**Classification:** UI Architecture Specification

---

## 1. Page Structure

### 1.1 Routes

| Route | Component | Purpose |
|-------|-----------|---------|
| `/login` | `LoginPage` | Token entry → authenticate |
| `/cameras` | `CameraPage` | Device table with actions |
| `/files` | `FilesPage` | Recordings/Snapshots management |
| `/about` | `AboutPage` | Server info/status display |

### 1.2 Navigation

- **Top-level routing**: React Router with protected routes
- **Authentication guard**: Redirect to `/login` if not authenticated
- **Default route**: `/cameras` (after login)

---

## 2. Page Specifications

### 2.1 Login Page (`/login`)

**Purpose**: Token-based authentication entry point

**UI Elements**:
- Token input field (password type)
- "Connect" button
- Error message display
- Loading state during authentication

**Behavior**:
- Submit token → call `authenticate` RPC
- On success → redirect to `/cameras`
- On failure → display error message
- Store token in session storage

### 2.2 Cameras Page (`/cameras`)

**Purpose**: Device discovery and control operations

**UI Elements**:
- Device table with columns:
  - Device ID
  - Status (Connected/Disconnected/Error)
  - Name
  - Resolution
  - FPS
  - Actions menu
- Action menu per device:
  - "Copy Link(s)" → copies HLS/WebRTC URLs
  - "Snapshot" → `take_snapshot`
  - "Record Start" → `start_recording` (unlimited)
  - "Record Stop" → `stop_recording`
  - "Timed Record" → `start_recording` with duration picker

**State Dependencies**:
- `get_camera_list()` → device data
- `get_stream_url(device)` → stream URLs for copying
- Real-time status updates via `camera_status_update` notifications

### 2.3 Files Page (`/files`)

**Purpose**: Server-side file management

**UI Elements**:
- Tab navigation: "Recordings" | "Snapshots"
- Paginated table with columns:
  - Filename
  - Size
  - Created/Modified time
  - Actions: "Download" | "Delete"
- Pagination controls (limit/offset)
- Search/filter (optional)

**State Dependencies**:
- `list_recordings(limit, offset)` → recordings data
- `list_snapshots(limit, offset)` → snapshots data
- `get_recording_info(filename)` → file details
- `get_snapshot_info(filename)` → file details

**Actions**:
- Download → open `download_url` in new tab
- Delete → confirmation dialog → `delete_recording`/`delete_snapshot`

### 2.4 About Page (`/about`)

**Purpose**: Server information and system status

**UI Elements**:
- Server info section:
  - Name, version, build date
  - Go version, architecture
  - Capabilities, supported formats
- System status section:
  - Overall status (Healthy/Degraded/Unhealthy)
  - Uptime, version
  - Component status (WebSocket, Camera Monitor, MediaMTX)
- Storage info section:
  - Total/used/available space
  - Usage percentage
  - Recordings/Snapshots size breakdown

**State Dependencies**:
- `get_server_info()` → server details
- `get_status()` → system health
- `get_storage_info()` → storage metrics

---

## 3. Global Elements

### 3.1 Top Bar

**Elements**:
- **Connection Status**: WebSocket connection indicator
  - Green: Connected
  - Yellow: Connecting/Reconnecting
  - Red: Disconnected
- **Server Info**: Server name and version
- **User Menu**: Sign out button

**State Dependencies**:
- WebSocket connection state
- `get_server_info()` → server name/version
- Authentication state → user role

### 3.2 Toast Notifications

**Types**:
- **Success**: Command completed successfully
- **Error**: Command failed with error message
- **Info**: Status updates, connection changes

**Triggers**:
- RPC command responses
- WebSocket connection state changes
- Server notification events

### 3.3 Confirmation Dialogs

**Delete Confirmations**:
- "Delete Recording" → `delete_recording(filename)`
- "Delete Snapshot" → `delete_snapshot(filename)`
- Show filename and size
- "Cancel" | "Delete" buttons

**Timed Recording Dialog**:
- Duration picker (seconds/minutes/hours)
- Format selector (fmp4/mp4/mkv)
- "Start Recording" | "Cancel" buttons

---

## 4. State Management

### 4.1 Authentication State

```typescript
interface AuthState {
  token: string | null;
  role: 'admin' | 'operator' | 'viewer' | null;
  session_id: string | null;
  isAuthenticated: boolean;
  expires_at: string | null;
}
```

### 4.2 Connection State

```typescript
interface ConnectionState {
  status: 'connected' | 'connecting' | 'disconnected' | 'error';
  lastError: string | null;
  reconnectAttempts: number;
}
```

### 4.3 Device State

```typescript
interface DeviceState {
  cameras: Camera[];
  streams: StreamInfo[];
  loading: boolean;
  error: string | null;
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

### 4.4 File State

```typescript
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
}

interface RecordingFile {
  filename: string;
  file_size: number;
  modified_time: string;
  download_url: string;
}
```

### 4.5 Server State

```typescript
interface ServerState {
  info: ServerInfo | null;
  status: SystemStatus | null;
  storage: StorageInfo | null;
  loading: boolean;
  error: string | null;
}
```

---

## 5. Component Architecture

### 5.1 Page Components

- `LoginPage` → `LoginForm`
- `CameraPage` → `CameraTable` + `DeviceActions`
- `FilesPage` → `FileTabs` + `FileTable` + `Pagination`
- `AboutPage` → `ServerInfo` + `SystemStatus` + `StorageInfo`

### 5.2 Shared Components

- `TopBar` → Connection status, server info, user menu
- `ToastContainer` → Success/error notifications
- `ConfirmDialog` → Delete confirmations
- `TimedRecordDialog` → Duration picker for recordings
- `LoadingSpinner` → Loading states
- `ErrorBoundary` → Error handling

### 5.3 Layout Components

- `AppLayout` → Main layout with top bar
- `PageContainer` → Page content wrapper
- `ActionMenu` → Device action dropdown
- `FileActions` → File row actions

---

## 6. User Interactions

### 6.1 Authentication Flow

1. User enters token → `authenticate` RPC
2. Store token in session storage
3. Redirect to `/cameras`
4. Show connection status in top bar

### 6.2 Device Control Flow

1. User clicks action → show loading state
2. Send RPC command → wait for response
3. Show success/error toast
4. Update device status if applicable

### 6.3 File Management Flow

1. User navigates to `/files`
2. Load recordings/snapshots → show loading
3. User clicks download → open URL in new tab
4. User clicks delete → show confirmation
5. Confirm → send delete RPC → refresh list

### 6.4 Real-time Updates

1. WebSocket notifications → update relevant state
2. Show toast for status changes
3. Refresh affected page data
4. Update connection status

---

## 7. Responsive Design

### 7.1 Breakpoints

- **Mobile**: < 768px
- **Tablet**: 768px - 1024px
- **Desktop**: > 1024px

### 7.2 Mobile Adaptations

- Collapsible navigation menu
- Stacked table layout
- Touch-friendly action buttons
- Swipe gestures for file actions

---

## 8. Accessibility

### 8.1 Keyboard Navigation

- Tab order through interactive elements
- Enter/Space for button activation
- Escape to close dialogs
- Arrow keys for table navigation

### 8.2 Screen Reader Support

- Semantic HTML structure
- ARIA labels for interactive elements
- Status announcements for dynamic content
- Alternative text for icons

---

## 9. Performance Considerations

### 9.1 Loading States

- Skeleton screens for initial load
- Progressive loading for large lists
- Debounced search/filter inputs
- Lazy loading for non-critical components

### 9.2 Caching Strategy

- Cache device list (5 minutes)
- Cache server info (10 minutes)
- No caching for file lists (always fresh)
- Session storage for auth token only

---

## 10. Error Handling

### 10.1 Error Types

- **Connection errors**: WebSocket disconnection
- **Authentication errors**: Invalid token, expired session
- **Command errors**: RPC failures, permission denied
- **File errors**: Download failures, delete failures

### 10.2 Error Recovery

- Automatic reconnection for WebSocket
- Token refresh for authentication
- Retry mechanisms for failed commands
- Graceful degradation for file operations

---

**Document Status:** Released  
**Classification:** UI Architecture Specification  
**Review Cycle:** Quarterly  
**Approval:** Architecture Board
