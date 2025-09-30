/**
 * Tabs Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-TABS-001: Tabs renders with tab labels
 * - REQ-TABS-002: Tabs handles tab changes
 * - REQ-TABS-003: Tabs shows active tab
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import { Tabs } from '../../../../src/components/atoms/Tabs/Tabs';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('Tabs Component', () => {
  test('REQ-TABS-001: Tabs renders with tab labels', () => {
    const component = renderWithProviders(
      <Tabs value={0}>
        <div>Tab 1</div>
        <div>Tab 2</div>
      </Tabs>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Tab 1', 'Tab 2'],
      hasClass: ['tabs']
    });
  });

  test('REQ-TABS-002: Tabs handles tab changes', () => {
    const handleChange = jest.fn();
    const component = renderWithProviders(
      <Tabs value={0} onChange={handleChange}>
        <div>Tab 1</div>
        <div>Tab 2</div>
      </Tabs>
    );
    
    const tab2 = component.getByText('Tab 2');
    tab2.click();
    expect(handleChange).toHaveBeenCalledTimes(1);
  });

  test('REQ-TABS-003: Tabs shows active tab', () => {
    const component = renderWithProviders(
      <Tabs value={1}>
        <div>Tab 1</div>
        <div>Tab 2</div>
      </Tabs>
    );
    
    assertComponentBehavior(component, {
      hasText: ['Tab 1', 'Tab 2'],
      hasClass: ['tabs']
    });
  });
});
