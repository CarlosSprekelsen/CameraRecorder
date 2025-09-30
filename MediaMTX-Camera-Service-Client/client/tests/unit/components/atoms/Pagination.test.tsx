/**
 * Pagination Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-PAGINATION-001: Pagination renders with correct page count
 * - REQ-PAGINATION-002: Pagination handles page changes
 * - REQ-PAGINATION-003: Pagination shows current page
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { Pagination } from '../../../../src/components/atoms/Pagination/Pagination';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('Pagination Component', () => {
  test('REQ-PAGINATION-001: Pagination renders with correct page count', () => {
    const component = renderWithProviders(
      <Pagination count={10} page={1} />
    );
    
    assertComponentBehavior(component, {
      hasClass: ['pagination']
    });
  });

  test('REQ-PAGINATION-002: Pagination handles page changes', () => {
    const handleChange = jest.fn();
    const component = renderWithProviders(
      <Pagination count={5} page={1} onChange={handleChange} />
    );
    
    const nextButton = component.getByRole('button', { name: /next/i });
    nextButton.click();
    expect(handleChange).toHaveBeenCalledTimes(1);
  });

  test('REQ-PAGINATION-003: Pagination shows current page', () => {
    const component = renderWithProviders(
      <Pagination count={10} page={3} />
    );
    
    assertComponentBehavior(component, {
      hasText: ['3'],
      hasClass: ['pagination']
    });
  });
});
