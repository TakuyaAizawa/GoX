package interfaces

import (
	"context"

	"github.com/TakuyaAizawa/gox/internal/domain/models"
	"github.com/google/uuid"
)

// UserRepository ユーザーデータアクセスのインターフェースを定義
type UserRepository interface {
	// 新しいユーザーを作成
	Create(ctx context.Context, user *models.User) error

	// IDによるユーザー取得
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)

	// ユーザー名によるユーザー取得
	GetByUsername(ctx context.Context, username string) (*models.User, error)

	// メールアドレスによるユーザー取得
	GetByEmail(ctx context.Context, email string) (*models.User, error)

	// ユーザー情報の更新
	Update(ctx context.Context, user *models.User) error

	// ユーザーの削除
	Delete(ctx context.Context, id uuid.UUID) error

	// ページネーション付きユーザー一覧取得
	List(ctx context.Context, offset, limit int) ([]*models.User, error)

	// 名前またはユーザー名による検索
	Search(ctx context.Context, query string, offset, limit int) ([]*models.User, error)

	// ユーザー名が利用可能か確認
	IsUsernameAvailable(ctx context.Context, username string) (bool, error)

	// メールアドレスが利用可能か確認
	IsEmailAvailable(ctx context.Context, email string) (bool, error)

	// ユーザー総数のカウント
	Count(ctx context.Context) (int64, error)

	// アバター画像URLの更新
	UpdateAvatar(ctx context.Context, userID uuid.UUID, avatarURL string) error

	// バナー画像URLの更新
	UpdateBanner(ctx context.Context, userID uuid.UUID, bannerURL string) error
}
