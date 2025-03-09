import { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Post } from '../services/postService';
import apiClient from '../api/client';
import ProfileHeader from '../components/profile/ProfileHeader';
import PostCard from '../components/post/PostCard';
import Button from '../components/ui/Button';
import PostForm from '../components/post/PostForm';

// ProfileHeaderコンポーネントと互換性のあるUserインターフェース
interface User {
  id: string;
  username: string;
  display_name: string;
  bio?: string;
  avatar_url: string | null;
  banner_url: string | null;
  created_at: string;
  followers_count: number;
  following_count: number;
  posts_count: number;
  is_following: boolean;
}

const ProfilePage = () => {
  const { username } = useParams<{ username: string }>();
  const navigate = useNavigate();
  const [user, setUser] = useState<User | null>(null);
  const [posts, setPosts] = useState<Post[]>([]);
  const [loading, setLoading] = useState(true);
  const [postsLoading, setPostsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const [showReplyForm, setShowReplyForm] = useState(false);
  const [selectedPost, setSelectedPost] = useState<Post | null>(null);
  
  // ユーザープロフィール取得
  useEffect(() => {
    const fetchProfile = async () => {
      if (!username) return;
      
      setLoading(true);
      setError(null);
      
      try {
        // ユーザープロフィール取得
        const profileResponse = await apiClient.get(`/users/${username}`);
        const userData = profileResponse.data.user || profileResponse.data.data?.user;
        setUser(userData);
        
        // 投稿取得
        const postsResponse = await apiClient.get(`/users/${username}/posts?page=1&limit=20`);
        const postsData = postsResponse.data.posts || postsResponse.data.data?.posts || [];
        setPosts(postsData);
        setPage(1);
        setHasMore(postsData.length === 20);
      } catch (err) {
        console.error('プロフィール取得エラー:', err);
        setError('ユーザープロフィールの取得に失敗しました。');
      } finally {
        setLoading(false);
      }
    };
    
    fetchProfile();
  }, [username]);
  
  // 追加の投稿を読み込む
  const loadMorePosts = async () => {
    if (!username || postsLoading || !hasMore) return;
    
    setPostsLoading(true);
    const nextPage = page + 1;
    
    try {
      const response = await apiClient.get(`/users/${username}/posts?page=${nextPage}&limit=20`);
      const newPosts = response.data.posts || response.data.data?.posts || [];
      setPosts(prev => [...prev, ...newPosts]);
      setPage(nextPage);
      setHasMore(newPosts.length === 20);
    } catch (err) {
      console.error('投稿読み込みエラー:', err);
    } finally {
      setPostsLoading(false);
    }
  };
  
  // リプライ処理
  const handleReply = (post: Post) => {
    setSelectedPost(post);
    setShowReplyForm(true);
  };
  
  // フォロー状態変更後の処理
  const handleFollowChange = async () => {
    if (!username) return;
    
    try {
      // ユーザー情報再取得
      const response = await apiClient.get(`/users/${username}`);
      const updatedUser = response.data.user || response.data.data?.user;
      setUser(updatedUser);
    } catch (err) {
      console.error('ユーザー情報再取得エラー:', err);
    }
  };
  
  // リプライ投稿後の処理
  const handlePostCreated = () => {
    setShowReplyForm(false);
    setSelectedPost(null);
  };
  
  // 戻るボタン
  const handleGoBack = () => {
    navigate(-1);
  };

  return (
    <div className="min-h-screen bg-white dark:bg-gray-900">
      {/* ヘッダー */}
      <header className="sticky top-0 z-10 bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-700">
        <div className="max-w-2xl mx-auto px-4 py-2 flex items-center">
          <button 
            onClick={handleGoBack}
            className="mr-4 p-2 rounded-full hover:bg-gray-100 dark:hover:bg-gray-800"
          >
            <svg className="w-5 h-5 text-gray-600 dark:text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18"></path>
            </svg>
          </button>
          <h1 className="text-xl font-bold text-gray-900 dark:text-white">プロフィール</h1>
        </div>
      </header>
      
      {/* メインコンテンツ */}
      <main className="max-w-2xl mx-auto">
        {loading ? (
          <div className="p-8 text-center text-gray-600 dark:text-gray-300">
            <p>読み込み中...</p>
          </div>
        ) : error || !user ? (
          <div className="p-8 text-center">
            <p className="text-red-500 dark:text-red-400 mb-4">{error || 'ユーザーが見つかりませんでした'}</p>
            <Button onClick={handleGoBack} variant="primary" size="sm">
              戻る
            </Button>
          </div>
        ) : (
          <>
            {/* プロフィールヘッダー */}
            <ProfileHeader user={user} onFollowChange={handleFollowChange} />
            
            {/* リプライフォーム */}
            {showReplyForm && selectedPost && (
              <div className="border-b border-gray-200 dark:border-gray-700 p-4 bg-white dark:bg-gray-900">
                <PostForm 
                  parentId={selectedPost.id}
                  onPostCreated={handlePostCreated}
                  onCancel={() => {
                    setShowReplyForm(false);
                    setSelectedPost(null);
                  }}
                  placeholder="リプライを入力..."
                />
              </div>
            )}
            
            {/* 投稿一覧 */}
            <div className="divide-y divide-gray-200 dark:divide-gray-700">
              {posts.length === 0 ? (
                <div className="p-8 text-center text-gray-500 dark:text-gray-400">
                  <p>投稿がありません</p>
                </div>
              ) : (
                <>
                  {posts.map(post => (
                    <PostCard 
                      key={post.id} 
                      post={post}
                      onReply={handleReply}
                    />
                  ))}
                  
                  {/* もっと読み込むボタン */}
                  {hasMore && (
                    <div className="p-4 text-center">
                      <Button 
                        onClick={loadMorePosts}
                        variant="secondary"
                        size="sm"
                        disabled={postsLoading}
                      >
                        {postsLoading ? '読み込み中...' : 'もっと読み込む'}
                      </Button>
                    </div>
                  )}
                </>
              )}
            </div>
          </>
        )}
      </main>
    </div>
  );
};

export default ProfilePage; 