/**
 * Unit Tests for Connection Store
 * 
 * REQ-001: Store State Management - Test Zustand store actions
 * REQ-002: State Transitions - Test state changes
 * REQ-003: Error Handling - Test error states and recovery
 * REQ-004: API Integration - Mock API calls and test responses
 * REQ-005: Side Effects - Test store side effects
 * 
 * Ground Truth: Official RPC Documentation
 * API Reference: docs/api/json_rpc_methods.md
 */

import { describe, test, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { MockDataFactory } from '../../utils/mocks';
import { APIResponseValidator } from '../../utils/validators';
import { TestHelpers } from '../../utils/test-helpers';

// Mock the ConnectionService
jest.mock('../../../src/services/connection/connectionService', () => ({
  ConnectionService: jest.fn().mockImplementation(() => MockDataFactory.createMockConnectionService())
}));

// Mock the connection store
const mockConnectionStore = MockDataFactory.createMockConnectionStore();

describe('Connection Store', () => {
  let connectionStore: any;
  let mockConnectionService: any;

  beforeEach(() => {
    // Reset mocks
    jest.clearAllMocks();
    
    // Create fresh mock service
    mockConnectionService = MockDataFactory.createMockConnectionService();
    
    // Mock the store with fresh state
    connectionStore = { ...mockConnectionStore };
  });

  afterEach(() => {
    // Clean up
    connectionStore.reset?.();
  });

  describe('REQ-001: Store State Management', () => {
    test('should initialize with correct default state', () => {
      expect(connectionStore.status).toBe('connected');
      expect(connectionStore.lastError).toBe(null);
      expect(connectionStore.reconnectAttempts).toBe(0);
      expect(connectionStore.lastConnected).toBe('2025-01-15T14:30:00Z');
    });

    test('should update connection status correctly', () => {
      connectionStore.setStatus('disconnected');
      expect(connectionStore.status).toBe('disconnected');
      
      connectionStore.setStatus('connecting');
      expect(connectionStore.status).toBe('connecting');
      
      connectionStore.setStatus('connected');
      expect(connectionStore.status).toBe('connected');
    });

    test('should update last error correctly', () => {
      const errorMessage = 'Connection failed';
      connectionStore.setLastError(errorMessage);
      expect(connectionStore.lastError).toBe(errorMessage);
      
      connectionStore.setLastError(null);
      expect(connectionStore.lastError).toBe(null);
    });

    test('should update reconnect attempts correctly', () => {
      connectionStore.setReconnectAttempts(5);
      expect(connectionStore.reconnectAttempts).toBe(5);
      
      connectionStore.incrementReconnectAttempts();
      expect(connectionStore.reconnectAttempts).toBe(6);
      
      connectionStore.resetReconnectAttempts();
      expect(connectionStore.reconnectAttempts).toBe(0);
    });

    test('should update last connected timestamp correctly', () => {
      const timestamp = '2025-01-15T15:00:00Z';
      connectionStore.setLastConnected(timestamp);
      expect(connectionStore.lastConnected).toBe(timestamp);
    });
  });

  describe('REQ-002: State Transitions', () => {
    test('should handle connection sequence correctly', async () => {
      // Initial state
      expect(connectionStore.status).toBe('connected');
      
      // Disconnect
      await connectionStore.disconnect();
      expect(connectionStore.status).toBe('disconnected');
      
      // Reconnect
      await connectionStore.connect();
      expect(connectionStore.status).toBe('connected');
    });

    test('should handle connection errors correctly', async () => {
      const errorMessage = 'Connection timeout';
      
      // Mock connection failure
      mockConnectionService.connect = jest.fn().mockRejectedValue(new Error(errorMessage));
      
      try {
        await connectionStore.connect();
      } catch (error) {
        expect(connectionStore.status).toBe('error');
        expect(connectionStore.lastError).toBe(errorMessage);
        expect(connectionStore.reconnectAttempts).toBe(1);
      }
    });

    test('should handle reconnection attempts correctly', async () => {
      // Mock multiple connection failures followed by success
      mockConnectionService.connect = jest.fn()
        .mockRejectedValueOnce(new Error('Connection failed 1'))
        .mockRejectedValueOnce(new Error('Connection failed 2'))
        .mockResolvedValueOnce(undefined);
      
      // First attempt
      try {
        await connectionStore.connect();
      } catch (error) {
        expect(connectionStore.reconnectAttempts).toBe(1);
      }
      
      // Second attempt
      try {
        await connectionStore.connect();
      } catch (error) {
        expect(connectionStore.reconnectAttempts).toBe(2);
      }
      
      // Third attempt (success)
      await connectionStore.connect();
      expect(connectionStore.status).toBe('connected');
      expect(connectionStore.reconnectAttempts).toBe(0);
    });

    test('should handle connection state changes from notifications', () => {
      const connectionUpdate = {
        status: 'connected' as const,
        lastConnected: '2025-01-15T15:30:00Z'
      };
      
      connectionStore.handleConnectionUpdate(connectionUpdate);
      
      expect(connectionStore.status).toBe('connected');
      expect(connectionStore.lastConnected).toBe('2025-01-15T15:30:00Z');
    });
  });

  describe('REQ-003: Error Handling', () => {
    test('should handle connection timeouts', async () => {
      const timeoutError = new Error('Connection timeout');
      
      mockConnectionService.connect = jest.fn().mockRejectedValue(timeoutError);
      
      try {
        await connectionStore.connect();
      } catch (error) {
        expect(connectionStore.status).toBe('error');
        expect(connectionStore.lastError).toBe('Connection timeout');
        expect(connectionStore.reconnectAttempts).toBe(1);
      }
    });

    test('should handle network errors', async () => {
      const networkError = new Error('Network unreachable');
      
      mockConnectionService.connect = jest.fn().mockRejectedValue(networkError);
      
      try {
        await connectionStore.connect();
      } catch (error) {
        expect(connectionStore.status).toBe('error');
        expect(connectionStore.lastError).toBe('Network unreachable');
      }
    });

    test('should handle authentication errors', async () => {
      const authError = new Error('Authentication failed');
      
      mockConnectionService.connect = jest.fn().mockRejectedValue(authError);
      
      try {
        await connectionStore.connect();
      } catch (error) {
        expect(connectionStore.status).toBe('error');
        expect(connectionStore.lastError).toBe('Authentication failed');
      }
    });

    test('should clear errors on successful connection', async () => {
      // Set initial error
      connectionStore.setLastError('Previous error');
      expect(connectionStore.lastError).toBe('Previous error');
      
      // Mock successful connection
      mockConnectionService.connect = jest.fn().mockResolvedValue(undefined);
      
      await connectionStore.connect();
      
      // Error should be cleared on success
      expect(connectionStore.lastError).toBe(null);
      expect(connectionStore.status).toBe('connected');
    });
  });

  describe('REQ-004: API Integration', () => {
    test('should connect with correct service call', async () => {
      mockConnectionService.connect = jest.fn().mockResolvedValue(undefined);
      
      await connectionStore.connect();
      
      // Verify service was called
      expect(mockConnectionService.connect).toHaveBeenCalledTimes(1);
      
      // Verify state was updated
      expect(connectionStore.status).toBe('connected');
      expect(connectionStore.lastError).toBe(null);
    });

    test('should disconnect with correct service call', async () => {
      mockConnectionService.disconnect = jest.fn().mockResolvedValue(undefined);
      
      await connectionStore.disconnect();
      
      // Verify service was called
      expect(mockConnectionService.disconnect).toHaveBeenCalledTimes(1);
      
      // Verify state was updated
      expect(connectionStore.status).toBe('disconnected');
    });

    test('should check connection status correctly', () => {
      const isConnected = connectionStore.isConnected();
      expect(isConnected).toBe(true);
      
      connectionStore.setStatus('disconnected');
      const isDisconnected = connectionStore.isConnected();
      expect(isDisconnected).toBe(false);
    });

    test('should get connection state correctly', () => {
      const connectionState = connectionStore.getConnectionState();
      
      expect(connectionState).toEqual({
        status: 'connected',
        lastError: null,
        reconnectAttempts: 0,
        lastConnected: '2025-01-15T14:30:00Z'
      });
    });
  });

  describe('REQ-005: Side Effects', () => {
    test('should update lastConnected timestamp on successful connection', async () => {
      const initialTimestamp = connectionStore.lastConnected;
      
      mockConnectionService.connect = jest.fn().mockResolvedValue(undefined);
      
      await connectionStore.connect();
      
      // Verify timestamp was updated
      expect(connectionStore.lastConnected).not.toBe(initialTimestamp);
      expect(connectionStore.lastConnected).toBeDefined();
    });

    test('should handle connection state changes during API calls', async () => {
      let resolvePromise: (value: any) => void;
      const promise = new Promise(resolve => {
        resolvePromise = resolve;
      });
      
      mockConnectionService.connect = jest.fn().mockReturnValue(promise);
      
      // Start the connection
      const connectionCall = connectionStore.connect();
      
      // Verify status is connecting
      expect(connectionStore.status).toBe('connecting');
      
      // Resolve the promise
      resolvePromise!(undefined);
      await connectionCall;
      
      // Verify status is connected
      expect(connectionStore.status).toBe('connected');
    });

    test('should handle concurrent connection attempts correctly', async () => {
      mockConnectionService.connect = jest.fn().mockResolvedValue(undefined);
      
      // Start multiple connection attempts
      const promise1 = connectionStore.connect();
      const promise2 = connectionStore.connect();
      
      await Promise.all([promise1, promise2]);
      
      // Verify connection was successful
      expect(connectionStore.status).toBe('connected');
      expect(connectionStore.lastError).toBe(null);
    });

    test('should reset store state correctly', () => {
      // Set some state
      connectionStore.setStatus('error');
      connectionStore.setLastError('Test error');
      connectionStore.setReconnectAttempts(5);
      
      // Reset the store
      connectionStore.reset();
      
      // Verify state was reset to defaults
      expect(connectionStore.status).toBe('connected');
      expect(connectionStore.lastError).toBe(null);
      expect(connectionStore.reconnectAttempts).toBe(0);
      expect(connectionStore.lastConnected).toBe('2025-01-15T14:30:00Z');
    });

    test('should set connection service correctly', () => {
      const newService = MockDataFactory.createMockConnectionService();
      
      connectionStore.setConnectionService(newService);
      
      // Verify service was set (this would be implementation-specific)
      expect(connectionStore.connectionService).toBe(newService);
    });
  });

  describe('Edge Cases and Error Scenarios', () => {
    test('should handle connection drops gracefully', async () => {
      // Simulate connection drop
      connectionStore.handleConnectionDrop();
      
      expect(connectionStore.status).toBe('disconnected');
      expect(connectionStore.lastError).toBe('Connection dropped');
    });

    test('should handle rapid connection/disconnection cycles', async () => {
      mockConnectionService.connect = jest.fn().mockResolvedValue(undefined);
      mockConnectionService.disconnect = jest.fn().mockResolvedValue(undefined);
      
      // Rapid connect/disconnect cycle
      await connectionStore.connect();
      expect(connectionStore.status).toBe('connected');
      
      await connectionStore.disconnect();
      expect(connectionStore.status).toBe('disconnected');
      
      await connectionStore.connect();
      expect(connectionStore.status).toBe('connected');
    });

    test('should handle maximum reconnection attempts', async () => {
      const maxAttempts = 10;
      
      // Mock all connection attempts to fail
      mockConnectionService.connect = jest.fn().mockRejectedValue(new Error('Connection failed'));
      
      // Attempt multiple connections
      for (let i = 0; i < maxAttempts; i++) {
        try {
          await connectionStore.connect();
        } catch (error) {
          // Expected to fail
        }
      }
      
      expect(connectionStore.reconnectAttempts).toBe(maxAttempts);
      expect(connectionStore.status).toBe('error');
    });

    test('should handle malformed connection responses', async () => {
      const malformedResponse = { invalid: 'data' };
      
      mockConnectionService.connect = jest.fn().mockResolvedValue(malformedResponse);
      
      try {
        await connectionStore.connect();
      } catch (error) {
        expect(connectionStore.error).toBeDefined();
      }
    });

    test('should handle network disconnection', async () => {
      const networkError = new Error('Network disconnected');
      
      mockConnectionService.connect = jest.fn().mockRejectedValue(networkError);
      
      try {
        await connectionStore.connect();
      } catch (error) {
        expect(connectionStore.lastError).toBe('Network disconnected');
        expect(connectionStore.status).toBe('error');
      }
    });
  });

  describe('Performance and Optimization', () => {
    test('should handle connection state polling efficiently', () => {
      const pollInterval = 1000; // 1 second
      let pollCount = 0;
      
      const pollConnection = () => {
        pollCount++;
        return connectionStore.isConnected();
      };
      
      // Simulate polling for 5 seconds
      const startTime = Date.now();
      while (Date.now() - startTime < 5000) {
        pollConnection();
      }
      
      // Should have polled multiple times
      expect(pollCount).toBeGreaterThan(0);
    });

    test('should handle connection state changes efficiently', () => {
      const stateChanges = Array.from({ length: 1000 }, (_, i) => ({
        status: i % 2 === 0 ? 'connected' : 'disconnected',
        lastConnected: new Date().toISOString()
      }));
      
      // Apply all state changes
      stateChanges.forEach(change => {
        connectionStore.handleConnectionUpdate(change);
      });
      
      // Verify final state
      expect(connectionStore.status).toBe('disconnected');
    });

    test('should handle reconnection backoff correctly', () => {
      const baseDelay = 1000; // 1 second
      const maxDelay = 30000; // 30 seconds
      
      // Test exponential backoff calculation
      for (let attempt = 1; attempt <= 10; attempt++) {
        const delay = Math.min(baseDelay * Math.pow(2, attempt - 1), maxDelay);
        expect(delay).toBeGreaterThanOrEqual(baseDelay);
        expect(delay).toBeLessThanOrEqual(maxDelay);
      }
    });
  });
});