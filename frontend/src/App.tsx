import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import LoginPage from './pages/LoginPage';
import RegisterPage from './pages/RegisterPage';
import HomePage from './pages/HomePage';
import PostPage from './pages/PostPage';
import ProfilePage from './pages/ProfilePage';
import NotificationsPage from './pages/NotificationsPage';
import PrivateRoute from './components/common/PrivateRoute';
import ConsoleLogger from './components/debug/ConsoleLogger';
import { WebSocketProvider } from './context/WebSocketContext';
import NotificationListener from './components/notification/NotificationListener';
import { Toaster } from 'react-hot-toast';

function App() {
  // 開発環境かどうかを判定
  const isDevelopment = import.meta.env.DEV;
  
  return (
    <WebSocketProvider>
      <Router>
        <Routes>
          {/* 認証不要のルート */}
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          
          {/* 認証が必要なルート */}
          <Route element={<PrivateRoute />}>
            <Route path="/" element={<HomePage />} />
            <Route path="/post/:postId" element={<PostPage />} />
            <Route path="/profile/:username" element={<ProfilePage />} />
            <Route path="/notifications" element={<NotificationsPage />} />
          </Route>
          
          {/* 404ページ */}
          <Route path="*" element={<div className="p-4 text-gray-900 dark:text-white">ページが見つかりません</div>} />
        </Routes>
        
        {/* 通知リスナー（認証済みユーザーのみ） */}
        <NotificationListener />
        
        {/* コンソールロガー（開発環境でのみ表示） */}
        {isDevelopment && <ConsoleLogger />}
        
        {/* トースト通知コンテナ */}
        <Toaster
          position="top-right"
          toastOptions={{
            className: 'dark:bg-gray-800 dark:text-white',
            style: {
              background: '#fff',
              color: '#333',
            },
            duration: 5000,
          }}
        />
      </Router>
    </WebSocketProvider>
  );
}

export default App;
