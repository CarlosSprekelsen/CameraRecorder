/**
 * Box Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-UI-004: Box renders with correct styling
 * - REQ-UI-005: Box handles custom component prop
 * - REQ-UI-006: Box passes through children correctly
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { Box } from '../../../../src/components/atoms/Box/Box';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('Box Component', () => {
  test('REQ-UI-004: Box renders with correct styling', () => {
    const component = renderWithProviders(
      <Box className="test-box" sx={{ padding: 16 }}>Test Content</Box>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Test Content'],
      hasClass: ['test-box']
    });
  });

  test('REQ-UI-005: Box handles custom component prop', () => {
    const component = renderWithProviders(
      <Box component="section" data-testid="section-box">Section Content</Box>
    );
    
    expect(component.getByTestId('section-box').tagName).toBe('SECTION');
  });

  test('REQ-UI-006: Box passes through children correctly', () => {
    const component = renderWithProviders(
      <Box>
        <span>Child 1</span>
        <span>Child 2</span>
      </Box>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Child 1', 'Child 2']
    });
  });
});
