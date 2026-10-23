import { create } from 'zustand';
import { adminLogin, adminLogout, checkAdminSession, type LoginCredentials } from '../services/api';

interface AuthState {
  isAuthenticated: boolean;
  email: string | null;
  // the admin's own name, if they set one
  name: string | null;
  isLoading: boolean;
  error: string | null;
  login: (credentials: LoginCredentials) => Promise<boolean>;
  logout: () => Promise<void>;
  checkSession: () => Promise<void>;
  setName: (name: string) => void;
  clearError: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  isAuthenticated: false,
  email: null,
  name: null,
  isLoading: false,
  error: null,

  login: async (credentials: LoginCredentials) => {
    set({ isLoading: true, error: null });
    try {
      const response = await adminLogin(credentials);
      if (response.success) {
        set({ isAuthenticated: true, email: response.email ?? null, name: response.name || null, isLoading: false });
        return true;
      }
      set({ error: response.message ?? 'Login failed', isLoading: false });
      return false;
    } catch (err) {
      set({ error: err instanceof Error ? err.message : 'Login failed', isLoading: false });
      return false;
    }
  },

  logout: async () => {
    try {
      await adminLogout();
    } finally {
      set({ isAuthenticated: false, email: null, name: null });
    }
  },

  checkSession: async () => {
    set({ isLoading: true });
    try {
      const response = await checkAdminSession();
      set({
        isAuthenticated: response.success,
        email: response.email ?? null,
        name: response.name || null,
        isLoading: false,
      });
    } catch {
      set({ isAuthenticated: false, email: null, name: null, isLoading: false });
    }
  },

  setName: (name) => set({ name: name || null }),

  clearError: () => set({ error: null }),
}));
