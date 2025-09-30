/**
 * FormControlLabel Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-FORMCONTROLLABEL-001: FormControlLabel renders with label
 * - REQ-FORMCONTROLLABEL-002: FormControlLabel handles control prop
 * - REQ-FORMCONTROLLABEL-003: FormControlLabel applies custom styling
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { FormControlLabel } from '../../../../src/components/atoms/FormControlLabel/FormControlLabel';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('FormControlLabel Component', () => {
  test('REQ-FORMCONTROLLABEL-001: FormControlLabel renders with label', () => {
    const component = renderWithProviders(
      <FormControlLabel label="Form label" />
    );
    
    assertComponentBehavior(component, {
      hasText: ['Form label']
    });
  });

  test('REQ-FORMCONTROLLABEL-002: FormControlLabel handles control prop', () => {
    const component = renderWithProviders(
      <FormControlLabel label="Checkbox label" control={<input type="checkbox" />} />
    );
    
    assertComponentBehavior(component, {
      hasText: ['Checkbox label']
    });
  });

  test('REQ-FORMCONTROLLABEL-003: FormControlLabel applies custom styling', () => {
    const component = renderWithProviders(
      <FormControlLabel label="Custom label" className="custom-label" />
    );
    
    assertComponentBehavior(component, {
      hasText: ['Custom label'],
      hasClass: ['custom-label']
    });
  });
});
