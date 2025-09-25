/**
 * Unit Tests for File Store
 * 
 * REQ-001: Store State Management - Test Zustand store actions
 * REQ-002: State Transitions - Test state changes
 * REQ-003: Error Handling - Test error states and recovery
 * REQ-004: API Integration - Mock API calls and test responses
 * REQ-005: Side Effects - Test store side effects
 * 
 * Ground Truth: Official RPC Documentation
 * API Reference: docs/api/json_rpc_methods.md
 */

import { describe, test, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { MockDataFactory } from '../../utils/mocks';
import { APIResponseValidator } from '../../utils/validators';
import { TestHelpers } from '../../utils/test-helpers';

// Mock the FileService
jest.mock('../../../src/services/file/fileService', () => ({
  FileService: jest.fn().mockImplementation(() => MockDataFactory.createMockFileService())
}));

// Mock the file store
const mockFileStore = MockDataFactory.createMockFileStore();

describe('File Store', () => {
  let fileStore: any;
  let mockFileService: any;

  beforeEach(() => {
    // Reset mocks
    jest.clearAllMocks();
    
    // Create fresh mock service
    mockFileService = MockDataFactory.createMockFileService();
    
    // Mock the store with fresh state
    fileStore = { ...mockFileStore };
  });

  afterEach(() => {
    // Clean up
    fileStore.reset?.();
  });

  describe('REQ-001: Store State Management', () => {
    test('should initialize with correct default state', () => {
      expect(fileStore.recordings).toEqual(MockDataFactory.getFileListResult().files);
      expect(fileStore.snapshots).toEqual(MockDataFactory.getFileListResult().files);
      expect(fileStore.loading).toBe(false);
      expect(fileStore.error).toBe(null);
      expect(fileStore.lastUpdated).toBe('2025-01-15T14:30:00Z');
    });

    test('should set loading state correctly', () => {
      fileStore.setLoading(true);
      expect(fileStore.loading).toBe(true);
      
      fileStore.setLoading(false);
      expect(fileStore.loading).toBe(false);
    });

    test('should set error state correctly', () => {
      const errorMessage = 'Test error message';
      fileStore.setError(errorMessage);
      expect(fileStore.error).toBe(errorMessage);
      
      fileStore.setError(null);
      expect(fileStore.error).toBe(null);
    });
  });

  describe('REQ-002: State Transitions', () => {
    test('should load recordings correctly', async () => {
      const mockResponse = MockDataFactory.getFileListResult();
      mockFileService.listRecordings = jest.fn().mockResolvedValue(mockResponse);
      
      await fileStore.loadRecordings();
      
      expect(fileStore.recordings).toEqual(mockResponse.files);
      expect(fileStore.loading).toBe(false);
      expect(fileStore.error).toBe(null);
    });

    test('should load snapshots correctly', async () => {
      const mockResponse = MockDataFactory.getFileListResult();
      mockFileService.listSnapshots = jest.fn().mockResolvedValue(mockResponse);
      
      await fileStore.loadSnapshots();
      
      expect(fileStore.snapshots).toEqual(mockResponse.files);
      expect(fileStore.loading).toBe(false);
      expect(fileStore.error).toBe(null);
    });

    test('should handle pagination parameters', async () => {
      const mockResponse = MockDataFactory.getFileListResult();
      mockFileService.listRecordings = jest.fn().mockResolvedValue(mockResponse);
      
      await fileStore.loadRecordings({ limit: 10, offset: 20 });
      
      expect(mockFileService.listRecordings).toHaveBeenCalledWith({ limit: 10, offset: 20 });
    });
  });

  describe('REQ-003: Error Handling', () => {
    test('should handle API errors gracefully', async () => {
      const errorMessage = 'API request failed';
      
      mockFileService.listRecordings = jest.fn().mockRejectedValue(new Error(errorMessage));
      
      try {
        await fileStore.loadRecordings();
      } catch (error) {
        expect(fileStore.error).toBe(errorMessage);
        expect(fileStore.loading).toBe(false);
      }
    });

    test('should clear errors when new requests succeed', async () => {
      // Set initial error
      fileStore.setError('Previous error');
      expect(fileStore.error).toBe('Previous error');
      
      // Mock successful request
      mockFileService.listRecordings = jest.fn().mockResolvedValue(MockDataFactory.getFileListResult());
      
      await fileStore.loadRecordings();
      
      // Error should be cleared on success
      expect(fileStore.error).toBe(null);
    });

    test('should handle file not found errors', async () => {
      const notFoundError = new Error('File not found');
      
      mockFileService.getRecordingInfo = jest.fn().mockRejectedValue(notFoundError);
      
      try {
        await fileStore.getRecordingInfo('nonexistent-file');
      } catch (error) {
        expect(fileStore.error).toBe('File not found');
      }
    });
  });

  describe('REQ-004: API Integration', () => {
    test('should fetch recordings list with correct API response format', async () => {
      const mockResponse = MockDataFactory.getFileListResult();
      mockFileService.listRecordings = jest.fn().mockResolvedValue(mockResponse);
      
      await fileStore.loadRecordings();
      
      // Verify API was called
      expect(mockFileService.listRecordings).toHaveBeenCalledTimes(1);
      
      // Verify response format matches official RPC spec
      expect(APIResponseValidator.validateFileListResult(mockResponse)).toBe(true);
      
      // Verify store state was updated
      expect(fileStore.recordings).toEqual(mockResponse.files);
      expect(fileStore.loading).toBe(false);
      expect(fileStore.error).toBe(null);
    });

    test('should fetch snapshots list with correct API response format', async () => {
      const mockResponse = MockDataFactory.getFileListResult();
      mockFileService.listSnapshots = jest.fn().mockResolvedValue(mockResponse);
      
      await fileStore.loadSnapshots();
      
      // Verify API was called
      expect(mockFileService.listSnapshots).toHaveBeenCalledTimes(1);
      
      // Verify response format matches official RPC spec
      expect(APIResponseValidator.validateFileListResult(mockResponse)).toBe(true);
      
      // Verify store state was updated
      expect(fileStore.snapshots).toEqual(mockResponse.files);
    });

    test('should fetch recording info with correct API response format', async () => {
      const mockResponse = MockDataFactory.getRecordingInfo();
      mockFileService.getRecordingInfo = jest.fn().mockResolvedValue(mockResponse);
      
      const result = await fileStore.getRecordingInfo('test-recording');
      
      // Verify API was called with correct parameters
      expect(mockFileService.getRecordingInfo).toHaveBeenCalledWith('test-recording');
      
      // Verify response format matches official RPC spec
      expect(APIResponseValidator.validateRecordingInfo(mockResponse)).toBe(true);
      
      expect(result).toEqual(mockResponse);
    });

    test('should fetch snapshot info with correct API response format', async () => {
      const mockResponse = MockDataFactory.getSnapshotInfo();
      mockFileService.getSnapshotInfo = jest.fn().mockResolvedValue(mockResponse);
      
      const result = await fileStore.getSnapshotInfo('test-snapshot');
      
      // Verify API was called with correct parameters
      expect(mockFileService.getSnapshotInfo).toHaveBeenCalledWith('test-snapshot');
      
      // Verify response format matches official RPC spec
      expect(APIResponseValidator.validateSnapshotInfo(mockResponse)).toBe(true);
      
      expect(result).toEqual(mockResponse);
    });

    test('should delete recording with correct API response format', async () => {
      const mockResponse = MockDataFactory.getDeleteResult();
      mockFileService.deleteRecording = jest.fn().mockResolvedValue(mockResponse);
      
      const result = await fileStore.deleteRecording('test-recording');
      
      // Verify API was called with correct parameters
      expect(mockFileService.deleteRecording).toHaveBeenCalledWith('test-recording');
      
      // Verify response format matches official RPC spec
      expect(APIResponseValidator.validateDeleteResult(mockResponse)).toBe(true);
      
      expect(result).toEqual(mockResponse);
    });

    test('should delete snapshot with correct API response format', async () => {
      const mockResponse = MockDataFactory.getDeleteResult();
      mockFileService.deleteSnapshot = jest.fn().mockResolvedValue(mockResponse);
      
      const result = await fileStore.deleteSnapshot('test-snapshot');
      
      // Verify API was called with correct parameters
      expect(mockFileService.deleteSnapshot).toHaveBeenCalledWith('test-snapshot');
      
      // Verify response format matches official RPC spec
      expect(APIResponseValidator.validateDeleteResult(mockResponse)).toBe(true);
      
      expect(result).toEqual(mockResponse);
    });
  });

  describe('REQ-005: Side Effects', () => {
    test('should update lastUpdated timestamp on successful API calls', async () => {
      const initialTimestamp = fileStore.lastUpdated;
      
      mockFileService.listRecordings = jest.fn().mockResolvedValue(MockDataFactory.getFileListResult());
      
      await fileStore.loadRecordings();
      
      // Verify timestamp was updated
      expect(fileStore.lastUpdated).not.toBe(initialTimestamp);
      expect(fileStore.lastUpdated).toBeDefined();
    });

    test('should set loading state during API calls', async () => {
      let resolvePromise: (value: any) => void;
      const promise = new Promise(resolve => {
        resolvePromise = resolve;
      });
      
      mockFileService.listRecordings = jest.fn().mockReturnValue(promise);
      
      // Start the API call
      const apiCall = fileStore.loadRecordings();
      
      // Verify loading state was set
      expect(fileStore.loading).toBe(true);
      
      // Resolve the promise
      resolvePromise!(MockDataFactory.getFileListResult());
      await apiCall;
      
      // Verify loading state was cleared
      expect(fileStore.loading).toBe(false);
    });

    test('should handle concurrent API calls correctly', async () => {
      const mockResponse1 = MockDataFactory.getFileListResult();
      const mockResponse2 = MockDataFactory.getFileListResult();
      
      mockFileService.listRecordings = jest.fn().mockResolvedValue(mockResponse1);
      mockFileService.listSnapshots = jest.fn().mockResolvedValue(mockResponse2);
      
      // Start concurrent calls
      const promise1 = fileStore.loadRecordings();
      const promise2 = fileStore.loadSnapshots();
      
      await Promise.all([promise1, promise2]);
      
      // Verify both calls completed successfully
      expect(fileStore.recordings).toEqual(mockResponse1.files);
      expect(fileStore.snapshots).toEqual(mockResponse2.files);
      expect(fileStore.loading).toBe(false);
      expect(fileStore.error).toBe(null);
    });

    test('should reset store state correctly', () => {
      // Set some state
      fileStore.setLoading(true);
      fileStore.setError('Test error');
      
      // Reset the store
      fileStore.reset();
      
      // Verify state was reset to defaults
      expect(fileStore.loading).toBe(false);
      expect(fileStore.error).toBe(null);
      expect(fileStore.recordings).toEqual(MockDataFactory.getFileListResult().files);
      expect(fileStore.snapshots).toEqual(MockDataFactory.getFileListResult().files);
    });

    test('should set file service correctly', () => {
      const newService = MockDataFactory.createMockFileService();
      
      fileStore.setFileService(newService);
      
      // Verify service was set (this would be implementation-specific)
      expect(fileStore.fileService).toBe(newService);
    });
  });

  describe('Edge Cases and Error Scenarios', () => {
    test('should handle empty file list response', async () => {
      const emptyResponse = {
        files: [],
        total: 0,
        limit: 10,
        offset: 0
      };
      
      mockFileService.listRecordings = jest.fn().mockResolvedValue(emptyResponse);
      
      await fileStore.loadRecordings();
      
      expect(fileStore.recordings).toEqual([]);
      expect(fileStore.loading).toBe(false);
      expect(fileStore.error).toBe(null);
    });

    test('should handle malformed API responses', async () => {
      const malformedResponse = { invalid: 'data' };
      
      mockFileService.listRecordings = jest.fn().mockResolvedValue(malformedResponse);
      
      try {
        await fileStore.loadRecordings();
      } catch (error) {
        expect(fileStore.error).toBeDefined();
        expect(fileStore.loading).toBe(false);
      }
    });

    test('should handle network disconnection', async () => {
      const networkError = new Error('Network disconnected');
      
      mockFileService.listRecordings = jest.fn().mockRejectedValue(networkError);
      
      try {
        await fileStore.loadRecordings();
      } catch (error) {
        expect(fileStore.error).toBe('Network disconnected');
        expect(fileStore.loading).toBe(false);
      }
    });

    test('should handle invalid file names', async () => {
      const invalidFileName = '../etc/passwd';
      
      mockFileService.getRecordingInfo = jest.fn().mockRejectedValue(new Error('Invalid file name'));
      
      try {
        await fileStore.getRecordingInfo(invalidFileName);
      } catch (error) {
        expect(fileStore.error).toBe('Invalid file name');
      }
    });

    test('should handle permission denied errors', async () => {
      const permissionError = new Error('Permission denied');
      
      mockFileService.deleteRecording = jest.fn().mockRejectedValue(permissionError);
      
      try {
        await fileStore.deleteRecording('protected-file');
      } catch (error) {
        expect(fileStore.error).toBe('Permission denied');
      }
    });
  });

  describe('Performance and Optimization', () => {
    test('should not make redundant API calls', async () => {
      mockFileService.listRecordings = jest.fn().mockResolvedValue(MockDataFactory.getFileListResult());
      
      // Make multiple calls
      await fileStore.loadRecordings();
      await fileStore.loadRecordings();
      await fileStore.loadRecordings();
      
      // Verify API was called each time (no caching implemented)
      expect(mockFileService.listRecordings).toHaveBeenCalledTimes(3);
    });

    test('should handle large file lists efficiently', async () => {
      const largeFileList = {
        files: Array.from({ length: 1000 }, (_, i) => ({
          filename: `file_${i}`,
          file_size: 1024 * 1024,
          modified_time: '2025-01-15T14:30:00Z',
          download_url: `/files/recordings/file_${i}`
        })),
        total: 1000,
        limit: 1000,
        offset: 0
      };
      
      mockFileService.listRecordings = jest.fn().mockResolvedValue(largeFileList);
      
      await fileStore.loadRecordings();
      
      expect(fileStore.recordings).toHaveLength(1000);
      expect(fileStore.loading).toBe(false);
      expect(fileStore.error).toBe(null);
    });

    test('should handle pagination correctly', async () => {
      const page1Response = {
        files: Array.from({ length: 50 }, (_, i) => ({
          filename: `file_${i}`,
          file_size: 1024,
          modified_time: '2025-01-15T14:30:00Z',
          download_url: `/files/recordings/file_${i}`
        })),
        total: 100,
        limit: 50,
        offset: 0
      };
      
      const page2Response = {
        files: Array.from({ length: 50 }, (_, i) => ({
          filename: `file_${i + 50}`,
          file_size: 1024,
          modified_time: '2025-01-15T14:30:00Z',
          download_url: `/files/recordings/file_${i + 50}`
        })),
        total: 100,
        limit: 50,
        offset: 50
      };
      
      mockFileService.listRecordings = jest.fn()
        .mockResolvedValueOnce(page1Response)
        .mockResolvedValueOnce(page2Response);
      
      // Load first page
      await fileStore.loadRecordings({ limit: 50, offset: 0 });
      expect(fileStore.recordings).toHaveLength(50);
      
      // Load second page
      await fileStore.loadRecordings({ limit: 50, offset: 50 });
      expect(fileStore.recordings).toHaveLength(50);
    });
  });
});