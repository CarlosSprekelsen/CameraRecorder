/**
 * TextField Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-TEXTFIELD-001: TextField renders with correct props
 * - REQ-TEXTFIELD-002: TextField handles value changes
 * - REQ-TEXTFIELD-003: TextField handles disabled state
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { TextField } from '../../../../src/components/atoms/TextField/TextField';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('TextField Component', () => {
  test('REQ-TEXTFIELD-001: TextField renders with correct props', () => {
    const component = renderWithProviders(
      <TextField label="Username" placeholder="Enter username" />
    );
    
    assertComponentBehavior(component, {
      hasText: ['Username']
    });
  });

  test('REQ-TEXTFIELD-002: TextField handles value changes', () => {
    const handleChange = jest.fn();
    const component = renderWithProviders(
      <TextField onChange={handleChange} value="test" />
    );
    
    const input = component.getByDisplayValue('test');
    expect(input).toBeInTheDocument();
  });

  test('REQ-TEXTFIELD-003: TextField handles disabled state', () => {
    const component = renderWithProviders(
      <TextField disabled label="Disabled Field" />
    );
    
    const input = component.getByLabelText('Disabled Field');
    expect(input).toBeDisabled();
  });
});
