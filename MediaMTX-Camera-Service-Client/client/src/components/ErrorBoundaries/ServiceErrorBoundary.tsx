/**
 * Service Error Boundary
 * 
 * Architecture: Error Boundary Pattern for Services
 * - Catches errors in service operations
 * - Provides service-specific error recovery
 * - Handles service degradation gracefully
 */

import React, { Component, ErrorInfo, ReactNode } from 'react';
import { Box, Typography, Button, Paper, Alert, Chip } from '@mui/material';
import { 
  Error as ErrorIcon, 
  Refresh as RefreshIcon, 
  WifiOff as OfflineIcon,
  Warning as WarningIcon 
} from '@mui/icons-material';
import { logger, loggers } from '../../services/loggerService';

interface Props {
  children: ReactNode;
  serviceName: string;
  fallback?: ReactNode;
  onError?: (error: Error, errorInfo: ErrorInfo) => void;
  retryable?: boolean;
  maxRetries?: number;
}

interface State {
  hasError: boolean;
  error?: Error;
  errorInfo?: ErrorInfo;
  retryCount: number;
  isRetrying: boolean;
}

class ServiceErrorBoundary extends Component<Props, State> {
  private maxRetries: number;

  constructor(props: Props) {
    super(props);
    this.maxRetries = props.maxRetries || 3;
    this.state = {
      hasError: false,
      retryCount: 0,
      isRetrying: false
    };
  }

  static getDerivedStateFromError(error: Error): Partial<State> {
    return {
      hasError: true,
      error
    };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    const { serviceName, onError } = this.props;
    
    // Log service error with context
    loggers.service.error(serviceName, 'operation', error);
    logger.error(`Service error in ${serviceName}`, error, 'service-error-boundary', {
      componentStack: errorInfo.componentStack,
      retryCount: this.state.retryCount,
      serviceName
    });

    // Update state with error info
    this.setState({
      error,
      errorInfo
    });

    // Call custom error handler if provided
    onError?.(error, errorInfo);
  }

  handleRetry = async () => {
    const { retryCount } = this.state;
    const { serviceName } = this.props;
    
    if (retryCount < this.maxRetries) {
      this.setState({ isRetrying: true });
      
      logger.info(`Retrying ${serviceName} service (attempt ${retryCount + 1}/${this.maxRetries})`, undefined, 'service-error-boundary');
      
      try {
        // Wait a bit before retrying
        await new Promise(resolve => setTimeout(resolve, 1000));
        
        this.setState({
          hasError: false,
          error: undefined,
          errorInfo: undefined,
          retryCount: retryCount + 1,
          isRetrying: false
        });
      } catch (error) {
        this.setState({ isRetrying: false });
        logger.error(`Retry failed for ${serviceName}`, error as Error, 'service-error-boundary');
      }
    }
  };

  handleFallback = () => {
    const { serviceName } = this.props;
    logger.info(`Using fallback mode for ${serviceName}`, undefined, 'service-error-boundary');
    
    // This would typically trigger a fallback service or offline mode
    // For now, we'll just log and continue
  };

  getErrorSeverity = (): 'error' | 'warning' | 'info' => {
    const { retryCount } = this.state;
    if (retryCount >= this.maxRetries) return 'error';
    if (retryCount > 0) return 'warning';
    return 'info';
  };

  getErrorIcon = () => {
    const { retryCount } = this.state;
    if (retryCount >= this.maxRetries) return <ErrorIcon />;
    if (retryCount > 0) return <WarningIcon />;
    return <OfflineIcon />;
  };

  render() {
    const { hasError, error, retryCount, isRetrying } = this.state;
    const { children, serviceName, fallback, retryable = true } = this.props;

    if (hasError) {
      // Use custom fallback if provided
      if (fallback) {
        return fallback;
      }

      const severity = this.getErrorSeverity();
      const canRetry = retryable && retryCount < this.maxRetries;

      return (
        <Box sx={{ p: 2 }}>
          <Paper elevation={1} sx={{ p: 3 }}>
            <Alert severity={severity} sx={{ mb: 2 }}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                {this.getErrorIcon()}
                <Typography variant="subtitle2">
                  {serviceName} Service Error
                </Typography>
                {retryCount > 0 && (
                  <Chip 
                    label={`Attempt ${retryCount}/${this.maxRetries}`} 
                    size="small" 
                    color={severity === 'error' ? 'error' : 'warning'}
                  />
                )}
              </Box>
            </Alert>
            
            <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
              The {serviceName} service encountered an error. 
              {canRetry ? ' You can try again or use fallback mode.' : ' Please check your connection or try again later.'}
            </Typography>

            <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap' }}>
              {canRetry && (
                <Button
                  variant="contained"
                  startIcon={<RefreshIcon />}
                  onClick={this.handleRetry}
                  disabled={isRetrying}
                  color="primary"
                >
                  {isRetrying ? 'Retrying...' : 'Try Again'}
                </Button>
              )}
              
              <Button
                variant="outlined"
                onClick={this.handleFallback}
                color="secondary"
              >
                Use Fallback
              </Button>
            </Box>

            {process.env.NODE_ENV === 'development' && error && (
              <Box sx={{ mt: 3 }}>
                <Typography variant="subtitle2" gutterBottom>
                  Error Details (Development):
                </Typography>
                <Typography
                  variant="body2"
                  component="pre"
                  sx={{
                    backgroundColor: 'grey.100',
                    p: 2,
                    borderRadius: 1,
                    overflow: 'auto',
                    fontSize: '0.75rem',
                    maxHeight: 150
                  }}
                >
                  {error.toString()}
                </Typography>
              </Box>
            )}
          </Paper>
        </Box>
      );
    }

    return children;
  }
}

export default ServiceErrorBoundary;
