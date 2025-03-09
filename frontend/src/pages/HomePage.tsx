import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuthStore } from '../store/authStore';
import Button from '../components/ui/Button';

const HomePage = () => {
  const navigate = useNavigate();
  const { user, logout } = useAuthStore();
  const [isLoggingOut, setIsLoggingOut] = useState(false);

  const handleLogout = async () => {
    setIsLoggingOut(true);
    try {
      await logout();
      navigate('/login');
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      setIsLoggingOut(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      {/* ヘッダー */}
      <header className="bg-white dark:bg-gray-800 shadow-sm">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex">
              <div className="flex-shrink-0 flex items-center">
                <h1 className="text-2xl font-bold text-primary-600">GoX</h1>
              </div>
              <nav className="ml-6 flex space-x-8">
                <Link
                  to="/"
                  className="border-primary-500 text-gray-900 dark:text-white inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium"
                >
                  ホーム
                </Link>
                <Link
                  to="/notifications"
                  className="border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-300 dark:hover:text-white inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium"
                >
                  通知
                </Link>
              </nav>
            </div>
            <div className="flex items-center">
              {user && (
                <div className="flex items-center space-x-4">
                  <Link
                    to={`/profile/${user.username}`}
                    className="text-gray-500 hover:text-gray-700 dark:text-gray-300 dark:hover:text-white"
                  >
                    <div className="flex items-center space-x-2">
                      <img
                        src={user.avatar_url || '/default-avatar.png'}
                        alt={user.display_name}
                        className="h-8 w-8 rounded-full"
                      />
                      <span>{user.display_name}</span>
                    </div>
                  </Link>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={handleLogout}
                    isLoading={isLoggingOut}
                  >
                    ログアウト
                  </Button>
                </div>
              )}
            </div>
          </div>
        </div>
      </header>

      {/* メインコンテンツ */}
      <main className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="bg-white dark:bg-gray-800 shadow rounded-lg p-6">
          <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">タイムライン</h2>
          
          {/* 投稿作成フォーム */}
          <div className="mb-6 bg-gray-50 dark:bg-gray-700 rounded-lg p-4">
            <textarea
              className="w-full p-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 dark:bg-gray-800 dark:text-white"
              placeholder="いまどうしてる？"
              rows={3}
            />
            <div className="mt-2 flex justify-end">
              <Button size="sm">投稿する</Button>
            </div>
          </div>
          
          {/* タイムラインの投稿（一時的なモックデータ） */}
          <div className="space-y-6">
            <div className="border-b border-gray-200 dark:border-gray-700 pb-4">
              <div className="flex items-start space-x-3">
                <img
                  src="/default-avatar.png"
                  alt="User"
                  className="h-10 w-10 rounded-full"
                />
                <div>
                  <div className="flex items-center space-x-2">
                    <span className="font-semibold text-gray-900 dark:text-white">サンプルユーザー</span>
                    <span className="text-gray-500 dark:text-gray-400 text-sm">@sample_user</span>
                    <span className="text-gray-500 dark:text-gray-400 text-sm">2時間前</span>
                  </div>
                  <p className="mt-1 text-gray-800 dark:text-gray-200">
                    GoXプラットフォームへようこそ！このマイクロブログで様々な情報を共有しましょう。
                  </p>
                  <div className="mt-2 flex space-x-4">
                    <button className="text-gray-500 dark:text-gray-400 flex items-center space-x-1 hover:text-primary-500">
                      <span>❤️</span>
                      <span>12</span>
                    </button>
                    <button className="text-gray-500 dark:text-gray-400 flex items-center space-x-1 hover:text-primary-500">
                      <span>💬</span>
                      <span>5</span>
                    </button>
                  </div>
                </div>
              </div>
            </div>
            
            <div className="border-b border-gray-200 dark:border-gray-700 pb-4">
              <div className="flex items-start space-x-3">
                <img
                  src="/default-avatar.png"
                  alt="User"
                  className="h-10 w-10 rounded-full"
                />
                <div>
                  <div className="flex items-center space-x-2">
                    <span className="font-semibold text-gray-900 dark:text-white">テストユーザー</span>
                    <span className="text-gray-500 dark:text-gray-400 text-sm">@test_user</span>
                    <span className="text-gray-500 dark:text-gray-400 text-sm">5時間前</span>
                  </div>
                  <p className="mt-1 text-gray-800 dark:text-gray-200">
                    フロントエンド開発が進んでいます。React + TypeScript + Tailwind CSSの組み合わせは最高です！
                  </p>
                  <div className="mt-2 flex space-x-4">
                    <button className="text-gray-500 dark:text-gray-400 flex items-center space-x-1 hover:text-primary-500">
                      <span>❤️</span>
                      <span>8</span>
                    </button>
                    <button className="text-gray-500 dark:text-gray-400 flex items-center space-x-1 hover:text-primary-500">
                      <span>💬</span>
                      <span>2</span>
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
};

export default HomePage; 