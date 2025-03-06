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

func TestFollowRepository(t *testing.T) {
	db := testing_helper.NewTestDB(t)
	defer db.Close()

	// テスト開始時にすべてのテーブルをクリーンアップ
	db.CleanupAllTables(t)

	userRepo := NewUserRepository(db.Pool)
	followRepo := NewFollowRepository(db.Pool)

	ctx := context.Background()

	// テストユーザーの作成
	user1 := &models.User{
		ID:             uuid.New(),
		Username:       "user1",
		Email:          "user1@example.com",
		Password:       "hashedpassword",
		Name:           "User 1",
		Bio:            "Test bio",
		ProfileImage:   "https://example.com/image.jpg",
		FollowerCount:  0,
		FollowingCount: 0,
		PostCount:      0,
		IsVerified:     false,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	user2 := &models.User{
		ID:             uuid.New(),
		Username:       "user2",
		Email:          "user2@example.com",
		Password:       "hashedpassword",
		Name:           "User 2",
		Bio:            "Test bio",
		ProfileImage:   "https://example.com/image.jpg",
		FollowerCount:  0,
		FollowingCount: 0,
		PostCount:      0,
		IsVerified:     false,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	// テストユーザーを作成
	err := userRepo.Create(ctx, user1)
	require.NoError(t, err)
	err = userRepo.Create(ctx, user2)
	require.NoError(t, err)

	// Follow のテスト
	t.Run("Follow", func(t *testing.T) {
		err := followRepo.Follow(ctx, user1.ID, user2.ID)
		require.NoError(t, err)

		// フォロー関係の確認
		isFollowing, err := followRepo.IsFollowing(ctx, user1.ID, user2.ID)
		require.NoError(t, err)
		assert.True(t, isFollowing)

		// カウントの確認
		updatedUser1, err := userRepo.GetByID(ctx, user1.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, updatedUser1.FollowingCount)

		updatedUser2, err := userRepo.GetByID(ctx, user2.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, updatedUser2.FollowerCount)

		// 自分自身をフォローできないことを確認
		err = followRepo.Follow(ctx, user1.ID, user1.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot follow yourself")
	})

	// Unfollow のテスト
	t.Run("Unfollow", func(t *testing.T) {
		err := followRepo.Unfollow(ctx, user1.ID, user2.ID)
		require.NoError(t, err)

		// フォロー関係の確認
		isFollowing, err := followRepo.IsFollowing(ctx, user1.ID, user2.ID)
		require.NoError(t, err)
		assert.False(t, isFollowing)

		// カウントの確認
		updatedUser1, err := userRepo.GetByID(ctx, user1.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, updatedUser1.FollowingCount)

		updatedUser2, err := userRepo.GetByID(ctx, user2.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, updatedUser2.FollowerCount)

		// 存在しないフォロー関係の解除を試みる
		err = followRepo.Unfollow(ctx, user1.ID, user2.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "follow relationship not found")
	})

	// GetFollowers のテスト
	t.Run("GetFollowers", func(t *testing.T) {
		// フォロー関係を作成
		err := followRepo.Follow(ctx, user1.ID, user2.ID)
		require.NoError(t, err)

		// フォロワー一覧を取得
		followers, err := followRepo.GetFollowers(ctx, user2.ID, 0, 10)
		require.NoError(t, err)
		assert.Len(t, followers, 1)
		assert.Equal(t, user1.ID, followers[0])

		// 存在しないユーザーのフォロワー一覧
		nonexistentID := uuid.New()
		followers, err = followRepo.GetFollowers(ctx, nonexistentID, 0, 10)
		require.NoError(t, err)
		assert.Empty(t, followers)
	})

	// GetFollowing のテスト
	t.Run("GetFollowing", func(t *testing.T) {
		// フォロー中一覧を取得
		following, err := followRepo.GetFollowing(ctx, user1.ID, 0, 10)
		require.NoError(t, err)
		assert.Len(t, following, 1)
		assert.Equal(t, user2.ID, following[0])

		// 存在しないユーザーのフォロー中一覧
		nonexistentID := uuid.New()
		following, err = followRepo.GetFollowing(ctx, nonexistentID, 0, 10)
		require.NoError(t, err)
		assert.Empty(t, following)
	})

	// IsFollowing のテスト
	t.Run("IsFollowing", func(t *testing.T) {
		// フォロー中の確認
		isFollowing, err := followRepo.IsFollowing(ctx, user1.ID, user2.ID)
		require.NoError(t, err)
		assert.True(t, isFollowing)

		// フォローしていない関係の確認
		isFollowing, err = followRepo.IsFollowing(ctx, user2.ID, user1.ID)
		require.NoError(t, err)
		assert.False(t, isFollowing)
	})

	// Count のテスト
	t.Run("Count", func(t *testing.T) {
		// フォロワー数の確認
		count, err := followRepo.CountFollowers(ctx, user2.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)

		// フォロー中数の確認
		count, err = followRepo.CountFollowing(ctx, user1.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)

		// 存在しないユーザーのカウント
		nonexistentID := uuid.New()
		count, err = followRepo.CountFollowers(ctx, nonexistentID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)

		count, err = followRepo.CountFollowing(ctx, nonexistentID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}
