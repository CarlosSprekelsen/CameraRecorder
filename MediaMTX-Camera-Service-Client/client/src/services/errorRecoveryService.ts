/**
 * Error Recovery Service
 * Provides automatic retry mechanisms and error handling strategies
 * Aligned with server error codes and recovery patterns
 */

import { ERROR_CODES } from '../types/rpc';

/**
 * Retry configuration
 */
export interface RetryConfig {
  maxAttempts: number;
  baseDelay: number;
  maxDelay: number;
  backoffMultiplier: number;
  retryableErrors: number[];
}

/**
 * Default retry configuration
 */
export const DEFAULT_RETRY_CONFIG: RetryConfig = {
  maxAttempts: 3,
  baseDelay: 1000, // 1 second
  maxDelay: 10000, // 10 seconds
  backoffMultiplier: 2,
  retryableErrors: [
    ERROR_CODES.INTERNAL_ERROR,
    ERROR_CODES.MEDIAMTX_SERVICE_UNAVAILABLE,
    ERROR_CODES.CAMERA_NOT_FOUND_OR_DISCONNECTED,
  ],
};

/**
 * Operation result
 */
export interface OperationResult<T> {
  success: boolean;
  data?: T;
  error?: string;
  attempts: number;
  lastAttempt: Date;
}

/**
 * Retryable operation function
 */
export type RetryableOperation<T> = () => Promise<T>;

/**
 * Error recovery strategies
 */
export enum RecoveryStrategy {
  RETRY = 'retry',
  FALLBACK = 'fallback',
  CIRCUIT_BREAKER = 'circuit_breaker',
  GRACEFUL_DEGRADATION = 'graceful_degradation',
}

/**
 * Circuit breaker state
 */
export interface CircuitBreakerState {
  isOpen: boolean;
  failureCount: number;
  lastFailureTime: Date | null;
  threshold: number;
  timeout: number;
}

/**
 * Error Recovery Service Class
 */
export class ErrorRecoveryService {
  private retryConfig: RetryConfig;
  private circuitBreakers: Map<string, CircuitBreakerState>;

  constructor(config: Partial<RetryConfig> = {}) {
    this.retryConfig = { ...DEFAULT_RETRY_CONFIG, ...config };
    this.circuitBreakers = new Map();
  }

  /**
   * Execute operation with retry mechanism
   */
  async executeWithRetry<T>(
    operation: RetryableOperation<T>,
    operationName: string = 'unknown',
    customConfig?: Partial<RetryConfig>
  ): Promise<OperationResult<T>> {
    const config = { ...this.retryConfig, ...customConfig };
    let lastError: Error | null = null;
    let attempt = 0;

    // Check circuit breaker
    if (this.isCircuitBreakerOpen(operationName)) {
      return {
        success: false,
        error: `Circuit breaker is open for ${operationName}`,
        attempts: 0,
        lastAttempt: new Date(),
      };
    }

    while (attempt < config.maxAttempts) {
      attempt++;
      
      try {
        const result = await operation();
        
        // Success - reset circuit breaker
        this.resetCircuitBreaker(operationName);
        
        return {
          success: true,
          data: result,
          attempts: attempt,
          lastAttempt: new Date(),
        };
        
      } catch (error) {
        lastError = error instanceof Error ? error : new Error(String(error));
        
        // Check if error is retryable
        if (!this.isRetryableError(lastError, config.retryableErrors)) {
          this.recordFailure(operationName);
          return {
            success: false,
            error: lastError.message,
            attempts: attempt,
            lastAttempt: new Date(),
          };
        }

        // If this is the last attempt, don't wait
        if (attempt >= config.maxAttempts) {
          this.recordFailure(operationName);
          return {
            success: false,
            error: lastError.message,
            attempts: attempt,
            lastAttempt: new Date(),
          };
        }

        // Calculate delay with exponential backoff
        const delay = this.calculateBackoffDelay(attempt, config);
        console.log(`🔄 Retrying ${operationName} in ${delay}ms (attempt ${attempt}/${config.maxAttempts})`);
        
        await this.sleep(delay);
      }
    }

    this.recordFailure(operationName);
    return {
      success: false,
      error: lastError?.message || 'Operation failed after all retry attempts',
      attempts: attempt,
      lastAttempt: new Date(),
    };
  }

  /**
   * Execute operation with fallback
   */
  async executeWithFallback<T>(
    primaryOperation: RetryableOperation<T>,
    fallbackOperation: RetryableOperation<T>,
    operationName: string = 'unknown'
  ): Promise<OperationResult<T>> {
    try {
      const result = await this.executeWithRetry(primaryOperation, operationName);
      if (result.success) {
        return result;
      }
    } catch (error) {
      console.warn(`Primary operation failed for ${operationName}, trying fallback`);
    }

    // Try fallback operation
    try {
      const fallbackResult = await this.executeWithRetry(fallbackOperation, `${operationName}_fallback`);
      return {
        ...fallbackResult,
        data: fallbackResult.data,
      };
    } catch (error) {
      return {
        success: false,
        error: `Both primary and fallback operations failed: ${error instanceof Error ? error.message : String(error)}`,
        attempts: 0,
        lastAttempt: new Date(),
      };
    }
  }

  /**
   * Execute operation with graceful degradation
   */
  async executeWithGracefulDegradation<T>(
    operation: RetryableOperation<T>,
    fallbackValue: T,
    operationName: string = 'unknown'
  ): Promise<OperationResult<T>> {
    const result = await this.executeWithRetry(operation, operationName);
    
    if (result.success) {
      return result;
    }

    // Return fallback value on failure
    return {
      success: true,
      data: fallbackValue,
      attempts: result.attempts,
      lastAttempt: result.lastAttempt,
    };
  }

  /**
   * Check if error is retryable
   */
  private isRetryableError(error: Error, retryableErrors: number[]): boolean {
    // Check for network errors
    if (error.message.includes('Network Error') || 
        error.message.includes('Failed to fetch') ||
        error.message.includes('WebSocket')) {
      return true;
    }

    // Check for specific error codes
    for (const errorCode of retryableErrors) {
      if (error.message.includes(errorCode.toString())) {
        return true;
      }
    }

    // Don't retry authentication errors
    if (error.message.includes('Authentication') || 
        error.message.includes('Unauthorized') ||
        error.message.includes('Forbidden')) {
      return false;
    }

    // Don't retry validation errors
    if (error.message.includes('Invalid') || 
        error.message.includes('Validation')) {
      return false;
    }

    return true;
  }

  /**
   * Calculate backoff delay
   */
  private calculateBackoffDelay(attempt: number, config: RetryConfig): number {
    const delay = config.baseDelay * Math.pow(config.backoffMultiplier, attempt - 1);
    return Math.min(delay, config.maxDelay);
  }

  /**
   * Sleep utility
   */
  private sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  /**
   * Circuit breaker methods
   */
  private isCircuitBreakerOpen(operationName: string): boolean {
    const circuitBreaker = this.circuitBreakers.get(operationName);
    if (!circuitBreaker) return false;

    if (circuitBreaker.isOpen) {
      const timeSinceLastFailure = Date.now() - (circuitBreaker.lastFailureTime?.getTime() || 0);
      if (timeSinceLastFailure > circuitBreaker.timeout) {
        // Reset circuit breaker
        this.resetCircuitBreaker(operationName);
        return false;
      }
      return true;
    }

    return false;
  }

  private recordFailure(operationName: string): void {
    let circuitBreaker = this.circuitBreakers.get(operationName);
    
    if (!circuitBreaker) {
      circuitBreaker = {
        isOpen: false,
        failureCount: 0,
        lastFailureTime: null,
        threshold: 5,
        timeout: 30000, // 30 seconds
      };
      this.circuitBreakers.set(operationName, circuitBreaker);
    }

    circuitBreaker.failureCount++;
    circuitBreaker.lastFailureTime = new Date();

    if (circuitBreaker.failureCount >= circuitBreaker.threshold) {
      circuitBreaker.isOpen = true;
      console.warn(`🚨 Circuit breaker opened for ${operationName}`);
    }
  }

  private resetCircuitBreaker(operationName: string): void {
    const circuitBreaker = this.circuitBreakers.get(operationName);
    if (circuitBreaker) {
      circuitBreaker.isOpen = false;
      circuitBreaker.failureCount = 0;
      circuitBreaker.lastFailureTime = null;
    }
  }

  /**
   * Get circuit breaker status
   */
  getCircuitBreakerStatus(operationName: string): CircuitBreakerState | null {
    return this.circuitBreakers.get(operationName) || null;
  }

  /**
   * Reset all circuit breakers
   */
  resetAllCircuitBreakers(): void {
    this.circuitBreakers.clear();
  }

  /**
   * Update retry configuration
   */
  updateRetryConfig(config: Partial<RetryConfig>): void {
    this.retryConfig = { ...this.retryConfig, ...config };
  }
}

/**
 * Global error recovery service instance
 */
export const errorRecoveryService = new ErrorRecoveryService();

/**
 * Usage Examples for stores:
 * 
 * Camera Store:
 * ```typescript
 * const result = await errorRecoveryService.executeWithRetry(
 *   () => wsService.call('get_camera_list', {}),
 *   'get_camera_list'
 * );
 * ```
 * 
 * File Store:
 * ```typescript
 * const result = await errorRecoveryService.executeWithRetry(
 *   () => wsService.call('delete_recording', { filename }),
 *   `delete_recording_${filename}`,
 *   { maxAttempts: 2 }
 * );
 * ```
 * 
 * Connection Store:
 * ```typescript
 * const result = await errorRecoveryService.executeWithRetry(
 *   () => wsService.connect(),
 *   'websocket_connect',
 *   { maxAttempts: 5, baseDelay: 2000 }
 * );
 * ```
 */

export default errorRecoveryService;
