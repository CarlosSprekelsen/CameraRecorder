import { errorRecoveryService } from './errorRecoveryService';
import type { JSONRPCError } from '../types/rpc';
import { ERROR_CODES } from '../types/rpc';

/**
 * Enhanced error information
 */
interface ErrorInfo {
  message: string;
  code?: number;
  context: string;
  timestamp: Date;
  userFriendly: string;
  severity: 'info' | 'warning' | 'error' | 'critical';
  recoverable: boolean;
  data?: any;
}

/**
 * Recovery result
 */
interface RecoveryResult {
  success: boolean;
  attempts: number;
  error?: ErrorInfo;
  context: string;
}

/**
 * Error log entry
 */
interface ErrorLog {
  timestamp: Date;
  error: ErrorInfo;
  stack?: string;
  userAgent?: string;
  url?: string;
}

/**
 * Error Handler Service
 * 
 * Provides enhanced error handling, recovery mechanisms, and user-friendly
 * error messages for the application.
 */
class ErrorHandlerService {
  private errorCallbacks: Set<(error: ErrorInfo) => void> = new Set();
  private recoveryCallbacks: Set<(result: RecoveryResult) => void> = new Set();
  private logCallbacks: Set<(log: ErrorLog) => void> = new Set();

  /**
   * Handle JSON-RPC error
   */
  handleJSONRPCError(error: JSONRPCError, context: string): ErrorInfo {
    const errorInfo = this.createErrorInfo(error, context);
    this.logError(errorInfo);
    this.notifyErrorCallbacks(errorInfo);
    return errorInfo;
  }

  /**
   * Handle generic error
   */
  handleError(error: Error, context: string): ErrorInfo {
    const errorInfo = this.createErrorInfoFromGeneric(error, context);
    this.logError(errorInfo);
    this.notifyErrorCallbacks(errorInfo);
    return errorInfo;
  }

  /**
   * Attempt error recovery
   */
  async attemptRecovery<T>(
    operation: () => Promise<T>,
    context: string,
    maxAttempts: number = 3
  ): Promise<RecoveryResult> {
    let attempts = 0;
    let lastError: ErrorInfo | undefined;

    while (attempts < maxAttempts) {
      attempts++;
      
      try {
        await operation();
        
        const result: RecoveryResult = {
          success: true,
          attempts,
          context
        };
        
        this.notifyRecoveryCallbacks(result);
        return result;
      } catch (error: any) {
        lastError = this.createErrorInfoFromGeneric(error, context);
        
        if (attempts === maxAttempts) {
          // Final attempt failed
          const result: RecoveryResult = {
            success: false,
            attempts,
            error: lastError,
            context
          };
          
          this.notifyRecoveryCallbacks(result);
          return result;
        }
        
        // Wait before retry (exponential backoff)
        await this.delay(Math.pow(2, attempts) * 1000);
      }
    }

    // This should never be reached, but TypeScript requires it
    const result: RecoveryResult = {
      success: false,
      attempts,
      error: lastError,
      context
    };
    
    this.notifyRecoveryCallbacks(result);
    return result;
  }

  /**
   * Create user-friendly error message
   */
  createUserFriendlyMessage(error: JSONRPCError | Error, context: string): string {
    if ('code' in error && error.code) {
      return this.getUserFriendlyMessageForCode(error.code, context);
    }
    
    return this.getUserFriendlyMessageForGeneric(error, context);
  }

  /**
   * Get error severity
   */
  getErrorSeverity(error: JSONRPCError | Error): 'info' | 'warning' | 'error' | 'critical' {
    if ('code' in error && error.code) {
      return this.getSeverityForCode(error.code);
    }
    
    return 'error';
  }

  /**
   * Check if error is recoverable
   */
  isErrorRecoverable(error: JSONRPCError | Error): boolean {
    if ('code' in error && error.code) {
      return this.isRecoverableCode(error.code);
    }
    
    return true; // Assume generic errors are recoverable
  }

  /**
   * Event handlers
   */
  onError(callback: (error: ErrorInfo) => void): void {
    this.errorCallbacks.add(callback);
  }

  onRecovery(callback: (result: RecoveryResult) => void): void {
    this.recoveryCallbacks.add(callback);
  }

  onErrorLog(callback: (log: ErrorLog) => void): void {
    this.logCallbacks.add(callback);
  }

  /**
   * Private methods
   */
  private createErrorInfo(error: JSONRPCError, context: string): ErrorInfo {
    return {
      message: error.message,
      code: error.code,
      context,
      timestamp: new Date(),
      userFriendly: this.createUserFriendlyMessage(error, context),
      severity: this.getErrorSeverity(error),
      recoverable: this.isErrorRecoverable(error),
      data: error.data
    };
  }

  private createErrorInfoFromGeneric(error: Error, context: string): ErrorInfo {
    return {
      message: error.message,
      context,
      timestamp: new Date(),
      userFriendly: this.createUserFriendlyMessage(error, context),
      severity: this.getErrorSeverity(error),
      recoverable: this.isErrorRecoverable(error),
      data: { stack: error.stack }
    };
  }

  private getUserFriendlyMessageForCode(code: number, context: string): string {
    switch (code) {
      case ERROR_CODES.CAMERA_ALREADY_RECORDING:
        return 'This camera is currently recording. Please stop the current recording before starting a new one.';
      
      case ERROR_CODES.STORAGE_SPACE_LOW:
        return 'Storage space is running low. Consider deleting old recordings to free up space.';
      
      case ERROR_CODES.STORAGE_SPACE_CRITICAL:
        return 'Storage space is critically low. Recording operations are blocked until space is freed.';
      
      case ERROR_CODES.CAMERA_NOT_FOUND_OR_DISCONNECTED:
        return 'The camera was not found or is currently disconnected. Please check the camera connection and try again.';
      
      case ERROR_CODES.RECORDING_ALREADY_IN_PROGRESS:
        return 'A recording is already in progress. Please wait for it to complete or stop it manually.';
      
      case ERROR_CODES.MEDIAMTX_SERVICE_UNAVAILABLE:
        return 'The MediaMTX service is currently unavailable. Please try again later.';
      
      case ERROR_CODES.AUTHENTICATION_REQUIRED:
        return 'Authentication is required. Please log in to continue.';
      
      case ERROR_CODES.INSUFFICIENT_STORAGE_SPACE:
        return 'There is insufficient storage space available. Please free up some space and try again.';
      
      case ERROR_CODES.CAMERA_CAPABILITY_NOT_SUPPORTED:
        return 'This camera does not support the requested operation. Please check the camera capabilities.';
      
      case ERROR_CODES.INVALID_PARAMS:
        return 'Invalid parameters provided. Please check your input and try again.';
      
      case ERROR_CODES.METHOD_NOT_FOUND:
        return 'The requested operation is not available. Please check if the service supports this feature.';
      
      case ERROR_CODES.INTERNAL_ERROR:
        return 'An internal error occurred. Please try again later or contact support.';
      
      default:
        return `An error occurred while ${context}. Please try again.`;
    }
  }

  private getUserFriendlyMessageForGeneric(error: JSONRPCError | Error, context: string): string {
    const message = error.message || 'An unknown error occurred';
    return `${message} (${context})`;
  }

  private getSeverityForCode(code: number): 'info' | 'warning' | 'error' | 'critical' {
    switch (code) {
      case ERROR_CODES.STORAGE_SPACE_CRITICAL:
        return 'critical';
      
      case ERROR_CODES.STORAGE_SPACE_LOW:
      case ERROR_CODES.CAMERA_ALREADY_RECORDING:
      case ERROR_CODES.RECORDING_ALREADY_IN_PROGRESS:
      case ERROR_CODES.CAMERA_CAPABILITY_NOT_SUPPORTED:
        return 'warning';
      
      case ERROR_CODES.AUTHENTICATION_REQUIRED:
      case ERROR_CODES.MEDIAMTX_SERVICE_UNAVAILABLE:
      case ERROR_CODES.INSUFFICIENT_STORAGE_SPACE:
      case ERROR_CODES.CAMERA_NOT_FOUND_OR_DISCONNECTED:
        return 'error';
      
      default:
        return 'error';
    }
  }

  private isRecoverableCode(code: number): boolean {
    switch (code) {
      case ERROR_CODES.STORAGE_SPACE_CRITICAL:
      case ERROR_CODES.CAMERA_NOT_FOUND_OR_DISCONNECTED:
      case ERROR_CODES.INVALID_PARAMS:
      case ERROR_CODES.AUTHENTICATION_REQUIRED:
      case ERROR_CODES.MEDIAMTX_SERVICE_UNAVAILABLE:
      case ERROR_CODES.INSUFFICIENT_STORAGE_SPACE:
        return false;
      
      default:
        return true;
    }
  }

  private logError(errorInfo: ErrorInfo): void {
    const log: ErrorLog = {
      timestamp: errorInfo.timestamp,
      error: errorInfo,
      stack: errorInfo.data?.stack,
      userAgent: navigator.userAgent,
      url: window.location.href
    };
    
    this.notifyLogCallbacks(log);
    
    // Also log to console for debugging
    console.error(`[${errorInfo.severity.toUpperCase()}] ${errorInfo.context}:`, errorInfo.message);
  }

  private notifyErrorCallbacks(error: ErrorInfo): void {
    this.errorCallbacks.forEach(callback => callback(error));
  }

  private notifyRecoveryCallbacks(result: RecoveryResult): void {
    this.recoveryCallbacks.forEach(callback => callback(result));
  }

  private notifyLogCallbacks(log: ErrorLog): void {
    this.logCallbacks.forEach(callback => callback(log));
  }

  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  /**
   * Cleanup
   */
  cleanup(): void {
    this.errorCallbacks.clear();
    this.recoveryCallbacks.clear();
    this.logCallbacks.clear();
  }
}

// Export singleton instance
export const errorHandlerService = new ErrorHandlerService();
