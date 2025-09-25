/**
 * REQ-ERR01-001: Error handling must provide comprehensive error management and recovery
 * REQ-ERR01-002: Error logging must track and categorize errors appropriately
 * Coverage: UNIT
 * Quality: HIGH
 */
/**
 * Unit tests for error store
 * 
 * Design Principles:
 * - Pure unit testing with complete isolation
 * - Direct store testing without React context dependency
 * - Focus on error management and recovery logic
 * - Test error categorization and user-friendly messaging
 * - Validate error logging and history tracking
 */

import { useErrorStore } from '../../../src/stores/errorStore';
import type { JSONRPCError } from '../../../src/types/rpc';

// Mock the error handler service
jest.mock('../../../src/services/errorHandlerService', () => ({
  errorHandlerService: {
    handleError: jest.fn(),
    recoverFromError: jest.fn(),
    logError: jest.fn(),
    getUserFriendlyMessage: jest.fn(),
    categorizeError: jest.fn()
  }
}));

describe('Error Store', () => {
  let store: ReturnType<typeof useErrorStore.getState>;
  let mockErrorService: any;

  beforeEach(() => {
    // Reset store state completely
    const currentStore = useErrorStore.getState();
    currentStore.reset();
    
    // Get fresh store instance after reset
    store = useErrorStore.getState();
    
    // Get mock service
    mockErrorService = require('../../../src/services/errorHandlerService').errorHandlerService;
    jest.clearAllMocks();
  });

  describe('Initialization', () => {
    it('should start with correct default state', () => {
      const state = useErrorStore.getState();
      expect(state.currentError).toBeNull();
      expect(state.errorHistory).toEqual([]);
      expect(state.recoveryInProgress).toBe(false);
      expect(state.recoveryAttempts).toBe(0);
      expect(state.maxRecoveryAttempts).toBe(3);
      expect(state.isLogging).toBe(false);
      expect(state.errorCount).toBe(0);
      expect(state.criticalErrorCount).toBe(0);
      expect(state.lastErrorTime).toBeNull();
      expect(state.errorRate).toBe(0);
      expect(state.maxHistorySize).toBe(100);
    });
  });

  describe('Error Management', () => {
    it('should set current error', () => {
      const errorInfo = {
        message: 'Test error',
        code: 1001,
        context: 'test-context',
        timestamp: new Date(),
        userFriendly: 'A test error occurred',
        severity: 'error' as const,
        recoverable: true,
        data: { test: 'data' }
      };

      store.setCurrentError(errorInfo);
      
      const state = useErrorStore.getState();
      expect(state.currentError).toEqual(errorInfo);
    });

    it('should clear current error', () => {
      const errorInfo = {
        message: 'Test error',
        code: 1001,
        context: 'test-context',
        timestamp: new Date(),
        userFriendly: 'A test error occurred',
        severity: 'error' as const,
        recoverable: true
      };

      store.setCurrentError(errorInfo);
      store.clearCurrentError();
      
      const state = useErrorStore.getState();
      expect(state.currentError).toBeNull();
    });

    it('should add error to history', () => {
      const errorInfo = {
        message: 'Test error',
        code: 1001,
        context: 'test-context',
        timestamp: new Date(),
        userFriendly: 'A test error occurred',
        severity: 'error' as const,
        recoverable: true
      };

      store.addToHistory(errorInfo);
      
      const state = useErrorStore.getState();
      expect(state.errorHistory).toHaveLength(1);
      expect(state.errorHistory[0]).toEqual(errorInfo);
    });

    it('should limit error history size', () => {
      store.setMaxHistorySize(3);
      
      // Add more errors than the limit
      for (let i = 0; i < 5; i++) {
        store.addToHistory({
          message: `Error ${i}`,
          context: 'test-context',
          timestamp: new Date(),
          userFriendly: `Error ${i} occurred`,
          severity: 'error' as const,
          recoverable: true
        });
      }
      
      const state = useErrorStore.getState();
      expect(state.errorHistory).toHaveLength(3);
      // Should keep the most recent ones
      expect(state.errorHistory[0].message).toBe('Error 4');
      expect(state.errorHistory[1].message).toBe('Error 3');
      expect(state.errorHistory[2].message).toBe('Error 2');
    });

    it('should clear error history', () => {
      store.addToHistory({
        message: 'Test error',
        context: 'test-context',
        timestamp: new Date(),
        userFriendly: 'A test error occurred',
        severity: 'error' as const,
        recoverable: true
      });
      
      store.clearHistory();
      
      const state = useErrorStore.getState();
      expect(state.errorHistory).toEqual([]);
    });
  });

  describe('Error Recovery', () => {
    it('should set recovery in progress', () => {
      store.setRecoveryInProgress(true);
      let state = useErrorStore.getState();
      expect(state.recoveryInProgress).toBe(true);

      store.setRecoveryInProgress(false);
      state = useErrorStore.getState();
      expect(state.recoveryInProgress).toBe(false);
    });

    it('should increment recovery attempts', () => {
      store.incrementRecoveryAttempts();
      let state = useErrorStore.getState();
      expect(state.recoveryAttempts).toBe(1);

      store.incrementRecoveryAttempts();
      state = useErrorStore.getState();
      expect(state.recoveryAttempts).toBe(2);
    });

    it('should reset recovery attempts', () => {
      store.incrementRecoveryAttempts();
      store.incrementRecoveryAttempts();
      store.resetRecoveryAttempts();
      
      const state = useErrorStore.getState();
      expect(state.recoveryAttempts).toBe(0);
    });

    it('should set max recovery attempts', () => {
      store.setMaxRecoveryAttempts(5);
      
      const state = useErrorStore.getState();
      expect(state.maxRecoveryAttempts).toBe(5);
    });

    it('should check if recovery is possible', () => {
      store.setMaxRecoveryAttempts(3);
      
      // Test with attempts under limit
      store.setRecoveryAttempts(2);
      expect(store.canRecover()).toBe(true);
      
      // Test with max attempts reached
      store.setRecoveryAttempts(3);
      expect(store.canRecover()).toBe(false);
    });
  });

  describe('Error Statistics', () => {
    it('should increment error count', () => {
      store.incrementErrorCount();
      let state = useErrorStore.getState();
      expect(state.errorCount).toBe(1);

      store.incrementErrorCount();
      state = useErrorStore.getState();
      expect(state.errorCount).toBe(2);
    });

    it('should increment critical error count', () => {
      store.incrementCriticalErrorCount();
      let state = useErrorStore.getState();
      expect(state.criticalErrorCount).toBe(1);

      store.incrementCriticalErrorCount();
      state = useErrorStore.getState();
      expect(state.criticalErrorCount).toBe(2);
    });

    it('should update last error time', () => {
      const now = new Date();
      store.updateLastErrorTime(now);
      
      const state = useErrorStore.getState();
      expect(state.lastErrorTime).toEqual(now);
    });

    it('should set error rate', () => {
      store.setErrorRate(0.05);
      
      const state = useErrorStore.getState();
      expect(state.errorRate).toBe(0.05);
    });

    it('should reset error statistics', () => {
      store.incrementErrorCount();
      store.incrementCriticalErrorCount();
      store.updateLastErrorTime(new Date());
      store.setErrorRate(0.1);
      
      store.resetErrorStatistics();
      
      const state = useErrorStore.getState();
      expect(state.errorCount).toBe(0);
      expect(state.criticalErrorCount).toBe(0);
      expect(state.lastErrorTime).toBeNull();
      expect(state.errorRate).toBe(0);
    });
  });

  describe('Error Operations', () => {
    it('should handle error successfully', async () => {
      const error = new Error('Test error');
      const context = 'test-context';
      const mockErrorInfo = {
        message: 'Test error',
        context: 'test-context',
        timestamp: new Date(),
        userFriendly: 'A test error occurred',
        severity: 'error' as const,
        recoverable: true
      };

      mockErrorService.handleError.mockResolvedValue(mockErrorInfo);
      mockErrorService.getUserFriendlyMessage.mockReturnValue('A test error occurred');
      mockErrorService.categorizeError.mockReturnValue('error');

      await store.handleError(error, context);

      const state = useErrorStore.getState();
      expect(state.currentError).toEqual(mockErrorInfo);
      expect(state.errorHistory).toContain(mockErrorInfo);
      expect(state.errorCount).toBe(1);
      expect(state.lastErrorTime).toBeInstanceOf(Date);
    });

    it('should handle JSON-RPC error', async () => {
      const rpcError: JSONRPCError = {
        code: -32001,
        message: 'Camera not found',
        data: { device: 'camera0' }
      };
      const context = 'camera-operation';
      const mockErrorInfo = {
        message: 'Camera not found',
        code: -32001,
        context: 'camera-operation',
        timestamp: new Date(),
        userFriendly: 'Camera device not found',
        severity: 'error' as const,
        recoverable: true,
        data: { device: 'camera0' }
      };

      mockErrorService.handleError.mockResolvedValue(mockErrorInfo);
      mockErrorService.getUserFriendlyMessage.mockReturnValue('Camera device not found');
      mockErrorService.categorizeError.mockReturnValue('error');

      await store.handleRPCError(rpcError, context);

      const state = useErrorStore.getState();
      expect(state.currentError).toEqual(mockErrorInfo);
      expect(state.errorHistory).toContain(mockErrorInfo);
    });

    it('should attempt recovery successfully', async () => {
      const errorInfo = {
        message: 'Connection lost',
        context: 'websocket',
        timestamp: new Date(),
        userFriendly: 'Connection to server lost',
        severity: 'error' as const,
        recoverable: true
      };

      const mockRecoveryResult = {
        success: true,
        attempts: 1,
        context: 'websocket'
      };

      mockErrorService.recoverFromError.mockResolvedValue(mockRecoveryResult);

      const result = await store.attemptRecovery(errorInfo);

      expect(result).toEqual(mockRecoveryResult);
      const state = useErrorStore.getState();
      expect(state.recoveryInProgress).toBe(false);
      expect(state.recoveryAttempts).toBe(1);
    });

    it('should handle recovery failure', async () => {
      const errorInfo = {
        message: 'Connection lost',
        context: 'websocket',
        timestamp: new Date(),
        userFriendly: 'Connection to server lost',
        severity: 'error' as const,
        recoverable: true
      };

      const mockRecoveryResult = {
        success: false,
        attempts: 3,
        error: errorInfo,
        context: 'websocket'
      };

      mockErrorService.recoverFromError.mockResolvedValue(mockRecoveryResult);

      const result = await store.attemptRecovery(errorInfo);

      expect(result).toEqual(mockRecoveryResult);
      const state = useErrorStore.getState();
      expect(state.recoveryInProgress).toBe(false);
      expect(state.recoveryAttempts).toBe(3);
    });

    it('should log error successfully', async () => {
      const errorInfo = {
        message: 'Test error',
        context: 'test-context',
        timestamp: new Date(),
        userFriendly: 'A test error occurred',
        severity: 'error' as const,
        recoverable: true
      };

      mockErrorService.logError.mockResolvedValue(undefined);

      await store.logError(errorInfo);

      const state = useErrorStore.getState();
      expect(state.isLogging).toBe(false);
      expect(mockErrorService.logError).toHaveBeenCalledWith(errorInfo);
    });
  });

  describe('Error Analysis', () => {
    it('should get error statistics', () => {
      store.incrementErrorCount();
      store.incrementErrorCount();
      store.incrementCriticalErrorCount();
      store.setErrorRate(0.05);
      store.updateLastErrorTime(new Date());

      const stats = store.getErrorStatistics();
      expect(stats).toEqual({
        total_errors: 2,
        critical_errors: 1,
        error_rate: 0.05,
        last_error_time: expect.any(Date),
        history_size: 0
      });
    });

    it('should get errors by severity', () => {
      store.addToHistory({
        message: 'Info message',
        context: 'test',
        timestamp: new Date(),
        userFriendly: 'Info',
        severity: 'info',
        recoverable: true
      });
      store.addToHistory({
        message: 'Warning message',
        context: 'test',
        timestamp: new Date(),
        userFriendly: 'Warning',
        severity: 'warning',
        recoverable: true
      });
      store.addToHistory({
        message: 'Error message',
        context: 'test',
        timestamp: new Date(),
        userFriendly: 'Error',
        severity: 'error',
        recoverable: true
      });

      const errorsBySeverity = store.getErrorsBySeverity();
      expect(errorsBySeverity.info).toHaveLength(1);
      expect(errorsBySeverity.warning).toHaveLength(1);
      expect(errorsBySeverity.error).toHaveLength(1);
      expect(errorsBySeverity.critical).toHaveLength(0);
    });

    it('should get recent errors', () => {
      const now = new Date();
      const recentError = {
        message: 'Recent error',
        context: 'test',
        timestamp: new Date(now.getTime() - 1000),
        userFriendly: 'Recent error',
        severity: 'error' as const,
        recoverable: true
      };
      const oldError = {
        message: 'Old error',
        context: 'test',
        timestamp: new Date(now.getTime() - 3600000),
        userFriendly: 'Old error',
        severity: 'error' as const,
        recoverable: true
      };

      store.addToHistory(recentError);
      store.addToHistory(oldError);

      const recentErrors = store.getRecentErrors(3000000); // 50 minutes
      expect(recentErrors).toHaveLength(1);
      expect(recentErrors[0].message).toBe('Recent error');
    });

    it('should check if has critical errors', () => {
      expect(store.hasCriticalErrors()).toBe(false);

      store.addToHistory({
        message: 'Critical error',
        context: 'test',
        timestamp: new Date(),
        userFriendly: 'Critical error',
        severity: 'critical',
        recoverable: false
      });

      expect(store.hasCriticalErrors()).toBe(true);
    });

    it('should get error trend', () => {
      const now = new Date();
      // Add errors over time
      for (let i = 0; i < 5; i++) {
        store.addToHistory({
          message: `Error ${i}`,
          context: 'test',
          timestamp: new Date(now.getTime() - (i * 60000)), // 1 minute apart
          userFriendly: `Error ${i}`,
          severity: 'error' as const,
          recoverable: true
        });
      }

      const trend = store.getErrorTrend();
      expect(trend).toBe('increasing'); // More recent errors
    });
  });

  describe('Configuration', () => {
    it('should set max history size', () => {
      store.setMaxHistorySize(50);
      
      const state = useErrorStore.getState();
      expect(state.maxHistorySize).toBe(50);
    });
  });

  describe('State Reset', () => {
    it('should reset all state to initial values', () => {
      // Set some state
      store.setCurrentError({
        message: 'Test error',
        context: 'test',
        timestamp: new Date(),
        userFriendly: 'Test error',
        severity: 'error',
        recoverable: true
      });
      store.addToHistory({
        message: 'History error',
        context: 'test',
        timestamp: new Date(),
        userFriendly: 'History error',
        severity: 'error',
        recoverable: true
      });
      store.setRecoveryInProgress(true);
      store.incrementErrorCount();
      
      // Reset
      store.reset();
      
      const state = useErrorStore.getState();
      expect(state.currentError).toBeNull();
      expect(state.errorHistory).toEqual([]);
      expect(state.recoveryInProgress).toBe(false);
      expect(state.recoveryAttempts).toBe(0);
      expect(state.isLogging).toBe(false);
      expect(state.errorCount).toBe(0);
      expect(state.criticalErrorCount).toBe(0);
      expect(state.lastErrorTime).toBeNull();
      expect(state.errorRate).toBe(0);
    });
  });
});
