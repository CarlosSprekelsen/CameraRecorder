// Service Layer - Status Receiver (NOTIF)
// Implements notification handling as required by architecture section 5.1

import { JsonRpcNotification } from '../../types/api';
import { WebSocketService } from '../websocket/WebSocketService';
import { logger } from '../logger/LoggerService';

export interface NotificationHandler {
  (notification: JsonRpcNotification): void;
}

export class NotificationService {
  private handlers = new Map<string, Set<NotificationHandler>>();
  private wsService: WebSocketService; // WebSocket service reference

  constructor(wsService: WebSocketService) {
    this.wsService = wsService;
    this.setupWebSocketHandlers();
  }

  private setupWebSocketHandlers(): void {
    if (this.wsService) {
      this.wsService.events = {
        ...this.wsService.events,
        onNotification: (notification: JsonRpcNotification) => {
          this.handleNotification(notification);
        },
      };
    }
  }

  subscribe(method: string, handler: NotificationHandler): () => void {
    if (!this.handlers.has(method)) {
      this.handlers.set(method, new Set());
    }

    this.handlers.get(method)!.add(handler);

    logger.info(`Subscribed to notifications: ${method}`);

    // Return unsubscribe function
    return () => {
      this.unsubscribe(method, handler);
    };
  }

  unsubscribe(method: string, handler: NotificationHandler): void {
    const methodHandlers = this.handlers.get(method);
    if (methodHandlers) {
      methodHandlers.delete(handler);
      if (methodHandlers.size === 0) {
        this.handlers.delete(method);
      }
    }

    logger.info(`Unsubscribed from notifications: ${method}`);
  }

  private handleNotification(notification: JsonRpcNotification): void {
    const method = notification.method;
    const methodHandlers = this.handlers.get(method);

    if (methodHandlers) {
      logger.debug(`Handling notification: ${method}`, { notification });

      methodHandlers.forEach((handler) => {
        try {
          handler(notification);
        } catch (error) {
          logger.error(`Error in notification handler for ${method}`, error as Record<string, unknown>);
        }
      });
    } else {
      logger.debug(`No handlers for notification: ${method}`);
    }
  }

  getSubscribedMethods(): string[] {
    return Array.from(this.handlers.keys());
  }

  getHandlerCount(method: string): number {
    return this.handlers.get(method)?.size || 0;
  }
}
