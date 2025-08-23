/**
 * Snapshot Settings Form
 * Manages image snapshot settings
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
} from '@mui/material';
import { type SnapshotSettings } from '../../../types/settings';

interface SnapshotSettingsFormProps {
  settings: SnapshotSettings;
  onChange: (settings: SnapshotSettings) => void;
}

const SnapshotSettingsForm: React.FC<SnapshotSettingsFormProps> = ({ settings, onChange }) => {
  const handleChange = (field: keyof SnapshotSettings, value: unknown) => {
    onChange({
      ...settings,
      [field]: value,
    });
  };

  return (
    <Box>
      <Typography variant="h6" gutterBottom>
        Snapshot Configuration
      </Typography>
      
      <Grid container spacing={3}>
        <Grid item xs={12} md={6}>
          <FormControl fullWidth>
            <InputLabel>Default Image Format</InputLabel>
            <Select
              value={settings.defaultFormat}
              onChange={(e) => handleChange('defaultFormat', e.target.value)}
              label="Default Image Format"
            >
              <MenuItem value="jpeg">JPEG</MenuItem>
              <MenuItem value="png">PNG</MenuItem>
              <MenuItem value="bmp">BMP</MenuItem>
              <MenuItem value="webp">WebP</MenuItem>
            </Select>
          </FormControl>
        </Grid>
        
        <Grid item xs={12} md={6}>
          <TextField
            fullWidth
            label="JPEG Quality (1-100)"
            type="number"
            value={settings.jpegQuality}
            onChange={(e) => handleChange('jpegQuality', parseInt(e.target.value) || 85)}
            helperText="Quality setting for JPEG images"
            variant="outlined"
            inputProps={{ min: 1, max: 100 }}
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <TextField
            fullWidth
            label="Default Resolution Width"
            type="number"
            value={settings.defaultWidth}
            onChange={(e) => handleChange('defaultWidth', parseInt(e.target.value) || 1920)}
            helperText="Default image width in pixels"
            variant="outlined"
            inputProps={{ min: 320, max: 7680 }}
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <TextField
            fullWidth
            label="Default Resolution Height"
            type="number"
            value={settings.defaultHeight}
            onChange={(e) => handleChange('defaultHeight', parseInt(e.target.value) || 1080)}
            helperText="Default image height in pixels"
            variant="outlined"
            inputProps={{ min: 240, max: 4320 }}
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <FormControlLabel
            control={
              <Switch
                checked={settings.enableTimestamp}
                onChange={(e) => handleChange('enableTimestamp', e.target.checked)}
              />
            }
            label="Add Timestamp to Images"
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
            label="Add Watermark to Images"
          />
        </Grid>
      </Grid>

      <Alert severity="info" sx={{ mt: 3 }}>
        Snapshot settings control how still images are captured and saved. 
        Format and quality settings affect file size and image clarity.
      </Alert>
    </Box>
  );
};

export default SnapshotSettingsForm;
