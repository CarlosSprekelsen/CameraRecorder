import { ServerInfo, SystemStatus, StorageInfo } from '../../types/api';
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

  async ping(): Promise<string> {
    if (!this.wsService.isConnected) {
      throw new Error('WebSocket not connected');
    }

    return this.wsService.sendRPC<string>('ping');
  }
}
