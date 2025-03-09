/**
 * GoX デバッグヘルパー
 * デバッグ用のユーティリティ関数を提供します
 */

/**
 * テスト用のログメッセージを出力する
 * コンソールロガーのテスト用
 */
export const generateTestLogs = () => {
  console.log('通常のログメッセージ');
  console.info('情報メッセージ', { timestamp: new Date().toISOString() });
  console.warn('警告メッセージ', '注意が必要です');
  
  try {
    // 意図的なエラーを発生させる
    const testObj: any = null;
    testObj.nonExistentMethod();
  } catch (error) {
    console.error('エラーが発生しました:', error instanceof Error ? error.message : '不明なエラー');
  }
  
  // 複雑なオブジェクトのログ
  console.log('複雑なオブジェクト:', {
    user: {
      id: 1,
      name: 'テストユーザー',
      roles: ['admin', 'editor'],
      settings: {
        theme: 'dark',
        notifications: true
      }
    },
    stats: [1, 2, 3, 4, 5]
  });
};

/**
 * APIレスポンスのロギング関数
 */
export const logApiResponse = <T>(endpoint: string, data: T) => {
  console.info(`API Response [${endpoint}]:`, data);
};

/**
 * APIエラーのロギング関数
 */
export const logApiError = (endpoint: string, error: unknown) => {
  console.error(`API Error [${endpoint}]:`, error);
};

/**
 * パフォーマンス計測のためのシンプルなタイマー
 */
export class PerformanceTimer {
  private startTime: number;
  private name: string;

  constructor(name: string) {
    this.name = name;
    this.startTime = performance.now();
  }

  stop() {
    const endTime = performance.now();
    const duration = endTime - this.startTime;
    console.info(`⏱️ パフォーマンス [${this.name}]: ${duration.toFixed(2)}ms`);
    return duration;
  }
}

/**
 * ブラウザとデバイス情報をログに出力
 */
export const logEnvironmentInfo = () => {
  console.info('環境情報:', {
    userAgent: navigator.userAgent,
    language: navigator.language,
    screenSize: {
      width: window.screen.width,
      height: window.screen.height
    },
    viewport: {
      width: window.innerWidth,
      height: window.innerHeight
    },
    cookiesEnabled: navigator.cookieEnabled,
    darkMode: window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches
  });
}; 