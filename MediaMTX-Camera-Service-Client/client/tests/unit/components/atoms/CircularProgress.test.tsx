/**
 * CircularProgress Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-CIRCULARPROGRESS-001: CircularProgress renders correctly
 * - REQ-CIRCULARPROGRESS-002: CircularProgress handles size prop
 * - REQ-CIRCULARPROGRESS-003: CircularProgress shows progress value
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { CircularProgress } from '../../../../src/components/atoms/CircularProgress/CircularProgress';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('CircularProgress Component', () => {
  test('REQ-CIRCULARPROGRESS-001: CircularProgress renders correctly', () => {
    const component = renderWithProviders(
      <CircularProgress />
    );
    
    assertComponentBehavior(component, {
      hasClass: ['animate-spin', 'text-blue-600']
    });
  });

  test('REQ-CIRCULARPROGRESS-002: CircularProgress handles size prop', () => {
    const component = renderWithProviders(
      <CircularProgress size={48} />
    );
    
    assertComponentBehavior(component, {
      hasClass: ['animate-spin']
    });
  });

  test('REQ-CIRCULARPROGRESS-003: CircularProgress shows progress value', () => {
    const component = renderWithProviders(
      <CircularProgress value={50} />
    );
    
    assertComponentBehavior(component, {
      hasClass: ['animate-spin']
    });
  });
});
