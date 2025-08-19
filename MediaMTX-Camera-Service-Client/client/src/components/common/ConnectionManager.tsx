import React, { useEffect, useRef } from 'react';
import { Box, Alert, Button, Typography, CircularProgress } from '@mui/material';
import { useConnectionStore } from '../../stores/connectionStore';
import ConnectionStatus from './ConnectionStatus';

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
  const {
    status,
    isConnecting,
    isReconnecting,
    error,
    autoReconnect,
    connect,
    forceReconnect,
    setAutoReconnect,
    clearError
  } = useConnectionStore();

  const hasInitialized = useRef(false);

  // Initialize connection on mount
  useEffect(() => {
    if (autoConnect && !hasInitialized.current) {
      hasInitialized.current = true;
      connect().catch(console.error);
    }
  }, [autoConnect, connect]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      // Don't disconnect on unmount to allow for reconnection attempts
      // The connection store will handle cleanup
    };
  }, []);

  const handleConnect = () => {
    connect().catch(console.error);
  };

  const handleForceReconnect = () => {
    forceReconnect().catch(console.error);
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
