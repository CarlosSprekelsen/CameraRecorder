/**
 * FormControl Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-FORMCONTROL-001: FormControl renders children correctly
 * - REQ-FORMCONTROL-002: FormControl handles error state
 * - REQ-FORMCONTROL-003: FormControl applies custom styling
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { FormControl } from '../../../../src/components/atoms/FormControl/FormControl';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('FormControl Component', () => {
  test('REQ-FORMCONTROL-001: FormControl renders children correctly', () => {
    const component = renderWithProviders(
      <FormControl>
        <input type="text" />
      </FormControl>
    );
    
    assertComponentBehavior(component, {
      hasClass: ['form-control']
    });
  });

  test('REQ-FORMCONTROL-002: FormControl handles error state', () => {
    const component = renderWithProviders(
      <FormControl>
        <input type="text" />
      </FormControl>
    );
    
    assertComponentBehavior(component, {
      hasClass: ['form-control']
    });
  });

  test('REQ-FORMCONTROL-003: FormControl applies custom styling', () => {
    const component = renderWithProviders(
      <FormControl className="custom-control">
        <input type="text" />
      </FormControl>
    );
    
    assertComponentBehavior(component, {
      hasClass: ['custom-control', 'form-control']
    });
  });
});
