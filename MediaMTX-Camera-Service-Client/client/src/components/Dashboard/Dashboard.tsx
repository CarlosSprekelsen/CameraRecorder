import React, { useEffect } from 'react';
import { Box, Typography, Paper, Alert, CircularProgress } from '@mui/material';
import { useCameraStore } from '../../stores/cameraStore';
import CameraGrid from './CameraGrid';
import ConnectionStatus from '../common/ConnectionStatus';
import RealTimeStatus from '../common/RealTimeStatus';

const Dashboard: React.FC = () => {
  const {
    cameras: storeCameras,
    isLoading: storeIsLoading,
    isRefreshing: storeIsRefreshing,
    isConnected: storeIsConnected,
    error: storeError,
    serverInfo: storeServerInfo,
    initialize: storeInitialize,
    refreshCameras: storeRefreshCameras,
    disconnect: storeDisconnect,
  } = useCameraStore();

  useEffect(() => {
    // Initialize connection on component mount
    try {
      storeInitialize();
    } catch (err) {
      console.error('Failed to initialize camera store:', err);
    }
    
    // Cleanup on unmount
    return () => {
      try {
        storeDisconnect();
      } catch (err) {
        console.error('Failed to disconnect:', err);
      }
    };
  }, [storeInitialize, storeDisconnect]);

  const handleRefresh = () => {
    try {
      storeRefreshCameras();
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
      {storeError && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {storeError}
        </Alert>
      )}

      {/* Server Info */}
      {storeServerInfo && (
        <Paper sx={{ p: 2, mb: 3 }}>
          <Typography variant="h6" gutterBottom>
            Server Information
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Version: {storeServerInfo.version} | Connected: {storeServerInfo.cameras_connected}
          </Typography>
        </Paper>
      )}

      {/* Loading State */}
      {storeIsLoading && (
        <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
          <CircularProgress />
        </Box>
      )}

      {/* Camera Grid */}
      {!storeIsLoading && (
        <Box>
          <Box sx={{ mb: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <Typography variant="h6">
            Available Cameras ({storeCameras?.length || 0})
          </Typography>
          {storeIsRefreshing && <CircularProgress size={20} />}
          </Box>
          
          {(!storeCameras || storeCameras.length === 0) && storeIsConnected ? (
            <Paper sx={{ p: 3, textAlign: 'center' }}>
              <Typography variant="body1" color="text.secondary">
                No cameras found. Please check your camera connections.
              </Typography>
            </Paper>
          ) : (
            <CameraGrid cameras={storeCameras || []} />
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