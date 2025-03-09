import { formatDistance } from 'date-fns';
import { ja } from 'date-fns/locale';
import { DisplayNotification } from '../../types/notification';

interface NotificationCardProps {
  notification: DisplayNotification;
  onClick: (notification: DisplayNotification) => void;
}

export const getNotificationIcon = (type: string) => {
  switch (type) {
    case 'like':
      return '❤️';
    case 'follow':
      return '👤';
    case 'reply':
      return '💬';
    case 'mention':
      return '@️';
    default:
      return '🔔';
  }
};

export const getNotificationMessage = (notification: DisplayNotification) => {
  const actorName = notification.actor_display_name || notification.actor_username;
  const countText = notification.count && notification.count > 1 
    ? `や他${notification.count - 1}件` 
    : '';
  
  switch (notification.type) {
    case 'like':
      return `${actorName}${countText}があなたの投稿にいいねしました`;
    case 'follow':
      return `${actorName}${countText}があなたをフォローしました`;
    case 'reply':
      return `${actorName}${countText}があなたの投稿に返信しました`;
    case 'mention':
      return `${actorName}${countText}があなたについて言及しました`;
    default:
      return '新しい通知があります';
  }
};

const formatDate = (dateString: string) => {
  return formatDistance(new Date(dateString), new Date(), {
    addSuffix: true,
    locale: ja
  });
};

const NotificationCard: React.FC<NotificationCardProps> = ({ notification, onClick }) => {
  return (
    <button
      onClick={() => onClick(notification)}
      className="w-full text-left p-4 flex items-start hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors"
    >
      {/* アクターのアバター */}
      <div className="mr-4 mt-1">
        <div className="relative">
          <img
            src={notification.actor_avatar_url || '/default-avatar.png'}
            alt={`${notification.actor_display_name || notification.actor_username}のアバター`}
            className="w-10 h-10 rounded-full object-cover"
          />
          {notification.count && notification.count > 1 && (
            <span className="absolute -bottom-1 -right-1 bg-blue-500 text-white text-xs rounded-full w-5 h-5 flex items-center justify-center">
              {notification.count > 9 ? '9+' : notification.count}
            </span>
          )}
        </div>
      </div>
      
      {/* 通知内容 */}
      <div className="flex-1">
        <div className="flex items-start justify-between">
          <div>
            <p className="text-gray-900 dark:text-white">
              <span className="mr-2">{getNotificationIcon(notification.type)}</span>
              {getNotificationMessage(notification)}
            </p>
            
            {notification.post_content && (
              <p className="mt-1 text-gray-600 dark:text-gray-300 text-sm">
                {notification.post_content.length > 100
                  ? `${notification.post_content.substring(0, 100)}...`
                  : notification.post_content}
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
  );
};

export default NotificationCard; 