/**
 * Card Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-UI-010: Card renders with correct styling
 * - REQ-UI-011: Card handles children correctly
 * - REQ-UI-012: Card handles custom className
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { Card } from '../../../../src/components/atoms/Card/Card';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('Card Component', () => {
  test('REQ-UI-010: Card renders with correct styling', () => {
    const component = renderWithProviders(
      <Card>Card Content</Card>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Card Content'],
      hasClass: ['bg-white', 'rounded-lg', 'shadow']
    });
  });

  test('REQ-UI-011: Card handles children correctly', () => {
    const component = renderWithProviders(
      <Card>
        <div>Child 1</div>
        <div>Child 2</div>
      </Card>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Child 1', 'Child 2']
    });
  });

  test('REQ-UI-012: Card handles custom className', () => {
    const component = renderWithProviders(
      <Card className="custom-card">Custom Card</Card>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Custom Card'],
      hasClass: ['custom-card', 'bg-white']
    });
  });
});
