/**
 * Connection Stores Index
 * 
 * Architecture: Modular Store Organization
 * - Exports all connection-related stores
 * - Provides unified interface for connection state
 * - Enables easy migration from monolithic connection store
 */

export { useConnectionStore } from './connectionStore';
export type { ConnectionStoreState, ConnectionStoreActions } from './connectionStore';

export { useHealthStore } from './healthStore';
export type { HealthStoreState, HealthStoreActions } from './healthStore';

export { useMetricsStore } from './metricsStore';
export type { MetricsStoreState, MetricsStoreActions } from './metricsStore';

/**
 * Unified connection state interface
 * Combines all connection-related stores for easy access
 */
export interface UnifiedConnectionState {
  // Core connection state
  connection: {
    status: string;
    isConnected: boolean;
    isConnecting: boolean;
    isReconnecting: boolean;
    url: string | null;
    lastConnected: Date | null;
    lastDisconnected: Date | null;
    error: string | null;
    errorCode: number | null;
  };
  
  // Health state
  health: {
    isHealthy: boolean;
    healthScore: number;
    connectionQuality: string;
    lastHeartbeat: Date | null;
    latency: number | null;
    packetLoss: number | null;
  };
  
  // Metrics state
  metrics: {
    messageCount: number;
    errorCount: number;
    averageResponseTime: number;
    connectionUptime: number | null;
    bytesSent: number;
    bytesReceived: number;
  };
}

/**
 * Hook to get unified connection state
 * Combines all connection-related stores into a single interface
 */
export function useUnifiedConnectionState(): UnifiedConnectionState {
  const connectionState = useConnectionStore();
  const healthState = useHealthStore();
  const metricsState = useMetricsStore();
  
  return {
    connection: {
      status: connectionState.status,
      isConnected: connectionState.isConnected,
      isConnecting: connectionState.isConnecting,
      isReconnecting: connectionState.isReconnecting,
      url: connectionState.url,
      lastConnected: connectionState.lastConnected,
      lastDisconnected: connectionState.lastDisconnected,
      error: connectionState.error,
      errorCode: connectionState.errorCode
    },
    health: {
      isHealthy: healthState.isHealthy,
      healthScore: healthState.healthScore,
      connectionQuality: healthState.connectionQuality,
      lastHeartbeat: healthState.lastHeartbeat,
      latency: healthState.latency,
      packetLoss: healthState.packetLoss
    },
    metrics: {
      messageCount: metricsState.messageCount,
      errorCount: metricsState.errorCount,
      averageResponseTime: metricsState.averageResponseTime,
      connectionUptime: metricsState.connectionUptime,
      bytesSent: metricsState.bytesSent,
      bytesReceived: metricsState.bytesReceived
    }
  };
}
