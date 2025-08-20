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

import type { CameraDevice, CameraListResponse } from '../types/camera';
import { authService } from './authService';

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

  /**
   * Get camera list via HTTP (single request)
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
