import axios from 'axios';

// 完全な基本URLを指定
const API_BASE_URL = 'http://localhost:8080/api/v1';

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
    // デバッグ用
    console.log('APIレスポンス:', {
      status: response.status,
      data: response.data
    });
    return response;
  },
  async (error) => {
    // デバッグ用
    console.error('APIエラー:', {
      status: error.response?.status,
      data: error.response?.data,
      message: error.message
    });

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