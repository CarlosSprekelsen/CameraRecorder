import { create } from 'zustand';
import { errorHandlerService } from '../services/errorHandlerService';
import type { JSONRPCError } from '../types/rpc';

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
 * Error Store State Interface
 */
interface ErrorStoreState {
  // Current errors
  currentError: ErrorInfo | null;
  
  // Error history
  errorHistory: ErrorInfo[];
  
  // Recovery state
  recoveryResults: RecoveryResult[];
  
  // Error logs
  errorLogs: ErrorLog[];
  
  // Loading states
  isRecovering: boolean;
  isLogging: boolean;
  
  // Error states
  hasErrors: boolean;
  hasRecoverableErrors: boolean;
  hasCriticalErrors: boolean;
  
  // Display states
  showErrorModal: boolean;
  showErrorHistory: boolean;
}

/**
 * Error Store Actions Interface
 */
interface ErrorStoreActions {
  // State management
  setCurrentError: (error: ErrorInfo | null) => void;
  addErrorToHistory: (error: ErrorInfo) => void;
  clearErrorHistory: () => void;
  addRecoveryResult: (result: RecoveryResult) => void;
  clearRecoveryResults: () => void;
  addErrorLog: (log: ErrorLog) => void;
  clearErrorLogs: () => void;
  
  // Loading states
  setRecovering: (recovering: boolean) => void;
  setLogging: (logging: boolean) => void;
  
  // Error states
  setHasErrors: (hasErrors: boolean) => void;
  setHasRecoverableErrors: (hasRecoverable: boolean) => void;
  setHasCriticalErrors: (hasCritical: boolean) => void;
  
  // Display states
  setShowErrorModal: (show: boolean) => void;
  setShowErrorHistory: (show: boolean) => void;
  
  // Error operations
  handleError: (error: Error, context: string) => ErrorInfo;
  handleJSONRPCError: (error: JSONRPCError, context: string) => ErrorInfo;
  attemptRecovery: <T>(operation: () => Promise<T>, context: string, maxAttempts?: number) => Promise<RecoveryResult>;
  clearCurrentError: () => void;
  dismissError: (errorId: string) => void;
  
  // State queries
  getCurrentError: () => ErrorInfo | null;
  getErrorHistory: () => ErrorInfo[];
  getRecoveryResults: () => RecoveryResult[];
  getErrorLogs: () => ErrorLog[];
  hasCurrentError: () => boolean;
  hasErrorHistory: () => boolean;
  hasRecoveryResults: () => boolean;
  hasErrorLogs: () => boolean;
  
  // Error analysis
  getErrorsBySeverity: (severity: 'info' | 'warning' | 'error' | 'critical') => ErrorInfo[];
  getErrorsByContext: (context: string) => ErrorInfo[];
  getRecoverableErrors: () => ErrorInfo[];
  getCriticalErrors: () => ErrorInfo[];
  
  // Service integration
  initialize: () => void;
  cleanup: () => void;
}

/**
 * Error Store Type
 */
type ErrorStore = ErrorStoreState & ErrorStoreActions;

/**
 * Error Store Implementation
 */
export const useErrorStore = create<ErrorStore>((set, get) => ({
  // Initial state
  currentError: null,
  errorHistory: [],
  recoveryResults: [],
  errorLogs: [],
  isRecovering: false,
  isLogging: false,
  hasErrors: false,
  hasRecoverableErrors: false,
  hasCriticalErrors: false,
  showErrorModal: false,
  showErrorHistory: false,

  // State management actions
  setCurrentError: (error: ErrorInfo | null) => {
    set({ 
      currentError: error,
      hasErrors: error !== null,
      showErrorModal: error !== null
    });
  },

  addErrorToHistory: (error: ErrorInfo) => {
    set((storeState) => ({
      errorHistory: [error, ...storeState.errorHistory.slice(0, 99)] // Keep last 100 errors
    }));
  },

  clearErrorHistory: () => {
    set({ errorHistory: [] });
  },

  addRecoveryResult: (result: RecoveryResult) => {
    set((storeState) => ({
      recoveryResults: [result, ...storeState.recoveryResults.slice(0, 49)] // Keep last 50 results
    }));
  },

  clearRecoveryResults: () => {
    set({ recoveryResults: [] });
  },

  addErrorLog: (log: ErrorLog) => {
    set((storeState) => ({
      errorLogs: [log, ...storeState.errorLogs.slice(0, 199)] // Keep last 200 logs
    }));
  },

  clearErrorLogs: () => {
    set({ errorLogs: [] });
  },

  // Loading state actions
  setRecovering: (recovering: boolean) => {
    set({ isRecovering: recovering });
  },

  setLogging: (logging: boolean) => {
    set({ isLogging: logging });
  },

  // Error state actions
  setHasErrors: (hasErrors: boolean) => {
    set({ hasErrors });
  },

  setHasRecoverableErrors: (hasRecoverable: boolean) => {
    set({ hasRecoverableErrors: hasRecoverable });
  },

  setHasCriticalErrors: (hasCritical: boolean) => {
    set({ hasCriticalErrors: hasCritical });
  },

  // Display state actions
  setShowErrorModal: (show: boolean) => {
    set({ showErrorModal: show });
  },

  setShowErrorHistory: (show: boolean) => {
    set({ showErrorHistory: show });
  },

  // Error operations
  handleError: (error: Error, context: string) => {
    const { setLogging } = get();
    setLogging(true);
    
    try {
      const errorInfo = errorHandlerService.handleError(error, context);
      
      const { setCurrentError, addErrorToHistory, addErrorLog, setHasRecoverableErrors, setHasCriticalErrors } = get();
      
      setCurrentError(errorInfo);
      addErrorToHistory(errorInfo);
      addErrorLog({
        timestamp: errorInfo.timestamp,
        error: errorInfo,
        stack: errorInfo.data?.stack,
        userAgent: navigator.userAgent,
        url: window.location.href
      });
      
      setHasRecoverableErrors(errorInfo.recoverable);
      setHasCriticalErrors(errorInfo.severity === 'critical');
      
      return errorInfo;
    } finally {
      setLogging(false);
    }
  },

  handleJSONRPCError: (error: JSONRPCError, context: string) => {
    const { setLogging } = get();
    setLogging(true);
    
    try {
      const errorInfo = errorHandlerService.handleJSONRPCError(error, context);
      
      const { setCurrentError, addErrorToHistory, addErrorLog, setHasRecoverableErrors, setHasCriticalErrors } = get();
      
      setCurrentError(errorInfo);
      addErrorToHistory(errorInfo);
      addErrorLog({
        timestamp: errorInfo.timestamp,
        error: errorInfo,
        userAgent: navigator.userAgent,
        url: window.location.href
      });
      
      setHasRecoverableErrors(errorInfo.recoverable);
      setHasCriticalErrors(errorInfo.severity === 'critical');
      
      return errorInfo;
    } finally {
      setLogging(false);
    }
  },

  attemptRecovery: async <T>(operation: () => Promise<T>, context: string, maxAttempts: number = 3) => {
    const { setRecovering } = get();
    setRecovering(true);
    
    try {
      const result = await errorHandlerService.attemptRecovery(operation, context, maxAttempts);
      
      const { addRecoveryResult } = get();
      addRecoveryResult(result);
      
      return result;
    } finally {
      setRecovering(false);
    }
  },

  clearCurrentError: () => {
    const { setCurrentError, setShowErrorModal } = get();
    setCurrentError(null);
    setShowErrorModal(false);
  },

  dismissError: (errorId: string) => {
    const { errorHistory } = get();
    const updatedHistory = errorHistory.filter(error => 
      error.timestamp.getTime().toString() !== errorId
    );
    set({ errorHistory: updatedHistory });
  },

  // State queries
  getCurrentError: () => {
    return get().currentError;
  },

  getErrorHistory: () => {
    return get().errorHistory;
  },

  getRecoveryResults: () => {
    return get().recoveryResults;
  },

  getErrorLogs: () => {
    return get().errorLogs;
  },

  hasCurrentError: () => {
    return get().currentError !== null;
  },

  hasErrorHistory: () => {
    return get().errorHistory.length > 0;
  },

  hasRecoveryResults: () => {
    return get().recoveryResults.length > 0;
  },

  hasErrorLogs: () => {
    return get().errorLogs.length > 0;
  },

  // Error analysis
  getErrorsBySeverity: (severity: 'info' | 'warning' | 'error' | 'critical') => {
    return get().errorHistory.filter(error => error.severity === severity);
  },

  getErrorsByContext: (context: string) => {
    return get().errorHistory.filter(error => error.context === context);
  },

  getRecoverableErrors: () => {
    return get().errorHistory.filter(error => error.recoverable);
  },

  getCriticalErrors: () => {
    return get().errorHistory.filter(error => error.severity === 'critical');
  },

  // Service integration
  initialize: () => {
    // Set up event handlers
    errorHandlerService.onError((error) => {
      const { setCurrentError, addErrorToHistory, addErrorLog } = get();
      setCurrentError(error);
      addErrorToHistory(error);
      addErrorLog({
        timestamp: error.timestamp,
        error,
        userAgent: navigator.userAgent,
        url: window.location.href
      });
    });

    errorHandlerService.onRecovery((result) => {
      get().addRecoveryResult(result);
    });

    errorHandlerService.onErrorLog((log) => {
      get().addErrorLog(log);
    });
  },

  cleanup: () => {
    errorHandlerService.cleanup();
    set({
      currentError: null,
      errorHistory: [],
      recoveryResults: [],
      errorLogs: [],
      isRecovering: false,
      isLogging: false,
      hasErrors: false,
      hasRecoverableErrors: false,
      hasCriticalErrors: false,
      showErrorModal: false,
      showErrorHistory: false
    });
  }
}));
