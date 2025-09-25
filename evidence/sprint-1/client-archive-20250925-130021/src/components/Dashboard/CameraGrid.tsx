import React from 'react';
import { Box, Typography } from '@mui/material';
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
    <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 3 }}>
      {cameras.map((camera) => (
        <Box sx={{ flex: '1 1 300px', minWidth: 0 }} key={camera.device}>
          <CameraCard camera={camera} />
        </Box>
      ))}
    </Box>
  );
};

export default CameraGrid; 