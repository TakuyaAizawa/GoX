import { useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useNotificationStore } from '../../store/notificationStore';
import { getUnreadCount } from '../../services/notificationService';

interface NotificationIconProps {
  className?: string;
}

/**
 * 通知アイコンコンポーネント
 * 未読通知がある場合はカウントバッジを表示します
 */
const NotificationIcon: React.FC<NotificationIconProps> = ({ className = '' }) => {
  const { unreadCount, setUnreadCount } = useNotificationStore();
  
  // コンポーネントマウント時に未読通知数を取得
  useEffect(() => {
    const fetchUnreadCount = async () => {
      try {
        const count = await getUnreadCount();
        setUnreadCount(count);
      } catch (error) {
        console.error('未読通知数の取得に失敗しました:', error);
      }
    };
    
    fetchUnreadCount();
    
    // 定期的に未読数を更新（フォールバック）
    const interval = setInterval(fetchUnreadCount, 5 * 60 * 1000); // 5分ごと
    
    return () => clearInterval(interval);
  }, [setUnreadCount]);
  
  return (
    <Link 
      to="/notifications"
      className={`relative p-2 flex items-center justify-center rounded-full hover:bg-gray-100 dark:hover:bg-gray-800 ${className}`}
      aria-label="通知"
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
          d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"
        />
      </svg>
      
      {/* 未読通知カウントバッジ */}
      {unreadCount > 0 && (
        <span className="absolute top-0.5 right-0.5 inline-flex items-center justify-center w-5 h-5 text-xs font-bold text-white bg-red-500 rounded-full">
          {unreadCount > 99 ? '99+' : unreadCount}
        </span>
      )}
    </Link>
  );
};

export default NotificationIcon; 