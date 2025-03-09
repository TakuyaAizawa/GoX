import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import * as authService from '../services/authService';

export interface User {
  id: string;
  username: string;
  display_name: string;
  email: string;
  avatar_url?: string;
  banner_url?: string;
  bio?: string;
  created_at: string;
}

interface AuthState {
  user: User | null;
  token: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  loading: boolean;
  error: string | null;
  
  // アクション
  login: (credentials: authService.LoginCredentials) => Promise<void>;
  register: (data: authService.RegisterData) => Promise<void>;
  logout: () => Promise<void>;
  refreshAuthToken: () => Promise<boolean>;
  clearError: () => void;
  updateUser: (userData: Partial<User>) => void;
  
  // 追加の直接アクセス関数
  setUser: (user: User | null) => void;
  setTokens: (token: string, refreshToken: string) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      token: null,
      refreshToken: null,
      isAuthenticated: false,
      loading: false,
      error: null,
      
      login: async (credentials) => {
        set({ loading: true, error: null });
        try {
          const response = await authService.login(credentials);
          set({
            user: response.user,
            token: response.token,
            refreshToken: response.refresh_token,
            isAuthenticated: true,
            loading: false,
          });
        } catch (error) {
          console.error('Login error:', error);
          set({
            error: error instanceof Error ? error.message : '認証に失敗しました',
            loading: false,
          });
        }
      },
      
      register: async (data) => {
        set({ loading: true, error: null });
        try {
          const response = await authService.register(data);
          set({
            user: response.user,
            token: response.token,
            refreshToken: response.refresh_token,
            isAuthenticated: true,
            loading: false,
          });
        } catch (error) {
          console.error('Register error:', error);
          set({
            error: error instanceof Error ? error.message : 'ユーザー登録に失敗しました',
            loading: false,
          });
        }
      },
      
      logout: async () => {
        set({ loading: true });
        try {
          await authService.logout();
        } catch (error) {
          console.error('Logout error:', error);
        } finally {
          set({
            user: null,
            token: null,
            refreshToken: null,
            isAuthenticated: false,
            loading: false,
          });
        }
      },
      
      refreshAuthToken: async () => {
        const currentRefreshToken = get().refreshToken;
        if (!currentRefreshToken) return false;
        
        try {
          const response = await authService.refreshToken(currentRefreshToken);
          set({
            token: response.token,
          });
          return true;
        } catch (error) {
          console.error('Token refresh error:', error);
          set({
            user: null,
            token: null,
            refreshToken: null,
            isAuthenticated: false,
          });
          return false;
        }
      },
      
      clearError: () => set({ error: null }),
      
      updateUser: (userData) => {
        const currentUser = get().user;
        if (currentUser) {
          set({
            user: { ...currentUser, ...userData },
          });
        }
      },
      
      // 追加の直接アクセス関数
      setUser: (user) => {
        set({
          user,
          isAuthenticated: !!user
        });
      },
      
      setTokens: (token, refreshToken) => {
        set({
          token,
          refreshToken,
          isAuthenticated: true
        });
      },
      
      setLoading: (loading) => {
        set({ loading });
      },
      
      setError: (error) => {
        set({ error });
      }
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        user: state.user,
        token: state.token,
        refreshToken: state.refreshToken,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
); 