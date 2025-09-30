/**
 * Checkbox Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-CHECKBOX-001: Checkbox renders with correct state
 * - REQ-CHECKBOX-002: Checkbox handles change events
 * - REQ-CHECKBOX-003: Checkbox handles disabled state
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { Checkbox } from '../../../../src/components/atoms/Checkbox/Checkbox';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('Checkbox Component', () => {
  test('REQ-CHECKBOX-001: Checkbox renders with correct state', () => {
    const component = renderWithProviders(
      <Checkbox checked={true} />
    );
    
    const checkbox = component.getByRole('checkbox');
    expect(checkbox).toBeChecked();
  });

  test('REQ-CHECKBOX-002: Checkbox handles change events', () => {
    const handleChange = jest.fn();
    const component = renderWithProviders(
      <Checkbox onChange={handleChange} checked={false} />
    );
    
    const checkbox = component.getByRole('checkbox');
    checkbox.click();
    expect(handleChange).toHaveBeenCalledTimes(1);
  });

  test('REQ-CHECKBOX-003: Checkbox handles disabled state', () => {
    const component = renderWithProviders(
      <Checkbox disabled checked={false} />
    );
    
    const checkbox = component.getByRole('checkbox');
    expect(checkbox).toBeDisabled();
  });
});
