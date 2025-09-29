import React, { useState } from 'react';
import { IconButton, Menu, MenuItem, ListItemIcon, ListItemText, Divider } from '@mui/material';
import {
  MoreVert as MoreIcon,
  CameraAlt as SnapshotIcon,
  Videocam as RecordIcon,
  Stop as StopIcon,
  AccessTime as TimedIcon,
  Settings as SettingsIcon,
  PlayArrow as StreamIcon,
  Pause as PauseIcon,
} from '@mui/icons-material';
// ARCHITECTURE FIX: Removed direct service import - use store hooks instead
import { Snackbar, Alert } from '@mui/material';
import TimedRecordDialog from './TimedRecordDialog';
import { useRecordingStore } from '../../stores/recording/recordingStore';
import { useStreamingStore } from '../../stores/streaming/streamingStore';
// ARCHITECTURE FIX: Removed serviceFactory import - components must use stores only
import PermissionGate from '../Security/PermissionGate';

interface DeviceActionsProps {
  device: string;
}

/**
 * DeviceActions - Per-device action menu
 * Provides device control actions for camera operations
 * Note: This component provides similar functionality to RecordingController as specified in architecture
 */
const DeviceActions: React.FC<DeviceActionsProps> = ({ device }) => {
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [timedOpen, setTimedOpen] = useState(false);
  const [snack, setSnack] = useState<{
    open: boolean;
    msg: string;
    sev: 'success' | 'error' | 'info';
  }>({ open: false, msg: '', sev: 'success' });
  const open = Boolean(anchorEl);

  const { takeSnapshot, startRecording, stopRecording } = useRecordingStore();
  // ARCHITECTURE FIX: Removed setService - components don't inject services

  // ARCHITECTURE FIX: Removed direct service initialization - stores handle service injection

  const handleClick = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const handleSnapshot = async () => {
    console.log(`Snapshot requested for device: ${device}`);
    try {
      await takeSnapshot(device);
      setSnack({ open: true, msg: 'Snapshot requested', sev: 'success' });
    } catch (e) {
      setSnack({ open: true, msg: 'Snapshot failed', sev: 'error' });
    }
    handleClose();
  };

  const handleStartRecording = async () => {
    console.log(`Start recording requested for device: ${device}`);
    try {
      await startRecording(device);
      setSnack({ open: true, msg: 'Recording started', sev: 'success' });
    } catch (e) {
      setSnack({ open: true, msg: 'Start recording failed', sev: 'error' });
    }
    handleClose();
  };

  const handleStopRecording = async () => {
    console.log(`Stop recording requested for device: ${device}`);
    try {
      await stopRecording(device);
      setSnack({ open: true, msg: 'Recording stop requested', sev: 'info' });
    } catch (e) {
      setSnack({ open: true, msg: 'Stop recording failed', sev: 'error' });
    }
    handleClose();
  };

  const handleTimedStart = async (duration: number, format: string) => {
    await startRecording(device, duration, format);
    setTimedOpen(false);
  };

  const handleSettings = () => {
    console.log(`Settings requested for device: ${device}`);
    handleClose();
  };

  return (
    <>
      <IconButton onClick={handleClick} size="small" color="primary">
        <MoreIcon />
      </IconButton>

      <Menu
        anchorEl={anchorEl}
        open={open}
        onClose={handleClose}
        anchorOrigin={{
          vertical: 'bottom',
          horizontal: 'right',
        }}
        transformOrigin={{
          vertical: 'top',
          horizontal: 'right',
        }}
      >
        <PermissionGate requirePermission="controlCameras">
          <MenuItem onClick={handleSnapshot}>
            <ListItemIcon>
              <SnapshotIcon fontSize="small" />
            </ListItemIcon>
            <ListItemText primary="Take Snapshot" />
          </MenuItem>
        </PermissionGate>

        <Divider />

        <PermissionGate requirePermission="controlCameras">
          <MenuItem onClick={handleStartRecording}>
            <ListItemIcon>
              <RecordIcon fontSize="small" />
            </ListItemIcon>
            <ListItemText primary="Start Recording" />
          </MenuItem>

          <MenuItem
            onClick={() => {
              setTimedOpen(true);
              handleClose();
            }}
          >
            <ListItemIcon>
              <TimedIcon fontSize="small" />
            </ListItemIcon>
            <ListItemText primary="Timed Recording" />
          </MenuItem>

          <MenuItem onClick={handleStopRecording}>
            <ListItemIcon>
              <StopIcon fontSize="small" />
            </ListItemIcon>
            <ListItemText primary="Stop Recording" />
          </MenuItem>
        </PermissionGate>

        <Divider />

        <MenuItem onClick={handleSettings}>
          <ListItemIcon>
            <SettingsIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText primary="Device Settings" />
        </MenuItem>
      </Menu>

      <TimedRecordDialog
        open={timedOpen}
        onCancel={() => setTimedOpen(false)}
        onStart={handleTimedStart}
      />

      <Snackbar
        open={snack.open}
        autoHideDuration={2500}
        onClose={() => setSnack({ ...snack, open: false })}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert severity={snack.sev} sx={{ width: '100%' }}>
          {snack.msg}
        </Alert>
      </Snackbar>
    </>
  );
};

export default DeviceActions;
