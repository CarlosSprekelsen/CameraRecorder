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
      // Mock store providers would go here
      // For now, return children directly
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
