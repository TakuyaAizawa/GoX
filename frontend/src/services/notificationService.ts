import client from '../api/client';

export interface Notification {
  id: string;
  type: 'like' | 'follow' | 'reply';
  actor_id: string;
  actor_username: string;
  actor_display_name: string;
  actor_avatar_url: string | null;
  target_id: string | null;
  target_type: 'post' | 'user' | null;
  content: string | null;
  is_read: boolean;
  created_at: string;
}

export interface NotificationParams {
  page?: number;
  limit?: number;
}

/**
 * 通知一覧を取得する
 */
export const getNotifications = async (params: NotificationParams = {}) => {
  try {
    const { page = 1, limit = 20 } = params;
    const response = await client.get<Notification[]>('/api/v1/notifications', {
      params: { page, limit }
    });
    return response.data;
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
    const response = await client.get<{ count: number }>('/api/v1/notifications/unread');
    return response.data.count;
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