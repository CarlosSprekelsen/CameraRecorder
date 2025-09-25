import React, { useState } from 'react';
import {
  IconButton,
  Menu,
  MenuItem,
  ListItemIcon,
  ListItemText,
  Divider,
} from '@mui/material';
import {
  MoreVert as MoreIcon,
  CameraAlt as SnapshotIcon,
  Videocam as RecordIcon,
  Stop as StopIcon,
  Settings as SettingsIcon,
} from '@mui/icons-material';
import { logger } from '../../services/logger/LoggerService';

interface DeviceActionsProps {
  device: string;
}

/**
 * DeviceActions - Per-device action menu following architecture section 5.1
 * Provides device control actions (will be enhanced in Sprint 3)
 */
const DeviceActions: React.FC<DeviceActionsProps> = ({ device }) => {
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const open = Boolean(anchorEl);

  const handleClick = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const handleSnapshot = () => {
    logger.info(`Snapshot requested for device: ${device}`);
    // TODO: Implement in Sprint 3
    handleClose();
  };

  const handleStartRecording = () => {
    logger.info(`Start recording requested for device: ${device}`);
    // TODO: Implement in Sprint 3
    handleClose();
  };

  const handleStopRecording = () => {
    logger.info(`Stop recording requested for device: ${device}`);
    // TODO: Implement in Sprint 3
    handleClose();
  };

  const handleSettings = () => {
    logger.info(`Settings requested for device: ${device}`);
    // TODO: Implement device settings
    handleClose();
  };

  return (
    <>
      <IconButton
        onClick={handleClick}
        size="small"
        color="primary"
      >
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
        <MenuItem onClick={handleSnapshot}>
          <ListItemIcon>
            <SnapshotIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText primary="Take Snapshot" />
        </MenuItem>

        <Divider />

        <MenuItem onClick={handleStartRecording}>
          <ListItemIcon>
            <RecordIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText primary="Start Recording" />
        </MenuItem>

        <MenuItem onClick={handleStopRecording}>
          <ListItemIcon>
            <StopIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText primary="Stop Recording" />
        </MenuItem>

        <Divider />

        <MenuItem onClick={handleSettings}>
          <ListItemIcon>
            <SettingsIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText primary="Device Settings" />
        </MenuItem>
      </Menu>
    </>
  );
};

export default DeviceActions;
