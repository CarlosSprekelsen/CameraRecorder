import React, { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  Card,
  CardContent,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Tabs,
  Tab,
  Button,
  IconButton,
  CircularProgress,
  Alert,
  Pagination,
  Stack,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  DialogContentText,
  Tooltip,
  Chip,
  Grid,
} from '@mui/material';
import {
  Download,
  Refresh,
  VideoFile,
  Image,
  Delete,
  Info,
  Warning,
} from '@mui/icons-material';
import { useFileStore } from '../../stores/fileStore';
import type { FileItem, FileType } from '../../types';

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

function TabPanel(props: TabPanelProps) {
  const { children, value, index, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`file-tabpanel-${index}`}
      aria-labelledby={`file-tab-${index}`}
      {...other}
    >
      {value === index && <Box sx={{ pt: 2 }}>{children}</Box>}
    </div>
  );
}

const FileManager: React.FC = () => {
  const [tabValue, setTabValue] = useState(0);
  const [page, setPage] = useState(1);
  const [limit] = useState(20);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [fileInfoDialogOpen, setFileInfoDialogOpen] = useState(false);
  const [localSelectedFile, setLocalSelectedFile] = useState<FileItem | null>(null);

  const {
    recordings: storeRecordings,
    snapshots: storeSnapshots,
    selectedFile: storeSelectedFile,
    isLoading: storeIsLoading,
    isDeleting: storeIsDeleting,
    isLoadingFileInfo: storeIsLoadingFileInfo,
    error: storeError,
    loadRecordings: storeLoadRecordings,
    loadSnapshots: storeLoadSnapshots,
    downloadFile: storeDownloadFile,
    deleteRecording: storeDeleteRecording,
    deleteSnapshot: storeDeleteSnapshot,
    getRecordingInfo: storeGetRecordingInfo,
    getSnapshotInfo: storeGetSnapshotInfo,
    setSelectedFile: setStoreSelectedFile,
    isDownloading: storeIsDownloading,
    canDeleteFiles: storeCanDeleteFiles,
  } = useFileStore();

  const currentFiles = tabValue === 0 ? storeRecordings : storeSnapshots;
  const fileType: FileType = tabValue === 0 ? 'recordings' : 'snapshots';

  useEffect(() => {
    const offset = (page - 1) * limit;
    if (tabValue === 0) {
      storeLoadRecordings(limit, offset);
    } else {
      storeLoadSnapshots(limit, offset);
    }
  }, [tabValue, page, limit, storeLoadRecordings, storeLoadSnapshots]);

  const handleTabChange = (_event: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue);
    setPage(1); // Reset to first page when switching tabs
  };

  const handleDownload = async (filename: string) => {
    try {
      await storeDownloadFile(fileType, filename);
    } catch (err) {
      console.error('Download failed:', err);
    }
  };

  const handleDelete = (file: FileItem) => {
    setStoreSelectedFile(file);
    setDeleteDialogOpen(true);
  };

  const handleConfirmDelete = async () => {
    if (!storeSelectedFile) return;

    try {
      if (fileType === 'recordings') {
        await storeDeleteRecording(storeSelectedFile.filename);
      } else {
        await storeDeleteSnapshot(storeSelectedFile.filename);
      }
      setDeleteDialogOpen(false);
      setStoreSelectedFile(null);
    } catch (err) {
      console.error('Delete failed:', err);
    }
  };

  const handleViewInfo = async (file: FileItem) => {
    setStoreSelectedFile(file);
    try {
      if (fileType === 'recordings') {
        await storeGetRecordingInfo(file.filename);
      } else {
        await storeGetSnapshotInfo(file.filename);
      }
      setFileInfoDialogOpen(true);
    } catch (err) {
      console.error('Failed to get file info:', err);
    }
  };

  const handleRefresh = () => {
    const offset = (page - 1) * limit;
    if (tabValue === 0) {
      storeLoadRecordings(limit, offset);
    } else {
      storeLoadSnapshots(limit, offset);
    }
  };

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes';
    
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    
    // Determine the appropriate unit
    let i = 0;
    if (bytes >= k * k) { // >= 1MB
      i = 2; // MB
    } else if (bytes >= k) { // >= 1KB
      i = 1; // KB
    }
    
    // Calculate the value in the selected unit
    const value = bytes / Math.pow(k, i);
    
    // For MB and GB, round to whole numbers; for KB and Bytes, show 2 decimal places
    const formattedValue = i >= 2 ? Math.round(value) : parseFloat(value.toFixed(2));
    
    return `${formattedValue} ${sizes[i]}`;
  };

  const formatDuration = (seconds: number): string => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;
    return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  };

  const formatDate = (dateString: string): string => {
    return new Date(dateString).toLocaleString();
  };

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ mb: 3 }}>
        <Typography variant="h4" component="h1" gutterBottom>
          File Manager
        </Typography>
        <Typography variant="body1" color="text.secondary">
          Browse and download recordings and snapshots
        </Typography>
        {!canDeleteFiles && (
          <Alert severity="info" sx={{ mt: 2 }}>
            You have read-only access. Admin or operator permissions required for file deletion.
          </Alert>
        )}
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {error}
        </Alert>
      )}

      <Card>
        <CardContent>
          <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 2 }}>
            <Tabs value={tabValue} onChange={handleTabChange}>
              <Tab 
                label={
                  <Stack direction="row" spacing={1} alignItems="center">
                    <VideoFile />
                    <span>Recordings ({recordings?.length || 0})</span>
                  </Stack>
                } 
              />
              <Tab 
                label={
                  <Stack direction="row" spacing={1} alignItems="center">
                    <Image />
                    <span>Snapshots ({snapshots?.length || 0})</span>
                  </Stack>
                } 
              />
            </Tabs>
          </Box>

          <TabPanel value={tabValue} index={0}>
            <FileTable 
              files={storeRecordings}
              fileType="recordings"
              isLoading={storeIsLoading}
              onDownload={handleDownload}
              onDelete={handleDelete}
              onViewInfo={handleViewInfo}
              isDownloading={storeIsDownloading}
              isDeleting={storeIsDeleting}
              canDeleteFiles={storeCanDeleteFiles}
              formatFileSize={formatFileSize}
              formatDuration={formatDuration}
              formatDate={formatDate}
            />
          </TabPanel>

          <TabPanel value={tabValue} index={1}>
            <FileTable 
              files={storeSnapshots}
              fileType="snapshots"
              isLoading={storeIsLoading}
              onDownload={handleDownload}
              onDelete={handleDelete}
              onViewInfo={handleViewInfo}
              isDownloading={storeIsDownloading}
              isDeleting={storeIsDeleting}
              canDeleteFiles={storeCanDeleteFiles}
              formatFileSize={formatFileSize}
              formatDuration={formatDuration}
              formatDate={formatDate}
            />
          </TabPanel>

          <Box sx={{ mt: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Button
              startIcon={<Refresh />}
              onClick={handleRefresh}
              disabled={storeIsLoading}
            >
              Refresh
            </Button>

            <Pagination
              count={Math.ceil((currentFiles?.length || 0) / limit)}
              page={page}
              onChange={(_event, value) => setPage(value)}
              disabled={storeIsLoading}
            />
          </Box>
        </CardContent>
      </Card>

      {/* Delete Confirmation Dialog */}
      <Dialog
        open={deleteDialogOpen}
        onClose={() => setDeleteDialogOpen(false)}
        aria-labelledby="delete-dialog-title"
        aria-describedby="delete-dialog-description"
      >
        <DialogTitle id="delete-dialog-title">
          Confirm File Deletion
        </DialogTitle>
        <DialogContent>
          <DialogContentText id="delete-dialog-description">
            Are you sure you want to delete "{storeSelectedFile?.filename}"? This action cannot be undone.
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteDialogOpen(false)}>Cancel</Button>
          <Button 
            onClick={handleConfirmDelete} 
            color="error" 
            variant="contained"
            disabled={isDeleting}
          >
            {storeIsDeleting ? 'Deleting...' : 'Delete'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* File Info Dialog */}
      <Dialog
        open={fileInfoDialogOpen}
        onClose={() => setFileInfoDialogOpen(false)}
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle>
          File Information
        </DialogTitle>
        <DialogContent>
          {storeIsLoadingFileInfo ? (
            <Box display="flex" justifyContent="center" p={2}>
              <CircularProgress />
            </Box>
          ) : storeSelectedFile ? (
            <Grid container spacing={2}>
              <Grid item xs={12}>
                <Typography variant="h6" gutterBottom>
                  {selectedFile.filename}
                </Typography>
              </Grid>
              <Grid item xs={6}>
                <Typography variant="body2" color="text.secondary">
                  File Size
                </Typography>
                <Typography variant="body1">
                  {formatFileSize(storeSelectedFile.file_size)}
                </Typography>
              </Grid>
              <Grid item xs={6}>
                <Typography variant="body2" color="text.secondary">
                  Created
                </Typography>
                <Typography variant="body1">
                  {formatDate(selectedFile.created_time)}
                </Typography>
              </Grid>
              {fileType === 'recordings' && selectedFile.duration && (
                <Grid item xs={6}>
                  <Typography variant="body2" color="text.secondary">
                    Duration
                  </Typography>
                  <Typography variant="body1">
                    {formatDuration(selectedFile.duration)}
                  </Typography>
                </Grid>
              )}
              {fileType === 'snapshots' && selectedFile.resolution && (
                <Grid item xs={6}>
                  <Typography variant="body2" color="text.secondary">
                    Resolution
                  </Typography>
                  <Typography variant="body1">
                    {selectedFile.resolution}
                  </Typography>
                </Grid>
              )}
              <Grid item xs={12}>
                <Typography variant="body2" color="text.secondary">
                  Download URL
                </Typography>
                <Typography variant="body2" sx={{ wordBreak: 'break-all' }}>
                  {selectedFile.download_url}
                </Typography>
              </Grid>
            </Grid>
          ) : (
            <Typography color="text.secondary">
              No file information available
            </Typography>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setFileInfoDialogOpen(false)}>Close</Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

interface FileTableProps {
  files: FileItem[] | null;
  fileType: FileType;
  isLoading: boolean;
  onDownload: (filename: string) => void;
  onDelete: (file: FileItem) => void;
  onViewInfo: (file: FileItem) => void;
  isDownloading: boolean;
  isDeleting: boolean;
  canDeleteFiles: boolean;
  formatFileSize: (bytes: number) => string;
  formatDuration: (seconds: number) => string;
  formatDate: (dateString: string) => string;
}

const FileTable: React.FC<FileTableProps> = ({
  files,
  fileType,
  isLoading,
  onDownload,
  onDelete,
  onViewInfo,
  isDownloading,
  isDeleting,
  canDeleteFiles,
  formatFileSize,
  formatDuration,
  formatDate
}) => {
  if (isLoading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
        <CircularProgress />
      </Box>
    );
  }

  if (!files || files.length === 0) {
    return (
      <Box sx={{ textAlign: 'center', p: 4 }}>
        <Typography variant="body1" color="text.secondary">
          No {fileType} found
        </Typography>
      </Box>
    );
  }

  return (
    <TableContainer component={Paper}>
      <Table>
        <TableHead>
          <TableRow>
            <TableCell>Filename</TableCell>
            <TableCell>Size</TableCell>
            {fileType === 'recordings' && <TableCell>Duration</TableCell>}
            <TableCell>Created</TableCell>
            <TableCell align="right">Actions</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {files.map((file) => (
            <TableRow key={file.filename}>
              <TableCell>
                <Stack direction="row" spacing={1} alignItems="center">
                  {fileType === 'recordings' ? <VideoFile data-testid="VideoFileIcon" /> : <Image data-testid="ImageIcon" />}
                  <Typography variant="body2">{file.filename}</Typography>
                </Stack>
              </TableCell>
              <TableCell>{formatFileSize(file.file_size || 0)}</TableCell>
              {fileType === 'recordings' && (
                <TableCell>
                  {file.duration ? formatDuration(file.duration) : 'N/A'}
                </TableCell>
              )}
                              <TableCell>{formatDate(file.created_time)}</TableCell>
              <TableCell align="right">
                <Stack direction="row" spacing={1} justifyContent="flex-end">
                  <Tooltip title="View file information">
                    <IconButton
                      onClick={() => onViewInfo(file)}
                      color="primary"
                      size="small"
                    >
                      <Info />
                    </IconButton>
                  </Tooltip>
                  
                  <Tooltip title="Download file">
                    <IconButton
                      onClick={() => onDownload(file.filename)}
                      disabled={isDownloading}
                      color="primary"
                      size="small"
                    >
                      <Download />
                    </IconButton>
                  </Tooltip>
                  
                  {canDeleteFiles && (
                    <Tooltip title="Delete file">
                      <IconButton
                        onClick={() => onDelete(file)}
                        disabled={isDeleting}
                        color="error"
                        size="small"
                      >
                        <Delete />
                      </IconButton>
                    </Tooltip>
                  )}
                </Stack>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
};

export default FileManager;
