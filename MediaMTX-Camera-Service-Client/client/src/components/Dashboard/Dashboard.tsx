import React from 'react';
import { Grid, Container, Typography, Box } from '@mui/material';
import CameraGrid from '../CameraGrid/CameraGrid';
import HealthMonitor from '../HealthMonitor/HealthMonitor';
import FileManager from '../FileManager/FileManager';
import ConnectionStatus from '../ConnectionStatus/ConnectionStatus';
import RecordingManager from '../RecordingManager/RecordingManager';
import StorageMonitor from '../StorageMonitor/StorageMonitor';
import ConfigurationManager from '../ConfigurationManager/ConfigurationManager';
import ErrorHandler from '../ErrorHandler/ErrorHandler';

const Dashboard: React.FC = () => {
  return (
    <Container maxWidth="xl">
      <Box sx={{ py: 3 }}>
        <Typography variant="h4" gutterBottom>
          MediaMTX Camera Service Dashboard
        </Typography>
        <Typography variant="body1" color="textSecondary" sx={{ mb: 3 }}>
          Real-time camera management and monitoring system
        </Typography>

        <Grid container spacing={3}>
          {/* Connection Status */}
          <Grid item xs={12}>
            <ConnectionStatus />
          </Grid>

          {/* Health Monitor */}
          <Grid item xs={12} md={6}>
            <HealthMonitor />
          </Grid>

          {/* Storage Monitor */}
          <Grid item xs={12} md={6}>
            <StorageMonitor />
          </Grid>

          {/* Camera Grid */}
          <Grid item xs={12}>
            <CameraGrid />
          </Grid>

          {/* Recording Manager */}
          <Grid item xs={12}>
            <RecordingManager />
          </Grid>

          {/* Configuration Manager */}
          <Grid item xs={12} md={6}>
            <ConfigurationManager />
          </Grid>

          {/* Error Handler */}
          <Grid item xs={12} md={6}>
            <ErrorHandler />
          </Grid>

          {/* File Manager */}
          <Grid item xs={12}>
            <FileManager />
          </Grid>
        </Grid>

        {/* Sprint Status */}
        <Box sx={{ mt: 4, p: 2, bgcolor: 'background.paper', borderRadius: 1 }}>
          <Typography variant="body2" color="textSecondary">
            🚀 Sprint Status: Architecture refactoring complete - 100% server API alignment achieved
          </Typography>
        </Box>
      </Box>
    </Container>
  );
};

export default Dashboard; 