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
    const token = await authHelper.generateTestToken('admin');
    const result = await apiClient.authenticate(token);
    
    expect(APIResponseValidator.validateAuthResult(result)).toBe(true);
    expect(result.authenticated).toBe(true);
    expect(result.role).toBe('admin');
  });

  test('REQ-INT-002: Get camera list with authentication', async () => {
    const token = await authHelper.generateTestToken('admin');
    await apiClient.authenticate(token);
    
    const result = await apiClient.call('get_camera_list');
    
    expect(APIResponseValidator.validateCameraListResult(result)).toBe(true);
    expect(Array.isArray(result.cameras)).toBe(true);
    expect(typeof result.total).toBe('number');
    expect(typeof result.connected).toBe('number');
  });

  test('REQ-INT-003: Get camera status for specific device', async () => {
    const token = await authHelper.generateTestToken('admin');
    await apiClient.authenticate(token);
    
    const result = await apiClient.call('get_camera_status', ['camera0']);
    
    expect(APIResponseValidator.validateCamera(result)).toBe(true);
    expect(result.device).toBe('camera0');
    expect(['CONNECTED', 'DISCONNECTED', 'ERROR']).toContain(result.status);
  });

  test('REQ-INT-004: Start recording with valid parameters', async () => {
    const token = await authHelper.generateTestToken('admin');
    await apiClient.authenticate(token);
    
    const result = await apiClient.call('start_recording', ['camera0', 60, 'mp4']);
    
    expect(APIResponseValidator.validateRecordingStartResult(result)).toBe(true);
    expect(result.device).toBe('camera0');
    expect(['RECORDING', 'STARTING', 'STOPPING', 'PAUSED', 'ERROR', 'FAILED']).toContain(result.status);
  });

  test('REQ-INT-005: Stop recording for active device', async () => {
    const token = await authHelper.generateTestToken('admin');
    await apiClient.authenticate(token);
    
    const result = await apiClient.call('stop_recording', ['camera0']);
    
    expect(APIResponseValidator.validateRecordingStopResult(result)).toBe(true);
    expect(result.device).toBe('camera0');
    expect(['STOPPED', 'FAILED']).toContain(result.status);
  });

  test('REQ-INT-006: Take snapshot with valid device', async () => {
    const token = await authHelper.generateTestToken('admin');
    await apiClient.authenticate(token);
    
    const result = await apiClient.call('take_snapshot', ['camera0']);
    
    expect(APIResponseValidator.validateSnapshotInfo(result)).toBe(true);
    expect(result.device).toBe('camera0');
    expect(['SUCCESS', 'FAILED']).toContain(result.status);
  });

  test('REQ-INT-007: List recordings with pagination', async () => {
    const token = await authHelper.generateTestToken('admin');
    await apiClient.authenticate(token);
    
    const result = await apiClient.call('list_recordings', [10, 0]);
    
    expect(APIResponseValidator.validateListFilesResult(result)).toBe(true);
    expect(Array.isArray(result.files)).toBe(true);
    expect(result.limit).toBe(10);
    expect(result.offset).toBe(0);
  });

  test('REQ-INT-008: List snapshots with pagination', async () => {
    const token = await authHelper.generateTestToken('admin');
    await apiClient.authenticate(token);
    
    const result = await apiClient.call('list_snapshots', [10, 0]);
    
    expect(APIResponseValidator.validateListFilesResult(result)).toBe(true);
    expect(Array.isArray(result.files)).toBe(true);
    expect(result.limit).toBe(10);
    expect(result.offset).toBe(0);
  });

  test('REQ-INT-009: Get system status', async () => {
    const token = await authHelper.generateTestToken('admin');
    await apiClient.authenticate(token);
    
    const result = await apiClient.call('get_status');
    
    expect(APIResponseValidator.validateStatusResult(result)).toBe(true);
    expect(['HEALTHY', 'DEGRADED', 'UNHEALTHY']).toContain(result.status);
  });

  test('REQ-INT-010: Get server information', async () => {
    const token = await authHelper.generateTestToken('admin');
    await apiClient.authenticate(token);
    
    const result = await apiClient.call('get_server_info');
    
    expect(APIResponseValidator.validateServerInfo(result)).toBe(true);
    expect(typeof result.name).toBe('string');
    expect(typeof result.version).toBe('string');
    expect(Array.isArray(result.capabilities)).toBe(true);
  });

  test('REQ-INT-011: Get storage information', async () => {
    const token = await authHelper.generateTestToken('admin');
    await apiClient.authenticate(token);
    
    const result = await apiClient.call('get_storage_info');
    
    expect(APIResponseValidator.validateStorageInfo(result)).toBe(true);
    expect(typeof result.total_space).toBe('number');
    expect(typeof result.used_space).toBe('number');
    expect(typeof result.available_space).toBe('number');
  });

  test('REQ-INT-012: Error handling for invalid device', async () => {
    const token = await authHelper.generateTestToken('admin');
    await apiClient.authenticate(token);
    
    await expect(apiClient.call('get_camera_status', ['invalid_device'])).rejects.toThrow();
  });

  test('REQ-INT-013: Error handling for unauthorized access', async () => {
    const token = await authHelper.generateTestToken('viewer');
    await apiClient.authenticate(token);
    
    // Viewer should not be able to start recording
    await expect(apiClient.call('start_recording', ['camera0'])).rejects.toThrow();
  });

  test('REQ-INT-014: Error handling for invalid parameters', async () => {
    const token = await authHelper.generateTestToken('admin');
    await apiClient.authenticate(token);
    
    await expect(apiClient.call('start_recording', ['camera0', 'invalid_duration'])).rejects.toThrow();
  });

  test('REQ-INT-015: Performance test - multiple rapid calls', async () => {
    const token = await authHelper.generateTestToken('admin');
    await apiClient.authenticate(token);
    
    const startTime = Date.now();
    const promises = Array(10).fill(null).map(() => apiClient.call('get_camera_list'));
    const results = await Promise.all(promises);
    const endTime = Date.now();
    
    expect(results).toHaveLength(10);
    expect(endTime - startTime).toBeLessThan(5000); // Should complete within 5 seconds
  });
});
