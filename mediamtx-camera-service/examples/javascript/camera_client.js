#!/usr/bin/env node
/**
 * MediaMTX Camera Service JavaScript/Node.js Client Example
 * 
 * This example demonstrates how to connect to the MediaMTX Camera Service
 * using WebSocket JSON-RPC 2.0 protocol with authentication support.
 * 
 * IMPORTANT: Port Configuration
 * - Production/Default: Port 8002 (config/default.yaml)
 * - Development: Port 8080 (config/development.yaml)
 * - Use --port argument to specify the correct port for your environment
 * 
 * Features:
 * - JWT and API Key authentication
 * - WebSocket connection management
 * - Camera discovery and control
 * - Snapshot and recording operations
 * - Real-time status notifications
 * - Comprehensive error handling
 * - Retry logic and connection recovery
 * 
 * Usage:
 *   node camera_client.js --host localhost --port 8002 --auth-type jwt --token your_jwt_token
 *   node camera_client.js --host localhost --port 8002 --auth-type api_key --key your_api_key
 */

const WebSocket = require('ws');
const { URL } = require('url');
const { v4: uuidv4 } = require('uuid');

/**
 * Camera device information
 */
class CameraInfo {
    constructor(devicePath, name, capabilities, status, streamUrl = null) {
        this.devicePath = devicePath;
        this.name = name;
        this.capabilities = capabilities;
        this.status = status;
        this.streamUrl = streamUrl;
    }
}

/**
 * Recording session information
 */
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

/**
 * Base exception for camera service errors
 */
class CameraServiceError extends Error {
    constructor(message) {
        super(message);
        this.name = 'CameraServiceError';
    }
}

/**
 * Authentication failed exception
 */
class AuthenticationError extends CameraServiceError {
    constructor(message) {
        super(message);
        this.name = 'AuthenticationError';
    }
}

/**
 * Connection failed exception
 */
class ConnectionError extends CameraServiceError {
    constructor(message) {
        super(message);
        this.name = 'ConnectionError';
    }
}

/**
 * Camera device not found exception
 */
class CameraNotFoundError extends CameraServiceError {
    constructor(message) {
        super(message);
        this.name = 'CameraNotFoundError';
    }
}

/**
 * MediaMTX operation failed exception
 */
class MediaMTXError extends CameraServiceError {
    constructor(message) {
        super(message);
        this.name = 'MediaMTXError';
    }
}

/**
 * JavaScript/Node.js client for MediaMTX Camera Service
 * 
 * Provides a high-level interface for camera control and monitoring
 * with support for JWT and API key authentication.
 */
class CameraClient {
    /**
     * Initialize the camera client
     * 
     * @param {Object} options - Client configuration options
     * @param {string} options.host - Server hostname (default: 'localhost')
     * @param {number} options.port - Server port (default: 8080)
     * @param {boolean} options.useSsl - Whether to use SSL/TLS (default: false)
     * @param {string} options.authType - Authentication type ('jwt' or 'api_key') (default: 'jwt')
     * @param {string} options.authToken - JWT token for authentication
     * @param {string} options.apiKey - API key for authentication
     * @param {number} options.maxRetries - Maximum number of connection retries (default: 3)
     * @param {number} options.retryDelay - Delay between retries in seconds (default: 1.0)
     */
    constructor(options = {}) {
        this.host = options.host || 'localhost';
        this.port = options.port || 8002;  // Changed from 8080 to 8002 (production default)
        this.useSsl = options.useSsl || false;
        this.authType = options.authType || 'jwt';
        this.authToken = options.authToken;
        this.apiKey = options.apiKey;
        this.maxRetries = options.maxRetries || 3;
        this.retryDelay = options.retryDelay || 1.0;
        
        // Connection state
        this.websocket = null;
        this.connected = false;
        this.authenticated = false;
        this.clientId = uuidv4();
        
        // Request tracking
        this.requestId = 0;
        this.pendingRequests = new Map();
        
        // Event handlers
        this.onCameraStatusUpdate = null;
        this.onRecordingStatusUpdate = null;
        this.onConnectionLost = null;
        
        // Setup logging
        this.logger = console;
    }

    /**
     * Get WebSocket URL
     * 
     * @returns {string} WebSocket URL
     */
    _getWsUrl() {
        const protocol = this.useSsl ? 'wss' : 'ws';
        return `${protocol}://${this.host}:${this.port}/ws`;
    }

    /**
     * Get authentication headers
     * 
     * @returns {Object} Authentication headers
     */
    _getAuthHeaders() {
        const headers = {};
        
        if (this.authType === 'jwt' && this.authToken) {
            headers['Authorization'] = `Bearer ${this.authToken}`;
        } else if (this.authType === 'api_key' && this.apiKey) {
            headers['X-API-Key'] = this.apiKey;
        }
        
        return headers;
    }

    /**
     * Connect to the camera service
     * 
     * @returns {Promise<void>}
     * @throws {ConnectionError} If connection fails
     * @throws {AuthenticationError} If authentication fails
     */
    async connect() {
        for (let attempt = 1; attempt <= this.maxRetries; attempt++) {
            try {
                this.logger.info(`Connecting to ${this._getWsUrl()} (attempt ${attempt})`);
                
                // Create WebSocket connection
                const wsUrl = this._getWsUrl();
                const headers = this._getAuthHeaders();
                
                this.websocket = new WebSocket(wsUrl, {
                    headers: headers
                });
                
                // Set up event handlers
                this._setupWebSocketHandlers();
                
                // Wait for connection
                await this._waitForConnection();
                
                this.connected = true;
                this.logger.info('Connected to camera service');
                
                // Authenticate if token provided
                if (this.authToken || this.apiKey) {
                    await this._authenticate();
                }
                
                // Test connection with ping
                await this.ping();
                this.logger.info('Connection test successful');
                
                return;
                
            } catch (error) {
                this.logger.error(`Connection attempt ${attempt} failed: ${error.message}`);
                if (attempt < this.maxRetries) {
                    await this._sleep(this.retryDelay * attempt);
                } else {
                    throw new ConnectionError(`Failed to connect after ${this.maxRetries} attempts: ${error.message}`);
                }
            }
        }
    }

    /**
     * Set up WebSocket event handlers
     */
    _setupWebSocketHandlers() {
        this.websocket.on('open', () => {
            this.logger.info('WebSocket connection opened');
        });

        this.websocket.on('message', (data) => {
            this._processMessage(data.toString());
        });

        this.websocket.on('close', (code, reason) => {
            this.logger.warn(`WebSocket connection closed: ${code} - ${reason}`);
            this.connected = false;
            if (this.onConnectionLost) {
                this.onConnectionLost();
            }
        });

        this.websocket.on('error', (error) => {
            this.logger.error(`WebSocket error: ${error.message}`);
            this.connected = false;
        });
    }

    /**
     * Wait for WebSocket connection to be established
     * 
     * @returns {Promise<void>}
     */
    _waitForConnection() {
        return new Promise((resolve, reject) => {
            const timeout = setTimeout(() => {
                reject(new Error('Connection timeout'));
            }, 10000);

            this.websocket.once('open', () => {
                clearTimeout(timeout);
                resolve();
            });

            this.websocket.once('error', (error) => {
                clearTimeout(timeout);
                reject(error);
            });
        });
    }

    /**
     * Authenticate with the camera service using JWT or API key
     * 
     * @returns {Promise<void>}
     * @throws {AuthenticationError} If authentication fails
     */
    async _authenticate() {
        if (!this.connected) {
            throw new ConnectionError('Not connected to camera service');
        }

        // Determine token to use
        let token = null;
        let authType = 'auto';

        if (this.authType === 'jwt' && this.authToken) {
            token = this.authToken;
            authType = 'jwt';
        } else if (this.authType === 'api_key' && this.apiKey) {
            token = this.apiKey;
            authType = 'api_key';
        } else {
            throw new AuthenticationError('No authentication token provided');
        }

        try {
            // Send authentication request
            const response = await this._sendRequest('authenticate', {
                token: token,
                auth_type: authType
            });

            if (response.authenticated) {
                this.authenticated = true;
                this.logger.info(`Authenticated successfully with role: ${response.role || 'unknown'}`);
            } else {
                const errorMsg = response.error || 'Authentication failed';
                throw new AuthenticationError(`Authentication failed: ${errorMsg}`);
            }
        } catch (error) {
            if (error instanceof AuthenticationError) {
                throw error;
            }
            throw new AuthenticationError(`Authentication error: ${error.message}`);
        }
    }

    /**
     * Sleep for specified milliseconds
     * 
     * @param {number} ms - Milliseconds to sleep
     * @returns {Promise<void>}
     */
    _sleep(ms) {
        return new Promise(resolve => setTimeout(resolve, ms * 1000));
    }

    /**
     * Disconnect from the camera service
     */
    async disconnect() {
        if (this.websocket) {
            this.websocket.close();
            this.websocket = null;
        }
        this.connected = false;
        this.authenticated = false;
        this.logger.info('Disconnected from camera service');
    }

    /**
     * Process incoming WebSocket message
     * 
     * @param {string} message - Raw message string
     */
    _processMessage(message) {
        try {
            const data = JSON.parse(message);
            
            // Handle JSON-RPC response
            if (data.id !== undefined && data.result !== undefined) {
                this._handleResponse(data);
            }
            // Handle JSON-RPC notification
            else if (data.method !== undefined && data.id === undefined) {
                this._handleNotification(data);
            }
            else {
                this.logger.warn(`Unknown message format: ${JSON.stringify(data)}`);
            }
            
        } catch (error) {
            this.logger.error(`Invalid JSON message: ${error.message}`);
        }
    }

    /**
     * Handle JSON-RPC response
     * 
     * @param {Object} response - JSON-RPC response object
     */
    _handleResponse(response) {
        const requestId = response.id;
        if (this.pendingRequests.has(requestId)) {
            const { resolve, reject } = this.pendingRequests.get(requestId);
            this.pendingRequests.delete(requestId);
            
            if (response.error) {
                reject(new CameraServiceError(response.error.message || 'Unknown error'));
            } else {
                resolve(response.result);
            }
        }
    }

    /**
     * Handle JSON-RPC notification
     * 
     * @param {Object} notification - JSON-RPC notification object
     */
    _handleNotification(notification) {
        const method = notification.method;
        const params = notification.params || {};
        
        if (method === 'camera_status_update' && this.onCameraStatusUpdate) {
            this.onCameraStatusUpdate(params);
        } else if (method === 'recording_status_update' && this.onRecordingStatusUpdate) {
            this.onRecordingStatusUpdate(params);
        } else {
            this.logger.info(`Received notification: ${method}`);
        }
    }

    /**
     * Send JSON-RPC request and wait for response
     * 
     * @param {string} method - RPC method name
     * @param {Object} params - Method parameters
     * @returns {Promise<any>} Response result
     * @throws {CameraServiceError} If request fails
     */
    async _sendRequest(method, params = {}) {
        if (!this.connected) {
            throw new ConnectionError('Not connected to camera service');
        }
        
        this.requestId++;
        const requestId = this.requestId;
        
        const request = {
            jsonrpc: '2.0',
            method: method,
            id: requestId,
            params: params
        };
        
        return new Promise((resolve, reject) => {
            // Store promise for response
            this.pendingRequests.set(requestId, { resolve, reject });
            
            // Send request
            this.websocket.send(JSON.stringify(request));
            
            // Set timeout
            setTimeout(() => {
                if (this.pendingRequests.has(requestId)) {
                    this.pendingRequests.delete(requestId);
                    reject(new CameraServiceError(`Request timeout: ${method}`));
                }
            }, 30000);
        });
    }

    /**
     * Send ping request to test connection
     * 
     * @returns {Promise<string>} Pong response
     */
    async ping() {
        return await this._sendRequest('ping');
    }

    /**
     * Get list of available cameras
     * 
     * @returns {Promise<CameraInfo[]>} List of camera information
     */
    async getCameraList() {
        const result = await this._sendRequest('get_camera_list');
        
        const cameras = [];
        for (const cameraData of result.cameras || []) {
            const camera = new CameraInfo(
                cameraData.device,
                cameraData.name,
                cameraData.capabilities || [],
                cameraData.status,
                cameraData.stream_url
            );
            cameras.push(camera);
        }
        
        return cameras;
    }

    /**
     * Get status of specific camera
     * 
     * @param {string} devicePath - Camera device path
     * @returns {Promise<CameraInfo>} Camera information
     * @throws {CameraNotFoundError} If camera not found
     */
    async getCameraStatus(devicePath) {
        const result = await this._sendRequest('get_camera_status', { device_path: devicePath });
        
        if (!result.found) {
            throw new CameraNotFoundError(`Camera not found: ${devicePath}`);
        }
        
        const cameraData = result.camera;
        return new CameraInfo(
            cameraData.device_path,
            cameraData.name,
            cameraData.capabilities || [],
            cameraData.status,
            cameraData.stream_url
        );
    }

    /**
     * Take a snapshot from camera
     * 
     * @param {string} devicePath - Camera device path
     * @param {string} customFilename - Optional custom filename
     * @returns {Promise<Object>} Snapshot information
     * @throws {CameraNotFoundError} If camera not found
     * @throws {MediaMTXError} If snapshot fails
     */
    async takeSnapshot(devicePath, customFilename = null) {
        const params = { device_path: devicePath };
        if (customFilename) {
            params.custom_filename = customFilename;
        }
        
        const result = await this._sendRequest('take_snapshot', params);
        
        if (!result.success) {
            const error = result.error || 'Unknown error';
            if (error.toLowerCase().includes('not found')) {
                throw new CameraNotFoundError(`Camera not found: ${devicePath}`);
            } else {
                throw new MediaMTXError(`Snapshot failed: ${error}`);
            }
        }
        
        return result;
    }

    /**
     * Start recording from camera
     * 
     * @param {string} devicePath - Camera device path
     * @param {number} duration - Recording duration in seconds (optional)
     * @param {string} customFilename - Optional custom filename
     * @returns {Promise<RecordingInfo>} Recording information
     * @throws {CameraNotFoundError} If camera not found
     * @throws {MediaMTXError} If recording fails
     */
    async startRecording(devicePath, duration = null, customFilename = null) {
        const params = { device_path: devicePath };
        if (duration) {
            params.duration = duration;
        }
        if (customFilename) {
            params.custom_filename = customFilename;
        }
        
        const result = await this._sendRequest('start_recording', params);
        
        if (!result.success) {
            const error = result.error || 'Unknown error';
            if (error.toLowerCase().includes('not found')) {
                throw new CameraNotFoundError(`Camera not found: ${devicePath}`);
            } else {
                throw new MediaMTXError(`Recording failed: ${error}`);
            }
        }
        
        const recordingData = result.recording;
        return new RecordingInfo(
            recordingData.device_path,
            recordingData.recording_id,
            recordingData.filename,
            recordingData.start_time,
            recordingData.duration,
            recordingData.status
        );
    }

    /**
     * Stop recording from camera
     * 
     * @param {string} devicePath - Camera device path
     * @returns {Promise<Object>} Recording stop information
     * @throws {CameraNotFoundError} If camera not found
     * @throws {MediaMTXError} If stop recording fails
     */
    async stopRecording(devicePath) {
        const result = await this._sendRequest('stop_recording', { device_path: devicePath });
        
        if (!result.success) {
            const error = result.error || 'Unknown error';
            if (error.toLowerCase().includes('not found')) {
                throw new CameraNotFoundError(`Camera not found: ${devicePath}`);
            } else {
                throw new MediaMTXError(`Stop recording failed: ${error}`);
            }
        }
        
        return result;
    }

    /**
     * Set callback for camera status updates
     * 
     * @param {Function} callback - Callback function
     */
    setCameraStatusCallback(callback) {
        this.onCameraStatusUpdate = callback;
    }

    /**
     * Set callback for recording status updates
     * 
     * @param {Function} callback - Callback function
     */
    setRecordingStatusCallback(callback) {
        this.onRecordingStatusUpdate = callback;
    }

    /**
     * Set callback for connection lost events
     * 
     * @param {Function} callback - Callback function
     */
    setConnectionLostCallback(callback) {
        this.onConnectionLost = callback;
    }
}

/**
 * Example usage of the camera client
 */
async function main() {
    // Parse command line arguments
    const args = process.argv.slice(2);
    const options = {};
    
    // Check for help argument first
    if (args.includes('--help') || args.includes('-h')) {
        console.log(`
MediaMTX Camera Service JavaScript Client

Usage:
  node camera_client.js --host localhost --port 8002 --auth-type jwt --token your_jwt_token
  node camera_client.js --host localhost --port 8002 --auth-type api_key --key your_api_key

Options:
  --host HOST           Server hostname (default: localhost)
  --port PORT           Server port (default: 8002)
  --ssl                 Use SSL/TLS
  --auth-type TYPE      Authentication type (jwt or api_key)
  --token TOKEN         JWT token
  --key KEY             API key
  --help, -h            Show this help message
        `);
        return;
    }
    
    for (let i = 0; i < args.length; i += 2) {
        const key = args[i];
        const value = args[i + 1];
        
        switch (key) {
            case '--host':
                options.host = value;
                break;
            case '--port':
                options.port = parseInt(value);
                break;
            case '--ssl':
                options.useSsl = true;
                i--; // No value for this flag
                break;
            case '--auth-type':
                options.authType = value;
                break;
            case '--token':
                options.authToken = value;
                break;
            case '--key':
                options.apiKey = value;
                break;
        }
    }
    
    // Create client
    const client = new CameraClient(options);
    
    try {
        // Connect to service
        await client.connect();
        console.log('✅ Connected to camera service');
        
        // Test ping
        const pong = await client.ping();
        console.log(`✅ Ping response: ${pong}`);
        
        // Get camera list
        const cameras = await client.getCameraList();
        console.log(`✅ Found ${cameras.length} cameras:`);
        for (const camera of cameras) {
            console.log(`  - ${camera.name} (${camera.devicePath}) - ${camera.status}`);
        }
        
        if (cameras.length > 0) {
            // Get status of first camera
            const camera = cameras[0];
            const status = await client.getCameraStatus(camera.devicePath);
            console.log(`✅ Camera status: ${status.status}`);
            
            // Take snapshot
            const snapshot = await client.takeSnapshot(camera.devicePath);
            console.log(`✅ Snapshot taken: ${snapshot.filename}`);
            
            // Start recording
            const recording = await client.startRecording(camera.devicePath, 10);
            console.log(`✅ Recording started: ${recording.filename}`);
            
            // Wait a bit
            await new Promise(resolve => setTimeout(resolve, 5000));
            
            // Stop recording
            const stopResult = await client.stopRecording(camera.devicePath);
            console.log(`✅ Recording stopped: ${stopResult.filename}`);
        }
        
    } catch (error) {
        console.error(`❌ Error: ${error.message}`);
    } finally {
        await client.disconnect();
        console.log('✅ Disconnected');
    }
}

// Export for use as module
module.exports = {
    CameraClient,
    CameraInfo,
    RecordingInfo,
    CameraServiceError,
    AuthenticationError,
    ConnectionError,
    CameraNotFoundError,
    MediaMTXError
};

// Run if called directly
if (require.main === module) {
    main().catch(console.error);
} 