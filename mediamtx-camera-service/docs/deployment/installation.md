# Installation Guide

## System Requirements

- **Operating System**: Ubuntu 22.04+ or similar Linux distribution
- **Python**: 3.10 or higher
- **Memory**: 512MB minimum, 1GB recommended
- **Storage**: 10GB minimum for recordings
- **Hardware**: USB cameras compatible with V4L2

## Prerequisites

### System Updates
`
sudo apt update && sudo apt upgrade -y
`

### Required Packages
`
sudo apt install -y python3 python3-pip python3-venv v4l-utils ffmpeg git
`

## Installation Methods

### Method 1: Automated Installation (Recommended)

`# Download and run installation script`
`curl -sSL https://raw.githubusercontent.com/your-org/mediamtx-camera-service/main/deployment/scripts/install.sh | sudo bash`

### Method 2: Manual Installation

#### Step 1: Clone Repository
`
git clone https://github.com/your-org/mediamtx-camera-service
cd mediamtx-camera-service
`

#### Step 2: Run Installation Script
`
sudo ./deployment/scripts/install.sh
`

#### Step 3: Verify Installation
`
sudo systemctl status camera-service
sudo systemctl status mediamtx
`

## Post-Installation Setup

### Start Services
`
sudo systemctl enable camera-service mediamtx
sudo systemctl start camera-service mediamtx
`

### Check Status
`
# Service status
sudo systemctl status camera-service
sudo systemctl status mediamtx

# View logs
sudo journalctl -u camera-service -f
sudo journalctl -u mediamtx -f
`

### Test Connection
`
# Test WebSocket connection
curl -H "Connection: Upgrade" -H "Upgrade: websocket" -H "Sec-WebSocket-Key: test" -H "Sec-WebSocket-Version: 13" http://localhost:8002/ws

# Check MediaMTX API
curl http://localhost:9997/v3/paths/list
`

## Configuration

### Main Configuration File
Edit /opt/camera-service/config/camera-service.yaml:

`yaml
server:
  host: "0.0.0.0"
  port: 8002

mediamtx:
  api_port: 9997
  rtsp_port: 8554
  webrtc_port: 8889
`

### Environment Variables  
Edit /etc/systemd/system/camera-service.env:

`
CAMERA_SERVICE_HOST=0.0.0.0
CAMERA_SERVICE_PORT=8002
MEDIAMTX_API_PORT=9997
LOG_LEVEL=INFO
`

After editing configuration:
`
sudo systemctl daemon-reload
sudo systemctl restart camera-service
`

## Verification

### Camera Detection
`
# List video devices
ls -la /dev/video*

# Test camera capabilities
v4l2-ctl --list-devices
v4l2-ctl --device=/dev/video0 --list-formats-ext
`

### Service Health
`
# Check process status
ps aux | grep camera-service
ps aux | grep mediamtx

# Check listening ports
sudo netstat -tlnp | grep -E ':(8002|8554|8889|9997)'

# Test API endpoints
curl http://localhost:9997/v3/config/global/get
`

## Troubleshooting

### Common Issues

#### Services Won't Start
`
# Check service logs
sudo journalctl -u camera-service -n 50
sudo journalctl -u mediamtx -n 50

# Check file permissions
sudo ls -la /opt/camera-service/
sudo ls -la /opt/camera-service/logs/
`

#### Camera Not Detected
`
# Check camera permissions
ls -la /dev/video*
groups camera-service

# Test camera access
sudo -u camera-service v4l2-ctl --list-devices
`

#### Port Conflicts
`
# Check port usage
sudo netstat -tlnp | grep -E ':(8002|8554|8889|9997)'

# Kill conflicting processes
sudo fuser -k 8002/tcp
`

### Log Files
- **Service logs**: /opt/camera-service/logs/camera-service.log
- **System logs**: journalctl -u camera-service
- **MediaMTX logs**: journalctl -u mediamtx

## Uninstallation

`
sudo ./deployment/scripts/uninstall.sh
`

Or manually:
`
sudo systemctl stop camera-service mediamtx
sudo systemctl disable camera-service mediamtx
sudo rm /etc/systemd/system/camera-service.service
sudo rm /etc/systemd/system/mediamtx.service
sudo rm -rf /opt/camera-service
sudo userdel camera-service
`
"@
}

# Deployment files
 = @{
    "deployment/systemd/camera-service.service" = @"
[Unit]
Description=Camera Service WebSocket JSON-RPC Server
Documentation=https://github.com/your-org/mediamtx-camera-service
After=network.target mediamtx.service
Requires=mediamtx.service

[Service]
Type=simple
User=camera-service
Group=camera-service

# Working directory
WorkingDirectory=/opt/camera-service

# Python executable and script
ExecStart=/opt/camera-service/bin/python -m camera_service.main

# Environment
EnvironmentFile=/etc/systemd/system/camera-service.env
Environment=PYTHONPATH=/opt/camera-service/src
Environment=PYTHONUNBUFFERED=1

# Restart policy
Restart=always
RestartSec=5
StartLimitInterval=60
StartLimitBurst=3

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/camera-service/logs /opt/camera-service/recordings /opt/camera-service/snapshots
ProtectControlGroups=true
ProtectKernelModules=true
ProtectKernelTunables=true
RestrictRealtime=true
RestrictSUIDSGID=true
RemoveIPC=true
RestrictNamespaces=true

# Process settings
LimitNOFILE=65536
LimitNPROC=4096

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=camera-service

# Graceful shutdown
TimeoutStopSec=30
KillMode=mixed
KillSignal=SIGTERM

[Install]
WantedBy=multi-user.target
