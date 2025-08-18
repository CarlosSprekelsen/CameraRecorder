/**
 * Unit tests for file store
 * Tests file operations, WebSocket integration, and error handling
 */

import { renderHook, act } from '@testing-library/react';
import { useFileStore } from '../../../src/stores/fileStore';
import { createWebSocketService } from '../../../src/services/websocket';
import { RPC_METHODS } from '../../../src/types';

// Mock the WebSocket service
jest.mock('../../../src/services/websocket');
const mockCreateWebSocketService = createWebSocketService as jest.MockedFunction<typeof createWebSocketService>;

// Mock WebSocket service instance
const mockWebSocketService = {
  connect: jest.fn(),
  disconnect: jest.fn(),
  call: jest.fn(),
  onConnect: jest.fn(),
  onDisconnect: jest.fn(),
  onError: jest.fn(),
  onMessage: jest.fn()
};

describe('File Store', () => {
  beforeEach(() => {
    mockCreateWebSocketService.mockReturnValue(mockWebSocketService);
    jest.clearAllMocks();
    
    // Reset store state
    const { result } = renderHook(() => useFileStore());
    act(() => {
      result.current.disconnect();
    });
  });

  describe('Initialization', () => {
    it('should initialize with default state', () => {
      const { result } = renderHook(() => useFileStore());

      expect(result.current.recordings).toBeNull();
      expect(result.current.snapshots).toBeNull();
      expect(result.current.isLoading).toBe(false);
      expect(result.current.isDownloading).toBe(false);
      expect(result.current.error).toBeNull();
      expect(result.current.isConnected).toBe(false);
    });

    it('should initialize WebSocket service', async () => {
      mockWebSocketService.connect.mockResolvedValue();

      const { result } = renderHook(() => useFileStore());

      await act(async () => {
        await result.current.initialize();
      });

      expect(mockCreateWebSocketService).toHaveBeenCalledWith({
        url: 'ws://localhost:8002/ws',
        reconnectTimeout: 5000,
        maxReconnectAttempts: 5,
      });
      expect(mockWebSocketService.connect).toHaveBeenCalled();
    });

    it('should set up event handlers during initialization', async () => {
      mockWebSocketService.connect.mockResolvedValue();

      const { result } = renderHook(() => useFileStore());

      await act(async () => {
        await result.current.initialize();
      });

      expect(mockWebSocketService.onConnect).toHaveBeenCalled();
      expect(mockWebSocketService.onDisconnect).toHaveBeenCalled();
      expect(mockWebSocketService.onError).toHaveBeenCalled();
    });

    it('should handle initialization errors', async () => {
      mockWebSocketService.connect.mockRejectedValue(new Error('Connection failed'));

      const { result } = renderHook(() => useFileStore());

      await act(async () => {
        await result.current.initialize();
      });

      expect(result.current.error).toBe('Connection failed');
      expect(result.current.isConnected).toBe(false);
    });

    it('should not initialize twice', async () => {
      mockWebSocketService.connect.mockResolvedValue();

      const { result } = renderHook(() => useFileStore());

      await act(async () => {
        await result.current.initialize();
        await result.current.initialize(); // Second call should be ignored
      });

      expect(mockCreateWebSocketService).toHaveBeenCalledTimes(1);
      expect(mockWebSocketService.connect).toHaveBeenCalledTimes(1);
    });
  });

  describe('Disconnection', () => {
    it('should disconnect WebSocket service', () => {
      const { result } = renderHook(() => useFileStore());

      act(() => {
        result.current.disconnect();
      });

      expect(mockWebSocketService.disconnect).toHaveBeenCalled();
    });

    it('should reset state on disconnect', () => {
      const { result } = renderHook(() => useFileStore());

      act(() => {
        result.current.disconnect();
      });

      expect(result.current.wsService).toBeNull();
      expect(result.current.isConnected).toBe(false);
      expect(result.current.recordings).toBeNull();
      expect(result.current.snapshots).toBeNull();
    });
  });

  describe('Load Recordings', () => {
    beforeEach(async () => {
      mockWebSocketService.connect.mockResolvedValue();
      const { result } = renderHook(() => useFileStore());
      await act(async () => {
        await result.current.initialize();
      });
    });

    it('should load recordings successfully', async () => {
      const mockRecordings = [
        {
          filename: 'recording-1.mp4',
          file_size: 1024000,
          created_at: '2024-01-01T00:00:00Z',
          modified_time: '2024-01-01T00:01:00Z',
          download_url: '/files/recordings/recording-1.mp4',
          duration: 60,
          format: 'mp4'
        }
      ];

      mockWebSocketService.call.mockResolvedValue({
        files: mockRecordings
      });

      const { result } = renderHook(() => useFileStore());

      await act(async () => {
        await result.current.loadRecordings(20, 0);
      });

      expect(mockWebSocketService.call).toHaveBeenCalledWith(RPC_METHODS.LIST_RECORDINGS, {
        limit: 20,
        offset: 0
      });
      expect(result.current.recordings).toEqual(mockRecordings);
      expect(result.current.isLoading).toBe(false);
      expect(result.current.error).toBeNull();
    });

    it('should handle empty recordings response', async () => {
      mockWebSocketService.call.mockResolvedValue({
        files: []
      });

      const { result } = renderHook(() => useFileStore());

      await act(async () => {
        await result.current.loadRecordings();
      });

      expect(result.current.recordings).toEqual([]);
    });

    it('should handle recordings loading error', async () => {
      mockWebSocketService.call.mockRejectedValue(new Error('Failed to load recordings'));

      const { result } = renderHook(() => useFileStore());

      await act(async () => {
        await result.current.loadRecordings();
      });

      expect(result.current.error).toBe('Failed to load recordings');
      expect(result.current.recordings).toBeNull();
      expect(result.current.isLoading).toBe(false);
    });

    it('should set loading state during operation', async () => {
      mockWebSocketService.call.mockImplementation(() => new Promise(resolve => setTimeout(resolve, 100)));

      const { result } = renderHook(() => useFileStore());

      const loadPromise = act(async () => {
        await result.current.loadRecordings();
      });

      expect(result.current.isLoading).toBe(true);

      await loadPromise;
      expect(result.current.isLoading).toBe(false);
    });

    it('should handle WebSocket not connected error', async () => {
      const { result } = renderHook(() => useFileStore());

      await act(async () => {
        await result.current.loadRecordings();
      });

      expect(result.current.error).toBe('WebSocket not connected');
    });
  });

  describe('Load Snapshots', () => {
    beforeEach(async () => {
      mockWebSocketService.connect.mockResolvedValue();
      const { result } = renderHook(() => useFileStore());
      await act(async () => {
        await result.current.initialize();
      });
    });

    it('should load snapshots successfully', async () => {
      const mockSnapshots = [
        {
          filename: 'snapshot-1.jpg',
          file_size: 512000,
          created_at: '2024-01-01T00:00:00Z',
          modified_time: '2024-01-01T00:00:00Z',
          download_url: '/files/snapshots/snapshot-1.jpg',
          format: 'jpg'
        }
      ];

      mockWebSocketService.call.mockResolvedValue({
        files: mockSnapshots
      });

      const { result } = renderHook(() => useFileStore());

      await act(async () => {
        await result.current.loadSnapshots(20, 0);
      });

      expect(mockWebSocketService.call).toHaveBeenCalledWith(RPC_METHODS.LIST_SNAPSHOTS, {
        limit: 20,
        offset: 0
      });
      expect(result.current.snapshots).toEqual(mockSnapshots);
      expect(result.current.isLoading).toBe(false);
      expect(result.current.error).toBeNull();
    });

    it('should handle empty snapshots response', async () => {
      mockWebSocketService.call.mockResolvedValue({
        files: []
      });

      const { result } = renderHook(() => useFileStore());

      await act(async () => {
        await result.current.loadSnapshots();
      });

      expect(result.current.snapshots).toEqual([]);
    });

    it('should handle snapshots loading error', async () => {
      mockWebSocketService.call.mockRejectedValue(new Error('Failed to load snapshots'));

      const { result } = renderHook(() => useFileStore());

      await act(async () => {
        await result.current.loadSnapshots();
      });

      expect(result.current.error).toBe('Failed to load snapshots');
      expect(result.current.snapshots).toBeNull();
      expect(result.current.isLoading).toBe(false);
    });
  });

  describe('Download File', () => {
    beforeEach(() => {
      // Mock document.createElement and related DOM methods
      const mockLink = {
        href: '',
        download: '',
        target: '',
        click: jest.fn()
      };

      Object.defineProperty(document, 'createElement', {
        value: jest.fn(() => mockLink),
        writable: true
      });

      Object.defineProperty(document.body, 'appendChild', {
        value: jest.fn(),
        writable: true
      });

      Object.defineProperty(document.body, 'removeChild', {
        value: jest.fn(),
        writable: true
      });
    });

    it('should download file successfully', async () => {
      const { result } = renderHook(() => useFileStore());

      await act(async () => {
        await result.current.downloadFile('recordings', 'test-recording.mp4');
      });

      expect(result.current.isDownloading).toBe(false);
      expect(result.current.error).toBeNull();
    });

    it('should construct correct download URL', async () => {
      // Mock window.location.origin
      Object.defineProperty(window, 'location', {
        value: { origin: 'https://localhost:3000' },
        writable: true
      });

      const { result } = renderHook(() => useFileStore());

      await act(async () => {
        await result.current.downloadFile('recordings', 'test-recording.mp4');
      });

      const mockLink = (document.createElement as jest.Mock).mock.results[0].value;
      expect(mockLink.href).toBe('https://localhost:3000/files/recordings/test-recording.mp4');
      expect(mockLink.download).toBe('test-recording.mp4');
    });

    it('should handle download errors', async () => {
      // Mock document.createElement to throw error
      Object.defineProperty(document, 'createElement', {
        value: jest.fn(() => {
          throw new Error('Download failed');
        }),
        writable: true
      });

      const { result } = renderHook(() => useFileStore());

      await act(async () => {
        await result.current.downloadFile('recordings', 'test-recording.mp4');
      });

      expect(result.current.error).toBe('Download failed');
      expect(result.current.isDownloading).toBe(false);
    });

    it('should set downloading state during operation', async () => {
      const { result } = renderHook(() => useFileStore());

      const downloadPromise = act(async () => {
        await result.current.downloadFile('recordings', 'test-recording.mp4');
      });

      expect(result.current.isDownloading).toBe(true);

      await downloadPromise;
      expect(result.current.isDownloading).toBe(false);
    });
  });

  describe('State Management', () => {
    it('should set error state', () => {
      const { result } = renderHook(() => useFileStore());

      act(() => {
        result.current.setError('Test error');
      });

      expect(result.current.error).toBe('Test error');
    });

    it('should clear error state', () => {
      const { result } = renderHook(() => useFileStore());

      act(() => {
        result.current.setError('Test error');
        result.current.clearError();
      });

      expect(result.current.error).toBeNull();
    });

    it('should set connection status', () => {
      const { result } = renderHook(() => useFileStore());

      act(() => {
        result.current.setConnectionStatus(true);
      });

      expect(result.current.isConnected).toBe(true);
    });

    it('should update file list for recordings', () => {
      const { result } = renderHook(() => useFileStore());

      const testFiles = [{ filename: 'test.mp4', file_size: 1024, created_at: '2024-01-01T00:00:00Z', modified_time: '2024-01-01T00:00:00Z', download_url: '/test.mp4' }];

      act(() => {
        result.current.updateFileList('recordings', testFiles);
      });

      expect(result.current.recordings).toEqual(testFiles);
    });

    it('should update file list for snapshots', () => {
      const { result } = renderHook(() => useFileStore());

      const testFiles = [{ filename: 'test.jpg', file_size: 512, created_at: '2024-01-01T00:00:00Z', modified_time: '2024-01-01T00:00:00Z', download_url: '/test.jpg' }];

      act(() => {
        result.current.updateFileList('snapshots', testFiles);
      });

      expect(result.current.snapshots).toEqual(testFiles);
    });
  });

  describe('WebSocket Event Handling', () => {
    it('should handle connect event', async () => {
      mockWebSocketService.connect.mockResolvedValue();

      const { result } = renderHook(() => useFileStore());

      await act(async () => {
        await result.current.initialize();
      });

      // Simulate connect event
      const connectHandler = mockWebSocketService.onConnect.mock.calls[0][0];
      act(() => {
        connectHandler();
      });

      expect(result.current.isConnected).toBe(true);
      expect(result.current.error).toBeNull();
    });

    it('should handle disconnect event', async () => {
      mockWebSocketService.connect.mockResolvedValue();

      const { result } = renderHook(() => useFileStore());

      await act(async () => {
        await result.current.initialize();
      });

      // Simulate disconnect event
      const disconnectHandler = mockWebSocketService.onDisconnect.mock.calls[0][0];
      act(() => {
        disconnectHandler();
      });

      expect(result.current.isConnected).toBe(false);
    });

    it('should handle error event', async () => {
      mockWebSocketService.connect.mockResolvedValue();

      const { result } = renderHook(() => useFileStore());

      await act(async () => {
        await result.current.initialize();
      });

      // Simulate error event
      const errorHandler = mockWebSocketService.onError.mock.calls[0][0];
      act(() => {
        errorHandler(new Error('WebSocket error'));
      });

      expect(result.current.error).toBe('WebSocket error');
      expect(result.current.isConnected).toBe(false);
    });
  });
});
