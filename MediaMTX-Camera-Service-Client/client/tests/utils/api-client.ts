/**
 * SINGLE WebSocket client abstraction for all tests
 * Environment-driven: real connections for integration, mocks for unit
 * 
 * Ground Truth References:
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * 
 * Requirements Coverage:
 * - REQ-API-001: WebSocket connection management
 * - REQ-API-002: JSON-RPC 2.0 protocol compliance
 * - REQ-API-003: Authentication token handling
 * 
 * Test Categories: Unit/Integration/Performance
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */

import WebSocket from 'ws';
import { AuthResult } from '@/types/api';

export interface TestAPIClientConfig {
  mockMode?: boolean;
  serverUrl?: string;
  timeout?: number;
}

export class TestAPIClient {
  private ws: WebSocket | null = null;
  private isMockMode: boolean;
  private serverUrl: string;
  private timeout: number;
  private messageId = 0;
  private pendingRequests = new Map<number, { resolve: Function; reject: Function }>();

  constructor(config: TestAPIClientConfig = {}) {
    this.isMockMode = config.mockMode ?? process.env.NODE_ENV === 'test';
    this.serverUrl = config.serverUrl ?? process.env.TEST_WEBSOCKET_URL ?? 'ws://localhost:8002/ws';
    this.timeout = config.timeout ?? 30000;
  }

  /**
   * Connect to WebSocket server
   * MANDATORY: Use this method for all connection tests
   */
  async connect(): Promise<void> {
    if (this.isMockMode) {
      // Mock WebSocket for unit tests
      this.ws = new MockWebSocket();
      return Promise.resolve();
    }
    
    // Real WebSocket for integration tests
    return this.connectReal();
  }

  private async connectReal(): Promise<void> {
    return new Promise((resolve, reject) => {
      const timeout = setTimeout(() => {
        reject(new Error('WebSocket connection timeout'));
      }, this.timeout);

      this.ws = new WebSocket(this.serverUrl);
      
      this.ws.on('open', () => {
        clearTimeout(timeout);
        resolve();
      });
      
      this.ws.on('error', (error) => {
        clearTimeout(timeout);
        reject(error);
      });
      
      this.ws.on('message', (data) => {
        this.handleMessage(data);
      });
    });
  }

  /**
   * Disconnect from WebSocket server
   * MANDATORY: Use this method for cleanup
   */
  async disconnect(): Promise<void> {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.pendingRequests.clear();
  }

  /**
   * Call JSON-RPC method
   * MANDATORY: Use this method for all API calls
   * Validates against documented API schema
   */
  async call(method: string, params: any[] = []): Promise<any> {
    if (!this.ws) {
      throw new Error('WebSocket not connected');
    }

    const id = ++this.messageId;
    const request = {
      jsonrpc: '2.0',
      method,
      params,
      id
    };

    return new Promise((resolve, reject) => {
      const timeout = setTimeout(() => {
        this.pendingRequests.delete(id);
        reject(new Error(`Request timeout for method: ${method}`));
      }, this.timeout);

      this.pendingRequests.set(id, {
        resolve: (result: any) => {
          clearTimeout(timeout);
          resolve(result);
        },
        reject: (error: any) => {
          clearTimeout(timeout);
          reject(error);
        }
      });

      this.ws!.send(JSON.stringify(request));
    });
  }

  /**
   * Authenticate using JWT token
   * MANDATORY: Use this method for all authentication tests
   * Follows documented authentication flow exactly
   */
  async authenticate(token: string): Promise<AuthResult> {
    const result = await this.call('authenticate', { auth_token: token });
    
    // Validate against documented AuthResult schema
    if (!this.validateAuthResult(result)) {
      throw new Error('Invalid authentication result format');
    }
    
    return result;
  }

  /**
   * Ping connectivity check
   * MANDATORY: Use this method for connection validation
   */
  async ping(): Promise<string> {
    const result = await this.call('ping');
    
    if (result !== 'pong') {
      throw new Error('Invalid ping response');
    }
    
    return result;
  }

  private handleMessage(data: WebSocket.Data): void {
    try {
      const message = JSON.parse(data.toString());
      
      if (message.id && this.pendingRequests.has(message.id)) {
        const { resolve, reject } = this.pendingRequests.get(message.id)!;
        this.pendingRequests.delete(message.id);
        
        if (message.error) {
          reject(new Error(message.error.message || 'RPC Error'));
        } else {
          resolve(message.result);
        }
      }
    } catch (error) {
      console.error('Failed to parse WebSocket message:', error);
    }
  }

  /**
   * Validate authentication result against documented schema
   * MANDATORY: Use this validation for all auth tests
   */
  private validateAuthResult(result: any): boolean {
    return (
      typeof result === 'object' &&
      typeof result.authenticated === 'boolean' &&
      typeof result.role === 'string' &&
      ['admin', 'operator', 'viewer'].includes(result.role) &&
      typeof result.session_id === 'string'
    );
  }
}

/**
 * Mock WebSocket for unit tests
 * MANDATORY: Use this mock for all unit tests
 */
class MockWebSocket {
  private listeners: { [key: string]: Function[] } = {};

  on(event: string, listener: Function): void {
    if (!this.listeners[event]) {
      this.listeners[event] = [];
    }
    this.listeners[event].push(listener);
  }

  send(data: string): void {
    // Mock response based on request
    const request = JSON.parse(data);
    const response = this.getMockResponse(request);
    
    setTimeout(() => {
      this.listeners['message']?.forEach(listener => {
        listener(JSON.stringify(response));
      });
    }, 10);
  }

  close(): void {
    this.listeners['close']?.forEach(listener => {
      listener();
    });
  }

  private getMockResponse(request: any): any {
    switch (request.method) {
      case 'ping':
        return { jsonrpc: '2.0', result: 'pong', id: request.id };
      case 'authenticate':
        return {
          jsonrpc: '2.0',
          result: {
            authenticated: true,
            role: 'admin',
            permissions: ['read', 'write', 'delete'],
            expires_at: new Date(Date.now() + 3600000).toISOString(),
            session_id: 'test-session-id'
          },
          id: request.id
        };
      case 'get_camera_list':
        return {
          jsonrpc: '2.0',
          result: {
            cameras: [
              {
                device: 'camera0',
                status: 'CONNECTED',
                name: 'Test Camera',
                resolution: '1920x1080',
                fps: 30,
                streams: {
                  rtsp: 'rtsp://localhost:8554/camera0',
                  hls: 'https://localhost/hls/camera0.m3u8'
                }
              }
            ],
            total: 1,
            connected: 1
          },
          id: request.id
        };
      default:
        return {
          jsonrpc: '2.0',
          error: {
            code: -32601,
            message: 'Method Not Found'
          },
          id: request.id
        };
    }
  }
}
