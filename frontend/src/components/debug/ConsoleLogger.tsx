import React, { useState, useEffect, useRef } from 'react';
import { generateTestLogs, logEnvironmentInfo } from './DebugHelper';
import './ConsoleLogger.css';

interface LogEntry {
  id: number;
  type: 'log' | 'warn' | 'error' | 'info';
  content: string;
  timestamp: string;
}

// コンソールログを表示するコンポーネント
const ConsoleLogger: React.FC = () => {
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [isVisible, setIsVisible] = useState<boolean>(false);
  const [filter, setFilter] = useState<string>('');
  const [isExpanded, setIsExpanded] = useState<boolean>(false);
  const logContainerRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    // オリジナルのコンソールメソッドを保存
    const originalConsole = {
      log: console.log,
      warn: console.warn,
      error: console.error,
      info: console.info
    };

    // コンソールメソッドをオーバーライド
    console.log = (...args: any[]) => {
      originalConsole.log(...args);
      addLog('log', args);
    };

    console.warn = (...args: any[]) => {
      originalConsole.warn(...args);
      addLog('warn', args);
    };

    console.error = (...args: any[]) => {
      originalConsole.error(...args);
      addLog('error', args);
    };

    console.info = (...args: any[]) => {
      originalConsole.info(...args);
      addLog('info', args);
    };

    // 初期ログメッセージ
    console.info('コンソールロガーが初期化されました');
    logEnvironmentInfo();

    // クリーンアップ関数でオリジナルのコンソールメソッドを復元
    return () => {
      console.log = originalConsole.log;
      console.warn = originalConsole.warn;
      console.error = originalConsole.error;
      console.info = originalConsole.info;
    };
  }, []);

  // ログを追加する関数
  const addLog = (type: 'log' | 'warn' | 'error' | 'info', args: any[]) => {
    const timestamp = new Date().toISOString().split('T')[1].slice(0, -1);
    
    // オブジェクトや配列を適切に文字列化
    const formattedArgs = args.map(arg => {
      if (typeof arg === 'object' && arg !== null) {
        try {
          return JSON.stringify(arg, null, 2);
        } catch (e) {
          return String(arg);
        }
      }
      return String(arg);
    });

    setLogs(prevLogs => [
      ...prevLogs,
      {
        id: Date.now() + Math.random(),
        type,
        content: formattedArgs.join(' '),
        timestamp
      }
    ]);
  };

  // ログをクリア
  const clearLogs = () => {
    setLogs([]);
  };

  // ログビューの表示/非表示を切り替え
  const toggleVisibility = () => {
    setIsVisible(!isVisible);
  };

  // パネルの拡大/縮小
  const toggleExpanded = () => {
    setIsExpanded(!isExpanded);
  };
  
  // テストログの生成
  const handleTestLogs = () => {
    generateTestLogs();
  };

  // 環境情報の出力
  const handleShowEnvironment = () => {
    logEnvironmentInfo();
  };

  // フィルタリングされたログ
  const filteredLogs = logs.filter(log => 
    log.content.toLowerCase().includes(filter.toLowerCase()) ||
    log.type.toLowerCase().includes(filter.toLowerCase())
  );

  // 自動スクロール
  useEffect(() => {
    if (logContainerRef.current) {
      logContainerRef.current.scrollTop = logContainerRef.current.scrollHeight;
    }
  }, [filteredLogs]);

  const getPanelClassName = () => {
    let className = "console-panel";
    if (isExpanded) className += " console-panel-expanded";
    return className;
  };

  return (
    <>
      {/* コンソール操作ボタン群 - 常に表示 */}
      <div className="console-buttons">
        <button 
          className="console-toggle main-button"
          onClick={toggleVisibility}
        >
          {isVisible ? 'コンソールを閉じる' : 'コンソールを開く'}
        </button>
        <button 
          className="console-test-button"
          onClick={handleTestLogs}
          title="テストログを出力します"
        >
          テストログ
        </button>
      </div>

      {/* コンソールパネル */}
      {isVisible && (
        <div className={getPanelClassName()}>
          <div className="console-header">
            <h3>コンソールログ</h3>
            <div className="console-controls">
              <input
                type="text"
                placeholder="ログをフィルター..."
                value={filter}
                onChange={(e) => setFilter(e.target.value)}
                className="console-filter"
              />
              <button onClick={handleTestLogs} className="console-action-btn">テストログ</button>
              <button onClick={handleShowEnvironment} className="console-action-btn">環境情報</button>
              <button onClick={toggleExpanded} className="console-action-btn">
                {isExpanded ? '縮小' : '拡大'}
              </button>
              <button onClick={clearLogs} className="console-clear">クリア</button>
            </div>
          </div>

          <div className="console-logs" ref={logContainerRef}>
            {filteredLogs.length > 0 ? (
              filteredLogs.map(log => (
                <div key={log.id} className={`console-log console-${log.type}`}>
                  <span className="console-timestamp">{log.timestamp}</span>
                  <span className="console-type">[{log.type.toUpperCase()}]</span>
                  <pre className="console-content">{log.content}</pre>
                </div>
              ))
            ) : (
              <div className="console-empty">ログはありません</div>
            )}
          </div>
          
          <div className="console-footer">
            <span className="console-stats">
              ログ数: {filteredLogs.length} / {logs.length}
              {filter && ` (フィルター: "${filter}")`}
            </span>
          </div>
        </div>
      )}
    </>
  );
}

export default ConsoleLogger; 