import React from 'react';
import {
  Paper,
  Typography,
  Box,
  Chip,
  Divider,
} from '@mui/material';
import {
  Videocam as CameraIcon,
  SignalCellularAlt as SignalIcon,
} from '@mui/icons-material';
import type { CameraDevice } from '../../types';

interface StatusDisplayProps {
  camera: CameraDevice;
}

const StatusDisplay: React.FC<StatusDisplayProps> = ({ camera }) => {
  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'connected':
        return 'success';
      case 'disconnected':
        return 'error';
      case 'recording':
        return 'warning';
      default:
        return 'default';
    }
  };

  return (
    <Paper sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', alignItems: 'center', mb: 3 }}>
        <CameraIcon sx={{ mr: 2, fontSize: 40, color: 'primary.main' }} />
        <Box>
          <Typography variant="h5" component="h2">
            {camera.name || camera.device}
          </Typography>
          <Chip
            label={camera.status}
            color={getStatusColor(camera.status) as 'success' | 'error' | 'warning' | 'info' | 'default'}
            size="small"
            sx={{ mt: 1 }}
          />
        </Box>
      </Box>

      <Divider sx={{ mb: 3 }} />

      <Box sx={{ display: 'flex', flexDirection: { xs: 'column', md: 'row' }, gap: 3 }}>
        <Box sx={{ flex: 1 }}>
          <Typography variant="h6" gutterBottom>
            Device Information
          </Typography>
          <Box sx={{ mt: 2 }}>
            <Typography variant="body2" color="text.secondary">
              Device Path: {camera.device}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Status: {camera.status}
            </Typography>
            {camera.capabilities && (
              <>
                <Typography variant="body2" color="text.secondary">
                  Resolution: {camera.capabilities.resolution}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  FPS: {camera.capabilities.fps}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  Formats: {camera.capabilities.formats?.join(', ')}
                </Typography>
              </>
            )}
          </Box>
        </Box>

        <Box sx={{ flex: 1 }}>
          <Typography variant="h6" gutterBottom>
            Stream Information
          </Typography>
          <Box sx={{ mt: 2 }}>
            {camera.streams && (
              <>
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                  <SignalIcon sx={{ mr: 1, fontSize: 16 }} />
                  <Typography variant="body2" color="text.secondary">
                    RTSP: {camera.streams.rtsp}
                  </Typography>
                </Box>
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                  <SignalIcon sx={{ mr: 1, fontSize: 16 }} />
                  <Typography variant="body2" color="text.secondary">
                    WebRTC: {camera.streams.webrtc}
                  </Typography>
                </Box>
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <SignalIcon sx={{ mr: 1, fontSize: 16 }} />
                  <Typography variant="body2" color="text.secondary">
                    HLS: {camera.streams.hls}
                  </Typography>
                </Box>
              </>
            )}
          </Box>
        </Box>
      </Box>

      {camera.metrics && (
        <>
          <Divider sx={{ my: 3 }} />
          <Typography variant="h6" gutterBottom>
            Performance Metrics
          </Typography>
          <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 2 }}>
            <Box sx={{ flex: '1 1 150px', textAlign: 'center' }}>
              <Typography variant="h4" color="primary">
                {camera.metrics.bytes_sent || 'N/A'}
              </Typography>
              <Typography variant="caption" color="text.secondary">
                Bytes Sent
              </Typography>
            </Box>
            <Box sx={{ flex: '1 1 150px', textAlign: 'center' }}>
              <Typography variant="h4" color="primary">
                {camera.metrics.readers || 'N/A'}
              </Typography>
              <Typography variant="caption" color="text.secondary">
                Active Readers
              </Typography>
            </Box>
            <Box sx={{ flex: '1 1 150px', textAlign: 'center' }}>
              <Typography variant="h4" color="primary">
                {camera.metrics.uptime || 'N/A'}
              </Typography>
              <Typography variant="caption" color="text.secondary">
                Uptime (s)
              </Typography>
            </Box>
          </Box>
        </>
      )}
    </Paper>
  );
};

export default StatusDisplay; 