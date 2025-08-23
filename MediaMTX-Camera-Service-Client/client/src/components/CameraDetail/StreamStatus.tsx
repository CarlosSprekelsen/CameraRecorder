/**
 * Stream Status Component
 * Displays MediaMTX stream information for cameras
 * Aligned with server get_streams method
 */

import React, { useEffect, useState } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Chip,
  Grid,
  LinearProgress,
  Tooltip,
  IconButton,
  Alert,
} from '@mui/material';
import {
  PlayArrow as ActiveIcon,
  Stop as InactiveIcon,
  Refresh as RefreshIcon,
  Videocam as StreamIcon,
  People as ViewersIcon,
  Storage as DataIcon,
} from '@mui/icons-material';
import { useCameraStore } from '../../stores/cameraStore';
import type { StreamInfo } from '../../types';

/**
 * Stream Status Component Props
 */
interface StreamStatusProps {
  deviceId: string;
  autoRefresh?: boolean;
  refreshInterval?: number;
}

/**
 * Stream Status Component
 */
const StreamStatus: React.FC<StreamStatusProps> = ({
  deviceId,
  autoRefresh = true,
  refreshInterval = 10000, // 10 seconds
}) => {
  const { streams: storeStreams, getStreams: storeGetStreams, isLoading: storeIsLoading, error: storeError } = useCameraStore();
  const [isRefreshing, setIsRefreshing] = useState(false);

  // Find stream for this device
  const deviceStream = storeStreams.find(stream => stream.name === deviceId);

  /**
   * Refresh stream data
   */
  const refreshStreams = async () => {
    if (isRefreshing) return;
    
    setIsRefreshing(true);
    try {
      await storeGetStreams();
    } catch (error) {
      console.error('Failed to refresh streams:', error);
    } finally {
      setIsRefreshing(false);
    }
  };

  /**
   * Auto-refresh streams
   */
  useEffect(() => {
    if (!autoRefresh) return;

    // Initial load
    refreshStreams();

    // Set up interval
    const interval = setInterval(refreshStreams, refreshInterval);
    return () => clearInterval(interval);
  }, [autoRefresh, refreshInterval, deviceId]);

  /**
   * Format bytes to human readable
   */
  const formatBytes = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`;
  };

  /**
   * Get stream status color
   */
  const getStatusColor = (ready: boolean): 'success' | 'error' | 'warning' => {
    if (ready) return 'success';
    return 'error';
  };

  /**
   * Get stream status icon
   */
  const getStatusIcon = (ready: boolean) => {
    return ready ? <ActiveIcon /> : <InactiveIcon />;
  };

  if (storeIsLoading && !deviceStream) {
    return (
      <Card>
        <CardContent>
          <Box display="flex" alignItems="center" gap={1} mb={2}>
            <StreamIcon />
            <Typography variant="h6">Stream Status</Typography>
          </Box>
          <LinearProgress />
        </CardContent>
      </Card>
    );
  }

  if (storeError && !deviceStream) {
    return (
      <Card>
        <CardContent>
          <Box display="flex" alignItems="center" gap={1} mb={2}>
            <StreamIcon />
            <Typography variant="h6">Stream Status</Typography>
          </Box>
          <Alert severity="error">
            Failed to load stream information: {storeError}
          </Alert>
        </CardContent>
      </Card>
    );
  }

  if (!deviceStream) {
    return (
      <Card>
        <CardContent>
          <Box display="flex" alignItems="center" justifyContent="space-between" mb={2}>
            <Box display="flex" alignItems="center" gap={1}>
              <StreamIcon />
              <Typography variant="h6">Stream Status</Typography>
            </Box>
            <Tooltip title="Refresh streams">
              <IconButton 
                onClick={refreshStreams} 
                disabled={isRefreshing}
                size="small"
              >
                <RefreshIcon />
              </IconButton>
            </Tooltip>
          </Box>
          <Alert severity="info">
            No stream information available for this camera
          </Alert>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardContent>
        <Box display="flex" alignItems="center" justifyContent="space-between" mb={2}>
          <Box display="flex" alignItems="center" gap={1}>
            <StreamIcon />
            <Typography variant="h6">Stream Status</Typography>
          </Box>
          <Tooltip title="Refresh streams">
            <IconButton 
              onClick={refreshStreams} 
              disabled={isRefreshing}
              size="small"
            >
              <RefreshIcon />
            </IconButton>
          </Tooltip>
        </Box>

        <Grid container spacing={2}>
          {/* Stream Status */}
          <Grid item xs={12} sm={6}>
            <Box display="flex" alignItems="center" gap={1} mb={1}>
              {getStatusIcon(deviceStream.ready)}
              <Typography variant="body2" color="text.secondary">
                Status
              </Typography>
            </Box>
            <Chip
              label={deviceStream.ready ? 'Active' : 'Inactive'}
              color={getStatusColor(deviceStream.ready)}
              size="small"
            />
          </Grid>

          {/* Viewers */}
          <Grid item xs={12} sm={6}>
            <Box display="flex" alignItems="center" gap={1} mb={1}>
              <ViewersIcon fontSize="small" />
              <Typography variant="body2" color="text.secondary">
                Viewers
              </Typography>
            </Box>
            <Typography variant="h6">
              {deviceStream.readers}
            </Typography>
          </Grid>

          {/* Data Sent */}
          <Grid item xs={12}>
            <Box display="flex" alignItems="center" gap={1} mb={1}>
              <DataIcon fontSize="small" />
              <Typography variant="body2" color="text.secondary">
                Data Sent
              </Typography>
            </Box>
            <Typography variant="h6">
              {formatBytes(deviceStream.bytes_sent)}
            </Typography>
          </Grid>

          {/* Stream Source */}
          <Grid item xs={12}>
            <Typography variant="body2" color="text.secondary" gutterBottom>
              Stream Source
            </Typography>
            <Typography 
              variant="body2" 
              sx={{ 
                fontFamily: 'monospace',
                backgroundColor: 'grey.100',
                padding: 1,
                borderRadius: 1,
                wordBreak: 'break-all'
              }}
            >
              {deviceStream.source}
            </Typography>
          </Grid>
        </Grid>

        {/* Stream Metrics */}
        {deviceStream.readers > 0 && (
          <Box mt={2}>
            <Typography variant="body2" color="text.secondary" gutterBottom>
              Stream Activity
            </Typography>
            <LinearProgress 
              variant="determinate" 
              value={Math.min((deviceStream.readers / 10) * 100, 100)} 
              color="primary"
            />
            <Typography variant="caption" color="text.secondary">
              {deviceStream.readers} active viewer{deviceStream.readers !== 1 ? 's' : ''}
            </Typography>
          </Box>
        )}
      </CardContent>
    </Card>
  );
};

export default StreamStatus;
