# MediaMTX Camera Service - Container Setup Guide

**Version:** 1.0  
**Date:** 2025-01-27  
**Status:** Approved  
**Target Platform:** Ubuntu 22.04 LTS with LXD  

---

## Overview

This guide provides instructions for setting up the MediaMTX Camera Service in LXD containers. Container deployment offers isolation, easy management, and simplified backup/restore procedures.

---

## Prerequisites

- Ubuntu 22.04 LTS host system
- LXD container runtime installed
- USB cameras available on host system
- Internet connection for package downloads

---

## LXD Container Setup

### Step 1: Install LXD

Install LXD container runtime:

```bash
# Install LXD
sudo apt install -y lxd

# Initialize LXD (use default settings)
sudo lxd init --auto

# Add current user to lxd group
sudo usermod -a -G lxd $USER

# Log out and back in for group changes to take effect
echo "Please log out and log back in, then continue with the next step"
```

### Step 2: Create Ubuntu 22.04 Container

Create a container for the camera service:

```bash
# Launch Ubuntu 22.04 container
lxc launch ubuntu:22.04 camera-service-container

# Wait for container to start
lxc exec camera-service-container -- wait-for-system

# Access the container
lxc exec camera-service-container -- bash
```

### Step 3: Install Services in Container

Inside the container, follow the installation steps:

```bash
# Update system
apt update && apt upgrade -y

# Install prerequisites
apt install -y curl wget git software-properties-common python3 python3-pip python3-venv python3-dev v4l-utils ffmpeg systemd systemd-sysv logrotate net-tools

# Install MediaMTX (follow MediaMTX installation steps from main guide)
# Install Camera Service (follow Camera Service installation steps from main guide)
```

### Step 4: Configure Container for Camera Access

Configure the container to access USB cameras:

```bash
# Exit container
exit

# Configure container for USB device access
lxc config device add camera-service-container video0 unix-char path=/dev/video0

# Configure container for additional devices if needed
lxc config device add camera-service-container video1 unix-char path=/dev/video1

# Start container
lxc start camera-service-container

# Access container
lxc exec camera-service-container -- bash
```

### Step 5: Verify Container Installation

Inside the container, verify the installation:

```bash
# Check services
systemctl status mediamtx camera-service

# Test camera access
ls /dev/video*
v4l2-ctl --list-devices

# Test WebSocket connection
curl -H "Connection: Upgrade" -H "Upgrade: websocket" -H "Sec-WebSocket-Key: test" -H "Sec-WebSocket-Version: 13" http://localhost:8002/ws
```

---

## Container Management

### Starting/Stopping Container

```bash
# Start container
lxc start camera-service-container

# Stop container
lxc stop camera-service-container

# Restart container
lxc restart camera-service-container
```

### Accessing Container

```bash
# Access container shell
lxc exec camera-service-container -- bash

# Execute single command
lxc exec camera-service-container -- systemctl status mediamtx
```

### Container Configuration

```bash
# View container configuration
lxc config show camera-service-container

# Add additional devices
lxc config device add camera-service-container device-name unix-char path=/dev/device

# Remove devices
lxc config device remove camera-service-container device-name
```

### Backup and Restore

```bash
# Create container backup
lxc snapshot camera-service-container backup-$(date +%Y%m%d)

# List snapshots
lxc list camera-service-container

# Restore from snapshot
lxc restore camera-service-container backup-20250127

# Export container
lxc export camera-service-container camera-service-backup.tar.gz

# Import container
lxc import camera-service-backup.tar.gz
```

---

## Troubleshooting

### Container Won't Start

```bash
# Check container status
lxc list camera-service-container

# View container logs
lxc info camera-service-container

# Start in debug mode
lxc start camera-service-container --debug
```

### Camera Access Issues

```bash
# Check device mapping
lxc config device list camera-service-container

# Verify device exists on host
ls -la /dev/video*

# Test device access from container
lxc exec camera-service-container -- ls -la /dev/video*
```

### Network Issues

```bash
# Check container network
lxc exec camera-service-container -- ip addr

# Test network connectivity
lxc exec camera-service-container -- ping -c 3 8.8.8.8

# Check port binding
lxc exec camera-service-container -- netstat -tlnp
```

---

## Security Considerations

### Container Isolation

- Containers provide process isolation
- Each container has its own network namespace
- File system is isolated from host

### Device Access

- USB devices must be explicitly mapped
- Camera permissions are inherited from host
- Consider security implications of device access

### Network Security

- Container network is isolated by default
- Port forwarding may be required for external access
- Consider firewall rules for container traffic

---

## Performance Optimization

### Resource Limits

```bash
# Set memory limit
lxc config set camera-service-container limits.memory 2GB

# Set CPU limit
lxc config set camera-service-container limits.cpu 2

# Set disk space limit
lxc config set camera-service-container limits.disk 10GB
```

### Network Optimization

```bash
# Use host network for better performance
lxc config device add camera-service-container eth0 nic nictype=bridged parent=lxdbr0

# Configure network interface
lxc exec camera-service-container -- ip link set eth0 up
```

---

## Conclusion

Container deployment provides excellent isolation and management capabilities for the MediaMTX Camera Service. The setup process is straightforward and offers significant advantages for production deployments.

For additional support or questions, please refer to the main installation guide or create an issue in the repository.

---

**Container Setup Guide**: Complete  
**Last Updated**: 2025-01-27  
**Version**: 1.0  
**Target Platform**: Ubuntu 22.04 LTS with LXD 