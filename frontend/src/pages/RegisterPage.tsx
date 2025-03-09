import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import Button from '../components/ui/Button';
import { useAuthStore } from '../store/authStore';

// バリデーションスキーマ
const registerSchema = z.object({
  username: z.string()
    .min(3, 'ユーザー名は3文字以上である必要があります')
    .max(20, 'ユーザー名は20文字以下である必要があります')
    .regex(/^[a-z0-9_]+$/, 'ユーザー名は小文字、数字、アンダースコアのみ使用できます'),
  display_name: z.string()
    .min(1, '表示名は必須です')
    .max(50, '表示名は50文字以下である必要があります'),
  email: z.string().email('有効なメールアドレスを入力してください'),
  password: z.string()
    .min(8, 'パスワードは8文字以上である必要があります')
    .regex(/[A-Z]/, 'パスワードは少なくとも1つの大文字を含む必要があります')
    .regex(/[0-9]/, 'パスワードは少なくとも1つの数字を含む必要があります'),
  confirmPassword: z.string(),
}).refine((data) => data.password === data.confirmPassword, {
  message: 'パスワードが一致しません',
  path: ['confirmPassword'],
});

type RegisterFormData = z.infer<typeof registerSchema>;

const RegisterPage = () => {
  const navigate = useNavigate();
  const { setUser, setTokens, error, clearError, loading, setLoading, setError } = useAuthStore();
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      username: '',
      display_name: '',
      email: '',
      password: '',
      confirmPassword: '',
    }
  });
  
  const onSubmit = async (data: RegisterFormData) => {
    // エラーをクリア
    clearError();
    setLoading(true);
    
    // confirmPasswordはAPIに送信しないため除外
    const { confirmPassword, ...registrationData } = data;
    
    // 登録処理
    try {
      console.log('登録リクエスト:', registrationData);
      
      // 直接Fetch APIを使用してリクエスト
      const response = await fetch('http://localhost:8080/api/v1/auth/register', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(registrationData)
      });
      
      const responseData = await response.json();
      console.log('登録レスポンス:', responseData);
      
      // レスポンスに直接ユーザー情報とトークン情報が含まれているか確認
      if (response.ok && responseData) {
        // ユーザー情報を取得
        const user = responseData.user;
        const token = responseData.token;
        const refresh_token = responseData.refresh_token;
        
        if (user && token) {
          // ストアに保存
          setUser(user);
          setTokens(token, refresh_token);
          
          // ローカルストレージにも保存（ページ更新時に利用）
          localStorage.setItem('token', token);
          localStorage.setItem('refreshToken', refresh_token);
          localStorage.setItem('user', JSON.stringify(user));
          
          console.log('登録成功:', user);
          
          // ホームページに遷移
          navigate('/');
          return;
        }
      }
      
      // 成功レスポンスの構造が異なる可能性を考慮（success フィールドがある場合）
      if (responseData.success && responseData.data) {
        const { user, token, refresh_token } = responseData.data;
        
        if (user && token) {
          // ストアに保存
          setUser(user);
          setTokens(token, refresh_token);
          
          // ローカルストレージにも保存
          localStorage.setItem('token', token);
          localStorage.setItem('refreshToken', refresh_token);
          localStorage.setItem('user', JSON.stringify(user));
          
          console.log('登録成功 (data):', user);
          
          // ホームページに遷移
          navigate('/');
          return;
        }
      }
      
      // ここまで来た場合はエラー
      const errorMessage = responseData.error?.message || 
                           responseData.message || 
                           'ユーザー登録に失敗しました。入力内容を確認してください。';
      setError(errorMessage);
      console.error('登録エラー:', errorMessage);
      
    } catch (error) {
      console.error('Registration error:', error);
      setError('サーバーとの通信に失敗しました。');
    } finally {
      setLoading(false);
    }
  };
  
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full p-6 bg-white dark:bg-gray-800 rounded-lg shadow-lg">
        <div className="text-center">
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-2">GoX</h1>
          <h2 className="text-2xl font-semibold text-gray-900 dark:text-white">アカウント作成</h2>
          <p className="mt-2 text-gray-600 dark:text-gray-400">新しいアカウントを作成してください</p>
        </div>
        
        {error && (
          <div className="mt-4 p-3 bg-red-100 text-red-700 rounded-md dark:bg-red-900 dark:text-red-200">
            {error}
          </div>
        )}
        
        <form onSubmit={handleSubmit(onSubmit)} className="mt-8 space-y-4">
          <div>
            <label htmlFor="username" className="block text-sm font-medium text-gray-700 dark:text-gray-200">
              ユーザー名
            </label>
            <input
              id="username"
              type="text"
              className="input mt-1"
              placeholder="username123"
              {...register('username')}
            />
            {errors.username && (
              <p className="mt-1 text-sm text-red-600 dark:text-red-400">{errors.username.message}</p>
            )}
          </div>
          
          <div>
            <label htmlFor="display_name" className="block text-sm font-medium text-gray-700 dark:text-gray-200">
              表示名
            </label>
            <input
              id="display_name"
              type="text"
              className="input mt-1"
              placeholder="表示名"
              {...register('display_name')}
            />
            {errors.display_name && (
              <p className="mt-1 text-sm text-red-600 dark:text-red-400">{errors.display_name.message}</p>
            )}
          </div>
          
          <div>
            <label htmlFor="email" className="block text-sm font-medium text-gray-700 dark:text-gray-200">
              メールアドレス
            </label>
            <input
              id="email"
              type="email"
              className="input mt-1"
              placeholder="your-email@example.com"
              {...register('email')}
            />
            {errors.email && (
              <p className="mt-1 text-sm text-red-600 dark:text-red-400">{errors.email.message}</p>
            )}
          </div>
          
          <div>
            <label htmlFor="password" className="block text-sm font-medium text-gray-700 dark:text-gray-200">
              パスワード
            </label>
            <div className="relative">
              <input
                id="password"
                type={showPassword ? 'text' : 'password'}
                className="input mt-1 pr-10"
                placeholder="********"
                {...register('password')}
              />
              <button
                type="button"
                className="absolute inset-y-0 right-0 pr-3 flex items-center"
                onClick={() => setShowPassword(!showPassword)}
              >
                {showPassword ? (
                  <span className="text-gray-500">非表示</span>
                ) : (
                  <span className="text-gray-500">表示</span>
                )}
              </button>
            </div>
            {errors.password && (
              <p className="mt-1 text-sm text-red-600 dark:text-red-400">{errors.password.message}</p>
            )}
          </div>
          
          <div>
            <label htmlFor="confirmPassword" className="block text-sm font-medium text-gray-700 dark:text-gray-200">
              パスワード（確認）
            </label>
            <div className="relative">
              <input
                id="confirmPassword"
                type={showConfirmPassword ? 'text' : 'password'}
                className="input mt-1 pr-10"
                placeholder="********"
                {...register('confirmPassword')}
              />
              <button
                type="button"
                className="absolute inset-y-0 right-0 pr-3 flex items-center"
                onClick={() => setShowConfirmPassword(!showConfirmPassword)}
              >
                {showConfirmPassword ? (
                  <span className="text-gray-500">非表示</span>
                ) : (
                  <span className="text-gray-500">表示</span>
                )}
              </button>
            </div>
            {errors.confirmPassword && (
              <p className="mt-1 text-sm text-red-600 dark:text-red-400">{errors.confirmPassword.message}</p>
            )}
          </div>
          
          <Button
            type="submit"
            fullWidth
            isLoading={loading}
          >
            登録する
          </Button>
          
          <div className="text-center mt-4">
            <p className="text-sm text-gray-600 dark:text-gray-400">
              すでにアカウントをお持ちですか？{' '}
              <Link to="/login" className="text-primary-500 hover:text-primary-600">
                ログインする
              </Link>
            </p>
          </div>
        </form>
      </div>
    </div>
  );
};

export default RegisterPage; 