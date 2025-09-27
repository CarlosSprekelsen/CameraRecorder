import { ServerInfo, SystemStatus, StorageInfo, MetricsResult } from '../../types/api';
import { IStatus } from '../interfaces/ServiceInterfaces';

/**
 * System Metrics Interface
 *
 * Defines the structure for system performance metrics including CPU, memory,
 * disk usage, and camera-specific metrics.
 *
 * @interface SystemMetrics
 */
export interface SystemMetrics {
  timestamp: string;
  system_metrics: {
    cpu_usage: number;
    memory_usage: number;
    disk_usage: number;
    goroutines: number;
  };
  camera_metrics: {
    connected_cameras: number;
    cameras: Record<string, Record<string, unknown>>;
  };
  recording_metrics: Record<string, Record<string, unknown>>;
  stream_metrics: {
    active_streams: number;
    total_streams: number;
    total_viewers: number;
  };
}
import { WebSocketService } from '../websocket/WebSocketService';
import { SubscriptionResult, UnsubscriptionResult, SubscriptionStatsResult } from '../../types/api';

/**
 * Server Service - System status and metrics management
 *
 * Implements the IStatus interface for server status monitoring, system metrics,
 * and event subscription management. Provides methods for retrieving server information,
 * system health, storage details, and managing real-time event subscriptions.
 *
 * @class ServerService
 * @implements {IStatus}
 *
 * @example
 * ```typescript
 * const serverService = new ServerService(wsService);
 *
 * // Get system status
 * const status = await serverService.getStatus();
 *
 * // Get server metrics
 * const metrics = await serverService.getMetrics();
 *
 * // Subscribe to events
 * const subscription = await serverService.subscribeEvents(['camera_status_update']);
 * ```
 *
 * @see {@link ../interfaces/ServiceInterfaces#IStatus} IStatus interface
 * @see {@link ../../docs/architecture/client-architechture.md} Client Architecture
 */
export class ServerService implements IStatus {
  private wsService: WebSocketService;

  constructor(wsService: WebSocketService) {
    this.wsService = wsService;
  }

  async getServerInfo(): Promise<ServerInfo> {
    if (!this.wsService.isConnected) {
      throw new Error('WebSocket not connected');
    }

    return this.wsService.sendRPC<ServerInfo>('get_server_info');
  }

  async getStatus(): Promise<SystemStatus> {
    if (!this.wsService.isConnected) {
      throw new Error('WebSocket not connected');
    }

    return this.wsService.sendRPC<SystemStatus>('get_status');
  }

  async getSystemStatus(): Promise<SystemStatus> {
    if (!this.wsService.isConnected) {
      throw new Error('WebSocket not connected');
    }

    return this.wsService.sendRPC<SystemStatus>('get_system_status');
  }

  async getStorageInfo(): Promise<StorageInfo> {
    if (!this.wsService.isConnected) {
      throw new Error('WebSocket not connected');
    }

    return this.wsService.sendRPC<StorageInfo>('get_storage_info');
  }

  async getMetrics(): Promise<MetricsResult> {
    if (!this.wsService.isConnected) {
      throw new Error('WebSocket not connected');
    }

    return this.wsService.sendRPC<MetricsResult>('get_metrics');
  }

  async ping(): Promise<string> {
    if (!this.wsService.isConnected) {
      throw new Error('WebSocket not connected');
    }

    return this.wsService.sendRPC<string>('ping');
  }

  // IStatus interface implementation
  async subscribeEvents(
    topics: string[],
    filters?: Record<string, unknown>,
  ): Promise<SubscriptionResult> {
    if (!this.wsService.isConnected) {
      throw new Error('WebSocket not connected');
    }

    return this.wsService.sendRPC('subscribe_events', { topics, filters });
  }

  async unsubscribeEvents(topics?: string[]): Promise<UnsubscriptionResult> {
    if (!this.wsService.isConnected) {
      throw new Error('WebSocket not connected');
    }

    return this.wsService.sendRPC('unsubscribe_events', { topics });
  }

  async getSubscriptionStats(): Promise<SubscriptionStatsResult> {
    if (!this.wsService.isConnected) {
      throw new Error('WebSocket not connected');
    }

    return this.wsService.sendRPC('get_subscription_stats');
  }
}
