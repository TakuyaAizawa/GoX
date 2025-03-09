import { useEffect } from 'react';
import { useAuthStore } from '../../store/authStore';
import { useNotificationStore } from '../../store/notificationStore';
import { useWebSocketContext } from '../../context/WebSocketContext';
import { toast } from 'react-hot-toast';

// 通知音
const notificationSound = new Audio('/notification.mp3');

/**
 * WebSocketからリアルタイム通知を受け取り、ストアに追加するコンポーネント
 * このコンポーネントは表示するUIはなく、バックグラウンドで動作します
 */
const NotificationListener: React.FC = () => {
  const { isAuthenticated, user } = useAuthStore();
  const { isConnected, addMessageHandler } = useWebSocketContext();
  const { addNotification, setUnreadCount } = useNotificationStore();
  
  // 通知のタイプに応じたメッセージを生成
  const getNotificationMessage = (notification: any) => {
    const actorName = notification.actor_display_name || notification.actor_username;
    
    switch (notification.type) {
      case 'like':
        return `${actorName}があなたの投稿にいいねしました`;
      case 'follow':
        return `${actorName}があなたをフォローしました`;
      case 'reply':
        return `${actorName}があなたの投稿に返信しました`;
      case 'mention':
        return `${actorName}があなたについて言及しました`;
      default:
        return '新しい通知があります';
    }
  };
  
  // WebSocketからの通知メッセージをリスニング
  useEffect(() => {
    if (!isAuthenticated || !isConnected) return;
    
    // notification タイプのメッセージをリスニング
    const removeHandler = addMessageHandler('notification', (data) => {
      console.log('新しい通知を受信:', data);
      
      // 通知データを検証
      if (!data || !data.id) {
        console.error('無効な通知データ:', data);
        return;
      }
      
      // 通知をストアに追加
      addNotification(data);
      
      // 通知音を再生
      try {
        notificationSound.play();
      } catch (error) {
        console.warn('通知音の再生に失敗しました:', error);
      }
      
      // トースト通知を表示
      toast(getNotificationMessage(data), {
        icon: '🔔',
        position: 'top-right',
        duration: 4000,
      });
    });
    
    // unread_count タイプのメッセージをリスニング
    const removeCountHandler = addMessageHandler('unread_count', (data) => {
      if (typeof data === 'number') {
        setUnreadCount(data);
      } else if (data && typeof data.count === 'number') {
        setUnreadCount(data.count);
      }
    });
    
    return () => {
      removeHandler();
      removeCountHandler();
    };
  }, [isAuthenticated, isConnected, addMessageHandler, addNotification, setUnreadCount]);
  
  // 必要であれば定期的に未読数を取得するためのポーリング
  useEffect(() => {
    if (!isAuthenticated) return;
    
    // WebSocketが切断されている場合のフォールバックとして、
    // 5分ごとに未読数を取得する
    const interval = setInterval(() => {
      if (!isConnected) {
        // 未読数を取得するAPIを直接呼び出す
        fetch(`${import.meta.env.VITE_API_BASE_URL}/api/v1/notifications/unread`, {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`
          }
        })
          .then(res => res.json())
          .then(data => {
            const count = data.unread_count || data.count || 0;
            setUnreadCount(count);
          })
          .catch(err => console.error('未読通知数の取得に失敗しました:', err));
      }
    }, 5 * 60 * 1000); // 5分ごと
    
    return () => clearInterval(interval);
  }, [isAuthenticated, isConnected, setUnreadCount]);
  
  // このコンポーネントは表示するUIはないため、nullを返す
  return null;
};

export default NotificationListener; 