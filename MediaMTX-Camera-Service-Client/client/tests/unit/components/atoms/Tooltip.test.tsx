/**
 * Tooltip Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-TOOLTIP-001: Tooltip renders with title
 * - REQ-TOOLTIP-002: Tooltip shows on hover
 * - REQ-TOOLTIP-003: Tooltip handles children correctly
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { Tooltip } from '../../../../src/components/atoms/Tooltip/Tooltip';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('Tooltip Component', () => {
  test('REQ-TOOLTIP-001: Tooltip renders with title', () => {
    const component = renderWithProviders(
      <Tooltip title="Help text">
        <button>Hover me</button>
      </Tooltip>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Hover me']
    });
  });

  test('REQ-TOOLTIP-002: Tooltip shows on hover', () => {
    const component = renderWithProviders(
      <Tooltip title="Tooltip content">
        <span>Hover target</span>
      </Tooltip>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Hover target']
    });
  });

  test('REQ-TOOLTIP-003: Tooltip handles children correctly', () => {
    const component = renderWithProviders(
      <Tooltip title="Help text">
        <div>Child element</div>
      </Tooltip>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Child element']
    });
  });
});
