/**
 * MediaMTX Camera Service JavaScript SDK Client.
 */

import WebSocket from 'ws';
import { v4 as uuidv4 } from 'uuid';
import {
    ClientConfig,
    CameraInfo,
    RecordingInfo,
    SnapshotInfo,
    JsonRpcRequest,
    JsonRpcResponse,
    JsonRpcNotification,
    JsonRpcMessage
} from './types';
import {
    CameraServiceError,
    AuthenticationError,
    ConnectionError,
    CameraNotFoundError,
    MediaMTXError,
    TimeoutError
} from './exceptions';

/**
 * JavaScript client for MediaMTX Camera Service.
 */
export class CameraClient {
    private host: string;
    private port: number;
    private useSsl: boolean;
    private authType: string;
    private authToken?: string;
    private apiKey?: string;
    private maxRetries: number;
    private retryDelay: number;
    
    private websocket?: WebSocket;
    private connected: boolean = false;
    private authenticated: boolean = false;
    private clientId: string;
    private requestId: number = 0;
    private pendingRequests: Map<number, { resolve: (value: any) => void; reject: (reason: any) => void }> = new Map();
    
    // Event handlers
    public onCameraStatusUpdate?: (cameraInfo: CameraInfo) => void;
    public onRecordingStatusUpdate?: (recordingInfo: RecordingInfo) => void;
    public onConnectionLost?: () => void;

    constructor(config: ClientConfig = {}) {
        this.host = config.host || 'localhost';
        this.port = config.port || 8080;
        this.useSsl = config.useSsl || false;
        this.authType = config.authType || 'jwt';
        this.authToken = config.authToken;
        this.apiKey = config.apiKey;
        this.maxRetries = config.maxRetries || 3;
        this.retryDelay = config.retryDelay || 1000;
        this.clientId = uuidv4();
    }

    private getWsUrl(): string {
        const protocol = this.useSsl ? 'wss' : 'ws';
        return `${protocol}://${this.host}:${this.port}/ws`;
    }

    private async authenticate(): Promise<void> {
        if (!this.connected) {
            throw new ConnectionError('Not connected to camera service');
        }

        let token: string | undefined;
        let authType: string = 'auto';

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
            const response = await this.sendRequest('authenticate', {
                token,
                auth_type: authType
            });

            if (response.authenticated) {
                this.authenticated = true;
                console.log(`Authenticated successfully with role: ${response.role || 'unknown'}`);
            } else {
                const errorMsg = response.error || 'Authentication failed';
                throw new AuthenticationError(`Authentication failed: ${errorMsg}`);
            }
        } catch (error) {
            if (error instanceof AuthenticationError) {
                throw error;
            }
            throw new AuthenticationError(`Authentication error: ${error}`);
        }
    }

    public async connect(): Promise<void> {
        for (let attempt = 1; attempt <= this.maxRetries; attempt++) {
            try {
                console.log(`Connecting to ${this.getWsUrl()} (attempt ${attempt})`);
                
                const wsUrl = this.getWsUrl();
                this.websocket = new WebSocket(wsUrl);
                
                // Set up event handlers
                this.setupWebSocketHandlers();
                
                // Wait for connection
                await this.waitForConnection();
                
                this.connected = true;
                console.log('Connected to camera service');
                
                // Authenticate if token provided
                if (this.authToken || this.apiKey) {
                    await this.authenticate();
                }
                
                // Test connection with ping
                await this.ping();
                console.log('Connection test successful');
                
                return;
                
            } catch (error) {
                console.error(`Connection attempt ${attempt} failed:`, error);
                if (attempt < this.maxRetries) {
                    await this.sleep(this.retryDelay * attempt);
                } else {
                    throw new ConnectionError(`Failed to connect after ${this.maxRetries} attempts: ${error}`);
                }
            }
        }
    }

    private setupWebSocketHandlers(): void {
        if (!this.websocket) return;

        this.websocket.on('open', () => {
            console.log('WebSocket connection opened');
        });

        this.websocket.on('message', (data: WebSocket.Data) => {
            this.processMessage(data.toString());
        });

        this.websocket.on('close', (code: number, reason: Buffer) => {
            console.warn(`WebSocket connection closed: ${code} - ${reason.toString()}`);
            this.connected = false;
            if (this.onConnectionLost) {
                this.onConnectionLost();
            }
        });

        this.websocket.on('error', (error: Error) => {
            console.error(`WebSocket error: ${error.message}`);
            this.connected = false;
        });
    }

    private waitForConnection(): Promise<void> {
        return new Promise((resolve, reject) => {
            if (!this.websocket) {
                reject(new Error('WebSocket not initialized'));
                return;
            }

            const timeout = setTimeout(() => {
                reject(new Error('Connection timeout'));
            }, 10000);

            this.websocket!.once('open', () => {
                clearTimeout(timeout);
                resolve();
            });

            this.websocket!.once('error', (error: Error) => {
                clearTimeout(timeout);
                reject(error);
            });
        });
    }

    private sleep(ms: number): Promise<void> {
        return new Promise(resolve => setTimeout(resolve, ms));
    }

    public async disconnect(): Promise<void> {
        if (this.websocket) {
            this.websocket.close();
            this.websocket = undefined;
        }
        this.connected = false;
        this.authenticated = false;
        console.log('Disconnected from camera service');
    }

    private processMessage(message: string): void {
        try {
            const data: JsonRpcMessage = JSON.parse(message);
            
            // Handle JSON-RPC response
            if ('id' in data && 'result' in data) {
                this.handleResponse(data as JsonRpcResponse);
            }
            // Handle JSON-RPC notification
            else if ('method' in data && !('id' in data)) {
                this.handleNotification(data as JsonRpcNotification);
            }
            else {
                console.warn(`Unknown message format: ${JSON.stringify(data)}`);
            }
            
        } catch (error) {
            console.error(`Invalid JSON message: ${error}`);
        }
    }

    private handleResponse(response: JsonRpcResponse): void {
        const requestId = response.id;
        if (this.pendingRequests.has(requestId)) {
            const { resolve, reject } = this.pendingRequests.get(requestId)!;
            this.pendingRequests.delete(requestId);
            
            if (response.error) {
                reject(new CameraServiceError(response.error.message || 'Unknown error'));
            } else {
                resolve(response.result);
            }
        }
    }

    private handleNotification(notification: JsonRpcNotification): void {
        const method = notification.method;
        const params = notification.params || {};
        
        if (method === 'camera_status_update' && this.onCameraStatusUpdate) {
            const cameraInfo: CameraInfo = {
                devicePath: params.device_path || '',
                name: params.name || '',
                capabilities: params.capabilities || [],
                status: params.status || '',
                streamUrl: params.stream_url
            };
            this.onCameraStatusUpdate(cameraInfo);
        } else if (method === 'recording_status_update' && this.onRecordingStatusUpdate) {
            const recordingInfo: RecordingInfo = {
                devicePath: params.device_path || '',
                recordingId: params.recording_id || '',
                filename: params.filename || '',
                startTime: params.start_time || 0,
                duration: params.duration,
                status: params.status || 'active'
            };
            this.onRecordingStatusUpdate(recordingInfo);
        } else {
            console.log(`Received notification: ${method}`);
        }
    }

    private async sendRequest(method: string, params: Record<string, any> = {}): Promise<any> {
        if (!this.connected) {
            throw new ConnectionError('Not connected to camera service');
        }
        
        this.requestId++;
        const requestId = this.requestId;
        
        const request: JsonRpcRequest = {
            jsonrpc: '2.0',
            id: requestId,
            method,
            params
        };
        
        return new Promise((resolve, reject) => {
            // Store the promise handlers
            this.pendingRequests.set(requestId, { resolve, reject });
            
            // Set timeout
            const timeout = setTimeout(() => {
                this.pendingRequests.delete(requestId);
                reject(new TimeoutError(`Request timeout: ${method}`));
            }, 30000);
            
            try {
                // Send request
                this.websocket!.send(JSON.stringify(request));
                
                // Override the resolve/reject to clear timeout
                const originalResolve = resolve;
                const originalReject = reject;
                
                this.pendingRequests.set(requestId, {
                    resolve: (value: any) => {
                        clearTimeout(timeout);
                        originalResolve(value);
                    },
                    reject: (reason: any) => {
                        clearTimeout(timeout);
                        originalReject(reason);
                    }
                });
                
            } catch (error) {
                clearTimeout(timeout);
                this.pendingRequests.delete(requestId);
                reject(new CameraServiceError(`Request failed: ${error}`));
            }
        });
    }

    public async ping(): Promise<string> {
        return await this.sendRequest('ping');
    }

    public async getCameraList(): Promise<CameraInfo[]> {
        const response = await this.sendRequest('get_camera_list');
        
        return response.map((cameraData: any) => ({
            devicePath: cameraData.device_path || '',
            name: cameraData.name || '',
            capabilities: cameraData.capabilities || [],
            status: cameraData.status || '',
            streamUrl: cameraData.stream_url
        }));
    }

    public async getCameraStatus(devicePath: string): Promise<CameraInfo> {
        const response = await this.sendRequest('get_camera_status', { device: devicePath });
        
        if (!response) {
            throw new CameraNotFoundError(`Camera not found: ${devicePath}`);
        }
        
        return {
            devicePath: response.device_path || devicePath,
            name: response.name || '',
            capabilities: response.capabilities || [],
            status: response.status || '',
            streamUrl: response.stream_url
        };
    }

    public async takeSnapshot(devicePath: string, filename?: string): Promise<SnapshotInfo> {
        const params: Record<string, any> = { device: devicePath };
        if (filename) {
            params.filename = filename;
        }
        
        const response = await this.sendRequest('take_snapshot', params);
        
        return {
            devicePath,
            filename: response.filename || '',
            timestamp: response.timestamp || 0,
            sizeBytes: response.size_bytes
        };
    }

    public async startRecording(devicePath: string, filename?: string): Promise<RecordingInfo> {
        const params: Record<string, any> = { device: devicePath };
        if (filename) {
            params.filename = filename;
        }
        
        const response = await this.sendRequest('start_recording', params);
        
        return {
            devicePath,
            recordingId: response.recording_id || '',
            filename: response.filename || '',
            startTime: response.start_time || 0,
            status: 'active'
        };
    }

    public async stopRecording(devicePath: string): Promise<RecordingInfo> {
        const response = await this.sendRequest('stop_recording', { device: devicePath });
        
        return {
            devicePath,
            recordingId: response.recording_id || '',
            filename: response.filename || '',
            startTime: response.start_time || 0,
            duration: response.duration,
            status: 'stopped'
        };
    }

    public async getRecordingStatus(devicePath: string): Promise<RecordingInfo | null> {
        const response = await this.sendRequest('get_recording_status', { device: devicePath });
        
        if (!response) {
            return null;
        }
        
        return {
            devicePath,
            recordingId: response.recording_id || '',
            filename: response.filename || '',
            startTime: response.start_time || 0,
            duration: response.duration,
            status: response.status || 'active'
        };
    }
}
