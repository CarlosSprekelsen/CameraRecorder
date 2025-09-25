/**
 * Connection Store - Core Connection State
 * 
 * Architecture: Single Responsibility Principle
 * - Handles only WebSocket connection state
 * - Separated from health, metrics, and UI concerns
 * - Provides clean interface for connection operations
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import type { ConnectionStatus } from '../../types';

/**
 * Core connection state interface
 */
export interface ConnectionStoreState {
  // Connection status
  status: ConnectionStatus;
  isConnecting: boolean;
  isReconnecting: boolean;
  isConnected: boolean;
  
  // Connection info
  url: string | null;
  lastConnected: Date | null;
  lastDisconnected: Date | null;
  
  // Reconnection info
  reconnectAttempts: number;
  maxReconnectAttempts: number;
  nextReconnectTime: Date | null;
  autoReconnect: boolean;
  
  // Error state
  error: string | null;
  errorCode: number | null;
  errorTimestamp: Date | null;
}

/**
 * Connection store actions interface
 */
export interface ConnectionStoreActions {
  // Connection management
  setStatus: (status: ConnectionStatus) => void;
  setConnecting: (connecting: boolean) => void;
  setReconnecting: (reconnecting: boolean) => void;
  setConnected: (connected: boolean) => void;
  
  // Connection info
  setUrl: (url: string | null) => void;
  setLastConnected: (date: Date | null) => void;
  setLastDisconnected: (date: Date | null) => void;
  
  // Reconnection management
  setReconnectAttempts: (attempts: number) => void;
  setMaxReconnectAttempts: (max: number) => void;
  setNextReconnectTime: (date: Date | null) => void;
  setAutoReconnect: (enabled: boolean) => void;
  
  // Error management
  setError: (error: string | null, code?: number) => void;
  clearError: () => void;
  
  // Utility actions
  reset: () => void;
}

/**
 * Connection store type
 */
type ConnectionStore = ConnectionStoreState & ConnectionStoreActions;

/**
 * Connection store implementation
 */
export const useConnectionStore = create<ConnectionStore>()(
  devtools(
    (set, get) => ({
      // Initial state
      status: 'disconnected',
      isConnecting: false,
      isReconnecting: false,
      isConnected: false,
      
      url: null,
      lastConnected: null,
      lastDisconnected: null,
      
      reconnectAttempts: 0,
      maxReconnectAttempts: 10,
      nextReconnectTime: null,
      autoReconnect: true,
      
      error: null,
      errorCode: null,
      errorTimestamp: null,

      // Connection management actions
      setStatus: (status: ConnectionStatus) => {
        set({ status });
      },

      setConnecting: (connecting: boolean) => {
        set({ isConnecting: connecting });
      },

      setReconnecting: (reconnecting: boolean) => {
        set({ isReconnecting: reconnecting });
      },

      setConnected: (connected: boolean) => {
        set({ isConnected: connected });
      },

      // Connection info actions
      setUrl: (url: string | null) => {
        set({ url });
      },

      setLastConnected: (date: Date | null) => {
        set({ lastConnected: date });
      },

      setLastDisconnected: (date: Date | null) => {
        set({ lastDisconnected: date });
      },

      // Reconnection management actions
      setReconnectAttempts: (attempts: number) => {
        set({ reconnectAttempts: attempts });
      },

      setMaxReconnectAttempts: (max: number) => {
        set({ maxReconnectAttempts: max });
      },

      setNextReconnectTime: (date: Date | null) => {
        set({ nextReconnectTime: date });
      },

      setAutoReconnect: (enabled: boolean) => {
        set({ autoReconnect: enabled });
      },

      // Error management actions
      setError: (error: string | null, code?: number) => {
        set({ 
          error, 
          errorCode: code || null,
          errorTimestamp: error ? new Date() : null
        });
      },

      clearError: () => {
        set({ 
          error: null, 
          errorCode: null, 
          errorTimestamp: null 
        });
      },

      // Utility actions
      reset: () => {
        set({
          status: 'disconnected',
          isConnecting: false,
          isReconnecting: false,
          isConnected: false,
          url: null,
          lastConnected: null,
          lastDisconnected: null,
          reconnectAttempts: 0,
          nextReconnectTime: null,
          error: null,
          errorCode: null,
          errorTimestamp: null
        });
      }
    }),
    {
      name: 'connection-store',
      partialize: (state) => ({
        // Only persist essential connection state
        url: state.url,
        autoReconnect: state.autoReconnect,
        maxReconnectAttempts: state.maxReconnectAttempts
      })
    }
  )
);
