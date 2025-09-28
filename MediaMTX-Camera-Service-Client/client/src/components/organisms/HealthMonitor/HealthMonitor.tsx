/**
 * HealthMonitor - Architecture Compliance
 * 
 * Architecture requirement: "HealthMonitor component" (Section 5.2)
 * Provides system health monitoring and alerting functionality
 */

import React, { useState, useEffect } from 'react';
import { 
  Box, 
  Typography, 
  Card, 
  CardContent, 
  Alert, 
  Chip, 
  LinearProgress,
  Grid 
} from '@mui/material';
import { 
  CheckCircle, 
  Error, 
  Warning, 
  Storage,
  Memory,
  Speed 
} from '@mui/icons-material';
import { useUnifiedStore } from '../../../stores/UnifiedStateStore';
import { APIClient } from '../../../services/abstraction/APIClient';
import { LoggerService } from '../../../services/logger/LoggerService';

interface HealthMonitorProps {
  apiClient: APIClient;
  logger: LoggerService;
}

export const HealthMonitor: React.FC<HealthMonitorProps> = ({ apiClient, logger }) => {
  const { serverStatus, systemMetrics, checkHealth, setHealthError } = useUnifiedStore();
  const [loading, setLoading] = useState(false);
  const [lastCheck, setLastCheck] = useState<Date | null>(null);

  useEffect(() => {
    const interval = setInterval(() => {
      handleHealthCheck();
    }, 30000); // Check every 30 seconds

    // Initial check
    handleHealthCheck();

    return () => clearInterval(interval);
  }, []);

  const handleHealthCheck = async () => {
    setLoading(true);
    try {
      await checkHealth();
      setLastCheck(new Date());
      logger.info('Health check completed successfully');
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Health check failed';
      setHealthError(errorMsg);
      logger.error('Health check failed:', err);
    } finally {
      setLoading(false);
    }
  };

  const getHealthStatus = () => {
    if (serverStatus?.status === 'online' && systemMetrics?.cpu_usage < 80) {
      return { status: 'healthy', color: 'success', icon: <CheckCircle /> };
    } else if (serverStatus?.status === 'online' && systemMetrics?.cpu_usage >= 80) {
      return { status: 'warning', color: 'warning', icon: <Warning /> };
    } else {
      return { status: 'error', color: 'error', icon: <Error /> };
    }
  };

  const healthStatus = getHealthStatus();

  return (
    <Card sx={{ mb: 2 }}>
      <CardContent>
        <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
          {healthStatus.icon}
          <Typography variant="h6" sx={{ ml: 1 }}>
            System Health Monitor
          </Typography>
          <Chip 
            label={healthStatus.status.toUpperCase()} 
            color={healthStatus.color as any}
            size="small"
            sx={{ ml: 'auto' }}
          />
        </Box>

        {loading && <LinearProgress sx={{ mb: 2 }} />}

        <Grid container spacing={2}>
          <Grid item xs={12} md={4}>
            <Card variant="outlined">
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                  <Speed sx={{ mr: 1, color: 'primary.main' }} />
                  <Typography variant="subtitle2">Server Status</Typography>
                </Box>
                <Typography variant="h6" color={serverStatus?.status === 'online' ? 'success.main' : 'error.main'}>
                  {serverStatus?.status || 'Unknown'}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  Last check: {lastCheck?.toLocaleTimeString() || 'Never'}
                </Typography>
              </CardContent>
            </Card>
          </Grid>

          <Grid item xs={12} md={4}>
            <Card variant="outlined">
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                  <Memory sx={{ mr: 1, color: 'primary.main' }} />
                  <Typography variant="subtitle2">CPU Usage</Typography>
                </Box>
                <Typography variant="h6">
                  {systemMetrics?.cpu_usage?.toFixed(1) || 'N/A'}%
                </Typography>
                <LinearProgress 
                  variant="determinate" 
                  value={systemMetrics?.cpu_usage || 0}
                  color={systemMetrics?.cpu_usage > 80 ? 'error' : 'primary'}
                />
              </CardContent>
            </Card>
          </Grid>

          <Grid item xs={12} md={4}>
            <Card variant="outlined">
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                  <Storage sx={{ mr: 1, color: 'primary.main' }} />
                  <Typography variant="subtitle2">Storage</Typography>
                </Box>
                <Typography variant="h6">
                  {systemMetrics?.storage_usage?.toFixed(1) || 'N/A'}%
                </Typography>
                <LinearProgress 
                  variant="determinate" 
                  value={systemMetrics?.storage_usage || 0}
                  color={systemMetrics?.storage_usage > 90 ? 'error' : 'primary'}
                />
              </CardContent>
            </Card>
          </Grid>
        </Grid>

        {serverStatus?.error && (
          <Alert severity="error" sx={{ mt: 2 }}>
            {serverStatus.error}
          </Alert>
        )}
      </CardContent>
    </Card>
  );
};
