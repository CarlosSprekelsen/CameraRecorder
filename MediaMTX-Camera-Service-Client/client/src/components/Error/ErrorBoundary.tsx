/**
 * @fileoverview ErrorBoundary component for error handling and recovery
 * @author MediaMTX Development Team
 * @version 1.0.0
 */

import React, { ReactNode, ErrorInfo } from 'react';
import { ErrorBoundary as ReactErrorBoundary, FallbackProps } from 'react-error-boundary';
import { Button } from '../atoms/Button/Button';
import { Alert } from '../atoms/Alert/Alert';
import { Icon } from '../atoms/Icon/Icon';
import { logger } from '../../services/logger/LoggerService';
// ARCHITECTURE FIX: Logger is infrastructure - components can import it directly

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
  onError?: (error: Error, errorInfo: ErrorInfo) => void;
}

/**
 * ErrorFallback - Error boundary fallback UI component
 *
 * Provides a user-friendly error display with recovery options when React
 * components encounter unhandled errors. Implements error handling patterns
 * from architecture section 8.1 with logging and recovery mechanisms.
 *
 * @component
 * @param {FallbackProps} props - Error boundary fallback props
 * @param {Error} props.error - The error that caused the boundary to trigger
 * @param {() => void} props.resetErrorBoundary - Function to reset the error boundary
 * @returns {JSX.Element} The error fallback UI
 *
 * @features
 * - User-friendly error display
 * - Error logging and reporting
 * - Recovery options (reload, retry)
 * - Development error details
 * - Production-safe error messages
 *
 * @example
 * ```tsx
 * <ErrorBoundary>
 *   <App />
 * </ErrorBoundary>
 * ```
 *
 * @see {@link ../../docs/architecture/client-architechture.md} Client Architecture
 */
const ErrorFallback: React.FC<FallbackProps> = ({ error, resetErrorBoundary }) => {
  const handleReload = () => {
    window.location.reload();
  };

  return (
    <div className="flex flex-col items-center justify-center min-h-[400px] p-6 text-center">
      <Alert variant="error" title="Something went wrong" className="mb-6 max-w-2xl">
        <p className="text-sm mt-2">
          An unexpected error occurred. Please try refreshing the page or contact support if the
          problem persists.
        </p>
      </Alert>

      <div className="flex gap-4 mt-4">
        <Button
          variant="primary"
          onClick={resetErrorBoundary}
          className="flex items-center gap-2"
        >
          <Icon name="refresh" size={16} />
          Try Again
        </Button>
        <Button 
          variant="secondary" 
          onClick={handleReload}
          className="flex items-center gap-2"
        >
          <Icon name="bug" size={16} />
          Reload Page
        </Button>
      </div>

      {process.env.NODE_ENV === 'development' && error && (
        <div className="mt-6 p-4 bg-gray-100 rounded-lg max-w-4xl">
          <h6 className="text-lg font-semibold text-red-600 mb-2">
            Development Error Details:
          </h6>
          <pre className="text-xs font-mono whitespace-pre-wrap break-words">
            {error.toString()}
          </pre>
        </div>
      )}
    </div>
  );
};

/**
 * ErrorBoundary - Functional error boundary using react-error-boundary
 * Implements error handling from architecture section 8.1
 * Converts class component to functional component for coding standards compliance
 */
const ErrorBoundary: React.FC<Props> = ({ children, fallback, onError }) => {
  const handleError = (error: Error, errorInfo: ErrorInfo) => {
    // Log error using professional logger
    logger.error('ErrorBoundary caught an error', { error: error.message, stack: error.stack });

    // Call custom error handler if provided
    if (onError) {
      onError(error, errorInfo);
    }
  };

  return (
    <ReactErrorBoundary
      FallbackComponent={fallback ? () => <>{fallback}</> : ErrorFallback}
      onError={handleError}
      onReset={() => {
        // Reset logic handled by react-error-boundary
      }}
    >
      {children}
    </ReactErrorBoundary>
  );
};

export default ErrorBoundary;
