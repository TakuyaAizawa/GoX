import apiClient from '../api/client';
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

export interface UpdateProfileData {
  display_name: string;
  bio?: string;
}

/**
 * ユーザープロフィールを取得する
 */
export const getUserProfile = async (username: string): Promise<User> => {
  try {
    const response = await apiClient.get(`users/${username}`);
    
    // APIレスポンス形式に応じてデータを取得
    let userData = null;
    
    if (response.data.user) {
      userData = response.data.user;
    } else if (response.data.data && response.data.data.user) {
      userData = response.data.data.user;
    } else if (response.data.success === true && response.data.data) {
      userData = response.data.data;
    }
    
    if (!userData) {
      console.error('ユーザーデータが応答に見つかりません:', response.data);
      throw new Error('ユーザーデータの形式が不正です');
    }
    
    return userData;
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
    const response = await apiClient.get(`users/${username}/posts${query}`);
    
    // APIレスポンス形式に応じてデータを取得
    let posts = null;
    
    if (Array.isArray(response.data)) {
      posts = response.data;
    } else if (Array.isArray(response.data.posts)) {
      posts = response.data.posts;
    } else if (response.data.data && Array.isArray(response.data.data.posts)) {
      posts = response.data.data.posts;
    } else if (response.data.success === true && Array.isArray(response.data.data)) {
      posts = response.data.data;
    }
    
    if (!posts) {
      console.error('投稿データが応答に見つかりません:', response.data);
      return [];
    }
    
    return posts;
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
    await apiClient.post(`users/${username}/follow`);
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
    await apiClient.delete(`users/${username}/follow`);
  } catch (error) {
    console.error('フォロー解除エラー:', error);
    throw error;
  }
};

/**
 * フォロワーリストを取得する
 */
export const getFollowers = async (username: string, params?: UserProfileParams): Promise<User[]> => {
  try {
    const queryParams = new URLSearchParams();
    if (params?.page) queryParams.append('page', params.page.toString());
    if (params?.limit) queryParams.append('limit', params.limit.toString());
    
    const query = queryParams.toString() ? `?${queryParams.toString()}` : '';
    const response = await apiClient.get(`users/${username}/followers${query}`);
    
    // APIレスポンス形式に応じてデータを取得
    let followers = null;
    
    if (Array.isArray(response.data)) {
      followers = response.data;
    } else if (Array.isArray(response.data.followers)) {
      followers = response.data.followers;
    } else if (response.data.data && Array.isArray(response.data.data.users)) {
      followers = response.data.data.users;
    } else if (response.data.success === true && Array.isArray(response.data.data)) {
      followers = response.data.data;
    }
    
    if (!followers) {
      console.error('フォロワーデータが応答に見つかりません:', response.data);
      return [];
    }
    
    return followers;
  } catch (error) {
    console.error('フォロワー取得エラー:', error);
    throw error;
  }
};

/**
 * フォロー中のユーザーリストを取得する
 */
export const getFollowing = async (username: string, params?: UserProfileParams): Promise<User[]> => {
  try {
    const queryParams = new URLSearchParams();
    if (params?.page) queryParams.append('page', params.page.toString());
    if (params?.limit) queryParams.append('limit', params.limit.toString());
    
    const query = queryParams.toString() ? `?${queryParams.toString()}` : '';
    const response = await apiClient.get(`users/${username}/following${query}`);
    
    // APIレスポンス形式に応じてデータを取得
    let following = null;
    
    if (Array.isArray(response.data)) {
      following = response.data;
    } else if (Array.isArray(response.data.following)) {
      following = response.data.following;
    } else if (response.data.data && Array.isArray(response.data.data.users)) {
      following = response.data.data.users;
    } else if (response.data.success === true && Array.isArray(response.data.data)) {
      following = response.data.data;
    }
    
    if (!following) {
      console.error('フォロー中データが応答に見つかりません:', response.data);
      return [];
    }
    
    return following;
  } catch (error) {
    console.error('フォロー中取得エラー:', error);
    throw error;
  }
};

/**
 * プロフィールを更新する
 */
export const updateProfile = async (profileData: UpdateProfileData): Promise<User> => {
  try {
    const response = await apiClient.put('users/me', profileData);
    
    // APIレスポンス形式に応じてデータを取得
    let userData = null;
    
    if (response.data.user) {
      userData = response.data.user;
    } else if (response.data.data && response.data.data.user) {
      userData = response.data.data.user;
    } else if (response.data.success === true && response.data.data) {
      userData = response.data.data;
    }
    
    if (!userData) {
      console.error('更新されたユーザーデータが応答に見つかりません:', response.data);
      throw new Error('ユーザーデータの形式が不正です');
    }
    
    return userData;
  } catch (error) {
    console.error('プロフィール更新エラー:', error);
    throw error;
  }
};

/**
 * アバター画像をアップロードする
 */
export const uploadAvatar = async (file: File): Promise<string> => {
  try {
    const formData = new FormData();
    formData.append('avatar', file);
    
    const response = await apiClient.post('users/me/avatar', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    });
    
    let avatarUrl = null;
    
    if (response.data.avatar_url) {
      avatarUrl = response.data.avatar_url;
    } else if (response.data.data && response.data.data.avatar_url) {
      avatarUrl = response.data.data.avatar_url;
    } else if (response.data.success === true && response.data.data && response.data.data.url) {
      avatarUrl = response.data.data.url;
    }
    
    if (!avatarUrl) {
      console.error('アバターURLが応答に見つかりません:', response.data);
      throw new Error('アバターのアップロードに失敗しました');
    }
    
    return avatarUrl;
  } catch (error) {
    console.error('アバターアップロードエラー:', error);
    throw error;
  }
};

/**
 * バナー画像をアップロードする
 */
export const uploadBanner = async (file: File): Promise<string> => {
  try {
    const formData = new FormData();
    formData.append('banner', file);
    
    const response = await apiClient.post('users/me/banner', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    });
    
    let bannerUrl = null;
    
    if (response.data.banner_url) {
      bannerUrl = response.data.banner_url;
    } else if (response.data.data && response.data.data.banner_url) {
      bannerUrl = response.data.data.banner_url;
    } else if (response.data.success === true && response.data.data && response.data.data.url) {
      bannerUrl = response.data.data.url;
    }
    
    if (!bannerUrl) {
      console.error('バナーURLが応答に見つかりません:', response.data);
      throw new Error('バナーのアップロードに失敗しました');
    }
    
    return bannerUrl;
  } catch (error) {
    console.error('バナーアップロードエラー:', error);
    throw error;
  }
}; 