/**
 * HealthMonitor - Architecture Compliance
 * 
 * Architecture requirement: "HealthMonitor component" (Section 5.2)
 * Provides system health monitoring and alerting functionality
 */

import React, { useState, useEffect } from 'react';
import { Grid } from '../../atoms/Grid/Grid';
import { Card } from '../../atoms/Card/Card';
import { Alert } from '../../atoms/Alert/Alert';
import { Badge } from '../../atoms/Badge/Badge';
import { 
  CheckCircle, 
  Error, 
  Warning, 
  Storage,
  Speed,
  Memory,
  Refresh
} from '@mui/icons-material';
import { useServerStore } from '../../../stores/server/serverStore';
import { useConnectionStore } from '../../../stores/connection/connectionStore';

export interface HealthMonitorProps {
  className?: string;
}

export const HealthMonitor: React.FC<HealthMonitorProps> = ({ className = '' }) => {
  const [loading, setLoading] = useState(false);
  const [lastCheck, setLastCheck] = useState<Date | null>(null);
  const [error, setError] = useState<string | null>(null);

  const { status, loadSystemStatus, setError: setServerError } = useServerStore();
  const { status: connectionStatus } = useConnectionStore();

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
      setServerError(errorMsg);
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
    <div className={`health-monitor ${className}`}>
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-xl font-semibold">System Health</h2>
        <button
          onClick={handleHealthCheck}
          disabled={loading}
          className="flex items-center gap-2 px-3 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
        >
          <Refresh className="h-4 w-4" />
          {loading ? 'Checking...' : 'Refresh'}
        </button>
      </div>

      {error && (
        <Alert variant="error" className="mb-4">
          {error}
        </Alert>
      )}

      {loading && (
        <div className="w-full bg-gray-200 rounded-full h-2 mb-4">
          <div className="bg-blue-600 h-2 rounded-full animate-pulse"></div>
        </div>
      )}

      <Grid container spacing={2}>
        <Grid item xs={12} md={4}>
          <Card variant="outlined">
            <div className="p-4">
              <div className="flex items-center mb-2">
                <Speed className="mr-2 text-blue-600" />
                <h3 className="text-sm font-medium">Server Status</h3>
              </div>
              <h2 className={`text-lg font-semibold ${healthStatus.color === 'success' ? 'text-green-600' : healthStatus.color === 'warning' ? 'text-yellow-600' : 'text-red-600'}`}>
                {status?.status || 'Unknown'}
              </h2>
              <p className="text-sm text-gray-500">
                Last check: {lastCheck?.toLocaleTimeString() || 'Never'}
              </p>
            </div>
          </Card>
        </Grid>

        <Grid item xs={12} md={4}>
          <Card variant="outlined">
            <div className="p-4">
              <div className="flex items-center mb-2">
                <Memory className="mr-2 text-blue-600" />
                <h3 className="text-sm font-medium">System Metrics</h3>
              </div>
              <div className="space-y-2">
                <div className="flex justify-between">
                  <span className="text-sm text-gray-600">CPU Usage:</span>
                  <Badge variant="info">{status?.cpu_usage || 'N/A'}%</Badge>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-gray-600">Memory:</span>
                  <Badge variant="info">{status?.memory_usage || 'N/A'}%</Badge>
                </div>
              </div>
            </div>
          </Card>
        </Grid>

        <Grid item xs={12} md={4}>
          <Card variant="outlined">
            <div className="p-4">
              <div className="flex items-center mb-2">
                <Storage className="mr-2 text-blue-600" />
                <h3 className="text-sm font-medium">Storage</h3>
              </div>
              <div className="space-y-2">
                <div className="flex justify-between">
                  <span className="text-sm text-gray-600">Available:</span>
                  <Badge variant="success">{status?.storage_available || 'N/A'}</Badge>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-gray-600">Used:</span>
                  <Badge variant="warning">{status?.storage_used || 'N/A'}</Badge>
                </div>
              </div>
            </div>
          </Card>
        </Grid>
      </Grid>
    </div>
  );
};