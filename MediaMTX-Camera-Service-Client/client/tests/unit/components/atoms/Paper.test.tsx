/**
 * Paper Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-PAPER-001: Paper renders with correct styling
 * - REQ-PAPER-002: Paper handles children correctly
 * - REQ-PAPER-003: Paper applies custom className
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { Paper } from '../../../../src/components/atoms/Paper/Paper';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('Paper Component', () => {
  test('REQ-PAPER-001: Paper renders with correct styling', () => {
    const component = renderWithProviders(
      <Paper>Paper content</Paper>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Paper content'],
      hasClass: ['bg-white', 'shadow', 'rounded']
    });
  });

  test('REQ-PAPER-002: Paper handles children correctly', () => {
    const component = renderWithProviders(
      <Paper>
        <div>Child 1</div>
        <div>Child 2</div>
      </Paper>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Child 1', 'Child 2']
    });
  });

  test('REQ-PAPER-003: Paper applies custom className', () => {
    const component = renderWithProviders(
      <Paper className="custom-paper">Custom paper</Paper>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Custom paper'],
      hasClass: ['custom-paper', 'bg-white']
    });
  });
});
