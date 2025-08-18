import React from 'react';
import {
  Paper,
  Typography,
  Button,
  Box,
  Divider,
  Stack,
} from '@mui/material';
import {
  PhotoCamera as SnapshotIcon,
  FiberManualRecord as RecordIcon,
  Stop as StopIcon,
  Settings as SettingsIcon,
} from '@mui/icons-material';
import type { CameraDevice } from '../../types';

interface ControlPanelProps {
  camera: CameraDevice;
}

const ControlPanel: React.FC<ControlPanelProps> = ({ camera }) => {
  const handleSnapshot = () => {
    // TODO: Implement snapshot functionality
    console.log('Take snapshot for camera:', camera.device);
  };

  const handleRecord = () => {
    // TODO: Implement recording functionality
    console.log('Start recording for camera:', camera.device);
  };

  const handleStopRecording = () => {
    // TODO: Implement stop recording functionality
    console.log('Stop recording for camera:', camera.device);
  };

  const isRecording = camera.status.toLowerCase() === 'recording';

  return (
    <Paper sx={{ p: 3 }}>
      <Typography variant="h6" gutterBottom>
        Controls
      </Typography>
      
      <Divider sx={{ mb: 3 }} />
      
      <Stack spacing={2}>
        <Button
          variant="outlined"
          startIcon={<SnapshotIcon />}
          onClick={handleSnapshot}
          fullWidth
          disabled={camera.status.toLowerCase() !== 'connected'}
        >
          Take Snapshot
        </Button>

        {!isRecording ? (
          <Button
            variant="contained"
            startIcon={<RecordIcon />}
            onClick={handleRecord}
            fullWidth
            disabled={camera.status.toLowerCase() !== 'connected'}
            color="error"
          >
            Start Recording
          </Button>
        ) : (
          <Button
            variant="contained"
            startIcon={<StopIcon />}
            onClick={handleStopRecording}
            fullWidth
            color="error"
          >
            Stop Recording
          </Button>
        )}

        <Button
          variant="text"
          startIcon={<SettingsIcon />}
          fullWidth
        >
          Camera Settings
        </Button>
      </Stack>

      <Divider sx={{ my: 3 }} />

      <Typography variant="subtitle2" gutterBottom>
        Camera Info
      </Typography>
      
      <Box sx={{ mt: 2 }}>
        <Typography variant="body2" color="text.secondary">
          Device: {camera.device}
        </Typography>
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
    </Paper>
  );
};

export default ControlPanel; 