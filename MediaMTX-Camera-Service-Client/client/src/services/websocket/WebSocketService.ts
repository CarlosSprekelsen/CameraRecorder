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
      try {
        this.ws = new WebSocket(this.config.url);

        this.ws.onopen = () => {
          // WebSocket connected - handled by connection store
          this.reconnectAttempts = 0;
          this.startPingInterval();
          this.events.onConnect?.();
          resolve();
        };

        this.ws.onclose = (event) => {
          // WebSocket closed - handled by connection store
          this.stopPingInterval();
          this.events.onDisconnect?.(new Error(`Connection closed: ${event.code} ${event.reason}`));
          this.handleReconnect();
        };

        this.ws.onerror = () => {
          // WebSocket error - handled by connection store
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
        reject(error);
      }
    });
  }

  disconnect(): void {
    this.stopPingInterval();
    this.clearReconnectTimeout();

    if (this.ws) {
      this.ws.close(1000, 'Client disconnect');
      this.ws = null;
    }

    // Reject all pending requests
    this.pendingRequests.forEach(({ reject }) => {
      reject(new Error('WebSocket disconnected'));
    });
    this.pendingRequests.clear();
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
      setTimeout(() => {
        if (this.pendingRequests.has(id)) {
          this.pendingRequests.delete(id);
          reject(new Error(`RPC request timeout: ${method}`));
        }
      }, 30000); // 30 second timeout

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

  private handleNotification(notification: JsonRpcNotification): void {
    this.events.onNotification?.(notification);
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
