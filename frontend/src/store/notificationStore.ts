import { create } from 'zustand';
import { getNotifications, markNotificationsAsRead } from '../services/notificationService';

export interface Notification {
  id: string;
  user_id: string;
  type: 'like' | 'follow' | 'reply' | 'mention';
  actor_id: string;
  actor_username: string;
  actor_display_name: string;
  actor_avatar_url: string | null;
  post_id?: string;
  post_content?: string;
  created_at: string;
  read: boolean;
}

interface NotificationState {
  notifications: Notification[];
  unreadCount: number;
  loading: boolean;
  error: string | null;
  hasMore: boolean;
  page: number;
  
  // アクション
  fetchNotifications: (page?: number) => Promise<void>;
  markAsRead: (ids?: string[]) => Promise<void>;
  markAllAsRead: () => Promise<void>;
  addNotification: (notification: Notification) => void;
  setUnreadCount: (count: number) => void;
  resetNotifications: () => void;
}

export const useNotificationStore = create<NotificationState>((set, get) => ({
  notifications: [],
  unreadCount: 0,
  loading: false,
  error: null,
  hasMore: true,
  page: 1,
  
  // 通知を取得
  fetchNotifications: async (page = 1) => {
    const isFirstPage = page === 1;
    const { loading } = get();
    
    if (loading) return;
    
    set({ loading: true, error: null });
    
    try {
      const response = await getNotifications({ page, limit: 20 });
      const newNotifications = response.notifications;
      const unreadCount = response.unread_count;
      
      set(state => ({
        notifications: isFirstPage 
          ? newNotifications
          : [...state.notifications, ...newNotifications],
        unreadCount,
        loading: false,
        hasMore: newNotifications.length === 20,
        page,
        error: null
      }));
    } catch (error) {
      console.error('通知の取得に失敗しました', error);
      set({ 
        loading: false, 
        error: error instanceof Error ? error.message : '通知の取得に失敗しました' 
      });
    }
  },
  
  // 特定の通知を既読にする
  markAsRead: async (ids) => {
    try {
      await markNotificationsAsRead(ids);
      
      // 既読にした通知を更新
      set(state => {
        const updatedNotifications = state.notifications.map(notification => {
          if (!ids || ids.includes(notification.id)) {
            return { ...notification, read: true };
          }
          return notification;
        });
        
        // 未読数を再計算
        const unreadCount = updatedNotifications.filter(n => !n.read).length;
        
        return {
          notifications: updatedNotifications,
          unreadCount
        };
      });
    } catch (error) {
      console.error('通知の既読処理に失敗しました', error);
    }
  },
  
  // すべての通知を既読にする
  markAllAsRead: async () => {
    try {
      await markNotificationsAsRead();
      
      set(state => ({
        notifications: state.notifications.map(notification => ({
          ...notification,
          read: true
        })),
        unreadCount: 0
      }));
    } catch (error) {
      console.error('すべての通知の既読処理に失敗しました', error);
    }
  },
  
  // リアルタイム通知を追加
  addNotification: (notification) => {
    set(state => {
      // 既に同じIDの通知がある場合は追加しない
      if (state.notifications.some(n => n.id === notification.id)) {
        return state;
      }
      
      // 新しい通知を先頭に追加
      const updatedNotifications = [notification, ...state.notifications];
      
      // 未読数を更新
      const unreadCount = updatedNotifications.filter(n => !n.read).length;
      
      return {
        notifications: updatedNotifications,
        unreadCount
      };
    });
  },
  
  // 未読数を直接設定
  setUnreadCount: (count) => {
    set({ unreadCount: count });
  },
  
  // 通知をリセット
  resetNotifications: () => {
    set({
      notifications: [],
      unreadCount: 0,
      loading: false,
      error: null,
      hasMore: true,
      page: 1
    });
  }
})); 