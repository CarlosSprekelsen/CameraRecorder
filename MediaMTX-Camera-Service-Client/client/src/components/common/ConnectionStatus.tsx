import React from 'react';
import { Chip, Box, Tooltip, IconButton } from '@mui/material';
import {
  WifiOff as WifiOffIcon,
  Wifi as WifiIcon,
  Refresh as RefreshIcon,
  HourglassEmpty as ConnectingIcon,
} from '@mui/icons-material';

interface ConnectionStatusProps {
  isConnected: boolean;
  isConnecting: boolean;
  onRefresh?: () => void;
}

const ConnectionStatus: React.FC<ConnectionStatusProps> = ({ 
  isConnected, 
  isConnecting, 
  onRefresh 
}) => {
  if (isConnecting) {
    return (
      <Tooltip title="Connecting to server..." arrow>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <Chip
            icon={<ConnectingIcon />}
            label="Connecting..."
            color="warning"
            size="small"
            variant="outlined"
          />
        </Box>
      </Tooltip>
    );
  }

  if (isConnected) {
    return (
      <Tooltip title="Connected to MediaMTX Camera Service" arrow>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <Chip
            icon={<WifiIcon />}
            label="Connected"
            color="success"
            size="small"
            variant="outlined"
          />
          {onRefresh && (
            <Tooltip title="Refresh cameras" arrow>
              <IconButton
                size="small"
                onClick={onRefresh}
                color="primary"
              >
                <RefreshIcon />
              </IconButton>
            </Tooltip>
          )}
        </Box>
      </Tooltip>
    );
  }

  return (
    <Tooltip title="WebSocket connection not established" arrow>
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
        <Chip
          icon={<WifiOffIcon />}
          label="Disconnected"
          color="error"
          size="small"
          variant="outlined"
        />
        {onRefresh && (
          <Tooltip title="Retry connection" arrow>
            <IconButton
              size="small"
              onClick={onRefresh}
              color="primary"
            >
              <RefreshIcon />
            </IconButton>
          </Tooltip>
        )}
      </Box>
    </Tooltip>
  );
};

export default ConnectionStatus; 