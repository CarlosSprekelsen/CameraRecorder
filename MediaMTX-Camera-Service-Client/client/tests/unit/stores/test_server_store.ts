/**
 * Unit Tests for Server Store
 * 
 * REQ-001: Store State Management - Test Zustand store actions
 * REQ-002: State Transitions - Test state changes
 * REQ-003: Error Handling - Test error states and recovery
 * REQ-004: Server Data Management - Test server info, status, and storage
 * REQ-005: Side Effects - Test store side effects
 * 
 * Ground Truth: Official RPC Documentation
 * API Reference: docs/api/json_rpc_methods.md
 */

import { describe, test, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { MockDataFactory } from '../../utils/mocks';
import { APIResponseValidator } from '../../utils/validators';
import { useServerStore } from '../../../src/stores/server/serverStore';

describe('Server Store', () => {
  beforeEach(() => {
    // Reset the store to initial state
    useServerStore.getState().reset();
  });

  afterEach(() => {
    useServerStore.getState().reset();
  });

  describe('REQ-001: Store State Management', () => {
    test('should initialize with correct initial state', () => {
      const state = useServerStore.getState();
      
      expect(state.info).toBe(null);
      expect(state.status).toBe(null);
      expect(state.storage).toBe(null);
      expect(state.loading).toBe(false);
      expect(state.error).toBe(null);
      expect(state.lastUpdated).toBe(null);
    });

    test('should set info correctly', () => {
      const { setInfo } = useServerStore.getState();
      const mockInfo = MockDataFactory.getServerInfo();
      
      setInfo(mockInfo);
      expect(useServerStore.getState().info).toEqual(mockInfo);
      
      setInfo(null);
      expect(useServerStore.getState().info).toBe(null);
    });

    test('should set status correctly', () => {
      const { setStatus } = useServerStore.getState();
      const mockStatus = MockDataFactory.getSystemStatus();
      
      setStatus(mockStatus);
      expect(useServerStore.getState().status).toEqual(mockStatus);
      
      setStatus(null);
      expect(useServerStore.getState().status).toBe(null);
    });

    test('should set storage correctly', () => {
      const { setStorage } = useServerStore.getState();
      const mockStorage = MockDataFactory.getStorageInfo();
      
      setStorage(mockStorage);
      expect(useServerStore.getState().storage).toEqual(mockStorage);
      
      setStorage(null);
      expect(useServerStore.getState().storage).toBe(null);
    });

    test('should set loading state correctly', () => {
      const { setLoading } = useServerStore.getState();
      
      setLoading(true);
      expect(useServerStore.getState().loading).toBe(true);
      
      setLoading(false);
      expect(useServerStore.getState().loading).toBe(false);
    });

    test('should set error state correctly', () => {
      const { setError } = useServerStore.getState();
      const errorMessage = 'Server error';
      
      setError(errorMessage);
      expect(useServerStore.getState().error).toBe(errorMessage);
      
      setError(null);
      expect(useServerStore.getState().error).toBe(null);
    });

    test('should set last updated timestamp correctly', () => {
      const { setLastUpdated } = useServerStore.getState();
      const timestamp = '2025-01-15T14:30:00Z';
      
      setLastUpdated(timestamp);
      expect(useServerStore.getState().lastUpdated).toBe(timestamp);
      
      setLastUpdated(null);
      expect(useServerStore.getState().lastUpdated).toBe(null);
    });
  });

  describe('REQ-004: Server Data Management', () => {
    test('should manage server info data', () => {
      const { setInfo } = useServerStore.getState();
      const mockInfo = MockDataFactory.getServerInfo();
      
      setInfo(mockInfo);
      
      const state = useServerStore.getState();
      expect(state.info).toEqual(mockInfo);
      expect(state.info?.version).toBe(mockInfo.version);
      expect(state.info?.build_time).toBe(mockInfo.build_time);
    });

    test('should manage system status data', () => {
      const { setStatus } = useServerStore.getState();
      const mockStatus = MockDataFactory.getSystemStatus();
      
      setStatus(mockStatus);
      
      const state = useServerStore.getState();
      expect(state.status).toEqual(mockStatus);
      expect(state.status?.uptime).toBe(mockStatus.uptime);
      expect(state.status?.cpu_usage).toBe(mockStatus.cpu_usage);
    });

    test('should manage storage info data', () => {
      const { setStorage } = useServerStore.getState();
      const mockStorage = MockDataFactory.getStorageInfo();
      
      setStorage(mockStorage);
      
      const state = useServerStore.getState();
      expect(state.storage).toEqual(mockStorage);
      expect(state.storage?.total_space).toBe(mockStorage.total_space);
      expect(state.storage?.used_space).toBe(mockStorage.used_space);
    });
  });

  describe('REQ-002: State Transitions', () => {
    test('should handle loading to success transition', () => {
      const { setLoading, setInfo, setLastUpdated } = useServerStore.getState();
      const mockInfo = MockDataFactory.getServerInfo();
      
      // Start loading
      setLoading(true);
      expect(useServerStore.getState().loading).toBe(true);
      
      // Set data
      setInfo(mockInfo);
      setLastUpdated('2025-01-15T14:30:00Z');
      
      // Stop loading
      setLoading(false);
      
      const state = useServerStore.getState();
      expect(state.loading).toBe(false);
      expect(state.info).toEqual(mockInfo);
      expect(state.lastUpdated).toBe('2025-01-15T14:30:00Z');
    });

    test('should handle loading to error transition', () => {
      const { setLoading, setError } = useServerStore.getState();
      
      // Start loading
      setLoading(true);
      expect(useServerStore.getState().loading).toBe(true);
      
      // Set error
      setError('Server unavailable');
      
      // Stop loading
      setLoading(false);
      
      const state = useServerStore.getState();
      expect(state.loading).toBe(false);
      expect(state.error).toBe('Server unavailable');
    });
  });

  describe('REQ-003: Error Handling', () => {
    test('should handle null values correctly', () => {
      const { setInfo, setStatus, setStorage } = useServerStore.getState();
      
      setInfo(null);
      setStatus(null);
      setStorage(null);
      
      const state = useServerStore.getState();
      expect(state.info).toBe(null);
      expect(state.status).toBe(null);
      expect(state.storage).toBe(null);
    });

    test('should handle error state correctly', () => {
      const { setError } = useServerStore.getState();
      
      setError('Connection failed');
      
      const state = useServerStore.getState();
      expect(state.error).toBe('Connection failed');
    });
  });

  describe('REQ-005: Side Effects', () => {
    test('should reset store to initial state', () => {
      const { reset, setInfo, setStatus, setStorage, setLoading, setError } = useServerStore.getState();
      
      // Modify state
      setInfo(MockDataFactory.getServerInfo());
      setStatus(MockDataFactory.getSystemStatus());
      setStorage(MockDataFactory.getStorageInfo());
      setLoading(true);
      setError('Test error');
      
      // Reset
      reset();
      
      // Check state is back to initial
      const state = useServerStore.getState();
      expect(state.info).toBe(null);
      expect(state.status).toBe(null);
      expect(state.storage).toBe(null);
      expect(state.loading).toBe(false);
      expect(state.error).toBe(null);
      expect(state.lastUpdated).toBe(null);
    });

    test('should maintain state consistency during updates', () => {
      const { setInfo, setStatus, setStorage, setLoading } = useServerStore.getState();
      
      // Set multiple properties
      setInfo(MockDataFactory.getServerInfo());
      setStatus(MockDataFactory.getSystemStatus());
      setStorage(MockDataFactory.getStorageInfo());
      setLoading(true);
      
      const state = useServerStore.getState();
      expect(state.info).toBeTruthy();
      expect(state.status).toBeTruthy();
      expect(state.storage).toBeTruthy();
      expect(state.loading).toBe(true);
    });
  });

  describe('API Compliance Validation', () => {
    test('should validate server info against RPC spec', () => {
      const { setInfo } = useServerStore.getState();
      const mockInfo = MockDataFactory.getServerInfo();
      
      setInfo(mockInfo);
      
      const state = useServerStore.getState();
      expect(APIResponseValidator.validateServerInfo(state.info!)).toBe(true);
    });

    test('should validate system status against RPC spec', () => {
      const { setStatus } = useServerStore.getState();
      const mockStatus = MockDataFactory.getSystemStatus();
      
      setStatus(mockStatus);
      
      const state = useServerStore.getState();
      expect(APIResponseValidator.validateSystemStatus(state.status!)).toBe(true);
    });

    test('should validate storage info against RPC spec', () => {
      const { setStorage } = useServerStore.getState();
      const mockStorage = MockDataFactory.getStorageInfo();
      
      setStorage(mockStorage);
      
      const state = useServerStore.getState();
      expect(APIResponseValidator.validateStorageInfo(state.storage!)).toBe(true);
    });
  });
});