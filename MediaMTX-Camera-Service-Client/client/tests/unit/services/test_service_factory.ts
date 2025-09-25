/**
 * ServiceFactory unit tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * 
 * Requirements Coverage:
 * - REQ-FACTORY-001: Singleton pattern implementation
 * - REQ-FACTORY-002: Service creation and caching
 * - REQ-FACTORY-003: Service retrieval
 * - REQ-FACTORY-004: Factory reset functionality
 * - REQ-FACTORY-005: Dependency injection
 * 
 * Test Categories: Unit
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */

import { ServiceFactory } from '../../../src/services/ServiceFactory';
import { WebSocketService } from '../../../src/services/websocket/WebSocketService';
import { AuthService } from '../../../src/services/auth/AuthService';
import { ServerService } from '../../../src/services/server/ServerService';
import { NotificationService } from '../../../src/services/notifications/NotificationService';
import { DeviceService } from '../../../src/services/device/DeviceService';
import { RecordingService } from '../../../src/services/recording/RecordingService';
import { FileService } from '../../../src/services/file/FileService';

// Use centralized mocks - eliminates duplication
import { MockDataFactory } from '../../utils/mocks';

// Mock all services using centralized approach - return constructor functions
jest.mock('../../../src/services/websocket/WebSocketService', () => ({
  WebSocketService: jest.fn().mockImplementation(() => MockDataFactory.createMockWebSocketService())
}));
jest.mock('../../../src/services/auth/AuthService', () => ({
  AuthService: jest.fn().mockImplementation(() => MockDataFactory.createMockAuthService())
}));
jest.mock('../../../src/services/server/ServerService', () => ({
  ServerService: jest.fn().mockImplementation(() => MockDataFactory.createMockServerService())
}));
jest.mock('../../../src/services/notifications/NotificationService', () => ({
  NotificationService: jest.fn().mockImplementation(() => MockDataFactory.createMockEventHandler())
}));
jest.mock('../../../src/services/device/DeviceService', () => ({
  DeviceService: jest.fn().mockImplementation(() => MockDataFactory.createMockDeviceService())
}));
jest.mock('../../../src/services/recording/RecordingService', () => ({
  RecordingService: jest.fn().mockImplementation(() => MockDataFactory.createMockRecordingService())
}));
jest.mock('../../../src/services/file/FileService', () => ({
  FileService: jest.fn().mockImplementation(() => MockDataFactory.createMockFileService())
}));
jest.mock('../../../src/services/logger/LoggerService', () => ({
  logger: {
    info: jest.fn(),
    warn: jest.fn(),
    error: jest.fn(),
    debug: jest.fn()
  }
}));

describe('ServiceFactory Unit Tests', () => {
  let factory: ServiceFactory;

  beforeEach(() => {
    // Reset factory instance for each test
    factory = ServiceFactory.getInstance();
    factory.reset();
  });

  describe('REQ-FACTORY-001: Singleton pattern implementation', () => {
    test('Should return same instance on multiple calls', () => {
      // Act
      const instance1 = ServiceFactory.getInstance();
      const instance2 = ServiceFactory.getInstance();

      // Assert
      expect(instance1).toBe(instance2);
      expect(instance1).toBeInstanceOf(ServiceFactory);
    });

    test('Should maintain state across multiple getInstance calls', () => {
      // Arrange
      const url = 'ws://localhost:8002/ws';
      const wsService = factory.createWebSocketService(url);

      // Act
      const newFactory = ServiceFactory.getInstance();
      const retrievedWsService = newFactory.getWebSocketService();

      // Assert
      expect(retrievedWsService).toBe(wsService);
    });
  });

  describe('REQ-FACTORY-002: Service creation and caching', () => {
    test('Should create WebSocket service on first call', () => {
      // Arrange
      const url = 'ws://localhost:8002/ws';

      // Act
      const wsService = factory.createWebSocketService(url);

      // Assert
      expect(wsService).toBeDefined();
      expect(WebSocketService).toHaveBeenCalledWith({ url });
    });

    test('Should return cached WebSocket service on subsequent calls', () => {
      // Arrange
      const url = 'ws://localhost:8002/ws';
      const firstService = factory.createWebSocketService(url);

      // Act
      const secondService = factory.createWebSocketService(url);

      // Assert
      expect(secondService).toBe(firstService);
      expect(WebSocketService).toHaveBeenCalledTimes(1);
    });

    test('Should create AuthService with WebSocket dependency', () => {
      // Arrange
      const url = 'ws://localhost:8002/ws';
      const wsService = factory.createWebSocketService(url);

      // Act
      const authService = factory.createAuthService(wsService);

      // Assert
      expect(authService).toBeDefined();
      expect(AuthService).toHaveBeenCalledWith(wsService);
    });

    test('Should create ServerService with WebSocket dependency', () => {
      // Arrange
      const url = 'ws://localhost:8002/ws';
      const wsService = factory.createWebSocketService(url);

      // Act
      const serverService = factory.createServerService(wsService);

      // Assert
      expect(serverService).toBeDefined();
      expect(ServerService).toHaveBeenCalledWith(wsService);
    });

    test('Should create NotificationService with WebSocket dependency', () => {
      // Arrange
      const url = 'ws://localhost:8002/ws';
      const wsService = factory.createWebSocketService(url);

      // Act
      const notificationService = factory.createNotificationService(wsService);

      // Assert
      expect(notificationService).toBeDefined();
      expect(NotificationService).toHaveBeenCalledWith(wsService);
    });

    test('Should create DeviceService with WebSocket and logger dependencies', () => {
      // Arrange
      const url = 'ws://localhost:8002/ws';
      const wsService = factory.createWebSocketService(url);

      // Act
      const deviceService = factory.createDeviceService(wsService);

      // Assert
      expect(deviceService).toBeDefined();
      expect(DeviceService).toHaveBeenCalledWith(wsService, expect.any(Object));
    });

    test('Should create RecordingService with WebSocket and logger dependencies', () => {
      // Arrange
      const url = 'ws://localhost:8002/ws';
      const wsService = factory.createWebSocketService(url);

      // Act
      const recordingService = factory.createRecordingService(wsService);

      // Assert
      expect(recordingService).toBeDefined();
      expect(RecordingService).toHaveBeenCalledWith(wsService, expect.any(Object));
    });

    test('Should create FileService with WebSocket and logger dependencies', () => {
      // Arrange
      const url = 'ws://localhost:8002/ws';
      const wsService = factory.createWebSocketService(url);

      // Act
      const fileService = factory.createFileService(wsService);

      // Assert
      expect(fileService).toBeDefined();
      expect(FileService).toHaveBeenCalledWith(wsService, expect.any(Object));
    });
  });

  describe('REQ-FACTORY-003: Service retrieval', () => {
    test('Should return null for uncreated services', () => {
      // Assert
      expect(factory.getWebSocketService()).toBeNull();
      expect(factory.getAuthService()).toBeNull();
      expect(factory.getServerService()).toBeNull();
      expect(factory.getNotificationService()).toBeNull();
      expect(factory.getDeviceService()).toBeNull();
      expect(factory.getRecordingService()).toBeNull();
      expect(factory.getFileService()).toBeNull();
    });

    test('Should return created services', () => {
      // Arrange
      const url = 'ws://localhost:8002/ws';
      const wsService = factory.createWebSocketService(url);
      const authService = factory.createAuthService(wsService);
      const serverService = factory.createServerService(wsService);
      const notificationService = factory.createNotificationService(wsService);
      const deviceService = factory.createDeviceService(wsService);
      const recordingService = factory.createRecordingService(wsService);
      const fileService = factory.createFileService(wsService);

      // Act & Assert
      expect(factory.getWebSocketService()).toBe(wsService);
      expect(factory.getAuthService()).toBe(authService);
      expect(factory.getServerService()).toBe(serverService);
      expect(factory.getNotificationService()).toBe(notificationService);
      expect(factory.getDeviceService()).toBe(deviceService);
      expect(factory.getRecordingService()).toBe(recordingService);
      expect(factory.getFileService()).toBe(fileService);
    });
  });

  describe('REQ-FACTORY-004: Factory reset functionality', () => {
    test('Should reset all services to null', () => {
      // Arrange
      const url = 'ws://localhost:8002/ws';
      const wsService = factory.createWebSocketService(url);
      factory.createAuthService(wsService);
      factory.createServerService(wsService);
      factory.createNotificationService(wsService);
      factory.createDeviceService(wsService);
      factory.createRecordingService(wsService);
      factory.createFileService(wsService);

      // Act
      factory.reset();

      // Assert
      expect(factory.getWebSocketService()).toBeNull();
      expect(factory.getAuthService()).toBeNull();
      expect(factory.getServerService()).toBeNull();
      expect(factory.getNotificationService()).toBeNull();
      expect(factory.getDeviceService()).toBeNull();
      expect(factory.getRecordingService()).toBeNull();
      expect(factory.getFileService()).toBeNull();
    });

    test('Should allow recreation of services after reset', () => {
      // Arrange
      const url = 'ws://localhost:8002/ws';
      factory.createWebSocketService(url);
      factory.reset();

      // Act
      const newWsService = factory.createWebSocketService(url);

      // Assert
      expect(newWsService).toBeDefined();
      expect(factory.getWebSocketService()).toBe(newWsService);
    });
  });

  describe('REQ-FACTORY-005: Dependency injection', () => {
    test('Should inject WebSocket service into all dependent services', () => {
      // Arrange
      const url = 'ws://localhost:8002/ws';
      const wsService = factory.createWebSocketService(url);

      // Act
      factory.createAuthService(wsService);
      factory.createServerService(wsService);
      factory.createNotificationService(wsService);
      factory.createDeviceService(wsService);
      factory.createRecordingService(wsService);
      factory.createFileService(wsService);

      // Assert
      expect(AuthService).toHaveBeenCalledWith(wsService);
      expect(ServerService).toHaveBeenCalledWith(wsService);
      expect(NotificationService).toHaveBeenCalledWith(wsService);
      expect(DeviceService).toHaveBeenCalledWith(wsService, expect.any(Object));
      expect(RecordingService).toHaveBeenCalledWith(wsService, expect.any(Object));
      expect(FileService).toHaveBeenCalledWith(wsService, expect.any(Object));
    });

    test('Should inject logger into services that require it', () => {
      // Arrange
      const url = 'ws://localhost:8002/ws';
      const wsService = factory.createWebSocketService(url);

      // Act
      factory.createDeviceService(wsService);
      factory.createRecordingService(wsService);
      factory.createFileService(wsService);

      // Assert
      expect(DeviceService).toHaveBeenCalledWith(wsService, expect.any(Object));
      expect(RecordingService).toHaveBeenCalledWith(wsService, expect.any(Object));
      expect(FileService).toHaveBeenCalledWith(wsService, expect.any(Object));
    });
  });

  describe('REQ-FACTORY-006: Service lifecycle management', () => {
    test('Should maintain service instances across multiple factory calls', () => {
      // Arrange
      const url = 'ws://localhost:8002/ws';
      const wsService = factory.createWebSocketService(url);
      const authService = factory.createAuthService(wsService);

      // Act
      const newFactory = ServiceFactory.getInstance();
      const retrievedAuthService = newFactory.getAuthService();

      // Assert
      expect(retrievedAuthService).toBe(authService);
    });

    test('Should handle service creation with different WebSocket instances', () => {
      // Arrange
      const url1 = 'ws://localhost:8002/ws';
      const url2 = 'ws://localhost:8003/ws';

      // Act
      const wsService1 = factory.createWebSocketService(url1);
      factory.reset();
      const wsService2 = factory.createWebSocketService(url2);

      // Assert
      expect(wsService1).not.toBe(wsService2);
      expect(WebSocketService).toHaveBeenCalledWith({ url: url1 });
      expect(WebSocketService).toHaveBeenCalledWith({ url: url2 });
    });
  });
});
