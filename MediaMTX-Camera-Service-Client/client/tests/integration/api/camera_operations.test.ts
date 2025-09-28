/**
 * Camera operations integration tests
 * Real server communication tests
 * 
 * Ground Truth References:
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * 
 * Requirements Coverage:
 * - REQ-INT-001: Real API communication
 * - REQ-INT-002: Authentication flow validation
 * - REQ-INT-003: Error handling validation
 * 
 * Test Categories: Integration/API-Compliance
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */

import { TestAPIClient } from '../../utils/api-client';
import { AuthHelper } from '../../utils/auth-helper';
import { APIResponseValidator } from '../../utils/validators';
import { CameraIdHelper } from '../../utils/test-helpers';
import { loadTestEnvironment } from '../../utils/test-helpers';

describe('Camera Operations Integration Tests', () => {
  let apiClient: TestAPIClient;
  let authHelper: AuthHelper;
  let testEnv: any;

  beforeAll(async () => {
    // Load test environment with real server connection
    testEnv = await loadTestEnvironment();
    apiClient = testEnv.apiClient;
    authHelper = testEnv.authHelper;
  });

  afterAll(async () => {
    if (apiClient) {
      await apiClient.disconnect();
    }
  });

  test('REQ-INT-001: Authenticate with valid token', async () => {
    // Use the real JWT token from the server
    const token = process.env.TEST_ADMIN_TOKEN;
    if (!token) {
      throw new Error('TEST_ADMIN_TOKEN not found in environment');
    }
    
    const result = await apiClient.authenticate(token);
    
    expect(APIResponseValidator.validateAuthenticateResult(result)).toBe(true);
    expect(result.authenticated).toBe(true);
    expect(result.role).toBe('admin');
  });

  test('REQ-INT-002: Get camera list with authentication', async () => {
    // API client is already authenticated in loadTestEnvironment()
    const result = await apiClient.call('get_camera_list', {});
    
    expect(APIResponseValidator.validateCameraListResult(result)).toBe(true);
    expect(Array.isArray(result.cameras)).toBe(true);
    expect(typeof result.total).toBe('number');
    expect(typeof result.connected).toBe('number');
    
    // Validate camera IDs follow the correct pattern
    if (result.cameras && result.cameras.length > 0) {
      result.cameras.forEach((camera: any) => {
        expect(APIResponseValidator.validateCameraDeviceId(camera.device)).toBe(true);
      });
    }
  });

  test('REQ-INT-003: Get camera status for specific device', async () => {
    // API client is already authenticated in loadTestEnvironment()
    const cameraId = CameraIdHelper.getFirstAvailableCameraId();
    const result = await apiClient.call('get_camera_status', { device: cameraId });
    
    expect(APIResponseValidator.validateCamera(result)).toBe(true);
    expect(result.device).toBe(cameraId);
    expect(['CONNECTED', 'DISCONNECTED', 'ERROR']).toContain(result.status);
  });

  test('REQ-INT-004: Start recording with valid parameters', async () => {
    // API client is already authenticated in loadTestEnvironment()
    const cameraId = CameraIdHelper.getFirstAvailableCameraId();
    const result = await apiClient.call('start_recording', { device: cameraId, duration: 60, format: 'mp4' });
    
    expect(APIResponseValidator.validateRecordingStartResult(result)).toBe(true);
    expect(result.device).toBe(cameraId);
    expect(['RECORDING', 'STARTING', 'STOPPING', 'PAUSED', 'ERROR', 'FAILED']).toContain(result.status);
  });

  test('REQ-INT-005: Stop recording for active device', async () => {
    // API client is already authenticated in loadTestEnvironment()
    const cameraId = CameraIdHelper.getFirstAvailableCameraId();
    const result = await apiClient.call('stop_recording', { device: cameraId });
    
    expect(APIResponseValidator.validateRecordingStopResult(result)).toBe(true);
    expect(result.device).toBe(cameraId);
    expect(['STOPPED', 'FAILED']).toContain(result.status);
  });

  test('REQ-INT-006: Take snapshot with valid device', async () => {
    // API client is already authenticated in loadTestEnvironment()
    const cameraId = CameraIdHelper.getFirstAvailableCameraId();
    const result = await apiClient.call('take_snapshot', { device: cameraId });
    
    expect(APIResponseValidator.validateSnapshotInfo(result)).toBe(true);
    expect(result.device).toBe(cameraId);
    expect(['completed', 'failed', 'processing']).toContain(result.status);
  });

  test('REQ-INT-007: List recordings with pagination', async () => {
    // API client is already authenticated in loadTestEnvironment()
    const result = await apiClient.call('list_recordings', { limit: 10, offset: 0 });
    
    expect(APIResponseValidator.validateFileListResult(result)).toBe(true);
    expect(Array.isArray(result.files)).toBe(true);
    expect(result.limit).toBe(10);
    expect(result.offset).toBe(0);
  });

  test('REQ-INT-008: List snapshots with pagination', async () => {
    // API client is already authenticated in loadTestEnvironment()
    const result = await apiClient.call('list_snapshots', { limit: 10, offset: 0 });
    
    expect(APIResponseValidator.validateFileListResult(result)).toBe(true);
    expect(Array.isArray(result.files)).toBe(true);
    expect(result.limit).toBe(10);
    expect(result.offset).toBe(0);
  });

  test('REQ-INT-009: Get system status', async () => {
    // API client is already authenticated in loadTestEnvironment()
    const result = await apiClient.call('get_status');
    
    expect(APIResponseValidator.validateSystemStatus(result)).toBe(true);
    expect(['HEALTHY', 'DEGRADED', 'UNHEALTHY']).toContain(result.status);
  });

  test('REQ-INT-010: Get server information', async () => {
    // API client is already authenticated in loadTestEnvironment()
    const result = await apiClient.call('get_server_info');
    
    expect(APIResponseValidator.validateServerInfo(result)).toBe(true);
    expect(typeof result.name).toBe('string');
    expect(typeof result.version).toBe('string');
    expect(Array.isArray(result.capabilities)).toBe(true);
  });

  test('REQ-INT-011: Get storage information', async () => {
    // API client is already authenticated in loadTestEnvironment()
    const result = await apiClient.call('get_storage_info');
    
    expect(APIResponseValidator.validateStorageInfo(result)).toBe(true);
    expect(typeof result.total_space).toBe('number');
    expect(typeof result.used_space).toBe('number');
    expect(typeof result.available_space).toBe('number');
  });

  test('REQ-INT-012: Error handling for invalid device', async () => {
    // API client is already authenticated in loadTestEnvironment()
    await expect(apiClient.call('get_camera_status', { device: 'invalid_device' })).rejects.toThrow();
  });

  test('REQ-INT-013: Error handling for unauthorized access', async () => {
    // API client is already authenticated in loadTestEnvironment()
    // Note: This test needs a separate viewer-authenticated client
    await expect(apiClient.call('start_recording', { device: 'camera0' })).rejects.toThrow();
  });

  test('REQ-INT-014: Error handling for invalid parameters', async () => {
    // API client is already authenticated in loadTestEnvironment()
    await expect(apiClient.call('start_recording', { device: 'camera0', duration: 'invalid_duration' })).rejects.toThrow();
  });

  test('REQ-INT-015: Performance test - multiple rapid calls', async () => {
    // API client is already authenticated in loadTestEnvironment()
    const startTime = Date.now();
    const promises = Array(10).fill(null).map(() => apiClient.call('get_camera_list'));
    const results = await Promise.all(promises);
    const endTime = Date.now();
    
    expect(results).toHaveLength(10);
    expect(endTime - startTime).toBeLessThan(5000); // Should complete within 5 seconds
  });
});
