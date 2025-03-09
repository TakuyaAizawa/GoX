import { useState } from 'react';
import { Link } from 'react-router-dom';
import { formatDistance } from 'date-fns';
import { ja } from 'date-fns/locale';
import { Post } from '../../services/postService';
import apiClient from '../../api/client';

interface PostCardProps {
  post: Post;
  onLike?: (postId: string) => void;
  onUnlike?: (postId: string) => void;
  onReply?: (post: Post) => void;
  isDetail?: boolean;
}

const PostCard = ({ post, onLike, onUnlike, onReply, isDetail = false }: PostCardProps) => {
  const [isLiked, setIsLiked] = useState(post.is_liked);
  const [likesCount, setLikesCount] = useState(post.likes_count);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 日付をフォーマット
  const formattedDate = formatDistance(new Date(post.created_at), new Date(), {
    addSuffix: true,
    locale: ja
  });

  // いいね処理
  const handleLike = async (e: React.MouseEvent) => {
    e.preventDefault(); // リンクのナビゲーションを防止
    e.stopPropagation(); // イベントの伝播を停止
    
    if (isSubmitting) return;
    
    setIsSubmitting(true);
    setError(null);
    
    try {
      if (isLiked) {
        await apiClient.delete(`/posts/${post.id}/like`);
        setIsLiked(false);
        setLikesCount(prev => Math.max(0, prev - 1)); // マイナスにならないよう保護
        onUnlike?.(post.id);
      } else {
        await apiClient.post(`/posts/${post.id}/like`);
        setIsLiked(true);
        setLikesCount(prev => prev + 1);
        onLike?.(post.id);
      }
    } catch (err) {
      console.error('いいね処理エラー:', err);
      setError('処理に失敗しました');
      // エラー状態を3秒後にクリア
      setTimeout(() => setError(null), 3000);
    } finally {
      setIsSubmitting(false);
    }
  };

  // リプライ処理
  const handleReply = (e: React.MouseEvent) => {
    e.preventDefault(); // リンクのナビゲーションを防止
    e.stopPropagation(); // イベントの伝播を停止
    onReply?.(post);
  };

  // 投稿カード全体をクリックしたときの処理
  const handleCardClick = () => {
    if (!isDetail) {
      // 詳細ページに遷移
      window.location.href = `/post/${post.id}`;
    }
  };

  return (
    <div 
      className={`border-b border-gray-200 dark:border-gray-700 p-4 hover:bg-gray-50 dark:hover:bg-gray-800 transition cursor-pointer ${isDetail ? 'bg-white dark:bg-gray-900' : ''}`}
      onClick={!isDetail ? handleCardClick : undefined}
    >
      <div className="flex space-x-3">
        {/* ユーザーアバター */}
        <Link 
          to={`/profile/${post.user.username}`} 
          className="flex-shrink-0"
          onClick={(e) => e.stopPropagation()} // 親リンクのナビゲーションを防止
        >
          <img
            src={post.user.avatar_url || '/default-avatar.png'}
            alt={post.user.display_name}
            className="h-10 w-10 rounded-full"
          />
        </Link>
        
        <div className="flex-1 min-w-0">
          {/* ユーザー情報と日付 */}
          <div className="flex items-center text-sm space-x-1">
            <Link 
              to={`/profile/${post.user.username}`} 
              className="font-semibold text-gray-900 dark:text-white hover:underline"
              onClick={(e) => e.stopPropagation()} // 親リンクのナビゲーションを防止
            >
              {post.user.display_name}
            </Link>
            <span className="text-gray-500 dark:text-gray-400">@{post.user.username}</span>
            <span className="text-gray-500 dark:text-gray-400">・</span>
            <span title={new Date(post.created_at).toLocaleString('ja-JP')} className="text-gray-500 dark:text-gray-400">
              {formattedDate}
            </span>
          </div>
          
          {/* 投稿内容 */}
          <div className="mt-1 mb-2">
            <div className="text-gray-900 dark:text-white text-sm whitespace-pre-wrap break-words">
              {post.content}
            </div>
          </div>
          
          {/* メディア表示（あれば） */}
          {post.media_urls && post.media_urls.length > 0 && (
            <div className="mt-2 mb-3">
              {/* メディア表示ロジックはここに実装 */}
              <p className="text-blue-500 text-sm">添付メディアあり</p>
            </div>
          )}
          
          {/* エラーメッセージ */}
          {error && (
            <div className="mb-2 text-red-500 text-xs">
              {error}
            </div>
          )}
          
          {/* アクションボタン */}
          <div className="flex mt-2 space-x-8 text-gray-500">
            {/* リプライボタン */}
            <button 
              className="flex items-center text-gray-500 dark:text-gray-400 hover:text-blue-500 dark:hover:text-blue-400"
              onClick={handleReply}
            >
              <svg className="w-5 h-5 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"></path>
              </svg>
              <span className="text-sm">{post.replies_count || 0}</span>
            </button>
            
            {/* いいねボタン */}
            <button
              className={`flex items-center ${isLiked ? 'text-red-500 dark:text-red-400' : 'text-gray-500 dark:text-gray-400'} hover:text-red-500 dark:hover:text-red-400`}
              onClick={handleLike}
              disabled={isSubmitting}
            >
              <svg className="w-5 h-5 mr-1" fill={isLiked ? "currentColor" : "none"} stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.5" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"></path>
              </svg>
              <span className="text-sm">{likesCount || 0}</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default PostCard; 