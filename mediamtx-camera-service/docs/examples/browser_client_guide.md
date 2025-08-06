# Browser Client Guide - MediaMTX Camera Service

## Overview

The browser-based client example demonstrates how to connect to the MediaMTX Camera Service using WebSocket JSON-RPC 2.0 protocol with JWT authentication and a modern, responsive user interface.

## Features

- **JWT and API Key Authentication**: Secure authentication support
- **Modern Responsive UI**: Beautiful, mobile-friendly interface
- **Real-time Camera Status**: Live updates and notifications
- **Camera Control Operations**: Snapshot, recording, and streaming
- **WebSocket Connection Management**: Automatic reconnection and error handling
- **Activity Logging**: Comprehensive operation tracking
- **Cross-browser Compatibility**: Works on all modern browsers

## Quick Start

### Prerequisites

- Modern web browser (Chrome, Firefox, Safari, Edge)
- MediaMTX Camera Service running on a server
- JWT token or API key for authentication

### Usage

1. **Open the HTML file**:
   ```bash
   # Open in browser
   open examples/browser/camera_client.html
   ```

2. **Configure connection**:
   - Enter server hostname (e.g., `localhost`)
   - Enter server port (e.g., `8080`)
   - Select authentication type (JWT or API Key)
   - Enter your authentication token

3. **Connect and use**:
   - Click "Connect" to establish WebSocket connection
   - View available cameras in the grid
   - Use camera control buttons for operations

## Authentication

### JWT Authentication

The browser client supports JWT token authentication:

```javascript
// JWT token is sent in the first WebSocket message
{
    "jsonrpc": "2.0",
    "method": "authenticate",
    "id": 1,
    "params": {
        "token": "your_jwt_token_here"
    }
}
```

### API Key Authentication

For API key authentication:

```javascript
// API key is sent in the first WebSocket message
{
    "jsonrpc": "2.0",
    "method": "authenticate",
    "id": 1,
    "params": {
        "api_key": "your_api_key_here"
    }
}
```

## User Interface

### Connection Panel

The connection panel allows you to configure the connection to the camera service:

- **Host**: Server hostname (default: localhost)
- **Port**: Server port (default: 8080)
- **Authentication Type**: JWT Token or API Key
- **Authentication Token**: Your JWT token or API key

### Camera Grid

The camera grid displays all available cameras with their status and capabilities:

- **Camera Name**: Human-readable camera name
- **Status**: Current camera status (available, busy, error)
- **Device Path**: System device path
- **Capabilities**: Available camera features
- **Stream URL**: Direct stream access (if available)

### Camera Actions

Each camera card provides action buttons:

- **📸 Snapshot**: Take a still image
- **🎥 Record**: Start video recording
- **⏹️ Stop**: Stop current recording
- **📺 Stream**: Open live stream in new tab

### Activity Log

The activity log shows real-time operation feedback:

- **Connection events**: Connect/disconnect messages
- **Camera operations**: Snapshot and recording results
- **Error messages**: Failed operations and reasons
- **Status updates**: Camera and recording status changes

## Browser Compatibility

### Supported Browsers

- **Chrome**: 80+ (recommended)
- **Firefox**: 75+
- **Safari**: 13+
- **Edge**: 80+

### WebSocket Support

All modern browsers support WebSocket connections. The client automatically detects the protocol:

- **HTTP**: Uses `ws://` protocol
- **HTTPS**: Uses `wss://` protocol

## Security Considerations

### JWT Token Security

```javascript
// Store tokens securely (not in localStorage for production)
const JWT_TOKEN = 'your_jwt_token';

// Consider using secure storage methods
// - HTTP-only cookies
// - Session storage (for session duration)
// - Secure token management systems
```

### HTTPS Requirements

For production use, always use HTTPS:

```javascript
// Automatic protocol detection
const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
const wsUrl = `${protocol}//${host}:${port}/ws`;
```

### CORS Considerations

Ensure your camera service allows WebSocket connections from your domain:

```javascript
// Server should allow WebSocket upgrade requests
// Check CORS headers for WebSocket connections
```

## Advanced Usage

### Custom Camera Operations

```javascript
// Access the client instance
const client = window.client;

// Custom snapshot with filename
await client.sendRequest('take_snapshot', {
    device_path: '/dev/video0',
    custom_filename: 'my_snapshot.jpg'
});

// Custom recording with duration
await client.sendRequest('start_recording', {
    device_path: '/dev/video0',
    duration: 60,
    custom_filename: 'my_recording.mp4'
});
```

### Real-time Event Handling

```javascript
// Listen for camera status updates
client.handleCameraStatusUpdate = (params) => {
    const camera = params.camera;
    console.log(`Camera ${camera.name} status: ${camera.status}`);
};

// Listen for recording status updates
client.handleRecordingStatusUpdate = (params) => {
    const recording = params.recording;
    console.log(`Recording ${recording.filename} status: ${recording.status}`);
};
```

### Error Handling

```javascript
// Custom error handling
client.sendRequest('get_camera_list').catch(error => {
    if (error.message.includes('authentication')) {
        // Handle authentication errors
        console.error('Authentication failed');
    } else if (error.message.includes('connection')) {
        // Handle connection errors
        console.error('Connection lost');
    } else {
        // Handle other errors
        console.error('Operation failed:', error.message);
    }
});
```

## Troubleshooting

### Common Issues

1. **Connection Refused**
   - Check if camera service is running
   - Verify host and port settings
   - Check firewall settings
   - Ensure WebSocket endpoint is accessible

2. **Authentication Failed**
   - Verify JWT token or API key
   - Check token expiration
   - Ensure correct authentication type
   - Check server authentication configuration

3. **No Cameras Available**
   - Check camera service logs
   - Verify camera devices are connected
   - Check camera permissions
   - Ensure camera discovery is working

4. **WebSocket Connection Issues**
   - Check browser console for errors
   - Verify WebSocket support in browser
   - Check network connectivity
   - Try different browser

### Debug Mode

Enable browser developer tools for debugging:

```javascript
// Open browser console (F12)
// Check for WebSocket connection logs
// Monitor network tab for WebSocket frames
// Check for JavaScript errors
```

## Performance Optimization

### Connection Management

```javascript
// Automatic reconnection
client.websocket.onclose = () => {
    console.log('Connection lost, attempting to reconnect...');
    setTimeout(() => client.connect(), 5000);
};
```

### Memory Management

```javascript
// Clean up event listeners
window.addEventListener('beforeunload', () => {
    if (client.websocket) {
        client.websocket.close();
    }
});
```

### UI Performance

```javascript
// Debounce camera updates
let updateTimeout;
client.handleCameraStatusUpdate = (params) => {
    clearTimeout(updateTimeout);
    updateTimeout = setTimeout(() => {
        // Update UI
        client.renderCameraGrid();
    }, 100);
};
```

## Best Practices

### 1. Secure Token Management

```javascript
// Don't store tokens in localStorage for production
// Use secure session management
// Implement token refresh mechanisms
```

### 2. Error Recovery

```javascript
// Implement automatic reconnection
// Show user-friendly error messages
// Provide retry mechanisms for failed operations
```

### 3. User Experience

```javascript
// Show loading states during operations
// Provide clear feedback for all actions
// Handle edge cases gracefully
```

### 4. Accessibility

```javascript
// Add ARIA labels to buttons
// Ensure keyboard navigation
// Provide screen reader support
```

## API Reference

### CameraServiceClient Class

#### Constructor

```javascript
new CameraServiceClient()
```

#### Properties

- `websocket`: WebSocket connection object
- `connected`: Connection status boolean
- `cameras`: Map of camera devices
- `recordings`: Map of active recordings

#### Methods

- `connect()`: Connect to camera service
- `disconnect()`: Disconnect from camera service
- `sendRequest(method, params)`: Send JSON-RPC request
- `loadCameras()`: Load camera list from server
- `renderCameraGrid()`: Update camera display
- `takeSnapshot(devicePath)`: Take camera snapshot
- `startRecording(devicePath)`: Start video recording
- `stopRecording(devicePath)`: Stop video recording

#### Event Handlers

- `handleCameraStatusUpdate(params)`: Handle camera status updates
- `handleRecordingStatusUpdate(params)`: Handle recording status updates
- `onWebSocketOpen()`: Handle WebSocket connection open
- `onWebSocketClose()`: Handle WebSocket connection close
- `onWebSocketError(error)`: Handle WebSocket errors

### UI Components

#### Connection Panel

- Host input field
- Port input field
- Authentication type selector
- Authentication token input
- Connect/Disconnect buttons

#### Camera Grid

- Camera cards with status
- Action buttons for each camera
- Real-time status indicators
- Recording indicators

#### Activity Log

- Timestamped log entries
- Color-coded message types
- Auto-scrolling log container
- Message filtering

## Examples

### Complete Integration Example

```html
<!DOCTYPE html>
<html>
<head>
    <title>Custom Camera Client</title>
</head>
<body>
    <div id="cameraContainer"></div>
    
    <script>
        // Custom camera client implementation
        class CustomCameraClient extends CameraServiceClient {
            constructor() {
                super();
                this.setupCustomHandlers();
            }
            
            setupCustomHandlers() {
                // Custom camera status handler
                this.handleCameraStatusUpdate = (params) => {
                    const camera = params.camera;
                    this.updateCustomUI(camera);
                };
                
                // Custom recording handler
                this.handleRecordingStatusUpdate = (params) => {
                    const recording = params.recording;
                    this.updateRecordingUI(recording);
                };
            }
            
            updateCustomUI(camera) {
                // Custom UI updates
                console.log(`Camera ${camera.name} updated`);
            }
            
            updateRecordingUI(recording) {
                // Custom recording UI updates
                console.log(`Recording ${recording.filename} updated`);
            }
        }
        
        // Initialize custom client
        const customClient = new CustomCameraClient();
    </script>
</body>
</html>
```

### Minimal Integration

```html
<!DOCTYPE html>
<html>
<head>
    <title>Minimal Camera Client</title>
</head>
<body>
    <button onclick="takeSnapshot()">Take Snapshot</button>
    <button onclick="startRecording()">Start Recording</button>
    
    <script>
        // Simple camera operations
        async function takeSnapshot() {
            try {
                const result = await client.sendRequest('take_snapshot', {
                    device_path: '/dev/video0'
                });
                alert(`Snapshot saved: ${result.filename}`);
            } catch (error) {
                alert(`Snapshot failed: ${error.message}`);
            }
        }
        
        async function startRecording() {
            try {
                const result = await client.sendRequest('start_recording', {
                    device_path: '/dev/video0',
                    duration: 30
                });
                alert(`Recording started: ${result.recording.filename}`);
            } catch (error) {
                alert(`Recording failed: ${error.message}`);
            }
        }
    </script>
</body>
</html>
```

## Support

For issues and questions:

1. Check browser console for JavaScript errors
2. Verify WebSocket connection in Network tab
3. Check camera service logs
4. Ensure authentication credentials are correct
5. Test with different browsers

### Browser Console Commands

```javascript
// Check connection status
console.log(client.connected);

// List available cameras
console.log(Array.from(client.cameras.values()));

// Test ping
client.sendRequest('ping').then(console.log);

// Get camera list
client.sendRequest('get_camera_list').then(console.log);
```

---

**Version:** 1.0  
**Last Updated:** 2025-08-06  
**Compatibility:** Modern browsers with WebSocket support 