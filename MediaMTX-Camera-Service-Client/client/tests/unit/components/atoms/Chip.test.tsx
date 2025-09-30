/**
 * Chip Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-CHIP-001: Chip renders with correct label
 * - REQ-CHIP-002: Chip handles different colors
 * - REQ-CHIP-003: Chip handles click events
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { Chip } from '../../../../src/components/atoms/Chip/Chip';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('Chip Component', () => {
  test('REQ-CHIP-001: Chip renders with correct label', () => {
    const component = renderWithProviders(
      <Chip label="Status tag" />
    );
    
    assertComponentBehavior(component, {
      hasText: ['Status tag'],
      hasClass: ['bg-gray-100', 'text-gray-800']
    });
  });

  test('REQ-CHIP-002: Chip handles different colors', () => {
    const component = renderWithProviders(
      <Chip label="Success" color="success" />
    );
    
    assertComponentBehavior(component, {
      hasText: ['Success'],
      hasClass: ['bg-green-100', 'text-green-800']
    });
  });

  test('REQ-CHIP-003: Chip handles click events', () => {
    const handleClick = jest.fn();
    const component = renderWithProviders(
      <Chip label="Clickable" onClick={handleClick} />
    );
    
    component.getByText('Clickable').click();
    expect(handleClick).toHaveBeenCalledTimes(1);
  });
});
