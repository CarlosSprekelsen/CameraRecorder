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
  Chip,
  CircularProgress,
  Alert,
  Pagination,
  TextField,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Stack
} from '@mui/material';
import {
  Download,
  Refresh,
  Search,
  Sort,
  VideoFile,
  Image
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
  const [sortBy, setSortBy] = useState<'date' | 'size' | 'name'>('date');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc');
  const [searchTerm, setSearchTerm] = useState('');

  const {
    recordings,
    snapshots,
    isLoading,
    error,
    loadRecordings,
    loadSnapshots,
    downloadFile,
    isDownloading
  } = useFileStore();

  const currentFiles = tabValue === 0 ? recordings : snapshots;
  const fileType: FileType = tabValue === 0 ? 'recordings' : 'snapshots';

  useEffect(() => {
    const offset = (page - 1) * limit;
    if (tabValue === 0) {
      loadRecordings(limit, offset);
    } else {
      loadSnapshots(limit, offset);
    }
  }, [tabValue, page, limit, loadRecordings, loadSnapshots]);

  const handleTabChange = (event: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue);
    setPage(1); // Reset to first page when switching tabs
  };

  const handleDownload = async (filename: string) => {
    try {
      await downloadFile(fileType, filename);
    } catch (err) {
      console.error('Download failed:', err);
    }
  };

  const handleRefresh = () => {
    const offset = (page - 1) * limit;
    if (tabValue === 0) {
      loadRecordings(limit, offset);
    } else {
      loadSnapshots(limit, offset);
    }
  };

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
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
              files={recordings}
              fileType="recordings"
              isLoading={isLoading}
              onDownload={handleDownload}
              isDownloading={isDownloading}
              formatFileSize={formatFileSize}
              formatDuration={formatDuration}
              formatDate={formatDate}
            />
          </TabPanel>

          <TabPanel value={tabValue} index={1}>
            <FileTable 
              files={snapshots}
              fileType="snapshots"
              isLoading={isLoading}
              onDownload={handleDownload}
              isDownloading={isDownloading}
              formatFileSize={formatFileSize}
              formatDuration={formatDuration}
              formatDate={formatDate}
            />
          </TabPanel>

          <Box sx={{ mt: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Button
              startIcon={<Refresh />}
              onClick={handleRefresh}
              disabled={isLoading}
            >
              Refresh
            </Button>

            <Pagination
              count={Math.ceil((currentFiles?.length || 0) / limit)}
              page={page}
              onChange={(event, value) => setPage(value)}
              disabled={isLoading}
            />
          </Box>
        </CardContent>
      </Card>
    </Box>
  );
};

interface FileTableProps {
  files: FileItem[] | null;
  fileType: FileType;
  isLoading: boolean;
  onDownload: (filename: string) => void;
  isDownloading: boolean;
  formatFileSize: (bytes: number) => string;
  formatDuration: (seconds: number) => string;
  formatDate: (dateString: string) => string;
}

const FileTable: React.FC<FileTableProps> = ({
  files,
  fileType,
  isLoading,
  onDownload,
  isDownloading,
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
                  {fileType === 'recordings' ? <VideoFile /> : <Image />}
                  <Typography variant="body2">{file.filename}</Typography>
                </Stack>
              </TableCell>
              <TableCell>{formatFileSize(file.file_size || 0)}</TableCell>
              {fileType === 'recordings' && (
                <TableCell>
                  {file.duration ? formatDuration(file.duration) : 'N/A'}
                </TableCell>
              )}
              <TableCell>{formatDate(file.created_at)}</TableCell>
              <TableCell align="right">
                <IconButton
                  onClick={() => onDownload(file.filename)}
                  disabled={isDownloading}
                  color="primary"
                >
                  <Download />
                </IconButton>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
};

export default FileManager;
