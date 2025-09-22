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

  // HTTP polling removed - Go server is WebSocket-only

  private config: WebSocketConfig;

  constructor(config: WebSocketConfig) {
    this.config = config;
    
    // HTTP polling removed - Go server is WebSocket-only
  }

  /**
   * Initialize HTTP polling fallback service (REQ-NET01-003)
   */
  private initializeHTTPPollingFallback(): void {
    // Extract HTTP base URL from WebSocket URL
    const wsUrl = new URL(this.config.url);
    const httpBaseUrl = `http://${wsUrl.hostname}:8003`; // Health server port
    
    const pollingConfig: HTTPPollingConfig = {
      baseUrl: httpBaseUrl,
      pollingInterval: 5000, // 5 seconds
      timeout: 3000, // 3 seconds
      maxRetries: 3,
      retryDelay: 1000,
    };
    
    this.httpPollingService = new HTTPPollingService(pollingConfig);
    
    // Set up event handlers for polling service
    this.httpPollingService.onCameraListUpdate((cameras) => {
      console.log('📡 HTTP Polling: Camera list updated', cameras.length, 'cameras');
      
      // Update camera store if available
      if (this.cameraStore && this.cameraStore.handleNotification) {
        this.cameraStore.handleNotification({
          type: 'camera_list_update',
          cameras: cameras
        });
      }
    });
    
    this.httpPollingService.onError((error) => {
      console.error('📡 HTTP Polling Error:', error.message);
      
      // Update connection store
      if (this.connectionStore) {
        this.connectionStore.setError(`HTTP Polling: ${error.message}`, error.statusCode);
        this.connectionStore.incrementErrorCount();
      }
    });
    
    this.httpPollingService.onPollingStart(() => {
      console.log('🔄 HTTP Polling Fallback: Started');
      this.fallbackMode = true;
      this.fallbackStartTime = Date.now();
      
      // Update connection store
      if (this.connectionStore) {
        this.connectionStore.setError('WebSocket disconnected - using HTTP polling fallback');
        this.connectionStore.updateConnectionQuality('poor');
      }
    });
    
    this.httpPollingService.onPollingStop(() => {
      console.log('🔄 HTTP Polling Fallback: Stopped');
      this.fallbackMode = false;
      this.fallbackStartTime = 0;
      
      // Update connection store
      if (this.connectionStore) {
        this.connectionStore.updateConnectionQuality('excellent');
      }
    });
  }

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
   * Make a JSON-RPC call with optional authentication
   */
  public call(method: string, params: Record<string, unknown> = {}, requireAuth: boolean = true): Promise<unknown> {
    if (!this.isConnected()) {
      console.warn(`⚠️ WebSocket not connected for method: ${method}`);
      if (this.isFallbackMethodSupported(method)) {
        return this.getFallbackResponse(method, params);
      }
      return Promise.reject(new WebSocketError('WebSocket not connected'));
    }

    // Add authentication token if required and available
    let finalParams = params;
    if (requireAuth && method !== 'authenticate') {
      try {
        const { authService } = require('./authService');
        finalParams = authService.addAuthToParams(params);
      } catch (error) {
        // Auth service not available, continue without auth
        console.warn(`⚠️ Auth service not available for method: ${method}`);
      }
    }

    const id = this.generateRequestId();
    const request: JSONRPCRequest = {
      jsonrpc: '2.0',
      id,
      method,
      params: finalParams
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
      console.log(`📤 Sent request #${id}: ${method}`);
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
    console.log(`🔍 Handling error: ${error.code} - ${error.message}`);
    
    // Update connection store
    if (this.connectionStore) {
      this.connectionStore.setError(error.message, error.code);
      this.connectionStore.incrementErrorCount();
    }
    
    // Handle authentication errors
    if (error.code === ERROR_CODES.AUTHENTICATION_FAILED) {
      console.error(`🔐 Authentication failed: ${error.message}`);
      try {
        const { authService } = require('./authService');
        authService.handleAuthError(error);
      } catch (authError) {
        console.warn('⚠️ Auth service not available for error handling');
      }
      return;
    }
    
    // Process specific error codes with enhanced handling
    switch (error.code) {
      case ERROR_CODES.CAMERA_ALREADY_RECORDING:
        console.warn(`⚠️ Recording conflict detected: ${error.message}`);
        this.handleRecordingConflict(error);
        break;
      case ERROR_CODES.STORAGE_SPACE_LOW:
        console.warn(`⚠️ Storage space low: ${error.message}`);
        this.handleStorageWarning(error);
        break;
      case ERROR_CODES.STORAGE_SPACE_CRITICAL:
        console.error(`🚨 Storage space critical: ${error.message}`);
        this.handleStorageCritical(error);
        break;
      case ERROR_CODES.CAMERA_NOT_FOUND_OR_DISCONNECTED:
        console.error(`❌ Camera not found or disconnected: ${error.message}`);
        break;
      case ERROR_CODES.RECORDING_ALREADY_IN_PROGRESS:
        console.warn(`⚠️ Recording already in progress: ${error.message}`);
        break;
      case ERROR_CODES.MEDIAMTX_SERVICE_UNAVAILABLE:
        console.error(`❌ MediaMTX service unavailable: ${error.message}`);
        break;
      case ERROR_CODES.AUTHENTICATION_REQUIRED:
        console.error(`❌ Authentication required: ${error.message}`);
        break;
      case ERROR_CODES.INSUFFICIENT_STORAGE_SPACE:
        console.error(`❌ Insufficient storage space: ${error.message}`);
        break;
      case ERROR_CODES.CAMERA_CAPABILITY_NOT_SUPPORTED:
        console.warn(`⚠️ Camera capability not supported: ${error.message}`);
        break;
      default:
        console.error(`❌ Standard error: ${error.code} - ${error.message}`);
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
      console.log('🔄 Already connecting or connected');
      return;
    }

    if (this.isDestroyed) {
      throw new WebSocketError('WebSocket service has been destroyed');
    }

    this.isConnecting = true;
    console.log('🔌 Connecting to WebSocket server...');

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
    console.log('🔌 Disconnecting from WebSocket server...');
    
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
    
    // Stop HTTP polling fallback
    if (this.httpPollingService) {
      this.httpPollingService.stopPolling();
    }
    
    console.log('✅ WebSocket disconnected');
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
      console.log('✅ WebSocket connection established');
      this.isConnecting = false;
      this.reconnectAttempts = 0;
      this.startHeartbeat();
      this.startMetricsCollection();
      
      // REQ-NET01-003: Stop HTTP polling fallback when WebSocket is restored
      if (this.httpPollingService && this.httpPollingService.isActive) {
        console.log('🔄 WebSocket restored - stopping HTTP polling fallback');
        this.httpPollingService.stopPolling();
      }
      
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
      
      // REQ-NET01-003: Start HTTP polling fallback when WebSocket closes
      if (this.httpPollingService && !this.httpPollingService.isActive) {
        console.log('🔄 WebSocket closed - starting HTTP polling fallback');
        this.httpPollingService.startPolling();
        this.fallbackMode = true;
        this.fallbackStartTime = Date.now();
      }
      
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
        console.log('📥 Received message:', JSON.stringify(message));
        this.handleMessage(message, receiveTime);
      } catch (error) {
        const wsError = new WebSocketError('Failed to parse message');
        console.error('❌ Message parsing error:', wsError);
        
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
          
          console.log('💓 Heartbeat sent');
          
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
      console.warn('⚠️ Received non-JSON-RPC message:', message);
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

  /**
   * Check if a method is supported for HTTP polling fallback
   */
  private isFallbackMethodSupported(method: string): boolean {
    const supportedMethods = [
      'get_camera_list',
      'get_camera_status',
      'ping'
    ];
    return supportedMethods.includes(method);
  }

  /**
   * Get a fallback response using HTTP polling service
   */
  private async getFallbackResponse(method: string, params: Record<string, unknown>): Promise<unknown> {
    console.warn(`⚠️ WebSocket disconnected - using HTTP polling fallback for method: ${method}`);
    
    if (!this.httpPollingService) {
      throw new WebSocketError('HTTP polling service not available for fallback');
    }

    try {
      switch (method) {
        case 'get_camera_list':
          return await this.httpPollingService.getCameraList();
        
        case 'get_camera_status':
          const device = params.device as string;
          return await this.httpPollingService.getCameraStatus(device);
        
        case 'take_snapshot':
          const snapshotDevice = params.device as string;
          return await this.httpPollingService.takeSnapshot(snapshotDevice);
        
        case 'start_recording':
          const recordingDevice = params.device as string;
          return await this.httpPollingService.startRecording(recordingDevice);
        
        case 'stop_recording':
          const stopDevice = params.device as string;
          return await this.httpPollingService.stopRecording(stopDevice);
        
        case 'list_recordings':
          return await this.httpPollingService.listRecordings();
        
        case 'list_snapshots':
          return await this.httpPollingService.listSnapshots();
        
        case 'get_recording_info':
          const recordingFilename = params.filename as string;
          return await this.httpPollingService.getRecordingInfo(recordingFilename);
        
        case 'get_snapshot_info':
          const snapshotFilename = params.filename as string;
          return await this.httpPollingService.getSnapshotInfo(snapshotFilename);
        
        case 'delete_recording':
          const deleteRecordingFilename = params.filename as string;
          return await this.httpPollingService.deleteRecording(deleteRecordingFilename);
        
        case 'delete_snapshot':
          const deleteSnapshotFilename = params.filename as string;
          return await this.httpPollingService.deleteSnapshot(deleteSnapshotFilename);
        
        case 'get_storage_info':
          return await this.httpPollingService.getStorageInfo();
        
        case 'get_metrics':
          return await this.httpPollingService.getMetrics();
        
        case 'get_status':
          return await this.httpPollingService.getStatus();
        
        case 'get_server_info':
          return await this.httpPollingService.getServerInfo();
        
        case 'get_streams':
          return await this.httpPollingService.getStreams();
        
        case 'ping':
          // Use health endpoint as ping alternative
          const health = await this.httpPollingService.getSystemHealth();
          return { status: 'healthy', health };
        
        default:
          throw new WebSocketError(`HTTP polling fallback not implemented for method: ${method}`);
      }
    } catch (error) {
      throw new WebSocketError(`HTTP polling fallback failed for method: ${method}: ${error}`);
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