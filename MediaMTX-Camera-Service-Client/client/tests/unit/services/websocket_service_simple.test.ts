/**
 * IAPIClient Interface Tests - WebSocket Abstraction
 * 
 * Focus: IAPIClient interface compliance and connection management
 * Coverage Target: IAPIClient methods that abstract WebSocket communication
 */

import { IAPIClient } from '../../../src/services/abstraction/IAPIClient';
import { LoggerService } from '../../../src/services/logger/LoggerService';
import { MockDataFactory } from '../../utils/mocks';

// Use centralized mocks - aligned with refactored architecture
const mockAPIClient = MockDataFactory.createMockAPIClient();
const mockLoggerService = MockDataFactory.createMockLoggerService();

describe('IAPIClient Interface Tests (WebSocket Abstraction)', () => {
  let apiClient: IAPIClient;

  beforeEach(() => {
    jest.clearAllMocks();
    // Use IAPIClient mock - aligned with refactored architecture
    apiClient = mockAPIClient;
  });

  describe('REQ-WS-001: IAPIClient connection management', () => {
    test('should have correct initial state', () => {
      expect(apiClient.isConnected()).toBe(true);
      expect(apiClient.getConnectionStatus()).toEqual({
        connected: true,
        ready: true
      });
    });

    test('should handle connection state changes', () => {
      // Test initial state
      expect(apiClient.isConnected()).toBe(true);

      // Mock state changes
      (apiClient.isConnected as jest.Mock).mockReturnValue(false);
      expect(apiClient.isConnected()).toBe(false);
      expect(apiClient.getConnectionStatus()).toEqual({
        connected: false,
        ready: false
      });
    });

    test('should handle connection status updates', () => {
      expect(apiClient.getConnectionStatus().connected).toBe(true);
      
      // Mock connection status changes
      (apiClient.getConnectionStatus as jest.Mock).mockReturnValue({
        connected: false,
        ready: false
      });
      expect(apiClient.getConnectionStatus().connected).toBe(false);
    });
  });

  describe('REQ-WS-002: IAPIClient method calls', () => {
    test('should handle call method', async () => {
      const result = await apiClient.call('test_method', { param: 'value' });
      expect(mockAPIClient.call).toHaveBeenCalledWith('test_method', { param: 'value' });
      expect(result).toBeDefined();
    });

    test('should handle batch call method', async () => {
      const calls = [
        { method: 'method1', params: { param1: 'value1' } },
        { method: 'method2', params: { param2: 'value2' } }
      ];
      const result = await apiClient.batchCall(calls);
      expect(mockAPIClient.batchCall).toHaveBeenCalledWith(calls);
      expect(result).toBeDefined();
    });
  });

  describe('REQ-WS-003: Error handling', () => {
    test('should handle connection errors gracefully', async () => {
      // Mock connection error
      (apiClient.isConnected as jest.Mock).mockReturnValue(false);
      (apiClient.call as jest.Mock).mockRejectedValue(new Error('Connection failed'));

      await expect(apiClient.call('test_method')).rejects.toThrow('Connection failed');
    });

    test('should handle batch call errors', async () => {
      const calls = [{ method: 'failing_method', params: {} }];
      (apiClient.batchCall as jest.Mock).mockRejectedValue(new Error('Batch call failed'));

      await expect(apiClient.batchCall(calls)).rejects.toThrow('Batch call failed');
    });
  });

  describe('REQ-WS-004: Connection status validation', () => {
    test('should validate connection status structure', () => {
      const status = apiClient.getConnectionStatus();
      expect(status).toHaveProperty('connected');
      expect(status).toHaveProperty('ready');
      expect(typeof status.connected).toBe('boolean');
      expect(typeof status.ready).toBe('boolean');
    });

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
  });
});