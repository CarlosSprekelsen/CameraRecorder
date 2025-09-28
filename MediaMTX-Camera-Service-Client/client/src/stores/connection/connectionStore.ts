import { create } from 'zustand';
import { ConnectionState, ConnectionStatus } from '../../types/api';
import { WebSocketService } from '../../services/websocket/WebSocketService';

interface ConnectionStore extends ConnectionState {
  // Service injection
  setWebSocketService: (service: WebSocketService) => void;
  
  // State setters
  setStatus: (status: ConnectionStatus) => void;
  setError: (error: string | null) => void;
  setReconnectAttempts: (attempts: number) => void;
  setLastConnected: (timestamp: string | null) => void;
  
  // Actions that call services
  connect: () => Promise<void>;
  disconnect: () => void;
  reconnect: () => Promise<void>;
  
  // Reset
  reset: () => void;
}

const initialState: ConnectionState = {
  status: 'disconnected',
  lastError: null,
  reconnectAttempts: 0,
  lastConnected: null,
};

export const useConnectionStore = create<ConnectionStore>((set) => {
  let wsService: WebSocketService | null = null;

  return {
    ...initialState,

    // Service injection
    setWebSocketService: (service: WebSocketService) => {
      wsService = service;
    },

    // State setters
    setStatus: (status: ConnectionStatus) =>
      set((state) => ({
        ...state,
        status,
        lastError: status === 'error' ? state.lastError : null,
      })),

    setError: (error: string | null) =>
      set((state) => ({
        ...state,
        lastError: error,
        status: error ? 'error' : 'disconnected',
      })),

    setReconnectAttempts: (attempts: number) =>
      set((state) => ({ ...state, reconnectAttempts: attempts })),

    setLastConnected: (timestamp: string | null) =>
      set((state) => ({ ...state, lastConnected: timestamp })),

    // Actions that call services
    connect: async () => {
      if (!wsService) throw new Error('WebSocket service not initialized');
      set({ status: 'connecting', lastError: null });
      try {
        await wsService.connect();
        set({ status: 'connected', lastConnected: new Date().toISOString() });
      } catch (error) {
        set({ 
          status: 'error', 
          lastError: error instanceof Error ? error.message : 'Connection failed' 
        });
      }
    },

    disconnect: () => {
      if (wsService) {
        wsService.disconnect();
      }
      set({ status: 'disconnected' });
    },

    reconnect: async () => {
      if (!wsService) throw new Error('WebSocket service not initialized');
      set((state) => ({ 
        ...state, 
        reconnectAttempts: state.reconnectAttempts + 1,
        status: 'connecting' 
      }));
      try {
        await wsService.connect();
        set({ status: 'connected', lastConnected: new Date().toISOString() });
      } catch (error) {
        set({ 
          status: 'error', 
          lastError: error instanceof Error ? error.message : 'Reconnection failed' 
        });
      }
    },

    reset: () => set(initialState),
  };
});
