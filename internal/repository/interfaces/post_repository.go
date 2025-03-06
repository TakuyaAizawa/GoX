package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/TakuyaAizawa/gox/internal/domain/models"
)

// PostRepository 投稿データアクセスのインターフェースを定義
type PostRepository interface {
	// 新しい投稿を作成
	Create(ctx context.Context, post *models.Post) error
	
	// IDによる投稿取得
	GetByID(ctx context.Context, id uuid.UUID) (*models.Post, error)
	
	// 投稿の更新
	Update(ctx context.Context, post *models.Post) error
	
	// 投稿の削除
	Delete(ctx context.Context, id uuid.UUID) error
	
	// ページネーション付き投稿一覧取得
	List(ctx context.Context, offset, limit int) ([]*models.Post, error)
	
	// ユーザーIDによる投稿取得
	GetByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*models.Post, error)
	
	// 投稿への返信を取得
	GetReplies(ctx context.Context, postID uuid.UUID, offset, limit int) ([]*models.Post, error)
	
	// 投稿のリポスト（再投稿）を取得
	GetReposts(ctx context.Context, postID uuid.UUID, offset, limit int) ([]*models.Post, error)
	
	// ユーザーIDによる投稿数のカウント
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
	
	// 投稿への返信数のカウント
	CountReplies(ctx context.Context, postID uuid.UUID) (int64, error)
	
	// 投稿のリポスト数のカウント
	CountReposts(ctx context.Context, postID uuid.UUID) (int64, error)
	
	// いいね数を増加
	IncrementLikeCount(ctx context.Context, postID uuid.UUID) error
	
	// いいね数を減少
	DecrementLikeCount(ctx context.Context, postID uuid.UUID) error
	
	// リポスト数を増加
	IncrementRepostCount(ctx context.Context, postID uuid.UUID) error
	
	// リポスト数を減少
	DecrementRepostCount(ctx context.Context, postID uuid.UUID) error
	
	// 返信数を増加
	IncrementReplyCount(ctx context.Context, postID uuid.UUID) error
	
	// 返信数を減少
	DecrementReplyCount(ctx context.Context, postID uuid.UUID) error
} 