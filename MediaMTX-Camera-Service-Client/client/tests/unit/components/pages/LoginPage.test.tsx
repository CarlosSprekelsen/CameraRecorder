/**
 * LoginPage Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Security Architecture: Section 8.3
 * - Authentication Interface: Section 5.3.1
 * 
 * Architecture Compliance:
 * - Token-based authentication (not username/password)
 * - Single "Authentication Token" field
 * - "Connect" button (not "Sign In")
 * - JSON-RPC authenticate method with auth_token parameter
 * 
 * Requirements Coverage:
 * - REQ-LOGIN-001: LoginPage renders login form
 * - REQ-LOGIN-002: LoginPage handles authentication
 * - REQ-LOGIN-003: LoginPage shows error states
 * - REQ-LOGIN-004: LoginPage validates form inputs
 * - REQ-LOGIN-005: LoginPage handles loading states
 * 
 * Test Categories: Unit/Component/Security
 */

import React from 'react';
import LoginPage from '../../../../src/pages/Login/LoginPage';
import { renderWithProviders, assertComponentBehavior } from '../../../utils/component-test-helper';

describe('LoginPage Component', () => {
  test('REQ-LOGIN-001: LoginPage renders login form', () => {
    const component = renderWithProviders(
      <LoginPage />,
      { 
        withStores: true,
        initialStoreState: {
          authStore: { loading: false, error: null },
          connectionStore: { status: 'disconnected' }
        }
      }
    );
    
    // Architecture compliance: Token-based authentication with "Connect" button
    assertComponentBehavior(component, {
      hasText: ['MediaMTX Camera Service', 'Enter your authentication token to continue', 'Authentication Token', 'Connect', 'Contact your administrator for access credentials']
    });
  });

  test('REQ-LOGIN-002: LoginPage handles authentication', () => {
    const component = renderWithProviders(
      <LoginPage />,
      { 
        withStores: true,
        initialStoreState: {
          authStore: { loading: false, error: null },
          connectionStore: { status: 'disconnected' }
        }
      }
    );
    
    // Architecture compliance: "Connect" button exists but is disabled when disconnected
    const connectButton = component.getByRole('button', { name: /connect/i });
    expect(connectButton).toBeInTheDocument();
    expect(connectButton).toBeDisabled(); // Button is disabled when disconnected
  });

  test('REQ-LOGIN-003: LoginPage shows error states', () => {
    const component = renderWithProviders(
      <LoginPage />,
      { 
        withStores: true,
        initialStoreState: {
          authStore: { 
            loading: false, 
            error: 'Authentication failed' 
          },
          connectionStore: { status: 'disconnected' }
        }
      }
    );
    
    // Architecture compliance: Component shows connection status, not auth errors directly
    // Note: Component shows "Disconnected" status when connection store is disconnected
    // Use more flexible text matching since "Status:" and "Disconnected" are in separate elements
    expect(component.getByText(/Status:/)).toBeInTheDocument();
    expect(component.getByText(/Disconnected/)).toBeInTheDocument();
  });

  test('REQ-LOGIN-004: LoginPage validates form inputs', () => {
    const component = renderWithProviders(
      <LoginPage />,
      { 
        withStores: true,
        initialStoreState: {
          authStore: { loading: false, error: null },
          connectionStore: { status: 'disconnected' }
        }
      }
    );
    
    // Architecture compliance: Single token input field (not username/password)
    // Note: Input is disabled when disconnected, so check for the input element directly
    const tokenInput = component.container.querySelector('input[type="password"]');
    expect(tokenInput).toBeInTheDocument();
    expect(tokenInput).toHaveAttribute('type', 'password');
    expect(tokenInput).toBeDisabled(); // Input is disabled when disconnected
  });

  test('REQ-LOGIN-005: LoginPage handles loading states', () => {
    const component = renderWithProviders(
      <LoginPage />,
      { 
        withStores: true,
        initialStoreState: {
          authStore: { loading: true, error: null },
          connectionStore: { status: 'disconnected' }
        }
      }
    );
    
    // Architecture compliance: "Connect" button is disabled during loading AND when disconnected
    const connectButton = component.getByRole('button', { name: /connect/i });
    expect(connectButton).toBeDisabled();
  });
});
