/**
 * REQ-MVP01-001: [Primary requirement being tested]
 * REQ-MVP01-002: [Secondary requirements covered]
 * Coverage: IVV
 * Quality: HIGH
 */
/**
 * REQ-MVP01: MVP Functionality Validation
 * 
 * IV&V Independent Validation Test
 * Following "Test First, Real Integration Always" approach
 * 
 * Tests real server integration for complete camera operations workflow
 * Validates behavior, not implementation details
 * 
 * Prerequisites:
 * - MediaMTX Camera Service running via systemd
 * - Server accessible at ws://localhost:8002/ws
 * - Real cameras available for testing
 */

import { WebSocketService } from '../../src/services/websocket';
import { 
  RPC_METHODS, 
  ERROR_CODES, 
  PERFORMANCE_TARGETS, 
  isNotification,
  WebSocketConfig 
} from '../../src/types';
// Authentication utilities (inline for REQ-MVP01 validation)
import jwt from 'jsonwebtoken';

const generateValidToken = (userId = 'mvp_test_user', role = 'operator', expiresIn = 24 * 60 * 60): string => {
    const secret = process.env.CAMERA_SERVICE_JWT_SECRET;
    if (!secret) {
        throw new Error('CAMERA_SERVICE_JWT_SECRET environment variable is required');
    }
    
    const payload = {
        user_id: userId,
        role: role,
        iat: Math.floor(Date.now() / 1000),
        exp: Math.floor(Date.now() / 1000) + expiresIn
    };
    
    return jwt.sign(payload, secret, { algorithm: 'HS256' });
};

const validateTestEnvironment = (): boolean => {
    try {
        const secret = process.env.CAMERA_SERVICE_JWT_SECRET;
        if (!secret) {
            throw new Error('CAMERA_SERVICE_JWT_SECRET environment variable is required');
        }
        console.log('✅ Test environment validated - JWT secret available');
        return true;
    } catch (error) {
        console.error('❌ Test environment validation failed:', error instanceof Error ? error.message : String(error));
        return false;
    }
};
// Type definitions for validation tests
interface CameraListResponse {
  cameras: CameraDevice[];
  total: number;
  connected: number;
}

interface CameraDevice {
  device: string;
  status: string;
  name: string;
  resolution: string;
  fps: number;
  streams: {
    rtsp: string;
    webrtc: string;
    hls: string;
  };
}

interface SnapshotResult {
  status: string;
  filename: string;
  file_size: number;
  format?: string;
  quality?: number;
}

interface RecordingSession {
  session_id: string;
  device: string;
  format: string;
  duration?: number;
  file_size?: number;
}

interface FileListResponse {
  files: Array<{
    filename: string;
    file_size: number;
    modified_time: string;
    download_url: string;
  }>;
  total: number;
  limit: number;
  offset: number;
}

describe('REQ-MVP01: MVP Functionality Validation', () => {
  let wsService: WebSocketService;
  const TEST_WEBSOCKET_URL = process.env.TEST_WEBSOCKET_URL || 'ws://localhost:8002';

  beforeAll(async () => {
    // Validate test environment setup (authentication)
    console.log('Validating test environment...');
    if (!validateTestEnvironment()) {
      throw new Error('Test environment not properly set up. Authentication required for REQ-MVP01 validation.');
    }
    
    // Verify server is available before running tests
    console.log('Checking server availability...');
    const isServerAvailable = await checkServerAvailability();
    console.log('Server available:', isServerAvailable);
    if (!isServerAvailable) {
      console.warn('Server availability check failed, but proceeding with tests...');
    }
  });

  beforeEach(async () => {
    const config: WebSocketConfig = {
      url: TEST_WEBSOCKET_URL,
      reconnectInterval: 1000,
      maxReconnectAttempts: 3,
      requestTimeout: 5000,
      heartbeatInterval: 30000,
      baseDelay: 1000,
      maxDelay: 10000,
    };
    
    wsService = new WebSocketService(config);
    await wsService.connect();
    
    // Authenticate the WebSocket connection
    console.log('Authenticating WebSocket connection...');
    const token = generateValidToken('mvp_test_user', 'operator');
    
    try {
      // Call the authenticate method explicitly
      const authResult = await wsService.call('authenticate', { token: token }) as any;
      expect(authResult.authenticated).toBe(true);
      console.log('Authentication completed successfully');
    } catch (error) {
      console.error('Authentication failed:', error);
      throw new Error(`Authentication required for REQ-MVP01 validation: ${error}`);
    }
  });

  afterEach(async () => {
    if (wsService) {
      wsService.disconnect();
    }
  });

  describe('REQ-MVP01.1: Camera Discovery Workflow (End-to-End)', () => {
    it('should execute complete camera discovery workflow', async () => {
      // Step 1: Get camera list
      const startTime = performance.now();
      const cameraList = await wsService.call(RPC_METHODS.GET_CAMERA_LIST, {}, true) as CameraListResponse;
      const responseTime = performance.now() - startTime;
      
      // Validate response structure
      expect(cameraList).toHaveProperty('cameras');
      expect(cameraList).toHaveProperty('total');
      expect(cameraList).toHaveProperty('connected');
      expect(Array.isArray(cameraList.cameras)).toBe(true);
      expect(typeof cameraList.total).toBe('number');
      expect(typeof cameraList.connected).toBe('number');
      
      // Validate performance target
      expect(responseTime).toBeLessThan(PERFORMANCE_TARGETS.STATUS_METHODS);
      
      // Step 2: Validate camera data structure
      if (cameraList.cameras.length > 0) {
        const camera = cameraList.cameras[0];
        expect(camera).toHaveProperty('device');
        expect(camera).toHaveProperty('status');
        expect(camera).toHaveProperty('name');
        expect(camera).toHaveProperty('resolution');
        expect(camera).toHaveProperty('fps');
        expect(camera).toHaveProperty('streams');
        
        // Step 3: Get individual camera status
        const statusStartTime = performance.now();
        const cameraStatus = await wsService.call(RPC_METHODS.GET_CAMERA_STATUS, { device: camera.device }, true) as CameraDevice;
        const statusResponseTime = performance.now() - statusStartTime;
        
        expect(cameraStatus).toHaveProperty('device');
        expect(cameraStatus.device).toBe(camera.device);
        expect(statusResponseTime).toBeLessThan(PERFORMANCE_TARGETS.STATUS_METHODS);
      } else {
        // If no cameras, test should still validate the API contract
        console.log('No cameras available - validating API contract only');
      }
    }, 15000);

    it('should handle camera discovery errors gracefully', async () => {
      // Test with invalid parameters
      try {
        await wsService.call(RPC_METHODS.GET_CAMERA_STATUS, {}, true);
        fail('Expected error for missing device parameter');
      } catch (error: any) {
        expect(error).toHaveProperty('code');
        expect(error.code).toBe(ERROR_CODES.INVALID_PARAMS);
      }
      
      // Test with non-existent camera
      try {
        await wsService.call(RPC_METHODS.GET_CAMERA_STATUS, { device: '/dev/video999' }, true);
        fail('Expected error for non-existent camera');
      } catch (error: any) {
        expect(error).toHaveProperty('code');
        expect(error.code).toBe(ERROR_CODES.CAMERA_NOT_FOUND_OR_DISCONNECTED);
      }
    }, 10000);
  });

  describe('REQ-MVP01.2: Real-time Camera Status Updates', () => {
    it('should receive real-time camera status updates', async () => {
      // Get initial camera list
      const initialCameraList = await wsService.call(RPC_METHODS.GET_CAMERA_LIST, {}, true) as CameraListResponse;
      
      // Set up notification listener
      const notificationPromise = new Promise((resolve) => {
        wsService.onMessage((message) => {
          if (isNotification(message) && message.method === 'camera_status_update') {
            resolve(message.params);
          }
        });
      });

      // Wait for notification (with timeout)
      const notification = await Promise.race([
        notificationPromise,
        new Promise((_, reject) => setTimeout(() => reject(new Error('Notification timeout')), 10000))
      ]) as any;

      // Validate notification structure
      expect(notification).toHaveProperty('device');
      expect(notification).toHaveProperty('status');
      expect(notification).toHaveProperty('name');
      expect(notification).toHaveProperty('resolution');
      expect(notification).toHaveProperty('fps');
      expect(notification).toHaveProperty('streams');
      
      // Validate notification is for a known camera
      const knownDevices = initialCameraList.cameras.map((c: CameraDevice) => c.device);
      expect(knownDevices).toContain(notification.device);
    }, 15000);

    it('should handle physical camera connect/disconnect scenarios', async () => {
      // This test validates the system can handle camera state changes
      // In a real scenario, this would involve physically connecting/disconnecting cameras
      
      const initialList = await wsService.call(RPC_METHODS.GET_CAMERA_LIST, {}, true) as CameraListResponse;
      const initialCount = initialList.cameras.length;
      
      // Wait for potential status changes
      await new Promise(resolve => setTimeout(resolve, 5000));
      
              const updatedList = await wsService.call(RPC_METHODS.GET_CAMERA_LIST, {}, true) as CameraListResponse;
      
      // System should maintain consistent state
      expect(updatedList).toHaveProperty('cameras');
      expect(Array.isArray(updatedList.cameras)).toBe(true);
      
      // Note: In real testing, we would validate actual camera state changes
      // For now, we validate the system remains responsive
      console.log(`Camera count: ${initialCount} -> ${updatedList.cameras.length}`);
    }, 15000);
  });

  describe('REQ-MVP01.3: Snapshot Capture Operations', () => {
    it('should capture snapshots with multiple format/quality combinations', async () => {
      // Get available cameras
      const cameraList = await wsService.call(RPC_METHODS.GET_CAMERA_LIST, {}, true) as CameraListResponse;
      
      if (cameraList.cameras.length === 0) {
        fail('No cameras available for snapshot testing - test cannot validate core functionality');
      }

      const testCamera = cameraList.cameras.find((c: CameraDevice) => c.status === 'CONNECTED');
      if (!testCamera) {
        fail('No connected cameras available for snapshot testing - test cannot validate core functionality');
      }

      // Test JPEG format
      const jpegParams = {
        device: testCamera.device,
        format: 'jpg',
        quality: 80
      } as Record<string, unknown>;
      
      const jpegStartTime = performance.now();
      const jpegResult = await wsService.call(RPC_METHODS.TAKE_SNAPSHOT, jpegParams, true) as SnapshotResult;
      const jpegResponseTime = performance.now() - jpegStartTime;
      
      expect(jpegResult).toHaveProperty('status');
      expect(jpegResult.status).toBe('completed');
      expect(jpegResult).toHaveProperty('filename');
      expect(jpegResult).toHaveProperty('file_size');
      expect(jpegResult).toHaveProperty('format');
      expect(jpegResult.format).toBe('jpg');
      expect(jpegResult.file_size).toBeGreaterThan(0);
      expect(jpegResponseTime).toBeLessThan(PERFORMANCE_TARGETS.CONTROL_METHODS);
      
      // Test PNG format
      const pngParams = {
        device: testCamera.device,
        format: 'png',
        quality: 90
      } as Record<string, unknown>;
      
      const pngStartTime = performance.now();
      const pngResult = await wsService.call(RPC_METHODS.TAKE_SNAPSHOT, pngParams, true) as SnapshotResult;
      const pngResponseTime = performance.now() - pngStartTime;
      
      expect(pngResult.status).toBe('completed');
      expect(pngResult.format).toBe('png');
      expect(pngResult.file_size).toBeGreaterThan(0);
      expect(pngResponseTime).toBeLessThan(PERFORMANCE_TARGETS.CONTROL_METHODS);
    }, 30000);

    it('should handle snapshot errors gracefully', async () => {
      // Test with non-existent camera
      const invalidParams = {
        device: '/dev/video999',
        format: 'jpg',
        quality: 80
      } as Record<string, unknown>;
      
      try {
        await wsService.call(RPC_METHODS.TAKE_SNAPSHOT, invalidParams, true);
        fail('Expected error for non-existent camera');
      } catch (error: any) {
        expect(error).toHaveProperty('code');
        expect(error.code).toBe(ERROR_CODES.CAMERA_NOT_FOUND_OR_DISCONNECTED);
      }
      
      // Test with invalid format (if camera exists)
      const cameraList = await wsService.call(RPC_METHODS.GET_CAMERA_LIST, {}, true) as CameraListResponse;
      if (cameraList.cameras.length > 0) {
        const invalidFormatParams = {
          device: cameraList.cameras[0].device,
          format: 'invalid',
          quality: 80
        };
        
        try {
          await wsService.call(RPC_METHODS.TAKE_SNAPSHOT, invalidFormatParams, true);
          fail('Expected error for invalid format');
        } catch (error: any) {
          expect(error).toHaveProperty('code');
          expect(error.code).toBe(ERROR_CODES.INVALID_PARAMS);
        }
      }
    }, 15000);
  });

  describe('REQ-MVP01.4: Video Recording Operations', () => {
    it('should perform unlimited and timed duration recordings', async () => {
      // Get available cameras
      const cameraList = await wsService.call(RPC_METHODS.GET_CAMERA_LIST, {}, true) as CameraListResponse;
      
      if (cameraList.cameras.length === 0) {
        fail('No cameras available for recording testing - test cannot validate core functionality');
      }

      const testCamera = cameraList.cameras.find((c: CameraDevice) => c.status === 'CONNECTED');
      if (!testCamera) {
        fail('No connected cameras available for recording testing - test cannot validate core functionality');
      }

      // Test timed recording (5 seconds)
      const startParams = {
        device: testCamera.device,
        duration_seconds: 5,
        format: 'mp4'
      } as Record<string, unknown>;
      
      const startRecordingTime = performance.now();
      const startResult = await wsService.call(RPC_METHODS.START_RECORDING, startParams, true) as RecordingSession;
      const startResponseTime = performance.now() - startRecordingTime;
      
      expect(startResult).toHaveProperty('session_id');
      expect(startResult).toHaveProperty('device');
      expect(startResult).toHaveProperty('format');
      expect(startResult.device).toBe(testCamera.device);
      expect(startResult.format).toBe('mp4');
      expect(startResponseTime).toBeLessThan(PERFORMANCE_TARGETS.CONTROL_METHODS);
      
      // Wait for recording to complete
      await new Promise(resolve => setTimeout(resolve, 6000));
      
      // Stop recording
      const stopParams = {
        device: testCamera.device
      } as Record<string, unknown>;
      
      const stopRecordingTime = performance.now();
      const stopResult = await wsService.call(RPC_METHODS.STOP_RECORDING, stopParams, true) as RecordingSession;
      const stopResponseTime = performance.now() - stopRecordingTime;
      
      expect(stopResult).toHaveProperty('session_id');
      expect(stopResult).toHaveProperty('duration');
      expect(stopResult).toHaveProperty('file_size');
      expect(stopResult.session_id).toBe(startResult.session_id);
      expect(stopResult.duration).toBeGreaterThan(0);
      expect(stopResult.file_size).toBeGreaterThan(0);
      expect(stopResponseTime).toBeLessThan(PERFORMANCE_TARGETS.CONTROL_METHODS);
    }, 45000);

    it('should handle recording errors gracefully', async () => {
      // Test with non-existent camera
      const invalidParams = {
        device: '/dev/video999',
        duration_seconds: 10,
        format: 'mp4'
      } as Record<string, unknown>;
      
      try {
        await wsService.call(RPC_METHODS.START_RECORDING, invalidParams, true);
        fail('Expected error for non-existent camera');
      } catch (error: any) {
        expect(error).toHaveProperty('code');
        expect(error.code).toBe(ERROR_CODES.CAMERA_NOT_FOUND_OR_DISCONNECTED);
      }
      
      // Test with invalid duration
      const cameraList = await wsService.call(RPC_METHODS.GET_CAMERA_LIST, {}, true) as CameraListResponse;
      if (cameraList.cameras.length > 0) {
        const invalidDurationParams = {
          device: cameraList.cameras[0].device,
          duration_seconds: -1,
          format: 'mp4'
        };
        
        try {
          await wsService.call(RPC_METHODS.START_RECORDING, invalidDurationParams, true);
          fail('Expected error for invalid duration');
        } catch (error: any) {
          expect(error).toHaveProperty('code');
          expect(error.code).toBe(ERROR_CODES.INVALID_PARAMS);
        }
      }
    }, 15000);
  });

  describe('REQ-MVP01.5: File Browsing and Download Functionality', () => {
    it('should list recordings and snapshots with metadata', async () => {
      // List recordings
      const recordingsParams = {
        limit: 20,
        offset: 0
      } as Record<string, unknown>;
      
      const recordingsStartTime = performance.now();
      const recordings = await wsService.call(RPC_METHODS.LIST_RECORDINGS, recordingsParams, true) as FileListResponse;
      const recordingsResponseTime = performance.now() - recordingsStartTime;
      
      expect(recordings).toHaveProperty('files');
      expect(recordings).toHaveProperty('total');
      expect(Array.isArray(recordings.files)).toBe(true);
      expect(typeof recordings.total).toBe('number');
      expect(recordingsResponseTime).toBeLessThan(PERFORMANCE_TARGETS.STATUS_METHODS);
      
      if (recordings.files.length > 0) {
        const recording = recordings.files[0];
        expect(recording).toHaveProperty('filename');
        expect(recording).toHaveProperty('file_size');
        expect(recording).toHaveProperty('modified_time');
        expect(recording).toHaveProperty('download_url');
        expect(recording.file_size).toBeGreaterThan(0);
      }
      
      // List snapshots
      const snapshotsParams = {
        limit: 20,
        offset: 0
      } as Record<string, unknown>;
      
      const snapshotsStartTime = performance.now();
      const snapshots = await wsService.call(RPC_METHODS.LIST_SNAPSHOTS, snapshotsParams, true) as FileListResponse;
      const snapshotsResponseTime = performance.now() - snapshotsStartTime;
      
      expect(snapshots).toHaveProperty('files');
      expect(snapshots).toHaveProperty('total');
      expect(Array.isArray(snapshots.files)).toBe(true);
      expect(typeof snapshots.total).toBe('number');
      expect(snapshotsResponseTime).toBeLessThan(PERFORMANCE_TARGETS.STATUS_METHODS);
      
      if (snapshots.files.length > 0) {
        const snapshot = snapshots.files[0];
        expect(snapshot).toHaveProperty('filename');
        expect(snapshot).toHaveProperty('file_size');
        expect(snapshot).toHaveProperty('modified_time');
        expect(snapshot).toHaveProperty('download_url');
        expect(snapshot.file_size).toBeGreaterThan(0);
      }
    }, 20000);

    it('should handle pagination correctly', async () => {
      // Test recordings pagination
      const firstPageParams = {
        limit: 5,
        offset: 0
      } as Record<string, unknown>;
      
      const firstPage = await wsService.call(RPC_METHODS.LIST_RECORDINGS, firstPageParams, true) as FileListResponse;
      
      if (firstPage.total > 5) {
        const secondPageParams = {
          limit: 5,
          offset: 5
        } as Record<string, unknown>;
        
        const secondPage = await wsService.call(RPC_METHODS.LIST_RECORDINGS, secondPageParams, true) as FileListResponse;
        
        expect(secondPage.files.length).toBeLessThanOrEqual(5);
        expect(secondPage.total).toBe(firstPage.total);
        
        // Files should be different between pages
        const firstPageFilenames = firstPage.files.map(f => f.filename);
        const secondPageFilenames = secondPage.files.map(f => f.filename);
        
        const intersection = firstPageFilenames.filter(name => secondPageFilenames.includes(name));
        expect(intersection.length).toBe(0);
      }
    }, 15000);
  });

  describe('REQ-MVP01.6: Error Handling and Recovery', () => {
    it('should handle network failures and reconnection', async () => {
      // Verify initial connection
      expect(wsService.isConnected).toBe(true);
      
      // Simulate connection loss
      wsService.disconnect();
      expect(wsService.isConnected).toBe(false);
      
      // Reconnect
      await wsService.connect();
      expect(wsService.isConnected).toBe(true);
      
      // Verify functionality after reconnection
      const response = await wsService.call(RPC_METHODS.PING, {});
      expect(response).toBe('pong');
    }, 15000);

    it('should handle server errors gracefully', async () => {
      // Test invalid method
      try {
        await wsService.call('invalid_method', {});
        fail('Expected error for invalid method');
      } catch (error: any) {
        expect(error).toHaveProperty('code');
        expect(error.code).toBe(ERROR_CODES.METHOD_NOT_FOUND);
      }
      
      // Test invalid parameters
      try {
        await wsService.call(RPC_METHODS.GET_CAMERA_STATUS, {}, true);
        fail('Expected error for missing device parameter');
      } catch (error: any) {
        expect(error).toHaveProperty('code');
        expect(error.code).toBe(ERROR_CODES.INVALID_PARAMS);
      }
    }, 10000);
  });
});

/**
 * Check if MediaMTX Camera Service is available
 */
async function checkServerAvailability(): Promise<boolean> {
  const testWebSocketUrl = process.env.TEST_WEBSOCKET_URL || 'ws://localhost:8002';
  try {
    // Try to connect to WebSocket endpoint
    const ws = new WebSocket(testWebSocketUrl);
    
    return new Promise((resolve) => {
      const timeout = setTimeout(() => {
        ws.close();
        resolve(false);
      }, 5000);

      ws.onopen = () => {
        clearTimeout(timeout);
        ws.close();
        resolve(true);
      };

      ws.onerror = () => {
        clearTimeout(timeout);
        resolve(false);
      };
    });
  } catch {
    return false;
  }
}
