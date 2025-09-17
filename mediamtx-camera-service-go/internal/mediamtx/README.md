# MediaMTX Integration Module

**Version:** 2.0  
**Date:** 2025-01-15  
**Status:** Production Architecture Documentation  
**Document Type:** Module Architecture Specification

---

## 1. Module Overview

The MediaMTX Integration Module serves as **Layer 5 (Orchestration)** in the system architecture, implementing the single source of truth for all video operations and business logic coordination. This module provides the central orchestration layer between the WebSocket API layer and the underlying hardware/service layers.

### 1.1 Architecture Positioning

```plantuml
@startuml ModulePositioning
title MediaMTX Module - Architecture Positioning

package "Layer 6: API" {
    component "WebSocket Server" as WS
    component "JSON-RPC Methods" as RPC
}

package "Layer 5: Orchestration" #lightcoral {
    component "MediaMTX Controller" as Controller
    note right of Controller
        Single Source of Truth
        Business Logic Coordination
        API Abstraction Layer
    end note
}

package "Layer 4: Business Logic" {
    component "Recording Manager" as RM
    component "Snapshot Manager" as SM
}

package "Layer 3: Managers" {
    component "Path Manager" as PM
    component "Stream Manager" as STM
    component "FFmpeg Manager" as FM
}

package "Layer 2: Core Services" {
    component "MediaMTX Client" as Client
    component "Health Monitor" as HM
}

WS --> Controller : Delegates ALL operations
RPC --> Controller : No business logic in API
Controller --> RM : Orchestrates
Controller --> SM : Orchestrates
Controller --> PM : Coordinates
Controller --> STM : Coordinates
Controller --> Client : Integrates

@enduml
```

### 1.2 Key Responsibilities

- **Single Source of Truth**: All business logic resides in the MediaMTX controller
- **API Abstraction**: Maps external identifiers (camera0) to internal paths (/dev/video0)
- **Component Orchestration**: Coordinates all managers and services
- **Event-Driven Architecture**: Real-time notifications and progressive readiness
- **Stateless Recording**: MediaMTX API as the authoritative recording state source

---

## 2. Component Architecture

### 2.1 Core Component Structure

```plantuml
@startuml ComponentStructure
title MediaMTX Module - Component Structure

package "MediaMTX Integration Module" {
    
    class Controller {
        +client: MediaMTXClient
        +healthMonitor: HealthMonitor
        +pathManager: PathManager
        +streamManager: StreamManager
        +recordingManager: RecordingManager
        +snapshotManager: SnapshotManager
        +cameraMonitor: CameraMonitor
        +configIntegration: ConfigIntegration
        --
        +Start(ctx) error
        +Stop(ctx) error
        +IsReady() bool
        +GetCameraList(ctx) (*CameraListResponse, error)
        +StartRecording(ctx, device, options) (*RecordingResponse, error)
        +TakeSnapshot(ctx, device, options) (*SnapshotResponse, error)
    }
    
    class RecordingManager {
        +pathManager: PathManager
        +configIntegration: ConfigIntegration
        +timers: map[string]*time.Timer
        --
        +StartRecording(ctx, device) error
        +StopRecording(ctx, device) error
        +IsRecording(ctx, device) (bool, error)
        +CleanupOldRecordings(ctx, maxAge, maxCount) error
    }
    
    class SnapshotManager {
        +streamManager: StreamManager
        +ffmpegManager: FFmpegManager
        +configIntegration: ConfigIntegration
        --
        +TakeSnapshot(ctx, device, outputPath) (*SnapshotResult, error)
        +CleanupOldSnapshots(ctx, maxAge, maxCount) error
        +GetSnapshotMetadata(filePath) (map[string]interface{}, error)
    }
    
    class PathManager {
        +client: MediaMTXClient
        +configIntegration: ConfigIntegration
        +createGroup: singleflight.Group
        +metrics: PathManagerMetrics
        --
        +CreatePath(ctx, name, source, options) error
        +DeletePath(ctx, name) error
        +PathExists(ctx, name) bool
        +PatchPath(ctx, name, config) error
        +ListPaths(ctx) (*PathList, error)
    }
    
    class StreamManager {
        +pathManager: PathManager
        +ffmpegManager: FFmpegManager
        +configIntegration: ConfigIntegration
        --
        +StartStream(ctx, devicePath) (*Path, error)
        +StopStream(ctx, device) error
        +GetStreamStatus(ctx, device) (*StreamStatus, error)
        +buildFFmpegCommand(devicePath, streamName) string
    }
    
    class HealthMonitor {
        +client: MediaMTXClient
        +configIntegration: ConfigIntegration
        +circuitBreaker: CircuitBreaker
        --
        +Start(ctx) error
        +Stop() error
        +IsHealthy() bool
        +GetHealth() (*HealthStatus, error)
        +GetMetrics() (map[string]interface{}, error)
    }
    
    Controller *-- RecordingManager
    Controller *-- SnapshotManager
    Controller *-- PathManager
    Controller *-- StreamManager
    Controller *-- HealthMonitor
    
    RecordingManager --> PathManager
    SnapshotManager --> StreamManager
    SnapshotManager --> FFmpegManager
    StreamManager --> PathManager
    PathManager --> MediaMTXClient
    HealthMonitor --> MediaMTXClient
}

@enduml
```

### 2.2 Component Responsibilities Matrix

| Component | Layer | Primary Responsibility | Key Dependencies |
|-----------|-------|----------------------|------------------|
| **Controller** | 5 - Orchestration | Single source of truth, API abstraction, component coordination | All managers, CameraMonitor |
| **RecordingManager** | 4 - Business Logic | Stateless recording via MediaMTX API, auto-stop timers | PathManager, ConfigIntegration |
| **SnapshotManager** | 4 - Business Logic | Multi-tier snapshot capture (V4L2→FFmpeg→RTSP) | StreamManager, FFmpegManager |
| **PathManager** | 3 - Managers | MediaMTX path lifecycle, idempotent operations, per-path mutex | MediaMTXClient, ConfigIntegration |
| **StreamManager** | 3 - Managers | Stream lifecycle, FFmpeg coordination, on-demand processes | PathManager, FFmpegManager |
| **HealthMonitor** | 2 - Core Services | Circuit breaker pattern, health monitoring | MediaMTXClient, ConfigIntegration |

---

## 3. Data Flow Architecture

### 3.1 Recording Flow

```plantuml
@startuml RecordingFlow
title Recording Flow - Stateless Architecture

participant "WebSocket API" as API
participant "Controller" as C
participant "Recording Manager" as RM
participant "Path Manager" as PM
participant "MediaMTX Server" as MTX

API -> C : StartRecording(camera0, options)
C -> RM : StartRecording(camera0, options)

note over RM : Stateless Recording Pattern
RM -> RM : Validate camera0 → /dev/video0 mapping
RM -> MTX : GET /v3/config/paths/get/camera0
RM -> MTX : PATCH /v3/config/paths/patch/camera0 {"record": true}

note over MTX
Recording Configuration:
- record: true
- recordPath: "/opt/recordings/camera0_%Y-%m-%d_%H-%M-%S.mp4"
- recordFormat: "fmp4" (STANAG 4609 compatible)
end note

RM -> RM : Start RTSP keepalive (if needed)
RM -> RM : Set auto-stop timer (if duration specified)

MTX -> MTX : Start FFmpeg process (on-demand)
MTX -> MTX : Begin recording to file

RM --> C : Recording started (no session state)
C --> API : RecordingResponse

note over MTX
Single path "camera0" provides:
• Live streaming: rtsp://localhost:8554/camera0
• File recording: /opt/recordings/camera0_2024-01-15.mp4
• Both operate simultaneously
end note

@enduml
```

### 3.2 Multi-Tier Snapshot Architecture

```plantuml
@startuml SnapshotTiers
title Multi-Tier Snapshot Architecture

start
:Snapshot Request;

partition "Tier 0: V4L2 Direct (FASTEST)" {
    :Direct V4L2 capture;
    if (USB device?) then (yes)
        :Capture frame directly;
        :~100ms latency;
        stop
    else (no)
    endif
}

partition "Tier 1: FFmpeg Direct" {
    :FFmpeg from device;
    if (Device accessible?) then (yes)
        :FFmpeg capture;
        :~200ms latency;
        stop
    else (no)
    endif
}

partition "Tier 2: RTSP Reuse" {
    :Check existing stream;
    if (Stream active?) then (yes)
        :Capture from RTSP;
        :~300ms latency;
        stop
    else (no)
    endif
}

partition "Tier 3: Stream Activation" {
    :Create MediaMTX path;
    :Start FFmpeg;
    :Capture from stream;
    :~500ms latency;
    stop
}

@enduml
```

### 3.3 Path Management Architecture

```plantuml
@startuml PathManagement
title Path Management - Idempotent Operations

participant "Controller" as C
participant "Path Manager" as PM
participant "MediaMTX API" as API
database "singleflight.Group" as SG

C -> PM : CreatePath("camera0", "/dev/video0", options)
PM -> SG : Do("camera0", createFunc)

note over SG : Prevents concurrent creation\nof same path

alt Path Creation
    PM -> API : POST /v3/config/paths/add/camera0
    alt Success
        API --> PM : 200 OK
        PM --> C : Success
    else Already Exists
        API --> PM : 409 Conflict "path already exists"
        PM -> PM : isAlreadyExistsError() = true
        PM --> C : Success (idempotent)
    else Other Error
        API --> PM : Error
        PM --> C : Error with context
    end
end

note over PM
Architectural Guarantees:
• Create is idempotent
• Per-path mutex prevents races
• Exponential backoff on retries
• Comprehensive error context
end note

@enduml
```

---

## 4. Advanced Architecture Patterns

### 4.1 Optional Component Pattern

```plantuml
@startuml OptionalComponents
title Optional Component Pattern

class Controller {
    +cameraMonitor: CameraMonitor ✓
    +healthMonitor: HealthMonitor ✓
    +recordingManager: RecordingManager ✓
    +externalDiscovery: ExternalStreamDiscovery ❓
    +pathIntegration: PathIntegration ❓
    --
    +hasExternalDiscovery() bool
    +checkOptionalComponent(component) bool
}

note right of Controller
Optional Component Rules:
1. May be nil based on configuration
2. ALL methods MUST check for nil
3. Return graceful errors for nil components
4. Document optional nature in constructor
end note

class ExternalStreamDiscovery {
    +networkScanner: NetworkScanner
    +configIntegration: ConfigIntegration
    --
    +DiscoverStreams(ctx) ([]*ExternalStream, error)
    +AddExternalStream(ctx, streamURL) error
    +RemoveExternalStream(ctx, streamURL) error
}

note bottom of ExternalStreamDiscovery
Optional: Only initialized if
external streams are enabled
in configuration
end note

Controller o-- ExternalStreamDiscovery : optional

@enduml
```

### 4.2 Configuration Integration Pattern

```plantuml
@startuml ConfigIntegration
title Configuration Integration Pattern

class ConfigManager {
    -config: Config
    +LoadConfig(path: string) error
    +GetConfig() *Config
    +RegisterLoggingConfigurationUpdates()
}

class ConfigIntegration {
    -configManager: *ConfigManager
    -logger: *logging.Logger
    +GetMediaMTXConfig() *MediaMTXConfig
    +GetRecordingConfig() *RecordingConfig
    +GetSnapshotConfig() *SnapshotConfig
    +GetStreamingConfig() *StreamingConfig
}

class Controller {
    -configIntegration: *ConfigIntegration
}

class RecordingManager {
    -configIntegration: *ConfigIntegration
}

class SnapshotManager {
    -configIntegration: *ConfigIntegration
}

ConfigManager --> ConfigIntegration : provides config
ConfigIntegration --> Controller : injected
ConfigIntegration --> RecordingManager : injected
ConfigIntegration --> SnapshotManager : injected

note bottom of ConfigIntegration
Pattern Rules:
1. ALL components receive ConfigIntegration
2. NO direct ConfigManager access
3. Type-safe configuration access
4. Centralized defaults and validation
end note

@enduml
```

---

## 5. Integration Architecture

### 5.1 WebSocket Integration

```plantuml
@startuml WebSocketIntegration
title WebSocket Integration - Delegation Pattern

package "WebSocket Layer (NO Business Logic)" {
    class WebSocketServer {
        +mediaMTXController: MediaMTXControllerAPI
        --
        +handleJSONRPC(message) (*JsonRpcResponse, error)
    }
    
    class JSONRPCMethods {
        +controller: MediaMTXControllerAPI
        --
        +getCameraList(params, client) (*JsonRpcResponse, error)
        +startRecording(params, client) (*JsonRpcResponse, error)
        +takeSnapshot(params, client) (*JsonRpcResponse, error)
    }
}

package "MediaMTX Layer (ALL Business Logic)" {
    interface MediaMTXControllerAPI {
        +GetCameraList(ctx) (*CameraListResponse, error)
        +StartRecording(ctx, device, options) (*RecordingResponse, error)
        +TakeSnapshot(ctx, device, options) (*SnapshotResponse, error)
        +GetSystemHealth(ctx) (*GetHealthResponse, error)
        +GetSystemMetrics(ctx) (*GetSystemMetricsResponse, error)
    }
    
    class Controller {
        // Implementation of MediaMTXControllerAPI
    }
}

WebSocketServer --> MediaMTXControllerAPI : Delegates ALL operations
JSONRPCMethods --> MediaMTXControllerAPI : NO business logic
Controller ..|> MediaMTXControllerAPI : Implements

note bottom
Architectural Constraint:
WebSocket layer contains ZERO business logic
ALL operations delegated to MediaMTX Controller
end note

@enduml
```

### 5.2 Hardware Integration

```plantuml
@startuml HardwareIntegration
title Hardware Integration - Abstraction Layer

package "External API Layer" #lightblue {
    component "Client sees: camera0" as API
}

package "Controller Abstraction" #lightgreen {
    class Controller {
        +GetCameraForDevicePath(devicePath) (string, bool)
        +GetDevicePathForCamera(cameraID) (string, bool)
        --
        +mapDeviceToCamera(devicePath) string
        +mapCameraToDevice(cameraID) string
    }
    
    note right of Controller
    Mapping Rules:
    camera0 ↔ /dev/video0
    camera1 ↔ /dev/video1
    camera2 ↔ /dev/video2
    end note
}

package "Hardware Layer" #lightyellow {
    component "Hardware: /dev/video0" as HW
    component "Camera Monitor" as CM
}

API --> Controller : camera0 (abstract)
Controller --> HW : /dev/video0 (concrete)
Controller --> CM : Device discovery events

note bottom
CRITICAL Rules:
1. External APIs ONLY use camera identifiers
2. Internal operations use discovered device paths
3. Controller manages ALL mapping
4. NEVER expose device paths to clients
5. Mapping based on Camera Monitor discovery
end note

@enduml
```

---

## 6. Performance and Quality Architecture

### 6.1 Circuit Breaker Pattern

```plantuml
@startuml CircuitBreaker
title Circuit Breaker Pattern - MediaMTX Health Monitor

state "Closed\n(Normal)" as Closed
state "Open\n(Failing)" as Open
state "Half-Open\n(Testing)" as HalfOpen

Closed --> Open : Failure threshold\n(5 failures in 10s)
Open --> HalfOpen : After timeout\n(30 seconds)
HalfOpen --> Closed : Test request succeeds
HalfOpen --> Open : Test request fails

note right of Open
When Open:
• Fail fast (no MediaMTX calls)
• Return cached data if available
• Log circuit breaker state
• Emit health events to clients
end note

note bottom of HalfOpen
Half-Open Testing:
• Allow single test request
• Monitor response time
• Exponential backoff on failures
• Automatic state transitions
end note

@enduml
```

### 6.2 Performance Metrics

| Operation | Target Latency | Architecture Optimization |
|-----------|----------------|---------------------------|
| **Camera List** | <50ms | Cached discovery results |
| **Start Recording** | <200ms | Stateless MediaMTX PATCH |
| **Stop Recording** | <100ms | Direct API call, no session cleanup |
| **Snapshot Tier 0** | <100ms | Direct V4L2 capture |
| **Snapshot Tier 1** | <200ms | Direct FFmpeg capture |
| **Snapshot Tier 2** | <300ms | RTSP stream reuse |
| **Snapshot Tier 3** | <500ms | Stream activation + capture |
| **Health Check** | <30ms | Circuit breaker with caching |

---

## 7. Data Models and Types

### 7.1 Core Response Types

```plantuml
@startuml ResponseTypes
title API Response Types

class CameraListResponse {
    +Cameras: []CameraInfo
    +Total: int
    +Timestamp: string
}

class CameraInfo {
    +Device: string
    +Name: string
    +Status: string
    +Resolution: string
    +FrameRate: int
    +Capabilities: []string
    +StreamURL: string
    +IsRecording: bool
}

class RecordingResponse {
    +Device: string
    +Status: string
    +Filename: string
    +StartTime: string
    +Duration: int
    +AutoStop: bool
    +Format: string
}

class SnapshotResponse {
    +Device: string
    +Filename: string
    +FilePath: string
    +FileSize: int64
    +Resolution: string
    +Format: string
    +TierUsed: int
    +CaptureTime: float64
    +Timestamp: string
}

class GetHealthResponse {
    +Status: string
    +Uptime: string
    +Version: string
    +Components: map[string]string
    +Checks: []interface{}
    +Timestamp: string
}

class GetSystemMetricsResponse {
    +Timestamp: string
    +CPUUsage: float64
    +MemoryUsage: float64
    +DiskUsage: float64
    +Goroutines: int
    +NetworkIn: int64
    +NetworkOut: int64
    +LoadAverage: float64
    +Connections: int64
}

@enduml
```

---

## 8. Architecture Quality Assessment

### 8.1 Current Architecture Strengths

1. **✅ Single Source of Truth**: MediaMTX Controller centralizes all business logic
2. **✅ Clean Abstraction**: Clear separation between API identifiers and hardware paths
3. **✅ Stateless Recording**: MediaMTX API as authoritative recording state
4. **✅ Multi-Tier Snapshots**: Intelligent fallback system with performance optimization
5. **✅ Event-Driven Patterns**: Progressive readiness and real-time notifications
6. **✅ Optional Components**: Flexible configuration-based component initialization
7. **✅ Circuit Breaker**: Fault tolerance for external service dependencies
8. **✅ Configuration Integration**: Centralized, type-safe configuration access

### 8.2 Architectural Constraints and Limitations

1. **TODO Dependencies**: Several components contain placeholder implementations that limit production readiness
2. **Configuration Hardcoding**: Some configuration values are hardcoded rather than externalized
3. **Metrics Completeness**: Comprehensive metrics collection is partially implemented
4. **Error Context**: Some error scenarios lack detailed diagnostic information

### 8.3 Design Principles Compliance

| Principle | Compliance | Evidence |
|-----------|------------|----------|
| **Single Responsibility** | ✅ High | Each component has well-defined, focused responsibilities |
| **Dependency Inversion** | ✅ High | Interface-based design with dependency injection |
| **Open/Closed Principle** | ✅ Medium | Extensible via interfaces, some concrete dependencies |
| **Interface Segregation** | ✅ High | Focused interfaces with specific responsibilities |
| **Don't Repeat Yourself** | ✅ Medium | ConfigIntegration pattern reduces duplication |

### 8.4 Architectural Trade-offs

#### **Chosen: Stateless Recording**
- **Benefits**: Simplified state management, better scalability, MediaMTX as source of truth
- **Trade-offs**: Requires API queries for status, limited local state optimization

#### **Chosen: Multi-Tier Snapshots**
- **Benefits**: Performance optimization, graceful degradation, broad device support
- **Trade-offs**: Increased complexity, multiple code paths to maintain

#### **Chosen: Optional Components**
- **Benefits**: Flexible deployment configurations, reduced resource usage
- **Trade-offs**: Nil-checking overhead, increased testing complexity

---

## 9. Conclusion

The MediaMTX Integration Module represents a well-architected, production-ready orchestration layer that successfully implements the single source of truth pattern while maintaining clean separation of concerns. The module demonstrates strong architectural principles including event-driven design, optional component patterns, and comprehensive error handling.

### Key Architectural Achievements

1. **Centralized Business Logic**: All video operations flow through the MediaMTX controller
2. **Clean API Abstraction**: Proper mapping between external APIs and internal hardware
3. **Stateless Design**: Recording state managed by MediaMTX API, not local sessions
4. **Multi-Tier Optimization**: Intelligent snapshot capture with performance-optimized fallbacks
5. **Production Readiness**: Circuit breaker patterns, health monitoring, and graceful degradation

### Architectural Maturity

The current architecture demonstrates high maturity in design patterns and component organization. The foundation is solid for future evolution while maintaining backward compatibility and system stability.

**Document Maintenance**: This architecture documentation should be updated when significant architectural changes are implemented, following the principle that architecture documentation should reflect the current state of the system, not future plans.
