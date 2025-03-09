import { useState, useEffect, useRef } from 'react';
import { useWebSocketContext } from '../../context/WebSocketContext';
import { Post } from '../../services/postService';
import { useAuthStore } from '../../store/authStore';
import { useNavigate } from 'react-router-dom';

interface RealtimeBroadcastProps {
  type: 'new_posts' | 'trending' | 'following_activity';
  onRefresh?: () => void;
}

const RealtimeBroadcast: React.FC<RealtimeBroadcastProps> = ({ type, onRefresh }) => {
  const { addMessageHandler, isConnected } = useWebSocketContext();
  const { user } = useAuthStore();
  const navigate = useNavigate();
  
  // 状態管理
  const [newPosts, setNewPosts] = useState<Post[]>([]);
  const [newPostsCount, setNewPostsCount] = useState(0);
  const [trendingPosts, setTrendingPosts] = useState<Post[]>([]);
  const [showBroadcast, setShowBroadcast] = useState(false);
  const [expanded, setExpanded] = useState(false);
  
  // タイマー参照
  const autoHideTimerRef = useRef<number | null>(null);
  
  // WebSocketメッセージのリスニング
  useEffect(() => {
    if (!isConnected) return;
    
    // 新しい投稿のリスニング
    const removeNewPostsHandler = addMessageHandler('new_post', (data) => {
      if (user && data.user_id !== user.id) { // 自分の投稿は除外
        setNewPosts(prev => [data, ...prev].slice(0, 5)); // 最大5件まで保持
        setNewPostsCount(prev => prev + 1);
        setShowBroadcast(true);
        
        // 自動的に非表示にするタイマーをセット
        if (autoHideTimerRef.current) {
          clearTimeout(autoHideTimerRef.current);
        }
        autoHideTimerRef.current = window.setTimeout(() => {
          setShowBroadcast(false);
          setExpanded(false);
        }, 10000); // 10秒後に非表示
      }
    });
    
    // トレンド投稿のリスニング
    const removeTrendingHandler = addMessageHandler('trending_post', (data) => {
      if (type === 'trending') {
        setTrendingPosts(prev => {
          // すでに存在する場合は追加しない
          if (prev.some(post => post.id === data.id)) {
            return prev;
          }
          return [data, ...prev].slice(0, 3); // 最大3件まで保持
        });
        setShowBroadcast(true);
        
        // 自動的に非表示にするタイマーをセット
        if (autoHideTimerRef.current) {
          clearTimeout(autoHideTimerRef.current);
        }
        autoHideTimerRef.current = window.setTimeout(() => {
          setShowBroadcast(false);
          setExpanded(false);
        }, 10000); // 10秒後に非表示
      }
    });
    
    return () => {
      removeNewPostsHandler();
      removeTrendingHandler();
      
      // タイマーをクリア
      if (autoHideTimerRef.current) {
        clearTimeout(autoHideTimerRef.current);
      }
    };
  }, [isConnected, addMessageHandler, type, user]);
  
  // 表示するメッセージを取得
  const getMessage = () => {
    switch (type) {
      case 'new_posts':
        return `${newPostsCount}件の新しい投稿`;
      case 'trending':
        return '人気の投稿があります';
      case 'following_activity':
        return 'フォロー中のユーザーの新しいアクティビティ';
      default:
        return '新しいコンテンツがあります';
    }
  };
  
  // 新しい投稿を確認する
  const handleViewNewPosts = () => {
    if (onRefresh) {
      onRefresh();
    }
    setNewPostsCount(0);
    setNewPosts([]);
    setShowBroadcast(false);
    setExpanded(false);
  };
  
  // 特定の投稿に移動する
  const handleGoToPost = (postId: string) => {
    navigate(`/post/${postId}`);
    setShowBroadcast(false);
    setExpanded(false);
  };
  
  // 表示を切り替える
  const toggleExpanded = () => {
    setExpanded(!expanded);
    
    // 拡張表示した場合、自動非表示タイマーをキャンセル
    if (!expanded && autoHideTimerRef.current) {
      clearTimeout(autoHideTimerRef.current);
      autoHideTimerRef.current = null;
    }
  };
  
  // 何も表示するものがない場合は何も描画しない
  if (!showBroadcast) return null;
  
  return (
    <div className="px-4 py-2 mb-4">
      <div className="bg-blue-100 dark:bg-blue-900/30 rounded-lg overflow-hidden transition-all duration-300">
        {/* ヘッダー部分 */}
        <button
          className="w-full px-4 py-3 flex items-center justify-between text-blue-700 dark:text-blue-300"
          onClick={toggleExpanded}
        >
          <div className="flex items-center">
            <span className="mr-2">
              {type === 'new_posts' ? '🔔' : type === 'trending' ? '🔥' : '👥'}
            </span>
            <span className="font-medium">{getMessage()}</span>
          </div>
          <div className="flex items-center">
            {type === 'new_posts' && (
              <button
                className="mr-2 text-sm bg-blue-500 hover:bg-blue-600 text-white px-3 py-1 rounded-full"
                onClick={(e) => {
                  e.stopPropagation();
                  handleViewNewPosts();
                }}
              >
                表示
              </button>
            )}
            <span className="transform transition-transform duration-200 block">
              {expanded ? (
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 15l7-7 7 7" />
                </svg>
              ) : (
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                </svg>
              )}
            </span>
          </div>
        </button>
        
        {/* 展開時に表示するコンテンツ */}
        {expanded && (
          <div className="px-4 pb-3 border-t border-blue-200 dark:border-blue-700/50">
            {(type === 'new_posts' && newPosts.length > 0) && (
              <ul className="mt-2 space-y-2">
                {newPosts.map(post => (
                  <li key={post.id}>
                    <button
                      className="w-full text-left p-2 rounded hover:bg-blue-200 dark:hover:bg-blue-800/50 flex items-start"
                      onClick={() => handleGoToPost(post.id)}
                    >
                      <img
                        src={post.user.avatar_url || '/default-avatar.png'}
                        alt={post.user.display_name}
                        className="w-8 h-8 rounded-full mr-2 flex-shrink-0"
                      />
                      <div>
                        <p className="font-medium text-blue-800 dark:text-blue-200">{post.user.display_name}</p>
                        <p className="text-sm text-blue-700 dark:text-blue-300 line-clamp-2">{post.content}</p>
                      </div>
                    </button>
                  </li>
                ))}
              </ul>
            )}
            
            {(type === 'trending' && trendingPosts.length > 0) && (
              <ul className="mt-2 space-y-2">
                {trendingPosts.map(post => (
                  <li key={post.id}>
                    <button
                      className="w-full text-left p-2 rounded hover:bg-blue-200 dark:hover:bg-blue-800/50 flex items-start"
                      onClick={() => handleGoToPost(post.id)}
                    >
                      <img
                        src={post.user.avatar_url || '/default-avatar.png'}
                        alt={post.user.display_name}
                        className="w-8 h-8 rounded-full mr-2 flex-shrink-0"
                      />
                      <div className="flex-1">
                        <div className="flex justify-between">
                          <p className="font-medium text-blue-800 dark:text-blue-200">{post.user.display_name}</p>
                          <span className="text-xs text-blue-600 dark:text-blue-400 flex items-center">
                            <span className="mr-1">❤️</span> {post.likes_count || 0}
                          </span>
                        </div>
                        <p className="text-sm text-blue-700 dark:text-blue-300 line-clamp-2">{post.content}</p>
                      </div>
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </div>
        )}
      </div>
    </div>
  );
};

export default RealtimeBroadcast; 