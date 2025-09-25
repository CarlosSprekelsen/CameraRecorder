import { create } from 'zustand';
import { AuthState } from '../../types/api';

interface AuthStore extends AuthState {
  setToken: (token: string | null) => void;
  setRole: (role: 'admin' | 'operator' | 'viewer' | null) => void;
  setSessionId: (sessionId: string | null) => void;
  setExpiresAt: (expiresAt: string | null) => void;
  setPermissions: (permissions: string[]) => void;
  setAuthenticated: (authenticated: boolean) => void;
  login: (token: string, role: string, sessionId: string, expiresAt: string, permissions: string[]) => void;
  logout: () => void;
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

export const useAuthStore = create<AuthStore>((set) => ({
  ...initialState,
  
  setToken: (token: string | null) => 
    set((state) => ({ ...state, token })),
  
  setRole: (role: 'admin' | 'operator' | 'viewer' | null) => 
    set((state) => ({ ...state, role })),
  
  setSessionId: (sessionId: string | null) => 
    set((state) => ({ ...state, session_id: sessionId })),
  
  setExpiresAt: (expiresAt: string | null) => 
    set((state) => ({ ...state, expires_at: expiresAt })),
  
  setPermissions: (permissions: string[]) => 
    set((state) => ({ ...state, permissions })),
  
  setAuthenticated: (authenticated: boolean) => 
    set((state) => ({ ...state, isAuthenticated: authenticated })),
  
  login: (token: string, role: string, sessionId: string, expiresAt: string, permissions: string[]) => 
    set((state) => ({
      ...state,
      token,
      role: role as 'admin' | 'operator' | 'viewer',
      session_id: sessionId,
      isAuthenticated: true,
      expires_at: expiresAt,
      permissions,
    })),
  
  logout: () => 
    set((state) => ({
      ...state,
      token: null,
      role: null,
      session_id: null,
      isAuthenticated: false,
      expires_at: null,
      permissions: [],
    })),
  
  reset: () => set(initialState),
}));
