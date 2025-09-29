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
import { useConnectionStore } from '../../../stores/connection/connectionStore';

interface HealthMonitorProps {
  // ARCHITECTURE FIX: Components use stores, not direct service props
}

export const HealthMonitor: React.FC<HealthMonitorProps> = () => {
  const { status, loadSystemStatus, setError } = useServerStore();
  const { status: connectionStatus } = useConnectionStore();
  const [loading, setLoading] = useState(false);
  const [lastCheck, setLastCheck] = useState<Date | null>(null);

  useEffect(() => {
    // ARCHITECTURE FIX: Removed client-side timer - health checks are server-authoritative
    // Initial check only - server will send health updates via WebSocket
    handleHealthCheck();
  }, []);

  const handleHealthCheck = async () => {
    setLoading(true);
    try {
      await loadSystemStatus();
      setLastCheck(new Date());
      console.log('Health check completed successfully');
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Health check failed';
      setError(errorMsg);
      console.error('Health check failed:', err instanceof Error ? err.message : String(err));
    } finally {
      setLoading(false);
    }
  };

  const getHealthStatus = () => {
    if (connectionStatus === 'connected' && status?.status === 'HEALTHY') {
      return { status: 'healthy', color: 'success', icon: <CheckCircle /> };
    } else if (connectionStatus === 'connected' && status?.status === 'DEGRADED') {
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
