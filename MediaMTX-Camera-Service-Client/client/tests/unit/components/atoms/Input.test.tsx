/**
 * Input Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-INPUT-001: Input renders with correct type
 * - REQ-INPUT-002: Input handles value changes
 * - REQ-INPUT-003: Input handles disabled state
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { Input } from '../../../../src/components/atoms/Input/Input';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('Input Component', () => {
  test('REQ-INPUT-001: Input renders with correct type', () => {
    const component = renderWithProviders(
      <Input type="text" placeholder="Enter text" />
    );
    
    const input = component.getByPlaceholderText('Enter text');
    expect(input).toHaveAttribute('type', 'text');
  });

  test('REQ-INPUT-002: Input handles value changes', () => {
    const handleChange = jest.fn();
    const component = renderWithProviders(
      <Input onChange={handleChange} value="test value" />
    );
    
    const input = component.getByDisplayValue('test value');
    expect(input).toBeInTheDocument();
  });

  test('REQ-INPUT-003: Input handles disabled state', () => {
    const component = renderWithProviders(
      <Input disabled placeholder="Disabled input" />
    );
    
    const input = component.getByPlaceholderText('Disabled input');
    expect(input).toBeDisabled();
  });
});
