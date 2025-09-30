/**
 * IconButton Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-ICONBUTTON-001: IconButton renders with icon
 * - REQ-ICONBUTTON-002: IconButton handles click events
 * - REQ-ICONBUTTON-003: IconButton handles disabled state
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { IconButton } from '../../../../src/components/atoms/IconButton/IconButton';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('IconButton Component', () => {
  test('REQ-ICONBUTTON-001: IconButton renders with icon', () => {
    const component = renderWithProviders(
      <IconButton icon="settings" aria-label="Settings" />
    );
    
    assertComponentBehavior(component, {
      hasRole: 'button'
    });
  });

  test('REQ-ICONBUTTON-002: IconButton handles click events', () => {
    const handleClick = jest.fn();
    const component = renderWithProviders(
      <IconButton icon="play" onClick={handleClick} aria-label="Play" />
    );
    
    component.getByRole('button').click();
    expect(handleClick).toHaveBeenCalledTimes(1);
  });

  test('REQ-ICONBUTTON-003: IconButton handles disabled state', () => {
    const component = renderWithProviders(
      <IconButton icon="stop" disabled aria-label="Stop" />
    );
    
    const button = component.getByRole('button');
    expect(button).toBeDisabled();
  });
});
