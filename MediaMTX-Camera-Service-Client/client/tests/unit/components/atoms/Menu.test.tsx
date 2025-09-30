/**
 * Menu Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-MENU-001: Menu renders when open
 * - REQ-MENU-002: Menu handles close events
 * - REQ-MENU-003: Menu displays menu items
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { Menu } from '../../../../src/components/atoms/Menu/Menu';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('Menu Component', () => {
  test('REQ-MENU-001: Menu renders when open', () => {
    const component = renderWithProviders(
      <Menu open={true} onClose={() => {}}>
        <div>Menu item 1</div>
        <div>Menu item 2</div>
      </Menu>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Menu item 1', 'Menu item 2'],
      hasClass: ['menu', 'open']
    });
  });

  test('REQ-MENU-002: Menu handles close events', () => {
    const handleClose = jest.fn();
    const component = renderWithProviders(
      <Menu open={true} onClose={handleClose}>
        <div>Menu content</div>
      </Menu>
    );
    
    const closeButton = component.getByRole('button', { name: /close/i });
    closeButton.click();
    expect(handleClose).toHaveBeenCalledTimes(1);
  });

  test('REQ-MENU-003: Menu displays menu items', () => {
    const component = renderWithProviders(
      <Menu open={true} onClose={() => {}}>
        <div>Dropdown item</div>
      </Menu>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Dropdown item'],
      hasClass: ['menu']
    });
  });
});
