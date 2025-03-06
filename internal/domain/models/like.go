package models

import (
	"time"
	"github.com/google/uuid"
)

// Like represents a like in the system
type Like struct {
	UserID    uuid.UUID `json:"user_id"`
	PostID    uuid.UUID `json:"post_id"`
	CreatedAt time.Time `json:"created_at"`
}

// NewLike creates a new like with default values
func NewLike(userID, postID uuid.UUID) *Like {
	return &Like{
		UserID:    userID,
		PostID:    postID,
		CreatedAt: time.Now().UTC(),
	}
}