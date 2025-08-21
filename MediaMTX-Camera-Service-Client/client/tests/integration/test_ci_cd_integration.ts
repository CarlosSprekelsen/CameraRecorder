/**
 * REQ-CICD01-001: CI/CD pipeline integration validation
 * REQ-CICD01-002: Secondary requirements covered
 * Coverage: INTEGRATION
 * Quality: HIGH
 */
/**
 * CI/CD Integration Tests
 * 
 * Validates automated testing pipeline with real server integration
 * Following the unified testing strategy CI/CD requirements
 * 
 * Prerequisites:
 * - MediaMTX Camera Service running via systemd
 * - Server accessible at ws://localhost:8002/ws
 * - Health check endpoint accessible
 */

import { WebSocketTestFixture, HealthTestFixture } from '../fixtures/stable-test-fixture';

describe('CI/CD Integration Tests', () => {
  let wsFixture: WebSocketTestFixture;
  let healthFixture: HealthTestFixture;

  beforeAll(async () => {
    // Initialize stable fixtures for authentication and server availability
    wsFixture = new WebSocketTestFixture();
    healthFixture = new HealthTestFixture();
    
    await wsFixture.initialize();
    await healthFixture.initialize();
    
    // Verify server is available using stable fixtures
    const serverAvailable = await wsFixture.testConnection();
    if (!serverAvailable) {
      throw new Error('MediaMTX Camera Service not available for CI/CD testing.');
    }
  });

  afterAll(async () => {
    wsFixture.cleanup();
    healthFixture.cleanup();
  });

  describe('Service Startup and Readiness', () => {
    it('should verify server service is running', async () => {
      // Check if MediaMTX service is active via systemd
      const isServiceActive = await checkSystemdServiceStatus();
      expect(isServiceActive).toBe(true);
    });

    it('should verify health check endpoint is accessible', async () => {
      const isHealthy = await healthFixture.testHealthEndpoint();
      expect(isHealthy).toBe(true);
    });

    it('should verify WebSocket endpoint is accessible', async () => {
      const isWebSocketAccessible = await wsFixture.testConnection();
      expect(isWebSocketAccessible).toBe(true);
    });
  });

  describe('Network Connectivity', () => {
    it('should validate network connectivity between components', async () => {
      const connectivityChecks = await performConnectivityChecks();
      
      expect(connectivityChecks.apiEndpoint).toBe(true);
      expect(connectivityChecks.websocketEndpoint).toBe(true);
      expect(connectivityChecks.rtspEndpoint).toBe(true);
    });

    it('should validate component communication', async () => {
      // Test basic communication using stable fixture
      const connectionResult = await wsFixture.testConnection();
      expect(connectionResult).toBe(true);
      
      // Test ping functionality
      const pingResult = await wsFixture.testPing();
      expect(pingResult).toBe(true);
    });
  });

  describe('Test Execution Sequencing', () => {
    it('should execute server-first, then client tests', async () => {
      // Step 1: Verify server is ready
      const serverReady = await verifyServerReadiness();
      expect(serverReady).toBe(true);
      
      // Step 2: Execute client integration tests
      const clientTestsPass = await executeClientIntegrationTests();
      expect(clientTestsPass).toBe(true);
    });

    it('should handle test isolation and cleanup', async () => {
      // Setup test state
      const testState = await setupTestState();
      expect(testState.initialized).toBe(true);
      
      // Execute tests
      const testResult = await executeIsolatedTests();
      expect(testResult.success).toBe(true);
      
      // Cleanup test state
      const cleanupResult = await cleanupTestState();
      expect(cleanupResult.cleaned).toBe(true);
    });
  });

  describe('End-to-End Workflow Validation', () => {
    it('should validate complete camera operations workflow', async () => {
      // Step 1: Get camera list
      const cameraListResult = await wsFixture.testCameraList();
      expect(cameraListResult).toBe(true);
      
      // Step 2: Get camera status
      const cameraStatusResult = await wsFixture.testCameraStatus();
      expect(cameraStatusResult).toBe(true);
      
      // Step 3: Take snapshot
      const snapshotResult = await wsFixture.testSnapshot();
      expect(snapshotResult).toBe(true);
      
      // Step 4: Start recording
      const recordingResult = await wsFixture.testRecording();
      expect(recordingResult).toBe(true);
    });

    it('should validate file management operations', async () => {
      // Test recordings list
      const recordingsResult = await wsFixture.testListRecordings();
      expect(recordingsResult).toBe(true);
      
      // Test snapshots list
      const snapshotsResult = await wsFixture.testListSnapshots();
      expect(snapshotsResult).toBe(true);
    });
  });

  describe('Performance Validation in CI/CD', () => {
    it('should validate performance targets in CI environment', async () => {
      // Test status method performance
      const statusStartTime = performance.now();
      const pingResult = await wsFixture.testPing();
      const statusTime = performance.now() - statusStartTime;
      expect(pingResult).toBe(true);
      expect(statusTime).toBeLessThan(50); // <50ms target
      
      // Test control method performance
      const controlStartTime = performance.now();
      const snapshotResult = await wsFixture.testSnapshot();
      const controlTime = performance.now() - controlStartTime;
      expect(snapshotResult).toBe(true);
      expect(controlTime).toBeLessThan(100); // <100ms target
    });
  });

  describe('Error Handling and Recovery', () => {
    it('should handle connection failures gracefully', async () => {
      // Test connection error handling
      const errorResult = await wsFixture.testConnectionError();
      expect(errorResult).toBe(true);
    });

    it('should validate recovery mechanisms', async () => {
      // Test recovery from connection failures
      const recoveryResult = await wsFixture.testConnectionRecovery();
      expect(recoveryResult).toBe(true);
    });
  });

  describe('Authentication and Security', () => {
    it('should validate authentication requirements', async () => {
      // Test authentication flow
      const authResult = await wsFixture.testAuthentication();
      expect(authResult).toBe(true);
    });

    it('should validate unauthorized access blocking', async () => {
      // Test unauthorized access blocking
      const unauthorizedResult = await wsFixture.testUnauthorizedAccess();
      expect(unauthorizedResult).toBe(true);
    });
  });
});

// Helper functions for CI/CD validation
async function checkSystemdServiceStatus(): Promise<boolean> {
  try {
    const { exec } = require('child_process');
    return new Promise((resolve) => {
      exec('systemctl is-active --quiet mediamtx-camera-service', (error: any) => {
        resolve(!error);
      });
    });
  } catch {
    return false;
  }
}

async function performConnectivityChecks(): Promise<{
  apiEndpoint: boolean;
  websocketEndpoint: boolean;
  rtspEndpoint: boolean;
}> {
  try {
    const { exec } = require('child_process');
    
    const checkPort = (port: number): Promise<boolean> => {
      return new Promise((resolve) => {
        exec(`nc -z localhost ${port}`, (error: any) => {
          resolve(!error);
        });
      });
    };

    const [apiEndpoint, websocketEndpoint, rtspEndpoint] = await Promise.all([
      checkPort(8003), // Health server
      checkPort(8002), // WebSocket server
      checkPort(8554)  // RTSP server
    ]);

    return { apiEndpoint, websocketEndpoint, rtspEndpoint };
  } catch {
    return { apiEndpoint: false, websocketEndpoint: false, rtspEndpoint: false };
  }
}

async function verifyServerReadiness(): Promise<boolean> {
  try {
    const healthResult = await healthFixture.testHealthEndpoint();
    const wsResult = await wsFixture.testConnection();
    return healthResult && wsResult;
  } catch {
    return false;
  }
}

async function executeClientIntegrationTests(): Promise<boolean> {
  try {
    // Test core functionality
    const pingResult = await wsFixture.testPing();
    const cameraListResult = await wsFixture.testCameraList();
    const cameraStatusResult = await wsFixture.testCameraStatus();
    
    return pingResult && cameraListResult && cameraStatusResult;
  } catch {
    return false;
  }
}

async function setupTestState(): Promise<{ initialized: boolean }> {
  try {
    await wsFixture.initialize();
    return { initialized: true };
  } catch {
    return { initialized: false };
  }
}

async function executeIsolatedTests(): Promise<{ success: boolean }> {
  try {
    const pingResult = await wsFixture.testPing();
    return { success: pingResult };
  } catch {
    return { success: false };
  }
}

async function cleanupTestState(): Promise<{ cleaned: boolean }> {
  try {
    wsFixture.cleanup();
    return { cleaned: true };
  } catch {
    return { cleaned: false };
  }
}
