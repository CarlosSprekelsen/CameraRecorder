/**
 * Notification Settings Form
 * Manages notification preferences
 */

import React from 'react';
import {
  Box,
  Switch,
  FormControlLabel,
  Typography,
  Grid,
  Alert,
  TextField,
  Divider,
} from '@mui/material';
import { type NotificationSettings } from '../../../types/settings';

interface NotificationSettingsFormProps {
  settings: NotificationSettings;
  onChange: (settings: NotificationSettings) => void;
}

const NotificationSettingsForm: React.FC<NotificationSettingsFormProps> = ({ settings, onChange }) => {
  const handleChange = (field: keyof NotificationSettings, value: unknown) => {
    onChange({
      ...settings,
      [field]: value,
    });
  };

  const handleNotificationTypeChange = (type: keyof NotificationSettings['notificationTypes'], value: boolean) => {
    onChange({
      ...settings,
      notificationTypes: {
        ...settings.notificationTypes,
        [type]: value,
      },
    });
  };

  return (
    <Box>
      <Typography variant="h6" gutterBottom>
        Notification Preferences
      </Typography>
      
      <Grid container spacing={3}>
        <Grid item xs={12} md={6}>
          <FormControlLabel
            control={
              <Switch
                checked={settings.enabled}
                onChange={(e) => handleChange('enabled', e.target.checked)}
              />
            }
            label="Enable Notifications"
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <FormControlLabel
            control={
              <Switch
                checked={settings.soundEnabled}
                onChange={(e) => handleChange('soundEnabled', e.target.checked)}
              />
            }
            label="Enable Sound"
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <FormControlLabel
            control={
              <Switch
                checked={settings.desktopNotifications}
                onChange={(e) => handleChange('desktopNotifications', e.target.checked)}
              />
            }
            label="Desktop Notifications"
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <FormControlLabel
            control={
              <Switch
                checked={settings.emailNotifications}
                onChange={(e) => handleChange('emailNotifications', e.target.checked)}
              />
            }
            label="Email Notifications"
          />
        </Grid>
        
        {settings.emailNotifications && (
          <Grid item xs={12} md={6}>
            <TextField
              fullWidth
              label="Email Address"
              type="email"
              value={settings.emailAddress}
              onChange={(e) => handleChange('emailAddress', e.target.value)}
              helperText="Email address for notifications"
              variant="outlined"
            />
          </Grid>
        )}
      </Grid>

      <Divider sx={{ my: 3 }} />

      <Typography variant="h6" gutterBottom>
        Notification Types
      </Typography>
      
      <Grid container spacing={3}>
        <Grid item xs={12} md={6}>
          <FormControlLabel
            control={
              <Switch
                checked={settings.notificationTypes.cameraStatus}
                onChange={(e) => handleNotificationTypeChange('cameraStatus', e.target.checked)}
              />
            }
            label="Camera Status Changes"
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <FormControlLabel
            control={
              <Switch
                checked={settings.notificationTypes.recordingEvents}
                onChange={(e) => handleNotificationTypeChange('recordingEvents', e.target.checked)}
              />
            }
            label="Recording Events"
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <FormControlLabel
            control={
              <Switch
                checked={settings.notificationTypes.systemAlerts}
                onChange={(e) => handleNotificationTypeChange('systemAlerts', e.target.checked)}
              />
            }
            label="System Alerts"
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <FormControlLabel
            control={
              <Switch
                checked={settings.notificationTypes.fileOperations}
                onChange={(e) => handleNotificationTypeChange('fileOperations', e.target.checked)}
              />
            }
            label="File Operations"
          />
        </Grid>
      </Grid>

      <Alert severity="info" sx={{ mt: 3 }}>
        Configure which types of notifications you want to receive. 
        Desktop notifications require browser permission.
      </Alert>
    </Box>
  );
};

export default NotificationSettingsForm;
