/**
 * Recording workflow E2E tests
 * Complete user workflows with real hardware/simulation
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * 
 * Requirements Coverage:
 * - REQ-E2E-001: Complete recording workflow
 * - REQ-E2E-002: Real hardware interaction
 * - REQ-E2E-003: Performance validation
 * 
 * Test Categories: E2E/Performance
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */

import { TestAPIClient } from '../../utils/api-client';
import { AuthHelper } from '../../utils/auth-helper';
import { APIResponseValidator } from '../../utils/validators';
import { loadTestEnvironment, waitFor, waitForCondition } from '../../utils/test-helpers';
import { AuthService } from '../../src/services/auth/AuthService';
import { LoggerService } from '../../src/services/logger/LoggerService';

describe('Recording Workflow E2E Tests', () => {
  let apiClient: TestAPIClient;
  let authService: AuthService;
  let authHelper: AuthHelper;
  let testEnv: any;

  beforeAll(async () => {
    // Load test environment with real server connection
    testEnv = await loadTestEnvironment();
    apiClient = testEnv.apiClient;
    
    // Create AuthService following architectural pattern
    authService = new AuthService(apiClient, LoggerService.getInstance());
    authHelper = testEnv.authHelper;
  });

  afterAll(async () => {
    if (apiClient) {
      await apiClient.disconnect();
    }
  });

  test('REQ-E2E-001: Complete recording workflow - start, monitor, stop', async () => {
    const token = await authHelper.generateTestToken('admin');
    await authHelper.authenticateWithToken(token);
    
    // Step 1: Get camera list and verify camera is available
    const cameraList = await apiClient.call('get_camera_list');
    expect(APIResponseValidator.validateCameraListResult(cameraList)).toBe(true);
    expect(cameraList.cameras.length).toBeGreaterThan(0);
    
    const camera = cameraList.cameras[0];
    expect(camera.status).toBe('CONNECTED');
    
    // Step 2: Start recording
    const startResult = await apiClient.call('start_recording', { device: camera.device, duration: 60, format: 'mp4' });
    expect(APIResponseValidator.validateRecordingStartResult(startResult)).toBe(true);
    expect(startResult.device).toBe(camera.device);
    expect(['RECORDING', 'STARTING', 'STOPPING', 'PAUSED', 'ERROR', 'FAILED']).toContain(startResult.status);
    
    // Step 3: Wait for recording to start
    await waitForCondition(
      () => startResult.status === 'RECORDING',
      10000, // 10 second timeout
      1000  // Check every second
    );
    
    // Step 4: Monitor recording status
    const statusResult = await apiClient.call('get_camera_status', { device: camera.device });
    expect(APIResponseValidator.validateCamera(statusResult)).toBe(true);
    expect(statusResult.device).toBe(camera.device);
    
    // Step 5: Stop recording
    const stopResult = await apiClient.call('stop_recording', { device: camera.device });
    expect(APIResponseValidator.validateRecordingStopResult(stopResult)).toBe(true);
    expect(stopResult.device).toBe(camera.device);
    expect(['STOPPED', 'FAILED']).toContain(stopResult.status);
    
    // Step 6: Verify recording file was created
    const recordings = await apiClient.call('list_recordings', { limit: 10, offset: 0 });
    expect(APIResponseValidator.validateFileListResult(recordings)).toBe(true);
    expect(recordings.files.length).toBeGreaterThan(0);
    
    const recordingFile = recordings.files.find(file => 
      file.filename.includes(camera.device) && 
      file.filename.includes('mp4')
    );
    expect(recordingFile).toBeDefined();
    expect(recordingFile?.file_size).toBeGreaterThan(0);
  });

  test('REQ-E2E-002: Snapshot capture workflow', async () => {
    const token = await authHelper.generateTestToken('admin');
    await authService.authenticate(token);
    
    // Step 1: Get camera list
    const cameraList = await apiClient.call('get_camera_list');
    expect(cameraList.cameras.length).toBeGreaterThan(0);
    
    const camera = cameraList.cameras[0];
    expect(camera.status).toBe('CONNECTED');
    
    // Step 2: Take snapshot
    const snapshotResult = await apiClient.call('take_snapshot', { device: camera.device });
    expect(APIResponseValidator.validateSnapshotInfo(snapshotResult)).toBe(true);
    expect(snapshotResult.device).toBe(camera.device);
    expect(['SUCCESS', 'FAILED']).toContain(snapshotResult.status);
    
    // Step 3: Verify snapshot file was created
    const snapshots = await apiClient.call('list_snapshots', { limit: 10, offset: 0 });
    expect(APIResponseValidator.validateFileListResult(snapshots)).toBe(true);
    expect(snapshots.files.length).toBeGreaterThan(0);
    
    const snapshotFile = snapshots.files.find(file => 
      file.filename.includes(camera.device) && 
      file.filename.includes('jpg')
    );
    expect(snapshotFile).toBeDefined();
    expect(snapshotFile?.file_size).toBeGreaterThan(0);
  });

  test('REQ-E2E-003: File management workflow - list, download, delete', async () => {
    const token = await authHelper.generateTestToken('admin');
    await authService.authenticate(token);
    
    // Step 1: List recordings
    const recordings = await apiClient.call('list_recordings', { limit: 50, offset: 0 });
    expect(APIResponseValidator.validateFileListResult(recordings)).toBe(true);
    
    if (recordings.files.length > 0) {
      const recording = recordings.files[0];
      
      // Step 2: Get recording info
      const recordingInfo = await apiClient.call('get_recording_info', { filename: recording.filename });
      expect(recordingInfo.filename).toBe(recording.filename);
      expect(typeof recordingInfo.file_size).toBe('number');
      expect(typeof recordingInfo.download_url).toBe('string');
      
      // Step 3: Delete recording
      const deleteResult = await apiClient.call('delete_recording', { filename: recording.filename });
      expect(deleteResult.filename).toBe(recording.filename);
      expect(deleteResult.deleted).toBe(true);
      
      // Step 4: Verify deletion
      const updatedRecordings = await apiClient.call('list_recordings', { limit: 50, offset: 0 });
      const deletedFile = updatedRecordings.files.find(file => file.filename === recording.filename);
      expect(deletedFile).toBeUndefined();
    }
  });

  test('REQ-E2E-004: System monitoring workflow', async () => {
    const token = await authHelper.generateTestToken('admin');
    await authService.authenticate(token);
    
    // Step 1: Get system status
    const status = await apiClient.call('get_status');
    expect(APIResponseValidator.validateSystemStatus(status)).toBe(true);
    expect(['HEALTHY', 'DEGRADED', 'UNHEALTHY']).toContain(status.status);
    
    // Step 2: Get server information
    const serverInfo = await apiClient.call('get_server_info');
    expect(APIResponseValidator.validateServerInfo(serverInfo)).toBe(true);
    expect(typeof serverInfo.name).toBe('string');
    expect(typeof serverInfo.version).toBe('string');
    
    // Step 3: Get storage information
    const storageInfo = await apiClient.call('get_storage_info');
    expect(APIResponseValidator.validateStorageInfo(storageInfo)).toBe(true);
    expect(typeof storageInfo.total_space).toBe('number');
    expect(typeof storageInfo.used_space).toBe('number');
    expect(storageInfo.total_space).toBeGreaterThan(storageInfo.used_space);
    
    // Step 4: Get metrics
    const metrics = await apiClient.call('get_metrics');
    expect(APIResponseValidator.validateMetricsResult(metrics)).toBe(true);
    expect(typeof metrics.timestamp).toBe('string');
    expect(typeof metrics.system_metrics).toBe('object');
    expect(typeof metrics.camera_metrics).toBe('object');
  });

  test('REQ-E2E-005: Performance test - concurrent operations', async () => {
    const token = await authHelper.generateTestToken('admin');
    await authService.authenticate(token);
    
    const startTime = Date.now();
    
    // Perform multiple concurrent operations
    const operations = [
      apiClient.call('get_camera_list'),
      apiClient.call('get_status'),
      apiClient.call('get_server_info'),
      apiClient.call('get_storage_info'),
      apiClient.call('get_metrics')
    ];
    
    const results = await Promise.all(operations);
    const endTime = Date.now();
    
    expect(results).toHaveLength(5);
    expect(endTime - startTime).toBeLessThan(10000); // Should complete within 10 seconds
    
    // Verify all results are valid
    expect(APIResponseValidator.validateCameraListResult(results[0])).toBe(true);
    expect(APIResponseValidator.validateSystemStatus(results[1])).toBe(true);
    expect(APIResponseValidator.validateServerInfo(results[2])).toBe(true);
    expect(APIResponseValidator.validateStorageInfo(results[3])).toBe(true);
    expect(APIResponseValidator.validateMetricsResult(results[4])).toBe(true);
  });

  test('REQ-E2E-006: Error recovery workflow', async () => {
    const token = await authHelper.generateTestToken('admin');
    await authService.authenticate(token);
    
    // Step 1: Attempt operation with invalid device
    await expect(apiClient.call('get_camera_status', { device: 'invalid_device' })).rejects.toThrow();
    
    // Step 2: Verify system is still operational
    const status = await apiClient.call('get_status');
    expect(APIResponseValidator.validateSystemStatus(status)).toBe(true);
    
    // Step 3: Verify camera list still works
    const cameraList = await apiClient.call('get_camera_list');
    expect(APIResponseValidator.validateCameraListResult(cameraList)).toBe(true);
  });

  test('REQ-E2E-007: Authentication workflow with different roles', async () => {
    // Test admin role
    const adminToken = await authHelper.generateTestToken('admin');
    await authService.authenticate(adminToken);
    
    const adminResult = await apiClient.call('get_camera_list');
    expect(APIResponseValidator.validateCameraListResult(adminResult)).toBe(true);
    
    // Test operator role
    const operatorToken = await authHelper.generateTestToken('operator');
    await authService.authenticate(operatorToken);
    
    const operatorResult = await apiClient.call('get_camera_list');
    expect(APIResponseValidator.validateCameraListResult(operatorResult)).toBe(true);
    
    // Test viewer role
    const viewerToken = await authHelper.generateTestToken('viewer');
    await authService.authenticate(viewerToken);
    
    const viewerResult = await apiClient.call('get_camera_list');
    expect(APIResponseValidator.validateCameraListResult(viewerResult)).toBe(true);
  });
});
