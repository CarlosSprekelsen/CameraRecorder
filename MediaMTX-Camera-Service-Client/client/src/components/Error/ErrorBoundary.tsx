/**
 * @fileoverview ErrorBoundary component for error handling and recovery
 * @author MediaMTX Development Team
 * @version 1.0.0
 */

import React, { ReactNode, ErrorInfo } from 'react';
import { ErrorBoundary as ReactErrorBoundary, FallbackProps } from 'react-error-boundary';
import { Box, Typography, Button, Alert, AlertTitle } from '@mui/material';
import { Refresh as RefreshIcon, BugReport as BugIcon } from '@mui/icons-material';
import { logger } from '../../services/logger/LoggerService';

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
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        minHeight: '400px',
        p: 3,
        textAlign: 'center',
      }}
    >
      <Alert severity="error" sx={{ mb: 3, maxWidth: 600 }}>
        <AlertTitle>Something went wrong</AlertTitle>
        <Typography variant="body2" sx={{ mt: 1 }}>
          An unexpected error occurred. Please try refreshing the page or contact support if the
          problem persists.
        </Typography>
      </Alert>

      <Box display="flex" gap={2} mt={2}>
        <Button
          variant="contained"
          startIcon={<RefreshIcon />}
          onClick={resetErrorBoundary}
          color="primary"
        >
          Try Again
        </Button>
        <Button variant="outlined" startIcon={<BugIcon />} onClick={handleReload} color="secondary">
          Reload Page
        </Button>
      </Box>

      {process.env.NODE_ENV === 'development' && error && (
        <Box sx={{ mt: 3, p: 2, bgcolor: 'grey.100', borderRadius: 1, maxWidth: 800 }}>
          <Typography variant="h6" color="error" gutterBottom>
            Development Error Details:
          </Typography>
          <Typography
            variant="body2"
            component="pre"
            sx={{ whiteSpace: 'pre-wrap', fontSize: '0.75rem' }}
          >
            {error.toString()}
          </Typography>
        </Box>
      )}
    </Box>
  );
};

/**
 * ErrorBoundary - Functional error boundary using react-error-boundary
 * Implements error handling from architecture section 8.1
 * Converts class component to functional component for coding standards compliance
 */
const ErrorBoundary: React.FC<Props> = ({ children, fallback, onError }) => {
  const handleError = (error: Error, errorInfo: ErrorInfo) => {
    // Log error to service
    logger.error('ErrorBoundary caught an error', error);

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
