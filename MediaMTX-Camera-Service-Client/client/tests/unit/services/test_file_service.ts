/**
 * FileService unit tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * 
 * Requirements Coverage:
 * - REQ-FILE-001: File listing with pagination
 * - REQ-FILE-002: File information retrieval
 * - REQ-FILE-003: File deletion operations
 * - REQ-FILE-004: File download functionality
 * - REQ-FILE-005: Error handling and validation
 * 
 * Test Categories: Unit
 * API Documentation Reference: ../mediamtx-camera-service-go/docs/api/json_rpc_methods.md
 */

import { FileService } from '../../../src/services/file/FileService';
import { WebSocketService } from '../../../src/services/websocket/WebSocketService';
import { LoggerService } from '../../../src/services/logger/LoggerService';
import { APIMocks } from '../../utils/mocks';
import { APIResponseValidator } from '../../utils/validators';

// Mock dependencies
const mockWebSocketService = {
  sendRPC: jest.fn(),
} as jest.Mocked<WebSocketService>;

const mockLoggerService = {
  info: jest.fn(),
  warn: jest.fn(),
  error: jest.fn(),
} as jest.Mocked<LoggerService>;

// Mock DOM methods
const mockCreateElement = jest.fn();
const mockAppendChild = jest.fn();
const mockRemoveChild = jest.fn();
const mockClick = jest.fn();

// Mock document for jsdom environment
if (typeof document === 'undefined') {
  (global as any).document = {
    createElement: mockCreateElement,
    body: {
      appendChild: mockAppendChild,
      removeChild: mockRemoveChild,
    },
  };
} else {
  Object.defineProperty(document, 'createElement', {
    value: mockCreateElement,
    writable: true,
  });
  Object.defineProperty(document.body, 'appendChild', {
    value: mockAppendChild,
    writable: true,
  });
  Object.defineProperty(document.body, 'removeChild', {
    value: mockRemoveChild,
    writable: true,
  });
}

describe('FileService Unit Tests', () => {
  let fileService: FileService;

  beforeEach(() => {
    jest.clearAllMocks();
    fileService = new FileService(mockWebSocketService, mockLoggerService);
  });

  describe('REQ-FILE-001: File listing with pagination', () => {
    test('should list recordings with pagination', async () => {
      const limit = 10;
      const offset = 0;
      const expectedResult = APIMocks.getListFilesResult('recordings');
      
      mockWebSocketService.sendRPC.mockResolvedValue(expectedResult);

      const result = await fileService.listRecordings(limit, offset);

      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('list_recordings', { limit, offset });
      expect(mockLoggerService.info).toHaveBeenCalledWith(`Listing recordings: limit=${limit}, offset=${offset}`);
      expect(mockLoggerService.info).toHaveBeenCalledWith(`Found ${expectedResult.files.length} recordings`);
      expect(result).toEqual(expectedResult);
      expect(APIResponseValidator.validateListFilesResult(result)).toBe(true);
    });

    test('should list snapshots with pagination', async () => {
      const limit = 20;
      const offset = 10;
      const expectedResult = APIMocks.getListFilesResult('snapshots');
      
      mockWebSocketService.sendRPC.mockResolvedValue(expectedResult);

      const result = await fileService.listSnapshots(limit, offset);

      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('list_snapshots', { limit, offset });
      expect(mockLoggerService.info).toHaveBeenCalledWith(`Listing snapshots: limit=${limit}, offset=${offset}`);
      expect(mockLoggerService.info).toHaveBeenCalledWith(`Found ${expectedResult.files.length} snapshots`);
      expect(result).toEqual(expectedResult);
    });

    test('should handle empty file lists', async () => {
      const emptyResult = { files: [], total: 0, limit: 10, offset: 0 };
      mockWebSocketService.sendRPC.mockResolvedValue(emptyResult);

      const result = await fileService.listRecordings(10, 0);

      expect(result.files).toEqual([]);
      expect(mockLoggerService.info).toHaveBeenCalledWith('Found 0 recordings');
    });

    test('should handle listing errors', async () => {
      const error = new Error('Failed to list files');
      mockWebSocketService.sendRPC.mockRejectedValue(error);

      await expect(fileService.listRecordings(10, 0)).rejects.toThrow('Failed to list files');
      expect(mockLoggerService.error).toHaveBeenCalledWith('Failed to list recordings', error);
    });

    test('should validate pagination parameters', async () => {
      const limit = 10;
      const offset = 0;
      const result = APIMocks.getListFilesResult('recordings');
      
      mockWebSocketService.sendRPC.mockResolvedValue(result);

      await fileService.listRecordings(limit, offset);

      expect(APIResponseValidator.validatePaginationParams(limit, offset)).toBe(true);
    });
  });

  describe('REQ-FILE-002: File information retrieval', () => {
    test('should get recording info', async () => {
      const filename = 'recording_camera0_1234567890.mp4';
      const expectedInfo = {
        filename,
        file_size: 1024000,
        modified_time: new Date().toISOString(),
        download_url: `https://localhost/downloads/${filename}`,
        duration: 60,
        format: 'mp4',
        device: 'camera0'
      };
      
      mockWebSocketService.sendRPC.mockResolvedValue(expectedInfo);

      const result = await fileService.getRecordingInfo(filename);

      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('get_recording_info', { filename });
      expect(mockLoggerService.info).toHaveBeenCalledWith(`Getting recording info for: ${filename}`);
      expect(mockLoggerService.info).toHaveBeenCalledWith(`Recording info retrieved for ${filename}`);
      expect(result).toEqual(expectedInfo);
    });

    test('should get snapshot info', async () => {
      const filename = 'snapshot_camera0_1234567890.jpg';
      const expectedInfo = {
        filename,
        file_size: 512000,
        modified_time: new Date().toISOString(),
        download_url: `https://localhost/downloads/${filename}`,
        format: 'jpg',
        device: 'camera0'
      };
      
      mockWebSocketService.sendRPC.mockResolvedValue(expectedInfo);

      const result = await fileService.getSnapshotInfo(filename);

      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('get_snapshot_info', { filename });
      expect(mockLoggerService.info).toHaveBeenCalledWith(`Getting snapshot info for: ${filename}`);
      expect(mockLoggerService.info).toHaveBeenCalledWith(`Snapshot info retrieved for ${filename}`);
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

    test('should validate file information structure', async () => {
      const filename = 'test.mp4';
      const fileInfo = {
        filename,
        file_size: 1024000,
        modified_time: new Date().toISOString(),
        download_url: `https://localhost/downloads/${filename}`
      };
      
      mockWebSocketService.sendRPC.mockResolvedValue(fileInfo);

      const result = await fileService.getRecordingInfo(filename);

      expect(APIResponseValidator.validateRecordingFile(result)).toBe(true);
    });
  });

  describe('REQ-FILE-003: File deletion operations', () => {
    test('should delete recording successfully', async () => {
      const filename = 'recording_camera0_1234567890.mp4';
      const expectedResult = { success: true, message: 'Recording deleted successfully' };
      
      mockWebSocketService.sendRPC.mockResolvedValue(expectedResult);

      const result = await fileService.deleteRecording(filename);

      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('delete_recording', { filename });
      expect(mockLoggerService.info).toHaveBeenCalledWith(`Deleting recording: ${filename}`);
      expect(mockLoggerService.info).toHaveBeenCalledWith(`Recording deleted: ${filename}`);
      expect(result).toEqual(expectedResult);
    });

    test('should delete snapshot successfully', async () => {
      const filename = 'snapshot_camera0_1234567890.jpg';
      const expectedResult = { success: true, message: 'Snapshot deleted successfully' };
      
      mockWebSocketService.sendRPC.mockResolvedValue(expectedResult);

      const result = await fileService.deleteSnapshot(filename);

      expect(mockWebSocketService.sendRPC).toHaveBeenCalledWith('delete_snapshot', { filename });
      expect(mockLoggerService.info).toHaveBeenCalledWith(`Deleting snapshot: ${filename}`);
      expect(mockLoggerService.info).toHaveBeenCalledWith(`Snapshot deleted: ${filename}`);
      expect(result).toEqual(expectedResult);
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
      const expectedResult = { success: false, message: 'File not found' };
      
      mockWebSocketService.sendRPC.mockResolvedValue(expectedResult);

      const result = await fileService.deleteRecording(filename);

      expect(result.success).toBe(false);
      expect(result.message).toBe('File not found');
    });
  });

  describe('REQ-FILE-004: File download functionality', () => {
    test('should download file via server URL', async () => {
      const downloadUrl = 'https://localhost/downloads/recording.mp4';
      const filename = 'recording.mp4';
      
      const mockLink = {
        href: '',
        download: '',
        target: '',
        click: mockClick
      };
      
      mockCreateElement.mockReturnValue(mockLink);

      await fileService.downloadFile(downloadUrl, filename);

      expect(mockCreateElement).toHaveBeenCalledWith('a');
      expect(mockLink.href).toBe(downloadUrl);
      expect(mockLink.download).toBe(filename);
      expect(mockLink.target).toBe('_blank');
      expect(mockAppendChild).toHaveBeenCalledWith(mockLink);
      expect(mockClick).toHaveBeenCalled();
      expect(mockRemoveChild).toHaveBeenCalledWith(mockLink);
      expect(mockLoggerService.info).toHaveBeenCalledWith(`Downloading file: ${filename}`);
      expect(mockLoggerService.info).toHaveBeenCalledWith(`Download initiated for: ${filename}`);
    });

    test('should handle download errors', async () => {
      const downloadUrl = 'https://localhost/downloads/recording.mp4';
      const filename = 'recording.mp4';
      
      mockCreateElement.mockImplementation(() => {
        throw new Error('DOM manipulation failed');
      });

      await expect(fileService.downloadFile(downloadUrl, filename)).rejects.toThrow('DOM manipulation failed');
      expect(mockLoggerService.error).toHaveBeenCalledWith(
        `Failed to download file ${filename}`,
        expect.any(Error)
      );
    });

    test('should validate download URL format', async () => {
      const downloadUrl = 'https://localhost/downloads/recording.mp4';
      const filename = 'recording.mp4';
      
      const mockLink = {
        href: '',
        download: '',
        target: '',
        click: mockClick
      };
      
      mockCreateElement.mockReturnValue(mockLink);

      await fileService.downloadFile(downloadUrl, filename);

      expect(APIResponseValidator.validateStreamUrl(downloadUrl)).toBe(true);
    });
  });

  describe('REQ-FILE-005: Error handling and validation', () => {
    test('should handle WebSocket service errors', async () => {
      const error = new Error('WebSocket connection lost');
      mockWebSocketService.sendRPC.mockRejectedValue(error);

      await expect(fileService.listRecordings(10, 0)).rejects.toThrow('WebSocket connection lost');
      expect(mockLoggerService.error).toHaveBeenCalledWith('Failed to list recordings', error);
    });

    test('should log all operations with appropriate levels', async () => {
      const result = APIMocks.getListFilesResult('recordings');
      mockWebSocketService.sendRPC.mockResolvedValue(result);

      await fileService.listRecordings(10, 0);

      expect(mockLoggerService.info).toHaveBeenCalledWith('Listing recordings: limit=10, offset=0');
      expect(mockLoggerService.info).toHaveBeenCalledWith(`Found ${result.files.length} recordings`);
    });

    test('should handle invalid file names', async () => {
      const invalidFilename = '';
      const error = new Error('Invalid filename');
      mockWebSocketService.sendRPC.mockRejectedValue(error);

      await expect(fileService.getRecordingInfo(invalidFilename)).rejects.toThrow('Invalid filename');
      expect(mockLoggerService.error).toHaveBeenCalledWith(
        `Failed to get recording info for ${invalidFilename}`,
        error
      );
    });

    test('should handle network timeouts', async () => {
      const error = new Error('Request timeout');
      mockWebSocketService.sendRPC.mockRejectedValue(error);

      await expect(fileService.deleteRecording('test.mp4')).rejects.toThrow('Request timeout');
      expect(mockLoggerService.error).toHaveBeenCalledWith(
        'Failed to delete recording test.mp4',
        error
      );
    });
  });

  describe('File format validation', () => {
    test('should handle different recording formats', async () => {
      const formats = ['mp4', 'fmp4', 'mkv'];
      
      for (const format of formats) {
        const filename = `recording.${format}`;
        const fileInfo = {
          filename,
          file_size: 1024000,
          modified_time: new Date().toISOString(),
          download_url: `https://localhost/downloads/${filename}`,
          format
        };
        
        mockWebSocketService.sendRPC.mockResolvedValue(fileInfo);

        const result = await fileService.getRecordingInfo(filename);

        expect(result.format).toBe(format);
        expect(APIResponseValidator.validateRecordingFormat(format)).toBe(true);
      }
    });

    test('should handle different snapshot formats', async () => {
      const formats = ['jpg', 'png'];
      
      for (const format of formats) {
        const filename = `snapshot.${format}`;
        const fileInfo = {
          filename,
          file_size: 512000,
          modified_time: new Date().toISOString(),
          download_url: `https://localhost/downloads/${filename}`,
          format
        };
        
        mockWebSocketService.sendRPC.mockResolvedValue(fileInfo);

        const result = await fileService.getSnapshotInfo(filename);

        expect(result.format).toBe(format);
      }
    });
  });
});
