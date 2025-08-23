/**
 * Recording Settings Form
 * Manages video recording settings
 */

import React from 'react';
import {
  Box,
  TextField,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Switch,
  FormControlLabel,
  Typography,
  Grid,
  Alert,
  Divider,
} from '@mui/material';
import { type RecordingSettings } from '../../../types/settings';

interface RecordingSettingsFormProps {
  settings: RecordingSettings;
  onChange: (settings: RecordingSettings) => void;
}

const RecordingSettingsForm: React.FC<RecordingSettingsFormProps> = ({ settings, onChange }) => {
  const handleChange = (field: keyof RecordingSettings, value: unknown) => {
    onChange({
      ...settings,
      [field]: value,
    });
  };

  return (
    <Box>
      <Typography variant="h6" gutterBottom>
        Recording Configuration
      </Typography>
      
      <Grid container spacing={3}>
        <Grid item xs={12} md={6}>
          <TextField
            fullWidth
            label="Default Recording Duration (seconds)"
            type="number"
            value={settings.defaultDuration}
            onChange={(e) => handleChange('defaultDuration', parseInt(e.target.value) || 30)}
            helperText="Default recording length when no duration specified"
            variant="outlined"
            inputProps={{ min: 5, max: 3600 }}
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <TextField
            fullWidth
            label="Max Recording Duration (seconds)"
            type="number"
            value={settings.maxDuration}
            onChange={(e) => handleChange('maxDuration', parseInt(e.target.value) || 300)}
            helperText="Maximum allowed recording duration"
            variant="outlined"
            inputProps={{ min: 10, max: 7200 }}
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <FormControl fullWidth>
            <InputLabel>Default Video Quality</InputLabel>
            <Select
              value={settings.defaultQuality}
              onChange={(e) => handleChange('defaultQuality', e.target.value)}
              label="Default Video Quality"
            >
              <MenuItem value="low">Low (480p)</MenuItem>
              <MenuItem value="medium">Medium (720p)</MenuItem>
              <MenuItem value="high">High (1080p)</MenuItem>
              <MenuItem value="ultra">Ultra (4K)</MenuItem>
            </Select>
          </FormControl>
        </Grid>
        
        <Grid item xs={12} md={6}>
          <FormControl fullWidth>
            <InputLabel>Default Video Format</InputLabel>
            <Select
              value={settings.defaultFormat}
              onChange={(e) => handleChange('defaultFormat', e.target.value)}
              label="Default Video Format"
            >
              <MenuItem value="mp4">MP4</MenuItem>
              <MenuItem value="avi">AVI</MenuItem>
              <MenuItem value="mov">MOV</MenuItem>
              <MenuItem value="mkv">MKV</MenuItem>
            </Select>
          </FormControl>
        </Grid>
        
        <Grid item xs={12} md={6}>
          <TextField
            fullWidth
            label="Default Frame Rate (fps)"
            type="number"
            value={settings.defaultFrameRate}
            onChange={(e) => handleChange('defaultFrameRate', parseInt(e.target.value) || 30)}
            helperText="Default frames per second for recordings"
            variant="outlined"
            inputProps={{ min: 1, max: 60 }}
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <TextField
            fullWidth
            label="Default Bitrate (kbps)"
            type="number"
            value={settings.defaultBitrate}
            onChange={(e) => handleChange('defaultBitrate', parseInt(e.target.value) || 2000)}
            helperText="Default video bitrate in kilobits per second"
            variant="outlined"
            inputProps={{ min: 100, max: 10000 }}
          />
        </Grid>
      </Grid>

      <Divider sx={{ my: 3 }} />

      <Typography variant="h6" gutterBottom>
        Storage Settings
      </Typography>
      
      <Grid container spacing={3}>
        <Grid item xs={12} md={6}>
          <TextField
            fullWidth
            label="Storage Directory"
            value={settings.storageDirectory}
            onChange={(e) => handleChange('storageDirectory', e.target.value)}
            helperText="Directory where recordings are saved"
            variant="outlined"
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <TextField
            fullWidth
            label="Max Storage Size (GB)"
            type="number"
            value={settings.maxStorageSize}
            onChange={(e) => handleChange('maxStorageSize', parseInt(e.target.value) || 10)}
            helperText="Maximum storage space for recordings"
            variant="outlined"
            inputProps={{ min: 1, max: 1000 }}
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <TextField
            fullWidth
            label="Auto Cleanup Age (days)"
            type="number"
            value={settings.autoCleanupAge}
            onChange={(e) => handleChange('autoCleanupAge', parseInt(e.target.value) || 30)}
            helperText="Automatically delete recordings older than this"
            variant="outlined"
            inputProps={{ min: 1, max: 365 }}
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <FormControlLabel
            control={
              <Switch
                checked={settings.enableAutoCleanup}
                onChange={(e) => handleChange('enableAutoCleanup', e.target.checked)}
              />
            }
            label="Enable Auto Cleanup"
          />
        </Grid>
      </Grid>

      <Divider sx={{ my: 3 }} />

      <Typography variant="h6" gutterBottom>
        Advanced Options
      </Typography>
      
      <Grid container spacing={3}>
        <Grid item xs={12} md={6}>
          <FormControlLabel
            control={
              <Switch
                checked={settings.enableAudio}
                onChange={(e) => handleChange('enableAudio', e.target.checked)}
              />
            }
            label="Enable Audio Recording"
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <FormControlLabel
            control={
              <Switch
                checked={settings.enableWatermark}
                onChange={(e) => handleChange('enableWatermark', e.target.checked)}
              />
            }
            label="Enable Watermark"
          />
        </Grid>
        
        {settings.enableWatermark && (
          <Grid item xs={12} md={6}>
            <TextField
              fullWidth
              label="Watermark Text"
              value={settings.watermarkText}
              onChange={(e) => handleChange('watermarkText', e.target.value)}
              helperText="Text to display as watermark on recordings"
              variant="outlined"
            />
          </Grid>
        )}
        
        <Grid item xs={12} md={6}>
          <FormControlLabel
            control={
              <Switch
                checked={settings.enableCompression}
                onChange={(e) => handleChange('enableCompression', e.target.checked)}
              />
            }
            label="Enable Video Compression"
          />
        </Grid>
      </Grid>

      <Alert severity="info" sx={{ mt: 3 }}>
        Recording settings control how video recordings are captured and stored. 
        Quality and format settings affect file size and playback compatibility.
      </Alert>
    </Box>
  );
};

export default RecordingSettingsForm;
