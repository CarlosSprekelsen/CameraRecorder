/**
 * Common test utilities for all test categories
 * Centralized helpers to prevent duplication
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Testing Guidelines: ../docs/development/client-testing-guidelines.md
 * 
 * Requirements Coverage:
 * - REQ-HELPER-001: Common test utilities
 * - REQ-HELPER-002: Environment setup
 * - REQ-HELPER-003: Test data generation
 * 
 * Test Categories: Unit/Integration/E2E
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */

import { TestAPIClient } from './api-client';
import { AuthHelper } from './auth-helper';
import { APIMocks } from './mocks';

export interface TestEnvironment {
  apiClient: TestAPIClient;
  authHelper: AuthHelper;
  mocks: typeof APIMocks;
}

/**
 * Load test environment with all utilities
 * MANDATORY: Use this method for all test setup
 */
export async function loadTestEnvironment(): Promise<TestEnvironment> {
  // Load environment variables
  if (process.env.NODE_ENV !== 'test') {
    throw new Error('Test environment not properly configured');
  }

  const apiClient = new TestAPIClient({
    mockMode: process.env.TEST_MOCK_MODE === 'true',
    serverUrl: process.env.TEST_WEBSOCKET_URL,
    timeout: parseInt(process.env.TEST_TIMEOUT || '30000')
  });

  const authHelper = new AuthHelper();
  const mocks = APIMocks;

  return {
    apiClient,
    authHelper,
    mocks
  };
}

/**
 * Setup test environment for unit tests
 * MANDATORY: Use this method for all unit test setup
 */
export async function setupUnitTestEnvironment(): Promise<TestEnvironment> {
  process.env.TEST_MOCK_MODE = 'true';
  return loadTestEnvironment();
}

/**
 * Setup test environment for integration tests
 * MANDATORY: Use this method for all integration test setup
 */
export async function setupIntegrationTestEnvironment(): Promise<TestEnvironment> {
  process.env.TEST_MOCK_MODE = 'false';
  return loadTestEnvironment();
}

/**
 * Setup test environment for E2E tests
 * MANDATORY: Use this method for all E2E test setup
 */
export async function setupE2ETestEnvironment(): Promise<TestEnvironment> {
  process.env.TEST_MOCK_MODE = 'false';
  return loadTestEnvironment();
}

/**
 * Generate test data for camera operations
 * MANDATORY: Use this method for all camera test data
 */
export function generateTestCameraData(count: number = 1): any[] {
  const cameras = [];
  for (let i = 0; i < count; i++) {
    cameras.push({
      device: `camera${i}`,
      status: i % 2 === 0 ? 'CONNECTED' : 'DISCONNECTED',
      name: `Test Camera ${i}`,
      resolution: i % 2 === 0 ? '1920x1080' : '1280x720',
      fps: i % 2 === 0 ? 30 : 25
    });
  }
  return cameras;
}

/**
 * Generate test data for file operations
 * MANDATORY: Use this method for all file test data
 */
export function generateTestFileData(type: 'recordings' | 'snapshots', count: number = 1): any[] {
  const files = [];
  for (let i = 0; i < count; i++) {
    files.push({
      filename: `${type}_camera0_${Date.now()}_${i}.${type === 'recordings' ? 'mp4' : 'jpg'}`,
      file_size: type === 'recordings' ? 1024000 * (i + 1) : 512000 * (i + 1),
      modified_time: new Date(Date.now() - i * 3600000).toISOString(),
      download_url: `https://localhost/downloads/${type}_camera0_${Date.now()}_${i}.${type === 'recordings' ? 'mp4' : 'jpg'}`
    });
  }
  return files;
}

/**
 * Generate test data for authentication
 * MANDATORY: Use this method for all auth test data
 */
export function generateTestAuthData(role: 'admin' | 'operator' | 'viewer' = 'admin'): any {
  return {
    token: `test-token-${role}-${Date.now()}`,
    role,
    permissions: getRolePermissions(role),
    sessionId: `test-session-${Date.now()}`
  };
}

/**
 * Generate test data for error scenarios
 * MANDATORY: Use this method for all error test data
 */
export function generateTestErrorData(code: number, message?: string): any {
  return {
    code,
    message: message || `Test error ${code}`,
    timestamp: new Date().toISOString()
  };
}

/**
 * Wait for async operations to complete
 * MANDATORY: Use this method for all async test operations
 */
export async function waitFor(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms));
}

/**
 * Wait for condition to be true
 * MANDATORY: Use this method for all conditional test operations
 */
export async function waitForCondition(
  condition: () => boolean,
  timeout: number = 5000,
  interval: number = 100
): Promise<void> {
  const startTime = Date.now();
  
  while (Date.now() - startTime < timeout) {
    if (condition()) {
      return;
    }
    await waitFor(interval);
  }
  
  throw new Error(`Condition not met within ${timeout}ms`);
}

/**
 * Retry operation with exponential backoff
 * MANDATORY: Use this method for all retry test operations
 */
export async function retryOperation<T>(
  operation: () => Promise<T>,
  maxRetries: number = 3,
  baseDelay: number = 1000
): Promise<T> {
  let lastError: Error;
  
  for (let i = 0; i < maxRetries; i++) {
    try {
      return await operation();
    } catch (error) {
      lastError = error as Error;
      if (i < maxRetries - 1) {
        const delay = baseDelay * Math.pow(2, i);
        await waitFor(delay);
      }
    }
  }
  
  throw lastError!;
}

/**
 * Validate test environment configuration
 * MANDATORY: Use this method for all test environment validation
 */
export function validateTestEnvironment(): boolean {
  const requiredEnvVars = [
    'NODE_ENV',
    'TEST_WEBSOCKET_URL',
    'TEST_JWT_SECRET'
  ];
  
  return requiredEnvVars.every(envVar => process.env[envVar]);
}

/**
 * Cleanup test resources
 * MANDATORY: Use this method for all test cleanup
 */
export async function cleanupTestResources(apiClient?: TestAPIClient): Promise<void> {
  if (apiClient) {
    await apiClient.disconnect();
  }
  
  // Clear any test data
  // Reset environment variables
  delete process.env.TEST_MOCK_MODE;
}

/**
 * Get role-specific permissions
 * MANDATORY: Use this method for all role permission tests
 */
function getRolePermissions(role: 'admin' | 'operator' | 'viewer'): string[] {
  switch (role) {
    case 'admin':
      return ['read', 'write', 'delete', 'admin'];
    case 'operator':
      return ['read', 'write'];
    case 'viewer':
      return ['read'];
    default:
      return [];
  }
}

/**
 * Create test timeout
 * MANDATORY: Use this method for all timeout tests
 */
export function createTestTimeout(ms: number): Promise<never> {
  return new Promise((_, reject) => {
    setTimeout(() => {
      reject(new Error(`Test timeout after ${ms}ms`));
    }, ms);
  });
}

/**
 * Generate test UUID
 * MANDATORY: Use this method for all UUID generation tests
 */
export function generateTestUUID(): string {
  return `test-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
}

/**
 * Generate test timestamp
 * MANDATORY: Use this method for all timestamp tests
 */
export function generateTestTimestamp(): string {
  return new Date().toISOString();
}

/**
 * Generate test duration
 * MANDATORY: Use this method for all duration tests
 */
export function generateTestDuration(minutes: number = 1): number {
  return minutes * 60;
}

/**
 * Generate test file size
 * MANDATORY: Use this method for all file size tests
 */
export function generateTestFileSize(mb: number = 1): number {
  return mb * 1024 * 1024;
}

/**
 * Generate test bitrate
 * MANDATORY: Use this method for all bitrate tests
 */
export function generateTestBitrate(mbps: number = 2): number {
  return mbps * 1000000;
}

/**
 * Generate test FPS
 * MANDATORY: Use this method for all FPS tests
 */
export function generateTestFPS(fps: number = 30): number {
  return fps;
}

/**
 * Generate test resolution
 * MANDATORY: Use this method for all resolution tests
 */
export function generateTestResolution(width: number = 1920, height: number = 1080): string {
  return `${width}x${height}`;
}

/**
 * Generate test device ID
 * MANDATORY: Use this method for all device ID tests
 */
export function generateTestDeviceId(index: number = 0): string {
  return `camera${index}`;
}

/**
 * Generate test filename
 * MANDATORY: Use this method for all filename tests
 */
export function generateTestFilename(type: 'recording' | 'snapshot', device: string = 'camera0'): string {
  const extension = type === 'recording' ? 'mp4' : 'jpg';
  return `${type}_${device}_${Date.now()}.${extension}`;
}

/**
 * Generate test download URL
 * MANDATORY: Use this method for all download URL tests
 */
export function generateTestDownloadUrl(filename: string): string {
  return `https://localhost/downloads/${filename}`;
}

/**
 * Generate test stream URL
 * MANDATORY: Use this method for all stream URL tests
 */
export function generateTestStreamUrl(device: string = 'camera0', protocol: 'rtsp' | 'hls' = 'hls'): string {
  if (protocol === 'rtsp') {
    return `rtsp://localhost:8554/${device}`;
  } else {
    return `https://localhost/hls/${device}.m3u8`;
  }
}
