import React, { useEffect } from 'react';
import {
  Box,
  Typography,
  Paper,
  Alert,
  CircularProgress,
  Container,
} from '@mui/material';
import { useDeviceStore } from '../../stores/device/deviceStore';
import { useAuthStore } from '../../stores/auth/authStore';
import { serviceFactory } from '../../services/ServiceFactory';
import CameraTable from '../../components/Cameras/CameraTable';
import { logger } from '../../services/logger/LoggerService';
import { JsonRpcNotification } from '../../types/api';
import { useRecordingStore } from '../../stores/recording/recordingStore';

/**
 * CameraPage - Main device table following architecture section 5.1
 * Implements I.Discovery interface for device discovery and stream links
 */
const CameraPage: React.FC = () => {
  const { 
    cameras, 
    streams, 
    loading, 
    error, 
    lastUpdated,
    getCameraList,
    getStreams,
    setDeviceService,
    handleCameraStatusUpdate
  } = useDeviceStore();

  const { isAuthenticated } = useAuthStore();
  const { handleRecordingStatusUpdate } = useRecordingStore();

  // Initialize device service and load data
  useEffect(() => {
    if (!isAuthenticated) {
      logger.warn('User not authenticated, skipping camera data load');
      return;
    }

    const initializeDeviceService = async () => {
      try {
        const wsService = serviceFactory.getWebSocketService();
        if (!wsService) {
          logger.error('WebSocket service not available');
          return;
        }

        const deviceService = serviceFactory.createDeviceService(wsService);
        setDeviceService(deviceService);

        // Set up notification service for real-time updates
        const notificationService = serviceFactory.createNotificationService(wsService);
        
        // Subscribe to camera status updates
        const unsubscribeCameraUpdates = notificationService.subscribe(
          'camera_status_update',
          (notification: JsonRpcNotification) => {
            if (notification.params) {
              handleCameraStatusUpdate(notification.params);
            }
          }
        );

        // Subscribe to recording status updates
        const unsubscribeRecordingUpdates = notificationService.subscribe(
          'recording_status_update',
          (notification: JsonRpcNotification) => {
            if (notification.params) {
              handleRecordingStatusUpdate(notification.params as any);
            }
          }
        );

        // Subscribe to real-time events
        await deviceService.subscribeToCameraEvents();

        // Load initial data
        await Promise.all([
          getCameraList(),
          getStreams()
        ]);

        logger.info('Camera page initialized successfully');

        // Cleanup function
        return () => {
          unsubscribeCameraUpdates();
          unsubscribeRecordingUpdates();
        };
      } catch (error) {
        logger.error('Failed to initialize camera page', error as Error);
      }
    };

    initializeDeviceService();
  }, [isAuthenticated, getCameraList, getStreams, setDeviceService]);

  // Redirect to login if not authenticated
  if (!isAuthenticated) {
    return (
      <Container maxWidth="lg">
        <Box sx={{ mt: 4 }}>
          <Alert severity="warning">
            Please log in to view camera devices.
          </Alert>
        </Box>
      </Container>
    );
  }

  return (
    <Container maxWidth="lg">
      <Box sx={{ mt: 4, mb: 4 }}>
        <Typography variant="h4" component="h1" gutterBottom>
          Camera Devices
        </Typography>
        
        {lastUpdated && (
          <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
            Last updated: {new Date(lastUpdated).toLocaleString()}
          </Typography>
        )}

        {error && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {error}
          </Alert>
        )}

        <Paper sx={{ p: 2 }}>
          {loading ? (
            <Box display="flex" justifyContent="center" alignItems="center" minHeight="200px">
              <CircularProgress />
              <Typography variant="body1" sx={{ ml: 2 }}>
                Loading camera devices...
              </Typography>
            </Box>
          ) : (
            <CameraTable 
              cameras={cameras}
              streams={streams}
              onRefresh={getCameraList}
            />
          )}
        </Paper>
      </Box>
    </Container>
  );
};

export default CameraPage;
