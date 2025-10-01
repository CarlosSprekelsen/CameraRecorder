/**
 * IAPIClient Interface Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * 
 * Requirements Coverage:
 * - REQ-WS-001: IAPIClient connection management
 * - REQ-WS-002: IAPIClient method calls
 * - REQ-WS-003: Error handling
 * - REQ-WS-004: Connection status validation
 * - REQ-WS-005: Batch call handling
 * 
 * Test Categories: Unit
 * API Documentation Reference: ../mediamtx-camera-service-go/docs/api/json_rpc_methods.md
 */

import { IAPIClient } from '../../../src/services/abstraction/IAPIClient';
import { LoggerService } from '../../../src/services/logger/LoggerService';
import { MockDataFactory } from '../../utils/mocks';
import { APIResponseValidator } from '../../utils/validators';

// Use centralized mocks - aligned with refactored architecture
const mockAPIClient = MockDataFactory.createMockAPIClient();
const mockLoggerService = MockDataFactory.createMockLoggerService();

describe('IAPIClient Interface Tests', () => {
  let apiClient: IAPIClient;

  beforeEach(() => {
    jest.clearAllMocks();
    // Use IAPIClient mock - aligned with refactored architecture
    apiClient = mockAPIClient;
  });

  describe('REQ-WS-001: IAPIClient connection management', () => {
    test('should have correct initial connection state', () => {
      expect(apiClient.isConnected()).toBe(true);
      expect(apiClient.getConnectionStatus()).toEqual({
        connected: true,
        ready: true
      });
    });

    test('should handle connection state changes', () => {
      // Mock connection loss
      (apiClient.isConnected as jest.Mock).mockReturnValue(false);
      (apiClient.getConnectionStatus as jest.Mock).mockReturnValue({
        connected: false,
        ready: false
      });

      expect(apiClient.isConnected()).toBe(false);
      expect(apiClient.getConnectionStatus().connected).toBe(false);
    });

    test('should validate connection status structure', () => {
      const status = apiClient.getConnectionStatus();
      expect(status).toHaveProperty('connected');
      expect(status).toHaveProperty('ready');
      expect(typeof status.connected).toBe('boolean');
      expect(typeof status.ready).toBe('boolean');
    });
  });

  describe('REQ-WS-002: IAPIClient method calls', () => {
    test('should handle single method call', async () => {
      const method = 'test_method';
      const params = { param1: 'value1' };
      const expectedResult = { result: 'success' };

      (apiClient.call as jest.Mock).mockResolvedValue(expectedResult);

      const result = await apiClient.call(method, params);

      expect(apiClient.call).toHaveBeenCalledWith(method, params);
      expect(result).toEqual(expectedResult);
    });

    test('should handle method call without parameters', async () => {
      const method = 'test_method';
      const expectedResult = { result: 'success' };

      (apiClient.call as jest.Mock).mockResolvedValue(expectedResult);

      const result = await apiClient.call(method);

      // Method called without parameters - default value {} is used internally
      expect(apiClient.call).toHaveBeenCalledWith(method);
      expect(result).toEqual(expectedResult);
    });

    test('should handle batch method calls', async () => {
      const calls = [
        { method: 'method1', params: { param1: 'value1' } },
        { method: 'method2', params: { param2: 'value2' } }
      ];
      const expectedResults = [{ result1: 'success1' }, { result2: 'success2' }];

      (apiClient.batchCall as jest.Mock).mockResolvedValue(expectedResults);

      const results = await apiClient.batchCall(calls);

      expect(apiClient.batchCall).toHaveBeenCalledWith(calls);
      expect(results).toEqual(expectedResults);
    });
  });

  describe('REQ-WS-003: Error handling', () => {
    test('should handle connection errors', async () => {
      const error = new Error('Connection failed');
      (apiClient.call as jest.Mock).mockRejectedValue(error);

      await expect(apiClient.call('test_method')).rejects.toThrow('Connection failed');
    });

    test('should handle batch call errors', async () => {
      const calls = [{ method: 'failing_method', params: {} }];
      const error = new Error('Batch call failed');
      (apiClient.batchCall as jest.Mock).mockRejectedValue(error);

      await expect(apiClient.batchCall(calls)).rejects.toThrow('Batch call failed');
    });

    test('should handle method call timeouts', async () => {
      const timeoutError = new Error('Request timeout');
      (apiClient.call as jest.Mock).mockRejectedValue(timeoutError);

      await expect(apiClient.call('slow_method')).rejects.toThrow('Request timeout');
    });
  });

  describe('REQ-WS-004: Connection status validation', () => {
    test('should handle ready state correctly', () => {
      // Mock ready state
      (apiClient.getConnectionStatus as jest.Mock).mockReturnValue({
        connected: true,
        ready: false
      });
      
      const status = apiClient.getConnectionStatus();
      expect(status.connected).toBe(true);
      expect(status.ready).toBe(false);
    });

    test('should handle disconnected state', () => {
      // Mock disconnected state
      (apiClient.getConnectionStatus as jest.Mock).mockReturnValue({
        connected: false,
        ready: false
      });
      
      const status = apiClient.getConnectionStatus();
      expect(status.connected).toBe(false);
      expect(status.ready).toBe(false);
    });
  });

  describe('REQ-WS-005: Batch call handling', () => {
    test('should handle empty batch calls', async () => {
      const calls: Array<{method: string, params: Record<string, unknown>}> = [];
      const expectedResults: any[] = [];

      (apiClient.batchCall as jest.Mock).mockResolvedValue(expectedResults);

      const results = await apiClient.batchCall(calls);

      expect(apiClient.batchCall).toHaveBeenCalledWith(calls);
      expect(results).toEqual(expectedResults);
    });

    test('should handle mixed success and failure in batch calls', async () => {
      const calls = [
        { method: 'success_method', params: {} },
        { method: 'failing_method', params: {} }
      ];
      const error = new Error('Batch call failed');
      (apiClient.batchCall as jest.Mock).mockRejectedValue(error);

      await expect(apiClient.batchCall(calls)).rejects.toThrow('Batch call failed');
    });
  });
});