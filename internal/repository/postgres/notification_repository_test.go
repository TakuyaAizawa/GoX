package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/TakuyaAizawa/gox/internal/domain/models"
	testing_helper "github.com/TakuyaAizawa/gox/internal/repository/postgres/testing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotificationRepository(t *testing.T) {
	db := testing_helper.NewTestDB(t)
	defer db.Close()

	// テスト開始時にすべてのテーブルをクリーンアップ
	db.CleanupAllTables(t)

	userRepo := NewUserRepository(db.Pool)
	postRepo := NewPostRepository(db.Pool)
	notificationRepo := NewNotificationRepository(db.Pool)

	ctx := context.Background()

	// テストユーザーの作成
	user1 := &models.User{
		ID:           uuid.New(),
		Username:     "user1",
		Email:        "user1@example.com",
		Password:     "hashedpassword",
		Name:         "User 1",
		Bio:          "Test bio",
		ProfileImage: "https://example.com/image.jpg",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	user2 := &models.User{
		ID:           uuid.New(),
		Username:     "user2",
		Email:        "user2@example.com",
		Password:     "hashedpassword",
		Name:         "User 2",
		Bio:          "Test bio",
		ProfileImage: "https://example.com/image.jpg",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	// テストユーザーを作成
	err := userRepo.Create(ctx, user1)
	require.NoError(t, err)
	err = userRepo.Create(ctx, user2)
	require.NoError(t, err)

	// テスト投稿の作成
	post := &models.Post{
		ID:        uuid.New(),
		UserID:    user1.ID,
		Content:   "Test content",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	err = postRepo.Create(ctx, post)
	require.NoError(t, err)

	// Create のテスト
	t.Run("Create", func(t *testing.T) {
		notification := models.NewNotification(user1.ID, user2.ID, models.NotificationTypeLike, &post.ID)
		err := notificationRepo.Create(ctx, notification)
		require.NoError(t, err)

		// 作成された通知を取得して確認
		created, err := notificationRepo.GetByID(ctx, notification.ID)
		require.NoError(t, err)
		assert.Equal(t, notification.ID, created.ID)
		assert.Equal(t, user1.ID, created.UserID)
		assert.Equal(t, user2.ID, created.ActorID)
		assert.Equal(t, models.NotificationTypeLike, created.Type)
		assert.Equal(t, post.ID, *created.PostID)
		assert.False(t, created.IsRead)
	})

	// GetByUserID のテスト
	t.Run("GetByUserID", func(t *testing.T) {
		notifications, err := notificationRepo.GetByUserID(ctx, user1.ID, 0, 10)
		require.NoError(t, err)
		assert.Len(t, notifications, 1)

		// 存在しないユーザーIDでの取得
		notifications, err = notificationRepo.GetByUserID(ctx, uuid.New(), 0, 10)
		require.NoError(t, err)
		assert.Empty(t, notifications)
	})

	// MarkAsRead のテスト
	t.Run("MarkAsRead", func(t *testing.T) {
		notifications, err := notificationRepo.GetByUserID(ctx, user1.ID, 0, 1)
		require.NoError(t, err)
		require.NotEmpty(t, notifications)

		notification := notifications[0]
		err = notificationRepo.MarkAsRead(ctx, notification.ID)
		require.NoError(t, err)

		// 既読状態を確認
		updated, err := notificationRepo.GetByID(ctx, notification.ID)
		require.NoError(t, err)
		assert.True(t, updated.IsRead)

		// 存在しない通知の既読化を試みる
		err = notificationRepo.MarkAsRead(ctx, uuid.New())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "notification not found")
	})

	// MarkAllAsRead のテスト
	t.Run("MarkAllAsRead", func(t *testing.T) {
		// 追加の未読通知を作成
		notification2 := models.NewNotification(user1.ID, user2.ID, models.NotificationTypeFollow, nil)
		err := notificationRepo.Create(ctx, notification2)
		require.NoError(t, err)

		// 全通知を既読化
		err = notificationRepo.MarkAllAsRead(ctx, user1.ID)
		require.NoError(t, err)

		// 未読通知数を確認
		count, err := notificationRepo.CountUnreadByUserID(ctx, user1.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	// GetWithRelations のテスト
	t.Run("GetWithRelations", func(t *testing.T) {
		// 新しいフォロー通知を作成（投稿関連なし）
		followNotification := models.NewNotification(user1.ID, user2.ID, models.NotificationTypeFollow, nil)
		err := notificationRepo.Create(ctx, followNotification)
		require.NoError(t, err)

		// いいね通知を作成（投稿関連あり）
		likeNotification := models.NewNotification(user1.ID, user2.ID, models.NotificationTypeLike, &post.ID)
		err = notificationRepo.Create(ctx, likeNotification)
		require.NoError(t, err)

		// いいね通知を取得して関連データを確認
		notificationWithRelations, err := notificationRepo.GetWithRelations(ctx, likeNotification.ID)
		require.NoError(t, err)

		// 関連データの確認
		assert.NotNil(t, notificationWithRelations)
		assert.NotNil(t, notificationWithRelations.Actor)
		assert.Equal(t, user2.ID, notificationWithRelations.Actor.ID)
		assert.Equal(t, user2.Username, notificationWithRelations.Actor.Username)

		assert.NotNil(t, notificationWithRelations.PostID)
		if notificationWithRelations.Post != nil {
			assert.Equal(t, post.ID, notificationWithRelations.Post.ID)
			assert.Equal(t, post.Content, notificationWithRelations.Post.Content)
		}

		// フォロー通知を取得して関連データを確認
		followWithRelations, err := notificationRepo.GetWithRelations(ctx, followNotification.ID)
		require.NoError(t, err)
		assert.NotNil(t, followWithRelations.Actor)
		assert.Nil(t, followWithRelations.PostID) // 投稿IDがない
		assert.Nil(t, followWithRelations.Post)   // 投稿データがない
	})

	// GetByUserIDWithRelations のテスト
	t.Run("GetByUserIDWithRelations", func(t *testing.T) {
		notifications, err := notificationRepo.GetByUserIDWithRelations(ctx, user1.ID, 0, 10)
		require.NoError(t, err)
		assert.NotEmpty(t, notifications)

		// 各通知をチェック
		for _, notification := range notifications {
			// アクター（通知の送信者）の情報を確認
			assert.NotNil(t, notification.Actor)
			assert.Equal(t, user2.ID, notification.Actor.ID)
			assert.Equal(t, user2.Username, notification.Actor.Username)

			// タイプに応じてPost情報を確認
			if notification.Type == models.NotificationTypeLike && notification.PostID != nil {
				if notification.Post != nil {
					assert.Equal(t, post.ID, notification.Post.ID)
					assert.Equal(t, post.Content, notification.Post.Content)
				}
			}

			if notification.Type == models.NotificationTypeFollow {
				assert.Nil(t, notification.PostID)
			}
		}

		// 存在しないユーザーIDでの取得
		notifications, err = notificationRepo.GetByUserIDWithRelations(ctx, uuid.New(), 0, 10)
		require.NoError(t, err)
		assert.Empty(t, notifications)
	})

	// Delete のテスト
	t.Run("Delete", func(t *testing.T) {
		notifications, err := notificationRepo.GetByUserID(ctx, user1.ID, 0, 1)
		require.NoError(t, err)
		require.NotEmpty(t, notifications)

		notification := notifications[0]
		err = notificationRepo.Delete(ctx, notification.ID)
		require.NoError(t, err)

		// 削除を確認
		_, err = notificationRepo.GetByID(ctx, notification.ID)
		assert.Error(t, err)

		// 存在しない通知の削除を試みる
		err = notificationRepo.Delete(ctx, uuid.New())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "notification not found")
	})

	// CountUnreadByUserID のテスト
	t.Run("CountUnreadByUserID", func(t *testing.T) {
		// まず既存の通知をすべて既読にする
		err = notificationRepo.MarkAllAsRead(ctx, user1.ID)
		require.NoError(t, err)

		// 現在の未読通知がゼロであることを確認
		count, err := notificationRepo.CountUnreadByUserID(ctx, user1.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)

		// 新しい未読通知を作成
		notification := models.NewNotification(user1.ID, user2.ID, models.NotificationTypeLike, &post.ID)
		err = notificationRepo.Create(ctx, notification)
		require.NoError(t, err)

		// 未読通知数を確認
		count, err = notificationRepo.CountUnreadByUserID(ctx, user1.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)

		// 存在しないユーザーの未読通知数
		count, err = notificationRepo.CountUnreadByUserID(ctx, uuid.New())
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}
