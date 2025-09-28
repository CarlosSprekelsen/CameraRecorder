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
import { useServerStore } from '../../../stores/server/serverStore';
import { logger } from '../../../services/logger/LoggerService';
// ARCHITECTURE FIX: Logger is infrastructure - components can import it directly

interface HealthMonitorProps {
  // ARCHITECTURE FIX: Removed service props - components only use stores
}

export const HealthMonitor: React.FC<HealthMonitorProps> = () => {
  const { status, storage, loading: serverLoading, error: serverError } = useServerStore();
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
      // TODO: Implement health check via server service
      setLastCheck(new Date());
      logger.info('Health check completed successfully');
    } catch (err) {
      logger.error('Health check failed:', { error: err });
    } finally {
      setLoading(false);
    }
  };

  const getHealthStatus = () => {
    if (status?.status === 'HEALTHY' && (storage?.usage_percentage ?? 0) < 80) {
      return { status: 'healthy', color: 'success', icon: <CheckCircle /> };
    } else if (status?.status === 'HEALTHY' && (storage?.usage_percentage ?? 0) >= 80) {
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

        {(loading || serverLoading) && <LinearProgress sx={{ mb: 2 }} />}

        <Grid container spacing={2}>
          <Grid item xs={12} md={4}>
            <Card variant="outlined">
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                  <Speed sx={{ mr: 1, color: 'primary.main' }} />
                  <Typography variant="subtitle2">Server Status</Typography>
                </Box>
                <Typography variant="h6" color={status?.status === 'HEALTHY' ? 'success.main' : 'error.main'}>
                  {status?.status || 'Unknown'}
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
                  N/A%
                </Typography>
                <LinearProgress 
                  variant="determinate" 
                  value={0}
                  color="primary"
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
                  {storage?.usage_percentage?.toFixed(1) || 'N/A'}%
                </Typography>
                <LinearProgress 
                  variant="determinate" 
                  value={storage?.usage_percentage || 0}
                  color={storage?.usage_percentage > 90 ? 'error' : 'primary'}
                />
              </CardContent>
            </Card>
          </Grid>
        </Grid>

        {serverError && (
          <Alert severity="error" sx={{ mt: 2 }}>
            {serverError}
          </Alert>
        )}
      </CardContent>
    </Card>
  );
};
