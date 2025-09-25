/**
 * Unit Tests for Device Store
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
import { useDeviceStore } from '../../../src/stores/device/deviceStore';

// Mock the DeviceService
const mockDeviceService = {
  getCameraList: jest.fn() as jest.MockedFunction<any>,
  getCameraStatus: jest.fn() as jest.MockedFunction<any>,
  getCameraCapabilities: jest.fn() as jest.MockedFunction<any>,
  getStreamUrl: jest.fn() as jest.MockedFunction<any>,
  getStreamStatus: jest.fn() as jest.MockedFunction<any>,
  getStreams: jest.fn() as jest.MockedFunction<any>
};

jest.mock('../../../src/services/device/DeviceService', () => ({
  DeviceService: jest.fn().mockImplementation(() => mockDeviceService)
}));

describe('Device Store', () => {
  beforeEach(() => {
    // Reset the store to initial state
    useDeviceStore.getState().reset();
    
    // Reset all mocks
    jest.clearAllMocks();
    
    // Set up default mock implementations
    mockDeviceService.getCameraList.mockResolvedValue(MockDataFactory.getCameraListResult());
    mockDeviceService.getStreamUrl.mockResolvedValue('rtsp://localhost:8554/camera0');
    mockDeviceService.getStreams.mockResolvedValue(MockDataFactory.getStreamsListResult());
  });

  afterEach(() => {
    useDeviceStore.getState().reset();
  });

  describe('REQ-001: Store State Management', () => {
    test('should initialize with correct initial state', () => {
      const state = useDeviceStore.getState();
      
      expect(state.cameras).toEqual([]);
      expect(state.streams).toEqual([]);
      expect(state.loading).toBe(false);
      expect(state.error).toBe(null);
      expect(state.lastUpdated).toBe(null);
    });

    test('should set loading state correctly', () => {
      const { setLoading } = useDeviceStore.getState();
      
      setLoading(true);
      expect(useDeviceStore.getState().loading).toBe(true);
      
      setLoading(false);
      expect(useDeviceStore.getState().loading).toBe(false);
    });

    test('should set error state correctly', () => {
      const { setError } = useDeviceStore.getState();
      const errorMessage = 'Test error';
      
      setError(errorMessage);
      expect(useDeviceStore.getState().error).toBe(errorMessage);
      
      setError(null);
      expect(useDeviceStore.getState().error).toBe(null);
    });

    test('should update camera status correctly', () => {
      const { updateCameraStatus } = useDeviceStore.getState();
      
      // Create a camera with all required fields for the store
      const camera = {
        device: 'camera0',
        status: 'CONNECTED' as const,
        name: 'Test Camera',
        resolution: '1920x1080',
        fps: 30,
        streams: {
          rtsp: 'rtsp://localhost:8554/camera0',
          hls: 'https://localhost/hls/camera0.m3u8'
        }
      };
      
      // Add the camera first
      useDeviceStore.setState({ cameras: [camera] });
      
      updateCameraStatus('camera0', 'DISCONNECTED');
      const cameras = useDeviceStore.getState().cameras;
      expect(cameras[0].status).toBe('DISCONNECTED');
    });

    test('should update stream status correctly', () => {
      const { updateStreamStatus } = useDeviceStore.getState();
      
      // Add a stream first
      useDeviceStore.setState({ 
        streams: [MockDataFactory.getStreamsListResult()[0]] 
      });
      
      updateStreamStatus('stream1', false, 5);
      const streams = useDeviceStore.getState().streams;
      expect(streams[0].ready).toBe(false);
      expect(streams[0].readers).toBe(5);
    });
  });

  describe('REQ-002: State Transitions', () => {
    test('should transition from loading to success state', async () => {
      const { getCameraList, setDeviceService } = useDeviceStore.getState();
      
      // Set up the service
      setDeviceService(mockDeviceService as any);
      
      // Start the action
      const promise = getCameraList();
      
      // Check loading state
      expect(useDeviceStore.getState().loading).toBe(true);
      expect(useDeviceStore.getState().error).toBe(null);
      
      // Wait for completion
      await promise;
      
      // Check final state
      const state = useDeviceStore.getState();
      expect(state.loading).toBe(false);
      expect(state.error).toBe(null);
      expect(state.cameras.length).toBeGreaterThan(0);
      expect(state.lastUpdated).toBeTruthy();
    });

    test('should transition from loading to error state', async () => {
      const { getCameraList, setDeviceService } = useDeviceStore.getState();
      
      // Set up the service to reject
      mockDeviceService.getCameraList.mockRejectedValue(new Error('Network error'));
      setDeviceService(mockDeviceService as any);
      
      // Start the action
      await getCameraList();
      
      // Check final state
      const state = useDeviceStore.getState();
      expect(state.loading).toBe(false);
      expect(state.error).toBe('Network error');
      expect(state.cameras).toEqual([]);
    });
  });

  describe('REQ-003: Error Handling', () => {
    test('should handle service not initialized error', async () => {
      const { getCameraList } = useDeviceStore.getState();
      
      // Don't set the service
      await getCameraList();
      
      const state = useDeviceStore.getState();
      expect(state.error).toBe('Device service not initialized');
      expect(state.loading).toBe(false);
    });

    test('should handle API errors gracefully', async () => {
      const { getCameraList, setDeviceService } = useDeviceStore.getState();
      
      // Set up the service to reject
      mockDeviceService.getCameraList.mockRejectedValue(new Error('API Error'));
      setDeviceService(mockDeviceService as any);
      
      await getCameraList();
      
      const state = useDeviceStore.getState();
      expect(state.error).toBe('API Error');
      expect(state.loading).toBe(false);
    });

    test('should handle non-Error exceptions', async () => {
      const { getCameraList, setDeviceService } = useDeviceStore.getState();
      
      // Set up the service to reject with a string
      mockDeviceService.getCameraList.mockRejectedValue('String error');
      setDeviceService(mockDeviceService as any);
      
      await getCameraList();
      
      const state = useDeviceStore.getState();
      expect(state.error).toBe('Failed to get camera list');
      expect(state.loading).toBe(false);
    });
  });

  describe('REQ-004: API Integration', () => {
    test('should call getCameraList and update state', async () => {
      const { getCameraList, setDeviceService } = useDeviceStore.getState();
      
      setDeviceService(mockDeviceService as any);
      await getCameraList();
      
      expect(mockDeviceService.getCameraList).toHaveBeenCalledTimes(1);
      
      const state = useDeviceStore.getState();
      expect(state.cameras.length).toBeGreaterThan(0);
    });

    test('should call getStreamUrl and return result', async () => {
      const { getStreamUrl, setDeviceService } = useDeviceStore.getState();
      
      setDeviceService(mockDeviceService as any);
      const result = await getStreamUrl('camera0');
      
      expect(mockDeviceService.getStreamUrl).toHaveBeenCalledWith('camera0');
      expect(result).toBe('rtsp://localhost:8554/camera0');
    });

    test('should call getStreams and update state', async () => {
      const { getStreams, setDeviceService } = useDeviceStore.getState();
      
      setDeviceService(mockDeviceService as any);
      await getStreams();
      
      expect(mockDeviceService.getStreams).toHaveBeenCalledTimes(1);
      
      const state = useDeviceStore.getState();
      expect(state.streams).toEqual(MockDataFactory.getStreamsListResult());
    });
  });

  describe('REQ-005: Side Effects', () => {
    test('should handle camera status updates', () => {
      const { handleCameraStatusUpdate } = useDeviceStore.getState();
      
      // Create a camera with all required fields for the store
      const camera = {
        device: 'camera0',
        status: 'CONNECTED' as const,
        name: 'Test Camera',
        resolution: '1920x1080',
        fps: 30,
        streams: {
          rtsp: 'rtsp://localhost:8554/camera0',
          hls: 'https://localhost/hls/camera0.m3u8'
        }
      };
      
      handleCameraStatusUpdate(camera);
      
      const state = useDeviceStore.getState();
      expect(state.cameras).toContain(camera);
    });

    test('should handle stream updates', () => {
      const { handleStreamUpdate } = useDeviceStore.getState();
      const stream = MockDataFactory.getStreamsListResult()[0];
      
      handleStreamUpdate(stream);
      
      const state = useDeviceStore.getState();
      expect(state.streams).toContain(stream);
    });

    test('should reset store to initial state', () => {
      const { reset, setLoading, setError } = useDeviceStore.getState();
      
      // Modify state
      setLoading(true);
      setError('Test error');
      useDeviceStore.setState({ 
        cameras: MockDataFactory.getCameraListResult().cameras,
        lastUpdated: '2025-01-15T14:30:00Z'
      });
      
      // Reset
      reset();
      
      // Check state is back to initial
      const state = useDeviceStore.getState();
      expect(state.cameras).toEqual([]);
      expect(state.streams).toEqual([]);
      expect(state.loading).toBe(false);
      expect(state.error).toBe(null);
      expect(state.lastUpdated).toBe(null);
    });
  });

  describe('API Compliance Validation', () => {
    test('should validate camera list response against RPC spec', async () => {
      const { getCameraList, setDeviceService } = useDeviceStore.getState();
      
      setDeviceService(mockDeviceService as any);
      await getCameraList();
      
      const cameras = useDeviceStore.getState().cameras;
      expect(cameras.length).toBeGreaterThan(0);
      
      // Validate each camera against RPC spec
      cameras.forEach(camera => {
        expect(APIResponseValidator.validateCamera(camera)).toBe(true);
      });
    });

    test('should validate stream response against RPC spec', async () => {
      const { getStreams, setDeviceService } = useDeviceStore.getState();
      
      setDeviceService(mockDeviceService as any);
      await getStreams();
      
      const streams = useDeviceStore.getState().streams;
      expect(streams.length).toBeGreaterThan(0);
      
      // Validate each stream against RPC spec
      streams.forEach(stream => {
        expect(APIResponseValidator.validateStreamsListResult([stream])).toBe(true);
      });
    });
  });
});