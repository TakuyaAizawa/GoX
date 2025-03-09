import { useState, useEffect, useCallback, useRef } from 'react';
import { useAuthStore } from '../store/authStore';

interface WebSocketMessage {
  type: string;
  data: any;
}

type MessageHandler = (data: any) => void;

const useWebSocket = () => {
  const { user, isAuthenticated } = useAuthStore();
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const socketRef = useRef<WebSocket | null>(null);
  const messageHandlersRef = useRef<Map<string, MessageHandler[]>>(new Map());
  
  // WebSocketサーバーURL
  const getWebSocketUrl = useCallback(() => {
    const baseUrl = import.meta.env.VITE_API_BASE_URL || '';
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    
    // baseUrlからホスト部分のみを取得
    let wsUrl = baseUrl.replace(/^https?:\/\//, '');
    // パスを/wsに設定
    wsUrl = wsUrl.replace(/\/+$/, '') + '/ws';
    
    return `${wsProtocol}//${wsUrl}`;
  }, []);
  
  // WebSocketコネクションを確立
  const connect = useCallback(() => {
    if (!isAuthenticated || !user) {
      setError('認証されていません');
      return;
    }
    
    try {
      const token = localStorage.getItem('token');
      if (!token) {
        setError('認証トークンがありません');
        return;
      }
      
      const wsUrl = `${getWebSocketUrl()}?token=${token}`;
      const socket = new WebSocket(wsUrl);
      
      socket.onopen = () => {
        console.log('WebSocket接続が確立されました');
        setIsConnected(true);
        setError(null);
      };
      
      socket.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data) as WebSocketMessage;
          console.log('WebSocketメッセージを受信:', message);
          
          // メッセージタイプに応じたハンドラーを呼び出す
          const handlers = messageHandlersRef.current.get(message.type) || [];
          handlers.forEach(handler => handler(message.data));
          
          // すべてのメッセージを処理する'*'イベントハンドラーを呼び出す
          const allHandlers = messageHandlersRef.current.get('*') || [];
          allHandlers.forEach(handler => handler(message));
        } catch (err) {
          console.error('WebSocketメッセージの解析に失敗しました:', err);
        }
      };
      
      socket.onclose = (event) => {
        console.log('WebSocket接続が閉じられました:', event);
        setIsConnected(false);
        
        // 異常終了の場合は、数秒後に再接続を試みる
        if (!event.wasClean) {
          setError('接続が切断されました。再接続します...');
          setTimeout(() => {
            connect();
          }, 3000);
        }
      };
      
      socket.onerror = (event) => {
        console.error('WebSocketエラー:', event);
        setError('接続エラーが発生しました');
      };
      
      socketRef.current = socket;
    } catch (err) {
      console.error('WebSocket接続の確立に失敗しました:', err);
      setError('接続の確立に失敗しました');
    }
  }, [isAuthenticated, user, getWebSocketUrl]);
  
  // WebSocketコネクションを切断
  const disconnect = useCallback(() => {
    if (socketRef.current && socketRef.current.readyState === WebSocket.OPEN) {
      socketRef.current.close();
      setIsConnected(false);
    }
  }, []);
  
  // メッセージハンドラーを登録
  const addMessageHandler = useCallback((type: string, handler: MessageHandler) => {
    const handlers = messageHandlersRef.current.get(type) || [];
    messageHandlersRef.current.set(type, [...handlers, handler]);
    
    // クリーンアップ関数を返す
    return () => {
      const currentHandlers = messageHandlersRef.current.get(type) || [];
      messageHandlersRef.current.set(
        type,
        currentHandlers.filter(h => h !== handler)
      );
    };
  }, []);
  
  // メッセージを送信
  const sendMessage = useCallback((type: string, data: any) => {
    if (socketRef.current && socketRef.current.readyState === WebSocket.OPEN) {
      const message: WebSocketMessage = { type, data };
      socketRef.current.send(JSON.stringify(message));
      return true;
    }
    return false;
  }, []);
  
  // 認証状態が変わったら自動的に接続/切断
  useEffect(() => {
    if (isAuthenticated && user) {
      connect();
    } else {
      disconnect();
    }
    
    return () => {
      disconnect();
    };
  }, [isAuthenticated, user, connect, disconnect]);
  
  return {
    isConnected,
    error,
    connect,
    disconnect,
    addMessageHandler,
    sendMessage
  };
};

export default useWebSocket; 