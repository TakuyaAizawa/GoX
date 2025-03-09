import { useState, useEffect } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { Post } from '../services/postService';
import apiClient from '../api/client';
import PostCard from '../components/post/PostCard';
import PostForm from '../components/post/PostForm';
import Button from '../components/ui/Button';
import { useAuthStore } from '../store/authStore';

const PostPage = () => {
  const { postId } = useParams<{ postId: string }>();
  const navigate = useNavigate();
  const { user } = useAuthStore();
  const [post, setPost] = useState<Post | null>(null);
  const [replies, setReplies] = useState<Post[]>([]);
  const [loading, setLoading] = useState(true);
  const [replyLoading, setReplyLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [replyTo, setReplyTo] = useState<Post | null>(null);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);

  // 投稿とリプライを取得
  useEffect(() => {
    const fetchPostAndReplies = async () => {
      if (!postId) return;
      
      setLoading(true);
      setError(null);
      
      try {
        // 投稿の取得
        const postResponse = await apiClient.get(`/posts/${postId}`);
        const postData = postResponse.data.post || postResponse.data.data?.post;
        setPost(postData);
        
        // リプライの取得
        const repliesResponse = await apiClient.get(`/posts/${postId}/replies?page=1&limit=20`);
        const repliesData = repliesResponse.data.posts || repliesResponse.data.data?.posts || [];
        setReplies(repliesData);
        setHasMore(repliesData.length === 20);
        setPage(1);
      } catch (err) {
        console.error('投稿取得エラー:', err);
        setError('投稿の取得に失敗しました。再読み込みしてください。');
      } finally {
        setLoading(false);
      }
    };
    
    fetchPostAndReplies();
  }, [postId]);
  
  // 追加のリプライを読み込む
  const loadMoreReplies = async () => {
    if (replyLoading || !hasMore || !postId) return;
    
    setReplyLoading(true);
    const nextPage = page + 1;
    
    try {
      const response = await apiClient.get(`/posts/${postId}/replies?page=${nextPage}&limit=20`);
      const newReplies = response.data.posts || response.data.data?.posts || [];
      setReplies(prev => [...prev, ...newReplies]);
      setPage(nextPage);
      setHasMore(newReplies.length === 20);
    } catch (err) {
      console.error('リプライ読み込みエラー:', err);
    } finally {
      setReplyLoading(false);
    }
  };
  
  // リプライ作成フォーム表示
  const handleReply = (parentPost: Post) => {
    setReplyTo(parentPost);
  };
  
  // リプライが投稿された後の処理
  const handlePostCreated = async () => {
    setReplyTo(null);
    
    // リプライ一覧を更新
    if (postId) {
      try {
        const response = await apiClient.get(`/posts/${postId}/replies?page=1&limit=20`);
        const repliesData = response.data.posts || response.data.data?.posts || [];
        setReplies(repliesData);
        setPage(1);
        setHasMore(repliesData.length === 20);
      } catch (err) {
        console.error('リプライ更新エラー:', err);
      }
    }
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
          <h1 className="text-xl font-bold text-gray-900 dark:text-white">投稿</h1>
        </div>
      </header>
      
      {/* メインコンテンツ */}
      <main className="max-w-2xl mx-auto divide-y divide-gray-200 dark:divide-gray-700">
        {loading ? (
          <div className="p-8 text-center text-gray-600 dark:text-gray-300">
            <p>読み込み中...</p>
          </div>
        ) : error || !post ? (
          <div className="p-8 text-center">
            <p className="text-red-500 dark:text-red-400 mb-4">{error || '投稿が見つかりませんでした'}</p>
            <Button onClick={handleGoBack} variant="primary" size="sm">
              戻る
            </Button>
          </div>
        ) : (
          <>
            {/* 親投稿 */}
            <div>
              <PostCard 
                post={post} 
                isDetail={true}
                onReply={handleReply}
              />
            </div>
            
            {/* リプライ作成フォーム */}
            {replyTo && (
              <div className="p-4 border-b border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900">
                <PostForm 
                  parentId={replyTo.id}
                  onPostCreated={handlePostCreated}
                  onCancel={() => setReplyTo(null)}
                  placeholder="リプライを入力..."
                />
              </div>
            )}
            
            {/* リプライ一覧 */}
            <div>
              <h2 className="px-4 py-3 text-lg font-semibold text-gray-900 dark:text-white">
                リプライ
              </h2>
              
              {replies.length === 0 ? (
                <div className="p-8 text-center text-gray-500 dark:text-gray-400">
                  <p>まだリプライはありません</p>
                </div>
              ) : (
                <div>
                  {replies.map(reply => (
                    <PostCard 
                      key={reply.id} 
                      post={reply}
                      onReply={handleReply}
                    />
                  ))}
                  
                  {hasMore && (
                    <div className="p-4 text-center">
                      <Button 
                        onClick={loadMoreReplies}
                        variant="secondary"
                        size="sm"
                        disabled={replyLoading}
                      >
                        {replyLoading ? '読み込み中...' : 'もっと読み込む'}
                      </Button>
                    </div>
                  )}
                </div>
              )}
            </div>
          </>
        )}
      </main>
    </div>
  );
};

export default PostPage; 