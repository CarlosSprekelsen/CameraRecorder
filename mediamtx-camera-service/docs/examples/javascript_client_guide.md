# JavaScript/Node.js Client Guide - MediaMTX Camera Service

## Overview

The JavaScript/Node.js client example demonstrates how to connect to the MediaMTX Camera Service using WebSocket JSON-RPC 2.0 protocol with comprehensive authentication support and error handling.

## Features

- **Dual Authentication Support**: JWT tokens and API keys
- **WebSocket Connection Management**: Automatic reconnection and retry logic
- **Camera Operations**: Discovery, status monitoring, snapshots, and recording
- **Real-time Notifications**: Camera and recording status updates
- **Comprehensive Error Handling**: Custom exceptions and recovery strategies
- **SSL/TLS Support**: Secure connections with certificate validation
- **Async/Await**: Modern JavaScript asynchronous programming

## Installation

### Prerequisites

```bash
# Install Node.js (version 14 or higher)
node --version

# Install required packages
npm install ws uuid
```

### Quick Start

```bash
# Clone the repository
git clone <repository-url>
cd mediamtx-camera-service

# Run the JavaScript client example
node examples/javascript/camera_client.js --host localhost --port 8080 --auth-type jwt --token your_jwt_token
```

## Authentication

### JWT Authentication

```javascript
const { CameraClient } = require('./examples/javascript/camera_client.js');

// Create client with JWT authentication
const client = new CameraClient({
    host: 'localhost',
    port: 8080,
    authType: 'jwt',
    authToken: 'your_jwt_token_here'
});
```

### API Key Authentication

```javascript
// Create client with API key authentication
const client = new CameraClient({
    host: 'localhost',
    port: 8080,
    authType: 'api_key',
    apiKey: 'your_api_key_here'
});
```

### SSL/TLS Support

```javascript
// Create client with SSL/TLS
const client = new CameraClient({
    host: 'localhost',
    port: 8080,
    useSsl: true,
    authType: 'jwt',
    authToken: 'your_jwt_token_here'
});
```

## Basic Usage

### Connection Management

```javascript
const { CameraClient } = require('./examples/javascript/camera_client.js');

async function main() {
    // Create client
    const client = new CameraClient({
        host: 'localhost',
        port: 8080,
        authType: 'jwt',
        authToken: 'your_token'
    });
    
    try {
        // Connect to service
        await client.connect();
        console.log('Connected successfully');
        
        // Your camera operations here
        
    } catch (error) {
        console.error(`Connection failed: ${error.message}`);
    } finally {
        // Always disconnect
        await client.disconnect();
    }
}

// Run the example
main().catch(console.error);
```

### Camera Discovery

```javascript
// Get list of available cameras
const cameras = await client.getCameraList();
console.log(`Found ${cameras.length} cameras:`);

for (const camera of cameras) {
    console.log(`  - ${camera.name} (${camera.devicePath})`);
    console.log(`    Status: ${camera.status}`);
    console.log(`    Capabilities: ${camera.capabilities.join(', ')}`);
    if (camera.streamUrl) {
        console.log(`    Stream URL: ${camera.streamUrl}`);
    }
}
```

### Camera Status Monitoring

```javascript
// Get status of specific camera
try {
    const status = await client.getCameraStatus('/dev/video0');
    console.log(`Camera status: ${status.status}`);
    console.log(`Capabilities: ${status.capabilities.join(', ')}`);
} catch (error) {
    if (error.name === 'CameraNotFoundError') {
        console.log('Camera not found');
    } else {
        console.error(`Error: ${error.message}`);
    }
}
```

### Taking Snapshots

```javascript
// Take a snapshot
try {
    const snapshot = await client.takeSnapshot('/dev/video0');
    console.log(`Snapshot saved: ${snapshot.filename}`);
    console.log(`File size: ${snapshot.file_size} bytes`);
} catch (error) {
    if (error.name === 'CameraNotFoundError') {
        console.log('Camera not found');
    } else if (error.name === 'MediaMTXError') {
        console.log(`Snapshot failed: ${error.message}`);
    } else {
        console.error(`Error: ${error.message}`);
    }
}
```

### Recording Operations

```javascript
// Start recording
try {
    const recording = await client.startRecording(
        '/dev/video0',
        60,  // 60 seconds
        'my_recording.mp4'
    );
    console.log(`Recording started: ${recording.filename}`);
    console.log(`Recording ID: ${recording.recordingId}`);
    
    // Wait for some time
    await new Promise(resolve => setTimeout(resolve, 30000));
    
    // Stop recording
    const stopResult = await client.stopRecording('/dev/video0');
    console.log(`Recording stopped: ${stopResult.filename}`);
    console.log(`Duration: ${stopResult.duration} seconds`);
    
} catch (error) {
    if (error.name === 'CameraNotFoundError') {
        console.log('Camera not found');
    } else if (error.name === 'MediaMTXError') {
        console.log(`Recording failed: ${error.message}`);
    } else {
        console.error(`Error: ${error.message}`);
    }
}
```

## Advanced Features

### Real-time Notifications

```javascript
// Set up event handlers
client.setCameraStatusCallback((params) => {
    const camera = params.camera || {};
    console.log(`Camera ${camera.name} status changed to ${camera.status}`);
});

client.setRecordingStatusCallback((params) => {
    const recording = params.recording || {};
    console.log(`Recording ${recording.filename} status: ${recording.status}`);
});

client.setConnectionLostCallback(() => {
    console.log('Connection lost - attempting to reconnect...');
});
```

### Error Handling

```javascript
const {
    CameraServiceError,
    AuthenticationError,
    ConnectionError,
    CameraNotFoundError,
    MediaMTXError
} = require('./examples/javascript/camera_client.js');

async function robustCameraOperations() {
    try {
        await client.connect();
        
        // Get cameras
        const cameras = await client.getCameraList();
        
        if (cameras.length === 0) {
            console.log('No cameras available');
            return;
        }
        
        // Work with first camera
        const camera = cameras[0];
        
        // Take snapshot
        const snapshot = await client.takeSnapshot(camera.devicePath);
        console.log(`Snapshot: ${snapshot.filename}`);
        
        // Start recording
        const recording = await client.startRecording(
            camera.devicePath,
            30
        );
        console.log(`Recording: ${recording.filename}`);
        
        // Wait and stop
        await new Promise(resolve => setTimeout(resolve, 10000));
        await client.stopRecording(camera.devicePath);
        
    } catch (error) {
        if (error.name === 'AuthenticationError') {
            console.log('Authentication failed - check your credentials');
        } else if (error.name === 'ConnectionError') {
            console.log('Connection failed - check server status');
        } else if (error.name === 'CameraNotFoundError') {
            console.log('Camera not found - check device path');
        } else if (error.name === 'MediaMTXError') {
            console.log(`MediaMTX operation failed: ${error.message}`);
        } else if (error.name === 'CameraServiceError') {
            console.log(`Camera service error: ${error.message}`);
        } else {
            console.log(`Unexpected error: ${error.message}`);
        }
    } finally {
        await client.disconnect();
    }
}
```

### Retry Logic

```javascript
// Client with custom retry settings
const client = new CameraClient({
    host: 'localhost',
    port: 8080,
    authType: 'jwt',
    authToken: 'your_token',
    maxRetries: 5,      // More retries
    retryDelay: 2.0     // Longer delay between retries
});
```

## Configuration Options

### Client Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `host` | string | 'localhost' | Server hostname |
| `port` | number | 8080 | Server port |
| `useSsl` | boolean | false | Use SSL/TLS |
| `authType` | string | 'jwt' | Authentication type ('jwt' or 'api_key') |
| `authToken` | string | null | JWT token |
| `apiKey` | string | null | API key |
| `maxRetries` | number | 3 | Maximum connection retries |
| `retryDelay` | number | 1.0 | Delay between retries (seconds) |

### Authentication Headers

The client automatically sends appropriate headers based on authentication type:

- **JWT**: `Authorization: Bearer <token>`
- **API Key**: `X-API-Key: <key>`

## Best Practices

### 1. Always Handle Exceptions

```javascript
try {
    await client.connect();
    // Your operations
} catch (error) {
    console.error(`Service error: ${error.message}`);
} finally {
    await client.disconnect();
}
```

### 2. Use Async/Await Properly

```javascript
async function cameraOperations() {
    const client = new CameraClient(options);
    
    try {
        await client.connect();
        
        const cameras = await client.getCameraList();
        console.log(`Found ${cameras.length} cameras`);
        
        if (cameras.length > 0) {
            const camera = cameras[0];
            const snapshot = await client.takeSnapshot(camera.devicePath);
            console.log(`Snapshot: ${snapshot.filename}`);
        }
        
    } catch (error) {
        console.error(`Error: ${error.message}`);
    } finally {
        await client.disconnect();
    }
}
```

### 3. Implement Proper Logging

```javascript
// Configure logging
const logger = {
    info: (message) => console.log(`[INFO] ${message}`),
    error: (message) => console.error(`[ERROR] ${message}`),
    warning: (message) => console.warn(`[WARN] ${message}`)
};

// Use with client
client.logger = logger;
```

### 4. Handle Connection Loss

```javascript
client.setConnectionLostCallback(async () => {
    console.log('Connection lost - attempting to reconnect...');
    try {
        await client.connect();
        console.log('Reconnected successfully');
    } catch (error) {
        console.error(`Reconnection failed: ${error.message}`);
    }
});
```

### 5. Use Timeouts

```javascript
// The client has built-in request timeouts (30 seconds)
// For long operations, consider implementing custom timeouts
async function longOperation() {
    try {
        const result = await Promise.race([
            client.startRecording('/dev/video0', 300),
            new Promise((_, reject) => 
                setTimeout(() => reject(new Error('Timeout')), 60000)
            )
        ]);
        console.log('Operation completed');
    } catch (error) {
        if (error.message === 'Timeout') {
            console.log('Operation timed out');
        } else {
            console.error(`Error: ${error.message}`);
        }
    }
}
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
   - Use `useSsl: false` for testing

### Debug Mode

```javascript
// Enable debug logging
const debugClient = new CameraClient({
    ...options,
    logger: {
        info: (msg) => console.log(`[DEBUG] ${msg}`),
        error: (msg) => console.error(`[DEBUG] ${msg}`),
        warning: (msg) => console.warn(`[DEBUG] ${msg}`)
    }
});
```

## Performance Considerations

### Connection Pooling

For high-throughput applications, consider implementing connection pooling:

```javascript
class CameraClientPool {
    constructor(poolSize = 5) {
        this.poolSize = poolSize;
        this.clients = [];
        this.available = [];
    }
    
    async initialize() {
        for (let i = 0; i < this.poolSize; i++) {
            const client = new CameraClient(options);
            await client.connect();
            this.available.push(client);
        }
    }
    
    async getClient() {
        if (this.available.length === 0) {
            throw new Error('No available clients');
        }
        return this.available.pop();
    }
    
    async returnClient(client) {
        this.available.push(client);
    }
}
```

### Batch Operations

For multiple cameras, consider parallel operations:

```javascript
async function processAllCameras() {
    const cameras = await client.getCameraList();
    
    // Parallel snapshot operations
    const tasks = cameras.map(camera => 
        client.takeSnapshot(camera.devicePath)
            .catch(error => ({ error: error.message, camera: camera.name }))
    );
    
    const results = await Promise.all(tasks);
    
    for (const result of results) {
        if (result.error) {
            console.log(`Failed to snapshot ${result.camera}: ${result.error}`);
        } else {
            console.log(`Snapshot ${result.camera}: ${result.filename}`);
        }
    }
}
```

## Security Considerations

### Token Management

```javascript
// Store tokens securely
const JWT_TOKEN = process.env.CAMERA_SERVICE_JWT_TOKEN;
const API_KEY = process.env.CAMERA_SERVICE_API_KEY;

// Rotate tokens regularly
function isTokenExpired(token) {
    // Implement token expiration check
    return false;
}
```

### SSL/TLS Best Practices

```javascript
// Create secure SSL context
const https = require('https');

const sslOptions = {
    rejectUnauthorized: true,
    checkServerIdentity: () => undefined
};

// Use with client
const client = new CameraClient({
    host: 'secure.example.com',
    port: 443,
    useSsl: true,
    // SSL options will be used internally
});
```

## Examples

### Complete Working Example

```javascript
#!/usr/bin/env node
/**
 * Complete camera service client example.
 */

const { CameraClient } = require('./examples/javascript/camera_client.js');

async function cameraMonitor() {
    // Setup logging
    const logger = {
        info: (msg) => console.log(`[INFO] ${msg}`),
        error: (msg) => console.error(`[ERROR] ${msg}`),
        warning: (msg) => console.warn(`[WARN] ${msg}`)
    };
    
    // Create client
    const client = new CameraClient({
        host: 'localhost',
        port: 8080,
        authType: 'jwt',
        authToken: 'your_token_here',
        logger: logger
    });
    
    // Set up event handlers
    client.setCameraStatusCallback((params) => {
        const camera = params.camera || {};
        console.log(`📹 Camera ${camera.name}: ${camera.status}`);
    });
    
    client.setRecordingStatusCallback((params) => {
        const recording = params.recording || {};
        console.log(`🎥 Recording ${recording.filename}: ${recording.status}`);
    });
    
    try {
        // Connect
        await client.connect();
        console.log('✅ Connected to camera service');
        
        // Get cameras
        const cameras = await client.getCameraList();
        console.log(`📹 Found ${cameras.length} cameras`);
        
        if (cameras.length > 0) {
            const camera = cameras[0];
            console.log(`🎯 Working with camera: ${camera.name}`);
            
            // Take snapshot
            const snapshot = await client.takeSnapshot(camera.devicePath);
            console.log(`📸 Snapshot: ${snapshot.filename}`);
            
            // Start recording
            const recording = await client.startRecording(
                camera.devicePath,
                30,
                'demo_recording.mp4'
            );
            console.log(`🎥 Recording started: ${recording.filename}`);
            
            // Monitor for 20 seconds
            await new Promise(resolve => setTimeout(resolve, 20000));
            
            // Stop recording
            const stopResult = await client.stopRecording(camera.devicePath);
            console.log(`⏹️ Recording stopped: ${stopResult.filename}`);
        }
        
        // Keep monitoring
        console.log('👀 Monitoring cameras (press Ctrl+C to stop)...');
        await new Promise(resolve => setTimeout(resolve, 60000));
        
    } catch (error) {
        console.error(`❌ Error: ${error.message}`);
    } finally {
        await client.disconnect();
        console.log('✅ Disconnected');
    }
}

// Handle graceful shutdown
process.on('SIGINT', () => {
    console.log('\n🛑 Stopping monitor...');
    process.exit(0);
});

cameraMonitor().catch(console.error);
```

## API Reference

### CameraClient Class

#### Constructor

```javascript
new CameraClient(options)
```

#### Methods

- `connect()` - Connect to camera service
- `disconnect()` - Disconnect from camera service
- `ping()` - Test connection
- `getCameraList()` - Get available cameras
- `getCameraStatus(devicePath)` - Get camera status
- `takeSnapshot(devicePath, customFilename)` - Take snapshot
- `startRecording(devicePath, duration, customFilename)` - Start recording
- `stopRecording(devicePath)` - Stop recording

#### Event Handlers

- `setCameraStatusCallback(callback)` - Set camera status update handler
- `setRecordingStatusCallback(callback)` - Set recording status update handler
- `setConnectionLostCallback(callback)` - Set connection lost handler

### Classes

#### CameraInfo

```javascript
class CameraInfo {
    constructor(devicePath, name, capabilities, status, streamUrl = null) {
        this.devicePath = devicePath;
        this.name = name;
        this.capabilities = capabilities;
        this.status = status;
        this.streamUrl = streamUrl;
    }
}
```

#### RecordingInfo

```javascript
class RecordingInfo {
    constructor(devicePath, recordingId, filename, startTime, duration = null, status = 'active') {
        this.devicePath = devicePath;
        this.recordingId = recordingId;
        this.filename = filename;
        this.startTime = startTime;
        this.duration = duration;
        this.status = status;
    }
}
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
**Compatibility:** Node.js 14+, ws, uuid 