# Epic E3 Development Plan: WebSocket JSON-RPC Server

**Version:** 1.0  
**Date:** 2025-01-15  
**Status:** Pending PM Approval  
**Related Epic/Story:** Epic E3 - WebSocket JSON-RPC Server  

## Executive Summary

This development plan outlines the implementation strategy for Epic E3: WebSocket JSON-RPC Server, focusing on high-performance WebSocket server with 1000+ concurrent connections and <50ms response time. The implementation will integrate with existing foundation components from Epic E1 and camera discovery from Epic E2.

### Implementation Goals
- **Performance**: 1000+ concurrent WebSocket connections
- **Response Time**: <50ms for status methods, <100ms for control methods
- **Integration**: Full integration with Configuration Management (Epic E1) and Camera Discovery (Epic E2)
- **API Compatibility**: 100% functional equivalence with Python implementation
- **Architecture Compliance**: Follow established patterns and reuse existing components

---

## Architecture Integration Analysis

### Existing Components to Integrate With

#### **Epic E1 Foundation Components**
1. **Configuration Management**: `internal/config/ConfigManager`
   - Server settings (host, port, WebSocket path)
   - Connection limits and timeouts
   - Authentication configuration
   - Hot-reload capability

2. **Logging Infrastructure**: `internal/logging/Logger`
   - Structured logging with correlation IDs
   - Log level management
   - Performance metrics logging

3. **Security Framework**: `internal/security/JWTHandler`
   - JWT token validation
   - Role-based access control (viewer, operator, admin)
   - Session management

#### **Epic E2 Camera Discovery Components**
4. **Camera Monitor**: `internal/camera/HybridCameraMonitor`
   - Real camera enumeration via `GetCameraList()`
   - Individual camera status via `GetCameraStatus()`
   - Real-time camera event notifications

### Integration Points

#### **Configuration Integration**
```go
// Server configuration from ConfigManager
type ServerConfig struct {
    Host            string `mapstructure:"host"`
    Port            int    `mapstructure:"port"`
    WebSocketPath   string `mapstructure:"websocket_path"`
    MaxConnections  int    `mapstructure:"max_connections"`
    ReadTimeout     time.Duration `mapstructure:"read_timeout"`
    WriteTimeout    time.Duration `mapstructure:"write_timeout"`
}
```

#### **Logging Integration**
```go
// Structured logging with correlation IDs
logger.WithFields(logrus.Fields{
    "client_id": clientID,
    "method":    method,
    "action":    "websocket_request",
}).Info("Processing WebSocket request")
```

#### **Camera Data Integration**
```go
// Use HybridCameraMonitor for API methods
cameras, err := cameraMonitor.GetCameraList()
if err != nil {
    return createErrorResponse(INTERNAL_ERROR, "Failed to get camera list")
}
```

#### **Authentication Integration**
```go
// JWT token validation and role checking
claims, err := jwtHandler.ValidateToken(authToken)
if err != nil {
    return createErrorResponse(AUTHENTICATION_REQUIRED, "Authentication failed")
}
```

---

## Existing Pattern Analysis

### Python WebSocket Patterns (Reference Implementation)

#### **Server Architecture**
- **Class Structure**: `WebSocketJsonRpcServer` with comprehensive connection management
- **Connection Handling**: Client connection tracking with authentication state
- **Method Registration**: Built-in method registration system with versioning
- **Error Handling**: Comprehensive JSON-RPC error codes and messages

#### **JSON-RPC Protocol**
```python
# JSON-RPC 2.0 request/response structure
@dataclass
class JsonRpcRequest:
    jsonrpc: str
    method: str
    id: Optional[Any] = None
    params: Optional[Dict[str, Any]] = None

@dataclass
class JsonRpcResponse:
    jsonrpc: str
    id: Optional[Any]
    result: Optional[Any] = None
    error: Optional[Dict[str, Any]] = None
```

#### **Performance Metrics**
```python
class PerformanceMetrics:
    def __init__(self) -> None:
        self.request_count = 0
        self.response_times = defaultdict(list)
        self.error_count = 0
        self.active_connections = 0
```

### Go Architecture Patterns (From Epic E1 & E2)

#### **Dependency Injection Pattern**
```go
type WebSocketServer struct {
    configManager   *config.ConfigManager
    logger          *logging.Logger
    cameraMonitor   *camera.HybridCameraMonitor
    jwtHandler      *security.JWTHandler
    // ... other dependencies
}

func NewWebSocketServer(
    configManager *config.ConfigManager,
    logger *logging.Logger,
    cameraMonitor *camera.HybridCameraMonitor,
    jwtHandler *security.JWTHandler,
) *WebSocketServer {
    // Constructor with dependency injection
}
```

#### **Interface-Based Design**
```go
type WebSocketHandler interface {
    HandleConnection(conn *websocket.Conn)
    RegisterMethod(name string, handler MethodHandler)
    BroadcastNotification(method string, params interface{})
}
```

---

## Component Reuse Plan

### **Reuse Existing Components**

#### **Configuration Management**
- **Component**: `internal/config/ConfigManager`
- **Usage**: Server settings, connection limits, authentication config
- **Integration**: Direct dependency injection in WebSocket server constructor
- **Benefits**: Consistent configuration management, hot-reload capability

#### **Logging Infrastructure**
- **Component**: `internal/logging/Logger`
- **Usage**: Structured logging with correlation IDs for all WebSocket operations
- **Integration**: Pass logger instance to WebSocket server
- **Benefits**: Consistent logging format, correlation ID tracking

#### **Camera Discovery System**
- **Component**: `internal/camera/HybridCameraMonitor`
- **Usage**: Real camera data for `get_camera_list` and `get_camera_status` methods
- **Integration**: Use monitor methods directly in JSON-RPC handlers
- **Benefits**: Real camera data, no duplicate discovery logic

#### **Security Framework**
- **Component**: `internal/security/JWTHandler`
- **Usage**: JWT token validation and role-based access control
- **Integration**: Authentication middleware for all protected methods
- **Benefits**: Consistent authentication, role-based permissions

### **No Duplicate Implementations**

#### **Prohibited Duplications**
- **Configuration System**: No new config loading or validation logic
- **Logging System**: No new logging infrastructure or formatters
- **Authentication System**: No new JWT handling or role management
- **Camera Discovery**: No new camera enumeration or status logic
- **Error Handling**: Reuse existing error types and patterns

#### **Architecture Compliance**
- **Single Responsibility**: Each component has one clear purpose
- **Dependency Injection**: Constructor-based dependency injection
- **Interface-Based Design**: Clear interfaces for testability
- **Real System Testing**: No over-mocking, use real components

---

## Test Strategy Aligned with Requirements

### **Real System Testing Approach**

#### **WebSocket Testing**
- **Real Connections**: Use gorilla/websocket for actual WebSocket connections
- **No Mocking**: Test with real WebSocket protocol implementation
- **Connection Management**: Test real connection lifecycle (connect, disconnect, reconnect)
- **Performance Testing**: Real connection stress testing (1000+ connections)

#### **Authentication Testing**
- **Real JWT Tokens**: Use actual JWT tokens with test secrets
- **Role-Based Access**: Test real role validation and permission checking
- **Session Management**: Test real session establishment and validation
- **Security Validation**: Test authentication failure scenarios

#### **Camera Data Integration**
- **Real Camera Discovery**: Use actual HybridCameraMonitor for camera data
- **Live Status Updates**: Test real camera status changes and notifications
- **Error Handling**: Test real camera discovery failures and edge cases
- **Performance Validation**: Test response times with real camera data

#### **Configuration Integration**
- **Real Config Loading**: Use actual ConfigManager for configuration
- **Hot-Reload Testing**: Test configuration changes during runtime
- **Environment Variables**: Test environment variable overrides
- **Validation Testing**: Test configuration validation and error handling

### **Test Categories**

#### **Unit Tests**
- **Method Logic**: Individual JSON-RPC method implementation
- **Error Handling**: Comprehensive error scenarios and edge cases
- **Authentication Logic**: JWT validation and role checking
- **Response Formatting**: JSON-RPC response structure validation

#### **Integration Tests**
- **Full WebSocket Server**: End-to-end WebSocket communication
- **Component Integration**: Integration with all existing components
- **Configuration Integration**: Full configuration system integration
- **Camera Data Integration**: Real camera discovery integration

#### **Performance Tests**
- **Connection Stress Testing**: 1000+ concurrent connections
- **Response Time Testing**: <50ms for status methods
- **Memory Usage Testing**: <60MB base footprint
- **CPU Usage Testing**: <50% under normal load

#### **Security Tests**
- **Authentication Testing**: JWT token validation
- **Authorization Testing**: Role-based access control
- **Input Validation**: Malicious input handling
- **Session Security**: Session management and security

---

## Implementation Plan

### **Story S3.1: WebSocket Infrastructure**

#### **Task T3.1.1: Implement gorilla/websocket server**
- **Technology**: Use gorilla/websocket library as specified
- **Architecture**: Follow Python WebSocketJsonRpcServer patterns
- **Features**: Connection management, message handling, error handling
- **Integration**: Use ConfigManager for server settings

#### **Task T3.1.2: Add connection management**
- **Client Tracking**: Track connected clients with authentication state
- **Connection Limits**: Implement max connections from configuration
- **Lifecycle Management**: Handle connect, disconnect, reconnect events
- **Resource Cleanup**: Proper cleanup on connection termination

#### **Task T3.1.3: Implement JSON-RPC 2.0 protocol**
- **Request/Response**: Full JSON-RPC 2.0 request and response handling
- **Error Codes**: Implement all JSON-RPC error codes from Python system
- **Method Registration**: Dynamic method registration system
- **Notification Support**: Support for JSON-RPC notifications

#### **Task T3.1.4: Add authentication middleware**
- **JWT Validation**: Integrate with existing JWTHandler
- **Role-Based Access**: Implement role checking for all methods
- **Session Management**: Track authenticated sessions
- **Security Logging**: Log authentication events and failures

#### **Task T3.1.5: Create WebSocket unit tests**
- **Method Testing**: Test individual WebSocket methods
- **Error Testing**: Test error handling and edge cases
- **Authentication Testing**: Test authentication scenarios
- **Performance Testing**: Test response times and throughput

#### **Integration Tasks (T3.1.8-T3.1.10)**
- **Configuration Integration**: Use ConfigManager for all server settings
- **Connection Limits**: Implement configuration-driven connection limits
- **Integration Testing**: End-to-end integration tests with all components

### **Story S3.2: Core JSON-RPC Methods**

#### **Task T3.2.1: Implement `ping` method**
- **Functionality**: Simple health check returning "pong"
- **Authentication**: Require viewer role
- **Performance**: <50ms response time
- **Testing**: Unit and integration tests

#### **Task T3.2.2: Implement `authenticate` method**
- **JWT Validation**: Use existing JWTHandler for token validation
- **Role Extraction**: Extract user role from JWT claims
- **Session Establishment**: Create authenticated session
- **Response Format**: Match Python implementation exactly

#### **Task T3.2.3: Implement `get_camera_list` method**
- **Camera Data**: Use HybridCameraMonitor.GetCameraList()
- **Response Format**: Match Python implementation exactly
- **Performance**: <50ms response time
- **Error Handling**: Handle camera discovery failures

#### **Task T3.2.4: Implement `get_camera_status` method**
- **Camera Data**: Use HybridCameraMonitor.GetCameraStatus()
- **Response Format**: Match Python implementation exactly
- **Performance**: <50ms response time
- **Error Handling**: Handle camera not found scenarios

#### **Integration Tasks (T3.2.8-T3.2.10)**
- **Camera Integration**: Full integration with camera discovery system
- **Configuration Integration**: Configuration-driven method behavior
- **End-to-End Testing**: Full flow from config to camera to API

---

## Performance Targets

### **Connection Performance**
- **Concurrent Connections**: 1000+ simultaneous WebSocket connections
- **Connection Establishment**: <100ms connection setup time
- **Connection Stability**: 99.9% connection uptime
- **Resource Usage**: <1MB memory per connection

### **Response Time Performance**
- **Status Methods**: <50ms response time (ping, get_camera_list, get_camera_status)
- **Control Methods**: <100ms response time (authenticate)
- **WebSocket Notifications**: <20ms delivery latency
- **Error Responses**: <10ms error response time

### **System Performance**
- **Memory Usage**: <60MB base footprint, <200MB with 1000 connections
- **CPU Usage**: <50% sustained usage under normal load
- **Throughput**: 1000+ requests/second
- **Error Rate**: <1% error rate under normal load

---

## Architecture Compliance

### **Single Responsibility Principle**
- **WebSocket Server**: Handle WebSocket connections and protocol
- **JSON-RPC Handler**: Handle JSON-RPC request/response logic
- **Authentication Middleware**: Handle authentication and authorization
- **Connection Manager**: Handle connection lifecycle and tracking

### **Dependency Injection Pattern**
- **Constructor Injection**: All dependencies injected via constructor
- **Interface-Based Design**: Clear interfaces for all dependencies
- **Testability**: Easy to mock dependencies for testing
- **Flexibility**: Easy to swap implementations

### **No Duplicate Implementations**
- **Configuration**: Use existing ConfigManager, no new config logic
- **Logging**: Use existing Logger, no new logging infrastructure
- **Authentication**: Use existing JWTHandler, no new auth logic
- **Camera Discovery**: Use existing HybridCameraMonitor, no new discovery logic

### **Real System Testing**
- **No Over-Mocking**: Use real components in integration tests
- **Real WebSocket**: Use actual gorilla/websocket connections
- **Real Authentication**: Use actual JWT tokens and validation
- **Real Camera Data**: Use actual camera discovery system

---

## Risk Mitigation

### **Integration Complexity**
- **Progressive Integration**: Integrate components one at a time
- **Comprehensive Testing**: Test each integration point thoroughly
- **Fallback Mechanisms**: Implement fallbacks for component failures
- **Monitoring**: Monitor integration points for issues

### **Performance Targets**
- **Continuous Benchmarking**: Regular performance testing
- **Optimization Cycles**: Iterative performance optimization
- **Resource Monitoring**: Monitor memory, CPU, and connection usage
- **Load Testing**: Regular load testing with realistic scenarios

### **API Compatibility**
- **Comprehensive Testing**: Test against Python implementation
- **Response Validation**: Validate all response formats match exactly
- **Error Code Validation**: Ensure error codes and messages match
- **Behavior Validation**: Ensure method behavior matches exactly

### **Authentication Security**
- **Proper JWT Validation**: Use existing secure JWT handler
- **Role-Based Access**: Implement proper role checking
- **Session Security**: Secure session management
- **Security Testing**: Comprehensive security testing

---

## Success Criteria

### **Functional Requirements**
- **API Compatibility**: 100% functional equivalence with Python system
- **Method Implementation**: All core methods implemented and tested
- **Authentication**: Full JWT authentication with role-based access
- **Integration**: Complete integration with all existing components

### **Performance Requirements**
- **Connection Capacity**: 1000+ concurrent WebSocket connections
- **Response Times**: <50ms for status methods, <100ms for control methods
- **Memory Usage**: <60MB base footprint, <200MB with 1000 connections
- **CPU Usage**: <50% under normal load

### **Quality Requirements**
- **Test Coverage**: >90% unit test coverage
- **Integration Testing**: Complete integration test suite
- **Performance Testing**: Comprehensive performance test suite
- **Security Testing**: Complete security test suite

### **Architecture Requirements**
- **Component Reuse**: No duplicate implementations
- **Dependency Injection**: Proper dependency injection pattern
- **Single Responsibility**: Each component has one clear purpose
- **Real System Testing**: No over-mocking, use real components

---

**Document Status**: Development plan ready for PM approval  
**Next Step**: Await PM approval before implementation  
**Implementation Timeline**: 3-4 sprints as specified in migration plan  
**Risk Level**: Low (follows established patterns, reuses existing components)
