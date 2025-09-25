/**
 * Auth Store - Architecture Compliant (<200 lines)
 * 
 * This store provides a thin wrapper around AuthService
 * following the modular store pattern established in connection/
 * 
 * Responsibilities:
 * - Authentication state management
 * - User session tracking
 * - Authentication operations
 * 
 * Architecture Compliance:
 * - Single responsibility (authentication only)
 * - Uses service layer abstraction
 * - Provides predictable state interface for components
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { logger } from '../services/loggerService';
import { authService } from '../services/authService';
import type { AuthState } from '../services/authService';

// State interface
interface AuthStoreState {
  // Authentication state
  isAuthenticated: boolean;
  role: 'admin' | 'operator' | 'viewer' | null;
  permissions: string[];
  sessionId: string | null;
  token: string | null;
  expiresAt: Date | null;
  
  // Loading states
  isLoading: boolean;
  isAuthenticating: boolean;
  
  // Error state
  error: string | null;
}

// Actions interface
interface AuthStoreActions {
  // Authentication operations
  login: (credentials: { username: string; password: string }) => Promise<boolean>;
  logout: () => Promise<void>;
  refreshToken: () => Promise<boolean>;
  
  // State management
  getAuthState: () => AuthState;
  updateAuthState: () => void;
  
  // Error handling
  clearError: () => void;
  setError: (error: string) => void;
}

// Combined store type
type AuthStore = AuthStoreState & AuthStoreActions;

// Initial state
const initialState: AuthStoreState = {
  isAuthenticated: false,
  role: null,
  permissions: [],
  sessionId: null,
  token: null,
  expiresAt: null,
  isLoading: false,
  isAuthenticating: false,
  error: null,
};

// Create store
export const useAuthStore = create<AuthStore>()(
  devtools(
    (set, get) => ({
      ...initialState,
      
      // Authentication operations
      login: async (credentials: { username: string; password: string }) => {
        set({ isAuthenticating: true, error: null });
        try {
          // Use existing AuthService
          const result = await authService.authenticate(credentials);
          if (result.success) {
            const authState = authService.getAuthState();
            set({
              isAuthenticated: authState.isAuthenticated,
              role: authState.role,
              permissions: authState.permissions,
              sessionId: authState.sessionId,
              token: authState.token,
              expiresAt: authState.expiresAt,
              isAuthenticating: false,
            });
            logger.info('User authenticated successfully', undefined, 'authStore');
            return true;
          } else {
            set({ 
              error: result.error || 'Authentication failed', 
              isAuthenticating: false 
            });
            return false;
          }
        } catch (error: any) {
          const errorMessage = error instanceof Error ? error.message : 'Authentication failed';
          set({ 
            error: errorMessage, 
            isAuthenticating: false 
          });
          return false;
        }
      },
      
      logout: async () => {
        set({ isLoading: true, error: null });
        try {
          // Use existing AuthService
          await authService.clearAuth();
          set({
            isAuthenticated: false,
            role: null,
            permissions: [],
            sessionId: null,
            token: null,
            expiresAt: null,
            isLoading: false,
          });
          logger.info('User logged out successfully', undefined, 'authStore');
        } catch (error: any) {
          const errorMessage = error instanceof Error ? error.message : 'Logout failed';
          set({ 
            error: errorMessage, 
            isLoading: false 
          });
        }
      },
      
      refreshToken: async () => {
        set({ isLoading: true, error: null });
        try {
          // Use existing AuthService
          const result = await authService.refreshToken();
          if (result.success) {
            const authState = authService.getAuthState();
            set({
              isAuthenticated: authState.isAuthenticated,
              role: authState.role,
              permissions: authState.permissions,
              sessionId: authState.sessionId,
              token: authState.token,
              expiresAt: authState.expiresAt,
              isLoading: false,
            });
            return true;
          } else {
            set({ 
              error: result.error || 'Token refresh failed', 
              isLoading: false 
            });
            return false;
          }
        } catch (error: any) {
          const errorMessage = error instanceof Error ? error.message : 'Token refresh failed';
          set({ 
            error: errorMessage, 
            isLoading: false 
          });
          return false;
        }
      },
      
      // State management
      getAuthState: () => {
        return authService.getAuthState();
      },
      
      updateAuthState: () => {
        const authState = authService.getAuthState();
        set({
          isAuthenticated: authState.isAuthenticated,
          role: authState.role,
          permissions: authState.permissions,
          sessionId: authState.sessionId,
          token: authState.token,
          expiresAt: authState.expiresAt,
        });
      },
      
      // Error handling
      clearError: () => set({ error: null }),
      setError: (error: string) => set({ error }),
    }),
    {
      name: 'auth-store',
    }
  )
);

// Export types for components
export type { AuthStoreState, AuthStoreActions };
