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
    // 直接環境変数からWebSocket URLを使用
    return import.meta.env.VITE_WS_URL;
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
      
      // 既存の接続を閉じる
      if (socketRef.current) {
        socketRef.current.close();
      }
      
      const wsUrl = `${getWebSocketUrl()}?token=${token}`;
      console.log('WebSocket接続URL:', wsUrl);
      
      const socket = new WebSocket(wsUrl);
      socketRef.current = socket;
      
      socket.onopen = () => {
        console.log('WebSocket接続が確立されました');
        setIsConnected(true);
        setError(null);
      };
      
      socket.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data) as WebSocketMessage;
          console.log('WebSocketメッセージ受信:', message.type);
          
          // メッセージタイプに対応するハンドラーを呼び出す
          const handlers = messageHandlersRef.current.get(message.type) || [];
          handlers.forEach(handler => handler(message.data));
        } catch (e) {
          console.error('WebSocketメッセージの解析に失敗しました:', e);
        }
      };
      
      socket.onerror = (event) => {
        console.error('WebSocketエラー:', event);
        setError('WebSocket接続エラー');
        setIsConnected(false);
      };
      
      socket.onclose = (event) => {
        console.log('WebSocket接続が閉じられました:', event.code, event.reason);
        setIsConnected(false);
        
        // 正常なクローズでない場合は再接続を試みる
        if (event.code !== 1000) {
          console.log('WebSocketの再接続を試みます...');
          setTimeout(() => {
            if (isAuthenticated && user) {
              connect();
            }
          }, 3000);
        }
      };
    } catch (e) {
      console.error('WebSocket接続の確立に失敗しました:', e);
      setError('WebSocket接続の確立に失敗しました');
    }
  }, [isAuthenticated, user, getWebSocketUrl]);
  
  // WebSocketを切断
  const disconnect = useCallback(() => {
    if (socketRef.current) {
      socketRef.current.close();
      socketRef.current = null;
      setIsConnected(false);
    }
  }, []);
  
  // メッセージハンドラーを追加
  const addMessageHandler = useCallback((messageType: string, handler: MessageHandler) => {
    const handlers = messageHandlersRef.current.get(messageType) || [];
    handlers.push(handler);
    messageHandlersRef.current.set(messageType, handlers);
    
    // クリーンアップ関数を返す
    return () => {
      const updatedHandlers = (messageHandlersRef.current.get(messageType) || [])
        .filter(h => h !== handler);
      messageHandlersRef.current.set(messageType, updatedHandlers);
    };
  }, []);
  
  // メッセージを送信
  const sendMessage = useCallback((messageType: string, data: any) => {
    if (!socketRef.current || socketRef.current.readyState !== WebSocket.OPEN) {
      console.error('WebSocketが接続されていません');
      return false;
    }
    
    try {
      const message: WebSocketMessage = {
        type: messageType,
        data
      };
      
      socketRef.current.send(JSON.stringify(message));
      return true;
    } catch (e) {
      console.error('メッセージの送信に失敗しました:', e);
      return false;
    }
  }, []);
  
  // 認証状態が変わったときに自動的に接続/切断
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