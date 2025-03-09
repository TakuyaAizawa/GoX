import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    host: '0.0.0.0',     // すべてのネットワークインターフェースでリッスン
    // port: 5173,          // デフォルトポート
    allowedHosts: [
      'win-t',           // エラーメッセージに出たホスト名を追加
      'localhost',
      'all',             // すべてのホストを許可する場合はこれを追加
    ]
  }
})