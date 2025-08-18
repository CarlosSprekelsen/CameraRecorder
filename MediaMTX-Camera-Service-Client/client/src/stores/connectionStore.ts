/**
 * Connection state management store
 * Handles WebSocket connection status and reconnection logic
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import type { ConnectionStatus } from '../types';
import type { WebSocketService } from '../services/websocket';

/**
 * Connection store state interface
 */
interface ConnectionState {
  // Connection status
  status: ConnectionStatus;
  isConnecting: boolean;
  isReconnecting: boolean;
  
  // Connection info
  url: string | null;
  lastConnected: Date | null;
  lastDisconnected: Date | null;
  
  // Reconnection info
  reconnectAttempts: number;
  maxReconnectAttempts: number;
  
  // Error state
  error: string | null;
  
  // WebSocket service reference
  wsService: WebSocketService | null;
}

/**
 * Connection store actions interface
 */
interface ConnectionActions {
  // Connection management
  connect: (url: string) => Promise<void>;
  disconnect: () => void;
  reconnect: () => Promise<void>;
  
  // Status updates
  setStatus: (status: ConnectionStatus) => void;
  setConnecting: (isConnecting: boolean) => void;
  setReconnecting: (isReconnecting: boolean) => void;
  
  // Error handling
  setError: (error: string | null) => void;
  clearError: () => void;
  
  // Reconnection management
  setReconnectAttempts: (attempts: number) => void;
  resetReconnectAttempts: () => void;
  
  // Service management
  setWebSocketService: (service: WebSocketService | null) => void;
  
  // Timestamps
  setLastConnected: (date: Date) => void;
  setLastDisconnected: (date: Date) => void;
}

/**
 * Connection store type
 */
type ConnectionStore = ConnectionState & ConnectionActions;

/**
 * Create connection store
 */
export const useConnectionStore = create<ConnectionStore>()(
  devtools(
    (set, get) => ({
      // Initial state
      status: 'disconnected',
      isConnecting: false,
      isReconnecting: false,
      url: null,
      lastConnected: null,
      lastDisconnected: null,
      reconnectAttempts: 0,
      maxReconnectAttempts: 5,
      error: null,
      wsService: null,

      // Connection management
      connect: async (url: string) => {
        try {
          set({ 
            isConnecting: true, 
            error: null, 
            url,
            status: 'connecting' 
          });

          const { wsService } = get();
          if (!wsService) {
            throw new Error('WebSocket service not available');
          }

          await wsService.connect();
          
          set({ 
            isConnecting: false,
            status: 'connected',
            lastConnected: new Date(),
            reconnectAttempts: 0
          });
          
        } catch (error) {
          set({ 
            isConnecting: false,
            status: 'error',
            error: error instanceof Error ? error.message : 'Failed to connect',
            lastDisconnected: new Date()
          });
        }
      },

      disconnect: () => {
        const { wsService } = get();
        if (wsService) {
          wsService.disconnect();
        }
        
        set({ 
          status: 'disconnected',
          isConnecting: false,
          isReconnecting: false,
          lastDisconnected: new Date(),
          reconnectAttempts: 0
        });
      },

      reconnect: async () => {
        const { url, reconnectAttempts, maxReconnectAttempts } = get();
        
        if (!url) {
          set({ error: 'No connection URL available for reconnection' });
          return;
        }

        if (reconnectAttempts >= maxReconnectAttempts) {
          set({ 
            status: 'error',
            isReconnecting: false,
            error: 'Max reconnection attempts reached'
          });
          return;
        }

        try {
          set({ 
            isReconnecting: true,
            status: 'connecting',
            reconnectAttempts: reconnectAttempts + 1
          });

          await get().connect(url);
          
          set({ isReconnecting: false });
          
        } catch (error) {
          set({ 
            isReconnecting: false,
            status: 'error',
            error: error instanceof Error ? error.message : 'Reconnection failed'
          });
        }
      },

      // Status updates
      setStatus: (status: ConnectionStatus) => {
        set({ status });
      },

      setConnecting: (isConnecting: boolean) => {
        set({ isConnecting });
      },

      setReconnecting: (isReconnecting: boolean) => {
        set({ isReconnecting });
      },

      // Error handling
      setError: (error: string | null) => {
        set({ error });
      },

      clearError: () => {
        set({ error: null });
      },

      // Reconnection management
      setReconnectAttempts: (attempts: number) => {
        set({ reconnectAttempts: attempts });
      },

      resetReconnectAttempts: () => {
        set({ reconnectAttempts: 0 });
      },

      // Service management
      setWebSocketService: (service: WebSocketService | null) => {
        set({ wsService: service });
      },

      // Timestamps
      setLastConnected: (date: Date) => {
        set({ lastConnected: date });
      },

      setLastDisconnected: (date: Date) => {
        set({ lastDisconnected: date });
      },
    }),
    {
      name: 'connection-store',
    }
  )
); 