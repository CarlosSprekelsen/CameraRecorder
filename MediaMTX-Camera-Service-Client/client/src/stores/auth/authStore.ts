import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { AuthenticateResult } from '../../types/api';
import { AuthService } from '../../services/auth/AuthService';

export interface AuthStoreState {
  // API state (from server)
  token: string | null;
  role: 'admin' | 'operator' | 'viewer' | null;
  session_id: string | null;
  isAuthenticated: boolean;
  expires_at: string | null;
  permissions: string[];
  
  // UI state (client-side)
  loading: boolean;
  error: string | null;
}

export interface AuthActions {
  // Service injection
  setAuthService: (service: AuthService) => void;
  
  // State setters
  setToken: (token: string | null) => void;
  setRole: (role: 'admin' | 'operator' | 'viewer' | null) => void;
  setSessionId: (sessionId: string | null) => void;
  setExpiresAt: (expiresAt: string | null) => void;
  setPermissions: (permissions: string[]) => void;
  setAuthenticated: (authenticated: boolean) => void;
  
  // Actions that call services
  login: (
    token: string,
    role: string,
    sessionId: string,
    expiresAt: string,
    permissions: string[],
  ) => void;
  logout: () => void;
  authenticate: (token: string) => Promise<AuthenticateResult>;
  
  // Reset
  reset: () => void;
}

const initialState: AuthStoreState = {
  // API state (from server)
  token: null,
  role: null,
  session_id: null,
  isAuthenticated: false,
  expires_at: null,
  permissions: [],
  
  // UI state (client-side)
  loading: false,
  error: null,
};

export const useAuthStore = create<AuthStoreState & AuthActions>()(
  devtools(
    (set) => {
  let authService: AuthService | null = null;

  return {
    ...initialState,

    // Service injection
    setAuthService: (service: AuthService) => {
      authService = service;
    },

    // State setters
    setToken: (token: string | null) => set((state) => ({ ...state, token })),

    setRole: (role: 'admin' | 'operator' | 'viewer' | null) => set((state) => ({ ...state, role })),

    setSessionId: (sessionId: string | null) => set((state) => ({ ...state, session_id: sessionId })),

    setExpiresAt: (expiresAt: string | null) => set((state) => ({ ...state, expires_at: expiresAt })),

    setPermissions: (permissions: string[]) => set((state) => ({ ...state, permissions })),

    setAuthenticated: (authenticated: boolean) =>
      set((state) => ({ ...state, isAuthenticated: authenticated })),

    // Actions that call services
    login: (
      token: string,
      role: string,
      sessionId: string,
      expiresAt: string,
      permissions: string[],
    ) =>
      set((state) => ({
        ...state,
        token,
        role: role as 'admin' | 'operator' | 'viewer',
        session_id: sessionId,
        isAuthenticated: true,
        expires_at: expiresAt,
        permissions,
      })),

    logout: () => {
      if (authService) {
        authService.logout();
      }
      set((state) => ({
        ...state,
        token: null,
        role: null,
        session_id: null,
        isAuthenticated: false,
        expires_at: null,
        permissions: [],
      }));
    },

    authenticate: async (token: string) => {
      // Synchronous guard - graceful error handling per ADR-002
      if (!authService) {
        set({ error: 'Auth service not initialized', loading: false });
        return undefined;
      }

      set({ loading: true, error: null });
      try {
        const result = await authService.authenticate(token);
        set({ loading: false });
        return result;
      } catch (error) {
        set({ loading: false, error: error instanceof Error ? error.message : 'Authentication failed' });
        // No re-throw - graceful degradation per ADR-002
        return undefined;
      }
    },


    reset: () => set(initialState),
  };
},
{
  name: 'auth-store',
},
),
);
