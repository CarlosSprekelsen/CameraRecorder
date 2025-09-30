/**
 * Typography Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-UI-007: Typography renders with correct variant
 * - REQ-UI-008: Typography handles color prop correctly
 * - REQ-UI-009: Typography handles component prop correctly
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { Typography } from '../../../../src/components/atoms/Typography/Typography';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('Typography Component', () => {
  test('REQ-UI-007: Typography renders with correct variant', () => {
    const component = renderWithProviders(
      <Typography variant="h1">Heading Text</Typography>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Heading Text'],
      hasClass: ['text-4xl', 'font-bold']
    });
  });

  test('REQ-UI-008: Typography handles color prop correctly', () => {
    const component = renderWithProviders(
      <Typography color="error">Error Text</Typography>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Error Text'],
      hasClass: ['text-red-600']
    });
  });

  test('REQ-UI-009: Typography handles component prop correctly', () => {
    const component = renderWithProviders(
      <Typography component="h2" data-testid="heading">Heading</Typography>
    );
    
    expect(component.getByTestId('heading').tagName).toBe('H2');
  });
});
