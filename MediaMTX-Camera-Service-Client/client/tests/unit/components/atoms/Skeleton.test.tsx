/**
 * Skeleton Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-SKELETON-001: Skeleton renders with correct animation
 * - REQ-SKELETON-002: Skeleton handles different variants
 * - REQ-SKELETON-003: Skeleton applies custom styling
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { Skeleton } from '../../../../src/components/atoms/Skeleton/Skeleton';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('Skeleton Component', () => {
  test('REQ-SKELETON-001: Skeleton renders with correct animation', () => {
    const component = renderWithProviders(
      <Skeleton />
    );
    
    assertComponentBehavior(component, {
      hasClass: ['animate-pulse', 'bg-gray-200']
    });
  });

  test('REQ-SKELETON-002: Skeleton handles different variants', () => {
    const component = renderWithProviders(
      <Skeleton variant="rectangular" width={200} height={100} />
    );
    
    assertComponentBehavior(component, {
      hasClass: ['animate-pulse', 'bg-gray-200']
    });
  });

  test('REQ-SKELETON-003: Skeleton applies custom styling', () => {
    const component = renderWithProviders(
      <Skeleton className="custom-skeleton" />
    );
    
    assertComponentBehavior(component, {
      hasClass: ['custom-skeleton', 'animate-pulse']
    });
  });
});
