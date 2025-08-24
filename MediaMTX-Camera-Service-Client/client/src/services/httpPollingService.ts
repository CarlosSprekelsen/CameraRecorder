/**
 * HTTP Polling Service for MediaMTX Camera Service
 * 
 * REQ-NET01-003: Polling Fallback Mechanism
 * 
 * Provides HTTP-based polling fallback when WebSocket connection fails.
 * Implements real integration with MediaMTX server health endpoints.
 * 
 * Architecture:
 * - Primary: WebSocket JSON-RPC (real-time, efficient)
 * - Fallback: HTTP polling to /api/cameras endpoint (reliable, slower)
 * - Automatic switch back to WebSocket when connection restored
 * 
 * Following "Test First, Real Integration Always" philosophy
 * No mocking - uses real server endpoints
 */

import type {
  CameraDevice, 
  CameraListResponse, 
  FileListResponse,
  FileItem,
  SnapshotResult,
  RecordingSession,
  StreamInfo,
  ServerInfo,
  StorageInfo
} from '../types/camera';
import type {
  JSONRPCError,
  RecordingConflictErrorData,
  StorageErrorData
} from '../types/rpc';
import { authService } from './authService';
import { ERROR_CODES } from '../types/rpc';

export interface HTTPPollingConfig {
  baseUrl: string;
  pollingInterval: number;
  timeout: number;
  maxRetries: number;
  retryDelay: number;
}

export class HTTPPollingError extends Error {
  public statusCode?: number;
  public response?: any;

  constructor(message: string, statusCode?: number, response?: any) {
    super(message);
    this.name = 'HTTPPollingError';
    this.statusCode = statusCode;
    this.response = response;
  }
}

export class HTTPPollingService {
  private config: HTTPPollingConfig;
  private isPolling = false;
  private pollingInterval: NodeJS.Timeout | null = null;
  private lastPollTime = 0;
  private pollCount = 0;
  private errorCount = 0;

  // Event handlers
  private onCameraListUpdateHandler?: (cameras: CameraDevice[]) => void;
  private onErrorHandler?: (error: HTTPPollingError) => void;
  private onPollingStartHandler?: () => void;
  private onPollingStopHandler?: () => void;

  constructor(config: HTTPPollingConfig) {
    this.config = config;
  }

  /**
   * Start HTTP polling fallback
   */
  public startPolling(): void {
    if (this.isPolling) {
      console.log('🔄 HTTP polling already active');
      return;
    }

    console.log('🔄 Starting HTTP polling fallback');
    this.isPolling = true;
    this.pollCount = 0;
    this.errorCount = 0;

    if (this.onPollingStartHandler) {
      this.onPollingStartHandler();
    }

    // Start immediate poll
    this.performPoll();

    // Set up polling interval
    this.pollingInterval = setInterval(() => {
      this.performPoll();
    }, this.config.pollingInterval);
  }

  /**
   * Stop HTTP polling fallback
   */
  public stopPolling(): void {
    if (!this.isPolling) {
      return;
    }

    console.log('🔄 Stopping HTTP polling fallback');
    this.isPolling = false;

    if (this.pollingInterval) {
      clearInterval(this.pollingInterval);
      this.pollingInterval = null;
    }

    if (this.onPollingStopHandler) {
      this.onPollingStopHandler();
    }
  }

  /**
   * Perform a single HTTP poll to get camera list
   */
  public async performPoll(): Promise<CameraListResponse | null> {
    if (!this.isPolling) {
      return null;
    }

    const startTime = performance.now();
    this.pollCount++;

    try {
      console.log(`📡 HTTP Poll #${this.pollCount} - Getting camera list`);
      
      const response = await this.fetchWithTimeout(
        `${this.config.baseUrl}/api/cameras`,
        {
          method: 'GET',
          headers: this.getAuthHeaders(),
        },
        this.config.timeout
      );

      if (!response.ok) {
        throw new HTTPPollingError(
          `HTTP ${response.status}: ${response.statusText}`,
          response.status
        );
      }

      const data = await response.json();
      const responseTime = performance.now() - startTime;

      console.log(`✅ HTTP Poll #${this.pollCount} successful (${responseTime.toFixed(2)}ms)`);
      
      this.lastPollTime = Date.now();
      this.errorCount = 0;

      // Parse camera list response
      const cameraList = this.parseCameraListResponse(data);

      // Notify listeners
      if (this.onCameraListUpdateHandler && cameraList.cameras) {
        this.onCameraListUpdateHandler(cameraList.cameras);
      }

      return cameraList;

    } catch (error) {
      this.errorCount++;
      const responseTime = performance.now() - startTime;
      
      console.error(`❌ HTTP Poll #${this.pollCount} failed (${responseTime.toFixed(2)}ms):`, error);

      if (this.onErrorHandler) {
        this.onErrorHandler(error as HTTPPollingError);
      }

      // Stop polling if too many consecutive errors
      if (this.errorCount >= this.config.maxRetries) {
        console.error(`🛑 Stopping HTTP polling after ${this.errorCount} consecutive errors`);
        this.stopPolling();
      }

      return null;
    }
  }

  // ============================================================================
  // JSON-RPC METHOD FALLBACKS - FULL IMPLEMENTATION
  // ============================================================================

  /**
   * ping - Health check method
   */
  public async ping(): Promise<{ pong: string }> {
    const response = await this.fetchWithTimeout(
      `${this.config.baseUrl}/api/ping`,
      {
        method: 'GET',
        headers: this.getAuthHeaders(),
      },
      this.config.timeout
    );

    if (!response.ok) {
      throw new HTTPPollingError(
        `HTTP ${response.status}: ${response.statusText}`,
        response.status
      );
    }

    const data = await response.json();
    return { pong: data.pong || 'pong' };
  }

  /**
   * get_camera_list - Get list of available cameras
   */
  public async getCameraList(): Promise<CameraListResponse> {
    const response = await this.fetchWithTimeout(
      `${this.config.baseUrl}/api/cameras`,
      {
        method: 'GET',
        headers: this.getAuthHeaders(),
      },
      this.config.timeout
    );

    if (!response.ok) {
      throw new HTTPPollingError(
        `HTTP ${response.status}: ${response.statusText}`,
        response.status
      );
    }

    const data = await response.json();
    return this.parseCameraListResponse(data);
  }

  /**
   * get_camera_status - Get status of specific camera
   */
  public async getCameraStatus(deviceId: string): Promise<CameraDevice> {
    const response = await this.fetchWithTimeout(
      `${this.config.baseUrl}/api/cameras/${encodeURIComponent(deviceId)}/status`,
      {
        method: 'GET',
        headers: this.getAuthHeaders(),
      },
      this.config.timeout
    );

    if (!response.ok) {
      throw new HTTPPollingError(
        `HTTP ${response.status}: ${response.statusText}`,
        response.status
      );
    }

    const data = await response.json();
    return {
      device: deviceId,
      status: data.status || 'DISCONNECTED',
      name: data.name || deviceId,
      resolution: data.resolution || 'unknown',
      fps: data.fps || 0,
      streams: data.streams || { rtsp: '', webrtc: '', hls: '' },
      metrics: data.metrics,
      capabilities: data.capabilities
    };
  }

  /**
   * take_snapshot - Take snapshot from camera
   */
  public async takeSnapshot(deviceId: string): Promise<SnapshotResult> {
    const response = await this.fetchWithTimeout(
      `${this.config.baseUrl}/api/cameras/${encodeURIComponent(deviceId)}/snapshot`,
      {
        method: 'POST',
        headers: this.getAuthHeaders(),
      },
      this.config.timeout
    );

    if (!response.ok) {
      throw new HTTPPollingError(
        `HTTP ${response.status}: ${response.statusText}`,
        response.status
      );
    }

    const data = await response.json();
    return {
      device: deviceId,
      filename: data.filename || '',
      status: data.status === 'completed' ? 'completed' : 'FAILED',
      timestamp: new Date().toISOString(),
      file_size: data.file_size || 0,
      format: data.format || 'jpg',
      quality: data.quality || 85,
      error: data.error || undefined
    };
  }

  /**
   * start_recording - Start recording from camera
   */
  public async startRecording(deviceId: string): Promise<RecordingSession> {
    const response = await this.fetchWithTimeout(
      `${this.config.baseUrl}/api/cameras/${encodeURIComponent(deviceId)}/recording/start`,
      {
        method: 'POST',
        headers: this.getAuthHeaders(),
      },
      this.config.timeout
    );

    const data = await this.handleHTTPResponse(response, 'start_recording');
    return {
      device: deviceId,
      session_id: data.session_id || '',
      filename: data.filename || '',
      status: data.status || 'STARTED',
      start_time: new Date().toISOString(),
      format: data.format || 'mp4'
    };
  }

  /**
   * stop_recording - Stop recording from camera
   */
  public async stopRecording(deviceId: string): Promise<RecordingSession> {
    const response = await this.fetchWithTimeout(
      `${this.config.baseUrl}/api/cameras/${encodeURIComponent(deviceId)}/recording/stop`,
      {
        method: 'POST',
        headers: this.getAuthHeaders(),
      },
      this.config.timeout
    );

    const data = await this.handleHTTPResponse(response, 'stop_recording');
    return {
      device: deviceId,
      session_id: data.session_id || '',
      filename: data.filename || '',
      status: data.status || 'STOPPED',
      start_time: data.start_time || new Date().toISOString(),
      end_time: new Date().toISOString(),
      duration: data.duration || 0,
      format: data.format || 'mp4',
      file_size: data.file_size
    };
  }

  /**
   * list_recordings - Get list of recordings
   */
  public async listRecordings(): Promise<FileListResponse> {
    const response = await this.fetchWithTimeout(
      `${this.config.baseUrl}/api/files/recordings`,
      {
        method: 'GET',
        headers: this.getAuthHeaders(),
      },
      this.config.timeout
    );

    if (!response.ok) {
      throw new HTTPPollingError(
        `HTTP ${response.status}: ${response.statusText}`,
        response.status
      );
    }

    const data = await response.json();
    return {
      files: data.files || [],
      total: data.total || 0,
      limit: data.limit || 100,
      offset: data.offset || 0
    };
  }

  /**
   * list_snapshots - Get list of snapshots
   */
  public async listSnapshots(): Promise<FileListResponse> {
    const response = await this.fetchWithTimeout(
      `${this.config.baseUrl}/api/files/snapshots`,
      {
        method: 'GET',
        headers: this.getAuthHeaders(),
      },
      this.config.timeout
    );

    if (!response.ok) {
      throw new HTTPPollingError(
        `HTTP ${response.status}: ${response.statusText}`,
        response.status
      );
    }

    const data = await response.json();
    return {
      files: data.files || [],
      total: data.total || 0,
      limit: data.limit || 100,
      offset: data.offset || 0
    };
  }

  /**
   * get_recording_info - Get information about specific recording
   */
  public async getRecordingInfo(filename: string): Promise<FileItem> {
    const response = await this.fetchWithTimeout(
      `${this.config.baseUrl}/api/files/recordings/${encodeURIComponent(filename)}/info`,
      {
        method: 'GET',
        headers: this.getAuthHeaders(),
      },
      this.config.timeout
    );

    if (!response.ok) {
      throw new HTTPPollingError(
        `HTTP ${response.status}: ${response.statusText}`,
        response.status
      );
    }

    const data = await response.json();
    return {
      filename: data.filename || filename,
      file_size: data.file_size || 0,
      created_time: data.created_time || new Date().toISOString(),
      modified_time: data.modified_time || new Date().toISOString(),
      download_url: data.download_url || '',
      duration: data.duration || 0,
      format: data.format || 'mp4'
    };
  }

  /**
   * get_snapshot_info - Get information about specific snapshot
   */
  public async getSnapshotInfo(filename: string): Promise<FileItem> {
    const response = await this.fetchWithTimeout(
      `${this.config.baseUrl}/api/files/snapshots/${encodeURIComponent(filename)}/info`,
      {
        method: 'GET',
        headers: this.getAuthHeaders(),
      },
      this.config.timeout
    );

    if (!response.ok) {
      throw new HTTPPollingError(
        `HTTP ${response.status}: ${response.statusText}`,
        response.status
      );
    }

    const data = await response.json();
    return {
      filename: data.filename || filename,
      file_size: data.file_size || 0,
      created_time: data.created_time || new Date().toISOString(),
      modified_time: data.modified_time || new Date().toISOString(),
      download_url: data.download_url || '',
      format: data.format || 'jpg'
    };
  }

  /**
   * delete_recording - Delete specific recording
   */
  public async deleteRecording(filename: string): Promise<{ status: string; error?: string }> {
    const response = await this.fetchWithTimeout(
      `${this.config.baseUrl}/api/files/recordings/${encodeURIComponent(filename)}`,
      {
        method: 'DELETE',
        headers: this.getAuthHeaders(),
      },
      this.config.timeout
    );

    if (!response.ok) {
      throw new HTTPPollingError(
        `HTTP ${response.status}: ${response.statusText}`,
        response.status
      );
    }

    const data = await response.json();
    return {
      status: data.status || 'deleted',
      error: data.error || undefined
    };
  }

  /**
   * delete_snapshot - Delete specific snapshot
   */
  public async deleteSnapshot(filename: string): Promise<{ status: string; error?: string }> {
    const response = await this.fetchWithTimeout(
      `${this.config.baseUrl}/api/files/snapshots/${encodeURIComponent(filename)}`,
      {
        method: 'DELETE',
        headers: this.getAuthHeaders(),
      },
      this.config.timeout
    );

    if (!response.ok) {
      throw new HTTPPollingError(
        `HTTP ${response.status}: ${response.statusText}`,
        response.status
      );
    }

    const data = await response.json();
    return {
      status: data.status || 'deleted',
      error: data.error || undefined
    };
  }

  /**
   * get_storage_info - Get storage information
   */
  public async getStorageInfo(): Promise<{
    total_space: number;
    used_space: number;
    available_space: number;
    usage_percent: number;
    threshold_status: 'normal' | 'warning' | 'critical';
  }> {
    const response = await this.fetchWithTimeout(
      `${this.config.baseUrl}/api/storage`,
      {
        method: 'GET',
        headers: this.getAuthHeaders(),
      },
      this.config.timeout
    );

    const data = await this.handleHTTPResponse(response, 'get_storage_info');
    
    // Calculate used space and threshold status
    const total_space = data.total_space || 0;
    const available_space = data.available_space || 0;
    const used_space = total_space - available_space;
    const usage_percent = total_space > 0 ? (used_space / total_space) * 100 : 0;
    
    // Determine threshold status based on usage percentage
    let threshold_status: 'normal' | 'warning' | 'critical' = 'normal';
    if (usage_percent >= 90) {
      threshold_status = 'critical';
    } else if (usage_percent >= 80) {
      threshold_status = 'warning';
    }
    
    return {
      total_space,
      used_space,
      available_space,
      usage_percent,
      threshold_status
    };
  }

  /**
   * get_metrics - Get system metrics
   */
  public async getMetrics(): Promise<{
    active_connections: number;
    total_requests: number;
    average_response_time: number;
    error_rate: number;
    cpu_usage: number;
    memory_usage: number;
  }> {
    const response = await this.fetchWithTimeout(
      `${this.config.baseUrl}/api/metrics`,
      {
        method: 'GET',
        headers: this.getAuthHeaders(),
      },
      this.config.timeout
    );

    if (!response.ok) {
      throw new HTTPPollingError(
        `HTTP ${response.status}: ${response.statusText}`,
        response.status
      );
    }

    const data = await response.json();
    return {
      active_connections: data.active_connections || 0,
      total_requests: data.total_requests || 0,
      average_response_time: data.average_response_time || 0,
      error_rate: data.error_rate || 0,
      cpu_usage: data.cpu_usage || 0,
      memory_usage: data.memory_usage || 0
    };
  }

  /**
   * get_status - Get system status
   */
  public async getStatus(): Promise<{
    status: string;
    uptime: number;
    version: string;
    components: Record<string, string>;
  }> {
    const response = await this.fetchWithTimeout(
      `${this.config.baseUrl}/api/status`,
      {
        method: 'GET',
        headers: this.getAuthHeaders(),
      },
      this.config.timeout
    );

    if (!response.ok) {
      throw new HTTPPollingError(
        `HTTP ${response.status}: ${response.statusText}`,
        response.status
      );
    }

    const data = await response.json();
    return {
      status: data.status || 'unknown',
      uptime: data.uptime || 0,
      version: data.version || 'unknown',
      components: data.components || {}
    };
  }

  /**
   * get_server_info - Get server information
   */
  public async getServerInfo(): Promise<ServerInfo> {
    const response = await this.fetchWithTimeout(
      `${this.config.baseUrl}/api/server/info`,
      {
        method: 'GET',
        headers: this.getAuthHeaders(),
      },
      this.config.timeout
    );

    if (!response.ok) {
      throw new HTTPPollingError(
        `HTTP ${response.status}: ${response.statusText}`,
        response.status
      );
    }

    const data = await response.json();
    return {
      version: data.version || 'unknown',
      uptime: data.uptime || 0,
      cameras_connected: data.cameras_connected || 0,
      total_recordings: data.total_recordings || 0,
      total_snapshots: data.total_snapshots || 0
    };
  }

  /**
   * get_streams - Get MediaMTX stream information
   */
  public async getStreams(): Promise<StreamInfo[]> {
    const response = await this.fetchWithTimeout(
      `${this.config.baseUrl}/api/streams`,
      {
        method: 'GET',
        headers: this.getAuthHeaders(),
      },
      this.config.timeout
    );

    if (!response.ok) {
      throw new HTTPPollingError(
        `HTTP ${response.status}: ${response.statusText}`,
        response.status
      );
    }

    const data = await response.json();
    return data.streams || [];
  }

  /**
   * Get system health status
   */
  public async getSystemHealth(): Promise<any> {
    const response = await this.fetchWithTimeout(
      `${this.config.baseUrl}/health/system`,
      {
        method: 'GET',
        headers: this.getAuthHeaders(),
      },
      this.config.timeout
    );

    if (!response.ok) {
      throw new HTTPPollingError(
        `HTTP ${response.status}: ${response.statusText}`,
        response.status
      );
    }

    return await response.json();
  }

  /**
   * Check if polling is active
   */
  public get isActive(): boolean {
    return this.isPolling;
  }

  /**
   * Get polling statistics
   */
  public getPollingStats() {
    return {
      isActive: this.isPolling,
      pollCount: this.pollCount,
      errorCount: this.errorCount,
      lastPollTime: this.lastPollTime,
      successRate: this.pollCount > 0 ? ((this.pollCount - this.errorCount) / this.pollCount) * 100 : 0
    };
  }

  /**
   * Event handlers
   */
  public onCameraListUpdate(handler: (cameras: CameraDevice[]) => void): void {
    this.onCameraListUpdateHandler = handler;
  }

  public onError(handler: (error: HTTPPollingError) => void): void {
    this.onErrorHandler = handler;
  }

  /**
   * Handle error with specific error code processing
   */
  private handleError(error: JSONRPCError): void {
    console.log(`🔍 HTTP Polling: Handling error: ${error.code} - ${error.message}`);
    
    // Process specific error codes with enhanced handling
    switch (error.code) {
      case ERROR_CODES.CAMERA_ALREADY_RECORDING:
        console.warn(`⚠️ HTTP Polling: Recording conflict detected: ${error.message}`);
        this.handleRecordingConflict(error);
        break;
      case ERROR_CODES.STORAGE_SPACE_LOW:
        console.warn(`⚠️ HTTP Polling: Storage space low: ${error.message}`);
        this.handleStorageWarning(error);
        break;
      case ERROR_CODES.STORAGE_SPACE_CRITICAL:
        console.error(`🚨 HTTP Polling: Storage space critical: ${error.message}`);
        this.handleStorageCritical(error);
        break;
      case ERROR_CODES.CAMERA_NOT_FOUND_OR_DISCONNECTED:
        console.error(`❌ HTTP Polling: Camera not found or disconnected: ${error.message}`);
        break;
      case ERROR_CODES.RECORDING_ALREADY_IN_PROGRESS:
        console.warn(`⚠️ HTTP Polling: Recording already in progress: ${error.message}`);
        break;
      case ERROR_CODES.MEDIAMTX_SERVICE_UNAVAILABLE:
        console.error(`❌ HTTP Polling: MediaMTX service unavailable: ${error.message}`);
        break;
      case ERROR_CODES.AUTHENTICATION_REQUIRED:
        console.error(`❌ HTTP Polling: Authentication required: ${error.message}`);
        break;
      case ERROR_CODES.INSUFFICIENT_STORAGE_SPACE:
        console.error(`❌ HTTP Polling: Insufficient storage space: ${error.message}`);
        break;
      case ERROR_CODES.CAMERA_CAPABILITY_NOT_SUPPORTED:
        console.warn(`⚠️ HTTP Polling: Camera capability not supported: ${error.message}`);
        break;
      default:
        console.error(`❌ HTTP Polling: Standard error: ${error.code} - ${error.message}`);
    }
    
    // Notify error handler about the error
    this.onErrorHandler?.(new HTTPPollingError(error.message, error.code, error.data));
  }

  /**
   * Handle recording conflict errors
   */
  private handleRecordingConflict(error: JSONRPCError): void {
    const conflictData = error.data as RecordingConflictErrorData;
    console.warn(`⚠️ HTTP Polling: Recording conflict for camera: ${conflictData?.camera_id || 'unknown'}`);
    console.warn(`⚠️ HTTP Polling: Active session: ${conflictData?.session_id || 'unknown'}`);
    
    // Notify error handler about the conflict
    this.onErrorHandler?.(new HTTPPollingError(
      `Camera ${conflictData?.camera_id || 'unknown'} is currently recording (Session: ${conflictData?.session_id || 'unknown'})`,
      ERROR_CODES.CAMERA_ALREADY_RECORDING,
      conflictData
    ));
  }

  /**
   * Handle storage warning errors
   */
  private handleStorageWarning(error: JSONRPCError): void {
    const storageData = error.data as StorageErrorData;
    const usagePercent = storageData?.usage_percent || 0;
    const availableSpace = storageData?.available_space || 0;
    
    console.warn(`⚠️ HTTP Polling: Storage space low: ${usagePercent.toFixed(1)}% used`);
    console.warn(`⚠️ HTTP Polling: Available space: ${this.formatBytes(availableSpace)}`);
    
    // Notify error handler about storage warning
    this.onErrorHandler?.(new HTTPPollingError(
      `Storage space is low (${usagePercent.toFixed(1)}% used). Available: ${this.formatBytes(availableSpace)}`,
      ERROR_CODES.STORAGE_SPACE_LOW,
      storageData
    ));
  }

  /**
   * Handle storage critical errors
   */
  private handleStorageCritical(error: JSONRPCError): void {
    const storageData = error.data as StorageErrorData;
    const usagePercent = storageData?.usage_percent || 0;
    const availableSpace = storageData?.available_space || 0;
    
    console.error(`🚨 HTTP Polling: Storage space critical: ${usagePercent.toFixed(1)}% used`);
    console.error(`🚨 HTTP Polling: Available space: ${this.formatBytes(availableSpace)}`);
    
    // Notify error handler about critical storage
    this.onErrorHandler?.(new HTTPPollingError(
      `Storage space is critical (${usagePercent.toFixed(1)}% used). Available: ${this.formatBytes(availableSpace)}`,
      ERROR_CODES.STORAGE_SPACE_CRITICAL,
      storageData
    ));
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
   * Handle HTTP response with error checking
   */
  private async handleHTTPResponse(response: Response, context: string): Promise<any> {
    if (!response.ok) {
      // Try to parse error response for specific error codes
      try {
        const errorData = await response.json();
        if (errorData.error && typeof errorData.error === 'object') {
          const error = errorData.error as JSONRPCError;
          this.handleError(error);
        }
      } catch (parseError) {
        // If we can't parse the error response, continue with standard error handling
      }
      
      throw new HTTPPollingError(
        `HTTP ${response.status}: ${response.statusText} (${context})`,
        response.status
      );
    }

    return await response.json();
  }

  public onPollingStart(handler: () => void): void {
    this.onPollingStartHandler = handler;
  }

  public onPollingStop(handler: () => void): void {
    this.onPollingStopHandler = handler;
  }

  /**
   * Private helper methods
   */
  private getAuthHeaders(): Record<string, string> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };

    // Add authentication token if available using existing proven pattern
    try {
      const authParams = authService.includeAuth({});
      if (authParams.auth_token) {
        headers['Authorization'] = `Bearer ${authParams.auth_token}`;
      }
    } catch (error) {
      // Not authenticated - continue without auth headers
      console.log('HTTP Polling: No authentication token available');
    }

    return headers;
  }

  private async fetchWithTimeout(
    url: string, 
    options: RequestInit, 
    timeout: number
  ): Promise<Response> {
    // Use Node.js http module for better compatibility
    const http = require('http');
    const https = require('https');
    const { URL } = require('url');
    
    const urlObj = new URL(url);
    const isHttps = urlObj.protocol === 'https:';
    const client = isHttps ? https : http;
    
    return new Promise((resolve, reject) => {
      const timeoutId = setTimeout(() => {
        reject(new HTTPPollingError(`Request timeout after ${timeout}ms`));
      }, timeout);
      
      const req = client.request(url, {
        method: options.method || 'GET',
        headers: options.headers || {},
      }, (res: any) => {
        clearTimeout(timeoutId);
        
        let data = '';
        res.on('data', (chunk: string) => {
          data += chunk;
        });
        
        res.on('end', () => {
          // Create a Response-like object
          const response = {
            ok: res.statusCode >= 200 && res.statusCode < 300,
            status: res.statusCode,
            statusText: res.statusMessage,
            headers: res.headers,
            json: async () => JSON.parse(data),
            text: async () => data,
          };
          resolve(response as Response);
        });
      });
      
      req.on('error', (error: Error) => {
        clearTimeout(timeoutId);
        reject(new HTTPPollingError(error.message));
      });
      
      req.on('timeout', () => {
        clearTimeout(timeoutId);
        req.destroy();
        reject(new HTTPPollingError(`Request timeout after ${timeout}ms`));
      });
      
      req.setTimeout(timeout);
      req.end();
    });
  }

  private parseCameraListResponse(data: any): CameraListResponse {
    // Handle different response formats from health endpoints
    if (data.cameras) {
      // Direct camera list format
      return {
        cameras: data.cameras,
        total: data.total || data.cameras.length,
        connected: data.connected || data.cameras.filter((c: any) => c.status === 'CONNECTED').length
      };
    } else if (data.details && typeof data.details === 'string') {
      // Health endpoint format - extract camera count from details
      const cameraMatch = data.details.match(/(\d+)\s*cameras?/i);
      const cameraCount = cameraMatch ? parseInt(cameraMatch[1]) : 0;
      
      return {
        cameras: [],
        total: cameraCount,
        connected: cameraCount
      };
    } else {
      // Fallback format
      return {
        cameras: Array.isArray(data) ? data : [],
        total: Array.isArray(data) ? data.length : 0,
        connected: Array.isArray(data) ? data.filter((c: any) => c.status === 'CONNECTED').length : 0
      };
    }
  }
}
