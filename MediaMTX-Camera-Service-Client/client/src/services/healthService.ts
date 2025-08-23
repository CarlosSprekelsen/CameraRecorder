/**
 * Health Service for MediaMTX Camera Service Client
 * Provides dedicated health endpoint integration
 * Aligned with server health endpoints API
 * 
 * Server API Reference: ../mediamtx-camera-service/docs/api/health-endpoints.md
 */

import type {
  SystemHealth,
  CameraHealth,
  MediaMTXHealth,
  ReadinessStatus,
  HealthStatus,
} from '../stores/healthStore';
import { authService } from './authService';

/**
 * Health service configuration
 */
export interface HealthServiceConfig {
  baseUrl: string;
  timeout: number;
  retryAttempts: number;
  retryDelay: number;
}

/**
 * Health service error
 */
export class HealthServiceError extends Error {
  public statusCode?: number;
  public endpoint?: string;

  constructor(message: string, statusCode?: number, endpoint?: string) {
    super(message);
    this.name = 'HealthServiceError';
    this.statusCode = statusCode;
    this.endpoint = endpoint;
  }
}

/**
 * Health service class
 * Provides methods to interact with server health endpoints
 */
export class HealthService {
  private config: HealthServiceConfig;
  private isPolling = false;
  private pollInterval: NodeJS.Timeout | null = null;

  constructor(config: Partial<HealthServiceConfig> = {}) {
    this.config = {
      baseUrl: config.baseUrl || 'http://localhost:8003',
      timeout: config.timeout || 10000,
      retryAttempts: config.retryAttempts || 3,
      retryDelay: config.retryDelay || 1000,
    };
  }

  /**
   * Get system health status
   * @returns Promise<SystemHealth> System health information
   */
  async getSystemHealth(): Promise<SystemHealth> {
    return this.makeRequest<SystemHealth>('/health/system');
  }

  /**
   * Get camera system health
   * @returns Promise<CameraHealth> Camera system health information
   */
  async getCameraHealth(): Promise<CameraHealth> {
    return this.makeRequest<CameraHealth>('/health/cameras');
  }

  /**
   * Get MediaMTX integration health
   * @returns Promise<MediaMTXHealth> MediaMTX health information
   */
  async getMediaMTXHealth(): Promise<MediaMTXHealth> {
    return this.makeRequest<MediaMTXHealth>('/health/mediamtx');
  }

  /**
   * Get Kubernetes readiness status
   * @returns Promise<ReadinessStatus> Readiness status information
   */
  async getReadinessStatus(): Promise<ReadinessStatus> {
    return this.makeRequest<ReadinessStatus>('/health/ready');
  }

  /**
   * Get all health information
   * @returns Promise<object> All health endpoints data
   */
  async getAllHealth(): Promise<{
    system: SystemHealth;
    cameras: CameraHealth;
    mediamtx: MediaMTXHealth;
    readiness: ReadinessStatus;
  }> {
    const [system, cameras, mediamtx, readiness] = await Promise.all([
      this.getSystemHealth(),
      this.getCameraHealth(),
      this.getMediaMTXHealth(),
      this.getReadinessStatus(),
    ]);

    return {
      system,
      cameras,
      mediamtx,
      readiness,
    };
  }

  /**
   * Start health polling
   * @param interval Polling interval in milliseconds
   * @param callback Callback function to handle health updates
   */
  startPolling(
    interval: number,
    callback: (health: {
      system: SystemHealth;
      cameras: CameraHealth;
      mediamtx: MediaMTXHealth;
      readiness: ReadinessStatus;
    }) => void
  ): void {
    if (this.isPolling) {
      this.stopPolling();
    }

    this.isPolling = true;
    this.pollInterval = setInterval(async () => {
      try {
        const health = await this.getAllHealth();
        callback(health);
      } catch (error) {
        console.error('Health polling error:', error);
      }
    }, interval);
  }

  /**
   * Stop health polling
   */
  stopPolling(): void {
    if (this.pollInterval) {
      clearInterval(this.pollInterval);
      this.pollInterval = null;
    }
    this.isPolling = false;
  }

  /**
   * Check if polling is active
   * @returns boolean True if polling is active
   */
  get isActive(): boolean {
    return this.isPolling;
  }

  /**
   * Make HTTP request with retry logic
   * @param endpoint API endpoint
   * @returns Promise<T> Response data
   */
  private async makeRequest<T>(endpoint: string): Promise<T> {
    let lastError: Error | null = null;

    for (let attempt = 1; attempt <= this.config.retryAttempts; attempt++) {
      try {
        const response = await this.fetchWithTimeout(
          `${this.config.baseUrl}${endpoint}`,
          {
            method: 'GET',
            headers: this.getAuthHeaders(),
          },
          this.config.timeout
        );

        if (!response.ok) {
          throw new HealthServiceError(
            `HTTP ${response.status}: ${response.statusText}`,
            response.status,
            endpoint
          );
        }

        return await response.json();
      } catch (error) {
        lastError = error as Error;
        
        if (attempt < this.config.retryAttempts) {
          await this.delay(this.config.retryDelay * attempt);
        }
      }
    }

    throw lastError || new HealthServiceError('Request failed after all retry attempts');
  }

  /**
   * Get authentication headers
   * @returns Record<string, string> Headers with authentication
   */
  private getAuthHeaders(): Record<string, string> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };

    try {
      const authParams = authService.includeAuth({});
      if (authParams.auth_token) {
        headers['Authorization'] = `Bearer ${authParams.auth_token}`;
      }
    } catch (error) {
      // Not authenticated - continue without auth headers
      console.log('Health Service: No authentication token available');
    }

    return headers;
  }

  /**
   * Fetch with timeout
   * @param url Request URL
   * @param options Request options
   * @param timeout Timeout in milliseconds
   * @returns Promise<Response> Fetch response
   */
  private async fetchWithTimeout(
    url: string,
    options: RequestInit,
    timeout: number
  ): Promise<Response> {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), timeout);

    try {
      const response = await fetch(url, {
        ...options,
        signal: controller.signal,
      });
      clearTimeout(timeoutId);
      return response;
    } catch (error) {
      clearTimeout(timeoutId);
      if (error instanceof Error && error.name === 'AbortError') {
        throw new HealthServiceError('Request timeout');
      }
      throw error;
    }
  }

  /**
   * Delay utility
   * @param ms Milliseconds to delay
   * @returns Promise<void>
   */
  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  /**
   * Get service configuration
   * @returns HealthServiceConfig Current configuration
   */
  getConfig(): HealthServiceConfig {
    return { ...this.config };
  }

  /**
   * Update service configuration
   * @param config Partial configuration to update
   */
  updateConfig(config: Partial<HealthServiceConfig>): void {
    this.config = { ...this.config, ...config };
  }
}

/**
 * Default health service instance
 */
export const healthService = new HealthService();

/**
 * Export health service for use in components
 */
export default healthService;
