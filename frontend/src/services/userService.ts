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
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/api/v1/users/${username}`, {
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
    console.log('ユーザープロフィール応答データ:', JSON.stringify(data, null, 2));
    
    // APIレスポンス形式に応じてデータを取得
    let userData = null;
    
    if (data.user) {
      userData = data.user;
    } else if (data.data && data.data.user) {
      userData = data.data.user;
    } else if (data.success === true && data.data) {
      userData = data.data;
    }
    
    if (!userData) {
      console.error('ユーザーデータが応答に見つかりません:', data);
      throw new Error('ユーザーデータの形式が不正です');
    }
    
    console.log('解析されたユーザーデータ:', userData);
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
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/api/v1/users/${username}/posts${query}`, {
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
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/api/v1/users/${username}/follow`, {
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
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/api/v1/users/${username}/follow`, {
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
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/api/v1/users/${username}/followers${query}`, {
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
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/api/v1/users/${username}/following${query}`, {
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

/**
 * プロフィール情報を更新する
 */
export const updateProfile = async (profileData: UpdateProfileData): Promise<User> => {
  try {
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/api/v1/users/me`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: JSON.stringify(profileData)
    });
    
    if (!response.ok) {
      throw new Error('プロフィールの更新に失敗しました');
    }
    
    const data = await response.json();
    
    // APIレスポンス形式に応じてデータを取得
    return data.user || data.data?.user;
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
    
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/api/v1/users/me/avatar`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: formData
    });
    
    if (!response.ok) {
      throw new Error('アバター画像のアップロードに失敗しました');
    }
    
    const data = await response.json();
    console.log('アバターアップロードレスポンス:', data);
    
    // アバターURLを返す
    const avatarUrl = data.data?.avatar_url || data.avatar_url;
    if (!avatarUrl) {
      throw new Error('アバターURLの取得に失敗しました');
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
    
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/api/v1/users/me/banner`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: formData
    });
    
    if (!response.ok) {
      throw new Error('バナー画像のアップロードに失敗しました');
    }
    
    const data = await response.json();
    console.log('バナーアップロードレスポンス:', data);
    
    // バナーURLを返す
    const bannerUrl = data.data?.banner_url || data.banner_url;
    if (!bannerUrl) {
      throw new Error('バナーURLの取得に失敗しました');
    }
    
    return bannerUrl;
  } catch (error) {
    console.error('バナーアップロードエラー:', error);
    throw error;
  }
}; 