/**
 * WebSocket JSON-RPC 2.0 Client for MediaMTX Camera Service
 * 
 * Implements:
 * - WebSocket connection management with auto-reconnect
 * - JSON-RPC 2.0 protocol client
 * - Exponential backoff for reconnection
 * - Error handling and timeout management
 * 
 * References:
 * - Server API: docs/api/json-rpc-methods.md
 * - WebSocket Protocol: docs/api/websocket-protocol.md
 */

// STOP: clarify JSON-RPC error payload shape [Client-S2]
// STOP: clarify WebSocket reconnection timeout values [Client-S2]
// STOP: npm test hangs - investigate Jest/WebSocket mocking [Client-S2]
// STOP: clarify reconnection timeout values [Client-S1] - Using 100ms for tests, 500ms default okay?
// STOP: clarify max reconnection attempts [Client-S1] - Should default be Infinity or finite number?
// STOP: clarify WebSocket readyState handling [Client-S1] - Should we check readyState before operations?

import type {
  WebSocketConfig,
  JSONRPCRequest,
  JSONRPCResponse,
  JSONRPCNotification,
  WebSocketMessage,
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
   * Connect to the WebSocket server
   */
  public connect(): Promise<void> {
    if (this.isDestroyed) {
      return Promise.resolve();
    }

    // Reset connection state for reconnection attempts
    this.isConnecting = true;

    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(this.config.url);
        this.setupEventHandlers(resolve, reject);
      } catch {
        this.isConnecting = false;
        reject(new WebSocketError('Failed to create WebSocket connection'));
      }
    });
  }

  /**
   * Disconnect from the WebSocket server
   */
  public disconnect(): void {
    this.isDestroyed = true;
    this.clearReconnectTimeout();
    
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

    return new Promise((resolve, reject) => {
      const timeout = setTimeout(() => {
        this.pendingRequests.delete(requestId);
        reject(new WebSocketError(`Request timeout for method: ${method}`));
      }, this.config.requestTimeout);

      this.pendingRequests.set(requestId, { resolve, reject, timeout });

      try {
        this.ws!.send(JSON.stringify(request));
      } catch {
        this.pendingRequests.delete(requestId);
        clearTimeout(timeout);
        reject(new WebSocketError('Failed to send request'));
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

  private setupEventHandlers(
    resolve: () => void,
    reject: (error: WebSocketError) => void
  ): void {
    if (!this.ws) return;

    this.ws.onopen = () => {
      this.isConnecting = false;
      this.reconnectAttempts = 0;
      this.onConnectHandler?.();
      resolve();
    };

    this.ws.onclose = (event) => {
      this.isConnecting = false;
      this.onDisconnectHandler?.();

      if (!this.isDestroyed && !event.wasClean) {
        this.scheduleReconnect();
      }
    };

    this.ws.onerror = () => {
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
        this.handleMessage(message);
      } catch {
        const wsError = new WebSocketError('Failed to parse message');
        this.onErrorHandler?.(wsError);
      }
    };
  }

  private handleMessage(message: WebSocketMessage): void {
    // Handle notifications (no id)
    if (!('id' in message)) {
      this.onMessageHandler?.(message as JSONRPCNotification);
      return;
    }

    // Handle responses
    const response = message as JSONRPCResponse;
    const pendingRequest = this.pendingRequests.get(response.id);

    if (!pendingRequest) {
      return; // Response for unknown request
    }

    const { resolve, reject, timeout } = pendingRequest;
    clearTimeout(timeout);
    this.pendingRequests.delete(response.id);

    if (response.error) {
      reject(new WebSocketError(
        response.error.message,
        response.error.code,
        response.error.data
      ));
    } else {
      resolve(response.result);
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= this.config.maxReconnectAttempts) {
      const error = new WebSocketError('Max reconnection attempts reached');
      this.onErrorHandler?.(error);
      return;
    }

    const delay = Math.min(
      this.config.reconnectInterval * Math.pow(2, this.reconnectAttempts),
      30000 // 30 second max delay
    );

    this.reconnectTimeout = setTimeout(() => {
      this.reconnectAttempts++;
      this.connect().then(() => {
        // Reconnection successful - onConnectHandler will be called by setupEventHandlers
      }).catch((error) => {
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
 * Updated for Sprint 3: Real server integration
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
 */
export function createWebSocketService(config: Partial<WebSocketConfig> = {}): WebSocketService {
  const finalConfig = { ...defaultWebSocketConfig, ...config };
  return new WebSocketService(finalConfig);
} 