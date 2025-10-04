# Camera Module

The camera module provides comprehensive camera device discovery, monitoring, and management capabilities for the MediaMTX Camera Service.

## Quick Start

```go
import "github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"

// Create real implementations
deviceChecker := &camera.RealDeviceChecker{}
commandExecutor := &camera.RealV4L2CommandExecutor{}
infoParser := &camera.RealDeviceInfoParser{}

// Create monitor with dependency injection
monitor, err := camera.NewHybridCameraMonitor(
    configManager,
    logger,
    deviceChecker,
    commandExecutor,
    infoParser,
)
```

## Documentation

For complete package documentation, see:
- **Package docs**: `go doc github.com/camerarecorder/mediamtx-camera-service-go/internal/camera`
- **Interface contracts**: `go doc github.com/camerarecorder/mediamtx-camera-service-go/internal/camera CameraMonitor`
- **Type definitions**: `go doc github.com/camerarecorder/mediamtx-camera-service-go/internal/camera CameraDevice`

## Architecture

- **Interface-based design** with dependency injection
- **Event-driven monitoring** with real-time device discovery
- **Hybrid camera support** (USB, IP, RTSP)
- **Performance targets**: <100ms discovery, <500ms capability probing

## Testing

The module uses integration testing with real hardware:
- `interfaces_test.go`: Interface contract validation
- `real_hardware_test.go`: Hardware integration testing
- `hybrid_monitor_test.go`: Monitor lifecycle testing

## Requirements

- Go 1.21+
- `v4l2-ctl` utility installed
- Camera device permissions (`/dev/video*`)

---

**Status**: Production Ready  
**Last Updated**: 2025-01-15
