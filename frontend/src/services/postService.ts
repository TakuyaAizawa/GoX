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
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/timeline/home${query}`, {
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
    console.log('ホームタイムラインデータ:', data);
    
    // APIレスポンス形式に応じてデータを取得
    return data.posts || data.data?.posts || [];
  } catch (error) {
    console.error('ホームタイムライン取得エラー:', error);
    throw error;
  }
};

/**
 * エクスプローラータイムラインを取得する（人気/最新の投稿）
 */
export const getExploreTimeline = async (params?: TimelineParams): Promise<Post[]> => {
  try {
    const queryParams = new URLSearchParams();
    if (params?.page) queryParams.append('page', params.page.toString());
    if (params?.limit) queryParams.append('limit', params.limit.toString());
    
    const query = queryParams.toString() ? `?${queryParams.toString()}` : '';
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/timeline/explore${query}`, {
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
    console.log('エクスプローラータイムラインデータ:', data);
    
    // APIレスポンス形式に応じてデータを取得
    return data.posts || data.data?.posts || [];
  } catch (error) {
    console.error('エクスプローラータイムライン取得エラー:', error);
    throw error;
  }
};

/**
 * 新しい投稿を作成する
 */
export const createPost = async (content: string, parentId?: string): Promise<Post> => {
  try {
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/posts`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: JSON.stringify({
        content,
        parent_id: parentId
      })
    });
    
    if (!response.ok) {
      throw new Error('投稿の作成に失敗しました');
    }
    
    const data = await response.json();
    console.log('投稿作成レスポンス:', data);
    
    // APIレスポンス形式に応じてデータを取得
    return data.post || data.data?.post;
  } catch (error) {
    console.error('投稿作成エラー:', error);
    throw error;
  }
};

/**
 * 指定した投稿にいいねをする
 */
export const likePost = async (postId: string): Promise<void> => {
  try {
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/posts/${postId}/like`, {
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
 * 指定した投稿のいいねを取り消す
 */
export const unlikePost = async (postId: string): Promise<void> => {
  try {
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/posts/${postId}/like`, {
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
 * 指定したIDの投稿を取得する
 */
export const getPost = async (postId: string): Promise<Post> => {
  try {
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/posts/${postId}`, {
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
 * 指定した投稿へのリプライを取得する
 */
export const getReplies = async (postId: string, params?: TimelineParams): Promise<Post[]> => {
  try {
    const queryParams = new URLSearchParams();
    if (params?.page) queryParams.append('page', params.page.toString());
    if (params?.limit) queryParams.append('limit', params.limit.toString());
    
    const query = queryParams.toString() ? `?${queryParams.toString()}` : '';
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/posts/${postId}/replies${query}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    });
    
    if (!response.ok) {
      throw new Error('リプライの取得に失敗しました');
    }
    
    const data = await response.json();
    
    // APIレスポンス形式に応じてデータを取得
    return data.posts || data.data?.posts || [];
  } catch (error) {
    console.error('リプライ取得エラー:', error);
    throw error;
  }
};

/**
 * 指定した投稿を削除する
 */
export const deletePost = async (postId: string): Promise<void> => {
  try {
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/posts/${postId}`, {
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