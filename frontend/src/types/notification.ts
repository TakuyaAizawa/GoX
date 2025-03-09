import { Notification } from '../store/notificationStore';

// 表示用に拡張した通知型
export interface DisplayNotification extends Notification {
  count?: number;
} 