/**
 * NotificationService unit tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * 
 * Requirements Coverage:
 * - REQ-NOTIF-001: Notification subscription management
 * - REQ-NOTIF-002: Notification handler execution
 * - REQ-NOTIF-003: Error handling in notification handlers
 * - REQ-NOTIF-004: Unsubscription functionality
 * - REQ-NOTIF-005: WebSocket integration
 * 
 * Test Categories: Unit
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */

import { NotificationService } from '../../../src/services/notifications/NotificationService';
import { APIClient } from '../../../src/services/abstraction/APIClient';
import { WebSocketService } from '../../../src/services/websocket/WebSocketService';
import { LoggerService } from '../../../src/services/logger/LoggerService';
import { EventBus } from '../../../src/services/events/EventBus';
import { JsonRpcNotification } from '../../../src/types/api';
import { MockDataFactory } from '../../utils/mocks';

// Use centralized mocks - eliminates duplication
const mockWebSocketService = MockDataFactory.createMockWebSocketService();
const mockLoggerService = MockDataFactory.createMockLoggerService();
const mockAPIClient = new APIClient(mockWebSocketService, mockLoggerService);
const mockEventBus = MockDataFactory.createMockEventBus();

describe('NotificationService Unit Tests', () => {
  let notificationService: NotificationService;
  let mockHandler: jest.Mock;

  beforeEach(() => {
    jest.clearAllMocks();
    mockHandler = MockDataFactory.createMockEventHandler();
    notificationService = new NotificationService(mockAPIClient, mockLoggerService, mockEventBus);
  });

  describe('REQ-NOTIF-001: Notification subscription management', () => {
    test('Should subscribe to notification method (server-generated)', () => {
      // Note: camera_status_update is a server-generated notification, not a callable method
      // This test validates the client can set up handlers for incoming notifications
      const unsubscribe = notificationService.subscribe('camera_status_update', mockHandler);

      // Assert
      expect(notificationService.getSubscribedMethods()).toContain('camera_status_update');
      expect(notificationService.getHandlerCount('camera_status_update')).toBe(1);
      expect(typeof unsubscribe).toBe('function');
    });

    test('Should handle multiple subscriptions to same method', () => {
      // Arrange
      const handler1 = MockDataFactory.createMockEventHandler();
      const handler2 = MockDataFactory.createMockEventHandler();

      // Act
      notificationService.subscribe('recording_status_update', handler1);
      notificationService.subscribe('recording_status_update', handler2);

      // Assert
      expect(notificationService.getHandlerCount('recording_status_update')).toBe(2);
    });

    test('Should handle subscriptions to different methods', () => {
      // Act
      notificationService.subscribe('camera_status_update', mockHandler);
      notificationService.subscribe('recording_status_update', mockHandler);

      // Assert
      expect(notificationService.getSubscribedMethods()).toHaveLength(2);
      expect(notificationService.getSubscribedMethods()).toContain('camera_status_update');
      expect(notificationService.getSubscribedMethods()).toContain('recording_status_update');
    });
  });

  describe('REQ-NOTIF-002: Notification handler execution', () => {
    test('Should execute handlers for subscribed method', () => {
      // Arrange
      const notification: JsonRpcNotification = {
        jsonrpc: '2.0',
        method: 'camera_status_update',
        params: {
          device: 'camera0',
          status: 'CONNECTED'
        }
      };

      notificationService.subscribe('camera_status_update', mockHandler);

      // Act
      notificationService['handleNotification'](notification);

      // Assert
      expect(mockHandler).toHaveBeenCalledWith(notification);
    });

    test('Should not execute handlers for unsubscribed method', () => {
      // Arrange
      const notification: JsonRpcNotification = {
        jsonrpc: '2.0',
        method: 'camera_status_update',
        params: {
          device: 'camera0',
          status: 'CONNECTED'
        }
      };

      // Act
      notificationService['handleNotification'](notification);

      // Assert
      expect(mockHandler).not.toHaveBeenCalled();
    });

    test('Should execute multiple handlers for same method', () => {
      // Arrange
      const handler1 = MockDataFactory.createMockEventHandler();
      const handler2 = MockDataFactory.createMockEventHandler();
      const notification: JsonRpcNotification = {
        jsonrpc: '2.0',
        method: 'recording_status_update',
        params: {
          device: 'camera0',
          status: 'STARTED'
        }
      };

      notificationService.subscribe('recording_status_update', handler1);
      notificationService.subscribe('recording_status_update', handler2);

      // Act
      notificationService['handleNotification'](notification);

      // Assert
      expect(handler1).toHaveBeenCalledWith(notification);
      expect(handler2).toHaveBeenCalledWith(notification);
    });
  });

  describe('REQ-NOTIF-003: Error handling in notification handlers', () => {
    test('Should handle errors in notification handlers gracefully', () => {
      // Arrange
      const errorHandler = MockDataFactory.createMockEventHandler().mockImplementation(() => {
        throw new Error('Handler error');
      });
      const notification: JsonRpcNotification = {
        jsonrpc: '2.0',
        method: 'camera_status_update',
        params: {
          device: 'camera0',
          status: 'CONNECTED'
        }
      };

      notificationService.subscribe('camera_status_update', errorHandler);

      // Act & Assert (should not throw)
      expect(() => {
        notificationService['handleNotification'](notification);
      }).not.toThrow();

      expect(errorHandler).toHaveBeenCalledWith(notification);
    });

    test('Should continue executing other handlers when one fails', () => {
      // Arrange
      const errorHandler = MockDataFactory.createMockEventHandler().mockImplementation(() => {
        throw new Error('Handler error');
      });
      const successHandler = MockDataFactory.createMockEventHandler();
      const notification: JsonRpcNotification = {
        jsonrpc: '2.0',
        method: 'camera_status_update',
        params: {
          device: 'camera0',
          status: 'CONNECTED'
        }
      };

      notificationService.subscribe('camera_status_update', errorHandler);
      notificationService.subscribe('camera_status_update', successHandler);

      // Act
      notificationService['handleNotification'](notification);

      // Assert
      expect(errorHandler).toHaveBeenCalledWith(notification);
      expect(successHandler).toHaveBeenCalledWith(notification);
    });
  });

  describe('REQ-NOTIF-004: Unsubscription functionality', () => {
    test('Should unsubscribe handler using returned function', () => {
      // Arrange
      const unsubscribe = notificationService.subscribe('camera_status_update', mockHandler);
      expect(notificationService.getHandlerCount('camera_status_update')).toBe(1);

      // Act
      unsubscribe();

      // Assert
      expect(notificationService.getHandlerCount('camera_status_update')).toBe(0);
      expect(notificationService.getSubscribedMethods()).not.toContain('camera_status_update');
    });

    test('Should unsubscribe handler using unsubscribe method', () => {
      // Arrange
      notificationService.subscribe('camera_status_update', mockHandler);
      expect(notificationService.getHandlerCount('camera_status_update')).toBe(1);

      // Act
      notificationService.unsubscribe('camera_status_update', mockHandler);

      // Assert
      expect(notificationService.getHandlerCount('camera_status_update')).toBe(0);
    });

    test('Should handle unsubscription of non-existent handler gracefully', () => {
      // Act & Assert (should not throw)
      expect(() => {
        notificationService.unsubscribe('non_existent_method', mockHandler);
      }).not.toThrow();
    });

    test('Should remove method from subscribed methods when last handler is removed', () => {
      // Arrange
      const handler1 = MockDataFactory.createMockEventHandler();
      const handler2 = MockDataFactory.createMockEventHandler();
      notificationService.subscribe('camera_status_update', handler1);
      notificationService.subscribe('camera_status_update', handler2);
      expect(notificationService.getSubscribedMethods()).toContain('camera_status_update');

      // Act
      notificationService.unsubscribe('camera_status_update', handler1);
      notificationService.unsubscribe('camera_status_update', handler2);

      // Assert
      expect(notificationService.getSubscribedMethods()).not.toContain('camera_status_update');
    });
  });

  describe('REQ-NOTIF-005: WebSocket integration', () => {
    test('Should setup WebSocket handlers on construction', () => {
      // Arrange - Use centralized mock instead of local duplication
      const mockWsService = MockDataFactory.createMockWebSocketService();
      mockWsService.events = {} as any;

      // Act
      new NotificationService(mockWsService);

      // Assert
      expect(mockWsService.events.onNotification).toBeDefined();
      expect(typeof mockWsService.events.onNotification).toBe('function');
    });

    test('Should handle WebSocket notifications through setup handler', () => {
      // Arrange - Use centralized mock instead of local duplication
      const mockWsService = MockDataFactory.createMockWebSocketService();
      mockWsService.events = {} as any;
      const service = new NotificationService(mockWsService);
      const notification: JsonRpcNotification = {
        jsonrpc: '2.0',
        method: 'camera_status_update',
        params: {
          device: 'camera0',
          status: 'CONNECTED'
        }
      };

      service.subscribe('camera_status_update', mockHandler);

      // Act
      mockWsService.events.onNotification(notification);

      // Assert
      expect(mockHandler).toHaveBeenCalledWith(notification);
    });
  });

  describe('REQ-NOTIF-006: Service state management', () => {
    test('Should return correct handler count for method', () => {
      // Arrange
      const handler1 = MockDataFactory.createMockEventHandler();
      const handler2 = MockDataFactory.createMockEventHandler();
      const handler3 = MockDataFactory.createMockEventHandler();

      // Act
      notificationService.subscribe('camera_status_update', handler1);
      notificationService.subscribe('camera_status_update', handler2);
      notificationService.subscribe('recording_status_update', handler3);

      // Assert
      expect(notificationService.getHandlerCount('camera_status_update')).toBe(2);
      expect(notificationService.getHandlerCount('recording_status_update')).toBe(1);
      expect(notificationService.getHandlerCount('non_existent')).toBe(0);
    });

    test('Should return all subscribed methods', () => {
      // Arrange
      const handler1 = MockDataFactory.createMockEventHandler();
      const handler2 = MockDataFactory.createMockEventHandler();

      // Act
      notificationService.subscribe('camera_status_update', handler1);
      notificationService.subscribe('recording_status_update', handler2);

      // Assert
      const methods = notificationService.getSubscribedMethods();
      expect(methods).toHaveLength(2);
      expect(methods).toContain('camera_status_update');
      expect(methods).toContain('recording_status_update');
    });
  });
});
