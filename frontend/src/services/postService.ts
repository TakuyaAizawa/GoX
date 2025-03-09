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
 * ホームタイムライン（フォロー中のユーザーの投稿）を取得する
 */
export const getHomeTimeline = async (params?: TimelineParams): Promise<Post[]> => {
  try {
    const queryParams = new URLSearchParams();
    if (params?.page) queryParams.append('page', params.page.toString());
    if (params?.limit) queryParams.append('limit', params.limit.toString());
    
    const query = queryParams.toString() ? `?${queryParams.toString()}` : '';
    const response = await apiClient.get(`timeline/home${query}`);
    
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
    console.error('ホームタイムライン取得エラー:', error);
    throw error;
  }
};

/**
 * エクスプローラータイムライン（人気/トレンド投稿）を取得する
 */
export const getExploreTimeline = async (params?: TimelineParams): Promise<Post[]> => {
  try {
    const queryParams = new URLSearchParams();
    if (params?.page) queryParams.append('page', params.page.toString());
    if (params?.limit) queryParams.append('limit', params.limit.toString());
    
    const query = queryParams.toString() ? `?${queryParams.toString()}` : '';
    const response = await apiClient.get(`timeline/explore${query}`);
    
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
    console.error('エクスプローラータイムライン取得エラー:', error);
    throw error;
  }
};

/**
 * 新しい投稿を作成する
 */
export const createPost = async (content: string, parentId?: string): Promise<Post> => {
  try {
    const postData: { content: string; parent_id?: string } = { content };
    if (parentId) {
      postData.parent_id = parentId;
    }
    
    const response = await apiClient.post('posts', postData);
    
    // APIレスポンス形式に応じてデータを取得
    let post = null;
    
    if (response.data.post) {
      post = response.data.post;
    } else if (response.data.data && response.data.data.post) {
      post = response.data.data.post;
    } else if (response.data.success === true && response.data.data) {
      post = response.data.data;
    }
    
    if (!post) {
      console.error('投稿データが応答に見つかりません:', response.data);
      throw new Error('投稿の作成に失敗しました');
    }
    
    return post;
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
    await apiClient.post(`posts/${postId}/like`);
  } catch (error) {
    console.error('いいねエラー:', error);
    throw error;
  }
};

/**
 * 投稿のいいねを解除する
 */
export const unlikePost = async (postId: string): Promise<void> => {
  try {
    await apiClient.delete(`posts/${postId}/like`);
  } catch (error) {
    console.error('いいね解除エラー:', error);
    throw error;
  }
};

/**
 * 投稿を取得する
 */
export const getPost = async (postId: string): Promise<Post> => {
  try {
    const response = await apiClient.get(`posts/${postId}`);
    
    // APIレスポンス形式に応じてデータを取得
    let post = null;
    
    if (response.data.post) {
      post = response.data.post;
    } else if (response.data.data && response.data.data.post) {
      post = response.data.data.post;
    } else if (response.data.success === true && response.data.data) {
      post = response.data.data;
    }
    
    if (!post) {
      console.error('投稿データが応答に見つかりません:', response.data);
      throw new Error('投稿の取得に失敗しました');
    }
    
    return post;
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
    const response = await apiClient.get(`posts/${postId}/replies${query}`);
    
    // APIレスポンス形式に応じてデータを取得
    let replies = null;
    
    if (Array.isArray(response.data)) {
      replies = response.data;
    } else if (Array.isArray(response.data.replies)) {
      replies = response.data.replies;
    } else if (response.data.data && Array.isArray(response.data.data.replies)) {
      replies = response.data.data.replies;
    } else if (response.data.success === true && Array.isArray(response.data.data)) {
      replies = response.data.data;
    }
    
    if (!replies) {
      console.error('リプライデータが応答に見つかりません:', response.data);
      return [];
    }
    
    return replies;
  } catch (error) {
    console.error('リプライ取得エラー:', error);
    throw error;
  }
};

/**
 * 投稿を削除する
 */
export const deletePost = async (postId: string): Promise<void> => {
  try {
    await apiClient.delete(`posts/${postId}`);
  } catch (error) {
    console.error('投稿削除エラー:', error);
    throw error;
  }
}; 