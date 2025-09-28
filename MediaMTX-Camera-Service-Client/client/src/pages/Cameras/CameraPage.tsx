import React, { useEffect, memo } from 'react';
import { Box, Typography, Paper, Alert, CircularProgress, Container } from '@mui/material';
import { useDeviceStore, Camera } from '../../stores/device/deviceStore';
import { useAuthStore } from '../../stores/auth/authStore';
// ARCHITECTURE FIX: Removed serviceFactory import - components must use stores only
import CameraTable from '../../components/Cameras/CameraTable';
import { logger } from '../../services/logger/LoggerService';
import { JsonRpcNotification } from '../../types/api';
import { useRecordingStore } from '../../stores/recording/recordingStore';

/**
 * CameraPage - Main device management interface
 *
 * Implements the I.Discovery interface for device discovery and stream links.
 * Provides a comprehensive view of all connected cameras with real-time status updates.
 *
 * @component
 * @returns {JSX.Element} The camera management page
 *
 * @features
 * - Real-time camera status monitoring
 * - Device discovery and enumeration
 * - Stream URL management
 * - Recording status tracking
 * - Error handling and loading states
 *
 * @example
 * ```tsx
 * <CameraPage />
 * ```
 *
 * @see {@link https://github.com/mediamtx/mediamtx} MediaMTX documentation
 * @see {@link ../../docs/architecture/client-architechture.md} Client Architecture
 */
const CameraPage: React.FC = memo(() => {
  const {
    cameras,
    streams,
    loading,
    error,
    lastUpdated,
    getCameraList,
    getStreams,
    // ARCHITECTURE FIX: Removed setDeviceService - components don't inject services
    handleCameraStatusUpdate,
  } = useDeviceStore();

  const { isAuthenticated } = useAuthStore();
  const { handleRecordingStatusUpdate } = useRecordingStore();
  // ARCHITECTURE FIX: Removed setService - components don't inject services

  // Initialize device service and load data
  useEffect(() => {
    if (!isAuthenticated) {
      logger.warn('User not authenticated, skipping camera data load');
      return;
    }

    const initializeDeviceService = async () => {
      // ARCHITECTURE FIX: Removed direct service calls - stores handle all service interactions
      // Load initial data through store actions only
      try {
        await Promise.all([getCameraList(), getStreams()]);
        logger.info('Camera page initialized successfully');
      } catch (error) {
        logger.error('Failed to initialize camera page', error as Record<string, unknown>);
      }
    };

    initializeDeviceService();
  }, [
    isAuthenticated,
    getCameraList,
    getStreams,
    // ARCHITECTURE FIX: Removed service injection dependencies
    handleCameraStatusUpdate,
    handleRecordingStatusUpdate,
  ]);

  // Redirect to login if not authenticated
  if (!isAuthenticated) {
    return (
      <Container maxWidth="lg">
        <Box sx={{ mt: 4 }}>
          <Alert severity="warning">Please log in to view camera devices.</Alert>
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
            <CameraTable cameras={cameras} streams={streams} onRefresh={getCameraList} />
          )}
        </Paper>
      </Box>
    </Container>
  );
});

CameraPage.displayName = 'CameraPage';

export default CameraPage;
