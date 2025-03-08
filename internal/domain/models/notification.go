package models

import (
	"time"

	"github.com/google/uuid"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeLike    NotificationType = "like"
	NotificationTypeFollow  NotificationType = "follow"
	NotificationTypeRepost  NotificationType = "repost"
	NotificationTypeReply   NotificationType = "reply"
	NotificationTypeMention NotificationType = "mention"
)

// Notification represents a notification in the system
type Notification struct {
	ID        uuid.UUID        `json:"id"`
	UserID    uuid.UUID        `json:"user_id"`
	ActorID   uuid.UUID        `json:"actor_id"`
	Type      NotificationType `json:"type"`
	PostID    *uuid.UUID       `json:"post_id,omitempty"`
	IsRead    bool             `json:"is_read"`
	CreatedAt time.Time        `json:"created_at"`

	// APIレスポンス用の関連データ
	Actor *UserResponse `json:"actor,omitempty"`
	Post  *PostResponse `json:"post,omitempty"`
}

// NewNotification creates a new notification with default values
func NewNotification(userID, actorID uuid.UUID, notificationType NotificationType, postID *uuid.UUID) *Notification {
	return &Notification{
		ID:        uuid.New(),
		UserID:    userID,
		ActorID:   actorID,
		Type:      notificationType,
		PostID:    postID,
		IsRead:    false,
		CreatedAt: time.Now().UTC(),
	}
}

// NotificationResponse represents the notification data sent to clients
type NotificationResponse struct {
	ID        uuid.UUID        `json:"id"`
	UserID    uuid.UUID        `json:"user_id"`
	ActorID   uuid.UUID        `json:"actor_id"`
	Type      NotificationType `json:"type"`
	PostID    *uuid.UUID       `json:"post_id,omitempty"`
	IsRead    bool             `json:"is_read"`
	CreatedAt time.Time        `json:"created_at"`
	Actor     *UserResponse    `json:"actor,omitempty"`
	Post      *PostResponse    `json:"post,omitempty"`
}

// ToResponse converts a Notification to NotificationResponse
func (n *Notification) ToResponse() *NotificationResponse {
	return &NotificationResponse{
		ID:        n.ID,
		UserID:    n.UserID,
		ActorID:   n.ActorID,
		Type:      n.Type,
		PostID:    n.PostID,
		IsRead:    n.IsRead,
		CreatedAt: n.CreatedAt,
		Actor:     n.Actor,
		Post:      n.Post,
	}
}
