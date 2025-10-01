/**
 * Dialog Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-DIALOG-001: Dialog renders when open
 * - REQ-DIALOG-002: Dialog handles close events
 * - REQ-DIALOG-003: Dialog displays content correctly
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { Dialog } from '../../../../src/components/atoms/Dialog/Dialog';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('Dialog Component', () => {
  test('REQ-DIALOG-001: Dialog renders when open', () => {
    const component = renderWithProviders(
      <Dialog open={true} onClose={() => {}}>
        <div>Dialog content</div>
      </Dialog>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Dialog content'],
      hasClass: ['fixed', 'inset-0']
    });
  });

  test('REQ-DIALOG-002: Dialog handles close events', () => {
    const handleClose = jest.fn();
    const component = renderWithProviders(
      <Dialog open={true} onClose={handleClose}>
        <div>Dialog content</div>
      </Dialog>
    );
    
    const closeButton = component.getByRole('button', { name: /close/i });
    closeButton.click();
    expect(handleClose).toHaveBeenCalledTimes(1);
  });

  test('REQ-DIALOG-003: Dialog displays content correctly', () => {
    const component = renderWithProviders(
      <Dialog open={true} onClose={() => {}}>
        <div>Modal content</div>
      </Dialog>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Modal content'],
      hasClass: ['dialog', 'fixed', 'inset-0']
    });
  });
});
