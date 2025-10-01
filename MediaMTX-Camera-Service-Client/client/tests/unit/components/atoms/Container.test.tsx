/**
 * Container Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-CONTAINER-001: Container renders with correct maxWidth
 * - REQ-CONTAINER-002: Container handles children correctly
 * - REQ-CONTAINER-003: Container applies correct styling
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { Container } from '../../../../src/components/atoms/Container/Container';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('Container Component', () => {
  test('REQ-CONTAINER-001: Container renders with correct maxWidth', () => {
    const component = renderWithProviders(
      <Container maxWidth="lg">Container content</Container>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Container content'],
      hasClass: ['max-w-4xl']
    });
  });

  test('REQ-CONTAINER-002: Container handles children correctly', () => {
    const component = renderWithProviders(
      <Container>
        <div>Child 1</div>
        <div>Child 2</div>
      </Container>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Child 1', 'Child 2']
    });
  });

  test('REQ-CONTAINER-003: Container applies correct styling', () => {
    const component = renderWithProviders(
      <Container>Styled container</Container>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Styled container'],
      hasClass: ['mx-auto', 'px-4']
    });
  });
});
