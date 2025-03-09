import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuthStore } from '../../store/authStore';
import NotificationIcon from '../notification/NotificationIcon';

interface HeaderProps {
  hideNav?: boolean;
}

const Header: React.FC<HeaderProps> = ({ hideNav = false }) => {
  const { user, logout } = useAuthStore();
  const navigate = useNavigate();
  const [showDropdown, setShowDropdown] = useState(false);
  
  const handleLogout = () => {
    logout();
    navigate('/login');
  };
  
  return (
    <header className="sticky top-0 z-10 bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-700">
      <div className="max-w-6xl mx-auto px-4 py-2">
        <div className="flex items-center justify-between h-14">
          {/* ロゴ */}
          <div className="flex items-center">
            <Link to="/" className="text-xl font-bold text-blue-500">
              GoX
            </Link>
          </div>
          
          {/* ナビゲーション - hideNav が true の場合は表示しない */}
          {!hideNav && user && (
            <nav className="flex items-center space-x-2">
              {/* ホームアイコン */}
              <Link 
                to="/"
                className="p-2 rounded-full hover:bg-gray-100 dark:hover:bg-gray-800"
                aria-label="ホーム"
              >
                <svg 
                  className="w-6 h-6 text-gray-700 dark:text-gray-300" 
                  fill="none" 
                  stroke="currentColor" 
                  viewBox="0 0 24 24" 
                  xmlns="http://www.w3.org/2000/svg"
                >
                  <path 
                    strokeLinecap="round" 
                    strokeLinejoin="round" 
                    strokeWidth={1.5} 
                    d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"
                  />
                </svg>
              </Link>
              
              {/* 通知アイコン */}
              <NotificationIcon />
              
              {/* プロフィールアイコン */}
              <div className="relative">
                <button
                  className="flex items-center p-1 rounded-full hover:bg-gray-100 dark:hover:bg-gray-800"
                  onClick={() => setShowDropdown(!showDropdown)}
                  aria-label="ユーザーメニュー"
                >
                  <div className="w-8 h-8 rounded-full overflow-hidden bg-gray-300 dark:bg-gray-600">
                    {user.avatar_url ? (
                      <img 
                        src={user.avatar_url} 
                        alt={user.display_name || user.username} 
                        className="w-full h-full object-cover"
                      />
                    ) : (
                      <div className="w-full h-full flex items-center justify-center text-gray-500 dark:text-gray-400">
                        {(user.display_name || user.username || 'U').charAt(0).toUpperCase()}
                      </div>
                    )}
                  </div>
                </button>
                
                {/* ドロップダウンメニュー */}
                {showDropdown && (
                  <div className="absolute right-0 mt-2 w-48 py-2 bg-white dark:bg-gray-800 rounded-md shadow-lg border border-gray-200 dark:border-gray-700">
                    <Link
                      to={`/profile/${user.username}`}
                      className="block px-4 py-2 text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700"
                      onClick={() => setShowDropdown(false)}
                    >
                      プロフィール
                    </Link>
                    <button
                      className="block w-full text-left px-4 py-2 text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700"
                      onClick={handleLogout}
                    >
                      ログアウト
                    </button>
                  </div>
                )}
              </div>
            </nav>
          )}
        </div>
      </div>
    </header>
  );
};

export default Header; 