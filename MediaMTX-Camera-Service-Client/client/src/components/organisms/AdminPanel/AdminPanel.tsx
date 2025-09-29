/**
 * AdminPanel - Architecture Compliance
 * 
 * Architecture requirement: "Admin panel for retention policies" (Priority 3)
 * Provides admin-only interface for system configuration and retention policy management
 * Follows existing component patterns and architecture guidelines
 */

import React, { useState, useEffect } from 'react';
import { 
  Box, 
  Card, 
  CardContent, 
  Typography, 
  Button, 
  TextField, 
  Switch, 
  FormControlLabel, 
  Grid, 
  Alert, 
  Divider,
  Chip,
  CircularProgress
} from '@mui/material';
import { 
  Settings, 
  Storage, 
  Delete, 
  Save, 
  Refresh,
  AdminPanelSettings,
  Warning
} from '@mui/icons-material';
import { useFileStore } from '../../../stores/file/fileStore';
import { usePermissions } from '../../../hooks/usePermissions';
import { logger } from '../../../services/logger/LoggerService';

export interface AdminPanelProps {
  className?: string;
}

export const AdminPanel: React.FC<AdminPanelProps> = ({ className = '' }) => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  
  // Retention policy form state
  const [policyType, setPolicyType] = useState<'age' | 'size' | 'manual'>('age');
  const [enabled, setEnabled] = useState(false);
  const [maxAgeDays, setMaxAgeDays] = useState<number>(30);
  const [maxSizeGb, setMaxSizeGb] = useState<number>(10);
  
  // Cleanup state
  const [cleanupLoading, setCleanupLoading] = useState(false);
  const [cleanupResult, setCleanupResult] = useState<{ files_deleted: number; space_freed: number } | null>(null);

  const { setRetentionPolicy, cleanupOldFiles } = useFileStore();
  const { canViewAdminPanel, isAdmin } = usePermissions();

  useEffect(() => {
    logger.info('AdminPanel initialized');
  }, []);

  // Security check - only admin users can access
  if (!canViewAdminPanel() || !isAdmin) {
    return (
      <Box className={`admin-panel ${className}`}>
        <Alert severity="error" icon={<Warning />}>
          Access denied. Admin privileges required to access this panel.
        </Alert>
      </Box>
    );
  }

  const handleSetRetentionPolicy = async () => {
    setLoading(true);
    setError(null);
    setSuccess(null);

    try {
      logger.info('Setting retention policy', { policyType, enabled, maxAgeDays, maxSizeGb });
      
      const result = await setRetentionPolicy(
        policyType,
        enabled,
        maxAgeDays,
        maxSizeGb
      );

      setSuccess(`Retention policy set successfully: ${result.policy_type} policy ${result.enabled ? 'enabled' : 'disabled'}`);
      logger.info('Retention policy set successfully', result);
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to set retention policy';
      setError(errorMsg);
      logger.error('Failed to set retention policy', { error: errorMsg });
    } finally {
      setLoading(false);
    }
  };

  const handleCleanupOldFiles = async () => {
    setCleanupLoading(true);
    setError(null);
    setSuccess(null);

    try {
      logger.info('Starting cleanup of old files');
      
      const result = await cleanupOldFiles();
      setCleanupResult(result);
      setSuccess(`Cleanup completed: ${result.files_deleted} files deleted, ${result.space_freed} bytes freed`);
      logger.info('Cleanup completed successfully', result);
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to cleanup old files';
      setError(errorMsg);
      logger.error('Failed to cleanup old files', { error: errorMsg });
    } finally {
      setCleanupLoading(false);
    }
  };

  const handleResetForm = () => {
    setPolicyType('age');
    setEnabled(false);
    setMaxAgeDays(30);
    setMaxSizeGb(10);
    setError(null);
    setSuccess(null);
    setCleanupResult(null);
  };

  return (
    <Box className={`admin-panel ${className}`}>
      <Box sx={{ mb: 3, display: 'flex', alignItems: 'center', gap: 2 }}>
        <AdminPanelSettings sx={{ fontSize: 32, color: 'primary.main' }} />
        <Typography variant="h4" component="h1">
          Admin Panel
        </Typography>
        <Chip 
          label="Admin Only" 
          color="error" 
          size="small" 
          icon={<Settings />}
        />
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }} onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      {success && (
        <Alert severity="success" sx={{ mb: 3 }} onClose={() => setSuccess(null)}>
          {success}
        </Alert>
      )}

      <Grid container spacing={3}>
        {/* Retention Policy Configuration */}
        <Grid item xs={12} md={8}>
          <Card variant="outlined">
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', mb: 3 }}>
                <Storage sx={{ mr: 1, color: 'primary.main' }} />
                <Typography variant="h6" component="h2">
                  File Retention Policy
                </Typography>
              </Box>

              <Grid container spacing={2}>
                <Grid item xs={12}>
                  <FormControlLabel
                    control={
                      <Switch
                        checked={enabled}
                        onChange={(e) => setEnabled(e.target.checked)}
                        color="primary"
                      />
                    }
                    label="Enable Retention Policy"
                  />
                </Grid>

                <Grid item xs={12} sm={6}>
                  <TextField
                    select
                    fullWidth
                    label="Policy Type"
                    value={policyType}
                    onChange={(e) => setPolicyType(e.target.value as 'age' | 'size' | 'manual')}
                    SelectProps={{ native: true }}
                    disabled={!enabled}
                  >
                    <option value="age">Age-based (days)</option>
                    <option value="size">Size-based (GB)</option>
                    <option value="manual">Manual cleanup only</option>
                  </TextField>
                </Grid>

                {policyType === 'age' && (
                  <Grid item xs={12} sm={6}>
                    <TextField
                      fullWidth
                      type="number"
                      label="Max Age (days)"
                      value={maxAgeDays}
                      onChange={(e) => setMaxAgeDays(parseInt(e.target.value) || 30)}
                      disabled={!enabled}
                      inputProps={{ min: 1, max: 365 }}
                    />
                  </Grid>
                )}

                {policyType === 'size' && (
                  <Grid item xs={12} sm={6}>
                    <TextField
                      fullWidth
                      type="number"
                      label="Max Size (GB)"
                      value={maxSizeGb}
                      onChange={(e) => setMaxSizeGb(parseInt(e.target.value) || 10)}
                      disabled={!enabled}
                      inputProps={{ min: 1, max: 1000 }}
                    />
                  </Grid>
                )}

                <Grid item xs={12}>
                  <Box sx={{ display: 'flex', gap: 2, mt: 2 }}>
                    <Button
                      variant="contained"
                      onClick={handleSetRetentionPolicy}
                      disabled={loading}
                      startIcon={loading ? <CircularProgress size={20} /> : <Save />}
                    >
                      {loading ? 'Setting...' : 'Set Policy'}
                    </Button>
                    <Button
                      variant="outlined"
                      onClick={handleResetForm}
                      disabled={loading}
                    >
                      Reset
                    </Button>
                  </Box>
                </Grid>
              </Grid>
            </CardContent>
          </Card>
        </Grid>

        {/* Cleanup Operations */}
        <Grid item xs={12} md={4}>
          <Card variant="outlined">
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', mb: 3 }}>
                <Delete sx={{ mr: 1, color: 'error.main' }} />
                <Typography variant="h6" component="h2">
                  File Cleanup
                </Typography>
              </Box>

              <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                Manually trigger cleanup of old files based on current retention policy.
              </Typography>

              <Button
                variant="contained"
                color="error"
                fullWidth
                onClick={handleCleanupOldFiles}
                disabled={cleanupLoading}
                startIcon={cleanupLoading ? <CircularProgress size={20} /> : <Delete />}
                sx={{ mb: 2 }}
              >
                {cleanupLoading ? 'Cleaning...' : 'Cleanup Old Files'}
              </Button>

              {cleanupResult && (
                <Box sx={{ mt: 2, p: 2, bgcolor: 'grey.50', borderRadius: 1 }}>
                  <Typography variant="body2" color="text.secondary">
                    Last Cleanup Results:
                  </Typography>
                  <Typography variant="body2">
                    • Files deleted: {cleanupResult.files_deleted}
                  </Typography>
                  <Typography variant="body2">
                    • Space freed: {(cleanupResult.space_freed / 1024 / 1024).toFixed(2)} MB
                  </Typography>
                </Box>
              )}
            </CardContent>
          </Card>
        </Grid>

        {/* System Information */}
        <Grid item xs={12}>
          <Card variant="outlined">
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                <Settings sx={{ mr: 1, color: 'primary.main' }} />
                <Typography variant="h6" component="h2">
                  System Information
                </Typography>
              </Box>

              <Divider sx={{ mb: 2 }} />

              <Grid container spacing={2}>
                <Grid item xs={12} sm={6} md={3}>
                  <Box sx={{ textAlign: 'center', p: 2, bgcolor: 'grey.50', borderRadius: 1 }}>
                    <Typography variant="h6" color="primary.main">
                      Admin Panel
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      Version 1.0.0
                    </Typography>
                  </Box>
                </Grid>
                <Grid item xs={12} sm={6} md={3}>
                  <Box sx={{ textAlign: 'center', p: 2, bgcolor: 'grey.50', borderRadius: 1 }}>
                    <Typography variant="h6" color="success.main">
                      Retention Policy
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      {enabled ? 'Enabled' : 'Disabled'}
                    </Typography>
                  </Box>
                </Grid>
                <Grid item xs={12} sm={6} md={3}>
                  <Box sx={{ textAlign: 'center', p: 2, bgcolor: 'grey.50', borderRadius: 1 }}>
                    <Typography variant="h6" color="info.main">
                      Policy Type
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      {policyType.charAt(0).toUpperCase() + policyType.slice(1)}
                    </Typography>
                  </Box>
                </Grid>
                <Grid item xs={12} sm={6} md={3}>
                  <Box sx={{ textAlign: 'center', p: 2, bgcolor: 'grey.50', borderRadius: 1 }}>
                    <Typography variant="h6" color="warning.main">
                      Last Cleanup
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      {cleanupResult ? `${cleanupResult.files_deleted} files` : 'Never'}
                    </Typography>
                  </Box>
                </Grid>
              </Grid>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Box>
  );
};
