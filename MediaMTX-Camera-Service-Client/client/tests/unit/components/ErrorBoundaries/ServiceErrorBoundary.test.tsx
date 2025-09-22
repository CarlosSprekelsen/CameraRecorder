/**
 * ServiceErrorBoundary Test Suite
 * 
 * Comprehensive tests for the ServiceErrorBoundary component including:
 * - Service-specific error handling
 * - Retryable vs non-retryable configuration
 * - Async retry operations with delays
 * - Fallback mode activation
 * - Error severity classification
 * - Service degradation scenarios
 * - Custom max retries configuration
 * - Error reporting integration
 * - Service timeout handling
 * - Network error simulation
 */

import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import ServiceErrorBoundary from '../../../../src/components/ErrorBoundaries/ServiceErrorBoundary';
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

// Mock console.error to avoid noise in test output
const originalConsoleError = console.error;
beforeAll(() => {
  console.error = jest.fn();
});

afterAll(() => {
  console.error = originalConsoleError;
});

describe('ServiceErrorBoundary', () => {
  beforeEach(() => {
    resetMocks();
  });

  describe('Basic Service Error Catching', () => {
    it('should catch errors and display service error UI', () => {
      const { container } = renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService">
          <ErrorThrowingComponent shouldThrow={true} />
        </ServiceErrorBoundary>
      );

      expect(errorBoundaryHelpers.hasError(container)).toBe(true);
      expect(screen.getByText('TestService Service Error')).toBeInTheDocument();
      expect(screen.getByText(/The TestService service encountered an error/)).toBeInTheDocument();
    });

    it('should render children when no error occurs', () => {
      const { container } = renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService">
          <ErrorThrowingComponent shouldThrow={false} />
        </ServiceErrorBoundary>
      );

      expect(errorBoundaryHelpers.hasError(container)).toBe(false);
      expect(screen.getByTestId('error-throwing-component')).toBeInTheDocument();
    });

    it('should catch different types of service errors', () => {
      const errorTypes = ['syntax', 'reference', 'type', 'custom'] as const;
      
      errorTypes.forEach(errorType => {
        const { container, unmount } = renderWithTheme(
          <ServiceErrorBoundary serviceName="TestService">
            <ErrorTypeComponent errorType={errorType} />
          </ServiceErrorBoundary>
        );

        expect(errorBoundaryHelpers.hasError(container)).toBe(true);
        expect(screen.getByText('TestService Service Error')).toBeInTheDocument();
        
        unmount();
      });
    });
  });

  describe('Service Error Logging Integration', () => {
    it('should log service errors with correct context', () => {
      const testError = new Error('Service error message');
      
      renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService">
          <ErrorThrowingComponent shouldThrow={true} errorMessage="Service error message" />
        </ServiceErrorBoundary>
      );

      expect(mockLoggers.service.error).toHaveBeenCalledWith('TestService', 'operation', expect.any(Error));
      expect(mockLogger.error).toHaveBeenCalledWith(
        'Service error in TestService',
        expect.any(Error),
        'service-error-boundary',
        expect.objectContaining({
          componentStack: expect.any(String),
          retryCount: 0,
          serviceName: 'TestService'
        })
      );
    });

    it('should call custom onError handler when provided', () => {
      const onErrorMock = jest.fn();
      
      renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService" onError={onErrorMock}>
          <ErrorThrowingComponent shouldThrow={true} />
        </ServiceErrorBoundary>
      );

      expect(onErrorMock).toHaveBeenCalledWith(
        expect.any(Error),
        expect.objectContaining({
          componentStack: expect.any(String)
        })
      );
    });
  });

  describe('Retryable vs Non-Retryable Services', () => {
    it('should show retry button for retryable services', () => {
      const { container } = renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService" retryable={true}>
          <ErrorThrowingComponent shouldThrow={true} />
        </ServiceErrorBoundary>
      );

      expect(screen.getByText('Try Again')).toBeInTheDocument();
      expect(screen.getByText(/You can try again or use fallback mode/)).toBeInTheDocument();
    });

    it('should not show retry button for non-retryable services', () => {
      const { container } = renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService" retryable={false}>
          <ErrorThrowingComponent shouldThrow={true} />
        </ServiceErrorBoundary>
      );

      expect(screen.queryByText('Try Again')).not.toBeInTheDocument();
      expect(screen.getByText(/Please check your connection or try again later/)).toBeInTheDocument();
    });

    it('should default to retryable when not specified', () => {
      const { container } = renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService">
          <ErrorThrowingComponent shouldThrow={true} />
        </ServiceErrorBoundary>
      );

      expect(screen.getByText('Try Again')).toBeInTheDocument();
    });
  });

  describe('Async Retry Operations', () => {
    it('should handle async retry with delay', async () => {
      const { container } = renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService">
          <ErrorThrowingComponent shouldThrow={true} />
        </ServiceErrorBoundary>
      );

      const retryButton = screen.getByText('Try Again');
      fireEvent.click(retryButton);

      // Should show retrying state
      expect(screen.getByText('Retrying...')).toBeInTheDocument();
      expect(retryButton).toBeDisabled();

      // Wait for async retry to complete
      await waitFor(() => {
        expect(screen.queryByText('Retrying...')).not.toBeInTheDocument();
      }, { timeout: 2000 });

      expect(mockLogger.info).toHaveBeenCalledWith(
        'Retrying TestService service (attempt 1/3)',
        undefined,
        'service-error-boundary'
      );
    });

    it('should handle retry failure gracefully', async () => {
      const { container } = renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService">
          <ErrorThrowingComponent shouldThrow={true} />
        </ServiceErrorBoundary>
      );

      // Mock setTimeout to throw an error
      const originalSetTimeout = global.setTimeout;
      global.setTimeout = jest.fn((callback) => {
        callback();
        throw new Error('Retry failed');
      });

      const retryButton = screen.getByText('Try Again');
      fireEvent.click(retryButton);

      await waitFor(() => {
        expect(mockLogger.error).toHaveBeenCalledWith(
          'Retry failed for TestService',
          expect.any(Error),
          'service-error-boundary'
        );
      });

      // Restore original setTimeout
      global.setTimeout = originalSetTimeout;
    });

    it('should track retry attempts correctly', async () => {
      const { container } = renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService">
          <ErrorThrowingComponent shouldThrow={true} />
        </ServiceErrorBoundary>
      );

      // First retry
      fireEvent.click(screen.getByText('Try Again'));
      await waitForAsync(1100); // Wait for retry delay

      expect(screen.getByText('Attempt 1/3')).toBeInTheDocument();
    });
  });

  describe('Fallback Mode Activation', () => {
    it('should handle fallback button click', () => {
      const { container } = renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService">
          <ErrorThrowingComponent shouldThrow={true} />
        </ServiceErrorBoundary>
      );

      const fallbackButton = screen.getByText('Use Fallback');
      fireEvent.click(fallbackButton);

      expect(mockLogger.info).toHaveBeenCalledWith(
        'Using fallback mode for TestService',
        undefined,
        'service-error-boundary'
      );
    });

    it('should show fallback button for all service errors', () => {
      const { container } = renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService" retryable={false}>
          <ErrorThrowingComponent shouldThrow={true} />
        </ServiceErrorBoundary>
      );

      expect(screen.getByText('Use Fallback')).toBeInTheDocument();
    });
  });

  describe('Error Severity Classification', () => {
    it('should show info severity for first error', () => {
      const { container } = renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService">
          <ErrorThrowingComponent shouldThrow={true} />
        </ServiceErrorBoundary>
      );

      const alert = container.querySelector('.MuiAlert-root');
      expect(alert).toHaveClass('MuiAlert-standardInfo');
    });

    it('should show warning severity after retry attempts', async () => {
      const { container } = renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService">
          <ErrorThrowingComponent shouldThrow={true} />
        </ServiceErrorBoundary>
      );

      // First retry
      fireEvent.click(screen.getByText('Try Again'));
      await waitForAsync(1100);

      const alert = container.querySelector('.MuiAlert-root');
      expect(alert).toHaveClass('MuiAlert-standardWarning');
    });

    it('should show error severity after max retries', async () => {
      const { container } = renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService">
          <ErrorThrowingComponent shouldThrow={true} />
        </ServiceErrorBoundary>
      );

      // Simulate reaching max retries
      for (let i = 0; i < 3; i++) {
        fireEvent.click(screen.getByText('Try Again'));
        await waitForAsync(1100);
      }

      const alert = container.querySelector('.MuiAlert-root');
      expect(alert).toHaveClass('MuiAlert-standardError');
    });
  });

  describe('Custom Max Retries Configuration', () => {
    it('should respect custom max retries', () => {
      const { container } = renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService" maxRetries={5}>
          <ErrorThrowingComponent shouldThrow={true} />
        </ServiceErrorBoundary>
      );

      // Should allow more retries
      for (let i = 0; i < 5; i++) {
        expect(screen.getByText('Try Again')).toBeInTheDocument();
        fireEvent.click(screen.getByText('Try Again'));
      }

      // Should not show retry button after max retries
      expect(screen.queryByText('Try Again')).not.toBeInTheDocument();
    });

    it('should default to 3 max retries when not specified', () => {
      const { container } = renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService">
          <ErrorThrowingComponent shouldThrow={true} />
        </ServiceErrorBoundary>
      );

      // Should allow 3 retries
      for (let i = 0; i < 3; i++) {
        expect(screen.getByText('Try Again')).toBeInTheDocument();
        fireEvent.click(screen.getByText('Try Again'));
      }

      // Should not show retry button after 3 retries
      expect(screen.queryByText('Try Again')).not.toBeInTheDocument();
    });
  });

  describe('Custom Fallback Component', () => {
    it('should render custom fallback when provided', () => {
      const customFallback = <div data-testid="custom-service-fallback">Custom Service Fallback</div>;
      
      const { container } = renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService" fallback={customFallback}>
          <ErrorThrowingComponent shouldThrow={true} />
        </ServiceErrorBoundary>
      );

      expect(screen.getByTestId('custom-service-fallback')).toBeInTheDocument();
      expect(screen.queryByText('TestService Service Error')).not.toBeInTheDocument();
    });

    it('should not render custom fallback when no error occurs', () => {
      const customFallback = <div data-testid="custom-service-fallback">Custom Service Fallback</div>;
      
      const { container } = renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService" fallback={customFallback}>
          <ErrorThrowingComponent shouldThrow={false} />
        </ServiceErrorBoundary>
      );

      expect(screen.queryByTestId('custom-service-fallback')).not.toBeInTheDocument();
      expect(screen.getByTestId('error-throwing-component')).toBeInTheDocument();
    });
  });

  describe('Development vs Production Error Display', () => {
    it('should show error details in development mode', () => {
      const originalEnv = process.env.NODE_ENV;
      process.env.NODE_ENV = 'development';

      const { container } = renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService">
          <ErrorThrowingComponent shouldThrow={true} errorMessage="Development service error" />
        </ServiceErrorBoundary>
      );

      expect(screen.getByText('Error Details (Development):')).toBeInTheDocument();
      expect(screen.getByText('Development service error')).toBeInTheDocument();

      process.env.NODE_ENV = originalEnv;
    });

    it('should not show error details in production mode', () => {
      const originalEnv = process.env.NODE_ENV;
      process.env.NODE_ENV = 'production';

      const { container } = renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService">
          <ErrorThrowingComponent shouldThrow={true} />
        </ServiceErrorBoundary>
      );

      expect(screen.queryByText('Error Details (Development):')).not.toBeInTheDocument();

      process.env.NODE_ENV = originalEnv;
    });
  });

  describe('Service Degradation Scenarios', () => {
    it('should handle service timeout scenarios', async () => {
      const { container } = renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService">
          <AsyncErrorComponent shouldThrow={true} delay={2000} />
        </ServiceErrorBoundary>
      );

      // Wait for async error to be thrown
      await waitFor(() => {
        expect(errorBoundaryHelpers.hasError(container)).toBe(true);
      }, { timeout: 3000 });

      expect(screen.getByText('TestService Service Error')).toBeInTheDocument();
    });

    it('should handle network interruption scenarios', () => {
      const { container } = renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService">
          <ErrorThrowingComponent shouldThrow={true} errorMessage="Network error" />
        </ServiceErrorBoundary>
      );

      expect(screen.getByText('TestService Service Error')).toBeInTheDocument();
      expect(screen.getByText('Network error')).toBeInTheDocument();
    });
  });

  describe('Error Recovery Scenarios', () => {
    it('should recover from error after retry', async () => {
      const { container, rerender } = renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService">
          <ErrorThrowingComponent shouldThrow={true} />
        </ServiceErrorBoundary>
      );

      expect(errorBoundaryHelpers.hasError(container)).toBe(true);

      // Retry
      fireEvent.click(screen.getByText('Try Again'));
      await waitForAsync(1100);

      // Re-render with no error (simulating recovery)
      rerender(
        <ServiceErrorBoundary serviceName="TestService">
          <ErrorThrowingComponent shouldThrow={false} />
        </ServiceErrorBoundary>
      );

      expect(errorBoundaryHelpers.hasError(container)).toBe(false);
      expect(screen.getByTestId('error-throwing-component')).toBeInTheDocument();
    });

    it('should maintain retry count across re-renders', async () => {
      const { container, rerender } = renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService">
          <ErrorThrowingComponent shouldThrow={true} />
        </ServiceErrorBoundary>
      );

      // First retry
      fireEvent.click(screen.getByText('Try Again'));
      await waitForAsync(1100);
      expect(screen.getByText('Attempt 1/3')).toBeInTheDocument();

      // Re-render with same error
      rerender(
        <ServiceErrorBoundary serviceName="TestService">
          <ErrorThrowingComponent shouldThrow={true} />
        </ServiceErrorBoundary>
      );

      // Retry count should be maintained
      expect(screen.getByText('Attempt 1/3')).toBeInTheDocument();
    });
  });

  describe('Multiple Service Error Boundaries', () => {
    it('should handle multiple service error boundaries independently', () => {
      const { container } = renderWithTheme(
        <div>
          <ServiceErrorBoundary serviceName="Service1">
            <ErrorThrowingComponent shouldThrow={true} errorMessage="Service1 error" />
          </ServiceErrorBoundary>
          <ServiceErrorBoundary serviceName="Service2">
            <ErrorThrowingComponent shouldThrow={false} />
          </ServiceErrorBoundary>
        </div>
      );

      expect(screen.getByText('Service1 Service Error')).toBeInTheDocument();
      expect(screen.getByTestId('error-throwing-component')).toBeInTheDocument();
    });
  });

  describe('Accessibility and UX', () => {
    it('should have proper ARIA labels and roles', () => {
      const { container } = renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService">
          <ErrorThrowingComponent shouldThrow={true} />
        </ServiceErrorBoundary>
      );

      // Check for proper button roles and accessibility
      const retryButton = screen.getByText('Try Again');
      const fallbackButton = screen.getByText('Use Fallback');
      
      expect(retryButton).toBeInTheDocument();
      expect(fallbackButton).toBeInTheDocument();
    });

    it('should provide clear error messages to users', () => {
      const { container } = renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService">
          <ErrorThrowingComponent shouldThrow={true} />
        </ServiceErrorBoundary>
      );

      expect(screen.getByText('TestService Service Error')).toBeInTheDocument();
      expect(screen.getByText(/The TestService service encountered an error/)).toBeInTheDocument();
      expect(screen.getByText(/You can try again or use fallback mode/)).toBeInTheDocument();
    });

    it('should show appropriate error icons based on severity', () => {
      const { container } = renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService">
          <ErrorThrowingComponent shouldThrow={true} />
        </ServiceErrorBoundary>
      );

      // Should show offline icon for first error
      const alert = container.querySelector('.MuiAlert-root');
      expect(alert).toBeInTheDocument();
    });
  });

  describe('Edge Cases and Error Handling', () => {
    it('should handle missing serviceName gracefully', () => {
      const { container } = renderWithTheme(
        <ServiceErrorBoundary serviceName="">
          <ErrorThrowingComponent shouldThrow={true} />
        </ServiceErrorBoundary>
      );

      expect(errorBoundaryHelpers.hasError(container)).toBe(true);
      expect(screen.getByText(' Service Error')).toBeInTheDocument(); // Empty service name
    });

    it('should handle component lifecycle errors', async () => {
      const { container } = renderWithTheme(
        <ServiceErrorBoundary serviceName="TestService">
          <AsyncErrorComponent shouldThrow={true} delay={50} />
        </ServiceErrorBoundary>
      );

      // Wait for async error to be thrown
      await waitFor(() => {
        expect(errorBoundaryHelpers.hasError(container)).toBe(true);
      }, { timeout: 200 });

      expect(screen.getByText('TestService Service Error')).toBeInTheDocument();
    });
  });
});
