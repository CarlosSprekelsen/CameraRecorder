/**
 * Grid Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-GRID-001: Grid renders with correct container
 * - REQ-GRID-002: Grid handles item sizing
 * - REQ-GRID-003: Grid handles spacing correctly
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { Grid } from '../../../../src/components/atoms/Grid/Grid';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('Grid Component', () => {
  test('REQ-GRID-001: Grid renders with correct container', () => {
    const component = renderWithProviders(
      <Grid container>Grid container content</Grid>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Grid container content'],
      hasClass: ['grid', 'grid-cols-12']
    });
  });

  test('REQ-GRID-002: Grid handles item sizing', () => {
    const component = renderWithProviders(
      <Grid item xs={6}>Grid item content</Grid>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Grid item content'],
      hasClass: ['col-span-6']
    });
  });

  test('REQ-GRID-003: Grid handles spacing correctly', () => {
    const component = renderWithProviders(
      <Grid container spacing={2}>Spaced grid content</Grid>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Spaced grid content'],
      hasClass: ['gap-4']
    });
  });
});
