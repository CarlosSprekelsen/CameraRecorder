# MediaMTX Camera Service - Complete Installation Guide

**Version:** 2.0  
**Authors:** MediaMTX Camera Service Team  
**Date:** 2025-01-27  
**Status:** Approved  
**Target Platform:** Ubuntu 22.04 LTS  

---

## Table of Contents

1. [Overview](#overview)
2. [System Requirements](#system-requirements)
3. [Prerequisites](#prerequisites)
4. [Installation Methods](#installation-methods)
5. [MediaMTX Server Installation](#mediamtx-server-installation)
6. [Camera Service Installation](#camera-service-installation)
7. [Post-Installation Configuration](#post-installation-configuration)
8. [Verification and Testing](#verification-and-testing)
9. [LXD Container Setup](CONTAINER_SETUP.md)
10. [Troubleshooting](#troubleshooting)
11. [Uninstallation](#uninstallation)

---

## Overview

This guide provides step-by-step instructions for installing the complete MediaMTX Camera Service system on Ubuntu 22.04 LTS. The system consists of two main components:

1. **MediaMTX Server**: The media streaming server that handles RTSP, WebRTC, and HLS streaming
2. **Camera Service**: The Python-based WebSocket service that manages camera discovery and provides JSON-RPC API

This guide is designed for technicians with minimal Linux experience and provides detailed, copy-paste commands for every step.

---

## System Requirements

### Minimum Requirements
- **Operating System**: Ubuntu 22.04 LTS (recommended) or Ubuntu 20.04 LTS
- **CPU**: 1 GHz dual-core processor
- **Memory**: 2 GB RAM (4 GB recommended)
- **Storage**: 20 GB available disk space
- **Network**: Internet connection for package downloads
- **Hardware**: USB cameras compatible with V4L2

### Recommended Requirements
- **Operating System**: Ubuntu 22.04 LTS
- **CPU**: 2 GHz quad-core processor
- **Memory**: 4 GB RAM
- **Storage**: 50 GB available disk space (for recordings)
- **Network**: Gigabit Ethernet connection
- **Hardware**: Multiple USB 3.0 cameras

---

## Prerequisites

### Step 1: System Update

First, update your Ubuntu system to ensure you have the latest packages:

```bash
# Update package list
sudo apt update

# Upgrade existing packages
sudo apt upgrade -y

# Install essential tools
sudo apt install -y curl wget git software-properties-common
```

### Step 2: Install Required System Packages

Install the necessary system packages for the camera service:

```bash
# Install Python and development tools
sudo apt install -y python3 python3-pip python3-venv python3-dev

# Install camera and media tools
sudo apt install -y v4l-utils ffmpeg

# Install system utilities
sudo apt install -y systemd systemd-sysv logrotate

# Install network tools
sudo apt install -y net-tools
```

### Step 3: Verify Python Installation

Check that Python 3.10+ is installed:

```bash
# Check Python version
python3 --version

# Expected output: Python 3.10.x or higher
```

---

## Installation Methods

### Method 1: Automated Installation (Recommended)

The automated installation script handles all dependencies and configuration automatically.

### Method 2: Manual Installation

For advanced users who want to understand each step or customize the installation.

---

## MediaMTX Server Installation

### Step 1: Download MediaMTX Server

Download the MediaMTX server binary for Ubuntu 22.04:

```bash
# Create installation directory
sudo mkdir -p /opt/mediamtx

# Download MediaMTX server
cd /tmp
wget https://github.com/bluenviron/mediamtx/releases/download/v1.13.1/mediamtx_v1.13.1_linux_amd64.tar.gz

# Verify download (optional)
echo "Download completed. File size should be approximately 15MB"
ls -lh mediamtx_v1.13.1_linux_amd64.tar.gz
```

### Step 2: Extract and Install MediaMTX

Extract the MediaMTX server and install it:

```bash
# Extract the archive
tar -xzf mediamtx_v1.13.1_linux_amd64.tar.gz

# Move to installation directory
sudo mv mediamtx /opt/mediamtx/

# Set executable permissions
sudo chmod +x /opt/mediamtx/mediamtx

# Create symbolic link for easy access
sudo ln -sf /opt/mediamtx/mediamtx /usr/local/bin/mediamtx
```

### Step 3: Create MediaMTX Configuration

Create the MediaMTX configuration file:

```bash
# Create configuration directory
sudo mkdir -p /opt/mediamtx/config

# Create MediaMTX configuration file
sudo tee /opt/mediamtx/config/mediamtx.yml > /dev/null << 'EOF'
# MediaMTX Configuration for Camera Service
# This file configures MediaMTX to work with the camera service

# API settings
api: yes
apiAddress: :9997

# RTSP settings
rtspAddress: :8554
rtspTransports: [tcp, udp]

# WebRTC settings
webrtcAddress: :8889

# HLS settings
hlsAddress: :8888
hlsVariant: lowLatency

# Logging
logLevel: info
logDestinations: [stdout]

# Paths configuration
paths:
  all:
    # Record format (use fmp4 for fragmented MP4)
    recordFormat: fmp4
    # Record segment duration (must be string with time unit)
    recordSegmentDuration: "3600s"
EOF
```

### Step 4: Create MediaMTX Systemd Service

Create a systemd service for MediaMTX:

```bash
# Create MediaMTX systemd service
sudo tee /etc/systemd/system/mediamtx.service > /dev/null << 'EOF'
[Unit]
Description=MediaMTX Media Server
Documentation=https://github.com/bluenviron/mediamtx
After=network.target
Wants=network.target

[Service]
Type=simple
User=mediamtx
Group=mediamtx
WorkingDirectory=/opt/mediamtx
ExecStart=/opt/mediamtx/mediamtx /opt/mediamtx/config/mediamtx.yml
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=mediamtx

[Install]
WantedBy=multi-user.target
EOF
```

### Step 5: Create MediaMTX User and Directories

Create the MediaMTX user and set up directories:

```bash
# Create MediaMTX user
sudo useradd -r -s /bin/false -d /opt/mediamtx mediamtx

# Create MediaMTX group
sudo groupadd -f mediamtx

# Add user to group
sudo usermod -a -G mediamtx mediamtx

# Set ownership
sudo chown -R mediamtx:mediamtx /opt/mediamtx

# Set permissions
sudo chmod 755 /opt/mediamtx
sudo chmod 750 /opt/mediamtx/config
```

### Step 6: Enable and Start MediaMTX Service

Enable and start the MediaMTX service:

```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable MediaMTX service
sudo systemctl enable mediamtx

# Start MediaMTX service
sudo systemctl start mediamtx

# Check service status
sudo systemctl status mediamtx
```

### Step 7: Verify MediaMTX Installation

Verify that MediaMTX is running correctly:

```bash
# Check if MediaMTX is running
sudo systemctl is-active mediamtx

# Check if ports are listening
sudo netstat -tlnp | grep -E ':(8554|8889|8888|9997)'

# Test MediaMTX API
curl http://localhost:9997/v3/config/global/get

# Check MediaMTX logs
sudo journalctl -u mediamtx -n 20
```

### Step 8: Understanding the MediaMTX Configuration

The MediaMTX configuration uses a minimal setup that avoids common configuration errors:

**Key Configuration Decisions:**

1. **No Global Recording Settings**: Recording is not enabled globally to avoid `recordPath` configuration issues
2. **API-Controlled Recording**: Recording is enabled per-stream through the MediaMTX REST API
3. **Updated RTSP Settings**: Uses `rtspTransports` instead of deprecated `protocols`
4. **Removed Problematic Settings**: Eliminated `sourceOnDemand` and global recording paths

**Recording Behavior:**
- Recording is disabled by default in the configuration
- The camera service will enable recording per camera stream via API calls
- Recording paths are set dynamically when recording starts
- This approach provides better control and avoids configuration conflicts

**Ports in Use:**
- **8554**: RTSP streaming
- **8888**: HLS streaming  
- **8889**: WebRTC streaming
- **9997**: REST API for control
- **1935**: RTMP (default, not used by camera service)
- **8890**: SRT (default, not used by camera service)
- **8189**: WebRTC ICE (default, not used by camera service)

---

## Camera Service Installation

### Method 1: Automated Installation (Recommended)

#### Step 1: Clone Repository and Run Installation Script

**For Public Repositories:**
```bash
# Clone the repository
git clone https://github.com/your-org/mediamtx-camera-service
cd mediamtx-camera-service

# Run the installation script
sudo ./deployment/scripts/install.sh
```

**For Private Repositories:**
```bash
# Clone your private repository
git clone https://github.com/YOUR_USERNAME/YOUR_REPO_NAME.git
cd YOUR_REPO_NAME

# Run the installation script
sudo ./deployment/scripts/install.sh
```

#### Step 2: Verify Installation

```bash
# Run the verification script
sudo ./deployment/scripts/verify_installation.sh
```

### Method 2: Manual Installation (Alternative)

#### Step 1: Download Required Files

Since the installation script requires the complete project files (requirements.txt, source code, etc.), you must have the full repository:

```bash
# Clone the repository (replace with your actual repository URL)
git clone https://github.com/YOUR_USERNAME/YOUR_REPO_NAME.git
cd YOUR_REPO_NAME

# Or download as ZIP and extract
wget https://github.com/YOUR_USERNAME/YOUR_REPO_NAME/archive/main.zip
unzip main.zip
cd YOUR_REPO_NAME-main
```

#### Step 2: Run Installation Script

```bash
# Run the installation script
sudo ./deployment/scripts/install.sh
```

#### Step 3: Verify Installation

```bash
# Run verification script
sudo ./deployment/scripts/verify_installation.sh
```

### Method 3: Remote Installation (Advanced)

If you need to install on a remote server without cloning the repository:

#### Step 1: Copy Required Files to Server

From your local machine, copy the entire project to the server:

```bash
# Copy the project directory to the server
scp -r /path/to/your/mediamtx-camera-service dts@your-server:/tmp/

# Or create a tarball and copy
tar -czf camera-service.tar.gz mediamtx-camera-service/
scp camera-service.tar.gz dts@your-server:/tmp/
```

#### Step 2: Install on Server

On the remote server:

```bash
# Extract if using tarball
cd /tmp
tar -xzf camera-service.tar.gz
cd mediamtx-camera-service

# Run installation
sudo ./deployment/scripts/install.sh

### Production Installation

For production deployment with enhanced security, monitoring, and backup features:

```bash
# Run production installation with enhanced features
sudo PRODUCTION_MODE=true ./deployment/scripts/install.sh
```

**Production features include:**
- HTTPS/SSL configuration with Nginx reverse proxy
- UFW firewall configuration and security hardening
- Automated backup and recovery procedures
- Production monitoring and alerting
- Enhanced logging and health monitoring

**Validation:**
```bash
# Run validation to verify all production features
sudo ./deployment/scripts/validate_production.sh

# Run complete production setup (all phases)
sudo ./deployment/scripts/setup_production.sh
```
```

---

## Post-Installation Configuration

### Step 1: Configure Camera Service

Edit the camera service configuration:

```bash
# Edit configuration file
sudo nano /opt/camera-service/config/camera-service.yaml
```

Example configuration:

```yaml
# MediaMTX Camera Service Configuration

server:
  host: "0.0.0.0"
  port: 8002
  websocket_path: "/ws"
  max_connections: 100

mediamtx:
  host: "127.0.0.1"
  api_port: 9997
  rtsp_port: 8554
  webrtc_port: 8889
  hls_port: 8888
  config_path: "/opt/mediamtx/config/mediamtx.yml"
  recordings_path: "/opt/camera-service/recordings"
  snapshots_path: "/opt/camera-service/snapshots"

camera:
  poll_interval: 0.1
  detection_timeout: 2.0
  device_range: [0, 9]
  enable_capability_detection: true
  auto_start_streams: true

logging:
  level: "INFO"
  format: "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
  file_enabled: true
  file_path: "/opt/camera-service/logs/camera-service.log"
  max_file_size: "10MB"
  backup_count: 5

recording:
  auto_record: false
  format: "mp4"
  quality: "high"
  max_duration: 3600
  cleanup_after_days: 30

snapshots:
  format: "jpg"
  quality: 90
  cleanup_after_days: 7
```

### Step 2: Restart Services

Restart both services to apply configuration changes:

```bash
# Restart MediaMTX service
sudo systemctl restart mediamtx

# Restart camera service
sudo systemctl restart camera-service

# Check both services
sudo systemctl status mediamtx camera-service
```

### Step 3: Configure Firewall (Optional)

If you have UFW firewall enabled, configure it to allow the required ports:

```bash
# Allow WebSocket port
sudo ufw allow 8002/tcp

# Allow RTSP port
sudo ufw allow 8554/tcp

# Allow WebRTC port
sudo ufw allow 8889/tcp

# Allow HLS port
sudo ufw allow 8888/tcp

# Allow MediaMTX API port
sudo ufw allow 9997/tcp

# Reload firewall
sudo ufw reload
```

---

## Verification and Testing

### Step 1: Automated Verification

Run the comprehensive verification script:

```bash
# Run verification script
sudo ./deployment/scripts/verify_installation.sh
```

### Step 2: Manual Testing

#### Test WebSocket Connection

```bash
# Test WebSocket endpoint
curl -H "Connection: Upgrade" -H "Upgrade: websocket" -H "Sec-WebSocket-Key: test" -H "Sec-WebSocket-Version: 13" http://localhost:8002/ws
```

#### Test Network Ports

```bash
# Check if all ports are listening
sudo netstat -tlnp | grep -E ':(8002|8554|8889|8888|9997)'
```

#### Test Camera Detection

```bash
# List available camera devices
ls /dev/video*

# Test camera capability detection
v4l2-ctl --list-devices

# Test specific camera (if available)
v4l2-ctl --device=/dev/video0 --list-formats-ext
```

#### Test MediaMTX API

```bash
# Test MediaMTX API endpoints
curl http://localhost:9997/v3/config/global/get
curl http://localhost:9997/v3/paths/list
```

### Step 3: Service Health Check

```bash
# Check service status
sudo systemctl status mediamtx camera-service

# Check service logs
sudo journalctl -u mediamtx -n 20
sudo journalctl -u camera-service -n 20

# Check application logs
sudo tail -f /opt/camera-service/logs/camera-service.log
```

---

## LXD Container Setup

Container deployment provides excellent isolation and management capabilities for the MediaMTX Camera Service. For detailed container setup instructions, please refer to the separate [Container Setup Guide](CONTAINER_SETUP.md).

The container setup includes:
- LXD container runtime installation
- Ubuntu 22.04 container creation
- Service installation within container
- USB camera device mapping
- Container management and troubleshooting
- Security considerations and performance optimization

---

## Troubleshooting

### Common Issues and Solutions

#### Issue 1: Services Won't Start

**Symptoms:**
- `systemctl status camera-service` shows failed
- `systemctl status mediamtx` shows failed

**Solution:**
```bash
# Check service logs
sudo journalctl -u camera-service -n 50
sudo journalctl -u mediamtx -n 50

# Check file permissions
sudo ls -la /opt/camera-service/
sudo ls -la /opt/mediamtx/

# Fix ownership if needed
sudo chown -R camera-service:camera-service /opt/camera-service
sudo chown -R mediamtx:mediamtx /opt/mediamtx

# Restart services
sudo systemctl restart camera-service mediamtx
```

#### Issue 2: Camera Not Detected

**Symptoms:**
- No cameras listed in `/dev/video*`
- Camera detection fails

**Solution:**
```bash
# Check camera permissions
ls -la /dev/video*
groups camera-service

# Test camera access
sudo -u camera-service v4l2-ctl --list-devices

# Check if camera is recognized by system
dmesg | grep -i camera
dmesg | grep -i video

# Install additional camera drivers if needed
sudo apt install -y v4l-utils
```

#### Issue 3: Port Conflicts

**Symptoms:**
- Services fail to start due to port already in use
- Connection refused errors

**Solution:**
```bash
# Check port usage
sudo netstat -tlnp | grep -E ':(8002|8554|8889|8888|9997)'

# Kill conflicting processes
sudo fuser -k 8002/tcp
sudo fuser -k 8554/tcp
sudo fuser -k 8889/tcp
sudo fuser -k 8888/tcp
sudo fuser -k 9997/tcp

# Restart services
sudo systemctl restart camera-service mediamtx
```

#### Issue 4: Python Environment Issues

**Symptoms:**
- Import errors
- Module not found errors
- "requirements.txt not found in current directory" error

**Solution:**
```bash
# Check if requirements.txt exists
ls -la requirements.txt

# If requirements.txt is missing, the installation script was run without the full project
# You need to clone the repository first:
git clone https://github.com/YOUR_USERNAME/YOUR_REPO_NAME.git
cd YOUR_REPO_NAME
sudo ./deployment/scripts/install.sh

# Or recreate virtual environment with proper requirements
sudo rm -rf /opt/camera-service/venv
sudo -u camera-service python3 -m venv /opt/camera-service/venv
sudo -u camera-service /opt/camera-service/venv/bin/pip install -r /opt/camera-service/requirements.txt

# Restart service
sudo systemctl restart camera-service
```

#### Issue 5: MediaMTX Configuration Errors

**Symptoms:**
- MediaMTX service fails to start with configuration errors
- Errors like "invalid record format", "unknown field", or "recordPath must contain %path"

**Solution:**
```bash
# Check MediaMTX configuration syntax
sudo -u mediamtx /opt/mediamtx/mediamtx /opt/mediamtx/config/mediamtx.yml

# Common fixes:
# 1. Use 'fmp4' instead of 'mp4' for recordFormat
# 2. Use string values for durations: "3600s" instead of 3600
# 3. Use 'rtspTransports' instead of deprecated 'protocols'
# 4. Remove global recording settings to avoid recordPath issues
# 5. Remove 'sourceOnDemand' which can cause conflicts

# Edit configuration if needed
sudo nano /opt/mediamtx/config/mediamtx.yml

# Restart MediaMTX after fixes
sudo systemctl restart mediamtx
```

**Recommended Minimal Configuration:**
```yaml
# MediaMTX Configuration for Camera Service
api: yes
apiAddress: :9997
rtspAddress: :8554
rtspTransports: [tcp, udp]
webrtcAddress: :8889
hlsAddress: :8888
hlsVariant: lowLatency
logLevel: info
logDestinations: [stdout]

paths:
  all:
    recordFormat: fmp4
    recordSegmentDuration: "3600s"
```

#### Issue 6: MediaMTX API Connection Issues

**Symptoms:**
- Camera service cannot connect to MediaMTX API
- MediaMTX API returns errors

**Solution:**
```bash
# Check MediaMTX service status
sudo systemctl status mediamtx

# Check MediaMTX logs
sudo journalctl -u mediamtx -n 50

# Test MediaMTX API directly
curl http://localhost:9997/v3/config/global/get

# Check MediaMTX configuration
sudo cat /opt/mediamtx/config/mediamtx.yml

# Restart MediaMTX
sudo systemctl restart mediamtx
```

#### Issue 7: MediaMTX Service Security Settings Conflict

**Symptoms:**
- MediaMTX service fails to start with NAMESPACE error
- Service shows "activating (auto-restart)" status
- Journal logs show "Failed to set up namespace"

**Root Cause:**
- Overly restrictive security settings in systemd service
- `NoNewPrivileges=true`, `PrivateTmp=true`, `ProtectSystem=strict`, `ProtectHome=true` cause namespace conflicts

**Solution:**
```bash
# Edit the MediaMTX service file
sudo nano /etc/systemd/system/mediamtx.service

# Remove these problematic lines:
# NoNewPrivileges=true
# PrivateTmp=true
# ProtectSystem=strict
# ProtectHome=true

# Reload and restart
sudo systemctl daemon-reload
sudo systemctl restart mediamtx
```

**Prevention:**
- Use minimal security settings for MediaMTX service
- Test service startup after configuration changes

#### Issue 8: WebSocket Server Binding Issues

**Symptoms:**
- Camera service starts but WebSocket port 8002 not listening
- Logs show "Starting WebSocket JSON-RPC server" but no confirmation
- No error messages in logs

**Diagnosis:**
```bash
# Check if port is blocked
sudo lsof -i :8002

# Test port availability
sudo -u camera-service /opt/camera-service/venv/bin/python3 -c "
import socket
s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
try:
    s.bind(('0.0.0.0', 8002))
    print('Port 8002 is available')
    s.close()
except Exception as e:
    print(f'Port 8002 blocked: {e}')
"

# Check firewall
sudo ufw status
```

**Potential Solutions:**
- Restart camera service: `sudo systemctl restart camera-service`
- Check for port conflicts with other services
- Verify firewall settings
- Check system resource limits

### Log Files and Debugging

#### Service Logs
```bash
# View camera service logs
sudo journalctl -u camera-service -f

# View MediaMTX logs
sudo journalctl -u mediamtx -f

# View application logs
sudo tail -f /opt/camera-service/logs/camera-service.log
```

#### System Logs
```bash
# View system messages
sudo dmesg | tail -50

# View USB device messages
sudo dmesg | grep -i usb

# View camera device messages
sudo dmesg | grep -i video
```

### Performance Monitoring

#### Resource Usage
```bash
# Check CPU and memory usage
htop

# Check disk usage
df -h

# Check network usage
netstat -i
```

#### Service Monitoring
```bash
# Monitor service status
watch -n 5 'systemctl status mediamtx camera-service'

# Monitor port usage
watch -n 5 'netstat -tlnp | grep -E ":(8002|8554|8889|8888|9997)"'
```

---

## Uninstallation

### Complete Uninstallation

To completely remove the MediaMTX Camera Service system:

```bash
# Stop and disable services
sudo systemctl stop camera-service mediamtx
sudo systemctl disable camera-service mediamtx

# Remove service files
sudo rm /etc/systemd/system/camera-service.service
sudo rm /etc/systemd/system/mediamtx.service
sudo rm /etc/systemd/system/camera-service.env
sudo rm /etc/logrotate.d/camera-service

# Remove installation directories
sudo rm -rf /opt/camera-service
sudo rm -rf /opt/mediamtx

# Remove service users
sudo userdel camera-service
sudo userdel mediamtx

# Remove symbolic links
sudo rm -f /usr/local/bin/mediamtx

# Reload systemd
sudo systemctl daemon-reload

# Clean up temporary files
sudo rm -f /tmp/mediamtx_v1.13.1_linux_amd64.tar.gz
```

### Partial Uninstallation

To remove only specific components:

#### Remove Camera Service Only
```bash
# Stop and disable camera service
sudo systemctl stop camera-service
sudo systemctl disable camera-service

# Remove camera service files
sudo rm /etc/systemd/system/camera-service.service
sudo rm /etc/systemd/system/camera-service.env
sudo rm -rf /opt/camera-service
sudo userdel camera-service
```

#### Remove MediaMTX Only
```bash
# Stop and disable MediaMTX service
sudo systemctl stop mediamtx
sudo systemctl disable mediamtx

# Remove MediaMTX files
sudo rm /etc/systemd/system/mediamtx.service
sudo rm -rf /opt/mediamtx
sudo userdel mediamtx
sudo rm -f /usr/local/bin/mediamtx
```

---

## Client Application Recommendations

### Web Client

The web client should be bundled with the server for ease of deployment. Recommended approach:

1. **Bundle with Server**: Include the web client files in the camera service installation
2. **Serve via Camera Service**: Configure the camera service to serve the web client on a dedicated port
3. **Auto-discovery**: Web client should auto-discover the camera service on the local network

### Android Client

The Android APK should be distributed separately but include:

1. **Auto-discovery**: Ability to discover camera service on local network
2. **Manual Configuration**: Option to manually enter server address and port
3. **Offline Mode**: Basic functionality when server is unavailable

### Installation Script Updates

The installation script should be updated to:

1. **Include Web Client**: Bundle and install web client files
2. **Configure Web Server**: Set up web server for client access
3. **Generate Certificates**: Create self-signed certificates for HTTPS
4. **Firewall Configuration**: Automatically configure firewall rules

---

## Support and Maintenance

### Regular Maintenance Tasks

#### Daily Tasks
- Monitor service logs for errors
- Check disk space usage
- Verify camera connectivity

#### Weekly Tasks
- Review and rotate log files
- Check for system updates
- Backup configuration files

#### Monthly Tasks
- Update system packages
- Review and clean old recordings
- Test backup and recovery procedures

### Monitoring and Alerting

#### Recommended Monitoring
- Service status monitoring
- Resource usage monitoring
- Network connectivity monitoring
- Camera availability monitoring

#### Alerting Setup
```bash
# Example: Simple monitoring script
sudo tee /opt/camera-service/scripts/monitor.sh > /dev/null << 'EOF'
#!/bin/bash
# Simple monitoring script for camera service

# Check services
if ! systemctl is-active --quiet camera-service; then
    echo "ALERT: Camera service is down!"
    systemctl restart camera-service
fi

if ! systemctl is-active --quiet mediamtx; then
    echo "ALERT: MediaMTX service is down!"
    systemctl restart mediamtx
fi

# Check disk space
DISK_USAGE=$(df /opt/camera-service/recordings | tail -1 | awk '{print $5}' | sed 's/%//')
if [ "$DISK_USAGE" -gt 90 ]; then
    echo "ALERT: Disk space is running low: ${DISK_USAGE}%"
fi
EOF

# Make executable
sudo chmod +x /opt/camera-service/scripts/monitor.sh

# Add to crontab for regular monitoring
echo "*/5 * * * * /opt/camera-service/scripts/monitor.sh" | sudo crontab -
```

---

## Conclusion

This installation guide provides comprehensive, step-by-step instructions for deploying the complete MediaMTX Camera Service system on Ubuntu 22.04 LTS. The guide is designed to be followed by technicians with minimal Linux experience and includes detailed troubleshooting and maintenance procedures.

For additional support or questions, please refer to the project documentation or create an issue in the repository.

---

**Installation Guide**: Complete  
**Last Updated**: 2025-01-27  
**Version**: 2.0  
**Target Platform**: Ubuntu 22.04 LTS 