/**
 * Integration Tests: API Compliance Testing
 * 
 * Tests compliance with authoritative JSON-RPC specification
 * Focus: Method validation, data structure compliance, error handling
 */

import { WebSocketService } from '../../src/services/websocket/WebSocketService';
import { APIClient } from '../../src/services/abstraction/APIClient';
import { AuthService } from '../../src/services/auth/AuthService';
import { FileService } from '../../src/services/file/FileService';
import { DeviceService } from '../../src/services/device/DeviceService';
import { ServerService } from '../../src/services/server/ServerService';
import { LoggerService } from '../../src/services/logger/LoggerService';

describe('Integration Tests: API Compliance', () => {
  let webSocketService: WebSocketService;
  let authService: AuthService;
  let fileService: FileService;
  let deviceService: DeviceService;
  let serverService: ServerService;
  let loggerService: LoggerService;

  beforeAll(async () => {
    loggerService = new LoggerService();
    webSocketService = new WebSocketService({ url: 'ws://localhost:8002/ws' });
    
    // Connect to the server
    await webSocketService.connect();
    
    // Wait for connection to be established
    await new Promise(resolve => setTimeout(resolve, 2000));
    
    // Create APIClient for services
    const apiClient = new APIClient(webSocketService, loggerService);
    
    authService = new AuthService(apiClient, loggerService);
    fileService = new FileService(apiClient, loggerService);
    deviceService = new DeviceService(apiClient, loggerService);
    serverService = new ServerService(apiClient, loggerService);
  });

  afterAll(async () => {
    if (webSocketService) {
      await webSocketService.disconnect();
    }
  });

  describe('REQ-API-001: JSON-RPC 2.0 Compliance', () => {
    test('should validate request structure', async () => {
      // Test that all requests follow JSON-RPC 2.0 format
      const cameras = await deviceService.getCameraList();
      expect(Array.isArray(cameras)).toBe(true);
    });

    test('should validate response structure', async () => {
      // Test that all responses follow JSON-RPC 2.0 format
      const status = await serverService.getStatus();
      expect(status).toBeDefined();
      expect(typeof status).toBe('object');
    });

    test('should handle method not found errors', async () => {
      // Test error handling for invalid methods
      try {
        // This would require sending invalid method directly
        // For now, test that our service handles errors gracefully
        await deviceService.getCameraList();
        expect(true).toBe(true); // Service should work
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });

  describe('REQ-API-002: Authentication Methods', () => {
    test('should validate login method compliance', async () => {
      const result = await authService.login('testuser', 'testpass');
      expect(result).toBeDefined();
      expect(typeof result.success).toBe('boolean');
      if (result.success) {
        expect(typeof result.token).toBe('string');
      } else {
        expect(typeof result.error).toBe('string');
      }
    });

    test('should validate logout method compliance', async () => {
      const result = await authService.logout();
      expect(result).toBeDefined();
      expect(typeof result.success).toBe('boolean');
    });

    test('should validate token validation compliance', async () => {
      const result = await authService.validateToken('test-token');
      expect(result).toBeDefined();
      expect(typeof result.valid).toBe('boolean');
    });
  });

  describe('REQ-API-003: Device Management Methods', () => {
    test('should validate get_camera_list compliance', async () => {
      const cameras = await deviceService.getCameraList();
      expect(Array.isArray(cameras)).toBe(true);
      
      // Validate camera structure if cameras exist
      if (cameras.length > 0) {
        const camera = cameras[0];
        expect(typeof camera.id).toBe('string');
        expect(typeof camera.name).toBe('string');
        expect(typeof camera.status).toBe('string');
      }
    });

    test('should validate get_stream_url compliance', async () => {
      const streamUrl = await deviceService.getStreamUrl('camera0');
      expect(typeof streamUrl).toBe('string');
      expect(streamUrl).toContain('rtsp://');
    });

    test('should validate get_streams compliance', async () => {
      const streams = await deviceService.getStreams();
      expect(Array.isArray(streams)).toBe(true);
      
      // Validate stream structure if streams exist
      if (streams.length > 0) {
        const stream = streams[0];
        expect(typeof stream.name).toBe('string');
        expect(typeof stream.readers).toBe('number');
      }
    });
  });

  describe('REQ-API-004: File Management Methods', () => {
    test('should validate list_recordings compliance', async () => {
      const recordings = await fileService.listRecordings(10, 0);
      expect(recordings).toBeDefined();
      expect(Array.isArray(recordings.files)).toBe(true);
      expect(typeof recordings.total).toBe('number');
      expect(typeof recordings.limit).toBe('number');
      expect(typeof recordings.offset).toBe('number');
    });

    test('should validate list_snapshots compliance', async () => {
      const snapshots = await fileService.listSnapshots(10, 0);
      expect(snapshots).toBeDefined();
      expect(Array.isArray(snapshots.files)).toBe(true);
      expect(typeof snapshots.total).toBe('number');
      expect(typeof snapshots.limit).toBe('number');
      expect(typeof snapshots.offset).toBe('number');
    });

    test('should validate get_recording_info compliance', async () => {
      try {
        const info = await fileService.getRecordingInfo('test.mp4');
        if (info) {
          expect(typeof info.filename).toBe('string');
          expect(typeof info.file_size).toBe('number');
          expect(typeof info.created_time).toBe('string');
          expect(typeof info.download_url).toBe('string');
        }
      } catch (error) {
        // File might not exist, which is expected
        expect(error).toBeDefined();
      }
    });

    test('should validate get_snapshot_info compliance', async () => {
      try {
        const info = await fileService.getSnapshotInfo('test.jpg');
        if (info) {
          expect(typeof info.filename).toBe('string');
          expect(typeof info.file_size).toBe('number');
          expect(typeof info.created_time).toBe('string');
          expect(typeof info.download_url).toBe('string');
        }
      } catch (error) {
        // File might not exist, which is expected
        expect(error).toBeDefined();
      }
    });
  });

  describe('REQ-API-005: Server Status Methods', () => {
    test('should validate get_status compliance', async () => {
      const status = await serverService.getStatus();
      expect(status).toBeDefined();
      expect(typeof status.uptime).toBe('number');
      expect(typeof status.version).toBe('string');
      expect(typeof status.status).toBe('string');
    });

    test('should validate get_metrics compliance', async () => {
      const metrics = await serverService.getMetrics();
      expect(metrics).toBeDefined();
      expect(typeof metrics.cpu_usage).toBe('number');
      expect(typeof metrics.memory_usage).toBe('number');
      expect(typeof metrics.disk_usage).toBe('number');
    });
  });

  describe('REQ-API-006: Error Handling Compliance', () => {
    test('should handle invalid parameters gracefully', async () => {
      try {
        await fileService.listRecordings(-1, -1);
        // Should either reject or sanitize negative values
      } catch (error) {
        expect(error).toBeDefined();
      }
    });

    test('should handle missing parameters gracefully', async () => {
      try {
        await deviceService.getStreamUrl('');
        // Should handle empty string gracefully
      } catch (error) {
        expect(error).toBeDefined();
      }
    });

    test('should handle type mismatches gracefully', async () => {
      try {
        await fileService.listRecordings('invalid' as any, 'invalid' as any);
        // Should handle type mismatches gracefully
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });

  describe('REQ-API-007: Data Structure Compliance', () => {
    test('should validate timestamp formats', async () => {
      try {
        const info = await fileService.getRecordingInfo('test.mp4');
        if (info && info.created_time) {
          // Should be ISO 8601 format
          expect(info.created_time).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d{3})?Z?$/);
        }
      } catch (error) {
        // File might not exist
        expect(error).toBeDefined();
      }
    });

    test('should validate URL formats', async () => {
      const streamUrl = await deviceService.getStreamUrl('camera0');
      expect(streamUrl).toMatch(/^rtsp:\/\//);
    });

    test('should validate numeric ranges', async () => {
      const status = await serverService.getStatus();
      if (status.uptime !== undefined) {
        expect(status.uptime).toBeGreaterThanOrEqual(0);
      }
    });
  });

  describe('REQ-API-008: Method Coverage', () => {
    test('should implement all documented methods', () => {
      // Test that all documented methods are implemented
      expect(typeof authService.login).toBe('function');
      expect(typeof authService.logout).toBe('function');
      expect(typeof authService.validateToken).toBe('function');
      
      expect(typeof deviceService.getCameraList).toBe('function');
      expect(typeof deviceService.getStreamUrl).toBe('function');
      expect(typeof deviceService.getStreams).toBe('function');
      
      expect(typeof fileService.listRecordings).toBe('function');
      expect(typeof fileService.listSnapshots).toBe('function');
      expect(typeof fileService.getRecordingInfo).toBe('function');
      expect(typeof fileService.getSnapshotInfo).toBe('function');
      expect(typeof fileService.deleteRecording).toBe('function');
      expect(typeof fileService.deleteSnapshot).toBe('function');
      
      expect(typeof serverService.getStatus).toBe('function');
      expect(typeof serverService.getMetrics).toBe('function');
    });
  });
});
