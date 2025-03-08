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

func TestLikeRepository(t *testing.T) {
	db := testing_helper.NewTestDB(t)
	defer db.Close()

	// テスト開始時にすべてのテーブルをクリーンアップ
	db.CleanupAllTables(t)

	userRepo := NewUserRepository(db.Pool)
	postRepo := NewPostRepository(db.Pool)
	likeRepo := NewLikeRepository(db.Pool)

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

	// Like のテスト
	t.Run("Like", func(t *testing.T) {
		like := models.NewLike(user2.ID, post.ID)
		err := likeRepo.Like(ctx, like)
		require.NoError(t, err)

		// いいね関係の確認
		hasLiked, err := likeRepo.HasLiked(ctx, user2.ID, post.ID)
		require.NoError(t, err)
		assert.True(t, hasLiked)

		// いいね数の確認
		count, err := likeRepo.CountLikesByPostID(ctx, post.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)

		// 投稿のいいね数の確認
		updatedPost, err := postRepo.GetByID(ctx, post.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, updatedPost.LikeCount)
	})

	// Unlike のテスト
	t.Run("Unlike", func(t *testing.T) {
		err := likeRepo.Unlike(ctx, user2.ID, post.ID)
		require.NoError(t, err)

		// いいね関係の確認
		hasLiked, err := likeRepo.HasLiked(ctx, user2.ID, post.ID)
		require.NoError(t, err)
		assert.False(t, hasLiked)

		// いいね数の確認
		count, err := likeRepo.CountLikesByPostID(ctx, post.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)

		// 投稿のいいね数の確認
		updatedPost, err := postRepo.GetByID(ctx, post.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, updatedPost.LikeCount)

		// 存在しないいいね関係の解除を試みる
		err = likeRepo.Unlike(ctx, user2.ID, post.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "like relationship not found")
	})

	// GetLikesByPostID のテスト
	t.Run("GetLikesByPostID", func(t *testing.T) {
		// いいねを作成
		like := models.NewLike(user2.ID, post.ID)
		err := likeRepo.Like(ctx, like)
		require.NoError(t, err)

		// 投稿に対するいいね一覧を取得
		likes, err := likeRepo.GetLikesByPostID(ctx, post.ID, 0, 10)
		require.NoError(t, err)
		assert.Len(t, likes, 1)
		assert.Equal(t, user2.ID, likes[0].UserID)
		assert.Equal(t, post.ID, likes[0].PostID)

		// 存在しない投稿のいいね一覧
		nonexistentID := uuid.New()
		likes, err = likeRepo.GetLikesByPostID(ctx, nonexistentID, 0, 10)
		require.NoError(t, err)
		assert.Empty(t, likes)
	})

	// GetLikesByUserID のテスト
	t.Run("GetLikesByUserID", func(t *testing.T) {
		// ユーザーのいいね一覧を取得
		likes, err := likeRepo.GetLikesByUserID(ctx, user2.ID, 0, 10)
		require.NoError(t, err)
		assert.Len(t, likes, 1)
		assert.Equal(t, user2.ID, likes[0].UserID)
		assert.Equal(t, post.ID, likes[0].PostID)

		// 存在しないユーザーのいいね一覧
		nonexistentID := uuid.New()
		likes, err = likeRepo.GetLikesByUserID(ctx, nonexistentID, 0, 10)
		require.NoError(t, err)
		assert.Empty(t, likes)
	})

	// Count のテスト
	t.Run("Count", func(t *testing.T) {
		// 投稿に対するいいね数の確認
		count, err := likeRepo.CountLikesByPostID(ctx, post.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)

		// ユーザーのいいね総数の確認
		count, err = likeRepo.CountLikesByUserID(ctx, user2.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)

		// 存在しない投稿のいいね数
		nonexistentID := uuid.New()
		count, err = likeRepo.CountLikesByPostID(ctx, nonexistentID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)

		// 存在しないユーザーのいいね数
		count, err = likeRepo.CountLikesByUserID(ctx, nonexistentID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}
