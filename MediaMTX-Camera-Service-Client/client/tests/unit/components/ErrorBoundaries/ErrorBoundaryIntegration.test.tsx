/**
 * Error Boundary Integration Test Suite
 * 
 * Tests the integration of Error Boundaries in the application including:
 * - App.tsx error boundary hierarchy
 * - Feature error boundary isolation
 * - Service error boundary recovery
 * - Cross-boundary error propagation
 * - Error recovery flow validation
 * - Real-world error scenarios
 */

import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import App from '../../../../src/App';
import FeatureErrorBoundary from '../../../../src/components/ErrorBoundaries/FeatureErrorBoundary';
import ServiceErrorBoundary from '../../../../src/components/ErrorBoundaries/ServiceErrorBoundary';
import ErrorBoundary from '../../../../src/components/common/ErrorBoundary';
import {
  ErrorThrowingComponent,
  AsyncErrorComponent,
  ErrorTypeComponent,
  renderWithTheme,
  userInteractions,
  errorBoundaryHelpers,
  mockWindowReload,
  waitForAsync,
  testData,
  resetMocks,
  mockLogger,
  mockLoggers,
} from './test-utils';

// Mock React Router
jest.mock('react-router-dom', () => ({
  BrowserRouter: ({ children }: { children: React.ReactNode }) => <div data-testid="router">{children}</div>,
  Routes: ({ children }: { children: React.ReactNode }) => <div data-testid="routes">{children}</div>,
  Route: ({ element }: { element: React.ReactNode }) => <div data-testid="route">{element}</div>,
}));

// Mock other components to focus on error boundary integration
jest.mock('../../../../src/components/common/ConnectionManager', () => {
  return function MockConnectionManager({ children }: { children: React.ReactNode }) {
    return <div data-testid="connection-manager">{children}</div>;
  };
});

jest.mock('../../../../src/components/Dashboard/Dashboard', () => {
  return function MockDashboard() {
    return <div data-testid="dashboard">Dashboard</div>;
  };
});

jest.mock('../../../../src/components/CameraDetail/CameraDetail', () => {
  return function MockCameraDetail() {
    return <div data-testid="camera-detail">Camera Detail</div>;
  };
});

jest.mock('../../../../src/components/FileManager/FileManager', () => {
  return function MockFileManager() {
    return <div data-testid="file-manager">File Manager</div>;
  };
});

jest.mock('../../../../src/components/HealthMonitor/HealthMonitor', () => {
  return function MockHealthMonitor() {
    return <div data-testid="health-monitor">Health Monitor</div>;
  };
});

jest.mock('../../../../src/components/AdminDashboard/AdminDashboard', () => {
  return function MockAdminDashboard() {
    return <div data-testid="admin-dashboard">Admin Dashboard</div>;
  };
});

jest.mock('../../../../src/components/Settings/Settings', () => {
  return function MockSettings() {
    return <div data-testid="settings">Settings</div>;
  };
});

// Mock console.error to avoid noise in test output
const originalConsoleError = console.error;
beforeAll(() => {
  console.error = jest.fn();
});

afterAll(() => {
  console.error = originalConsoleError;
});

describe('Error Boundary Integration', () => {
  beforeEach(() => {
    resetMocks();
  });

  describe('App.tsx Error Boundary Hierarchy', () => {
    it('should have proper error boundary hierarchy in App component', () => {
      const { container } = renderWithTheme(<App />);

      // Should render without errors
      expect(screen.getByTestId('router')).toBeInTheDocument();
      expect(screen.getByTestId('connection-manager')).toBeInTheDocument();
    });

    it('should catch errors at the top level ErrorBoundary', () => {
      // Mock a component that throws an error
      const ErrorComponent = () => {
        throw new Error('Top level error');
      };

      const { container } = renderWithTheme(
        <ErrorBoundary>
          <ErrorComponent />
        </ErrorBoundary>
      );

      expect(screen.getByText('Something went wrong')).toBeInTheDocument();
    });
  });

  describe('Feature Error Boundary Isolation', () => {
    it('should isolate feature errors without affecting other features', () => {
      const { container } = renderWithTheme(
        <div>
          <FeatureErrorBoundary featureName="Feature1">
            <ErrorThrowingComponent shouldThrow={true} errorMessage="Feature1 error" />
          </FeatureErrorBoundary>
          <FeatureErrorBoundary featureName="Feature2">
            <ErrorThrowingComponent shouldThrow={false} />
          </FeatureErrorBoundary>
        </div>
      );

      expect(screen.getByText('Feature1 Error')).toBeInTheDocument();
      expect(screen.getByTestId('error-throwing-component')).toBeInTheDocument();
    });

    it('should handle multiple feature errors independently', () => {
      const { container } = renderWithTheme(
        <div>
          <FeatureErrorBoundary featureName="Feature1">
            <ErrorThrowingComponent shouldThrow={true} errorMessage="Feature1 error" />
          </FeatureErrorBoundary>
          <FeatureErrorBoundary featureName="Feature2">
            <ErrorThrowingComponent shouldThrow={true} errorMessage="Feature2 error" />
          </FeatureErrorBoundary>
        </div>
      );

      expect(screen.getByText('Feature1 Error')).toBeInTheDocument();
      expect(screen.getByText('Feature2 Error')).toBeInTheDocument();
    });
  });

  describe('Service Error Boundary Recovery', () => {
    it('should handle service errors with proper recovery flow', async () => {
      const { container, rerender } = renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService">
          <ErrorThrowingComponent shouldThrow={true} />
        </ServiceErrorBoundary>
      );

      expect(screen.getByText('TestService Service Error')).toBeInTheDocument();

      // Retry
      fireEvent.click(screen.getByText('Try Again'));
      await waitForAsync(1100);

      // Re-render with no error (simulating recovery)
      rerender(
        <ServiceErrorBoundary serviceName="TestService">
          <ErrorThrowingComponent shouldThrow={false} />
        </ServiceErrorBoundary>
      );

      expect(screen.queryByText('TestService Service Error')).not.toBeInTheDocument();
      expect(screen.getByTestId('error-throwing-component')).toBeInTheDocument();
    });

    it('should handle service fallback mode activation', () => {
      const { container } = renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService">
          <ErrorThrowingComponent shouldThrow={true} />
        </ServiceErrorBoundary>
      );

      fireEvent.click(screen.getByText('Use Fallback'));

      expect(mockLogger.info).toHaveBeenCalledWith(
        'Using fallback mode for TestService',
        undefined,
        'service-error-boundary'
      );
    });
  });

  describe('Cross-Boundary Error Propagation', () => {
    it('should prevent error propagation between different boundary levels', () => {
      const { container } = renderWithTheme(
        <ErrorBoundary>
          <FeatureErrorBoundary featureName="Feature1">
            <ErrorThrowingComponent shouldThrow={true} errorMessage="Feature error" />
          </FeatureErrorBoundary>
          <FeatureErrorBoundary featureName="Feature2">
            <ErrorThrowingComponent shouldThrow={false} />
          </FeatureErrorBoundary>
        </ErrorBoundary>
      );

      // Feature1 should show error, Feature2 should render normally
      expect(screen.getByText('Feature1 Error')).toBeInTheDocument();
      expect(screen.getByTestId('error-throwing-component')).toBeInTheDocument();
    });

    it('should handle nested error boundaries correctly', () => {
      const { container } = renderWithTheme(
        <ErrorBoundary>
          <ServiceErrorBoundary serviceName="Service1">
            <FeatureErrorBoundary featureName="Feature1">
              <ErrorThrowingComponent shouldThrow={true} errorMessage="Nested error" />
            </FeatureErrorBoundary>
          </ServiceErrorBoundary>
        </ErrorBoundary>
      );

      // Should catch at the FeatureErrorBoundary level
      expect(screen.getByText('Feature1 Error')).toBeInTheDocument();
    });
  });

  describe('Error Recovery Flow Validation', () => {
    it('should validate complete error recovery flow', async () => {
      const { container, rerender } = renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService">
          <FeatureErrorBoundary featureName="TestFeature">
            <ErrorThrowingComponent shouldThrow={true} />
          </FeatureErrorBoundary>
        </ServiceErrorBoundary>
      );

      // Initial error state
      expect(screen.getByText('TestFeature Error')).toBeInTheDocument();

      // Retry at feature level
      fireEvent.click(screen.getByText('Try Again'));
      await waitForAsync(100);

      // Re-render with no error
      rerender(
        <ServiceErrorBoundary serviceName="TestService">
          <FeatureErrorBoundary featureName="TestFeature">
            <ErrorThrowingComponent shouldThrow={false} />
          </FeatureErrorBoundary>
        </ServiceErrorBoundary>
      );

      // Should recover completely
      expect(screen.queryByText('TestFeature Error')).not.toBeInTheDocument();
      expect(screen.getByTestId('error-throwing-component')).toBeInTheDocument();
    });

    it('should handle partial recovery scenarios', async () => {
      const { container, rerender } = renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService">
          <FeatureErrorBoundary featureName="TestFeature">
            <ErrorThrowingComponent shouldThrow={true} />
          </FeatureErrorBoundary>
        </ServiceErrorBoundary>
      );

      // Initial error state
      expect(screen.getByText('TestFeature Error')).toBeInTheDocument();

      // Retry at feature level
      fireEvent.click(screen.getByText('Try Again'));
      await waitForAsync(100);

      // Re-render with different error
      rerender(
        <ServiceErrorBoundary serviceName="TestService">
          <FeatureErrorBoundary featureName="TestFeature">
            <ErrorThrowingComponent shouldThrow={true} errorMessage="Different error" />
          </FeatureErrorBoundary>
        </ServiceErrorBoundary>
      );

      // Should still show error but with updated retry count
      expect(screen.getByText('TestFeature Error')).toBeInTheDocument();
      expect(screen.getByText('Retry attempt 1 of 3')).toBeInTheDocument();
    });
  });

  describe('Real-World Error Scenarios', () => {
    it('should handle WebSocket connection errors', async () => {
      const { container } = renderWithTheme(
        <ServiceErrorBoundary serviceName="WebSocketService">
          <AsyncErrorComponent shouldThrow={true} delay={100} />
        </ServiceErrorBoundary>
      );

      // Wait for async error
      await waitForAsync(150);

      expect(screen.getByText('WebSocketService Service Error')).toBeInTheDocument();
    });

    it('should handle component rendering errors', () => {
      const { container } = renderWithTheme(
        <FeatureErrorBoundary featureName="Dashboard">
          <ErrorTypeComponent errorType="syntax" />
        </FeatureErrorBoundary>
      );

      expect(screen.getByText('Dashboard Error')).toBeInTheDocument();
    });

    it('should handle service timeout scenarios', async () => {
      const { container } = renderWithTheme(
        <ServiceErrorBoundary serviceName="TimeoutService">
          <AsyncErrorComponent shouldThrow={true} delay={2000} />
        </ServiceErrorBoundary>
      );

      // Wait for timeout error
      await waitFor(() => {
        expect(screen.getByText('TimeoutService Service Error')).toBeInTheDocument();
      }, { timeout: 3000 });
    });

    it('should handle network interruption scenarios', () => {
      const { container } = renderWithTheme(
        <ServiceErrorBoundary serviceName="NetworkService">
          <ErrorThrowingComponent shouldThrow={true} errorMessage="Network error" />
        </ServiceErrorBoundary>
      );

      expect(screen.getByText('NetworkService Service Error')).toBeInTheDocument();
    });
  });

  describe('Error Boundary Performance', () => {
    it('should handle rapid error state changes efficiently', () => {
      const { container, rerender } = renderWithTheme(
        <FeatureErrorBoundary featureName="PerformanceTest">
          <ErrorThrowingComponent shouldThrow={true} />
        </FeatureErrorBoundary>
      );

      // Rapid re-renders
      const startTime = performance.now();
      for (let i = 0; i < 10; i++) {
        rerender(
          <FeatureErrorBoundary featureName="PerformanceTest">
            <ErrorThrowingComponent shouldThrow={i % 2 === 0} />
          </FeatureErrorBoundary>
        );
      }
      const endTime = performance.now();

      // Should complete within reasonable time (less than 100ms)
      expect(endTime - startTime).toBeLessThan(100);
    });

    it('should handle multiple concurrent error boundaries', () => {
      const { container } = renderWithTheme(
        <div>
          {Array.from({ length: 5 }, (_, i) => (
            <FeatureErrorBoundary key={i} featureName={`Feature${i}`}>
              <ErrorThrowingComponent shouldThrow={i % 2 === 0} />
            </FeatureErrorBoundary>
          ))}
        </div>
      );

      // Should handle multiple boundaries without performance issues
      expect(container).toBeInTheDocument();
    });
  });

  describe('Error Boundary Accessibility', () => {
    it('should maintain accessibility across error boundary hierarchy', () => {
      const { container } = renderWithTheme(
        <ErrorBoundary>
          <ServiceErrorBoundary serviceName="TestService">
            <FeatureErrorBoundary featureName="TestFeature">
              <ErrorThrowingComponent shouldThrow={true} />
            </FeatureErrorBoundary>
          </ServiceErrorBoundary>
        </ErrorBoundary>
      );

      // Should have proper accessibility attributes
      const retryButton = screen.getByText('Try Again');
      const reloadButton = screen.getByText('Reload Page');
      
      expect(retryButton).toBeInTheDocument();
      expect(reloadButton).toBeInTheDocument();
    });

    it('should provide clear error messages in nested boundaries', () => {
      const { container } = renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService">
          <FeatureErrorBoundary featureName="TestFeature">
            <ErrorThrowingComponent shouldThrow={true} />
          </FeatureErrorBoundary>
        </ServiceErrorBoundary>
      );

      expect(screen.getByText('TestFeature Error')).toBeInTheDocument();
      expect(screen.getByText(/Something went wrong in the TestFeature feature/)).toBeInTheDocument();
    });
  });
});
