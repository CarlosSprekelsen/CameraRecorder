/**
 * @fileoverview FileTable component for displaying files in table format
 * @author MediaMTX Development Team
 * @version 1.0.0
 */

import React, { useState } from 'react';
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  IconButton,
  Chip,
  Box,
  Typography,
  CircularProgress,
  Checkbox,
  Tooltip,
} from '@mui/material';
import {
  Download as DownloadIcon,
  Delete as DeleteIcon,
  Info as InfoIcon,
} from '@mui/icons-material';
import { useFileStore } from '../../stores/file/fileStore';
import { FileInfo } from '../../stores/file/fileStore';
import ConfirmDialog from './ConfirmDialog';
import { logger } from '../../services/logger/LoggerService';
import PermissionGate from '../Security/PermissionGate';

interface FileTableProps {
  files: FileInfo[];
  fileType: 'recordings' | 'snapshots';
  loading: boolean;
}

/**
 * FileTable - File listing with download and delete actions
 *
 * Displays files in a table format with download and delete capabilities.
 * Implements I.FileActions interface for file operations including download
 * via server-provided URLs and file deletion with confirmation.
 *
 * @component
 * @param {FileTableProps} props - Component props
 * @param {FileInfo[]} props.files - Array of file information to display
 * @param {'recordings' | 'snapshots'} props.fileType - Type of files being displayed
 * @param {boolean} props.loading - Loading state indicator
 * @returns {JSX.Element} The file table component
 *
 * @features
 * - File listing with metadata (size, date, format)
 * - Download functionality via server URLs
 * - Delete operations with confirmation
 * - Loading states and error handling
 * - Responsive table design
 *
 * @example
 * ```tsx
 * <FileTable
 *   files={recordings}
 *   fileType="recordings"
 *   loading={false}
 * />
 * ```
 *
 * @see {@link ../../docs/architecture/client-architechture.md} Client Architecture
 */
const FileTable: React.FC<FileTableProps> = ({ files, fileType, loading }) => {
  const [deleteDialog, setDeleteDialog] = useState<{
    open: boolean;
    filename: string;
    fileType: 'recordings' | 'snapshots';
  }>({ open: false, filename: '', fileType: 'recordings' });

  const {
    selectedFiles,
    downloadFile,
    deleteRecording,
    deleteSnapshot,
    toggleFileSelection,
    clearSelection,
  } = useFileStore();

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const formatDate = (dateString: string): string => {
    return new Date(dateString).toLocaleString();
  };

  const handleDownload = async (file: FileInfo) => {
    try {
      await downloadFile(file.download_url, file.filename);
      logger.info(`Download initiated for: ${file.filename}`);
    } catch (error) {
      logger.error(`Download failed for: ${file.filename}`, error as Record<string, unknown>);
    }
  };

  const handleDeleteClick = (filename: string) => {
    setDeleteDialog({
      open: true,
      filename,
      fileType,
    });
  };

  const handleDeleteConfirm = async () => {
    const { filename, fileType } = deleteDialog;
    try {
      let success = false;
      if (fileType === 'recordings') {
        success = await deleteRecording(filename);
      } else {
        success = await deleteSnapshot(filename);
      }

      if (success) {
        logger.info(`File deleted: ${filename}`);
        // Remove from selection if it was selected
        if (selectedFiles.includes(filename)) {
          toggleFileSelection(filename);
        }
      }
    } catch (error) {
      logger.error(`Delete failed for: ${filename}`, error as Record<string, unknown>);
    }
    setDeleteDialog({ open: false, filename: '', fileType: 'recordings' });
  };

  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      files.forEach((file) => {
        if (!selectedFiles.includes(file.filename)) {
          toggleFileSelection(file.filename);
        }
      });
    } else {
      clearSelection();
    }
  };

  const isAllSelected =
    files.length > 0 && files.every((file) => selectedFiles.includes(file.filename));
  const isIndeterminate = selectedFiles.length > 0 && !isAllSelected;

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="200px">
        <CircularProgress />
      </Box>
    );
  }

  if (files.length === 0) {
    return (
      <Paper elevation={1} sx={{ p: 3, textAlign: 'center' }}>
        <Typography variant="h6" color="textSecondary">
          No {fileType} found.
        </Typography>
        <Typography variant="body2" color="textSecondary">
          {fileType === 'recordings'
            ? 'Start recording to see files here.'
            : 'Take snapshots to see files here.'}
        </Typography>
      </Paper>
    );
  }

  return (
    <>
      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell padding="checkbox">
                <Checkbox
                  indeterminate={isIndeterminate}
                  checked={isAllSelected}
                  onChange={(e) => handleSelectAll(e.target.checked)}
                />
              </TableCell>
              <TableCell>Filename</TableCell>
              <TableCell>Size</TableCell>
              <TableCell>Modified</TableCell>
              <TableCell>Format</TableCell>
              <TableCell>Device</TableCell>
              <TableCell align="center">Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {files.map((file) => (
              <TableRow key={file.filename} hover>
                <TableCell padding="checkbox">
                  <Checkbox
                    checked={selectedFiles.includes(file.filename)}
                    onChange={() => toggleFileSelection(file.filename)}
                  />
                </TableCell>
                <TableCell>
                  <Box>
                    <Typography variant="body2" noWrap sx={{ maxWidth: 200 }}>
                      {file.filename}
                    </Typography>
                  </Box>
                </TableCell>
                <TableCell>
                  <Typography variant="body2">{formatFileSize(file.file_size)}</Typography>
                </TableCell>
                <TableCell>
                  <Typography variant="body2">{formatDate(file.modified_time)}</Typography>
                </TableCell>
                <TableCell>
                  <Chip
                    label={file.format || 'Unknown'}
                    size="small"
                    color="primary"
                    variant="outlined"
                  />
                </TableCell>
                <TableCell>
                  <Typography variant="body2">{file.device || 'Unknown'}</Typography>
                </TableCell>
                <TableCell align="center">
                  <Box display="flex" gap={1} justifyContent="center">
                    <PermissionGate requirePermission="manageFiles">
                      <Tooltip title="Download">
                        <IconButton
                          size="small"
                          onClick={() => handleDownload(file)}
                          color="primary"
                        >
                          <DownloadIcon />
                        </IconButton>
                      </Tooltip>
                    </PermissionGate>

                    <PermissionGate requirePermission="deleteFiles">
                      <Tooltip title="Delete">
                        <IconButton
                          size="small"
                          onClick={() => handleDeleteClick(file.filename)}
                          color="error"
                        >
                          <DeleteIcon />
                        </IconButton>
                      </Tooltip>
                    </PermissionGate>

                    <Tooltip title="Info">
                      <IconButton size="small" color="info">
                        <InfoIcon />
                      </IconButton>
                    </Tooltip>
                  </Box>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>

      <ConfirmDialog
        open={deleteDialog.open}
        title={`Delete ${fileType === 'recordings' ? 'Recording' : 'Snapshot'}`}
        message={`Are you sure you want to delete "${deleteDialog.filename}"? This action cannot be undone.`}
        onConfirm={handleDeleteConfirm}
        onCancel={() => setDeleteDialog({ open: false, filename: '', fileType: 'recordings' })}
        confirmText="Delete"
        cancelText="Cancel"
        severity="error"
      />
    </>
  );
};

export default FileTable;
