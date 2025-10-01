/**
 * Icon Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-ICON-001: Icon renders with correct name
 * - REQ-ICON-002: Icon handles size prop correctly
 * - REQ-ICON-003: Icon applies custom styling
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { Icon } from '../../../../src/components/atoms/Icon/Icon';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('Icon Component', () => {
  test('REQ-ICON-001: Icon renders with correct name', () => {
    const component = renderWithProviders(
      <Icon name="settings" />
    );
    
    assertComponentBehavior(component, {
      hasClass: ['icon-settings']
    });
  });

  test('REQ-ICON-002: Icon handles size prop correctly', () => {
    const component = renderWithProviders(
      <Icon name="camera" size={32} />
    );
    
    assertComponentBehavior(component, {
      hasClass: ['icon-camera']
    });
  });

  test('REQ-ICON-003: Icon applies custom styling', () => {
    const component = renderWithProviders(
      <Icon name="stop" className="custom-icon" />
    );
    
    assertComponentBehavior(component, {
      hasClass: ['custom-icon', 'icon-stop']
    });
  });
});
