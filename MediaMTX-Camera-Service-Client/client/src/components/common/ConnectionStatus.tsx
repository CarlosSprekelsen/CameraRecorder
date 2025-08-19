import React from 'react';
import { 
  Chip, 
  Box, 
  Tooltip, 
  IconButton, 
  Typography, 
  LinearProgress,
  Alert,
  Collapse,
  Button
} from '@mui/material';
import {
  WifiOff as WifiOffIcon,
  Refresh as RefreshIcon,
  HourglassEmpty as ConnectingIcon,
  Error as ErrorIcon,
  CheckCircle as HealthyIcon,
  Warning as WarningIcon,
  Speed as SpeedIcon,
  SignalCellular4Bar as ExcellentIcon,
  SignalCellular3Bar as GoodIcon,
  SignalCellular2Bar as PoorIcon,
  ExpandMore as ExpandMoreIcon,
  ExpandLess as ExpandLessIcon,
  Settings as SettingsIcon,
} from '@mui/icons-material';
import { useConnectionStore } from '../../stores/connectionStore';

interface ConnectionStatusProps {
  onRefresh?: () => void;
  showDetails?: boolean;
  compact?: boolean;
}

const ConnectionStatus: React.FC<ConnectionStatusProps> = ({ 
  onRefresh,
  showDetails = false,
  compact = false
}) => {
  const [showAdvancedDetails, setShowAdvancedDetails] = React.useState(false);
  
  const {
    status,
    isConnecting,
    isReconnecting,
    isHealthy,
    error,
    errorCode,
    errorTimestamp,
    reconnectAttempts,
    maxReconnectAttempts,
    nextReconnectTime,
    lastConnected,
    lastDisconnected,
    url,
    healthScore,
    connectionQuality,
    latency,
    messageCount,
    errorCount,
    connectionUptime,
    autoReconnect
  } = useConnectionStore();

  const handleRefresh = () => {
    if (onRefresh) {
      onRefresh();
    } else {
      // Default refresh behavior - force reconnect
      useConnectionStore.getState().forceReconnect();
    }
  };

  const handleToggleAutoReconnect = () => {
    useConnectionStore.getState().setAutoReconnect(!autoReconnect);
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
        if (isHealthy) {
          switch (connectionQuality) {
            case 'excellent': return <ExcellentIcon />;
            case 'good': return <GoodIcon />;
            case 'poor': return <PoorIcon />;
            default: return <HealthyIcon />;
          }
        }
        return <WarningIcon />;
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
    if (status === 'connected') {
      if (isHealthy) {
        return `${connectionQuality.charAt(0).toUpperCase() + connectionQuality.slice(1)} (${healthScore}%)`;
      }
      return 'Connected (Unhealthy)';
    }
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
    
    if (connectionUptime) {
      const uptimeMinutes = Math.floor(connectionUptime / 60000);
      tooltip += `\nUptime: ${uptimeMinutes} minutes`;
    }
    
    if (healthScore !== null) {
      tooltip += `\nHealth Score: ${healthScore}%`;
    }
    
    if (latency !== null) {
      tooltip += `\nLatency: ${latency.toFixed(1)}ms`;
    }
    
    if (messageCount > 0) {
      tooltip += `\nMessages: ${messageCount}`;
    }
    
    if (errorCount > 0) {
      tooltip += `\nErrors: ${errorCount}`;
    }
    
    if (reconnectAttempts > 0) {
      tooltip += `\nReconnection Attempts: ${reconnectAttempts}/${maxReconnectAttempts}`;
    }
    
    if (nextReconnectTime) {
      tooltip += `\nNext Reconnect: ${nextReconnectTime.toLocaleTimeString()}`;
    }
    
    if (error) {
      tooltip += `\nError: ${error}`;
      if (errorCode) {
        tooltip += ` (Code: ${errorCode})`;
      }
    }
    
    return tooltip;
  };

  const formatUptime = (uptime: number | null) => {
    if (!uptime) return 'N/A';
    const minutes = Math.floor(uptime / 60000);
    const hours = Math.floor(minutes / 60);
    const remainingMinutes = minutes % 60;
    if (hours > 0) {
      return `${hours}h ${remainingMinutes}m`;
    }
    return `${minutes}m`;
  };

  const formatLatency = (latency: number | null) => {
    if (!latency) return 'N/A';
    return `${latency.toFixed(1)}ms`;
  };

  const getQualityColor = () => {
    switch (connectionQuality) {
      case 'excellent': return 'success';
      case 'good': return 'info';
      case 'poor': return 'warning';
      case 'unstable': return 'error';
      default: return 'default';
    }
  };

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
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
          <Tooltip title="Force reconnect" arrow>
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

        {!compact && (
          <Tooltip title="Connection settings" arrow>
            <IconButton
              size="small"
              onClick={() => setShowAdvancedDetails(!showAdvancedDetails)}
              color="primary"
              sx={{ ml: 0.5 }}
            >
              {showAdvancedDetails ? <ExpandLessIcon /> : <ExpandMoreIcon />}
            </IconButton>
          </Tooltip>
        )}
      </Box>

      {/* Error Alert */}
      {error && (
        <Collapse in={true}>
          <Alert 
            severity="error" 
            sx={{ mt: 1 }}
            action={
              <Button color="inherit" size="small" onClick={() => useConnectionStore.getState().clearError()}>
                Dismiss
              </Button>
            }
          >
            <Typography variant="body2">
              {error}
              {errorCode && ` (Code: ${errorCode})`}
            </Typography>
            {errorTimestamp && (
              <Typography variant="caption" display="block">
                {errorTimestamp.toLocaleTimeString()}
              </Typography>
            )}
          </Alert>
        </Collapse>
      )}

      {/* Advanced Details */}
      {showAdvancedDetails && !compact && (
        <Collapse in={showAdvancedDetails}>
          <Box sx={{ 
            mt: 1, 
            p: 2, 
            border: 1, 
            borderColor: 'divider', 
            borderRadius: 1,
            backgroundColor: 'background.paper'
          }}>
            <Typography variant="subtitle2" gutterBottom>
              Connection Details
            </Typography>
            
            <Box sx={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 2, mb: 2 }}>
              <Box>
                <Typography variant="caption" color="text.secondary">
                  Health Score
                </Typography>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                  <LinearProgress 
                    variant="determinate" 
                    value={healthScore} 
                    sx={{ flexGrow: 1, height: 8, borderRadius: 4 }}
                    color={healthScore >= 90 ? 'success' : healthScore >= 70 ? 'info' : 'warning'}
                  />
                  <Typography variant="body2">{healthScore}%</Typography>
                </Box>
              </Box>
              
              <Box>
                <Typography variant="caption" color="text.secondary">
                  Quality
                </Typography>
                <Chip 
                  label={connectionQuality} 
                  size="small" 
                  color={getQualityColor()}
                  variant="outlined"
                />
              </Box>
            </Box>

            <Box sx={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 2, mb: 2 }}>
              <Box>
                <Typography variant="caption" color="text.secondary">
                  Latency
                </Typography>
                <Typography variant="body2" sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                  <SpeedIcon fontSize="small" />
                  {formatLatency(latency)}
                </Typography>
              </Box>
              
              <Box>
                <Typography variant="caption" color="text.secondary">
                  Uptime
                </Typography>
                <Typography variant="body2">
                  {formatUptime(connectionUptime)}
                </Typography>
              </Box>
            </Box>

            <Box sx={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 2, mb: 2 }}>
              <Box>
                <Typography variant="caption" color="text.secondary">
                  Messages
                </Typography>
                <Typography variant="body2">
                  {messageCount} sent
                </Typography>
              </Box>
              
              <Box>
                <Typography variant="caption" color="text.secondary">
                  Errors
                </Typography>
                <Typography variant="body2" color={errorCount > 0 ? 'error' : 'text.primary'}>
                  {errorCount} errors
                </Typography>
              </Box>
            </Box>

            {reconnectAttempts > 0 && (
              <Box sx={{ mb: 2 }}>
                <Typography variant="caption" color="text.secondary">
                  Reconnection Progress
                </Typography>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                  <LinearProgress 
                    variant="determinate" 
                    value={(reconnectAttempts / maxReconnectAttempts) * 100}
                    sx={{ flexGrow: 1, height: 6, borderRadius: 3 }}
                    color="warning"
                  />
                  <Typography variant="body2">
                    {reconnectAttempts}/{maxReconnectAttempts}
                  </Typography>
                </Box>
                {nextReconnectTime && (
                  <Typography variant="caption" color="text.secondary">
                    Next attempt: {nextReconnectTime.toLocaleTimeString()}
                  </Typography>
                )}
              </Box>
            )}

            <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
              <Button
                size="small"
                variant="outlined"
                startIcon={<SettingsIcon />}
                onClick={handleToggleAutoReconnect}
                color={autoReconnect ? 'success' : 'primary'}
              >
                Auto-reconnect {autoReconnect ? 'ON' : 'OFF'}
              </Button>
              
              <Button
                size="small"
                variant="outlined"
                onClick={() => useConnectionStore.getState().resetMetrics()}
              >
                Reset Metrics
              </Button>
            </Box>
          </Box>
        </Collapse>
      )}

      {/* Compact Details */}
      {showDetails && compact && (
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
          {healthScore !== null && (
            <Typography variant="caption" color="text.secondary">
              Health: {healthScore}%
            </Typography>
          )}
          {latency !== null && (
            <Typography variant="caption" color="text.secondary">
              Latency: {formatLatency(latency)}
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