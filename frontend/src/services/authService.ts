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
    const response = await apiClient.post('/auth/login', credentials);
    console.log('ログインレスポンス:', response.data);
    
    // GoXのAPIはsuccessフィールドを含むレスポンス形式を使用
    if (response.data && response.data.success) {
      return response.data.data;
    } else {
      throw new Error(response.data?.error?.message || 'ログインに失敗しました');
    }
  } catch (error) {
    console.error('ログインエラー詳細:', error);
    throw error;
  }
};

/**
 * ユーザー登録処理
 */
export const register = async (userData: RegisterData): Promise<AuthResponse> => {
  try {
    console.log('登録リクエスト:', userData);
    
    // フォームデータを作成し、必要に応じてデータ形式を調整
    const requestData = {
      username: userData.username,
      email: userData.email,
      password: userData.password,
      display_name: userData.display_name
    };
    
    // 直接APIエンドポイントにPOSTリクエストを送信
    const response = await fetch('http://localhost:8080/api/v1/auth/register', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(requestData)
    });
    
    const data = await response.json();
    console.log('登録レスポンス:', data);
    
    if (data && data.success) {
      return data.data;
    } else {
      throw new Error(data?.error?.message || 'ユーザー登録に失敗しました');
    }
  } catch (error) {
    console.error('登録エラー詳細:', error);
    throw error;
  }
};

/**
 * トークンリフレッシュ処理
 */
export const refreshToken = async (refreshToken: string): Promise<{ token: string }> => {
  try {
    console.log('トークンリフレッシュリクエスト');
    const response = await apiClient.post('/auth/refresh', { refresh_token: refreshToken });
    console.log('リフレッシュレスポンス:', response.data);
    
    if (response.data && response.data.success) {
      return response.data.data;
    } else {
      throw new Error(response.data?.error?.message || 'トークンのリフレッシュに失敗しました');
    }
  } catch (error) {
    console.error('リフレッシュエラー詳細:', error);
    throw error;
  }
};

/**
 * ログアウト処理
 */
export const logout = async (): Promise<void> => {
  try {
    console.log('ログアウトリクエスト');
    await apiClient.post('/auth/logout');
    console.log('ログアウト成功');
  } catch (error) {
    console.error('ログアウトエラー:', error);
    throw error;
  }
}; 