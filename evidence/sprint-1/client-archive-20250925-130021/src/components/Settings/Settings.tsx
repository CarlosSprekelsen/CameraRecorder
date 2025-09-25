/**
 * Settings Component
 * Comprehensive application settings interface
 */

import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Tabs,
  Tab,
  Button,
  Alert,
  CircularProgress,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Grid,
  Chip,
  Stack,
} from '@mui/material';
import {
  Save as SaveIcon,
  Refresh as ResetIcon,
  Download as ExportIcon,
  Upload as ImportIcon,
  Settings as SettingsIcon,
} from '@mui/icons-material';
import { useSettingsStore } from '../../stores/settingsStore';
import { useNotifications } from '../common/NotificationSystem';
import { SETTINGS_CATEGORIES, type SettingsCategory } from '../../types/settings';
import ConnectionSettingsForm from './forms/ConnectionSettingsForm';
import RecordingSettingsForm from './forms/RecordingSettingsForm';
import SnapshotSettingsForm from './forms/SnapshotSettingsForm';
import InterfaceSettingsForm from './forms/InterfaceSettingsForm';
import NotificationSettingsForm from './forms/NotificationSettingsForm';
import SecuritySettingsForm from './forms/SecuritySettingsForm';
import PerformanceSettingsForm from './forms/PerformanceSettingsForm';

/**
 * Settings Component Props
 */
interface SettingsProps {
  onClose?: () => void;
}

/**
 * Settings Component
 */
const Settings: React.FC<SettingsProps> = ({ onClose }) => {
  const [activeTab, setActiveTab] = useState<SettingsCategory>('connection');
  const [showResetDialog, setShowResetDialog] = useState(false);
  const [showImportDialog, setShowImportDialog] = useState(false);
  const [importText, setImportText] = useState('');

  const { showSuccess, showError } = useNotifications();

  const {
    settings: storeSettings,
    isLoading: storeIsLoading,
    isSaving: storeIsSaving,
    error: storeError,
    hasUnsavedChanges: storeHasUnsavedChanges,
    loadSettings: storeLoadSettings,
    saveSettings: storeSaveSettings,
    resetSettings: storeResetSettings,
    updateSettings: storeUpdateSettings,
    exportSettings: storeExportSettings,
    importSettings: storeImportSettings,
    clearError: storeClearError,
  } = useSettingsStore();

  // Load settings on mount
  useEffect(() => {
    storeLoadSettings();
  }, [storeLoadSettings]);

  // Handle tab change
  const handleTabChange = (event: React.SyntheticEvent, newValue: SettingsCategory) => {
    setActiveTab(newValue);
  };

  // Handle save settings
  const handleSave = async () => {
    try {
      await storeSaveSettings();
      showSuccess('Settings Saved', 'Your settings have been saved successfully');
    } catch (error: unknown) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to save settings';
      showError('Save Failed', errorMessage);
    }
  };

  // Handle reset settings
  const handleReset = () => {
    storeResetSettings();
    setShowResetDialog(false);
    showSuccess('Settings Reset', 'Settings have been reset to defaults');
  };

  // Handle export settings
  const handleExport = () => {
    const exportedSettings = storeExportSettings();
    const blob = new Blob([exportedSettings], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `camera-app-settings-${new Date().toISOString().split('T')[0]}.json`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    showSuccess('Settings Exported', 'Settings have been exported successfully');
  };

  // Handle import settings
  const handleImport = async () => {
    if (!importText.trim()) {
      showError('Import Failed', 'Please provide settings data to import');
      return;
    }

    try {
      const success = await storeImportSettings(importText);
      if (success) {
        setShowImportDialog(false);
        setImportText('');
        showSuccess('Settings Imported', 'Settings have been imported successfully');
      }
    } catch (error: unknown) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to import settings';
      showError('Import Failed', errorMessage);
    }
  };

  if (storeIsLoading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="400px">
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box sx={{ p: 3 }}>
      {/* Header */}
      <Box sx={{ mb: 3 }}>
        <Box display="flex" alignItems="center" gap={2} mb={2}>
          <SettingsIcon fontSize="large" />
          <Typography variant="h4" component="h1">
            Settings
          </Typography>
        </Box>
        
        <Typography variant="body1" color="text.secondary">
          Configure application preferences and connection settings
        </Typography>
      </Box>

      {/* Error Display */}
      {storeError && (
        <Alert severity="error" sx={{ mb: 3 }} onClose={storeClearError}>
          {storeError}
        </Alert>
      )}

      {/* Action Buttons */}
      <Box sx={{ mb: 3 }}>
        <Stack direction="row" spacing={2} alignItems="center">
          <Button
            variant="contained"
            startIcon={<SaveIcon />}
            onClick={handleSave}
            disabled={storeIsSaving || !storeHasUnsavedChanges}
          >
                          {storeIsSaving ? <CircularProgress size={20} /> : 'Save Settings'}
          </Button>
          
                      <Button
              variant="outlined"
              startIcon={<ResetIcon />}
              onClick={() => setShowResetDialog(true)}
              disabled={storeIsSaving}
            >
            Reset to Defaults
          </Button>
          
                      <Button
              variant="outlined"
              startIcon={<ExportIcon />}
              onClick={handleExport}
              disabled={storeIsSaving}
            >
            Export
          </Button>
          
                      <Button
              variant="outlined"
              startIcon={<ImportIcon />}
              onClick={() => setShowImportDialog(true)}
              disabled={storeIsSaving}
            >
            Import
          </Button>

          {storeHasUnsavedChanges && (
            <Chip
              label="Unsaved Changes"
              color="warning"
              size="small"
            />
          )}
        </Stack>
      </Box>

      {/* Settings Tabs */}
      <Card>
        <CardContent sx={{ p: 0 }}>
          <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
            <Tabs
              value={activeTab}
              onChange={handleTabChange}
              variant="scrollable"
              scrollButtons="auto"
            >
              {Object.entries(SETTINGS_CATEGORIES).map(([key, category]) => (
                <Tab
                  key={key}
                  label={
                    <Box display="flex" alignItems="center" gap={1}>
                      <span>{category.icon}</span>
                      <span>{category.title}</span>
                    </Box>
                  }
                  value={key as SettingsCategory}
                />
              ))}
            </Tabs>
          </Box>

          {/* Tab Content */}
          <Box sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              {SETTINGS_CATEGORIES[activeTab].title}
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
              {SETTINGS_CATEGORIES[activeTab].description}
            </Typography>
            
            {/* Connection Settings */}
            {activeTab === 'connection' && (
                          <ConnectionSettingsForm
              settings={storeSettings.connection}
              onChange={(connectionSettings) => storeUpdateSettings('connection', connectionSettings)}
            />
            )}

            {/* Recording Settings */}
            {activeTab === 'recording' && (
                          <RecordingSettingsForm
              settings={storeSettings.recording}
              onChange={(recordingSettings) => storeUpdateSettings('recording', recordingSettings)}
            />
            )}

            {/* Snapshot Settings */}
            {activeTab === 'snapshot' && (
              <SnapshotSettingsForm 
                settings={storeSettings.snapshot}
                onChange={(snapshotSettings) => storeUpdateSettings('snapshot', snapshotSettings)}
              />
            )}

            {/* Interface Settings */}
            {activeTab === 'ui' && (
              <InterfaceSettingsForm 
                settings={storeSettings.ui}
                onChange={(uiSettings) => storeUpdateSettings('ui', uiSettings)}
              />
            )}

            {/* Notification Settings */}
            {activeTab === 'notifications' && (
              <NotificationSettingsForm 
                settings={storeSettings.notifications}
                onChange={(notificationSettings) => storeUpdateSettings('notifications', notificationSettings)}
              />
            )}

            {/* Security Settings */}
            {activeTab === 'security' && (
              <SecuritySettingsForm 
                settings={storeSettings.security}
                onChange={(securitySettings) => storeUpdateSettings('security', securitySettings)}
              />
            )}

            {/* Performance Settings */}
            {activeTab === 'performance' && (
              <PerformanceSettingsForm 
                settings={storeSettings.performance}
                onChange={(performanceSettings) => storeUpdateSettings('performance', performanceSettings)}
              />
            )}
          </Box>
        </CardContent>
      </Card>

      {/* Reset Confirmation Dialog */}
      <Dialog open={showResetDialog} onClose={() => setShowResetDialog(false)}>
        <DialogTitle>Reset Settings</DialogTitle>
        <DialogContent>
          <Typography>
            Are you sure you want to reset all settings to their default values? This action cannot be undone.
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setShowResetDialog(false)}>Cancel</Button>
          <Button onClick={handleReset} color="error" variant="contained">
            Reset
          </Button>
        </DialogActions>
      </Dialog>

      {/* Import Dialog */}
      <Dialog 
        open={showImportDialog} 
        onClose={() => setShowImportDialog(false)}
        maxWidth="md"
        fullWidth
      >
        <DialogTitle>Import Settings</DialogTitle>
        <DialogContent>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
            Paste your exported settings JSON data below:
          </Typography>
          <TextField
            multiline
            rows={10}
            fullWidth
            value={importText}
            onChange={(e) => setImportText(e.target.value)}
            placeholder="Paste settings JSON here..."
            variant="outlined"
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setShowImportDialog(false)}>Cancel</Button>
          <Button onClick={handleImport} variant="contained">
            Import
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default Settings;
