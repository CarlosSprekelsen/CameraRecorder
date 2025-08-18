import React from 'react';
import { Chip, Box, Tooltip, IconButton, Typography } from '@mui/material';
import {
  WifiOff as WifiOffIcon,
  Wifi as WifiIcon,
  Refresh as RefreshIcon,
  HourglassEmpty as ConnectingIcon,
  Error as ErrorIcon,
  CheckCircle as HealthyIcon,
  Warning as WarningIcon,
} from '@mui/icons-material';
import { useConnectionStore } from '../../stores/connectionStore';

interface ConnectionStatusProps {
  onRefresh?: () => void;
  showDetails?: boolean;
}

const ConnectionStatus: React.FC<ConnectionStatusProps> = ({ 
  onRefresh,
  showDetails = false
}) => {
  const {
    status,
    isConnecting,
    isReconnecting,
    isHealthy,
    error,
    reconnectAttempts,
    maxReconnectAttempts,
    lastConnected,
    lastDisconnected,
    url
  } = useConnectionStore();

  const handleRefresh = () => {
    if (onRefresh) {
      onRefresh();
    } else {
      // Default refresh behavior
      useConnectionStore.getState().reconnect();
    }
  };

  const getStatusColor = () => {
    switch (status) {
      case 'connected':
        return isHealthy ? 'success' : 'warning';
      case 'connecting':
        return 'warning';
      case 'error':
        return 'error';
      default:
        return 'error';
    }
  };

  const getStatusIcon = () => {
    switch (status) {
      case 'connected':
        return isHealthy ? <HealthyIcon /> : <WarningIcon />;
      case 'connecting':
        return <ConnectingIcon />;
      case 'error':
        return <ErrorIcon />;
      default:
        return <WifiOffIcon />;
    }
  };

  const getStatusLabel = () => {
    if (isConnecting) return 'Connecting...';
    if (isReconnecting) return `Reconnecting (${reconnectAttempts}/${maxReconnectAttempts})`;
    if (status === 'connected') return isHealthy ? 'Connected' : 'Connected (Unhealthy)';
    if (status === 'error') return 'Connection Error';
    return 'Disconnected';
  };

  const getTooltipText = () => {
    let tooltip = `Status: ${status}`;
    
    if (url) {
      tooltip += `\nServer: ${url}`;
    }
    
    if (lastConnected) {
      tooltip += `\nLast Connected: ${lastConnected.toLocaleTimeString()}`;
    }
    
    if (lastDisconnected) {
      tooltip += `\nLast Disconnected: ${lastDisconnected.toLocaleTimeString()}`;
    }
    
    if (reconnectAttempts > 0) {
      tooltip += `\nReconnection Attempts: ${reconnectAttempts}/${maxReconnectAttempts}`;
    }
    
    if (error) {
      tooltip += `\nError: ${error}`;
    }
    
    return tooltip;
  };

  return (
    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
      <Tooltip title={getTooltipText()} arrow placement="bottom">
        <Chip
          icon={getStatusIcon()}
          label={getStatusLabel()}
          color={getStatusColor()}
          size="small"
          variant="outlined"
          sx={{ 
            minWidth: 'fit-content',
            '& .MuiChip-icon': {
              fontSize: '1rem'
            }
          }}
        />
      </Tooltip>
      
      {(status === 'disconnected' || status === 'error') && (
        <Tooltip title="Retry connection" arrow>
          <IconButton
            size="small"
            onClick={handleRefresh}
            color="primary"
            sx={{ ml: 0.5 }}
          >
            <RefreshIcon />
          </IconButton>
        </Tooltip>
      )}
      
      {status === 'connected' && onRefresh && (
        <Tooltip title="Refresh cameras" arrow>
          <IconButton
            size="small"
            onClick={onRefresh}
            color="primary"
            sx={{ ml: 0.5 }}
          >
            <RefreshIcon />
          </IconButton>
        </Tooltip>
      )}

      {showDetails && (
        <Box sx={{ ml: 2, display: 'flex', flexDirection: 'column', gap: 0.5 }}>
          {url && (
            <Typography variant="caption" color="text.secondary">
              Server: {url}
            </Typography>
          )}
          {lastConnected && (
            <Typography variant="caption" color="text.secondary">
              Connected: {lastConnected.toLocaleTimeString()}
            </Typography>
          )}
          {reconnectAttempts > 0 && (
            <Typography variant="caption" color="text.secondary">
              Reconnections: {reconnectAttempts}/{maxReconnectAttempts}
            </Typography>
          )}
          {error && (
            <Typography variant="caption" color="error">
              Error: {error}
            </Typography>
          )}
        </Box>
      )}
    </Box>
  );
};

export default ConnectionStatus; 