import React, { useState } from 'react';
import {
  IconButton,
  Menu,
  MenuItem,
  ListItemIcon,
  ListItemText,
  Snackbar,
  Alert,
  Tooltip,
} from '@mui/material';
import {
  Link as LinkIcon,
  ContentCopy as CopyIcon,
  OpenInNew as OpenIcon,
} from '@mui/icons-material';
import { logger } from '../../services/logger/LoggerService';

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
      logger.info(`Copied ${label} for device ${device}`);
    } catch (error) {
      setSnackbarMessage(`Failed to copy ${label}`);
      setSnackbarSeverity('error');
      setSnackbarOpen(true);
      logger.error(`Failed to copy ${label} for device ${device}`, error as Error);
    }
    handleClose();
  };

  const openInNewTab = (url: string, label: string) => {
    try {
      window.open(url, '_blank', 'noopener,noreferrer');
      logger.info(`Opened ${label} for device ${device} in new tab`);
    } catch (error) {
      setSnackbarMessage(`Failed to open ${label}`);
      setSnackbarSeverity('error');
      setSnackbarOpen(true);
      logger.error(`Failed to open ${label} for device ${device}`, error as Error);
    }
    handleClose();
  };

  const handleSnackbarClose = () => {
    setSnackbarOpen(false);
  };

  return (
    <>
      <Tooltip title="Copy stream links">
        <IconButton onClick={handleClick} size="small" color="primary">
          <LinkIcon />
        </IconButton>
      </Tooltip>

      <Menu
        anchorEl={anchorEl}
        open={open}
        onClose={handleClose}
        anchorOrigin={{
          vertical: 'bottom',
          horizontal: 'left',
        }}
        transformOrigin={{
          vertical: 'top',
          horizontal: 'left',
        }}
      >
        <MenuItem onClick={() => copyToClipboard(streams.rtsp, 'RTSP URL')}>
          <ListItemIcon>
            <CopyIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText primary="Copy RTSP URL" secondary={streams.rtsp} />
        </MenuItem>

        <MenuItem onClick={() => copyToClipboard(streams.hls, 'HLS URL')}>
          <ListItemIcon>
            <CopyIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText primary="Copy HLS URL" secondary={streams.hls} />
        </MenuItem>

        <MenuItem onClick={() => openInNewTab(streams.hls, 'HLS stream')}>
          <ListItemIcon>
            <OpenIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText primary="Open HLS in new tab" secondary="View stream in browser" />
        </MenuItem>
      </Menu>

      <Snackbar
        open={snackbarOpen}
        autoHideDuration={3000}
        onClose={handleSnackbarClose}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert onClose={handleSnackbarClose} severity={snackbarSeverity} sx={{ width: '100%' }}>
          {snackbarMessage}
        </Alert>
      </Snackbar>
    </>
  );
};

export default CopyLinkButton;
