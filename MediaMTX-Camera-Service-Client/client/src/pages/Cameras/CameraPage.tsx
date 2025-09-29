import React, { useEffect, memo } from 'react';
import { Box } from '../../components/atoms/Box/Box';
import { Typography } from '../../components/atoms/Typography/Typography';
import { Card } from '../../components/atoms/Card/Card';
import { Alert } from '../../components/atoms/Alert/Alert';
import { CircularProgress } from '../../components/atoms/CircularProgress/CircularProgress';
import { useDeviceStore } from '../../stores/device/deviceStore';
// ARCHITECTURE FIX: Removed unused Camera import
import { useAuthStore } from '../../stores/auth/authStore';
// ARCHITECTURE FIX: Removed direct service import - use stores only
import CameraTable from '../../components/Cameras/CameraTable';
import { logger } from '../../services/logger/LoggerService';
// ARCHITECTURE FIX: Removed unused JsonRpcNotification import
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
    setDeviceService,
    handleCameraStatusUpdate,
  } = useDeviceStore();

  const { isAuthenticated } = useAuthStore();
  const { handleRecordingStatusUpdate, setRecordingService } = useRecordingStore();

  // Initialize device service and load data
  useEffect(() => {
    if (!isAuthenticated) {
      logger.warn('User not authenticated, skipping camera data load');
      return;
    }

    const initializeDeviceService = async () => {
      try {
        // ARCHITECTURE FIX: Services are managed by stores, not created directly
        logger.info('Services managed by stores - no direct service creation needed');

        // ARCHITECTURE FIX: Real-time updates handled by stores
        logger.info('Real-time updates managed by stores');

        // Load initial data
        await Promise.all([getCameraList(), getStreams()]);

        logger.info('Camera page initialized successfully');

        // ARCHITECTURE FIX: Cleanup handled by stores
      } catch (error) {
        logger.error('Failed to initialize camera page', error as Record<string, unknown>);
      }
    };

    initializeDeviceService();
  }, [
    isAuthenticated,
    getCameraList,
    getStreams,
    setDeviceService,
    handleCameraStatusUpdate,
    handleRecordingStatusUpdate,
    setRecordingService,
  ]);

  // Redirect to login if not authenticated
  if (!isAuthenticated) {
    return (
      <Box className="max-w-6xl mx-auto">
        <Box className="mt-4">
          <Alert variant="warning">Please log in to view camera devices.</Alert>
        </Box>
      </Box>
    );
  }

  return (
    <Box className="max-w-6xl mx-auto">
      <Box className="mt-4 mb-4">
        <Typography variant="h4" component="h1" className="mb-4">
          Camera Devices
        </Typography>

        {lastUpdated && (
          <Typography variant="body2" color="secondary" className="mb-2">
            Last updated: {new Date(lastUpdated).toLocaleString()}
          </Typography>
        )}

        {error && (
          <Alert variant="error" className="mb-2">
            {error}
          </Alert>
        )}

        <Card className="p-2">
          {loading ? (
            <Box className="flex justify-center items-center min-h-[200px]">
              <CircularProgress />
              <Typography variant="body1" className="ml-2">
                Loading camera devices...
              </Typography>
            </Box>
          ) : (
            <CameraTable cameras={cameras} streams={streams} onRefresh={getCameraList} />
          )}
        </Card>
      </Box>
    </Box>
  );
});

CameraPage.displayName = 'CameraPage';

export default CameraPage;
