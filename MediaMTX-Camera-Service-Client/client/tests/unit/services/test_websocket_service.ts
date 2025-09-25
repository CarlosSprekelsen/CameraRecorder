/**
 * WebSocket service unit tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * 
 * Requirements Coverage:
 * - REQ-UNIT-001: WebSocket connection management
 * - REQ-UNIT-002: JSON-RPC message handling
 * - REQ-UNIT-003: Error handling
 * 
 * Test Categories: Unit
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */

import { TestAPIClient } from '../../utils/api-client';
import { AuthHelper } from '../../utils/auth-helper';
import { APIMocks } from '../../utils/mocks';
import { APIResponseValidator } from '../../utils/validators';

describe('WebSocket Service Unit Tests', () => {
  let apiClient: TestAPIClient;
  let authHelper: AuthHelper;

  beforeEach(() => {
    // Use centralized utilities
    apiClient = new TestAPIClient({ mockMode: true });
    authHelper = new AuthHelper();
  });

  afterEach(async () => {
    await apiClient.disconnect();
  });

  test('REQ-UNIT-001: WebSocket connection establishment', async () => {
    await expect(apiClient.connect()).resolves.not.toThrow();
  });

  test('REQ-UNIT-002: Ping connectivity check', async () => {
    await apiClient.connect();
    const result = await apiClient.ping();
    
    expect(result).toBe('pong');
  });

  test('REQ-UNIT-003: Authentication with valid token', async () => {
    await apiClient.connect();
    const token = await authHelper.generateTestToken('admin');
    const result = await apiClient.authenticate(token);
    
    expect(APIResponseValidator.validateAuthResult(result)).toBe(true);
    expect(result.authenticated).toBe(true);
    expect(result.role).toBe('admin');
  });

  test('REQ-UNIT-004: Camera list retrieval', async () => {
    await apiClient.connect();
    const result = await apiClient.call('get_camera_list');
    
    expect(APIResponseValidator.validateCameraListResult(result)).toBe(true);
    expect(Array.isArray(result.cameras)).toBe(true);
    expect(typeof result.total).toBe('number');
    expect(typeof result.connected).toBe('number');
  });

  test('REQ-UNIT-005: Error handling for invalid method', async () => {
    await apiClient.connect();
    
    await expect(apiClient.call('invalid_method')).rejects.toThrow('Method Not Found');
  });

  test('REQ-UNIT-006: Recording start command', async () => {
    await apiClient.connect();
    const result = await apiClient.call('start_recording', ['camera0']);
    
    expect(APIResponseValidator.validateRecordingStartResult(result)).toBe(true);
    expect(result.device).toBe('camera0');
    expect(['RECORDING', 'STARTING', 'STOPPING', 'PAUSED', 'ERROR', 'FAILED']).toContain(result.status);
  });

  test('REQ-UNIT-007: Recording stop command', async () => {
    await apiClient.connect();
    const result = await apiClient.call('stop_recording', ['camera0']);
    
    expect(APIResponseValidator.validateRecordingStopResult(result)).toBe(true);
    expect(result.device).toBe('camera0');
    expect(['STOPPED', 'FAILED']).toContain(result.status);
  });

  test('REQ-UNIT-008: Snapshot capture command', async () => {
    await apiClient.connect();
    const result = await apiClient.call('take_snapshot', ['camera0']);
    
    expect(APIResponseValidator.validateSnapshotInfo(result)).toBe(true);
    expect(result.device).toBe('camera0');
    expect(['SUCCESS', 'FAILED']).toContain(result.status);
  });

  test('REQ-UNIT-009: File list retrieval', async () => {
    await apiClient.connect();
    const result = await apiClient.call('list_recordings', [50, 0]);
    
    expect(APIResponseValidator.validateListFilesResult(result)).toBe(true);
    expect(Array.isArray(result.files)).toBe(true);
    expect(typeof result.total).toBe('number');
  });

  test('REQ-UNIT-010: System status retrieval', async () => {
    await apiClient.connect();
    const result = await apiClient.call('get_status');
    
    expect(APIResponseValidator.validateStatusResult(result)).toBe(true);
    expect(['HEALTHY', 'DEGRADED', 'UNHEALTHY']).toContain(result.status);
  });
});
