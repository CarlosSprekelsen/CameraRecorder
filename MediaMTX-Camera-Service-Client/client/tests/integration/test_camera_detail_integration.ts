/**
 * REQ-UNIT01-001: Camera information display must be clear and accessible
 * REQ-UNIT01-002: Camera controls must be functional and responsive
 * Coverage: INTEGRATION
 * Quality: HIGH
 */
/**
 * Integration tests for CameraDetail component requirements
 * Tests real server integration for camera information display and controls
 * 
 * These tests require a running MediaMTX server for full validation
 */

import { WebSocketTestFixture } from '../fixtures/stable-test-fixture';

describe('CameraDetail Integration', () => {
  let wsFixture: WebSocketTestFixture;

  beforeAll(async () => {
    // Initialize stable fixtures for authentication and server availability
    wsFixture = new WebSocketTestFixture();
    
    await wsFixture.initialize();
  });

  afterAll(async () => {
    wsFixture.cleanup();
  });

  describe('Camera Information Display', () => {
    it('should retrieve and display camera information correctly', async () => {
      // Test camera list retrieval - validates camera information structure
      const cameraListResult = await wsFixture.testCameraList();
      expect(cameraListResult).toBe(true);
    }, 10000);

    it('should get individual camera status with all required fields', async () => {
      // Test camera status retrieval - validates detailed camera information
      const cameraStatusResult = await wsFixture.testCameraStatus();
      expect(cameraStatusResult).toBe(true);
    }, 10000);
  });

  describe('Camera Controls Functionality', () => {
    it('should take snapshots with different formats', async () => {
      // Test snapshot functionality with JPEG format
      const jpegResult = await wsFixture.testSnapshot();
      expect(jpegResult).toBe(true);

      // Test snapshot functionality with PNG format
      const pngResult = await wsFixture.testSnapshotPNG();
      expect(pngResult).toBe(true);
    }, 30000);

    it('should start and stop recordings', async () => {
      // Test recording start functionality
      const startResult = await wsFixture.testRecording();
      expect(startResult).toBe(true);
    }, 30000);

    it('should handle unlimited duration recordings', async () => {
      // Test unlimited recording functionality
      const result = await wsFixture.testUnlimitedRecording();
      expect(result).toBe(true);
    }, 20000);
  });

  describe('Error Handling and Edge Cases', () => {
    it('should handle snapshot errors gracefully', async () => {
      // Test error handling for snapshot operations
      const result = await wsFixture.testSnapshotError();
      expect(result).toBe(true);
    }, 10000);

    it('should handle recording errors gracefully', async () => {
      // Test error handling for recording operations
      const result = await wsFixture.testRecordingError();
      expect(result).toBe(true);
    }, 10000);

    it('should handle invalid camera operations', async () => {
      // Test error handling for invalid operations
      const result = await wsFixture.testInvalidCameraOperations();
      expect(result).toBe(true);
    }, 10000);
  });

  describe('Real-time Status Updates', () => {
    it('should receive camera status updates', async () => {
      // Test real-time status update functionality
      const result = await wsFixture.testStatusUpdates();
      expect(result).toBe(true);
    }, 10000);

    it('should handle connection loss and recovery', async () => {
      // Test connection resilience
      const result = await wsFixture.testConnectionRecovery();
      expect(result).toBe(true);
    }, 10000);
  });

  describe('File Management Integration', () => {
    it('should list recordings after creating them', async () => {
      // Test recording file management
      const result = await wsFixture.testListRecordings();
      expect(result).toBe(true);
    }, 10000);

    it('should list snapshots after creating them', async () => {
      // Test snapshot file management
      const result = await wsFixture.testListSnapshots();
      expect(result).toBe(true);
    }, 10000);
  });

  describe('Performance Validation', () => {
    it('should meet performance targets for camera operations', async () => {
      // Test performance requirements for camera operations
      const result = await wsFixture.testCameraPerformance();
      expect(result).toBe(true);
    }, 10000);

    it('should meet performance targets for file operations', async () => {
      // Test performance requirements for file operations
      const result = await wsFixture.testFilePerformance();
      expect(result).toBe(true);
    }, 10000);
  });

  describe('WebSocket Health Monitoring', () => {
    it('should monitor health via WebSocket connection', async () => {
      // Health monitoring is done via WebSocket, not separate HTTP endpoints
      const connectionHealthy = await wsFixture.testConnection();
      expect(connectionHealthy).toBe(true);
    });
  });
});
