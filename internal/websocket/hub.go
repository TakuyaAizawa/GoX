package websocket

import (
	"encoding/json"
	"sync"

	"github.com/TakuyaAizawa/gox/pkg/logger"
	"github.com/google/uuid"
)

// Hub はWebSocket接続の中央管理を行う
type Hub struct {
	// すべてのアクティブなクライアント
	clients map[*Client]bool

	// ユーザーID別のクライアントマップ
	userClients map[uuid.UUID][]*Client

	// ユーザーマップの排他制御
	userMutex sync.RWMutex

	// すべてのクライアントへのブロードキャストメッセージ
	broadcast chan []byte

	// 特定ユーザーへの通知メッセージ
	notify chan *NotificationMessage

	// クライアント登録リクエスト
	register chan *Client

	// クライアント登録解除リクエスト
	unregister chan *Client

	// ロガー
	log logger.Logger
}

// NotificationMessage はユーザーへの通知メッセージを表す
type NotificationMessage struct {
	// 通知の受信者ID
	UserID uuid.UUID

	// JSON形式の通知データ
	Payload []byte
}

// NewHub は新しいHubを作成する
func NewHub(log logger.Logger) *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		userClients: make(map[uuid.UUID][]*Client),
		broadcast:   make(chan []byte),
		notify:      make(chan *NotificationMessage),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		log:         log,
	}
}

// Run はハブの主要ループを開始する
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			// クライアントを登録
			h.clients[client] = true

			// ユーザーIDでインデックス化
			h.userMutex.Lock()
			h.userClients[client.ID] = append(h.userClients[client.ID], client)
			h.userMutex.Unlock()

			h.log.Info("WebSocketクライアント接続", "user_id", client.ID)

		case client := <-h.unregister:
			// クライアントの登録解除
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)

				// ユーザーのクライアントリストからも削除
				h.userMutex.Lock()
				userClients := h.userClients[client.ID]
				for i, c := range userClients {
					if c == client {
						// スライスから削除
						h.userClients[client.ID] = append(userClients[:i], userClients[i+1:]...)
						break
					}
				}
				// クライアントがなくなったらマップからも削除
				if len(h.userClients[client.ID]) == 0 {
					delete(h.userClients, client.ID)
				}
				h.userMutex.Unlock()

				h.log.Info("WebSocketクライアント切断", "user_id", client.ID)
			}

		case message := <-h.broadcast:
			// すべてのクライアントにブロードキャスト
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// 送信バッファがいっぱいの場合はクライアントを切断
					close(client.send)
					delete(h.clients, client)

					h.userMutex.Lock()
					userClients := h.userClients[client.ID]
					for i, c := range userClients {
						if c == client {
							h.userClients[client.ID] = append(userClients[:i], userClients[i+1:]...)
							break
						}
					}
					if len(h.userClients[client.ID]) == 0 {
						delete(h.userClients, client.ID)
					}
					h.userMutex.Unlock()
				}
			}

		case notification := <-h.notify:
			// 特定ユーザーへの通知
			h.userMutex.RLock()
			clients := h.userClients[notification.UserID]
			h.userMutex.RUnlock()

			if len(clients) > 0 {
				h.log.Debug("通知送信",
					"user_id", notification.UserID,
					"client_count", len(clients))

				// ユーザーの全クライアントに送信
				for _, client := range clients {
					select {
					case client.send <- notification.Payload:
					default:
						// バッファがいっぱいの場合はこのクライアントをスキップ
						h.log.Warn("通知送信失敗: バッファがいっぱい", "user_id", client.ID)
					}
				}
			}
		}
	}
}

// NotifyUser は特定のユーザーに通知を送信する
func (h *Hub) NotifyUser(userID uuid.UUID, notification interface{}) error {
	payload, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	h.notify <- &NotificationMessage{
		UserID:  userID,
		Payload: payload,
	}

	return nil
}

// Register はクライアントをハブに登録する
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Broadcast はすべての接続クライアントにメッセージを送信する
func (h *Hub) Broadcast(message interface{}) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}

	h.broadcast <- payload
	return nil
}
