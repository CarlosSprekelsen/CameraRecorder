/**
 * Security Settings Form
 * Manages security and authentication settings
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
} from '@mui/material';
import { type SecuritySettings } from '../../../types/settings';

interface SecuritySettingsFormProps {
  settings: SecuritySettings;
  onChange: (settings: SecuritySettings) => void;
}

const SecuritySettingsForm: React.FC<SecuritySettingsFormProps> = ({ settings, onChange }) => {
  const handleChange = (field: keyof SecuritySettings, value: unknown) => {
    onChange({
      ...settings,
      [field]: value,
    });
  };

  return (
    <Box>
      <Typography variant="h6" gutterBottom>
        Security Configuration
      </Typography>
      
      <Grid container spacing={3}>
        <Grid item xs={12} md={6}>
          <FormControlLabel
            control={
              <Switch
                checked={settings.autoLogout}
                onChange={(e) => handleChange('autoLogout', e.target.checked)}
              />
            }
            label="Auto Logout"
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <TextField
            fullWidth
            label="Session Timeout (minutes)"
            type="number"
            value={Math.floor(settings.sessionTimeout / 60000)}
            onChange={(e) => handleChange('sessionTimeout', parseInt(e.target.value) * 60000 || 3600000)}
            helperText="Session timeout in minutes"
            variant="outlined"
            inputProps={{ min: 5, max: 480 }}
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <FormControlLabel
            control={
              <Switch
                checked={settings.rememberCredentials}
                onChange={(e) => handleChange('rememberCredentials', e.target.checked)}
              />
            }
            label="Remember Credentials"
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <FormControlLabel
            control={
              <Switch
                checked={settings.requireReauthForSensitive}
                onChange={(e) => handleChange('requireReauthForSensitive', e.target.checked)}
              />
            }
            label="Require Re-authentication for Sensitive Operations"
          />
        </Grid>
      </Grid>

      <Alert severity="warning" sx={{ mt: 3 }}>
        Security settings affect how the application handles authentication and sessions. 
        Changes may require you to log in again.
      </Alert>
    </Box>
  );
};

export default SecuritySettingsForm;
