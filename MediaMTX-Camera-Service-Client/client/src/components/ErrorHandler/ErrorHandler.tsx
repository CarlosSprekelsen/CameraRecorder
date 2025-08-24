import React, { useEffect, useState } from 'react';
import { useErrorStore } from '../../stores/errorStore';
import './ErrorHandler.module.css';

/**
 * Error Handler Component
 * 
 * Enhanced error handling and recovery with user-friendly error messages,
 * error logging, and recovery mechanisms.
 */
const ErrorHandler: React.FC = () => {
  // Local state with prefixes
  const [localShowErrorModal, setLocalShowErrorModal] = useState(false);
  const [localShowErrorHistory, setLocalShowErrorHistory] = useState(false);
  const [localShowRecoveryResults, setLocalShowRecoveryResults] = useState(false);
  const [localShowErrorLogs, setLocalShowErrorLogs] = useState(false);
  const [localSelectedError, setLocalSelectedError] = useState<string | null>(null);

  // Store state with aliases
  const {
    currentError: storeCurrentError,
    errorHistory: storeErrorHistory,
    recoveryResults: storeRecoveryResults,
    errorLogs: storeErrorLogs,
    isRecovering: storeIsRecovering,
    isLogging: storeIsLogging,
    hasErrors: storeHasErrors,
    hasRecoverableErrors: storeHasRecoverableErrors,
    hasCriticalErrors: storeHasCriticalErrors,
    showErrorModal: storeShowErrorModal,
    showErrorHistory: storeShowErrorHistory,
    handleError: storeHandleError,
    handleJSONRPCError: storeHandleJSONRPCError,
    attemptRecovery: storeAttemptRecovery,
    clearCurrentError: storeClearCurrentError,
    dismissError: storeDismissError,
    getCurrentError: storeGetCurrentError,
    getErrorHistory: storeGetErrorHistory,
    getRecoveryResults: storeGetRecoveryResults,
    getErrorLogs: storeGetErrorLogs,
    hasCurrentError: storeHasCurrentError,
    hasErrorHistory: storeHasErrorHistory,
    hasRecoveryResults: storeHasRecoveryResults,
    hasErrorLogs: storeHasErrorLogs,
    getErrorsBySeverity: storeGetErrorsBySeverity,
    getErrorsByContext: storeGetErrorsByContext,
    getRecoverableErrors: storeGetRecoverableErrors,
    getCriticalErrors: storeGetCriticalErrors,
    initialize: storeInitialize,
    cleanup: storeCleanup
  } = useErrorStore();

  // Local handlers
  const handleLocalToggleErrorModal = () => {
    setLocalShowErrorModal(!localShowErrorModal);
  };

  const handleLocalToggleErrorHistory = () => {
    setLocalShowErrorHistory(!localShowErrorHistory);
  };

  const handleLocalToggleRecoveryResults = () => {
    setLocalShowRecoveryResults(!localShowRecoveryResults);
  };

  const handleLocalToggleErrorLogs = () => {
    setLocalShowErrorLogs(!localShowErrorLogs);
  };

  const handleLocalClearCurrentError = () => {
    storeClearCurrentError();
    setLocalShowErrorModal(false);
  };

  const handleLocalDismissError = (errorId: string) => {
    storeDismissError(errorId);
  };

  const handleLocalSelectError = (errorId: string) => {
    setLocalSelectedError(errorId);
  };

  // Format timestamp
  const formatTimestamp = (timestamp: Date): string => {
    return timestamp.toLocaleString();
  };

  // Get severity color class
  const getSeverityColorClass = (severity: string): string => {
    switch (severity) {
      case 'critical': return 'critical';
      case 'error': return 'error';
      case 'warning': return 'warning';
      case 'info': return 'info';
      default: return 'info';
    }
  };

  // Get severity icon
  const getSeverityIcon = (severity: string): string => {
    switch (severity) {
      case 'critical': return '🚨';
      case 'error': return '❌';
      case 'warning': return '⚠️';
      case 'info': return 'ℹ️';
      default: return 'ℹ️';
    }
  };

  // Initialize component
  useEffect(() => {
    storeInitialize();
  }, []);

  return (
    <div className="error-handler">
      <div className="error-handler-header">
        <h2>Error Handler</h2>
        <div className="error-handler-controls">
          <button
            onClick={handleLocalToggleErrorModal}
            className={`error-modal-toggle ${localShowErrorModal ? 'active' : ''}`}
          >
            Current Error {storeHasCurrentError() ? '(1)' : '(0)'}
          </button>
          <button
            onClick={handleLocalToggleErrorHistory}
            className={`history-toggle ${localShowErrorHistory ? 'active' : ''}`}
          >
            History ({storeErrorHistory.length})
          </button>
          <button
            onClick={handleLocalToggleRecoveryResults}
            className={`recovery-toggle ${localShowRecoveryResults ? 'active' : ''}`}
          >
            Recovery ({storeRecoveryResults.length})
          </button>
          <button
            onClick={handleLocalToggleErrorLogs}
            className={`logs-toggle ${localShowErrorLogs ? 'active' : ''}`}
          >
            Logs ({storeErrorLogs.length})
          </button>
        </div>
      </div>

      {/* Error Status Overview */}
      <div className="error-status-overview">
        <div className="status-indicators">
          <div className="status-item">
            <span className="status-label">Current Errors:</span>
            <span className={`status-value ${storeHasErrors ? 'has-errors' : 'no-errors'}`}>
              {storeHasErrors ? '❌ Yes' : '✅ No'}
            </span>
          </div>
          <div className="status-item">
            <span className="status-label">Recoverable Errors:</span>
            <span className={`status-value ${storeHasRecoverableErrors ? 'has-recoverable' : 'no-recoverable'}`}>
              {storeHasRecoverableErrors ? '⚠️ Yes' : '✅ No'}
            </span>
          </div>
          <div className="status-item">
            <span className="status-label">Critical Errors:</span>
            <span className={`status-value ${storeHasCriticalErrors ? 'has-critical' : 'no-critical'}`}>
              {storeHasCriticalErrors ? '🚨 Yes' : '✅ No'}
            </span>
          </div>
          <div className="status-item">
            <span className="status-label">Recovery Status:</span>
            <span className={`status-value ${storeIsRecovering ? 'recovering' : 'idle'}`}>
              {storeIsRecovering ? '🔄 Recovering...' : '✅ Idle'}
            </span>
          </div>
        </div>
      </div>

      {/* Current Error Modal */}
      {localShowErrorModal && storeCurrentError && (
        <div className="error-modal">
          <div className="modal-header">
            <h3>Current Error</h3>
            <button onClick={handleLocalClearCurrentError} className="close-button">
              ✕
            </button>
          </div>
          
          <div className={`error-content ${getSeverityColorClass(storeCurrentError.severity)}`}>
            <div className="error-header">
              <span className="error-icon">{getSeverityIcon(storeCurrentError.severity)}</span>
              <span className="error-severity">{storeCurrentError.severity.toUpperCase()}</span>
              <span className="error-timestamp">{formatTimestamp(storeCurrentError.timestamp)}</span>
            </div>
            
            <div className="error-message">
              <h4>Error Message</h4>
              <p className="message-text">{storeCurrentError.message}</p>
            </div>
            
            <div className="error-details">
              <h4>User-Friendly Message</h4>
              <p className="user-message">{storeCurrentError.userFriendly}</p>
            </div>
            
            <div className="error-context">
              <h4>Context</h4>
              <p className="context-text">{storeCurrentError.context}</p>
            </div>
            
            {storeCurrentError.code && (
              <div className="error-code">
                <h4>Error Code</h4>
                <p className="code-text">{storeCurrentError.code}</p>
              </div>
            )}
            
            <div className="error-recovery">
              <h4>Recovery Status</h4>
              <p className={`recovery-status ${storeCurrentError.recoverable ? 'recoverable' : 'not-recoverable'}`}>
                {storeCurrentError.recoverable ? '✅ Recoverable' : '❌ Not Recoverable'}
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Error History */}
      {localShowErrorHistory && (
        <div className="error-history">
          <h3>Error History</h3>
          {storeErrorHistory.length === 0 ? (
            <p className="no-errors">No errors in history</p>
          ) : (
            <div className="history-list">
              {storeErrorHistory.map((error, index) => (
                <div key={index} className={`history-item ${getSeverityColorClass(error.severity)}`}>
                  <div className="history-header">
                    <span className="history-icon">{getSeverityIcon(error.severity)}</span>
                    <span className="history-severity">{error.severity.toUpperCase()}</span>
                    <span className="history-timestamp">{formatTimestamp(error.timestamp)}</span>
                    <button
                      onClick={() => handleLocalDismissError(error.timestamp.getTime().toString())}
                      className="dismiss-button"
                    >
                      ✕
                    </button>
                  </div>
                  
                  <div className="history-content">
                    <div className="history-message">
                      <strong>Message:</strong> {error.message}
                    </div>
                    <div className="history-context">
                      <strong>Context:</strong> {error.context}
                    </div>
                    <div className="history-user-message">
                      <strong>User Message:</strong> {error.userFriendly}
                    </div>
                    {error.code && (
                      <div className="history-code">
                        <strong>Code:</strong> {error.code}
                      </div>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Recovery Results */}
      {localShowRecoveryResults && (
        <div className="recovery-results">
          <h3>Recovery Results</h3>
          {storeRecoveryResults.length === 0 ? (
            <p className="no-recoveries">No recovery attempts recorded</p>
          ) : (
            <div className="recovery-list">
              {storeRecoveryResults.map((result, index) => (
                <div key={index} className={`recovery-item ${result.success ? 'success' : 'failure'}`}>
                  <div className="recovery-header">
                    <span className="recovery-icon">
                      {result.success ? '✅' : '❌'}
                    </span>
                    <span className="recovery-status">
                      {result.success ? 'Success' : 'Failure'}
                    </span>
                    <span className="recovery-attempts">
                      Attempts: {result.attempts}
                    </span>
                    <span className="recovery-context">
                      Context: {result.context}
                    </span>
                  </div>
                  
                  {result.error && (
                    <div className="recovery-error">
                      <strong>Error:</strong> {result.error.message}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Error Logs */}
      {localShowErrorLogs && (
        <div className="error-logs">
          <h3>Error Logs</h3>
          {storeErrorLogs.length === 0 ? (
            <p className="no-logs">No error logs recorded</p>
          ) : (
            <div className="logs-list">
              {storeErrorLogs.map((log, index) => (
                <div key={index} className={`log-item ${getSeverityColorClass(log.error.severity)}`}>
                  <div className="log-header">
                    <span className="log-icon">{getSeverityIcon(log.error.severity)}</span>
                    <span className="log-timestamp">{formatTimestamp(log.timestamp)}</span>
                    <span className="log-url">{log.url}</span>
                  </div>
                  
                  <div className="log-content">
                    <div className="log-error">
                      <strong>Error:</strong> {log.error.message}
                    </div>
                    <div className="log-context">
                      <strong>Context:</strong> {log.error.context}
                    </div>
                    {log.stack && (
                      <div className="log-stack">
                        <strong>Stack Trace:</strong>
                        <pre className="stack-trace">{log.stack}</pre>
                      </div>
                    )}
                    <div className="log-user-agent">
                      <strong>User Agent:</strong> {log.userAgent}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Error Analysis */}
      <div className="error-analysis">
        <h3>Error Analysis</h3>
        <div className="analysis-grid">
          <div className="analysis-section">
            <h4>Errors by Severity</h4>
            <div className="severity-breakdown">
              <div className="severity-item">
                <span className="severity-label">Critical:</span>
                <span className="severity-count">{storeGetErrorsBySeverity('critical').length}</span>
              </div>
              <div className="severity-item">
                <span className="severity-label">Error:</span>
                <span className="severity-count">{storeGetErrorsBySeverity('error').length}</span>
              </div>
              <div className="severity-item">
                <span className="severity-label">Warning:</span>
                <span className="severity-count">{storeGetErrorsBySeverity('warning').length}</span>
              </div>
              <div className="severity-item">
                <span className="severity-label">Info:</span>
                <span className="severity-count">{storeGetErrorsBySeverity('info').length}</span>
              </div>
            </div>
          </div>
          
          <div className="analysis-section">
            <h4>Recovery Statistics</h4>
            <div className="recovery-stats">
              <div className="stat-item">
                <span className="stat-label">Recoverable Errors:</span>
                <span className="stat-count">{storeGetRecoverableErrors().length}</span>
              </div>
              <div className="stat-item">
                <span className="stat-label">Critical Errors:</span>
                <span className="stat-count">{storeGetCriticalErrors().length}</span>
              </div>
              <div className="stat-item">
                <span className="stat-label">Total Errors:</span>
                <span className="stat-count">{storeErrorHistory.length}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ErrorHandler;
