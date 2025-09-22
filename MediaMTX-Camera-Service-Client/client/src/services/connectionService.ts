/**
 * Connection Service for MediaMTX Camera Service Client
 * Provides abstraction layer between components and connection store
 * 
 * Architecture Pattern: Service Layer Abstraction
 * - Components should use this service instead of direct store access
 * - Encapsulates connection logic and state management
 * - Provides consistent API for connection operations
 * 
 * Usage:
 * ```typescript
 * const connectionService = ConnectionService.getInstance();
 * await connectionService.connect();
 * ```
 */

import type { ConnectionStoreInterface } from './websocket';

/**
 * Connection service configuration
 */
export interface ConnectionServiceConfig {
  autoReconnect: boolean;
  maxReconnectAttempts: number;
  reconnectInterval: number;
}

/**
 * Connection service error
 */
export class ConnectionServiceError extends Error {
  public code?: number;
  public context?: string;

  constructor(message: string, code?: number, context?: string) {
    super(message);
    this.name = 'ConnectionServiceError';
    this.code = code;
    this.context = context;
  }
}

/**
 * Connection service class
 * Provides abstraction layer for connection operations
 */
export class ConnectionService {
  private static instance: ConnectionService;
  private connectionStore: ConnectionStoreInterface | null = null;
  private config: ConnectionServiceConfig;

  private constructor(config: Partial<ConnectionServiceConfig> = {}) {
    this.config = {
      autoReconnect: true,
      maxReconnectAttempts: 10,
      reconnectInterval: 1000,
      ...config
    };
  }

  /**
   * Get singleton instance
   */
  public static getInstance(config?: Partial<ConnectionServiceConfig>): ConnectionService {
    if (!ConnectionService.instance) {
      ConnectionService.instance = new ConnectionService(config);
    }
    return ConnectionService.instance;
  }

  /**
   * Initialize service with connection store
   */
  public async initialize(): Promise<void> {
    try {
      // Dynamically import to avoid circular dependencies
      const { useConnectionStore } = await import('../stores/connection');
      this.connectionStore = useConnectionStore.getState();
    } catch (error) {
      throw new ConnectionServiceError(
        'Failed to initialize connection service',
        undefined,
        'store_initialization'
      );
    }
  }

  /**
   * Connect to WebSocket server
   */
  public async connect(): Promise<void> {
    if (!this.connectionStore) {
      throw new ConnectionServiceError(
        'Connection service not initialized',
        undefined,
        'not_initialized'
      );
    }

    try {
      // Use WebSocket service for connection
      const { createWebSocketService } = await import('./websocket');
      const wsService = await createWebSocketService();
      await wsService.connect();
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown connection error';
      throw new ConnectionServiceError(
        `Failed to connect: ${errorMessage}`,
        undefined,
        'connection_failed'
      );
    }
  }

  /**
   * Disconnect from WebSocket server
   */
  public async disconnect(): Promise<void> {
    if (!this.connectionStore) {
      throw new ConnectionServiceError(
        'Connection service not initialized',
        undefined,
        'not_initialized'
      );
    }

    try {
      // Use WebSocket service for disconnection
      const { createWebSocketService } = await import('./websocket');
      const wsService = await createWebSocketService();
      await wsService.disconnect();
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown disconnection error';
      throw new ConnectionServiceError(
        `Failed to disconnect: ${errorMessage}`,
        undefined,
        'disconnection_failed'
      );
    }
  }

  /**
   * Force reconnection
   */
  public async forceReconnect(): Promise<void> {
    if (!this.connectionStore) {
      throw new ConnectionServiceError(
        'Connection service not initialized',
        undefined,
        'not_initialized'
      );
    }

    try {
      await this.disconnect();
      await this.connect();
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown reconnection error';
      throw new ConnectionServiceError(
        `Failed to reconnect: ${errorMessage}`,
        undefined,
        'reconnection_failed'
      );
    }
  }

  /**
   * Get connection status
   */
  public getStatus(): {
    isConnected: boolean;
    status: string;
    error: string | null;
    healthScore: number;
  } {
    if (!this.connectionStore) {
      return {
        isConnected: false,
        status: 'not_initialized',
        error: 'Connection service not initialized',
        healthScore: 0
      };
    }

    return {
      isConnected: this.connectionStore.isConnected,
      status: this.connectionStore.status,
      error: this.connectionStore.error,
      healthScore: this.connectionStore.healthScore
    };
  }

  /**
   * Clear connection error
   */
  public clearError(): void {
    if (!this.connectionStore) {
      throw new ConnectionServiceError(
        'Connection service not initialized',
        undefined,
        'not_initialized'
      );
    }

    this.connectionStore.setError(null);
  }

  /**
   * Set auto-reconnect setting
   */
  public setAutoReconnect(enabled: boolean): void {
    if (!this.connectionStore) {
      throw new ConnectionServiceError(
        'Connection service not initialized',
        undefined,
        'not_initialized'
      );
    }

    this.connectionStore.setAutoReconnect?.(enabled);
  }

  /**
   * Get connection metrics
   */
  public getMetrics(): {
    messageCount: number;
    errorCount: number;
    connectionUptime: number | null;
    latency: number | null;
  } {
    if (!this.connectionStore) {
      return {
        messageCount: 0,
        errorCount: 0,
        connectionUptime: null,
        latency: null
      };
    }

    return {
      messageCount: this.connectionStore.messageCount,
      errorCount: this.connectionStore.errorCount,
      connectionUptime: this.connectionStore.connectionUptime,
      latency: this.connectionStore.latency
    };
  }
}

/**
 * Singleton instance for easy access
 */
export const connectionService = ConnectionService.getInstance();

/**
 * Initialize connection service
 * Call this during app startup
 */
export async function initializeConnectionService(): Promise<void> {
  await connectionService.initialize();
}
