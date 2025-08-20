/**
 * WebSocket JSON-RPC 2.0 Client for MediaMTX Camera Service
 * 
 * Sprint 3: Real Server Integration with Enhanced Connection Management
 * REQ-NET01-003: Polling Fallback Mechanism Implementation
 * 
 * Implements:
 * - Real WebSocket connection to MediaMTX server at ws://localhost:8002/ws
 * - JSON-RPC 2.0 protocol client with full error handling
 * - Exponential backoff for reconnection with production settings
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
  WebSocketConfig,
  JSONRPCRequest,
  JSONRPCResponse,
  JSONRPCNotification,
  WebSocketMessage,
  NotificationMessage,
  CameraStatusNotification,
  RecordingStatusNotification,
} from '../types';
import { authService } from './authService';
import { NOTIFICATION_METHODS } from '../types';
import { HTTPPollingService, HTTPPollingConfig, HTTPPollingError } from './httpPollingService';

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
}

interface CameraStoreInterface {
  handleNotification: (notification: unknown) => void;
  updateCameraStatus?: (device: string, status: import('../types/camera').CameraStatus) => void;
  addRecording?: (device: string, recording: import('../types/camera').RecordingSession) => void;
  removeRecording?: (device: string) => void;
  updateRecordingProgress?: (device: string, progress: number) => void;
  clearRecordingProgress?: (device: string) => void;
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

export class WebSocketService {
  private ws: any | null = null;
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

  // Event handlers
  private onMessageHandler?: (message: WebSocketMessage) => void;
  private onConnectHandler?: () => void;
  private onDisconnectHandler?: () => void;
  private onErrorHandler?: (error: WebSocketError) => void;

  // Enhanced notification handlers
  private onCameraStatusUpdateHandler?: (notification: CameraStatusNotification) => void;
  private onRecordingStatusUpdateHandler?: (notification: RecordingStatusNotification) => void;
  private onNotificationHandler?: (notification: NotificationMessage) => void;

  // Connection store integration
  private connectionStore: ConnectionStoreInterface | null = null;
  private cameraStore: CameraStoreInterface | null = null;

  // Performance tracking
  private notificationCount = 0;
  private lastNotificationTime = 0;
  private notificationLatency: number[] = [];

  // REQ-NET01-003: HTTP Polling Fallback
  private httpPollingService: HTTPPollingService | null = null;
  private fallbackMode = false;
  private fallbackStartTime = 0;

  private config: WebSocketConfig;

  constructor(config: WebSocketConfig) {
    this.config = config;
    
    // Initialize HTTP polling fallback service
    this.initializeHTTPPollingFallback();
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

  public onNotification(handler: (notification: NotificationMessage) => void): void {
    this.onNotificationHandler = handler;
  }

  /**
   * Connect to the MediaMTX WebSocket server
   * Sprint 3: Real server integration with enhanced error handling and metrics
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
        
        // Use Node.js ws library in Node.js environment, browser WebSocket in browser
        if (typeof window === 'undefined') {
          // Node.js environment - use 'ws' library
          const WebSocket = require('ws');
          this.ws = new WebSocket(this.config.url);
        } else {
          // Browser environment - use native WebSocket
          this.ws = new WebSocket(this.config.url);
        }
        
        this.setupEventHandlers(resolve, reject);
      } catch (error) {
        this.isConnecting = false;
        const wsError = new WebSocketError('Failed to create WebSocket connection');
        console.error('❌ WebSocket connection failed:', wsError);
        
        // Update connection store
        if (this.connectionStore) {
          this.connectionStore.setError(wsError.message, wsError.code);
          this.connectionStore.incrementErrorCount();
        }
        
        reject(wsError);
      }
    });
  }

  /**
   * Disconnect from the WebSocket server
   */
  public disconnect(): void {
    console.log('🔌 Disconnecting from MediaMTX server');
    this.clearReconnectTimeout();
    this.clearHeartbeat();
    this.clearMetricsInterval();
    
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }

    // REQ-NET01-003: Start HTTP polling fallback when manually disconnected
    if (this.httpPollingService && !this.httpPollingService.isActive) {
      console.log('🔄 Manual disconnect - starting HTTP polling fallback');
      this.httpPollingService.startPolling();
      this.fallbackMode = true;
      this.fallbackStartTime = Date.now();
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

    // Update connection store
    if (this.connectionStore) {
      this.connectionStore.setLastDisconnected(new Date());
    }
  }

  /**
   * Send a JSON-RPC method call with optional authentication
   * Sprint 3: Enhanced for real server integration with metrics tracking
   * REQ-NET01-003: HTTP polling fallback when WebSocket fails
   */
  public async call(method: string, params: Record<string, unknown> | object = {}, requireAuth: boolean = false): Promise<unknown> {
    if (!this.ws || this.ws.readyState !== 1) { // WebSocket.OPEN = 1
      // REQ-NET01-003: Try HTTP polling fallback for supported methods
      if (this.httpPollingService && this.isFallbackMethodSupported(method)) {
        console.log(`🔄 WebSocket disconnected - using HTTP polling fallback for ${method}`);
        
        try {
          // Start polling if not already active
          if (!this.httpPollingService.isActive) {
            this.httpPollingService.startPolling();
          }
          
          // Use HTTP polling for camera list
          if (method === 'get_camera_list') {
            const result = await this.httpPollingService.getCameraList();
            console.log(`✅ HTTP Polling: ${method} successful`);
            return result;
          }
          
          // For other methods, return a fallback response
          return this.getFallbackResponse(method, params);
          
        } catch (error) {
          console.error(`❌ HTTP Polling fallback failed for ${method}:`, error);
          const fallbackError = new WebSocketError(`HTTP polling fallback failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
          
          // Update connection store
          if (this.connectionStore) {
            this.connectionStore.setError(fallbackError.message, fallbackError.code);
            this.connectionStore.incrementErrorCount();
          }
          
          throw fallbackError;
        }
      }
      
      const error = new WebSocketError('WebSocket not connected');
      
      // Update connection store
      if (this.connectionStore) {
        this.connectionStore.setError(error.message, error.code);
        this.connectionStore.incrementErrorCount();
      }
      
      throw error;
    }

    // Add authentication token if required and available
    let finalParams = params as Record<string, unknown>;
    if (requireAuth) {
      try {
        finalParams = authService.includeAuth(params as Record<string, unknown>);
      } catch (error) {
        const authError = new WebSocketError(`Authentication required: ${error instanceof Error ? error.message : 'Unknown error'}`);
        
        // Update connection store
        if (this.connectionStore) {
          this.connectionStore.setError(authError.message, authError.code);
          this.connectionStore.incrementErrorCount();
        }
        
        throw authError;
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
        
        // Update connection store
        if (this.connectionStore) {
          this.connectionStore.setError(timeoutError.message, timeoutError.code);
          this.connectionStore.incrementErrorCount();
        }
        
        reject(timeoutError);
      }, this.config.requestTimeout);

      this.pendingRequests.set(requestId, { resolve, reject, timeout });

      try {
        this.ws!.send(JSON.stringify(request));
        
        // Update metrics
        if (this.connectionStore) {
          this.connectionStore.incrementMessageCount();
        }
      } catch {
        this.pendingRequests.delete(requestId);
        clearTimeout(timeout);
        const sendError = new WebSocketError('Failed to send request');
        console.error('❌ Failed to send request:', sendError);
        
        // Update connection store
        if (this.connectionStore) {
          this.connectionStore.setError(sendError.message, sendError.code);
          this.connectionStore.incrementErrorCount();
        }
        
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
    return this.ws?.readyState === 1; // WebSocket.OPEN = 1
  }

  public get isConnectingStatus(): boolean {
    return this.isConnecting;
  }

  /**
   * Get notification performance metrics
   */
  public getNotificationMetrics() {
    return {
      count: this.notificationCount,
      averageLatency: this.notificationLatency.length > 0 
        ? this.notificationLatency.reduce((a, b) => a + b, 0) / this.notificationLatency.length 
        : 0,
      lastNotificationTime: this.lastNotificationTime
    };
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
   * Test-only accessor for internal state
   * Only available in test environment
   */
  public getTestState() {
    if (process.env.NODE_ENV === 'test') {
      return {
        ws: this.ws,
        isConnecting: this.isConnecting,
        isDestroyed: this.isDestroyed,
        reconnectAttempts: this.reconnectAttempts,
        pendingRequests: this.pendingRequests,
        connectionStore: this.connectionStore,
        cameraStore: this.cameraStore,
        httpPollingService: this.httpPollingService,
        fallbackMode: this.fallbackMode
      };
    }
    return null;
  }

  /**
   * REQ-NET01-003: Check if method is supported in HTTP polling fallback
   */
  private isFallbackMethodSupported(method: string): boolean {
    const supportedMethods = [
      'get_camera_list',
      'ping',
      'get_camera_status'
    ];
    return supportedMethods.includes(method);
  }

  /**
   * REQ-NET01-003: Get fallback response for methods not fully supported via HTTP
   */
  private getFallbackResponse(method: string, params: Record<string, unknown> | object): unknown {
    switch (method) {
      case 'ping':
        return 'pong';
      
      case 'get_camera_status':
        // Return a basic status indicating fallback mode
        return {
          device: (params as any).device || 'unknown',
          status: 'UNKNOWN',
          name: 'Camera (Fallback Mode)',
          resolution: 'unknown',
          fps: 0,
          streams: {},
          fallback_mode: true,
          message: 'Status unavailable in HTTP polling fallback mode'
        };
      
      default:
        throw new WebSocketError(`Method ${method} not supported in HTTP polling fallback mode`);
    }
  }

  /**
   * REQ-NET01-003: Get fallback mode status
   */
  public get isInFallbackMode(): boolean {
    return this.fallbackMode;
  }

  /**
   * REQ-NET01-003: Get fallback statistics
   */
  public getFallbackStats() {
    return {
      isInFallbackMode: this.fallbackMode,
      fallbackStartTime: this.fallbackStartTime,
      fallbackDuration: this.fallbackStartTime > 0 ? Date.now() - this.fallbackStartTime : 0,
      httpPollingStats: this.httpPollingService?.getPollingStats() || null
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
      if (this.isConnected) {
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
   * Clear heartbeat interval
   */
  private clearHeartbeat(): void {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
  }

  /**
   * Start metrics collection interval
   */
  private startMetricsCollection(): void {
    if (this.metricsInterval) {
      clearInterval(this.metricsInterval);
    }

    this.metricsInterval = setInterval(() => {
      if (this.isConnected && this.connectionStore) {
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
   * Clear metrics collection interval
   */
  private clearMetricsInterval(): void {
    if (this.metricsInterval) {
      clearInterval(this.metricsInterval);
      this.metricsInterval = null;
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
        this.scheduleReconnect();
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
   * Enhanced message handling with notification processing
   */
  private handleMessage(message: WebSocketMessage, receiveTime: number): void {
    // Handle notifications (no id)
    if (!('id' in message)) {
      this.handleNotification(message as JSONRPCNotification, receiveTime);
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
      
      // Update connection store
      if (this.connectionStore) {
        this.connectionStore.setError(error.message, error.code);
        this.connectionStore.incrementErrorCount();
      }
      
      reject(error);
    } else {
      console.log('✅ RPC Response received for request:', response.id);
      
      // Update connection store with successful response
      if (this.connectionStore) {
        this.connectionStore.updateHealthScore(Math.min(100, this.connectionStore.healthScore + 2));
      }
      
      resolve(response.result);
    }
  }

  /**
   * Enhanced notification handling with real-time updates
   */
  private handleNotification(notification: JSONRPCNotification, receiveTime: number): void {
    console.log('📢 Notification received:', notification.method);
    
    // Update notification metrics
    this.notificationCount++;
    this.lastNotificationTime = receiveTime;
    
    // Calculate notification latency (if we have a reference time)
    if (this.lastNotificationTime > 0) {
      const latency = receiveTime - this.lastNotificationTime;
      this.notificationLatency.push(latency);
      
      // Keep only last 100 latency measurements for performance
      if (this.notificationLatency.length > 100) {
        this.notificationLatency.shift();
      }
    }

    // Handle specific notification types
    switch (notification.method) {
      case NOTIFICATION_METHODS.CAMERA_STATUS_UPDATE:
        this.handleCameraStatusUpdate(notification as CameraStatusNotification);
        break;
        
      case NOTIFICATION_METHODS.RECORDING_STATUS_UPDATE:
        this.handleRecordingStatusUpdate(notification as RecordingStatusNotification);
        break;
        
      default:
        console.warn('⚠️ Unknown notification method:', notification.method);
        break;
    }

    // Call general notification handler
    this.onNotificationHandler?.(notification as NotificationMessage);
    
    // Call legacy message handler for backward compatibility
    this.onMessageHandler?.(notification);
  }

  /**
   * Handle camera status update notifications
   */
  private handleCameraStatusUpdate(notification: CameraStatusNotification): void {
    console.log('📹 Camera status update:', notification.params);
    
    // Call specific handler
    this.onCameraStatusUpdateHandler?.(notification);
    
    // Update camera store if available
    if (this.cameraStore && this.cameraStore.handleNotification) {
      try {
        this.cameraStore.handleNotification(notification);
      } catch (error) {
        console.error('❌ Error updating camera store:', error);
      }
    }
  }

  /**
   * Handle recording status update notifications
   */
  private handleRecordingStatusUpdate(notification: RecordingStatusNotification): void {
    console.log('🎥 Recording status update:', notification.params);
    
    // Call specific handler
    this.onRecordingStatusUpdateHandler?.(notification);
    
    // Update camera store if available
    if (this.cameraStore && this.cameraStore.handleNotification) {
      try {
        this.cameraStore.handleNotification(notification);
      } catch (error) {
        console.error('❌ Error updating camera store:', error);
      }
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= this.config.maxReconnectAttempts) {
      const error = new WebSocketError('Max reconnection attempts reached');
      console.error('❌ Max reconnection attempts reached');
      
      // Update connection store
      if (this.connectionStore) {
        this.connectionStore.setError(error.message, error.code);
        this.connectionStore.updateConnectionQuality('unstable');
      }
      
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
        
        // Update connection store
        if (this.connectionStore) {
          this.connectionStore.setError(error.message, error.code);
          this.connectionStore.incrementErrorCount();
        }
        
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
 * Sprint 3: Enhanced for real server integration with connection store integration
 */
export async function createWebSocketService(config: Partial<WebSocketConfig> = {}): Promise<WebSocketService> {
  const finalConfig = { ...defaultWebSocketConfig, ...config };
  console.log('🔧 Creating WebSocket service with config:', finalConfig);
  const service = new WebSocketService(finalConfig);
  
  // Integrate with connection store if available (only in non-test environment)
  if (process.env.NODE_ENV !== 'test') {
    try {
      const { useConnectionStore } = await import('../stores/connectionStore');
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