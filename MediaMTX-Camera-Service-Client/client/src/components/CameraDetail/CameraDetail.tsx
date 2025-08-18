import React from 'react';
import { useParams, Navigate } from 'react-router-dom';
import { Box, Typography, Paper } from '@mui/material';


const CameraDetail: React.FC = () => {
  const { deviceId } = useParams<{ deviceId: string }>();

  if (!deviceId) {
    return <Navigate to="/" replace />;
  }

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" component="h1" gutterBottom>
        Camera Details
      </Typography>

      <Paper sx={{ p: 3 }}>
        <Typography variant="h6" gutterBottom>
          Device: {decodeURIComponent(deviceId)}
        </Typography>
        <Typography variant="body2" color="text.secondary">
          Camera detail view coming soon...
        </Typography>
      </Paper>
    </Box>
  );
};

export default CameraDetail; 