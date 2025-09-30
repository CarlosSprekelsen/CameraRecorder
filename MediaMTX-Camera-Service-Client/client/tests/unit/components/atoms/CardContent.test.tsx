/**
 * CardContent Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-CARDCONTENT-001: CardContent renders children correctly
 * - REQ-CARDCONTENT-002: CardContent handles custom className
 * - REQ-CARDCONTENT-003: CardContent applies correct styling
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { CardContent } from '../../../../src/components/atoms/CardContent/CardContent';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('CardContent Component', () => {
  test('REQ-CARDCONTENT-001: CardContent renders children correctly', () => {
    const component = renderWithProviders(
      <CardContent>Card content text</CardContent>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Card content text']
    });
  });

  test('REQ-CARDCONTENT-002: CardContent handles custom className', () => {
    const component = renderWithProviders(
      <CardContent className="custom-content">Custom content</CardContent>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Custom content'],
      hasClass: ['custom-content']
    });
  });

  test('REQ-CARDCONTENT-003: CardContent applies correct styling', () => {
    const component = renderWithProviders(
      <CardContent>Styled content</CardContent>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Styled content'],
      hasClass: ['p-6']
    });
  });
});
