// Service Layer - Factory Pattern
// Implements dependency injection as required by architecture

import { WebSocketService } from './websocket/WebSocketService';
import { AuthService } from './auth/AuthService';
import { ServerService } from './server/ServerService';
import { NotificationService } from './notifications/NotificationService';
import { logger } from './logger/LoggerService';

export class ServiceFactory {
  private static instance: ServiceFactory;
  private wsService: WebSocketService | null = null;
  private authService: AuthService | null = null;
  private serverService: ServerService | null = null;
  private notificationService: NotificationService | null = null;

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

  // Cleanup method for testing
  reset(): void {
    this.wsService = null;
    this.authService = null;
    this.serverService = null;
    this.notificationService = null;
    logger.info('Service factory reset');
  }
}

// Export singleton instance
export const serviceFactory = ServiceFactory.getInstance();
