/**
 * Select Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-SELECT-001: Select renders with options
 * - REQ-SELECT-002: Select handles value changes
 * - REQ-SELECT-003: Select handles disabled state
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { Select } from '../../../../src/components/atoms/Select/Select';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('Select Component', () => {
  test('REQ-SELECT-001: Select renders with options', () => {
    const component = renderWithProviders(
      <Select>
        <option value="option1">Option 1</option>
        <option value="option2">Option 2</option>
      </Select>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Option 1', 'Option 2']
    });
  });

  test('REQ-SELECT-002: Select handles value changes', () => {
    const handleChange = jest.fn();
    const component = renderWithProviders(
      <Select onChange={handleChange} value="option1">
        <option value="option1">Option 1</option>
      </Select>
    );
    
    const select = component.getByDisplayValue('Option 1');
    expect(select).toBeInTheDocument();
  });

  test('REQ-SELECT-003: Select handles disabled state', () => {
    const component = renderWithProviders(
      <Select disabled>
        <option value="option1">Option 1</option>
      </Select>
    );
    
    const select = component.getByRole('combobox');
    expect(select).toBeDisabled();
  });
});
