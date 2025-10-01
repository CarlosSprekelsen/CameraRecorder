import {
  JsonRpcRequest,
  JsonRpcResponse,
  JsonRpcNotification,
  RpcMethod,
  ERROR_CODES,
} from '../../types/api';

export interface WebSocketServiceConfig {
  url: string;
  maxReconnectAttempts?: number;
  reconnectDelay?: number;
  maxReconnectDelay?: number;
  pingInterval?: number;
  pongTimeout?: number;
}

export interface WebSocketServiceEvents {
  onConnect: () => void;
  onDisconnect: (error?: Error) => void;
  onError: (error: Error) => void;
  onNotification: (notification: JsonRpcNotification) => void;
  onResponse: (response: JsonRpcResponse) => void;
}

/**
 * WebSocket Service
 *
 * Manages WebSocket connections and JSON-RPC 2.0 communication with the MediaMTX server.
 * Provides automatic reconnection, ping/pong heartbeat, and request/response handling.
 *
 * @class WebSocketService
 * @implements {WebSocketServiceConfig}
 *
 * @example
 * ```typescript
 * const wsService = new WebSocketService({
 *   url: 'ws://localhost:8002/ws',
 *   reconnectInterval: 5000,
 *   maxReconnectAttempts: 5
 * });
 *
 * wsService.events.onConnect = () => console.log('Connected');
 * await wsService.connect();
 * ```
 *
 * @see {@link https://www.jsonrpc.org/specification} JSON-RPC 2.0 Specification
 * @see {@link ../../docs/architecture/client-architechture.md} Client Architecture
 */
export class WebSocketService {
  private ws: WebSocket | null = null;
  private config: Required<WebSocketServiceConfig>;
  public events: Partial<WebSocketServiceEvents> = {};
  private reconnectAttempts = 0;
  private reconnectTimeout: NodeJS.Timeout | null = null;
  private pingInterval: NodeJS.Timeout | null = null;
  private pongTimeout: NodeJS.Timeout | null = null;
  private pendingRequests = new Map<
    string | number,
    {
      resolve: <T>(value: T) => void;
      reject: (error: Error) => void;
      timestamp: number;
      timeout?: NodeJS.Timeout;
    }
  >();
  private requestId = 0;

  constructor(config: WebSocketServiceConfig, events?: Partial<WebSocketServiceEvents>) {
    this.config = {
      maxReconnectAttempts: 5,
      reconnectDelay: 1000,
      maxReconnectDelay: 30000,
      pingInterval: 30000,
      pongTimeout: 5000,
      ...config,
    };
    this.events = events || {};
  }

  async connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      let connectionTimeout: NodeJS.Timeout | null = null;
      
      try {
        // FIXED: Store timeout reference for proper cleanup
        connectionTimeout = setTimeout(() => {
          if (this.ws && this.ws.readyState !== WebSocket.OPEN) {
            this.ws.close();
            reject(new Error('WebSocket connection timeout'));
          }
        }, 10000); // 10 second timeout

        this.ws = new WebSocket(this.config.url);

        this.ws.onopen = () => {
          // FIXED: Clear timeout on successful connection
          if (connectionTimeout) {
            clearTimeout(connectionTimeout);
            connectionTimeout = null;
          }
          this.reconnectAttempts = 0;
          this.startPingInterval();
          
          // Architecture requirement: Auto-subscribe to events after authentication
          // Note: autoSubscribeToEvents() will be called after successful authentication
          
          this.events.onConnect?.();
          resolve();
        };

        this.ws.onclose = (event) => {
          // FIXED: Clear timeout on close
          if (connectionTimeout) {
            clearTimeout(connectionTimeout);
            connectionTimeout = null;
          }
          this.stopPingInterval();
          this.events.onDisconnect?.(new Error(`Connection closed: ${event.code} ${event.reason}`));
          this.handleReconnect();
        };

        this.ws.onerror = (error) => {
          // FIXED: Clear timeout on error
          if (connectionTimeout) {
            clearTimeout(connectionTimeout);
            connectionTimeout = null;
          }
          console.error('WebSocket connection error:', error);
          this.events.onError?.(new Error('WebSocket connection error'));
          reject(new Error('WebSocket connection failed'));
        };

        this.ws.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data);
            this.handleMessage(data);
          } catch (error) {
            // Failed to parse WebSocket message - handled by error boundary
          }
        };
      } catch (error) {
        // FIXED: Clear timeout on exception
        if (connectionTimeout) {
          clearTimeout(connectionTimeout);
        }
        reject(error);
      }
    });
  }

  disconnect(): void {
    // FIXED: Proper resource cleanup to prevent leaks
    this.stopPingInterval();
    this.clearReconnectTimeout();

    // Clear all pending requests with proper timeout cleanup
    this.pendingRequests.forEach(({ reject, timeout }) => {
      if (timeout) {
        clearTimeout(timeout);
      }
      reject(new Error('WebSocket disconnected'));
    });
    this.pendingRequests.clear();

    // FIXED: Remove all event listeners before closing
    if (this.ws) {
      this.ws.onopen = null;
      this.ws.onclose = null;
      this.ws.onerror = null;
      this.ws.onmessage = null;
      
      // Close with proper cleanup
      if (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING) {
        this.ws.close(1000, 'Client disconnect');
      }
      this.ws = null;
    }

    // FIXED: Clear all notification handlers
    this.notificationHandlers.clear();
    
    // FIXED: Clear all event handlers
    this.events = {};
  }

  async sendRPC<T = unknown>(method: RpcMethod, params?: Record<string, unknown>): Promise<T> {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      throw new Error('WebSocket not connected');
    }

    const id = ++this.requestId;
    const request: JsonRpcRequest = {
      jsonrpc: '2.0',
      method,
      params,
      id,
    };

    return new Promise((resolve, reject) => {
      // Store pending request
      this.pendingRequests.set(id, {
        resolve: resolve as <T>(value: T) => void,
        reject,
        timestamp: Date.now(),
      });

      // Set timeout for request
      const requestTimeout = setTimeout(() => {
        if (this.pendingRequests.has(id)) {
          this.pendingRequests.delete(id);
          reject(new Error(`RPC request timeout: ${method}`));
        }
      }, 30000); // 30 second timeout
      
      // Store timeout for cleanup
      const pendingRequest = this.pendingRequests.get(id);
      if (pendingRequest) {
        pendingRequest.timeout = requestTimeout;
      }

      try {
        this.ws?.send(JSON.stringify(request));
      } catch (error) {
        this.pendingRequests.delete(id);
        reject(error);
      }
    });
  }

  sendNotification(method: string, params?: Record<string, unknown>): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      throw new Error('WebSocket not connected');
    }

    const notification: JsonRpcNotification = {
      jsonrpc: '2.0',
      method,
      params,
    };

    this.ws?.send(JSON.stringify(notification));
  }

  private handleMessage(data: unknown): void {
    const message = data as any;
    if (message.jsonrpc !== '2.0') {
      // Invalid JSON-RPC message - handled by error boundary
      return;
    }

    if (message.id !== undefined) {
      // Response to a request
      this.handleResponse(message as JsonRpcResponse);
    } else {
      // Notification
      this.handleNotification(message as JsonRpcNotification);
    }
  }

  private handleResponse(response: JsonRpcResponse): void {
    const pending = this.pendingRequests.get(response.id);
    if (!pending) {
      // Received response for unknown request - handled by error boundary
      return;
    }

    // Clear timeout if it exists
    if (pending.timeout) {
      clearTimeout(pending.timeout);
    }

    this.pendingRequests.delete(response.id);

    if (response.error) {
      // Handle specific error codes according to server API specification
      const errorCode = response.error.code;
      let errorMessage = response.error.message;

      switch (errorCode) {
        case ERROR_CODES.AUTH_FAILED:
          errorMessage = 'Authentication failed. Please log in again.';
          break;
        case ERROR_CODES.PERMISSION_DENIED:
          errorMessage = 'Permission denied. You do not have access to this operation.';
          break;
        case ERROR_CODES.NOT_FOUND:
          errorMessage = 'Resource not found.';
          break;
        case ERROR_CODES.INVALID_STATE:
          errorMessage = 'Operation not allowed in current state.';
          break;
        case ERROR_CODES.UNSUPPORTED:
          errorMessage = 'Feature not supported.';
          break;
        case ERROR_CODES.RATE_LIMITED:
          errorMessage = 'Rate limit exceeded. Please try again later.';
          break;
        case ERROR_CODES.DEPENDENCY_FAILED:
          errorMessage = 'External service unavailable.';
          break;
        default:
          errorMessage = `RPC Error ${errorCode}: ${response.error.message}`;
      }

      const error = new Error(errorMessage) as Error & { code: number; data: unknown };
      error.code = errorCode;
      error.data = response.error.data;
      pending.reject(error);
    } else {
      pending.resolve(response.result);
    }

    this.events.onResponse?.(response);
  }

  private notificationHandlers: Map<string, ((data: any) => void)[]> = new Map();

  private handleNotification(notification: JsonRpcNotification): void {
    // Handle server notifications
    console.log('Received notification:', notification);
    
    // Architecture requirement: Route notifications to appropriate store handlers
    this.routeNotification(notification);
    
    const handlers = this.notificationHandlers.get(notification.method);
    if (handlers) {
      handlers.forEach(handler => {
        try {
          handler(notification.params);
        } catch (error) {
          console.error('Error in notification handler:', error);
        }
      });
    }

    // Also call the legacy event handler
    this.events.onNotification?.(notification);
  }

  /**
   * Route notifications to appropriate store handlers
   * Architecture requirement: Route notifications to appropriate store handlers
   */
  private routeNotification(notification: JsonRpcNotification): void {
    switch (notification.method) {
      case 'camera_status_update':
        // Import dynamically to avoid circular dependencies
        import('../notifications/RealTimeNotificationHandler').then(({ RealTimeNotificationHandler }) => {
          const handler = new RealTimeNotificationHandler();
          handler.handleCameraStatusUpdate(notification.params as any);
        });
        break;
      case 'recording_status_update':
        import('../notifications/RealTimeNotificationHandler').then(({ RealTimeNotificationHandler }) => {
          const handler = new RealTimeNotificationHandler();
          handler.handleRecordingStatusUpdate(notification.params as any);
        });
        break;
      case 'system_health_update':
        import('../notifications/RealTimeNotificationHandler').then(({ RealTimeNotificationHandler }) => {
          const handler = new RealTimeNotificationHandler();
          handler.handleSystemHealthUpdate(notification.params);
        });
        break;
      case 'mediamtx.stream':
      case 'mediamtx.path':
      case 'mediamtx.stream_started':
      case 'mediamtx.stream_stopped':
        import('../notifications/RealTimeNotificationHandler').then(({ RealTimeNotificationHandler }) => {
          const handler = new RealTimeNotificationHandler();
          handler.handleStreamUpdate(notification.params as any);
        });
        break;
      default:
        console.log('Unhandled notification method:', notification.method);
    }
  }

  /**
   * Register a handler for server notifications
   */
  onNotification(method: string, handler: (data: any) => void): void {
    if (!this.notificationHandlers.has(method)) {
      this.notificationHandlers.set(method, []);
    }
    this.notificationHandlers.get(method)!.push(handler);
  }

  /**
   * Remove a notification handler
   */
  offNotification(method: string, handler: (data: any) => void): void {
    const handlers = this.notificationHandlers.get(method);
    if (handlers) {
      const index = handlers.indexOf(handler);
      if (index > -1) {
        handlers.splice(index, 1);
      }
    }
  }

  private handleReconnect(): void {
    if (this.reconnectAttempts >= this.config.maxReconnectAttempts) {
      // Max reconnection attempts reached - handled by connection store
      return;
    }

    this.reconnectAttempts++;
    const delay = Math.min(
      this.config.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1),
      this.config.maxReconnectDelay,
    );

    // Reconnecting - handled by connection store

    this.reconnectTimeout = setTimeout(() => {
      this.connect().catch(() => {
        // Reconnection failed - handled by connection store
        this.handleReconnect();
      });
    }, delay);
  }

  /**
   * Auto-subscribe to events on WebSocket connection
   * Architecture requirement: Auto-subscribe to events after successful authentication
   */
  autoSubscribeToEvents(): void {
    console.log('Auto-subscribing to real-time events');
    
    // Subscribe to camera status updates
    this.sendRPC('subscribe_events', { 
      topics: ['camera_status_update', 'recording_status_update', 'system_health_update'] 
    }).catch((error: any) => {
      console.error('Failed to subscribe to events:', error);
    });
  }

  private startPingInterval(): void {
    this.stopPingInterval();

    this.pingInterval = setInterval(() => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        this.sendRPC('ping').catch(() => {
          // Ping failed - handled by connection store
        });
      }
    }, this.config.pingInterval);
  }

  private stopPingInterval(): void {
    if (this.pingInterval) {
      clearInterval(this.pingInterval);
      this.pingInterval = null;
    }
    if (this.pongTimeout) {
      clearTimeout(this.pongTimeout);
      this.pongTimeout = null;
    }
  }

  private clearReconnectTimeout(): void {
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }
  }

  get isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  get connectionState(): number {
    return this.ws?.readyState ?? WebSocket.CLOSED;
  }
}
