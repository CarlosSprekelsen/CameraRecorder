import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Button,
  Grid,
  Chip,
  Alert,
  CircularProgress,
  IconButton,
  Tooltip,
} from '@mui/material';
import {
  Videocam,
  VideocamOff,
  PhotoCamera,
  PlayArrow,
  Stop,
  Warning,
  Error,
  CheckCircle,
} from '@mui/icons-material';
import { useCameraStore } from '../../stores/cameraStore';
import { useRecordingStore } from '../../stores/recordingStore';
import { CameraDevice, CameraStatus } from '../../types/camera';

const CameraGrid: React.FC = () => {
  const [localLoading, setLocalLoading] = useState(false);
  const [localError, setLocalError] = useState<string | null>(null);

  // Store state
  const {
    cameras,
    isLoading,
    error: storeError,
    refreshCameras,
  } = useCameraStore();

  const {
    isRecording,
    getRecordingState,
  } = useRecordingStore();

  // Local handlers
  const handleRefresh = async () => {
    setLocalLoading(true);
    setLocalError(null);
    try {
      await refreshCameras();
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to refresh cameras';
      setLocalError(errorMessage);
    } finally {
      setLocalLoading(false);
    }
  };

  const getStatusColor = (status: CameraStatus): 'success' | 'warning' | 'error' | 'default' => {
    switch (status) {
      case 'CONNECTED':
        return 'success';
      case 'DISCONNECTED':
        return 'default';
      case 'ERROR':
        return 'error';
      default:
        return 'warning';
    }
  };

  const getStatusIcon = (status: CameraStatus) => {
    switch (status) {
      case 'CONNECTED':
        return <Videocam color="success" />;
      case 'DISCONNECTED':
        return <VideocamOff color="disabled" />;
      case 'ERROR':
        return <Error color="error" />;
      default:
        return <Warning color="warning" />;
    }
  };

  const getRecordingStatusIcon = (device: string) => {
    if (isRecording(device)) {
      return <PlayArrow color="success" />;
    }
    return <Stop color="disabled" />;
  };

  // Initialize component
  useEffect(() => {
    handleRefresh();
  }, []);

  if (isLoading || localLoading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="200px">
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h5" gutterBottom>
          Camera Grid
        </Typography>
        <Button
          variant="outlined"
          onClick={handleRefresh}
          disabled={localLoading}
          startIcon={<CheckCircle />}
        >
          Refresh Cameras
        </Button>
      </Box>

      {/* Error Display */}
      {(localError || storeError) && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {localError || storeError}
        </Alert>
      )}

      {/* Camera Grid */}
      <Grid container spacing={3}>
        {cameras.length === 0 ? (
          <Grid item xs={12}>
            <Card>
              <CardContent>
                <Typography variant="body1" color="textSecondary" textAlign="center">
                  No cameras found. Please check your camera connections.
                </Typography>
              </CardContent>
            </Card>
          </Grid>
        ) : (
          cameras.map((camera: CameraDevice) => (
            <Grid item xs={12} sm={6} md={4} key={camera.device}>
              <Card variant="outlined">
                <CardContent>
                  <Box display="flex" alignItems="center" justifyContent="space-between" mb={2}>
                    <Typography variant="subtitle1" fontWeight="bold">
                      {camera.name || camera.device}
                    </Typography>
                    <Chip
                      icon={getStatusIcon(camera.status)}
                      label={camera.status}
                      color={getStatusColor(camera.status)}
                      size="small"
                    />
                  </Box>

                  {/* Camera Details */}
                  <Box mb={2}>
                    <Typography variant="body2" color="textSecondary">
                      Device: {camera.device}
                    </Typography>
                    <Typography variant="body2" color="textSecondary">
                      Resolution: {camera.resolution}
                    </Typography>
                    <Typography variant="body2" color="textSecondary">
                      FPS: {camera.fps}
                    </Typography>
                    {camera.recording && (
                      <Typography variant="body2" color="success.main">
                        ● Recording Active
                      </Typography>
                    )}
                  </Box>

                  {/* Recording Status */}
                  <Box display="flex" alignItems="center" mb={2}>
                    <Typography variant="body2" color="textSecondary" mr={1}>
                      Recording:
                    </Typography>
                    {getRecordingStatusIcon(camera.device)}
                    <Typography variant="body2" ml={1}>
                      {isRecording(camera.device) ? 'Active' : 'Inactive'}
                    </Typography>
                  </Box>

                  {/* Camera Controls */}
                  <Box display="flex" gap={1}>
                    <Button
                      variant="contained"
                      color="primary"
                      startIcon={<PhotoCamera />}
                      disabled={camera.status !== 'CONNECTED'}
                      fullWidth
                    >
                      Snapshot
                    </Button>
                    <Button
                      variant="contained"
                      color="secondary"
                      startIcon={isRecording(camera.device) ? <Stop /> : <PlayArrow />}
                      disabled={camera.status !== 'CONNECTED'}
                      fullWidth
                    >
                      {isRecording(camera.device) ? 'Stop' : 'Record'}
                    </Button>
                  </Box>
                </CardContent>
              </Card>
            </Grid>
          ))
        )}
      </Grid>
    </Box>
  );
};

export default CameraGrid;
