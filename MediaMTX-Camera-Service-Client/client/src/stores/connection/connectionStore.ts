import { create } from 'zustand';
import { ConnectionState, ConnectionStatus } from '../../types/api';

interface ConnectionStore extends ConnectionState {
  setStatus: (status: ConnectionStatus) => void;
  setError: (error: string | null) => void;
  setReconnectAttempts: (attempts: number) => void;
  setLastConnected: (timestamp: string | null) => void;
  reset: () => void;
}

const initialState: ConnectionState = {
  status: 'disconnected',
  lastError: null,
  reconnectAttempts: 0,
  lastConnected: null,
};

export const useConnectionStore = create<ConnectionStore>((set) => ({
  ...initialState,
  
  setStatus: (status: ConnectionStatus) => 
    set((state) => ({ 
      ...state, 
      status,
      lastError: status === 'error' ? state.lastError : null 
    })),
  
  setError: (error: string | null) => 
    set((state) => ({ 
      ...state, 
      lastError: error,
      status: error ? 'error' : state.status 
    })),
  
  setReconnectAttempts: (attempts: number) => 
    set((state) => ({ ...state, reconnectAttempts: attempts })),
  
  setLastConnected: (timestamp: string | null) => 
    set((state) => ({ ...state, lastConnected: timestamp })),
  
  reset: () => set(initialState),
}));
