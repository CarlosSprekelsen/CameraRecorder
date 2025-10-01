/**
 * AboutPage Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Atomic Design Pattern: ADR-003
 * 
 * Requirements Coverage:
 * - REQ-ABOUT-001: AboutPage renders information content
 * - REQ-ABOUT-002: AboutPage displays version information
 * - REQ-ABOUT-003: AboutPage shows system details
 * 
 * Test Categories: Unit/Component
 */

import React from 'react';
import AboutPage from '../../../../src/pages/About/AboutPage';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('AboutPage Component', () => {
  test('REQ-ABOUT-001: AboutPage renders information content', () => {
    const component = renderWithProviders(
      <AboutPage />
    );
    
    assertComponentBehavior(component, {
      hasText: ['Server Information']
    });
  });

  test('REQ-ABOUT-002: AboutPage displays version information', () => {
    const component = renderWithProviders(
      <AboutPage />,
      { 
        withStores: true,
        initialStoreState: {
          serverStore: { 
            serverInfo: { 
              version: '1.0.0',
              build_date: '2025-01-25'
            }
          }
        }
      }
    );
    
    assertComponentBehavior(component, {
      hasText: ['Version:', 'Build Date:']
    });
  });

  test('REQ-ABOUT-003: AboutPage shows system details', () => {
    const component = renderWithProviders(
      <AboutPage />,
      { 
        withStores: true,
        initialStoreState: {
          serverStore: { 
            serverInfo: { 
              version: '1.0.0',
              system_info: 'MediaMTX Camera Service Client'
            }
          }
        }
      }
    );
    
    assertComponentBehavior(component, {
      hasText: ['System Status']
    });
  });
});
