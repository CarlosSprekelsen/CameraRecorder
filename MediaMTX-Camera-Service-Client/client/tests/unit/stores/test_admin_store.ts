/**
 * REQ-ADMIN01-001: Admin functionality must provide comprehensive system management
 * REQ-ADMIN01-002: System monitoring must provide accurate performance metrics
 * Coverage: UNIT
 * Quality: HIGH
 */
/**
 * Unit tests for admin store
 * 
 * Design Principles:
 * - Pure unit testing with complete isolation
 * - Direct store testing without React context dependency
 * - Focus on admin functionality and system management
 * - Test system metrics and status monitoring
 * - Validate admin operations and configuration management
 */

import { useAdminStore } from '../../../src/stores/adminStore';
import type { SystemMetrics, SystemStatus, ServerInfo, StorageInfo } from '../../../src/stores/adminStore';

// Mock the admin service
jest.mock('../../../src/services/adminService', () => ({
  adminService: {
    getSystemMetrics: jest.fn(),
    getSystemStatus: jest.fn(),
    getServerInfo: jest.fn(),
    getStorageInfo: jest.fn(),
    restartService: jest.fn(),
    updateConfiguration: jest.fn(),
    clearLogs: jest.fn(),
    exportLogs: jest.fn()
  }
}));

describe('Admin Store', () => {
  let store: ReturnType<typeof useAdminStore.getState>;
  let mockAdminService: any;

  beforeEach(() => {
    // Reset store state completely
    const currentStore = useAdminStore.getState();
    currentStore.reset();
    
    // Get fresh store instance after reset
    store = useAdminStore.getState();
    
    // Get mock service
    mockAdminService = require('../../../src/services/adminService').adminService;
    jest.clearAllMocks();
  });

  describe('Initialization', () => {
    it('should start with correct default state', () => {
      const state = useAdminStore.getState();
      expect(state.systemMetrics).toBeNull();
      expect(state.systemStatus).toBeNull();
      expect(state.serverInfo).toBeNull();
      expect(state.storageInfo).toBeNull();
      expect(state.isLoading).toBe(false);
      expect(state.isRefreshing).toBe(false);
      expect(state.isRestarting).toBe(false);
      expect(state.isUpdatingConfig).toBe(false);
      expect(state.isExportingLogs).toBe(false);
      expect(state.isClearingLogs).toBe(false);
      expect(state.error).toBeNull();
      expect(state.lastError).toBeNull();
      expect(state.lastRefresh).toBeNull();
    });
  });

  describe('System Metrics Management', () => {
    it('should set system metrics', () => {
      const metrics: SystemMetrics = {
        active_connections: 5,
        total_requests: 1000,
        average_response_time: 50,
        error_rate: 0.02,
        memory_usage: 75.5,
        cpu_usage: 45.2
      };

      store.setSystemMetrics(metrics);
      
      const state = useAdminStore.getState();
      expect(state.systemMetrics).toEqual(metrics);
    });

    it('should get system metrics', () => {
      const metrics: SystemMetrics = {
        active_connections: 5,
        total_requests: 1000,
        average_response_time: 50,
        error_rate: 0.02,
        memory_usage: 75.5,
        cpu_usage: 45.2
      };

      store.setSystemMetrics(metrics);
      
      expect(store.getSystemMetrics()).toEqual(metrics);
    });

    it('should check if system is healthy based on metrics', () => {
      const healthyMetrics: SystemMetrics = {
        active_connections: 5,
        total_requests: 1000,
        average_response_time: 50,
        error_rate: 0.02,
        memory_usage: 75.5,
        cpu_usage: 45.2
      };

      store.setSystemMetrics(healthyMetrics);
      expect(store.isSystemHealthy()).toBe(true);

      const unhealthyMetrics: SystemMetrics = {
        active_connections: 5,
        total_requests: 1000,
        average_response_time: 2000,
        error_rate: 0.5,
        memory_usage: 95.0,
        cpu_usage: 90.0
      };

      store.setSystemMetrics(unhealthyMetrics);
      expect(store.isSystemHealthy()).toBe(false);
    });

    it('should get performance summary', () => {
      const metrics: SystemMetrics = {
        active_connections: 5,
        total_requests: 1000,
        average_response_time: 50,
        error_rate: 0.02,
        memory_usage: 75.5,
        cpu_usage: 45.2
      };

      store.setSystemMetrics(metrics);
      
      const summary = store.getPerformanceSummary();
      expect(summary).toEqual({
        connections: 5,
        requests: 1000,
        response_time: 50,
        error_rate: 0.02,
        memory_usage: 75.5,
        cpu_usage: 45.2,
        is_healthy: true
      });
    });
  });

  describe('System Status Management', () => {
    it('should set system status', () => {
      const status: SystemStatus = {
        status: 'healthy',
        uptime: 86400,
        version: '1.0.0',
        components: {
          websocket_server: 'running',
          camera_monitor: 'running',
          mediamtx_controller: 'running'
        }
      };

      store.setSystemStatus(status);
      
      const state = useAdminStore.getState();
      expect(state.systemStatus).toEqual(status);
    });

    it('should get system status', () => {
      const status: SystemStatus = {
        status: 'healthy',
        uptime: 86400,
        version: '1.0.0',
        components: {
          websocket_server: 'running',
          camera_monitor: 'running',
          mediamtx_controller: 'running'
        }
      };

      store.setSystemStatus(status);
      
      expect(store.getSystemStatus()).toEqual(status);
    });

    it('should check if all components are running', () => {
      const healthyStatus: SystemStatus = {
        status: 'healthy',
        uptime: 86400,
        version: '1.0.0',
        components: {
          websocket_server: 'running',
          camera_monitor: 'running',
          mediamtx_controller: 'running'
        }
      };

      store.setSystemStatus(healthyStatus);
      expect(store.areAllComponentsRunning()).toBe(true);

      const degradedStatus: SystemStatus = {
        status: 'degraded',
        uptime: 86400,
        version: '1.0.0',
        components: {
          websocket_server: 'running',
          camera_monitor: 'stopped',
          mediamtx_controller: 'running'
        }
      };

      store.setSystemStatus(degradedStatus);
      expect(store.areAllComponentsRunning()).toBe(false);
    });

    it('should get system uptime in human readable format', () => {
      const status: SystemStatus = {
        status: 'healthy',
        uptime: 90061, // 1 day, 1 hour, 1 minute, 1 second
        version: '1.0.0',
        components: {
          websocket_server: 'running',
          camera_monitor: 'running',
          mediamtx_controller: 'running'
        }
      };

      store.setSystemStatus(status);
      
      const uptime = store.getFormattedUptime();
      expect(uptime).toContain('1 day');
      expect(uptime).toContain('1 hour');
      expect(uptime).toContain('1 minute');
    });
  });

  describe('Server Information Management', () => {
    it('should set server info', () => {
      const serverInfo: ServerInfo = {
        name: 'MediaMTX Camera Service',
        version: '1.0.0',
        capabilities: ['camera_control', 'recording', 'streaming'],
        supported_formats: ['mp4', 'h264', 'h265'],
        max_cameras: 10
      };

      store.setServerInfo(serverInfo);
      
      const state = useAdminStore.getState();
      expect(state.serverInfo).toEqual(serverInfo);
    });

    it('should get server info', () => {
      const serverInfo: ServerInfo = {
        name: 'MediaMTX Camera Service',
        version: '1.0.0',
        capabilities: ['camera_control', 'recording', 'streaming'],
        supported_formats: ['mp4', 'h264', 'h265'],
        max_cameras: 10
      };

      store.setServerInfo(serverInfo);
      
      expect(store.getServerInfo()).toEqual(serverInfo);
    });

    it('should check if server supports capability', () => {
      const serverInfo: ServerInfo = {
        name: 'MediaMTX Camera Service',
        version: '1.0.0',
        capabilities: ['camera_control', 'recording', 'streaming'],
        supported_formats: ['mp4', 'h264', 'h265'],
        max_cameras: 10
      };

      store.setServerInfo(serverInfo);
      
      expect(store.supportsCapability('camera_control')).toBe(true);
      expect(store.supportsCapability('admin_access')).toBe(false);
    });

    it('should check if server supports format', () => {
      const serverInfo: ServerInfo = {
        name: 'MediaMTX Camera Service',
        version: '1.0.0',
        capabilities: ['camera_control', 'recording', 'streaming'],
        supported_formats: ['mp4', 'h264', 'h265'],
        max_cameras: 10
      };

      store.setServerInfo(serverInfo);
      
      expect(store.supportsFormat('mp4')).toBe(true);
      expect(store.supportsFormat('avi')).toBe(false);
    });
  });

  describe('Storage Information Management', () => {
    it('should set storage info', () => {
      const storageInfo: StorageInfo = {
        total_space: 1000000000,
        available_space: 500000000,
        used_space: 500000000,
        mount_point: '/storage',
        filesystem: 'ext4'
      };

      store.setStorageInfo(storageInfo);
      
      const state = useAdminStore.getState();
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

    it('should get storage usage percentage', () => {
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
  });

  describe('Loading State Management', () => {
    it('should set loading state', () => {
      store.setLoading(true);
      let state = useAdminStore.getState();
      expect(state.isLoading).toBe(true);

      store.setLoading(false);
      state = useAdminStore.getState();
      expect(state.isLoading).toBe(false);
    });

    it('should set refreshing state', () => {
      store.setRefreshing(true);
      let state = useAdminStore.getState();
      expect(state.isRefreshing).toBe(true);

      store.setRefreshing(false);
      state = useAdminStore.getState();
      expect(state.isRefreshing).toBe(false);
    });

    it('should set restarting state', () => {
      store.setRestarting(true);
      let state = useAdminStore.getState();
      expect(state.isRestarting).toBe(true);

      store.setRestarting(false);
      state = useAdminStore.getState();
      expect(state.isRestarting).toBe(false);
    });

    it('should set updating config state', () => {
      store.setUpdatingConfig(true);
      let state = useAdminStore.getState();
      expect(state.isUpdatingConfig).toBe(true);

      store.setUpdatingConfig(false);
      state = useAdminStore.getState();
      expect(state.isUpdatingConfig).toBe(false);
    });

    it('should set exporting logs state', () => {
      store.setExportingLogs(true);
      let state = useAdminStore.getState();
      expect(state.isExportingLogs).toBe(true);

      store.setExportingLogs(false);
      state = useAdminStore.getState();
      expect(state.isExportingLogs).toBe(false);
    });

    it('should set clearing logs state', () => {
      store.setClearingLogs(true);
      let state = useAdminStore.getState();
      expect(state.isClearingLogs).toBe(true);

      store.setClearingLogs(false);
      state = useAdminStore.getState();
      expect(state.isClearingLogs).toBe(false);
    });
  });

  describe('Error Management', () => {
    it('should set error', () => {
      store.setError('Admin operation failed');
      let state = useAdminStore.getState();
      expect(state.error).toBe('Admin operation failed');

      store.setError(null);
      state = useAdminStore.getState();
      expect(state.error).toBeNull();
    });

    it('should set last error', () => {
      store.setLastError('Last admin error');
      let state = useAdminStore.getState();
      expect(state.lastError).toBe('Last admin error');

      store.setLastError(null);
      state = useAdminStore.getState();
      expect(state.lastError).toBeNull();
    });

    it('should clear all errors', () => {
      store.setError('Current error');
      store.setLastError('Last error');
      
      store.clearAllErrors();
      
      const state = useAdminStore.getState();
      expect(state.error).toBeNull();
      expect(state.lastError).toBeNull();
    });
  });

  describe('Admin Operations', () => {
    it('should refresh system data successfully', async () => {
      const mockMetrics: SystemMetrics = {
        active_connections: 5,
        total_requests: 1000,
        average_response_time: 50,
        error_rate: 0.02,
        memory_usage: 75.5,
        cpu_usage: 45.2
      };

      const mockStatus: SystemStatus = {
        status: 'healthy',
        uptime: 86400,
        version: '1.0.0',
        components: {
          websocket_server: 'running',
          camera_monitor: 'running',
          mediamtx_controller: 'running'
        }
      };

      mockAdminService.getSystemMetrics.mockResolvedValue(mockMetrics);
      mockAdminService.getSystemStatus.mockResolvedValue(mockStatus);

      await store.refreshSystemData();

      const state = useAdminStore.getState();
      expect(state.systemMetrics).toEqual(mockMetrics);
      expect(state.systemStatus).toEqual(mockStatus);
      expect(state.isRefreshing).toBe(false);
      expect(state.error).toBeNull();
      expect(state.lastRefresh).toBeInstanceOf(Date);
    });

    it('should handle refresh system data failure', async () => {
      const error = new Error('System data unavailable');
      mockAdminService.getSystemMetrics.mockRejectedValue(error);

      await store.refreshSystemData();

      const state = useAdminStore.getState();
      expect(state.isRefreshing).toBe(false);
      expect(state.error).toBe('System data unavailable');
    });

    it('should restart service successfully', async () => {
      mockAdminService.restartService.mockResolvedValue(undefined);

      await store.restartService();

      const state = useAdminStore.getState();
      expect(state.isRestarting).toBe(false);
      expect(state.error).toBeNull();
    });

    it('should handle restart service failure', async () => {
      const error = new Error('Service restart failed');
      mockAdminService.restartService.mockRejectedValue(error);

      await store.restartService();

      const state = useAdminStore.getState();
      expect(state.isRestarting).toBe(false);
      expect(state.error).toBe('Service restart failed');
    });

    it('should update configuration successfully', async () => {
      const config = { max_cameras: 15, timeout: 30 };
      mockAdminService.updateConfiguration.mockResolvedValue(undefined);

      await store.updateConfiguration(config);

      const state = useAdminStore.getState();
      expect(state.isUpdatingConfig).toBe(false);
      expect(state.error).toBeNull();
    });

    it('should export logs successfully', async () => {
      const mockLogData = 'log content here';
      mockAdminService.exportLogs.mockResolvedValue(mockLogData);

      const result = await store.exportLogs();

      expect(result).toBe(mockLogData);
      const state = useAdminStore.getState();
      expect(state.isExportingLogs).toBe(false);
      expect(state.error).toBeNull();
    });

    it('should clear logs successfully', async () => {
      mockAdminService.clearLogs.mockResolvedValue(undefined);

      await store.clearLogs();

      const state = useAdminStore.getState();
      expect(state.isClearingLogs).toBe(false);
      expect(state.error).toBeNull();
    });
  });

  describe('System Health Analysis', () => {
    it('should get overall system health', () => {
      const metrics: SystemMetrics = {
        active_connections: 5,
        total_requests: 1000,
        average_response_time: 50,
        error_rate: 0.02,
        memory_usage: 75.5,
        cpu_usage: 45.2
      };

      const status: SystemStatus = {
        status: 'healthy',
        uptime: 86400,
        version: '1.0.0',
        components: {
          websocket_server: 'running',
          camera_monitor: 'running',
          mediamtx_controller: 'running'
        }
      };

      store.setSystemMetrics(metrics);
      store.setSystemStatus(status);

      const health = store.getOverallSystemHealth();
      expect(health).toEqual({
        status: 'healthy',
        components_running: true,
        performance_healthy: true,
        uptime: 86400,
        version: '1.0.0'
      });
    });

    it('should get system summary', () => {
      const metrics: SystemMetrics = {
        active_connections: 5,
        total_requests: 1000,
        average_response_time: 50,
        error_rate: 0.02,
        memory_usage: 75.5,
        cpu_usage: 45.2
      };

      const status: SystemStatus = {
        status: 'healthy',
        uptime: 86400,
        version: '1.0.0',
        components: {
          websocket_server: 'running',
          camera_monitor: 'running',
          mediamtx_controller: 'running'
        }
      };

      const serverInfo: ServerInfo = {
        name: 'MediaMTX Camera Service',
        version: '1.0.0',
        capabilities: ['camera_control', 'recording'],
        supported_formats: ['mp4', 'h264'],
        max_cameras: 10
      };

      store.setSystemMetrics(metrics);
      store.setSystemStatus(status);
      store.setServerInfo(serverInfo);

      const summary = store.getSystemSummary();
      expect(summary).toHaveProperty('metrics');
      expect(summary).toHaveProperty('status');
      expect(summary).toHaveProperty('server_info');
      expect(summary).toHaveProperty('health');
    });
  });

  describe('State Reset', () => {
    it('should reset all state to initial values', () => {
      // Set some state
      store.setSystemMetrics({
        active_connections: 5,
        total_requests: 1000,
        average_response_time: 50,
        error_rate: 0.02,
        memory_usage: 75.5,
        cpu_usage: 45.2
      });
      store.setError('Test error');
      store.setLoading(true);
      
      // Reset
      store.reset();
      
      const state = useAdminStore.getState();
      expect(state.systemMetrics).toBeNull();
      expect(state.systemStatus).toBeNull();
      expect(state.serverInfo).toBeNull();
      expect(state.storageInfo).toBeNull();
      expect(state.isLoading).toBe(false);
      expect(state.isRefreshing).toBe(false);
      expect(state.isRestarting).toBe(false);
      expect(state.isUpdatingConfig).toBe(false);
      expect(state.isExportingLogs).toBe(false);
      expect(state.isClearingLogs).toBe(false);
      expect(state.error).toBeNull();
      expect(state.lastError).toBeNull();
      expect(state.lastRefresh).toBeNull();
    });
  });
});
