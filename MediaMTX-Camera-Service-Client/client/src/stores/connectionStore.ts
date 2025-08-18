/**
 * Connection state management store
 * Sprint 3: Enhanced for real server integration
 * Handles WebSocket connection status and reconnection logic
 * 
 * Sprint 3 Updates:
 * - Comprehensive connection state tracking (CONNECTING, CONNECTED, DISCONNECTED, ERROR)
 * - Enhanced error handling and recovery mechanisms
 * - Connection retry logic with user control
 * - Connection status indicators throughout UI
 * - Graceful degradation when disconnected
 * - Connection health monitoring and alerts
 * - Real-time connection metrics
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import type { ConnectionStatus } from '../types';
import type { WebSocketService } from '../services/websocket';
import { createWebSocketService, defaultWebSocketConfig } from '../services/websocket';

/**
 * Enhanced connection state interface
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
  nextReconnectTime: Date | null;
  
  // Error state
  error: string | null;
  errorCode: number | null;
  errorTimestamp: Date | null;
  
  // WebSocket service reference
  wsService: WebSocketService | null;
  
  // Connection health
  isHealthy: boolean;
  lastHeartbeat: Date | null;
  connectionUptime: number | null;
  healthScore: number; // 0-100
  
  // Performance metrics
  responseTime: number | null;
  messageCount: number;
  errorCount: number;
  
  // User preferences
  autoReconnect: boolean;
  showConnectionAlerts: boolean;
  
  // Connection quality
  connectionQuality: 'excellent' | 'good' | 'poor' | 'unstable';
  latency: number | null;
  packetLoss: number | null;
}

/**
 * Enhanced connection store actions interface
 */
interface ConnectionActions {
  // Connection management
  connect: (url?: string) => Promise<void>;
  disconnect: () => void;
  reconnect: () => Promise<void>;
  forceReconnect: () => Promise<void>;
  
  // Status updates
  setStatus: (status: ConnectionStatus) => void;
  setConnecting: (isConnecting: boolean) => void;
  setReconnecting: (isReconnecting: boolean) => void;
  
  // Error handling
  setError: (error: string | null, code?: number) => void;
  clearError: () => void;
  handleConnectionError: (error: Error) => void;
  
  // Reconnection management
  setReconnectAttempts: (attempts: number) => void;
  resetReconnectAttempts: () => void;
  setNextReconnectTime: (time: Date | null) => void;
  
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
  updateHealthScore: (score: number) => void;
  
  // Performance tracking
  updateResponseTime: (time: number) => void;
  incrementMessageCount: () => void;
  incrementErrorCount: () => void;
  resetMetrics: () => void;
  
  // User preferences
  setAutoReconnect: (enabled: boolean) => void;
  setShowConnectionAlerts: (enabled: boolean) => void;
  
  // Connection quality
  updateConnectionQuality: (quality: ConnectionState['connectionQuality']) => void;
  updateLatency: (latency: number) => void;
  updatePacketLoss: (loss: number) => void;
  
  // Connection testing
  testConnection: () => Promise<boolean>;
  performHealthCheck: () => Promise<boolean>;
  
  // Utility methods
  getConnectionSummary: () => {
    status: ConnectionStatus;
    isHealthy: boolean;
    uptime: number | null;
    errorCount: number;
    healthScore: number;
    quality: string;
  };
  
  shouldAttemptReconnect: () => boolean;
  getReconnectDelay: () => number;
}

/**
 * Connection store type
 */
type ConnectionStore = ConnectionState & ConnectionActions;

/**
 * Create enhanced connection store
 * Sprint 3: Comprehensive connection state management
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
      nextReconnectTime: null,
      error: null,
      errorCode: null,
      errorTimestamp: null,
      wsService: null,
      isHealthy: false,
      lastHeartbeat: null,
      connectionUptime: null,
      healthScore: 0,
      responseTime: null,
      messageCount: 0,
      errorCount: 0,
      autoReconnect: true,
      showConnectionAlerts: true,
      connectionQuality: 'unstable',
      latency: null,
      packetLoss: null,

      // Connection management
      connect: async (url = defaultWebSocketConfig.url) => {
        try {
          set({ 
            isConnecting: true, 
            error: null,
            errorCode: null,
            errorTimestamp: null,
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

          // Set up enhanced event handlers
          service.onConnect(() => {
            const now = new Date();
            set({ 
              isConnecting: false,
              status: 'connected',
              lastConnected: now,
              reconnectAttempts: 0,
              isHealthy: true,
              healthScore: 100,
              error: null,
              errorCode: null,
              errorTimestamp: null,
              connectionQuality: 'excellent',
              nextReconnectTime: null
            });
            
            // Start health monitoring
            get().startHealthMonitoring();
          });

          service.onDisconnect(() => {
            const now = new Date();
            set({ 
              status: 'disconnected',
              isConnecting: false,
              isReconnecting: false,
              lastDisconnected: now,
              isHealthy: false,
              healthScore: 0,
              connectionQuality: 'unstable'
            });
            
            // Stop health monitoring
            get().stopHealthMonitoring();
            
            // Attempt reconnection if auto-reconnect is enabled
            if (get().autoReconnect && get().shouldAttemptReconnect()) {
              setTimeout(() => get().reconnect(), get().getReconnectDelay());
            }
          });

          service.onError((error) => {
            const now = new Date();
            set({ 
              error: error.message,
              errorCode: error.code || null,
              errorTimestamp: now,
              isConnecting: false,
              isHealthy: false,
              healthScore: Math.max(0, get().healthScore - 20),
              connectionQuality: 'poor'
            });
            
            get().incrementErrorCount();
            get().handleConnectionError(error);
          });

          await service.connect();
          
        } catch (error) {
          const now = new Date();
          const errorMessage = error instanceof Error ? error.message : 'Failed to connect';
          set({ 
            isConnecting: false,
            status: 'error',
            error: errorMessage,
            errorTimestamp: now,
            lastDisconnected: now,
            isHealthy: false,
            healthScore: 0,
            connectionQuality: 'unstable'
          });
          
          get().incrementErrorCount();
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
          healthScore: 0,
          connectionQuality: 'unstable',
          nextReconnectTime: null,
          wsService: null
        });
        
        get().stopHealthMonitoring();
      },

      reconnect: async () => {
        const { url, reconnectAttempts, maxReconnectAttempts, autoReconnect } = get();
        
        if (!autoReconnect) {
          set({ error: 'Auto-reconnect is disabled' });
          return;
        }
        
        if (!url) {
          set({ error: 'No connection URL available for reconnection' });
          return;
        }

        if (reconnectAttempts >= maxReconnectAttempts) {
          set({ 
            status: 'error',
            isReconnecting: false,
            error: 'Max reconnection attempts reached',
            connectionQuality: 'unstable'
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
            error: error instanceof Error ? error.message : 'Reconnection failed',
            connectionQuality: 'unstable'
          });
          
          // Schedule next reconnection attempt
          const delay = get().getReconnectDelay();
          const nextTime = new Date(Date.now() + delay);
          set({ nextReconnectTime: nextTime });
          
          if (get().shouldAttemptReconnect()) {
            setTimeout(() => get().reconnect(), delay);
          }
        }
      },

      forceReconnect: async () => {
        get().disconnect();
        await new Promise(resolve => setTimeout(resolve, 1000)); // Wait 1 second
        await get().connect();
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
      setError: (error: string | null, code?: number) => {
        set({ 
          error, 
          errorCode: code || null,
          errorTimestamp: error ? new Date() : null
        });
        if (error) {
          get().incrementErrorCount();
        }
      },

      clearError: () => {
        set({ 
          error: null,
          errorCode: null,
          errorTimestamp: null
        });
      },

      handleConnectionError: (error: Error) => {
        const { showConnectionAlerts } = get();
        
        // Update connection quality based on error type
        let quality: ConnectionState['connectionQuality'] = 'poor';
        if (error.message.includes('timeout')) {
          quality = 'unstable';
        } else if (error.message.includes('network')) {
          quality = 'poor';
        }
        
        set({ connectionQuality: quality });
        
        // Show alert if enabled
        if (showConnectionAlerts) {
          console.warn('Connection error:', error.message);
        }
      },

      // Reconnection management
      setReconnectAttempts: (attempts: number) => {
        set({ reconnectAttempts: attempts });
      },

      resetReconnectAttempts: () => {
        set({ reconnectAttempts: 0 });
      },

      setNextReconnectTime: (time: Date | null) => {
        set({ nextReconnectTime: time });
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

      updateHealthScore: (score: number) => {
        const clampedScore = Math.max(0, Math.min(100, score));
        set({ healthScore: clampedScore });
        
        // Update connection quality based on health score
        let quality: ConnectionState['connectionQuality'] = 'unstable';
        if (clampedScore >= 90) quality = 'excellent';
        else if (clampedScore >= 70) quality = 'good';
        else if (clampedScore >= 30) quality = 'poor';
        
        set({ connectionQuality: quality });
      },

      // Performance tracking
      updateResponseTime: (time: number) => {
        set({ responseTime: time });
        get().updateLatency(time);
      },

      incrementMessageCount: () => {
        set(state => ({ messageCount: state.messageCount + 1 }));
      },

      incrementErrorCount: () => {
        set(state => ({ errorCount: state.errorCount + 1 }));
      },

      resetMetrics: () => {
        set({ 
          messageCount: 0,
          errorCount: 0,
          responseTime: null,
          latency: null,
          packetLoss: null
        });
      },

      // User preferences
      setAutoReconnect: (enabled: boolean) => {
        set({ autoReconnect: enabled });
      },

      setShowConnectionAlerts: (enabled: boolean) => {
        set({ showConnectionAlerts: enabled });
      },

      // Connection quality
      updateConnectionQuality: (quality: ConnectionState['connectionQuality']) => {
        set({ connectionQuality: quality });
      },

      updateLatency: (latency: number) => {
        set({ latency });
      },

      updatePacketLoss: (loss: number) => {
        set({ packetLoss: loss });
      },

      // Connection testing
      testConnection: async (): Promise<boolean> => {
        try {
          const { wsService } = get();
          if (!wsService || !wsService.isConnected) {
            return false;
          }

          const startTime = performance.now();
          await wsService.call('ping', {});
          const responseTime = performance.now() - startTime;
          
          get().updateResponseTime(responseTime);
          get().incrementMessageCount();
          set({ 
            isHealthy: true, 
            lastHeartbeat: new Date(),
            healthScore: Math.min(100, get().healthScore + 10)
          });
          
          return true;
        } catch (error) {
          set({ isHealthy: false });
          get().incrementErrorCount();
          return false;
        }
      },

      performHealthCheck: async (): Promise<boolean> => {
        const isHealthy = await get().testConnection();
        
        if (isHealthy) {
          // Update uptime
          const { lastConnected } = get();
          if (lastConnected) {
            const uptime = Date.now() - lastConnected.getTime();
            set({ connectionUptime: uptime });
          }
        }
        
        return isHealthy;
      },

      // Utility methods
      getConnectionSummary: () => {
        const { status, isHealthy, connectionUptime, errorCount, healthScore, connectionQuality } = get();
        return {
          status,
          isHealthy,
          uptime: connectionUptime,
          errorCount,
          healthScore,
          quality: connectionQuality
        };
      },

      shouldAttemptReconnect: () => {
        const { autoReconnect, reconnectAttempts, maxReconnectAttempts } = get();
        return autoReconnect && reconnectAttempts < maxReconnectAttempts;
      },

      getReconnectDelay: () => {
        const { reconnectAttempts } = get();
        const baseDelay = defaultWebSocketConfig.reconnectInterval;
        const maxDelay = defaultWebSocketConfig.maxDelay;
        return Math.min(baseDelay * Math.pow(2, reconnectAttempts), maxDelay);
      },

      // Health monitoring
      startHealthMonitoring: () => {
        // Start periodic health checks
        const healthInterval = setInterval(async () => {
          const { status } = get();
          if (status === 'connected') {
            await get().performHealthCheck();
          } else {
            clearInterval(healthInterval);
          }
        }, 30000); // Check every 30 seconds
        
        // Store interval reference for cleanup
        (get() as any).healthInterval = healthInterval;
      },

      stopHealthMonitoring: () => {
        const interval = (get() as any).healthInterval;
        if (interval) {
          clearInterval(interval);
          (get() as any).healthInterval = null;
        }
      },
    }),
    {
      name: 'connection-store',
    }
  )
); 