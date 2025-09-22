import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Button,
  Chip,
  Alert,
  CircularProgress,
  Grid,
  Switch,
  FormControlLabel,
} from '@mui/material';
import {
  Storage,
  Warning,
  Error,
  CheckCircle,
  Refresh,
  Visibility,
  VisibilityOff,
} from '@mui/icons-material';
import { useStorageStore } from '../../stores/storageStore';
import { StorageInfo, ThresholdStatus } from '../../types/camera';

const StorageMonitor: React.FC = () => {
  const [localLoading, setLocalLoading] = useState(false);
  const [localError, setLocalError] = useState<string | null>(null);
  const [localShowDetails, setLocalShowDetails] = useState(false);
  const [localMonitoringEnabled, setLocalMonitoringEnabled] = useState(true);

  // Store state
  const {
    storageInfo: storeStorageInfo,
    thresholdStatus: storeThresholdStatus,
    warnings: storeWarnings,
    isLoading,
    error: storeError,
    refreshStorage,
    startMonitoring,
    stopMonitoring,
  } = useStorageStore();

  // Local handlers
  const handleRefreshStorage = async () => {
    setLocalLoading(true);
    setLocalError(null);
    try {
      await refreshStorage();
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to refresh storage';
      setLocalError(errorMessage);
    } finally {
      setLocalLoading(false);
    }
  };

  const handleToggleMonitoring = () => {
    setLocalMonitoringEnabled(!localMonitoringEnabled);
    if (localMonitoringEnabled) {
      stopMonitoring();
    } else {
      startMonitoring();
    }
  };

  const handleToggleDetails = () => {
    setLocalShowDetails(!localShowDetails);
  };

  const formatBytes = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const calculateUsagePercentage = (): number => {
    if (!storeStorageInfo) return 0;
    return ((storeStorageInfo.total_space - storeStorageInfo.available_space) / storeStorageInfo.total_space) * 100;
  };

  const getThresholdStatus = () => {
    const storageInfo = storeStorageInfo;
    if (!storageInfo) return null;

    const usagePercent = storageInfo.usage_percent;
    const thresholdStatus = storageInfo.threshold_status;
    
    if (usagePercent >= thresholdStatus.critical_threshold) {
      return {
        level: 'critical',
        message: `Storage critical: ${usagePercent.toFixed(1)}% used`,
        isCritical: true,
        isWarning: false
      };
    } else if (usagePercent >= thresholdStatus.warning_threshold) {
      return {
        level: 'warning',
        message: `Storage warning: ${usagePercent.toFixed(1)}% used`,
        isCritical: false,
        isWarning: true
      };
    } else {
      return {
        level: 'normal',
        message: `Storage normal: ${usagePercent.toFixed(1)}% used`,
        isCritical: false,
        isWarning: false
      };
    }
  };

  const getThresholdMessage = () => {
    const thresholdStatus = getThresholdStatus();
    if (!thresholdStatus) return 'Storage information unavailable';
    
    return thresholdStatus.message;
  };

  const getThresholdColor = () => {
    const thresholdStatus = getThresholdStatus();
    if (!thresholdStatus) return 'info';
    
    if (thresholdStatus.isCritical) return 'error';
    if (thresholdStatus.isWarning) return 'warning';
    return 'success';
  };

  const isStorageAvailable = (): boolean => {
    if (!storeStorageInfo) return false;
    return storeStorageInfo.usage_percent < storeStorageInfo.threshold_status.critical_threshold;
  };

  // Initialize component
  useEffect(() => {
    handleRefreshStorage();
  }, []);

  if (isLoading || localLoading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="200px">
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h5" gutterBottom>
          Storage Monitor
        </Typography>
        <Box display="flex" gap={1}>
          <Button
            variant="outlined"
            onClick={handleRefreshStorage}
            disabled={localLoading}
            startIcon={<Refresh />}
          >
            Refresh Storage
          </Button>
          <Button
            variant="outlined"
            onClick={handleToggleDetails}
            startIcon={localShowDetails ? <VisibilityOff /> : <Visibility />}
          >
            {localShowDetails ? 'Hide Details' : 'Show Details'}
          </Button>
        </Box>
      </Box>

      {/* Error Display */}
      {(localError || storeError) && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {localError || storeError}
        </Alert>
      )}

      {/* Storage Overview */}
      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Storage Overview
          </Typography>
          
          {storeStorageInfo ? (
            <Grid container spacing={2}>
              <Grid item xs={12} md={6}>
                <Box>
                  <Typography variant="body2" color="textSecondary">
                    Total Space
                  </Typography>
                  <Typography variant="h6">
                    {formatBytes(storeStorageInfo.total_space)}
                  </Typography>
                </Box>
              </Grid>
              <Grid item xs={12} md={6}>
                <Box>
                  <Typography variant="body2" color="textSecondary">
                    Used Space
                  </Typography>
                  <Typography variant="h6">
                    {formatBytes(storeStorageInfo.used_space)}
                  </Typography>
                </Box>
              </Grid>
              <Grid item xs={12} md={6}>
                <Box>
                  <Typography variant="body2" color="textSecondary">
                    Available Space
                  </Typography>
                  <Typography variant="h6">
                    {formatBytes(storeStorageInfo.available_space)}
                  </Typography>
                </Box>
              </Grid>
              <Grid item xs={12} md={6}>
                <Box>
                  <Typography variant="body2" color="textSecondary">
                    Usage Percentage
                  </Typography>
                  <Typography variant="h6">
                    {storeStorageInfo.usage_percent.toFixed(1)}%
                  </Typography>
                </Box>
              </Grid>
            </Grid>
          ) : (
            <Typography color="textSecondary">
              No storage information available
            </Typography>
          )}
        </CardContent>
      </Card>

      {/* Threshold Status */}
      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Threshold Status
          </Typography>
          
          {storeStorageInfo ? (
            <Box>
              <Box display="flex" alignItems="center" justifyContent="space-between" mb={2}>
                <Typography variant="body2">
                  Current Status
                </Typography>
                <Chip
                  label={storeStorageInfo.threshold_status.current_status}
                  color={getThresholdColor()}
                  size="small"
                />
              </Box>
              
              <Box mb={2}>
                <Typography variant="body2" color="textSecondary">
                  Warning Threshold: {storeStorageInfo.threshold_status.warning_threshold}%
                </Typography>
                <Typography variant="body2" color="textSecondary">
                  Critical Threshold: {storeStorageInfo.threshold_status.critical_threshold}%
                </Typography>
              </Box>
              
              <Alert severity={getThresholdColor()}>
                {getThresholdMessage()}
              </Alert>
            </Box>
          ) : (
            <Typography color="textSecondary">
              No threshold information available
            </Typography>
          )}
        </CardContent>
      </Card>

      {/* Storage Availability */}
      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Storage Availability
          </Typography>
          
          {storeStorageInfo ? (
            <Box>
              <Box display="flex" alignItems="center" justifyContent="space-between" mb={2}>
                <Typography variant="body2">
                  Recording Available
                </Typography>
                <Chip
                  icon={isStorageAvailable() ? <CheckCircle /> : <Error />}
                  label={isStorageAvailable() ? 'Yes' : 'No'}
                  color={isStorageAvailable() ? 'success' : 'error'}
                  size="small"
                />
              </Box>
              
              <Typography variant="body2" color="textSecondary">
                Storage is {isStorageAvailable() ? 'available' : 'unavailable'} for recording operations.
              </Typography>
            </Box>
          ) : (
            <Typography color="textSecondary">
              Storage availability unknown
            </Typography>
          )}
        </CardContent>
      </Card>

      {/* Monitoring Controls */}
      <Card>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Monitoring Controls
          </Typography>
          
          <Box display="flex" alignItems="center" justifyContent="space-between">
            <Typography variant="body2">
              Storage Monitoring
            </Typography>
            <FormControlLabel
              control={
                <Switch
                  checked={localMonitoringEnabled}
                  onChange={handleToggleMonitoring}
                  color="primary"
                />
              }
              label={localMonitoringEnabled ? 'Active' : 'Inactive'}
            />
          </Box>
          
          {storeWarnings.length > 0 && (
            <Box mt={2}>
              <Typography variant="body2" color="textSecondary" gutterBottom>
                Recent Warnings:
              </Typography>
              {storeWarnings.map((warning, index) => (
                <Alert key={index} severity="warning" sx={{ mb: 1 }}>
                  {warning}
                </Alert>
              ))}
            </Box>
          )}
        </CardContent>
      </Card>
    </Box>
  );
};

export default StorageMonitor;
