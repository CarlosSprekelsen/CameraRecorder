/**
 * Storage Monitor Service Unit Tests
 * 
 * Ground Truth References:
 * - Server API: ../mediamtx-camera-service/docs/api/json-rpc-methods.md
 * - Client Architecture: ../docs/architecture/client-architecture.md
 * - Client Requirements: ../docs/requirements/client-requirements.md
 * 
 * Requirements Coverage:
 * - REQ-STOR01-001: Storage monitoring must be accurate and real-time
 * - REQ-STOR01-002: Storage thresholds must trigger appropriate warnings
 * - REQ-STOR01-003: Storage cleanup must be automatic and safe
 * - REQ-STOR01-004: Storage metrics must be comprehensive
 * 
 * Coverage: UNIT
 * Quality: HIGH
 */

import { StorageMonitorService } from '../../../src/services/storageMonitorService';
import { websocketService } from '../../../src/services/websocket';
import { RPC_METHODS } from '../../../src/types/rpc';

// Mock dependencies
jest.mock('../../../src/services/websocket', () => ({
  websocketService: {
    call: jest.fn(),
    isConnected: jest.fn(() => true),
  },
}));

jest.mock('../../../src/services/loggerService', () => ({
  logger: {
    info: jest.fn(),
    error: jest.fn(),
    warn: jest.fn(),
    debug: jest.fn(),
  },
  loggers: {
    storage: {
      info: jest.fn(),
      error: jest.fn(),
      warn: jest.fn(),
      debug: jest.fn(),
    },
  },
}));

describe('Storage Monitor Service', () => {
  let storageMonitor: StorageMonitorService;
  const mockWebSocketService = websocketService as jest.Mocked<typeof websocketService>;

  beforeEach(() => {
    storageMonitor = new StorageMonitorService();
    jest.clearAllMocks();
  });

  afterEach(() => {
    storageMonitor.cleanup();
  });

  describe('REQ-STOR01-001: Accurate and Real-time Storage Monitoring', () => {
    it('should get current storage information', async () => {
      const mockStorageInfo = {
        total_space: 1000000000, // 1GB
        used_space: 500000000,   // 500MB
        free_space: 500000000,   // 500MB
        usage_percentage: 50,
        mount_point: '/storage',
        device: '/dev/sda1',
      };

      const mockResponse = {
        jsonrpc: '2.0' as const,
        result: mockStorageInfo,
        id: 1,
      };

      mockWebSocketService.call.mockResolvedValue(mockResponse);

      const result = await storageMonitor.getStorageInfo();

      expect(mockWebSocketService.call).toHaveBeenCalledWith(RPC_METHODS.GET_STORAGE_INFO, {});
      expect(result).toEqual(mockStorageInfo);
    });

    it('should monitor storage usage in real-time', async () => {
      const mockStorageInfo = {
        total_space: 1000000000,
        used_space: 600000000,
        free_space: 400000000,
        usage_percentage: 60,
        mount_point: '/storage',
        device: '/dev/sda1',
      };

      const mockResponse = {
        jsonrpc: '2.0' as const,
        result: mockStorageInfo,
        id: 1,
      };

      mockWebSocketService.call.mockResolvedValue(mockResponse);

      const result = await storageMonitor.getStorageUsage();

      expect(result.usage_percentage).toBe(60);
      expect(result.free_space).toBe(400000000);
    });

    it('should handle storage monitoring errors', async () => {
      mockWebSocketService.call.mockRejectedValue(new Error('Storage monitoring failed'));

      await expect(storageMonitor.getStorageInfo()).rejects.toThrow('Storage monitoring failed');
    });

    it('should provide storage statistics', async () => {
      const mockStats = {
        total_files: 1000,
        total_size: 500000000,
        average_file_size: 500000,
        oldest_file: '2024-01-01T00:00:00Z',
        newest_file: '2024-01-15T12:00:00Z',
      };

      const mockResponse = {
        jsonrpc: '2.0' as const,
        result: mockStats,
        id: 1,
      };

      mockWebSocketService.call.mockResolvedValue(mockResponse);

      const result = await storageMonitor.getStorageStatistics();

      expect(result).toEqual(mockStats);
    });
  });

  describe('REQ-STOR01-002: Storage Threshold Warnings', () => {
    it('should check storage thresholds', async () => {
      const mockThresholdStatus = {
        warning_threshold: 80,
        critical_threshold: 90,
        current_usage: 85,
        status: 'warning',
        message: 'Storage usage is above warning threshold',
      };

      const mockResponse = {
        jsonrpc: '2.0' as const,
        result: mockThresholdStatus,
        id: 1,
      };

      mockWebSocketService.call.mockResolvedValue(mockResponse);

      const result = await storageMonitor.checkStorageThresholds();

      expect(result.status).toBe('warning');
      expect(result.current_usage).toBe(85);
    });

    it('should trigger warning notifications', async () => {
      const mockThresholdStatus = {
        warning_threshold: 80,
        critical_threshold: 90,
        current_usage: 85,
        status: 'warning',
        message: 'Storage usage is above warning threshold',
      };

      const mockResponse = {
        jsonrpc: '2.0' as const,
        result: mockThresholdStatus,
        id: 1,
      };

      mockWebSocketService.call.mockResolvedValue(mockResponse);

      const onThresholdExceeded = jest.fn();
      storageMonitor.onStorageThresholdExceeded(onThresholdExceeded);

      await storageMonitor.checkStorageThresholds();

      expect(onThresholdExceeded).toHaveBeenCalledWith(mockThresholdStatus);
    });

    it('should trigger critical notifications', async () => {
      const mockThresholdStatus = {
        warning_threshold: 80,
        critical_threshold: 90,
        current_usage: 95,
        status: 'critical',
        message: 'Storage usage is critical',
      };

      const mockResponse = {
        jsonrpc: '2.0' as const,
        result: mockThresholdStatus,
        id: 1,
      };

      mockWebSocketService.call.mockResolvedValue(mockResponse);

      const onThresholdExceeded = jest.fn();
      storageMonitor.onStorageThresholdExceeded(onThresholdExceeded);

      await storageMonitor.checkStorageThresholds();

      expect(onThresholdExceeded).toHaveBeenCalledWith(mockThresholdStatus);
    });

    it('should configure custom thresholds', () => {
      const customThresholds = {
        warning_threshold: 70,
        critical_threshold: 85,
      };

      storageMonitor.configureThresholds(customThresholds);

      const config = storageMonitor.getThresholdConfig();
      expect(config.warning_threshold).toBe(70);
      expect(config.critical_threshold).toBe(85);
    });
  });

  describe('REQ-STOR01-003: Automatic and Safe Storage Cleanup', () => {
    it('should cleanup old files', async () => {
      const mockCleanupResult = {
        files_deleted: 50,
        space_freed: 100000000,
        errors: [],
        duration: 5000,
      };

      const mockResponse = {
        jsonrpc: '2.0' as const,
        result: mockCleanupResult,
        id: 1,
      };

      mockWebSocketService.call.mockResolvedValue(mockResponse);

      const result = await storageMonitor.cleanupOldFiles();

      expect(mockWebSocketService.call).toHaveBeenCalledWith(RPC_METHODS.CLEANUP_OLD_FILES, {});
      expect(result.files_deleted).toBe(50);
      expect(result.space_freed).toBe(100000000);
    });

    it('should cleanup files older than specified age', async () => {
      const mockCleanupResult = {
        files_deleted: 25,
        space_freed: 50000000,
        errors: [],
        duration: 3000,
      };

      const mockResponse = {
        jsonrpc: '2.0' as const,
        result: mockCleanupResult,
        id: 1,
      };

      mockWebSocketService.call.mockResolvedValue(mockResponse);

      const result = await storageMonitor.cleanupOldFiles({ older_than_days: 30 });

      expect(mockWebSocketService.call).toHaveBeenCalledWith(RPC_METHODS.CLEANUP_OLD_FILES, {
        older_than_days: 30,
      });
      expect(result.files_deleted).toBe(25);
    });

    it('should handle cleanup errors safely', async () => {
      const mockCleanupResult = {
        files_deleted: 10,
        space_freed: 20000000,
        errors: ['Permission denied for file1.mp4', 'File in use: file2.mp4'],
        duration: 2000,
      };

      const mockResponse = {
        jsonrpc: '2.0' as const,
        result: mockCleanupResult,
        id: 1,
      };

      mockWebSocketService.call.mockResolvedValue(mockResponse);

      const result = await storageMonitor.cleanupOldFiles();

      expect(result.errors).toHaveLength(2);
      expect(result.files_deleted).toBe(10);
    });

    it('should prevent cleanup of active recordings', async () => {
      const mockCleanupResult = {
        files_deleted: 0,
        space_freed: 0,
        errors: ['Cannot delete active recording files'],
        duration: 1000,
      };

      const mockResponse = {
        jsonrpc: '2.0' as const,
        result: mockCleanupResult,
        id: 1,
      };

      mockWebSocketService.call.mockResolvedValue(mockResponse);

      const result = await storageMonitor.cleanupOldFiles({ exclude_active: true });

      expect(mockWebSocketService.call).toHaveBeenCalledWith(RPC_METHODS.CLEANUP_OLD_FILES, {
        exclude_active: true,
      });
      expect(result.files_deleted).toBe(0);
    });
  });

  describe('REQ-STOR01-004: Comprehensive Storage Metrics', () => {
    it('should provide detailed storage metrics', async () => {
      const mockMetrics = {
        total_space: 1000000000,
        used_space: 600000000,
        free_space: 400000000,
        usage_percentage: 60,
        files_by_type: {
          mp4: 100,
          jpg: 500,
          png: 50,
        },
        size_by_type: {
          mp4: 500000000,
          jpg: 80000000,
          png: 20000000,
        },
        growth_rate: 0.05, // 5% per day
        projected_full_date: '2024-02-15T00:00:00Z',
      };

      const mockResponse = {
        jsonrpc: '2.0' as const,
        result: mockMetrics,
        id: 1,
      };

      mockWebSocketService.call.mockResolvedValue(mockResponse);

      const result = await storageMonitor.getDetailedMetrics();

      expect(result.usage_percentage).toBe(60);
      expect(result.files_by_type.mp4).toBe(100);
      expect(result.growth_rate).toBe(0.05);
    });

    it('should track storage trends over time', async () => {
      const mockTrends = {
        daily_usage: [
          { date: '2024-01-01', usage: 50 },
          { date: '2024-01-02', usage: 52 },
          { date: '2024-01-03', usage: 55 },
        ],
        weekly_growth: 0.1,
        monthly_projection: 80,
      };

      const mockResponse = {
        jsonrpc: '2.0' as const,
        result: mockTrends,
        id: 1,
      };

      mockWebSocketService.call.mockResolvedValue(mockResponse);

      const result = await storageMonitor.getStorageTrends();

      expect(result.daily_usage).toHaveLength(3);
      expect(result.weekly_growth).toBe(0.1);
    });

    it('should provide storage health score', async () => {
      const mockHealthScore = {
        score: 75,
        factors: {
          usage_percentage: 60,
          growth_rate: 0.05,
          fragmentation: 0.1,
          error_rate: 0.02,
        },
        recommendations: [
          'Consider increasing storage capacity',
          'Monitor growth rate closely',
        ],
      };

      const mockResponse = {
        jsonrpc: '2.0' as const,
        result: mockHealthScore,
        id: 1,
      };

      mockWebSocketService.call.mockResolvedValue(mockResponse);

      const result = await storageMonitor.getStorageHealthScore();

      expect(result.score).toBe(75);
      expect(result.recommendations).toHaveLength(2);
    });

    it('should generate storage reports', async () => {
      const mockReport = {
        summary: {
          total_space: 1000000000,
          used_space: 600000000,
          free_space: 400000000,
          usage_percentage: 60,
        },
        trends: {
          daily_growth: 0.05,
          weekly_growth: 0.1,
        },
        recommendations: [
          'Storage usage is within normal limits',
          'Monitor growth rate',
        ],
        generated_at: '2024-01-15T12:00:00Z',
      };

      const mockResponse = {
        jsonrpc: '2.0' as const,
        result: mockReport,
        id: 1,
      };

      mockWebSocketService.call.mockResolvedValue(mockResponse);

      const result = await storageMonitor.generateStorageReport();

      expect(result.summary.usage_percentage).toBe(60);
      expect(result.recommendations).toHaveLength(2);
    });
  });

  describe('Event Handling', () => {
    it('should emit storage events', () => {
      const onStorageUpdate = jest.fn();
      const onThresholdExceeded = jest.fn();

      storageMonitor.on('storageUpdate', onStorageUpdate);
      storageMonitor.on('thresholdExceeded', onThresholdExceeded);

      // Simulate storage update
      storageMonitor.emit('storageUpdate', { usage_percentage: 70 });

      expect(onStorageUpdate).toHaveBeenCalledWith({ usage_percentage: 70 });
    });

    it('should remove event listeners', () => {
      const onStorageUpdate = jest.fn();
      storageMonitor.on('storageUpdate', onStorageUpdate);
      storageMonitor.off('storageUpdate', onStorageUpdate);

      storageMonitor.emit('storageUpdate', { usage_percentage: 70 });

      expect(onStorageUpdate).not.toHaveBeenCalled();
    });
  });

  describe('Configuration', () => {
    it('should configure monitoring settings', () => {
      const config = {
        monitoring_interval: 30000,
        threshold_check_interval: 60000,
        cleanup_interval: 3600000,
        enable_automatic_cleanup: true,
      };

      storageMonitor.configure(config);

      expect(storageMonitor.getConfig()).toMatchObject(config);
    });

    it('should validate configuration options', () => {
      const invalidConfig = {
        monitoring_interval: -1,
        threshold_check_interval: 0,
      };

      expect(() => {
        storageMonitor.configure(invalidConfig);
      }).toThrow('Invalid configuration');
    });
  });

  describe('Performance', () => {
    it('should handle high-frequency monitoring', async () => {
      const mockStorageInfo = {
        total_space: 1000000000,
        used_space: 500000000,
        free_space: 500000000,
        usage_percentage: 50,
        mount_point: '/storage',
        device: '/dev/sda1',
      };

      const mockResponse = {
        jsonrpc: '2.0' as const,
        result: mockStorageInfo,
        id: 1,
      };

      mockWebSocketService.call.mockResolvedValue(mockResponse);

      const startTime = Date.now();
      
      // Perform multiple monitoring calls
      const promises = Array.from({ length: 100 }, () => storageMonitor.getStorageInfo());
      await Promise.all(promises);

      const endTime = Date.now();
      const duration = endTime - startTime;

      // Should complete within reasonable time
      expect(duration).toBeLessThan(5000);
      expect(mockWebSocketService.call).toHaveBeenCalledTimes(100);
    });
  });
});
