/**
 * Interface Settings Form
 * Manages UI and interface settings
 */

import React from 'react';
import {
  Box,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Switch,
  FormControlLabel,
  Typography,
  Grid,
  Alert,
  TextField,
} from '@mui/material';
import { type UISettings } from '../../../types/settings';

interface InterfaceSettingsFormProps {
  settings: UISettings;
  onChange: (settings: UISettings) => void;
}

const InterfaceSettingsForm: React.FC<InterfaceSettingsFormProps> = ({ settings, onChange }) => {
  const handleChange = (field: keyof UISettings, value: unknown) => {
    onChange({
      ...settings,
      [field]: value,
    });
  };

  return (
    <Box>
      <Typography variant="h6" gutterBottom>
        Interface Configuration
      </Typography>
      
      <Grid container spacing={3}>
        <Grid item xs={12} md={6}>
          <FormControl fullWidth>
            <InputLabel>Theme</InputLabel>
            <Select
              value={settings.theme}
              onChange={(e) => handleChange('theme', e.target.value)}
              label="Theme"
            >
              <MenuItem value="light">Light</MenuItem>
              <MenuItem value="dark">Dark</MenuItem>
              <MenuItem value="auto">Auto (System)</MenuItem>
            </Select>
          </FormControl>
        </Grid>
        
        <Grid item xs={12} md={6}>
          <FormControl fullWidth>
            <InputLabel>Language</InputLabel>
            <Select
              value={settings.language}
              onChange={(e) => handleChange('language', e.target.value)}
              label="Language"
            >
              <MenuItem value="en">English</MenuItem>
              <MenuItem value="es">Español</MenuItem>
              <MenuItem value="fr">Français</MenuItem>
              <MenuItem value="de">Deutsch</MenuItem>
            </Select>
          </FormControl>
        </Grid>
        
        <Grid item xs={12} md={6}>
          <TextField
            fullWidth
            label="Refresh Interval (seconds)"
            type="number"
            value={settings.refreshInterval}
            onChange={(e) => handleChange('refreshInterval', parseInt(e.target.value) || 30)}
            helperText="How often to refresh camera data"
            variant="outlined"
            inputProps={{ min: 5, max: 300 }}
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <FormControlLabel
            control={
              <Switch
                checked={settings.autoRefresh}
                onChange={(e) => handleChange('autoRefresh', e.target.checked)}
              />
            }
            label="Enable Auto Refresh"
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <FormControlLabel
            control={
              <Switch
                checked={settings.showNotifications}
                onChange={(e) => handleChange('showNotifications', e.target.checked)}
              />
            }
            label="Show System Notifications"
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <FormControlLabel
            control={
              <Switch
                checked={settings.compactMode}
                onChange={(e) => handleChange('compactMode', e.target.checked)}
              />
            }
            label="Compact Mode"
          />
        </Grid>
      </Grid>

      <Alert severity="info" sx={{ mt: 3 }}>
        Interface settings control the appearance and behavior of the user interface. 
        Changes take effect immediately.
      </Alert>
    </Box>
  );
};

export default InterfaceSettingsForm;
