/**
 * REQ-UNIT01-001: [Primary requirement being tested]
 * REQ-UNIT01-002: [Secondary requirements covered]
 * Coverage: UNIT
 * Quality: HIGH
 */
/**
 * Simple component test to validate React testing infrastructure
 * This test validates that the testing environment is properly configured
 */

import React from 'react';
import { render, screen } from '@testing-library/react';

// Simple test component
const TestComponent = ({ message }: { message: string }) => {
  return <div data-testid="test-component">{message}</div>;
};

describe('Simple Component Test', () => {
  it('should render a simple component', () => {
    render(<TestComponent message="Hello World" />);
    
    const element = screen.getByTestId('test-component');
    expect(element).toBeInTheDocument();
    expect(element).toHaveTextContent('Hello World');
  });

  it('should handle component props correctly', () => {
    const testMessage = 'Test Message';
    render(<TestComponent message={testMessage} />);
    
    expect(screen.getByText(testMessage)).toBeInTheDocument();
  });
});
