/**
 * REQ-STOR01-001: Storage monitoring must provide accurate disk usage information
 * REQ-STOR01-002: Storage thresholds must trigger appropriate warnings and alerts
 * Coverage: UNIT
 * Quality: HIGH
 */
/**
 * Unit tests for storage store
 * 
 * Design Principles:
 * - Pure unit testing with complete isolation
 * - Direct store testing without React context dependency
 * - Focus on storage monitoring and threshold management
 * - Test storage usage calculations and warning systems
 * - Validate storage monitoring lifecycle
 */

import { useStorageStore } from '../../../src/stores/storageStore';
import type { StorageInfo, StorageUsage, ThresholdStatus } from '../../../src/types/camera';

// Mock the storage monitor service
jest.mock('../../../src/services/storageMonitorService', () => ({
  storageMonitorService: {
    getStorageInfo: jest.fn(),
    getStorageUsage: jest.fn(),
    checkThresholds: jest.fn(),
    startMonitoring: jest.fn(),
    stopMonitoring: jest.fn()
  }
}));

// Mock the logger service
jest.mock('../../../src/services/loggerService', () => ({
  logger: {
    info: jest.fn(),
    warn: jest.fn(),
    error: jest.fn()
  },
  loggers: {
    storage: {
      info: jest.fn(),
      warn: jest.fn(),
      error: jest.fn()
    }
  }
}));

describe('Storage Store', () => {
  let store: ReturnType<typeof useStorageStore.getState>;
  let mockStorageService: any;

  beforeEach(() => {
    // Reset store state completely
    const currentStore = useStorageStore.getState();
    currentStore.reset();
    
    // Get fresh store instance after reset
    store = useStorageStore.getState();
    
    // Get mock service
    mockStorageService = require('../../../src/services/storageMonitorService').storageMonitorService;
    jest.clearAllMocks();
  });

  describe('Initialization', () => {
    it('should start with correct default state', () => {
      const state = useStorageStore.getState();
      expect(state.storageInfo).toBeNull();
      expect(state.storageUsage).toBeNull();
      expect(state.thresholdStatus).toBeNull();
      expect(state.isMonitoring).toBe(false);
      expect(state.monitoringInterval).toBe(30000);
      expect(state.isLoading).toBe(false);
      expect(state.isCheckingThresholds).toBe(false);
      expect(state.error).toBeNull();
      expect(state.lastError).toBeNull();
      expect(state.hasWarnings).toBe(false);
      expect(state.hasCriticalIssues).toBe(false);
      expect(state.warnings).toEqual([]);
    });
  });

  describe('Storage Information Management', () => {
    it('should set storage info', () => {
      const storageInfo: StorageInfo = {
        total_space: 1000000000, // 1GB
        available_space: 500000000, // 500MB
        used_space: 500000000, // 500MB
        mount_point: '/storage',
        filesystem: 'ext4'
      };

      store.setStorageInfo(storageInfo);
      
      const state = useStorageStore.getState();
      expect(state.storageInfo).toEqual(storageInfo);
    });

    it('should get storage info', () => {
      const storageInfo: StorageInfo = {
        total_space: 1000000000,
        available_space: 500000000,
        used_space: 500000000,
        mount_point: '/storage',
        filesystem: 'ext4'
      };

      store.setStorageInfo(storageInfo);
      
      expect(store.getStorageInfo()).toEqual(storageInfo);
    });

    it('should calculate storage usage percentage', () => {
      const storageInfo: StorageInfo = {
        total_space: 1000000000,
        available_space: 300000000,
        used_space: 700000000,
        mount_point: '/storage',
        filesystem: 'ext4'
      };

      store.setStorageInfo(storageInfo);
      
      expect(store.getStorageUsagePercentage()).toBe(70);
    });

    it('should get available space percentage', () => {
      const storageInfo: StorageInfo = {
        total_space: 1000000000,
        available_space: 300000000,
        used_space: 700000000,
        mount_point: '/storage',
        filesystem: 'ext4'
      };

      store.setStorageInfo(storageInfo);
      
      expect(store.getAvailableSpacePercentage()).toBe(30);
    });
  });

  describe('Storage Usage Management', () => {
    it('should set storage usage', () => {
      const storageUsage: StorageUsage = {
        recordings_size: 200000000,
        snapshots_size: 50000000,
        logs_size: 10000000,
        temp_size: 5000000,
        other_size: 235000000
      };

      store.setStorageUsage(storageUsage);
      
      const state = useStorageStore.getState();
      expect(state.storageUsage).toEqual(storageUsage);
    });

    it('should get storage usage', () => {
      const storageUsage: StorageUsage = {
        recordings_size: 200000000,
        snapshots_size: 50000000,
        logs_size: 10000000,
        temp_size: 5000000,
        other_size: 235000000
      };

      store.setStorageUsage(storageUsage);
      
      expect(store.getStorageUsage()).toEqual(storageUsage);
    });

    it('should calculate total used space from usage breakdown', () => {
      const storageUsage: StorageUsage = {
        recordings_size: 200000000,
        snapshots_size: 50000000,
        logs_size: 10000000,
        temp_size: 5000000,
        other_size: 235000000
      };

      store.setStorageUsage(storageUsage);
      
      expect(store.getTotalUsedSpace()).toBe(500000000);
    });

    it('should get usage breakdown by category', () => {
      const storageUsage: StorageUsage = {
        recordings_size: 200000000,
        snapshots_size: 50000000,
        logs_size: 10000000,
        temp_size: 5000000,
        other_size: 235000000
      };

      store.setStorageUsage(storageUsage);
      
      const breakdown = store.getUsageBreakdown();
      expect(breakdown).toEqual({
        recordings: 200000000,
        snapshots: 50000000,
        logs: 10000000,
        temp: 5000000,
        other: 235000000
      });
    });
  });

  describe('Threshold Management', () => {
    it('should set threshold status', () => {
      const thresholdStatus: ThresholdStatus = {
        warning_threshold: 80,
        critical_threshold: 90,
        current_usage: 75,
        status: 'normal',
        warnings: [],
        critical_issues: []
      };

      store.setThresholdStatus(thresholdStatus);
      
      const state = useStorageStore.getState();
      expect(state.thresholdStatus).toEqual(thresholdStatus);
    });

    it('should get threshold status', () => {
      const thresholdStatus: ThresholdStatus = {
        warning_threshold: 80,
        critical_threshold: 90,
        current_usage: 75,
        status: 'normal',
        warnings: [],
        critical_issues: []
      };

      store.setThresholdStatus(thresholdStatus);
      
      expect(store.getThresholdStatus()).toEqual(thresholdStatus);
    });

    it('should check if storage is at warning threshold', () => {
      const thresholdStatus: ThresholdStatus = {
        warning_threshold: 80,
        critical_threshold: 90,
        current_usage: 85,
        status: 'warning',
        warnings: ['Storage usage is high'],
        critical_issues: []
      };

      store.setThresholdStatus(thresholdStatus);
      
      expect(store.isAtWarningThreshold()).toBe(true);
    });

    it('should check if storage is at critical threshold', () => {
      const thresholdStatus: ThresholdStatus = {
        warning_threshold: 80,
        critical_threshold: 90,
        current_usage: 95,
        status: 'critical',
        warnings: ['Storage usage is high'],
        critical_issues: ['Storage is critically low']
      };

      store.setThresholdStatus(thresholdStatus);
      
      expect(store.isAtCriticalThreshold()).toBe(true);
    });

    it('should get current threshold level', () => {
      const thresholdStatus: ThresholdStatus = {
        warning_threshold: 80,
        critical_threshold: 90,
        current_usage: 75,
        status: 'normal',
        warnings: [],
        critical_issues: []
      };

      store.setThresholdStatus(thresholdStatus);
      
      expect(store.getCurrentThresholdLevel()).toBe('normal');
    });
  });

  describe('Warning Management', () => {
    it('should set warnings', () => {
      const warnings = ['Storage usage is high', 'Temp files accumulating'];
      store.setWarnings(warnings);
      
      const state = useStorageStore.getState();
      expect(state.warnings).toEqual(warnings);
      expect(state.hasWarnings).toBe(true);
    });

    it('should add warning', () => {
      store.addWarning('Storage usage is high');
      
      const state = useStorageStore.getState();
      expect(state.warnings).toContain('Storage usage is high');
      expect(state.hasWarnings).toBe(true);
    });

    it('should clear warnings', () => {
      store.setWarnings(['Warning 1', 'Warning 2']);
      store.clearWarnings();
      
      const state = useStorageStore.getState();
      expect(state.warnings).toEqual([]);
      expect(state.hasWarnings).toBe(false);
    });

    it('should set critical issues', () => {
      const issues = ['Storage critically low', 'Cannot write files'];
      store.setCriticalIssues(issues);
      
      const state = useStorageStore.getState();
      expect(state.criticalIssues).toEqual(issues);
      expect(state.hasCriticalIssues).toBe(true);
    });

    it('should add critical issue', () => {
      store.addCriticalIssue('Storage critically low');
      
      const state = useStorageStore.getState();
      expect(state.criticalIssues).toContain('Storage critically low');
      expect(state.hasCriticalIssues).toBe(true);
    });

    it('should clear critical issues', () => {
      store.setCriticalIssues(['Issue 1', 'Issue 2']);
      store.clearCriticalIssues();
      
      const state = useStorageStore.getState();
      expect(state.criticalIssues).toEqual([]);
      expect(state.hasCriticalIssues).toBe(false);
    });
  });

  describe('Monitoring Management', () => {
    it('should set monitoring state', () => {
      store.setMonitoring(true);
      let state = useStorageStore.getState();
      expect(state.isMonitoring).toBe(true);

      store.setMonitoring(false);
      state = useStorageStore.getState();
      expect(state.isMonitoring).toBe(false);
    });

    it('should set monitoring interval', () => {
      store.setMonitoringInterval(60000);
      
      const state = useStorageStore.getState();
      expect(state.monitoringInterval).toBe(60000);
    });

    it('should start monitoring', async () => {
      mockStorageService.startMonitoring.mockResolvedValue(undefined);

      await store.startMonitoring();

      const state = useStorageStore.getState();
      expect(state.isMonitoring).toBe(true);
      expect(mockStorageService.startMonitoring).toHaveBeenCalled();
    });

    it('should stop monitoring', async () => {
      mockStorageService.stopMonitoring.mockResolvedValue(undefined);

      await store.stopMonitoring();

      const state = useStorageStore.getState();
      expect(state.isMonitoring).toBe(false);
      expect(mockStorageService.stopMonitoring).toHaveBeenCalled();
    });
  });

  describe('Loading State Management', () => {
    it('should set loading state', () => {
      store.setLoading(true);
      let state = useStorageStore.getState();
      expect(state.isLoading).toBe(true);

      store.setLoading(false);
      state = useStorageStore.getState();
      expect(state.isLoading).toBe(false);
    });

    it('should set checking thresholds state', () => {
      store.setCheckingThresholds(true);
      let state = useStorageStore.getState();
      expect(state.isCheckingThresholds).toBe(true);

      store.setCheckingThresholds(false);
      state = useStorageStore.getState();
      expect(state.isCheckingThresholds).toBe(false);
    });
  });

  describe('Error Management', () => {
    it('should set error', () => {
      store.setError('Storage check failed');
      let state = useStorageStore.getState();
      expect(state.error).toBe('Storage check failed');

      store.setError(null);
      state = useStorageStore.getState();
      expect(state.error).toBeNull();
    });

    it('should set last error', () => {
      store.setLastError('Last storage error');
      let state = useStorageStore.getState();
      expect(state.lastError).toBe('Last storage error');

      store.setLastError(null);
      state = useStorageStore.getState();
      expect(state.lastError).toBeNull();
    });

    it('should clear all errors', () => {
      store.setError('Current error');
      store.setLastError('Last error');
      
      store.clearAllErrors();
      
      const state = useStorageStore.getState();
      expect(state.error).toBeNull();
      expect(state.lastError).toBeNull();
    });
  });

  describe('Storage Operations', () => {
    it('should refresh storage info successfully', async () => {
      const mockStorageInfo: StorageInfo = {
        total_space: 1000000000,
        available_space: 500000000,
        used_space: 500000000,
        mount_point: '/storage',
        filesystem: 'ext4'
      };

      mockStorageService.getStorageInfo.mockResolvedValue(mockStorageInfo);

      await store.refreshStorageInfo();

      const state = useStorageStore.getState();
      expect(state.storageInfo).toEqual(mockStorageInfo);
      expect(state.isLoading).toBe(false);
      expect(state.error).toBeNull();
    });

    it('should handle refresh storage info failure', async () => {
      const error = new Error('Storage info unavailable');
      mockStorageService.getStorageInfo.mockRejectedValue(error);

      await store.refreshStorageInfo();

      const state = useStorageStore.getState();
      expect(state.isLoading).toBe(false);
      expect(state.error).toBe('Storage info unavailable');
    });

    it('should refresh storage usage successfully', async () => {
      const mockStorageUsage: StorageUsage = {
        recordings_size: 200000000,
        snapshots_size: 50000000,
        logs_size: 10000000,
        temp_size: 5000000,
        other_size: 235000000
      };

      mockStorageService.getStorageUsage.mockResolvedValue(mockStorageUsage);

      await store.refreshStorageUsage();

      const state = useStorageStore.getState();
      expect(state.storageUsage).toEqual(mockStorageUsage);
      expect(state.isLoading).toBe(false);
      expect(state.error).toBeNull();
    });

    it('should check thresholds successfully', async () => {
      const mockThresholdStatus: ThresholdStatus = {
        warning_threshold: 80,
        critical_threshold: 90,
        current_usage: 75,
        status: 'normal',
        warnings: [],
        critical_issues: []
      };

      mockStorageService.checkThresholds.mockResolvedValue(mockThresholdStatus);

      await store.checkThresholds();

      const state = useStorageStore.getState();
      expect(state.thresholdStatus).toEqual(mockThresholdStatus);
      expect(state.isCheckingThresholds).toBe(false);
      expect(state.error).toBeNull();
    });

    it('should handle check thresholds failure', async () => {
      const error = new Error('Threshold check failed');
      mockStorageService.checkThresholds.mockRejectedValue(error);

      await store.checkThresholds();

      const state = useStorageStore.getState();
      expect(state.isCheckingThresholds).toBe(false);
      expect(state.error).toBe('Threshold check failed');
    });
  });

  describe('Storage Analysis', () => {
    it('should get storage health status', () => {
      const storageInfo: StorageInfo = {
        total_space: 1000000000,
        available_space: 200000000,
        used_space: 800000000,
        mount_point: '/storage',
        filesystem: 'ext4'
      };

      const thresholdStatus: ThresholdStatus = {
        warning_threshold: 80,
        critical_threshold: 90,
        current_usage: 80,
        status: 'warning',
        warnings: ['Storage usage is high'],
        critical_issues: []
      };

      store.setStorageInfo(storageInfo);
      store.setThresholdStatus(thresholdStatus);

      const healthStatus = store.getStorageHealthStatus();
      expect(healthStatus).toEqual({
        status: 'warning',
        usage_percentage: 80,
        available_percentage: 20,
        has_warnings: true,
        has_critical_issues: false,
        warnings: ['Storage usage is high'],
        critical_issues: []
      });
    });

    it('should get storage summary', () => {
      const storageInfo: StorageInfo = {
        total_space: 1000000000,
        available_space: 500000000,
        used_space: 500000000,
        mount_point: '/storage',
        filesystem: 'ext4'
      };

      const storageUsage: StorageUsage = {
        recordings_size: 200000000,
        snapshots_size: 50000000,
        logs_size: 10000000,
        temp_size: 5000000,
        other_size: 235000000
      };

      store.setStorageInfo(storageInfo);
      store.setStorageUsage(storageUsage);

      const summary = store.getStorageSummary();
      expect(summary).toEqual({
        total_space: 1000000000,
        available_space: 500000000,
        used_space: 500000000,
        usage_percentage: 50,
        available_percentage: 50,
        usage_breakdown: {
          recordings: 200000000,
          snapshots: 50000000,
          logs: 10000000,
          temp: 5000000,
          other: 235000000
        }
      });
    });
  });

  describe('State Reset', () => {
    it('should reset all state to initial values', () => {
      // Set some state
      store.setStorageInfo({
        total_space: 1000000000,
        available_space: 500000000,
        used_space: 500000000,
        mount_point: '/storage',
        filesystem: 'ext4'
      });
      store.setWarnings(['Test warning']);
      store.setError('Test error');
      store.setMonitoring(true);
      
      // Reset
      store.reset();
      
      const state = useStorageStore.getState();
      expect(state.storageInfo).toBeNull();
      expect(state.storageUsage).toBeNull();
      expect(state.thresholdStatus).toBeNull();
      expect(state.isMonitoring).toBe(false);
      expect(state.isLoading).toBe(false);
      expect(state.isCheckingThresholds).toBe(false);
      expect(state.error).toBeNull();
      expect(state.lastError).toBeNull();
      expect(state.hasWarnings).toBe(false);
      expect(state.hasCriticalIssues).toBe(false);
      expect(state.warnings).toEqual([]);
    });
  });
});
