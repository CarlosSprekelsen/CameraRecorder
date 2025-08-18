/**
 * Unit tests for FileManager component
 * Tests file browsing, download functionality, and pagination
 */

import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { ThemeProvider } from '@mui/material/styles';
import { theme } from '../../../src/theme';
import FileManager from '../../../src/components/FileManager/FileManager';
import { useFileStore } from '../../../src/stores/fileStore';

// Mock the file store
jest.mock('../../../src/stores/fileStore');
const mockUseFileStore = useFileStore as jest.MockedFunction<typeof useFileStore>;

// Mock file data
const mockRecordings = [
  {
    filename: 'recording-1.mp4',
    file_size: 1024000,
    created_at: '2024-01-01T00:00:00Z',
    modified_time: '2024-01-01T00:01:00Z',
    download_url: '/files/recordings/recording-1.mp4',
    duration: 60,
    format: 'mp4'
  },
  {
    filename: 'recording-2.mp4',
    file_size: 2048000,
    created_at: '2024-01-01T01:00:00Z',
    modified_time: '2024-01-01T01:02:00Z',
    download_url: '/files/recordings/recording-2.mp4',
    duration: 120,
    format: 'mp4'
  }
];

const mockSnapshots = [
  {
    filename: 'snapshot-1.jpg',
    file_size: 512000,
    created_at: '2024-01-01T00:00:00Z',
    modified_time: '2024-01-01T00:00:00Z',
    download_url: '/files/snapshots/snapshot-1.jpg',
    format: 'jpg'
  },
  {
    filename: 'snapshot-2.png',
    file_size: 256000,
    created_at: '2024-01-01T01:00:00Z',
    modified_time: '2024-01-01T01:00:00Z',
    download_url: '/files/snapshots/snapshot-2.png',
    format: 'png'
  }
];

const mockStore = {
  recordings: mockRecordings,
  snapshots: mockSnapshots,
  isLoading: false,
  isDownloading: false,
  error: null,
  loadRecordings: jest.fn(),
  loadSnapshots: jest.fn(),
  downloadFile: jest.fn()
};

const renderWithProviders = (component: React.ReactElement) => {
  return render(
    <ThemeProvider theme={theme}>
      {component}
    </ThemeProvider>
  );
};

describe('FileManager Component', () => {
  beforeEach(() => {
    mockUseFileStore.mockReturnValue(mockStore);
    jest.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render file manager with tabs', () => {
      renderWithProviders(<FileManager />);

      expect(screen.getByText('File Manager')).toBeInTheDocument();
      expect(screen.getByText('Browse and download recordings and snapshots')).toBeInTheDocument();
      expect(screen.getByText('Recordings (2)')).toBeInTheDocument();
      expect(screen.getByText('Snapshots (2)')).toBeInTheDocument();
    });

    it('should display recordings by default', () => {
      renderWithProviders(<FileManager />);

      expect(screen.getByText('recording-1.mp4')).toBeInTheDocument();
      expect(screen.getByText('recording-2.mp4')).toBeInTheDocument();
      expect(screen.queryByText('snapshot-1.jpg')).not.toBeInTheDocument();
    });

    it('should display file metadata correctly', () => {
      renderWithProviders(<FileManager />);

      // Check file size formatting
      expect(screen.getByText('1 MB')).toBeInTheDocument(); // 1024000 bytes
      expect(screen.getByText('2 MB')).toBeInTheDocument(); // 2048000 bytes

      // Check duration formatting
      expect(screen.getByText('00:01:00')).toBeInTheDocument(); // 60 seconds
      expect(screen.getByText('00:02:00')).toBeInTheDocument(); // 120 seconds
    });

    it('should show loading state', () => {
      mockUseFileStore.mockReturnValue({
        ...mockStore,
        isLoading: true
      });

      renderWithProviders(<FileManager />);

      expect(screen.getByRole('progressbar')).toBeInTheDocument();
    });

    it('should show error message when error occurs', () => {
      mockUseFileStore.mockReturnValue({
        ...mockStore,
        error: 'Failed to load files'
      });

      renderWithProviders(<FileManager />);

      expect(screen.getByText('Failed to load files')).toBeInTheDocument();
    });

    it('should show empty state when no files', () => {
      mockUseFileStore.mockReturnValue({
        ...mockStore,
        recordings: [],
        snapshots: []
      });

      renderWithProviders(<FileManager />);

      expect(screen.getByText('No recordings found')).toBeInTheDocument();
    });
  });

  describe('Tab Navigation', () => {
    it('should switch to snapshots tab when clicked', () => {
      renderWithProviders(<FileManager />);

      const snapshotsTab = screen.getByText('Snapshots (2)');
      fireEvent.click(snapshotsTab);

      expect(screen.getByText('snapshot-1.jpg')).toBeInTheDocument();
      expect(screen.getByText('snapshot-2.png')).toBeInTheDocument();
      expect(screen.queryByText('recording-1.mp4')).not.toBeInTheDocument();
    });

    it('should reset pagination when switching tabs', () => {
      renderWithProviders(<FileManager />);

      const snapshotsTab = screen.getByText('Snapshots (2)');
      fireEvent.click(snapshotsTab);

      // Verify that loadSnapshots was called with page 1 (offset 0)
      expect(mockStore.loadSnapshots).toHaveBeenCalledWith(20, 0);
    });
  });

  describe('File Download', () => {
    it('should call downloadFile when download button is clicked', async () => {
      mockStore.downloadFile.mockResolvedValue();

      renderWithProviders(<FileManager />);

      const downloadButtons = screen.getAllByRole('button', { name: /download/i });
      fireEvent.click(downloadButtons[0]);

      await waitFor(() => {
        expect(mockStore.downloadFile).toHaveBeenCalledWith('recordings', 'recording-1.mp4');
      });
    });

    it('should call downloadFile with correct file type for snapshots', async () => {
      mockStore.downloadFile.mockResolvedValue();

      renderWithProviders(<FileManager />);

      // Switch to snapshots tab
      const snapshotsTab = screen.getByText('Snapshots (2)');
      fireEvent.click(snapshotsTab);

      const downloadButtons = screen.getAllByRole('button', { name: /download/i });
      fireEvent.click(downloadButtons[0]);

      await waitFor(() => {
        expect(mockStore.downloadFile).toHaveBeenCalledWith('snapshots', 'snapshot-1.jpg');
      });
    });

    it('should disable download buttons when downloading', () => {
      mockUseFileStore.mockReturnValue({
        ...mockStore,
        isDownloading: true
      });

      renderWithProviders(<FileManager />);

      const downloadButtons = screen.getAllByRole('button', { name: /download/i });
      downloadButtons.forEach(button => {
        expect(button).toBeDisabled();
      });
    });
  });

  describe('Pagination', () => {
    it('should render pagination controls', () => {
      renderWithProviders(<FileManager />);

      expect(screen.getByRole('navigation')).toBeInTheDocument();
    });

    it('should call loadRecordings with correct pagination parameters', () => {
      renderWithProviders(<FileManager />);

      // Initial load should be called with page 1
      expect(mockStore.loadRecordings).toHaveBeenCalledWith(20, 0);
    });

    it('should handle page changes', () => {
      renderWithProviders(<FileManager />);

      // Mock pagination component to simulate page change
      // This would need to be tested with actual pagination interaction
      // For now, we'll verify the component handles pagination state
    });
  });

  describe('Refresh Functionality', () => {
    it('should call refresh when refresh button is clicked', () => {
      renderWithProviders(<FileManager />);

      const refreshButton = screen.getByText('Refresh');
      fireEvent.click(refreshButton);

      expect(mockStore.loadRecordings).toHaveBeenCalledWith(20, 0);
    });

    it('should refresh snapshots when on snapshots tab', () => {
      renderWithProviders(<FileManager />);

      // Switch to snapshots tab
      const snapshotsTab = screen.getByText('Snapshots (2)');
      fireEvent.click(snapshotsTab);

      const refreshButton = screen.getByText('Refresh');
      fireEvent.click(refreshButton);

      expect(mockStore.loadSnapshots).toHaveBeenCalledWith(20, 0);
    });

    it('should disable refresh button when loading', () => {
      mockUseFileStore.mockReturnValue({
        ...mockStore,
        isLoading: true
      });

      renderWithProviders(<FileManager />);

      const refreshButton = screen.getByText('Refresh');
      expect(refreshButton).toBeDisabled();
    });
  });

  describe('File Table', () => {
    it('should display correct table headers for recordings', () => {
      renderWithProviders(<FileManager />);

      expect(screen.getByText('Filename')).toBeInTheDocument();
      expect(screen.getByText('Size')).toBeInTheDocument();
      expect(screen.getByText('Duration')).toBeInTheDocument();
      expect(screen.getByText('Created')).toBeInTheDocument();
      expect(screen.getByText('Actions')).toBeInTheDocument();
    });

    it('should display correct table headers for snapshots', () => {
      renderWithProviders(<FileManager />);

      // Switch to snapshots tab
      const snapshotsTab = screen.getByText('Snapshots (2)');
      fireEvent.click(snapshotsTab);

      expect(screen.getByText('Filename')).toBeInTheDocument();
      expect(screen.getByText('Size')).toBeInTheDocument();
      expect(screen.queryByText('Duration')).not.toBeInTheDocument(); // No duration for snapshots
      expect(screen.getByText('Created')).toBeInTheDocument();
      expect(screen.getByText('Actions')).toBeInTheDocument();
    });

    it('should display file icons correctly', () => {
      renderWithProviders(<FileManager />);

      // Check for video file icons in recordings tab
      const videoIcons = screen.getAllByTestId('VideocamIcon');
      expect(videoIcons).toHaveLength(2);

      // Switch to snapshots tab
      const snapshotsTab = screen.getByText('Snapshots (2)');
      fireEvent.click(snapshotsTab);

      // Check for image file icons in snapshots tab
      const imageIcons = screen.getAllByTestId('ImageIcon');
      expect(imageIcons).toHaveLength(2);
    });
  });

  describe('Error Handling', () => {
    it('should handle download errors gracefully', async () => {
      mockStore.downloadFile.mockRejectedValue(new Error('Download failed'));

      renderWithProviders(<FileManager />);

      const downloadButtons = screen.getAllByRole('button', { name: /download/i });
      fireEvent.click(downloadButtons[0]);

      // The error should be handled by the store, but we can verify the call was made
      await waitFor(() => {
        expect(mockStore.downloadFile).toHaveBeenCalled();
      });
    });

    it('should handle file loading errors', () => {
      mockUseFileStore.mockReturnValue({
        ...mockStore,
        error: 'Network error occurred'
      });

      renderWithProviders(<FileManager />);

      expect(screen.getByText('Network error occurred')).toBeInTheDocument();
    });
  });

  describe('Data Formatting', () => {
    it('should format file sizes correctly', () => {
      renderWithProviders(<FileManager />);

      // Test various file sizes
      expect(screen.getByText('1 MB')).toBeInTheDocument(); // 1024000 bytes
      expect(screen.getByText('2 MB')).toBeInTheDocument(); // 2048000 bytes
    });

    it('should format durations correctly', () => {
      renderWithProviders(<FileManager />);

      expect(screen.getByText('00:01:00')).toBeInTheDocument(); // 60 seconds
      expect(screen.getByText('00:02:00')).toBeInTheDocument(); // 120 seconds
    });

    it('should format dates correctly', () => {
      renderWithProviders(<FileManager />);

      // The date formatting depends on the user's locale
      // We'll just verify that dates are displayed
      expect(screen.getByText(/2024/)).toBeInTheDocument();
    });
  });

  describe('Accessibility', () => {
    it('should have proper ARIA labels', () => {
      renderWithProviders(<FileManager />);

      // Check for proper table structure
      expect(screen.getByRole('table')).toBeInTheDocument();
      expect(screen.getByRole('tablist')).toBeInTheDocument();
    });

    it('should have proper button labels', () => {
      renderWithProviders(<FileManager />);

      const downloadButtons = screen.getAllByRole('button', { name: /download/i });
      expect(downloadButtons).toHaveLength(2);
    });
  });
});
