import React, { useState } from 'react';
import { Button } from '../atoms/Button/Button';
import { Icon } from '../atoms/Icon/Icon';
import { Alert } from '../atoms/Alert/Alert';
import { Divider } from '../atoms/Divider/Divider';
import { Menu, MenuItem, ListItemIcon, ListItemText } from '../atoms/Menu/Menu';
import { Snackbar } from '../atoms/Snackbar/Snackbar';
import TimedRecordDialog from './TimedRecordDialog';
import { useRecordingStore } from '../../stores/recording/recordingStore';
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
      <Button onClick={handleClick} size="small" variant="secondary">
        <Icon name="settings" size={16} />
      </Button>

      <Menu
        anchorEl={anchorEl}
        open={open}
        onClose={handleClose}
      >
        <PermissionGate requirePermission="controlCameras">
          <MenuItem onClick={handleSnapshot}>
            <ListItemIcon>
              <Icon name="settings" size={16} />
            </ListItemIcon>
            <ListItemText>Take Snapshot</ListItemText>
          </MenuItem>
        </PermissionGate>

        <Divider />

        <PermissionGate requirePermission="controlCameras">
          <MenuItem onClick={handleStartRecording}>
            <ListItemIcon>
              <Icon name="settings" size={16} />
            </ListItemIcon>
            <ListItemText>Start Recording</ListItemText>
          </MenuItem>

          <MenuItem
            onClick={() => {
              setTimedOpen(true);
              handleClose();
            }}
          >
            <ListItemIcon>
              <Icon name="settings" size={16} />
            </ListItemIcon>
            <ListItemText>Timed Recording</ListItemText>
          </MenuItem>

          <MenuItem onClick={handleStopRecording}>
            <ListItemIcon>
              <Icon name="settings" size={16} />
            </ListItemIcon>
            <ListItemText>Stop Recording</ListItemText>
          </MenuItem>
        </PermissionGate>

        <Divider />

        <MenuItem onClick={handleSettings}>
          <ListItemIcon>
            <Icon name="settings" size={16} />
          </ListItemIcon>
          <ListItemText>Device Settings</ListItemText>
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
      >
        <Alert variant={snack.sev} className="w-full">
          {snack.msg}
        </Alert>
      </Snackbar>
    </>
  );
};

export default DeviceActions;
