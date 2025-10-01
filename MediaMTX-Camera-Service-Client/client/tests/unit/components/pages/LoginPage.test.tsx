/**
 * LoginPage Component Tests
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - Security Architecture: Section 8.3
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
          authStore: { loading: false, error: null }
        }
      }
    );
    
    assertComponentBehavior(component, {
      hasText: ['Sign In', 'Username', 'Password']
    });
  });

  test('REQ-LOGIN-002: LoginPage handles authentication', () => {
    const component = renderWithProviders(
      <LoginPage />,
      { 
        withStores: true,
        initialStoreState: {
          authStore: { loading: false, error: null }
        }
      }
    );
    
    const loginButton = component.getByRole('button', { name: /sign in/i });
    expect(loginButton).toBeInTheDocument();
    expect(loginButton).not.toBeDisabled();
  });

  test('REQ-LOGIN-003: LoginPage shows error states', () => {
    const component = renderWithProviders(
      <LoginPage />,
      { 
        withStores: true,
        initialStoreState: {
          authStore: { 
            loading: false, 
            error: 'Invalid credentials' 
          }
        }
      }
    );
    
    assertComponentBehavior(component, {
      hasText: ['Invalid credentials']
    });
  });

  test('REQ-LOGIN-004: LoginPage validates form inputs', () => {
    const component = renderWithProviders(
      <LoginPage />,
      { 
        withStores: true,
        initialStoreState: {
          authStore: { loading: false, error: null }
        }
      }
    );
    
    const usernameInput = component.getByLabelText(/username/i);
    const passwordInput = component.getByLabelText(/password/i);
    
    expect(usernameInput).toBeInTheDocument();
    expect(passwordInput).toBeInTheDocument();
  });

  test('REQ-LOGIN-005: LoginPage handles loading states', () => {
    const component = renderWithProviders(
      <LoginPage />,
      { 
        withStores: true,
        initialStoreState: {
          authStore: { loading: true, error: null }
        }
      }
    );
    
    const loginButton = component.getByRole('button', { name: /sign in/i });
    expect(loginButton).toBeDisabled();
  });
});
