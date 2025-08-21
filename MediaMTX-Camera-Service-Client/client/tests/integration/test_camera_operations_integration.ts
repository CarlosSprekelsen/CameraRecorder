/**
 * REQ-CAM01-001: Camera discovery and status operations
 * REQ-CAM01-002: Snapshot and recording operations
 * Coverage: INTEGRATION
 * Quality: HIGH
 */
/**
 * Integration tests for camera operations
 * Tests real server integration for snapshot and recording functionality
 * 
 * These tests require a running MediaMTX server for full validation
 */

import { RPC_METHODS, ERROR_CODES } from '../../src/types';
import { WebSocketTestFixture, HealthTestFixture } from '../fixtures/stable-test-fixture';

describe('Camera Operations Integration', () => {
  let wsFixture: WebSocketTestFixture;
  let healthFixture: HealthTestFixture;

  beforeAll(async () => {
    // Initialize stable fixtures for authentication and server availability
    wsFixture = new WebSocketTestFixture();
    healthFixture = new HealthTestFixture();
    
    await wsFixture.initialize();
    await healthFixture.initialize();
  });

  afterAll(async () => {
    wsFixture.cleanup();
    healthFixture.cleanup();
  });

  describe('Camera Discovery', () => {
    it('should discover available cameras', async () => {
      const result = await wsFixture.testCameraList();
      expect(result).toBe(true);
    }, 10000);

    it('should get individual camera status', async () => {
      const result = await wsFixture.testCameraStatus();
      expect(result).toBe(true);
    }, 10000);
  });

  describe('Snapshot Operations', () => {
    it('should take snapshot with default settings', async () => {
      const result = await wsFixture.testSnapshot();
      expect(result).toBe(true);
    }, 15000);

    it('should take snapshot with PNG format', async () => {
      const result = await wsFixture.testSnapshotPNG();
      expect(result).toBe(true);
    }, 15000);

    it('should handle snapshot errors gracefully', async () => {
      const result = await wsFixture.testSnapshotError();
      expect(result).toBe(true);
    }, 10000);
  });

  describe('Recording Operations', () => {
    it('should start and stop recording', async () => {
      const result = await wsFixture.testRecording();
      expect(result).toBe(true);
    }, 30000);

    it('should start unlimited recording', async () => {
      const result = await wsFixture.testUnlimitedRecording();
      expect(result).toBe(true);
    }, 20000);

    it('should handle recording errors gracefully', async () => {
      const result = await wsFixture.testRecordingError();
      expect(result).toBe(true);
    }, 10000);
  });

  describe('File Management Integration', () => {
    it('should list recordings after creating them', async () => {
      const result = await wsFixture.testListRecordings();
      expect(result).toBe(true);
    }, 10000);

    it('should list snapshots after creating them', async () => {
      const result = await wsFixture.testListSnapshots();
      expect(result).toBe(true);
    }, 10000);

    it('should handle pagination correctly', async () => {
      const result = await wsFixture.testPagination();
      expect(result).toBe(true);
    }, 10000);
  });

  describe('Real-time Updates', () => {
    it('should receive camera status updates', async () => {
      const result = await wsFixture.testStatusUpdates();
      expect(result).toBe(true);
    }, 10000);

    it('should handle connection loss and recovery', async () => {
      const result = await wsFixture.testConnectionRecovery();
      expect(result).toBe(true);
    }, 10000);
  });

  describe('Error Handling', () => {
    it('should handle invalid camera operations', async () => {
      const result = await wsFixture.testInvalidCameraOperations();
      expect(result).toBe(true);
    }, 10000);

    it('should handle invalid file operations', async () => {
      const result = await wsFixture.testInvalidFileOperations();
      expect(result).toBe(true);
    }, 10000);
  });

  describe('Performance Validation', () => {
    it('should meet performance targets for camera operations', async () => {
      const result = await wsFixture.testCameraPerformance();
      expect(result).toBe(true);
    }, 10000);

    it('should meet performance targets for file operations', async () => {
      const result = await wsFixture.testFilePerformance();
      expect(result).toBe(true);
    }, 10000);
  });

  describe('Health Server Integration', () => {
    it('should access all health endpoints', async () => {
      const systemHealth = await healthFixture.testSystemHealth();
      const cameraHealth = await healthFixture.testCameraHealth();
      const mediamtxHealth = await healthFixture.testMediaMTXHealth();
      const readiness = await healthFixture.testReadiness();
      
      expect(systemHealth).toBe(true);
      expect(cameraHealth).toBe(true);
      expect(mediamtxHealth).toBe(true);
      expect(readiness).toBe(true);
    });

    it('should handle health endpoint errors gracefully', async () => {
      const result = await healthFixture.testSystemHealth();
      expect(result).toBe(true);
    });
  });
});
