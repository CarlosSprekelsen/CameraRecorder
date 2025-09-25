/**
 * Mock Server Fallback Strategy
 * 
 * Provides mock responses that match real server behavior
 * Used only when real server is unavailable (CI/offline scenarios)
 * 
 * Environment variable: USE_MOCK_SERVER=true
 */

import { RPC_METHODS, ERROR_CODES } from '../../src/types';

export interface MockResponse {
  jsonrpc: '2.0';
  result?: any;
  error?: {
    code: number;
    message: string;
  };
  id: number;
}

export interface MockNotification {
  jsonrpc: '2.0';
  method: string;
  params: any;
}

/**
 * Mock server responses that match real server behavior
 */
export const MOCK_RESPONSES = {
  [RPC_METHODS.PING]: 'pong',
  
  [RPC_METHODS.GET_CAMERA_LIST]: {
    cameras: [
      {
        device: 'camera0',
        status: 'CONNECTED',
        name: 'Test Camera 1',
        resolution: '1920x1080',
        fps: 30,
        streams: {
          rtsp: 'rtsp://localhost:8554/camera0',
          webrtc: 'http://localhost:8889/camera0/webrtc',
          hls: 'http://localhost:8002/hls/camera0.m3u8'
        }
      },
      {
        device: 'camera1',
        status: 'CONNECTED',
        name: 'Test Camera 2',
        resolution: '1280x720',
        fps: 25,
        streams: {
          rtsp: 'rtsp://localhost:8554/camera1',
          webrtc: 'http://localhost:8889/camera1/webrtc',
          hls: 'http://localhost:8002/hls/camera1.m3u8'
        }
      }
    ],
    total: 2,
    connected: 2
  },
  
  [RPC_METHODS.GET_CAMERA_STATUS]: {
    device: 'camera0',
    status: 'CONNECTED',
    name: 'Test Camera 1',
    resolution: '1920x1080',
    fps: 30,
    streams: {
      rtsp: 'rtsp://localhost:8554/camera0',
      webrtc: 'http://localhost:8889/camera0/webrtc',
      hls: 'http://localhost:8002/hls/camera0.m3u8'
    },
    metrics: {
      bytes_sent: 12345678,
      readers: 2,
      uptime: 3600
    },
    capabilities: {
      formats: ['YUYV', 'MJPEG'],
      resolutions: ['1920x1080', '1280x720']
    }
  },
  
  [RPC_METHODS.TAKE_SNAPSHOT]: {
    device: 'camera0',
    filename: 'snapshot_2025-01-15_14-30-00.jpg',
    status: 'COMPLETED',
    timestamp: '2025-01-15T14:30:00Z',
    file_size: 204800,
    file_path: '/opt/camera-service/snapshots/snapshot_2025-01-15_14-30-00.jpg'
  },
  
  [RPC_METHODS.START_RECORDING]: {
    device: 'camera0',
    session_id: '550e8400-e29b-41d4-a716-446655440000',
    filename: 'camera0_2025-01-15_14-30-00.mp4',
    status: 'STARTED',
    start_time: '2025-01-15T14:30:00Z',
    duration: 3600,
    format: 'mp4'
  },
  
  [RPC_METHODS.STOP_RECORDING]: {
    device: 'camera0',
    session_id: '550e8400-e29b-41d4-a716-446655440000',
    filename: 'camera0_2025-01-15_14-30-00.mp4',
    status: 'STOPPED',
    start_time: '2025-01-15T14:30:00Z',
    end_time: '2025-01-15T15:00:00Z',
    duration: 1800,
    file_size: 1073741824
  },
  
  [RPC_METHODS.LIST_RECORDINGS]: {
    files: [
      {
        filename: 'camera0_2025-01-15_14-30-00.mp4',
        file_size: 1073741824,
        modified_time: '2025-01-15T14:30:00Z',
        download_url: '/files/recordings/camera0_2025-01-15_14-30-00.mp4'
      }
    ],
    total: 1,
    limit: 10,
    offset: 0
  },
  
  [RPC_METHODS.LIST_SNAPSHOTS]: {
    files: [
      {
        filename: 'snapshot_2025-01-15_14-30-00.jpg',
        file_size: 204800,
        modified_time: '2025-01-15T14:30:00Z',
        download_url: '/files/snapshots/snapshot_2025-01-15_14-30-00.jpg'
      }
    ],
    total: 1,
    limit: 10,
    offset: 0
  }
};

/**
 * Mock error responses
 */
export const MOCK_ERRORS = {
  [ERROR_CODES.CAMERA_NOT_FOUND_OR_DISCONNECTED]: {
    code: ERROR_CODES.CAMERA_NOT_FOUND_OR_DISCONNECTED,
    message: 'Camera not found'
  },
  [ERROR_CODES.METHOD_NOT_FOUND]: {
    code: ERROR_CODES.METHOD_NOT_FOUND,
    message: 'Method not found'
  },
  [ERROR_CODES.INVALID_PARAMS]: {
    code: ERROR_CODES.INVALID_PARAMS,
    message: 'Invalid parameters'
  }
};

/**
 * Mock WebSocket service for testing
 */
export class MockWebSocketService {
  private requestId = 0;
  private _isConnected = false;
  private messageHandlers: ((message: any) => void)[] = [];
  private notificationInterval: NodeJS.Timeout | null = null;

  constructor() {
    this._isConnected = true;
    this.startNotificationSimulation();
  }

  /**
   * Simulate WebSocket connection
   */
  async connect(): Promise<void> {
    // Simulate connection delay
    await new Promise(resolve => setTimeout(resolve, 10));
    this._isConnected = true;
  }

  /**
   * Simulate WebSocket disconnection
   */
  disconnect(): void {
    this._isConnected = false;
    if (this.notificationInterval) {
      clearInterval(this.notificationInterval);
      this.notificationInterval = null;
    }
  }

  /**
   * Check connection status
   */
  isConnected(): boolean {
    return this._isConnected;
  }

  /**
   * Simulate JSON-RPC call
   */
  async call(method: string, params: any = {}): Promise<any> {
    if (!this._isConnected) {
      throw new Error('WebSocket is not connected');
    }

    // Simulate network delay
    await new Promise(resolve => setTimeout(resolve, 10));

    const requestId = ++this.requestId;

    // Handle specific error cases
    if (method === RPC_METHODS.GET_CAMERA_STATUS && (!params.device || params.device === 'camera999')) {
      throw new Error('Camera not found');
    }

    if (method === 'invalid_method') {
      throw new Error('Method not found');
    }

    if (method === RPC_METHODS.GET_CAMERA_STATUS && !params.device) {
      throw new Error('Invalid parameters');
    }

    // Return mock response
    const mockResponse = MOCK_RESPONSES[method as keyof typeof MOCK_RESPONSES];
    if (mockResponse) {
      return mockResponse;
    }

    throw new Error('Method not implemented in mock');
  }

  /**
   * Register message handler
   */
  onMessage(handler: (message: any) => void): void {
    this.messageHandlers.push(handler);
  }

  /**
   * Start notification simulation
   */
  private startNotificationSimulation(): void {
    this.notificationInterval = setInterval(() => {
      if (!this._isConnected) return;

      // Simulate camera status update
      const cameraNotification: MockNotification = {
        jsonrpc: '2.0',
        method: 'camera_status_update',
        params: {
          device: 'camera0',
          status: 'CONNECTED',
          name: 'Test Camera 1',
          resolution: '1920x1080',
          fps: 30,
          streams: {
            rtsp: 'rtsp://localhost:8554/camera0',
            webrtc: 'http://localhost:8889/camera0/webrtc',
            hls: 'http://localhost:8002/hls/camera0.m3u8'
          }
        }
      };

      this.messageHandlers.forEach(handler => handler(cameraNotification));
    }, 5000); // Send notification every 5 seconds
  }
}

/**
 * Check if mock server should be used
 */
export function shouldUseMockServer(): boolean {
  return process.env.USE_MOCK_SERVER === 'true';
}

/**
 * Create appropriate WebSocket service (real or mock)
 */
export function createWebSocketService(): any {
  if (shouldUseMockServer()) {
    console.log('Using mock WebSocket service (USE_MOCK_SERVER=true)');
    return new MockWebSocketService();
  } else {
    // Import real service
    const { WebSocketService } = require('../../src/services/websocket');
    return new WebSocketService({
      url: process.env.TEST_WEBSOCKET_URL || 'ws://localhost:8002/ws',
      reconnectInterval: 1000,
      maxReconnectAttempts: 3,
      requestTimeout: 5000,
      heartbeatInterval: 30000,
    });
  }
}

/**
 * Validate mock responses against real server responses
 */
export function validateMockResponseAccuracy(
  mockResponse: any,
  realResponse: any,
  method: string
): boolean {
  // Basic structure validation
  if (typeof mockResponse !== typeof realResponse) {
    console.warn(`Mock response type mismatch for ${method}`);
    return false;
  }

  if (Array.isArray(mockResponse) !== Array.isArray(realResponse)) {
    console.warn(`Mock response array mismatch for ${method}`);
    return false;
  }

  // For objects, check key structure
  if (typeof mockResponse === 'object' && mockResponse !== null) {
    const mockKeys = Object.keys(mockResponse);
    const realKeys = Object.keys(realResponse);
    
    const missingKeys = realKeys.filter(key => !mockKeys.includes(key));
    if (missingKeys.length > 0) {
      console.warn(`Mock response missing keys for ${method}:`, missingKeys);
      return false;
    }
  }

  return true;
}
