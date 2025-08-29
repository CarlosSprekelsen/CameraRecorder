# MediaMTX Camera Service (Go) - Deployment Scripts

This directory contains deployment scripts for the MediaMTX Camera Service Go implementation, adapted from the Python version.

## Scripts Overview

### 1. `install.sh` - Installation Script
Installs the complete MediaMTX Camera Service (Go) system including:
- System dependencies (Go, V4L2, FFmpeg, etc.)
- MediaMTX server
- Camera Service (Go binary)
- Systemd services
- User/group setup
- Video device permissions
- HTTPS configuration (optional)

**Usage:**
```bash
sudo ./deployment/scripts/install.sh
```

**Environment Variables:**
- `PRODUCTION_MODE=true` - Enable production mode with HTTPS and monitoring
- `ENABLE_HTTPS=true` - Enable HTTPS configuration
- `ENABLE_MONITORING=true` - Enable monitoring services

### 2. `uninstall.sh` - Uninstallation Script
Completely removes all components of the MediaMTX Camera Service installation:
- Stops and disables systemd services
- Removes service files and symlinks
- Deletes installation directories
- Removes service users and groups
- Cleans up SSL certificates and nginx configuration
- Generates uninstall report

**Usage:**
```bash
sudo ./deployment/scripts/uninstall.sh
```

### 3. `verify_installation.sh` - Verification Script
Comprehensive verification of the installation:
- System dependencies check
- Go installation verification
- Service status validation
- File and directory existence
- User/group verification
- File permissions validation
- API accessibility testing
- Video device permissions
- Configuration file validation
- Generates verification report

**Usage:**
```bash
sudo ./deployment/scripts/verify_installation.sh
```

## Installation Process

### Prerequisites
- Ubuntu/Debian-based system
- Root access (sudo)
- Internet connection for downloading dependencies

### Step-by-Step Installation

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd mediamtx-camera-service-go
   ```

2. **Run the installation script:**
   ```bash
   sudo ./deployment/scripts/install.sh
   ```

3. **Verify the installation:**
   ```bash
   sudo ./deployment/scripts/verify_installation.sh
   ```

### Production Installation

For production deployment, use:
```bash
sudo PRODUCTION_MODE=true ./deployment/scripts/install.sh
```

This will:
- Enable HTTPS with SSL certificates
- Configure nginx as reverse proxy
- Enable monitoring services
- Apply stricter security settings

## Systemd Services

The installation creates two systemd services:

### 1. `mediamtx.service`
- **Description:** MediaMTX Media Server
- **User:** mediamtx
- **Port:** 9997 (API), 8554 (RTSP), 1935 (RTMP), 8888 (HLS), 8889 (WebRTC)
- **Dependencies:** network.target

### 2. `camera-service.service`
- **Description:** MediaMTX Camera Service (Go)
- **User:** camera-service
- **Port:** 8080 (WebSocket/HTTP)
- **Dependencies:** network.target, mediamtx.service

## Service Management

### Start Services
```bash
sudo systemctl start mediamtx
sudo systemctl start camera-service
```

### Stop Services
```bash
sudo systemctl stop camera-service
sudo systemctl stop mediamtx
```

### Enable Services (auto-start)
```bash
sudo systemctl enable mediamtx
sudo systemctl enable camera-service
```

### Check Service Status
```bash
sudo systemctl status mediamtx
sudo systemctl status camera-service
```

### View Service Logs
```bash
sudo journalctl -u mediamtx -f
sudo journalctl -u camera-service -f
```

## Directory Structure

After installation, the following structure is created:

```
/opt/camera-service/
├── mediamtx-camera-service-go    # Go binary
├── cmd/                          # Command source files
├── internal/                     # Internal source files
├── config/
│   └── default.yaml             # Configuration file
├── recordings/                   # Recording storage
├── snapshots/                    # Snapshot storage
├── ssl/                         # SSL certificates (if HTTPS enabled)
├── go.mod                       # Go module file
└── go.sum                       # Go sum file

/opt/mediamtx/
├── mediamtx                     # MediaMTX binary
└── config/
    └── mediamtx.yml            # MediaMTX configuration
```

## Configuration

### Camera Service Configuration
The default configuration is created at `/opt/camera-service/config/default.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8080

mediamtx:
  host: "localhost"
  api_port: 9997
  health_check_url: "http://localhost:9997/v3/paths/list"

camera:
  scan_interval: "30s"
  device_paths:
    - "/dev/video0"
    - "/dev/video1"

logging:
  level: "info"
  format: "json"

security:
  jwt_secret_key: "your-secret-key-change-in-production"
  rate_limit_requests: 100
  rate_limit_window: "1m"

recording:
  default_rotation_size: 100
  segment_duration: "1h"
  max_segment_size: 1000
  cleanup_interval: "24h"

snapshots:
  interval: "5s"
  cleanup_interval: "1h"

storage:
  warn_percent: 80
  block_percent: 90
  default_path: "/opt/camera-service/recordings"
  fallback_path: "/tmp/recordings"
```

### MediaMTX Configuration
The MediaMTX configuration is created at `/opt/mediamtx/config/mediamtx.yml`:

```yaml
# MediaMTX Configuration
paths:
  all:
    source: publisher
    sourceOnDemand: yes
    sourceOnDemandStartTimeout: 10s
    sourceOnDemandCloseAfter: 10s

# API Configuration
api: yes
apiAddress: :9997

# RTSP Configuration
rtsp: yes
rtspAddress: :8554

# RTMP Configuration
rtmp: yes
rtmpAddress: :1935

# HLS Configuration
hls: yes
hlsAddress: :8888

# WebRTC Configuration
webrtc: yes
webrtcAddress: :8889

# Logging
logLevel: info
logDestinations: stdout
```

## Troubleshooting

### Common Issues

1. **Service fails to start:**
   ```bash
   sudo journalctl -u camera-service -n 50
   sudo journalctl -u mediamtx -n 50
   ```

2. **Permission denied errors:**
   ```bash
   sudo ./deployment/scripts/verify_installation.sh
   ```

3. **Video device access issues:**
   ```bash
   sudo usermod -a -G video camera-service
   sudo usermod -a -G video mediamtx
   ```

4. **API not accessible:**
   ```bash
   curl http://localhost:9997/v3/paths/list
   curl http://localhost:8080/health
   ```

### Uninstallation

To completely remove the installation:
```bash
sudo ./deployment/scripts/uninstall.sh
```

This will remove all components and generate a detailed uninstall report.

## Security Considerations

1. **Change default JWT secret key** in production
2. **Configure proper SSL certificates** for HTTPS
3. **Restrict video device access** to necessary users only
4. **Use firewall rules** to restrict API access
5. **Regular security updates** for system packages

## Monitoring

The installation includes basic monitoring capabilities:
- Systemd service monitoring
- Log aggregation via journalctl
- Health check endpoints
- Resource usage monitoring (if monitoring enabled)

## Support

For issues with the deployment scripts:
1. Check the verification script output
2. Review service logs
3. Ensure all prerequisites are met
4. Verify system compatibility
