import { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { formatDistance } from 'date-fns';
import { ja } from 'date-fns/locale';
import { getNotifications, markAllAsRead, Notification } from '../services/notificationService';
import Button from '../components/ui/Button';

const NotificationsPage = () => {
  const navigate = useNavigate();
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  
  // 通知を取得
  useEffect(() => {
    const fetchNotifications = async () => {
      setLoading(true);
      setError(null);
      
      try {
        const data = await getNotifications({ page: 1, limit: 20 });
        setNotifications(data);
        setHasMore(data.length === 20);
      } catch (error) {
        console.error('通知取得エラー:', error);
        setError('通知の取得に失敗しました');
      } finally {
        setLoading(false);
      }
    };
    
    fetchNotifications();
  }, []);
  
  // 追加の通知を読み込む
  const loadMoreNotifications = async () => {
    if (loadingMore || !hasMore) return;
    
    setLoadingMore(true);
    
    try {
      const nextPage = page + 1;
      const data = await getNotifications({ page: nextPage, limit: 20 });
      
      if (data.length > 0) {
        setNotifications(prev => [...prev, ...data]);
        setPage(nextPage);
      }
      
      setHasMore(data.length === 20);
    } catch (error) {
      console.error('追加通知取得エラー:', error);
    } finally {
      setLoadingMore(false);
    }
  };
  
  // 全ての通知を既読にする
  const handleMarkAllAsRead = async () => {
    try {
      await markAllAsRead();
      // 既読状態を更新
      setNotifications(prev => 
        prev.map(notification => ({ ...notification, is_read: true }))
      );
    } catch (error) {
      console.error('通知既読エラー:', error);
    }
  };
  
  // 通知に基づくアクション
  const handleNotificationClick = (notification: Notification) => {
    if (notification.type === 'like' || notification.type === 'reply') {
      // 投稿詳細ページに移動
      if (notification.target_id) {
        navigate(`/post/${notification.target_id}`);
      }
    } else if (notification.type === 'follow') {
      // ユーザープロフィールページに移動
      navigate(`/profile/${notification.actor_username}`);
    }
  };
  
  // 日付をフォーマット
  const formatDate = (dateString: string) => {
    return formatDistance(new Date(dateString), new Date(), {
      addSuffix: true,
      locale: ja
    });
  };
  
  // 通知タイプに基づくアイコンを取得
  const getNotificationIcon = (type: string) => {
    switch (type) {
      case 'like':
        return '❤️';
      case 'follow':
        return '👤';
      case 'reply':
        return '💬';
      default:
        return '🔔';
    }
  };
  
  // 通知メッセージを取得
  const getNotificationMessage = (notification: Notification) => {
    switch (notification.type) {
      case 'like':
        return `${notification.actor_display_name}さんがあなたの投稿にいいねしました`;
      case 'follow':
        return `${notification.actor_display_name}さんがあなたをフォローしました`;
      case 'reply':
        return `${notification.actor_display_name}さんがあなたの投稿にリプライしました`;
      default:
        return '新しい通知があります';
    }
  };
  
  // 戻るボタンのハンドラー
  const handleGoBack = () => {
    navigate(-1); // ブラウザの履歴で1つ前に戻る
  };
  
  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      {/* ヘッダー */}
      <header className="bg-white dark:bg-gray-800 shadow-sm sticky top-0 z-10">
        <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-14">
            <button
              onClick={handleGoBack}
              className="text-gray-500 hover:text-gray-900 dark:text-gray-300 dark:hover:text-white"
            >
              ← 戻る
            </button>
            <h1 className="text-lg font-bold text-gray-900 dark:text-white">通知</h1>
            <Button 
              variant="ghost" 
              size="sm" 
              onClick={handleMarkAllAsRead}
              disabled={notifications.every(n => n.is_read) || notifications.length === 0}
            >
              すべて既読
            </Button>
          </div>
        </div>
      </header>
      
      {/* メインコンテンツ */}
      <main className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
        <div className="bg-white dark:bg-gray-800 shadow rounded-lg">
          {loading && notifications.length === 0 ? (
            <div className="p-8 text-center text-gray-500 dark:text-gray-400">
              読み込み中...
            </div>
          ) : error ? (
            <div className="p-8 text-center text-red-500">
              {error}
            </div>
          ) : notifications.length === 0 ? (
            <div className="p-8 text-center text-gray-500 dark:text-gray-400">
              通知はありません
            </div>
          ) : (
            <ul>
              {notifications.map(notification => (
                <li
                  key={notification.id}
                  className={`border-b border-gray-200 dark:border-gray-700 last:border-b-0 ${
                    !notification.is_read ? 'bg-blue-50 dark:bg-blue-900/20' : ''
                  }`}
                >
                  <button
                    onClick={() => handleNotificationClick(notification)}
                    className="w-full text-left p-4 flex items-start hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors"
                  >
                    {/* アクターのアバター */}
                    <div className="mr-4 mt-1">
                      <img
                        src={notification.actor_avatar_url || '/default-avatar.png'}
                        alt={`${notification.actor_display_name}のアバター`}
                        className="w-10 h-10 rounded-full object-cover"
                      />
                    </div>
                    
                    {/* 通知内容 */}
                    <div className="flex-1">
                      <div className="flex items-start justify-between">
                        <div>
                          <p className="text-gray-900 dark:text-white">
                            <span className="mr-2">{getNotificationIcon(notification.type)}</span>
                            {getNotificationMessage(notification)}
                          </p>
                          
                          {notification.content && (
                            <p className="mt-1 text-gray-600 dark:text-gray-300 text-sm">
                              {notification.content.length > 100
                                ? `${notification.content.substring(0, 100)}...`
                                : notification.content}
                            </p>
                          )}
                        </div>
                        
                        {/* 日付 */}
                        <span className="text-xs text-gray-500 dark:text-gray-400 whitespace-nowrap ml-2">
                          {formatDate(notification.created_at)}
                        </span>
                      </div>
                    </div>
                  </button>
                </li>
              ))}
            </ul>
          )}
          
          {/* もっと読み込むボタン */}
          {hasMore && notifications.length > 0 && (
            <div className="p-4 text-center">
              <Button
                variant="outline"
                onClick={loadMoreNotifications}
                isLoading={loadingMore}
                disabled={loadingMore}
              >
                もっと読み込む
              </Button>
            </div>
          )}
        </div>
      </main>
    </div>
  );
};

export default NotificationsPage; 