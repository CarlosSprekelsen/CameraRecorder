/**
 * WebSocket Cleanup Utilities
 * 
 * Provides proper cleanup for WebSocket connections to prevent resource leaks
 */

import { WebSocketService } from '../../src/services/websocket/WebSocketService';

export class WebSocketCleanupManager {
  private static activeConnections: WebSocketService[] = [];

  /**
   * Register a WebSocket connection for cleanup tracking
   */
  static registerConnection(connection: WebSocketService): void {
    this.activeConnections.push(connection);
  }

  /**
   * Unregister a WebSocket connection from cleanup tracking
   */
  static unregisterConnection(connection: WebSocketService): void {
    const index = this.activeConnections.indexOf(connection);
    if (index > -1) {
      this.activeConnections.splice(index, 1);
    }
  }

  /**
   * Clean up all registered WebSocket connections
   */
  static async cleanupAllConnections(): Promise<void> {
    const cleanupPromises = this.activeConnections.map(async (connection) => {
      try {
        if (connection && typeof connection.disconnect === 'function') {
          connection.disconnect();
        }
      } catch (error) {
        console.warn('Error cleaning up WebSocket connection:', error);
      }
    });

    await Promise.all(cleanupPromises);
    this.activeConnections = [];
  }

  /**
   * Force cleanup with timeout
   */
  static async forceCleanup(timeoutMs: number = 1000): Promise<void> {
    await this.cleanupAllConnections();
    
    // Wait for cleanup to complete
    await new Promise(resolve => setTimeout(resolve, timeoutMs));
    
    // Force garbage collection if available
    if (typeof global.gc === 'function') {
      global.gc();
    }
  }

  /**
   * Get count of active connections
   */
  static getActiveConnectionCount(): number {
    return this.activeConnections.length;
  }
}

// FIXED: Global cleanup on process exit
process.on('exit', () => {
  WebSocketCleanupManager.cleanupAllConnections();
});

process.on('SIGINT', () => {
  WebSocketCleanupManager.cleanupAllConnections();
  process.exit(0);
});

process.on('SIGTERM', () => {
  WebSocketCleanupManager.cleanupAllConnections();
  process.exit(0);
});
