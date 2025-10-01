import { IAPIClient } from '../abstraction/IAPIClient';
import { LoggerService } from '../logger/LoggerService';

/**
 * Base Service Class - Common service patterns and utilities
 * 
 * Provides common functionality for all service classes including:
 * - WebSocket connection validation
 * - Standardized logging patterns
 * - Error handling with consistent messaging
 * - RPC call wrapper with logging
 * 
 * @abstract
 * @class BaseService
 * 
 * @example
 * ```typescript
 * class MyService extends BaseService {
 *   async myMethod(): Promise<MyResult> {
 *     return this.callWithLogging('my_method', {}, 'MyMethod');
 *   }
 * }
 * ```
 */
export abstract class BaseService {
  constructor(
    protected apiClient: IAPIClient,
    protected logger: LoggerService,
  ) {}

  /**
   * Validates WebSocket connection before making API calls
   * @throws {Error} When WebSocket is not connected
   */
  protected validateConnection(): void {
    if (!this.apiClient.isConnected()) {
      throw new Error('WebSocket not connected');
    }
  }

  /**
   * Makes RPC call with standardized logging and error handling
   * @param method - RPC method name
   * @param params - Method parameters
   * @param operationName - Human-readable operation name for logging
   * @returns Promise with RPC response
   */
  protected async callWithLogging<T>(
    method: string,
    params: Record<string, unknown> = {},
    operationName?: string
  ): Promise<T> {
    const logPrefix = operationName || method;
    
    try {
      this.validateConnection();
      this.logger.info(`${method} request`, params);
      
      const result = await this.apiClient.call<T>(method, params);
      return result;
    } catch (error) {
      this.logger.error(`${method} failed`, error as Record<string, unknown>);
      throw error;
    }
  }

  /**
   * Makes RPC call with connection validation only (no logging)
   * For methods that don't need request/response logging
   * @param method - RPC method name
   * @param params - Method parameters (optional)
   * @returns Promise with RPC response
   */
  protected async callWithValidation<T>(
    method: string,
    params?: Record<string, unknown>
  ): Promise<T> {
    this.validateConnection();
    return this.apiClient.call<T>(method, params || {});
  }

  /**
   * Logs service initialization
   * @param serviceName - Name of the service being initialized
   */
  protected logInitialization(serviceName: string): void {
    this.logger.info(`${serviceName} initialized`);
  }
}
