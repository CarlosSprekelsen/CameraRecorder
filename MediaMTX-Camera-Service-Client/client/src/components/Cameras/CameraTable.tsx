/**
 * @fileoverview CameraTable component for displaying camera devices in a table format
 * @author MediaMTX Development Team
 * @version 1.0.0
 */

import React from 'react';
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Chip,
  Box,
  Typography,
  Button,
} from '@mui/material';
import {
  Refresh as RefreshIcon,
  Videocam as CameraIcon,
  CheckCircle as ConnectedIcon,
  Error as ErrorIcon,
  Cancel as DisconnectedIcon,
} from '@mui/icons-material';
import { Camera, StreamsListResult } from '../../types/api';
import DeviceActions from './DeviceActions';
import CopyLinkButton from './CopyLinkButton';
import { useRecordingStore } from '../../stores/recording/recordingStore';

interface CameraTableProps {
  cameras: Camera[];
  streams: StreamsListResult[];
  onRefresh: () => void;
}

/**
 * CameraTable - Device list with status and stream management
 *
 * Displays cameras in a table format with real-time status updates and stream links.
 * Implements the I.Discovery interface for device enumeration and stream URL management.
 *
 * @component
 * @param {CameraTableProps} props - Component props
 * @param {Camera[]} props.cameras - Array of camera devices to display
 * @param {StreamsListResult[]} props.streams - Array of active stream information
 * @param {() => void} props.onRefresh - Callback function to refresh camera list
 * @returns {JSX.Element} The camera table component
 *
 * @features
 * - Real-time camera status display
 * - Stream URL management and copying
 * - Recording status indicators
 * - Device action buttons
 * - Responsive table design
 *
 * @example
 * ```tsx
 * <CameraTable
 *   cameras={cameras}
 *   streams={streams}
 *   onRefresh={handleRefresh}
 * />
 * ```
 *
 * @see {@link ../../docs/architecture/client-architechture.md} Client Architecture
 */
const CameraTable: React.FC<CameraTableProps> = ({ cameras, streams, onRefresh }) => {
  const { activeRecordings } = useRecordingStore();
  const getStatusIcon = (status: Camera['status']) => {
    switch (status) {
      case 'CONNECTED':
        return <ConnectedIcon color="success" />;
      case 'ERROR':
        return <ErrorIcon color="error" />;
      case 'DISCONNECTED':
        return <DisconnectedIcon color="disabled" />;
      default:
        return <CameraIcon color="disabled" />;
    }
  };

  const getStatusColor = (status: Camera['status']) => {
    switch (status) {
      case 'CONNECTED':
        return 'success';
      case 'ERROR':
        return 'error';
      case 'DISCONNECTED':
        return 'default';
      default:
        return 'default';
    }
  };

  const getStreamStatus = (device: string) => {
    const stream = streams.find((s) => s.name === device);
    return stream
      ? {
          active: stream.ready,
          readers: stream.readers,
          bytesSent: stream.bytes_sent,
        }
      : {
          active: false,
          readers: 0,
          bytesSent: 0,
        };
  };

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  if (cameras.length === 0) {
    return (
      <Box textAlign="center" py={4}>
        <CameraIcon sx={{ fontSize: 64, color: 'text.secondary', mb: 2 }} />
        <Typography variant="h6" color="text.secondary" gutterBottom>
          No cameras found
        </Typography>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
          Make sure cameras are connected and the service is running.
        </Typography>
        <Button variant="outlined" startIcon={<RefreshIcon />} onClick={onRefresh}>
          Refresh
        </Button>
      </Box>
    );
  }

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={2}>
        <Typography variant="h6">Connected Devices ({cameras.length})</Typography>
        <Button variant="outlined" startIcon={<RefreshIcon />} onClick={onRefresh} size="small">
          Refresh
        </Button>
      </Box>

      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Device</TableCell>
              <TableCell>Status</TableCell>
              <TableCell>Resolution</TableCell>
              <TableCell>FPS</TableCell>
              <TableCell>Stream Status</TableCell>
              <TableCell>Recording</TableCell>
              <TableCell>Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {cameras.map((camera) => {
              const streamStatus = getStreamStatus(camera.device);
              const recording = activeRecordings[camera.device];

              return (
                <TableRow key={camera.device}>
                  <TableCell>
                    <Box display="flex" alignItems="center">
                      {getStatusIcon(camera.status)}
                      <Box ml={1}>
                        <Typography variant="body2" fontWeight="medium">
                          {camera.name}
                        </Typography>
                        <Typography variant="caption" color="text.secondary">
                          {camera.device}
                        </Typography>
                      </Box>
                    </Box>
                  </TableCell>

                  <TableCell>
                    <Chip
                      label={camera.status}
                      color={
                        getStatusColor(camera.status) as 'success' | 'error' | 'warning' | 'info'
                      }
                      size="small"
                    />
                  </TableCell>

                  <TableCell>
                    <Typography variant="body2">{camera.resolution}</Typography>
                  </TableCell>

                  <TableCell>
                    <Typography variant="body2">{camera.fps} fps</Typography>
                  </TableCell>

                  <TableCell>
                    <Box>
                      <Chip
                        label={streamStatus.active ? 'Active' : 'Inactive'}
                        color={streamStatus.active ? 'success' : 'default'}
                        size="small"
                        sx={{ mb: 0.5 }}
                      />
                      {streamStatus.active && (
                        <Typography variant="caption" display="block" color="text.secondary">
                          {streamStatus.readers} readers • {formatBytes(streamStatus.bytesSent)}{' '}
                          sent
                        </Typography>
                      )}
                    </Box>
                  </TableCell>
                  <TableCell>
                    {recording ? (
                      <Chip label="Recording" color="error" size="small" />
                    ) : (
                      <Chip label="Idle" size="small" />
                    )}
                  </TableCell>

                  <TableCell>
                    <Box display="flex" gap={1}>
                      {camera.streams && (
                        <CopyLinkButton device={camera.device} streams={camera.streams} />
                      )}
                      <DeviceActions device={camera.device} />
                    </Box>
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  );
};

export default CameraTable;
