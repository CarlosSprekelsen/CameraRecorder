/**
 * Connection Settings Form
 * Manages WebSocket and HTTP connection settings
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
import { type ConnectionSettings } from '../../../types/settings';

interface ConnectionSettingsFormProps {
  settings: ConnectionSettings;
  onChange: (settings: ConnectionSettings) => void;
}

const ConnectionSettingsForm: React.FC<ConnectionSettingsFormProps> = ({ settings, onChange }) => {
  const handleChange = (field: keyof ConnectionSettings, value: unknown) => {
    onChange({
      ...settings,
      [field]: value,
    });
  };

  return (
    <Box>
      <Typography variant="h6" gutterBottom>
        WebSocket Connection
      </Typography>
      
      <Grid container spacing={3}>
        <Grid item xs={12} md={6}>
          <TextField
            fullWidth
            label="WebSocket URL"
            value={settings.websocketUrl}
            onChange={(e) => handleChange('websocketUrl', e.target.value)}
            helperText="WebSocket server endpoint (e.g., ws://localhost:8080)"
            variant="outlined"
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <TextField
            fullWidth
            label="HTTP Base URL"
            value={settings.httpBaseUrl}
            onChange={(e) => handleChange('httpBaseUrl', e.target.value)}
            helperText="HTTP server endpoint (e.g., http://localhost:8080)"
            variant="outlined"
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <TextField
            fullWidth
            type="number"
            label="Connection Timeout (ms)"
            value={settings.connectionTimeout}
            onChange={(e) => handleChange('connectionTimeout', parseInt(e.target.value) || 5000)}
            helperText="Maximum time to wait for connection"
            variant="outlined"
            inputProps={{ min: 1000, max: 30000 }}
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <TextField
            fullWidth
            type="number"
            label="Request Timeout (ms)"
            value={settings.requestTimeout}
            onChange={(e) => handleChange('requestTimeout', parseInt(e.target.value) || 10000)}
            helperText="Maximum time to wait for request response"
            variant="outlined"
            inputProps={{ min: 1000, max: 60000 }}
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <TextField
            fullWidth
            type="number"
            label="Reconnect Interval (ms)"
            value={settings.reconnectInterval}
            onChange={(e) => handleChange('reconnectInterval', parseInt(e.target.value) || 5000)}
            helperText="Time between reconnection attempts"
            variant="outlined"
            inputProps={{ min: 1000, max: 30000 }}
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <TextField
            fullWidth
            type="number"
            label="Max Reconnect Attempts"
            value={settings.maxReconnectAttempts}
            onChange={(e) => handleChange('maxReconnectAttempts', parseInt(e.target.value) || 5)}
            helperText="Maximum number of reconnection attempts"
            variant="outlined"
            inputProps={{ min: 1, max: 20 }}
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <TextField
            fullWidth
            type="number"
            label="Heartbeat Interval (ms)"
            value={settings.heartbeatInterval}
            onChange={(e) => handleChange('heartbeatInterval', parseInt(e.target.value) || 30000)}
            helperText="Interval for connection heartbeat"
            variant="outlined"
            inputProps={{ min: 5000, max: 120000 }}
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <FormControl fullWidth>
            <InputLabel>Connection Quality Threshold</InputLabel>
            <Select
              value={settings.qualityThreshold}
              onChange={(e) => handleChange('qualityThreshold', e.target.value)}
              label="Connection Quality Threshold"
            >
              <MenuItem value="excellent">Excellent (90%+)</MenuItem>
              <MenuItem value="good">Good (70%+)</MenuItem>
              <MenuItem value="poor">Poor (50%+)</MenuItem>
              <MenuItem value="unstable">Unstable (&lt;50%)</MenuItem>
            </Select>
          </FormControl>
        </Grid>
      </Grid>

      <Divider sx={{ my: 3 }} />

      <Typography variant="h6" gutterBottom>
        HTTP Polling Fallback
      </Typography>
      
      <Grid container spacing={3}>
        <Grid item xs={12} md={6}>
          <FormControlLabel
            control={
              <Switch
                checked={settings.enableHttpFallback}
                onChange={(e) => handleChange('enableHttpFallback', e.target.checked)}
              />
            }
            label="Enable HTTP Polling Fallback"
          />
        </Grid>
        
        {settings.enableHttpFallback && (
          <>
            <Grid item xs={12} md={6}>
              <TextField
                fullWidth
                type="number"
                label="Polling Interval (ms)"
                value={settings.pollingInterval}
                onChange={(e) => handleChange('pollingInterval', parseInt(e.target.value) || 5000)}
                helperText="Interval for HTTP polling when WebSocket is unavailable"
                variant="outlined"
                inputProps={{ min: 1000, max: 30000 }}
              />
            </Grid>
            
            <Grid item xs={12} md={6}>
              <TextField
                fullWidth
                type="number"
                label="Max Polling Duration (ms)"
                value={settings.maxPollingDuration}
                onChange={(e) => handleChange('maxPollingDuration', parseInt(e.target.value) || 300000)}
                helperText="Maximum time to use HTTP polling before giving up"
                variant="outlined"
                inputProps={{ min: 60000, max: 1800000 }}
              />
            </Grid>
          </>
        )}
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
                checked={settings.enableMetrics}
                onChange={(e) => handleChange('enableMetrics', e.target.checked)}
              />
            }
            label="Enable Connection Metrics"
          />
        </Grid>
        
        <Grid item xs={12} md={6}>
          <FormControlLabel
            control={
              <Switch
                checked={settings.enableCircuitBreaker}
                onChange={(e) => handleChange('enableCircuitBreaker', e.target.checked)}
              />
            }
            label="Enable Circuit Breaker"
          />
        </Grid>
        
        {settings.enableCircuitBreaker && (
          <>
            <Grid item xs={12} md={6}>
              <TextField
                fullWidth
                type="number"
                label="Circuit Breaker Threshold"
                value={settings.circuitBreakerThreshold}
                onChange={(e) => handleChange('circuitBreakerThreshold', parseInt(e.target.value) || 3)}
                helperText="Number of failures before opening circuit breaker"
                variant="outlined"
                inputProps={{ min: 1, max: 10 }}
              />
            </Grid>
            
            <Grid item xs={12} md={6}>
              <TextField
                fullWidth
                type="number"
                label="Circuit Breaker Timeout (ms)"
                value={settings.circuitBreakerTimeout}
                onChange={(e) => handleChange('circuitBreakerTimeout', parseInt(e.target.value) || 60000)}
                helperText="Time before attempting to close circuit breaker"
                variant="outlined"
                inputProps={{ min: 10000, max: 300000 }}
              />
            </Grid>
          </>
        )}
      </Grid>

      <Alert severity="info" sx={{ mt: 3 }}>
        Connection settings affect how the application communicates with the MediaMTX server. 
        Changes will take effect after saving and restarting the application.
      </Alert>
    </Box>
  );
};

export default ConnectionSettingsForm;
