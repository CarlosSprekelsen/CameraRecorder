/**
 * Unit tests for file store
 * 
 * Design Principles:
 * - Pure unit testing with complete isolation
 * - Clean mocking with dependency injection pattern
 * - Focus on business logic, not implementation details
 * - Strong test boundaries and predictable behavior
 */

import { renderHook, act } from '@testing-library/react';
import { useFileStore } from '../../../src/stores/fileStore';
import { RPC_METHODS } from '../../../src/types';

// Simple, clean mock for WebSocket service
const mockWebSocketService = {
  connect: jest.fn(),
  disconnect: jest.fn(),
  call: jest.fn(),
  onConnect: jest.fn(),
  onDisconnect: jest.fn(),
  onError: jest.fn(),
  onMessage: jest.fn()
};

// Mock the WebSocket service factory with simple, predictable behavior
jest.mock('../../../src/services/websocket', () => ({
  createWebSocketService: jest.fn(() => Promise.resolve(mockWebSocketService))
}));

describe('File Store', () => {
  beforeEach(() => {
    // Reset all mocks to ensure test isolation
    jest.clearAllMocks();
    
    // Set up default mock behaviors
    mockWebSocketService.connect.mockResolvedValue(undefined);
    mockWebSocketService.call.mockResolvedValue({ files: [] });
    
    // Reset store state by creating fresh instance
    const { result } = renderHook(() => useFileStore());
    act(() => {
      result.current.disconnect();
    });
  });

  describe('Initialization', () => {
    it('should start with correct default state', () => {
      const { result } = renderHook(() => useFileStore());

      expect(result.current.recordings).toBeNull();
      expect(result.current.snapshots).toBeNull();
      expect(result.current.isLoading).toBe(false);
      expect(result.current.isDownloading).toBe(false);
      expect(result.current.error).toBeNull();
      expect(result.current.isConnected).toBe(false);
    });

    it('should initialize WebSocket service with correct configuration', async () => {
      const { result } = renderHook(() => useFileStore());

      await act(async () => {
        await result.current.initialize();
      });

      const createWebSocketService = require('../../../src/services/websocket').createWebSocketService;
      expect(createWebSocketService).toHaveBeenCalledWith({
        url: 'ws://localhost:8002/ws',
        reconnectInterval: 5000,
        maxReconnectAttempts: 5,
      });
    });

    it('should set up event handlers during initialization', async () => {
      const { result } = renderHook(() => useFileStore());

      await act(async () => {
        await result.current.initialize();
      });

      expect(mockWebSocketService.onConnect).toHaveBeenCalled();
      expect(mockWebSocketService.onDisconnect).toHaveBeenCalled();
      expect(mockWebSocketService.onError).toHaveBeenCalled();
    });

    it('should handle initialization errors gracefully', async () => {
      const initError = new Error('Connection failed');
      mockWebSocketService.connect.mockRejectedValue(initError);

      const { result } = renderHook(() => useFileStore());

      await act(async () => {
        await result.current.initialize();
      });

      expect(result.current.error).toBe('Connection failed');
      expect(result.current.isConnected).toBe(false);
    });

    it('should prevent double initialization', async () => {
      const { result } = renderHook(() => useFileStore());

      await act(async () => {
        await result.current.initialize();
        await result.current.initialize(); // Second call should be ignored
      });

      const createWebSocketService = require('../../../src/services/websocket').createWebSocketService;
      expect(createWebSocketService).toHaveBeenCalledTimes(1);
    });
  });

  describe('Connection Management', () => {
    it('should disconnect WebSocket service', async () => {
      const { result } = renderHook(() => useFileStore());
      
      await act(async () => {
        await result.current.initialize();
        result.current.disconnect();
      });

      expect(mockWebSocketService.disconnect).toHaveBeenCalled();
    });

    it('should update connection state when connected', async () => {
      const { result } = renderHook(() => useFileStore());
      
      await act(async () => {
        await result.current.initialize();
      });

      // Simulate connection event
      const connectHandler = mockWebSocketService.onConnect.mock.calls[0][0];
      
      act(() => {
        connectHandler();
      });

      expect(result.current.isConnected).toBe(true);
      expect(result.current.error).toBeNull();
    });

    it('should update connection state when disconnected', async () => {
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

    it('should handle connection errors', async () => {
      const { result } = renderHook(() => useFileStore());
      
      await act(async () => {
        await result.current.initialize();
      });

      // Simulate error event
      const errorHandler = mockWebSocketService.onError.mock.calls[0][0];
      const testError = new Error('WebSocket error');
      
      act(() => {
        errorHandler(testError);
      });

      expect(result.current.error).toBe('WebSocket error');
    });
  });

  describe('File Operations', () => {
    beforeEach(async () => {
      // Set up connected state for file operations
      const { result } = renderHook(() => useFileStore());
      await act(async () => {
        await result.current.initialize();
      });
      
      // Simulate connection
      const connectHandler = mockWebSocketService.onConnect.mock.calls[0][0];
      act(() => {
        connectHandler();
      });
    });

    it('should load recordings successfully', async () => {
      const mockRecordings = [
        { 
          filename: 'recording1.mp4', 
          file_size: 1024, 
          created_at: '2023-01-01T00:00:00Z',
          modified_time: '2023-01-01T00:00:00Z',
          download_url: '/api/files/recordings/recording1.mp4'
        },
        { 
          filename: 'recording2.mp4', 
          file_size: 2048, 
          created_at: '2023-01-02T00:00:00Z',
          modified_time: '2023-01-02T00:00:00Z',
          download_url: '/api/files/recordings/recording2.mp4'
        }
      ];
      
      mockWebSocketService.call.mockResolvedValue({ files: mockRecordings });

      const { result } = renderHook(() => useFileStore());

      await act(async () => {
        await result.current.loadRecordings();
      });

      expect(mockWebSocketService.call).toHaveBeenCalledWith(RPC_METHODS.LIST_RECORDINGS, {
        limit: 20,
        offset: 0
      });
      expect(result.current.recordings).toEqual(mockRecordings);
      expect(result.current.isLoading).toBe(false);
      expect(result.current.error).toBeNull();
    });

    it('should load snapshots successfully', async () => {
      const mockSnapshots = [
        { 
          filename: 'snapshot1.jpg', 
          file_size: 512, 
          created_at: '2023-01-01T00:00:00Z',
          modified_time: '2023-01-01T00:00:00Z',
          download_url: '/api/files/snapshots/snapshot1.jpg'
        },
        { 
          filename: 'snapshot2.jpg', 
          file_size: 768, 
          created_at: '2023-01-02T00:00:00Z',
          modified_time: '2023-01-02T00:00:00Z',
          download_url: '/api/files/snapshots/snapshot2.jpg'
        }
      ];
      
      mockWebSocketService.call.mockResolvedValue({ files: mockSnapshots });

      const { result } = renderHook(() => useFileStore());

      await act(async () => {
        await result.current.loadSnapshots();
      });

      expect(mockWebSocketService.call).toHaveBeenCalledWith(RPC_METHODS.LIST_SNAPSHOTS, {
        limit: 20,
        offset: 0
      });
      expect(result.current.snapshots).toEqual(mockSnapshots);
    });

    it('should handle empty file lists gracefully', async () => {
      mockWebSocketService.call.mockResolvedValue({ files: [] });

      const { result } = renderHook(() => useFileStore());

      await act(async () => {
        await result.current.loadRecordings();
      });

      expect(result.current.recordings).toEqual([]);
      expect(result.current.error).toBeNull();
    });

    it('should handle file loading errors', async () => {
      const { result } = renderHook(() => useFileStore());
      
      // Initialize and connect first
      await act(async () => {
        await result.current.initialize();
      });
      
      const connectHandler = mockWebSocketService.onConnect.mock.calls[0][0];
      act(() => {
        connectHandler();
      });

      // Now mock the error
      mockWebSocketService.call.mockRejectedValue(new Error('Failed to load files'));

      await act(async () => {
        await result.current.loadRecordings();
      });

      expect(result.current.error).toBe('Failed to load files');
      expect(result.current.isLoading).toBe(false);
    });

    it('should prevent operations when not connected', async () => {
      // Create store without connecting
      const { result } = renderHook(() => useFileStore());

      await act(async () => {
        await result.current.loadRecordings();
      });

      expect(result.current.error).toBe('WebSocket not connected');
      expect(mockWebSocketService.call).not.toHaveBeenCalled();
    });
  });

  describe('Download Operations', () => {
    beforeEach(() => {
      // Mock URL methods
      global.URL.createObjectURL = jest.fn(() => 'blob:mock-url');
      global.URL.revokeObjectURL = jest.fn();
      
      // Mock document methods
      const mockLink = {
        href: '',
        download: '',
        target: '',
        click: jest.fn()
      };
      
      jest.spyOn(document, 'createElement').mockReturnValue(mockLink as any);
      jest.spyOn(document.body, 'appendChild').mockImplementation(() => mockLink as any);
      jest.spyOn(document.body, 'removeChild').mockImplementation(() => mockLink as any);
    });

    it('should handle download success', async () => {
      // Mock successful HEAD and GET requests
      global.fetch = jest.fn()
        .mockResolvedValueOnce({ ok: true }) // HEAD request
        .mockResolvedValueOnce({ 
          ok: true, 
          blob: () => Promise.resolve(new Blob(['file content']))
        }); // GET request

      const { result } = renderHook(() => useFileStore());

      await act(async () => {
        await result.current.downloadFile('recordings', 'test.mp4');
      });

      expect(result.current.isDownloading).toBe(false);
      expect(result.current.error).toBeNull();
      expect(global.fetch).toHaveBeenCalledTimes(2);
    });

    it('should handle file not found error', async () => {
      // Mock 404 HEAD response
      global.fetch = jest.fn().mockResolvedValueOnce({ 
        ok: false, 
        status: 404,
        statusText: 'Not Found'
      });

      const { result } = renderHook(() => useFileStore());

      await act(async () => {
        await result.current.downloadFile('recordings', 'missing.mp4');
      });

      expect(result.current.error).toBe('File not found');
      expect(result.current.isDownloading).toBe(false);
    });

    it('should handle download network errors', async () => {
      // Mock network error
      global.fetch = jest.fn().mockRejectedValueOnce(new Error('Network error'));

      const { result } = renderHook(() => useFileStore());

      await act(async () => {
        await result.current.downloadFile('recordings', 'test.mp4');
      });

      expect(result.current.error).toBe('Network error');
      expect(result.current.isDownloading).toBe(false);
    });

    it('should handle server error responses', async () => {
      // Mock server error
      global.fetch = jest.fn().mockResolvedValueOnce({ 
        ok: false, 
        status: 500,
        statusText: 'Internal Server Error'
      });

      const { result } = renderHook(() => useFileStore());

      await act(async () => {
        await result.current.downloadFile('recordings', 'test.mp4');
      });

      expect(result.current.error).toBe('Download failed: 500 Internal Server Error');
      expect(result.current.isDownloading).toBe(false);
    });
  });

  describe('State Management', () => {
    it('should set and clear errors', () => {
      const { result } = renderHook(() => useFileStore());

      act(() => {
        result.current.setError('Test error');
      });
      expect(result.current.error).toBe('Test error');

      act(() => {
        result.current.clearError();
      });
      expect(result.current.error).toBeNull();
    });

    it('should update connection status', () => {
      const { result } = renderHook(() => useFileStore());

      act(() => {
        result.current.setConnectionStatus(true);
      });
      expect(result.current.isConnected).toBe(true);

      act(() => {
        result.current.setConnectionStatus(false);
      });
      expect(result.current.isConnected).toBe(false);
    });

    it('should update file lists', () => {
      const { result } = renderHook(() => useFileStore());
      const testFiles = [{ 
        filename: 'test.mp4', 
        file_size: 1024, 
        created_at: '2023-01-01T00:00:00Z',
        modified_time: '2023-01-01T00:00:00Z',
        download_url: '/api/files/recordings/test.mp4'
      }];

      act(() => {
        result.current.updateFileList('recordings', testFiles);
      });
      expect(result.current.recordings).toEqual(testFiles);

      act(() => {
        result.current.updateFileList('snapshots', testFiles);
      });
      expect(result.current.snapshots).toEqual(testFiles);
    });
  });
});