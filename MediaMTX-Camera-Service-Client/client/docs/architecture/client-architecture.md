# MediaMTX Camera Service - Client Architecture

## **Project Overview**

### **Purpose**
The MediaMTX Camera Service Client is a React/TypeScript Progressive Web App (PWA) that provides a modern web interface for managing USB cameras, monitoring their status, and controlling recording/snapshot operations through the MediaMTX Camera Service WebSocket JSON-RPC API.

### **Project Objectives**
1. **Real-time Camera Management**: Provide instant visibility into camera status and capabilities
2. **Recording Control**: Enable snapshot capture and recording start/stop operations
3. **Mobile-First Design**: Responsive PWA that works seamlessly on smartphones and desktops
4. **Intuitive UX**: Clean, modern interface that requires minimal training
5. **Reliable Communication**: WebSocket-based real-time updates with polling fallback

### **Target Users**
- **System Administrators**: Monitor camera health and manage recordings
- **Security Personnel**: Quick access to camera status and snapshot capture
- **Mobile Users**: Responsive interface for on-the-go camera management
- **Developers**: API integration reference and testing interface

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
├─────────────────────────────────────────────────────────────┤
│                React Component Architecture                 │
│     • Dashboard (camera grid, status overview)            │
│     • Camera Detail (capabilities, controls, history)     │
│     • Settings (server configuration, PWA settings)       │
│     • Notifications (real-time status updates)            │
├─────────────────────────────────────────────────────────────┤
│                State Management (Zustand)                  │
│     • Camera state (connected devices, status, metadata)  │
│     • UI state (selected camera, view mode, settings)     │
│     • Connection state (WebSocket status, error handling)  │
└─────────────────────┬───────────────────────────────────────┘
                      │ WebSocket JSON-RPC 2.0
┌─────────────────────▼───────────────────────────────────────┐
│              MediaMTX Camera Service                        │
│                (Backend Server)                            │
└─────────────────────────────────────────────────────────────┘
```

### **Technology Stack**

#### **Frontend Framework**
- **React 18+**: Modern React with hooks and concurrent features
- **TypeScript**: Type safety and better developer experience
- **Vite**: Fast build tool and development server
- **PWA Support**: Service workers and offline capabilities

#### **UI Framework**
- **Material-UI (MUI)**: Comprehensive component library
- **Emotion**: CSS-in-JS for styling
- **React Router**: Client-side routing

#### **State Management**
- **Zustand**: Lightweight state management
- **React Query**: Server state management and caching

#### **Communication**
- **WebSocket**: Real-time bidirectional communication
- **JSON-RPC 2.0**: Structured API calls and responses
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
- Header (navigation, connection status)
- Sidebar (camera list, settings)
- Main Content Area (dashboard, detail views)
- Notification System (real-time updates)
```

#### **2. Dashboard**
```typescript
// Dashboard.tsx - Camera overview
- Camera Grid (status cards for all cameras)
- Quick Actions (snapshot, record buttons)
- Status Summary (connected/disconnected counts)
- Real-time Updates (WebSocket notifications)
```

#### **3. Camera Detail**
```typescript
// CameraDetail.tsx - Individual camera view
- Camera Information (name, device path, capabilities)
- Status Indicators (connection, recording, streaming)
- Control Panel (snapshot, record start/stop)
- History (recent snapshots, recordings)
```

#### **4. WebSocket Client**
```typescript
// WebSocketClient.ts - Communication layer
- Connection Management (connect, disconnect, reconnect)
- JSON-RPC Protocol (method calls, responses)
- Notification Handling (real-time updates)
- Error Handling (connection failures, timeouts)
- Polling Fallback (backup for missed notifications)
```

#### **5. State Management**
```typescript
// stores/
- cameraStore.ts (camera state, status, metadata)
- uiStore.ts (selected camera, view mode, settings)
- connectionStore.ts (WebSocket status, errors)
```

### **Data Flow**

#### **Real-time Updates**
1. **WebSocket Connection**: Client connects to server WebSocket endpoint
2. **Notification Subscription**: Client listens for `camera_status_update` events
3. **State Updates**: Incoming notifications update Zustand stores
4. **UI Re-renders**: React components reflect new state automatically
5. **Polling Fallback**: If no notification received in 5 seconds, poll for updates

#### **User Actions**
1. **User Interaction**: User clicks button (e.g., "Take Snapshot")
2. **RPC Call**: Client sends JSON-RPC method call via WebSocket
3. **Server Processing**: Server executes action (camera snapshot)
4. **Response**: Server sends success/error response
5. **UI Update**: Client updates UI based on response
6. **Notification**: Server broadcasts status update to all clients

## **API Integration**

### **WebSocket JSON-RPC Methods**

#### **Core Camera Operations**
```typescript
// Camera discovery and status
await rpc.call("get_camera_list", {})
await rpc.call("get_camera_status", { device: "/dev/video0" })

// Recording operations
await rpc.call("start_recording", { 
  device: "/dev/video0", 
  duration: 60, 
  format: "mp4" 
})
await rpc.call("stop_recording", { device: "/dev/video0" })

// Snapshot operations
await rpc.call("take_snapshot", { 
  device: "/dev/video0", 
  format: "jpg", 
  quality: 85 
})
```

#### **Real-time Notifications**
```typescript
// Subscribe to camera status updates
{
  "jsonrpc": "2.0",
  "method": "camera_status_update",
  "params": {
    "device": "/dev/video0",
    "status": "CONNECTED",
    "capabilities": { ... },
    "streams": { ... }
  }
}
```

### **Error Handling**
- **Connection Failures**: Automatic reconnection with exponential backoff
- **RPC Errors**: User-friendly error messages and retry options
- **Network Issues**: Offline detection and graceful degradation
- **Server Errors**: Fallback to polling for critical operations

## **Development Phases**

### **Phase 1: MVP (S1-S2)**
**Timeline**: 2-3 weeks

#### **S1: Architecture & Scaffolding**
- [ ] **Project Setup**: Initialize React/TypeScript project with Vite
- [ ] **WebSocket Client**: Implement JSON-RPC client with connection management
- [ ] **State Management**: Set up Zustand stores for camera and UI state
- [ ] **Component Scaffolding**: Create basic component structure
- [ ] **Material-UI Integration**: Set up theme and component library
- [ ] **PWA Configuration**: Add service worker and manifest

#### **S2: Core Implementation**
- [ ] **Dashboard**: Camera grid with status cards and quick actions
- [ ] **Camera Detail**: Individual camera view with controls and information
- [ ] **Real-time Updates**: WebSocket notification handling and UI updates
- [ ] **Recording Controls**: Start/stop recording with status feedback
- [ ] **Snapshot Controls**: Take snapshots with format/quality options
- [ ] **Responsive Design**: Mobile-first layout with PWA capabilities

### **Phase 2: Enhancement (S3-S4)**
**Timeline**: 2-3 weeks

#### **S3: Testing & Validation**
- [ ] **Unit Tests**: Jest + React Testing Library for components
- [ ] **Integration Tests**: MSW for API mocking and testing
- [ ] **E2E Tests**: Cypress for full user workflow testing
- [ ] **Performance Testing**: Load testing with multiple cameras
- [ ] **Accessibility Testing**: WCAG compliance and screen reader support

#### **S4: Polish & Release**
- [ ] **Error Handling**: Comprehensive error states and user feedback
- [ ] **Offline Support**: Service worker for offline functionality
- [ ] **Performance Optimization**: Code splitting and lazy loading
- [ ] **Documentation**: User guide and API reference
- [ ] **Deployment**: Build pipeline and hosting configuration

### **Future Phases**

#### **Phase 3: Advanced Features**
- [ ] **Live Streaming**: HLS/WebRTC video preview integration
- [ ] **Authentication**: JWT-based user authentication
- [ ] **Settings Management**: Server configuration and preferences
- [ ] **Advanced Controls**: Camera configuration and capability management

#### **Phase 4: Mobile Enhancement**
- [ ] **Native App**: React Native or Flutter for Android/iOS
- [ ] **Push Notifications**: Real-time alerts for camera events
- [ ] **Offline Recording**: Local storage for critical operations

## **Project Structure**

```
camera-service-client/
├── public/
│   ├── manifest.json          # PWA manifest
│   └── icons/                 # App icons
├── src/
│   ├── components/            # React components
│   │   ├── Dashboard/         # Camera overview
│   │   ├── CameraDetail/      # Individual camera view
│   │   ├── Settings/          # Configuration
│   │   └── common/            # Shared components
│   ├── hooks/                 # Custom React hooks
│   │   ├── useWebSocket.ts    # WebSocket connection
│   │   ├── useCamera.ts       # Camera operations
│   │   └── useNotifications.ts # Real-time updates
│   ├── stores/                # Zustand state management
│   │   ├── cameraStore.ts     # Camera state
│   │   ├── uiStore.ts         # UI state
│   │   └── connectionStore.ts # Connection state
│   ├── services/              # API and external services
│   │   ├── websocket.ts       # WebSocket client
│   │   ├── rpc.ts             # JSON-RPC protocol
│   │   └── api.ts             # API utilities
│   ├── types/                 # TypeScript type definitions
│   │   ├── camera.ts          # Camera-related types
│   │   ├── rpc.ts             # JSON-RPC types
│   │   └── ui.ts              # UI-related types
│   ├── utils/                 # Utility functions
│   │   ├── polling.ts         # Polling fallback
│   │   ├── notifications.ts   # Notification handling
│   │   └── validation.ts      # Input validation
│   ├── styles/                # Global styles and themes
│   │   ├── theme.ts           # Material-UI theme
│   │   └── global.css         # Global styles
│   └── App.tsx                # Main application component
├── tests/                     # Test files
│   ├── unit/                  # Unit tests
│   ├── integration/           # Integration tests
│   └── e2e/                  # End-to-end tests
├── docs/                      # Documentation
│   ├── api-reference.md       # API documentation
│   ├── deployment.md          # Deployment guide
│   └── user-guide.md          # User documentation
├── package.json               # Dependencies and scripts
├── vite.config.ts             # Vite configuration
├── tsconfig.json              # TypeScript configuration
├── cypress.config.ts          # Cypress configuration
└── README.md                  # Project documentation
```

## **Integration with Server**

### **API Reference**
The client integrates directly with the MediaMTX Camera Service WebSocket JSON-RPC API. For complete API documentation, see:
- **Server API Reference**: `docs/api/json-rpc-methods.md`
- **WebSocket Protocol**: `docs/api/websocket-protocol.md`
- **Error Codes**: `docs/api/error-codes.md`

### **Key Integration Points**
1. **WebSocket Connection**: `ws://localhost:8002/ws`
2. **JSON-RPC Methods**: Direct method calls for camera operations
3. **Real-time Notifications**: Subscribe to status update events
4. **Error Handling**: Handle server errors and connection issues
5. **Polling Fallback**: Backup mechanism for missed notifications

### **Configuration**
- **Server URL**: Configurable WebSocket endpoint
- **Reconnection**: Automatic reconnection with exponential backoff
- **Polling Interval**: 5-second fallback for missed notifications
- **Timeout Settings**: Configurable request timeouts

## **Security Considerations**

### **Current Phase (MVP)**
- **No Authentication**: Direct WebSocket connection to server
- **Local Network**: Assumes server is on local network
- **HTTPS**: Use HTTPS in production for secure communication

### **Future Phases**
- **JWT Authentication**: Token-based user authentication
- **API Key Support**: Alternative authentication method
- **Role-based Access**: Different permissions for different users
- **Secure WebSocket**: WSS (WebSocket Secure) in production

## **Performance Considerations**

### **Real-time Updates**
- **WebSocket Efficiency**: Minimal overhead for status updates
- **State Management**: Efficient updates with Zustand
- **UI Optimization**: React.memo and useMemo for performance
- **Polling Fallback**: Minimal polling to reduce server load

### **Mobile Performance**
- **PWA Optimization**: Service worker for offline support
- **Responsive Design**: Mobile-first approach
- **Touch Interactions**: Optimized for touch devices
- **Battery Efficiency**: Minimal background processing

## **Testing Strategy**

### **Unit Testing**
- **Components**: React Testing Library for component testing
- **Hooks**: Custom hook testing with React Hooks Testing Library
- **Utilities**: Jest for utility function testing
- **Coverage**: Target 80%+ code coverage

### **Integration Testing**
- **API Mocking**: MSW for WebSocket and JSON-RPC mocking
- **State Management**: Zustand store testing
- **Error Scenarios**: Network failures and server errors
- **Real-time Updates**: Notification handling testing

### **End-to-End Testing**
- **Cypress**: Full user workflow testing
- **Real Server**: Tests against actual MediaMTX Camera Service
- **Mobile Testing**: Responsive design validation
- **PWA Testing**: Offline functionality and installation

## **Deployment Strategy**

### **Development**
- **Vite Dev Server**: Hot reload and fast development
- **Local Server**: Connect to local MediaMTX Camera Service
- **Environment Variables**: Configuration for different environments

### **Production**
- **Static Build**: Optimized production build
- **CDN Deployment**: Fast global distribution
- **HTTPS**: Secure communication with server
- **PWA Deployment**: Service worker and manifest for mobile

### **CI/CD Pipeline**
- **Automated Testing**: Run all tests on pull requests
- **Build Validation**: Ensure production build succeeds
- **Deployment**: Automated deployment to staging/production
- **Monitoring**: Performance and error monitoring

---

**Client Architecture**: Complete  
**Status**: Ready for Implementation  
**Next Step**: Begin S1 (Architecture & Scaffolding)  
**Estimated Timeline**: 4-6 weeks for MVP completion 