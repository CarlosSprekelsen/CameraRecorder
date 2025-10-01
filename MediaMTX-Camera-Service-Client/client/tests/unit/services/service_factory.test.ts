/**
 * ServiceFactory unit tests - DRY Refactored
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
import { ServiceFactoryTestHelper, TestConstants } from '../../utils/service-factory-test-helper';

// Setup common mocks using DRY utilities
ServiceFactoryTestHelper.setupCommonMocks();

describe('ServiceFactory Unit Tests - DRY Refactored', () => {
  let factory: ServiceFactory;
  let mockAPIClient: any;

  beforeEach(() => {
    factory = ServiceFactoryTestHelper.createFreshFactory();
    mockAPIClient = ServiceFactoryTestHelper.createMockAPIClient();
  });

  describe('REQ-FACTORY-001: Singleton pattern implementation', () => {
    test('Should return same instance on multiple calls', () => {
      ServiceFactoryTestHelper.validateSingletonPattern();
    });

    test('Should maintain state across multiple getInstance calls', () => {
      // Arrange
      const apiClient = factory.createAPIClient(ServiceFactoryTestHelper.createMockWebSocketService());

      // Act
      const newFactory = ServiceFactory.getInstance();
      const retrievedAPIClient = newFactory.getAPIClient();

      // Assert
      expect(retrievedAPIClient).toBe(apiClient);
    });
  });

  describe('REQ-FACTORY-002: Service creation and caching', () => {
    test('Should create API client on first call', () => {
      // Act
      const apiClient = factory.createAPIClient(ServiceFactoryTestHelper.createMockWebSocketService());

      // Assert
      expect(apiClient).toBeDefined();
      expect(factory.getAPIClient()).toBe(apiClient);
    });

    test('Should return cached API client on subsequent calls', () => {
      // Arrange
      const mockWsService = ServiceFactoryTestHelper.createMockWebSocketService();
      const firstClient = factory.createAPIClient(mockWsService);

      // Act
      const secondClient = factory.createAPIClient(mockWsService);

      // Assert
      expect(secondClient).toBe(firstClient);
    });

    // Parameterized tests for all services
    describe.each(TestConstants.SERVICE_CONFIGS)(
      'Service creation for $serviceName',
      (config) => {
        test(`Should create ${config.serviceName} with API client dependency`, () => {
          // Act
          const createdService = (factory as any)[config.createMethod](mockAPIClient);

          // Assert
          ServiceFactoryTestHelper.validateServiceCreation(factory, config, mockAPIClient, createdService);
        });

        test(`Should return cached ${config.serviceName} on subsequent calls`, () => {
          // Arrange
          const firstService = (factory as any)[config.createMethod](mockAPIClient);

          // Act
          const secondService = (factory as any)[config.createMethod](mockAPIClient);

          // Assert
          ServiceFactoryTestHelper.validateServiceCaching(factory, config, mockAPIClient);
          expect(secondService).toBe(firstService);
        });
      }
    );
  });

  describe('REQ-FACTORY-003: Service retrieval', () => {
    test('Should return null for uncreated services', () => {
      // Assert
      ServiceFactoryTestHelper.validateFactoryReset(factory, TestConstants.SERVICE_CONFIGS);
      expect(factory.getAPIClient()).toBeNull();
    });

    test('Should return created services', () => {
      // Arrange
      const apiClient = factory.createAPIClient(ServiceFactoryTestHelper.createMockWebSocketService());
      const createdServices = TestConstants.SERVICE_CONFIGS.map(config => {
        const service = (factory as any)[config.createMethod](apiClient);
        return { config, service };
      });

      // Act & Assert
      createdServices.forEach(({ config, service }) => {
        ServiceFactoryTestHelper.validateServiceRetrieval(factory, config, service);
      });
      expect(factory.getAPIClient()).toBe(apiClient);
    });
  });

  describe('REQ-FACTORY-004: Factory reset functionality', () => {
    test('Should reset all services to null', () => {
      // Arrange
      const apiClient = factory.createAPIClient(ServiceFactoryTestHelper.createMockWebSocketService());
      TestConstants.SERVICE_CONFIGS.forEach(config => {
        (factory as any)[config.createMethod](apiClient);
      });

      // Act
      factory.reset();

      // Assert
      ServiceFactoryTestHelper.validateFactoryReset(factory, TestConstants.SERVICE_CONFIGS);
      expect(factory.getAPIClient()).toBeNull();
    });

    test('Should allow recreation of services after reset', () => {
      // Arrange
      const apiClient = factory.createAPIClient(ServiceFactoryTestHelper.createMockWebSocketService());
      factory.reset();

      // Act
      const newAPIClient = factory.createAPIClient(ServiceFactoryTestHelper.createMockWebSocketService());

      // Assert
      expect(newAPIClient).toBeDefined();
      expect(factory.getAPIClient()).toBe(newAPIClient);
    });
  });

  describe('REQ-FACTORY-005: Dependency injection', () => {
    test('Should inject API client into all dependent services', () => {
      // Act
      TestConstants.SERVICE_CONFIGS.forEach(config => {
        (factory as any)[config.createMethod](mockAPIClient);
      });

      // Assert
      ServiceFactoryTestHelper.validateDependencyInjection(factory, TestConstants.SERVICE_CONFIGS, mockAPIClient);
    });
  });

  describe('REQ-FACTORY-006: Service lifecycle management', () => {
    test('Should maintain service instances across multiple factory calls', () => {
      // Arrange
      const apiClient = factory.createAPIClient(ServiceFactoryTestHelper.createMockWebSocketService());
      const authService = factory.createAuthService(apiClient);

      // Act
      const newFactory = ServiceFactory.getInstance();
      const retrievedAuthService = newFactory.getAuthService();

      // Assert
      expect(retrievedAuthService).toBe(authService);
    });

    test('Should handle service creation with different WebSocket instances', () => {
      // Arrange
      const wsService1 = ServiceFactoryTestHelper.createMockWebSocketService();
      const wsService2 = ServiceFactoryTestHelper.createMockWebSocketService();

      // Act
      const apiClient1 = factory.createAPIClient(wsService1);
      factory.reset();
      const apiClient2 = factory.createAPIClient(wsService2);

      // Assert
      expect(apiClient1).not.toBe(apiClient2);
      expect(factory.getAPIClient()).toBe(apiClient2);
    });
  });
});
