# MediaMTX Camera Service JavaScript SDK

**Version:** 1.0.0  
**Status:** Production Ready  
**Epic:** E3 Client API & SDK Ecosystem  

## Overview

The MediaMTX Camera Service JavaScript SDK provides a high-level interface for interacting with the MediaMTX Camera Service via WebSocket JSON-RPC protocol. The SDK supports both JWT token and API key authentication, making it suitable for both Node.js applications and browser environments.

## Features

- **WebSocket JSON-RPC Client**: Full support for the MediaMTX Camera Service API
- **Authentication**: JWT token and API key authentication support
- **Camera Control**: Discover, monitor, and control camera devices
- **Media Operations**: Take snapshots, start/stop recordings
- **Real-time Notifications**: Receive camera status and recording updates
- **Error Handling**: Comprehensive error handling with custom exceptions
- **Connection Management**: Automatic reconnection and retry logic
- **TypeScript Support**: Full type definitions for better development experience
- **Cross-platform**: Works in Node.js and browser environments

## Requirements

- Node.js >= 12.0.0 (tested with v12.22.9)
- npm or yarn package manager
- TypeScript >= 4.0.0 (for development)

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/mediamtx/camera-service.git
cd camera-service/sdk/javascript

# Install dependencies
npm install

# Build the SDK
npm run build

# Install in development mode
npm link
```

### From NPM (Future)

```bash
npm install mediamtx-camera-sdk
```

## Quick Start

### Basic Usage

```javascript
import { CameraClient } from 'mediamtx-camera-sdk';

async function main() {
    // Create client with JWT authentication
    const client = new CameraClient({
        host: 'localhost',
        port: 8080,
        authType: 'jwt',
        authToken: 'your_jwt_token_here'
    });
    
    try {
        // Connect to service
        await client.connect();
        
        // List available cameras
        const cameras = await client.getCameraList();
        console.log(`Found ${cameras.length} cameras`);
        
        // Take a snapshot
        if (cameras.length > 0) {
            const snapshotInfo = await client.takeSnapshot(cameras[0].devicePath);
            console.log(`Snapshot saved: ${snapshotInfo.filename}`);
        }
        
    } finally {
        await client.disconnect();
    }
}

// Run the example
main().catch(console.error);
```

### API Key Authentication

```javascript
import { CameraClient } from 'mediamtx-camera-sdk';

async function main() {
    // Create client with API key authentication
    const client = new CameraClient({
        host: 'localhost',
        port: 8080,
        authType: 'api_key',
        apiKey: 'your_api_key_here'
    });
    
    await client.connect();
    // ... use the client
    await client.disconnect();
}
```

### Real-time Notifications

```javascript
import { CameraClient } from 'mediamtx-camera-sdk';

async function onCameraStatusUpdate(cameraInfo) {
    console.log(`Camera ${cameraInfo.devicePath} status: ${cameraInfo.status}`);
}

async function onRecordingStatusUpdate(recordingInfo) {
    console.log(`Recording ${recordingInfo.recordingId}: ${recordingInfo.status}`);
}

async function main() {
    const client = new CameraClient({
        host: 'localhost',
        port: 8080,
        authType: 'jwt',
        authToken: 'your_jwt_token_here'
    });
    
    // Set up event handlers
    client.onCameraStatusUpdate = onCameraStatusUpdate;
    client.onRecordingStatusUpdate = onRecordingStatusUpdate;
    
    await client.connect();
    
    // Keep connection alive to receive notifications
    try {
        while (true) {
            await new Promise(resolve => setTimeout(resolve, 1000));
        }
    } catch (error) {
        console.error('Error:', error);
    } finally {
        await client.disconnect();
    }
}
```

## API Reference

### CameraClient

The main client class for interacting with the MediaMTX Camera Service.

#### Constructor

```javascript
new CameraClient({
    host?: string,
    port?: number,
    useSsl?: boolean,
    authType?: 'jwt' | 'api_key',
    authToken?: string,
    apiKey?: string,
    maxRetries?: number,
    retryDelay?: number,
})
```

#### Methods

##### Connection Management

- `connect()` - Connect to the camera service
- `disconnect()` - Disconnect from the camera service
- `ping()` - Test connection with ping

##### Camera Operations

- `getCameraList()` - Get list of available cameras
- `getCameraStatus(devicePath: string)` - Get camera status
- `takeSnapshot(devicePath: string, filename?: string)` - Take camera snapshot
- `startRecording(devicePath: string, filename?: string)` - Start recording
- `stopRecording(devicePath: string)` - Stop recording
- `getRecordingStatus(devicePath: string)` - Get recording status

##### Event Handlers

- `onCameraStatusUpdate(cameraInfo: CameraInfo)` - Called when camera status changes
- `onRecordingStatusUpdate(recordingInfo: RecordingInfo)` - Called when recording status changes
- `onConnectionLost()` - Called when connection is lost

### Data Classes

#### CameraInfo

```typescript
interface CameraInfo {
    devicePath: string;
    name: string;
    capabilities: string[];
    status: string;
    streamUrl?: string;
}
```

#### RecordingInfo

```typescript
interface RecordingInfo {
    devicePath: string;
    recordingId: string;
    filename: string;
    startTime: number;
    duration?: number;
    status: string;
}
```

#### SnapshotInfo

```typescript
interface SnapshotInfo {
    devicePath: string;
    filename: string;
    timestamp: number;
    sizeBytes?: number;
}
```

### Exceptions

- `CameraServiceError` - Base exception for all camera service errors
- `AuthenticationError` - Authentication failed
- `ConnectionError` - Connection failed
- `CameraNotFoundError` - Camera device not found
- `MediaMTXError` - MediaMTX operation failed

## Error Handling

The SDK provides comprehensive error handling with custom exceptions:

```javascript
import { CameraClient, AuthenticationError, ConnectionError } from 'mediamtx-camera-sdk';

async function main() {
    const client = new CameraClient({
        host: 'localhost',
        port: 8080,
        authType: 'jwt',
        authToken: 'invalid_token'
    });
    
    try {
        await client.connect();
    } catch (error) {
        if (error instanceof AuthenticationError) {
            console.error(`Authentication failed: ${error.message}`);
        } else if (error instanceof ConnectionError) {
            console.error(`Connection failed: ${error.message}`);
        } else {
            console.error(`Unexpected error: ${error.message}`);
        }
    }
}
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

```javascript
const client = new CameraClient({
    host: 'localhost',
    port: 8080,
    useSsl: true,  // Enable SSL/TLS
    authType: 'jwt',
    authToken: 'your_token'
});
```

## Browser Usage

The SDK can be used in browser environments with WebSocket support:

```html
<!DOCTYPE html>
<html>
<head>
    <title>MediaMTX Camera Service</title>
</head>
<body>
    <script type="module">
        import { CameraClient } from './node_modules/mediamtx-camera-sdk/dist/index.js';
        
        const client = new CameraClient({
            host: 'localhost',
            port: 8080,
            authType: 'jwt',
            authToken: 'your_token'
        });
        
        client.connect()
            .then(() => client.getCameraList())
            .then(cameras => console.log('Cameras:', cameras))
            .catch(console.error);
    </script>
</body>
</html>
```

## Development

### Running Tests

```bash
cd sdk/javascript
npm install
npm test
```

### Code Quality

```bash
# Format code
npm run format

# Lint code
npm run lint

# Build
npm run build
```

### TypeScript Configuration

The SDK includes TypeScript definitions and can be used with TypeScript projects:

```typescript
import { CameraClient, CameraInfo, RecordingInfo } from 'mediamtx-camera-sdk';

const client: CameraClient = new CameraClient({
    host: 'localhost',
    port: 8080,
    authType: 'jwt',
    authToken: 'your_token'
});

async function handleCameraUpdate(cameraInfo: CameraInfo): Promise<void> {
    console.log(`Camera ${cameraInfo.devicePath} updated: ${cameraInfo.status}`);
}

client.onCameraStatusUpdate = handleCameraUpdate;
```

## License

MIT License - see LICENSE file for details.

## Support

For support and questions:

- GitHub Issues: https://github.com/mediamtx/camera-service/issues
- Documentation: https://mediamtx-camera-service.readthedocs.io
- Email: team@mediamtx-camera-service.com
