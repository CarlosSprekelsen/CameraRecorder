/**
 * REQ-UNIT01-001: [Primary requirement being tested]
 * REQ-UNIT01-002: [Secondary requirements covered]
 * Coverage: UNIT
 * Quality: HIGH
 */
/**
 * Unit tests for file store
 * 
 * Design Principles:
 * - Pure unit testing with complete isolation
 * - Direct store testing without React context dependency
 * - Uses proven mock server fixture for consistency
 * - Focus on business logic, not implementation details
 * - Strong test boundaries and predictable behavior
 */

import { useFileStore } from '../../../src/stores/fileStore';
import { RPC_METHODS } from '../../../src/types';

// Import proven mock server fixture
const { MockWebSocketService } = require('../../fixtures/mock-server');

// Create a proper mock WebSocket service that matches the real interface
const mockWebSocketService = {
  connect: jest.fn().mockResolvedValue(undefined),
  disconnect: jest.fn(),
  call: jest.fn().mockResolvedValue({ files: [] }),
  onConnect: jest.fn(),
  onDisconnect: jest.fn(),
  onError: jest.fn(),
  onMessage: jest.fn(),
  setConnectionStore: jest.fn(),
  setCameraStore: jest.fn(),
  onCameraStatusUpdate: jest.fn(),
  onRecordingStatusUpdate: jest.fn(),
  onNotification: jest.fn()
};

// Mock the WebSocket service factory
jest.mock('../../../src/services/websocket', () => ({
  createWebSocketService: jest.fn(() => Promise.resolve(mockWebSocketService))
}));

// Mock normalizer to return files directly for easier testing
jest.mock('../../../src/services/apiNormalizer', () => ({
  normalizeFileListResponse: jest.fn((response) => ({ files: response.files || [] }))
}));

describe('File Store', () => {
  let store: ReturnType<typeof useFileStore.getState>;

  beforeEach(() => {
    // Reset all mocks to ensure test isolation
    jest.clearAllMocks();
    
    // Reset store state completely
    const currentStore = useFileStore.getState();
    currentStore.disconnect();
    
    // Get fresh store instance after reset
    store = useFileStore.getState();
  });

  describe('Initialization', () => {
    it('should start with correct default state', () => {
      const state = useFileStore.getState();
      expect(state.recordings).toBeNull();
      expect(state.snapshots).toBeNull();
      expect(state.isLoading).toBe(false);
      expect(state.isDownloading).toBe(false);
      expect(state.error).toBeNull();
      expect(state.isConnected).toBe(false);
    });

    it('should initialize WebSocket service with correct configuration', async () => {
      await store.initialize();

      const createWebSocketService = require('../../../src/services/websocket').createWebSocketService;
      expect(createWebSocketService).toHaveBeenCalledWith({
        url: 'ws://localhost:8002/ws',
        reconnectInterval: 5000,
        maxReconnectAttempts: 5,
      });
    });

    it('should handle initialization errors gracefully', async () => {
      // Mock the websocket service factory to throw
      const createWebSocketService = require('../../../src/services/websocket').createWebSocketService;
      createWebSocketService.mockRejectedValueOnce(new Error('Connection failed'));

      await store.initialize();

      const state = useFileStore.getState();
      expect(state.error).toBe('Connection failed');
      expect(state.isConnected).toBe(false);
    });

    it('should prevent double initialization', async () => {
      await store.initialize();
      await store.initialize(); // Second call should be ignored

      const createWebSocketService = require('../../../src/services/websocket').createWebSocketService;
      expect(createWebSocketService).toHaveBeenCalledTimes(1);
    });
  });

  describe('Connection Management', () => {
    it('should disconnect WebSocket service', async () => {
      await store.initialize();
      store.disconnect();

      const state = useFileStore.getState();
      expect(state.isConnected).toBe(false);
      expect(state.recordings).toBeNull();
      expect(state.snapshots).toBeNull();
    });

    it('should update connection status', () => {
      // Test setting connection status
      store.setConnectionStatus(true);
      let state = useFileStore.getState();
      expect(state.isConnected).toBe(true);

      store.setConnectionStatus(false);
      state = useFileStore.getState();
      expect(state.isConnected).toBe(false);
    });
  });

  describe('File Operations', () => {
    beforeEach(async () => {
      // Set up connected state for file operations
      await store.initialize();
      // Manually set connection status since the onConnect callback might not be called in tests
      store.setConnectionStatus(true);
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

      // Configure mock to return recordings
      mockWebSocketService.call.mockResolvedValueOnce({ files: mockRecordings });

      await store.loadRecordings();

      const state = useFileStore.getState();
      expect(state.recordings).toEqual(mockRecordings);
      expect(state.isLoading).toBe(false);
      expect(state.error).toBeNull();
    });

    it('should load snapshots successfully', async () => {
      const mockSnapshots = [
        { 
          filename: 'snapshot1.jpg', 
          file_size: 512, 
          created_at: '2023-01-01T00:00:00Z',
          modified_time: '2023-01-01T00:00:00Z',
          download_url: '/api/files/snapshots/snapshot1.jpg'
        }
      ];

      // Configure mock to return snapshots
      mockWebSocketService.call.mockResolvedValueOnce({ files: mockSnapshots });

      await store.loadSnapshots();

      const state = useFileStore.getState();
      expect(state.snapshots).toEqual(mockSnapshots);
      expect(state.isLoading).toBe(false);
      expect(state.error).toBeNull();
    });

    it('should handle empty file lists gracefully', async () => {
      // Mock empty response
      mockWebSocketService.call.mockResolvedValueOnce({ files: [] });

      await store.loadRecordings();

      const state = useFileStore.getState();
      expect(state.recordings).toEqual([]);
      expect(state.error).toBeNull();
    });

    it('should prevent operations when not connected', async () => {
      const disconnectedStore = useFileStore.getState();
      disconnectedStore.disconnect();
      
      await disconnectedStore.loadRecordings();

      const state = useFileStore.getState();
      expect(state.error).toBe('WebSocket not connected');
    });
  });

  describe('Download Operations', () => {
    beforeEach(async () => {
      // Initialize store for download operations
      await store.initialize();
      store.setConnectionStatus(true);
      
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

      await store.downloadFile('recordings', 'test.mp4');

      const state = useFileStore.getState();
      expect(state.isDownloading).toBe(false);
      expect(state.error).toBeNull();
      expect(global.fetch).toHaveBeenCalledTimes(2);
    });

    it('should handle file not found error', async () => {
      // Mock 404 HEAD response
      global.fetch = jest.fn().mockResolvedValueOnce({ 
        ok: false, 
        status: 404,
        statusText: 'Not Found'
      });

      await store.downloadFile('recordings', 'missing.mp4');

      const state = useFileStore.getState();
      expect(state.error).toBe('File not found');
      expect(state.isDownloading).toBe(false);
    });

    it('should handle download network errors', async () => {
      // Mock network error
      global.fetch = jest.fn().mockRejectedValueOnce(new Error('Network error'));

      await store.downloadFile('recordings', 'test.mp4');

      const state = useFileStore.getState();
      expect(state.error).toBe('Network error');
      expect(state.isDownloading).toBe(false);
    });

    it('should handle server error responses', async () => {
      // Mock server error
      global.fetch = jest.fn().mockResolvedValueOnce({ 
        ok: false, 
        status: 500,
        statusText: 'Internal Server Error'
      });

      await store.downloadFile('recordings', 'test.mp4');

      const state = useFileStore.getState();
      expect(state.error).toBe('Download failed: 500 Internal Server Error');
      expect(state.isDownloading).toBe(false);
    });
  });

  describe('State Management', () => {
    it('should set and clear errors', () => {
      store.setError('Test error');
      let state = useFileStore.getState();
      expect(state.error).toBe('Test error');

      store.clearError();
      state = useFileStore.getState();
      expect(state.error).toBeNull();
    });

    it('should update file lists', () => {
      const testFiles = [{ 
        filename: 'test.mp4', 
        file_size: 1024, 
        created_at: '2023-01-01T00:00:00Z',
        modified_time: '2023-01-01T00:00:00Z',
        download_url: '/api/files/recordings/test.mp4'
      }];

      store.updateFileList('recordings', testFiles);
      let state = useFileStore.getState();
      expect(state.recordings).toEqual(testFiles);

      store.updateFileList('snapshots', testFiles);
      state = useFileStore.getState();
      expect(state.snapshots).toEqual(testFiles);
    });
  });
});