/**
 * IAPIClient Interface - Architecture Compliance
 * 
 * Architecture requirement: "IAPIClient interface between business services and transport layer" (ADR-007)
 * Defines the contract for API client implementations
 */

import { RpcMethod } from '../../types/api';

export interface ConnectionStatus {
  connected: boolean;
  ready: boolean;
}

/**
 * API Client Interface
 * Architecture requirement: Section 5.3.1 - API Client Interface
 * ADR-007 requirement: "IAPIClient interface between business services and transport layer"
 */
export interface IAPIClient {
  /**
   * Execute RPC call
   * Architecture requirement: call<T>(method: string, params?: Record<string, any>): Promise<T>
   */
  call<T = any>(method: RpcMethod, params?: Record<string, unknown>): Promise<T>;

  /**
   * Execute batch RPC calls
   * Architecture requirement: batchCall<T>(calls: Array<{method: string, params: Record<string, unknown>}>): Promise<T[]>
   */
  batchCall<T = any>(calls: Array<{method: RpcMethod, params: Record<string, unknown>}>): Promise<T[]>;

  /**
   * Connect to the server
   * Architecture requirement: APIClient should handle connection management
   * ADR-007 requirement: Services need connection management through abstraction layer
   */
  connect(): Promise<void>;

  /**
   * Disconnect from the server
   * Architecture requirement: APIClient should handle connection management
   * ADR-007 requirement: Services need connection management through abstraction layer
   */
  disconnect(): Promise<void>;

  /**
   * Check if client is connected
   * Architecture requirement: isConnected(): boolean
   */
  isConnected(): boolean;

  /**
   * Get connection status
   * Architecture requirement: getConnectionStatus(): ConnectionStatus
   */
  getConnectionStatus(): ConnectionStatus;
}
