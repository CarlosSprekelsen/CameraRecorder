/**
 * Unit Tests for File Store
 * 
 * REQ-001: Store State Management - Test Zustand store actions
 * REQ-002: State Transitions - Test state changes
 * REQ-003: Error Handling - Test error states and recovery
 * REQ-004: API Integration - Mock API calls and test responses
 * REQ-005: Side Effects - Test store side effects
 * REQ-006: Pagination Logic - Test pagination state management
 * REQ-007: File Selection - Test file selection functionality
 * 
 * Ground Truth: Official RPC Documentation
 * API Reference: docs/api/json_rpc_methods.md
 */

import { describe, test, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { MockDataFactory } from '../../utils/mocks';
import { APIResponseValidator } from '../../utils/validators';
import { useFileStore } from '../../../src/stores/file/fileStore';

// Mock the FileService
const mockFileService = {
  listRecordings: jest.fn(),
  listSnapshots: jest.fn(),
  getRecordingInfo: jest.fn(),
  getSnapshotInfo: jest.fn(),
  deleteRecording: jest.fn(),
  deleteSnapshot: jest.fn()
};

jest.mock('../../../src/services/file/FileService', () => ({
  FileService: jest.fn().mockImplementation(() => mockFileService)
}));

describe('File Store', () => {
  beforeEach(() => {
    // Reset the store to initial state
    useFileStore.getState().reset();
    
    // Reset all mocks
    jest.clearAllMocks();
    
    // Set up default mock implementations
    (mockFileService.listRecordings as jest.Mock).mockResolvedValue(MockDataFactory.getFileListResult());
    (mockFileService.listSnapshots as jest.Mock).mockResolvedValue(MockDataFactory.getFileListResult());
    (mockFileService.getRecordingInfo as jest.Mock).mockResolvedValue(MockDataFactory.getFileListResult().files[0]);
    (mockFileService.getSnapshotInfo as jest.Mock).mockResolvedValue(MockDataFactory.getFileListResult().files[0]);
    (mockFileService.deleteRecording as jest.Mock).mockResolvedValue(true);
    (mockFileService.deleteSnapshot as jest.Mock).mockResolvedValue(true);
  });

  afterEach(() => {
    useFileStore.getState().reset();
  });

  describe('REQ-001: Store State Management', () => {
    test('should initialize with correct initial state', () => {
      const state = useFileStore.getState();
      
      expect(state.recordings).toEqual([]);
      expect(state.snapshots).toEqual([]);
      expect(state.loading).toBe(false);
      expect(state.error).toBe(null);
      expect(state.pagination).toEqual({
        limit: 50,
        offset: 0,
        total: 0
      });
      expect(state.selectedFiles).toEqual([]);
      expect(state.currentTab).toBe('recordings');
    });

    test('should set loading state correctly', () => {
      const { setLoading } = useFileStore.getState();
      
      setLoading(true);
      expect(useFileStore.getState().loading).toBe(true);
      
      setLoading(false);
      expect(useFileStore.getState().loading).toBe(false);
    });

    test('should set error state correctly', () => {
      const { setError } = useFileStore.getState();
      const errorMessage = 'Test error';
      
      setError(errorMessage);
      expect(useFileStore.getState().error).toBe(errorMessage);
      
      setError(null);
      expect(useFileStore.getState().error).toBe(null);
    });

    test('should set current tab correctly', () => {
      const { setCurrentTab } = useFileStore.getState();
      
      setCurrentTab('snapshots');
      expect(useFileStore.getState().currentTab).toBe('snapshots');
      
      setCurrentTab('recordings');
      expect(useFileStore.getState().currentTab).toBe('recordings');
    });
  });

  describe('REQ-006: Pagination Logic', () => {
    test('should set pagination correctly', () => {
      const { setPagination } = useFileStore.getState();
      
      setPagination(25, 50, 100);
      const pagination = useFileStore.getState().pagination;
      
      expect(pagination.limit).toBe(25);
      expect(pagination.offset).toBe(50);
      expect(pagination.total).toBe(100);
    });

    test('should update pagination when loading files', async () => {
      const { loadRecordings, setFileService } = useFileStore.getState();
      
      setFileService(mockFileService as any);
      await loadRecordings(25, 50);
      
      const pagination = useFileStore.getState().pagination;
      expect(pagination.limit).toBe(25);
      expect(pagination.offset).toBe(50);
    });
  });

  describe('REQ-007: File Selection', () => {
    test('should set selected files correctly', () => {
      const { setSelectedFiles } = useFileStore.getState();
      const files = ['file1.mp4', 'file2.mp4'];
      
      setSelectedFiles(files);
      expect(useFileStore.getState().selectedFiles).toEqual(files);
    });

    test('should toggle file selection correctly', () => {
      const { toggleFileSelection } = useFileStore.getState();
      
      // Add first file
      toggleFileSelection('file1.mp4');
      expect(useFileStore.getState().selectedFiles).toContain('file1.mp4');
      
      // Add second file
      toggleFileSelection('file2.mp4');
      expect(useFileStore.getState().selectedFiles).toEqual(['file1.mp4', 'file2.mp4']);
      
      // Remove first file
      toggleFileSelection('file1.mp4');
      expect(useFileStore.getState().selectedFiles).toEqual(['file2.mp4']);
    });

    test('should clear selection correctly', () => {
      const { setSelectedFiles, clearSelection } = useFileStore.getState();
      
      setSelectedFiles(['file1.mp4', 'file2.mp4']);
      clearSelection();
      
      expect(useFileStore.getState().selectedFiles).toEqual([]);
    });
  });

  describe('REQ-002: State Transitions', () => {
    test('should transition from loading to success state for recordings', async () => {
      const { loadRecordings, setFileService } = useFileStore.getState();
      
      setFileService(mockFileService as any);
      
      const promise = loadRecordings();
      
      // Check loading state
      expect(useFileStore.getState().loading).toBe(true);
      expect(useFileStore.getState().error).toBe(null);
      
      await promise;
      
      // Check final state
      const state = useFileStore.getState();
      expect(state.loading).toBe(false);
      expect(state.error).toBe(null);
      expect(state.recordings).toEqual(MockDataFactory.getFileListResult().files);
    });

    test('should transition from loading to error state', async () => {
      const { loadRecordings, setFileService } = useFileStore.getState();
      
      (mockFileService.listRecordings as jest.Mock).mockRejectedValue(new Error('Network error'));
      setFileService(mockFileService as any);
      
      await loadRecordings();
      
      const state = useFileStore.getState();
      expect(state.loading).toBe(false);
      expect(state.error).toBe('Network error');
    });
  });

  describe('REQ-003: Error Handling', () => {
    test('should handle service not initialized error', async () => {
      const { loadRecordings } = useFileStore.getState();
      
      // Don't set the service
      await loadRecordings();
      
      const state = useFileStore.getState();
      expect(state.error).toBe('File service not initialized');
      expect(state.loading).toBe(false);
    });

    test('should handle API errors gracefully', async () => {
      const { loadRecordings, setFileService } = useFileStore.getState();
      
      (mockFileService.listRecordings as jest.Mock).mockRejectedValue(new Error('API Error'));
      setFileService(mockFileService as any);
      
      await loadRecordings();
      
      const state = useFileStore.getState();
      expect(state.error).toBe('API Error');
      expect(state.loading).toBe(false);
    });
  });

  describe('REQ-004: API Integration', () => {
    test('should call listRecordings and update state', async () => {
      const { loadRecordings, setFileService } = useFileStore.getState();
      
      setFileService(mockFileService as any);
      await loadRecordings();
      
      expect(mockFileService.listRecordings).toHaveBeenCalledTimes(1);
      
      const state = useFileStore.getState();
      expect(state.recordings).toEqual(MockDataFactory.getFileListResult().files);
    });

    test('should call listSnapshots and update state', async () => {
      const { loadSnapshots, setFileService } = useFileStore.getState();
      
      setFileService(mockFileService as any);
      await loadSnapshots();
      
      expect(mockFileService.listSnapshots).toHaveBeenCalledTimes(1);
      
      const state = useFileStore.getState();
      expect(state.snapshots).toEqual(MockDataFactory.getFileListResult().files);
    });

    test('should call deleteRecording and update state', async () => {
      const { deleteRecording, setFileService } = useFileStore.getState();
      
      // Set up some recordings first
      useFileStore.setState({
        recordings: MockDataFactory.getFileListResult().files
      });
      
      setFileService(mockFileService as any);
      const result = await deleteRecording('test.mp4');
      
      expect(mockFileService.deleteRecording).toHaveBeenCalledWith('test.mp4');
      expect(result).toBe(true);
    });

    test('should call deleteSnapshot and update state', async () => {
      const { deleteSnapshot, setFileService } = useFileStore.getState();
      
      // Set up some snapshots first
      useFileStore.setState({
        snapshots: MockDataFactory.getFileListResult().files
      });
      
      setFileService(mockFileService as any);
      const result = await deleteSnapshot('test.jpg');
      
      expect(mockFileService.deleteSnapshot).toHaveBeenCalledWith('test.jpg');
      expect(result).toBe(true);
    });
  });

  describe('REQ-005: Side Effects', () => {
    test('should reset store to initial state', () => {
      const { reset, setLoading, setError } = useFileStore.getState();
      
      // Modify state
      setLoading(true);
      setError('Test error');
      useFileStore.setState({
        recordings: MockDataFactory.getFileListResult().files,
        selectedFiles: ['file1.mp4'],
        currentTab: 'snapshots'
      });
      
      // Reset
      reset();
      
      // Check state is back to initial
      const state = useFileStore.getState();
      expect(state.recordings).toEqual([]);
      expect(state.snapshots).toEqual([]);
      expect(state.loading).toBe(false);
      expect(state.error).toBe(null);
      expect(state.selectedFiles).toEqual([]);
      expect(state.currentTab).toBe('recordings');
      expect(state.pagination).toEqual({
        limit: 50,
        offset: 0,
        total: 0
      });
    });

    test('should handle file deletion side effects', async () => {
      const { deleteRecording, setFileService } = useFileStore.getState();
      
      const files = MockDataFactory.getFileListResult().files;
      useFileStore.setState({ recordings: files });
      
      setFileService(mockFileService as any);
      await deleteRecording('test.mp4');
      
      // The store should remove the deleted file from the list
      const remainingFiles = useFileStore.getState().recordings;
      expect(remainingFiles).not.toContainEqual(
        expect.objectContaining({ filename: 'test.mp4' })
      );
    });
  });

  describe('API Compliance Validation', () => {
    test('should validate file list response against RPC spec', async () => {
      const { loadRecordings, setFileService } = useFileStore.getState();
      
      setFileService(mockFileService as any);
      await loadRecordings();
      
      const recordings = useFileStore.getState().recordings;
      expect(recordings.length).toBeGreaterThan(0);
      
      // Validate each file against RPC spec
      recordings.forEach(file => {
        expect(APIResponseValidator.validateFileListResult({ files: [file] })).toBe(true);
      });
    });

    test('should validate snapshot list response against RPC spec', async () => {
      const { loadSnapshots, setFileService } = useFileStore.getState();
      
      setFileService(mockFileService as any);
      await loadSnapshots();
      
      const snapshots = useFileStore.getState().snapshots;
      expect(snapshots.length).toBeGreaterThan(0);
      
      // Validate each snapshot against RPC spec
      snapshots.forEach(file => {
        expect(APIResponseValidator.validateFileListResult({ files: [file] })).toBe(true);
      });
    });
  });
});