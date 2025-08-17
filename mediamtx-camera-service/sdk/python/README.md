# MediaMTX Camera Service Python SDK

**Version:** 1.0.0  
**Status:** Production Ready  
**Epic:** E3 Client API & SDK Ecosystem  

## Overview

The MediaMTX Camera Service Python SDK provides a high-level interface for interacting with the MediaMTX Camera Service via WebSocket JSON-RPC protocol. The SDK supports both JWT token and API key authentication, making it suitable for both user applications and service-to-service communication.

## Features

- **WebSocket JSON-RPC Client**: Full support for the MediaMTX Camera Service API
- **Authentication**: JWT token and API key authentication support
- **Camera Control**: Discover, monitor, and control camera devices
- **Media Operations**: Take snapshots, start/stop recordings
- **Real-time Notifications**: Receive camera status and recording updates
- **Error Handling**: Comprehensive error handling with custom exceptions
- **Connection Management**: Automatic reconnection and retry logic
- **Type Safety**: Full type hints for better development experience

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/mediamtx/camera-service.git
cd camera-service/sdk/python

# Install in development mode
pip install -e .
```

### From PyPI (Future)

```bash
pip install mediamtx-camera-sdk
```

## Quick Start

### Basic Usage

```python
import asyncio
from mediamtx_camera_sdk import CameraClient

async def main():
    # Create client with JWT authentication
    client = CameraClient(
        host="localhost",
        port=8080,
        auth_type="jwt",
        auth_token="your_jwt_token_here"
    )
    
    try:
        # Connect to service
        await client.connect()
        
        # List available cameras
        cameras = await client.get_camera_list()
        print(f"Found {len(cameras)} cameras")
        
        # Take a snapshot
        if cameras:
            snapshot_info = await client.take_snapshot(cameras[0].device_path)
            print(f"Snapshot saved: {snapshot_info.filename}")
            
    finally:
        await client.disconnect()

# Run the example
asyncio.run(main())
```

### API Key Authentication

```python
from mediamtx_camera_sdk import CameraClient

async def main():
    # Create client with API key authentication
    client = CameraClient(
        host="localhost",
        port=8080,
        auth_type="api_key",
        api_key="your_api_key_here"
    )
    
    await client.connect()
    # ... use the client
    await client.disconnect()
```

### Real-time Notifications

```python
from mediamtx_camera_sdk import CameraClient

async def on_camera_status_update(camera_info):
    print(f"Camera {camera_info.device_path} status: {camera_info.status}")

async def on_recording_status_update(recording_info):
    print(f"Recording {recording_info.recording_id}: {recording_info.status}")

async def main():
    client = CameraClient(
        host="localhost",
        port=8080,
        auth_type="jwt",
        auth_token="your_jwt_token_here"
    )
    
    # Set up event handlers
    client.on_camera_status_update = on_camera_status_update
    client.on_recording_status_update = on_recording_status_update
    
    await client.connect()
    
    # Keep connection alive to receive notifications
    try:
        while True:
            await asyncio.sleep(1)
    except KeyboardInterrupt:
        pass
    finally:
        await client.disconnect()
```

## API Reference

### CameraClient

The main client class for interacting with the MediaMTX Camera Service.

#### Constructor

```python
CameraClient(
    host: str = "localhost",
    port: int = 8080,
    use_ssl: bool = False,
    auth_type: str = "jwt",
    auth_token: Optional[str] = None,
    api_key: Optional[str] = None,
    max_retries: int = 3,
    retry_delay: float = 1.0,
)
```

#### Methods

##### Connection Management

- `async connect()` - Connect to the camera service
- `async disconnect()` - Disconnect from the camera service
- `async ping()` - Test connection with ping

##### Camera Operations

- `async get_camera_list()` - Get list of available cameras
- `async get_camera_status(device_path: str)` - Get camera status
- `async take_snapshot(device_path: str, filename: Optional[str] = None)` - Take camera snapshot
- `async start_recording(device_path: str, filename: Optional[str] = None)` - Start recording
- `async stop_recording(device_path: str)` - Stop recording
- `async get_recording_status(device_path: str)` - Get recording status

##### Event Handlers

- `on_camera_status_update(camera_info: CameraInfo)` - Called when camera status changes
- `on_recording_status_update(recording_info: RecordingInfo)` - Called when recording status changes
- `on_connection_lost()` - Called when connection is lost

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

- `CameraServiceError` - Base exception for all camera service errors
- `AuthenticationError` - Authentication failed
- `ConnectionError` - Connection failed
- `CameraNotFoundError` - Camera device not found
- `MediaMTXError` - MediaMTX operation failed

## Error Handling

The SDK provides comprehensive error handling with custom exceptions:

```python
from mediamtx_camera_sdk import CameraClient, AuthenticationError, ConnectionError

async def main():
    client = CameraClient(
        host="localhost",
        port=8080,
        auth_type="jwt",
        auth_token="invalid_token"
    )
    
    try:
        await client.connect()
    except AuthenticationError as e:
        print(f"Authentication failed: {e}")
    except ConnectionError as e:
        print(f"Connection failed: {e}")
    except Exception as e:
        print(f"Unexpected error: {e}")
```

## Configuration

### Environment Variables

The SDK respects the following environment variables:

- `MEDIAMTX_HOST` - Default host (default: localhost)
- `MEDIAMTX_PORT` - Default port (default: 8080)
- `MEDIAMTX_USE_SSL` - Use SSL/TLS (default: false)
- `MEDIAMTX_AUTH_TYPE` - Default auth type (default: jwt)
- `MEDIAMTX_AUTH_TOKEN` - Default JWT token
- `MEDIAMTX_API_KEY` - Default API key

### SSL/TLS Configuration

```python
client = CameraClient(
    host="localhost",
    port=8080,
    use_ssl=True,  # Enable SSL/TLS
    auth_type="jwt",
    auth_token="your_token"
)
```

## Development

### Running Tests

```bash
cd sdk/python
pip install -e ".[dev]"
pytest
```

### Code Quality

```bash
# Format code
black .

# Lint code
flake8 .

# Type checking
mypy .
```

## License

MIT License - see LICENSE file for details.

## Support

For support and questions:

- GitHub Issues: https://github.com/mediamtx/camera-service/issues
- Documentation: https://mediamtx-camera-service.readthedocs.io
- Email: team@mediamtx-camera-service.com
