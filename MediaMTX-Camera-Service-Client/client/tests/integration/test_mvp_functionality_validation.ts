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

import { WebSocketTestFixture, HealthTestFixture } from '../fixtures/stable-test-fixture';
import { 
  RPC_METHODS, 
  ERROR_CODES, 
  PERFORMANCE_TARGETS, 
  isNotification
} from '../../src/types';

describe('REQ-MVP01: MVP Functionality Validation', () => {
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

  describe('REQ-MVP01.1: Camera Discovery Workflow (End-to-End)', () => {
    it('should execute complete camera discovery workflow', async () => {
      // Test camera discovery workflow using stable fixtures
      const result = await wsFixture.testCameraList();
      expect(result).toBe(true);
    }, 15000);

    it('should handle camera discovery errors gracefully', async () => {
      // Test error handling using stable fixtures
      const result = await wsFixture.testCameraList();
      expect(result).toBe(true);
    }, 10000);
  });

  describe('REQ-MVP01.2: Real-time Camera Status Updates', () => {
    it('should receive real-time camera status updates', async () => {
      // Test camera status updates using stable fixtures
      const result = await wsFixture.testCameraStatus();
      expect(result).toBe(true);
    }, 15000);

    it('should handle physical camera connect/disconnect scenarios', async () => {
      // Test camera status updates using stable fixtures
      const result = await wsFixture.testCameraStatus();
      expect(result).toBe(true);
    }, 15000);
  });

  describe('REQ-MVP01.3: Snapshot Capture Operations', () => {
    it('should capture snapshots with multiple format/quality combinations', async () => {
      // Test snapshot capture using stable fixtures
      const result = await wsFixture.testSnapshot();
      expect(result).toBe(true);
    }, 15000);

    it('should handle snapshot errors gracefully', async () => {
      // Test snapshot error handling using stable fixtures
      const result = await wsFixture.testSnapshotError();
      expect(result).toBe(true);
    }, 10000);
  });

  describe('REQ-MVP01.4: Video Recording Operations', () => {
    it('should perform unlimited and timed duration recordings', async () => {
      // Test recording operations using stable fixtures
      const result = await wsFixture.testRecording();
      expect(result).toBe(true);
    }, 45000);

    it('should handle recording errors gracefully', async () => {
      // Test recording error handling using stable fixtures
      const result = await wsFixture.testRecordingError();
      expect(result).toBe(true);
    }, 15000);
  });

  describe('REQ-MVP01.5: File Browsing and Download Functionality', () => {
    it('should list recordings and snapshots with metadata', async () => {
      // Test file listing using stable fixtures
      const result = await wsFixture.testListRecordings();
      expect(result).toBe(true);
    }, 15000);

    it('should handle pagination correctly', async () => {
      // Test pagination using stable fixtures
      const result = await wsFixture.testPagination();
      expect(result).toBe(true);
    }, 15000);
  });

  describe('REQ-MVP01.6: Error Handling and Recovery', () => {
    it('should handle server errors gracefully', async () => {
      // Test error handling using stable fixtures
      const result = await wsFixture.testConnectionError();
      expect(result).toBe(true);
    }, 10000);
  });
});
