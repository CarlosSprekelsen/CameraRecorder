/**
 * Snackbar Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-SNACKBAR-001: Snackbar renders when open
 * - REQ-SNACKBAR-002: Snackbar displays message
 * - REQ-SNACKBAR-003: Snackbar handles close events
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { Snackbar } from '../../../../src/components/atoms/Snackbar/Snackbar';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('Snackbar Component', () => {
  test('REQ-SNACKBAR-001: Snackbar renders when open', () => {
    const component = renderWithProviders(
      <Snackbar open={true} message="Notification message" />
    );
    
    assertComponentBehavior(component, {
      hasText: ['Notification message'],
      hasClass: ['snackbar', 'fixed', 'bottom-4']
    });
  });

  test('REQ-SNACKBAR-002: Snackbar displays message', () => {
    const component = renderWithProviders(
      <Snackbar open={true} message="Success notification" />
    );
    
    assertComponentBehavior(component, {
      hasText: ['Success notification'],
      hasClass: ['snackbar']
    });
  });

  test('REQ-SNACKBAR-003: Snackbar handles close events', () => {
    const handleClose = jest.fn();
    const component = renderWithProviders(
      <Snackbar open={true} message="Test message" onClose={handleClose} />
    );
    
    const closeButton = component.getByRole('button', { name: /close/i });
    closeButton.click();
    expect(handleClose).toHaveBeenCalledTimes(1);
  });
});
