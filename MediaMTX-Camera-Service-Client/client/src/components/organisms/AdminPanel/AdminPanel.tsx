/**
 * AdminPanel - Architecture Compliance
 * 
 * Architecture requirement: "Admin panel for retention policies" (Priority 3)
 * Provides admin-only interface for system configuration and retention policy management
 * Follows existing component patterns and architecture guidelines
 */

import React, { useState, useEffect } from 'react';
import { Box } from '../../atoms/Box/Box';
import { Card } from '../../atoms/Card/Card';
import { CardContent } from '../../atoms/CardContent/CardContent';
import { Typography } from '../../atoms/Typography/Typography';
import { Button } from '../../atoms/Button/Button';
import { TextField } from '../../atoms/TextField/TextField';
import { Switch } from '../../atoms/Switch/Switch';
import { FormControlLabel } from '../../atoms/FormControlLabel/FormControlLabel';
import { Grid } from '../../atoms/Grid/Grid';
import { Alert } from '../../atoms/Alert/Alert';
import { Divider } from '../../atoms/Divider/Divider';
import { Chip } from '../../atoms/Chip/Chip';
import { Icon } from '../../atoms/Icon/Icon';
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
        <Alert variant="error">
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
      <Box className="mb-3 flex items-center gap-2">
        <Icon name="admin" size={32} color="#1976d2" />
        <Typography variant="h4" component="h1">
          Admin Panel
        </Typography>
        <Chip 
          label="Admin Only" 
          color="error" 
          size="small" 
          icon={<Icon name="settings" size={16} />}
        />
      </Box>

      {error && (
        <Alert variant="error" className="mb-3">
          {error}
        </Alert>
      )}

      {success && (
        <Alert variant="success" className="mb-3">
          {success}
        </Alert>
      )}

      <Grid container spacing={3}>
        {/* Retention Policy Configuration */}
        <Grid item xs={12} md={8}>
          <Card variant="outlined">
            <CardContent>
              <Box className="flex items-center mb-3">
                <Icon name="storage" size={20} color="#1976d2" />
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
                        onChange={setEnabled}
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
                    onChange={(value) => setPolicyType(value as 'age' | 'size' | 'manual')}
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
                      onChange={(value) => setMaxAgeDays(parseInt(value) || 30)}
                      disabled={!enabled}
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
                      onChange={(value) => setMaxSizeGb(parseInt(value) || 10)}
                      disabled={!enabled}
                    />
                  </Grid>
                )}

                <Grid item xs={12}>
                  <Box className="flex gap-2 mt-2">
                    <Button
                      variant="primary"
                      onClick={handleSetRetentionPolicy}
                      disabled={loading}
                      loading={loading}
                    >
                      {loading ? 'Setting...' : 'Set Policy'}
                    </Button>
                    <Button
                      variant="secondary"
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
              <Box className="flex items-center mb-3">
                <Icon name="delete" size={20} color="#d32f2f" />
                <Typography variant="h6" component="h2">
                  File Cleanup
                </Typography>
              </Box>

              <Typography variant="body2" color="secondary" className="mb-2">
                Manually trigger cleanup of old files based on current retention policy.
              </Typography>

              <Button
                variant="danger"
                onClick={handleCleanupOldFiles}
                disabled={cleanupLoading}
                loading={cleanupLoading}
                className="mb-2"
              >
                {cleanupLoading ? 'Cleaning...' : 'Cleanup Old Files'}
              </Button>

              {cleanupResult && (
                <Box className="mt-2 p-2 bg-gray-50 rounded">
                  <Typography variant="body2" color="secondary">
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
              <Box className="flex items-center mb-2">
                <Icon name="settings" size={20} color="#1976d2" />
                <Typography variant="h6" component="h2">
                  System Information
                </Typography>
              </Box>

              <Divider className="mb-2" />

              <Grid container spacing={2}>
                <Grid item xs={12} sm={6} md={3}>
                  <Box className="text-center p-2 bg-gray-50 rounded">
                    <Typography variant="h6" color="primary">
                      Admin Panel
                    </Typography>
                    <Typography variant="body2" color="secondary">
                      Version 1.0.0
                    </Typography>
                  </Box>
                </Grid>
                <Grid item xs={12} sm={6} md={3}>
                  <Box className="text-center p-2 bg-gray-50 rounded">
                    <Typography variant="h6" color="success">
                      Retention Policy
                    </Typography>
                    <Typography variant="body2" color="secondary">
                      {enabled ? 'Enabled' : 'Disabled'}
                    </Typography>
                  </Box>
                </Grid>
                <Grid item xs={12} sm={6} md={3}>
                  <Box className="text-center p-2 bg-gray-50 rounded">
                    <Typography variant="h6" color="primary">
                      Policy Type
                    </Typography>
                    <Typography variant="body2" color="secondary">
                      {policyType.charAt(0).toUpperCase() + policyType.slice(1)}
                    </Typography>
                  </Box>
                </Grid>
                <Grid item xs={12} sm={6} md={3}>
                  <Box className="text-center p-2 bg-gray-50 rounded">
                    <Typography variant="h6" color="warning">
                      Last Cleanup
                    </Typography>
                    <Typography variant="body2" color="secondary">
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
