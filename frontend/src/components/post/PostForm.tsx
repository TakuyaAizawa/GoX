import { useState } from 'react';
import { useAuthStore } from '../../store/authStore';
import Button from '../ui/Button';
import { createPost } from '../../services/postService';

interface PostFormProps {
  onPostCreated?: () => void;
  onCancel?: () => void;
  parentId?: string;
  placeholder?: string;
}

const MAX_CONTENT_LENGTH = 280; // 最大文字数

const PostForm = ({ onPostCreated, onCancel, parentId, placeholder = 'いまどうしてる？' }: PostFormProps) => {
  const { user } = useAuthStore();
  const [content, setContent] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  // 文字数
  const contentLength = content.length;
  const remainingChars = MAX_CONTENT_LENGTH - contentLength;
  const isContentValid = contentLength > 0 && contentLength <= MAX_CONTENT_LENGTH;
  
  // 投稿処理
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!isContentValid || isSubmitting) return;
    
    setIsSubmitting(true);
    setError(null);
    
    try {
      await createPost(content, parentId);
      setContent(''); // フォームをクリア
      onPostCreated?.(); // コールバック実行
    } catch (error) {
      console.error('投稿エラー:', error);
      setError('投稿に失敗しました。もう一度お試しください。');
    } finally {
      setIsSubmitting(false);
    }
  };

  // キャンセル処理
  const handleCancel = () => {
    onCancel?.();
  };
  
  return (
    <div className="border-b border-gray-200 dark:border-gray-700 p-4">
      <form onSubmit={handleSubmit}>
        <div className="flex">
          {/* ユーザーアバター */}
          <div className="flex-shrink-0 mr-3">
            <img
              src={user?.avatar_url || '/default-avatar.png'}
              alt={user?.display_name || 'ユーザー'}
              className="h-10 w-10 rounded-full"
            />
          </div>
          
          {/* 投稿入力エリア */}
          <div className="flex-1">
            <textarea
              value={content}
              onChange={(e) => setContent(e.target.value)}
              placeholder={placeholder}
              className="w-full p-2 border-none focus:ring-0 resize-none bg-transparent text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400"
              rows={3}
              maxLength={MAX_CONTENT_LENGTH}
            />
            
            {error && (
              <p className="text-red-500 text-sm mt-2">{error}</p>
            )}
            
            <div className="flex items-center justify-between mt-3">
              <div className={`${remainingChars < 0 ? 'text-red-500' : remainingChars <= 20 ? 'text-yellow-500' : 'text-gray-500'}`}>
                {remainingChars}
              </div>
              <div className="flex space-x-2">
                {onCancel && (
                  <Button
                    type="button"
                    variant="secondary"
                    size="sm"
                    onClick={handleCancel}
                    disabled={isSubmitting}
                  >
                    キャンセル
                  </Button>
                )}
                <Button
                  type="submit"
                  variant="primary"
                  size="sm"
                  disabled={!isContentValid || isSubmitting}
                >
                  {isSubmitting ? '送信中...' : '投稿する'}
                </Button>
              </div>
            </div>
          </div>
        </div>
      </form>
    </div>
  );
};

export default PostForm; 