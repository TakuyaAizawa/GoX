import { Navigate, Outlet } from 'react-router-dom';
import { useEffect } from 'react';
import { useAuthStore } from '../../store/authStore';

/**
 * 認証が必要なルートを保護するためのコンポーネント
 * 認証されていない場合はログインページにリダイレクトする
 */
const PrivateRoute = () => {
  const { isAuthenticated, setUser, setTokens } = useAuthStore();
  
  // コンポーネントマウント時にローカルストレージをチェック
  useEffect(() => {
    const token = localStorage.getItem('token');
    const refreshToken = localStorage.getItem('refreshToken');
    const userJson = localStorage.getItem('user');
    
    // トークンとユーザー情報があれば認証状態を復元
    if (token && refreshToken && userJson) {
      try {
        const user = JSON.parse(userJson);
        setUser(user);
        setTokens(token, refreshToken);
      } catch (e) {
        console.error('ユーザー情報の解析に失敗しました:', e);
        // 不正なユーザー情報はクリア
        localStorage.removeItem('user');
        localStorage.removeItem('token');
        localStorage.removeItem('refreshToken');
      }
    }
  }, [setUser, setTokens]);
  
  // 認証状態を確認
  const hasToken = !!localStorage.getItem('token');
  
  // 認証されていない場合はログインページにリダイレクト
  if (!isAuthenticated && !hasToken) {
    return <Navigate to="/login" replace />;
  }
  
  // 認証されている場合は子コンポーネントをレンダリング
  return <Outlet />;
};

export default PrivateRoute; 