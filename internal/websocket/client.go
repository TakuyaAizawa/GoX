package websocket

import (
	"time"

	"github.com/TakuyaAizawa/gox/pkg/logger"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// 書き込みのタイムアウト時間
	writeWait = 10 * time.Second

	// クライアントからの次のpingを待つ最大時間
	pongWait = 60 * time.Second

	// pingを送信する間隔（pongWaitより小さくすること）
	pingPeriod = (pongWait * 9) / 10

	// 許容最大メッセージサイズ
	maxMessageSize = 512
)

// Client はWebSocket接続とイベント配信を管理する
type Client struct {
	// クライアントID（ユーザーID）
	ID uuid.UUID

	// 所属するHub
	hub *Hub

	// WebSocket接続
	conn *websocket.Conn

	// 送信メッセージチャネル
	send chan []byte

	// ロガー
	log logger.Logger
}

// NewClient は新しいWebSocketクライアントを作成する
func NewClient(hub *Hub, conn *websocket.Conn, userID uuid.UUID, log logger.Logger) *Client {
	return &Client{
		ID:   userID,
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
		log:  log,
	}
}

// ReadPump はクライアントからのメッセージを処理する
// 各クライアント接続ごとに1つのgoroutineで実行される必要がある
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// クライアントからのメッセージ読み取りループ
	// 現在の実装では、クライアントからのメッセージは単に破棄される
	// 必要に応じて、ここでクライアントからのメッセージを処理することができる
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.log.Warn("WebSocket読み取りエラー", "error", err)
			}
			break
		}
		// 現在はクライアントからのメッセージは処理しない
	}
}

// WritePump はクライアントへのメッセージ送信を処理する
// 各クライアント接続ごとに1つのgoroutineで実行される必要がある
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hubがチャネルを閉じた
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// キューにあるすべてのメッセージをまとめて送信
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
