/**
 * Admin Service for MediaMTX Camera Service Client
 * Provides system administration and management functionality
 * Aligned with server JSON-RPC methods for admin operations
 * 
 * Server API Reference: ../mediamtx-camera-service/docs/api/json-rpc-methods.md
 */

import type {
  SystemMetrics,
  SystemStatus,
  ServerInfo,
  StorageInfo,
  RetentionPolicy,
  CleanupResults,
} from '../stores/adminStore';
import type { JSONRPCRequest, JSONRPCResponse } from '../types/rpc';
import type { WebSocketService } from './websocket';
import { authService } from './authService';

/**
 * Admin service configuration
 */
export interface AdminServiceConfig {
  timeout: number;
  retryAttempts: number;
  retryDelay: number;
}

/**
 * Admin service error
 */
export class AdminServiceError extends Error {
  public code?: number;
  public method?: string;

  constructor(message: string, code?: number, method?: string) {
    super(message);
    this.name = 'AdminServiceError';
    this.code = code;
    this.method = method;
  }
}

/**
 * Admin service class
 * Provides methods to interact with server admin JSON-RPC methods
 */
export class AdminService {
  private config: AdminServiceConfig;
  private wsService: WebSocketService | null = null;

  constructor(config: Partial<AdminServiceConfig> = {}) {
    this.config = {
      timeout: config.timeout || 10000,
      retryAttempts: config.retryAttempts || 3,
      retryDelay: config.retryDelay || 1000,
    };
  }

  /**
   * Set WebSocket service reference
   * @param wsService WebSocket service instance
   */
  setWebSocketService(wsService: WebSocketService): void {
    this.wsService = wsService;
  }

  /**
   * Get system metrics
   * @returns Promise<SystemMetrics> System performance metrics
   */
  async getMetrics(): Promise<SystemMetrics> {
    return this.callRPC<SystemMetrics>('get_metrics');
  }

  /**
   * Get system status
   * @returns Promise<SystemStatus> System health status
   */
  async getStatus(): Promise<SystemStatus> {
    return this.callRPC<SystemStatus>('get_status');
  }

  /**
   * Get server information
   * @returns Promise<ServerInfo> Server configuration and capabilities
   */
  async getServerInfo(): Promise<ServerInfo> {
    return this.callRPC<ServerInfo>('get_server_info');
  }

  /**
   * Get storage information
   * @returns Promise<StorageInfo> Storage space and usage information
   */
  async getStorageInfo(): Promise<StorageInfo> {
    return this.callRPC<StorageInfo>('get_storage_info');
  }

  /**
   * Set retention policy
   * @param policy Retention policy configuration
   * @returns Promise<RetentionPolicy> Updated retention policy
   */
  async setRetentionPolicy(policy: {
    policy_type: 'age' | 'size' | 'manual';
    max_age_days?: number;
    max_size_gb?: number;
    enabled: boolean;
  }): Promise<RetentionPolicy> {
    return this.callRPC<RetentionPolicy>('set_retention_policy', policy);
  }

  /**
   * Cleanup old files
   * @returns Promise<CleanupResults> Cleanup operation results
   */
  async cleanupOldFiles(): Promise<CleanupResults> {
    return this.callRPC<CleanupResults>('cleanup_old_files');
  }

  /**
   * Get all system information
   * @returns Promise<object> All system information
   */
  async getAllSystemInfo(): Promise<{
    metrics: SystemMetrics;
    status: SystemStatus;
    serverInfo: ServerInfo;
    storageInfo: StorageInfo;
  }> {
    const [metrics, status, serverInfo, storageInfo] = await Promise.all([
      this.getMetrics(),
      this.getStatus(),
      this.getServerInfo(),
      this.getStorageInfo(),
    ]);

    return {
      metrics,
      status,
      serverInfo,
      storageInfo,
    };
  }

  /**
   * Call JSON-RPC method with authentication
   * @param method RPC method name
   * @param params Method parameters
   * @returns Promise<T> Method response
   */
  private async callRPC<T>(method: string, params: Record<string, unknown> = {}): Promise<T> {
    if (!this.wsService) {
      throw new AdminServiceError('WebSocket service not available', undefined, method);
    }

    let lastError: Error | null = null;

    for (let attempt = 1; attempt <= this.config.retryAttempts; attempt++) {
      try {
        // Include authentication in parameters
        const authParams = authService.includeAuth(params);
        
        const response = await this.wsService.call(method, authParams);
        return response as T;
      } catch (error) {
        lastError = error as Error;
        
        // Check if it's an authentication error
        if (this.isAuthError(error)) {
          throw new AdminServiceError(
            'Authentication required for admin operations',
            -32003,
            method
          );
        }
        
        // Check if it's a permission error
        if (this.isPermissionError(error)) {
          throw new AdminServiceError(
            'Insufficient permissions for admin operations',
            -32003,
            method
          );
        }
        
        if (attempt < this.config.retryAttempts) {
          await this.delay(this.config.retryDelay * attempt);
        }
      }
    }

    throw lastError || new AdminServiceError('Request failed after all retry attempts', undefined, method);
  }

  /**
   * Check if error is authentication related
   * @param error Error object
   * @returns boolean True if authentication error
   */
  private isAuthError(error: unknown): boolean {
    if (error && typeof error === 'object' && 'code' in error) {
      const code = (error as { code: number }).code;
      return code === -32001 || code === -32004; // Authentication failed or token expired
    }
    return false;
  }

  /**
   * Check if error is permission related
   * @param error Error object
   * @returns boolean True if permission error
   */
  private isPermissionError(error: unknown): boolean {
    if (error && typeof error === 'object' && 'code' in error) {
      const code = (error as { code: number }).code;
      return code === -32003; // Insufficient permissions
    }
    return false;
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
   * @returns AdminServiceConfig Current configuration
   */
  getConfig(): AdminServiceConfig {
    return { ...this.config };
  }

  /**
   * Update service configuration
   * @param config Partial configuration to update
   */
  updateConfig(config: Partial<AdminServiceConfig>): void {
    this.config = { ...this.config, ...config };
  }

  /**
   * Check if user has admin permissions
   * @returns boolean True if user has admin permissions
   */
  hasAdminPermissions(): boolean {
    try {
      const token = authService.getToken();
      if (!token) return false;
      
      // Basic check - in a real implementation, you might decode the JWT
      // to check the role claim
      return true; // Simplified for now
    } catch {
      return false;
    }
  }
}

/**
 * Default admin service instance
 */
export const adminService = new AdminService();

/**
 * Export admin service for use in components
 */
export default adminService;
