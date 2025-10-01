/**
 * Divider Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-DIVIDER-001: Divider renders correctly
 * - REQ-DIVIDER-002: Divider applies correct styling
 * - REQ-DIVIDER-003: Divider handles orientation
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { Divider } from '../../../../src/components/atoms/Divider/Divider';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('Divider Component', () => {
  test('REQ-DIVIDER-001: Divider renders correctly', () => {
    const component = renderWithProviders(
      <Divider />
    );
    
    assertComponentBehavior(component, {
      hasClass: ['border-gray-300', 'w-full', 'border-t']
    });
  });

  test('REQ-DIVIDER-002: Divider applies correct styling', () => {
    const component = renderWithProviders(
      <Divider className="custom-divider" />
    );
    
    assertComponentBehavior(component, {
      hasClass: ['custom-divider', 'border-gray-300', 'w-full', 'border-t']
    });
  });

  test('REQ-DIVIDER-003: Divider handles orientation', () => {
    const component = renderWithProviders(
      <Divider orientation="vertical" />
    );
    
    assertComponentBehavior(component, {
      hasClass: ['border-gray-300', 'h-full', 'border-l']
    });
  });
});
