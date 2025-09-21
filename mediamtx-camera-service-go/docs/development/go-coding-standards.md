# Go Coding Standards

**Version:** 1.0  
**Authors:** Project Team  
**Date:** 2025-01-15  
**Status:** Approved  
**Related Epic/Story:** Go Implementation Standards  

**Purpose:**  
Define Go-specific coding standards, package structure, and concurrency patterns for the MediaMTX Camera Service Go implementation. These standards ensure consistency, maintainability, and performance optimization.

---

## 1. Code Organization and Package Structure

### Directory Layout
```
mediamtx-camera-service-go/
├── cmd/server/main.go              # Application entry point
├── internal/                       # Private application code
│   ├── websocket/                  # WebSocket JSON-RPC server
│   ├── camera/                     # Camera discovery and management
│   ├── mediamtx/                   # MediaMTX integration
│   ├── config/                     # Configuration management
│   └── auth/                       # Authentication and authorization
├── pkg/                           # Public packages (reusable)
│   ├── jsonrpc/                   # JSON-RPC protocol handling
│   └── types/                     # Shared data types
├── docs/                          # Documentation
└── tests/                         # Integration tests
```

### Package Naming Conventions

- Use **snake_case** for package names (e.g., `camera_discovery`, `json_rpc`)
- Keep package names short and descriptive
- Avoid generic names like `utils`, `helpers`, or `common`
- Use plural names for packages containing multiple related types

### File Naming Conventions

- Use **snake_case** for Go files (e.g., `camera_monitor.go`, `websocket_server.go`)
- Group related functionality in the same file
- Keep files under 500 lines when possible
- Use descriptive names that indicate the file's purpose

---

## 2. Code Style and Formatting

### Go Formatting

- Use `gofmt` for automatic formatting
- Follow standard Go formatting conventions
- Use `golangci-lint` for comprehensive linting
- Configure IDE to format on save

### Naming Conventions

- **Variables:** Use camelCase (e.g., `cameraList`, `maxConnections`)
- **Constants:** Use PascalCase for exported, camelCase for unexported
- **Functions:** Use camelCase (e.g., `getCameraStatus`, `startRecording`)
- **Types:** Use PascalCase (e.g., `CameraStatus`, `RecordingSession`)
- **Interfaces:** Use PascalCase with descriptive names (e.g., `CameraMonitor`, `StreamManager`)

### Comments and Documentation

- Follow Go documentation conventions
- Use `//` for single-line comments
- Use `/* */` for multi-line comments
- Document all exported functions, types, and packages
- Include examples in documentation where appropriate

### Example Function Documentation
```go
// CameraStatus represents the current status of a camera device.
type CameraStatus struct {
    Device     string            `json:"device"`
    Status     string            `json:"status"`
    Name       string            `json:"name"`
    Resolution string            `json:"resolution"`
    FPS        int               `json:"fps"`
    Streams    map[string]string `json:"streams"`
}

// GetCameraStatus retrieves the current status of a specific camera.
// Returns an error if the camera is not found or unavailable.
func (m *CameraMonitor) GetCameraStatus(device string) (*CameraStatus, error) {
    // Implementation
}
```

---

## 3. Error Handling

### Error Handling Patterns
- Always check errors and handle them appropriately
- Use `fmt.Errorf` with `%w` verb for error wrapping
- Create custom error types for specific error conditions
- Use `errors.Is` and `errors.As` for error type checking

### Custom Error Types
```go
// CameraError represents camera-specific errors
type CameraError struct {
    Device string
    Op     string
    Err    error
}

func (e *CameraError) Error() string {
    return fmt.Sprintf("camera %s: %s: %v", e.Device, e.Op, e.Err)
}

func (e *CameraError) Unwrap() error {
    return e.Err
}

// Error constants
var (
    ErrCameraNotFound = errors.New("camera not found")
    ErrCameraBusy     = errors.New("camera is busy")
    ErrRecordingInProgress = errors.New("recording already in progress")
)
```

### Error Handling Examples
```go
func (m *CameraMonitor) StartRecording(device string) error {
    camera, err := m.getCamera(device)
    if err != nil {
        return fmt.Errorf("failed to get camera %s: %w", device, err)
    }
    
    if camera.IsRecording {
        return &CameraError{
            Device: device,
            Op:     "start_recording",
            Err:    ErrRecordingInProgress,
        }
    }
    
    // Continue with recording logic
    return nil
}
```

---

## 4. Concurrency Patterns

### Goroutines and Channels
- Use goroutines for concurrent operations
- Use channels for communication between goroutines
- Prefer `select` statements for non-blocking operations
- Use `context.Context` for cancellation and timeouts

### Worker Pool Pattern
```go
type WorkerPool struct {
    workers    int
    jobQueue   chan Job
    resultChan chan Result
    wg         sync.WaitGroup
    ctx        context.Context
    cancel     context.CancelFunc
}

func NewWorkerPool(workers int) *WorkerPool {
    ctx, cancel := context.WithCancel(context.Background())
    return &WorkerPool{
        workers:    workers,
        jobQueue:   make(chan Job, 100),
        resultChan: make(chan Result, 100),
        ctx:        ctx,
        cancel:     cancel,
    }
}

func (wp *WorkerPool) Start() {
    for i := 0; i < wp.workers; i++ {
        wp.wg.Add(1)
        go wp.worker()
    }
}

func (wp *WorkerPool) worker() {
    defer wp.wg.Done()
    
    for {
        select {
        case job := <-wp.jobQueue:
            result := wp.processJob(job)
            wp.resultChan <- result
        case <-wp.ctx.Done():
            return
        }
    }
}
```

### Context Usage
```go
func (m *CameraMonitor) MonitorCameras(ctx context.Context) error {
    ticker := time.NewTicker(testutils.UniversalTimeoutVeryLong)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            if err := m.scanCameras(); err != nil {
                log.Printf("camera scan failed: %v", err)
            }
        }
    }
}
```

---

## 5. Configuration Management

### Configuration Structure
```go
type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Camera   CameraConfig   `mapstructure:"camera"`
    MediaMTX MediaMTXConfig `mapstructure:"mediamtx"`
    Security SecurityConfig `mapstructure:"security"`
}

type ServerConfig struct {
    Port         int    `mapstructure:"port"`
    WebSocketPort int   `mapstructure:"websocket_port"`
    LogLevel     string `mapstructure:"log_level"`
}

type CameraConfig struct {
    DiscoveryInterval time.Duration `mapstructure:"discovery_interval"`
    MaxCameras        int           `mapstructure:"max_cameras"`
    PollingEnabled    bool          `mapstructure:"polling_enabled"`
}
```

### Configuration Loading
```go
func LoadConfig(configPath string) (*Config, error) {
    viper.SetConfigFile(configPath)
    viper.SetConfigType("yaml")
    
    // Set defaults
    viper.SetDefault("server.port", 8003)
    viper.SetDefault("server.websocket_port", 8002)
    viper.SetDefault("camera.discovery_interval", "5s")
    viper.SetDefault("camera.max_cameras", 16)
    
    // Read environment variables
    viper.AutomaticEnv()
    viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    
    if err := viper.ReadInConfig(); err != nil {
        return nil, fmt.Errorf("failed to read config: %w", err)
    }
    
    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }
    
    return &config, nil
}
```

---

## 6. Logging Standards

### Structured Logging
```go
import "github.com/sirupsen/logrus"

var logger = logrus.New()

func init() {
    logger.SetFormatter(&logrus.JSONFormatter{
        TimestampFormat: time.RFC3339,
    })
}

func (m *CameraMonitor) StartRecording(device string) error {
    logger.WithFields(logrus.Fields{
        "device": device,
        "action": "start_recording",
    }).Info("Starting camera recording")
    
    // Implementation
    
    logger.WithFields(logrus.Fields{
        "device": device,
        "action": "start_recording",
        "status": "success",
    }).Info("Camera recording started")
    
    return nil
}
```

### Log Levels
- **ERROR:** System errors that require immediate attention
- **WARN:** Unexpected conditions that don't prevent operation
- **INFO:** General operational information
- **DEBUG:** Detailed information for troubleshooting

---

## 7. Testing Standards

### Unit Testing
```go
func TestCameraMonitor_GetCameraStatus(t *testing.T) {
    tests := []struct {
        name     string
        device   string
        wantErr  bool
        expected *CameraStatus
    }{
        {
            name:    "valid camera",
            device:  "/dev/video0",
            wantErr: false,
            expected: &CameraStatus{
                Device: "/dev/video0",
                Status: "CONNECTED",
            },
        },
        {
            name:    "invalid camera",
            device:  "/dev/video999",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            monitor := NewCameraMonitor()
            result, err := monitor.GetCameraStatus(tt.device)
            
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Integration Testing
```go
func TestWebSocketServer_Integration(t *testing.T) {
    // Setup test server
    server := NewWebSocketServer()
    go server.Start()
    defer server.Stop()
    
    // Connect client
    conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8002/ws", nil)
    require.NoError(t, err)
    defer conn.Close()
    
    // Send test message
    message := JSONRPCMessage{
        JSONRPC: "2.0",
        Method:  "ping",
        ID:      1,
    }
    
    err = conn.WriteJSON(message)
    require.NoError(t, err)
    
    // Read response
    var response JSONRPCResponse
    err = conn.ReadJSON(&response)
    require.NoError(t, err)
    
    assert.Equal(t, "pong", response.Result)
}
```

---

## 8. Performance Optimization

### Memory Management
- Use object pools for frequently allocated objects
- Minimize allocations in hot paths
- Use `sync.Pool` for temporary objects
- Profile memory usage with `pprof`

### Goroutine Management
- Limit the number of concurrent goroutines
- Use worker pools for CPU-intensive tasks
- Implement proper cleanup and shutdown
- Monitor goroutine leaks

### Example Object Pool
```go
type CameraStatusPool struct {
    pool sync.Pool
}

func NewCameraStatusPool() *CameraStatusPool {
    return &CameraStatusPool{
        pool: sync.Pool{
            New: func() interface{} {
                return &CameraStatus{}
            },
        },
    }
}

func (p *CameraStatusPool) Get() *CameraStatus {
    return p.pool.Get().(*CameraStatus)
}

func (p *CameraStatusPool) Put(status *CameraStatus) {
    // Reset fields
    status.Device = ""
    status.Status = ""
    status.Name = ""
    status.Resolution = ""
    status.FPS = 0
    status.Streams = nil
    
    p.pool.Put(status)
}
```

---

## 9. Security Best Practices

### Input Validation
```go
func validateDevicePath(device string) error {
    if device == "" {
        return errors.New("device path cannot be empty")
    }
    
    if !strings.HasPrefix(device, "/dev/video") {
        return errors.New("invalid device path format")
    }
    
    // Additional validation as needed
    return nil
}
```

### Secure Configuration
```go
type SecurityConfig struct {
    JWTSecret     string        `mapstructure:"jwt_secret"`
    TokenExpiry   time.Duration `mapstructure:"token_expiry"`
    MaxLoginAttempts int        `mapstructure:"max_login_attempts"`
}

func (c *SecurityConfig) Validate() error {
    if len(c.JWTSecret) < 32 {
        return errors.New("JWT secret must be at least 32 characters")
    }
    
    if c.TokenExpiry < time.Minute {
        return errors.New("token expiry must be at least 1 minute")
    }
    
    return nil
}
```

---

## 10. Code Review Checklist

Before submitting code for review, ensure:

- [ ] Code follows Go formatting standards
- [ ] All exported functions and types are documented
- [ ] Error handling is comprehensive and appropriate
- [ ] Concurrency patterns are used correctly
- [ ] Tests cover new functionality
- [ ] Performance implications are considered
- [ ] Security best practices are followed
- [ ] Configuration is properly validated
- [ ] Logging is structured and appropriate
- [ ] No TODO comments remain (unless tracked in roadmap)

---

## 11. Performance Targets

### Response Time Targets
- **Status Methods:** <50ms for 95% of requests
- **Control Methods:** <100ms for 95% of requests
- **WebSocket Notifications:** <20ms delivery latency

### Resource Usage Targets
- **Memory:** <60MB base footprint, <200MB with 10 cameras
- **CPU:** <50% sustained usage under normal load
- **Goroutines:** <1000 concurrent goroutines maximum

### Concurrency Targets
- **WebSocket Connections:** 1000+ simultaneous connections
- **Camera Operations:** 16 concurrent camera operations
- **API Requests:** 1000+ requests/second throughput

---

**References:**
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Concurrency Patterns](https://golang.org/doc/effective_go.html#concurrency)
