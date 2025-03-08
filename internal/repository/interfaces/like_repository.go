package interfaces

import (
	"context"

	"github.com/TakuyaAizawa/gox/internal/domain/models"
	"github.com/google/uuid"
)

// LikeRepository いいね関連のデータアクセスのインターフェースを定義
type LikeRepository interface {
	// 投稿にいいねをする
	Like(ctx context.Context, like *models.Like) error

	// いいねを取り消す
	Unlike(ctx context.Context, userID, postID uuid.UUID) error

	// いいね済みかどうかを確認
	HasLiked(ctx context.Context, userID, postID uuid.UUID) (bool, error)

	// 投稿に対するいいね一覧を取得
	GetLikesByPostID(ctx context.Context, postID uuid.UUID, offset, limit int) ([]*models.Like, error)

	// ユーザーがいいねした投稿一覧を取得
	GetLikesByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*models.Like, error)

	// 投稿に対するいいね数を取得
	CountLikesByPostID(ctx context.Context, postID uuid.UUID) (int64, error)

	// ユーザーのいいね総数を取得
	CountLikesByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
}
