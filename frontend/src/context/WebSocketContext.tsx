import { createContext, useContext, ReactNode } from 'react';
import useWebSocket from '../hooks/useWebSocket';

// WebSocketコンテキストの型定義
interface WebSocketContextType {
  isConnected: boolean;
  error: string | null;
  connect: () => void;
  disconnect: () => void;
  addMessageHandler: (type: string, handler: (data: any) => void) => () => void;
  sendMessage: (type: string, data: any) => boolean;
}

// デフォルト値
const defaultContext: WebSocketContextType = {
  isConnected: false,
  error: null,
  connect: () => {},
  disconnect: () => {},
  addMessageHandler: () => () => {},
  sendMessage: () => false
};

// コンテキストの作成
const WebSocketContext = createContext<WebSocketContextType>(defaultContext);

// プロバイダーコンポーネント
export const WebSocketProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const webSocket = useWebSocket();
  
  return (
    <WebSocketContext.Provider value={webSocket}>
      {children}
    </WebSocketContext.Provider>
  );
};

// カスタムフック
export const useWebSocketContext = () => useContext(WebSocketContext);

export default WebSocketContext; 