/**
 * Switch Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-SWITCH-001: Switch renders with correct state
 * - REQ-SWITCH-002: Switch handles toggle events
 * - REQ-SWITCH-003: Switch handles disabled state
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { Switch } from '../../../../src/components/atoms/Switch/Switch';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('Switch Component', () => {
  test('REQ-SWITCH-001: Switch renders with correct state', () => {
    const component = renderWithProviders(
      <Switch checked={true} />
    );
    
    const switchElement = component.getByRole('switch');
    expect(switchElement).toBeChecked();
  });

  test('REQ-SWITCH-002: Switch handles toggle events', () => {
    const handleChange = jest.fn();
    const component = renderWithProviders(
      <Switch onChange={handleChange} checked={false} />
    );
    
    const switchElement = component.getByRole('switch');
    switchElement.click();
    expect(handleChange).toHaveBeenCalledTimes(1);
  });

  test('REQ-SWITCH-003: Switch handles disabled state', () => {
    const component = renderWithProviders(
      <Switch disabled checked={false} />
    );
    
    const switchElement = component.getByRole('switch');
    expect(switchElement).toBeDisabled();
  });
});
