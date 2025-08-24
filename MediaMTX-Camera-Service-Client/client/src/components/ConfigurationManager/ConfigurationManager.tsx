import React from 'react';
import { Box, Typography, Alert } from '@mui/material';
import { Info } from '@mui/icons-material';

const ConfigurationManager: React.FC = () => {
  return (
    <Box>
      <Typography variant="h5" gutterBottom>
        Configuration Manager
      </Typography>
      
      <Alert severity="info" icon={<Info />}>
        <Typography variant="body1">
          Configuration management is not available in the current server API.
        </Typography>
        <Typography variant="body2" sx={{ mt: 1 }}>
          The server does not provide configuration management endpoints. 
          All configuration is handled server-side through environment variables.
        </Typography>
      </Alert>
    </Box>
  );
};

export default ConfigurationManager;
