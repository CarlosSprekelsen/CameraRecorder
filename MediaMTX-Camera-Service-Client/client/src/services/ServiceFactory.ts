// Service Layer - Factory Pattern
// Implements dependency injection as required by architecture

import { WebSocketService } from './websocket/WebSocketService';
import { AuthService } from './auth/AuthService';
import { ServerService } from './server/ServerService';
import { NotificationService } from './notifications/NotificationService';
import { DeviceService } from './device/DeviceService';
import { RecordingService } from './recording/RecordingService';
import { FileService } from './file/FileService';
import { ExternalStreamService } from './external/ExternalStreamService';
import { logger } from './logger/LoggerService';

export class ServiceFactory {
  private static instance: ServiceFactory;
  private wsService: WebSocketService | null = null;
  private authService: AuthService | null = null;
  private serverService: ServerService | null = null;
  private notificationService: NotificationService | null = null;
  private deviceService: DeviceService | null = null;
  private recordingService: RecordingService | null = null;
  private fileService: FileService | null = null;
  private externalStreamService: ExternalStreamService | null = null;

  private constructor() {}

  static getInstance(): ServiceFactory {
    if (!ServiceFactory.instance) {
      ServiceFactory.instance = new ServiceFactory();
    }
    return ServiceFactory.instance;
  }

  createWebSocketService(url: string): WebSocketService {
    if (!this.wsService) {
      this.wsService = new WebSocketService({ url });
      logger.info('WebSocket service created', { url });
    }
    return this.wsService;
  }

  createAuthService(wsService: WebSocketService): AuthService {
    if (!this.authService) {
      this.authService = new AuthService(wsService);
      logger.info('Auth service created');
    }
    return this.authService;
  }

  createServerService(wsService: WebSocketService): ServerService {
    if (!this.serverService) {
      this.serverService = new ServerService(wsService);
      logger.info('Server service created');
    }
    return this.serverService;
  }

  createNotificationService(wsService: WebSocketService): NotificationService {
    if (!this.notificationService) {
      this.notificationService = new NotificationService(wsService);
      logger.info('Notification service created');
    }
    return this.notificationService;
  }

  createDeviceService(wsService: WebSocketService): DeviceService {
    if (!this.deviceService) {
      this.deviceService = new DeviceService(wsService, logger);
      logger.info('Device service created');
    }
    return this.deviceService;
  }

  createRecordingService(wsService: WebSocketService): RecordingService {
    if (!this.recordingService) {
      this.recordingService = new RecordingService(wsService, logger);
      logger.info('Recording service created');
    }
    return this.recordingService;
  }

  createFileService(wsService: WebSocketService): FileService {
    if (!this.fileService) {
      this.fileService = new FileService(wsService, logger);
      logger.info('File service created');
    }
    return this.fileService;
  }

  createExternalStreamService(wsService: WebSocketService): ExternalStreamService {
    if (!this.externalStreamService) {
      this.externalStreamService = new ExternalStreamService(wsService, logger);
      logger.info('External stream service created');
    }
    return this.externalStreamService;
  }

  getWebSocketService(): WebSocketService | null {
    return this.wsService;
  }

  getAuthService(): AuthService | null {
    return this.authService;
  }

  getServerService(): ServerService | null {
    return this.serverService;
  }

  getNotificationService(): NotificationService | null {
    return this.notificationService;
  }

  getDeviceService(): DeviceService | null {
    return this.deviceService;
  }

  getRecordingService(): RecordingService | null {
    return this.recordingService;
  }

  getFileService(): FileService | null {
    return this.fileService;
  }

  getExternalStreamService(): ExternalStreamService | null {
    return this.externalStreamService;
  }

  // Cleanup method for testing
  reset(): void {
    this.wsService = null;
    this.authService = null;
    this.serverService = null;
    this.notificationService = null;
    this.deviceService = null;
    this.recordingService = null;
    this.fileService = null;
    this.externalStreamService = null;
    logger.info('Service factory reset');
  }
}

// Export singleton instance
export const serviceFactory = ServiceFactory.getInstance();
