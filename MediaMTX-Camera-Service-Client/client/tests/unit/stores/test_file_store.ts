/**
 * FileStore unit tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * 
 * Requirements Coverage:
 * - REQ-FS-001: File catalog operations (list recordings/snapshots)
 * - REQ-FS-002: File actions (download, delete)
 * - REQ-FS-003: File selection and UI state management
 * - REQ-FS-004: Pagination handling
 * - REQ-FS-005: Error handling and state recovery
 * 
 * Test Categories: Unit
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */

import { useFileStore, FileInfo } from '../../../src/stores/file/fileStore';
import { FileService } from '../../../src/services/file/FileService';
import { APIMocks } from '../../utils/mocks';
import { APIResponseValidator } from '../../utils/validators';

// Mock the FileService
jest.mock('../../../src/services/file/FileService');

describe('FileStore Unit Tests', () => {
  let mockFileService: any;
  let store: ReturnType<typeof useFileStore>;

  beforeEach(() => {
    // Reset store state before each test
    store = useFileStore.getState();
    store.reset();

    // Create mock file service
    mockFileService = APIMocks.createMockFileService() as jest.Mocked<FileService>;
    
    // Clear all mocks
    jest.clearAllMocks();
  });

  afterEach(() => {
    // Reset store after each test
    store.reset();
  });

  describe('REQ-FS-001: File catalog operations', () => {
    beforeEach(() => {
      store.setFileService(mockFileService);
    });

    test('should load recordings successfully', async () => {
      const mockResponse = APIMocks.getListFilesResult('recordings');
      mockFileService.listRecordings.mockResolvedValue(mockResponse);

      await store.loadRecordings(20, 0);

      expect(mockFileService.listRecordings).toHaveBeenCalledWith(20, 0);
      expect(store.recordings).toEqual(mockResponse.files);
      expect(store.pagination).toEqual({
        limit: 20,
        offset: 0,
        total: mockResponse.total
      });
      expect(store.loading).toBe(false);
      expect(store.error).toBeNull();
      expect(APIResponseValidator.validateListFilesResult(mockResponse)).toBe(true);
    });

    test('should load snapshots successfully', async () => {
      const mockResponse = APIMocks.getListFilesResult('snapshots');
      mockFileService.listSnapshots.mockResolvedValue(mockResponse);

      await store.loadSnapshots(50, 0);

      expect(mockFileService.listSnapshots).toHaveBeenCalledWith(50, 0);
      expect(store.snapshots).toEqual(mockResponse.files);
      expect(store.pagination).toEqual({
        limit: 50,
        offset: 0,
        total: mockResponse.total
      });
      expect(store.loading).toBe(false);
      expect(store.error).toBeNull();
    });

    test('should use default pagination parameters', async () => {
      const mockResponse = APIMocks.getListFilesResult('recordings');
      mockFileService.listRecordings.mockResolvedValue(mockResponse);

      await store.loadRecordings();

      expect(mockFileService.listRecordings).toHaveBeenCalledWith(20, 0);
    });

    test('should get recording info successfully', async () => {
      const filename = 'recording_camera0_123456.mp4';
      const mockInfo: FileInfo = {
        filename,
        file_size: 1024000,
        modified_time: new Date().toISOString(),
        download_url: `https://localhost/downloads/${filename}`,
        duration: 60,
        format: 'mp4',
        device: 'camera0'
      };
      mockFileService.getRecordingInfo.mockResolvedValue(mockInfo);

      const result = await store.getRecordingInfo(filename);

      expect(mockFileService.getRecordingInfo).toHaveBeenCalledWith(filename);
      expect(result).toEqual(mockInfo);
      expect(store.error).toBeNull();
    });

    test('should get snapshot info successfully', async () => {
      const filename = 'snapshot_camera0_123456.jpg';
      const mockInfo: FileInfo = {
        filename,
        file_size: 512000,
        modified_time: new Date().toISOString(),
        download_url: `https://localhost/downloads/${filename}`
      };
      mockFileService.getSnapshotInfo.mockResolvedValue(mockInfo);

      const result = await store.getSnapshotInfo(filename);

      expect(mockFileService.getSnapshotInfo).toHaveBeenCalledWith(filename);
      expect(result).toEqual(mockInfo);
      expect(store.error).toBeNull();
    });

    test('should handle recordings loading error', async () => {
      const errorMessage = 'Failed to load recordings';
      mockFileService.listRecordings.mockRejectedValue(new Error(errorMessage));

      await store.loadRecordings();

      expect(store.loading).toBe(false);
      expect(store.error).toBe(errorMessage);
      expect(store.recordings).toEqual([]);
    });

    test('should handle snapshots loading error', async () => {
      const errorMessage = 'Failed to load snapshots';
      mockFileService.listSnapshots.mockRejectedValue(new Error(errorMessage));

      await store.loadSnapshots();

      expect(store.loading).toBe(false);
      expect(store.error).toBe(errorMessage);
      expect(store.snapshots).toEqual([]);
    });
  });

  describe('REQ-FS-002: File actions', () => {
    beforeEach(() => {
      store.setFileService(mockFileService);
    });

    test('should download file successfully', async () => {
      const downloadUrl = 'https://localhost/downloads/test.mp4';
      const filename = 'test.mp4';
      mockFileService.downloadFile.mockResolvedValue(undefined);

      await store.downloadFile(downloadUrl, filename);

      expect(mockFileService.downloadFile).toHaveBeenCalledWith(downloadUrl, filename);
      expect(store.error).toBeNull();
    });

    test('should handle download error', async () => {
      const downloadUrl = 'https://localhost/downloads/test.mp4';
      const filename = 'test.mp4';
      const errorMessage = 'Download failed';
      mockFileService.downloadFile.mockRejectedValue(new Error(errorMessage));

      await store.downloadFile(downloadUrl, filename);

      expect(store.error).toBe(errorMessage);
    });

    test('should delete recording successfully', async () => {
      const filename = 'recording_camera0_123456.mp4';
      mockFileService.deleteRecording.mockResolvedValue({ success: true, message: 'Deleted' });
      mockFileService.listRecordings.mockResolvedValue(APIMocks.getListFilesResult('recordings'));

      const result = await store.deleteRecording(filename);

      expect(mockFileService.deleteRecording).toHaveBeenCalledWith(filename);
      expect(result).toBe(true);
      expect(store.loading).toBe(false);
      expect(store.error).toBeNull();
    });

    test('should handle recording deletion failure', async () => {
      const filename = 'recording_camera0_123456.mp4';
      mockFileService.deleteRecording.mockResolvedValue({ success: false, message: 'Delete failed' });

      const result = await store.deleteRecording(filename);

      expect(result).toBe(false);
      expect(store.loading).toBe(false);
      expect(store.error).toBe('Delete failed');
    });

    test('should delete snapshot successfully', async () => {
      const filename = 'snapshot_camera0_123456.jpg';
      mockFileService.deleteSnapshot.mockResolvedValue({ success: true, message: 'Deleted' });
      mockFileService.listSnapshots.mockResolvedValue(APIMocks.getListFilesResult('snapshots'));

      const result = await store.deleteSnapshot(filename);

      expect(mockFileService.deleteSnapshot).toHaveBeenCalledWith(filename);
      expect(result).toBe(true);
      expect(store.loading).toBe(false);
      expect(store.error).toBeNull();
    });

    test('should handle snapshot deletion failure', async () => {
      const filename = 'snapshot_camera0_123456.jpg';
      mockFileService.deleteSnapshot.mockResolvedValue({ success: false, message: 'Delete failed' });

      const result = await store.deleteSnapshot(filename);

      expect(result).toBe(false);
      expect(store.loading).toBe(false);
      expect(store.error).toBe('Delete failed');
    });

    test('should handle missing file service for download', async () => {
      await store.downloadFile('url', 'filename');
      expect(store.error).toBe('File service not initialized');
    });

    test('should handle missing file service for delete operations', async () => {
      const result1 = await store.deleteRecording('filename');
      const result2 = await store.deleteSnapshot('filename');
      
      expect(result1).toBe(false);
      expect(result2).toBe(false);
      expect(store.error).toBe('File service not initialized');
    });
  });

  describe('REQ-FS-003: File selection and UI state management', () => {
    test('should set current tab correctly', () => {
      store.setCurrentTab('snapshots');
      expect(store.currentTab).toBe('snapshots');
      
      store.setCurrentTab('recordings');
      expect(store.currentTab).toBe('recordings');
    });

    test('should set selected files correctly', () => {
      const files = ['file1.mp4', 'file2.mp4'];
      store.setSelectedFiles(files);
      expect(store.selectedFiles).toEqual(files);
    });

    test('should toggle file selection correctly', () => {
      const filename = 'test.mp4';
      
      // Add file to selection
      store.toggleFileSelection(filename);
      expect(store.selectedFiles).toContain(filename);
      
      // Remove file from selection
      store.toggleFileSelection(filename);
      expect(store.selectedFiles).not.toContain(filename);
    });

    test('should add multiple files to selection', () => {
      store.toggleFileSelection('file1.mp4');
      store.toggleFileSelection('file2.mp4');
      
      expect(store.selectedFiles).toHaveLength(2);
      expect(store.selectedFiles).toContain('file1.mp4');
      expect(store.selectedFiles).toContain('file2.mp4');
    });

    test('should clear selection', () => {
      store.setSelectedFiles(['file1.mp4', 'file2.mp4']);
      store.clearSelection();
      expect(store.selectedFiles).toEqual([]);
    });

    test('should handle mixed selection operations', () => {
      // Set initial selection
      store.setSelectedFiles(['file1.mp4']);
      
      // Add more files
      store.toggleFileSelection('file2.mp4');
      expect(store.selectedFiles).toHaveLength(2);
      
      // Remove one file
      store.toggleFileSelection('file1.mp4');
      expect(store.selectedFiles).toEqual(['file2.mp4']);
    });
  });

  describe('REQ-FS-004: Pagination handling', () => {
    test('should set pagination correctly', () => {
      const limit = 50;
      const offset = 100;
      const total = 500;
      
      store.setPagination(limit, offset, total);
      
      expect(store.pagination).toEqual({ limit, offset, total });
    });

    test('should go to next page correctly', () => {
      store.setPagination(20, 0, 100);
      store.nextPage();
      
      expect(store.pagination.offset).toBe(20);
    });

    test('should not go beyond total records', () => {
      store.setPagination(20, 80, 100);
      store.nextPage();
      
      // Should not change offset since 80 + 20 = 100 (at total)
      expect(store.pagination.offset).toBe(80);
    });

    test('should go to previous page correctly', () => {
      store.setPagination(20, 40, 100);
      store.prevPage();
      
      expect(store.pagination.offset).toBe(20);
    });

    test('should not go below zero offset', () => {
      store.setPagination(20, 0, 100);
      store.prevPage();
      
      expect(store.pagination.offset).toBe(0);
    });

    test('should go to specific page correctly', () => {
      store.setPagination(20, 0, 100);
      store.goToPage(3);
      
      expect(store.pagination.offset).toBe(40); // (3-1) * 20
    });

    test('should handle page navigation edge cases', () => {
      store.setPagination(20, 0, 100);
      
      // Go to page 0 (should stay at offset 0)
      store.goToPage(0);
      expect(store.pagination.offset).toBe(0);
      
      // Go to page 1 (should stay at offset 0)
      store.goToPage(1);
      expect(store.pagination.offset).toBe(0);
      
      // Go to very high page number
      store.goToPage(100);
      expect(store.pagination.offset).toBe(1980); // (100-1) * 20
    });
  });

  describe('REQ-FS-005: Error handling and state recovery', () => {
    test('should handle missing file service for load operations', async () => {
      await store.loadRecordings();
      await store.loadSnapshots();
      
      expect(store.error).toBe('File service not initialized');
    });

    test('should handle missing file service for info operations', async () => {
      const result1 = await store.getRecordingInfo('filename');
      const result2 = await store.getSnapshotInfo('filename');
      
      expect(result1).toBeNull();
      expect(result2).toBeNull();
      expect(store.error).toBe('File service not initialized');
    });

    test('should handle unknown error types', async () => {
      store.setFileService(mockFileService);
      mockFileService.listRecordings.mockRejectedValue('Unknown error');

      await store.loadRecordings();

      expect(store.loading).toBe(false);
      expect(store.error).toBe('Failed to load recordings');
    });

    test('should handle deletion error', async () => {
      store.setFileService(mockFileService);
      const errorMessage = 'Delete failed';
      mockFileService.deleteRecording.mockRejectedValue(new Error(errorMessage));

      const result = await store.deleteRecording('filename');

      expect(result).toBe(false);
      expect(store.loading).toBe(false);
      expect(store.error).toBe(errorMessage);
    });

    test('should reset to initial state', () => {
      // Set some state
      store.setLoading(true);
      store.setError('Test error');
      store.setSelectedFiles(['file1.mp4']);
      store.setCurrentTab('snapshots');
      store.setPagination(50, 100, 500);
      
      // Reset
      store.reset();
      
      expect(store.loading).toBe(false);
      expect(store.error).toBeNull();
      expect(store.selectedFiles).toEqual([]);
      expect(store.currentTab).toBe('recordings');
      expect(store.pagination).toEqual({ limit: 20, offset: 0, total: 0 });
      expect(store.recordings).toEqual([]);
      expect(store.snapshots).toEqual([]);
    });
  });

  describe('API Compliance Tests', () => {
    beforeEach(() => {
      store.setFileService(mockFileService);
    });

    test('should return files that match API schema', async () => {
      const mockResponse = APIMocks.getListFilesResult('recordings');
      mockFileService.listRecordings.mockResolvedValue(mockResponse);

      await store.loadRecordings();

      expect(APIResponseValidator.validateListFilesResult(mockResponse)).toBe(true);
      mockResponse.files.forEach(file => {
        expect(APIResponseValidator.validateRecordingFile(file)).toBe(true);
      });
    });

    test('should handle pagination parameters correctly', async () => {
      const mockResponse = APIMocks.getListFilesResult('recordings');
      mockFileService.listRecordings.mockResolvedValue(mockResponse);

      await store.loadRecordings(100, 50);

      expect(mockFileService.listRecordings).toHaveBeenCalledWith(100, 50);
      expect(APIResponseValidator.validatePaginationParams(100, 50)).toBe(true);
    });
  });

  describe('Edge Cases and Complex Scenarios', () => {
    beforeEach(() => {
      store.setFileService(mockFileService);
    });

    test('should handle empty file lists', async () => {
      const emptyResponse = {
        files: [],
        total: 0,
        limit: 20,
        offset: 0
      };
      mockFileService.listRecordings.mockResolvedValue(emptyResponse);

      await store.loadRecordings();

      expect(store.recordings).toEqual([]);
      expect(store.pagination.total).toBe(0);
    });

    test('should handle concurrent file operations', async () => {
      const mockResponse = APIMocks.getListFilesResult('recordings');
      mockFileService.listRecordings.mockResolvedValue(mockResponse);

      // Start concurrent operations
      const promises = [
        store.loadRecordings(),
        store.loadSnapshots(),
        store.getRecordingInfo('file1.mp4')
      ];

      await Promise.all(promises);

      expect(mockFileService.listRecordings).toHaveBeenCalledTimes(1);
      expect(mockFileService.listSnapshots).toHaveBeenCalledTimes(1);
      expect(mockFileService.getRecordingInfo).toHaveBeenCalledTimes(1);
    });

    test('should handle rapid selection changes', () => {
      const files = ['file1.mp4', 'file2.mp4', 'file3.mp4'];
      
      // Rapid selection changes
      files.forEach(file => store.toggleFileSelection(file));
      expect(store.selectedFiles).toEqual(files);
      
      // Rapid deselection
      files.forEach(file => store.toggleFileSelection(file));
      expect(store.selectedFiles).toEqual([]);
    });

    test('should handle large file lists', async () => {
      const largeFileList = Array.from({ length: 1000 }, (_, i) => ({
        filename: `recording_${i}.mp4`,
        file_size: 1024000,
        modified_time: new Date().toISOString(),
        download_url: `https://localhost/downloads/recording_${i}.mp4`
      }));
      
      const largeResponse = {
        files: largeFileList,
        total: 1000,
        limit: 20,
        offset: 0
      };
      
      mockFileService.listRecordings.mockResolvedValue(largeResponse);

      await store.loadRecordings();

      expect(store.recordings).toHaveLength(1000);
      expect(store.pagination.total).toBe(1000);
    });

    test('should handle file info with missing optional fields', async () => {
      const filename = 'recording_camera0_123456.mp4';
      const minimalInfo: FileInfo = {
        filename,
        file_size: 1024000,
        modified_time: new Date().toISOString(),
        download_url: `https://localhost/downloads/${filename}`
        // Missing optional fields: duration, format, device
      };
      mockFileService.getRecordingInfo.mockResolvedValue(minimalInfo);

      const result = await store.getRecordingInfo(filename);

      expect(result).toEqual(minimalInfo);
      expect(result?.duration).toBeUndefined();
      expect(result?.format).toBeUndefined();
      expect(result?.device).toBeUndefined();
    });
  });
});
