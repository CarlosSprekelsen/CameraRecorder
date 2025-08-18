/**
 * Connection state management store
 * Sprint 3: Enhanced for real server integration
 * Handles WebSocket connection status and reconnection logic
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import type { ConnectionStatus } from '../types';
import type { WebSocketService } from '../services/websocket';
import { createWebSocketService, defaultWebSocketConfig } from '../services/websocket';

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
  
  // Connection health
  isHealthy: boolean;
  lastHeartbeat: Date | null;
  connectionUptime: number | null;
}

/**
 * Connection store actions interface
 */
interface ConnectionActions {
  // Connection management
  connect: (url?: string) => Promise<void>;
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
  initializeWebSocketService: (url?: string) => void;
  
  // Timestamps
  setLastConnected: (date: Date) => void;
  setLastDisconnected: (date: Date) => void;
  
  // Health monitoring
  setHealthy: (isHealthy: boolean) => void;
  setLastHeartbeat: (date: Date) => void;
  setConnectionUptime: (uptime: number) => void;
  
  // Connection testing
  testConnection: () => Promise<boolean>;
}

/**
 * Connection store type
 */
type ConnectionStore = ConnectionState & ConnectionActions;

/**
 * Create connection store
 * Sprint 3: Enhanced for real server integration
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
      maxReconnectAttempts: defaultWebSocketConfig.maxReconnectAttempts,
      error: null,
      wsService: null,
      isHealthy: false,
      lastHeartbeat: null,
      connectionUptime: null,

      // Connection management
      connect: async (url = defaultWebSocketConfig.url) => {
        try {
          set({ 
            isConnecting: true, 
            error: null, 
            url,
            status: 'connecting' 
          });

          const { wsService } = get();
          if (!wsService) {
            // Initialize WebSocket service if not available
            get().initializeWebSocketService(url);
            const { wsService: newService } = get();
            if (!newService) {
              throw new Error('Failed to initialize WebSocket service');
            }
          }

          const { wsService: service } = get();
          if (!service) {
            throw new Error('WebSocket service not available');
          }

          // Set up event handlers if not already set
          service.onConnect(() => {
            set({ 
              isConnecting: false,
              status: 'connected',
              lastConnected: new Date(),
              reconnectAttempts: 0,
              isHealthy: true,
              error: null
            });
          });

          service.onDisconnect(() => {
            set({ 
              status: 'disconnected',
              isConnecting: false,
              isReconnecting: false,
              lastDisconnected: new Date(),
              isHealthy: false
            });
          });

          service.onError((error) => {
            set({ 
              error: error.message,
              isConnecting: false,
              isConnected: false,
              isHealthy: false,
              status: 'error'
            });
          });

          await service.connect();
          
        } catch (error) {
          set({ 
            isConnecting: false,
            status: 'error',
            error: error instanceof Error ? error.message : 'Failed to connect',
            lastDisconnected: new Date(),
            isHealthy: false
          });
          throw error;
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
          reconnectAttempts: 0,
          isHealthy: false,
          wsService: null
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

      initializeWebSocketService: (url = defaultWebSocketConfig.url) => {
        const wsService = createWebSocketService({
          ...defaultWebSocketConfig,
          url
        });
        set({ wsService, url });
      },

      // Timestamps
      setLastConnected: (date: Date) => {
        set({ lastConnected: date });
      },

      setLastDisconnected: (date: Date) => {
        set({ lastDisconnected: date });
      },

      // Health monitoring
      setHealthy: (isHealthy: boolean) => {
        set({ isHealthy });
      },

      setLastHeartbeat: (date: Date) => {
        set({ lastHeartbeat: date });
      },

      setConnectionUptime: (uptime: number) => {
        set({ connectionUptime: uptime });
      },

      // Connection testing
      testConnection: async (): Promise<boolean> => {
        try {
          const { wsService } = get();
          if (!wsService || !wsService.isConnected) {
            return false;
          }

          // Test with ping
          await wsService.call('ping', {});
          set({ isHealthy: true, lastHeartbeat: new Date() });
          return true;
        } catch (error) {
          set({ isHealthy: false });
          return false;
        }
      },
    }),
    {
      name: 'connection-store',
    }
  )
); 