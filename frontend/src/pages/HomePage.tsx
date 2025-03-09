import { useState } from 'react';
import Timeline from '../components/timeline/Timeline';
import Header from '../components/layout/Header';

const HomePage = () => {
  const [activeTab, setActiveTab] = useState<'home' | 'explore'>('home');

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      {/* グローバルヘッダー */}
      <Header />
      
      {/* タブナビゲーション */}
      <div className="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
        <div className="max-w-6xl mx-auto px-4">
          <div className="flex space-x-8">
            <button
              onClick={() => setActiveTab('home')}
              className={`py-4 px-1 text-sm font-medium border-b-2 ${
                activeTab === 'home'
                  ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                  : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300'
              }`}
            >
              ホーム
            </button>
            <button
              onClick={() => setActiveTab('explore')}
              className={`py-4 px-1 text-sm font-medium border-b-2 ${
                activeTab === 'explore'
                  ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                  : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300'
              }`}
            >
              エクスプローラー
            </button>
          </div>
        </div>
      </div>

      {/* メインコンテンツ */}
      <main className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
        <div className="bg-white dark:bg-gray-800 rounded-lg overflow-hidden">
          {/* タイムライン切り替え */}
          {activeTab === 'home' ? (
            <Timeline type="home" showPostForm={true} />
          ) : (
            <Timeline type="explore" showPostForm={false} />
          )}
        </div>
      </main>
    </div>
  );
};

export default HomePage; 