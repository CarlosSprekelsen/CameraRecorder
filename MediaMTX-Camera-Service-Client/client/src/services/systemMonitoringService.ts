/**
 * System Monitoring Service
 * Handles system status, metrics, and server information
 */

import { websocketService } from './websocket';
import { RPC_METHODS } from '../types/rpc';
import { logger, loggers } from './loggerService';

export interface SystemStatus {
  status: string;
  uptime: number;
  version: string;
  components: {
    websocket_server: string;
    camera_monitor: string;
    mediamtx: string;
  };
}

export interface ServerInfo {
  name: string;
  version: string;
  build_date: string;
  go_version: string;
  architecture: string;
  capabilities: string[];
  supported_formats: string[];
  max_cameras: number;
}

export interface SystemMetrics {
  active_connections: number;
  total_requests: number;
  average_response_time: number;
  error_rate: number;
  memory_usage: number;
  cpu_usage: number;
  disk_usage: number;
  goroutines: number;
  heap_alloc: number;
}

export class SystemMonitoringService {
  private static instance: SystemMonitoringService;
  private updateCallbacks: Set<(data: any) => void> = new Set();
  private monitoringInterval: NodeJS.Timeout | null = null;
  private currentStatus: SystemStatus | null = null;
  private currentMetrics: SystemMetrics | null = null;
  private currentServerInfo: ServerInfo | null = null;

  private constructor() {}

  static getInstance(): SystemMonitoringService {
    if (!SystemMonitoringService.instance) {
      SystemMonitoringService.instance = new SystemMonitoringService();
    }
    return SystemMonitoringService.instance;
  }

  /**
   * Get system status
   */
  public async getSystemStatus(): Promise<SystemStatus> {
    loggers.service.start('SystemMonitoringService', 'getSystemStatus');

    try {
      if (!websocketService.isConnected()) {
        throw new Error('WebSocket not connected');
      }

      const response = await websocketService.call(RPC_METHODS.GET_STATUS, {});
      const status = response as SystemStatus;
      
      this.currentStatus = status;
      this.notifyUpdateCallbacks({ type: 'status', data: status });

      loggers.service.success('SystemMonitoringService', 'getSystemStatus', {
        status: status.status,
        uptime: status.uptime
      });

      return status;
    } catch (error) {
      loggers.service.error('SystemMonitoringService', 'getSystemStatus', error as Error);
      throw error;
    }
  }

  /**
   * Get server information
   */
  public async getServerInfo(): Promise<ServerInfo> {
    loggers.service.start('SystemMonitoringService', 'getServerInfo');

    try {
      if (!websocketService.isConnected()) {
        throw new Error('WebSocket not connected');
      }

      const response = await websocketService.call(RPC_METHODS.GET_SERVER_INFO, {});
      const serverInfo = response as ServerInfo;
      
      this.currentServerInfo = serverInfo;
      this.notifyUpdateCallbacks({ type: 'serverInfo', data: serverInfo });

      loggers.service.success('SystemMonitoringService', 'getServerInfo', {
        name: serverInfo.name,
        version: serverInfo.version
      });

      return serverInfo;
    } catch (error) {
      loggers.service.error('SystemMonitoringService', 'getServerInfo', error as Error);
      throw error;
    }
  }

  /**
   * Get system metrics
   */
  public async getSystemMetrics(): Promise<SystemMetrics> {
    loggers.service.start('SystemMonitoringService', 'getSystemMetrics');

    try {
      if (!websocketService.isConnected()) {
        throw new Error('WebSocket not connected');
      }

      const response = await websocketService.call(RPC_METHODS.GET_METRICS, {});
      const metrics = response as SystemMetrics;
      
      this.currentMetrics = metrics;
      this.notifyUpdateCallbacks({ type: 'metrics', data: metrics });

      loggers.service.success('SystemMonitoringService', 'getSystemMetrics', {
        activeConnections: metrics.active_connections,
        averageResponseTime: metrics.average_response_time
      });

      return metrics;
    } catch (error) {
      loggers.service.error('SystemMonitoringService', 'getSystemMetrics', error as Error);
      throw error;
    }
  }

  /**
   * Start continuous monitoring
   */
  public startMonitoring(intervalMs: number = 30000): void {
    loggers.service.info('SystemMonitoringService', 'startMonitoring', { intervalMs });

    if (this.monitoringInterval) {
      clearInterval(this.monitoringInterval);
    }

    this.monitoringInterval = setInterval(async () => {
      try {
        await Promise.all([
          this.getSystemStatus(),
          this.getSystemMetrics()
        ]);
      } catch (error) {
        loggers.service.error('SystemMonitoringService', 'monitoringInterval', error as Error);
      }
    }, intervalMs);
  }

  /**
   * Stop continuous monitoring
   */
  public stopMonitoring(): void {
    loggers.service.info('SystemMonitoringService', 'stopMonitoring');

    if (this.monitoringInterval) {
      clearInterval(this.monitoringInterval);
      this.monitoringInterval = null;
    }
  }

  /**
   * Subscribe to system updates
   */
  public subscribeToUpdates(callback: (data: any) => void): () => void {
    this.updateCallbacks.add(callback);
    
    // Return unsubscribe function
    return () => {
      this.updateCallbacks.delete(callback);
    };
  }

  /**
   * Get current cached data
   */
  public getCurrentData(): {
    status: SystemStatus | null;
    metrics: SystemMetrics | null;
    serverInfo: ServerInfo | null;
  } {
    return {
      status: this.currentStatus,
      metrics: this.currentMetrics,
      serverInfo: this.currentServerInfo
    };
  }

  /**
   * Notify all update callbacks
   */
  private notifyUpdateCallbacks(data: any): void {
    this.updateCallbacks.forEach(callback => {
      try {
        callback(data);
      } catch (error) {
        loggers.service.error('SystemMonitoringService', 'notifyUpdateCallbacks', error as Error);
      }
    });
  }

  /**
   * Cleanup resources
   */
  public cleanup(): void {
    loggers.service.info('SystemMonitoringService', 'cleanup');
    
    this.stopMonitoring();
    this.updateCallbacks.clear();
    this.currentStatus = null;
    this.currentMetrics = null;
    this.currentServerInfo = null;
  }
}

// Export singleton instance
export const systemMonitoringService = SystemMonitoringService.getInstance();

// Export types
export type { SystemStatus, ServerInfo, SystemMetrics };
