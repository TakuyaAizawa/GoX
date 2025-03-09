import { useState, useEffect } from 'react';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import apiClient from '../../api/client';
import Button from '../ui/Button';
import { useAuthStore } from '../../store/authStore';

// バリデーションスキーマ
const profileSchema = z.object({
  display_name: z.string().min(1, '表示名は必須です').max(50, '表示名は50文字以内で入力してください'),
  bio: z.string().max(160, 'プロフィールは160文字以内で入力してください').optional(),
});

type ProfileFormData = z.infer<typeof profileSchema>;

interface User {
  id: string;
  username: string;
  display_name: string;
  bio?: string;
  avatar_url: string | null;
  banner_url: string | null;
}

interface EditProfileFormProps {
  user: User;
  onSuccess?: () => void;
  onCancel?: () => void;
}

const EditProfileForm: React.FC<EditProfileFormProps> = ({ user, onSuccess, onCancel }) => {
  const { user: authUser, setUser: setAuthUser } = useAuthStore();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  // アバターとバナーのアップロード状態
  const [avatar, setAvatar] = useState<File | null>(null);
  const [banner, setBanner] = useState<File | null>(null);
  const [avatarPreview, setAvatarPreview] = useState<string | null>(user.avatar_url);
  const [bannerPreview, setBannerPreview] = useState<string | null>(user.banner_url);
  const [isUploadingAvatar, setIsUploadingAvatar] = useState(false);
  const [isUploadingBanner, setIsUploadingBanner] = useState(false);
  
  // フォームの初期化
  const { control, handleSubmit, formState: { errors } } = useForm<ProfileFormData>({
    resolver: zodResolver(profileSchema),
    defaultValues: {
      display_name: user.display_name || '',
      bio: user.bio || '',
    },
  });

  // アバターファイル選択ハンドラー
  const handleAvatarChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      // 2MB以下のファイルサイズチェック
      if (file.size > 2 * 1024 * 1024) {
        setError('アバター画像は2MB以下のサイズにしてください');
        return;
      }
      
      // JPG、PNG、GIF形式のみ許可
      const validTypes = ['image/jpeg', 'image/png', 'image/gif'];
      if (!validTypes.includes(file.type)) {
        setError('JPG、PNG、GIF形式の画像のみ使用できます');
        return;
      }
      
      setAvatar(file);
      setAvatarPreview(URL.createObjectURL(file));
      setError(null);
    }
  };

  // バナーファイル選択ハンドラー
  const handleBannerChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      // 5MB以下のファイルサイズチェック
      if (file.size > 5 * 1024 * 1024) {
        setError('バナー画像は5MB以下のサイズにしてください');
        return;
      }
      
      // JPG、PNG、GIF形式のみ許可
      const validTypes = ['image/jpeg', 'image/png', 'image/gif'];
      if (!validTypes.includes(file.type)) {
        setError('JPG、PNG、GIF形式の画像のみ使用できます');
        return;
      }
      
      setBanner(file);
      setBannerPreview(URL.createObjectURL(file));
      setError(null);
    }
  };

  // アバターアップロード
  const uploadAvatar = async () => {
    if (!avatar) return null;
    
    setIsUploadingAvatar(true);
    const formData = new FormData();
    formData.append('avatar', avatar);
    
    try {
      const response = await apiClient.post('/users/me/avatar', formData, {
        headers: {
          'Content-Type': 'multipart/form-data'
        }
      });
      
      const avatarUrl = response.data.data?.avatar_url || response.data.avatar_url;
      setIsUploadingAvatar(false);
      return avatarUrl;
    } catch (err) {
      console.error('アバターのアップロードに失敗しました', err);
      setError('アバター画像のアップロードに失敗しました');
      setIsUploadingAvatar(false);
      return null;
    }
  };

  // バナーアップロード
  const uploadBanner = async () => {
    if (!banner) return null;
    
    setIsUploadingBanner(true);
    const formData = new FormData();
    formData.append('banner', banner);
    
    try {
      const response = await apiClient.post('/users/me/banner', formData, {
        headers: {
          'Content-Type': 'multipart/form-data'
        }
      });
      
      const bannerUrl = response.data.data?.banner_url || response.data.banner_url;
      setIsUploadingBanner(false);
      return bannerUrl;
    } catch (err) {
      console.error('バナーのアップロードに失敗しました', err);
      setError('バナー画像のアップロードに失敗しました');
      setIsUploadingBanner(false);
      return null;
    }
  };

  // プロフィール情報の保存
  const saveProfile = async (data: ProfileFormData) => {
    try {
      const response = await apiClient.put('/users/me', data);
      return response.data.data?.user || response.data.user;
    } catch (err) {
      console.error('プロフィールの更新に失敗しました', err);
      setError('プロフィール情報の更新に失敗しました');
      return null;
    }
  };

  // フォーム送信
  const onSubmit = async (data: ProfileFormData) => {
    setIsSubmitting(true);
    setError(null);
    
    try {
      // 両方のアップロードと情報保存を並行して実行
      const [avatarUrl, bannerUrl, updatedUser] = await Promise.all([
        avatar ? uploadAvatar() : null,
        banner ? uploadBanner() : null,
        saveProfile(data)
      ]);
      
      if (updatedUser) {
        // 成功した場合は更新されたユーザー情報を反映
        const newUserData = {
          ...updatedUser,
          // アップロードに成功した場合は新しいURL、それ以外は現在の値を維持
          avatar_url: avatarUrl || updatedUser.avatar_url,
          banner_url: bannerUrl || updatedUser.banner_url
        };
        
        // 認証ユーザーデータを更新
        if (authUser && authUser.id === user.id) {
          setAuthUser({
            ...authUser,
            display_name: newUserData.display_name,
            bio: newUserData.bio,
            avatar_url: newUserData.avatar_url,
            banner_url: newUserData.banner_url
          });
        }
        
        // 成功コールバック
        if (onSuccess) {
          onSuccess();
        }
      }
    } catch (err) {
      console.error('プロフィール更新中にエラーが発生しました', err);
      setError('プロフィールの更新に失敗しました。もう一度お試しください。');
    } finally {
      setIsSubmitting(false);
    }
  };

  // コンポーネントのアンマウント時にオブジェクトURLをリリース
  useEffect(() => {
    return () => {
      if (avatarPreview && avatarPreview !== user.avatar_url) {
        URL.revokeObjectURL(avatarPreview);
      }
      if (bannerPreview && bannerPreview !== user.banner_url) {
        URL.revokeObjectURL(bannerPreview);
      }
    };
  }, [avatarPreview, bannerPreview, user.avatar_url, user.banner_url]);

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 md:p-6">
      <h2 className="text-xl font-bold mb-4 text-gray-900 dark:text-white">プロフィール編集</h2>
      
      {error && (
        <div className="bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 p-3 rounded-md mb-4">
          {error}
        </div>
      )}
      
      <div className="mb-6">
        <h3 className="text-lg font-medium mb-2 text-gray-800 dark:text-gray-200">プロフィール画像</h3>
        
        <div className="flex flex-col space-y-4 md:flex-row md:space-y-0 md:space-x-4">
          {/* アバター編集セクション */}
          <div className="flex-1">
            <div className="mb-2 text-sm font-medium text-gray-700 dark:text-gray-300">アバター</div>
            <div className="flex items-center mb-3">
              <div className="w-24 h-24 rounded-full overflow-hidden bg-gray-200 dark:bg-gray-700 mr-4">
                {avatarPreview ? (
                  <img 
                    src={avatarPreview} 
                    alt="アバタープレビュー" 
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <div className="w-full h-full flex items-center justify-center text-gray-400">
                    画像なし
                  </div>
                )}
              </div>
              <div>
                <label className="block">
                  <Button
                    type="button"
                    variant="outline"
                    className="relative mb-2"
                    disabled={isUploadingAvatar}
                  >
                    {isUploadingAvatar ? 'アップロード中...' : '画像を選択'}
                    <input
                      type="file"
                      className="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
                      accept="image/jpeg,image/png,image/gif"
                      onChange={handleAvatarChange}
                      disabled={isUploadingAvatar}
                    />
                  </Button>
                </label>
                <p className="text-xs text-gray-500 dark:text-gray-400">
                  JPG、PNG、GIF、2MB以下
                </p>
              </div>
            </div>
          </div>
          
          {/* バナー編集セクション */}
          <div className="flex-1">
            <div className="mb-2 text-sm font-medium text-gray-700 dark:text-gray-300">バナー</div>
            <div className="mb-3">
              <div className="w-full h-32 rounded-md overflow-hidden bg-gray-200 dark:bg-gray-700 mb-3">
                {bannerPreview ? (
                  <img 
                    src={bannerPreview} 
                    alt="バナープレビュー" 
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <div className="w-full h-full flex items-center justify-center text-gray-400">
                    バナー画像なし
                  </div>
                )}
              </div>
              <div>
                <label className="block">
                  <Button
                    type="button"
                    variant="outline"
                    className="relative mb-2"
                    disabled={isUploadingBanner}
                  >
                    {isUploadingBanner ? 'アップロード中...' : '画像を選択'}
                    <input
                      type="file"
                      className="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
                      accept="image/jpeg,image/png,image/gif"
                      onChange={handleBannerChange}
                      disabled={isUploadingBanner}
                    />
                  </Button>
                </label>
                <p className="text-xs text-gray-500 dark:text-gray-400">
                  JPG、PNG、GIF、5MB以下
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <form onSubmit={handleSubmit(onSubmit)}>
        {/* 表示名 */}
        <div className="mb-4">
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            表示名
          </label>
          <Controller
            name="display_name"
            control={control}
            render={({ field }) => (
              <input
                {...field}
                type="text"
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:text-white"
                placeholder="あなたの表示名"
              />
            )}
          />
          {errors.display_name && (
            <p className="mt-1 text-sm text-red-600 dark:text-red-400">
              {errors.display_name.message}
            </p>
          )}
        </div>
        
        {/* 自己紹介 */}
        <div className="mb-6">
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            自己紹介
          </label>
          <Controller
            name="bio"
            control={control}
            render={({ field }) => (
              <textarea
                {...field}
                rows={4}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:text-white"
                placeholder="あなたについて教えてください（160文字以内）"
              />
            )}
          />
          {errors.bio && (
            <p className="mt-1 text-sm text-red-600 dark:text-red-400">
              {errors.bio.message}
            </p>
          )}
        </div>
        
        {/* アクションボタン */}
        <div className="flex justify-end space-x-3">
          <Button
            type="button"
            variant="outline"
            onClick={onCancel}
            disabled={isSubmitting}
          >
            キャンセル
          </Button>
          <Button
            type="submit"
            disabled={isSubmitting || isUploadingAvatar || isUploadingBanner}
          >
            {isSubmitting ? '保存中...' : '保存する'}
          </Button>
        </div>
      </form>
    </div>
  );
};

export default EditProfileForm; 