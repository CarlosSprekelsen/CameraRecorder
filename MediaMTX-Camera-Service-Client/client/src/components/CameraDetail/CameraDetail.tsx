import React, { useEffect, useState } from 'react';
import { useParams, Navigate } from 'react-router-dom';
import { 
  Box, 
  Typography, 
  Card, 
  CardContent, 
  Button, 
  Alert,
  CircularProgress,
  Chip,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  TextField,
  Switch,
  FormControlLabel,
  Stack,
  IconButton,
  Tooltip,
  Grid
} from '@mui/material';

import { 
  CameraAlt, 
  Videocam, 
  Stop, 
  Refresh,
  Info
} from '@mui/icons-material';
import { useCameraStore } from '../../stores/cameraStore';
import { useNotifications, notificationUtils } from '../common/NotificationSystem';
import StreamStatus from './StreamStatus';
import type { SnapshotFormat, RecordingFormat } from '../../types';

const CameraDetail: React.FC = () => {
  const { deviceId } = useParams<{ deviceId: string }>();
  const [isLoading, setIsLoading] = useState(false);
  const [localError, setLocalError] = useState<string | null>(null);
  const [snapshotFormat, setSnapshotFormat] = useState<SnapshotFormat>('jpg');
  const [snapshotQuality, setSnapshotQuality] = useState<number>(80);
  const [recordingFormat, setRecordingFormat] = useState<RecordingFormat>('mp4');
  const [recordingDuration, setRecordingDuration] = useState<number | undefined>(undefined);
  const [isUnlimitedRecording, setIsUnlimitedRecording] = useState(false);

  const { showSuccess, showError } = useNotifications();

  const {
    cameras: storeCameras,
    error: storeError,
    selectCamera: storeSelectCamera,
    takeSnapshot: storeTakeSnapshot,
    startRecording: storeStartRecording,
    stopRecording: storeStopRecording,
    getCameraStatus: storeGetCameraStatus,
    activeRecordings: storeActiveRecordings,
    isConnected: storeIsConnected,
  } = useCameraStore();

  const camera = storeCameras.find(c => c.device === deviceId);

  useEffect(() => {
    if (deviceId) {
      storeSelectCamera(deviceId);
    }
  }, [deviceId, storeSelectCamera]);

  const handleTakeSnapshot = async () => {
    if (!deviceId) return;
    
    setIsLoading(true);
    setLocalError(null);
    
    try {
      const result = await storeTakeSnapshot(deviceId, snapshotFormat, snapshotQuality);
      if (result) {
        console.log('Snapshot taken:', result);
        const notification = notificationUtils.camera.snapshotTaken(camera?.name || deviceId);
        showSuccess(notification.title, notification.message);
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to take snapshot';
      setLocalError(errorMessage);
      const notification = notificationUtils.camera.snapshotFailed(camera?.name || deviceId, errorMessage);
      showError(notification.title, notification.message);
    } finally {
      setIsLoading(false);
    }
  };

  const handleStartRecording = async () => {
    if (!deviceId) return;
    
    setIsLoading(true);
    setLocalError(null);
    
    try {
      const duration = isUnlimitedRecording ? undefined : recordingDuration;
      const result = await storeStartRecording(deviceId, duration, recordingFormat);
      if (result) {
        console.log('Recording started:', result);
        const notification = notificationUtils.camera.recordingStarted(camera?.name || deviceId);
        showSuccess(notification.title, notification.message);
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to start recording';
      setLocalError(errorMessage);
      const notification = notificationUtils.camera.recordingFailed(camera?.name || deviceId, errorMessage);
      showError(notification.title, notification.message);
    } finally {
      setIsLoading(false);
    }
  };

  const handleStopRecording = async () => {
    if (!deviceId) return;
    
    setIsLoading(true);
    setLocalError(null);
    
    try {
      const result = await storeStopRecording(deviceId);
      if (result) {
        console.log('Recording stopped:', result);
        // TODO: Show success notification
      }
    } catch (err) {
      setLocalError(err instanceof Error ? err.message : 'Failed to stop recording');
    } finally {
      setIsLoading(false);
    }
  };

  const handleRefreshCameraStatus = async () => {
    if (!deviceId) return;
    
    setIsLoading(true);
    setLocalError(null);
    
    try {
      const updatedCamera = await storeGetCameraStatus(deviceId);
      if (updatedCamera) {
        console.log('Camera status refreshed:', updatedCamera);
        // TODO: Show success notification
      }
    } catch (err) {
      setLocalError(err instanceof Error ? err.message : 'Failed to refresh camera status');
    } finally {
      setIsLoading(false);
    }
  };

  const isRecording = storeActiveRecordings.has(deviceId || '');

  if (!deviceId) {
    return <Navigate to="/" replace />;
  }

  if (!camera) {
    return (
      <Box sx={{ p: 3 }}>
        <Alert severity="warning">
          Camera not found. Please check the camera connection.
        </Alert>
      </Box>
    );
  }

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ mb: 3 }}>
        <Typography variant="h4" component="h1" gutterBottom>
          Camera: {camera.name}
        </Typography>
        <Typography variant="body1" color="text.secondary" gutterBottom>
          Device: {camera.device}
        </Typography>
        
        <Stack direction="row" spacing={2} alignItems="center" sx={{ mt: 2 }}>
          <Chip 
            label={camera.status} 
            color={camera.status === 'CONNECTED' ? 'success' : 'error'}
            icon={<Info />}
          />
          <Typography variant="body2">
            Resolution: {camera.resolution} | FPS: {camera.fps}
          </Typography>
        </Stack>
      </Box>

      {(storeError || localError) && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {storeError || localError}
        </Alert>
      )}

      <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 3 }}>
        <Grid container spacing={3}>
          {/* Camera Status */}
          <Grid item xs={12} md={6}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
                  <Typography variant="h6">
                    Camera Status
                  </Typography>
                  <Tooltip title="Refresh camera status">
                    <IconButton 
                      onClick={handleRefreshCameraStatus}
                      disabled={isLoading || !isConnected}
                      size="small"
                    >
                      <Refresh />
                    </IconButton>
                  </Tooltip>
                </Box>
                <Stack spacing={2}>
                  <Box>
                    <Typography variant="body2" color="text.secondary">
                      Status: {camera.status}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      Resolution: {camera.resolution}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      FPS: {camera.fps}
                    </Typography>
                  </Box>
                  
                  {camera.metrics && (
                    <Box>
                      <Typography variant="subtitle2" gutterBottom>
                        Metrics
                      </Typography>
                      <Typography variant="body2" color="text.secondary">
                        Bytes Sent: {camera.metrics.bytes_sent}
                      </Typography>
                      <Typography variant="body2" color="text.secondary">
                        Readers: {camera.metrics.readers}
                      </Typography>
                      <Typography variant="body2" color="text.secondary">
                        Uptime: {camera.metrics.uptime}s
                      </Typography>
                    </Box>
                  )}
                </Stack>
              </CardContent>
            </Card>
          </Grid>

          {/* Stream Status */}
          <Grid item xs={12} md={6}>
            <StreamStatus deviceId={deviceId} />
          </Grid>

          {/* Recording Status */}
          <Grid item xs={12} md={6}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  Recording Status
                </Typography>
                <Stack spacing={2}>
                  <Chip 
                    label={isRecording ? 'Recording Active' : 'Not Recording'} 
                    color={isRecording ? 'error' : 'default'}
                    icon={isRecording ? <Videocam /> : <Stop />}
                  />
                  
                  {isRecording && (
                    <Box>
                      <Typography variant="body2" color="text.secondary">
                        Recording in progress...
                      </Typography>
                    </Box>
                  )}
                </Stack>
              </CardContent>
            </Card>
          </Grid>

          {/* Snapshot Controls */}
          <Grid item xs={12} md={6}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  Snapshot Controls
                </Typography>
                <Stack spacing={2}>
                  <FormControl fullWidth>
                    <InputLabel>Format</InputLabel>
                    <Select
                      value={snapshotFormat}
                      label="Format"
                      onChange={(e) => setSnapshotFormat(e.target.value as SnapshotFormat)}
                    >
                      <MenuItem value="jpg">JPEG</MenuItem>
                      <MenuItem value="png">PNG</MenuItem>
                    </Select>
                  </FormControl>
                  
                  <TextField
                    label="Quality (1-100)"
                    type="number"
                    value={snapshotQuality}
                    onChange={(e) => setSnapshotQuality(Number(e.target.value))}
                    inputProps={{ min: 1, max: 100 }}
                    fullWidth
                  />
                  
                  <Button
                    variant="contained"
                    startIcon={<CameraAlt />}
                    onClick={handleTakeSnapshot}
                    disabled={isLoading || !isConnected || camera.status !== 'CONNECTED'}
                    fullWidth
                  >
                    {isLoading ? <CircularProgress size={20} /> : 'Take Snapshot'}
                  </Button>
                </Stack>
              </CardContent>
            </Card>
          </Grid>

          {/* Recording Controls */}
          <Grid item xs={12} md={6}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  Recording Controls
                </Typography>
                <Stack spacing={2}>
                  <FormControl fullWidth>
                    <InputLabel>Format</InputLabel>
                    <Select
                      value={recordingFormat}
                      label="Format"
                      onChange={(e) => setRecordingFormat(e.target.value as RecordingFormat)}
                    >
                      <MenuItem value="mp4">MP4</MenuItem>
                      <MenuItem value="mkv">MKV</MenuItem>
                    </Select>
                  </FormControl>
                  
                  <FormControlLabel
                    control={
                      <Switch
                        checked={isUnlimitedRecording}
                        onChange={(e) => setIsUnlimitedRecording(e.target.checked)}
                      />
                    }
                    label="Unlimited Duration"
                  />
                  
                  {!isUnlimitedRecording && (
                    <TextField
                      label="Duration (seconds)"
                      type="number"
                      value={recordingDuration || ''}
                      onChange={(e) => setRecordingDuration(Number(e.target.value) || undefined)}
                      inputProps={{ min: 1 }}
                      fullWidth
                    />
                  )}
                  
                  <Stack direction="row" spacing={1}>
                    <Button
                      variant="contained"
                      color="primary"
                      startIcon={<Videocam />}
                      onClick={handleStartRecording}
                      disabled={isLoading || isRecording || !isConnected || camera.status !== 'CONNECTED'}
                      fullWidth
                    >
                      {isLoading ? <CircularProgress size={20} /> : 'Start Recording'}
                    </Button>
                    
                    <Button
                      variant="contained"
                      color="error"
                      startIcon={<Stop />}
                      onClick={handleStopRecording}
                      disabled={isLoading || !isRecording || !isConnected}
                      fullWidth
                    >
                      {isLoading ? <CircularProgress size={20} /> : 'Stop Recording'}
                    </Button>
                  </Stack>
                </Stack>
              </CardContent>
            </Card>
          </Grid>

          {/* Stream URLs */}
          <Grid item xs={12}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  Stream URLs
                </Typography>
                <Stack spacing={1}>
                  <Typography variant="body2">
                    <strong>RTSP:</strong> {camera.streams.rtsp}
                  </Typography>
                  <Typography variant="body2">
                    <strong>WebRTC:</strong> {camera.streams.webrtc}
                  </Typography>
                  <Typography variant="body2">
                    <strong>HLS:</strong> {camera.streams.hls}
                  </Typography>
                </Stack>
              </CardContent>
            </Card>
          </Grid>
        </Grid>
      </Box>
    </Box>
  );
};

export default CameraDetail; 