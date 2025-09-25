import { JsonRpcRequest, JsonRpcResponse, JsonRpcNotification, RpcMethod, ERROR_CODES } from '../../types/api';

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

export class WebSocketService {
  private ws: WebSocket | null = null;
  private config: Required<WebSocketServiceConfig>;
  public events: Partial<WebSocketServiceEvents> = {};
  private reconnectAttempts = 0;
  private reconnectTimeout: NodeJS.Timeout | null = null;
  private pingInterval: NodeJS.Timeout | null = null;
  private pongTimeout: NodeJS.Timeout | null = null;
  private pendingRequests = new Map<string | number, {
    resolve: (value: any) => void;
    reject: (error: Error) => void;
    timestamp: number;
  }>();
  private requestId = 0;

  constructor(config: WebSocketServiceConfig, events?: Partial<WebSocketServiceEvents>) {
    this.config = {
      maxReconnectAttempts: 5,
      reconnectDelay: 1000,
      maxReconnectDelay: 30000,
      pingInterval: 30000,
      pongTimeout: 5000,
      ...config
    };
    this.events = events || {};
  }

  async connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(this.config.url);
        
        this.ws.onopen = () => {
          console.log('WebSocket connected');
          this.reconnectAttempts = 0;
          this.startPingInterval();
          this.events.onConnect?.();
          resolve();
        };

        this.ws.onclose = (event) => {
          console.log('WebSocket closed', event.code, event.reason);
          this.stopPingInterval();
          this.events.onDisconnect?.(new Error(`Connection closed: ${event.code} ${event.reason}`));
          this.handleReconnect();
        };

        this.ws.onerror = (error) => {
          console.error('WebSocket error:', error);
          this.events.onError?.(new Error('WebSocket connection error'));
          reject(new Error('WebSocket connection failed'));
        };

        this.ws.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data);
            this.handleMessage(data);
          } catch (error) {
            console.error('Failed to parse WebSocket message:', error);
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

  async sendRPC<T = any>(method: RpcMethod, params?: any): Promise<T> {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      throw new Error('WebSocket not connected');
    }

    const id = ++this.requestId;
    const request: JsonRpcRequest = {
      jsonrpc: '2.0',
      method,
      params,
      id
    };

    return new Promise((resolve, reject) => {
      // Store pending request
      this.pendingRequests.set(id, {
        resolve,
        reject,
        timestamp: Date.now()
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

  sendNotification(method: string, params?: any): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      throw new Error('WebSocket not connected');
    }

    const notification: JsonRpcNotification = {
      jsonrpc: '2.0',
      method,
      params
    };

    this.ws?.send(JSON.stringify(notification));
  }

  private handleMessage(data: any): void {
    if (data.jsonrpc !== '2.0') {
      console.warn('Invalid JSON-RPC message:', data);
      return;
    }

    if (data.id !== undefined) {
      // Response to a request
      this.handleResponse(data as JsonRpcResponse);
    } else {
      // Notification
      this.handleNotification(data as JsonRpcNotification);
    }
  }

  private handleResponse(response: JsonRpcResponse): void {
    const pending = this.pendingRequests.get(response.id);
    if (!pending) {
      console.warn('Received response for unknown request:', response.id);
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
      
      const error = new Error(errorMessage);
      (error as any).code = errorCode;
      (error as any).data = response.error.data;
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
      console.error('Max reconnection attempts reached');
      return;
    }

    this.reconnectAttempts++;
    const delay = Math.min(
      this.config.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1),
      this.config.maxReconnectDelay
    );

    console.log(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts}/${this.config.maxReconnectAttempts})`);

    this.reconnectTimeout = setTimeout(() => {
      this.connect().catch((error) => {
        console.error('Reconnection failed:', error);
        this.handleReconnect();
      });
    }, delay);
  }

  private startPingInterval(): void {
    this.stopPingInterval();
    
    this.pingInterval = setInterval(() => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        this.sendRPC('ping').catch((error) => {
          console.error('Ping failed:', error);
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
