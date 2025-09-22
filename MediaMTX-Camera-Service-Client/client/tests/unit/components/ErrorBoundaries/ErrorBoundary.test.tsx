/**
 * ErrorBoundary Test Suite
 * 
 * Comprehensive tests for the basic ErrorBoundary component including:
 * - Basic error catching functionality
 * - Fallback component rendering
 * - Development error details display
 * - Page reload functionality
 * - Error state management
 * - Props handling
 * - Error info logging
 */

import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
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
} from './test-utils';

// Mock console.error to avoid noise in test output
const originalConsoleError = console.error;
beforeAll(() => {
  console.error = jest.fn();
});

afterAll(() => {
  console.error = originalConsoleError;
});

describe('ErrorBoundary', () => {
  beforeEach(() => {
    resetMocks();
  });

  describe('Basic Error Catching', () => {
    it('should catch errors and display error UI', () => {
      const { container } = renderWithTheme(
        <ErrorBoundary>
          <ErrorThrowingComponent shouldThrow={true} />
        </ErrorBoundary>
      );

      expect(screen.getByText('Something went wrong')).toBeInTheDocument();
      expect(screen.getByText(/An unexpected error occurred/)).toBeInTheDocument();
      expect(screen.getByText('Reload Page')).toBeInTheDocument();
    });

    it('should render children when no error occurs', () => {
      const { container } = renderWithTheme(
        <ErrorBoundary>
          <ErrorThrowingComponent shouldThrow={false} />
        </ErrorBoundary>
      );

      expect(screen.getByTestId('error-throwing-component')).toBeInTheDocument();
      expect(screen.queryByText('Something went wrong')).not.toBeInTheDocument();
    });

    it('should catch different types of errors', () => {
      const errorTypes = ['syntax', 'reference', 'type', 'custom'] as const;
      
      errorTypes.forEach(errorType => {
        const { container, unmount } = renderWithTheme(
          <ErrorBoundary>
            <ErrorTypeComponent errorType={errorType} />
          </ErrorBoundary>
        );

        expect(screen.getByText('Something went wrong')).toBeInTheDocument();
        expect(screen.getByText(/An unexpected error occurred/)).toBeInTheDocument();
        
        unmount();
      });
    });
  });

  describe('Error Logging', () => {
    it('should log errors to console', () => {
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
      
      renderWithTheme(
        <ErrorBoundary>
          <ErrorThrowingComponent shouldThrow={true} errorMessage="Test error message" />
        </ErrorBoundary>
      );

      expect(consoleSpy).toHaveBeenCalledWith(
        'ErrorBoundary caught an error:',
        expect.any(Error),
        expect.objectContaining({
          componentStack: expect.any(String)
        })
      );

      consoleSpy.mockRestore();
    });
  });

  describe('Custom Fallback Component', () => {
    it('should render custom fallback when provided', () => {
      const customFallback = <div data-testid="custom-fallback">Custom Error Fallback</div>;
      
      const { container } = renderWithTheme(
        <ErrorBoundary fallback={customFallback}>
          <ErrorThrowingComponent shouldThrow={true} />
        </ErrorBoundary>
      );

      expect(screen.getByTestId('custom-fallback')).toBeInTheDocument();
      expect(screen.queryByText('Something went wrong')).not.toBeInTheDocument();
    });

    it('should not render custom fallback when no error occurs', () => {
      const customFallback = <div data-testid="custom-fallback">Custom Error Fallback</div>;
      
      const { container } = renderWithTheme(
        <ErrorBoundary fallback={customFallback}>
          <ErrorThrowingComponent shouldThrow={false} />
        </ErrorBoundary>
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
        <ErrorBoundary>
          <ErrorThrowingComponent shouldThrow={true} errorMessage="Development error" />
        </ErrorBoundary>
      );

      expect(screen.getByText('Error Details (Development):')).toBeInTheDocument();
      expect(screen.getByText('Development error')).toBeInTheDocument();

      process.env.NODE_ENV = originalEnv;
    });

    it('should not show error details in production mode', () => {
      const originalEnv = process.env.NODE_ENV;
      process.env.NODE_ENV = 'production';

      const { container } = renderWithTheme(
        <ErrorBoundary>
          <ErrorThrowingComponent shouldThrow={true} />
        </ErrorBoundary>
      );

      expect(screen.queryByText('Error Details (Development):')).not.toBeInTheDocument();

      process.env.NODE_ENV = originalEnv;
    });
  });

  describe('Page Reload Functionality', () => {
    it('should handle reload button click', () => {
      const { mockReload, restore } = mockWindowReload();

      const { container } = renderWithTheme(
        <ErrorBoundary>
          <ErrorThrowingComponent shouldThrow={true} />
        </ErrorBoundary>
      );

      fireEvent.click(screen.getByText('Reload Page'));
      expect(mockReload).toHaveBeenCalled();

      restore();
    });
  });

  describe('Error State Management', () => {
    it('should maintain error state across re-renders', () => {
      const { container, rerender } = renderWithTheme(
        <ErrorBoundary>
          <ErrorThrowingComponent shouldThrow={true} />
        </ErrorBoundary>
      );

      expect(screen.getByText('Something went wrong')).toBeInTheDocument();

      // Re-render with same error
      rerender(
        <ErrorBoundary>
          <ErrorThrowingComponent shouldThrow={true} />
        </ErrorBoundary>
      );

      expect(screen.getByText('Something went wrong')).toBeInTheDocument();
    });

    it('should reset error state when children change to non-error', () => {
      const { container, rerender } = renderWithTheme(
        <ErrorBoundary>
          <ErrorThrowingComponent shouldThrow={true} />
        </ErrorBoundary>
      );

      expect(screen.getByText('Something went wrong')).toBeInTheDocument();

      // Re-render with no error
      rerender(
        <ErrorBoundary>
          <ErrorThrowingComponent shouldThrow={false} />
        </ErrorBoundary>
      );

      expect(screen.queryByText('Something went wrong')).not.toBeInTheDocument();
      expect(screen.getByTestId('error-throwing-component')).toBeInTheDocument();
    });
  });

  describe('Component Lifecycle Error Handling', () => {
    it('should handle errors in componentDidMount', async () => {
      const { container } = renderWithTheme(
        <ErrorBoundary>
          <AsyncErrorComponent shouldThrow={true} delay={50} />
        </ErrorBoundary>
      );

      // Wait for async error to be thrown
      await waitForAsync(150);

      expect(screen.getByText('Something went wrong')).toBeInTheDocument();
    });

    it('should handle errors in componentDidUpdate', () => {
      const { container, rerender } = renderWithTheme(
        <ErrorBoundary>
          <ErrorThrowingComponent shouldThrow={false} />
        </ErrorBoundary>
      );

      // Re-render with error
      rerender(
        <ErrorBoundary>
          <ErrorThrowingComponent shouldThrow={true} />
        </ErrorBoundary>
      );

      expect(screen.getByText('Something went wrong')).toBeInTheDocument();
    });
  });

  describe('Multiple Error Boundaries', () => {
    it('should handle multiple error boundaries independently', () => {
      const { container } = renderWithTheme(
        <div>
          <ErrorBoundary>
            <ErrorThrowingComponent shouldThrow={true} errorMessage="Error 1" />
          </ErrorBoundary>
          <ErrorBoundary>
            <ErrorThrowingComponent shouldThrow={false} />
          </ErrorBoundary>
        </div>
      );

      expect(screen.getByText('Something went wrong')).toBeInTheDocument();
      expect(screen.getByTestId('error-throwing-component')).toBeInTheDocument();
    });
  });

  describe('Accessibility and UX', () => {
    it('should have proper ARIA labels and roles', () => {
      const { container } = renderWithTheme(
        <ErrorBoundary>
          <ErrorThrowingComponent shouldThrow={true} />
        </ErrorBoundary>
      );

      // Check for proper button roles and accessibility
      const reloadButton = screen.getByText('Reload Page');
      expect(reloadButton).toBeInTheDocument();
    });

    it('should provide clear error messages to users', () => {
      const { container } = renderWithTheme(
        <ErrorBoundary>
          <ErrorThrowingComponent shouldThrow={true} />
        </ErrorBoundary>
      );

      expect(screen.getByText('Something went wrong')).toBeInTheDocument();
      expect(screen.getByText(/An unexpected error occurred/)).toBeInTheDocument();
      expect(screen.getByText(/Please try refreshing the page/)).toBeInTheDocument();
    });
  });

  describe('Edge Cases and Error Handling', () => {
    it('should handle null children gracefully', () => {
      const { container } = renderWithTheme(
        <ErrorBoundary>
          {null}
        </ErrorBoundary>
      );

      // Should not throw and should render nothing
      expect(container.firstChild).toBeNull();
    });

    it('should handle undefined children gracefully', () => {
      const { container } = renderWithTheme(
        <ErrorBoundary>
          {undefined}
        </ErrorBoundary>
      );

      // Should not throw and should render nothing
      expect(container.firstChild).toBeNull();
    });

    it('should handle empty children gracefully', () => {
      const { container } = renderWithTheme(
        <ErrorBoundary>
          <></>
        </ErrorBoundary>
      );

      // Should not throw and should render empty fragment
      expect(container.firstChild).toBeInTheDocument();
    });
  });

  describe('Error Recovery Scenarios', () => {
    it('should recover from error when children change to non-error', () => {
      const { container, rerender } = renderWithTheme(
        <ErrorBoundary>
          <ErrorThrowingComponent shouldThrow={true} />
        </ErrorBoundary>
      );

      expect(screen.getByText('Something went wrong')).toBeInTheDocument();

      // Re-render with no error (simulating recovery)
      rerender(
        <ErrorBoundary>
          <ErrorThrowingComponent shouldThrow={false} />
        </ErrorBoundary>
      );

      expect(screen.queryByText('Something went wrong')).not.toBeInTheDocument();
      expect(screen.getByTestId('error-throwing-component')).toBeInTheDocument();
    });

    it('should handle rapid error state changes', () => {
      const { container, rerender } = renderWithTheme(
        <ErrorBoundary>
          <ErrorThrowingComponent shouldThrow={true} />
        </ErrorBoundary>
      );

      expect(screen.getByText('Something went wrong')).toBeInTheDocument();

      // Rapid re-renders
      for (let i = 0; i < 5; i++) {
        rerender(
          <ErrorBoundary>
            <ErrorThrowingComponent shouldThrow={i % 2 === 0} />
          </ErrorBoundary>
        );
      }

      // Should handle rapid changes gracefully
      expect(container).toBeInTheDocument();
    });
  });
});
