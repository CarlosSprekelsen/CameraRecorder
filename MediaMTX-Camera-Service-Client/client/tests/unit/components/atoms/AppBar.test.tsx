/**
 * AppBar Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-APPBAR-001: AppBar renders with correct styling
 * - REQ-APPBAR-002: AppBar handles children correctly
 * - REQ-APPBAR-003: AppBar applies custom className
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { AppBar } from '../../../../src/components/atoms/AppBar/AppBar';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('AppBar Component', () => {
  test('REQ-APPBAR-001: AppBar renders with correct styling', () => {
    const component = renderWithProviders(
      <AppBar>Navigation header</AppBar>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Navigation header'],
      hasClass: ['bg-blue-600', 'text-white']
    });
  });

  test('REQ-APPBAR-002: AppBar handles children correctly', () => {
    const component = renderWithProviders(
      <AppBar>
        <div>Logo</div>
        <div>Navigation</div>
      </AppBar>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Logo', 'Navigation']
    });
  });

  test('REQ-APPBAR-003: AppBar applies custom className', () => {
    const component = renderWithProviders(
      <AppBar className="custom-appbar">Custom header</AppBar>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Custom header'],
      hasClass: ['custom-appbar', 'bg-blue-600']
    });
  });
});
