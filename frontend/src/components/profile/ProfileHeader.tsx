import { useState } from 'react';
import { Link } from 'react-router-dom';
import { formatDistance } from 'date-fns';
import { ja } from 'date-fns/locale';
import { followUser, unfollowUser } from '../../services/userService';
import Button from '../ui/Button';
import { useAuthStore } from '../../store/authStore';
import FollowList from './FollowList';

// ユーザーの型定義
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

interface ProfileHeaderProps {
  user: User;
  onFollowChange?: () => void;
}

const ProfileHeader: React.FC<ProfileHeaderProps> = ({ user, onFollowChange }) => {
  const { user: currentUser } = useAuthStore();
  const [isFollowing, setIsFollowing] = useState(user.is_following);
  const [followersCount, setFollowersCount] = useState(user.followers_count);
  const [isLoading, setIsLoading] = useState(false);
  const [showFollowers, setShowFollowers] = useState(false);
  const [showFollowing, setShowFollowing] = useState(false);
  
  // 自分自身のプロフィールかどうか
  const isOwnProfile = currentUser?.id === user.id;
  
  // 日付のフォーマット
  const formattedJoinDate = formatDistance(
    new Date(user.created_at),
    new Date(),
    { addSuffix: true, locale: ja }
  );
  
  // フォロー/フォロー解除の処理
  const handleToggleFollow = async () => {
    if (isLoading) return;
    
    setIsLoading(true);
    
    try {
      if (isFollowing) {
        await unfollowUser(user.id);
        setFollowersCount(prev => prev - 1);
      } else {
        await followUser(user.id);
        setFollowersCount(prev => prev + 1);
      }
      
      setIsFollowing(!isFollowing);
      
      // 親コンポーネントに変更を通知
      if (onFollowChange) {
        onFollowChange();
      }
    } catch (error) {
      console.error('フォロー操作エラー:', error);
    } finally {
      setIsLoading(false);
    }
  };
  
  return (
    <div className="bg-white dark:bg-gray-800 shadow rounded-t-lg overflow-hidden">
      {/* バナー */}
      <div className="h-32 sm:h-48 relative bg-gray-200 dark:bg-gray-700">
        {user.banner_url && (
          <img
            src={user.banner_url}
            alt="プロフィールバナー"
            className="w-full h-full object-cover"
          />
        )}
      </div>
      
      {/* プロフィール情報 */}
      <div className="px-4 pb-4 relative">
        {/* アバター */}
        <div className="absolute -top-16 left-4">
          <div className="w-24 h-24 rounded-full border-4 border-white dark:border-gray-800 overflow-hidden bg-gray-300 dark:bg-gray-600">
            <img
              src={user.avatar_url || '/default-avatar.png'}
              alt={`${user.display_name}のアバター`}
              className="w-full h-full object-cover"
            />
          </div>
        </div>
        
        {/* アクションボタン */}
        <div className="flex justify-end pt-2 mb-4">
          {isOwnProfile ? (
            <Link to="/settings/profile">
              <Button
                variant="outline"
                size="sm"
              >
                プロフィールを編集
              </Button>
            </Link>
          ) : (
            <Button
              variant={isFollowing ? 'outline' : 'primary'}
              size="sm"
              onClick={handleToggleFollow}
              isLoading={isLoading}
              disabled={isLoading}
            >
              {isFollowing ? 'フォロー中' : 'フォロー'}
            </Button>
          )}
        </div>
        
        {/* ユーザー情報 */}
        <div className="mt-3">
          <h1 className="text-xl font-bold text-gray-900 dark:text-white">
            {user.display_name}
          </h1>
          <p className="text-gray-500 dark:text-gray-400">
            @{user.username}
          </p>
          
          {user.bio && (
            <p className="mt-2 text-gray-700 dark:text-gray-300">
              {user.bio}
            </p>
          )}
          
          <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">
            <span className="mr-4">
              {formattedJoinDate}に登録
            </span>
          </p>
          
          {/* フォローカウント */}
          <div className="mt-3 flex space-x-4 text-sm">
            <button
              onClick={() => setShowFollowing(true)}
              className="text-gray-700 dark:text-gray-300 hover:underline"
            >
              <span className="font-semibold">{user.following_count}</span>
              <span className="ml-1 text-gray-500 dark:text-gray-400">フォロー中</span>
            </button>
            
            <button
              onClick={() => setShowFollowers(true)}
              className="text-gray-700 dark:text-gray-300 hover:underline"
            >
              <span className="font-semibold">{followersCount}</span>
              <span className="ml-1 text-gray-500 dark:text-gray-400">フォロワー</span>
            </button>
            
            <div className="text-gray-700 dark:text-gray-300">
              <span className="font-semibold">{user.posts_count}</span>
              <span className="ml-1 text-gray-500 dark:text-gray-400">投稿</span>
            </div>
          </div>
        </div>
      </div>
      
      {/* フォロワーリスト モーダル */}
      {showFollowers && (
        <FollowList
          username={user.username}
          type="followers"
          onClose={() => setShowFollowers(false)}
        />
      )}
      
      {/* フォロー中リスト モーダル */}
      {showFollowing && (
        <FollowList
          username={user.username}
          type="following"
          onClose={() => setShowFollowing(false)}
        />
      )}
    </div>
  );
};

export default ProfileHeader; 