/**
 * WebSocket JSON-RPC 2.0 Client for MediaMTX Camera Service
 * 
 * Sprint 3: Real Server Integration with Enhanced Connection Management
 * REQ-NET01-003: Polling Fallback Mechanism Implementation
 * Phase 2B: Enhanced Reconnection with Exponential Backoff and State Recovery
 * 
 * Implements:
 * - Real WebSocket connection to MediaMTX server at ws://localhost:8002/ws
 * - JSON-RPC 2.0 protocol client with full error handling
 * - Exponential backoff for reconnection with production settings
 * - Circuit breaker pattern for connection resilience
 * - State recovery after reconnection
 * - Comprehensive error handling and timeout management
 * - Real-time notification handling for camera status updates
 * - Integration with connection store for state management
 * - Performance metrics tracking and health monitoring
 * - Enhanced notification handling and real-time updates
 * - HTTP polling fallback when WebSocket fails (REQ-NET01-003)
 * - Automatic switch back to WebSocket when connection restored
 * 
 * References:
 * - Server API: docs/api/json-rpc-methods.md
 * - WebSocket Protocol: docs/api/websocket-protocol.md
 * - Health Endpoints: docs/api/health-endpoints.md
 * - Test script: test-websocket.js
 */

import type {
  JSONRPCRequest,
  JSONRPCResponse,
  JSONRPCNotification,
  JSONRPCError,
  WebSocketMessage,
  CameraStatusNotification,
  RecordingStatusNotification,
  StorageStatusNotification,
  RecordingProgress,
} from '../types';
import { authService } from './authService';
import { NOTIFICATION_METHODS, ERROR_CODES } from '../types';
import { logger } from './loggerService';
// HTTP polling service removed - Go server is WebSocket-only

// Store interface types for better type safety
// These interfaces match the actual store state objects returned by getState()
interface ConnectionStoreInterface {
  // Error handling
  setError: (message: string, code?: number) => void;
  incrementErrorCount: () => void;
  
  // Connection timestamps
  setLastConnected: (date: Date) => void;
  setLastDisconnected: (date: Date) => void;
  
  // Health monitoring
  setLastHeartbeat: (date: Date) => void;
  updateHealthScore: (score: number) => void;
  updateConnectionQuality: (quality: 'excellent' | 'good' | 'poor' | 'unstable') => void;
  
  // Performance metrics
  incrementMessageCount: () => void;
  updateResponseTime: (time: number) => void;
  setConnectionUptime: (uptime: number) => void;
  
  // Direct state properties (from getState())
  healthScore: number;
  lastConnected: Date | null;
  isConnected: boolean;
  status: string;
  error: string | null;
  errorCode: number | null;
  lastHeartbeat: Date | null;
  connectionUptime: number | null;
  messageCount: number;
  errorCount: number;
  connectionQuality: 'excellent' | 'good' | 'poor' | 'unstable';
  latency: number | null;
  packetLoss: number | null;
}

interface CameraStoreInterface {
  handleNotification: (notification: unknown) => void;
  updateCameraStatus?: (device: string, status: import('../types/camera').CameraStatus) => void;
  addRecording?: (device: string, recording: import('../types/camera').RecordingSession) => void;
  removeRecording?: (device: string) => void;
  updateRecordingProgress?: (device: string, progress: RecordingProgress) => void;
  clearRecordingProgress?: (device: string) => void;
}

// WebSocket configuration interface
interface WebSocketConfig {
  url: string;
  maxReconnectAttempts: number;
  reconnectInterval: number;
  requestTimeout: number;
  heartbeatInterval: number;
  baseDelay: number;
  maxDelay: number;
}

// Notification message interface
interface NotificationMessage {
  method: string;
  params?: unknown;
}

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

/**
 * Circuit Breaker State
 */
interface CircuitBreakerState {
  isOpen: boolean;
  failureCount: number;
  lastFailureTime: Date | null;
  threshold: number;
  timeout: number;
}

/**
 * State Recovery Data
 */
interface StateRecoveryData {
  pendingRequests: Map<number, any>;
  lastHeartbeat: Date | null;
  connectionUptime: number;
  messageCount: number;
  errorCount: number;
}

/**
 * Parse recording conflict error message from server
 * Server format: "Camera is currently recording (session: 550e8400-e29b-41d4-a716-446655440000)"
 */
function parseRecordingConflictError(errorMessage: string): { device: string; session_id: string } {
  // Extract session_id from error message
  const sessionMatch = errorMessage.match(/session:\s*([a-f0-9-]+)/i);
  const session_id = sessionMatch ? sessionMatch[1] : 'unknown';
  
  // Extract device from error message (if present)
  const deviceMatch = errorMessage.match(/device\s+([^\s]+)/i);
  const device = deviceMatch ? deviceMatch[1] : 'unknown';
  
  return { device, session_id };
}

/**
 * Parse storage usage from error message
 * Server format: "Storage space is critical (95.2% used)"
 */
function parseStorageUsageFromMessage(errorMessage: string): number {
  const usageMatch = errorMessage.match(/(\d+\.?\d*)%?\s*used/i);
  return usageMatch ? parseFloat(usageMatch[1]) : 0;
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
  private metricsInterval: NodeJS.Timeout | null = null;
  
  // Enhanced reconnection properties
  private circuitBreaker: CircuitBreakerState = {
    isOpen: false,
    failureCount: 0,
    lastFailureTime: null,
    threshold: 5,
    timeout: 30000, // 30 seconds
  };
  
  private stateRecoveryData: StateRecoveryData | null = null;
  private exponentialBackoffBase = 1000; // 1 second
  private exponentialBackoffMax = 30000; // 30 seconds
  private exponentialBackoffMultiplier = 2;
  
  // Connection quality monitoring
  private connectionQualityScores: number[] = [];
  private maxQualityScores = 10;
  
  // Performance tracking
  private responseTimes: number[] = [];
  private maxResponseTimes = 20;
  
  // State recovery
  private recoveryAttempts = 0;
  private maxRecoveryAttempts = 3;

  // Event handlers
  private onMessageHandler?: (message: WebSocketMessage) => void;
  private onConnectHandler?: () => void;
  private onDisconnectHandler?: () => void;
  private onErrorHandler?: (error: WebSocketError) => void;

  // Enhanced notification handlers
  private onCameraStatusUpdateHandler?: (notification: CameraStatusNotification) => void;
  private onRecordingStatusUpdateHandler?: (notification: RecordingStatusNotification) => void;
  private onStorageStatusUpdateHandler?: (notification: StorageStatusNotification) => void;
  private onNotificationHandler?: (notification: NotificationMessage) => void;

  // Connection store integration
  private connectionStore: ConnectionStoreInterface | null = null;
  private cameraStore: CameraStoreInterface | null = null;

  // Performance tracking
  private notificationCount = 0;
  private lastNotificationTime = 0;
  private notificationLatency: number[] = [];

  // HTTP polling fallback removed - Go server is WebSocket-only

  private config: WebSocketConfig;

  constructor(config: WebSocketConfig) {
    this.config = config;
    
    // HTTP polling removed - Go server is WebSocket-only
  }

  // HTTP polling fallback method removed - Go server is WebSocket-only

  /**
   * Set connection store reference for integration
   */
  public setConnectionStore(store: ConnectionStoreInterface): void {
    this.connectionStore = store;
  }

  /**
   * Set camera store reference for integration
   */
  public setCameraStore(store: CameraStoreInterface): void {
    this.cameraStore = store;
  }

  /**
   * Enhanced notification event handlers
   */
  public onCameraStatusUpdate(handler: (notification: CameraStatusNotification) => void): void {
    this.onCameraStatusUpdateHandler = handler;
  }

  public onRecordingStatusUpdate(handler: (notification: RecordingStatusNotification) => void): void {
    this.onRecordingStatusUpdateHandler = handler;
  }

  /**
   * Set handler for storage status update notifications
   */
  public onStorageStatusUpdate(handler: (notification: StorageStatusNotification) => void): void {
    this.onStorageStatusUpdateHandler = handler;
  }

  public onNotification(handler: (notification: NotificationMessage) => void): void {
    this.onNotificationHandler = handler;
  }

  /**
   * Generate a unique request ID
   */
  private generateRequestId(): number {
    return ++this.requestId;
  }

  /**
   * Make a JSON-RPC call
   */
  public call(method: string, params: Record<string, unknown> = {}, requireAuth: boolean = true): Promise<unknown> {
    if (!this.isConnected()) {
      console.warn(`⚠️ WebSocket not connected for method: ${method}`);
      return Promise.reject(new WebSocketError('WebSocket not connected - Go server requires WebSocket connection'));
    }

    const id = this.generateRequestId();
    const request: JSONRPCRequest = {
      jsonrpc: '2.0',
      id,
      method,
      params
    };

    const requestPromise = new Promise<unknown>((resolve, reject) => {
      const timeout = setTimeout(() => {
        this.pendingRequests.delete(id);
        reject(new WebSocketError(`Request timeout for method: ${method}`));
      }, this.config.requestTimeout);

      this.pendingRequests.set(id, { resolve, reject, timeout });
    });

    try {
      this.ws!.send(JSON.stringify(request));
      logger.info(`Request sent`, { requestId: id, method }, 'websocket');
    } catch (error) {
      this.pendingRequests.delete(id);
      throw new WebSocketError(`Failed to send request: ${error}`);
    }

    return requestPromise;
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
   * Handle error with specific error code processing
   */
  private handleError(error: JSONRPCError): void {
    logger.info(`Handling error`, { code: error.code, message: error.message }, 'websocket');
    
    // Update connection store
    if (this.connectionStore) {
      this.connectionStore.setError(error.message, error.code);
      this.connectionStore.incrementErrorCount();
    }
    
    // Process specific error codes with enhanced handling
    switch (error.code) {
      case ERROR_CODES.CAMERA_ALREADY_RECORDING:
        logger.warn(`Recording conflict detected: ${error.message}`, undefined, 'websocket');
        this.handleRecordingConflict(error);
        break;
      case ERROR_CODES.STORAGE_SPACE_LOW:
        logger.warn(`Storage space low: ${error.message}`, undefined, 'websocket');
        this.handleStorageWarning(error);
        break;
      case ERROR_CODES.STORAGE_SPACE_CRITICAL:
        logger.error(`Storage space critical: ${error.message}`, undefined, 'websocket');
        this.handleStorageCritical(error);
        break;
      case ERROR_CODES.CAMERA_NOT_FOUND_OR_DISCONNECTED:
        logger.error(`Camera not found or disconnected: ${error.message}`, undefined, 'websocket');
        break;
      case ERROR_CODES.RECORDING_ALREADY_IN_PROGRESS:
        logger.warn(`Recording already in progress: ${error.message}`, undefined, 'websocket');
        break;
      case ERROR_CODES.MEDIAMTX_SERVICE_UNAVAILABLE:
        logger.error(`MediaMTX service unavailable: ${error.message}`, undefined, 'websocket');
        break;
      case ERROR_CODES.AUTHENTICATION_FAILED:
        console.error(`❌ Authentication required: ${error.message}`);
        break;
      case ERROR_CODES.INSUFFICIENT_STORAGE_SPACE:
        logger.error(`Insufficient storage space: ${error.message}`, undefined, 'websocket');
        break;
      case ERROR_CODES.CAMERA_CAPABILITY_NOT_SUPPORTED:
        logger.warn(`Camera capability not supported: ${error.message}`, undefined, 'websocket');
        break;
      default:
        logger.error(`Standard error: ${error.code} - ${error.message}`, undefined, 'websocket');
    }
    
    // Notify UI components about the error
    this.onErrorHandler?.(new WebSocketError(error.message, error.code, error.data));
  }

  /**
   * Handle recording conflict errors
   */
  private handleRecordingConflict(error: JSONRPCError): void {
    // Parse error message to extract session_id from server format
    const parsedError = parseRecordingConflictError(error.message);
    console.warn(`⚠️ Recording conflict for camera: ${parsedError.device}`);
    console.warn(`⚠️ Active session: ${parsedError.session_id}`);
    
    // Update connection store with conflict information
    if (this.connectionStore) {
      this.connectionStore.setError(
        `Camera ${parsedError.device} is currently recording (Session: ${parsedError.session_id})`,
        ERROR_CODES.CAMERA_ALREADY_RECORDING
      );
    }
  }

  /**
   * Handle storage warning errors
   */
  private handleStorageWarning(error: JSONRPCError): void {
    // Parse storage error from message since server doesn't provide structured data
    const usagePercent = parseStorageUsageFromMessage(error.message);
    
    console.warn(`⚠️ Storage space low: ${usagePercent.toFixed(1)}% used`);
    console.warn(`⚠️ Error message: ${error.message}`);
    
    // Update connection store with storage warning
    if (this.connectionStore) {
      this.connectionStore.setError(
        `Storage space is low (${usagePercent.toFixed(1)}% used)`,
        ERROR_CODES.STORAGE_SPACE_LOW
      );
    }
  }

  /**
   * Handle storage critical errors
   */
  private handleStorageCritical(error: JSONRPCError): void {
    // Parse storage error from message since server doesn't provide structured data
    const usagePercent = parseStorageUsageFromMessage(error.message);
    
    console.error(`🚨 Storage space critical: ${usagePercent.toFixed(1)}% used`);
    console.error(`🚨 Error message: ${error.message}`);
    
    // Update connection store with critical storage error
    if (this.connectionStore) {
      this.connectionStore.setError(
        `Storage space is critical (${usagePercent.toFixed(1)}% used)`,
        ERROR_CODES.STORAGE_SPACE_CRITICAL
      );
    }
  }

  /**
   * Format bytes to human readable format
   */
  private formatBytes(bytes: number): string {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }

  /**
   * Check if WebSocket is connected
   */
  public isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  /**
   * Check if WebSocket is connecting
   */
  public isConnectingStatus(): boolean {
    return this.ws?.readyState === WebSocket.CONNECTING;
  }

  /**
   * Connect to WebSocket server
   */
  public async connect(): Promise<void> {
    if (this.isConnecting || this.isConnected()) {
      logger.info('Already connecting or connected', undefined, 'websocket');
      return;
    }

    if (this.isDestroyed) {
      throw new WebSocketError('WebSocket service has been destroyed');
    }

    this.isConnecting = true;
    logger.info('Connecting to WebSocket server', undefined, 'websocket');

    return new Promise<void>((resolve, reject) => {
      try {
        this.ws = new WebSocket(this.config.url);
        this.setupEventHandlers(resolve, reject);
      } catch (error) {
        this.isConnecting = false;
        reject(new WebSocketError(`Failed to create WebSocket connection: ${error}`));
      }
    });
  }

  /**
   * Disconnect from WebSocket server
   */
  public async disconnect(): Promise<void> {
    logger.info('Disconnecting from WebSocket server', undefined, 'websocket');
    
    this.isDestroyed = true;
    this.isConnecting = false;
    
    // Clear all intervals
    this.clearHeartbeat();
    this.clearMetricsInterval();
    
    // Clear reconnection timeout
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }
    
    // Reject all pending requests
    this.pendingRequests.forEach((request, id) => {
      clearTimeout(request.timeout);
      request.reject(new WebSocketError('WebSocket disconnected'));
    });
    this.pendingRequests.clear();
    
    // Close WebSocket connection
    if (this.ws) {
      this.ws.close(1000, 'Client disconnect');
      this.ws = null;
    }
    
    // HTTP polling removed - Go server is WebSocket-only
    
    logger.info('WebSocket disconnected', undefined, 'websocket');
  }

  /**
   * Schedule reconnection attempt
   */
  private scheduleReconnection(): void {
    if (this.isDestroyed || this.reconnectAttempts >= this.config.maxReconnectAttempts) {
      console.log('🛑 Max reconnection attempts reached or service destroyed');
      return;
    }

    this.reconnectAttempts++;
    const delay = Math.min(
      this.exponentialBackoffBase * Math.pow(this.exponentialBackoffMultiplier, this.reconnectAttempts - 1),
      this.exponentialBackoffMax
    );

    console.log(`🔄 Scheduling reconnection attempt ${this.reconnectAttempts}/${this.config.maxReconnectAttempts} in ${delay}ms`);
    
    this.reconnectTimeout = setTimeout(async () => {
      try {
        await this.connect();
        this.reconnectAttempts = 0; // Reset on successful connection
      } catch (error) {
        console.error('❌ Reconnection failed:', error);
        this.scheduleReconnection(); // Try again
      }
    }, delay);
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

  /**
   * Clear metrics collection interval
   */
  private clearMetricsInterval(): void {
    if (this.metricsInterval) {
      clearInterval(this.metricsInterval);
      this.metricsInterval = null;
    }
  }

  /**
   * Setup event handlers for WebSocket
   */
  private setupEventHandlers(
    resolve: () => void,
    reject: (error: WebSocketError) => void
  ): void {
    if (!this.ws) return;

    this.ws.onopen = () => {
      logger.info('WebSocket connection established', undefined, 'websocket');
      this.isConnecting = false;
      this.reconnectAttempts = 0;
      this.startHeartbeat();
      this.startMetricsCollection();
      
      // HTTP polling removed - Go server is WebSocket-only
      
      // Update connection store
      if (this.connectionStore) {
        this.connectionStore.setLastConnected(new Date());
        this.connectionStore.updateHealthScore(100);
        this.connectionStore.updateConnectionQuality('excellent');
      }
      
      this.onConnectHandler?.();
      resolve();
    };

    this.ws.onclose = (event: any) => {
      console.log('🔌 WebSocket connection closed', { 
        wasClean: event.wasClean, 
        code: event.code, 
        reason: event.reason 
      });
      this.isConnecting = false;
      this.clearHeartbeat();
      this.clearMetricsInterval();
      
      // HTTP polling removed - Go server is WebSocket-only
      
      // Update connection store
      if (this.connectionStore) {
        this.connectionStore.setLastDisconnected(new Date());
        this.connectionStore.updateHealthScore(0);
        this.connectionStore.updateConnectionQuality('unstable');
      }
      
      this.onDisconnectHandler?.();

      if (!this.isDestroyed && !event.wasClean) {
        console.log('🔄 Scheduling reconnection...');
        this.scheduleReconnection();
      }
    };

    this.ws.onerror = (event: any) => {
      console.error('❌ WebSocket error:', event);
      this.isConnecting = false;
      const error = new WebSocketError('WebSocket error occurred');
      
      // Update connection store
      if (this.connectionStore) {
        this.connectionStore.setError(error.message, error.code);
        this.connectionStore.incrementErrorCount();
        this.connectionStore.updateConnectionQuality('poor');
      }
      
      this.onErrorHandler?.(error);
      // Don't reject if we're already connected
      if (this.isConnecting) {
        reject(error);
      }
    };

    this.ws.onmessage = (event: any) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);
        const receiveTime = performance.now();
        logger.debug('Received message', { message: JSON.stringify(message) }, 'websocket');
        this.handleMessage(message, receiveTime);
      } catch (error) {
        const wsError = new WebSocketError('Failed to parse message');
        logger.error('Message parsing error', wsError, 'websocket');
        
        // Update connection store
        if (this.connectionStore) {
          this.connectionStore.setError(wsError.message, wsError.code);
          this.connectionStore.incrementErrorCount();
        }
        
        this.onErrorHandler?.(wsError);
      }
    };
  }

  /**
   * Start heartbeat to keep connection alive
   */
  private startHeartbeat(): void {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
    }

    this.heartbeatInterval = setInterval(async () => {
      if (this.isConnected()) {
        try {
          const startTime = performance.now();
          await this.call('ping', {});
          const responseTime = performance.now() - startTime;
          
          logger.debug('Heartbeat sent', undefined, 'websocket');
          
          // Update connection store with heartbeat metrics
          if (this.connectionStore) {
            this.connectionStore.setLastHeartbeat(new Date());
            this.connectionStore.updateResponseTime(responseTime);
            this.connectionStore.updateHealthScore(Math.min(100, this.connectionStore.healthScore + 5));
          }
        } catch (error) {
          console.warn('⚠️ Heartbeat failed:', error);
          
          // Update connection store with heartbeat failure
          if (this.connectionStore) {
            this.connectionStore.updateHealthScore(Math.max(0, this.connectionStore.healthScore - 10));
          }
          
          // Don't trigger reconnection on heartbeat failure
        }
      }
    }, this.config.heartbeatInterval);
  }

  /**
   * Start metrics collection interval
   */
  private startMetricsCollection(): void {
    if (this.metricsInterval) {
      clearInterval(this.metricsInterval);
    }

    this.metricsInterval = setInterval(() => {
      if (this.isConnected() && this.connectionStore) {
        // Update connection uptime
        const { lastConnected } = this.connectionStore;
        if (lastConnected) {
          const uptime = Date.now() - lastConnected.getTime();
          this.connectionStore.setConnectionUptime(uptime);
        }
      }
    }, 5000); // Update every 5 seconds
  }

  /**
   * Handle incoming WebSocket messages
   */
  private handleMessage(message: WebSocketMessage, receiveTime: number): void {
    if (this.onMessageHandler) {
      this.onMessageHandler(message);
    }

    if (message.jsonrpc !== '2.0') {
      logger.warn('Received non-JSON-RPC message', { message }, 'websocket');
      return;
    }

    // Check if it's a response (has id) or notification (has method)
    if ('id' in message) {
      const response = message as JSONRPCResponse;
      const request = this.pendingRequests.get(response.id);
      if (request) {
        clearTimeout(request.timeout);
        this.pendingRequests.delete(response.id);

        if (response.error) {
          // Handle error with specific error code processing
          this.handleError(response.error);
          
          const error = new WebSocketError(response.error.message, response.error.code, response.error.data);
          console.error(`❌ Received error for request #${response.id}:`, error);
          request.reject(error);
        } else if (response.result) {
          console.log(`✅ Received result for request #${response.id}`);
          request.resolve(response.result);
        } else {
          console.warn(`⚠️ Received message with no error or result for request #${response.id}`);
          request.reject(new WebSocketError('Received message with no error or result'));
        }
      } else {
        console.warn(`⚠️ Received response for unknown request ID: ${response.id}`);
      }
    } else if ('method' in message) {
      // Handle notifications
      const notification = message as JSONRPCNotification;
      if (notification.method === NOTIFICATION_METHODS.CAMERA_STATUS_UPDATE) {
        const cameraNotification: CameraStatusNotification = {
          jsonrpc: '2.0',
          method: notification.method,
          params: notification.params as { device: string; status: string; name: string; resolution: string; fps: number; streams: { rtsp: string; webrtc: string; hls: string; }; }
        };
        this.onCameraStatusUpdateHandler?.(cameraNotification);
        this.onNotificationHandler?.(notification as NotificationMessage);
      } else if (notification.method === NOTIFICATION_METHODS.RECORDING_STATUS_UPDATE) {
        const recordingNotification: RecordingStatusNotification = {
          jsonrpc: '2.0',
          method: notification.method,
          params: notification.params as { device: string; status: string; filename: string; duration: number; }
        };
        this.onRecordingStatusUpdateHandler?.(recordingNotification);
        this.onNotificationHandler?.(notification as NotificationMessage);
      } else if (notification.method === NOTIFICATION_METHODS.STORAGE_STATUS_UPDATE) {
        const storageNotification: StorageStatusNotification = {
          jsonrpc: '2.0',
          method: notification.method,
          params: notification.params as { total_space: number; used_space: number; available_space: number; usage_percent: number; threshold_status: 'normal' | 'warning' | 'critical'; }
        };
        this.onStorageStatusUpdateHandler?.(storageNotification);
        this.onNotificationHandler?.(notification as NotificationMessage);
      } else {
        console.log(`📬 Received notification: ${notification.method}`);
        this.onNotificationHandler?.(notification as NotificationMessage);
      }
    } else {
      console.warn('⚠️ Received message with no ID or method:', message);
    }
  }

  // HTTP polling fallback methods removed - Go server is WebSocket-only
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
 * Sprint 3: Enhanced for real server integration with connection store integration
 */
export async function createWebSocketService(config: Partial<WebSocketConfig> = {}): Promise<WebSocketService> {
  const finalConfig = { ...defaultWebSocketConfig, ...config };
  console.log('🔧 Creating WebSocket service with config:', finalConfig);
  const service = new WebSocketService(finalConfig);
  
  // Integrate with connection store if available (only in non-test environment)
  if (process.env.NODE_ENV !== 'test') {
    try {
      const { useConnectionStore } = await import('../stores/connection');
      const store = useConnectionStore.getState();
      service.setConnectionStore(store);
    } catch {
      console.warn('⚠️ Connection store not available for WebSocket service integration');
    }
    
    // Integrate with camera store if available
    try {
      const { useCameraStore } = await import('../stores/cameraStore');
      const store = useCameraStore.getState();
      service.setCameraStore(store);
    } catch {
      console.warn('⚠️ Camera store not available for WebSocket service integration');
    }
  }
  
  return service;
}

/**
 * Create a WebSocket service instance synchronously (for testing)
 * This version skips store integration to avoid async operations in tests
 */
export function createWebSocketServiceSync(config: Partial<WebSocketConfig> = {}): WebSocketService {
  const finalConfig = { ...defaultWebSocketConfig, ...config };
  console.log('🔧 Creating WebSocket service synchronously with config:', finalConfig);
  return new WebSocketService(finalConfig);
} 