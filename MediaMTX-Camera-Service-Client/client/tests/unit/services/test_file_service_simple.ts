/**
 * Simplified FileService unit tests - avoiding complex document mocking
 * 
 * Focus: Core functionality without DOM operations
 * Coverage Target: FileService methods that don't require document manipulation
 */

import { FileService } from '../../../src/services/file/FileService';
import { WebSocketService } from '../../../src/services/websocket/WebSocketService';
import { LoggerService } from '../../../src/services/logger/LoggerService';
import { MockDataFactory } from '../../utils/mocks';
import { APIResponseValidator } from '../../utils/validators';

// Use centralized mocks - eliminates duplication
const mockWebSocketService = MockDataFactory.createMockWebSocketService();
const mockLoggerService = MockDataFactory.createMockLoggerService();

describe('FileService Simplified Tests (No DOM)', () => {
  let fileService: FileService;

  beforeEach(() => {
    jest.clearAllMocks();
    fileService = new FileService(mockWebSocketService, mockLoggerService);
  });

  describe('REQ-FILE-001: File listing with pagination', () => {
    test('should list recordings with pagination', async () => {
      const limit = 10;
      const offset = 0;
      const expectedResult = MockDataFactory.getFileListResult();
      
      mockWebSocketService.sendRPC.mockResolvedValue(expectedResult);

      const result = await fileService.listRecordings(limit, offset);

      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('list_recordings', {
        limit,
        offset
      });
      expect(result).toEqual(expectedResult);
      expect(APIResponseValidator.validateFileListResult(result)).toBe(true);
    });

    test('should list snapshots with pagination', async () => {
      const limit = 5;
      const offset = 10;
      const expectedResult = MockDataFactory.getFileListResult();
      
      mockWebSocketService.sendRPC.mockResolvedValue(expectedResult);

      const result = await fileService.listSnapshots(limit, offset);

      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('list_snapshots', {
        limit,
        offset
      });
      expect(result).toEqual(expectedResult);
    });

    test('should handle empty file lists', async () => {
      const emptyResult = { files: [], total: 0, limit: 10, offset: 0 };
      mockWebSocketService.sendRPC.mockResolvedValue(emptyResult);

      const result = await fileService.listRecordings(10, 0);

      expect(result.files).toEqual([]);
      expect(result.total).toBe(0);
    });

    test('should handle listing errors', async () => {
      const error = new Error('Failed to list files');
      mockWebSocketService.sendRPC.mockRejectedValue(error);

      await expect(fileService.listRecordings(10, 0)).rejects.toThrow('Failed to list files');
      expect(mockLoggerService.error).toHaveBeenCalledWith(
        'Failed to list recordings',
        error
      );
    });
  });

  describe('REQ-FILE-002: File information retrieval', () => {
    test('should get recording info', async () => {
      const filename = 'recording.mp4';
      const expectedInfo = MockDataFactory.getRecordingInfo();
      mockWebSocketService.sendRPC.mockResolvedValue(expectedInfo);

      const result = await fileService.getRecordingInfo(filename);

      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('get_recording_info', { filename });
      expect(result).toEqual(expectedInfo);
      expect(APIResponseValidator.validateRecordingFile(result)).toBe(true);
    });

    test('should get snapshot info', async () => {
      const filename = 'snapshot.jpg';
      const expectedInfo = MockDataFactory.getSnapshotInfo();
      mockWebSocketService.sendRPC.mockResolvedValue(expectedInfo);

      const result = await fileService.getSnapshotInfo(filename);

      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('get_snapshot_info', { filename });
      expect(result).toEqual(expectedInfo);
    });

    test('should handle file info errors', async () => {
      const filename = 'nonexistent.mp4';
      const error = new Error('File not found');
      mockWebSocketService.sendRPC.mockRejectedValue(error);

      await expect(fileService.getRecordingInfo(filename)).rejects.toThrow('File not found');
      expect(mockLoggerService.error).toHaveBeenCalledWith(
        `Failed to get recording info for ${filename}`,
        error
      );
    });
  });

  describe('REQ-FILE-003: File deletion operations', () => {
    test('should delete recording successfully', async () => {
      const filename = 'recording.mp4';
      const expectedResult = MockDataFactory.getDeleteResult();
      mockWebSocketService.sendRPC.mockResolvedValue(expectedResult);

      const result = await fileService.deleteRecording(filename);

      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('delete_recording', { filename });
      expect(result).toEqual(expectedResult);
      expect(mockLoggerService.info).toHaveBeenCalledWith(`Recording deleted: ${filename}`);
    });

    test('should delete snapshot successfully', async () => {
      const filename = 'snapshot.jpg';
      const expectedResult = MockDataFactory.getDeleteResult();
      mockWebSocketService.sendRPC.mockResolvedValue(expectedResult);

      const result = await fileService.deleteSnapshot(filename);

      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('delete_snapshot', { filename });
      expect(result).toEqual(expectedResult);
      expect(mockLoggerService.info).toHaveBeenCalledWith(`Snapshot deleted: ${filename}`);
    });

    test('should handle deletion errors', async () => {
      const filename = 'protected.mp4';
      const error = new Error('Permission denied');
      mockWebSocketService.sendRPC.mockRejectedValue(error);

      await expect(fileService.deleteRecording(filename)).rejects.toThrow('Permission denied');
      expect(mockLoggerService.error).toHaveBeenCalledWith(
        `Failed to delete recording ${filename}`,
        error
      );
    });

    test('should handle deletion failures', async () => {
      const filename = 'nonexistent.mp4';
      const failedResult = { deleted: false, message: 'File not found' };
      mockWebSocketService.sendRPC.mockResolvedValue(failedResult);

      const result = await fileService.deleteRecording(filename);

      expect(result.deleted).toBe(false);
      expect(result.message).toBe('File not found');
    });
  });

  describe('REQ-FILE-005: Error handling and validation', () => {
    test('should handle WebSocket service errors', async () => {
      const error = new Error('WebSocket connection lost');
      mockWebSocketService.sendRPC.mockRejectedValue(error);

      await expect(fileService.listRecordings(10, 0)).rejects.toThrow('WebSocket connection lost');
    });

    test('should log all operations with appropriate levels', async () => {
      const expectedResult = MockDataFactory.getFileListResult();
      mockWebSocketService.sendRPC.mockResolvedValue(expectedResult);

      await fileService.listRecordings(10, 0);

      expect(mockLoggerService.info).toHaveBeenCalledWith('Listing recordings: limit=10, offset=0');
    });

    test('should handle invalid file names', async () => {
      const invalidFilename = '';
      const error = new Error('Invalid filename');
      mockWebSocketService.sendRPC.mockRejectedValue(error);

      await expect(fileService.getRecordingInfo(invalidFilename)).rejects.toThrow('Invalid filename');
    });

    test('should handle network timeouts', async () => {
      const timeoutError = new Error('Request timeout');
      mockWebSocketService.sendRPC.mockRejectedValue(timeoutError);

      await expect(fileService.listRecordings(10, 0)).rejects.toThrow('Request timeout');
    });
  });
});
