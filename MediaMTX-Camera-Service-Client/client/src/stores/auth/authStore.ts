import { create } from 'zustand';
import { AuthState } from '../../types/api';
import { AuthService } from '../../services/auth/AuthService';

interface AuthStore extends AuthState {
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
  authenticate: (token: string) => Promise<void>;
  refreshToken: () => Promise<void>;
  
  // Reset
  reset: () => void;
}

const initialState: AuthState = {
  token: null,
  role: null,
  session_id: null,
  isAuthenticated: false,
  expires_at: null,
  permissions: [],
};

export const useAuthStore = create<AuthStore>((set) => {
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
      if (!authService) throw new Error('Auth service not initialized');
      set({ loading: true, error: null });
      try {
        const result = await authService.authenticate(token);
        set({ loading: false });
        return result;
      } catch (error) {
        set({ loading: false, error: error instanceof Error ? error.message : 'Authentication failed' });
        throw error;
      }
    },

    refreshToken: async () => {
      if (!authService) throw new Error('Auth service not initialized');
      set({ loading: true, error: null });
      try {
        const result = await authService.refreshToken();
        set({ loading: false });
        return result;
      } catch (error) {
        set({ loading: false, error: error instanceof Error ? error.message : 'Token refresh failed' });
        throw error;
      }
    },

    reset: () => set(initialState),
  };
});
