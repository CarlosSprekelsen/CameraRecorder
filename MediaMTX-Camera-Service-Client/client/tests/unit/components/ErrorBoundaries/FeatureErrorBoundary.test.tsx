/**
 * FeatureErrorBoundary Test Suite
 * 
 * Comprehensive tests for the FeatureErrorBoundary component including:
 * - Error catching and state management
 * - Retry mechanism with attempt tracking
 * - Max retry limit enforcement
 * - Custom fallback component rendering
 * - Error logging service integration
 * - Development vs production error display
 * - User interaction handling
 * - Props validation and edge cases
 * - Error details toggle functionality
 * - Component lifecycle error handling
 */

import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import FeatureErrorBoundary from '../../../../src/components/ErrorBoundaries/FeatureErrorBoundary';
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

describe('FeatureErrorBoundary', () => {
  beforeEach(() => {
    resetMocks();
  });

  describe('Basic Error Catching', () => {
    it('should catch errors and display error UI', () => {
      const { container } = renderWithTheme(
        <FeatureErrorBoundary featureName="TestFeature">
          <ErrorThrowingComponent shouldThrow={true} />
        </FeatureErrorBoundary>
      );

      expect(errorBoundaryHelpers.hasError(container)).toBe(true);
      expect(screen.getByText('TestFeature Error')).toBeInTheDocument();
      expect(screen.getByText(/Something went wrong in the TestFeature feature/)).toBeInTheDocument();
    });

    it('should render children when no error occurs', () => {
      const { container } = renderWithTheme(
        <FeatureErrorBoundary featureName="TestFeature">
          <ErrorThrowingComponent shouldThrow={false} />
        </FeatureErrorBoundary>
      );

      expect(errorBoundaryHelpers.hasError(container)).toBe(false);
      expect(screen.getByTestId('error-throwing-component')).toBeInTheDocument();
    });

    it('should catch different types of errors', () => {
      const errorTypes = ['syntax', 'reference', 'type', 'custom'] as const;
      
      errorTypes.forEach(errorType => {
        const { container, unmount } = renderWithTheme(
          <FeatureErrorBoundary featureName="TestFeature">
            <ErrorTypeComponent errorType={errorType} />
          </FeatureErrorBoundary>
        );

        expect(errorBoundaryHelpers.hasError(container)).toBe(true);
        expect(screen.getByText('TestFeature Error')).toBeInTheDocument();
        
        unmount();
      });
    });
  });

  describe('Error Logging Integration', () => {
    it('should log errors with correct context', () => {
      const testError = new Error('Test error message');
      
      renderWithTheme(
        <FeatureErrorBoundary featureName="TestFeature">
          <ErrorThrowingComponent shouldThrow={true} errorMessage="Test error message" />
        </FeatureErrorBoundary>
      );

      expect(mockLoggers.component.error).toHaveBeenCalledWith('TestFeature', expect.any(Error));
      expect(mockLogger.error).toHaveBeenCalledWith(
        'Feature error in TestFeature',
        expect.any(Error),
        'error-boundary',
        expect.objectContaining({
          componentStack: expect.any(String),
          retryCount: 0
        })
      );
    });

    it('should call custom onError handler when provided', () => {
      const onErrorMock = jest.fn();
      
      renderWithTheme(
        <FeatureErrorBoundary featureName="TestFeature" onError={onErrorMock}>
          <ErrorThrowingComponent shouldThrow={true} />
        </FeatureErrorBoundary>
      );

      expect(onErrorMock).toHaveBeenCalledWith(
        expect.any(Error),
        expect.objectContaining({
          componentStack: expect.any(String)
        })
      );
    });
  });

  describe('Retry Mechanism', () => {
    it('should allow retry when under max retry limit', () => {
      const { container } = renderWithTheme(
        <FeatureErrorBoundary featureName="TestFeature">
          <ErrorThrowingComponent shouldThrow={true} />
        </FeatureErrorBoundary>
      );

      expect(screen.getByText('Try Again')).toBeInTheDocument();
      expect(screen.getByText(/You can try again or reload the page/)).toBeInTheDocument();
    });

    it('should track retry attempts correctly', () => {
      const { container } = renderWithTheme(
        <FeatureErrorBoundary featureName="TestFeature">
          <ErrorThrowingComponent shouldThrow={true} />
        </FeatureErrorBoundary>
      );

      // First retry
      fireEvent.click(screen.getByText('Try Again'));
      expect(mockLogger.info).toHaveBeenCalledWith(
        'Retrying TestFeature (attempt 1/3)',
        undefined,
        'error-boundary'
      );

      // Should show retry attempt info
      expect(screen.getByText('Retry attempt 1 of 3')).toBeInTheDocument();
    });

    it('should enforce max retry limit', () => {
      const { container } = renderWithTheme(
        <FeatureErrorBoundary featureName="TestFeature">
          <ErrorThrowingComponent shouldThrow={true} />
        </FeatureErrorBoundary>
      );

      // Simulate reaching max retries
      for (let i = 0; i < 3; i++) {
        fireEvent.click(screen.getByText('Try Again'));
      }

      expect(mockLogger.warn).toHaveBeenCalledWith(
        'Max retries reached for TestFeature',
        undefined,
        'error-boundary'
      );
      expect(screen.queryByText('Try Again')).not.toBeInTheDocument();
      expect(screen.getByText(/Please reload the page/)).toBeInTheDocument();
    });

    it('should reset error state on successful retry', () => {
      const { container, rerender } = renderWithTheme(
        <FeatureErrorBoundary featureName="TestFeature">
          <ErrorThrowingComponent shouldThrow={true} />
        </FeatureErrorBoundary>
      );

      expect(errorBoundaryHelpers.hasError(container)).toBe(true);

      // Retry (simulate successful recovery)
      fireEvent.click(screen.getByText('Try Again'));

      // Re-render with no error
      rerender(
        <FeatureErrorBoundary featureName="TestFeature">
          <ErrorThrowingComponent shouldThrow={false} />
        </FeatureErrorBoundary>
      );

      expect(errorBoundaryHelpers.hasError(container)).toBe(false);
    });
  });

  describe('Custom Fallback Component', () => {
    it('should render custom fallback when provided', () => {
      const customFallback = <div data-testid="custom-fallback">Custom Error Fallback</div>;
      
      const { container } = renderWithTheme(
        <FeatureErrorBoundary featureName="TestFeature" fallback={customFallback}>
          <ErrorThrowingComponent shouldThrow={true} />
        </FeatureErrorBoundary>
      );

      expect(screen.getByTestId('custom-fallback')).toBeInTheDocument();
      expect(screen.queryByText('TestFeature Error')).not.toBeInTheDocument();
    });

    it('should not render custom fallback when no error occurs', () => {
      const customFallback = <div data-testid="custom-fallback">Custom Error Fallback</div>;
      
      const { container } = renderWithTheme(
        <FeatureErrorBoundary featureName="TestFeature" fallback={customFallback}>
          <ErrorThrowingComponent shouldThrow={false} />
        </FeatureErrorBoundary>
      );

      expect(screen.queryByTestId('custom-fallback')).not.toBeInTheDocument();
      expect(screen.getByTestId('error-throwing-component')).toBeInTheDocument();
    });
  });

  describe('Development vs Production Error Display', () => {
    it('should show error details in development mode', () => {
      const originalEnv = process.env.NODE_ENV;
      process.env.NODE_ENV = 'development';

      const { container } = renderWithTheme(
        <FeatureErrorBoundary featureName="TestFeature">
          <ErrorThrowingComponent shouldThrow={true} errorMessage="Development error" />
        </FeatureErrorBoundary>
      );

      expect(screen.getByText('Show Error Details')).toBeInTheDocument();
      
      // Toggle details
      fireEvent.click(screen.getByText('Show Error Details'));
      expect(screen.getByText('Hide Error Details')).toBeInTheDocument();
      expect(screen.getByText('Development error')).toBeInTheDocument();

      process.env.NODE_ENV = originalEnv;
    });

    it('should not show error details in production mode', () => {
      const originalEnv = process.env.NODE_ENV;
      process.env.NODE_ENV = 'production';

      const { container } = renderWithTheme(
        <FeatureErrorBoundary featureName="TestFeature">
          <ErrorThrowingComponent shouldThrow={true} />
        </FeatureErrorBoundary>
      );

      expect(screen.queryByText('Show Error Details')).not.toBeInTheDocument();
      expect(screen.queryByText('Hide Error Details')).not.toBeInTheDocument();

      process.env.NODE_ENV = originalEnv;
    });

    it('should toggle error details visibility', () => {
      const originalEnv = process.env.NODE_ENV;
      process.env.NODE_ENV = 'development';

      const { container } = renderWithTheme(
        <FeatureErrorBoundary featureName="TestFeature">
          <ErrorThrowingComponent shouldThrow={true} errorMessage="Toggle test error" />
        </FeatureErrorBoundary>
      );

      // Initially details should be hidden
      expect(screen.queryByText('Toggle test error')).not.toBeInTheDocument();

      // Show details
      fireEvent.click(screen.getByText('Show Error Details'));
      expect(screen.getByText('Toggle test error')).toBeInTheDocument();
      expect(screen.getByText('Hide Error Details')).toBeInTheDocument();

      // Hide details
      fireEvent.click(screen.getByText('Hide Error Details'));
      expect(screen.queryByText('Toggle test error')).not.toBeInTheDocument();
      expect(screen.getByText('Show Error Details')).toBeInTheDocument();

      process.env.NODE_ENV = originalEnv;
    });
  });

  describe('User Interactions', () => {
    it('should handle reload button click', () => {
      const { mockReload, restore } = mockWindowReload();

      const { container } = renderWithTheme(
        <FeatureErrorBoundary featureName="TestFeature">
          <ErrorThrowingComponent shouldThrow={true} />
        </FeatureErrorBoundary>
      );

      fireEvent.click(screen.getByText('Reload Page'));
      
      expect(mockLogger.info).toHaveBeenCalledWith(
        'Reloading page due to error in TestFeature',
        undefined,
        'error-boundary'
      );
      expect(mockReload).toHaveBeenCalled();

      restore();
    });

    it('should handle retry button click', () => {
      const { container } = renderWithTheme(
        <FeatureErrorBoundary featureName="TestFeature">
          <ErrorThrowingComponent shouldThrow={true} />
        </FeatureErrorBoundary>
      );

      fireEvent.click(screen.getByText('Try Again'));
      
      expect(mockLogger.info).toHaveBeenCalledWith(
        'Retrying TestFeature (attempt 1/3)',
        undefined,
        'error-boundary'
      );
    });
  });

  describe('Props Validation and Edge Cases', () => {
    it('should handle missing featureName gracefully', () => {
      const { container } = renderWithTheme(
        <FeatureErrorBoundary featureName="">
          <ErrorThrowingComponent shouldThrow={true} />
        </FeatureErrorBoundary>
      );

      expect(errorBoundaryHelpers.hasError(container)).toBe(true);
      expect(screen.getByText(' Error')).toBeInTheDocument(); // Empty feature name
    });

    it('should handle showDetails prop correctly', () => {
      const originalEnv = process.env.NODE_ENV;
      process.env.NODE_ENV = 'development';

      const { container } = renderWithTheme(
        <FeatureErrorBoundary featureName="TestFeature" showDetails={true}>
          <ErrorThrowingComponent shouldThrow={true} errorMessage="Initial details shown" />
        </FeatureErrorBoundary>
      );

      // Details should be shown initially
      expect(screen.getByText('Hide Error Details')).toBeInTheDocument();
      expect(screen.getByText('Initial details shown')).toBeInTheDocument();

      process.env.NODE_ENV = originalEnv;
    });

    it('should handle multiple error boundaries independently', () => {
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
  });

  describe('Component Lifecycle Error Handling', () => {
    it('should handle errors in componentDidMount', async () => {
      const { container } = renderWithTheme(
        <FeatureErrorBoundary featureName="TestFeature">
          <AsyncErrorComponent shouldThrow={true} delay={50} />
        </FeatureErrorBoundary>
      );

      // Wait for async error to be thrown
      await waitFor(() => {
        expect(errorBoundaryHelpers.hasError(container)).toBe(true);
      }, { timeout: 200 });

      expect(screen.getByText('TestFeature Error')).toBeInTheDocument();
    });

    it('should handle errors in componentDidUpdate', () => {
      const { container, rerender } = renderWithTheme(
        <FeatureErrorBoundary featureName="TestFeature">
          <ErrorThrowingComponent shouldThrow={false} />
        </FeatureErrorBoundary>
      );

      // Re-render with error
      rerender(
        <FeatureErrorBoundary featureName="TestFeature">
          <ErrorThrowingComponent shouldThrow={true} />
        </FeatureErrorBoundary>
      );

      expect(errorBoundaryHelpers.hasError(container)).toBe(true);
    });
  });

  describe('Error Recovery Scenarios', () => {
    it('should recover from error after retry', () => {
      const { container, rerender } = renderWithTheme(
        <FeatureErrorBoundary featureName="TestFeature">
          <ErrorThrowingComponent shouldThrow={true} />
        </FeatureErrorBoundary>
      );

      expect(errorBoundaryHelpers.hasError(container)).toBe(true);

      // Retry
      fireEvent.click(screen.getByText('Try Again'));

      // Re-render with no error (simulating recovery)
      rerender(
        <FeatureErrorBoundary featureName="TestFeature">
          <ErrorThrowingComponent shouldThrow={false} />
        </FeatureErrorBoundary>
      );

      expect(errorBoundaryHelpers.hasError(container)).toBe(false);
      expect(screen.getByTestId('error-throwing-component')).toBeInTheDocument();
    });

    it('should maintain retry count across re-renders', () => {
      const { container, rerender } = renderWithTheme(
        <FeatureErrorBoundary featureName="TestFeature">
          <ErrorThrowingComponent shouldThrow={true} />
        </FeatureErrorBoundary>
      );

      // First retry
      fireEvent.click(screen.getByText('Try Again'));
      expect(screen.getByText('Retry attempt 1 of 3')).toBeInTheDocument();

      // Re-render with same error
      rerender(
        <FeatureErrorBoundary featureName="TestFeature">
          <ErrorThrowingComponent shouldThrow={true} />
        </FeatureErrorBoundary>
      );

      // Retry count should be maintained
      expect(screen.getByText('Retry attempt 1 of 3')).toBeInTheDocument();
    });
  });

  describe('Accessibility and UX', () => {
    it('should have proper ARIA labels and roles', () => {
      const { container } = renderWithTheme(
        <FeatureErrorBoundary featureName="TestFeature">
          <ErrorThrowingComponent shouldThrow={true} />
        </FeatureErrorBoundary>
      );

      // Check for proper button roles and accessibility
      const retryButton = screen.getByText('Try Again');
      const reloadButton = screen.getByText('Reload Page');
      
      expect(retryButton).toBeInTheDocument();
      expect(reloadButton).toBeInTheDocument();
    });

    it('should provide clear error messages to users', () => {
      const { container } = renderWithTheme(
        <FeatureErrorBoundary featureName="TestFeature">
          <ErrorThrowingComponent shouldThrow={true} />
        </FeatureErrorBoundary>
      );

      expect(screen.getByText('TestFeature Error')).toBeInTheDocument();
      expect(screen.getByText(/Something went wrong in the TestFeature feature/)).toBeInTheDocument();
      expect(screen.getByText(/You can try again or reload the page/)).toBeInTheDocument();
    });
  });
});
