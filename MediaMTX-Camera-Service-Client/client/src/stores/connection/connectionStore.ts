import { create } from 'zustand';
import { ConnectionState, ConnectionStatus } from '../../types/api';
import { IAPIClient } from '../../services/abstraction/IAPIClient';

interface ConnectionStore extends ConnectionState {
  // Service injection
  setAPIClient: (client: IAPIClient) => void;
  
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
  let apiClient: IAPIClient | null = null;

  return {
    ...initialState,

    // Service injection
    setAPIClient: (client: IAPIClient) => {
      apiClient = client;
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
      if (!apiClient) throw new Error('API client not initialized');
      set({ status: 'connecting', lastError: null });
      try {
        await apiClient.connect();
        set({ status: 'connected', lastConnected: new Date().toISOString() });
      } catch (error) {
        set({ 
          status: 'error', 
          lastError: error instanceof Error ? error.message : 'Connection failed' 
        });
      }
    },

    disconnect: () => {
      if (apiClient) {
        apiClient.disconnect();
      }
      set({ status: 'disconnected' });
    },

    reconnect: async () => {
      if (!apiClient) throw new Error('API client not initialized');
      set((state) => ({ 
        ...state, 
        reconnectAttempts: state.reconnectAttempts + 1,
        status: 'connecting' 
      }));
      try {
        await apiClient.connect();
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
