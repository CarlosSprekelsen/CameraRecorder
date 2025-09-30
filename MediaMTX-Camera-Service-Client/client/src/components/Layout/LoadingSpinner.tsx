import React from 'react';
import { Box } from '../atoms/Box/Box';
import { CircularProgress } from '../atoms/CircularProgress/CircularProgress';
import { Typography } from '../atoms/Typography/Typography';

interface LoadingSpinnerProps {
  message?: string;
}

const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({ message = 'Loading...' }) => {
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', gap: 8 }}>
      <CircularProgress />
      <Typography variant="body2" color="secondary">
        {message}
      </Typography>
    </Box>
  );
};

export default LoadingSpinner;
