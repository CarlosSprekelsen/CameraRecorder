import { Component, ErrorInfo, ReactNode } from 'react';
import { Box, Typography, Button, Alert, AlertTitle } from '@mui/material';
import { Refresh as RefreshIcon, BugReport as BugIcon } from '@mui/icons-material';
import { logger } from '../../services/logger/LoggerService';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
  onError?: (error: Error, errorInfo: ErrorInfo) => void;
}

interface State {
  hasError: boolean;
  error: Error | null;
  errorInfo: ErrorInfo | null;
}

/**
 * ErrorBoundary - Comprehensive error boundary for React components
 * Implements error handling from architecture section 8.1
 */
class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
    };
  }

  static getDerivedStateFromError(error: Error): State {
    return {
      hasError: true,
      error,
      errorInfo: null,
    };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    this.setState({
      error,
      errorInfo,
    });

    // Log error to service
    logger.error('ErrorBoundary caught an error', error);

    // Call custom error handler if provided
    if (this.props.onError) {
      this.props.onError(error, errorInfo);
    }
  }

  handleRetry = () => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
    });
  };

  handleReload = () => {
    window.location.reload();
  };

  render() {
    if (this.state.hasError) {
      // Custom fallback UI
      if (this.props.fallback) {
        return this.props.fallback;
      }

      // Default error UI
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
              An unexpected error occurred. Please try refreshing the page or contact support if the problem persists.
            </Typography>
          </Alert>

          <Box display="flex" gap={2} mt={2}>
            <Button
              variant="contained"
              startIcon={<RefreshIcon />}
              onClick={this.handleRetry}
              color="primary"
            >
              Try Again
            </Button>
            <Button
              variant="outlined"
              startIcon={<BugIcon />}
              onClick={this.handleReload}
              color="secondary"
            >
              Reload Page
            </Button>
          </Box>

          {process.env.NODE_ENV === 'development' && this.state.error && (
            <Box sx={{ mt: 3, p: 2, bgcolor: 'grey.100', borderRadius: 1, maxWidth: 800 }}>
              <Typography variant="h6" color="error" gutterBottom>
                Development Error Details:
              </Typography>
              <Typography variant="body2" component="pre" sx={{ whiteSpace: 'pre-wrap', fontSize: '0.75rem' }}>
                {this.state.error.toString()}
              </Typography>
              {this.state.errorInfo && (
                <Typography variant="body2" component="pre" sx={{ whiteSpace: 'pre-wrap', fontSize: '0.75rem', mt: 1 }}>
                  {this.state.errorInfo.componentStack}
                </Typography>
              )}
            </Box>
          )}
        </Box>
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary;
