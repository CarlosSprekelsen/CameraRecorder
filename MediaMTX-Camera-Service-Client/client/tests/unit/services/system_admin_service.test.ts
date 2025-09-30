/**
 * SystemAdminService unit tests for missing RPC methods
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * 
 * Requirements Coverage:
 * - REQ-ADMIN-001: get_metrics RPC method
 * - REQ-ADMIN-002: get_storage_info RPC method
 * - REQ-ADMIN-003: set_retention_policy RPC method
 * - REQ-ADMIN-004: cleanup_old_files RPC method
 * 
 * Test Categories: Unit
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */

import { ServerService } from '../../../src/services/server/ServerService';
import { IAPIClient } from '../../../src/services/abstraction/IAPIClient';
import { LoggerService } from '../../../src/services/logger/LoggerService';
import { MetricsResult, StorageInfo, RetentionPolicySetResult, CleanupResult } from '../../../src/types/api';
import { MockDataFactory } from '../../utils/mocks';

// Use centralized mocks - aligned with refactored architecture
const mockAPIClient = MockDataFactory.createMockAPIClient();
const mockLoggerService = MockDataFactory.createMockLoggerService();

// Use real ServerService with APIClient for admin functions

describe('SystemAdminService Unit Tests', () => {
  let systemAdminService: ServerService;

  beforeEach(() => {
    jest.clearAllMocks();
    systemAdminService = new ServerService(mockAPIClient, mockLoggerService);
  });

  describe('REQ-ADMIN-001: get_metrics RPC method', () => {
    test('Should call WebSocket service with correct parameters', async () => {
      // Arrange
      const expectedResult: MetricsResult = {
        timestamp: '2025-01-15T14:30:00Z',
        system_metrics: {
          cpu_usage: 23.1,
          memory_usage: 85.5,
          disk_usage: 45.5,
          goroutines: 150
        },
        camera_metrics: {
          connected_cameras: 2,
          cameras: {
            'camera0': {
              path: 'camera0',
              name: 'USB 2.0 Camera: USB 2.0 Camera',
              status: 'CONNECTED',
              device_num: 0,
              last_seen: '2025-01-15T14:30:00Z',
              capabilities: {
                driver_name: 'uvcvideo',
                card_name: 'USB 2.0 Camera: USB 2.0 Camera',
                bus_info: 'usb-0000:00:1a.0-1.2',
                version: '6.14.8',
                capabilities: ['0x84a00001', 'Video Capture', 'Metadata Capture', 'Streaming', 'Extended Pix Format'],
                device_caps: ['0x04200001', 'Video Capture', 'Streaming', 'Extended Pix Format']
              },
              formats: [
                {
                  pixel_format: 'YUYV',
                  width: 640,
                  height: 480,
                  frame_rates: ['30.000', '20.000', '15.000', '10.000', '5.000']
                }
              ]
            }
          }
        },
        recording_metrics: {},
        stream_metrics: {
          active_streams: 0,
          total_streams: 4,
          total_viewers: 0
        }
      };

      (mockAPIClient.call as jest.Mock).mockResolvedValue(expectedResult);

      // Act
      const result = await systemAdminService.getMetrics();

      // Assert
      expect(mockLoggerService.info).toHaveBeenCalledWith('get_metrics request');
      expect(mockAPIClient.call).toHaveBeenCalledWith('get_metrics');
      expect(result).toEqual(expectedResult);
    });

    test('Should handle errors correctly', async () => {
      // Arrange
      const error = new Error('Get metrics failed');
      (mockAPIClient.call as jest.Mock).mockRejectedValue(error);

      // Act & Assert
      await expect(systemAdminService.getMetrics()).rejects.toThrow(error);
      expect(mockLoggerService.error).toHaveBeenCalledWith('get_metrics failed', error);
    });
  });

  describe('REQ-ADMIN-002: get_storage_info RPC method', () => {
    test('Should call WebSocket service with correct parameters', async () => {
      // Arrange
      const expectedResult: StorageInfo = {
        total_space: 107374182400,
        used_space: 53687091200,
        available_space: 53687091200,
        usage_percentage: 50.0,
        recordings_size: 42949672960,
        snapshots_size: 10737418240,
        low_space_warning: false
      };

      (mockAPIClient.call as jest.Mock).mockResolvedValue(expectedResult);

      // Act
      const result = await systemAdminService.getStorageInfo();

      // Assert
      expect(mockLoggerService.info).toHaveBeenCalledWith('get_storage_info request');
      expect(mockAPIClient.call).toHaveBeenCalledWith('get_storage_info');
      expect(result).toEqual(expectedResult);
    });

    test('Should handle errors correctly', async () => {
      // Arrange
      const error = new Error('Get storage info failed');
      (mockAPIClient.call as jest.Mock).mockRejectedValue(error);

      // Act & Assert
      await expect(systemAdminService.getStorageInfo()).rejects.toThrow(error);
      expect(mockLoggerService.error).toHaveBeenCalledWith('get_storage_info failed', error);
    });
  });

  describe('REQ-ADMIN-003: set_retention_policy RPC method', () => {
    test('Should call WebSocket service with age-based policy', async () => {
      // Arrange
      const policy = {
        policy_type: 'age',
        max_age_days: 30,
        enabled: true
      };
      const expectedResult: RetentionPolicySetResult = {
        policy_type: 'age',
        max_age_days: 30,
        enabled: true,
        message: 'Retention policy configured successfully'
      };

      (mockAPIClient.call as jest.Mock).mockResolvedValue(expectedResult);

      // Act
      const result = await systemAdminService.setRetentionPolicy(policy);

      // Assert
      expect(mockLoggerService.info).toHaveBeenCalledWith('set_retention_policy request', policy);
      expect(mockAPIClient.call).toHaveBeenCalledWith('set_retention_policy', policy);
      expect(result).toEqual(expectedResult);
    });

    test('Should call WebSocket service with size-based policy', async () => {
      // Arrange
      const policy = {
        policy_type: 'size',
        max_size_gb: 100,
        enabled: true
      };
      const expectedResult: RetentionPolicySetResult = {
        policy_type: 'size',
        max_size_gb: 100,
        enabled: true,
        message: 'Retention policy configured successfully'
      };

      (mockAPIClient.call as jest.Mock).mockResolvedValue(expectedResult);

      // Act
      const result = await systemAdminService.setRetentionPolicy(policy);

      // Assert
      expect(mockLoggerService.info).toHaveBeenCalledWith('set_retention_policy request', policy);
      expect(mockAPIClient.call).toHaveBeenCalledWith('set_retention_policy', policy);
      expect(result).toEqual(expectedResult);
    });

    test('Should call WebSocket service with manual policy', async () => {
      // Arrange
      const policy = {
        policy_type: 'manual',
        enabled: false
      };
      const expectedResult: RetentionPolicySetResult = {
        policy_type: 'manual',
        enabled: false,
        message: 'Retention policy configured successfully'
      };

      (mockAPIClient.call as jest.Mock).mockResolvedValue(expectedResult);

      // Act
      const result = await systemAdminService.setRetentionPolicy(policy);

      // Assert
      expect(mockLoggerService.info).toHaveBeenCalledWith('set_retention_policy request', policy);
      expect(mockAPIClient.call).toHaveBeenCalledWith('set_retention_policy', policy);
      expect(result).toEqual(expectedResult);
    });

    test('Should handle errors correctly', async () => {
      // Arrange
      const policy = {
        policy_type: 'age',
        max_age_days: 30,
        enabled: true
      };
      const error = new Error('Set retention policy failed');
      (mockAPIClient.call as jest.Mock).mockRejectedValue(error);

      // Act & Assert
      await expect(systemAdminService.setRetentionPolicy(policy)).rejects.toThrow(error);
      expect(mockLoggerService.error).toHaveBeenCalledWith('set_retention_policy failed', error);
    });
  });

  describe('REQ-ADMIN-004: cleanup_old_files RPC method', () => {
    test('Should call WebSocket service with correct parameters', async () => {
      // Arrange
      const expectedResult: CleanupResult = {
        cleanup_executed: true,
        files_deleted: 15,
        space_freed: 10737418240,
        message: 'Cleanup completed successfully'
      };

      (mockAPIClient.call as jest.Mock).mockResolvedValue(expectedResult);

      // Act
      const result = await systemAdminService.cleanupOldFiles();

      // Assert
      expect(mockLoggerService.info).toHaveBeenCalledWith('cleanup_old_files request');
      expect(mockAPIClient.call).toHaveBeenCalledWith('cleanup_old_files');
      expect(result).toEqual(expectedResult);
    });

    test('Should handle errors correctly', async () => {
      // Arrange
      const error = new Error('Cleanup old files failed');
      (mockAPIClient.call as jest.Mock).mockRejectedValue(error);

      // Act & Assert
      await expect(systemAdminService.cleanupOldFiles()).rejects.toThrow(error);
      expect(mockLoggerService.error).toHaveBeenCalledWith('cleanup_old_files failed', error);
    });
  });

  describe('REQ-ADMIN-005: System administration workflow', () => {
    test('Should handle complete admin workflow', async () => {
      // Arrange
      const metricsResult: MetricsResult = {
        timestamp: '2025-01-15T14:30:00Z',
        system_metrics: { cpu_usage: 23.1, memory_usage: 85.5, disk_usage: 45.5, goroutines: 150 },
        camera_metrics: { connected_cameras: 2, cameras: {} },
        recording_metrics: {},
        stream_metrics: { active_streams: 0, total_streams: 4, total_viewers: 0 }
      };

      const storageResult: StorageInfo = {
        total_space: 107374182400,
        used_space: 53687091200,
        available_space: 53687091200,
        usage_percentage: 50.0,
        recordings_size: 42949672960,
        snapshots_size: 10737418240,
        low_space_warning: false
      };

      const policyResult: RetentionPolicySetResult = {
        policy_type: 'age',
        max_age_days: 30,
        enabled: true,
        message: 'Retention policy configured successfully'
      };

      const cleanupResult: CleanupResult = {
        cleanup_executed: true,
        files_deleted: 15,
        space_freed: 10737418240,
        message: 'Cleanup completed successfully'
      };

      (mockAPIClient.call as jest.Mock)
        .mockResolvedValueOnce(metricsResult)
        .mockResolvedValueOnce(storageResult)
        .mockResolvedValueOnce(policyResult)
        .mockResolvedValueOnce(cleanupResult);

      // Act
      const metrics = await systemAdminService.getMetrics();
      const storage = await systemAdminService.getStorageInfo();
      const policy = await systemAdminService.setRetentionPolicy({
        policy_type: 'age',
        max_age_days: 30,
        enabled: true
      });
      const cleanup = await systemAdminService.cleanupOldFiles();

      // Assert
      expect(metrics).toEqual(metricsResult);
      expect(storage).toEqual(storageResult);
      expect(policy).toEqual(policyResult);
      expect(cleanup).toEqual(cleanupResult);
      expect(mockAPIClient.call).toHaveBeenCalledTimes(4);
    });
  });
});
