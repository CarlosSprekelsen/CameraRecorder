/**
 * Badge Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-BADGE-001: Badge renders with correct color
 * - REQ-BADGE-002: Badge displays content correctly
 * - REQ-BADGE-003: Badge handles different variants
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { Badge } from '../../../../src/components/atoms/Badge/Badge';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('Badge Component', () => {
  test('REQ-BADGE-001: Badge renders with correct color', () => {
    const component = renderWithProviders(
      <Badge color="success">Success badge</Badge>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Success badge'],
      hasClass: ['bg-green-100', 'text-green-800']
    });
  });

  test('REQ-BADGE-002: Badge displays content correctly', () => {
    const component = renderWithProviders(
      <Badge color="error">Error badge</Badge>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Error badge'],
      hasClass: ['bg-red-100', 'text-red-800']
    });
  });

  test('REQ-BADGE-003: Badge handles different variants', () => {
    const component = renderWithProviders(
      <Badge color="warning">Warning badge</Badge>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Warning badge'],
      hasClass: ['bg-yellow-100', 'text-yellow-800']
    });
  });
});
