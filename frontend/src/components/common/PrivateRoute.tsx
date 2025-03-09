import { Navigate, Outlet } from 'react-router-dom';
import { useAuthStore } from '../../store/authStore';

/**
 * 認証が必要なルートを保護するためのコンポーネント
 * 認証されていない場合はログインページにリダイレクトする
 */
const PrivateRoute = () => {
  const { isAuthenticated } = useAuthStore();
  
  // 認証されていない場合はログインページにリダイレクト
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  
  // 認証されている場合は子コンポーネントをレンダリング
  return <Outlet />;
};

export default PrivateRoute; 