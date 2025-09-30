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
import { APIResponseValidator } from './validators';

export interface TestEnvironment {
  apiClient: TestAPIClient;
  authHelper: AuthHelper;
  mocks: typeof APIMocks;
}

/**
 * Camera ID Helper - manages dynamic camera ID discovery
 * Replaces hardcoded 'camera0' with actual available camera IDs
 */
export class CameraIdHelper {
  private static availableCameraIds: string[] = [];
  private static initialized = false;

  /**
   * Initialize camera IDs by querying the server
   */
  static async initialize(apiClient: TestAPIClient): Promise<void> {
    if (this.initialized) return;

    try {
      const cameraList = await apiClient.call('get_camera_list', {});
      if (cameraList && cameraList.cameras && Array.isArray(cameraList.cameras)) {
        this.availableCameraIds = cameraList.cameras
          .map((camera: any) => camera.device)
          .filter((device: string) => APIResponseValidator.validateCameraDeviceId(device));
      }
      this.initialized = true;
    } catch (error) {
      console.warn('Failed to initialize camera IDs, using fallback:', error);
      this.availableCameraIds = ['camera0']; // Fallback
      this.initialized = true;
    }
  }

  /**
   * Get the first available camera ID
   */
  static getFirstAvailableCameraId(): string {
    if (!this.initialized) {
      console.warn('CameraIdHelper not initialized, using fallback camera0');
      return 'camera0';
    }
    
    if (this.availableCameraIds.length === 0) {
      throw new Error('No valid camera IDs available');
    }
    
    return this.availableCameraIds[0];
  }

  /**
   * Get all available camera IDs
   */
  static getAvailableCameraIds(): string[] {
    if (!this.initialized) {
      console.warn('CameraIdHelper not initialized, using fallback camera0');
      return ['camera0'];
    }
    
    return [...this.availableCameraIds];
  }

  /**
   * Validate if a camera ID is available
   */
  static isCameraIdAvailable(deviceId: string): boolean {
    return this.availableCameraIds.includes(deviceId);
  }
}

/**
 * Load test environment with all utilities
 * CRITICAL: Connect and authenticate during setup
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

  const authHelper = AuthHelper;
  const mocks = APIMocks;

  // CRITICAL: Connect and authenticate for integration tests
  if (!process.env.TEST_MOCK_MODE || process.env.TEST_MOCK_MODE === 'false') {
    await apiClient.connect();
    
    // Get real JWT token from environment
    const token = authHelper.generateTestToken('admin');
    if (!token) {
      throw new Error('No admin token available. Check test environment setup.');
    }
    
    // Authenticate the client
    const authResult = await apiClient.authenticate(token);
    if (!authResult.authenticated) {
      throw new Error('Failed to authenticate test client');
    }

    // Initialize camera IDs for dynamic discovery
    await CameraIdHelper.initialize(apiClient);
  }

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

// Unused test data generation functions removed - use MockDataFactory instead

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

// Unused test generation functions removed - use MockDataFactory instead
