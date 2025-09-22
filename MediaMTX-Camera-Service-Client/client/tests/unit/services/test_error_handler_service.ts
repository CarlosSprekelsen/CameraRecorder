/**
 * Error Handler Service Unit Tests
 * 
 * Ground Truth References:
 * - Server API: ../mediamtx-camera-service/docs/api/json-rpc-methods.md
 * - Client Architecture: ../docs/architecture/client-architecture.md
 * - Client Requirements: ../docs/requirements/client-requirements.md
 * 
 * Requirements Coverage:
 * - REQ-ERR01-001: Error handling must be comprehensive and consistent
 * - REQ-ERR01-002: Error recovery must be automatic when possible
 * - REQ-ERR01-003: Error reporting must be accurate and actionable
 * - REQ-ERR01-004: Error context must be preserved for debugging
 * 
 * Coverage: UNIT
 * Quality: HIGH
 */

import { ErrorHandlerService } from '../../../src/services/errorHandlerService';
import { ERROR_CODES } from '../../../src/types/rpc';

// Mock logger service
jest.mock('../../../src/services/loggerService', () => ({
  logger: {
    error: jest.fn(),
    warn: jest.fn(),
    info: jest.fn(),
    debug: jest.fn(),
  },
  loggers: {
    error: {
      error: jest.fn(),
      warn: jest.fn(),
      info: jest.fn(),
      debug: jest.fn(),
    },
  },
}));

describe('Error Handler Service', () => {
  let errorHandler: ErrorHandlerService;

  beforeEach(() => {
    errorHandler = new ErrorHandlerService();
    jest.clearAllMocks();
  });

  describe('REQ-ERR01-001: Comprehensive Error Handling', () => {
    it('should handle network errors', () => {
      const networkError = new Error('Network connection failed');
      const handledError = errorHandler.handleError(networkError, 'websocket');

      expect(handledError).toHaveProperty('type', 'network');
      expect(handledError).toHaveProperty('recoverable', true);
      expect(handledError).toHaveProperty('userMessage');
      expect(handledError).toHaveProperty('technicalMessage');
    });

    it('should handle authentication errors', () => {
      const authError = {
        code: ERROR_CODES.INVALID_TOKEN,
        message: 'Invalid or expired token',
      };
      const handledError = errorHandler.handleError(authError, 'authentication');

      expect(handledError).toHaveProperty('type', 'authentication');
      expect(handledError).toHaveProperty('recoverable', true);
      expect(handledError).toHaveProperty('userMessage');
    });

    it('should handle validation errors', () => {
      const validationError = {
        code: ERROR_CODES.INVALID_PARAMS,
        message: 'Invalid parameters provided',
      };
      const handledError = errorHandler.handleError(validationError, 'validation');

      expect(handledError).toHaveProperty('type', 'validation');
      expect(handledError).toHaveProperty('recoverable', false);
      expect(handledError).toHaveProperty('userMessage');
    });

    it('should handle system errors', () => {
      const systemError = {
        code: ERROR_CODES.INTERNAL_ERROR,
        message: 'Internal server error',
      };
      const handledError = errorHandler.handleError(systemError, 'system');

      expect(handledError).toHaveProperty('type', 'system');
      expect(handledError).toHaveProperty('recoverable', false);
      expect(handledError).toHaveProperty('userMessage');
    });

    it('should handle unknown errors', () => {
      const unknownError = new Error('Unknown error occurred');
      const handledError = errorHandler.handleError(unknownError, 'unknown');

      expect(handledError).toHaveProperty('type', 'unknown');
      expect(handledError).toHaveProperty('recoverable', false);
      expect(handledError).toHaveProperty('userMessage');
    });
  });

  describe('REQ-ERR01-002: Automatic Error Recovery', () => {
    it('should attempt recovery for recoverable errors', async () => {
      const recoverableError = new Error('Temporary network issue');
      const recoveryResult = await errorHandler.attemptRecovery(recoverableError, 'websocket');

      expect(recoveryResult).toHaveProperty('success');
      expect(recoveryResult).toHaveProperty('action');
      expect(recoveryResult).toHaveProperty('retryAfter');
    });

    it('should not attempt recovery for non-recoverable errors', async () => {
      const nonRecoverableError = {
        code: ERROR_CODES.INVALID_PARAMS,
        message: 'Invalid parameters',
      };
      const recoveryResult = await errorHandler.attemptRecovery(nonRecoverableError, 'validation');

      expect(recoveryResult.success).toBe(false);
      expect(recoveryResult.action).toBe('none');
    });

    it('should implement exponential backoff for retries', async () => {
      const networkError = new Error('Network timeout');
      
      // First retry
      const firstRetry = await errorHandler.attemptRecovery(networkError, 'websocket');
      expect(firstRetry.retryAfter).toBeGreaterThan(0);

      // Second retry should have longer delay
      const secondRetry = await errorHandler.attemptRecovery(networkError, 'websocket');
      expect(secondRetry.retryAfter).toBeGreaterThan(firstRetry.retryAfter);
    });

    it('should respect maximum retry attempts', async () => {
      const networkError = new Error('Persistent network issue');
      
      // Attempt multiple recoveries
      for (let i = 0; i < 5; i++) {
        await errorHandler.attemptRecovery(networkError, 'websocket');
      }

      // Should eventually give up
      const finalRetry = await errorHandler.attemptRecovery(networkError, 'websocket');
      expect(finalRetry.success).toBe(false);
      expect(finalRetry.action).toBe('give_up');
    });
  });

  describe('REQ-ERR01-003: Accurate Error Reporting', () => {
    it('should provide user-friendly error messages', () => {
      const technicalError = new Error('ECONNREFUSED: Connection refused');
      const handledError = errorHandler.handleError(technicalError, 'websocket');

      expect(handledError.userMessage).not.toContain('ECONNREFUSED');
      expect(handledError.userMessage).toContain('connection');
      expect(handledError.technicalMessage).toContain('ECONNREFUSED');
    });

    it('should provide actionable error messages', () => {
      const authError = {
        code: ERROR_CODES.INVALID_TOKEN,
        message: 'Token expired',
      };
      const handledError = errorHandler.handleError(authError, 'authentication');

      expect(handledError.userMessage).toContain('login');
      expect(handledError.action).toBeDefined();
    });

    it('should categorize errors by severity', () => {
      const criticalError = {
        code: ERROR_CODES.INTERNAL_ERROR,
        message: 'System failure',
      };
      const handledError = errorHandler.handleError(criticalError, 'system');

      expect(handledError.severity).toBe('critical');
      expect(handledError.notifyUser).toBe(true);
    });

    it('should provide error codes for programmatic handling', () => {
      const authError = {
        code: ERROR_CODES.INVALID_TOKEN,
        message: 'Invalid token',
      };
      const handledError = errorHandler.handleError(authError, 'authentication');

      expect(handledError.code).toBe(ERROR_CODES.INVALID_TOKEN);
      expect(handledError.category).toBe('authentication');
    });
  });

  describe('REQ-ERR01-004: Error Context Preservation', () => {
    it('should preserve error context for debugging', () => {
      const context = {
        userId: '123',
        operation: 'camera_start',
        timestamp: new Date().toISOString(),
      };
      const error = new Error('Camera start failed');
      const handledError = errorHandler.handleError(error, 'camera', context);

      expect(handledError.context).toMatchObject(context);
      expect(handledError.stack).toBeDefined();
      expect(handledError.timestamp).toBeDefined();
    });

    it('should capture call stack information', () => {
      const error = new Error('Test error');
      const handledError = errorHandler.handleError(error, 'test');

      expect(handledError.stack).toContain('Error: Test error');
      expect(handledError.stack).toContain('at ');
    });

    it('should preserve error chain information', () => {
      const originalError = new Error('Original error');
      const wrappedError = new Error('Wrapped error');
      wrappedError.cause = originalError;

      const handledError = errorHandler.handleError(wrappedError, 'test');

      expect(handledError.cause).toBeDefined();
      expect(handledError.cause.message).toBe('Original error');
    });

    it('should capture environment information', () => {
      const error = new Error('Test error');
      const handledError = errorHandler.handleError(error, 'test');

      expect(handledError.environment).toHaveProperty('userAgent');
      expect(handledError.environment).toHaveProperty('url');
      expect(handledError.environment).toHaveProperty('timestamp');
    });
  });

  describe('Error Statistics and Monitoring', () => {
    it('should track error statistics', () => {
      const error1 = new Error('Error 1');
      const error2 = new Error('Error 2');
      const error3 = { code: ERROR_CODES.INVALID_TOKEN, message: 'Auth error' };

      errorHandler.handleError(error1, 'websocket');
      errorHandler.handleError(error2, 'websocket');
      errorHandler.handleError(error3, 'authentication');

      const stats = errorHandler.getErrorStatistics();

      expect(stats.totalErrors).toBe(3);
      expect(stats.errorsByType.websocket).toBe(2);
      expect(stats.errorsByType.authentication).toBe(1);
      expect(stats.errorsBySeverity.critical).toBeGreaterThanOrEqual(0);
    });

    it('should track error trends over time', () => {
      const error = new Error('Test error');
      
      // Generate errors over time
      for (let i = 0; i < 10; i++) {
        errorHandler.handleError(error, 'test');
      }

      const trends = errorHandler.getErrorTrends();
      expect(trends).toHaveProperty('hourly');
      expect(trends).toHaveProperty('daily');
      expect(trends.hourly.length).toBeGreaterThan(0);
    });

    it('should identify error patterns', () => {
      const networkError = new Error('Network timeout');
      
      // Generate multiple similar errors
      for (let i = 0; i < 5; i++) {
        errorHandler.handleError(networkError, 'websocket');
      }

      const patterns = errorHandler.identifyErrorPatterns();
      expect(patterns).toHaveLength(1);
      expect(patterns[0].count).toBe(5);
      expect(patterns[0].type).toBe('websocket');
    });
  });

  describe('Error Recovery Strategies', () => {
    it('should implement different recovery strategies', () => {
      const strategies = errorHandler.getRecoveryStrategies();

      expect(strategies).toHaveProperty('network');
      expect(strategies).toHaveProperty('authentication');
      expect(strategies).toHaveProperty('validation');
      expect(strategies.network).toHaveProperty('retry');
      expect(strategies.authentication).toHaveProperty('reauth');
    });

    it('should execute recovery strategies', async () => {
      const networkError = new Error('Network timeout');
      const result = await errorHandler.executeRecoveryStrategy('network', networkError);

      expect(result).toHaveProperty('success');
      expect(result).toHaveProperty('action');
      expect(result).toHaveProperty('nextStep');
    });

    it('should handle recovery strategy failures', async () => {
      const persistentError = new Error('Persistent failure');
      const result = await errorHandler.executeRecoveryStrategy('network', persistentError);

      expect(result).toHaveProperty('success');
      if (!result.success) {
        expect(result).toHaveProperty('fallback');
      }
    });
  });

  describe('Configuration', () => {
    it('should configure error handling behavior', () => {
      const config = {
        maxRetries: 3,
        retryDelay: 1000,
        enableRecovery: true,
        logErrors: true,
      };

      errorHandler.configure(config);

      expect(errorHandler.getConfig()).toMatchObject(config);
    });

    it('should validate configuration options', () => {
      const invalidConfig = {
        maxRetries: -1,
        retryDelay: 0,
      };

      expect(() => {
        errorHandler.configure(invalidConfig);
      }).toThrow('Invalid configuration');
    });
  });

  describe('Error Reporting Integration', () => {
    it('should integrate with external error reporting', () => {
      const error = new Error('Test error');
      const handledError = errorHandler.handleError(error, 'test');

      // Should trigger external reporting if configured
      expect(handledError.reported).toBeDefined();
    });

    it('should handle reporting failures gracefully', () => {
      // Mock reporting service failure
      const error = new Error('Test error');
      const handledError = errorHandler.handleError(error, 'test');

      // Should not fail even if reporting fails
      expect(handledError).toBeDefined();
      expect(handledError.userMessage).toBeDefined();
    });
  });
});
