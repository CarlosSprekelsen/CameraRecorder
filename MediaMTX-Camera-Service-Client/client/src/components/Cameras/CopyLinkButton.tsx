/**
 * @fileoverview CopyLinkButton component for stream URL management
 * @author MediaMTX Development Team
 * @version 1.0.0
 */

import React, { useState } from 'react';
import { Button } from '../atoms/Button/Button';
import { Icon } from '../atoms/Icon/Icon';
import { Alert } from '../atoms/Alert/Alert';
import { Menu, MenuItem, ListItemIcon, ListItemText } from '../atoms/Menu/Menu';
import { Snackbar } from '../atoms/Snackbar/Snackbar';
// ARCHITECTURE FIX: Removed direct service import - use store hooks instead

interface CopyLinkButtonProps {
  device: string;
  streams: {
    rtsp: string;
    hls: string;
  };
}

/**
 * CopyLinkButton - Stream URL copying following architecture section 5.1
 * Exposes HLS/WebRTC links for external playback (no embedded playback)
 */
const CopyLinkButton: React.FC<CopyLinkButtonProps> = ({ device, streams }) => {
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [snackbarOpen, setSnackbarOpen] = useState(false);
  const [snackbarMessage, setSnackbarMessage] = useState('');
  const [snackbarSeverity, setSnackbarSeverity] = useState<'success' | 'error'>('success');

  const open = Boolean(anchorEl);

  const handleClick = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const copyToClipboard = async (text: string, label: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setSnackbarMessage(`${label} copied to clipboard`);
      setSnackbarSeverity('success');
      setSnackbarOpen(true);
      console.log(`Copied ${label} for device ${device}`);
    } catch (error) {
      setSnackbarMessage(`Failed to copy ${label}`);
      setSnackbarSeverity('error');
      setSnackbarOpen(true);
      console.error(`Failed to copy ${label} for device ${device}`, error);
    }
    handleClose();
  };

  const openInNewTab = (url: string, label: string) => {
    try {
      window.open(url, '_blank', 'noopener,noreferrer');
      console.log(`Opened ${label} for device ${device} in new tab`);
    } catch (error) {
      setSnackbarMessage(`Failed to open ${label}`);
      setSnackbarSeverity('error');
      setSnackbarOpen(true);
      console.error(`Failed to open ${label} for device ${device}`, error);
    }
    handleClose();
  };

  const handleSnackbarClose = () => {
    setSnackbarOpen(false);
  };

  return (
    <>
      <Button onClick={handleClick} size="small" variant="primary" title="Copy stream links">
        <Icon name="settings" size={16} />
      </Button>

      <Menu
        anchorEl={anchorEl}
        open={open}
        onClose={handleClose}
      >
        <MenuItem onClick={() => copyToClipboard(streams.rtsp, 'RTSP URL')}>
          <ListItemIcon>
            <Icon name="settings" size={16} />
          </ListItemIcon>
          <ListItemText>Copy RTSP URL - {streams.rtsp}</ListItemText>
        </MenuItem>

        <MenuItem onClick={() => copyToClipboard(streams.hls, 'HLS URL')}>
          <ListItemIcon>
            <Icon name="settings" size={16} />
          </ListItemIcon>
          <ListItemText>Copy HLS URL - {streams.hls}</ListItemText>
        </MenuItem>

        <MenuItem onClick={() => openInNewTab(streams.hls, 'HLS stream')}>
          <ListItemIcon>
            <Icon name="settings" size={16} />
          </ListItemIcon>
          <ListItemText>Open HLS in new tab - View stream in browser</ListItemText>
        </MenuItem>
      </Menu>

      <Snackbar
        open={snackbarOpen}
        autoHideDuration={3000}
        onClose={handleSnackbarClose}
      >
        <Alert variant={snackbarSeverity} className="w-full">
          {snackbarMessage}
        </Alert>
      </Snackbar>
    </>
  );
};

export default CopyLinkButton;
