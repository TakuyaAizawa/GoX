package websocket

import (
	"time"

	"github.com/google/uuid"
)

// EventType は通知イベントの種類を表す
type EventType string

const (
	// EventTypeNotification は通常の通知イベント
	EventTypeNotification EventType = "notification"

	// EventTypeLike はいいねの通知イベント
	EventTypeLike EventType = "like"

	// EventTypeFollow はフォロー通知イベント
	EventTypeFollow EventType = "follow"

	// EventTypeReply は返信通知イベント
	EventTypeReply EventType = "reply"

	// EventTypeRepost はリポスト通知イベント
	EventTypeRepost EventType = "repost"

	// EventTypeMention はメンション通知イベント
	EventTypeMention EventType = "mention"

	// EventTypeSystem はシステム通知イベント
	EventTypeSystem EventType = "system"
)

// WebSocketMessage はWebSocketを通じて送信されるメッセージの基本構造
type WebSocketMessage struct {
	// メッセージの種類
	Type string `json:"type"`

	// メッセージの内容
	Data interface{} `json:"data"`
}

// NotificationEvent は通知イベントの詳細を表す
type NotificationEvent struct {
	// 通知ID
	ID uuid.UUID `json:"id"`

	// 通知タイプ
	Type EventType `json:"type"`

	// 通知のアクター（送信者）情報
	Actor ActorInfo `json:"actor"`

	// 関連する投稿情報（あれば）
	Post *PostInfo `json:"post,omitempty"`

	// 通知生成時刻
	CreatedAt time.Time `json:"created_at"`

	// 通知内容の概要
	Message string `json:"message"`
}

// ActorInfo は通知アクターの情報
type ActorInfo struct {
	// ユーザーID
	ID uuid.UUID `json:"id"`

	// ユーザー名
	Username string `json:"username"`

	// 表示名
	DisplayName string `json:"display_name"`

	// プロフィール画像URL
	AvatarURL string `json:"avatar_url"`
}

// PostInfo は通知に関連する投稿情報
type PostInfo struct {
	// 投稿ID
	ID uuid.UUID `json:"id"`

	// 投稿内容のプレビュー
	Content string `json:"content"`
}

// NewNotificationMessage は通知メッセージを作成する
func NewNotificationMessage(event NotificationEvent) *WebSocketMessage {
	return &WebSocketMessage{
		Type: "notification",
		Data: event,
	}
}

// NewSystemMessage はシステムメッセージを作成する
func NewSystemMessage(message string) *WebSocketMessage {
	return &WebSocketMessage{
		Type: "system",
		Data: map[string]string{
			"message": message,
		},
	}
}
