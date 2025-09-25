/**
 * ServerService unit tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * 
 * Requirements Coverage:
 * - REQ-SERVER-001: Server information retrieval
 * - REQ-SERVER-002: System status monitoring
 * - REQ-SERVER-003: Storage information
 * - REQ-SERVER-004: System metrics collection
 * - REQ-SERVER-005: Event subscription management
 * 
 * Test Categories: Unit
 * API Documentation Reference: ../mediamtx-camera-service-go/docs/api/json_rpc_methods.md
 */

import { ServerService } from '../../../src/services/server/ServerService';
import { WebSocketService } from '../../../src/services/websocket/WebSocketService';
import { MockDataFactory } from '../../utils/mocks';
import { APIResponseValidator } from '../../utils/validators';

// Use centralized mocks - eliminates duplication
const mockWebSocketService = MockDataFactory.createMockWebSocketService();

describe('ServerService Unit Tests', () => {
  let serverService: ServerService;

  beforeEach(() => {
    jest.clearAllMocks();
    mockWebSocketService.isConnected = true;
    serverService = new ServerService(mockWebSocketService);
  });

  describe('REQ-SERVER-001: Server information retrieval', () => {
    test('should get server info successfully', async () => {
      const expectedInfo = MockDataFactory.getServerInfo();
      mockWebSocketService.sendRPC.mockResolvedValue(expectedInfo);

      const result = await serverService.getServerInfo();

      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('get_server_info');
      expect(result).toEqual(expectedInfo);
      expect(APIResponseValidator.validateServerInfo(result)).toBe(true);
    });

    test('should throw error when WebSocket not connected', async () => {
      mockWebSocketService.isConnected = false;

      await expect(serverService.getServerInfo()).rejects.toThrow('WebSocket not connected');
    });

    test('should handle server info errors', async () => {
      const error = new Error('Server info unavailable');
      mockWebSocketService.sendRPC.mockRejectedValue(error);

      await expect(serverService.getServerInfo()).rejects.toThrow('Server info unavailable');
    });

    test('should validate server info structure', async () => {
      const serverInfo = MockDataFactory.getServerInfo();
      mockWebSocketService.sendRPC.mockResolvedValue(serverInfo);

      const result = await serverService.getServerInfo();

      expect(result.name).toBeDefined();
      expect(result.version).toBeDefined();
      expect(result.build_date).toBeDefined();
      expect(result.go_version).toBeDefined();
      expect(result.architecture).toBeDefined();
      expect(Array.isArray(result.capabilities)).toBe(true);
      expect(Array.isArray(result.supported_formats)).toBe(true);
      expect(typeof result.max_cameras).toBe('number');
    });
  });

  describe('REQ-SERVER-002: System status monitoring', () => {
    test('should get system status successfully', async () => {
      const expectedStatus = MockDataFactory.getSystemStatus();
      mockWebSocketService.sendRPC.mockResolvedValue(expectedStatus);

      const result = await serverService.getStatus();

      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('get_status');
      expect(result).toEqual(expectedStatus);
      expect(APIResponseValidator.validateSystemStatus(result)).toBe(true);
    });

    test('should throw error when WebSocket not connected', async () => {
      mockWebSocketService.isConnected = false;

      await expect(serverService.getStatus()).rejects.toThrow('WebSocket not connected');
    });

    test('should handle status errors', async () => {
      const error = new Error('Status unavailable');
      mockWebSocketService.sendRPC.mockRejectedValue(error);

      await expect(serverService.getStatus()).rejects.toThrow('Status unavailable');
    });

    test('should validate status structure', async () => {
      const status = MockDataFactory.getServerInfo();
      mockWebSocketService.sendRPC.mockResolvedValue(status);

      const result = await serverService.getStatus();

      expect(['HEALTHY', 'DEGRADED', 'UNHEALTHY']).toContain(result.status);
      expect(typeof result.uptime).toBe('number');
      expect(typeof result.version).toBe('string');
      expect(typeof result.components).toBe('object');
    });
  });

  describe('REQ-SERVER-003: Storage information', () => {
    test('should get storage info successfully', async () => {
      const expectedStorage = MockDataFactory.getStorageInfo();
      mockWebSocketService.sendRPC.mockResolvedValue(expectedStorage);

      const result = await serverService.getStorageInfo();

      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('get_storage_info');
      expect(result).toEqual(expectedStorage);
      expect(APIResponseValidator.validateStorageInfo(result)).toBe(true);
    });

    test('should throw error when WebSocket not connected', async () => {
      mockWebSocketService.isConnected = false;

      await expect(serverService.getStorageInfo()).rejects.toThrow('WebSocket not connected');
    });

    test('should handle storage info errors', async () => {
      const error = new Error('Storage info unavailable');
      mockWebSocketService.sendRPC.mockRejectedValue(error);

      await expect(serverService.getStorageInfo()).rejects.toThrow('Storage info unavailable');
    });

    test('should validate storage info structure', async () => {
      const storage = MockDataFactory.getStorageInfo();
      mockWebSocketService.sendRPC.mockResolvedValue(storage);

      const result = await serverService.getStorageInfo();

      expect(typeof result.total_space).toBe('number');
      expect(typeof result.used_space).toBe('number');
      expect(typeof result.available_space).toBe('number');
      expect(typeof result.usage_percentage).toBe('number');
      expect(typeof result.recordings_size).toBe('number');
      expect(typeof result.snapshots_size).toBe('number');
      expect(typeof result.low_space_warning).toBe('boolean');
    });
  });

  describe('REQ-SERVER-004: System metrics collection', () => {
    test('should get metrics successfully', async () => {
      const expectedMetrics = MockDataFactory.getMetricsResult();
      mockWebSocketService.sendRPC.mockResolvedValue(expectedMetrics);

      const result = await serverService.getMetrics();

      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('get_metrics');
      expect(result).toEqual(expectedMetrics);
      expect(APIResponseValidator.validateMetricsResult(result)).toBe(true);
    });

    test('should throw error when WebSocket not connected', async () => {
      mockWebSocketService.isConnected = false;

      await expect(serverService.getMetrics()).rejects.toThrow('WebSocket not connected');
    });

    test('should handle metrics errors', async () => {
      const error = new Error('Metrics unavailable');
      mockWebSocketService.sendRPC.mockRejectedValue(error);

      await expect(serverService.getMetrics()).rejects.toThrow('Metrics unavailable');
    });

    test('should validate metrics structure', async () => {
      const metrics = MockDataFactory.getMetricsResult();
      mockWebSocketService.sendRPC.mockResolvedValue(metrics);

      const result = await serverService.getMetrics();

      expect(APIResponseValidator.validateIsoTimestamp(result.timestamp)).toBe(true);
      expect(typeof result.system_metrics).toBe('object');
      expect(typeof result.camera_metrics).toBe('object');
      expect(typeof result.recording_metrics).toBe('object');
      expect(typeof result.stream_metrics).toBe('object');
    });
  });

  describe('REQ-SERVER-005: Event subscription management', () => {
    test('should subscribe to events successfully', async () => {
      const topics = ['camera_status_update', 'recording_complete'];
      const filters = { device: 'camera0' };
      const expectedResult = {
        subscribed: true,
        topics,
        subscription_id: 'sub-123'
      };
      
      mockWebSocketService.sendRPC.mockResolvedValue(expectedResult);

      const result = await serverService.subscribeEvents(topics, filters);

      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('subscribe_events', {
        topics,
        filters
      });
      expect(result).toEqual(expectedResult);
    });

    test('should subscribe to events without filters', async () => {
      const topics = ['system_status_update'];
      const expectedResult = {
        subscribed: true,
        topics,
        subscription_id: 'sub-456'
      };
      
      mockWebSocketService.sendRPC.mockResolvedValue(expectedResult);

      const result = await serverService.subscribeEvents(topics);

      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('subscribe_events', {
        topics,
        filters: undefined
      });
      expect(result).toEqual(expectedResult);
    });

    test('should unsubscribe from events successfully', async () => {
      const topics = ['camera_status_update'];
      const expectedResult = {
        unsubscribed: true,
        topics,
        subscription_id: 'sub-123'
      };
      
      mockWebSocketService.sendRPC.mockResolvedValue(expectedResult);

      const result = await serverService.unsubscribeEvents(topics);

      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('unsubscribe_events', {
        topics
      });
      expect(result).toEqual(expectedResult);
    });

    test('should unsubscribe from all events', async () => {
      const expectedResult = {
        unsubscribed: true,
        topics: null,
        subscription_id: 'sub-123'
      };
      
      mockWebSocketService.sendRPC.mockResolvedValue(expectedResult);

      const result = await serverService.unsubscribeEvents();

      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('unsubscribe_events', {
        topics: undefined
      });
      expect(result).toEqual(expectedResult);
    });

    test('should get subscription stats', async () => {
      const expectedStats = {
        active_subscriptions: 2,
        total_events_received: 150,
        topics: {
          'camera_status_update': 100,
          'recording_complete': 50
        }
      };
      
      mockWebSocketService.sendRPC.mockResolvedValue(expectedStats);

      const result = await serverService.getSubscriptionStats();

      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('get_subscription_stats');
      expect(result).toEqual(expectedStats);
    });

    test('should throw error when WebSocket not connected for subscriptions', async () => {
      mockWebSocketService.isConnected = false;

      await expect(serverService.subscribeEvents(['test'])).rejects.toThrow('WebSocket not connected');
      await expect(serverService.unsubscribeEvents(['test'])).rejects.toThrow('WebSocket not connected');
      await expect(serverService.getSubscriptionStats()).rejects.toThrow('WebSocket not connected');
    });

    test('should handle subscription errors', async () => {
      const error = new Error('Subscription failed');
      mockWebSocketService.sendRPC.mockRejectedValue(error);

      await expect(serverService.subscribeEvents(['test'])).rejects.toThrow('Subscription failed');
      await expect(serverService.unsubscribeEvents(['test'])).rejects.toThrow('Subscription failed');
      await expect(serverService.getSubscriptionStats()).rejects.toThrow('Subscription failed');
    });
  });

  describe('Ping functionality', () => {
    test('should ping server successfully', async () => {
      const expectedPong = 'pong';
      mockWebSocketService.sendRPC.mockResolvedValue(expectedPong);

      const result = await serverService.ping();

      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('ping');
      expect(result).toBe(expectedPong);
    });

    test('should throw error when WebSocket not connected for ping', async () => {
      mockWebSocketService.isConnected = false;

      await expect(serverService.ping()).rejects.toThrow('WebSocket not connected');
    });

    test('should handle ping errors', async () => {
      const error = new Error('Ping failed');
      mockWebSocketService.sendRPC.mockRejectedValue(error);

      await expect(serverService.ping()).rejects.toThrow('Ping failed');
    });
  });

  describe('Error handling', () => {
    test('should handle WebSocket connection loss', async () => {
      mockWebSocketService.isConnected = false;

      await expect(serverService.getServerInfo()).rejects.toThrow('WebSocket not connected');
      await expect(serverService.getStatus()).rejects.toThrow('WebSocket not connected');
      await expect(serverService.getStorageInfo()).rejects.toThrow('WebSocket not connected');
      await expect(serverService.getMetrics()).rejects.toThrow('WebSocket not connected');
      await expect(serverService.ping()).rejects.toThrow('WebSocket not connected');
    });

    test('should handle RPC errors', async () => {
      const error = new Error('RPC method failed');
      mockWebSocketService.sendRPC.mockRejectedValue(error);

      await expect(serverService.getServerInfo()).rejects.toThrow('RPC method failed');
      await expect(serverService.getStatus()).rejects.toThrow('RPC method failed');
      await expect(serverService.getStorageInfo()).rejects.toThrow('RPC method failed');
      await expect(serverService.getMetrics()).rejects.toThrow('RPC method failed');
      await expect(serverService.ping()).rejects.toThrow('RPC method failed');
    });
  });
});
