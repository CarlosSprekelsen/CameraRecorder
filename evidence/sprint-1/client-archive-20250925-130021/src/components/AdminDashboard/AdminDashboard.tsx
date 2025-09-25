/**
 * Admin Dashboard Component
 * Provides system administration and management functionality
 * Aligned with server JSON-RPC methods for admin operations
 * 
 * Server API Reference: ../mediamtx-camera-service/docs/api/json-rpc-methods.md
 */

import React, { useEffect, useCallback } from 'react';
import { logger, loggers } from '../../services/loggerService';
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
    serverStats: storeServerStats,
    error: storeError,
    getSystemInfo: storeGetSystemInfo,
    getServerStats: storeGetServerStats,
    clearError: storeClearError,
    setError: storeSetError,
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
      await storeGetServerStats();
      
      storeClearError();
    } catch (error: unknown) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to refresh system info';
      logger.error('Failed to refresh system info', error instanceof Error ? error : undefined, 'adminDashboard', { errorMessage });
      storeSetError('Failed to load system information');
    } finally {
      setIsRefreshing(false);
    }
  }, [isRefreshing, storeGetServerStats, storeSetError, storeClearError]);

  /**
   * Perform cleanup operation
   */
  const performCleanup = useCallback(async () => {
    try {
      const results = await adminService.cleanupOldFiles();
      console.log('Cleanup completed:', results);
    } catch (error: unknown) {
      const errorMessage = error instanceof Error ? error.message : 'Cleanup failed';
      logger.error('Cleanup failed', error instanceof Error ? error : undefined, 'adminDashboard', { errorMessage });
      storeSetError('Cleanup operation failed');
    }
  }, [storeSetError]);

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
      console.log('Retention policy updated:', updatedPolicy);
    } catch (error: unknown) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to save retention policy';
      logger.error('Failed to save retention policy', error instanceof Error ? error : undefined, 'adminDashboard', { errorMessage });
      storeSetError('Failed to save retention policy');
    }
  }, [storeSetError]);

  /**
   * Check admin permissions on mount
   */
  useEffect(() => {
    const hasPermissions = adminService.hasAdminPermissions();
    console.log('Admin permissions:', hasPermissions);
  }, []);

  /**
   * Auto-refresh system information
   */
  useEffect(() => {
    if (autoRefresh) {
      refreshSystemInfo();
      
      const interval = setInterval(refreshSystemInfo, refreshInterval);
      return () => clearInterval(interval);
    }
  }, [autoRefresh, refreshInterval, refreshSystemInfo]);

  // Note: Admin permissions check removed - component will render for all users

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
      {storeError && (
        <Alert severity="error" sx={{ mb: 2 }} onClose={storeClearError}>
          {storeError}
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
              
              {false ? (
                <LinearProgress />
              ) : storeServerStats?.metrics ? (
                <Grid container spacing={2}>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="text.secondary">
                      Active Connections
                    </Typography>
                    <Typography variant="h6">
                      {storeServerStats.metrics.active_connections}
                    </Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="text.secondary">
                      Total Requests
                    </Typography>
                    <Typography variant="h6">
                      {storeServerStats.metrics.total_requests.toLocaleString()}
                    </Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="text.secondary">
                      Avg Response Time
                    </Typography>
                    <Typography variant="h6">
                      {storeServerStats.metrics.average_response_time.toFixed(1)}ms
                    </Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="text.secondary">
                      Error Rate
                    </Typography>
                    <Typography variant="h6">
                      {(storeServerStats.metrics.error_rate * 100).toFixed(2)}%
                    </Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="text.secondary">
                      CPU Usage
                    </Typography>
                    <Typography variant="h6">
                      {storeServerStats.metrics.cpu_usage.toFixed(1)}%
                    </Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="text.secondary">
                      Memory Usage
                    </Typography>
                    <Typography variant="h6">
                      {storeServerStats.metrics.memory_usage.toFixed(1)}%
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
                {false && <WarningIcon color="warning" />}
              </Box>
              
              {false ? (
                <LinearProgress />
              ) : storeServerStats?.storageInfo ? (
                <Box>
                  <Box display="flex" justifyContent="space-between" mb={1}>
                    <Typography variant="body2">Storage Usage</Typography>
                    <Typography variant="body2">
                      {((storeServerStats.storageInfo.used_space / storeServerStats.storageInfo.total_space) * 100).toFixed(1)}%
                    </Typography>
                  </Box>
                  <LinearProgress
                    variant="determinate"
                    value={(storeServerStats.storageInfo.used_space / storeServerStats.storageInfo.total_space) * 100}
                    color="primary"
                    sx={{ height: 8, borderRadius: 4, mb: 2 }}
                  />
                  
                  <Grid container spacing={2}>
                    <Grid item xs={6}>
                      <Typography variant="body2" color="text.secondary">
                        Total Space
                      </Typography>
                      <Typography variant="body1">
                        {storeServerStats.storageInfo.total_space}
                      </Typography>
                    </Grid>
                    <Grid item xs={6}>
                      <Typography variant="body2" color="text.secondary">
                        Available Space
                      </Typography>
                      <Typography variant="body1">
                        {storeFormatBytes(storeStorageInfo.available_space)}
                      </Typography>
                    </Grid>
                    <Grid item xs={6}>
                      <Typography variant="body2" color="text.secondary">
                        Recordings
                      </Typography>
                      <Typography variant="body1">
                        {storeFormatBytes(storeStorageInfo.recordings_size)}
                      </Typography>
                    </Grid>
                    <Grid item xs={6}>
                      <Typography variant="body2" color="text.secondary">
                        Snapshots
                      </Typography>
                      <Typography variant="body1">
                        {storeFormatBytes(storeStorageInfo.snapshots_size)}
                      </Typography>
                    </Grid>
                  </Grid>
                  
                  {false && (
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
              
              {storeIsLoadingStatus ? (
                <LinearProgress />
              ) : storeServerStats?.status ? (
                <Box>
                  <Box display="flex" alignItems="center" gap={1} mb={2}>
                    <Chip
                      label={storeServerStats.status.status}
                      color={storeServerStats.status.status === 'healthy' ? 'success' : 'warning'}
                      size="small"
                    />
                    <Typography variant="body2" color="text.secondary">
                      Uptime: {storeServerStats.status.uptime}
                    </Typography>
                  </Box>
                  
                  <Typography variant="body2" color="text.secondary" gutterBottom>
                    Version: {storeServerStats.status.version}
                  </Typography>
                  
                  <Typography variant="body2" color="text.secondary" gutterBottom>
                    Components:
                  </Typography>
                  <Box>
                    {Object.entries(storeServerStats.status.components).map(([component, status]) => (
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
              
              {storeServerInfo ? (
                <Box>
                  <Typography variant="body1" gutterBottom>
                    {storeServerInfo.name} v{storeServerInfo.version}
                  </Typography>
                  
                  <Typography variant="body2" color="text.secondary" gutterBottom>
                    Max Cameras: {storeServerInfo.max_cameras}
                  </Typography>
                  
                  <Typography variant="body2" color="text.secondary" gutterBottom>
                    Supported Formats:
                  </Typography>
                  <Box display="flex" gap={0.5} flexWrap="wrap" mb={1}>
                    {storeServerInfo.supported_formats.map((format) => (
                      <Chip key={format} label={format} size="small" />
                    ))}
                  </Box>
                  
                  <Typography variant="body2" color="text.secondary" gutterBottom>
                    Capabilities:
                  </Typography>
                  <Box display="flex" gap={0.5} flexWrap="wrap">
                    {storeServerInfo.capabilities.map((capability) => (
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
                disabled={storeIsPerformingCleanup}
                fullWidth
              >
                {storeIsPerformingCleanup ? 'Cleaning...' : 'Cleanup Old Files'}
              </Button>
            </Grid>
          </Grid>
          
          {/* Cleanup Results */}
          {storeLastCleanupResults && (
            <Alert severity="info" sx={{ mt: 2 }}>
              {storeLastCleanupResults.message}
              {storeLastCleanupResults.files_deleted > 0 && (
                <Typography variant="body2">
                  Files deleted: {storeLastCleanupResults.files_deleted} | 
                  Space freed: {storeFormatBytes(storeLastCleanupResults.space_freed)}
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
        currentPolicy={storeRetentionPolicy || undefined}
      />
    </Box>
  );
};

export default AdminDashboard;
