# MediaMTX Camera Service

A distributed video sensor management service designed for OCI-compliant container environments. This service provides real-time video source discovery, streaming, recording, and management capabilities as part of a larger multi-sensor ecosystem with centralized service discovery. It will allow users to take snapshots and record videos from USB-V4L2 devices and STANAG 4609 UAV streams from external UAVs connected to the container.

## System Overview

The MediaMTX Camera Service is an always-on containerized service that manages both USB video devices and external RTSP feeds within a coordinated sensor ecosystem. It operates as a specialized video sensor container that registers with a central service discovery aggregator and provides standardized video services to client applications.

**Version:** 3.2  
**Date:** 2025-01-15  
**Status:** Production Architecture Documentation  
**Document Type:** System Architecture Specification

---

## 1. System Context

### 1.1 System Boundaries

```plantuml
@startuml SystemContext
title System Context - MediaMTX Camera Service

actor "Client Applications" as client
rectangle "MediaMTX Camera Service" as service
database "MediaMTX Server" as mediamtx
component "USB V4L2 Cameras" as cameras
cloud "RTSP UAV Sources" as rtsp

client --> service : WebSocket JSON-RPC 2.0
service --> mediamtx : HTTP REST API\n(Path Management)
service --> cameras : V4L2 System Calls\n(Direct Access)
rtsp --> mediamtx : RTSP Streams\n(UAV Sources)
mediamtx --> cameras : FFmpeg Processing

note right of service
  Core Capabilities:
  • Real-time video source discovery
  • Snapshot capture from USB cameras
  • Video recording capabilities
  • MediaMTX path management for UAV streams
  • Service discovery registration
end note

note right of rtsp
  RTSP UAV Sources:
  • STANAG 4609 compliant streams
  • External video sources
  • Consumed by MediaMTX Server
  • NOT directly accessed by service
end note

@enduml
```

### 1.2 Quality Attributes

| Attribute | Target | Measurement |
|-----------|--------|-------------|
| **Performance** | <100ms response time | 95th percentile API calls |
| **Concurrency** | 1000+ connections | Simultaneous WebSocket clients |
| **Availability** | 99.9% uptime | System operational time |
| **Reliability** | <0.1% error rate | Failed operations ratio |

---

## 2. External Interface Architecture

### 2.1 Exposed Interfaces (Inbound)

**JSON-RPC 2.0 API (Primary External Interface)**
- **Protocol:** WebSocket over TCP
- **Port:** 8002
- **Documentation:** `docs/api/json_rpc_methods.md`
- **Authentication:** JWT Bearer tokens
- **Clients:** Web browsers, mobile apps, desktop applications

```plantuml
@startuml ExposedInterface
title Exposed Interface - JSON-RPC 2.0 API

actor "Client" as client
interface "WebSocket\nPort 8002" as ws
component "JSON-RPC Handler" as rpc
component "Authentication" as auth
component "MediaMTX Controller" as controller

client --> ws : WebSocket Connection
ws --> auth : Token Validation
auth --> rpc : Authenticated Request
rpc --> controller : Business Logic

note bottom of rpc
  Supported Methods:
  • get_camera_list()
  • take_snapshot()
  • start_recording()
  • stop_recording()
  • get_camera_status()
  
  Documentation: docs/api/json_rpc_methods.md
end note

@enduml
```

### 2.2 Consumed Interfaces (Outbound)

**MediaMTX REST API (External Dependency)**
- **Protocol:** HTTP/1.1
- **Endpoint:** http://localhost:9997/v3/
- **Purpose:** Stream path management, configuration
- **Required Version:** MediaMTX v1.0+

**V4L2 Hardware Interface**
- **Protocol:** Linux system calls
- **Devices:** /dev/video* character devices
- **Purpose:** Direct camera hardware access

```plantuml
@startuml ConsumedInterfaces
title Consumed Interfaces - External Dependencies

component "MediaMTX Controller" as controller
interface "HTTP REST\nPort 9997" as http
database "MediaMTX Server" as mediamtx
interface "V4L2 System Calls" as v4l2
component "USB Cameras" as cameras

controller --> http : Path Management
http --> mediamtx : Stream Configuration
controller --> v4l2 : Hardware Access
v4l2 --> cameras : Device Control

note bottom of http
  MediaMTX API Endpoints:
  • GET /v3/config/paths/list
  • POST /v3/config/paths/add/{name}
  • DELETE /v3/config/paths/delete/{name}
end note

@enduml
```

---

## 3. Internal Component Architecture

### 3.1 Component Structure

```plantuml
@startuml InternalComponents
title Internal Component Architecture

package "MediaMTX Camera Service" {
    component "WebSocket Server" as ws
    component "MediaMTX Controller" as controller
    component "Camera Monitor" as camera
    component "Security Framework" as security
    
    interface "ControllerAPI" as ctrl_api
    interface "CameraAPI" as cam_api
    interface "SecurityAPI" as sec_api
}

ws --> ctrl_api
controller ..> ctrl_api
controller --> cam_api
camera ..> cam_api
ws --> sec_api
security ..> sec_api

note right of ws
  Architecture Rules:
  • WebSocket Server contains NO business logic
  • All operations delegate to MediaMTX Controller
  • No direct component-to-component calls
  • Interface-based dependency injection
end note

@enduml
```

### 3.2 Component Responsibilities

**WebSocket Server (Protocol Layer)**
- JSON-RPC 2.0 protocol implementation
- WebSocket connection management (1000+ concurrent)
- Authentication enforcement
- **Constraint:** NO business logic - delegates all operations

**MediaMTX Controller (Business Logic Layer)**
- Camera operations coordination
- Stream lifecycle management
- API abstraction (camera0 ↔ /dev/video0)
- **Pattern:** Single Source of Truth for all operations

**Camera Monitor (Hardware Abstraction Layer)**
- USB camera detection via V4L2
- Real-time status monitoring
- Hardware capability probing
- **Integration:** Interface-based design with dependency injection

**Security Framework (Cross-Cutting Layer)**
- JWT token management
- Role-based access control (viewer/operator/admin)
- Session management
- **Pattern:** Middleware integration with existing configuration

### 3.3 Internal Interface Contracts

```plantuml
@startuml InterfaceContracts
title Internal Interface Contracts

interface ControllerAPI {
    +GetCameraList() : CameraListResponse
    +GetCameraStatus(device : string) : CameraStatusResponse
    +TakeSnapshot(device : string, path : string) : SnapshotResponse
    +StartRecording(device : string) : RecordingResponse
    +StopRecording(device : string) : RecordingResponse
}

interface CameraAPI {
    +Start(ctx : Context) : error
    +Stop() : error
    +GetConnectedCameras() : map[string]CameraDevice
    +GetDevice(path : string) : CameraDevice
}

interface SecurityAPI {
    +ValidateToken(token : string) : JWTClaims
    +CheckPermission(role : string, method : string) : bool
    +CreateSession(userID : string) : Session
}

@enduml
```

---

## 4. Process Architecture

### 4.1 Authentication Flow

```plantuml
@startuml AuthenticationFlow
title Authentication Flow

participant "Client" as C
participant "WebSocket Server" as WS
participant "Security Framework" as S
participant "MediaMTX Controller" as MC

C -> WS : WebSocket Connection
WS -> S : Validate Connection
S --> WS : Connection Authorized
WS --> C : Connection Established

C -> WS : authenticate(credentials)
WS -> S : Validate Credentials
S -> S : Generate JWT Token
S --> WS : JWT Token + Role
WS --> C : Authentication Success

note over C, MC
All subsequent calls include JWT token
for authentication and authorization
end note

@enduml
```

### 4.2 Snapshot Capture Flow

```plantuml
@startuml SnapshotFlow
title Snapshot Capture Flow (Multi-Tier)

participant "Client" as C
participant "WebSocket Server" as WS
participant "MediaMTX Controller" as MC
participant "Camera Monitor" as CM
participant "Hardware" as H

C -> WS : take_snapshot(camera0)
WS -> MC : TakeSnapshot(camera0)

alt Tier 1: Direct V4L2 Capture
    MC -> CM : CaptureDirectV4L2(/dev/video0)
    CM -> H : Direct hardware access
    H --> CM : Frame data
    CM --> MC : Snapshot created
else Tier 2: RTSP Stream Reuse
    MC -> MC : Check existing RTSP stream
    MC -> MC : Capture from stream
else Tier 3: On-Demand Activation
    MC -> MC : Activate MediaMTX path
    MC -> MC : Capture from activated stream
end

MC --> WS : SnapshotResponse
WS --> C : Snapshot result

@enduml
```

### 4.3 System Startup Coordination

```plantuml
@startuml StartupFlow
title System Startup Coordination

start

:Load Configuration;
:Initialize Security Framework;
:Start Camera Monitor;
:Initialize MediaMTX Controller;
:Start WebSocket Server;

:System Operational;

note right
Progressive Readiness Pattern:
• System accepts connections immediately
• Features become available as components initialize
• No blocking startup dependencies
• Clear status communication to clients
end note

stop

@enduml
```

---

## 5. Physical Architecture

### 5.1 Deployment Architecture

```plantuml
@startuml DeploymentArchitecture
title Deployment Architecture

node "Container Host" {
    node "Camera Service Container" {
        artifact "MediaMTX Camera Service" as service
        database "Configuration Files" as config
        database "Recording Storage" as storage
    }
    
    node "MediaMTX Container" {
        artifact "MediaMTX Server" as mediamtx
    }
    
    component "USB Cameras" as cameras
}

cloud "External Network" {
    actor "Client Applications" as clients
    cloud "UAV RTSP Sources" as uav
}

service --> mediamtx : HTTP API\nContainer Network
service --> cameras : V4L2 API\nDevice Passthrough
uav --> mediamtx : RTSP Streams\nExternal Network
clients --> service : WebSocket API\nHost Network

@enduml
```

### 5.3 Container Deployment Strategy

**Option 1: Separate Containers (Recommended)**
- **Advantages:**
  - Independent scaling of MediaMTX and camera service
  - Separate lifecycle management and updates
  - Better resource isolation and fault isolation
  - Follows microservices architecture principles

**Option 2: Single Container**
- **Advantages:**
  - Simpler deployment and management
  - Faster inter-process communication
  - Shared resource utilization
  - Reduced network overhead

**Recommendation:** Separate containers for production deployments to enable independent scaling and lifecycle management. Single container acceptable for development or resource-constrained environments.

### 5.2 Network Architecture

| Port | Protocol | Purpose | Security |
|------|----------|---------|----------|
| 8002 | WebSocket | Client API | JWT Authentication |
| 8003 | HTTP | Health checks | Internal only |
| 9997 | HTTP | MediaMTX API | Internal only |
| 8554 | RTSP | Media streaming | Internal only |

---

## 6. Data Architecture

### 6.1 Core Data Models

```plantuml
@startuml DataModels
title Core Data Models

class CameraDevice {
    +Path : string
    +Name : string
    +Status : string
    +Capabilities : V4L2Capabilities
    +LastSeen : time
    +Error : error
}

class V4L2Capabilities {
    +DriverName : string
    +CardName : string
    +BusInfo : string
    +Version : string
    +Capabilities : array
    +DeviceCaps : array
}

class Session {
    +ID : string
    +UserID : string
    +Role : string
    +Created : time
    +LastActivity : time
    +IsActive : bool
}

class SnapshotResult {
    +Device : string
    +FilePath : string
    +Size : int64
    +Created : time
    +TierUsed : int
    +CaptureTime : float64
}

CameraDevice *-- V4L2Capabilities
Session ||--o{ SnapshotResult

@enduml
```

### 6.2 Configuration Schema

```plantuml
@startuml ConfigurationSchema
title Configuration Schema

class Config {
    +ServerConfig Server
    +CameraConfig Camera
    +MediaMTXConfig MediaMTX
    +SecurityConfig Security
    +LoggingConfig Logging
}

class ServerConfig {
    +string Host
    +int Port
    +string WebSocketPath
    +int MaxConnections
}

class SecurityConfig {
    +string JWTSecretKey
    +int JWTExpiryHours
    +int RateLimitRequests
    +string RateLimitWindow
}

class CameraConfig {
    +float64 PollInterval
    +[]int DeviceRange
    +float64 DetectionTimeout
    +bool EnableCapabilityDetection
}

Config *-- ServerConfig
Config *-- SecurityConfig
Config *-- CameraConfig

@enduml
```

---

## 7. Security Architecture

### 7.1 Security Model

```plantuml
@startuml SecurityModel
title Security Architecture

rectangle "Security Layers" {
    (Network Security) --> (Authentication)
    (Authentication) --> (Authorization)
    (Authorization) --> (Input Validation)
    (Input Validation) --> (Audit Logging)
}

rectangle "Authentication Components" {
    (JWT Tokens) --> (Session Management)
    (Session Management) --> (Role Assignment)
}

rectangle "Authorization Components" {
    (Permission Matrix) --> (Method-Level RBAC)
    (Method-Level RBAC) --> (Resource Access Control)
}

@enduml
```

### 7.2 Role-Based Access Control

| Role | Permissions | Use Case |
|------|-------------|----------|
| **viewer** | Read-only access to status and listings | Monitoring dashboards |
| **operator** | Camera control + viewer permissions | Day-to-day operations |
| **admin** | Full system access + metrics | System administration |

### 7.3 Security Implementation

```plantuml
@startuml SecurityImplementation
title Security Implementation Flow

start

:Client Request;
:Rate Limiting Check;

if (Rate Limit Exceeded?) then (yes)
    :Return Rate Limit Error;
    stop
else (no)
    :JWT Token Validation;
endif

if (Token Valid?) then (yes)
    :Extract Role from Token;
    :Check Method Permissions;
    
    if (Permission Granted?) then (yes)
        :Execute Method;
        :Log Success Event;
        :Return Response;
    else (no)
        :Log Authorization Failure;
        :Return Authorization Error;
    endif
else (no)
    :Log Authentication Failure;
    :Return Authentication Error;
endif

stop

@enduml
```

---

## 8. Quality Attributes

### 8.1 Performance Architecture

**Response Time Optimization:**
- **Tier 0 Snapshots:** Direct V4L2 access (<200ms)
- **Connection Pooling:** Reuse MediaMTX connections
- **Event System:** O(log n) client notification scaling
- **Memory Management:** Object pooling for high-frequency operations

**Concurrency Design:**
- **Goroutine-Based:** Non-blocking concurrent operations
- **Channel Communication:** Lock-free inter-component communication
- **Context Cancellation:** Graceful operation termination
- **Resource Limiting:** Bounded goroutine pools

### 8.2 Reliability Architecture

**Fault Tolerance:**
- **Multi-Tier Fallback:** Snapshot capture tier degradation
- **Circuit Breaker:** MediaMTX communication protection
- **Health Monitoring:** Component status tracking
- **Graceful Degradation:** Partial functionality under failure

**Error Handling:**
- **Structured Errors:** Consistent error response format
- **Error Propagation:** Clean error context preservation
- **Recovery Mechanisms:** Automatic retry with exponential backoff
- **Failure Isolation:** Component failures don't cascade

### 8.3 Scalability Architecture

**Horizontal Scaling Readiness:**
- **Stateless Design:** Session state externalization capability
- **Resource Separation:** Compute vs storage separation
- **Event Distribution:** External event system integration ready
- **Service Discovery:** Container orchestration compatibility

---

## 9. Design Principles

### 9.1 Architectural Principles Applied

**Single Responsibility Principle:**
- Each component has one clear responsibility
- Clean separation between protocol, business logic, and hardware
- Interface-based design enables component substitution

**Dependency Inversion Principle:**
- High-level modules don't depend on low-level modules
- Both depend on abstractions (interfaces)
- Enables testing and component replacement

**Open/Closed Principle:**
- Components open for extension via interfaces
- Closed for modification through stable contracts
- Plugin architecture ready for future extensions

---

## 10. Architectural Debt

### 10.1 Current Technical Debt

**Performance Optimization Debt:**
- FFmpeg process management could be optimized with process pooling
- Memory allocation patterns could benefit from object pooling
- Network connection pooling not yet implemented

**Monitoring and Observability Debt:**
- Distributed tracing not implemented
- Advanced metrics collection could be enhanced
- Performance analytics could be more comprehensive

**Extensibility Debt:**
- Plugin architecture interfaces defined but not fully implemented
- External authentication providers not yet supported
- Advanced camera types (IP cameras) have basic support only

### 10.2 Debt Prioritization

**High Priority:**
- Process management optimization for production scalability
- Enhanced error handling and recovery mechanisms

**Medium Priority:**
- Advanced monitoring and observability features
- External authentication provider integration

**Low Priority:**
- Plugin architecture full implementation
- Advanced analytics integration points

---

**Document Status:** Production Architecture Documentation  
**Last Updated:** 2025-01-15  
**Review Cycle:** Quarterly architecture reviews  
**Document Maintenance:** Architecture changes require PM and IV&V approval