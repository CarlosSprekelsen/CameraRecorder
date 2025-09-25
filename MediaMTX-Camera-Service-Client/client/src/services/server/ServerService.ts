import { ServerInfo, SystemStatus, StorageInfo } from '../../types/api';

// Metrics interface based on server API specification
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
    cameras: Record<string, any>;
  };
  recording_metrics: Record<string, any>;
  stream_metrics: {
    active_streams: number;
    total_streams: number;
    total_viewers: number;
  };
}
import { WebSocketService } from '../websocket/WebSocketService';

export class ServerService {
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

  async getStorageInfo(): Promise<StorageInfo> {
    if (!this.wsService.isConnected) {
      throw new Error('WebSocket not connected');
    }

    return this.wsService.sendRPC<StorageInfo>('get_storage_info');
  }

  async getMetrics(): Promise<SystemMetrics> {
    if (!this.wsService.isConnected) {
      throw new Error('WebSocket not connected');
    }

    return this.wsService.sendRPC<SystemMetrics>('get_metrics');
  }

  async ping(): Promise<string> {
    if (!this.wsService.isConnected) {
      throw new Error('WebSocket not connected');
    }

    return this.wsService.sendRPC<string>('ping');
  }
}
