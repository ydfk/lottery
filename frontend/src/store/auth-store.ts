import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { login as apiLogin, logout as apiLogout } from '../lib/api/methods/user';

interface LoginRequest {
  username: string;
  password: string;
}

interface AuthState {
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  login: (data: LoginRequest) => Promise<void>;
  logout: () => void;
  clearError: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: localStorage.getItem('auth_token'),
      isAuthenticated: !!localStorage.getItem('auth_token'),
      isLoading: false,
      error: null,
      
      login: async (data: LoginRequest) => {
        set({ isLoading: true, error: null });
        try {
          const response = await apiLogin(data);
          localStorage.setItem('auth_token', response.token);
          set({ 
            token: response.token,
            isAuthenticated: true, 
            isLoading: false,
            error: null,
          });
        } catch (error) {
          let errorMessage = 'Login failed';
          if (error instanceof Error) {
            errorMessage = error.message;
          }
          set({ isLoading: false, error: errorMessage, isAuthenticated: false });
        }
      },
      
      logout: () => {
        apiLogout();
        set({ token: null, isAuthenticated: false });
      },
      
      clearError: () => {
        set({ error: null });
      },
    }),
    {
      name: 'auth-storage', // name of the item in localStorage
      partialize: (state) => ({ token: state.token }), // only persist token
    }
  )
);