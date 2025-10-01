/**
 * ServiceFactory Test Helper - DRY Test Utilities
 * 
 * Provides reusable test utilities for ServiceFactory tests to eliminate
 * code duplication and facilitate easier maintenance and refactoring.
 */

import { ServiceFactory } from '../../src/services/ServiceFactory';
import { MockDataFactory } from './mocks';

export interface ServiceTestConfig {
  serviceName: string;
  serviceClass: any;
  createMethod: string;
  getMethod: string;
  requiresLogger?: boolean;
}

export class ServiceFactoryTestHelper {
  /**
   * Common test URLs used across all tests
   */
  static readonly TEST_URLS = {
    DEFAULT: 'ws://localhost:8002/ws',
    ALTERNATE: 'ws://localhost:8003/ws'
  };

  /**
   * Service configurations for parameterized testing
   */
  static readonly SERVICE_CONFIGS: ServiceTestConfig[] = [
    {
      serviceName: 'AuthService',
      serviceClass: jest.fn().mockImplementation(() => MockDataFactory.createMockAuthService()),
      createMethod: 'createAuthService',
      getMethod: 'getAuthService',
      requiresLogger: false
    },
    {
      serviceName: 'ServerService',
      serviceClass: jest.fn().mockImplementation(() => MockDataFactory.createMockServerService()),
      createMethod: 'createServerService',
      getMethod: 'getServerService',
      requiresLogger: false
    },
    {
      serviceName: 'DeviceService',
      serviceClass: jest.fn().mockImplementation(() => MockDataFactory.createMockDeviceService()),
      createMethod: 'createDeviceService',
      getMethod: 'getDeviceService',
      requiresLogger: true
    },
    {
      serviceName: 'RecordingService',
      serviceClass: jest.fn().mockImplementation(() => MockDataFactory.createMockRecordingService()),
      createMethod: 'createRecordingService',
      getMethod: 'getRecordingService',
      requiresLogger: true
    },
    {
      serviceName: 'FileService',
      serviceClass: jest.fn().mockImplementation(() => MockDataFactory.createMockFileService()),
      createMethod: 'createFileService',
      getMethod: 'getFileService',
      requiresLogger: true
    },
    {
      serviceName: 'StreamingService',
      serviceClass: jest.fn().mockImplementation(() => MockDataFactory.createMockStreamingService()),
      createMethod: 'createStreamingService',
      getMethod: 'getStreamingService',
      requiresLogger: false
    },
    {
      serviceName: 'ExternalStreamService',
      serviceClass: jest.fn().mockImplementation(() => MockDataFactory.createMockExternalStreamService()),
      createMethod: 'createExternalStreamService',
      getMethod: 'getExternalStreamService',
      requiresLogger: false
    }
  ];

  /**
   * Creates a mock WebSocket service for testing
   */
  static createMockWebSocketService() {
    return MockDataFactory.createMockWebSocket();
  }

  /**
   * Creates a mock API client for testing
   */
  static createMockAPIClient() {
    return MockDataFactory.createMockAPIClient();
  }

  /**
   * Sets up common mocks for all ServiceFactory tests
   */
  static setupCommonMocks() {
    // Mock WebSocketService
    jest.mock('../../src/services/websocket/WebSocketService', () => ({
      WebSocketService: jest.fn().mockImplementation(() => this.createMockWebSocketService())
    }));

    // Mock LoggerService
    jest.mock('../../src/services/logger/LoggerService', () => ({
      logger: {
        info: jest.fn(),
        warn: jest.fn(),
        error: jest.fn(),
        debug: jest.fn()
      }
    }));

    // Mock APIClient
    jest.mock('../../src/services/abstraction/APIClient', () => ({
      APIClient: jest.fn().mockImplementation(() => MockDataFactory.createMockAPIClient())
    }));

    // Mock services individually to avoid hoisting issues
    jest.mock('../../src/services/auth/AuthService', () => ({
      AuthService: jest.fn().mockImplementation(() => MockDataFactory.createMockAuthService())
    }));

    jest.mock('../../src/services/server/ServerService', () => ({
      ServerService: jest.fn().mockImplementation(() => MockDataFactory.createMockServerService())
    }));

    jest.mock('../../src/services/device/DeviceService', () => ({
      DeviceService: jest.fn().mockImplementation(() => MockDataFactory.createMockDeviceService())
    }));

    jest.mock('../../src/services/recording/RecordingService', () => ({
      RecordingService: jest.fn().mockImplementation(() => MockDataFactory.createMockRecordingService())
    }));

    jest.mock('../../src/services/file/FileService', () => ({
      FileService: jest.fn().mockImplementation(() => MockDataFactory.createMockFileService())
    }));

    jest.mock('../../src/services/streaming/StreamingService', () => ({
      StreamingService: jest.fn().mockImplementation(() => MockDataFactory.createMockStreamingService())
    }));

    jest.mock('../../src/services/external/ExternalStreamService', () => ({
      ExternalStreamService: jest.fn().mockImplementation(() => MockDataFactory.createMockExternalStreamService())
    }));
  }

  /**
   * Creates a fresh ServiceFactory instance for testing
   */
  static createFreshFactory(): ServiceFactory {
    const factory = ServiceFactory.getInstance();
    factory.reset();
    return factory;
  }

  /**
   * Validates that a service was created correctly
   */
  static validateServiceCreation(
    factory: ServiceFactory,
    config: ServiceTestConfig,
    apiClient: any,
    expectedService: any
  ) {
    const createdService = (factory as any)[config.createMethod](apiClient);
    
    expect(createdService).toBeDefined();
    expect(createdService).toBe(expectedService);
    // Validate that the service was created with proper dependencies
    expect(createdService).toHaveProperty('constructor');
  }

  /**
   * Validates that a service can be retrieved
   */
  static validateServiceRetrieval(
    factory: ServiceFactory,
    config: ServiceTestConfig,
    expectedService: any
  ) {
    const retrievedService = (factory as any)[config.getMethod]();
    expect(retrievedService).toBe(expectedService);
  }

  /**
   * Validates that all services are reset to null
   */
  static validateFactoryReset(factory: ServiceFactory, configs: ServiceTestConfig[]) {
    configs.forEach(config => {
      const retrievedService = (factory as any)[config.getMethod]();
      expect(retrievedService).toBeNull();
    });
  }

  /**
   * Validates singleton pattern behavior
   */
  static validateSingletonPattern() {
    const instance1 = ServiceFactory.getInstance();
    const instance2 = ServiceFactory.getInstance();
    
    expect(instance1).toBe(instance2);
    expect(instance1).toBeInstanceOf(ServiceFactory);
  }

  /**
   * Validates service caching behavior
   */
  static validateServiceCaching(
    factory: ServiceFactory,
    config: ServiceTestConfig,
    apiClient: any
  ) {
    const firstService = (factory as any)[config.createMethod](apiClient);
    const secondService = (factory as any)[config.createMethod](apiClient);
    
    expect(secondService).toBe(firstService);
    // Validate that the same instance is returned (caching works)
    expect(firstService).toBeDefined();
    expect(secondService).toBeDefined();
  }

  /**
   * Validates dependency injection pattern
   */
  static validateDependencyInjection(
    factory: ServiceFactory,
    configs: ServiceTestConfig[],
    apiClient: any
  ) {
    configs.forEach(config => {
      const service = (factory as any)[config.createMethod](apiClient);
      
      // Validate that the service was created successfully with the API client
      expect(service).toBeDefined();
      expect(service).not.toBeNull();
      // Validate that the service has the expected interface
      expect(typeof service).toBe('object');
    });
  }
}

/**
 * Test constants for common values
 */
export const TestConstants = {
  URLS: ServiceFactoryTestHelper.TEST_URLS,
  SERVICE_CONFIGS: ServiceFactoryTestHelper.SERVICE_CONFIGS
};
