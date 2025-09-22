/**
 * REQ-NET01-001: Connection state management must be reliable and predictable
 * REQ-NET01-002: Connection recovery mechanisms must handle failures gracefully
 * Coverage: UNIT
 * Quality: HIGH
 */
/**
 * Unit tests for connection store
 * 
 * Design Principles:
 * - Pure unit testing with complete isolation
 * - Direct store testing without React context dependency
 * - Focus on connection state management logic
 * - Test connection lifecycle and error handling
 * - Validate reconnection mechanisms
 */

import { useConnectionStore } from '../../../src/stores/connection/connectionStore';
import type { ConnectionStatus } from '../../../src/types';

describe('Connection Store', () => {
  let store: ReturnType<typeof useConnectionStore.getState>;

  beforeEach(() => {
    // Reset store state completely
    const currentStore = useConnectionStore.getState();
    currentStore.reset();
    
    // Get fresh store instance after reset
    store = useConnectionStore.getState();
  });

  describe('Initialization', () => {
    it('should start with correct default state', () => {
      const state = useConnectionStore.getState();
      expect(state.status).toBe('disconnected');
      expect(state.isConnecting).toBe(false);
      expect(state.isReconnecting).toBe(false);
      expect(state.isConnected).toBe(false);
      expect(state.url).toBeNull();
      expect(state.lastConnected).toBeNull();
      expect(state.lastDisconnected).toBeNull();
      expect(state.reconnectAttempts).toBe(0);
      expect(state.maxReconnectAttempts).toBe(5);
      expect(state.nextReconnectTime).toBeNull();
      expect(state.autoReconnect).toBe(true);
      expect(state.error).toBeNull();
      expect(state.errorCode).toBeNull();
      expect(state.errorTimestamp).toBeNull();
    });
  });

  describe('Connection Status Management', () => {
    it('should set connection status correctly', () => {
      store.setStatus('connecting');
      let state = useConnectionStore.getState();
      expect(state.status).toBe('connecting');

      store.setStatus('connected');
      state = useConnectionStore.getState();
      expect(state.status).toBe('connected');
      expect(state.isConnected).toBe(true);

      store.setStatus('disconnected');
      state = useConnectionStore.getState();
      expect(state.status).toBe('disconnected');
      expect(state.isConnected).toBe(false);
    });

    it('should set connecting state', () => {
      store.setConnecting(true);
      let state = useConnectionStore.getState();
      expect(state.isConnecting).toBe(true);

      store.setConnecting(false);
      state = useConnectionStore.getState();
      expect(state.isConnecting).toBe(false);
    });

    it('should set reconnecting state', () => {
      store.setReconnecting(true);
      let state = useConnectionStore.getState();
      expect(state.isReconnecting).toBe(true);

      store.setReconnecting(false);
      state = useConnectionStore.getState();
      expect(state.isReconnecting).toBe(false);
    });

    it('should set connected state', () => {
      store.setConnected(true);
      let state = useConnectionStore.getState();
      expect(state.isConnected).toBe(true);

      store.setConnected(false);
      state = useConnectionStore.getState();
      expect(state.isConnected).toBe(false);
    });
  });

  describe('Connection Information Management', () => {
    it('should set connection URL', () => {
      const testUrl = 'ws://localhost:8002/ws';
      store.setUrl(testUrl);
      
      const state = useConnectionStore.getState();
      expect(state.url).toBe(testUrl);
    });

    it('should update connection timestamps', () => {
      const now = new Date();
      
      store.setLastConnected(now);
      let state = useConnectionStore.getState();
      expect(state.lastConnected).toEqual(now);

      store.setLastDisconnected(now);
      state = useConnectionStore.getState();
      expect(state.lastDisconnected).toEqual(now);
    });
  });

  describe('Reconnection Management', () => {
    it('should increment reconnect attempts', () => {
      store.incrementReconnectAttempts();
      let state = useConnectionStore.getState();
      expect(state.reconnectAttempts).toBe(1);

      store.incrementReconnectAttempts();
      state = useConnectionStore.getState();
      expect(state.reconnectAttempts).toBe(2);
    });

    it('should reset reconnect attempts', () => {
      store.incrementReconnectAttempts();
      store.incrementReconnectAttempts();
      store.resetReconnectAttempts();
      
      const state = useConnectionStore.getState();
      expect(state.reconnectAttempts).toBe(0);
    });

    it('should set max reconnect attempts', () => {
      store.setMaxReconnectAttempts(10);
      
      const state = useConnectionStore.getState();
      expect(state.maxReconnectAttempts).toBe(10);
    });

    it('should set next reconnect time', () => {
      const futureTime = new Date(Date.now() + 5000);
      store.setNextReconnectTime(futureTime);
      
      const state = useConnectionStore.getState();
      expect(state.nextReconnectTime).toEqual(futureTime);
    });

    it('should set auto reconnect', () => {
      store.setAutoReconnect(false);
      let state = useConnectionStore.getState();
      expect(state.autoReconnect).toBe(false);

      store.setAutoReconnect(true);
      state = useConnectionStore.getState();
      expect(state.autoReconnect).toBe(true);
    });

    it('should check if should reconnect', () => {
      // Test with auto reconnect enabled and attempts under limit
      store.setAutoReconnect(true);
      store.setReconnectAttempts(2);
      store.setMaxReconnectAttempts(5);
      
      let state = useConnectionStore.getState();
      expect(state.shouldReconnect()).toBe(true);

      // Test with auto reconnect disabled
      store.setAutoReconnect(false);
      state = useConnectionStore.getState();
      expect(state.shouldReconnect()).toBe(false);

      // Test with max attempts reached
      store.setAutoReconnect(true);
      store.setReconnectAttempts(5);
      state = useConnectionStore.getState();
      expect(state.shouldReconnect()).toBe(false);
    });
  });

  describe('Error Management', () => {
    it('should set error with message and code', () => {
      const errorMessage = 'Connection failed';
      const errorCode = 1006;
      
      store.setError(errorMessage, errorCode);
      
      const state = useConnectionStore.getState();
      expect(state.error).toBe(errorMessage);
      expect(state.errorCode).toBe(errorCode);
      expect(state.errorTimestamp).toBeInstanceOf(Date);
    });

    it('should clear error', () => {
      store.setError('Test error', 1000);
      store.clearError();
      
      const state = useConnectionStore.getState();
      expect(state.error).toBeNull();
      expect(state.errorCode).toBeNull();
      expect(state.errorTimestamp).toBeNull();
    });

    it('should increment error count', () => {
      store.incrementErrorCount();
      let state = useConnectionStore.getState();
      expect(state.errorCount).toBe(1);

      store.incrementErrorCount();
      state = useConnectionStore.getState();
      expect(state.errorCount).toBe(2);
    });
  });

  describe('Connection Lifecycle', () => {
    it('should handle connection start', () => {
      const url = 'ws://localhost:8002/ws';
      store.startConnection(url);
      
      const state = useConnectionStore.getState();
      expect(state.status).toBe('connecting');
      expect(state.isConnecting).toBe(true);
      expect(state.url).toBe(url);
      expect(state.error).toBeNull();
    });

    it('should handle successful connection', () => {
      const url = 'ws://localhost:8002/ws';
      store.startConnection(url);
      store.connectionSuccess();
      
      const state = useConnectionStore.getState();
      expect(state.status).toBe('connected');
      expect(state.isConnecting).toBe(false);
      expect(state.isConnected).toBe(true);
      expect(state.lastConnected).toBeInstanceOf(Date);
      expect(state.reconnectAttempts).toBe(0);
      expect(state.error).toBeNull();
    });

    it('should handle connection failure', () => {
      const url = 'ws://localhost:8002/ws';
      const errorMessage = 'Connection refused';
      const errorCode = 1006;
      
      store.startConnection(url);
      store.connectionFailed(errorMessage, errorCode);
      
      const state = useConnectionStore.getState();
      expect(state.status).toBe('disconnected');
      expect(state.isConnecting).toBe(false);
      expect(state.isConnected).toBe(false);
      expect(state.error).toBe(errorMessage);
      expect(state.errorCode).toBe(errorCode);
      expect(state.lastDisconnected).toBeInstanceOf(Date);
    });

    it('should handle disconnection', () => {
      // First establish a connection
      store.startConnection('ws://localhost:8002/ws');
      store.connectionSuccess();
      
      // Then disconnect
      store.disconnect();
      
      const state = useConnectionStore.getState();
      expect(state.status).toBe('disconnected');
      expect(state.isConnected).toBe(false);
      expect(state.lastDisconnected).toBeInstanceOf(Date);
    });

    it('should handle reconnection start', () => {
      store.startReconnection();
      
      const state = useConnectionStore.getState();
      expect(state.status).toBe('reconnecting');
      expect(state.isReconnecting).toBe(true);
      expect(state.reconnectAttempts).toBe(1);
    });

    it('should handle reconnection success', () => {
      store.startReconnection();
      store.reconnectionSuccess();
      
      const state = useConnectionStore.getState();
      expect(state.status).toBe('connected');
      expect(state.isReconnecting).toBe(false);
      expect(state.isConnected).toBe(true);
      expect(state.reconnectAttempts).toBe(0);
    });

    it('should handle reconnection failure', () => {
      store.startReconnection();
      store.reconnectionFailed('Reconnection failed');
      
      const state = useConnectionStore.getState();
      expect(state.status).toBe('disconnected');
      expect(state.isReconnecting).toBe(false);
      expect(state.isConnected).toBe(false);
      expect(state.error).toBe('Reconnection failed');
    });
  });

  describe('State Reset', () => {
    it('should reset all state to initial values', () => {
      // Set some state
      store.setStatus('connected');
      store.setConnecting(true);
      store.setUrl('ws://test');
      store.setError('Test error', 1000);
      store.incrementReconnectAttempts();
      
      // Reset
      store.reset();
      
      const state = useConnectionStore.getState();
      expect(state.status).toBe('disconnected');
      expect(state.isConnecting).toBe(false);
      expect(state.isReconnecting).toBe(false);
      expect(state.isConnected).toBe(false);
      expect(state.url).toBeNull();
      expect(state.error).toBeNull();
      expect(state.errorCode).toBeNull();
      expect(state.reconnectAttempts).toBe(0);
    });
  });

  describe('Connection State Queries', () => {
    it('should check if connection is active', () => {
      store.setStatus('connected');
      expect(store.isConnectionActive()).toBe(true);

      store.setStatus('connecting');
      expect(store.isConnectionActive()).toBe(false);

      store.setStatus('disconnected');
      expect(store.isConnectionActive()).toBe(false);
    });

    it('should check if connection is stable', () => {
      store.setStatus('connected');
      store.setError(null);
      expect(store.isConnectionStable()).toBe(true);

      store.setError('Some error');
      expect(store.isConnectionStable()).toBe(false);
    });

    it('should get connection duration', () => {
      const startTime = new Date(Date.now() - 5000); // 5 seconds ago
      store.setLastConnected(startTime);
      
      const duration = store.getConnectionDuration();
      expect(duration).toBeGreaterThanOrEqual(5000);
      expect(duration).toBeLessThan(6000); // Allow some tolerance
    });

    it('should get time until next reconnect', () => {
      const futureTime = new Date(Date.now() + 5000);
      store.setNextReconnectTime(futureTime);
      
      const timeUntilReconnect = store.getTimeUntilReconnect();
      expect(timeUntilReconnect).toBeGreaterThan(4000);
      expect(timeUntilReconnect).toBeLessThanOrEqual(5000);
    });
  });
});
