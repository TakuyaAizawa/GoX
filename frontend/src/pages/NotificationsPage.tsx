import { useState, useEffect, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { useNotificationStore } from '../store/notificationStore';
import Button from '../components/ui/Button';
import Header from '../components/layout/Header';
import NotificationCard from '../components/notification/NotificationCard';
import { DisplayNotification } from '../types/notification';

type FilterType = 'all' | 'follow' | 'like' | 'reply' | 'mention';

const NotificationsPage = () => {
  const navigate = useNavigate();
  const { 
    notifications, 
    loading, 
    error, 
    hasMore, 
    fetchNotifications, 
    markAsRead, 
    markAllAsRead 
  } = useNotificationStore();
  
  const [loadingMore, setLoadingMore] = useState(false);
  const [filterType, setFilterType] = useState<FilterType>('all');
  
  // 通知を取得
  useEffect(() => {
    fetchNotifications(1);
  }, [fetchNotifications]);
  
  // 通知をフィルタリング
  const filteredNotifications = useMemo(() => {
    if (filterType === 'all') {
      return notifications;
    }
    return notifications.filter(notification => notification.type === filterType);
  }, [notifications, filterType]);
  
  // 通知をグループ化（同じタイプ、同じアクターからの短時間の通知をグループ化）
  const groupedNotifications = useMemo(() => {
    // 最終的なグループ化された通知リスト
    const result: DisplayNotification[] = [];
    
    // コピーして降順ソート
    const sortedNotifications = [...filteredNotifications].sort(
      (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
    );
    
    sortedNotifications.forEach(notification => {
      // すでにグループがあるか探す（過去30分以内、同じアクターと同じタイプ）
      const existingGroupIndex = result.findIndex(group => {
        if (group.type !== notification.type) return false;
        if (group.actor_id !== notification.actor_id) return false;
        
        // 30分以内かどうか
        const timeDiff = Math.abs(
          new Date(group.created_at).getTime() - new Date(notification.created_at).getTime()
        );
        return timeDiff < 30 * 60 * 1000; // 30分 = 30 * 60 * 1000ミリ秒
      });
      
      if (existingGroupIndex !== -1) {
        // グループが見つかった場合はカウントを増やす
        const group = result[existingGroupIndex];
        result[existingGroupIndex] = {
          ...group,
          count: (group.count || 1) + 1,
          // 既読状態：グループ内で1つでも未読があれば未読とする
          read: group.read && notification.read
        };
      } else {
        // 新しいグループを作成
        result.push({ ...notification, count: 1 });
      }
    });
    
    return result;
  }, [filteredNotifications]);
  
  // 追加の通知を読み込む
  const loadMoreNotifications = async () => {
    if (loadingMore || !hasMore) return;
    
    setLoadingMore(true);
    try {
      await fetchNotifications(Math.ceil(notifications.length / 20) + 1);
    } finally {
      setLoadingMore(false);
    }
  };
  
  // 特定の通知を既読にする
  const handleMarkAsRead = async (notificationId: string) => {
    await markAsRead([notificationId]);
  };
  
  // 通知に基づくアクション
  const handleNotificationClick = (notification: DisplayNotification) => {
    // 未読なら既読にする
    if (!notification.read) {
      handleMarkAsRead(notification.id);
    }
    
    if (notification.type === 'like' || notification.type === 'reply') {
      // 投稿詳細ページに移動
      if (notification.post_id) {
        navigate(`/post/${notification.post_id}`);
      }
    } else if (notification.type === 'follow') {
      // ユーザープロフィールページに移動
      navigate(`/profile/${notification.actor_username}`);
    }
  };
  
  // フィルターを変更する
  const handleFilterChange = (type: FilterType) => {
    setFilterType(type);
  };
  
  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      {/* ヘッダー */}
      <Header />
      
      {/* サブヘッダー */}
      <div className="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 sticky top-14 z-10">
        <div className="max-w-3xl mx-auto px-4">
          <div className="flex items-center justify-between py-2">
            <h1 className="text-lg font-bold text-gray-900 dark:text-white">通知</h1>
            <Button 
              variant="ghost" 
              size="sm" 
              onClick={markAllAsRead}
              disabled={notifications.every(n => n.read) || notifications.length === 0}
            >
              すべて既読
            </Button>
          </div>
          
          {/* フィルタータブ */}
          <div className="flex space-x-1 pb-1 overflow-x-auto">
            <button
              onClick={() => handleFilterChange('all')}
              className={`px-3 py-2 text-sm rounded-md ${
                filterType === 'all'
                  ? 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-200'
                  : 'text-gray-600 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700'
              }`}
            >
              すべて
            </button>
            <button
              onClick={() => handleFilterChange('follow')}
              className={`px-3 py-2 text-sm rounded-md ${
                filterType === 'follow'
                  ? 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-200'
                  : 'text-gray-600 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700'
              }`}
            >
              <span className="mr-1">👤</span>フォロー
            </button>
            <button
              onClick={() => handleFilterChange('like')}
              className={`px-3 py-2 text-sm rounded-md ${
                filterType === 'like'
                  ? 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-200'
                  : 'text-gray-600 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700'
              }`}
            >
              <span className="mr-1">❤️</span>いいね
            </button>
            <button
              onClick={() => handleFilterChange('reply')}
              className={`px-3 py-2 text-sm rounded-md ${
                filterType === 'reply'
                  ? 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-200'
                  : 'text-gray-600 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700'
              }`}
            >
              <span className="mr-1">💬</span>返信
            </button>
            <button
              onClick={() => handleFilterChange('mention')}
              className={`px-3 py-2 text-sm rounded-md ${
                filterType === 'mention'
                  ? 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-200'
                  : 'text-gray-600 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700'
              }`}
            >
              <span className="mr-1">@️</span>メンション
            </button>
          </div>
        </div>
      </div>
      
      {/* メインコンテンツ */}
      <main className="max-w-3xl mx-auto px-4 py-4">
        <div className="bg-white dark:bg-gray-800 shadow rounded-lg">
          {loading && notifications.length === 0 ? (
            <div className="p-8 text-center text-gray-500 dark:text-gray-400">
              読み込み中...
            </div>
          ) : error ? (
            <div className="p-8 text-center text-red-500">
              {error}
            </div>
          ) : groupedNotifications.length === 0 ? (
            <div className="p-8 text-center text-gray-500 dark:text-gray-400">
              {filterType === 'all' 
                ? '通知はありません' 
                : `${filterType}タイプの通知はありません`}
            </div>
          ) : (
            <ul>
              {groupedNotifications.map(notification => (
                <li
                  key={notification.id}
                  className={`border-b border-gray-200 dark:border-gray-700 last:border-b-0 ${
                    !notification.read ? 'bg-blue-50 dark:bg-blue-900/20' : ''
                  }`}
                >
                  <NotificationCard
                    notification={notification}
                    onClick={handleNotificationClick}
                  />
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