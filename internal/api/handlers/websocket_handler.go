package handlers

import (
	"net/http"

	"github.com/TakuyaAizawa/gox/internal/util/response"
	"github.com/TakuyaAizawa/gox/internal/websocket"
	"github.com/TakuyaAizawa/gox/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	gorillaWs "github.com/gorilla/websocket"
)

// WebSocketHandler WebSocket接続を管理するハンドラー
type WebSocketHandler struct {
	hub *websocket.Hub
	log logger.Logger
}

// WebSocketのアップグレード設定
var upgrader = gorillaWs.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// CORS対応のため、接続元をチェックしない
	// 本番環境では適切なオリジン検証を行うべき
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// NewWebSocketHandler 新しいWebSocketハンドラーを作成する
func NewWebSocketHandler(log logger.Logger) *WebSocketHandler {
	hub := websocket.NewHub(log)
	go hub.Run()

	return &WebSocketHandler{
		hub: hub,
		log: log,
	}
}

// HandleWSConnection WebSocket接続をハンドリングする
func (h *WebSocketHandler) HandleWSConnection(c *gin.Context) {
	// ユーザー認証の確認
	userIDStr, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "認証が必要です")
		return
	}

	userID, ok := userIDStr.(uuid.UUID)
	if !ok {
		h.log.Error("ユーザーIDのフォーマットが不正です", "user_id", userIDStr)
		response.InternalServerError(c, "内部エラーが発生しました")
		return
	}

	// WebSocketへのアップグレード
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.log.Error("WebSocketアップグレードに失敗しました", "error", err)
		return
	}

	// 新しいクライアントの作成
	client := websocket.NewClient(h.hub, conn, userID, h.log)

	// クライアントをハブに登録
	h.hub.Register(client)

	// 接続を確認する簡単なシステムメッセージ
	welcomeMsg := websocket.NewSystemMessage("WebSocket接続が確立されました")
	err = h.hub.NotifyUser(userID, welcomeMsg)
	if err != nil {
		h.log.Error("ウェルカムメッセージの送信に失敗しました", "error", err)
	}

	// メッセージの読み書きはそれぞれ別のgoroutineで実行
	go client.WritePump()
	go client.ReadPump()
}

// GetNotificationHub 通知ハブを取得する（他のサービスからの利用用）
func (h *WebSocketHandler) GetNotificationHub() *websocket.Hub {
	return h.hub
}
