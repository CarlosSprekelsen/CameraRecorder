/**
 * ServerStore unit tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * 
 * Requirements Coverage:
 * - REQ-SS-001: Server info management
 * - REQ-SS-002: System status tracking
 * - REQ-SS-003: Storage information handling
 * - REQ-SS-004: Loading and error state management
 * - REQ-SS-005: Last updated timestamp tracking
 * 
 * Test Categories: Unit
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */

import { useServerStore } from '../../../src/stores/server/serverStore';
import { ServerInfo, SystemStatus, StorageInfo } from '../../../src/types/api';
import { APIMocks } from '../../utils/mocks';
import { APIResponseValidator } from '../../utils/validators';

describe('ServerStore Unit Tests', () => {
  let store: ReturnType<typeof useServerStore>;

  beforeEach(() => {
    // Reset store state before each test
    store = useServerStore.getState();
    store.reset();
  });

  afterEach(() => {
    // Reset store after each test
    store.reset();
  });

  describe('REQ-SS-001: Server info management', () => {
    test('should initialize with correct initial state', () => {
      const state = useServerStore.getState();
      
      expect(state.info).toBeNull();
      expect(state.status).toBeNull();
      expect(state.storage).toBeNull();
      expect(state.loading).toBe(false);
      expect(state.error).toBeNull();
      expect(state.lastUpdated).toBeNull();
    });

    test('should set server info correctly', () => {
      const serverInfo: ServerInfo = APIMocks.getServerInfo();
      store.setInfo(serverInfo);
      
      expect(store.info).toEqual(serverInfo);
      expect(store.info?.name).toBe('MediaMTX Camera Service');
      expect(store.info?.version).toBe('1.0.0');
      expect(store.info?.capabilities).toContain('recording');
    });

    test('should clear server info correctly', () => {
      store.setInfo(APIMocks.getServerInfo());
      store.setInfo(null);
      
      expect(store.info).toBeNull();
    });

    test('should handle server info with all fields', () => {
      const serverInfo: ServerInfo = {
        name: 'Test Server',
        version: '2.0.0',
        build_date: '2023-12-01T10:00:00Z',
        go_version: '1.21.5',
        architecture: 'linux/arm64',
        capabilities: ['recording', 'streaming', 'snapshots', 'file_management', 'admin'],
        supported_formats: ['fmp4', 'mp4', 'mkv', 'avi'],
        max_cameras: 20
      };
      
      store.setInfo(serverInfo);
      
      expect(store.info).toEqual(serverInfo);
      expect(store.info?.capabilities).toHaveLength(5);
      expect(store.info?.supported_formats).toHaveLength(4);
      expect(store.info?.max_cameras).toBe(20);
    });

    test('should handle minimal server info', () => {
      const minimalInfo: ServerInfo = {
        name: 'Minimal Server',
        version: '1.0.0',
        build_date: '2023-01-01T00:00:00Z',
        go_version: '1.20.0',
        architecture: 'linux/amd64',
        capabilities: [],
        supported_formats: [],
        max_cameras: 1
      };
      
      store.setInfo(minimalInfo);
      
      expect(store.info).toEqual(minimalInfo);
      expect(store.info?.capabilities).toEqual([]);
      expect(store.info?.supported_formats).toEqual([]);
    });

    test('should update server info correctly', () => {
      const initialInfo = APIMocks.getServerInfo();
      store.setInfo(initialInfo);
      
      const updatedInfo: ServerInfo = {
        ...initialInfo,
        version: '1.1.0',
        capabilities: [...initialInfo.capabilities, 'new_feature']
      };
      
      store.setInfo(updatedInfo);
      
      expect(store.info).toEqual(updatedInfo);
      expect(store.info?.version).toBe('1.1.0');
      expect(store.info?.capabilities).toContain('new_feature');
    });
  });

  describe('REQ-SS-002: System status tracking', () => {
    test('should set system status correctly', () => {
      const systemStatus: SystemStatus = APIMocks.getStatusResult();
      store.setStatus(systemStatus);
      
      expect(store.status).toEqual(systemStatus);
      expect(store.status?.status).toBe('HEALTHY');
      expect(store.status?.version).toBe('1.0.0');
    });

    test('should clear system status correctly', () => {
      store.setStatus(APIMocks.getStatusResult());
      store.setStatus(null);
      
      expect(store.status).toBeNull();
    });

    test('should handle different system statuses', () => {
      const statuses: SystemStatus[] = [
        { status: 'HEALTHY', uptime: 3600, version: '1.0.0' },
        { status: 'DEGRADED', uptime: 7200, version: '1.0.0', components: { websocket_server: 'HEALTHY', camera_monitor: 'DEGRADED' } },
        { status: 'UNHEALTHY', uptime: 1800, version: '1.0.0', components: { websocket_server: 'UNHEALTHY', camera_monitor: 'UNHEALTHY' } }
      ];
      
      statuses.forEach(status => {
        store.setStatus(status);
        expect(store.status).toEqual(status);
      });
    });

    test('should handle system status with components', () => {
      const statusWithComponents: SystemStatus = {
        status: 'HEALTHY',
        uptime: 3600.5,
        version: '1.0.0',
        components: {
          websocket_server: 'HEALTHY',
          camera_monitor: 'HEALTHY',
          mediamtx: 'HEALTHY'
        }
      };
      
      store.setStatus(statusWithComponents);
      
      expect(store.status?.components).toEqual(statusWithComponents.components);
      expect(store.status?.components?.websocket_server).toBe('HEALTHY');
      expect(store.status?.components?.camera_monitor).toBe('HEALTHY');
      expect(store.status?.components?.mediamtx).toBe('HEALTHY');
    });

    test('should handle system status without components', () => {
      const statusWithoutComponents: SystemStatus = {
        status: 'HEALTHY',
        uptime: 3600,
        version: '1.0.0'
      };
      
      store.setStatus(statusWithoutComponents);
      
      expect(store.status?.status).toBe('HEALTHY');
      expect(store.status?.components).toBeUndefined();
    });
  });

  describe('REQ-SS-003: Storage information handling', () => {
    test('should set storage info correctly', () => {
      const storageInfo: StorageInfo = APIMocks.getStorageInfo();
      store.setStorage(storageInfo);
      
      expect(store.storage).toEqual(storageInfo);
      expect(store.storage?.total_space).toBe(1000000000000);
      expect(store.storage?.usage_percentage).toBe(25.0);
      expect(store.storage?.low_space_warning).toBe(false);
    });

    test('should clear storage info correctly', () => {
      store.setStorage(APIMocks.getStorageInfo());
      store.setStorage(null);
      
      expect(store.storage).toBeNull();
    });

    test('should handle storage info with low space warning', () => {
      const storageWithWarning: StorageInfo = {
        total_space: 1000000000000,
        used_space: 900000000000,
        available_space: 100000000000,
        usage_percentage: 90.0,
        recordings_size: 800000000000,
        snapshots_size: 100000000000,
        low_space_warning: true
      };
      
      store.setStorage(storageWithWarning);
      
      expect(store.storage?.low_space_warning).toBe(true);
      expect(store.storage?.usage_percentage).toBe(90.0);
    });

    test('should handle storage info with different sizes', () => {
      const storageInfo: StorageInfo = {
        total_space: 500000000000, // 500GB
        used_space: 250000000000,  // 250GB
        available_space: 250000000000, // 250GB
        usage_percentage: 50.0,
        recordings_size: 200000000000, // 200GB
        snapshots_size: 50000000000,   // 50GB
        low_space_warning: false
      };
      
      store.setStorage(storageInfo);
      
      expect(store.storage?.total_space).toBe(500000000000);
      expect(store.storage?.used_space).toBe(250000000000);
      expect(store.storage?.available_space).toBe(250000000000);
      expect(store.storage?.usage_percentage).toBe(50.0);
    });

    test('should handle zero storage values', () => {
      const emptyStorage: StorageInfo = {
        total_space: 0,
        used_space: 0,
        available_space: 0,
        usage_percentage: 0.0,
        recordings_size: 0,
        snapshots_size: 0,
        low_space_warning: false
      };
      
      store.setStorage(emptyStorage);
      
      expect(store.storage).toEqual(emptyStorage);
      expect(store.storage?.usage_percentage).toBe(0.0);
    });
  });

  describe('REQ-SS-004: Loading and error state management', () => {
    test('should set loading state correctly', () => {
      store.setLoading(true);
      expect(store.loading).toBe(true);
      
      store.setLoading(false);
      expect(store.loading).toBe(false);
    });

    test('should set error state correctly', () => {
      const errorMessage = 'Server connection failed';
      store.setError(errorMessage);
      
      expect(store.error).toBe(errorMessage);
    });

    test('should clear error state correctly', () => {
      store.setError('Test error');
      store.setError(null);
      
      expect(store.error).toBeNull();
    });

    test('should handle different error types', () => {
      const errors = [
        'Network timeout',
        'Authentication failed',
        'Server unavailable',
        'Invalid response format',
        'Permission denied'
      ];
      
      errors.forEach(error => {
        store.setError(error);
        expect(store.error).toBe(error);
      });
    });

    test('should handle loading and error states together', () => {
      // Set loading
      store.setLoading(true);
      expect(store.loading).toBe(true);
      expect(store.error).toBeNull();
      
      // Set error while loading
      store.setError('Connection failed');
      expect(store.loading).toBe(true);
      expect(store.error).toBe('Connection failed');
      
      // Clear loading
      store.setLoading(false);
      expect(store.loading).toBe(false);
      expect(store.error).toBe('Connection failed');
      
      // Clear error
      store.setError(null);
      expect(store.loading).toBe(false);
      expect(store.error).toBeNull();
    });

    test('should reset loading and error states', () => {
      store.setLoading(true);
      store.setError('Test error');
      
      store.reset();
      
      expect(store.loading).toBe(false);
      expect(store.error).toBeNull();
    });
  });

  describe('REQ-SS-005: Last updated timestamp tracking', () => {
    test('should set last updated timestamp correctly', () => {
      const timestamp = new Date().toISOString();
      store.setLastUpdated(timestamp);
      
      expect(store.lastUpdated).toBe(timestamp);
    });

    test('should clear last updated timestamp correctly', () => {
      store.setLastUpdated(new Date().toISOString());
      store.setLastUpdated(null);
      
      expect(store.lastUpdated).toBeNull();
    });

    test('should handle timestamp updates with server info', () => {
      const timestamp = new Date().toISOString();
      const serverInfo = APIMocks.getServerInfo();
      
      store.setInfo(serverInfo);
      store.setLastUpdated(timestamp);
      
      expect(store.info).toEqual(serverInfo);
      expect(store.lastUpdated).toBe(timestamp);
    });

    test('should handle timestamp updates with status', () => {
      const timestamp = new Date().toISOString();
      const status = APIMocks.getStatusResult();
      
      store.setStatus(status);
      store.setLastUpdated(timestamp);
      
      expect(store.status).toEqual(status);
      expect(store.lastUpdated).toBe(timestamp);
    });

    test('should handle timestamp updates with storage', () => {
      const timestamp = new Date().toISOString();
      const storage = APIMocks.getStorageInfo();
      
      store.setStorage(storage);
      store.setLastUpdated(timestamp);
      
      expect(store.storage).toEqual(storage);
      expect(store.lastUpdated).toBe(timestamp);
    });

    test('should handle invalid timestamp formats', () => {
      const invalidTimestamps = [
        'invalid-date',
        '',
        '2023-13-45T25:70:90Z',
        null,
        undefined
      ];
      
      invalidTimestamps.forEach(timestamp => {
        store.setLastUpdated(timestamp as any);
        expect(store.lastUpdated).toBe(timestamp);
      });
    });
  });

  describe('Integration Tests', () => {
    test('should handle complete server data update', () => {
      const timestamp = new Date().toISOString();
      const serverInfo = APIMocks.getServerInfo();
      const status = APIMocks.getStatusResult();
      const storage = APIMocks.getStorageInfo();
      
      // Set all server data
      store.setInfo(serverInfo);
      store.setStatus(status);
      store.setStorage(storage);
      store.setLastUpdated(timestamp);
      
      expect(store.info).toEqual(serverInfo);
      expect(store.status).toEqual(status);
      expect(store.storage).toEqual(storage);
      expect(store.lastUpdated).toBe(timestamp);
      expect(store.loading).toBe(false);
      expect(store.error).toBeNull();
    });

    test('should handle server data update with loading state', () => {
      const timestamp = new Date().toISOString();
      const serverInfo = APIMocks.getServerInfo();
      
      // Start loading
      store.setLoading(true);
      expect(store.loading).toBe(true);
      
      // Update data
      store.setInfo(serverInfo);
      store.setLastUpdated(timestamp);
      
      // Still loading
      expect(store.loading).toBe(true);
      expect(store.info).toEqual(serverInfo);
      expect(store.lastUpdated).toBe(timestamp);
      
      // Finish loading
      store.setLoading(false);
      expect(store.loading).toBe(false);
    });

    test('should handle server data update with error', () => {
      const serverInfo = APIMocks.getServerInfo();
      const errorMessage = 'Failed to fetch server data';
      
      // Set data
      store.setInfo(serverInfo);
      
      // Error occurs
      store.setError(errorMessage);
      
      expect(store.info).toEqual(serverInfo); // Data should remain
      expect(store.error).toBe(errorMessage);
    });

    test('should handle complete reset', () => {
      // Set all data
      store.setInfo(APIMocks.getServerInfo());
      store.setStatus(APIMocks.getStatusResult());
      store.setStorage(APIMocks.getStorageInfo());
      store.setLoading(true);
      store.setError('Test error');
      store.setLastUpdated(new Date().toISOString());
      
      // Reset
      store.reset();
      
      expect(store.info).toBeNull();
      expect(store.status).toBeNull();
      expect(store.storage).toBeNull();
      expect(store.loading).toBe(false);
      expect(store.error).toBeNull();
      expect(store.lastUpdated).toBeNull();
    });
  });

  describe('API Compliance Tests', () => {
    test('should handle server info that matches API schema', () => {
      const serverInfo = APIMocks.getServerInfo();
      store.setInfo(serverInfo);
      
      expect(APIResponseValidator.validateServerInfo(serverInfo)).toBe(true);
      expect(store.info).toEqual(serverInfo);
    });

    test('should handle status result that matches API schema', () => {
      const status = APIMocks.getStatusResult();
      store.setStatus(status);
      
      expect(APIResponseValidator.validateStatusResult(status)).toBe(true);
      expect(store.status).toEqual(status);
    });

    test('should handle storage info that matches API schema', () => {
      const storage = APIMocks.getStorageInfo();
      store.setStorage(storage);
      
      expect(APIResponseValidator.validateStorageInfo(storage)).toBe(true);
      expect(store.storage).toEqual(storage);
    });

    test('should handle valid system status values', () => {
      const validStatuses = ['HEALTHY', 'DEGRADED', 'UNHEALTHY'];
      
      validStatuses.forEach(status => {
        const systemStatus: SystemStatus = { status: status as any, uptime: 3600, version: '1.0.0' };
        store.setStatus(systemStatus);
        expect(store.status?.status).toBe(status);
      });
    });
  });

  describe('Edge Cases and Complex Scenarios', () => {
    test('should handle rapid data updates', () => {
      const serverInfos = [
        APIMocks.getServerInfo(),
        { ...APIMocks.getServerInfo(), version: '1.1.0' },
        { ...APIMocks.getServerInfo(), version: '1.2.0' }
      ];
      
      serverInfos.forEach(info => {
        store.setInfo(info);
        expect(store.info?.version).toBe(info.version);
      });
      
      expect(store.info?.version).toBe('1.2.0');
    });

    test('should handle concurrent loading and error states', () => {
      // Rapid state changes
      store.setLoading(true);
      store.setError('Error 1');
      store.setLoading(false);
      store.setError('Error 2');
      store.setLoading(true);
      store.setError(null);
      
      expect(store.loading).toBe(true);
      expect(store.error).toBeNull();
    });

    test('should handle large storage values', () => {
      const largeStorage: StorageInfo = {
        total_space: 1000000000000000, // 1PB
        used_space: 500000000000000,   // 500TB
        available_space: 500000000000000, // 500TB
        usage_percentage: 50.0,
        recordings_size: 400000000000000, // 400TB
        snapshots_size: 100000000000000,  // 100TB
        low_space_warning: false
      };
      
      store.setStorage(largeStorage);
      
      expect(store.storage?.total_space).toBe(1000000000000000);
      expect(store.storage?.usage_percentage).toBe(50.0);
    });

    test('should handle server info with many capabilities', () => {
      const serverInfoWithManyCapabilities: ServerInfo = {
        name: 'Feature-Rich Server',
        version: '2.0.0',
        build_date: '2023-12-01T10:00:00Z',
        go_version: '1.21.5',
        architecture: 'linux/amd64',
        capabilities: [
          'recording', 'streaming', 'snapshots', 'file_management',
          'admin', 'monitoring', 'analytics', 'automation',
          'backup', 'restore', 'scheduling', 'notifications'
        ],
        supported_formats: ['fmp4', 'mp4', 'mkv', 'avi', 'mov', 'wmv'],
        max_cameras: 100
      };
      
      store.setInfo(serverInfoWithManyCapabilities);
      
      expect(store.info?.capabilities).toHaveLength(12);
      expect(store.info?.supported_formats).toHaveLength(6);
      expect(store.info?.max_cameras).toBe(100);
    });

    test('should handle status with all components', () => {
      const statusWithAllComponents: SystemStatus = {
        status: 'HEALTHY',
        uptime: 86400.5,
        version: '1.0.0',
        components: {
          websocket_server: 'HEALTHY',
          camera_monitor: 'HEALTHY',
          mediamtx: 'HEALTHY',
          database: 'HEALTHY',
          storage: 'HEALTHY',
          network: 'HEALTHY'
        }
      };
      
      store.setStatus(statusWithAllComponents);
      
      expect(store.status?.components?.websocket_server).toBe('HEALTHY');
      expect(store.status?.components?.camera_monitor).toBe('HEALTHY');
      expect(store.status?.components?.mediamtx).toBe('HEALTHY');
      expect(store.status?.components?.database).toBe('HEALTHY');
      expect(store.status?.components?.storage).toBe('HEALTHY');
      expect(store.status?.components?.network).toBe('HEALTHY');
    });

    test('should handle mixed data types and null values', () => {
      // Set some data
      store.setInfo(APIMocks.getServerInfo());
      store.setStatus(APIMocks.getStatusResult());
      
      // Set some to null
      store.setStorage(null);
      store.setError(null);
      store.setLastUpdated(null);
      
      expect(store.info).toBeTruthy();
      expect(store.status).toBeTruthy();
      expect(store.storage).toBeNull();
      expect(store.error).toBeNull();
      expect(store.lastUpdated).toBeNull();
    });
  });
});
