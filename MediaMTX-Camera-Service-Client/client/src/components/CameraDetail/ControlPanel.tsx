import React, { useState } from 'react';
import {
  Paper,
  Typography,
  Button,
  Box,
  Divider,
  Stack,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Slider,
  TextField,
  Alert,
  CircularProgress,
} from '@mui/material';
import {
  PhotoCamera as SnapshotIcon,
  FiberManualRecord as RecordIcon,
  Stop as StopIcon,
  Settings as SettingsIcon,
} from '@mui/icons-material';
import type { CameraDevice, SnapshotFormat } from '../../types';
import { useCameraStore } from '../../stores/cameraStore';

interface ControlPanelProps {
  camera: CameraDevice;
}

interface SnapshotDialogProps {
  open: boolean;
  onClose: () => void;
  onTakeSnapshot: (format: SnapshotFormat, quality: number, filename?: string) => void;
  loading: boolean;
}

interface RecordingDialogProps {
  open: boolean;
  onClose: () => void;
  onStartRecording: (duration: number, format: string) => void;
  loading: boolean;
}

const SnapshotDialog: React.FC<SnapshotDialogProps> = ({
  open,
  onClose,
  onTakeSnapshot,
  loading
}) => {
  const [format, setFormat] = useState<SnapshotFormat>('jpg');
  const [quality, setQuality] = useState<number>(85);
  const [filename, setFilename] = useState<string>('');

  const handleTakeSnapshot = () => {
    onTakeSnapshot(format, quality, filename || undefined);
  };

  const handleClose = () => {
    if (!loading) {
      setFormat('jpg');
      setQuality(85);
      setFilename('');
      onClose();
    }
  };

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle>Take Snapshot</DialogTitle>
      <DialogContent>
        <Stack spacing={3} sx={{ mt: 1 }}>
          <FormControl fullWidth>
            <InputLabel>Format</InputLabel>
            <Select
              value={format}
              label="Format"
              onChange={(e) => setFormat(e.target.value as SnapshotFormat)}
              disabled={loading}
            >
              <MenuItem value="jpg">JPEG</MenuItem>
              <MenuItem value="png">PNG</MenuItem>
            </Select>
          </FormControl>

          <Box>
            <Typography gutterBottom>Quality: {quality}%</Typography>
            <Slider
              value={quality}
              onChange={(_, value) => setQuality(value as number)}
              min={1}
              max={100}
              marks={[
                { value: 1, label: '1%' },
                { value: 50, label: '50%' },
                { value: 85, label: '85%' },
                { value: 100, label: '100%' },
              ]}
              disabled={loading}
            />
          </Box>

          <TextField
            label="Custom Filename (optional)"
            value={filename}
            onChange={(e) => setFilename(e.target.value)}
            placeholder="Leave empty for auto-generated filename"
            fullWidth
            disabled={loading}
          />
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose} disabled={loading}>
          Cancel
        </Button>
        <Button
          onClick={handleTakeSnapshot}
          variant="contained"
          disabled={loading}
          startIcon={loading ? <CircularProgress size={20} /> : <SnapshotIcon />}
        >
          {loading ? 'Taking Snapshot...' : 'Take Snapshot'}
        </Button>
      </DialogActions>
    </Dialog>
  );
};

const RecordingDialog: React.FC<RecordingDialogProps> = ({
  open,
  onClose,
  onStartRecording,
  loading
}) => {
  const [duration, setDuration] = useState<number>(60);
  const [format, setFormat] = useState<string>('mp4');
  const [durationType, setDurationType] = useState<'seconds' | 'minutes' | 'hours' | 'unlimited'>('seconds');

  const handleStartRecording = () => {
    const actualDuration = durationType === 'unlimited' ? 0 : duration;
    onStartRecording(actualDuration, format);
  };

  const handleClose = () => {
    if (!loading) {
      setDuration(60);
      setFormat('mp4');
      setDurationType('seconds');
      onClose();
    }
  };

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle>Start Recording</DialogTitle>
      <DialogContent>
        <Stack spacing={3} sx={{ mt: 1 }}>
          <FormControl fullWidth>
            <InputLabel>Duration Type</InputLabel>
            <Select
              value={durationType}
              label="Duration Type"
              onChange={(e) => setDurationType(e.target.value as any)}
              disabled={loading}
            >
              <MenuItem value="seconds">Seconds</MenuItem>
              <MenuItem value="minutes">Minutes</MenuItem>
              <MenuItem value="hours">Hours</MenuItem>
              <MenuItem value="unlimited">Unlimited</MenuItem>
            </Select>
          </FormControl>

          {durationType !== 'unlimited' && (
            <Box>
              <Typography gutterBottom>
                Duration: {duration} {durationType}
              </Typography>
              <Slider
                value={duration}
                onChange={(_, value) => setDuration(value as number)}
                min={1}
                max={durationType === 'seconds' ? 3600 : durationType === 'minutes' ? 1440 : 24}
                marks={[
                  { value: 1, label: '1' },
                  { value: durationType === 'seconds' ? 1800 : durationType === 'minutes' ? 720 : 12, label: durationType === 'seconds' ? '30m' : durationType === 'minutes' ? '12h' : '12h' },
                  { value: durationType === 'seconds' ? 3600 : durationType === 'minutes' ? 1440 : 24, label: durationType === 'seconds' ? '1h' : durationType === 'minutes' ? '24h' : '24h' },
                ]}
                disabled={loading}
              />
            </Box>
          )}

          <FormControl fullWidth>
            <InputLabel>Format</InputLabel>
            <Select
              value={format}
              label="Format"
              onChange={(e) => setFormat(e.target.value)}
              disabled={loading}
            >
              <MenuItem value="mp4">MP4</MenuItem>
              <MenuItem value="avi">AVI</MenuItem>
              <MenuItem value="mkv">MKV</MenuItem>
            </Select>
          </FormControl>
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose} disabled={loading}>
          Cancel
        </Button>
        <Button
          onClick={handleStartRecording}
          variant="contained"
          disabled={loading}
          startIcon={loading ? <CircularProgress size={20} /> : <RecordIcon />}
          color="error"
        >
          {loading ? 'Starting Recording...' : `Start Recording${durationType !== 'unlimited' ? ` (${duration} ${durationType})` : ' (Unlimited)'}`}
        </Button>
      </DialogActions>
    </Dialog>
  );
};

const ControlPanel: React.FC<ControlPanelProps> = ({ camera }) => {
  const { takeSnapshot, startRecording, stopRecording, error, clearError, activeRecordings } = useCameraStore();
  const [snapshotDialogOpen, setSnapshotDialogOpen] = useState(false);
  const [recordingDialogOpen, setRecordingDialogOpen] = useState(false);
  const [snapshotLoading, setSnapshotLoading] = useState(false);
  const [recordingLoading, setRecordingLoading] = useState(false);
  const [snapshotResult, setSnapshotResult] = useState<string | null>(null);
  const [recordingResult, setRecordingResult] = useState<string | null>(null);

  const handleSnapshot = () => {
    setSnapshotDialogOpen(true);
    setSnapshotResult(null);
    clearError();
  };

  const handleTakeSnapshot = async (format: SnapshotFormat, quality: number, filename?: string) => {
    setSnapshotLoading(true);
    clearError();
    
    try {
      const result = await takeSnapshot(camera.device, format, quality, filename);
      
      if (result) {
        if (result.status === 'completed') {
          setSnapshotResult(`✅ Snapshot saved: ${result.filename} (${result.file_size} bytes)`);
          // Close dialog after successful snapshot
          setTimeout(() => {
            setSnapshotDialogOpen(false);
            setSnapshotResult(null);
          }, 2000);
        } else {
          setSnapshotResult(`❌ Snapshot failed: ${result.error || 'Unknown error'}`);
        }
      } else {
        setSnapshotResult('❌ Snapshot failed: No response from server');
      }
    } catch (error) {
      setSnapshotResult(`❌ Snapshot failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
    } finally {
      setSnapshotLoading(false);
    }
  };

  const handleRecord = () => {
    setRecordingDialogOpen(true);
    setRecordingResult(null);
    clearError();
  };

  const handleStartRecording = async (duration: number, format: string) => {
    setRecordingLoading(true);
    clearError();
    
    try {
      const result = await startRecording(camera.device, duration, format);
      
      if (result) {
        setRecordingResult(`✅ Recording started: ${result.filename} (Session ID: ${result.session_id})`);
        // Close dialog after successful recording start
        setTimeout(() => {
          setRecordingDialogOpen(false);
          setRecordingResult(null);
        }, 2000);
      } else {
        setRecordingResult('❌ Recording failed: No response from server');
      }
    } catch (error) {
      setRecordingResult(`❌ Recording failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
    } finally {
      setRecordingLoading(false);
    }
  };

  const handleStopRecording = async () => {
    setRecordingLoading(true);
    clearError();
    
    try {
      const result = await stopRecording(camera.device);
      
      if (result) {
        const fileSize = result.file_size ? `${result.file_size} bytes` : 'unknown size';
        const duration = result.duration ? `${result.duration}s` : 'unknown duration';
        setRecordingResult(`✅ Recording stopped: ${result.filename} (${fileSize}, ${duration})`);
        // Clear result after successful stop
        setTimeout(() => {
          setRecordingResult(null);
        }, 3000);
      } else {
        setRecordingResult('❌ Stop recording failed: No response from server');
      }
    } catch (error) {
      setRecordingResult(`❌ Stop recording failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
    } finally {
      setRecordingLoading(false);
    }
  };

  const isRecording = camera.status.toLowerCase() === 'recording' || activeRecordings.has(camera.device);
  const isConnected = camera.status.toLowerCase() === 'connected';

  return (
    <>
      <Paper sx={{ p: 3 }}>
        <Typography variant="h6" gutterBottom>
          Controls
        </Typography>
        
        <Divider sx={{ mb: 3 }} />
        
        {error && (
          <Alert severity="error" sx={{ mb: 2 }} onClose={clearError}>
            {error}
          </Alert>
        )}

        {snapshotResult && (
          <Alert 
            severity={snapshotResult.includes('✅') ? 'success' : 'error'} 
            sx={{ mb: 2 }}
            onClose={() => setSnapshotResult(null)}
          >
            {snapshotResult}
          </Alert>
        )}

        {recordingResult && (
          <Alert 
            severity={recordingResult.includes('✅') ? 'success' : 'error'} 
            sx={{ mb: 2 }}
            onClose={() => setRecordingResult(null)}
          >
            {recordingResult}
          </Alert>
        )}
        
        <Stack spacing={2}>
          <Button
            variant="outlined"
            startIcon={<SnapshotIcon />}
            onClick={handleSnapshot}
            fullWidth
            disabled={!isConnected}
          >
            Take Snapshot
          </Button>

          {!isRecording ? (
            <Button
              variant="contained"
              startIcon={<RecordIcon />}
              onClick={handleRecord}
              fullWidth
              disabled={!isConnected || recordingLoading}
              color="error"
            >
              {recordingLoading ? 'Starting...' : 'Start Recording'}
            </Button>
          ) : (
            <Button
              variant="contained"
              startIcon={<StopIcon />}
              onClick={handleStopRecording}
              fullWidth
              disabled={recordingLoading}
              color="error"
            >
              {recordingLoading ? 'Stopping...' : 'Stop Recording'}
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

      <SnapshotDialog
        open={snapshotDialogOpen}
        onClose={() => setSnapshotDialogOpen(false)}
        onTakeSnapshot={handleTakeSnapshot}
        loading={snapshotLoading}
      />

      <RecordingDialog
        open={recordingDialogOpen}
        onClose={() => setRecordingDialogOpen(false)}
        onStartRecording={handleStartRecording}
        loading={recordingLoading}
      />
    </>
  );
};

export default ControlPanel; 