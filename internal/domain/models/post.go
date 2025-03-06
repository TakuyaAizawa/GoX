package models

import (
	"time"

	"github.com/google/uuid"
)

// Post represents a post in the system
type Post struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Content     string    `json:"content"`
	MediaURLs   []string  `json:"media_urls"`
	LikeCount   int       `json:"like_count"`
	RepostCount int       `json:"repost_count"`
	ReplyCount  int       `json:"reply_count"`
	IsRepost    bool      `json:"is_repost"`
	RepostID    *uuid.UUID `json:"repost_id,omitempty"`
	IsReply     bool      `json:"is_reply"`
	ReplyToID   *uuid.UUID `json:"reply_to_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewPost creates a new post with default values
func NewPost(userID uuid.UUID, content string, mediaURLs []string) *Post {
	now := time.Now()
	return &Post{
		ID:          uuid.New(),
		UserID:      userID,
		Content:     content,
		MediaURLs:   mediaURLs,
		LikeCount:   0,
		RepostCount: 0,
		ReplyCount:  0,
		IsRepost:    false,
		RepostID:    nil,
		IsReply:     false,
		ReplyToID:   nil,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// NewReply creates a new reply post
func NewReply(userID uuid.UUID, replyToID uuid.UUID, content string, mediaURLs []string) *Post {
	post := NewPost(userID, content, mediaURLs)
	post.IsReply = true
	post.ReplyToID = &replyToID
	return post
}

// NewRepost creates a new repost
func NewRepost(userID uuid.UUID, repostID uuid.UUID, content string) *Post {
	post := NewPost(userID, content, nil)
	post.IsRepost = true
	post.RepostID = &repostID
	return post
}

// PostResponse represents the post data sent to clients
type PostResponse struct {
	ID          uuid.UUID    `json:"id"`
	UserID      uuid.UUID    `json:"user_id"`
	User        *UserResponse `json:"user,omitempty"`
	Content     string       `json:"content"`
	MediaURLs   []string     `json:"media_urls"`
	LikeCount   int          `json:"like_count"`
	RepostCount int          `json:"repost_count"`
	ReplyCount  int          `json:"reply_count"`
	IsRepost    bool         `json:"is_repost"`
	RepostID    *uuid.UUID   `json:"repost_id,omitempty"`
	Repost      *PostResponse `json:"repost,omitempty"`
	IsReply     bool         `json:"is_reply"`
	ReplyToID   *uuid.UUID   `json:"reply_to_id,omitempty"`
	ReplyTo     *PostResponse `json:"reply_to,omitempty"`
	IsLiked     bool         `json:"is_liked"`
	IsReposted  bool         `json:"is_reposted"`
	CreatedAt   time.Time    `json:"created_at"`
}

// ToResponse converts a Post to PostResponse
func (p *Post) ToResponse() *PostResponse {
	return &PostResponse{
		ID:          p.ID,
		UserID:      p.UserID,
		Content:     p.Content,
		MediaURLs:   p.MediaURLs,
		LikeCount:   p.LikeCount,
		RepostCount: p.RepostCount,
		ReplyCount:  p.ReplyCount,
		IsRepost:    p.IsRepost,
		RepostID:    p.RepostID,
		IsReply:     p.IsReply,
		ReplyToID:   p.ReplyToID,
		IsLiked:     false, // このフィールドはサービス層で設定する
		IsReposted:  false, // このフィールドはサービス層で設定する
		CreatedAt:   p.CreatedAt,
	}
} 