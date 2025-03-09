import client from '../api/client';

export interface Notification {
  id: string;
  user_id: string;
  type: 'like' | 'follow' | 'reply' | 'mention';
  actor_id: string;
  actor_username: string;
  actor_display_name: string;
  actor_avatar_url: string | null;
  target_id: string | null;
  target_type: 'post' | 'user' | null;
  post_id?: string;
  post_content?: string;
  content: string | null;
  read: boolean;
  is_read: boolean; // APIの互換性のため
  created_at: string;
}

export interface NotificationResponse {
  notifications: Notification[];
  unread_count: number;
}

export interface NotificationParams {
  page?: number;
  limit?: number;
}

/**
 * 通知一覧を取得する
 */
export const getNotifications = async (params: NotificationParams = {}): Promise<NotificationResponse> => {
  try {
    const { page = 1, limit = 20 } = params;
    const response = await client.get<{ notifications: Notification[], unread_count: number } | Notification[]>('/api/v1/notifications', {
      params: { page, limit }
    });
    
    // レスポンス形式の違いを吸収
    if (Array.isArray(response.data)) {
      // 古い形式のAPI
      return {
        notifications: response.data,
        unread_count: response.data.filter(n => !n.is_read && !n.read).length
      };
    }
    
    // 新しい形式のAPI（オブジェクトで返る）
    const notifications = response.data.notifications.map(n => ({
      ...n,
      read: n.read || n.is_read // is_readとreadの互換性を確保
    }));
    
    return {
      notifications,
      unread_count: response.data.unread_count
    };
  } catch (error) {
    console.error('通知取得エラー:', error);
    throw error;
  }
};

/**
 * 未読通知数を取得する
 */
export const getUnreadCount = async () => {
  try {
    const response = await client.get<{ count: number } | { unread_count: number }>('/api/v1/notifications/unread');
    
    // レスポンス形式の違いを吸収
    if ('count' in response.data) {
      return response.data.count;
    }
    
    return response.data.unread_count;
  } catch (error) {
    console.error('未読通知数取得エラー:', error);
    throw error;
  }
};

/**
 * 全ての通知を既読にする
 */
export const markAllAsRead = async () => {
  try {
    await client.put('/api/v1/notifications/read');
    return true;
  } catch (error) {
    console.error('通知既読エラー:', error);
    throw error;
  }
};

/**
 * 特定の通知を既読にする
 */
export const markAsRead = async (notificationId: string) => {
  try {
    await client.put(`/api/v1/notifications/${notificationId}/read`);
    return true;
  } catch (error) {
    console.error('通知既読エラー:', error);
    throw error;
  }
};

/**
 * 複数の通知を既読にする
 * @param ids 既読にする通知ID配列。省略時は全ての通知が対象
 */
export const markNotificationsAsRead = async (ids?: string[]) => {
  try {
    if (ids?.length) {
      // 特定の複数の通知を既読にする
      await client.put('/api/v1/notifications/read', { notification_ids: ids });
    } else {
      // すべての通知を既読にする
      await client.put('/api/v1/notifications/read');
    }
    return true;
  } catch (error) {
    console.error('通知既読エラー:', error);
    throw error;
  }
}; 