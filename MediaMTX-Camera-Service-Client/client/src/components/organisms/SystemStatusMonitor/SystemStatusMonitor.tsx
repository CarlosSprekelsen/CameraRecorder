/**
 * SystemStatusMonitor - Architecture Compliance
 * 
 * Architecture requirement: "SystemStatusMonitor component" (Section 5.2)
 * Provides WebSocket-based system status monitoring and alerting functionality
 * Prevents confusion with HTTP health endpoints (Port 8003)
 */

import React, { useState, useEffect } from 'react';
import { Grid } from '../../atoms/Grid/Grid';
import { Card } from '../../atoms/Card/Card';
import { Alert } from '../../atoms/Alert/Alert';
import { Badge } from '../../atoms/Badge/Badge';
import { Icon } from '../../atoms/Icon/Icon';
import { useServerStore } from '../../../stores/server/serverStore';
import { useConnectionStore } from '../../../stores/connection/connectionStore';

export interface SystemStatusMonitorProps {
  className?: string;
}

export const SystemStatusMonitor: React.FC<SystemStatusMonitorProps> = ({ className = '' }) => {
  const [loading, setLoading] = useState(false);
  const [lastCheck, setLastCheck] = useState<Date | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [pingStatus, setPingStatus] = useState<'idle' | 'checking' | 'success' | 'failed'>('idle');
  const [pingResult, setPingResult] = useState<string | null>(null);

  const { systemReadiness, loadSystemReadiness, setError: setServerError, ping } = useServerStore();
  const { status: connectionStatus } = useConnectionStore();

  useEffect(() => {
    // ARCHITECTURE FIX: Removed client-side timer - health checks are server-authoritative
    // Initial check only - server will send health updates via WebSocket
    handleSystemReadinessCheck();
  }, []);

  const handleSystemReadinessCheck = async () => {
    setLoading(true);
    try {
      await loadSystemReadiness();
      setLastCheck(new Date());
      console.log('System readiness check completed successfully');
    } catch (err: unknown) {
      let errorMsg: string;
      if (err instanceof Error) {
        errorMsg = (err as Error).message;
      } else {
        errorMsg = String(err);
      }
      setServerError(errorMsg);
      setError(errorMsg);
      console.error('System readiness check failed:', errorMsg);
    } finally {
      setLoading(false);
    }
  };

  const handlePing = async () => {
    setPingStatus('checking');
    try {
      const result = await ping();
      setPingResult(result);
      setPingStatus('success');
      console.log('Ping successful:', result);
    } catch (err: unknown) {
      setPingStatus('failed');
      console.error('Ping failed:', err);
    }
  };

  const getSystemReadinessStatus = () => {
    if (connectionStatus === 'connected' && systemReadiness?.status === 'ready') {
      return { status: 'ready', color: 'success', icon: <Icon name="checkCircle" /> };
    } else if (connectionStatus === 'connected' && systemReadiness?.status === 'partial') {
      return { status: 'partial', color: 'warning', icon: <Icon name="warning" /> };
    } else if (connectionStatus === 'connected' && systemReadiness?.status === 'starting') {
      return { status: 'starting', color: 'info', icon: <Icon name="autorenew" className="animate-spin" /> };
    } else if (connectionStatus === 'connecting') {
      return { status: 'connecting', color: 'info', icon: <Icon name="autorenew" className="animate-spin" /> };
    } else {
      return { status: 'disconnected', color: 'error', icon: <Icon name="error" /> };
    }
  };

  const systemReadinessStatus = getSystemReadinessStatus();

  return (
    <div className={`health-monitor ${className}`}>
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-xl font-semibold">System Readiness</h2>
        <button
          onClick={handleSystemReadinessCheck}
          disabled={loading}
          className="flex items-center gap-2 px-3 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
        >
          <Icon name="refresh" size={16} />
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
                <Icon name="speed" size={20} color="rgb(37 99 235)" className="mr-2" />
                <h3 className="text-sm font-medium">Server Status</h3>
              </div>
              <h2 className={`text-lg font-semibold ${systemReadinessStatus.color === 'success' ? 'text-green-600' : systemReadinessStatus.color === 'warning' ? 'text-yellow-600' : systemReadinessStatus.color === 'info' ? 'text-blue-600' : 'text-red-600'}`}>
                {systemReadiness?.status || 'Unknown'}
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
                <Icon name="memory" size={20} color="rgb(37 99 235)" className="mr-2" />
                <h3 className="text-sm font-medium">System Metrics</h3>
              </div>
              <div className="space-y-2">
                <div className="flex justify-between">
                  <span className="text-sm text-gray-600">Readiness:</span>
                  <Badge variant={systemReadiness?.status === 'ready' ? 'success' : systemReadiness?.status === 'partial' ? 'warning' : systemReadiness?.status === 'starting' ? 'info' : 'error'}>
                    {systemReadiness?.status || 'UNKNOWN'}
                  </Badge>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-gray-600">Cameras:</span>
                  <Badge variant="info">{systemReadiness?.available_cameras?.length || 0}</Badge>
                </div>
                <div className="flex justify-between">
                  <span className="text-sm text-gray-600">Discovery:</span>
                  <Badge variant={systemReadiness?.discovery_active ? 'warning' : 'success'}>
                    {systemReadiness?.discovery_active ? 'Active' : 'Complete'}
                  </Badge>
                </div>
              </div>
            </div>
          </Card>
        </Grid>

        <Grid item xs={12} md={4}>
          <Card variant="outlined">
            <div className="p-4">
              <div className="flex items-center mb-2">
                <Icon name="storage" size={20} color="rgb(37 99 235)" className="mr-2" />
                <h3 className="text-sm font-medium">Storage</h3>
              </div>
              <div className="space-y-2">
                <div className="flex justify-between">
                  <span className="text-sm text-gray-600">Storage Info:</span>
                  <Badge variant="info">Available</Badge>
                </div>
              </div>
            </div>
          </Card>
        </Grid>

        <Grid item xs={12} md={4}>
          <Card variant="outlined">
            <div className="p-4">
              <div className="flex items-center mb-2">
                <Icon name="speed" size={20} color="rgb(37 99 235)" className="mr-2" />
                <h3 className="text-sm font-medium">Connection Health</h3>
              </div>
              <div className="space-y-2">
                <div className="flex justify-between">
                  <span className="text-sm text-gray-600">Ping Status:</span>
                  <Badge variant={
                    pingStatus === 'success' ? 'success' : 
                    pingStatus === 'failed' ? 'error' : 
                    pingStatus === 'checking' ? 'info' : 'default'
                  }>
                    {pingStatus === 'checking' ? 'Checking...' : 
                     pingStatus === 'success' ? 'Connected' : 
                     pingStatus === 'failed' ? 'Failed' : 'Not Checked'}
                  </Badge>
                </div>
                {pingResult && (
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-600">Response:</span>
                    <span className="text-sm text-green-600">{pingResult}</span>
                  </div>
                )}
                <button
                  onClick={handlePing}
                  disabled={pingStatus === 'checking'}
                  className="w-full mt-2 px-3 py-1 bg-blue-600 text-white text-sm rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {pingStatus === 'checking' ? 'Pinging...' : 'Test Connection'}
                </button>
              </div>
            </div>
          </Card>
        </Grid>
      </Grid>

      {/* Available Cameras Section */}
      {systemReadiness?.available_cameras && systemReadiness.available_cameras.length > 0 && (
        <div className="mt-4">
          <Card variant="outlined">
            <div className="p-4">
              <div className="flex items-center mb-2">
                <Icon name="camera" size={20} color="rgb(37 99 235)" className="mr-2" />
                <h3 className="text-sm font-medium">Available Cameras</h3>
              </div>
              <div className="flex flex-wrap gap-2">
                {systemReadiness.available_cameras.map((camera, index) => (
                  <Badge key={index} variant="success" className="text-xs">
                    {camera}
                  </Badge>
                ))}
              </div>
            </div>
          </Card>
        </div>
      )}

      {/* System Message */}
      {systemReadiness?.message && (
        <div className="mt-4">
          <Alert variant="info">
            {systemReadiness.message}
          </Alert>
        </div>
      )}
    </div>
  );
};