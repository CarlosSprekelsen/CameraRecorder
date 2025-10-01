/**
 * Component Test Helper - Lean utility for React component testing
 * 
 * REUSES existing utilities to minimize code bloat:
 * - MockDataFactory for store mocks
 * - AuthHelper for authentication setup
 * - APIResponseValidator for assertions
 * 
 * Ground Truth References:
 * - Testing Guidelines: ../docs/development/client-testing-guidelines.md
 * - Testing Implementation Plan: ../docs/development/testing-implementation-plan.md
 * 
 * Requirements Coverage:
 * - REQ-COMP-001: Component rendering with providers
 * - REQ-COMP-002: Mock prop generation
 * - REQ-COMP-003: Behavior validation
 * 
 * Test Categories: Unit/Component
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */

import { render, RenderOptions, RenderResult } from '@testing-library/react';
import { ReactElement } from 'react';
import { MockDataFactory } from './mocks';
import { AuthHelper } from './auth-helper';

// Service mocking will be handled by individual test files

export interface ComponentTestConfig {
  withStores?: boolean;
  withAuth?: boolean;
  withPermissions?: string[];
  initialStoreState?: Record<string, any>;
}

export interface MockPropsConfig {
  componentName: string;
  overrides?: Record<string, any>;
}

/**
 * Render component with all necessary providers and mocks
 * REUSES: MockDataFactory for store mocks, AuthHelper for auth setup
 */
export const renderWithProviders = (
  ui: ReactElement,
  config: ComponentTestConfig = {}
): RenderResult => {
  const {
    withStores = true,
    withAuth = true,
    withPermissions = ['admin'],
    initialStoreState = {}
  } = config;

  // REUSE: MockDataFactory for store mocks
  const mockStores = withStores ? {
    authStore: MockDataFactory.createMockAuthStore(initialStoreState.authStore),
    deviceStore: MockDataFactory.createMockDeviceStore(initialStoreState.deviceStore),
    recordingStore: MockDataFactory.createMockRecordingStore(initialStoreState.recordingStore),
    fileStore: MockDataFactory.createMockFileStore(initialStoreState.fileStore)
  } : {};

  // REUSE: AuthHelper for authentication setup
  const authSetup = withAuth ? {
    token: AuthHelper.generateTestToken('admin'),
    permissions: withPermissions
  } : null;

  // Render with mocked providers
  return render(ui, {
    wrapper: ({ children }) => {
      // Initialize stores with mocked services if needed
      if (withStores) {
        // CRITICAL: Mock authentication state first
        if (withAuth) {
          const { useAuthStore } = require('../../src/stores/auth/authStore');
          const authStore = useAuthStore.getState();
          
          // Determine role from withPermissions parameter
          const role = withPermissions.includes('admin') ? 'admin' : 
                      withPermissions.includes('operator') ? 'operator' : 'viewer';
          
          // Set authenticated state for tests
          authStore.setAuthenticated(true);
          authStore.setRole(role);
          authStore.setPermissions(withPermissions);
          authStore.setToken('test-token');
          authStore.setSessionId('test-session');
          authStore.setExpiresAt(new Date(Date.now() + 3600000).toISOString()); // 1 hour from now
        }
        
        // Initialize device store - always inject service for component tests
        const { useDeviceStore } = require('../../src/stores/device/deviceStore');
        const deviceStore = useDeviceStore.getState();
        
        // Inject mock service
        const mockDeviceService = MockDataFactory.createMockDeviceService();
        deviceStore.setDeviceService(mockDeviceService);
        
        // Set initial state if provided
        if (initialStoreState?.deviceStore) {
          if (initialStoreState.deviceStore.loading !== undefined) {
            deviceStore.setLoading(initialStoreState.deviceStore.loading);
          }
          if (initialStoreState.deviceStore.error !== undefined) {
            deviceStore.setError(initialStoreState.deviceStore.error);
          }
          if (initialStoreState.deviceStore.cameras) {
            deviceStore.cameras = initialStoreState.deviceStore.cameras;
          }
        }
        
        // Initialize recording store - always inject service for component tests
        const { useRecordingStore } = require('../../src/stores/recording/recordingStore');
        const recordingStore = useRecordingStore.getState();
        
        // Inject mock service
        const mockRecordingService = MockDataFactory.createMockRecordingService();
        recordingStore.setRecordingService(mockRecordingService);
        
        // Set initial state if provided
        if (initialStoreState?.recordingStore) {
          if (initialStoreState.recordingStore.activeRecordings) {
            recordingStore.activeRecordings = initialStoreState.recordingStore.activeRecordings;
          }
          if (initialStoreState.recordingStore.loading !== undefined) {
            recordingStore.loading = initialStoreState.recordingStore.loading;
          }
          if (initialStoreState.recordingStore.error !== undefined) {
            recordingStore.error = initialStoreState.recordingStore.error;
          }
        }
        
        // Initialize file store - always inject service for component tests
        const { useFileStore } = require('../../src/stores/file/fileStore');
        const fileStore = useFileStore.getState();
        
        // Inject mock service
        const mockFileService = MockDataFactory.createMockFileService();
        fileStore.setFileService(mockFileService);
        
        // Set initial state if provided
        if (initialStoreState?.fileStore) {
          if (initialStoreState.fileStore.recordings) {
            fileStore.recordings = initialStoreState.fileStore.recordings;
          }
          if (initialStoreState.fileStore.snapshots) {
            fileStore.snapshots = initialStoreState.fileStore.snapshots;
          }
          if (initialStoreState.fileStore.loading !== undefined) {
            fileStore.loading = initialStoreState.fileStore.loading;
          }
          if (initialStoreState.fileStore.error !== undefined) {
            fileStore.error = initialStoreState.fileStore.error;
          }
          if (initialStoreState.fileStore.pagination) {
            fileStore.pagination = initialStoreState.fileStore.pagination;
          }
          if (initialStoreState.fileStore.currentTab) {
            fileStore.currentTab = initialStoreState.fileStore.currentTab;
          }
        }
        
        // Initialize server store - always inject service for component tests
        const { useServerStore } = require('../../src/stores/server/serverStore');
        const serverStore = useServerStore.getState();
        
        // Inject mock service
        const mockServerService = MockDataFactory.createMockServerService();
        serverStore.setServerService(mockServerService);
        
        // Set initial state if provided
        if (initialStoreState?.serverStore) {
          if (initialStoreState.serverStore.serverInfo) {
            serverStore.serverInfo = initialStoreState.serverStore.serverInfo;
          }
          if (initialStoreState.serverStore.loading !== undefined) {
            serverStore.loading = initialStoreState.serverStore.loading;
          }
          if (initialStoreState.serverStore.error !== undefined) {
            serverStore.error = initialStoreState.serverStore.error;
          }
        }
      }
      
      return children as ReactElement;
    }
  } as RenderOptions);
};

/**
 * Create mock props for any component
 * REUSES: Existing mock patterns from MockDataFactory
 */
export const createMockProps = <T = Record<string, any>>(
  componentName: string,
  overrides: Partial<T> = {}
): T => {
  // REUSE: MockDataFactory patterns for common prop types
  const baseProps = {
    // Common props for all components
    className: 'test-class',
    'data-testid': `test-${componentName.toLowerCase()}`,
    ...overrides
  } as T;

  return baseProps;
};

/**
 * Assert component behavior matches expectations
 * REUSES: APIResponseValidator patterns for validation
 */
export const assertComponentBehavior = (
  component: RenderResult,
  expectedBehavior: {
    isVisible?: boolean;
    hasText?: string[];
    hasRole?: string;
    isDisabled?: boolean;
    hasClass?: string[];
  }
): void => {
  const { isVisible = true, hasText = [], hasRole, isDisabled = false, hasClass = [] } = expectedBehavior;

  if (isVisible) {
    expect(component.container.firstChild).toBeInTheDocument();
  }

  hasText.forEach(text => {
    expect(component.getByText(text)).toBeInTheDocument();
  });

  if (hasRole) {
    expect(component.getByRole(hasRole)).toBeInTheDocument();
  }

  if (isDisabled) {
    expect(component.getByRole('button')).toBeDisabled();
  }

  hasClass.forEach(className => {
    expect(component.container.firstChild).toHaveClass(className);
  });
};
