/**
 * ConnectionStore unit tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * 
 * Requirements Coverage:
 * - REQ-CS-001: Connection state management
 * - REQ-CS-002: Connection status tracking
 * - REQ-CS-003: Error handling and recovery
 * - REQ-CS-004: Reconnection attempts tracking
 * - REQ-CS-005: Last connected timestamp management
 * 
 * Test Categories: Unit
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */

import { useConnectionStore } from '../../../src/stores/connection/connectionStore';
import { ConnectionStatus } from '../../../src/types/api';

describe('ConnectionStore Unit Tests', () => {
  let store: ReturnType<typeof useConnectionStore>;

  beforeEach(() => {
    // Reset store state before each test
    store = useConnectionStore.getState();
    store.reset();
  });

  afterEach(() => {
    // Reset store after each test
    store.reset();
  });

  describe('REQ-CS-001: Connection state management', () => {
    test('should initialize with correct initial state', () => {
      const state = useConnectionStore.getState();
      
      expect(state.status).toBe('disconnected');
      expect(state.lastError).toBeNull();
      expect(state.reconnectAttempts).toBe(0);
      expect(state.lastConnected).toBeNull();
    });

    test('should set status correctly', () => {
      const statuses: ConnectionStatus[] = ['disconnected', 'connecting', 'connected', 'error'];
      
      statuses.forEach(status => {
        store.setStatus(status);
        expect(store.status).toBe(status);
      });
    });

    test('should clear error when setting non-error status', () => {
      // Set an error first
      store.setError('Test error');
      expect(store.lastError).toBe('Test error');
      expect(store.status).toBe('error');
      
      // Set a non-error status
      store.setStatus('connected');
      expect(store.lastError).toBeNull();
      expect(store.status).toBe('connected');
    });

    test('should preserve error when setting error status', () => {
      const errorMessage = 'Test error';
      store.setError(errorMessage);
      
      // Set error status again
      store.setStatus('error');
      expect(store.lastError).toBe(errorMessage);
      expect(store.status).toBe('error');
    });

    test('should reset to initial state', () => {
      // Set some state
      store.setStatus('connected');
      store.setError('Test error');
      store.setReconnectAttempts(5);
      store.setLastConnected(new Date().toISOString());
      
      // Reset
      store.reset();
      
      expect(store.status).toBe('disconnected');
      expect(store.lastError).toBeNull();
      expect(store.reconnectAttempts).toBe(0);
      expect(store.lastConnected).toBeNull();
    });
  });

  describe('REQ-CS-002: Connection status tracking', () => {
    test('should track connection flow correctly', () => {
      // Initial state
      expect(store.status).toBe('disconnected');
      
      // Start connecting
      store.setStatus('connecting');
      expect(store.status).toBe('connecting');
      
      // Connected successfully
      store.setStatus('connected');
      expect(store.status).toBe('connected');
      expect(store.lastError).toBeNull();
    });

    test('should handle connection error', () => {
      // Start connecting
      store.setStatus('connecting');
      
      // Connection fails
      store.setError('Connection failed');
      expect(store.status).toBe('error');
      expect(store.lastError).toBe('Connection failed');
    });

    test('should handle disconnection', () => {
      // Connect first
      store.setStatus('connected');
      store.setLastConnected(new Date().toISOString());
      
      // Disconnect
      store.setStatus('disconnected');
      expect(store.status).toBe('disconnected');
      // lastConnected should remain set
      expect(store.lastConnected).toBeTruthy();
    });

    test('should handle multiple status changes', () => {
      const statusSequence: ConnectionStatus[] = [
        'disconnected',
        'connecting',
        'connected',
        'disconnected',
        'connecting',
        'error'
      ];
      
      statusSequence.forEach((status, index) => {
        store.setStatus(status);
        expect(store.status).toBe(status);
      });
    });
  });

  describe('REQ-CS-003: Error handling and recovery', () => {
    test('should set error correctly', () => {
      const errorMessage = 'WebSocket connection failed';
      store.setError(errorMessage);
      
      expect(store.lastError).toBe(errorMessage);
      expect(store.status).toBe('error');
    });

    test('should clear error when setting null', () => {
      // Set an error first
      store.setError('Test error');
      expect(store.lastError).toBe('Test error');
      expect(store.status).toBe('error');
      
      // Clear error
      store.setError(null);
      expect(store.lastError).toBeNull();
      expect(store.status).toBe('error'); // Status remains error until explicitly changed
    });

    test('should handle different error types', () => {
      const errors = [
        'Connection timeout',
        'Network unreachable',
        'Authentication failed',
        'Server unavailable'
      ];
      
      errors.forEach(error => {
        store.setError(error);
        expect(store.lastError).toBe(error);
        expect(store.status).toBe('error');
      });
    });

    test('should handle error recovery flow', () => {
      // Initial error
      store.setError('Connection failed');
      expect(store.status).toBe('error');
      
      // Clear error and try reconnecting
      store.setError(null);
      store.setStatus('connecting');
      expect(store.status).toBe('connecting');
      expect(store.lastError).toBeNull();
      
      // Successful reconnection
      store.setStatus('connected');
      expect(store.status).toBe('connected');
      expect(store.lastError).toBeNull();
    });
  });

  describe('REQ-CS-004: Reconnection attempts tracking', () => {
    test('should set reconnect attempts correctly', () => {
      store.setReconnectAttempts(0);
      expect(store.reconnectAttempts).toBe(0);
      
      store.setReconnectAttempts(5);
      expect(store.reconnectAttempts).toBe(5);
      
      store.setReconnectAttempts(10);
      expect(store.reconnectAttempts).toBe(10);
    });

    test('should handle reconnection attempt sequence', () => {
      // Initial connection fails
      store.setError('Connection failed');
      store.setReconnectAttempts(1);
      
      expect(store.reconnectAttempts).toBe(1);
      expect(store.status).toBe('error');
      
      // Try reconnecting
      store.setStatus('connecting');
      store.setReconnectAttempts(2);
      
      expect(store.reconnectAttempts).toBe(2);
      expect(store.status).toBe('connecting');
      
      // Reconnection fails again
      store.setError('Reconnection failed');
      store.setReconnectAttempts(3);
      
      expect(store.reconnectAttempts).toBe(3);
      expect(store.status).toBe('error');
    });

    test('should reset reconnect attempts on successful connection', () => {
      // Failed connection with attempts
      store.setReconnectAttempts(5);
      store.setError('Connection failed');
      
      // Successful reconnection
      store.setStatus('connected');
      store.setReconnectAttempts(0); // Reset attempts
      
      expect(store.reconnectAttempts).toBe(0);
      expect(store.status).toBe('connected');
    });

    test('should handle maximum reconnect attempts', () => {
      const maxAttempts = 10;
      
      for (let i = 1; i <= maxAttempts; i++) {
        store.setReconnectAttempts(i);
        expect(store.reconnectAttempts).toBe(i);
      }
      
      // Should handle even higher numbers
      store.setReconnectAttempts(100);
      expect(store.reconnectAttempts).toBe(100);
    });
  });

  describe('REQ-CS-005: Last connected timestamp management', () => {
    test('should set last connected timestamp correctly', () => {
      const timestamp = new Date().toISOString();
      store.setLastConnected(timestamp);
      
      expect(store.lastConnected).toBe(timestamp);
    });

    test('should clear last connected timestamp', () => {
      // Set timestamp first
      const timestamp = new Date().toISOString();
      store.setLastConnected(timestamp);
      expect(store.lastConnected).toBe(timestamp);
      
      // Clear timestamp
      store.setLastConnected(null);
      expect(store.lastConnected).toBeNull();
    });

    test('should handle connection flow with timestamps', () => {
      // Initial state
      expect(store.lastConnected).toBeNull();
      
      // First connection
      const firstConnectionTime = new Date().toISOString();
      store.setStatus('connected');
      store.setLastConnected(firstConnectionTime);
      expect(store.lastConnected).toBe(firstConnectionTime);
      
      // Disconnection (timestamp should remain)
      store.setStatus('disconnected');
      expect(store.lastConnected).toBe(firstConnectionTime);
      
      // Reconnection
      const secondConnectionTime = new Date().toISOString();
      store.setStatus('connected');
      store.setLastConnected(secondConnectionTime);
      expect(store.lastConnected).toBe(secondConnectionTime);
    });

    test('should handle invalid timestamp formats', () => {
      const invalidTimestamps = [
        'invalid-date',
        '',
        '2023-13-45T25:70:90Z',
        null,
        undefined
      ];
      
      invalidTimestamps.forEach(timestamp => {
        store.setLastConnected(timestamp as any);
        expect(store.lastConnected).toBe(timestamp);
      });
    });
  });

  describe('Integration Tests', () => {
    test('should handle complete connection lifecycle', () => {
      // Initial state
      expect(store.status).toBe('disconnected');
      expect(store.lastError).toBeNull();
      expect(store.reconnectAttempts).toBe(0);
      expect(store.lastConnected).toBeNull();
      
      // Start connecting
      store.setStatus('connecting');
      expect(store.status).toBe('connecting');
      
      // Connection succeeds
      const connectionTime = new Date().toISOString();
      store.setStatus('connected');
      store.setLastConnected(connectionTime);
      expect(store.status).toBe('connected');
      expect(store.lastConnected).toBe(connectionTime);
      expect(store.lastError).toBeNull();
      
      // Connection drops
      store.setStatus('disconnected');
      expect(store.status).toBe('disconnected');
      expect(store.lastConnected).toBe(connectionTime); // Should remain
      
      // Try reconnecting
      store.setStatus('connecting');
      store.setReconnectAttempts(1);
      expect(store.status).toBe('connecting');
      expect(store.reconnectAttempts).toBe(1);
      
      // Reconnection fails
      store.setError('Network timeout');
      expect(store.status).toBe('error');
      expect(store.lastError).toBe('Network timeout');
      
      // Try again
      store.setReconnectAttempts(2);
      store.setStatus('connecting');
      store.setError(null);
      expect(store.reconnectAttempts).toBe(2);
      expect(store.status).toBe('connecting');
      expect(store.lastError).toBeNull();
      
      // Success
      const reconnectionTime = new Date().toISOString();
      store.setStatus('connected');
      store.setLastConnected(reconnectionTime);
      store.setReconnectAttempts(0);
      expect(store.status).toBe('connected');
      expect(store.lastConnected).toBe(reconnectionTime);
      expect(store.reconnectAttempts).toBe(0);
      expect(store.lastError).toBeNull();
    });

    test('should handle rapid status changes', () => {
      const statuses: ConnectionStatus[] = [
        'disconnected',
        'connecting',
        'connected',
        'disconnected',
        'connecting',
        'error',
        'connecting',
        'connected'
      ];
      
      statuses.forEach((status, index) => {
        store.setStatus(status);
        expect(store.status).toBe(status);
      });
      
      expect(store.status).toBe('connected');
    });

    test('should maintain state consistency during operations', () => {
      // Set up a connected state
      store.setStatus('connected');
      store.setLastConnected(new Date().toISOString());
      store.setReconnectAttempts(0);
      
      // Verify state is consistent
      expect(store.status).toBe('connected');
      expect(store.lastConnected).toBeTruthy();
      expect(store.reconnectAttempts).toBe(0);
      expect(store.lastError).toBeNull();
      
      // Simulate network issues
      store.setError('Network instability');
      expect(store.status).toBe('error');
      expect(store.lastError).toBe('Network instability');
      expect(store.lastConnected).toBeTruthy(); // Should remain
      
      // Recover
      store.setError(null);
      store.setStatus('connected');
      expect(store.status).toBe('connected');
      expect(store.lastError).toBeNull();
      expect(store.lastConnected).toBeTruthy(); // Should remain
    });
  });

  describe('Edge Cases', () => {
    test('should handle empty string errors', () => {
      store.setError('');
      expect(store.lastError).toBe('');
      expect(store.status).toBe('error');
    });

    test('should handle very large reconnect attempt numbers', () => {
      const largeNumber = 999999;
      store.setReconnectAttempts(largeNumber);
      expect(store.reconnectAttempts).toBe(largeNumber);
    });

    test('should handle negative reconnect attempts', () => {
      store.setReconnectAttempts(-1);
      expect(store.reconnectAttempts).toBe(-1);
    });

    test('should handle multiple rapid error changes', () => {
      const errors = ['Error 1', 'Error 2', 'Error 3', null, 'Error 4'];
      
      errors.forEach(error => {
        store.setError(error);
        expect(store.lastError).toBe(error);
      });
      
      expect(store.lastError).toBe('Error 4');
    });

    test('should handle timestamp precision', () => {
      const preciseTimestamp = '2023-12-01T12:34:56.789Z';
      store.setLastConnected(preciseTimestamp);
      expect(store.lastConnected).toBe(preciseTimestamp);
    });
  });
});
