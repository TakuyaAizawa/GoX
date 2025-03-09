import { Post } from './postService';

export interface User {
  id: string;
  username: string;
  display_name: string;
  email: string;
  avatar_url: string | null;
  banner_url: string | null;
  bio: string | null;
  created_at: string;
  followers_count: number;
  following_count: number;
  posts_count: number;
  is_following: boolean;
}

export interface UserProfileParams {
  page?: number;
  limit?: number;
}

/**
 * ユーザープロフィールを取得する
 */
export const getUserProfile = async (username: string): Promise<User> => {
  try {
    const response = await fetch(`http://localhost:8080/api/v1/users/${username}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    });
    
    if (!response.ok) {
      throw new Error('ユーザープロフィールの取得に失敗しました');
    }
    
    const data = await response.json();
    console.log('ユーザープロフィール:', data);
    
    // APIレスポンス形式に応じてデータを取得
    return data.user || data.data?.user;
  } catch (error) {
    console.error('ユーザープロフィール取得エラー:', error);
    throw error;
  }
};

/**
 * ユーザーの投稿を取得する
 */
export const getUserPosts = async (username: string, params?: UserProfileParams): Promise<Post[]> => {
  try {
    const queryParams = new URLSearchParams();
    if (params?.page) queryParams.append('page', params.page.toString());
    if (params?.limit) queryParams.append('limit', params.limit.toString());
    
    const query = queryParams.toString() ? `?${queryParams.toString()}` : '';
    const response = await fetch(`http://localhost:8080/api/v1/users/${username}/posts${query}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    });
    
    if (!response.ok) {
      throw new Error('ユーザー投稿の取得に失敗しました');
    }
    
    const data = await response.json();
    console.log('ユーザー投稿:', data);
    
    // APIレスポンス形式に応じてデータを取得
    return data.posts || data.data?.posts || [];
  } catch (error) {
    console.error('ユーザー投稿取得エラー:', error);
    throw error;
  }
};

/**
 * ユーザーをフォローする
 */
export const followUser = async (username: string): Promise<void> => {
  try {
    const response = await fetch(`http://localhost:8080/api/v1/users/${username}/follow`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    });
    
    if (!response.ok) {
      throw new Error('ユーザーのフォローに失敗しました');
    }
  } catch (error) {
    console.error('フォローエラー:', error);
    throw error;
  }
};

/**
 * ユーザーのフォローを解除する
 */
export const unfollowUser = async (username: string): Promise<void> => {
  try {
    const response = await fetch(`http://localhost:8080/api/v1/users/${username}/follow`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    });
    
    if (!response.ok) {
      throw new Error('フォロー解除に失敗しました');
    }
  } catch (error) {
    console.error('フォロー解除エラー:', error);
    throw error;
  }
};

/**
 * フォロワー一覧を取得する
 */
export const getFollowers = async (username: string, params?: UserProfileParams): Promise<User[]> => {
  try {
    const queryParams = new URLSearchParams();
    if (params?.page) queryParams.append('page', params.page.toString());
    if (params?.limit) queryParams.append('limit', params.limit.toString());
    
    const query = queryParams.toString() ? `?${queryParams.toString()}` : '';
    const response = await fetch(`http://localhost:8080/api/v1/users/${username}/followers${query}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    });
    
    if (!response.ok) {
      throw new Error('フォロワー一覧の取得に失敗しました');
    }
    
    const data = await response.json();
    
    // APIレスポンス形式に応じてデータを取得
    return data.users || data.data?.users || [];
  } catch (error) {
    console.error('フォロワー取得エラー:', error);
    throw error;
  }
};

/**
 * フォロー中ユーザー一覧を取得する
 */
export const getFollowing = async (username: string, params?: UserProfileParams): Promise<User[]> => {
  try {
    const queryParams = new URLSearchParams();
    if (params?.page) queryParams.append('page', params.page.toString());
    if (params?.limit) queryParams.append('limit', params.limit.toString());
    
    const query = queryParams.toString() ? `?${queryParams.toString()}` : '';
    const response = await fetch(`http://localhost:8080/api/v1/users/${username}/following${query}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    });
    
    if (!response.ok) {
      throw new Error('フォロー中ユーザー一覧の取得に失敗しました');
    }
    
    const data = await response.json();
    
    // APIレスポンス形式に応じてデータを取得
    return data.users || data.data?.users || [];
  } catch (error) {
    console.error('フォロー中ユーザー取得エラー:', error);
    throw error;
  }
}; 