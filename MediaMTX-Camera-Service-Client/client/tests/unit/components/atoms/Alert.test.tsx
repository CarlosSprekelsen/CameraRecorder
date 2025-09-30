/**
 * Alert Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-ALERT-001: Alert renders with correct severity
 * - REQ-ALERT-002: Alert handles different severity levels
 * - REQ-ALERT-003: Alert displays message content correctly
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { Alert } from '../../../../src/components/atoms/Alert/Alert';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('Alert Component', () => {
  test('REQ-ALERT-001: Alert renders with correct severity', () => {
    const component = renderWithProviders(
      <Alert severity="error">Error message</Alert>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Error message'],
      hasClass: ['bg-red-50', 'text-red-800']
    });
  });

  test('REQ-ALERT-002: Alert handles different severity levels', () => {
    const component = renderWithProviders(
      <Alert severity="warning">Warning message</Alert>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Warning message'],
      hasClass: ['bg-yellow-50', 'text-yellow-800']
    });
  });

  test('REQ-ALERT-003: Alert displays message content correctly', () => {
    const component = renderWithProviders(
      <Alert severity="info">Information message</Alert>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Information message'],
      hasClass: ['bg-blue-50', 'text-blue-800']
    });
  });
});
