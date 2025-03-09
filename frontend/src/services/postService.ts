import apiClient from '../api/client';

export interface Post {
  id: string;
  content: string;
  user: {
    id: string;
    username: string;
    display_name: string;
    avatar_url: string | null;
  };
  created_at: string;
  likes_count: number;
  replies_count: number;
  is_liked: boolean;
  media_urls: string[];
  parent_id: string | null;
}

export interface TimelineParams {
  page?: number;
  limit?: number;
}

/**
 * ホームタイムラインを取得する（フォロー中のユーザーの投稿）
 */
export const getHomeTimeline = async (params?: TimelineParams): Promise<Post[]> => {
  try {
    const queryParams = new URLSearchParams();
    if (params?.page) queryParams.append('page', params.page.toString());
    if (params?.limit) queryParams.append('limit', params.limit.toString());
    
    const query = queryParams.toString() ? `?${queryParams.toString()}` : '';
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/api/v1/timeline/home${query}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    });
    
    if (!response.ok) {
      throw new Error('タイムラインの取得に失敗しました');
    }
    
    const data = await response.json();
    console.log('ホームタイムラインレスポンス:', data);
    
    // APIレスポンス形式に応じてデータを取得
    return data.posts || data.data?.posts || [];
  } catch (error) {
    console.error('タイムライン取得エラー:', error);
    throw error;
  }
};

/**
 * エクスプローラータイムラインを取得する（すべてのユーザーの人気投稿）
 */
export const getExploreTimeline = async (params?: TimelineParams): Promise<Post[]> => {
  try {
    const queryParams = new URLSearchParams();
    if (params?.page) queryParams.append('page', params.page.toString());
    if (params?.limit) queryParams.append('limit', params.limit.toString());
    
    const query = queryParams.toString() ? `?${queryParams.toString()}` : '';
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/api/v1/timeline/explore${query}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    });
    
    if (!response.ok) {
      throw new Error('エクスプローラータイムラインの取得に失敗しました');
    }
    
    const data = await response.json();
    
    // APIレスポンス形式に応じてデータを取得
    return data.posts || data.data?.posts || [];
  } catch (error) {
    console.error('エクスプローラータイムライン取得エラー:', error);
    throw error;
  }
};

/**
 * 投稿を作成する
 */
export const createPost = async (content: string, parentId?: string): Promise<Post> => {
  try {
    const postData: any = { content };
    if (parentId) {
      postData.parent_id = parentId;
    }
    
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/api/v1/posts`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: JSON.stringify(postData)
    });
    
    if (!response.ok) {
      throw new Error('投稿の作成に失敗しました');
    }
    
    const data = await response.json();
    
    // APIレスポンス形式に応じてデータを取得
    return data.post || data.data?.post;
  } catch (error) {
    console.error('投稿作成エラー:', error);
    throw error;
  }
};

/**
 * 投稿にいいねする
 */
export const likePost = async (postId: string): Promise<void> => {
  try {
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/api/v1/posts/${postId}/like`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    });
    
    if (!response.ok) {
      throw new Error('いいねに失敗しました');
    }
  } catch (error) {
    console.error('いいねエラー:', error);
    throw error;
  }
};

/**
 * いいねを取り消す
 */
export const unlikePost = async (postId: string): Promise<void> => {
  try {
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/api/v1/posts/${postId}/like`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    });
    
    if (!response.ok) {
      throw new Error('いいね取り消しに失敗しました');
    }
  } catch (error) {
    console.error('いいね取り消しエラー:', error);
    throw error;
  }
};

/**
 * 投稿を取得する
 */
export const getPost = async (postId: string): Promise<Post> => {
  try {
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/api/v1/posts/${postId}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    });
    
    if (!response.ok) {
      throw new Error('投稿の取得に失敗しました');
    }
    
    const data = await response.json();
    
    // APIレスポンス形式に応じてデータを取得
    return data.post || data.data?.post;
  } catch (error) {
    console.error('投稿取得エラー:', error);
    throw error;
  }
};

/**
 * 投稿へのリプライを取得する
 */
export const getReplies = async (postId: string, params?: TimelineParams): Promise<Post[]> => {
  try {
    const queryParams = new URLSearchParams();
    if (params?.page) queryParams.append('page', params.page.toString());
    if (params?.limit) queryParams.append('limit', params.limit.toString());
    
    const query = queryParams.toString() ? `?${queryParams.toString()}` : '';
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/api/v1/posts/${postId}/replies${query}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    });
    
    if (!response.ok) {
      throw new Error('返信の取得に失敗しました');
    }
    
    const data = await response.json();
    
    // APIレスポンス形式に応じてデータを取得
    return data.posts || data.data?.posts || [];
  } catch (error) {
    console.error('返信取得エラー:', error);
    throw error;
  }
};

/**
 * 投稿を削除する
 */
export const deletePost = async (postId: string): Promise<void> => {
  try {
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/api/v1/posts/${postId}`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    });
    
    if (!response.ok) {
      throw new Error('投稿の削除に失敗しました');
    }
  } catch (error) {
    console.error('投稿削除エラー:', error);
    throw error;
  }
}; 