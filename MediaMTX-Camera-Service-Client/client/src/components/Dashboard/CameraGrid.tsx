import React from 'react';
import { Box, Typography, Grid } from '@mui/material';
import type { CameraDevice } from '../../types';
import CameraCard from './CameraCard';

interface CameraGridProps {
  cameras: CameraDevice[];
}

const CameraGrid: React.FC<CameraGridProps> = ({ cameras }) => {
  if (cameras.length === 0) {
    return (
      <Box sx={{ textAlign: 'center', py: 8 }}>
        <Typography variant="h6" color="text.secondary" gutterBottom>
          No cameras available
        </Typography>
        <Typography variant="body2" color="text.secondary">
          Connect a camera to get started
        </Typography>
      </Box>
    );
  }

  return (
    <Grid container spacing={3}>
      {cameras.map((camera) => (
        <Grid item xs={12} sm={6} md={4} lg={3} key={camera.device}>
          <CameraCard camera={camera} />
        </Grid>
      ))}
    </Grid>
  );
};

export default CameraGrid; 