import React, { useEffect } from 'react';
import { Box, Typography, Paper, Alert, CircularProgress } from '@mui/material';
import { useCameraStore } from '../../stores/cameraStore';
import CameraGrid from './CameraGrid';
import ConnectionStatus from '../common/ConnectionStatus';
import RealTimeStatus from '../common/RealTimeStatus';

const Dashboard: React.FC = () => {
  const {
    cameras,
    isLoading,
    isRefreshing,
    isConnecting,
    isConnected,
    error,
    serverInfo,
    initialize,
    refreshCameras,
    disconnect,
  } = useCameraStore();

  useEffect(() => {
    // Initialize connection on component mount
    try {
      initialize();
    } catch (err) {
      console.error('Failed to initialize camera store:', err);
    }
    
    // Cleanup on unmount
    return () => {
      try {
        disconnect();
      } catch (err) {
        console.error('Failed to disconnect:', err);
      }
    };
  }, [initialize, disconnect]);

  const handleRefresh = () => {
    try {
      refreshCameras();
    } catch (err) {
      console.error('Failed to refresh cameras:', err);
    }
  };

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ mb: 3 }}>
        <Typography variant="h4" component="h1" gutterBottom>
          Camera Dashboard
        </Typography>
        
        <ConnectionStatus 
          isConnected={isConnected}
          isConnecting={isConnecting}
          onRefresh={handleRefresh}
        />
        
        {/* Real-time Status */}
        <RealTimeStatus 
          showDetails={true}
          showRecordingProgress={true}
          showConnectionMetrics={true}
        />
      </Box>

      {/* Error Display */}
      {error && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {error}
        </Alert>
      )}

      {/* Server Info */}
      {serverInfo && (
        <Paper sx={{ p: 2, mb: 3 }}>
          <Typography variant="h6" gutterBottom>
            Server Information
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Version: {serverInfo.version} | Connected: {serverInfo.cameras_connected}
          </Typography>
        </Paper>
      )}

      {/* Loading State */}
      {isLoading && (
        <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
          <CircularProgress />
        </Box>
      )}

      {/* Camera Grid */}
      {!isLoading && (
        <Box>
          <Box sx={{ mb: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Typography variant="h6">
              Available Cameras ({cameras?.length || 0})
            </Typography>
            {isRefreshing && <CircularProgress size={20} />}
          </Box>
          
          {(!cameras || cameras.length === 0) && isConnected ? (
            <Paper sx={{ p: 3, textAlign: 'center' }}>
              <Typography variant="body1" color="text.secondary">
                No cameras found. Please check your camera connections.
              </Typography>
            </Paper>
          ) : (
            <CameraGrid cameras={cameras || []} />
          )}
        </Box>
      )}

      {/* Sprint 3 Status */}
      <Paper sx={{ p: 3, mt: 3, textAlign: 'center' }}>
        <Typography variant="h6" gutterBottom>
          Sprint 3: Server Integration Complete! 🚀
        </Typography>
        <Typography variant="body2" color="text.secondary">
          Real camera data integration and WebSocket connection working successfully.
        </Typography>
        <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
          Next: Real-time updates and camera operations
        </Typography>
      </Paper>
    </Box>
  );
};

export default Dashboard; 