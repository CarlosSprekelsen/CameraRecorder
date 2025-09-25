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
  Divider,
  List,
  ListItem,
  ListItemText,
  ListItemSecondaryAction,
} from '@mui/material';
import {
  PlayArrow,
  Stop,
  Videocam,
  VideocamOff,
  Warning,
  Error,
  CheckCircle,
  Timer,
  Storage,
} from '@mui/icons-material';
import { useRecordingStore } from '../../stores/recordingStore';
import { useCameraStore } from '../../stores/cameraStore';
import { RecordingSession, RecordingStatus } from '../../types/camera';
import { JSONRPCError } from '../../types/rpc';

// Error handling utility function
const getErrorMessage = (error: unknown): string => {
  if (error instanceof Error) {
    return error.message;
  }
  if (typeof error === 'string') {
    return error;
  }
  return 'An unknown error occurred';
};

const RecordingManager: React.FC = () => {
  const [localLoading, setLocalLoading] = useState(false);
  const [localError, setLocalError] = useState<string | null>(null);

  // Store state
  const {
    recordingStates,
    activeSessions,
    errors,
    progress,
    isLoading,
    error: storeError,
    startRecording,
    stopRecording,
    getRecordingState,
    isRecording,
    getRecordingProgress,
  } = useRecordingStore();

  const { cameras } = useCameraStore();

  // Local handlers
  const handleStartRecording = async (device: string) => {
    setLocalLoading(true);
    setLocalError(null);
    try {
      await startRecording(device);
    } catch (error: unknown) {
      const errorMessage = getErrorMessage(error);
      setLocalError(errorMessage);
    } finally {
      setLocalLoading(false);
    }
  };

  const handleStopRecording = async (device: string) => {
    setLocalLoading(true);
    setLocalError(null);
    try {
      await stopRecording(device);
    } catch (error: unknown) {
      const errorMessage = getErrorMessage(error);
      setLocalError(errorMessage);
    } finally {
      setLocalLoading(false);
    }
  };

  const getRecordingError = (device: string): JSONRPCError | null => {
    return errors.get(device) || null;
  };

  const getRecordingProgressData = (device: string) => {
    return getRecordingProgress(device);
  };

  const formatDuration = (seconds: number): string => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;
    return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  };

  const getStatusColor = (status: RecordingStatus): 'success' | 'warning' | 'error' | 'default' => {
    switch (status) {
      case 'RECORDING':
        return 'success';
      case 'STOPPED':
        return 'default';
      case 'ERROR':
        return 'error';
      default:
        return 'warning';
    }
  };

  const getStatusIcon = (status: RecordingStatus) => {
    switch (status) {
      case 'RECORDING':
        return <Videocam color="success" />;
      case 'STOPPED':
        return <VideocamOff color="disabled" />;
      case 'ERROR':
        return <Error color="error" />;
      default:
        return <Warning color="warning" />;
    }
  };

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="200px">
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box>
      <Typography variant="h5" gutterBottom>
        Recording Manager
      </Typography>

      {/* Error Display */}
      {(localError || storeError) && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {localError || storeError}
        </Alert>
      )}

      {/* Active Sessions */}
      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Active Recording Sessions ({activeSessions.size})
          </Typography>
          
          {activeSessions.size === 0 ? (
            <Typography color="textSecondary">
              No active recording sessions
            </Typography>
          ) : (
            <List>
              {Array.from(activeSessions.values()).map((session: RecordingSession) => (
                <ListItem key={session.device} divider>
                  <ListItemText
                    primary={`Camera: ${session.device}`}
                    secondary={
                      <Box>
                        <Typography variant="body2" color="textSecondary">
                          Started: {new Date(session.start_time).toLocaleString()}
                        </Typography>
                        {session.filename && (
                          <Typography variant="body2" color="textSecondary">
                            File: {session.filename}
                          </Typography>
                        )}
                      </Box>
                    }
                  />
                  <ListItemSecondaryAction>
                    <Button
                      variant="contained"
                      color="error"
                      startIcon={<Stop />}
                      onClick={() => handleStopRecording(session.device)}
                      disabled={localLoading}
                    >
                      Stop Recording
                    </Button>
                  </ListItemSecondaryAction>
                </ListItem>
              ))}
            </List>
          )}
        </CardContent>
      </Card>

      {/* Camera Recording States */}
      <Card>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Camera Recording States
          </Typography>
          
          <Grid container spacing={2}>
            {cameras.map((camera) => {
              const recordingState = getRecordingState(camera.device);
              const recordingError = getRecordingError(camera.device);
              const recordingProgress = getRecordingProgressData(camera.device);
              const isCurrentlyRecording = isRecording(camera.device);

              return (
                <Grid item xs={12} sm={6} md={4} key={camera.device}>
                  <Card variant="outlined">
                    <CardContent>
                      <Box display="flex" alignItems="center" justifyContent="space-between" mb={2}>
                        <Typography variant="subtitle1" fontWeight="bold">
                          {camera.device}
                        </Typography>
                        <Chip
                          icon={getStatusIcon(recordingState)}
                          label={recordingState}
                          color={getStatusColor(recordingState)}
                          size="small"
                        />
                      </Box>

                      {/* Recording Progress */}
                      {recordingProgress && (
                        <Box mb={2}>
                          <Typography variant="body2" color="textSecondary">
                            Duration: {formatDuration(recordingProgress.elapsed_time)}
                          </Typography>
                          {recordingProgress.file_size && (
                            <Typography variant="body2" color="textSecondary">
                              File Size: {recordingProgress.file_size}
                            </Typography>
                          )}
                        </Box>
                      )}

                      {/* Recording Error */}
                      {recordingError && (
                        <Alert severity="error" sx={{ mb: 2 }}>
                          <Typography variant="body2">
                            {recordingError.message || 'Recording error occurred'}
                          </Typography>
                          {recordingError.code && (
                            <Typography variant="caption" display="block">
                              Code: {recordingError.code}
                            </Typography>
                          )}
                        </Alert>
                      )}

                      {/* Recording Controls */}
                      <Box display="flex" gap={1}>
                        <Button
                          variant="contained"
                          color="primary"
                          startIcon={<PlayArrow />}
                          onClick={() => handleStartRecording(camera.device)}
                          disabled={isCurrentlyRecording || localLoading}
                          fullWidth
                        >
                          Start Recording
                        </Button>
                        <Button
                          variant="contained"
                          color="error"
                          startIcon={<Stop />}
                          onClick={() => handleStopRecording(camera.device)}
                          disabled={!isCurrentlyRecording || localLoading}
                          fullWidth
                        >
                          Stop Recording
                        </Button>
                      </Box>
                    </CardContent>
                  </Card>
                </Grid>
              );
            })}
          </Grid>
        </CardContent>
      </Card>
    </Box>
  );
};

export default RecordingManager;
