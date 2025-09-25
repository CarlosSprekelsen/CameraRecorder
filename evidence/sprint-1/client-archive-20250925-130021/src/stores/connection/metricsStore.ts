/**
 * Metrics Store - Connection Performance Metrics
 * 
 * Architecture: Single Responsibility Principle
 * - Handles only connection performance metrics
 * - Separated from connection state and health
 * - Provides metrics tracking and analysis
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';

/**
 * Connection metrics state interface
 */
export interface MetricsStoreState {
  // Message metrics
  messageCount: number;
  errorCount: number;
  lastMessageTime: Date | null;
  
  // Performance metrics
  responseTimes: number[];
  averageResponseTime: number;
  maxResponseTime: number;
  minResponseTime: number;
  
  // Connection metrics
  connectionUptime: number | null;
  totalConnections: number;
  successfulConnections: number;
  failedConnections: number;
  
  // Data transfer metrics
  bytesSent: number;
  bytesReceived: number;
  messagesPerSecond: number;
  
  // Configuration
  maxResponseTimeHistory: number;
  metricsResetTime: Date | null;
}

/**
 * Metrics store actions interface
 */
export interface MetricsStoreActions {
  // Message metrics
  incrementMessageCount: () => void;
  incrementErrorCount: () => void;
  setLastMessageTime: (date: Date | null) => void;
  
  // Performance metrics
  addResponseTime: (time: number) => void;
  updateResponseTime: (time: number) => void;
  calculateAverageResponseTime: () => void;
  
  // Connection metrics
  setConnectionUptime: (uptime: number | null) => void;
  incrementTotalConnections: () => void;
  incrementSuccessfulConnections: () => void;
  incrementFailedConnections: () => void;
  
  // Data transfer metrics
  addBytesSent: (bytes: number) => void;
  addBytesReceived: (bytes: number) => void;
  updateMessagesPerSecond: () => void;
  
  // Utility actions
  reset: () => void;
  resetMetrics: () => void;
  exportMetrics: () => any;
}

/**
 * Metrics store type
 */
type MetricsStore = MetricsStoreState & MetricsStoreActions;

/**
 * Metrics store implementation
 */
export const useMetricsStore = create<MetricsStore>()(
  devtools(
    (set, get) => ({
      // Initial state
      messageCount: 0,
      errorCount: 0,
      lastMessageTime: null,
      
      responseTimes: [],
      averageResponseTime: 0,
      maxResponseTime: 0,
      minResponseTime: 0,
      
      connectionUptime: null,
      totalConnections: 0,
      successfulConnections: 0,
      failedConnections: 0,
      
      bytesSent: 0,
      bytesReceived: 0,
      messagesPerSecond: 0,
      
      maxResponseTimeHistory: 100,
      metricsResetTime: null,

      // Message metrics actions
      incrementMessageCount: () => {
        const current = get().messageCount;
        set({ 
          messageCount: current + 1,
          lastMessageTime: new Date()
        });
      },

      incrementErrorCount: () => {
        const current = get().errorCount;
        set({ errorCount: current + 1 });
      },

      setLastMessageTime: (date: Date | null) => {
        set({ lastMessageTime: date });
      },

      // Performance metrics actions
      addResponseTime: (time: number) => {
        const { responseTimes, maxResponseTimeHistory } = get();
        const newTimes = [...responseTimes, time];
        
        // Keep only the most recent response times
        if (newTimes.length > maxResponseTimeHistory) {
          newTimes.shift();
        }
        
        set({ responseTimes: newTimes });
        get().calculateAverageResponseTime();
      },

      updateResponseTime: (time: number) => {
        get().addResponseTime(time);
      },

      calculateAverageResponseTime: () => {
        const { responseTimes } = get();
        if (responseTimes.length === 0) {
          set({ 
            averageResponseTime: 0,
            maxResponseTime: 0,
            minResponseTime: 0
          });
          return;
        }
        
        const sum = responseTimes.reduce((acc, time) => acc + time, 0);
        const average = sum / responseTimes.length;
        const max = Math.max(...responseTimes);
        const min = Math.min(...responseTimes);
        
        set({ 
          averageResponseTime: average,
          maxResponseTime: max,
          minResponseTime: min
        });
      },

      // Connection metrics actions
      setConnectionUptime: (uptime: number | null) => {
        set({ connectionUptime: uptime });
      },

      incrementTotalConnections: () => {
        const current = get().totalConnections;
        set({ totalConnections: current + 1 });
      },

      incrementSuccessfulConnections: () => {
        const current = get().successfulConnections;
        set({ successfulConnections: current + 1 });
      },

      incrementFailedConnections: () => {
        const current = get().failedConnections;
        set({ failedConnections: current + 1 });
      },

      // Data transfer metrics actions
      addBytesSent: (bytes: number) => {
        const current = get().bytesSent;
        set({ bytesSent: current + bytes });
      },

      addBytesReceived: (bytes: number) => {
        const current = get().bytesReceived;
        set({ bytesReceived: current + bytes });
      },

      updateMessagesPerSecond: () => {
        // This would typically be calculated based on a time window
        // For now, we'll use a simple calculation
        const { messageCount } = get();
        const now = Date.now();
        const resetTime = get().metricsResetTime;
        
        if (resetTime) {
          const timeDiff = (now - resetTime.getTime()) / 1000; // seconds
          const mps = timeDiff > 0 ? messageCount / timeDiff : 0;
          set({ messagesPerSecond: mps });
        }
      },

      // Utility actions
      reset: () => {
        set({
          messageCount: 0,
          errorCount: 0,
          lastMessageTime: null,
          responseTimes: [],
          averageResponseTime: 0,
          maxResponseTime: 0,
          minResponseTime: 0,
          connectionUptime: null,
          totalConnections: 0,
          successfulConnections: 0,
          failedConnections: 0,
          bytesSent: 0,
          bytesReceived: 0,
          messagesPerSecond: 0,
          metricsResetTime: new Date()
        });
      },

      resetMetrics: () => {
        get().reset();
      },

      exportMetrics: () => {
        const state = get();
        return {
          timestamp: new Date().toISOString(),
          messageCount: state.messageCount,
          errorCount: state.errorCount,
          averageResponseTime: state.averageResponseTime,
          maxResponseTime: state.maxResponseTime,
          minResponseTime: state.minResponseTime,
          connectionUptime: state.connectionUptime,
          totalConnections: state.totalConnections,
          successfulConnections: state.successfulConnections,
          failedConnections: state.failedConnections,
          bytesSent: state.bytesSent,
          bytesReceived: state.bytesReceived,
          messagesPerSecond: state.messagesPerSecond,
          successRate: state.totalConnections > 0 
            ? (state.successfulConnections / state.totalConnections) * 100 
            : 0
        };
      }
    }),
    {
      name: 'metrics-store',
      partialize: (state) => ({
        // Persist metrics configuration
        maxResponseTimeHistory: state.maxResponseTimeHistory
      })
    }
  )
);
