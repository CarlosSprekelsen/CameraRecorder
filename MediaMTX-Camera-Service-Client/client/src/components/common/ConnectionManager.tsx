import React, { useEffect, useRef } from 'react';
import { Box, Alert, Button, Typography, CircularProgress } from '@mui/material';
import { useConnectionStore, useHealthStore } from '../../stores/connection';
import ConnectionStatus from '../ConnectionStatus/ConnectionStatus';
import { connectionService } from '../../services/connectionService';
import { logger, loggers } from '../../services/loggerService';

interface ConnectionManagerProps {
  children: React.ReactNode;
  autoConnect?: boolean;
  showConnectionUI?: boolean;
}

const ConnectionManager: React.FC<ConnectionManagerProps> = ({ 
  children, 
  autoConnect = true,
  showConnectionUI = true
}) => {
  // Use new modular stores
  const {
    status,
    isConnecting,
    isReconnecting,
    error,
    autoReconnect,
    setAutoReconnect,
    clearError
  } = useConnectionStore();

  const {
    isHealthy,
    healthScore
  } = useHealthStore();

  const hasInitialized = useRef(false);

  // Initialize connection on mount
  useEffect(() => {
    if (autoConnect && !hasInitialized.current) {
      hasInitialized.current = true;
      loggers.service.start('ConnectionManager', 'initialize');
      
      connectionService.connect()
        .then(() => {
          loggers.service.success('ConnectionManager', 'initialize');
        })
        .catch((error) => {
          loggers.service.error('ConnectionManager', 'initialize', error);
        });
    }
  }, [autoConnect]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      // Don't disconnect on unmount to allow for reconnection attempts
      // The connection service will handle cleanup
    };
  }, []);

  const handleConnect = () => {
    loggers.service.start('ConnectionManager', 'connect');
    connectionService.connect()
      .then(() => {
        loggers.service.success('ConnectionManager', 'connect');
      })
      .catch((error) => {
        loggers.service.error('ConnectionManager', 'connect', error);
      });
  };

  const handleForceReconnect = () => {
    loggers.service.start('ConnectionManager', 'forceReconnect');
    connectionService.forceReconnect()
      .then(() => {
        loggers.service.success('ConnectionManager', 'forceReconnect');
      })
      .catch((error) => {
        loggers.service.error('ConnectionManager', 'forceReconnect', error);
      });
  };

  const handleToggleAutoReconnect = () => {
    setAutoReconnect(!autoReconnect);
  };

  // Show loading state while connecting
  if (isConnecting && status === 'connecting') {
    return (
      <Box sx={{ 
        display: 'flex', 
        flexDirection: 'column', 
        alignItems: 'center', 
        justifyContent: 'center', 
        minHeight: '200px',
        gap: 2
      }}>
        <CircularProgress size={40} />
        <Typography variant="h6" color="text.secondary">
          Connecting to MediaMTX Camera Service...
        </Typography>
        <Typography variant="body2" color="text.secondary">
          Please wait while we establish a connection to the server.
        </Typography>
      </Box>
    );
  }

  // Show reconnecting state
  if (isReconnecting) {
    return (
      <Box sx={{ 
        display: 'flex', 
        flexDirection: 'column', 
        alignItems: 'center', 
        justifyContent: 'center', 
        minHeight: '200px',
        gap: 2
      }}>
        <CircularProgress size={40} />
        <Typography variant="h6" color="text.secondary">
          Reconnecting...
        </Typography>
        <Typography variant="body2" color="text.secondary">
          Attempting to restore connection to the server.
        </Typography>
      </Box>
    );
  }

  // Show error state with recovery options
  if (status === 'error' && error) {
    return (
      <Box sx={{ 
        display: 'flex', 
        flexDirection: 'column', 
        alignItems: 'center', 
        justifyContent: 'center', 
        minHeight: '200px',
        gap: 3,
        p: 3
      }}>
        <Alert 
          severity="error" 
          sx={{ width: '100%', maxWidth: 600 }}
          action={
            <Button color="inherit" size="small" onClick={clearError}>
              Dismiss
            </Button>
          }
        >
          <Typography variant="h6" gutterBottom>
            Connection Error
          </Typography>
          <Typography variant="body2" paragraph>
            {error}
          </Typography>
          <Typography variant="caption" color="text.secondary">
            Unable to connect to the MediaMTX Camera Service. Please check your network connection and try again.
          </Typography>
        </Alert>

        <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap', justifyContent: 'center' }}>
          <Button 
            variant="contained" 
            onClick={handleConnect}
            disabled={isConnecting}
          >
            Try Again
          </Button>
          <Button 
            variant="outlined" 
            onClick={handleForceReconnect}
            disabled={isConnecting}
          >
            Force Reconnect
          </Button>
          <Button 
            variant="outlined" 
            onClick={handleToggleAutoReconnect}
            color={autoReconnect ? 'success' : 'primary'}
          >
            Auto-reconnect {autoReconnect ? 'ON' : 'OFF'}
          </Button>
        </Box>

        {showConnectionUI && (
          <Box sx={{ mt: 2 }}>
            <ConnectionStatus showDetails={true} />
          </Box>
        )}
      </Box>
    );
  }

  // Show disconnected state with connection options
  if (status === 'disconnected') {
    return (
      <Box sx={{ 
        display: 'flex', 
        flexDirection: 'column', 
        alignItems: 'center', 
        justifyContent: 'center', 
        minHeight: '200px',
        gap: 3,
        p: 3
      }}>
        <Alert severity="warning" sx={{ width: '100%', maxWidth: 600 }}>
          <Typography variant="h6" gutterBottom>
            Not Connected
          </Typography>
          <Typography variant="body2" paragraph>
            The application is not connected to the MediaMTX Camera Service.
          </Typography>
          <Typography variant="caption" color="text.secondary">
            Click "Connect" to establish a connection and start using the camera service.
          </Typography>
        </Alert>

        <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap', justifyContent: 'center' }}>
          <Button 
            variant="contained" 
            onClick={handleConnect}
            disabled={isConnecting}
          >
            Connect
          </Button>
          <Button 
            variant="outlined" 
            onClick={handleToggleAutoReconnect}
            color={autoReconnect ? 'success' : 'primary'}
          >
            Auto-reconnect {autoReconnect ? 'ON' : 'OFF'}
          </Button>
        </Box>

        {showConnectionUI && (
          <Box sx={{ mt: 2 }}>
            <ConnectionStatus showDetails={true} />
          </Box>
        )}
      </Box>
    );
  }

  // Show connected state with children
  if (status === 'connected') {
    return (
      <Box>
        {children}
      </Box>
    );
  }

  // Fallback loading state
  return (
    <Box sx={{ 
      display: 'flex', 
      flexDirection: 'column', 
      alignItems: 'center', 
      justifyContent: 'center', 
      minHeight: '200px',
      gap: 2
    }}>
      <CircularProgress size={40} />
      <Typography variant="h6" color="text.secondary">
        Initializing...
      </Typography>
    </Box>
  );
};

export default ConnectionManager;
