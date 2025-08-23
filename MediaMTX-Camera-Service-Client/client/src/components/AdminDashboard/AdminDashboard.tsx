/**
 * Admin Dashboard Component
 * Provides system administration and management functionality
 * Aligned with server JSON-RPC methods for admin operations
 * 
 * Server API Reference: ../mediamtx-camera-service/docs/api/json-rpc-methods.md
 */

import React, { useEffect, useCallback } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Grid,
  Button,
  LinearProgress,
  Alert,
  IconButton,
  Tooltip,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Switch,
  FormControlLabel,
  Chip,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
} from '@mui/material';
import {
  Refresh as RefreshIcon,
  Storage as StorageIcon,
  Speed as SpeedIcon,
  Memory as MemoryIcon,
  Settings as SettingsIcon,
  Delete as DeleteIcon,
  Warning as WarningIcon,
} from '@mui/icons-material';
import { useAdminStore } from '../../stores/adminStore';
import { adminService } from '../../services/adminService';

/**
 * Admin Dashboard Component Props
 */
interface AdminDashboardProps {
  autoRefresh?: boolean;
  refreshInterval?: number;
}

/**
 * Retention Policy Dialog Props
 */
interface RetentionPolicyDialogProps {
  open: boolean;
  onClose: () => void;
  onSave: (policy: {
    policy_type: 'age' | 'size' | 'manual';
    max_age_days?: number;
    max_size_gb?: number;
    enabled: boolean;
  }) => void;
  currentPolicy?: {
    policy_type: 'age' | 'size' | 'manual';
    max_age_days?: number;
    max_size_gb?: number;
    enabled: boolean;
  };
}

/**
 * Retention Policy Dialog Component
 */
const RetentionPolicyDialog: React.FC<RetentionPolicyDialogProps> = ({
  open,
  onClose,
  onSave,
  currentPolicy,
}) => {
  const [policy, setPolicy] = React.useState({
    policy_type: currentPolicy?.policy_type || 'age',
    max_age_days: currentPolicy?.max_age_days || 30,
    max_size_gb: currentPolicy?.max_size_gb || 10,
    enabled: currentPolicy?.enabled || false,
  });

  const handleSave = () => {
    onSave(policy);
    onClose();
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>Configure Retention Policy</DialogTitle>
      <DialogContent>
        <Grid container spacing={2} sx={{ mt: 1 }}>
          <Grid item xs={12}>
            <FormControl fullWidth>
              <InputLabel>Policy Type</InputLabel>
              <Select
                value={policy.policy_type}
                onChange={(e) => setPolicy({ ...policy, policy_type: e.target.value as 'age' | 'size' | 'manual' })}
                label="Policy Type"
              >
                <MenuItem value="age">Age-based (delete files older than X days)</MenuItem>
                <MenuItem value="size">Size-based (delete files when storage exceeds X GB)</MenuItem>
                <MenuItem value="manual">Manual (no automatic cleanup)</MenuItem>
              </Select>
            </FormControl>
          </Grid>
          
          {policy.policy_type === 'age' && (
            <Grid item xs={12}>
              <TextField
                fullWidth
                type="number"
                label="Maximum Age (days)"
                value={policy.max_age_days}
                onChange={(e) => setPolicy({ ...policy, max_age_days: parseInt(e.target.value) || 30 })}
                inputProps={{ min: 1, max: 365 }}
              />
            </Grid>
          )}
          
          {policy.policy_type === 'size' && (
            <Grid item xs={12}>
              <TextField
                fullWidth
                type="number"
                label="Maximum Size (GB)"
                value={policy.max_size_gb}
                onChange={(e) => setPolicy({ ...policy, max_size_gb: parseInt(e.target.value) || 10 })}
                inputProps={{ min: 1, max: 1000 }}
              />
            </Grid>
          )}
          
          <Grid item xs={12}>
            <FormControlLabel
              control={
                <Switch
                  checked={policy.enabled}
                  onChange={(e) => setPolicy({ ...policy, enabled: e.target.checked })}
                />
              }
              label="Enable automatic cleanup"
            />
          </Grid>
        </Grid>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button onClick={handleSave} variant="contained">
          Save Policy
        </Button>
      </DialogActions>
    </Dialog>
  );
};

/**
 * Admin Dashboard Component
 */
const AdminDashboard: React.FC<AdminDashboardProps> = ({
  autoRefresh = true,
  refreshInterval = 60000, // 1 minute
}) => {
  const {
    systemMetrics,
    systemStatus,
    serverInfo,
    storageInfo,
    retentionPolicy,
    isPerformingCleanup,
    lastCleanupResults,
    isAdmin,
    hasAdminPermissions,
    isLoadingMetrics,
    isLoadingStatus,
    isLoadingStorage,
    error,
    setSystemMetrics,
    setSystemStatus,
    setServerInfo,
    setStorageInfo,
    setRetentionPolicy,
    setCleanupResults,
    setPerformingCleanup,
    setAdminStatus,
    setLoadingMetrics,
    setLoadingStatus,
    setLoadingStorage,
    setError,
    clearError,
    getStorageUsagePercentage,
    getStorageUsageColor,
    isLowSpace,
    formatBytes,
    formatUptime,
  } = useAdminStore();

  const [retentionDialogOpen, setRetentionDialogOpen] = React.useState(false);
  const [isRefreshing, setIsRefreshing] = React.useState(false);

  /**
   * Refresh all system information
   */
  const refreshSystemInfo = useCallback(async () => {
    if (isRefreshing) return;

    setIsRefreshing(true);
    try {
      setLoadingMetrics(true);
      setLoadingStatus(true);
      setLoadingStorage(true);

      const systemInfo = await adminService.getAllSystemInfo();
      
      setSystemMetrics(systemInfo.metrics);
      setSystemStatus(systemInfo.status);
      setServerInfo(systemInfo.serverInfo);
      setStorageInfo(systemInfo.storageInfo);
      
      clearError();
    } catch (error) {
      console.error('Failed to refresh system info:', error);
      setError('Failed to load system information');
    } finally {
      setIsRefreshing(false);
      setLoadingMetrics(false);
      setLoadingStatus(false);
      setLoadingStorage(false);
    }
  }, [isRefreshing, setSystemMetrics, setSystemStatus, setServerInfo, setStorageInfo, setLoadingMetrics, setLoadingStatus, setLoadingStorage, setError, clearError]);

  /**
   * Perform cleanup operation
   */
  const performCleanup = useCallback(async () => {
    try {
      setPerformingCleanup(true);
      const results = await adminService.cleanupOldFiles();
      setCleanupResults(results);
    } catch (error) {
      console.error('Cleanup failed:', error);
      setError('Cleanup operation failed');
    } finally {
      setPerformingCleanup(false);
    }
  }, [setPerformingCleanup, setCleanupResults, setError]);

  /**
   * Save retention policy
   */
  const saveRetentionPolicy = useCallback(async (policy: {
    policy_type: 'age' | 'size' | 'manual';
    max_age_days?: number;
    max_size_gb?: number;
    enabled: boolean;
  }) => {
    try {
      const updatedPolicy = await adminService.setRetentionPolicy(policy);
      setRetentionPolicy(updatedPolicy);
    } catch (error) {
      console.error('Failed to save retention policy:', error);
      setError('Failed to save retention policy');
    }
  }, [setRetentionPolicy, setError]);

  /**
   * Check admin permissions on mount
   */
  useEffect(() => {
    const hasPermissions = adminService.hasAdminPermissions();
    setAdminStatus(true, hasPermissions);
  }, [setAdminStatus]);

  /**
   * Auto-refresh system information
   */
  useEffect(() => {
    if (autoRefresh && hasAdminPermissions) {
      refreshSystemInfo();
      
      const interval = setInterval(refreshSystemInfo, refreshInterval);
      return () => clearInterval(interval);
    }
  }, [autoRefresh, refreshInterval, hasAdminPermissions, refreshSystemInfo]);

  // Don't render if user doesn't have admin permissions
  if (!hasAdminPermissions) {
    return (
      <Alert severity="warning">
        You don't have permission to access the admin dashboard.
      </Alert>
    );
  }

  return (
    <Box>
      {/* Header */}
      <Box display="flex" alignItems="center" justifyContent="space-between" mb={3}>
        <Typography variant="h4" component="h1">
          Admin Dashboard
        </Typography>
        <Box display="flex" gap={1}>
          <Tooltip title="Refresh system information">
            <IconButton onClick={refreshSystemInfo} disabled={isRefreshing}>
              <RefreshIcon />
            </IconButton>
          </Tooltip>
        </Box>
      </Box>

      {/* Error Alert */}
      {error && (
        <Alert severity="error" sx={{ mb: 2 }} onClose={clearError}>
          {error}
        </Alert>
      )}

      {/* System Overview */}
      <Grid container spacing={3}>
        {/* System Metrics */}
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Box display="flex" alignItems="center" gap={1} mb={2}>
                <SpeedIcon />
                <Typography variant="h6">System Performance</Typography>
              </Box>
              
              {isLoadingMetrics ? (
                <LinearProgress />
              ) : systemMetrics ? (
                <Grid container spacing={2}>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="text.secondary">
                      Active Connections
                    </Typography>
                    <Typography variant="h6">
                      {systemMetrics.active_connections}
                    </Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="text.secondary">
                      Total Requests
                    </Typography>
                    <Typography variant="h6">
                      {systemMetrics.total_requests.toLocaleString()}
                    </Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="text.secondary">
                      Avg Response Time
                    </Typography>
                    <Typography variant="h6">
                      {systemMetrics.average_response_time.toFixed(1)}ms
                    </Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="text.secondary">
                      Error Rate
                    </Typography>
                    <Typography variant="h6">
                      {(systemMetrics.error_rate * 100).toFixed(2)}%
                    </Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="text.secondary">
                      CPU Usage
                    </Typography>
                    <Typography variant="h6">
                      {systemMetrics.cpu_usage.toFixed(1)}%
                    </Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="text.secondary">
                      Memory Usage
                    </Typography>
                    <Typography variant="h6">
                      {systemMetrics.memory_usage.toFixed(1)}%
                    </Typography>
                  </Grid>
                </Grid>
              ) : (
                <Typography color="text.secondary">No metrics available</Typography>
              )}
            </CardContent>
          </Card>
        </Grid>

        {/* Storage Information */}
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Box display="flex" alignItems="center" gap={1} mb={2}>
                <StorageIcon />
                <Typography variant="h6">Storage Status</Typography>
                {isLowSpace() && <WarningIcon color="warning" />}
              </Box>
              
              {isLoadingStorage ? (
                <LinearProgress />
              ) : storageInfo ? (
                <Box>
                  <Box display="flex" justifyContent="space-between" mb={1}>
                    <Typography variant="body2">Storage Usage</Typography>
                    <Typography variant="body2">
                      {getStorageUsagePercentage().toFixed(1)}%
                    </Typography>
                  </Box>
                  <LinearProgress
                    variant="determinate"
                    value={getStorageUsagePercentage()}
                    color={getStorageUsageColor()}
                    sx={{ height: 8, borderRadius: 4, mb: 2 }}
                  />
                  
                  <Grid container spacing={2}>
                    <Grid item xs={6}>
                      <Typography variant="body2" color="text.secondary">
                        Total Space
                      </Typography>
                      <Typography variant="body1">
                        {formatBytes(storageInfo.total_space)}
                      </Typography>
                    </Grid>
                    <Grid item xs={6}>
                      <Typography variant="body2" color="text.secondary">
                        Available Space
                      </Typography>
                      <Typography variant="body1">
                        {formatBytes(storageInfo.available_space)}
                      </Typography>
                    </Grid>
                    <Grid item xs={6}>
                      <Typography variant="body2" color="text.secondary">
                        Recordings
                      </Typography>
                      <Typography variant="body1">
                        {formatBytes(storageInfo.recordings_size)}
                      </Typography>
                    </Grid>
                    <Grid item xs={6}>
                      <Typography variant="body2" color="text.secondary">
                        Snapshots
                      </Typography>
                      <Typography variant="body1">
                        {formatBytes(storageInfo.snapshots_size)}
                      </Typography>
                    </Grid>
                  </Grid>
                  
                  {isLowSpace() && (
                    <Alert severity="warning" sx={{ mt: 2 }}>
                      Low storage space detected. Consider cleaning up old files.
                    </Alert>
                  )}
                </Box>
              ) : (
                <Typography color="text.secondary">No storage information available</Typography>
              )}
            </CardContent>
          </Card>
        </Grid>

        {/* System Status */}
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Box display="flex" alignItems="center" gap={1} mb={2}>
                <MemoryIcon />
                <Typography variant="h6">System Status</Typography>
              </Box>
              
              {isLoadingStatus ? (
                <LinearProgress />
              ) : systemStatus ? (
                <Box>
                  <Box display="flex" alignItems="center" gap={1} mb={2}>
                    <Chip
                      label={systemStatus.status}
                      color={systemStatus.status === 'healthy' ? 'success' : 'warning'}
                      size="small"
                    />
                    <Typography variant="body2" color="text.secondary">
                      Uptime: {formatUptime(systemStatus.uptime)}
                    </Typography>
                  </Box>
                  
                  <Typography variant="body2" color="text.secondary" gutterBottom>
                    Version: {systemStatus.version}
                  </Typography>
                  
                  <Typography variant="body2" color="text.secondary" gutterBottom>
                    Components:
                  </Typography>
                  <Box>
                    {Object.entries(systemStatus.components).map(([component, status]) => (
                      <Box key={component} display="flex" justifyContent="space-between" mb={0.5}>
                        <Typography variant="body2">
                          {component.replace('_', ' ').replace(/\b\w/g, l => l.toUpperCase())}
                        </Typography>
                        <Chip
                          label={status}
                          color={status === 'running' ? 'success' : 'error'}
                          size="small"
                        />
                      </Box>
                    ))}
                  </Box>
                </Box>
              ) : (
                <Typography color="text.secondary">No status information available</Typography>
              )}
            </CardContent>
          </Card>
        </Grid>

        {/* Server Information */}
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Box display="flex" alignItems="center" gap={1} mb={2}>
                <SettingsIcon />
                <Typography variant="h6">Server Information</Typography>
              </Box>
              
              {serverInfo ? (
                <Box>
                  <Typography variant="body1" gutterBottom>
                    {serverInfo.name} v{serverInfo.version}
                  </Typography>
                  
                  <Typography variant="body2" color="text.secondary" gutterBottom>
                    Max Cameras: {serverInfo.max_cameras}
                  </Typography>
                  
                  <Typography variant="body2" color="text.secondary" gutterBottom>
                    Supported Formats:
                  </Typography>
                  <Box display="flex" gap={0.5} flexWrap="wrap" mb={1}>
                    {serverInfo.supported_formats.map((format) => (
                      <Chip key={format} label={format} size="small" />
                    ))}
                  </Box>
                  
                  <Typography variant="body2" color="text.secondary" gutterBottom>
                    Capabilities:
                  </Typography>
                  <Box display="flex" gap={0.5} flexWrap="wrap">
                    {serverInfo.capabilities.map((capability) => (
                      <Chip key={capability} label={capability} size="small" variant="outlined" />
                    ))}
                  </Box>
                </Box>
              ) : (
                <Typography color="text.secondary">No server information available</Typography>
              )}
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Management Actions */}
      <Card sx={{ mt: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Management Actions
          </Typography>
          
          <Grid container spacing={2}>
            <Grid item xs={12} sm={6} md={4}>
              <Button
                variant="outlined"
                startIcon={<SettingsIcon />}
                onClick={() => setRetentionDialogOpen(true)}
                fullWidth
              >
                Configure Retention Policy
              </Button>
            </Grid>
            
            <Grid item xs={12} sm={6} md={4}>
              <Button
                variant="outlined"
                startIcon={<DeleteIcon />}
                onClick={performCleanup}
                disabled={isPerformingCleanup}
                fullWidth
              >
                {isPerformingCleanup ? 'Cleaning...' : 'Cleanup Old Files'}
              </Button>
            </Grid>
          </Grid>
          
          {/* Cleanup Results */}
          {lastCleanupResults && (
            <Alert severity="info" sx={{ mt: 2 }}>
              {lastCleanupResults.message}
              {lastCleanupResults.files_deleted > 0 && (
                <Typography variant="body2">
                  Files deleted: {lastCleanupResults.files_deleted} | 
                  Space freed: {formatBytes(lastCleanupResults.space_freed)}
                </Typography>
              )}
            </Alert>
          )}
        </CardContent>
      </Card>

      {/* Retention Policy Dialog */}
      <RetentionPolicyDialog
        open={retentionDialogOpen}
        onClose={() => setRetentionDialogOpen(false)}
        onSave={saveRetentionPolicy}
        currentPolicy={retentionPolicy || undefined}
      />
    </Box>
  );
};

export default AdminDashboard;
