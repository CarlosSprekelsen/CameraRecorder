/**
 * Health Store - Connection Health Monitoring
 * 
 * Architecture: Single Responsibility Principle
 * - Handles only connection health and quality metrics
 * - Separated from core connection state
 * - Provides health monitoring and quality assessment
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';

/**
 * Connection health state interface
 */
export interface HealthStoreState {
  // Health status
  isHealthy: boolean;
  healthScore: number;
  connectionQuality: 'excellent' | 'good' | 'poor' | 'unstable';
  
  // Health monitoring
  lastHeartbeat: Date | null;
  heartbeatInterval: number;
  missedHeartbeats: number;
  maxMissedHeartbeats: number;
  
  // Quality metrics
  latency: number | null;
  packetLoss: number | null;
  jitter: number | null;
  
  // Health history
  healthHistory: Array<{
    timestamp: Date;
    score: number;
    quality: string;
  }>;
  maxHistorySize: number;
}

/**
 * Health store actions interface
 */
export interface HealthStoreActions {
  // Health status management
  setHealthy: (healthy: boolean) => void;
  setHealthScore: (score: number) => void;
  updateHealthScore: (delta: number) => void;
  setConnectionQuality: (quality: 'excellent' | 'good' | 'poor' | 'unstable') => void;
  
  // Heartbeat management
  setLastHeartbeat: (date: Date | null) => void;
  setHeartbeatInterval: (interval: number) => void;
  incrementMissedHeartbeats: () => void;
  resetMissedHeartbeats: () => void;
  
  // Quality metrics
  setLatency: (latency: number | null) => void;
  setPacketLoss: (loss: number | null) => void;
  setJitter: (jitter: number | null) => void;
  
  // Health history
  addHealthRecord: (score: number, quality: string) => void;
  clearHealthHistory: () => void;
  
  // WebSocket-based health methods
  getSystemStatus: () => Promise<any>;
  getSystemMetrics: () => Promise<any>;
  refreshHealth: () => Promise<void>;
  
  // Utility actions
  reset: () => void;
  calculateHealthScore: () => number;
}

/**
 * Health store type
 */
type HealthStore = HealthStoreState & HealthStoreActions;

/**
 * Health store implementation
 */
export const useHealthStore = create<HealthStore>()(
  devtools(
    (set, get) => ({
      // Initial state
      isHealthy: false,
      healthScore: 0,
      connectionQuality: 'unstable',
      
      lastHeartbeat: null,
      heartbeatInterval: 30000, // 30 seconds
      missedHeartbeats: 0,
      maxMissedHeartbeats: 3,
      
      latency: null,
      packetLoss: null,
      jitter: null,
      
      healthHistory: [],
      maxHistorySize: 100,

      // Health status management actions
      setHealthy: (healthy: boolean) => {
        set({ isHealthy: healthy });
      },

      setHealthScore: (score: number) => {
        const clampedScore = Math.max(0, Math.min(100, score));
        set({ healthScore: clampedScore });
      },

      updateHealthScore: (delta: number) => {
        const currentScore = get().healthScore;
        const newScore = Math.max(0, Math.min(100, currentScore + delta));
        set({ healthScore: newScore });
      },

      setConnectionQuality: (quality: 'excellent' | 'good' | 'poor' | 'unstable') => {
        set({ connectionQuality: quality });
      },

      // Heartbeat management actions
      setLastHeartbeat: (date: Date | null) => {
        set({ lastHeartbeat: date });
      },

      setHeartbeatInterval: (interval: number) => {
        set({ heartbeatInterval: interval });
      },

      incrementMissedHeartbeats: () => {
        const current = get().missedHeartbeats;
        set({ missedHeartbeats: current + 1 });
      },

      resetMissedHeartbeats: () => {
        set({ missedHeartbeats: 0 });
      },

      // Quality metrics actions
      setLatency: (latency: number | null) => {
        set({ latency });
      },

      setPacketLoss: (loss: number | null) => {
        set({ packetLoss: loss });
      },

      setJitter: (jitter: number | null) => {
        set({ jitter });
      },

      // Health history actions
      addHealthRecord: (score: number, quality: string) => {
        const { healthHistory, maxHistorySize } = get();
        const newRecord = {
          timestamp: new Date(),
          score,
          quality
        };
        
        const updatedHistory = [...healthHistory, newRecord];
        if (updatedHistory.length > maxHistorySize) {
          updatedHistory.shift(); // Remove oldest record
        }
        
        set({ healthHistory: updatedHistory });
      },

      clearHealthHistory: () => {
        set({ healthHistory: [] });
      },

      // WebSocket-based health methods
      getSystemStatus: async () => {
        try {
          const { websocketService } = await import('../../services/websocket');
          const wsService = websocketService;

          if (!wsService.isConnected()) {
            throw new Error('WebSocket not connected');
          }

          console.log('Getting system status via WebSocket');
          const result = await wsService.call('get_status', {});
          
          // Update health based on system status
          const isHealthy = result.status === 'healthy';
          set({ isHealthy });
          
          return result;
        } catch (error) {
          console.error('Failed to get system status:', error);
          set({ isHealthy: false });
          throw error;
        }
      },

      getSystemMetrics: async () => {
        try {
          const { websocketService } = await import('../../services/websocket');
          const wsService = websocketService;

          if (!wsService.isConnected()) {
            throw new Error('WebSocket not connected');
          }

          console.log('Getting system metrics via WebSocket');
          const result = await wsService.call('get_metrics', {});
          
          // Update health score based on metrics
          if (result.average_response_time) {
            const latency = result.average_response_time;
            set({ latency });
            
            // Calculate health score based on response time
            let healthScore = 100;
            if (latency > 100) {
              healthScore -= Math.min(30, (latency - 100) / 10);
            }
            set({ healthScore: Math.max(0, healthScore) });
          }
          
          return result;
        } catch (error) {
          console.error('Failed to get system metrics:', error);
          set({ isHealthy: false });
          throw error;
        }
      },

      refreshHealth: async () => {
        try {
          // Get both system status and metrics
          const [status, metrics] = await Promise.all([
            get().getSystemStatus(),
            get().getSystemMetrics()
          ]);
          
          // Update connection quality based on health score
          const { healthScore } = get();
          let connectionQuality: 'excellent' | 'good' | 'poor' | 'unstable';
          
          if (healthScore >= 90) {
            connectionQuality = 'excellent';
          } else if (healthScore >= 70) {
            connectionQuality = 'good';
          } else if (healthScore >= 40) {
            connectionQuality = 'poor';
          } else {
            connectionQuality = 'unstable';
          }
          
          set({ connectionQuality });
          
          // Add to health history
          get().addHealthRecord(healthScore, connectionQuality);
          
          console.log('Health refreshed successfully', { status, metrics, healthScore, connectionQuality });
        } catch (error) {
          console.error('Failed to refresh health:', error);
          set({ isHealthy: false, connectionQuality: 'unstable' });
          throw error;
        }
      },

      // Utility actions
      reset: () => {
        set({
          isHealthy: false,
          healthScore: 0,
          connectionQuality: 'unstable',
          lastHeartbeat: null,
          missedHeartbeats: 0,
          latency: null,
          packetLoss: null,
          jitter: null,
          healthHistory: []
        });
      },

      calculateHealthScore: () => {
        const { latency, packetLoss, jitter, missedHeartbeats } = get();
        let score = 100;

        // Penalize high latency
        if (latency !== null) {
          if (latency > 1000) score -= 30; // > 1s
          else if (latency > 500) score -= 20; // > 500ms
          else if (latency > 200) score -= 10; // > 200ms
        }

        // Penalize packet loss
        if (packetLoss !== null) {
          if (packetLoss > 5) score -= 40; // > 5%
          else if (packetLoss > 2) score -= 20; // > 2%
          else if (packetLoss > 0.5) score -= 10; // > 0.5%
        }

        // Penalize high jitter
        if (jitter !== null) {
          if (jitter > 100) score -= 20; // > 100ms
          else if (jitter > 50) score -= 10; // > 50ms
        }

        // Penalize missed heartbeats
        if (missedHeartbeats > 0) {
          score -= missedHeartbeats * 10;
        }

        return Math.max(0, Math.min(100, score));
      }
    }),
    {
      name: 'health-store',
      partialize: (state) => ({
        // Persist health configuration
        heartbeatInterval: state.heartbeatInterval,
        maxMissedHeartbeats: state.maxMissedHeartbeats,
        maxHistorySize: state.maxHistorySize
      })
    }
  )
);
