# MediaMTX Camera Service

A lightweight WebSocket JSON-RPC 2.0 service that provides real-time USB camera monitoring and control using MediaMTX as the media server backend.

## Features

- **Real-time Camera Discovery**: Automatic USB camera connect/disconnect detection
- **WebSocket JSON-RPC 2.0 API**: Standard protocol for camera control
- **MediaMTX Integration**: Proven media server for streaming and recording
- **Multi-protocol Streaming**: RTSP, WebRTC, HLS support via MediaMTX
- **Recording & Snapshots**: Video recording and image capture capabilities
- **Production Ready**: Systemd services, logging, monitoring

## Quick Start

### Prerequisites
- Ubuntu 22.04+ or similar Linux distribution
- Python 3.10+
- USB cameras compatible with V4L2

### Installation
```bash
# Clone the repository
git clone https://github.com/your-org/mediamtx-camera-service
cd mediamtx-camera-service

# Basic installation
sudo ./deployment/scripts/install.sh

# Production installation
sudo PRODUCTION_MODE=true ./deployment/scripts/install.sh

# Automated deployment cycle (uninstall + install)
sudo ./deployment/scripts/deploy.sh --force-uninstall
```

### Validation

The installation process includes comprehensive validation to catch configuration and component mismatches early:

- **Configuration Schema Validation**: Ensures all required parameters are present and correctly typed
- **Component Instantiation Validation**: Tests that configuration can instantiate all components
- **API Interface Validation**: Verifies parameter compatibility between dataclasses and constructors
- **Python Compatibility Validation**: Checks for python3 availability and dependency installation

Run validation manually:
```bash
python3 scripts/validate_deployment.py
```

For detailed information about installation fixes and validation, see [Installation Fixes](docs/deployment/installation_fixes.md).

### Usage
```bash
# Start services
sudo systemctl start camera-service
sudo systemctl start mediamtx

# Check status
sudo systemctl status camera-service

# View logs
sudo journalctl -u camera-service -f

# Uninstall (if needed)
sudo ./deployment/scripts/uninstall.sh --force
```

## Architecture

`
┌─────────────────┐    WebSocket     ┌─────────────────┐
│   Web Clients   │ ◄──JSON-RPC────► │ Camera Service  │
└─────────────────┘                  └─────────┬───────┘
                                               │ REST API
┌─────────────────┐    USB Events    ┌─────────▼───────┐
│  USB Cameras    │ ◄──────────────► │    MediaMTX     │
└─────────────────┘                  └─────────────────┘
`

## API Examples

### Connect to WebSocket
```javascript
const ws = new WebSocket('ws://localhost:8002/ws');
```

### Get Camera List
```json
{
  "jsonrpc": "2.0",
  "method": "get_camera_list", 
  "id": 1
}
```

### Camera Status Notification
```json
{
  "jsonrpc": "2.0",
  "method": "camera_status_update",
  "params": {
    "device": "/dev/video0",
    "status": "CONNECTED",
    "streams": {
      "rtsp": "rtsp://localhost:8554/camera0",
      "webrtc": "http://localhost:8889/camera0/webrtc"
    }
  }
}
```

## Documentation

- [API Reference](docs/api/json-rpc-methods.md)
- [Installation Guide](docs/deployment/INSTALLATION_GUIDE.md)
- [Architecture Overview](docs/architecture/overview.md)
- [Development Setup](docs/development/setup.md)

## License

MIT License - see [LICENSE](LICENSE)
