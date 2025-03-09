import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { getFollowers, getFollowing, followUser, unfollowUser, User as UserType } from '../../services/userService';
import Button from '../ui/Button';
import { useAuthStore } from '../../store/authStore';

interface User {
  id: string;
  username: string;
  display_name: string;
  avatar_url: string | null;
  is_following: boolean;
}

interface FollowListProps {
  username: string;
  type: 'followers' | 'following';
  onClose: () => void;
}

const FollowList: React.FC<FollowListProps> = ({ username, type, onClose }) => {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const { user: currentUser } = useAuthStore();
  
  useEffect(() => {
    const fetchUsers = async () => {
      setLoading(true);
      setError(null);
      
      try {
        let data: UserType[];
        if (type === 'followers') {
          data = await getFollowers(username, { page: 1, limit: 20 });
        } else {
          data = await getFollowing(username, { page: 1, limit: 20 });
        }
        
        // 型変換を行う
        const formattedUsers: User[] = data.map(user => ({
          id: user.id,
          username: user.username,
          display_name: user.display_name,
          avatar_url: user.avatar_url,
          is_following: user.is_following
        }));
        
        setUsers(formattedUsers);
        setHasMore(data.length === 20);
      } catch (error) {
        console.error('ユーザーリスト取得エラー:', error);
        setError('ユーザーリストの取得に失敗しました');
      } finally {
        setLoading(false);
      }
    };
    
    fetchUsers();
  }, [username, type]);
  
  const loadMore = async () => {
    if (loading || !hasMore) return;
    
    setLoading(true);
    
    try {
      const nextPage = page + 1;
      let data: UserType[];
      
      if (type === 'followers') {
        data = await getFollowers(username, { page: nextPage, limit: 20 });
      } else {
        data = await getFollowing(username, { page: nextPage, limit: 20 });
      }
      
      if (data.length > 0) {
        // 型変換を行う
        const formattedUsers: User[] = data.map(user => ({
          id: user.id,
          username: user.username,
          display_name: user.display_name,
          avatar_url: user.avatar_url,
          is_following: user.is_following
        }));
        
        setUsers(prev => [...prev, ...formattedUsers]);
        setPage(nextPage);
      }
      
      setHasMore(data.length === 20);
    } catch (error) {
      console.error('追加ユーザーリスト取得エラー:', error);
    } finally {
      setLoading(false);
    }
  };
  
  const handleToggleFollow = async (userId: string, isFollowing: boolean) => {
    try {
      if (isFollowing) {
        await unfollowUser(userId);
      } else {
        await followUser(userId);
      }
      
      // ユーザーリスト内のフォロー状態を更新
      setUsers(prev => 
        prev.map(user => 
          user.id === userId 
            ? { ...user, is_following: !isFollowing } 
            : user
        )
      );
    } catch (error) {
      console.error('フォロー操作エラー:', error);
    }
  };
  
  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-800 rounded-lg w-full max-w-md max-h-[80vh] flex flex-col">
        {/* ヘッダー */}
        <div className="p-4 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
            {type === 'followers' ? 'フォロワー' : 'フォロー中'}
          </h2>
          <button
            onClick={onClose}
            className="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
          >
            ✕
          </button>
        </div>
        
        {/* ユーザーリスト */}
        <div className="overflow-y-auto flex-grow">
          {loading && users.length === 0 ? (
            <div className="p-4 text-center text-gray-500 dark:text-gray-400">
              読み込み中...
            </div>
          ) : error ? (
            <div className="p-4 text-center text-red-500">
              {error}
            </div>
          ) : users.length === 0 ? (
            <div className="p-4 text-center text-gray-500 dark:text-gray-400">
              {type === 'followers' ? 'フォロワーがいません' : 'フォロー中のユーザーがいません'}
            </div>
          ) : (
            <ul>
              {users.map(user => (
                <li 
                  key={user.id}
                  className="border-b border-gray-200 dark:border-gray-700 last:border-b-0"
                >
                  <div className="p-4 flex items-center justify-between">
                    <Link 
                      to={`/profile/${user.username}`}
                      className="flex items-center flex-grow"
                      onClick={onClose}
                    >
                      <img
                        src={user.avatar_url || '/default-avatar.png'}
                        alt={`${user.display_name}のアバター`}
                        className="w-10 h-10 rounded-full object-cover mr-3"
                      />
                      <div>
                        <p className="font-semibold text-gray-900 dark:text-white">
                          {user.display_name}
                        </p>
                        <p className="text-sm text-gray-500 dark:text-gray-400">
                          @{user.username}
                        </p>
                      </div>
                    </Link>
                    
                    {/* 自分自身以外にフォローボタンを表示 */}
                    {currentUser && currentUser.id !== user.id && (
                      <Button
                        variant={user.is_following ? 'outline' : 'primary'}
                        size="sm"
                        onClick={() => handleToggleFollow(user.id, user.is_following)}
                      >
                        {user.is_following ? 'フォロー中' : 'フォロー'}
                      </Button>
                    )}
                  </div>
                </li>
              ))}
            </ul>
          )}
          
          {/* もっと読み込むボタン */}
          {hasMore && (
            <div className="p-4 text-center">
              <Button
                variant="outline"
                onClick={loadMore}
                isLoading={loading && users.length > 0}
                disabled={loading}
              >
                もっと読み込む
              </Button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default FollowList; 