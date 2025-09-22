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
