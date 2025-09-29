/**
 * @fileoverview AboutPage component for server information display
 * @author MediaMTX Development Team
 * @version 1.0.0
 */

import React from 'react';
import { Grid } from '../../components/atoms/Grid/Grid';
import { Card } from '../../components/atoms/Card/Card';
import { Alert } from '../../components/atoms/Alert/Alert';
import { Badge } from '../../components/atoms/Badge/Badge';
import {
  Info as InfoIcon,
  Storage as StorageIcon,
  MonitorHeart as HealthIcon,
} from '@mui/icons-material';
import { useServerStore } from '../../stores/server/serverStore';

/**
 * AboutPage - Server information and system status display
 *
 * Displays comprehensive server information including system status, storage details,
 * and server metadata. Provides real-time health monitoring and system metrics.
 *
 * @component
 * @returns {JSX.Element} The about page component
 *
 * @features
 * - Server information display (version, build, uptime)
 * - System status monitoring (health, performance)
 * - Storage information (usage, available space)
 * - Real-time status updates
 * - Error handling and loading states
 *
 * @example
 * ```tsx
 * <AboutPage />
 * ```
 *
 * @see {@link ../../docs/architecture/client-architechture.md} Client Architecture
 */
const AboutPage: React.FC = () => {
  const { info, status, storage, loading, error } = useServerStore();

  const formatBytes = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const formatUptime = (seconds: number): string => {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);

    if (days > 0) {
      return `${days}d ${hours}h ${minutes}m`;
    } else if (hours > 0) {
      return `${hours}h ${minutes}m`;
    } else {
      return `${minutes}m`;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'HEALTHY':
        return 'success';
      case 'DEGRADED':
        return 'warning';
      case 'UNHEALTHY':
        return 'error';
      default:
        return 'default';
    }
  };

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="400px">
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return (
      <Alert severity="error" sx={{ m: 2 }}>
        Failed to load server information: {error}
      </Alert>
    );
  }

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" component="h1" gutterBottom>
        Server Information
      </Typography>

      <Grid container spacing={3}>
        {/* Server Info */}
        <Grid size={{ xs: 12, md: 6 }}>
          <Card>
            <CardContent>
              <Box display="flex" alignItems="center" mb={2}>
                <InfoIcon sx={{ mr: 1 }} />
                <Typography variant="h6">Server Details</Typography>
              </Box>

              {info ? (
                <Box>
                  <Typography variant="body2" color="text.secondary" gutterBottom>
                    <strong>Name:</strong> {info.name}
                  </Typography>
                  <Typography variant="body2" color="text.secondary" gutterBottom>
                    <strong>Version:</strong> {info.version}
                  </Typography>
                  <Typography variant="body2" color="text.secondary" gutterBottom>
                    <strong>Build Date:</strong> {info.build_date}
                  </Typography>
                  <Typography variant="body2" color="text.secondary" gutterBottom>
                    <strong>Go Version:</strong> {info.go_version}
                  </Typography>
                  <Typography variant="body2" color="text.secondary" gutterBottom>
                    <strong>Architecture:</strong> {info.architecture}
                  </Typography>
                  <Typography variant="body2" color="text.secondary" gutterBottom>
                    <strong>Max Cameras:</strong> {info.max_cameras}
                  </Typography>

                  <Divider sx={{ my: 2 }} />

                  <Typography variant="body2" color="text.secondary" gutterBottom>
                    <strong>Capabilities:</strong>
                  </Typography>
                  <Box display="flex" flexWrap="wrap" gap={1} mb={2}>
                    {info.capabilities.map((capability) => (
                      <Chip key={capability} label={capability} size="small" />
                    ))}
                  </Box>

                  <Typography variant="body2" color="text.secondary" gutterBottom>
                    <strong>Supported Formats:</strong>
                  </Typography>
                  <Box display="flex" flexWrap="wrap" gap={1}>
                    {info.supported_formats.map((format) => (
                      <Chip key={format} label={format} size="small" variant="outlined" />
                    ))}
                  </Box>
                </Box>
              ) : (
                <Typography color="text.secondary">No server information available</Typography>
              )}
            </CardContent>
          </Card>
        </Grid>

        {/* System Status */}
        <Grid size={{ xs: 12, md: 6 }}>
          <Card>
            <CardContent>
              <Box display="flex" alignItems="center" mb={2}>
                <HealthIcon sx={{ mr: 1 }} />
                <Typography variant="h6">System Status</Typography>
              </Box>

              {status ? (
                <Box>
                  <Box display="flex" alignItems="center" mb={2}>
                    <Typography variant="body2" color="text.secondary" sx={{ mr: 1 }}>
                      <strong>Status:</strong>
                    </Typography>
                    <Chip
                      label={status.status}
                      color={
                        getStatusColor(status.status) as 'success' | 'error' | 'warning' | 'info'
                      }
                      size="small"
                    />
                  </Box>

                  <Typography variant="body2" color="text.secondary" gutterBottom>
                    <strong>Uptime:</strong> {formatUptime(status.uptime)}
                  </Typography>
                  <Typography variant="body2" color="text.secondary" gutterBottom>
                    <strong>Version:</strong> {status.version}
                  </Typography>

                  <Divider sx={{ my: 2 }} />

                  <Typography variant="body2" color="text.secondary" gutterBottom>
                    <strong>Components:</strong>
                  </Typography>
                  <Box>
                    <Typography variant="body2" color="text.secondary">
                      WebSocket Server:{' '}
                      <Chip label={status.components.websocket_server} size="small" />
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      Camera Monitor: <Chip label={status.components.camera_monitor} size="small" />
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      MediaMTX: <Chip label={status.components.mediamtx} size="small" />
                    </Typography>
                  </Box>
                </Box>
              ) : (
                <Typography color="text.secondary">No status information available</Typography>
              )}
            </CardContent>
          </Card>
        </Grid>

        {/* Storage Info */}
        {storage && (
          <Grid item xs={12}>
            <Card>
              <div className="p-4">
                <Box display="flex" alignItems="center" mb={2}>
                  <StorageIcon sx={{ mr: 1 }} />
                  <Typography variant="h6">Storage Information</Typography>
                </Box>

                <Grid container spacing={2}>
                  <Grid item xs={12} sm={6} md={3}>
                    <Typography variant="body2" color="text.secondary">
                      <strong>Total Space:</strong>
                      <br />
                      {formatBytes(storage.total_space)}
                    </Typography>
                  </Grid>
                  <Grid item xs={12} sm={6} md={3}>
                    <Typography variant="body2" color="text.secondary">
                      <strong>Used Space:</strong>
                      <br />
                      {formatBytes(storage.used_space)}
                    </Typography>
                  </Grid>
                  <Grid item xs={12} sm={6} md={3}>
                    <Typography variant="body2" color="text.secondary">
                      <strong>Available Space:</strong>
                      <br />
                      {formatBytes(storage.available_space)}
                    </Typography>
                  </Grid>
                  <Grid item xs={12} sm={6} md={3}>
                    <Typography variant="body2" color="text.secondary">
                      <strong>Usage:</strong>
                      <br />
                      {storage.usage_percentage.toFixed(1)}%
                    </Typography>
                  </Grid>
                </Grid>

                <Divider sx={{ my: 2 }} />

                <Grid container spacing={2}>
                  <Grid item xs={12} sm={6}>
                    <Typography variant="body2" color="text.secondary">
                      <strong>Recordings Size:</strong> {formatBytes(storage.recordings_size)}
                    </Typography>
                  </Grid>
                  <Grid item xs={12} sm={6}>
                    <Typography variant="body2" color="text.secondary">
                      <strong>Snapshots Size:</strong> {formatBytes(storage.snapshots_size)}
                    </Typography>
                  </Grid>
                </Grid>

                {storage.low_space_warning && (
                  <Alert severity="warning" sx={{ mt: 2 }}>
                    Low storage space warning is active
                  </Alert>
                )}
              </div>
            </Card>
          </Grid>
        )}
      </Grid>
    </Box>
  );
};

export default AboutPage;
