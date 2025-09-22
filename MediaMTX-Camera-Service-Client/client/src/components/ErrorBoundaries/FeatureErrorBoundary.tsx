/**
 * Feature Error Boundary
 * 
 * Architecture: Error Boundary Pattern
 * - Catches errors within specific features
 * - Provides feature-specific error recovery
 * - Maintains app stability while isolating failures
 */

import React, { Component, ErrorInfo, ReactNode } from 'react';
import { Box, Typography, Button, Paper, Alert, Collapse } from '@mui/material';
import { Error as ErrorIcon, Refresh as RefreshIcon, BugReport as BugIcon } from '@mui/icons-material';
import { logger, loggers } from '../../services/loggerService';

interface Props {
  children: ReactNode;
  featureName: string;
  fallback?: ReactNode;
  onError?: (error: Error, errorInfo: ErrorInfo) => void;
  showDetails?: boolean;
}

interface State {
  hasError: boolean;
  error?: Error;
  errorInfo?: ErrorInfo;
  retryCount: number;
  showDetails: boolean;
}

class FeatureErrorBoundary extends Component<Props, State> {
  private maxRetries = 3;

  constructor(props: Props) {
    super(props);
    this.state = {
      hasError: false,
      retryCount: 0,
      showDetails: props.showDetails || false
    };
  }

  static getDerivedStateFromError(error: Error): Partial<State> {
    return {
      hasError: true,
      error
    };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    const { featureName, onError } = this.props;
    
    // Log error with context
    loggers.component.error(featureName, error);
    logger.error(`Feature error in ${featureName}`, error, 'error-boundary', {
      componentStack: errorInfo.componentStack,
      retryCount: this.state.retryCount
    });

    // Update state with error info
    this.setState({
      error,
      errorInfo
    });

    // Call custom error handler if provided
    onError?.(error, errorInfo);
  }

  handleRetry = () => {
    const { retryCount } = this.state;
    
    if (retryCount < this.maxRetries) {
      logger.info(`Retrying ${this.props.featureName} (attempt ${retryCount + 1}/${this.maxRetries})`, undefined, 'error-boundary');
      
      this.setState({
        hasError: false,
        error: undefined,
        errorInfo: undefined,
        retryCount: retryCount + 1
      });
    } else {
      logger.warn(`Max retries reached for ${this.props.featureName}`, undefined, 'error-boundary');
    }
  };

  handleReload = () => {
    logger.info(`Reloading page due to error in ${this.props.featureName}`, undefined, 'error-boundary');
    window.location.reload();
  };

  toggleDetails = () => {
    this.setState(prevState => ({
      showDetails: !prevState.showDetails
    }));
  };

  render() {
    const { hasError, error, errorInfo, retryCount, showDetails } = this.state;
    const { children, featureName, fallback } = this.props;

    if (hasError) {
      // Use custom fallback if provided
      if (fallback) {
        return fallback;
      }

      return (
        <Box sx={{ p: 2 }}>
          <Paper elevation={2} sx={{ p: 3, textAlign: 'center' }}>
            <ErrorIcon sx={{ fontSize: 48, color: 'error.main', mb: 2 }} />
            
            <Typography variant="h6" gutterBottom>
              {featureName} Error
            </Typography>
            
            <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
              Something went wrong in the {featureName} feature. 
              {retryCount < this.maxRetries ? ' You can try again or reload the page.' : ' Please reload the page.'}
            </Typography>

            {retryCount > 0 && (
              <Alert severity="info" sx={{ mb: 2 }}>
                Retry attempt {retryCount} of {this.maxRetries}
              </Alert>
            )}

            <Box sx={{ display: 'flex', gap: 2, justifyContent: 'center', mb: 2 }}>
              {retryCount < this.maxRetries && (
                <Button
                  variant="contained"
                  startIcon={<RefreshIcon />}
                  onClick={this.handleRetry}
                  color="primary"
                >
                  Try Again
                </Button>
              )}
              
              <Button
                variant="outlined"
                onClick={this.handleReload}
                color="primary"
              >
                Reload Page
              </Button>
            </Box>

            {process.env.NODE_ENV === 'development' && error && (
              <Box sx={{ mt: 3 }}>
                <Button
                  variant="text"
                  startIcon={<BugIcon />}
                  onClick={this.toggleDetails}
                  size="small"
                >
                  {showDetails ? 'Hide' : 'Show'} Error Details
                </Button>
                
                <Collapse in={showDetails}>
                  <Box sx={{ mt: 2, textAlign: 'left' }}>
                    <Typography variant="subtitle2" gutterBottom>
                      Error Details:
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
                        maxHeight: 200
                      }}
                    >
                      {error.toString()}
                    </Typography>
                    
                    {errorInfo && (
                      <>
                        <Typography variant="subtitle2" gutterBottom sx={{ mt: 2 }}>
                          Component Stack:
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
                            maxHeight: 200
                          }}
                        >
                          {errorInfo.componentStack}
                        </Typography>
                      </>
                    )}
                  </Box>
                </Collapse>
              </Box>
            )}
          </Paper>
        </Box>
      );
    }

    return children;
  }
}

export default FeatureErrorBoundary;
