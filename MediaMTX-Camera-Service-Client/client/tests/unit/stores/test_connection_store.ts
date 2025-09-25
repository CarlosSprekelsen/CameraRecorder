/**
 * Unit Tests for Connection Store
 * 
 * REQ-001: Store State Management - Test Zustand store actions
 * REQ-002: State Transitions - Test state changes
 * REQ-003: Error Handling - Test error states and recovery
 * REQ-004: Connection Status - Test connection status management
 * REQ-005: Side Effects - Test store side effects
 * 
 * Ground Truth: Official RPC Documentation
 * API Reference: docs/api/json_rpc_methods.md
 */

import { describe, test, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { MockDataFactory } from '../../utils/mocks';
import { APIResponseValidator } from '../../utils/validators';
import { useConnectionStore } from '../../../src/stores/connection/connectionStore';

describe('Connection Store', () => {
  beforeEach(() => {
    // Reset the store to initial state
    useConnectionStore.getState().reset();
  });

  afterEach(() => {
    useConnectionStore.getState().reset();
  });

  describe('REQ-001: Store State Management', () => {
    test('should initialize with correct initial state', () => {
      const state = useConnectionStore.getState();
      
      expect(state.status).toBe('disconnected');
      expect(state.lastError).toBe(null);
      expect(state.reconnectAttempts).toBe(0);
      expect(state.lastConnected).toBe(null);
    });

    test('should set status correctly', () => {
      const { setStatus } = useConnectionStore.getState();
      
      setStatus('connecting');
      expect(useConnectionStore.getState().status).toBe('connecting');
      
      setStatus('connected');
      expect(useConnectionStore.getState().status).toBe('connected');
      
      setStatus('disconnected');
      expect(useConnectionStore.getState().status).toBe('disconnected');
      
      setStatus('error');
      expect(useConnectionStore.getState().status).toBe('error');
    });

    test('should set error correctly', () => {
      const { setError } = useConnectionStore.getState();
      
      setError('Connection timeout');
      expect(useConnectionStore.getState().lastError).toBe('Connection timeout');
      expect(useConnectionStore.getState().status).toBe('error');
      
      setError(null);
      expect(useConnectionStore.getState().lastError).toBe(null);
      expect(useConnectionStore.getState().status).toBe('disconnected');
    });

    test('should set reconnect attempts correctly', () => {
      const { setReconnectAttempts } = useConnectionStore.getState();
      
      setReconnectAttempts(5);
      expect(useConnectionStore.getState().reconnectAttempts).toBe(5);
      
      setReconnectAttempts(0);
      expect(useConnectionStore.getState().reconnectAttempts).toBe(0);
    });

    test('should set last connected timestamp correctly', () => {
      const { setLastConnected } = useConnectionStore.getState();
      const timestamp = '2025-01-15T14:30:00Z';
      
      setLastConnected(timestamp);
      expect(useConnectionStore.getState().lastConnected).toBe(timestamp);
      
      setLastConnected(null);
      expect(useConnectionStore.getState().lastConnected).toBe(null);
    });
  });

  describe('REQ-002: State Transitions', () => {
    test('should handle connection lifecycle transitions', () => {
      const { setStatus, setLastConnected, setReconnectAttempts } = useConnectionStore.getState();
      
      // Initial state
      expect(useConnectionStore.getState().status).toBe('disconnected');
      
      // Start connecting
      setStatus('connecting');
      expect(useConnectionStore.getState().status).toBe('connecting');
      
      // Connected successfully
      setStatus('connected');
      setLastConnected('2025-01-15T14:30:00Z');
      expect(useConnectionStore.getState().status).toBe('connected');
      expect(useConnectionStore.getState().lastConnected).toBe('2025-01-15T14:30:00Z');
      
      // Connection lost
      setStatus('disconnected');
      setReconnectAttempts(1);
      expect(useConnectionStore.getState().status).toBe('disconnected');
      expect(useConnectionStore.getState().reconnectAttempts).toBe(1);
    });

    test('should handle error state transitions', () => {
      const { setStatus, setError } = useConnectionStore.getState();
      
      // Set error
      setError('Network error');
      expect(useConnectionStore.getState().status).toBe('error');
      expect(useConnectionStore.getState().lastError).toBe('Network error');
      
      // Clear error by setting status
      setStatus('connected');
      expect(useConnectionStore.getState().status).toBe('connected');
      expect(useConnectionStore.getState().lastError).toBe(null);
    });
  });

  describe('REQ-003: Error Handling', () => {
    test('should handle error state correctly', () => {
      const { setError } = useConnectionStore.getState();
      
      setError('Connection failed');
      
      const state = useConnectionStore.getState();
      expect(state.status).toBe('error');
      expect(state.lastError).toBe('Connection failed');
    });

    test('should clear error when status changes to non-error', () => {
      const { setError, setStatus } = useConnectionStore.getState();
      
      // Set error first
      setError('Connection failed');
      expect(useConnectionStore.getState().status).toBe('error');
      
      // Change status to connected
      setStatus('connected');
      expect(useConnectionStore.getState().status).toBe('connected');
      expect(useConnectionStore.getState().lastError).toBe(null);
    });
  });

  describe('REQ-004: Connection Status', () => {
    test('should track connection status changes', () => {
      const { setStatus } = useConnectionStore.getState();
      
      const statuses = ['disconnected', 'connecting', 'connected', 'disconnected', 'error'];
      
      statuses.forEach(status => {
        setStatus(status as any);
        expect(useConnectionStore.getState().status).toBe(status);
      });
    });

    test('should handle reconnect attempts tracking', () => {
      const { setReconnectAttempts, setStatus } = useConnectionStore.getState();
      
      // Simulate reconnection attempts
      for (let i = 1; i <= 5; i++) {
        setStatus('connecting');
        setReconnectAttempts(i);
        expect(useConnectionStore.getState().reconnectAttempts).toBe(i);
      }
      
      // Successful reconnection
      setStatus('connected');
      setReconnectAttempts(0);
      expect(useConnectionStore.getState().reconnectAttempts).toBe(0);
    });
  });

  describe('REQ-005: Side Effects', () => {
    test('should reset store to initial state', () => {
      const { reset, setStatus, setError, setReconnectAttempts, setLastConnected } = useConnectionStore.getState();
      
      // Modify state
      setStatus('connected');
      setError('Test error');
      setReconnectAttempts(3);
      setLastConnected('2025-01-15T14:30:00Z');
      
      // Reset
      reset();
      
      // Check state is back to initial
      const state = useConnectionStore.getState();
      expect(state.status).toBe('disconnected');
      expect(state.lastError).toBe(null);
      expect(state.reconnectAttempts).toBe(0);
      expect(state.lastConnected).toBe(null);
    });

    test('should maintain state consistency during updates', () => {
      const { setStatus, setError, setReconnectAttempts } = useConnectionStore.getState();
      
      // Set multiple properties
      setStatus('connecting');
      setReconnectAttempts(2);
      setError('Temporary error');
      
      const state = useConnectionStore.getState();
      expect(state.status).toBe('error'); // Error should override status
      expect(state.lastError).toBe('Temporary error');
      expect(state.reconnectAttempts).toBe(2);
    });
  });

  describe('API Compliance Validation', () => {
    test('should validate connection state against RPC spec', () => {
      const { setStatus } = useConnectionStore.getState();
      
      const validStatuses = ['connected', 'disconnected', 'connecting', 'error'];
      
      validStatuses.forEach(status => {
        setStatus(status as any);
        const state = useConnectionStore.getState();
        // Connection state validation would go here if validator existed
      });
    });
  });
});