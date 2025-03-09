import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import Button from '../components/ui/Button';
import { useAuthStore } from '../store/authStore';

// バリデーションスキーマ
const loginSchema = z.object({
  email: z.string().email('有効なメールアドレスを入力してください'),
  password: z.string().min(6, 'パスワードは6文字以上である必要があります'),
});

type LoginFormData = z.infer<typeof loginSchema>;

const LoginPage = () => {
  const navigate = useNavigate();
  const { setUser, setTokens, error, clearError, loading, setLoading, setError } = useAuthStore();
  const [showPassword, setShowPassword] = useState(false);
  
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
  });
  
  const onSubmit = async (data: LoginFormData) => {
    // エラーをクリア
    clearError();
    setLoading(true);
    
    // ログイン処理
    try {
      console.log('ログインリクエスト:', data);
      
      // 直接Fetch APIを使用してリクエスト
      const response = await fetch('http://localhost:8080/api/v1/auth/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(data)
      });
      
      const responseData = await response.json();
      console.log('ログインレスポンス:', responseData);
      
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
          
          console.log('ログイン成功:', user);
          
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
          
          console.log('ログイン成功 (data):', user);
          
          // ホームページに遷移
          navigate('/');
          return;
        }
      }
      
      // ここまで来た場合はエラー
      const errorMessage = responseData.error?.message || 
                           responseData.message || 
                           'ログインに失敗しました。認証情報を確認してください。';
      setError(errorMessage);
      console.error('ログインエラー:', errorMessage);
      
    } catch (error) {
      console.error('Login error:', error);
      setError('サーバーとの通信に失敗しました。');
    } finally {
      setLoading(false);
    }
  };
  
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
      <div className="max-w-md w-full p-6 bg-white dark:bg-gray-800 rounded-lg shadow-lg">
        <div className="text-center">
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-2">GoX</h1>
          <h2 className="text-2xl font-semibold text-gray-900 dark:text-white">ログイン</h2>
          <p className="mt-2 text-gray-600 dark:text-gray-400">アカウントにログインしてください</p>
        </div>
        
        {error && (
          <div className="mt-4 p-3 bg-red-100 text-red-700 rounded-md dark:bg-red-900 dark:text-red-200">
            {error}
          </div>
        )}
        
        <form onSubmit={handleSubmit(onSubmit)} className="mt-8 space-y-6">
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
          
          <Button
            type="submit"
            fullWidth
            isLoading={loading}
          >
            ログイン
          </Button>
          
          <div className="text-center mt-4">
            <p className="text-sm text-gray-600 dark:text-gray-400">
              アカウントをお持ちでないですか？{' '}
              <Link to="/register" className="text-primary-500 hover:text-primary-600">
                登録する
              </Link>
            </p>
          </div>
        </form>
      </div>
    </div>
  );
};

export default LoginPage; 