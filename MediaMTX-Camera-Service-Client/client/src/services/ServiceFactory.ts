// Service Layer - Factory Pattern
// Implements dependency injection as required by architecture

import { APIClient } from './abstraction/APIClient';
import { EventBus } from './events/EventBus';
import { AuthService } from './auth/AuthService';
import { ServerService } from './server/ServerService';
import { NotificationService } from './notifications/NotificationService';
import { DeviceService } from './device/DeviceService';
import { RecordingService } from './recording/RecordingService';
import { FileService } from './file/FileService';
import { StreamingService } from './streaming/StreamingService';
import { ExternalStreamService } from './external/ExternalStreamService';
import { logger } from './logger/LoggerService';

export class ServiceFactory {
  private static instance: ServiceFactory;
  private apiClient: APIClient | null = null;
  private authService: AuthService | null = null;
  private serverService: ServerService | null = null;
  private notificationService: NotificationService | null = null;
  private deviceService: DeviceService | null = null;
  private recordingService: RecordingService | null = null;
  private fileService: FileService | null = null;
  private streamingService: StreamingService | null = null;
  private externalStreamService: ExternalStreamService | null = null;

  private constructor() {}

  static getInstance(): ServiceFactory {
    if (!ServiceFactory.instance) {
      ServiceFactory.instance = new ServiceFactory();
    }
    return ServiceFactory.instance;
  }

  createAPIClient(wsService: any): APIClient {
    if (!this.apiClient) {
      this.apiClient = new APIClient(wsService, logger);
      logger.info('API Client created');
    }
    return this.apiClient;
  }

  createAuthService(apiClient: APIClient): AuthService {
    if (!this.authService) {
      this.authService = new AuthService(apiClient, logger);
      logger.info('Auth service created');
    }
    return this.authService;
  }

  createServerService(apiClient: APIClient): ServerService {
    if (!this.serverService) {
      this.serverService = new ServerService(apiClient, logger);
      logger.info('Server service created');
    }
    return this.serverService;
  }

  createNotificationService(apiClient: APIClient, eventBus: EventBus): NotificationService {
    if (!this.notificationService) {
      this.notificationService = new NotificationService(apiClient, logger, eventBus);
      logger.info('Notification service created');
    }
    return this.notificationService;
  }

  createDeviceService(apiClient: APIClient): DeviceService {
    if (!this.deviceService) {
      this.deviceService = new DeviceService(apiClient, logger);
      logger.info('Device service created');
    }
    return this.deviceService;
  }

  createRecordingService(apiClient: APIClient): RecordingService {
    if (!this.recordingService) {
      this.recordingService = new RecordingService(apiClient, logger);
      logger.info('Recording service created');
    }
    return this.recordingService;
  }

  createFileService(apiClient: APIClient): FileService {
    if (!this.fileService) {
      this.fileService = new FileService(apiClient, logger);
      logger.info('File service created');
    }
    return this.fileService;
  }

  createStreamingService(apiClient: APIClient): StreamingService {
    if (!this.streamingService) {
      this.streamingService = new StreamingService(apiClient, logger);
      logger.info('Streaming service created');
    }
    return this.streamingService;
  }

  createExternalStreamService(apiClient: APIClient): ExternalStreamService {
    if (!this.externalStreamService) {
      this.externalStreamService = new ExternalStreamService(apiClient, logger);
      logger.info('External stream service created');
    }
    return this.externalStreamService;
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

  getStreamingService(): StreamingService | null {
    return this.streamingService;
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
