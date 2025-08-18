/**
 * WebSocket JSON-RPC 2.0 Client for MediaMTX Camera Service
 * 
 * Sprint 3: Real Server Integration
 * Implements:
 * - Real WebSocket connection to MediaMTX server at ws://localhost:8002/ws
 * - JSON-RPC 2.0 protocol client with full error handling
 * - Exponential backoff for reconnection with production settings
 * - Comprehensive error handling and timeout management
 * - Real-time notification handling for camera status updates
 * 
 * References:
 * - Server API: docs/api/json-rpc-methods.md
 * - WebSocket Protocol: docs/api/websocket-protocol.md
 * - Test script: test-websocket.js
 */

import type {
  WebSocketConfig,
  JSONRPCRequest,
  JSONRPCResponse,
  JSONRPCNotification,
  WebSocketMessage,
  JSONRPCError,
} from '../types';
import { authService } from './authService';

export class WebSocketError extends Error {
  public code?: number;
  public data?: unknown;

  constructor(message: string, code?: number, data?: unknown) {
    super(message);
    this.name = 'WebSocketError';
    this.code = code;
    this.data = data;
  }
}

export class WebSocketService {
  private ws: WebSocket | null = null;
  private requestId = 0;
  private pendingRequests = new Map<number, {
    resolve: (value: unknown) => void;
    reject: (reason: WebSocketError) => void;
    timeout: NodeJS.Timeout;
  }>();
  private reconnectAttempts = 0;
  private reconnectTimeout: NodeJS.Timeout | null = null;
  private isConnecting = false;
  private isDestroyed = false;
  private heartbeatInterval: NodeJS.Timeout | null = null;

  // Event handlers
  private onMessageHandler?: (message: WebSocketMessage) => void;
  private onConnectHandler?: () => void;
  private onDisconnectHandler?: () => void;
  private onErrorHandler?: (error: WebSocketError) => void;

  private config: WebSocketConfig;

  constructor(config: WebSocketConfig) {
    this.config = config;
  }

  /**
   * Connect to the MediaMTX WebSocket server
   * Sprint 3: Real server integration with enhanced error handling
   */
  public connect(): Promise<void> {
    if (this.isDestroyed) {
      return Promise.resolve();
    }

    // Reset connection state for reconnection attempts
    this.isConnecting = true;

    return new Promise((resolve, reject) => {
      try {
        console.log(`🔌 Connecting to MediaMTX server: ${this.config.url}`);
        this.ws = new WebSocket(this.config.url);
        this.setupEventHandlers(resolve, reject);
      } catch (error) {
        this.isConnecting = false;
        const wsError = new WebSocketError('Failed to create WebSocket connection');
        console.error('❌ WebSocket connection failed:', wsError);
        reject(wsError);
      }
    });
  }

  /**
   * Disconnect from the WebSocket server
   */
  public disconnect(): void {
    console.log('🔌 Disconnecting from MediaMTX server');
    this.isDestroyed = true;
    this.clearReconnectTimeout();
    this.clearHeartbeat();
    
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }

    // Reset connection state
    this.isConnecting = false;
    this.reconnectAttempts = 0;

    // Reject all pending requests
    for (const [, { reject, timeout }] of this.pendingRequests) {
      clearTimeout(timeout);
      reject(new WebSocketError('Connection closed'));
    }
    this.pendingRequests.clear();
  }

  /**
   * Send a JSON-RPC method call with optional authentication
   * Sprint 3: Enhanced for real server integration
   */
  public async call(method: string, params: Record<string, unknown> = {}, requireAuth: boolean = false): Promise<unknown> {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      throw new WebSocketError('WebSocket not connected');
    }

    // Add authentication token if required and available
    let finalParams = params;
    if (requireAuth) {
      try {
        finalParams = authService.includeAuth(params);
      } catch (error) {
        throw new WebSocketError(`Authentication required: ${error instanceof Error ? error.message : 'Unknown error'}`);
      }
    }

    const requestId = ++this.requestId;
    const request: JSONRPCRequest = {
      jsonrpc: '2.0',
      method,
      params: finalParams,
      id: requestId
    };

    console.log(`📤 Sending ${method} (#${requestId})`, params ? JSON.stringify(params) : '');

    return new Promise((resolve, reject) => {
      const timeout = setTimeout(() => {
        this.pendingRequests.delete(requestId);
        const timeoutError = new WebSocketError(`Request timeout for method: ${method}`);
        console.error(`⏰ Request timeout: ${method}`, timeoutError);
        reject(timeoutError);
      }, this.config.requestTimeout);

      this.pendingRequests.set(requestId, { resolve, reject, timeout });

      try {
        this.ws!.send(JSON.stringify(request));
      } catch (error) {
        this.pendingRequests.delete(requestId);
        clearTimeout(timeout);
        const sendError = new WebSocketError('Failed to send request');
        console.error('❌ Failed to send request:', sendError);
        reject(sendError);
      }
    });
  }

  /**
   * Set event handlers
   */
  public onMessage(handler: (message: WebSocketMessage) => void): void {
    this.onMessageHandler = handler;
  }

  public onConnect(handler: () => void): void {
    this.onConnectHandler = handler;
  }

  public onDisconnect(handler: () => void): void {
    this.onDisconnectHandler = handler;
  }

  public onError(handler: (error: WebSocketError) => void): void {
    this.onErrorHandler = handler;
  }

  /**
   * Get connection status
   */
  public get isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  public get isConnectingStatus(): boolean {
    return this.isConnecting;
  }

  /**
   * Test-only accessor for WebSocket instance
   * Only available in test environment
   */
  public getWebSocket(): WebSocket | null {
    if (process.env.NODE_ENV === 'test') {
      return this.ws;
    }
    return null;
  }

  /**
   * Start heartbeat to keep connection alive
   */
  private startHeartbeat(): void {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
    }

    this.heartbeatInterval = setInterval(async () => {
      if (this.isConnected) {
        try {
          await this.call('ping', {});
          console.log('💓 Heartbeat sent');
        } catch (error) {
          console.warn('⚠️ Heartbeat failed:', error);
          // Don't trigger reconnection on heartbeat failure
        }
      }
    }, this.config.heartbeatInterval);
  }

  /**
   * Clear heartbeat interval
   */
  private clearHeartbeat(): void {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
  }

  private setupEventHandlers(
    resolve: () => void,
    reject: (error: WebSocketError) => void
  ): void {
    if (!this.ws) return;

    this.ws.onopen = () => {
      console.log('✅ WebSocket connection established');
      this.isConnecting = false;
      this.reconnectAttempts = 0;
      this.startHeartbeat();
      this.onConnectHandler?.();
      resolve();
    };

    this.ws.onclose = (event) => {
      console.log('🔌 WebSocket connection closed', { 
        wasClean: event.wasClean, 
        code: event.code, 
        reason: event.reason 
      });
      this.isConnecting = false;
      this.clearHeartbeat();
      this.onDisconnectHandler?.();

      if (!this.isDestroyed && !event.wasClean) {
        console.log('🔄 Scheduling reconnection...');
        this.scheduleReconnect();
      }
    };

    this.ws.onerror = (event) => {
      console.error('❌ WebSocket error:', event);
      this.isConnecting = false;
      const error = new WebSocketError('WebSocket error occurred');
      this.onErrorHandler?.(error);
      // Don't reject if we're already connected
      if (this.isConnecting) {
        reject(error);
      }
    };

    this.ws.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);
        console.log('📥 Received message:', JSON.stringify(message));
        this.handleMessage(message);
      } catch (error) {
        const wsError = new WebSocketError('Failed to parse message');
        console.error('❌ Message parsing error:', wsError);
        this.onErrorHandler?.(wsError);
      }
    };
  }

  private handleMessage(message: WebSocketMessage): void {
    // Handle notifications (no id)
    if (!('id' in message)) {
      console.log('📢 Notification received:', message.method);
      this.onMessageHandler?.(message as JSONRPCNotification);
      return;
    }

    // Handle responses
    const response = message as JSONRPCResponse;
    const pendingRequest = this.pendingRequests.get(response.id);

    if (!pendingRequest) {
      console.warn('⚠️ Response for unknown request:', response.id);
      return; // Response for unknown request
    }

    const { resolve, reject, timeout } = pendingRequest;
    clearTimeout(timeout);
    this.pendingRequests.delete(response.id);

    if (response.error) {
      const error = new WebSocketError(
        response.error.message,
        response.error.code,
        response.error.data
      );
      console.error('❌ RPC Error:', error);
      reject(error);
    } else {
      console.log('✅ RPC Response received for request:', response.id);
      resolve(response.result);
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= this.config.maxReconnectAttempts) {
      const error = new WebSocketError('Max reconnection attempts reached');
      console.error('❌ Max reconnection attempts reached');
      this.onErrorHandler?.(error);
      return;
    }

    const delay = Math.min(
      this.config.reconnectInterval * Math.pow(2, this.reconnectAttempts),
      this.config.maxDelay
    );

    console.log(`🔄 Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts + 1}/${this.config.maxReconnectAttempts})`);

    this.reconnectTimeout = setTimeout(() => {
      this.reconnectAttempts++;
      this.connect().then(() => {
        console.log('✅ Reconnection successful');
        // Reconnection successful - onConnectHandler will be called by setupEventHandlers
      }).catch((error) => {
        console.error('❌ Reconnection failed:', error);
        this.onErrorHandler?.(error);
      });
    }, delay);
  }

  private clearReconnectTimeout(): void {
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }
  }
}

/**
 * Default configuration for WebSocket service
 * Sprint 3: Production-ready configuration for real MediaMTX server
 */
export const defaultWebSocketConfig: WebSocketConfig = {
  url: 'ws://localhost:8002/ws', // Real MediaMTX Camera Service endpoint
  maxReconnectAttempts: 10, // Finite attempts for production
  reconnectInterval: 1000, // 1 second base delay
  requestTimeout: 15000, // 15 second timeout for real server calls
  heartbeatInterval: 30000, // 30 second heartbeat
  baseDelay: 1000, // 1 second base delay
  maxDelay: 30000, // 30 second max delay
};

/**
 * Create a WebSocket service instance
 * Sprint 3: Enhanced for real server integration
 */
export function createWebSocketService(config: Partial<WebSocketConfig> = {}): WebSocketService {
  const finalConfig = { ...defaultWebSocketConfig, ...config };
  console.log('🔧 Creating WebSocket service with config:', finalConfig);
  return new WebSocketService(finalConfig);
} 