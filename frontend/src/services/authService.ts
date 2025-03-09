import apiClient from '../api/client';

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface RegisterData {
  username: string;
  email: string;
  password: string;
  display_name: string;
}

export interface AuthResponse {
  user: {
    id: string;
    username: string;
    display_name: string;
    email: string;
    avatar_url?: string;
    banner_url?: string;
    bio?: string;
    created_at: string;
  };
  token: string;
  refresh_token: string;
}

/**
 * ログイン処理
 */
export const login = async (credentials: LoginCredentials): Promise<AuthResponse> => {
  const response = await apiClient.post('/auth/login', credentials);
  return response.data;
};

/**
 * ユーザー登録処理
 */
export const register = async (userData: RegisterData): Promise<AuthResponse> => {
  const response = await apiClient.post('/auth/register', userData);
  return response.data;
};

/**
 * トークンリフレッシュ処理
 */
export const refreshToken = async (refreshToken: string): Promise<{ token: string }> => {
  const response = await apiClient.post('/auth/refresh', { refresh_token: refreshToken });
  return response.data;
};

/**
 * ログアウト処理
 */
export const logout = async (): Promise<void> => {
  await apiClient.post('/auth/logout');
}; 