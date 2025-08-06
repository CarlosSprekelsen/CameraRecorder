# CLI Tool Guide - MediaMTX Camera Service

## Overview

The MediaMTX Camera Service CLI tool provides a comprehensive command-line interface for controlling cameras through the MediaMTX Camera Service. It supports JWT and API key authentication with all camera operations.

## Features

- **Dual Authentication Support**: JWT tokens and API keys
- **Comprehensive Camera Operations**: List, status, snapshot, recording, monitoring
- **Multiple Output Formats**: Table, JSON, CSV output
- **Real-time Monitoring**: Live camera and recording status updates
- **Verbose Logging**: Detailed operation feedback
- **Error Handling**: Comprehensive error reporting and recovery

## Installation

### Prerequisites

```bash
# Install Python 3.8+
python3 --version

# Install required packages
pip install websockets asyncio
```

### Quick Start

```bash
# Clone the repository
git clone <repository-url>
cd mediamtx-camera-service

# Make CLI executable
chmod +x examples/cli/camera_cli.py

# Test connection
python examples/cli/camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token ping
```

## Authentication

### JWT Authentication

```bash
# Use JWT token authentication
python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token list
```

### API Key Authentication

```bash
# Use API key authentication
python camera_cli.py --host localhost --port 8080 --auth-type api_key --key your_api_key list
```

### SSL/TLS Support

```bash
# Use SSL/TLS connection
python camera_cli.py --host secure.example.com --port 443 --ssl --auth-type jwt --token your_jwt_token list
```

## Commands

### List Cameras

List all available cameras:

```bash
# Basic list
python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token list

# JSON output
python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token list --format json

# CSV output
python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token list --format csv

# Verbose output
python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token list --verbose
```

**Output Example:**
```
📹 Found 2 camera(s):

1. USB Camera
   Device: /dev/video0
   Status: available
   Capabilities: snapshot, recording, streaming
   Stream: rtsp://localhost:8554/camera0

2. Built-in Camera
   Device: /dev/video1
   Status: busy
   Capabilities: snapshot, recording
```

### Camera Status

Get detailed status of a specific camera:

```bash
# Get camera status
python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token status /dev/video0

# JSON output
python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token status /dev/video0 --format json
```

**Output Example:**
```
📹 Camera Status: USB Camera
   Device: /dev/video0
   Status: available
   Capabilities: snapshot, recording, streaming
   Stream: rtsp://localhost:8554/camera0
```

### Take Snapshot

Take a snapshot from a camera:

```bash
# Basic snapshot
python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token snapshot /dev/video0

# Custom filename
python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token snapshot /dev/video0 --filename my_snapshot.jpg

# Verbose output
python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token snapshot /dev/video0 --verbose
```

**Output Example:**
```
✅ Snapshot saved: /opt/camera-service/snapshots/snapshot_20250806_143022.jpg
   Size: 245760 bytes
   Duration: 0.15 seconds
```

### Start Recording

Start recording from a camera:

```bash
# Basic recording
python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token record /dev/video0

# With duration
python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token record /dev/video0 --duration 30

# Custom filename
python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token record /dev/video0 --filename my_recording.mp4

# Verbose output
python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token record /dev/video0 --duration 30 --verbose
```

**Output Example:**
```
✅ Recording started: /opt/camera-service/recordings/recording_20250806_143022.mp4
   Recording ID: rec_1234567890
   Start Time: 2025-08-06 14:30:22
   Duration: 30 seconds
```

### Stop Recording

Stop recording from a camera:

```bash
# Stop recording
python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token stop /dev/video0

# Verbose output
python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token stop /dev/video0 --verbose
```

**Output Example:**
```
✅ Recording stopped: /opt/camera-service/recordings/recording_20250806_143022.mp4
   Duration: 25.67 seconds
   Size: 1048576 bytes
```

### Ping Test

Test connection to the camera service:

```bash
# Test connection
python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token ping

# Verbose output
python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token ping --verbose
```

**Output Example:**
```
✅ Ping response: pong
```

### Monitor Cameras

Monitor cameras in real-time:

```bash
# Monitor cameras
python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token monitor
```

**Output Example:**
```
👀 Monitoring cameras (press Ctrl+C to stop)...

📹 Monitoring 2 camera(s):
   - USB Camera (/dev/video0): available
   - Built-in Camera (/dev/video1): busy

[14:30:25] 📹 Camera USB Camera: available
[14:30:30] 🎥 Recording recording_20250806_143022.mp4: active
[14:30:35] 📹 Camera Built-in Camera: available
```

## Output Formats

### Table Format (Default)

Human-readable table format:

```bash
python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token list --format table
```

### JSON Format

Machine-readable JSON format:

```bash
python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token list --format json
```

**Output Example:**
```json
[
  {
    "device_path": "/dev/video0",
    "name": "USB Camera",
    "status": "available",
    "capabilities": ["snapshot", "recording", "streaming"],
    "stream_url": "rtsp://localhost:8554/camera0"
  }
]
```

### CSV Format

Comma-separated values format:

```bash
python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token list --format csv
```

**Output Example:**
```csv
device_path,name,status,capabilities,stream_url
/dev/video0,USB Camera,available,"snapshot,recording,streaming",rtsp://localhost:8554/camera0
```

## Advanced Usage

### Scripting Examples

#### Batch Snapshot Script

```bash
#!/bin/bash
# Take snapshots from all cameras

CAMERAS=$(python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token list --format json | jq -r '.[].device_path')

for camera in $CAMERAS; do
    echo "Taking snapshot from $camera..."
    python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token snapshot "$camera"
done
```

#### Recording Monitor Script

```bash
#!/bin/bash
# Monitor and stop recordings after duration

CAMERA="/dev/video0"
DURATION=60

echo "Starting recording for $DURATION seconds..."
python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token record "$CAMERA" --duration "$DURATION"

echo "Waiting for recording to complete..."
sleep "$DURATION"

echo "Stopping recording..."
python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token stop "$CAMERA"
```

#### Health Check Script

```bash
#!/bin/bash
# Health check script

HOST="localhost"
PORT="8080"
TOKEN="your_jwt_token"

# Test connection
if python camera_cli.py --host "$HOST" --port "$PORT" --auth-type jwt --token "$TOKEN" ping; then
    echo "✅ Connection OK"
else
    echo "❌ Connection failed"
    exit 1
fi

# Check camera availability
CAMERAS=$(python camera_cli.py --host "$HOST" --port "$PORT" --auth-type jwt --token "$TOKEN" list --format json | jq length)

if [ "$CAMERAS" -gt 0 ]; then
    echo "✅ Found $CAMERAS camera(s)"
else
    echo "❌ No cameras available"
    exit 1
fi
```

### Environment Variables

Use environment variables for configuration:

```bash
# Set environment variables
export CAMERA_HOST="localhost"
export CAMERA_PORT="8080"
export CAMERA_TOKEN="your_jwt_token"

# Use in commands
python camera_cli.py --host "$CAMERA_HOST" --port "$CAMERA_PORT" --auth-type jwt --token "$CAMERA_TOKEN" list
```

### Configuration File

Create a configuration file for repeated use:

```bash
# Create config file
cat > camera_config.sh << 'EOF'
#!/bin/bash
export CAMERA_HOST="localhost"
export CAMERA_PORT="8080"
export CAMERA_TOKEN="your_jwt_token"
export CAMERA_AUTH_TYPE="jwt"
EOF

# Source config and use
source camera_config.sh
python camera_cli.py --host "$CAMERA_HOST" --port "$CAMERA_PORT" --auth-type "$CAMERA_AUTH_TYPE" --token "$CAMERA_TOKEN" list
```

## Error Handling

### Common Error Messages

1. **Authentication Failed**
   ```
   ❌ Error: JWT token required for jwt authentication
   ```
   - Solution: Provide valid JWT token or API key

2. **Connection Refused**
   ```
   ❌ Error: Failed to connect after 3 attempts: Connection refused
   ```
   - Solution: Check if camera service is running

3. **Camera Not Found**
   ```
   ❌ Camera not found: /dev/video0
   ```
   - Solution: Check device path and camera availability

4. **Operation Failed**
   ```
   ❌ Snapshot failed: Camera is busy
   ```
   - Solution: Wait for camera to become available

### Exit Codes

- `0`: Success
- `1`: General error
- `130`: Interrupted by user (Ctrl+C)

### Verbose Debugging

Use `--verbose` flag for detailed output:

```bash
python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token list --verbose
```

## Best Practices

### 1. Use Configuration Files

```bash
# Create reusable configuration
cat > camera_env.sh << 'EOF'
export CAMERA_HOST="localhost"
export CAMERA_PORT="8080"
export CAMERA_TOKEN="your_jwt_token"
export CAMERA_AUTH_TYPE="jwt"
EOF

# Source before use
source camera_env.sh
```

### 2. Implement Error Handling

```bash
#!/bin/bash
# Robust error handling

set -e  # Exit on error

function camera_operation() {
    local operation="$1"
    local device="$2"
    
    if python camera_cli.py --host "$CAMERA_HOST" --port "$CAMERA_PORT" --auth-type "$CAMERA_AUTH_TYPE" --token "$CAMERA_TOKEN" "$operation" "$device"; then
        echo "✅ $operation successful"
    else
        echo "❌ $operation failed"
        return 1
    fi
}

# Use with error handling
camera_operation "snapshot" "/dev/video0" || exit 1
```

### 3. Use JSON Output for Scripting

```bash
# Parse JSON output in scripts
CAMERAS=$(python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token list --format json)

# Extract camera names
echo "$CAMERAS" | jq -r '.[].name'

# Check camera status
echo "$CAMERAS" | jq -r '.[] | select(.status == "available") | .device_path'
```

### 4. Implement Timeouts

```bash
#!/bin/bash
# Timeout wrapper

function timeout_command() {
    local timeout="$1"
    shift
    
    timeout "$timeout" "$@"
    local exit_code=$?
    
    if [ $exit_code -eq 124 ]; then
        echo "❌ Operation timed out after ${timeout}s"
        return 1
    fi
    
    return $exit_code
}

# Use with timeout
timeout_command 30 python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token snapshot /dev/video0
```

## Troubleshooting

### Connection Issues

1. **Check Service Status**
   ```bash
   # Test basic connectivity
   python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token ping
   ```

2. **Verify Authentication**
   ```bash
   # Test with verbose output
   python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token ping --verbose
   ```

3. **Check Network**
   ```bash
   # Test network connectivity
   telnet localhost 8080
   ```

### Camera Issues

1. **List Available Cameras**
   ```bash
   python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token list
   ```

2. **Check Camera Status**
   ```bash
   python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token status /dev/video0
   ```

3. **Monitor Camera Changes**
   ```bash
   python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token monitor
   ```

### Performance Issues

1. **Use JSON Output for Large Lists**
   ```bash
   python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token list --format json | jq .
   ```

2. **Implement Timeouts**
   ```bash
   timeout 30 python camera_cli.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token snapshot /dev/video0
   ```

## API Reference

### Command Line Options

#### Connection Options

- `--host`: Server hostname (default: localhost)
- `--port`: Server port (default: 8080)
- `--ssl`: Use SSL/TLS connection
- `--auth-type`: Authentication type (jwt, api_key)
- `--token`: JWT token for authentication
- `--key`: API key for authentication
- `--timeout`: Request timeout in seconds (default: 30)

#### Output Options

- `--format`: Output format (table, json, csv)
- `--verbose`: Verbose output
- `--filename`: Custom filename for operations
- `--duration`: Recording duration in seconds

#### Commands

- `list`: List available cameras
- `status <device>`: Get camera status
- `snapshot <device>`: Take camera snapshot
- `record <device>`: Start recording
- `stop <device>`: Stop recording
- `ping`: Test connection
- `monitor`: Monitor cameras in real-time

### Exit Codes

- `0`: Success
- `1`: General error
- `130`: Interrupted by user

### Error Types

- `CameraServiceError`: General camera service error
- `AuthenticationError`: Authentication failed
- `ConnectionError`: Connection failed
- `CameraNotFoundError`: Camera device not found
- `MediaMTXError`: MediaMTX operation failed

## Support

For issues and questions:

1. Check the troubleshooting section above
2. Use `--verbose` flag for detailed output
3. Verify authentication credentials
4. Ensure camera service is running
5. Check camera device permissions

### Getting Help

```bash
# Show help
python camera_cli.py --help

# Show command help
python camera_cli.py list --help
```

---

**Version:** 1.0  
**Last Updated:** 2025-08-06  
**Compatibility:** Python 3.8+, websockets, asyncio 