package interfaces

import (
	"context"

	"github.com/TakuyaAizawa/gox/internal/domain/models"
	"github.com/google/uuid"
)

// NotificationRepository 通知関連のデータアクセスのインターフェースを定義
type NotificationRepository interface {
	// 通知を作成
	Create(ctx context.Context, notification *models.Notification) error

	// IDによる通知取得
	GetByID(ctx context.Context, id uuid.UUID) (*models.Notification, error)

	// ユーザーIDによる通知一覧取得
	GetByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*models.Notification, error)

	// 通知を既読にする
	MarkAsRead(ctx context.Context, id uuid.UUID) error

	// ユーザーの全通知を既読にする
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error

	// 通知の削除
	Delete(ctx context.Context, id uuid.UUID) error

	// ユーザーの未読通知数を取得
	CountUnreadByUserID(ctx context.Context, userID uuid.UUID) (int64, error)

	// 通知を取得して関連データ（Actor, Post）を含める
	GetWithRelations(ctx context.Context, id uuid.UUID) (*models.Notification, error)

	// ユーザーIDによる通知一覧を取得して関連データを含める
	GetByUserIDWithRelations(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*models.Notification, error)
}
