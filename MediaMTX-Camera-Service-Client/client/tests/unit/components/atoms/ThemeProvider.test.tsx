/**
 * ThemeProvider Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-THEMEPROVIDER-001: ThemeProvider renders children
 * - REQ-THEMEPROVIDER-002: ThemeProvider applies theme context
 * - REQ-THEMEPROVIDER-003: ThemeProvider handles custom theme
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { ThemeProvider } from '../../../../src/components/atoms/ThemeProvider/ThemeProvider';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('ThemeProvider Component', () => {
  test('REQ-THEMEPROVIDER-001: ThemeProvider renders children', () => {
    const component = renderWithProviders(
      <ThemeProvider>
        <div>Theme content</div>
      </ThemeProvider>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Theme content']
    });
  });

  test('REQ-THEMEPROVIDER-002: ThemeProvider applies theme context', () => {
    const component = renderWithProviders(
      <ThemeProvider>
        <div>Themed content</div>
      </ThemeProvider>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Themed content']
    });
  });

  test('REQ-THEMEPROVIDER-003: ThemeProvider handles custom theme', () => {
    const customTheme = { primary: '#ff0000' };
    const component = renderWithProviders(
      <ThemeProvider theme={customTheme}>
        <div>Custom themed content</div>
      </ThemeProvider>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Custom themed content']
    });
  });
});
