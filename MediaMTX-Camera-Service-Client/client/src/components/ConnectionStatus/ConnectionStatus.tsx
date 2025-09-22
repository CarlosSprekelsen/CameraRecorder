/**
 * ConnectionStatus Component
 * 
 * Architecture: Service Layer Pattern
 * - Uses ConnectionService instead of direct store access
 * - Follows proper abstraction layer principles
 * - Maintains separation of concerns
 */

import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Button,
  Chip,
  Alert,
  CircularProgress,
  Grid,
  Divider,
} from '@mui/material';
import {
  Wifi,
  WifiOff,
  CheckCircle,
  Error,
  Warning,
  Refresh,
  Settings,
} from '@mui/icons-material';
import { useConnectionStore } from '../../stores/connectionStore';
import { useHealthStore } from '../../stores/healthStore';
import { connectionService } from '../../services/connectionService';
import { logger, loggers } from '../../services/loggerService';

const ConnectionStatus: React.FC = () => {
  const [localLoading, setLocalLoading] = useState(false);
  const [localError, setLocalError] = useState<string | null>(null);

  // Store state
  const {
    websocketStatus,
    healthStatus,
    lastConnected,
    lastError,
    isConnected,
    connect,
    disconnect,
    reconnect,
  } = useConnectionStore();

  const {
    systemHealth,
    cameraHealth,
    mediamtxHealth,
    isLoading,
    error: healthError,
    refreshHealth,
  } = useHealthStore();

  // Local handlers using service layer
  const handleConnect = async () => {
    setLocalLoading(true);
    setLocalError(null);
    loggers.service.start('ConnectionService', 'connect');
    
    try {
      await connectionService.connect();
      loggers.service.success('ConnectionService', 'connect');
    } catch (error: unknown) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to connect';
      setLocalError(errorMessage);
      loggers.service.error('ConnectionService', 'connect', error as Error);
    } finally {
      setLocalLoading(false);
    }
  };

  const handleDisconnect = async () => {
    setLocalLoading(true);
    setLocalError(null);
    loggers.service.start('ConnectionService', 'disconnect');
    
    try {
      await connectionService.disconnect();
      loggers.service.success('ConnectionService', 'disconnect');
    } catch (error: unknown) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to disconnect';
      setLocalError(errorMessage);
      loggers.service.error('ConnectionService', 'disconnect', error as Error);
    } finally {
      setLocalLoading(false);
    }
  };

  const handleReconnect = async () => {
    setLocalLoading(true);
    setLocalError(null);
    loggers.service.start('ConnectionService', 'reconnect');
    
    try {
      await connectionService.forceReconnect();
      loggers.service.success('ConnectionService', 'reconnect');
    } catch (error: unknown) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to reconnect';
      setLocalError(errorMessage);
      loggers.service.error('ConnectionService', 'reconnect', error as Error);
    } finally {
      setLocalLoading(false);
    }
  };

  const handleRefreshHealth = async () => {
    setLocalLoading(true);
    setLocalError(null);
    try {
      await refreshHealth();
    } catch (error: unknown) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to refresh health';
      setLocalError(errorMessage);
    } finally {
      setLocalLoading(false);
    }
  };

  const getConnectionStatusColor = (): 'success' | 'warning' | 'error' | 'default' => {
    if (isConnected) return 'success';
    if (websocketStatus === 'connecting') return 'warning';
    return 'error';
  };

  const getConnectionStatusIcon = () => {
    if (isConnected) return <Wifi color="success" />;
    if (websocketStatus === 'connecting') return <CircularProgress size={20} />;
    return <WifiOff color="error" />;
  };

  const getHealthStatusColor = (status: string): 'success' | 'warning' | 'error' | 'default' => {
    switch (status) {
      case 'healthy':
        return 'success';
      case 'degraded':
        return 'warning';
      case 'unhealthy':
        return 'error';
      default:
        return 'default';
    }
  };

  const getHealthStatusIcon = (status: string) => {
    switch (status) {
      case 'healthy':
        return <CheckCircle color="success" />;
      case 'degraded':
        return <Warning color="warning" />;
      case 'unhealthy':
        return <Error color="error" />;
      default:
        return <Warning color="warning" />;
    }
  };

  // Initialize component
  useEffect(() => {
    handleRefreshHealth();
  }, []);

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h5" gutterBottom>
          Connection Status
        </Typography>
        <Box display="flex" gap={1}>
          <Button
            variant="outlined"
            onClick={handleRefreshHealth}
            disabled={localLoading}
            startIcon={<Refresh />}
          >
            Refresh
          </Button>
          <Button
            variant="outlined"
            startIcon={<Settings />}
          >
            Settings
          </Button>
        </Box>
      </Box>

      {/* Error Display */}
      {(localError || lastError || healthError) && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {localError || lastError || healthError}
        </Alert>
      )}

      <Grid container spacing={3}>
        {/* WebSocket Connection Status */}
        <Grid item xs={12} md={6}>
          <Card variant="outlined">
            <CardContent>
              <Box display="flex" alignItems="center" justifyContent="space-between" mb={2}>
                <Typography variant="h6">
                  WebSocket Connection
                </Typography>
                <Chip
                  icon={getConnectionStatusIcon()}
                  label={websocketStatus}
                  color={getConnectionStatusColor()}
                  size="small"
                />
              </Box>

              <Box mb={2}>
                <Typography variant="body2" color="textSecondary">
                  Status: {websocketStatus}
                </Typography>
                {lastConnected && (
                  <Typography variant="body2" color="textSecondary">
                    Last Connected: {new Date(lastConnected).toLocaleString()}
                  </Typography>
                )}
              </Box>

              <Box display="flex" gap={1}>
                {!isConnected ? (
                  <Button
                    variant="contained"
                    color="primary"
                    onClick={handleConnect}
                    disabled={localLoading}
                    fullWidth
                  >
                    Connect
                  </Button>
                ) : (
                  <>
                    <Button
                      variant="contained"
                      color="error"
                      onClick={handleDisconnect}
                      disabled={localLoading}
                    >
                      Disconnect
                    </Button>
                    <Button
                      variant="outlined"
                      onClick={handleReconnect}
                      disabled={localLoading}
                    >
                      Reconnect
                    </Button>
                  </>
                )}
              </Box>
            </CardContent>
          </Card>
        </Grid>

        {/* Health Status */}
        <Grid item xs={12} md={6}>
          <Card variant="outlined">
            <CardContent>
              <Typography variant="h6" gutterBottom>
                System Health
              </Typography>

              {isLoading ? (
                <Box display="flex" justifyContent="center" p={2}>
                  <CircularProgress />
                </Box>
              ) : (
                <Box>
                  {/* System Health */}
                  {systemHealth && (
                    <Box mb={2}>
                      <Box display="flex" alignItems="center" justifyContent="space-between">
                        <Typography variant="body2">
                          System
                        </Typography>
                        <Chip
                          icon={getHealthStatusIcon(systemHealth.status)}
                          label={systemHealth.status}
                          color={getHealthStatusColor(systemHealth.status)}
                          size="small"
                        />
                      </Box>
                    </Box>
                  )}

                  {/* Camera Health */}
                  {cameraHealth && (
                    <Box mb={2}>
                      <Box display="flex" alignItems="center" justifyContent="space-between">
                        <Typography variant="body2">
                          Cameras
                        </Typography>
                        <Chip
                          icon={getHealthStatusIcon(cameraHealth.status)}
                          label={cameraHealth.status}
                          color={getHealthStatusColor(cameraHealth.status)}
                          size="small"
                        />
                      </Box>
                    </Box>
                  )}

                  {/* MediaMTX Health */}
                  {mediamtxHealth && (
                    <Box mb={2}>
                      <Box display="flex" alignItems="center" justifyContent="space-between">
                        <Typography variant="body2">
                          MediaMTX
                        </Typography>
                        <Chip
                          icon={getHealthStatusIcon(mediamtxHealth.status)}
                          label={mediamtxHealth.status}
                          color={getHealthStatusColor(mediamtxHealth.status)}
                          size="small"
                        />
                      </Box>
                    </Box>
                  )}

                  <Divider sx={{ my: 2 }} />

                  {/* Connection Summary */}
                  <Box>
                    <Typography variant="body2" color="textSecondary">
                      Overall Status: {isConnected ? 'Connected' : 'Disconnected'}
                    </Typography>
                    <Typography variant="body2" color="textSecondary">
                      Health Check: {isLoading ? 'Checking...' : 'Complete'}
                    </Typography>
                  </Box>
                </Box>
              )}
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Box>
  );
};

export default ConnectionStatus;
