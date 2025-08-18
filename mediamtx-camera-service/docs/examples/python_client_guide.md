# Python Client Guide - MediaMTX Camera Service

## Overview

The Python client example demonstrates how to connect to the MediaMTX Camera Service using WebSocket JSON-RPC 2.0 protocol with comprehensive authentication support and error handling.

## Features

- **Dual Authentication Support**: JWT tokens and API keys
- **WebSocket Connection Management**: Automatic reconnection and retry logic
- **Camera Operations**: Discovery, status monitoring, snapshots, and recording
- **Real-time Notifications**: Camera and recording status updates
- **Comprehensive Error Handling**: Custom exceptions and recovery strategies
- **SSL/TLS Support**: Secure connections with certificate validation
- **Async/Await**: Modern Python asynchronous programming

## Installation

### Prerequisites

```bash
# Install required packages
pip install websockets asyncio
```

### Quick Start

```bash
# Clone the repository
git clone <repository-url>
cd mediamtx-camera-service

# Run the Python client example
python examples/python/camera_client.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token
```

## Authentication

### JWT Authentication

```python
from examples.python.camera_client import CameraClient

# Create client with JWT authentication
client = CameraClient(
    host="localhost",
    port=8080,
    auth_type="jwt",
    auth_token="your_jwt_token_here"
)
```

### API Key Authentication

```python
# Create client with API key authentication
client = CameraClient(
    host="localhost",
    port=8080,
    auth_type="api_key",
    api_key="your_api_key_here"
)
```

### SSL/TLS Support

```python
# Create client with SSL/TLS
client = CameraClient(
    host="localhost",
    port=8080,
    use_ssl=True,
    auth_type="jwt",
    auth_token="your_jwt_token_here"
)
```

## Basic Usage

### Connection Management

```python
import asyncio
from examples.python.camera_client import CameraClient

async def main():
    # Create client
    client = CameraClient(
        host="localhost",
        port=8080,
        auth_type="jwt",
        auth_token="your_token"
    )
    
    try:
        # Connect to service
        await client.connect()
        print("Connected successfully")
        
        # Your camera operations here
        
    except Exception as e:
        print(f"Connection failed: {e}")
    finally:
        # Always disconnect
        await client.disconnect()

# Run the example
asyncio.run(main())
```

### Camera Discovery

```python
# Get list of available cameras
cameras = await client.get_camera_list()
print(f"Found {len(cameras)} cameras:")

for camera in cameras:
    print(f"  - {camera.name} ({camera.device_path})")
    print(f"    Status: {camera.status}")
    print(f"    Capabilities: {camera.capabilities}")
    if camera.stream_url:
        print(f"    Stream URL: {camera.stream_url}")
```

### Camera Status Monitoring

```python
# Get status of specific camera
try:
    status = await client.get_camera_status("/dev/video0")
    print(f"Camera status: {status.status}")
    print(f"Capabilities: {status.capabilities}")
except CameraNotFoundError:
    print("Camera not found")
```

### Taking Snapshots

```python
# Take a snapshot
try:
    snapshot = await client.take_snapshot("/dev/video0")
    print(f"Snapshot saved: {snapshot['filename']}")
    print(f"File size: {snapshot['file_size']} bytes")
except CameraNotFoundError:
    print("Camera not found")
except MediaMTXError as e:
    print(f"Snapshot failed: {e}")
```

### Recording Operations

```python
# Start recording
try:
    recording = await client.start_recording(
        device_path="/dev/video0",
        duration=60,  # 60 seconds
        custom_filename="my_recording.mp4"
    )
    print(f"Recording started: {recording.filename}")
    print(f"Recording ID: {recording.recording_id}")
    
    # Wait for some time
    await asyncio.sleep(30)
    
    # Stop recording
    stop_result = await client.stop_recording("/dev/video0")
    print(f"Recording stopped: {stop_result['filename']}")
    print(f"Duration: {stop_result['duration']} seconds")
    
except CameraNotFoundError:
    print("Camera not found")
except MediaMTXError as e:
    print(f"Recording failed: {e}")
```

## Advanced Features

### Real-time Notifications

```python
async def on_camera_status_update(params):
    """Handle camera status updates."""
    camera = params.get("camera", {})
    print(f"Camera {camera.get('name')} status changed to {camera.get('status')}")

async def on_recording_status_update(params):
    """Handle recording status updates."""
    recording = params.get("recording", {})
    print(f"Recording {recording.get('filename')} status: {recording.get('status')}")

async def on_connection_lost():
    """Handle connection lost events."""
    print("Connection lost - attempting to reconnect...")

# Set up event handlers
client.set_camera_status_callback(on_camera_status_update)
client.set_recording_status_callback(on_recording_status_update)
client.set_connection_lost_callback(on_connection_lost)
```

### Error Handling

```python
from examples.python.camera_client import (
    CameraServiceError,
    AuthenticationError,
    ConnectionError,
    CameraNotFoundError,
    MediaMTXError
)

async def robust_camera_operations():
    try:
        await client.connect()
        
        # Get cameras
        cameras = await client.get_camera_list()
        
        if not cameras:
            print("No cameras available")
            return
        
        # Work with first camera
        camera = cameras[0]
        
        # Take snapshot
        snapshot = await client.take_snapshot(camera.device_path)
        print(f"Snapshot: {snapshot['filename']}")
        
        # Start recording
        recording = await client.start_recording(
            camera.device_path,
            duration=30
        )
        print(f"Recording: {recording.filename}")
        
        # Wait and stop
        await asyncio.sleep(10)
        await client.stop_recording(camera.device_path)
        
    except AuthenticationError:
        print("Authentication failed - check your credentials")
    except ConnectionError:
        print("Connection failed - check server status")
    except CameraNotFoundError:
        print("Camera not found - check device path")
    except MediaMTXError as e:
        print(f"MediaMTX operation failed: {e}")
    except CameraServiceError as e:
        print(f"Camera service error: {e}")
    except Exception as e:
        print(f"Unexpected error: {e}")
    finally:
        await client.disconnect()
```

### Retry Logic

```python
# Client with custom retry settings
client = CameraClient(
    host="localhost",
    port=8080,
    auth_type="jwt",
    auth_token="your_token",
    max_retries=5,      # More retries
    retry_delay=2.0     # Longer delay between retries
)
```

## Configuration Options

### Client Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `host` | str | "localhost" | Server hostname |
| `port` | int | 8080 | Server port |
| `use_ssl` | bool | False | Use SSL/TLS |
| `auth_type` | str | "jwt" | Authentication type ("jwt" or "api_key") |
| `auth_token` | str | None | JWT token |
| `api_key` | str | None | API key |
| `max_retries` | int | 3 | Maximum connection retries |
| `retry_delay` | float | 1.0 | Delay between retries (seconds) |

### Authentication Headers

The client automatically sends appropriate headers based on authentication type:

- **JWT**: `Authorization: Bearer <token>`
- **API Key**: `X-API-Key: <key>`

## Best Practices

### 1. Always Handle Exceptions

```python
try:
    await client.connect()
    # Your operations
except CameraServiceError as e:
    print(f"Service error: {e}")
finally:
    await client.disconnect()
```

### 2. Use Context Managers (Recommended)

```python
class CameraClientContext:
    def __init__(self, client):
        self.client = client
    
    async def __aenter__(self):
        await self.client.connect()
        return self.client
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        await self.client.disconnect()

# Usage
async with CameraClientContext(client) as camera:
    cameras = await camera.get_camera_list()
    # ... other operations
```

### 3. Implement Proper Logging

```python
import logging

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)

# Client will use the configured logger
client = CameraClient(...)
```

### 4. Handle Connection Loss

```python
async def on_connection_lost():
    print("Connection lost - attempting to reconnect...")
    try:
        await client.connect()
        print("Reconnected successfully")
    except Exception as e:
        print(f"Reconnection failed: {e}")

client.set_connection_lost_callback(on_connection_lost)
```

### 5. Use Timeouts

```python
# The client has built-in request timeouts (30 seconds)
# For long operations, consider implementing custom timeouts
import asyncio

try:
    # Custom timeout for long operations
    result = await asyncio.wait_for(
        client.start_recording("/dev/video0", duration=300),
        timeout=60.0
    )
except asyncio.TimeoutError:
    print("Operation timed out")
```

## Troubleshooting

### Common Issues

1. **Connection Refused**
   - Check if the camera service is running
   - Verify host and port settings
   - Check firewall settings

2. **Authentication Failed**
   - Verify JWT token or API key
   - Check token expiration
   - Ensure correct authentication type

3. **Camera Not Found**
   - Verify device path exists
   - Check camera permissions
   - Ensure camera is not in use by another process

4. **SSL/TLS Errors**
   - Check certificate validity
   - Verify SSL configuration
   - Use `use_ssl=False` for testing

### Debug Mode

```python
import logging

# Enable debug logging
logging.getLogger().setLevel(logging.DEBUG)

# Create client with debug info
client = CameraClient(...)
```

## Performance Considerations

### Connection Pooling

For high-throughput applications, consider implementing connection pooling:

```python
class CameraClientPool:
    def __init__(self, pool_size=5):
        self.pool_size = pool_size
        self.clients = []
        self.available = asyncio.Queue()
    
    async def initialize(self):
        for _ in range(self.pool_size):
            client = CameraClient(...)
            await client.connect()
            await self.available.put(client)
    
    async def get_client(self):
        return await self.available.get()
    
    async def return_client(self, client):
        await self.available.put(client)
```

### Batch Operations

For multiple cameras, consider parallel operations:

```python
async def process_all_cameras():
    cameras = await client.get_camera_list()
    
    # Parallel snapshot operations
    tasks = [
        client.take_snapshot(camera.device_path)
        for camera in cameras
    ]
    
    results = await asyncio.gather(*tasks, return_exceptions=True)
    
    for camera, result in zip(cameras, results):
        if isinstance(result, Exception):
            print(f"Failed to snapshot {camera.name}: {result}")
        else:
            print(f"Snapshot {camera.name}: {result['filename']}")
```

## Security Considerations

### Token Management

```python
import os
from datetime import datetime, timedelta

# Store tokens securely
JWT_TOKEN = os.environ.get('CAMERA_SERVICE_JWT_TOKEN')
API_KEY = os.environ.get('CAMERA_SERVICE_API_KEY')

# Rotate tokens regularly
def is_token_expired(token):
    # Implement token expiration check
    pass
```

### SSL/TLS Best Practices

```python
import ssl

# Create secure SSL context
ssl_context = ssl.create_default_context()
ssl_context.check_hostname = True
ssl_context.verify_mode = ssl.CERT_REQUIRED

# Use with client
client = CameraClient(
    host="secure.example.com",
    port=443,
    use_ssl=True,
    # SSL context will be used internally
)
```

## Examples

### File Management

The Python client supports comprehensive file management operations for recordings and snapshots.

#### Listing Files

```python
# List recordings with pagination
recordings = await client.list_recordings(limit=10, offset=0)
print(f"Found {len(recordings['files'])} recordings (total: {recordings['total']})")

for recording in recordings['files']:
    print(f"📹 {recording['filename']}")
    print(f"   Size: {recording['file_size']} bytes")
    print(f"   Modified: {recording['modified_time']}")
    print(f"   Download: {recording['download_url']}")

# List snapshots with pagination
snapshots = await client.list_snapshots(limit=10, offset=0)
print(f"Found {len(snapshots['files'])} snapshots (total: {snapshots['total']})")

for snapshot in snapshots['files']:
    print(f"📸 {snapshot['filename']}")
    print(f"   Size: {snapshot['file_size']} bytes")
    print(f"   Modified: {snapshot['modified_time']}")
    print(f"   Download: {snapshot['download_url']}")
```

#### Downloading Files

```python
# Download a recording file
try:
    local_path = await client.download_file(
        file_type='recordings',
        filename='camera0_2025-01-15_14-30-00.mp4',
        local_path='./downloads/recording.mp4'
    )
    print(f"✅ Recording downloaded: {local_path}")
except Exception as e:
    print(f"❌ Download failed: {e}")

# Download a snapshot file
try:
    local_path = await client.download_file(
        file_type='snapshots',
        filename='snapshot_2025-01-15_14-30-00.jpg'
    )
    print(f"✅ Snapshot downloaded: {local_path}")
except Exception as e:
    print(f"❌ Download failed: {e}")
```

#### Complete File Management Example

```python
async def file_management_demo():
    """Demonstrate file management operations."""
    client = CameraClient(
        host="localhost",
        port=8080,
        auth_type="jwt",
        auth_token="your_token"
    )
    
    try:
        await client.connect()
        
        # List recent recordings
        recordings = await client.list_recordings(limit=5)
        print(f"📹 Recent recordings: {len(recordings['files'])}")
        
        # List recent snapshots
        snapshots = await client.list_snapshots(limit=5)
        print(f"📸 Recent snapshots: {len(snapshots['files'])}")
        
        # Download the most recent snapshot
        if snapshots['files']:
            latest_snapshot = snapshots['files'][0]
            local_path = await client.download_file(
                'snapshots',
                latest_snapshot['filename'],
                f"./downloads/{latest_snapshot['filename']}"
            )
            print(f"✅ Downloaded: {local_path}")
        
        # Download the most recent recording
        if recordings['files']:
            latest_recording = recordings['files'][0]
            local_path = await client.download_file(
                'recordings',
                latest_recording['filename'],
                f"./downloads/{latest_recording['filename']}"
            )
            print(f"✅ Downloaded: {local_path}")
            
    except Exception as e:
        print(f"❌ Error: {e}")
    finally:
        await client.disconnect()

# Run the demo
asyncio.run(file_management_demo())
```

### Complete Working Example

```python
#!/usr/bin/env python3
"""
Complete camera service client example.
"""

import asyncio
import logging
from examples.python.camera_client import CameraClient

async def camera_monitor():
    """Monitor and control cameras."""
    
    # Setup logging
    logging.basicConfig(level=logging.INFO)
    
    # Create client
    client = CameraClient(
        host="localhost",
        port=8080,
        auth_type="jwt",
        auth_token="your_token_here"
    )
    
    # Set up event handlers
    async def on_camera_update(params):
        camera = params.get("camera", {})
        print(f"📹 Camera {camera.get('name')}: {camera.get('status')}")
    
    async def on_recording_update(params):
        recording = params.get("recording", {})
        print(f"🎥 Recording {recording.get('filename')}: {recording.get('status')}")
    
    client.set_camera_status_callback(on_camera_update)
    client.set_recording_status_callback(on_recording_update)
    
    try:
        # Connect
        await client.connect()
        print("✅ Connected to camera service")
        
        # Get cameras
        cameras = await client.get_camera_list()
        print(f"📹 Found {len(cameras)} cameras")
        
        if cameras:
            camera = cameras[0]
            print(f"🎯 Working with camera: {camera.name}")
            
            # Take snapshot
            snapshot = await client.take_snapshot(camera.device_path)
            print(f"📸 Snapshot: {snapshot['filename']}")
            
            # Start recording
            recording = await client.start_recording(
                camera.device_path,
                duration=30,
                custom_filename="demo_recording.mp4"
            )
            print(f"🎥 Recording started: {recording.filename}")
            
            # Monitor for 20 seconds
            await asyncio.sleep(20)
            
            # Stop recording
            stop_result = await client.stop_recording(camera.device_path)
            print(f"⏹️ Recording stopped: {stop_result['filename']}")
            
            # List recordings
            recordings = await client.list_recordings(limit=5)
            print(f"📹 Found {len(recordings.get('files', []))} recordings:")
            for recording in recordings.get('files', [])[:3]:
                print(f"  - {recording['filename']} ({recording['file_size']} bytes)")
            
            # List snapshots
            snapshots = await client.list_snapshots(limit=5)
            print(f"📸 Found {len(snapshots.get('files', []))} snapshots:")
            for snapshot in snapshots.get('files', [])[:3]:
                print(f"  - {snapshot['filename']} ({snapshot['file_size']} bytes)")
            
            # Download a snapshot if available
            if snapshots.get('files'):
                snapshot_file = snapshots['files'][0]['filename']
                try:
                    local_path = await client.download_file('snapshots', snapshot_file)
                    print(f"✅ Downloaded snapshot: {local_path}")
                except Exception as e:
                    print(f"⚠️ Download failed: {e}")
        
        # Keep monitoring
        print("👀 Monitoring cameras (press Ctrl+C to stop)...")
        await asyncio.sleep(60)
        
    except KeyboardInterrupt:
        print("\n🛑 Stopping monitor...")
    except Exception as e:
        print(f"❌ Error: {e}")
    finally:
        await client.disconnect()
        print("✅ Disconnected")

if __name__ == "__main__":
    asyncio.run(camera_monitor())
```

## API Reference

### CameraClient Class

#### Methods

- `connect()` - Connect to camera service
- `disconnect()` - Disconnect from camera service
- `ping()` - Test connection
- `get_camera_list()` - Get available cameras
- `get_camera_status(device_path)` - Get camera status
- `take_snapshot(device_path, custom_filename)` - Take snapshot
- `start_recording(device_path, duration, custom_filename)` - Start recording
- `stop_recording(device_path)` - Stop recording
- `list_recordings(limit, offset)` - List available recording files
- `list_snapshots(limit, offset)` - List available snapshot files
- `download_file(file_type, filename, local_path)` - Download a file from the server

#### Event Handlers

- `set_camera_status_callback(callback)` - Set camera status update handler
- `set_recording_status_callback(callback)` - Set recording status update handler
- `set_connection_lost_callback(callback)` - Set connection lost handler

### Data Classes

#### CameraInfo

```python
@dataclass
class CameraInfo:
    device_path: str
    name: str
    capabilities: List[str]
    status: str
    stream_url: Optional[str] = None
```

#### RecordingInfo

```python
@dataclass
class RecordingInfo:
    device_path: str
    recording_id: str
    filename: str
    start_time: float
    duration: Optional[float] = None
    status: str = "active"
```

### Exceptions

- `CameraServiceError` - Base exception for camera service errors
- `AuthenticationError` - Authentication failed
- `ConnectionError` - Connection failed
- `CameraNotFoundError` - Camera device not found
- `MediaMTXError` - MediaMTX operation failed

## Support

For issues and questions:

1. Check the troubleshooting section above
2. Review the error messages and logs
3. Verify your authentication credentials
4. Ensure the camera service is running
5. Check camera device permissions

---

**Version:** 1.0  
**Last Updated:** 2025-08-06  
**Compatibility:** Python 3.8+, asyncio, websockets 