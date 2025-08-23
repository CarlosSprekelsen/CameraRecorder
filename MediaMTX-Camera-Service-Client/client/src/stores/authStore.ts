/**
 * Authentication state management store
 * Handles user authentication and role-based access control
 * Aligned with server authentication API
 * 
 * Server API Reference: ../mediamtx-camera-service/docs/api/json-rpc-methods.md
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { authService } from '../services/authService';

/**
 * User information from authentication
 */
export interface User {
  role: 'viewer' | 'operator' | 'admin';
  user_id?: string;
  permissions?: string[];
  expires_at?: string;
  session_id?: string;
}

/**
 * Authentication response from server
 */
export interface AuthResponse {
  authenticated: boolean;
  role: string;
  permissions: string[];
  expires_at: string;
  session_id: string;
}

/**
 * Auth store state interface
 */
interface AuthState {
  // Authentication state
  isAuthenticated: boolean;
  user: User | null;
  
  // Loading states
  isLoading: boolean;
  
  // Error state
  error: string | null;
  
  // Token management
  token: string | null;
  tokenExpiry: Date | null;
}

/**
 * Auth store actions interface
 */
interface AuthActions {
  // Authentication operations
  login: (token: string) => Promise<void>;
  logout: () => void;
  
  // State management
  setError: (error: string | null) => void;
  clearError: () => void;
  setLoading: (loading: boolean) => void;
  
  // Permission checking
  hasPermission: (permission: string) => boolean;
  hasRole: (role: string) => boolean;
  canDeleteFiles: () => boolean;
  canManageSystem: () => boolean;
  canControlCameras: () => boolean;
}

/**
 * Auth store type
 */
type AuthStore = AuthState & AuthActions;

/**
 * Create auth store
 */
export const useAuthStore = create<AuthStore>()(
  devtools(
    (set, get) => ({
      // Initial state
      isAuthenticated: false,
      user: null,
      isLoading: false,
      error: null,
      token: null,
      tokenExpiry: null,

      // Authentication operations
      login: async (token: string) => {
        set({ isLoading: true, error: null });

        try {
          // Store token
          authService.setToken(token);
          
          // Authenticate with server
          const response = await authService.authenticate(token);
          
          if (!response.authenticated) {
            throw new Error('Authentication failed');
          }

          const user: User = {
            role: response.role as 'viewer' | 'operator' | 'admin',
            user_id: response.user_id,
            permissions: response.permissions,
            expires_at: response.expires_at,
            session_id: response.session_id,
          };

          const tokenExpiry = response.expires_at ? new Date(response.expires_at) : null;

          set({
            isAuthenticated: true,
            user,
            token,
            tokenExpiry,
            isLoading: false,
            error: null,
          });

        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : 'Authentication failed';
          set({
            isAuthenticated: false,
            user: null,
            token: null,
            tokenExpiry: null,
            isLoading: false,
            error: errorMessage,
          });
          
          // Clear invalid token
          authService.clearToken();
          throw error;
        }
      },

      logout: () => {
        authService.clearToken();
        set({
          isAuthenticated: false,
          user: null,
          token: null,
          tokenExpiry: null,
          error: null,
        });
      },

      // State management
      setError: (error: string | null) => {
        set({ error });
      },

      clearError: () => {
        set({ error: null });
      },

      setLoading: (loading: boolean) => {
        set({ isLoading: loading });
      },

      // Permission checking
      hasPermission: (permission: string) => {
        const { user } = get();
        if (!user || !user.permissions) return false;
        return user.permissions.includes(permission);
      },

      hasRole: (role: string) => {
        const { user } = get();
        if (!user) return false;
        return user.role === role;
      },

      canDeleteFiles: () => {
        const { user } = get();
        if (!user) return false;
        return user.role === 'admin' || user.role === 'operator';
      },

      canManageSystem: () => {
        const { user } = get();
        if (!user) return false;
        return user.role === 'admin';
      },

      canControlCameras: () => {
        const { user } = get();
        if (!user) return false;
        return user.role === 'admin' || user.role === 'operator';
      },
    }),
    {
      name: 'auth-store',
    }
  )
);
