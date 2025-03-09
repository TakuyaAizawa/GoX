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
  try {
    console.log('ログインリクエスト:', credentials);
    const response = await apiClient.post('auth/login', credentials);
    console.log('ログインレスポンス:', response.data);
    
    // GoXのAPIはsuccessフィールドを含むレスポンス形式を使用
    if (response.data && response.data.success) {
      return response.data.data;
    }
    
    // 古いAPI形式との互換性のため
    return response.data;
  } catch (error) {
    console.error('ログインエラー:', error);
    throw error;
  }
};

/**
 * ユーザー登録処理
 */
export const register = async (userData: RegisterData): Promise<AuthResponse> => {
  try {
    const response = await apiClient.post('auth/register', userData);
    
    // GoXのAPIはsuccessフィールドを含むレスポンス形式を使用
    if (response.data && response.data.success) {
      return response.data.data;
    }
    
    // 古いAPI形式との互換性のため
    return response.data;
  } catch (error) {
    console.error('登録エラー:', error);
    throw error;
  }
};

/**
 * トークンリフレッシュ処理
 */
export const refreshToken = async (refreshToken: string): Promise<{ token: string }> => {
  try {
    const response = await apiClient.post('auth/refresh', { refresh_token: refreshToken });
    
    // GoXのAPIはsuccessフィールドを含むレスポンス形式を使用
    if (response.data && response.data.success) {
      return response.data.data;
    }
    
    // 古いAPI形式との互換性のため
    return response.data;
  } catch (error) {
    console.error('トークンリフレッシュエラー:', error);
    throw error;
  }
};

/**
 * ログアウト処理
 */
export const logout = async (): Promise<void> => {
  try {
    await apiClient.post('auth/logout');
  } catch (error) {
    console.error('ログアウトエラー:', error);
    // ログアウトエラーは無視する（サーバーが応答しなくても、クライアント側でのログアウトは進める）
  }
}; 