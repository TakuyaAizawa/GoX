import axios, { InternalAxiosRequestConfig, AxiosResponse } from 'axios';
import { logApiResponse, logApiError, PerformanceTimer } from '../components/debug/DebugHelper';

// 基本URLを指定
const API_BASE_URL = `${import.meta.env.VITE_API_BASE_URL}/api/v1`;

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
  if (import.meta.env.DEV) {
    console.log(`API Request: ${config.method?.toUpperCase() ?? 'REQUEST'} ${config.url}`, config);
  }
  
  return config;
}, (error) => {
  return Promise.reject(error);
});

// レスポンスインターセプター - トークンリフレッシュ
apiClient.interceptors.response.use(
  (response) => {
    // パフォーマンス計測終了
    const duration = new Date().getTime() - (response.config.metadata?.startTime ?? 0);
    response.metadata = {
      ...response.metadata,
      duration,
      endpoint: response.config.url
    };
    
    // タイマー停止
    response.config.metadata?.timer?.stop();
    
    // レスポンスをログに出力（デバッグ用）
    if (import.meta.env.DEV) {
      logApiResponse(response.config.url || 'unknown', response.data);
    }
    
    return response;
  },
  async (error) => {
    // タイマー停止
    error.config?.metadata?.timer?.stop();
    
    // エラーをログに出力（デバッグ用）
    if (import.meta.env.DEV) {
      logApiError(error.config?.url || 'unknown', {
        status: error.response?.status,
        data: error.response?.data,
        message: error.message
      });
    }
    
    const originalRequest = error.config;
    
    // 401エラーで、リフレッシュトークンが有効な場合は再試行
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      
      const refreshToken = localStorage.getItem('refreshToken');
      
      if (refreshToken) {
        try {
          // トークンリフレッシュAPIを呼び出す
          const response = await apiClient.post('auth/refresh', {
            refresh_token: refreshToken
          });
          
          let newToken = null;
          
          if (response.data.success && response.data.data && response.data.data.token) {
            // GoX API形式の応答
            newToken = response.data.data.token;
          } else if (response.data.token) {
            // 従来の応答形式
            newToken = response.data.token;
          }
          
          if (newToken) {
            // 新しいトークンを保存
            localStorage.setItem('token', newToken);
            
            // リクエストのヘッダーにトークンを追加
            originalRequest.headers.Authorization = `Bearer ${newToken}`;
            
            // 元のリクエストを再試行
            return axios(originalRequest);
          }
        } catch (refreshError) {
          console.error('トークンリフレッシュ失敗:', refreshError);
          // リフレッシュトークンが無効になった場合はログアウト処理
          localStorage.removeItem('token');
          localStorage.removeItem('refreshToken');
          localStorage.removeItem('user');
          window.location.href = '/login';
        }
      }
    }
    
    return Promise.reject(error);
  }
);

export default apiClient; 