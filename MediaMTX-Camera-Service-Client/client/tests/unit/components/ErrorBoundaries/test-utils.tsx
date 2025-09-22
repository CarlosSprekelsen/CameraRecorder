/**
 * Error Boundary Test Utilities
 * 
 * Provides helper functions and components for testing Error Boundaries
 * in isolation and with controlled error scenarios.
 */

import React, { Component, ReactNode } from 'react';
import { render, RenderOptions } from '@testing-library/react';
import { ThemeProvider } from '@mui/material/styles';
import { theme } from '../../../../src/theme';

// Mock logger service for testing
export const mockLogger = {
  error: jest.fn(),
  warn: jest.fn(),
  info: jest.fn(),
  debug: jest.fn(),
};

export const mockLoggers = {
  component: {
    error: jest.fn(),
  },
  service: {
    error: jest.fn(),
  },
};

// Mock the logger service
jest.mock('../../../../src/services/loggerService', () => ({
  logger: mockLogger,
  loggers: mockLoggers,
}));

/**
 * Component that throws an error when shouldThrow is true
 */
interface ErrorThrowingComponentProps {
  shouldThrow: boolean;
  errorMessage?: string;
  children?: ReactNode;
}

export class ErrorThrowingComponent extends Component<ErrorThrowingComponentProps> {
  constructor(props: ErrorThrowingComponentProps) {
    super(props);
  }

  render() {
    if (this.props.shouldThrow) {
      throw new Error(this.props.errorMessage || 'Test error thrown intentionally');
    }
    return <div data-testid="error-throwing-component">{this.props.children}</div>;
  }
}

/**
 * Component that throws an error after a delay (for async error testing)
 */
interface AsyncErrorComponentProps {
  shouldThrow: boolean;
  delay?: number;
  errorMessage?: string;
}

export class AsyncErrorComponent extends Component<AsyncErrorComponentProps> {
  private timeoutId: NodeJS.Timeout | null = null;

  componentDidMount() {
    if (this.props.shouldThrow) {
      this.timeoutId = setTimeout(() => {
        throw new Error(this.props.errorMessage || 'Async test error');
      }, this.props.delay || 100);
    }
  }

  componentWillUnmount() {
    if (this.timeoutId) {
      clearTimeout(this.timeoutId);
    }
  }

  render() {
    return <div data-testid="async-error-component">Async Component</div>;
  }
}

/**
 * Component that throws different types of errors
 */
interface ErrorTypeComponentProps {
  errorType: 'syntax' | 'reference' | 'type' | 'promise' | 'custom';
  customError?: Error;
}

export class ErrorTypeComponent extends Component<ErrorTypeComponentProps> {
  render() {
    const { errorType, customError } = this.props;

    switch (errorType) {
      case 'syntax':
        // This will cause a syntax error during render
        throw new SyntaxError('Syntax error for testing');
      
      case 'reference':
        // This will cause a reference error
        throw new ReferenceError('Reference error for testing');
      
      case 'type':
        // This will cause a type error
        throw new TypeError('Type error for testing');
      
      case 'promise':
        // This will cause a promise rejection
        Promise.reject(new Error('Promise rejection for testing'));
        return <div data-testid="promise-error-component">Promise Error Component</div>;
      
      case 'custom':
        throw customError || new Error('Custom error for testing');
      
      default:
        return <div data-testid="error-type-component">Error Type Component</div>;
    }
  }
}

/**
 * Test wrapper that provides theme context
 */
interface TestWrapperProps {
  children: ReactNode;
}

const TestWrapper: React.FC<TestWrapperProps> = ({ children }) => {
  return (
    <ThemeProvider theme={theme}>
      {children}
    </ThemeProvider>
  );
};

/**
 * Custom render function with theme provider
 */
export const renderWithTheme = (
  ui: React.ReactElement,
  options?: Omit<RenderOptions, 'wrapper'>
) => {
  return render(ui, { wrapper: TestWrapper, ...options });
};

/**
 * Helper to simulate user interactions
 */
export const userInteractions = {
  clickRetry: (container: HTMLElement) => {
    const retryButton = container.querySelector('[data-testid="retry-button"]') as HTMLButtonElement;
    if (retryButton) {
      retryButton.click();
    }
  },
  
  clickReload: (container: HTMLElement) => {
    const reloadButton = container.querySelector('[data-testid="reload-button"]') as HTMLButtonElement;
    if (reloadButton) {
      reloadButton.click();
    }
  },
  
  clickFallback: (container: HTMLElement) => {
    const fallbackButton = container.querySelector('[data-testid="fallback-button"]') as HTMLButtonElement;
    if (fallbackButton) {
      fallbackButton.click();
    }
  },
  
  toggleDetails: (container: HTMLElement) => {
    const detailsButton = container.querySelector('[data-testid="details-button"]') as HTMLButtonElement;
    if (detailsButton) {
      detailsButton.click();
    }
  },
};

/**
 * Helper to check error boundary state
 */
export const errorBoundaryHelpers = {
  hasError: (container: HTMLElement) => {
    return container.querySelector('[data-testid="error-boundary"]') !== null;
  },
  
  getErrorMessage: (container: HTMLElement) => {
    const errorMessage = container.querySelector('[data-testid="error-message"]');
    return errorMessage?.textContent || '';
  },
  
  getRetryCount: (container: HTMLElement) => {
    const retryInfo = container.querySelector('[data-testid="retry-info"]');
    return retryInfo?.textContent || '';
  },
  
  isRetrying: (container: HTMLElement) => {
    const retryButton = container.querySelector('[data-testid="retry-button"]') as HTMLButtonElement;
    return retryButton?.disabled || false;
  },
  
  hasDetails: (container: HTMLElement) => {
    return container.querySelector('[data-testid="error-details"]') !== null;
  },
};

/**
 * Mock window.location.reload for testing
 */
export const mockWindowReload = () => {
  const originalReload = window.location.reload;
  const mockReload = jest.fn();
  window.location.reload = mockReload;
  
  return {
    mockReload,
    restore: () => {
      window.location.reload = originalReload;
    },
  };
};

/**
 * Helper to wait for async operations
 */
export const waitForAsync = (ms: number = 100) => {
  return new Promise(resolve => setTimeout(resolve, ms));
};

/**
 * Test data for Error Boundaries
 */
export const testData = {
  featureNames: ['Dashboard', 'CameraDetail', 'FileManager', 'HealthMonitor', 'AdminDashboard', 'Settings'],
  serviceNames: ['ConnectionManager', 'WebSocketService', 'FileService', 'CameraService'],
  errorMessages: [
    'Network connection failed',
    'Service unavailable',
    'Authentication error',
    'Permission denied',
    'Timeout error',
    'Invalid response format',
  ],
  customErrors: [
    new Error('Custom error 1'),
    new Error('Custom error 2'),
    new SyntaxError('Syntax error'),
    new ReferenceError('Reference error'),
    new TypeError('Type error'),
  ],
};

/**
 * Reset all mocks before each test
 */
export const resetMocks = () => {
  jest.clearAllMocks();
  mockLogger.error.mockClear();
  mockLogger.warn.mockClear();
  mockLogger.info.mockClear();
  mockLogger.debug.mockClear();
  mockLoggers.component.error.mockClear();
  mockLoggers.service.error.mockClear();
};
