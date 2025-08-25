import React from 'react';
import {
  Card,
  CardContent,
  CardActions,
  Typography,
  Chip,
  Box,
  IconButton,
  Tooltip,
  CircularProgress,
} from '@mui/material';
import {
  Videocam as CameraIcon,
  PhotoCamera as SnapshotIcon,
  FiberManualRecord as RecordIcon,
  Stop as StopIcon,
} from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import { useCameraStore } from '../../stores/cameraStore';
import type { CameraDevice } from '../../types';

interface CameraCardProps {
  camera: CameraDevice;
}

const CameraCard: React.FC<CameraCardProps> = ({ camera }) => {
  const navigate = useNavigate();
  const { 
    activeRecordings: storeActiveRecordings, 
    takeSnapshot: storeTakeSnapshot, 
    startRecording: storeStartRecording, 
    stopRecording: storeStopRecording,
    selectCamera: storeSelectCamera 
  } = useCameraStore();

  const [isSnapshotLoading, setIsSnapshotLoading] = React.useState(false);
  const [isRecordingLoading, setIsRecordingLoading] = React.useState(false);

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'connected':
        return 'success';
      case 'disconnected':
        return 'error';
      case 'recording':
        return 'warning';
      case 'capturing':
        return 'info';
      default:
        return 'default';
    }
  };

  const handleCardClick = () => {
    storeSelectCamera(camera.device);
    navigate(`/camera/${encodeURIComponent(camera.device)}`);
  };

  const handleSnapshot = async (e: React.MouseEvent) => {
    e.stopPropagation();
    setIsSnapshotLoading(true);
    
    try {
      const result = await storeTakeSnapshot(camera.device);
      if (result?.status === 'completed') {
        console.log('Snapshot taken successfully:', result);
        // TODO: Show success notification
      } else {
        console.error('Snapshot failed:', result);
        // TODO: Show error notification
      }
    } catch (error: unknown) {
      const errorMessage = error instanceof Error ? error.message : 'Snapshot failed';
      console.error('Snapshot error:', errorMessage);
      // TODO: Show error notification
    } finally {
      setIsSnapshotLoading(false);
    }
  };

  const handleRecord = async (e: React.MouseEvent) => {
    e.stopPropagation();
    setIsRecordingLoading(true);
    
    try {
      if (isRecording) {
        const result = await storeStopRecording(camera.device);
        if (result?.status === 'STOPPED') {
          console.log('Recording stopped successfully:', result);
          // TODO: Show success notification
        } else {
          console.error('Stop recording failed:', result);
          // TODO: Show error notification
        }
      } else {
        const result = await storeStartRecording(camera.device);
        if (result?.status === 'STARTED') {
          console.log('Recording started successfully:', result);
          // TODO: Show success notification
        } else {
          console.error('Start recording failed:', result);
          // TODO: Show error notification
        }
      }
    } catch (error: unknown) {
      const errorMessage = error instanceof Error ? error.message : 'Recording failed';
      console.error('Recording error:', errorMessage);
      // TODO: Show error notification
    } finally {
      setIsRecordingLoading(false);
    }
  };

  const isRecording = storeActiveRecordings.has(camera.device) || camera.status.toLowerCase() === 'recording';

  return (
    <Card
      sx={{
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        cursor: 'pointer',
        '&:hover': {
          boxShadow: 4,
        },
      }}
      onClick={handleCardClick}
    >
      <CardContent sx={{ flexGrow: 1 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
          <CameraIcon sx={{ mr: 1, color: 'primary.main' }} />
          <Typography variant="h6" component="h3" noWrap>
            {camera.name || camera.device}
          </Typography>
        </Box>

        <Typography variant="body2" color="text.secondary" gutterBottom>
          {camera.device}
        </Typography>

        <Box sx={{ mt: 2 }}>
          <Chip
            label={camera.status}
            color={getStatusColor(camera.status) as 'success' | 'error' | 'warning' | 'info' | 'default'}
            size="small"
            variant="outlined"
          />
        </Box>

        <Box sx={{ mt: 2 }}>
          <Typography variant="caption" color="text.secondary">
            Resolution: {camera.resolution}
          </Typography>
          <br />
          <Typography variant="caption" color="text.secondary">
            FPS: {camera.fps}
          </Typography>
          {camera.capabilities && (
            <>
              <br />
              <Typography variant="caption" color="text.secondary">
                Formats: {camera.capabilities.formats?.join(', ')}
              </Typography>
              <br />
              <Typography variant="caption" color="text.secondary">
                Supported Resolutions: {camera.capabilities.resolutions?.join(', ')}
              </Typography>
            </>
          )}
        </Box>

        {/* Recording Status */}
        {isRecording && (
          <Box sx={{ mt: 2 }}>
            <Chip
              label="Recording"
              color="error"
              size="small"
              icon={<RecordIcon />}
            />
          </Box>
        )}
      </CardContent>

      <CardActions sx={{ justifyContent: 'space-between', p: 2 }}>
        <Tooltip title="Take Snapshot">
          <IconButton
            size="small"
            onClick={handleSnapshot}
            color="primary"
            disabled={isSnapshotLoading}
          >
            {isSnapshotLoading ? <CircularProgress size={20} /> : <SnapshotIcon />}
          </IconButton>
        </Tooltip>

        <Tooltip title={isRecording ? 'Stop Recording' : 'Start Recording'}>
          <IconButton
            size="small"
            onClick={handleRecord}
            color={isRecording ? 'error' : 'primary'}
            disabled={isRecordingLoading}
          >
            {isRecordingLoading ? (
              <CircularProgress size={20} />
            ) : isRecording ? (
              <StopIcon />
            ) : (
              <RecordIcon />
            )}
          </IconButton>
        </Tooltip>
      </CardActions>
    </Card>
  );
};

export default CameraCard; 