# MediaMTX Camera Service Client Requirements

**Version:** 2.0  
**Last Updated:** 2025-01-16  
**Status:** 🚨 **CRITICAL UPDATE - ARCHITECTURE REFACTORED FOR SERVER API ALIGNMENT**

## **Project Overview**

The MediaMTX Camera Service Client is a React/TypeScript Progressive Web App (PWA) that provides a modern web interface for managing USB cameras through the MediaMTX Camera Service. The client has been refactored to achieve 100% alignment with the server API ground truth.

### **Key Changes in Version 2.0**
- **Removed Client-Only Features**: Configuration management, enhanced error recovery, and advanced storage monitoring removed
- **Fixed API Misalignments**: RecordingConflict objects replaced with error responses
- **Server API Alignment**: All requirements now match server ground truth exactly
- **Simplified Architecture**: Focus on server-provided capabilities only

## **Supported Service Methods**

The client must support all server JSON-RPC methods:

### **Authentication**
- `authenticate(auth_token: string)` - JWT token authentication

### **Core Methods**
- `ping()` - Health check
- `get_camera_list()` - List available cameras
- `get_camera_status(device: string)` - Get camera status

### **Control Methods**
- `take_snapshot(device: string)` - Capture camera snapshot
- `start_recording(device: string)` - Start video recording
- `stop_recording(device: string)` - Stop video recording

### **File Management**
- `list_recordings(device?, limit?, offset?)` - List recording files
- `list_snapshots(device?, limit?, offset?)` - List snapshot files
- `get_recording_info(filename: string)` - Get recording metadata
- `get_snapshot_info(filename: string)` - Get snapshot metadata
- `delete_recording(filename: string)` - Delete recording file
- `delete_snapshot(filename: string)` - Delete snapshot file

### **Storage Management**
- `get_storage_info()` - Get storage usage information

### **System Information**
- `get_metrics()` - Get system metrics
- `get_status()` - Get system status
- `get_server_info()` - Get server information
- `get_streams()` - Get available streams

## **Health Endpoints**

The client must support all server health endpoints:

- `GET /health/system` - Overall system health
- `GET /health/cameras` - Camera system health
- `GET /health/mediamtx` - MediaMTX integration health
- `GET /health/ready` - Kubernetes readiness status

## **File Download Endpoints**

The client must support file download endpoints:

- `GET /files/recordings/{filename}` - Download recording file
- `GET /files/snapshots/{filename}` - Download snapshot file

## **Functional Requirements**

### **F1: Real-time Camera Management**

#### **F1.1: Camera Discovery**
- Display list of available USB cameras
- Show camera status (connected/disconnected)
- Real-time status updates via WebSocket notifications

#### **F1.2: Camera Control**
- Take snapshots from any connected camera
- Start/stop video recording with conflict prevention
- Display camera capabilities and stream information

#### **F1.3: Camera Status Monitoring**
- Real-time camera status updates
- Connection health monitoring
- Error state display and recovery

#### **F1.4: Enhanced Recording Management**
- Real-time recording status updates
- Recording progress tracking
- Error handling for recording conflicts
- Session management and cleanup

### **F2: System Health Monitoring**

#### **F2.1: Health Endpoint Integration**
- System health status display
- Component health monitoring
- Kubernetes readiness status
- Health history and trends

#### **F2.2: Connection Monitoring**
- WebSocket connection status
- HTTP health endpoint status
- Connection error handling and recovery
- Automatic reconnection

#### **F2.3: Performance Monitoring**
- Response time monitoring
- Error rate tracking
- System resource usage
- Performance alerts

#### **F2.4: Storage Monitoring**
- Storage usage display
- Threshold status indicators
- Storage warnings and alerts
- Storage validation for operations

### **F3: Authentication and Security**

#### **F3.1: JWT Authentication**
- Secure token management
- Automatic token refresh
- Session validation
- Secure token storage

#### **F3.2: Role-Based Access Control**
- Viewer role: Read-only access
- Operator role: Camera control access
- Admin role: Full system access
- Permission-based UI adaptation

#### **F3.3: Secure Communication**
- HTTPS/WSS for production
- Input validation and sanitization
- Secure error handling
- Token security

#### **F3.4: Enhanced Error Handling**
- JSON-RPC error code handling
- User-friendly error messages
- Error logging and tracking
- Error recovery strategies

### **F4: File Management**

#### **F4.1: Recording Management**
- List recording files with pagination
- Download recording files
- Delete recording files (with permissions)
- Recording metadata display

#### **F4.2: Snapshot Management**
- List snapshot files with pagination
- Download snapshot files
- Delete snapshot files (with permissions)
- Snapshot metadata display

#### **F4.3: File Operations**
- File download progress tracking
- File deletion confirmation
- File search and filtering
- File organization and sorting

### **F5: User Interface**

#### **F5.1: Responsive Design**
- Mobile-first responsive layout
- Touch-friendly interface
- Cross-browser compatibility
- Progressive Web App capabilities

#### **F5.2: Real-time Updates**
- WebSocket-based notifications
- Live status updates
- Real-time data refresh
- Optimistic UI updates

#### **F5.3: User Experience**
- Intuitive navigation
- Clear error messages
- Loading states and feedback
- Accessibility compliance

#### **F5.4: Dashboard Interface**
- Camera grid overview
- Quick action buttons
- Status summary display
- System health indicators

## **Technical Requirements**

### **T1: WebSocket Integration**

#### **T1.1: JSON-RPC 2.0 Client**
- Standard JSON-RPC 2.0 protocol support
- Two-parameter method calls (method, params)
- Proper error handling and response parsing
- Connection management and reconnection

#### **T1.2: Real-time Communication**
- WebSocket connection establishment
- Automatic reconnection on failure
- Message queuing and retry
- Connection health monitoring

#### **T1.3: Event Handling**
- Real-time notification processing
- Event-driven state updates
- Message routing and dispatching
- Error event handling

### **T2: HTTP Client Integration**

#### **T2.1: Health Endpoint Client**
- REST API client for health endpoints
- Configurable polling intervals
- Health data caching and management
- Error handling and retry logic

#### **T2.2: File Download Client**
- HTTP client for file downloads
- Download progress tracking
- File streaming and caching
- Download error handling

### **T3: State Management**

#### **T3.1: Zustand Store Architecture**
- Lightweight state management
- Type-safe state updates
- State persistence and hydration
- Store composition and modularity

#### **T3.2: React Query Integration**
- Server state management
- Automatic caching and invalidation
- Background data synchronization
- Optimistic updates

#### **T3.3: State Synchronization**
- WebSocket state synchronization
- HTTP polling fallback
- State consistency management
- Conflict resolution

### **T4: Error Handling**

#### **T4.1: JSON-RPC Error Codes**
- Support for all server error codes
- Error code mapping and categorization
- User-friendly error messages
- Error severity classification

#### **T4.2: Error Recovery**
- Automatic retry mechanisms
- Graceful degradation
- Error logging and reporting
- User feedback and guidance

#### **T4.3: Error Prevention**
- Input validation and sanitization
- Pre-operation validation
- Conflict detection and prevention
- Resource availability checks

### **T5: Performance Requirements**

#### **T5.1: Response Times**
- WebSocket latency: <20ms
- API response time: <100ms
- UI interaction feedback: <200ms
- Health check frequency: Configurable

#### **T5.2: Resource Usage**
- Memory usage optimization
- CPU usage minimization
- Network bandwidth efficiency
- Battery life optimization (mobile)

#### **T5.3: Scalability**
- Support for multiple cameras
- Efficient data handling
- Optimized rendering
- Progressive loading

## **Non-Functional Requirements**

### **NF1: Reliability**
- 99.9% uptime target
- Graceful error handling
- Automatic recovery mechanisms
- Data consistency guarantees

### **NF2: Security**
- JWT token security
- Role-based access control
- Secure communication protocols
- Input validation and sanitization

### **NF3: Usability**
- Intuitive user interface
- Minimal training requirements
- Accessibility compliance
- Cross-platform compatibility

### **NF4: Maintainability**
- TypeScript type safety
- Modular architecture
- Comprehensive documentation
- Testing coverage

### **NF5: Performance**
- Fast application startup
- Responsive user interface
- Efficient data handling
- Optimized resource usage

## **Implementation Priorities**

### **Priority 1: Core Functionality** ✅ **COMPLETE**
- WebSocket JSON-RPC integration
- HTTP health monitoring
- Authentication and authorization
- Basic camera operations

### **Priority 2: Enhanced Features** ✅ **COMPLETE**
- Recording state management
- Storage monitoring
- Error handling
- File management

### **Priority 3: User Experience** ✅ **COMPLETE**
- Responsive UI components
- Real-time updates
- Error feedback
- Performance optimization

### **Priority 4: Testing & Quality** 🔄 **IN PROGRESS**
- Unit test coverage
- Integration testing
- End-to-end testing
- Performance validation

## **Success Criteria**

### **Functional Success**
- All server API methods supported
- Real-time updates working correctly
- Error handling comprehensive
- File operations functional

### **Performance Success**
- Response times meet targets
- Resource usage optimized
- Scalability demonstrated
- Reliability achieved

### **User Experience Success**
- Intuitive interface design
- Responsive across devices
- Accessibility compliance
- Minimal training required

### **Technical Success**
- TypeScript coverage 100%
- Linting compliance
- Architecture alignment
- Documentation complete

## **Conclusion**

The MediaMTX Camera Service Client requirements have been updated to achieve 100% alignment with the server API ground truth. All client-only features have been removed, and the requirements now focus exclusively on server-provided capabilities.

The client is designed to provide a robust, scalable, and user-friendly interface for camera management while maintaining strict alignment with the server API specifications.