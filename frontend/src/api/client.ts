import axios, { InternalAxiosRequestConfig, AxiosResponse } from 'axios';
import { logApiResponse, logApiError, PerformanceTimer } from '../components/debug/DebugHelper';

// 完全な基本URLを指定
const API_BASE_URL = `${import.meta.env.VITE_API_BASE_URL}`;

// Axiosタイプ拡張
declare module 'axios' {
  export interface InternalAxiosRequestConfig {
    metadata?: {
      startTime?: number;
      timer?: PerformanceTimer;
      [key: string]: any;
    };
  }
  
  export interface AxiosResponse {
    metadata?: {
      duration?: number;
      endpoint?: string;
      [key: string]: any;
    };
  }
}

const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  }
});

// リクエストインターセプター - トークン追加
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }

  // パフォーマンス計測開始
  config.metadata = { 
    ...config.metadata,
    startTime: new Date().getTime(),
    timer: new PerformanceTimer(`API ${config.method?.toUpperCase() ?? 'REQUEST'} ${config.url}`)
  };

  // デバッグ用
  console.log('APIリクエスト:', {
    url: config.url,
    method: config.method,
    headers: config.headers,
    data: config.data
  });

  return config;
});

// レスポンスインターセプター - エラーハンドリングとトークンリフレッシュ
apiClient.interceptors.response.use(
  (response) => {
    // パフォーマンス計測終了
    const { config } = response;
    if (config.metadata?.timer) {
      const duration = config.metadata.timer.stop();
      response.metadata = { 
        ...response.metadata, 
        duration,
        endpoint: config.url
      };
    }

    // API応答をログに記録
    logApiResponse(response.config.url || 'unknown', response.data);
    
    return response;
  },
  async (error) => {
    // パフォーマンス計測終了（エラー時）
    const { config } = error;
    if (config?.metadata?.timer) {
      config.metadata.timer.stop();
    }

    // APIエラーをログに記録
    logApiError(
      error.config?.url || 'unknown', 
      {
        status: error.response?.status,
        data: error.response?.data,
        message: error.message
      }
    );

    const originalRequest = error.config;
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      try {
        // トークンのリフレッシュ処理
        const refreshToken = localStorage.getItem('refreshToken');
        const response = await axios.post(`${API_BASE_URL}/auth/refresh`, {
          refresh_token: refreshToken,
        });
        const { token } = response.data;
        localStorage.setItem('token', token);
        
        // 新しいトークンでリクエストを再試行
        originalRequest.headers.Authorization = `Bearer ${token}`;
        return apiClient(originalRequest);
      } catch (refreshError) {
        // リフレッシュに失敗した場合はログアウト
        localStorage.removeItem('token');
        localStorage.removeItem('refreshToken');
        window.location.href = '/login';
        return Promise.reject(refreshError);
      }
    }
    return Promise.reject(error);
  }
);

export default apiClient; 