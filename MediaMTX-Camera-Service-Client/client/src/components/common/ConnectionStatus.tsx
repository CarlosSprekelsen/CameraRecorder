/**
 * @deprecated This component is deprecated and will be removed in v2.0
 * Use /components/ConnectionStatus/ConnectionStatus.tsx instead
 * 
 * Migration Guide:
 * - Replace import: import ConnectionStatus from '../common/ConnectionStatus'
 * - With: import ConnectionStatus from '../ConnectionStatus/ConnectionStatus'
 * - Props interface remains the same
 * 
 * This component will be removed in the next major version.
 * Please migrate to the new implementation to avoid breaking changes.
 */

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
import { useConnectionStore, useHealthStore, useMetricsStore } from '../../stores/connection';
import { connectionService } from '../../services/connectionService';

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
  
  // Use new modular stores
  const {
    status: storeStatus,
    isConnecting: storeIsConnecting,
    isReconnecting: storeIsReconnecting,
    error: storeError,
    errorCode: storeErrorCode,
    errorTimestamp: storeErrorTimestamp,
    reconnectAttempts: storeReconnectAttempts,
    maxReconnectAttempts: storeMaxReconnectAttempts,
    nextReconnectTime: storeNextReconnectTime,
    lastConnected: storeLastConnected,
    lastDisconnected: storeLastDisconnected,
    url: storeUrl,
    autoReconnect: storeAutoReconnect
  } = useConnectionStore();

  const {
    isHealthy: storeIsHealthy,
    healthScore: storeHealthScore,
    connectionQuality: storeConnectionQuality
  } = useHealthStore();

  const {
    averageResponseTime: storeLatency,
    messageCount: storeMessageCount,
    errorCount: storeErrorCount,
    connectionUptime: storeConnectionUptime
  } = useMetricsStore();

  const handleRefresh = () => {
    if (onRefresh) {
      onRefresh();
    } else {
      // Default refresh behavior - force reconnect
      connectionService.forceReconnect();
    }
  };

  const handleToggleAutoReconnect = () => {
    connectionService.setAutoReconnect(!storeAutoReconnect);
  };

  const getStatusColor = () => {
    switch (storeStatus) {
      case 'connected':
        return storeIsHealthy ? 'success' : 'warning';
      case 'connecting':
        return 'warning';
      case 'error':
        return 'error';
      default:
        return 'error';
    }
  };

  const getStatusIcon = () => {
    switch (storeStatus) {
      case 'connected':
        if (storeIsHealthy) {
          switch (storeConnectionQuality) {
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
    if (storeIsConnecting) return 'Connecting...';
    if (storeIsReconnecting) return `Reconnecting (${storeReconnectAttempts}/${storeMaxReconnectAttempts})`;
    if (storeStatus === 'connected') {
      if (storeIsHealthy) {
        return `${storeConnectionQuality.charAt(0).toUpperCase() + storeConnectionQuality.slice(1)} (${storeHealthScore}%)`;
      }
      return 'Connected (Unhealthy)';
    }
    if (storeStatus === 'error') return 'Connection Error';
    return 'Disconnected';
  };

  const getTooltipText = () => {
    let tooltip = `Status: ${storeStatus}`;
    
    if (storeUrl) {
      tooltip += `\nServer: ${storeUrl}`;
    }
    
    if (storeLastConnected) {
      tooltip += `\nLast Connected: ${storeLastConnected.toLocaleTimeString()}`;
    }
    
    if (storeLastDisconnected) {
      tooltip += `\nLast Disconnected: ${storeLastDisconnected.toLocaleTimeString()}`;
    }
    
    if (storeConnectionUptime) {
      const uptimeMinutes = Math.floor(storeConnectionUptime / 60000);
      tooltip += `\nUptime: ${uptimeMinutes} minutes`;
    }
    
    if (storeHealthScore !== null) {
      tooltip += `\nHealth Score: ${storeHealthScore}%`;
    }
    
    if (storeLatency !== null) {
      tooltip += `\nLatency: ${storeLatency.toFixed(1)}ms`;
    }
    
    if (storeMessageCount > 0) {
      tooltip += `\nMessages: ${storeMessageCount}`;
    }
    
    if (storeErrorCount > 0) {
      tooltip += `\nErrors: ${storeErrorCount}`;
    }
    
    if (storeReconnectAttempts > 0) {
      tooltip += `\nReconnection Attempts: ${storeReconnectAttempts}/${storeMaxReconnectAttempts}`;
    }
    
    if (storeNextReconnectTime) {
      tooltip += `\nNext Reconnect: ${storeNextReconnectTime.toLocaleTimeString()}`;
    }
    
    if (storeError) {
      tooltip += `\nError: ${storeError}`;
      if (storeErrorCode) {
        tooltip += ` (Code: ${storeErrorCode})`;
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
    switch (storeConnectionQuality) {
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
      {storeError && (
        <Collapse in={true}>
          <Alert 
            severity="error" 
            sx={{ mt: 1 }}
            action={
              <Button color="inherit" size="small" onClick={() => connectionService.clearError()}>
                Dismiss
              </Button>
            }
          >
            <Typography variant="body2">
              {storeError}
              {storeErrorCode && ` (Code: ${storeErrorCode})`}
            </Typography>
            {storeErrorTimestamp && (
              <Typography variant="caption" display="block">
                {storeErrorTimestamp.toLocaleTimeString()}
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
                    value={storeHealthScore} 
                    sx={{ flexGrow: 1, height: 8, borderRadius: 4 }}
                    color={storeHealthScore >= 90 ? 'success' : storeHealthScore >= 70 ? 'info' : 'warning'}
                  />
                  <Typography variant="body2">{storeHealthScore}%</Typography>
                </Box>
              </Box>
              
              <Box>
                <Typography variant="caption" color="text.secondary">
                  Quality
                </Typography>
                <Chip 
                  label={storeConnectionQuality} 
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
                  {formatLatency(storeLatency)}
                </Typography>
              </Box>
              
              <Box>
                <Typography variant="caption" color="text.secondary">
                  Uptime
                </Typography>
                <Typography variant="body2">
                  {formatUptime(storeConnectionUptime)}
                </Typography>
              </Box>
            </Box>

            <Box sx={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 2, mb: 2 }}>
              <Box>
                <Typography variant="caption" color="text.secondary">
                  Messages
                </Typography>
                <Typography variant="body2">
                  {storeMessageCount} sent
                </Typography>
              </Box>
              
              <Box>
                <Typography variant="caption" color="text.secondary">
                  Errors
                </Typography>
                <Typography variant="body2" color={storeErrorCount > 0 ? 'error' : 'text.primary'}>
                  {storeErrorCount} errors
                </Typography>
              </Box>
            </Box>

            {storeReconnectAttempts > 0 && (
              <Box sx={{ mb: 2 }}>
                <Typography variant="caption" color="text.secondary">
                  Reconnection Progress
                </Typography>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                  <LinearProgress 
                    variant="determinate" 
                    value={(storeReconnectAttempts / storeMaxReconnectAttempts) * 100}
                    sx={{ flexGrow: 1, height: 6, borderRadius: 3 }}
                    color="warning"
                  />
                  <Typography variant="body2">
                    {storeReconnectAttempts}/{storeMaxReconnectAttempts}
                  </Typography>
                </Box>
                {storeNextReconnectTime && (
                  <Typography variant="caption" color="text.secondary">
                    Next attempt: {storeNextReconnectTime.toLocaleTimeString()}
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
                color={storeAutoReconnect ? 'success' : 'primary'}
              >
                Auto-reconnect {storeAutoReconnect ? 'ON' : 'OFF'}
              </Button>
              
              <Button
                size="small"
                variant="outlined"
                onClick={() => useMetricsStore.getState().resetMetrics()}
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
          {storeUrl && (
            <Typography variant="caption" color="text.secondary">
              Server: {storeUrl}
            </Typography>
          )}
          {storeLastConnected && (
            <Typography variant="caption" color="text.secondary">
              Connected: {storeLastConnected.toLocaleTimeString()}
            </Typography>
          )}
          {storeHealthScore !== null && (
            <Typography variant="caption" color="text.secondary">
              Health: {storeHealthScore}%
            </Typography>
          )}
          {storeLatency !== null && (
            <Typography variant="caption" color="text.secondary">
              Latency: {formatLatency(storeLatency)}
            </Typography>
          )}
          {storeReconnectAttempts > 0 && (
            <Typography variant="caption" color="text.secondary">
              Reconnections: {storeReconnectAttempts}/{storeMaxReconnectAttempts}
            </Typography>
          )}
          {storeError && (
            <Typography variant="caption" color="error">
              Error: {storeError}
            </Typography>
          )}
        </Box>
      )}
    </Box>
  );
};

export default ConnectionStatus; 