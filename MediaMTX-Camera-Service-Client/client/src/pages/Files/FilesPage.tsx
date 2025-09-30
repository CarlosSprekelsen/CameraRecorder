import React, { useEffect } from 'react';
import { Box } from '../../components/atoms/Box/Box';
import { Typography } from '../../components/atoms/Typography/Typography';
import { Container } from '../../components/atoms/Container/Container';
import { Alert } from '../../components/atoms/Alert/Alert';
import { CircularProgress } from '../../components/atoms/CircularProgress/CircularProgress';
import { useFileStore } from '../../stores/file/fileStore';
// ARCHITECTURE FIX: Removed serviceFactory import - components must use stores only
import { logger } from '../../services/logger/LoggerService';
import FileTabs from '../../components/Files/FileTabs';
import FileTable from '../../components/Files/FileTable';
import Pagination from '../../components/Files/Pagination';

/**
 * FilesPage - File management interface
 *
 * Provides comprehensive file management capabilities for recordings and snapshots.
 * Implements I.FileCatalog and I.FileActions interfaces for file operations including
 * listing, downloading, and deleting files with pagination support.
 *
 * @component
 * @returns {JSX.Element} The file management page
 *
 * @features
 * - File listing with pagination (recordings and snapshots)
 * - File download via server-provided URLs
 * - File deletion with confirmation
 * - File information display
 * - Loading states and error handling
 *
 * @example
 * ```tsx
 * <FilesPage />
 * ```
 *
 * @see {@link ../../docs/architecture/client-architechture.md} Client Architecture
 */
const FilesPage: React.FC = () => {
  const {
    recordings,
    snapshots,
    loading,
    error,
    pagination,
    currentTab,
    loadRecordings,
    loadSnapshots,
    // ARCHITECTURE FIX: Removed setFileService - stores handle service injection
    setCurrentTab,
  } = useFileStore();

  // Initialize file service and load data
  useEffect(() => {
    const initializeServiceAndLoadData = async () => {
      // ARCHITECTURE FIX: Removed direct service initialization - stores handle service injection
      // Services are initialized in App.tsx and injected into stores
      try {
        // Load initial data based on current tab
        if (currentTab === 'recordings') {
          await loadRecordings(pagination.limit, pagination.offset);
        } else {
          await loadSnapshots(pagination.limit, pagination.offset);
        }

        logger.info('Files page initialized successfully');
      } catch (error) {
        logger.error('Failed to initialize files page', error as Record<string, unknown>);
      }
    };

    initializeServiceAndLoadData();
  }, [
    // ARCHITECTURE FIX: Removed setFileService - stores handle service injection
    loadRecordings,
    loadSnapshots,
    currentTab,
    pagination.limit,
    pagination.offset,
  ]);

  // Handle tab change
  const handleTabChange = async (tab: 'recordings' | 'snapshots') => {
    setCurrentTab(tab);
    if (tab === 'recordings') {
      await loadRecordings(pagination.limit, 0);
    } else {
      await loadSnapshots(pagination.limit, 0);
    }
  };

  if (loading && recordings.length === 0 && snapshots.length === 0) {
    return (
      <Container
        maxWidth="lg"
        className="mt-4 flex justify-center items-center min-h-[60vh]"
      >
        <CircularProgress />
      </Container>
    );
  }

  return (
    <Container maxWidth="lg" className="mt-4">
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
        <Typography variant="h4" component="h1">
          Files
        </Typography>
      </Box>

      {error && (
        <Alert severity="error" className="mb-2">
          {error}
        </Alert>
      )}

      <FileTabs
        currentTab={currentTab}
        onTabChange={handleTabChange}
        recordingsCount={recordings.length}
        snapshotsCount={snapshots.length}
      />

      <Box sx={{ marginTop: 8 }}>
        <FileTable
          files={currentTab === 'recordings' ? recordings : snapshots}
          fileType={currentTab}
          loading={loading}
        />
      </Box>

      <Box sx={{ marginTop: 8, display: 'flex', justifyContent: 'center' }}>
        <Pagination
          pagination={pagination}
          onPageChange={(page) => {
            const newOffset = (page - 1) * pagination.limit;
            if (currentTab === 'recordings') {
              loadRecordings(pagination.limit, newOffset);
            } else {
              loadSnapshots(pagination.limit, newOffset);
            }
          }}
        />
      </Box>
    </Container>
  );
};

export default FilesPage;
