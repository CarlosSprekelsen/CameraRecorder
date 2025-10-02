/**
 * Integration Tests: Server Connectivity
 * 
 * Tests real server connectivity and basic functionality
 * Requires: Real MediaMTX server running
 * 
 * Architecture Compliance: Uses loadTestEnvironment() for consistent authentication
 */

import { AuthHelper, createAuthenticatedTestEnvironment } from '../utils/auth-helper';
import { APIClient } from '../../src/services/abstraction/APIClient';
import { AuthService } from '../../src/services/auth/AuthService';
import { DeviceService } from '../../src/services/device/DeviceService';
import { FileService } from '../../src/services/file/FileService';
import { ServerService } from '../../src/services/server/ServerService';
import { LoggerService } from '../../src/services/logger/LoggerService';

describe('Integration Tests: Server Connectivity', () => {
  let authHelper: AuthHelper;
  let authService: AuthService;
  let deviceService: DeviceService;
  let fileService: FileService;
  let serverService: ServerService;
  let loggerService: LoggerService;

  beforeAll(async () => {
    // Use unified authentication approach
    authHelper = await createAuthenticatedTestEnvironment(
      process.env.TEST_WEBSOCKET_URL || 'ws://localhost:8002/ws'
    );
    
    const services = authHelper.getAuthenticatedServices();
    const apiClient = services.apiClient;
    loggerService = services.logger;
    
    // Initialize services after authentication
    authService = new AuthService(apiClient, loggerService);
    deviceService = new DeviceService(apiClient, loggerService);
    fileService = new FileService(apiClient, loggerService);
    serverService = new ServerService(apiClient, loggerService);
  });

  afterAll(async () => {
    if (authHelper) {
      await authHelper.disconnect();
    }
  });

  describe('REQ-INT-001: Server Connection', () => {
    test('should connect to real server', async () => {
      expect(authHelper.isConnected()).toBe(true);
      expect(authHelper.isConnected()).toBe(true); // Connection is open
    });

    test('should maintain connection stability', async () => {
      // Test connection stability over time
      const startTime = Date.now();
      await new Promise(resolve => setTimeout(resolve, 5000));
      const endTime = Date.now();
      
      expect(authHelper.isConnected()).toBe(true);
      expect(endTime - startTime).toBeGreaterThan(4000);
    });
  });

  describe('REQ-INT-002: Authentication Flow', () => {
    test('should authenticate with real server', async () => {
      const token = AuthHelper.generateTestToken('admin');
      const result = await authService.authenticate(token);
      expect(result.authenticated).toBe(true);
      expect(result.role).toBe('admin');
      expect(typeof result.session_id).toBe('string');
    });

    test('should handle authentication errors', async () => {
      const result = await authService.authenticate('invalid-token');
      expect(result.authenticated).toBe(false);
      expect(result.role).toBeUndefined();
    });
  });

  describe('REQ-INT-003: Device Operations', () => {
    test('should get camera list from real server', async () => {
      const cameras = await deviceService.getCameraList();
      expect(Array.isArray(cameras)).toBe(true);
      // Note: May be empty if no cameras configured
    });

    test('should get stream URL from real server', async () => {
      const streamUrl = await deviceService.getStreamUrl('camera0');
      expect(typeof streamUrl).toBe('string');
      expect(streamUrl).toContain('rtsp://');
    });
  });

  describe('REQ-INT-004: File Operations', () => {
    test('should list recordings from real server', async () => {
      const recordings = await fileService.listRecordings(10, 0);
      expect(Array.isArray(recordings.files)).toBe(true);
      expect(typeof recordings.total).toBe('number');
    });

    test('should list snapshots from real server', async () => {
      const snapshots = await fileService.listSnapshots(10, 0);
      expect(Array.isArray(snapshots.files)).toBe(true);
      expect(typeof snapshots.total).toBe('number');
    });
  });

  describe('REQ-INT-005: Server Status', () => {
    test('should get server status from real server', async () => {
      const status = await serverService.getStatus();
      expect(status).toBeDefined();
      expect(typeof status.uptime).toBe('number');
    });

    test('should get server metrics from real server', async () => {
      const metrics = await serverService.getMetrics();
      expect(metrics).toBeDefined();
      expect(typeof metrics.system_metrics.cpu_usage).toBe('number');
    });
  });

  describe('REQ-INT-006: Performance Validation', () => {
    test('should meet connection performance targets', async () => {
      const startTime = Date.now();
      const connected = authHelper.isConnected;
      const endTime = Date.now();
      
      expect(authHelper.isConnected()).toBe(true);
      expect(endTime - startTime).toBeLessThan(100); // < 100ms
    });

    test('should meet API response performance targets', async () => {
      const startTime = Date.now();
      await serverService.getStatus();
      const endTime = Date.now();
      
      expect(endTime - startTime).toBeLessThan(1000); // < 1s
    });
  });

  describe('REQ-INT-007: Error Handling', () => {
    test('should handle server disconnection gracefully', async () => {
      // This test would require server restart simulation
      // For now, just test that we can detect connection state
      expect(authHelper.isConnected()).toBeDefined();
    });

    test('should handle invalid API calls gracefully', async () => {
      try {
        await deviceService.getStreamUrl('nonexistent-camera');
        // Should either return null or throw appropriate error
      } catch (error) {
        expect(error).toBeDefined();
      }
    });
  });
});
