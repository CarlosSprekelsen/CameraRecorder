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
} from '../atoms/Table/Table';
import { Paper } from '../atoms/Paper/Paper';
import { IconButton } from '../atoms/IconButton/IconButton';
import { Box } from '../atoms/Box/Box';
import { Typography } from '../atoms/Typography/Typography';
import { CircularProgress } from '../atoms/CircularProgress/CircularProgress';
import { Checkbox } from '../atoms/Checkbox/Checkbox';
import { Tooltip } from '../atoms/Tooltip/Tooltip';
import { Icon } from '../atoms/Icon/Icon';
import { useFileStore } from '../../stores/file/fileStore';
import { FileListItem } from '../../stores/file/fileStore';
import ConfirmDialog from './ConfirmDialog';
// ARCHITECTURE FIX: Removed direct service import - use store hooks instead
import PermissionGate from '../Security/PermissionGate';

interface FileTableProps {
  files: FileListItem[];
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

  const handleDownload = async (file: FileListItem) => {
    try {
      await downloadFile(file.download_url, file.filename);
      console.log(`Download initiated for: ${file.filename}`);
    } catch (error) {
      console.error(`Download failed for: ${file.filename}`, error);
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
        console.log(`File deleted: ${filename}`);
        // Remove from selection if it was selected
        if (selectedFiles.includes(filename)) {
          toggleFileSelection(filename);
        }
      }
    } catch (error) {
      console.error(`Delete failed for: ${filename}`, error);
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
      <Box className="flex justify-center items-center min-h-[200px]">
        <CircularProgress />
      </Box>
    );
  }

  if (files.length === 0) {
    return (
      <Paper className="p-6 text-center">
        <Typography variant="h6" color="secondary">
          No {fileType} found.
        </Typography>
        <Typography variant="body2" color="secondary">
          {fileType === 'recordings'
            ? 'Start recording to see files here.'
            : 'Take snapshots to see files here.'}
        </Typography>
      </Paper>
    );
  }

  return (
    <>
      <TableContainer>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>
                <Checkbox
                  indeterminate={isIndeterminate}
                  checked={isAllSelected}
                  onChange={handleSelectAll}
                />
              </TableCell>
              <TableCell>Filename</TableCell>
              <TableCell>Size</TableCell>
              <TableCell>Modified</TableCell>
              <TableCell align="center">Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {files.map((file) => (
              <TableRow key={file.filename} hover>
                <TableCell>
                  <Checkbox
                    checked={selectedFiles.includes(file.filename)}
                    onChange={() => toggleFileSelection(file.filename)}
                  />
                </TableCell>
                <TableCell>
                  <Box>
                    <Typography variant="body2" className="truncate max-w-[200px]">
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
                <TableCell align="center">
                  <Box className="flex gap-1 justify-center">
                    <PermissionGate requirePermission="manageFiles">
                      <Tooltip title="Download">
                        <IconButton
                          size="small"
                          onClick={() => handleDownload(file)}
                          color="primary"
                        >
                          <Icon name="download" size={16} />
                        </IconButton>
                      </Tooltip>
                    </PermissionGate>

                    <PermissionGate requirePermission="deleteFiles">
                      <Tooltip title="Delete">
                        <IconButton
                          size="small"
                          onClick={() => handleDeleteClick(file.filename)}
                          color="default"
                        >
                          <Icon name="delete" size={16} />
                        </IconButton>
                      </Tooltip>
                    </PermissionGate>

                    <Tooltip title="Info">
                      <IconButton size="small" color="default">
                        <Icon name="info" size={16} />
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
