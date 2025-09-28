/**
 * API Client Abstraction Layer
 * 
 * Architecture requirement: "Clear separation between presentation, application control, and infrastructure concerns" (Section 4.1)
 * This layer abstracts the WebSocket communication from business logic
 */

import { WebSocketService } from '../websocket/WebSocketService';
import { LoggerService } from '../logger/LoggerService';
import { PerformanceMonitor } from '../monitoring/PerformanceMonitor';

export interface APIClientConfig {
  timeout?: number;
  retryAttempts?: number;
  retryDelay?: number;
}

export class APIClient {
  private performanceMonitor: PerformanceMonitor;

  constructor(
    private wsService: WebSocketService,
    private logger: LoggerService,
    private config: APIClientConfig = {}
  ) {
    this.performanceMonitor = new PerformanceMonitor(logger);
  }

  /**
   * Execute RPC call with proper abstraction
   * Architecture requirement: Services should not directly access WebSocket
   */
  async call<T = any>(method: string, params: Record<string, unknown> = {}): Promise<T> {
    // Start performance monitoring
    const endTimer = this.performanceMonitor.startCommandTimer();
    
    try {
      this.logger.info(`Executing RPC call: ${method}`, params);
      
      const result = await this.wsService.sendRPC<T>(method, params);
      
      // Record successful operation
      this.performanceMonitor.recordSuccess();
      endTimer();
      
      this.logger.info(`RPC call successful: ${method}`);
      return result;
    } catch (error) {
      // Record failed operation
      this.performanceMonitor.recordFailure();
      endTimer();
      
      this.logger.error(`RPC call failed: ${method}`, error as Record<string, unknown>);
      throw error;
    }
  }

  /**
   * Execute authenticated RPC call
   * Architecture requirement: Centralized authentication handling
   */
  async authenticatedCall<T = any>(method: string, params: Record<string, unknown> = {}): Promise<T> {
    // Authentication is handled by the WebSocket service session
    return this.call<T>(method, params);
  }

  /**
   * Execute batch RPC calls
   * Architecture requirement: Efficient batch operations
   */
  async batchCall<T = any>(calls: Array<{method: string, params: Record<string, unknown>}>): Promise<T[]> {
    const results: T[] = [];
    
    for (const call of calls) {
      const result = await this.call<T>(call.method, call.params);
      results.push(result);
    }
    
    return results;
  }

  /**
   * Check if client is connected
   * Architecture requirement: Connection state management
   */
  isConnected(): boolean {
    return this.wsService.isConnected;
  }

  /**
   * Get connection status
   * Architecture requirement: Status monitoring
   */
  getConnectionStatus(): { connected: boolean; ready: boolean } {
    return {
      connected: this.wsService.isConnected,
      ready: this.wsService.isConnected
    };
  }
}
