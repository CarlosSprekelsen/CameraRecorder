/**
 * Button Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-UI-001: Button renders with correct variant styles
 * - REQ-UI-002: Button handles disabled state correctly
 * - REQ-UI-003: Button handles click events properly
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { Button } from '../../../../src/components/atoms/Button/Button';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('Button Component', () => {
  test('REQ-UI-001: Button renders with correct variant styles', () => {
    const component = renderWithProviders(
      <Button variant="primary">Test Button</Button>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Test Button'],
      hasClass: ['bg-blue-600']
    });
  });

  test('REQ-UI-002: Button handles disabled state correctly', () => {
    const component = renderWithProviders(
      <Button disabled>Disabled Button</Button>
    );
    
    assertComponentBehavior(component, {
      isDisabled: true,
      hasText: ['Disabled Button']
    });
  });

  test('REQ-UI-003: Button handles click events properly', () => {
    const handleClick = jest.fn();
    const component = renderWithProviders(
      <Button onClick={handleClick}>Clickable Button</Button>
    );
    
    component.getByRole('button').click();
    expect(handleClick).toHaveBeenCalledTimes(1);
  });
});
