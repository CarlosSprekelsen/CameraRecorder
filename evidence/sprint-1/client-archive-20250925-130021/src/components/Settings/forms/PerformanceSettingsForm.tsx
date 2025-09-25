/**
 * Performance Settings Form
 * Manages performance and optimization settings
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
import { type PerformanceSettings } from '../../../types/settings';

interface PerformanceSettingsFormProps {
  settings: PerformanceSettings;
  onChange: (settings: PerformanceSettings) => void;
}

const PerformanceSettingsForm: React.FC<PerformanceSettingsFormProps> = ({ settings, onChange }) => {
  const handleChange = (field: keyof PerformanceSettings, value: unknown) => {
    onChange({
      ...settings,
      [field]: value,
    });
  };

  return (
    <Box>
      <Typography variant="h6" gutterBottom>
        Performance Configuration
      </Typography>
      
      <Grid container spacing={3}>
        <Grid item xs={12} md={6}>
          <FormControlLabel
            control={
              <Switch
                checked={settings.enableCaching}
                onChange={(e) => handleChange('enableCaching', e.target.checked)}
              />
            }
            label="Enable Caching"
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <TextField
            fullWidth
            label="Cache Size (MB)"
            type="number"
            value={settings.cacheSize}
            onChange={(e) => handleChange('cacheSize', parseInt(e.target.value) || 100)}
            helperText="Maximum cache size in megabytes"
            variant="outlined"
            inputProps={{ min: 10, max: 1000 }}
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <FormControlLabel
            control={
              <Switch
                checked={settings.enableCompression}
                onChange={(e) => handleChange('enableCompression', e.target.checked)}
              />
            }
            label="Enable Compression"
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <TextField
            fullWidth
            label="Max Concurrent Downloads"
            type="number"
            value={settings.maxConcurrentDownloads}
            onChange={(e) => handleChange('maxConcurrentDownloads', parseInt(e.target.value) || 3)}
            helperText="Maximum number of simultaneous downloads"
            variant="outlined"
            inputProps={{ min: 1, max: 10 }}
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <FormControlLabel
            control={
              <Switch
                checked={settings.enableBackgroundSync}
                onChange={(e) => handleChange('enableBackgroundSync', e.target.checked)}
              />
            }
            label="Enable Background Sync"
          />
        </Grid>
      </Grid>

      <Alert severity="info" sx={{ mt: 3 }}>
        Performance settings control how the application optimizes for speed and resource usage. 
        Changes may affect memory usage and network performance.
      </Alert>
    </Box>
  );
};

export default PerformanceSettingsForm;
