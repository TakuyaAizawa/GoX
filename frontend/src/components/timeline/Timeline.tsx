import { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { getHomeTimeline, getExploreTimeline, Post } from '../../services/postService';
import PostCard from '../post/PostCard';
import PostForm from '../post/PostForm';
import RealtimeBroadcast from '../broadcast/RealtimeBroadcast';

interface TimelineProps {
  type: 'home' | 'explore';
  showPostForm?: boolean;
}

const Timeline = ({ type, showPostForm = true }: TimelineProps) => {
  const navigate = useNavigate();
  const [posts, setPosts] = useState<Post[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  
  // データ取得関数
  const fetchTimeline = useCallback(async (pageNum = 1, refresh = false) => {
    try {
      setLoading(true);
      setError(null);
      
      const fetchFunction = type === 'home' ? getHomeTimeline : getExploreTimeline;
      const newPosts = await fetchFunction({ page: pageNum, limit: 20 });
      
      if (refresh) {
        setPosts(newPosts);
      } else {
        setPosts(prev => [...prev, ...newPosts]);
      }
      
      // もし20件未満なら最後まで取得したとみなす
      setHasMore(newPosts.length === 20);
    } catch (error) {
      console.error('タイムライン取得エラー:', error);
      setError('タイムラインの取得に失敗しました。再読み込みしてください。');
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, [type]);
  
  // 初期データ読み込み
  useEffect(() => {
    fetchTimeline(1, true);
  }, [fetchTimeline]);
  
  // 次のページを読み込む
  const loadMore = () => {
    if (!loading && hasMore) {
      const nextPage = page + 1;
      setPage(nextPage);
      fetchTimeline(nextPage);
    }
  };
  
  // タイムラインを更新
  const refreshTimeline = () => {
    setRefreshing(true);
    setPage(1);
    fetchTimeline(1, true);
  };
  
  // 投稿作成後のコールバック
  const handlePostCreated = () => {
    refreshTimeline();
  };
  
  // リプライ処理
  const handleReply = (post: Post) => {
    // 投稿詳細ページに遷移
    navigate(`/post/${post.id}`);
  };
  
  // スクロールイベントハンドラ
  const handleScroll = useCallback((e: React.UIEvent<HTMLDivElement>) => {
    const { scrollTop, clientHeight, scrollHeight } = e.currentTarget;
    // 下部から100px以内までスクロールしたら追加読み込み
    if (scrollHeight - scrollTop <= clientHeight + 100 && !loading && hasMore) {
      loadMore();
    }
  }, [loading, hasMore]);
  
  return (
    <div className="flex flex-col h-full">
      {/* 投稿フォーム */}
      {showPostForm && (
        <PostForm onPostCreated={handlePostCreated} />
      )}
      
      {/* リアルタイムブロードキャスト - ホームタイムラインの場合は新規投稿、探索の場合はトレンド */}
      {type === 'home' ? (
        <RealtimeBroadcast type="new_posts" onRefresh={refreshTimeline} />
      ) : (
        <RealtimeBroadcast type="trending" />
      )}
      
      {/* エラーメッセージ */}
      {error && (
        <div className="p-4 text-red-500 text-center">
          <p>{error}</p>
          <button 
            onClick={refreshTimeline}
            className="mt-2 text-primary-500 hover:text-primary-600"
          >
            再読み込み
          </button>
        </div>
      )}
      
      {/* タイムライン */}
      <div 
        className="flex-1 overflow-y-auto" 
        onScroll={handleScroll}
      >
        {/* 更新中インジケーター */}
        {refreshing && (
          <div className="p-4 text-center text-gray-500">
            更新中...
          </div>
        )}
        
        {/* 投稿リスト */}
        {posts.length > 0 ? (
          posts.map(post => (
            <PostCard 
              key={post.id} 
              post={post}
              onReply={() => handleReply(post)}
            />
          ))
        ) : !loading ? (
          <div className="p-8 text-center text-gray-500 dark:text-gray-400">
            {type === 'home' 
              ? '表示する投稿がありません。誰かをフォローするか、新しい投稿を作成してください。' 
              : '表示する投稿がありません。'
            }
          </div>
        ) : null}
        
        {/* ローディングインジケーター */}
        {loading && !refreshing && (
          <div className="p-4 text-center text-gray-500">
            読み込み中...
          </div>
        )}
        
        {/* 最後に到達 */}
        {!hasMore && posts.length > 0 && (
          <div className="p-4 text-center text-gray-500 dark:text-gray-400">
            これ以上の投稿はありません
          </div>
        )}
      </div>
    </div>
  );
};

export default Timeline; 