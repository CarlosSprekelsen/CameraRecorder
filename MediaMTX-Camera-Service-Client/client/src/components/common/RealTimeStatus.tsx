import React, { useEffect, useState } from 'react';
import { 
  Box, 
  Typography, 
  LinearProgress, 
  Chip, 
  Alert, 
  Paper
} from '@mui/material';
import { 
  Wifi, 
  WifiOff, 
  Camera, 
  Videocam, 
  Refresh, 
  CheckCircle, 
  Error, 
  Warning 
} from '@mui/icons-material';
import { useConnectionStore } from '../../stores/connectionStore';
import { useCameraStore } from '../../stores/cameraStore';


interface RealTimeStatusProps {
  showDetails?: boolean;
  showRecordingProgress?: boolean;
  showConnectionMetrics?: boolean;
}

const RealTimeStatus: React.FC<RealTimeStatusProps> = ({
  showDetails = false,
  showRecordingProgress = true,
  showConnectionMetrics = false
}) => {
    const {
    status,
    isConnected,
    isConnecting,
    connectionQuality,
    healthScore,
    notificationCount,
    averageNotificationLatency,
    lastNotificationTime,
    realTimeUpdatesEnabled,
    componentSyncStatus,
    updateComponentSyncStatus
  } = useConnectionStore();

  const {
    cameras,
    activeRecordings,
    recordingProgress,
    notificationCount: cameraNotificationCount,
    realTimeUpdatesEnabled: cameraRealTimeEnabled
  } = useCameraStore();

  const [lastUpdate, setLastUpdate] = useState<Date>(new Date());

  // Update component sync status
  useEffect(() => {
    updateComponentSyncStatus('real-time-status', true);
    return () => updateComponentSyncStatus('real-time-status', false);
  }, [updateComponentSyncStatus]);

  // Update last update time when notifications are received
  useEffect(() => {
    if (notificationCount > 0 || cameraNotificationCount > 0) {
      setLastUpdate(new Date());
    }
  }, [notificationCount, cameraNotificationCount]);

  // Note: handleRecordingStatusUpdate is handled by the connection store
  // This callback is not needed in this component

  const getConnectionStatusColor = () => {
    switch (status) {
      case 'connected':
        return connectionQuality === 'excellent' ? 'success' : 
               connectionQuality === 'good' ? 'primary' : 'warning';
      case 'connecting':
        return 'info';
      case 'disconnected':
        return 'error';
      default:
        return 'default';
    }
  };

  const getConnectionStatusIcon = () => {
    if (isConnecting) return <Refresh sx={{ animation: 'spin 1s linear infinite' }} />;
    if (isConnected) return <Wifi />;
    return <WifiOff />;
  };

  const getHealthScoreColor = () => {
    if (healthScore >= 90) return 'success';
    if (healthScore >= 70) return 'primary';
    if (healthScore >= 30) return 'warning';
    return 'error';
  };

  const formatLatency = (latency: number) => {
    if (latency < 1000) return `${latency.toFixed(1)}ms`;
    return `${(latency / 1000).toFixed(2)}s`;
  };

  const getActiveRecordingsCount = () => {
    return activeRecordings.size;
  };

  const getConnectedCamerasCount = () => {
    return cameras.filter(camera => camera.status === 'CONNECTED').length;
  };

  return (
    <Paper sx={{ p: 2, mb: 2 }}>
      <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 2 }}>
        <Typography variant="h6" component="h3">
          Real-Time Status
        </Typography>
        <Box sx={{ display: 'flex', gap: 1, alignItems: 'center' }}>
          <Chip
            icon={getConnectionStatusIcon()}
            label={status.toUpperCase()}
            color={getConnectionStatusColor()}
            size="small"
          />
          {realTimeUpdatesEnabled && (
            <Chip
              label="LIVE"
              color="success"
              size="small"
              variant="outlined"
            />
          )}
        </Box>
      </Box>

      {/* Connection Health */}
      <Box sx={{ mb: 2 }}>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1 }}>
          <Typography variant="body2" color="text.secondary">
            Connection Health
          </Typography>
          <Typography variant="body2" color={`${getHealthScoreColor()}.main`}>
            {healthScore}%
          </Typography>
        </Box>
        <LinearProgress
          variant="determinate"
          value={healthScore}
          color={getHealthScoreColor()}
          sx={{ height: 6, borderRadius: 3 }}
        />
      </Box>

      {/* Real-time Updates Status */}
      <Box sx={{ mb: 2 }}>
        <Typography variant="body2" color="text.secondary" gutterBottom>
          Real-Time Updates
        </Typography>
        <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
          <Chip
            icon={realTimeUpdatesEnabled ? <CheckCircle /> : <Error />}
            label={realTimeUpdatesEnabled ? 'Enabled' : 'Disabled'}
            color={realTimeUpdatesEnabled ? 'success' : 'error'}
            size="small"
          />
          <Chip
            icon={cameraRealTimeEnabled ? <CheckCircle /> : <Error />}
            label="Camera Updates"
            color={cameraRealTimeEnabled ? 'success' : 'error'}
            size="small"
          />
          {notificationCount > 0 && (
            <Chip
              label={`${notificationCount} updates`}
              color="info"
              size="small"
              variant="outlined"
            />
          )}
        </Box>
      </Box>

      {/* Camera Status */}
      <Box sx={{ mb: 2 }}>
        <Typography variant="body2" color="text.secondary" gutterBottom>
          Camera Status
        </Typography>
        <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
          <Chip
            icon={<Camera />}
            label={`${getConnectedCamerasCount()}/${cameras.length} Connected`}
            color={getConnectedCamerasCount() > 0 ? 'success' : 'warning'}
            size="small"
          />
          {getActiveRecordingsCount() > 0 && (
            <Chip
              icon={<Videocam />}
              label={`${getActiveRecordingsCount()} Recording`}
              color="error"
              size="small"
            />
          )}
        </Box>
      </Box>

      {/* Recording Progress */}
      {showRecordingProgress && getActiveRecordingsCount() > 0 && (
        <Box sx={{ mb: 2 }}>
          <Typography variant="body2" color="text.secondary" gutterBottom>
            Recording Progress
          </Typography>
          {Array.from(activeRecordings.entries()).map(([device, _recording]) => {
            const progress = recordingProgress.get(device) || 0;
            return (
              <Box key={device} sx={{ mb: 1 }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 0.5 }}>
                  <Typography variant="caption" color="text.secondary">
                    {device}
                  </Typography>
                  <Typography variant="caption" color="text.secondary">
                    {progress.toFixed(1)}%
                  </Typography>
                </Box>
                <LinearProgress
                  variant="determinate"
                  value={progress}
                  color="error"
                  sx={{ height: 4, borderRadius: 2 }}
                />
              </Box>
            );
          })}
        </Box>
      )}

      {/* Connection Metrics */}
      {showConnectionMetrics && (
        <Box sx={{ mb: 2 }}>
          <Typography variant="body2" color="text.secondary" gutterBottom>
            Connection Metrics
          </Typography>
          <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
            {averageNotificationLatency > 0 && (
              <Chip
                label={`${formatLatency(averageNotificationLatency)} avg latency`}
                color="info"
                size="small"
                variant="outlined"
              />
            )}
            {lastNotificationTime && (
              <Chip
                label={`Last: ${lastNotificationTime.toLocaleTimeString()}`}
                color="info"
                size="small"
                variant="outlined"
              />
            )}
          </Box>
        </Box>
      )}

      {/* Component Sync Status */}
      {showDetails && (
        <Box sx={{ mb: 2 }}>
          <Typography variant="body2" color="text.secondary" gutterBottom>
            Component Sync Status
          </Typography>
          <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
            {Array.from(componentSyncStatus.entries()).map(([componentId, synced]) => (
              <Chip
                key={componentId}
                icon={synced ? <CheckCircle /> : <Warning />}
                label={componentId}
                color={synced ? 'success' : 'warning'}
                size="small"
                variant="outlined"
              />
            ))}
          </Box>
        </Box>
      )}

      {/* Last Update */}
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Typography variant="caption" color="text.secondary">
          Last update: {lastUpdate.toLocaleTimeString()}
        </Typography>
        <Typography variant="caption" color="text.secondary">
          Quality: {connectionQuality}
        </Typography>
      </Box>

      {/* Alerts */}
      {!realTimeUpdatesEnabled && (
        <Alert severity="warning" sx={{ mt: 2 }}>
          Real-time updates are disabled. Some features may not work correctly.
        </Alert>
      )}

      {status === 'error' && (
        <Alert severity="error" sx={{ mt: 2 }}>
          Connection error detected. Please check your network connection.
        </Alert>
      )}
    </Paper>
  );
};

export default RealTimeStatus;
