# MediaMTX Camera Service Installation Guide

## Overview

This guide provides comprehensive instructions for installing the MediaMTX Camera Service on Ubuntu 22.04+ systems. The installation process is automated through a bash script that handles all dependencies, configuration, and service setup.

## Prerequisites

### System Requirements
- **Operating System**: Ubuntu 22.04+ or similar Linux distribution
- **Python**: 3.10 or higher (automatically installed)
- **Memory**: 512MB minimum, 1GB recommended
- **Storage**: 10GB minimum for recordings
- **Hardware**: USB cameras compatible with V4L2

### Network Requirements
- Internet connection for package downloads
- Ports 8002 (WebSocket), 9997 (MediaMTX API), 8554 (RTSP), 8889 (WebRTC), 8888 (HLS) available

## Installation Methods

### Method 1: Automated Installation (Recommended)

#### Step 1: Download and Run Installation Script
```bash
# Download and run the installation script
curl -sSL https://raw.githubusercontent.com/your-org/mediamtx-camera-service/main/deployment/scripts/install.sh | sudo bash
```

#### Step 2: Verify Installation
```bash
# Run the verification script
sudo ./deployment/scripts/verify_installation.sh
```

### Method 2: Manual Installation

#### Step 1: Clone Repository
```bash
git clone https://github.com/your-org/mediamtx-camera-service
cd mediamtx-camera-service
```

#### Step 2: Run Installation Script
```bash
sudo ./deployment/scripts/install.sh
```

#### Step 3: Verify Installation
```bash
sudo ./deployment/scripts/verify_installation.sh
```

## Installation Details

### What the Install Script Does

The installation script performs the following steps:

1. **System Dependencies Installation**
   - Updates package list
   - Installs Python 3, pip, venv, dev tools
   - Installs v4l-utils for camera detection
   - Installs ffmpeg for media processing
   - Installs systemd, logrotate, and other utilities

2. **User and Directory Setup**
   - Creates service user `camera-service`
   - Creates installation directory `/opt/camera-service`
   - Creates subdirectories for config, logs, recordings, snapshots
   - Sets appropriate ownership and permissions

3. **Python Environment Setup**
   - Creates Python virtual environment
   - Installs Python dependencies from `requirements.txt`
   - Upgrades pip to latest version

4. **Application Files Installation**
   - Copies source code to installation directory
   - Copies configuration files
   - Sets proper ownership

5. **Systemd Service Creation**
   - Creates systemd service file
   - Creates environment file
   - Enables and starts the service

6. **Logging Configuration**
   - Creates logrotate configuration
   - Sets up log rotation

7. **Verification**
   - Checks service status
   - Verifies network ports
   - Validates configuration

### Installation Directory Structure

```
/opt/camera-service/
├── config/
│   └── camera-service.yaml          # Main configuration file
├── logs/
│   └── camera-service.log           # Application logs
├── recordings/                      # Video recordings directory
├── snapshots/                       # Image snapshots directory
├── src/                            # Application source code
├── venv/                           # Python virtual environment
└── requirements.txt                 # Python dependencies
```

### Service Configuration

#### Systemd Service File
- **Location**: `/etc/systemd/system/camera-service.service`
- **User**: `camera-service`
- **Working Directory**: `/opt/camera-service`
- **Restart Policy**: Always restart on failure
- **Security**: Runs with restricted privileges

#### Environment File
- **Location**: `/etc/systemd/system/camera-service.env`
- **Purpose**: Environment variables for the service
- **Variables**: Configuration paths, log levels, network settings

## Post-Installation Configuration

### Basic Configuration

Edit the main configuration file:
```bash
sudo nano /opt/camera-service/config/camera-service.yaml
```

Key configuration sections:
- **server**: WebSocket server settings
- **mediamtx**: MediaMTX integration settings
- **camera**: Camera detection and management
- **logging**: Logging configuration
- **recording**: Recording settings
- **snapshots**: Snapshot settings

### Environment Variables

Edit environment variables:
```bash
sudo nano /etc/systemd/system/camera-service.env
```

Available variables:
- `CAMERA_SERVICE_CONFIG_PATH`: Path to configuration file
- `CAMERA_SERVICE_LOG_LEVEL`: Logging level (INFO, DEBUG, etc.)
- `CAMERA_SERVICE_HOST`: WebSocket server host
- `CAMERA_SERVICE_PORT`: WebSocket server port
- `MEDIAMTX_API_PORT`: MediaMTX API port

## Service Management

### Basic Commands

```bash
# Check service status
sudo systemctl status camera-service

# Start the service
sudo systemctl start camera-service

# Stop the service
sudo systemctl stop camera-service

# Restart the service
sudo systemctl restart camera-service

# Enable service to start on boot
sudo systemctl enable camera-service

# Disable service from starting on boot
sudo systemctl disable camera-service
```

### Logging

```bash
# View service logs (systemd)
sudo journalctl -u camera-service -f

# View application logs
sudo tail -f /opt/camera-service/logs/camera-service.log

# View recent logs
sudo journalctl -u camera-service -n 50
```

## Verification and Testing

### Automated Verification

Run the verification script to check all components:
```bash
sudo ./deployment/scripts/verify_installation.sh
```

The verification script checks:
- Service status and enabled state
- Network port availability
- Directory structure and permissions
- Python environment and dependencies
- Configuration file validity
- Log file status
- WebSocket connection
- System resources

### Manual Testing

#### Test WebSocket Connection
```bash
# Test WebSocket endpoint
curl -H "Connection: Upgrade" -H "Upgrade: websocket" -H "Sec-WebSocket-Key: test" -H "Sec-WebSocket-Version: 13" http://localhost:8002/ws
```

#### Test Network Ports
```bash
# Check if ports are listening
sudo netstat -tlnp | grep -E ':(8002|9997|8554|8889|8888)'
```

#### Test Camera Detection
```bash
# List available camera devices
ls /dev/video*

# Test camera capability detection
v4l2-ctl --list-devices
```

## Troubleshooting

### Common Issues

#### Service Won't Start
```bash
# Check service status
sudo systemctl status camera-service

# View detailed logs
sudo journalctl -u camera-service -n 100

# Check configuration
sudo cat /opt/camera-service/config/camera-service.yaml
```

#### Permission Issues
```bash
# Fix ownership
sudo chown -R camera-service:camera-service /opt/camera-service

# Fix permissions
sudo chmod 755 /opt/camera-service
sudo chmod 750 /opt/camera-service/config
```

#### Network Port Issues
```bash
# Check if ports are in use
sudo netstat -tlnp | grep -E ':(8002|9997|8554|8889|8888)'

# Kill processes using ports (if needed)
sudo fuser -k 8002/tcp
```

#### Python Environment Issues
```bash
# Recreate virtual environment
sudo rm -rf /opt/camera-service/venv
sudo -u camera-service python3 -m venv /opt/camera-service/venv
sudo -u camera-service /opt/camera-service/venv/bin/pip install -r /opt/camera-service/requirements.txt
```

### Reinstallation

To completely reinstall the service:
```bash
# Stop and disable service
sudo systemctl stop camera-service
sudo systemctl disable camera-service

# Remove service files
sudo rm /etc/systemd/system/camera-service.service
sudo rm /etc/systemd/system/camera-service.env
sudo rm /etc/logrotate.d/camera-service

# Remove installation directory
sudo rm -rf /opt/camera-service

# Remove service user
sudo userdel camera-service

# Reload systemd
sudo systemctl daemon-reload

# Run installation script again
sudo ./deployment/scripts/install.sh
```

## Security Considerations

### Service Security
- Service runs as dedicated user `camera-service`
- Restricted file system access
- No new privileges allowed
- Private temporary directory
- Protected system directories

### Network Security
- WebSocket server bound to specific interface
- MediaMTX API accessible only locally
- Firewall rules should be configured as needed

### File Permissions
- Configuration files readable only by service user
- Log files with appropriate permissions
- Recording directories with proper access controls

## Performance Tuning

### Memory Optimization
- Monitor memory usage: `free -h`
- Adjust log levels to reduce memory usage
- Configure log rotation to prevent disk space issues

### Network Optimization
- Configure appropriate buffer sizes
- Monitor network usage: `netstat -i`
- Adjust WebSocket connection limits

### Storage Optimization
- Configure recording cleanup policies
- Monitor disk usage: `df -h`
- Set up log rotation to prevent disk space issues

## Support and Maintenance

### Regular Maintenance
- Monitor service logs for errors
- Check disk space usage
- Update system packages regularly
- Backup configuration files

### Updates
- Pull latest code from repository
- Run installation script again (idempotent)
- Restart service after updates

### Monitoring
- Set up monitoring for service status
- Monitor resource usage
- Configure alerting for failures

## Conclusion

The MediaMTX Camera Service installation script provides a complete, automated setup process that handles all dependencies, configuration, and service management. The installation is idempotent and can be safely re-run if needed.

For additional support or questions, please refer to the project documentation or create an issue in the repository.

---

**Installation Guide**: Complete  
**Last Updated**: 2025-01-27  
**Version**: 1.0.0 