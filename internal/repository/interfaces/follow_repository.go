package interfaces

import (
	"context"

	"github.com/google/uuid"
)

// FollowRepository フォロー関連のデータアクセスのインターフェースを定義
type FollowRepository interface {
	// フォローする
	Follow(ctx context.Context, followerID, followeeID uuid.UUID) error

	// フォロー解除する
	Unfollow(ctx context.Context, followerID, followeeID uuid.UUID) error

	// フォロー中かどうかを確認
	IsFollowing(ctx context.Context, followerID, followeeID uuid.UUID) (bool, error)

	// フォロワー一覧を取得
	GetFollowers(ctx context.Context, userID uuid.UUID, offset, limit int) ([]uuid.UUID, error)

	// フォロー中のユーザー一覧を取得
	GetFollowing(ctx context.Context, userID uuid.UUID, offset, limit int) ([]uuid.UUID, error)

	// フォロワー数を取得
	CountFollowers(ctx context.Context, userID uuid.UUID) (int64, error)

	// フォロー中のユーザー数を取得
	CountFollowing(ctx context.Context, userID uuid.UUID) (int64, error)
}
