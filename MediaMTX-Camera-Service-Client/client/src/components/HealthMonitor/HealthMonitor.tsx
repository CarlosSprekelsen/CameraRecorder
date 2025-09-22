/**
 * Health Monitor Component
 * Displays system health status and component health monitoring
 * Aligned with server health endpoints API
 * 
 * Server API Reference: ../mediamtx-camera-service/docs/api/health-endpoints.md
 */

import React, { useEffect, useCallback } from 'react';
import { logger, loggers } from '../../services/loggerService';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Grid,
  Chip,
  LinearProgress,
  Alert,
  IconButton,
  Tooltip,
  Collapse,
} from '@mui/material';
import {
  CheckCircle as HealthyIcon,
  Warning as DegradedIcon,
  Error as UnhealthyIcon,
  Refresh as RefreshIcon,
  ExpandMore as ExpandMoreIcon,
  ExpandLess as ExpandLessIcon,
} from '@mui/icons-material';
import { useHealthStore } from '../../stores/connection/healthStore';

/**
 * Health status color mapping
 */
const getHealthColor = (status: string): 'success' | 'warning' | 'error' => {
  switch (status) {
    case 'healthy':
      return 'success';
    case 'degraded':
      return 'warning';
    case 'unhealthy':
      return 'error';
    default:
      return 'error';
  }
};

/**
 * Health status icon mapping
 */
const getHealthIcon = (status: string) => {
  switch (status) {
    case 'healthy':
      return <HealthyIcon color="success" />;
    case 'degraded':
      return <DegradedIcon color="warning" />;
    case 'unhealthy':
      return <UnhealthyIcon color="error" />;
    default:
      return <UnhealthyIcon color="error" />;
  }
};

/**
 * Health Monitor Component Props
 */
interface HealthMonitorProps {
  autoRefresh?: boolean;
  refreshInterval?: number;
  showDetails?: boolean;
}

/**
 * Health Monitor Component
 */
const HealthMonitor: React.FC<HealthMonitorProps> = ({
  autoRefresh = true,
  refreshInterval = 30000, // 30 seconds
  showDetails = true,
}) => {
  const {
    isHealthy: storeIsHealthy,
    healthScore: storeHealthScore,
    connectionQuality: storeConnectionQuality,
    latency: storeLatency,
    refreshHealth: storeRefreshHealth,
  } = useHealthStore();

  const [expanded, setExpanded] = React.useState(false);
  const [isRefreshing, setIsRefreshing] = React.useState(false);

  /**
   * Refresh health data using WebSocket
   */
  const refreshHealth = useCallback(async () => {
    if (isRefreshing) return;

    setIsRefreshing(true);
    try {
      await storeRefreshHealth();
    } catch (error: unknown) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to refresh health data';
      logger.error('Failed to refresh health data', { errorMessage }, 'healthMonitor');
    } finally {
      setIsRefreshing(false);
    }
  }, [isRefreshing, storeRefreshHealth]);

  /**
   * Start health monitoring
   */
  useEffect(() => {
    if (autoRefresh) {
      // Initial load
      refreshHealth();

      // Set up periodic refresh
      const interval = setInterval(refreshHealth, refreshInterval);

      return () => {
        clearInterval(interval);
      };
    }
  }, [autoRefresh, refreshInterval, refreshHealth]);

  /**
   * Get overall health status from new health store
   */
  const overallHealth = storeIsHealthy ? 'healthy' : 'unhealthy';
  const healthScoreValue = storeHealthScore;
  const systemReady = storeIsHealthy;

  /**
   * Format timestamp
   */
  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString();
  };

  /**
   * Toggle expanded view
   */
  const handleToggleExpanded = () => {
    setExpanded(!expanded);
  };

  return (
    <Box>
      {/* Health Overview Card */}
      <Card sx={{ mb: 2 }}>
        <CardContent>
          <Box display="flex" alignItems="center" justifyContent="space-between" mb={2}>
            <Typography variant="h6" component="h2">
              System Health
            </Typography>
            <Box display="flex" alignItems="center" gap={1}>
              <Tooltip title="Refresh health data">
                <IconButton
                  onClick={refreshHealth}
                  disabled={isRefreshing}
                  size="small"
                >
                  <RefreshIcon />
                </IconButton>
              </Tooltip>
              {showDetails && (
                <Tooltip title={expanded ? 'Hide details' : 'Show details'}>
                  <IconButton onClick={handleToggleExpanded} size="small">
                    {expanded ? <ExpandLessIcon /> : <ExpandMoreIcon />}
                  </IconButton>
                </Tooltip>
              )}
            </Box>
          </Box>

          {/* Health Score */}
          <Box display="flex" alignItems="center" gap={2} mb={2}>
            {getHealthIcon(overallHealth)}
            <Box flex={1}>
              <Typography variant="body2" color="text.secondary">
                Health Score
              </Typography>
              <LinearProgress
                variant="determinate"
                value={healthScoreValue}
                color={getHealthColor(overallHealth)}
                sx={{ height: 8, borderRadius: 4 }}
              />
            </Box>
            <Typography variant="h6" color={getHealthColor(overallHealth)}>
              {healthScoreValue}%
            </Typography>
          </Box>

          {/* Status Summary */}
          <Grid container spacing={2}>
            <Grid item xs={12} sm={6} md={3}>
              <Box textAlign="center">
                <Typography variant="body2" color="text.secondary">
                  Overall Status
                </Typography>
                <Chip
                  label={overallHealth}
                  color={getHealthColor(overallHealth)}
                  size="small"
                  sx={{ mt: 0.5 }}
                />
              </Box>
            </Grid>
            <Grid item xs={12} sm={6} md={3}>
              <Box textAlign="center">
                <Typography variant="body2" color="text.secondary">
                  System Ready
                </Typography>
                <Chip
                  label={systemReady ? 'Ready' : 'Not Ready'}
                  color={systemReady ? 'success' : 'error'}
                  size="small"
                  sx={{ mt: 0.5 }}
                />
              </Box>
            </Grid>
            <Grid item xs={12} sm={6} md={3}>
              <Box textAlign="center">
                <Typography variant="body2" color="text.secondary">
                  Monitoring
                </Typography>
                <Chip
                  label={storeIsHealthy ? 'Active' : 'Inactive'}
                  color={storeIsHealthy ? 'success' : 'default'}
                  size="small"
                  sx={{ mt: 0.5 }}
                />
              </Box>
            </Grid>
            <Grid item xs={12} sm={6} md={3}>
              <Box textAlign="center">
                <Typography variant="body2" color="text.secondary">
                  Last Update
                </Typography>
                <Typography variant="caption" display="block">
                  {storeLatency ? `${storeLatency}ms` : 'Unknown'}
                </Typography>
              </Box>
            </Grid>
          </Grid>
        </CardContent>
      </Card>

      {/* Detailed Health Information */}
      <Collapse in={expanded && showDetails}>
        <Grid container spacing={2}>
          {/* System Health - Temporarily disabled during migration */}
          {/* 
          {storeSystemHealth && (
            <Grid item xs={12} md={6}>
              <Card>
                <CardContent>
                  <Typography variant="h6" gutterBottom>
                    System Components
                  </Typography>
                  <Box>
                    {Object.entries(storeSystemHealth.components).map(([component, health]) => (
                      <Box key={component} display="flex" alignItems="center" gap={1} mb={1}>
                        {getHealthIcon(health.status)}
                        <Typography variant="body2" flex={1}>
                          {component.replace('_', ' ').replace(/\b\w/g, l => l.toUpperCase())}
                        </Typography>
                        <Chip
                          label={health.status}
                          color={getHealthColor(health.status)}
                          size="small"
                        />
                      </Box>
                    ))}
                  </Box>
                  <Typography variant="caption" color="text.secondary" display="block" mt={1}>
                    Updated: {formatTimestamp(storeSystemHealth.timestamp)}
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
          )}
          */}

          {/* Camera Health - Temporarily disabled during migration */}
          {/* 
          {storeCameraHealth && (
            <Grid item xs={12} md={6}>
              <Card>
                <CardContent>
                  <Typography variant="h6" gutterBottom>
                    Camera System
                  </Typography>
                  <Box display="flex" alignItems="center" gap={1} mb={1}>
                                    {getHealthIcon(storeCameraHealth.status)}
                <Chip
                  label={storeCameraHealth.status}
                  color={getHealthColor(storeCameraHealth.status)}
                      size="small"
                    />
                  </Box>
                  <Typography variant="body2" color="text.secondary">
                    {storeCameraHealth.details}
                  </Typography>
                  <Typography variant="caption" color="text.secondary" display="block" mt={1}>
                    Updated: {formatTimestamp(storeCameraHealth.timestamp)}
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
          )}
          */}

          {/* MediaMTX Health - Temporarily disabled during migration */}
          {/* 
          {storeMediaMTXHealth && (
            <Grid item xs={12} md={6}>
              <Card>
                <CardContent>
                  <Typography variant="h6" gutterBottom>
                    MediaMTX Integration
                  </Typography>
                  <Box display="flex" alignItems="center" gap={1} mb={1}>
                                    {getHealthIcon(storeMediaMTXHealth.status)}
                <Chip
                  label={storeMediaMTXHealth.status}
                  color={getHealthColor(storeMediaMTXHealth.status)}
                      size="small"
                    />
                  </Box>
                  <Typography variant="body2" color="text.secondary">
                    {storeMediaMTXHealth.details}
                  </Typography>
                  <Typography variant="caption" color="text.secondary" display="block" mt={1}>
                    Updated: {formatTimestamp(storeMediaMTXHealth.timestamp)}
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
          )}

          {/* Readiness Status - Temporarily disabled during migration */}
          {false && storeReadinessStatus && (
            <Grid item xs={12} md={6}>
              <Card>
                <CardContent>
                  <Typography variant="h6" gutterBottom>
                    Kubernetes Readiness
                  </Typography>
                  <Box display="flex" alignItems="center" gap={1} mb={1}>
                                    {getHealthIcon(storeReadinessStatus.status)}
                <Chip
                  label={storeReadinessStatus.status}
                  color={getHealthColor(storeReadinessStatus.status)}
                      size="small"
                    />
                  </Box>
                  {storeReadinessStatus.details && (
                    <Box mt={1}>
                                              {Object.entries(storeReadinessStatus.details).map(([component, status]) => (
                        <Typography key={component} variant="body2" color="text.secondary">
                          {component}: {status}
                        </Typography>
                      ))}
                    </Box>
                  )}
                  <Typography variant="caption" color="text.secondary" display="block" mt={1}>
                    Updated: {formatTimestamp(storeReadinessStatus.timestamp)}
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
          )}
        </Grid>

        {/* Error Alert */}
        {overallHealth === 'unhealthy' && (
          <Alert severity="error" sx={{ mt: 2 }}>
            System health is critical. Please check component status and logs.
          </Alert>
        )}

        {overallHealth === 'degraded' && (
          <Alert severity="warning" sx={{ mt: 2 }}>
            System health is degraded. Some components may not be functioning optimally.
          </Alert>
        )}
      </Collapse>
    </Box>
  );
};

export default HealthMonitor;
