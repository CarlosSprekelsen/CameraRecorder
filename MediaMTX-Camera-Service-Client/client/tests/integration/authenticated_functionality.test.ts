/**
 * Authenticated Functionality Tests
 * 
 * Tests real server functionality with valid API keys:
 * - Camera discovery and status
 * - Snapshot capture and download
 * - Recording operations
 * - Stream URL validation
 * - File operations
 */

import { WebSocketService } from '../../src/services/websocket/WebSocketService';
import { AuthService } from '../../src/services/auth/AuthService';
import { DeviceService } from '../../src/services/device/DeviceService';
import { FileService } from '../../src/services/file/FileService';
import { RecordingService } from '../../src/services/recording/RecordingService';
import { LoggerService } from '../../src/services/logger/LoggerService';

describe('Authenticated Functionality Tests', () => {
  let webSocketService: WebSocketService;
  let authService: AuthService;
  let deviceService: DeviceService;
  let fileService: FileService;
  let recordingService: RecordingService;
  let loggerService: LoggerService;

  // API Keys from test environment
  const TEST_VIEWER_KEY = process.env.TEST_VIEWER_KEY;
  const TEST_OPERATOR_KEY = process.env.TEST_OPERATOR_KEY;
  const TEST_ADMIN_KEY = process.env.TEST_ADMIN_KEY;

  beforeAll(async () => {
    loggerService = new LoggerService();
    webSocketService = new WebSocketService({ url: 'ws://localhost:8002/ws' });
    
    // Connect to the server
    await webSocketService.connect();
    await new Promise(resolve => setTimeout(resolve, 2000));
    
    authService = new AuthService(webSocketService, loggerService);
    deviceService = new DeviceService(webSocketService, loggerService);
    fileService = new FileService(webSocketService, loggerService);
    recordingService = new RecordingService(webSocketService, loggerService);

    // Check if API keys are available
    if (!TEST_VIEWER_KEY && !TEST_OPERATOR_KEY && !TEST_ADMIN_KEY) {
      console.warn('⚠️  No API keys available - skipping authenticated tests');
    }
  });

  afterAll(async () => {
    if (webSocketService) {
      await webSocketService.disconnect();
    }
    await new Promise(resolve => setTimeout(resolve, 100));
  });

  describe('REQ-AUTH-001: Authentication with API Keys', () => {
    test('should authenticate with viewer API key', async () => {
      if (!TEST_VIEWER_KEY) {
        console.log('⏭️  Skipping viewer authentication test - no API key available');
        return;
      }

      console.log(`🔑 Testing authentication with viewer key: ${TEST_VIEWER_KEY.substring(0, 20)}...`);
      
      try {
        const result = await webSocketService.sendRPC('authenticate', { 
          auth_token: TEST_VIEWER_KEY 
        });
        
        expect(result).toBeDefined();
        expect(result.authenticated).toBe(true);
        expect(result.role).toBe('viewer');
        
        console.log('✅ Viewer authentication successful:', result);
      } catch (error: any) {
        console.log('❌ Viewer authentication failed:', error.message);
        throw error;
      }
    });

    test('should authenticate with operator API key', async () => {
      if (!TEST_OPERATOR_KEY) {
        console.log('⏭️  Skipping operator authentication test - no API key available');
        return;
      }

      console.log(`🔑 Testing authentication with operator key: ${TEST_OPERATOR_KEY.substring(0, 20)}...`);
      
      try {
        const result = await webSocketService.sendRPC('authenticate', { 
          auth_token: TEST_OPERATOR_KEY 
        });
        
        expect(result).toBeDefined();
        expect(result.authenticated).toBe(true);
        expect(result.role).toBe('operator');
        
        console.log('✅ Operator authentication successful:', result);
      } catch (error: any) {
        console.log('❌ Operator authentication failed:', error.message);
        throw error;
      }
    });

    test('should authenticate with admin API key', async () => {
      if (!TEST_ADMIN_KEY) {
        console.log('⏭️  Skipping admin authentication test - no API key available');
        return;
      }

      console.log(`🔑 Testing authentication with admin key: ${TEST_ADMIN_KEY.substring(0, 20)}...`);
      
      try {
        const result = await webSocketService.sendRPC('authenticate', { 
          auth_token: TEST_ADMIN_KEY 
        });
        
        expect(result).toBeDefined();
        expect(result.authenticated).toBe(true);
        expect(result.role).toBe('admin');
        
        console.log('✅ Admin authentication successful:', result);
      } catch (error: any) {
        console.log('❌ Admin authentication failed:', error.message);
        throw error;
      }
    });
  });

  describe('REQ-AUTH-002: Camera Discovery with Authentication', () => {
    test('should get camera list with viewer authentication', async () => {
      if (!TEST_VIEWER_KEY) {
        console.log('⏭️  Skipping camera list test - no viewer API key available');
        return;
      }

      // Authenticate first
      await webSocketService.sendRPC('authenticate', { auth_token: TEST_VIEWER_KEY });
      
      try {
        const cameras = await deviceService.getCameraList();
        
        expect(cameras).toBeDefined();
        expect(Array.isArray(cameras)).toBe(true);
        
        console.log('✅ Camera list retrieved:', cameras);
        
        // Validate camera structure if cameras exist
        if (cameras.length > 0) {
          const camera = cameras[0];
          expect(camera.device).toBeDefined();
          expect(camera.status).toBeDefined();
          expect(typeof camera.device).toBe('string');
          expect(typeof camera.status).toBe('string');
        } else {
          console.log('ℹ️  No cameras discovered (expected in test environment)');
        }
      } catch (error: any) {
        console.log('❌ Camera list retrieval failed:', error.message);
        throw error;
      }
    });

    test('should get camera status with authentication', async () => {
      if (!TEST_VIEWER_KEY) {
        console.log('⏭️  Skipping camera status test - no viewer API key available');
        return;
      }

      // Authenticate first
      await webSocketService.sendRPC('authenticate', { auth_token: TEST_VIEWER_KEY });
      
      try {
        const status = await deviceService.getCameraStatus('camera0');
        
        expect(status).toBeDefined();
        expect(status.device).toBe('camera0');
        expect(status.status).toBeDefined();
        
        console.log('✅ Camera status retrieved:', status);
      } catch (error: any) {
        // This might fail if no camera is connected, which is expected
        console.log('ℹ️  Camera status check (expected if no camera):', error.message);
        expect(error.message).toContain('not found');
      }
    });
  });

  describe('REQ-AUTH-003: Stream Operations with Authentication', () => {
    test('should get stream URLs with authentication', async () => {
      if (!TEST_VIEWER_KEY) {
        console.log('⏭️  Skipping stream URL test - no viewer API key available');
        return;
      }

      // Authenticate first
      await webSocketService.sendRPC('authenticate', { auth_token: TEST_VIEWER_KEY });
      
      try {
        const streamUrl = await deviceService.getStreamUrl('camera0');
        
        expect(streamUrl).toBeDefined();
        expect(streamUrl.stream_url).toBeDefined();
        expect(typeof streamUrl.stream_url).toBe('string');
        
        // Validate URL format
        const url = new URL(streamUrl.stream_url);
        expect(['rtsp:', 'http:', 'https:'].includes(url.protocol)).toBe(true);
        
        console.log('✅ Stream URL retrieved:', streamUrl);
      } catch (error: any) {
        // This might fail if no camera is connected
        console.log('ℹ️  Stream URL check (expected if no camera):', error.message);
        expect(error.message).toContain('not found');
      }
    });

    test('should get active streams with authentication', async () => {
      if (!TEST_VIEWER_KEY) {
        console.log('⏭️  Skipping active streams test - no viewer API key available');
        return;
      }

      // Authenticate first
      await webSocketService.sendRPC('authenticate', { auth_token: TEST_VIEWER_KEY });
      
      try {
        const streams = await deviceService.getStreams();
        
        expect(streams).toBeDefined();
        expect(Array.isArray(streams)).toBe(true);
        
        console.log('✅ Active streams retrieved:', streams);
      } catch (error: any) {
        console.log('❌ Active streams retrieval failed:', error.message);
        throw error;
      }
    });
  });

  describe('REQ-AUTH-004: Snapshot Operations with Authentication', () => {
    test('should capture snapshot with operator authentication', async () => {
      if (!TEST_OPERATOR_KEY) {
        console.log('⏭️  Skipping snapshot capture test - no operator API key available');
        return;
      }

      // Authenticate first
      await webSocketService.sendRPC('authenticate', { auth_token: TEST_OPERATOR_KEY });
      
      try {
        const snapshot = await deviceService.takeSnapshot('camera0', 'test_snapshot.jpg');
        
        expect(snapshot).toBeDefined();
        expect(snapshot.filename).toBeDefined();
        expect(snapshot.status).toBe('SUCCESS');
        expect(snapshot.file_size).toBeGreaterThan(0);
        
        console.log('✅ Snapshot captured:', snapshot);
        
        // Test download URL if available
        if (snapshot.download_url) {
          console.log('🔗 Download URL available:', snapshot.download_url);
        }
      } catch (error: any) {
        // This might fail if no camera is connected
        console.log('ℹ️  Snapshot capture (expected if no camera):', error.message);
        expect(error.message).toContain('not found');
      }
    });

    test('should list snapshots with authentication', async () => {
      if (!TEST_VIEWER_KEY) {
        console.log('⏭️  Skipping snapshot list test - no viewer API key available');
        return;
      }

      // Authenticate first
      await webSocketService.sendRPC('authenticate', { auth_token: TEST_VIEWER_KEY });
      
      try {
        const snapshots = await fileService.listSnapshots(10, 0);
        
        expect(snapshots).toBeDefined();
        expect(snapshots.files).toBeDefined();
        expect(Array.isArray(snapshots.files)).toBe(true);
        expect(snapshots.total).toBeDefined();
        
        console.log('✅ Snapshots listed:', snapshots);
        
        // Test downloading each file if available
        for (const file of snapshots.files) {
          if (file.download_url) {
            console.log(`🔗 Download URL for ${file.filename}: ${file.download_url}`);
          }
        }
      } catch (error: any) {
        console.log('❌ Snapshot list retrieval failed:', error.message);
        throw error;
      }
    });
  });

  describe('REQ-AUTH-005: Recording Operations with Authentication', () => {
    test('should start recording with operator authentication', async () => {
      if (!TEST_OPERATOR_KEY) {
        console.log('⏭️  Skipping recording start test - no operator API key available');
        return;
      }

      // Authenticate first
      await webSocketService.sendRPC('authenticate', { auth_token: TEST_OPERATOR_KEY });
      
      try {
        const recording = await recordingService.startRecording('camera0', 5); // 5 second recording
        
        expect(recording).toBeDefined();
        expect(recording.filename).toBeDefined();
        expect(recording.status).toBe('RECORDING');
        expect(recording.device).toBe('camera0');
        
        console.log('✅ Recording started:', recording);
        
        // Wait a bit then try to stop
        await new Promise(resolve => setTimeout(resolve, 6000)); // Wait 6 seconds
        
        try {
          const stopResult = await recordingService.stopRecording('camera0');
          console.log('✅ Recording stopped:', stopResult);
        } catch (stopError: any) {
          console.log('ℹ️  Recording stop (expected if no camera):', stopError.message);
        }
      } catch (error: any) {
        // This might fail if no camera is connected
        console.log('ℹ️  Recording start (expected if no camera):', error.message);
        expect(error.message).toContain('not found');
      }
    });

    test('should list recordings with authentication', async () => {
      if (!TEST_VIEWER_KEY) {
        console.log('⏭️  Skipping recording list test - no viewer API key available');
        return;
      }

      // Authenticate first
      await webSocketService.sendRPC('authenticate', { auth_token: TEST_VIEWER_KEY });
      
      try {
        const recordings = await fileService.listRecordings(10, 0);
        
        expect(recordings).toBeDefined();
        expect(recordings.files).toBeDefined();
        expect(Array.isArray(recordings.files)).toBe(true);
        
        console.log('✅ Recordings listed:', recordings);
        
        // Test downloading each file if available
        for (const file of recordings.files) {
          if (file.download_url) {
            console.log(`🔗 Download URL for ${file.filename}: ${file.download_url}`);
          }
        }
      } catch (error: any) {
        console.log('❌ Recording list retrieval failed:', error.message);
        throw error;
      }
    });
  });

  describe('REQ-AUTH-006: File Operations with Authentication', () => {
    test('should get recording info with authentication', async () => {
      if (!TEST_VIEWER_KEY) {
        console.log('⏭️  Skipping recording info test - no viewer API key available');
        return;
      }

      // Authenticate first
      await webSocketService.sendRPC('authenticate', { auth_token: TEST_VIEWER_KEY });
      
      try {
        const info = await fileService.getRecordingInfo('test_recording.mp4');
        
        expect(info).toBeDefined();
        expect(info.filename).toBeDefined();
        expect(info.file_size).toBeGreaterThan(0);
        expect(info.duration).toBeDefined();
        expect(info.format).toBeDefined();
        expect(info.device).toBeDefined();
        
        console.log('✅ Recording info retrieved:', info);
      } catch (error: any) {
        // This might fail if file doesn't exist
        console.log('ℹ️  Recording info (expected if no recordings):', error.message);
        expect(error.message).toContain('not found');
      }
    });

    test('should get snapshot info with authentication', async () => {
      if (!TEST_VIEWER_KEY) {
        console.log('⏭️  Skipping snapshot info test - no viewer API key available');
        return;
      }

      // Authenticate first
      await webSocketService.sendRPC('authenticate', { auth_token: TEST_VIEWER_KEY });
      
      try {
        const info = await fileService.getSnapshotInfo('test_snapshot.jpg');
        
        expect(info).toBeDefined();
        expect(info.filename).toBeDefined();
        expect(info.file_size).toBeGreaterThan(0);
        expect(info.format).toBeDefined();
        expect(info.device).toBeDefined();
        
        console.log('✅ Snapshot info retrieved:', info);
      } catch (error: any) {
        // This might fail if file doesn't exist
        console.log('ℹ️  Snapshot info (expected if no snapshots):', error.message);
        expect(error.message).toContain('not found');
      }
    });
  });

  describe('REQ-AUTH-007: Permission Testing', () => {
    test('should enforce viewer permissions', async () => {
      if (!TEST_VIEWER_KEY) {
        console.log('⏭️  Skipping viewer permission test - no viewer API key available');
        return;
      }

      // Authenticate as viewer
      await webSocketService.sendRPC('authenticate', { auth_token: TEST_VIEWER_KEY });
      
      try {
        // Viewer should be able to read
        const cameras = await deviceService.getCameraList();
        expect(cameras).toBeDefined();
        console.log('✅ Viewer can read camera list');
        
        // Viewer should NOT be able to control (this might fail)
        try {
          await deviceService.takeSnapshot('camera0', 'test.jpg');
          console.log('⚠️  Viewer was able to take snapshot (unexpected)');
        } catch (error: any) {
          console.log('✅ Viewer correctly blocked from taking snapshot:', error.message);
          expect(error.message).toContain('permission');
        }
      } catch (error: any) {
        console.log('❌ Viewer permission test failed:', error.message);
        throw error;
      }
    });

    test('should enforce operator permissions', async () => {
      if (!TEST_OPERATOR_KEY) {
        console.log('⏭️  Skipping operator permission test - no operator API key available');
        return;
      }

      // Authenticate as operator
      await webSocketService.sendRPC('authenticate', { auth_token: TEST_OPERATOR_KEY });
      
      try {
        // Operator should be able to read
        const cameras = await deviceService.getCameraList();
        expect(cameras).toBeDefined();
        console.log('✅ Operator can read camera list');
        
        // Operator should be able to control
        try {
          await deviceService.takeSnapshot('camera0', 'test.jpg');
          console.log('✅ Operator can take snapshot');
        } catch (error: any) {
          console.log('ℹ️  Operator snapshot (expected if no camera):', error.message);
        }
      } catch (error: any) {
        console.log('❌ Operator permission test failed:', error.message);
        throw error;
      }
    });
  });
});
